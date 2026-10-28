// Package telegram is the agent-independent presentation layer: a long-polling
// Telegram bot that consumes only the normalized agent.Event stream and the
// Agent interface. It routes commands, runs tasks on per-task git branches,
// renders a live status message, and brokers interactive approvals.
package telegram

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/telepath/telepath/internal/agent"
	"github.com/telepath/telepath/internal/config"
	"github.com/telepath/telepath/internal/permission"
	"github.com/telepath/telepath/internal/session"
)

// Bot is the Telegram bridge.
type Bot struct {
	api     *bot.Bot
	log     *slog.Logger
	cfg     *config.Config
	reg     *agent.Registry
	store   session.Store
	broker  *permission.Broker
	version string

	allowed map[int64]bool

	rootCtx context.Context

	// taskSlot serializes tasks: all chats share one ProjectsDir working tree,
	// so only one task may touch the repo at a time.
	taskSlot chan struct{}

	approvalMu  sync.Mutex
	approvalMsg map[string]approvalRef
}

type approvalRef struct {
	chatID int64
	msgID  int
}

// New constructs the bot and registers the update handler.
func New(cfg *config.Config, reg *agent.Registry, store session.Store, broker *permission.Broker, log *slog.Logger, version string) (*Bot, error) {
	bt := &Bot{
		log:         log,
		cfg:         cfg,
		reg:         reg,
		store:       store,
		broker:      broker,
		version:     version,
		allowed:     make(map[int64]bool, len(cfg.AllowedChatIDs)),
		taskSlot:    make(chan struct{}, 1),
		approvalMsg: make(map[string]approvalRef),
	}
	for _, id := range cfg.AllowedChatIDs {
		bt.allowed[id] = true
	}

	api, err := bot.New(cfg.BotToken, bot.WithDefaultHandler(bt.handleUpdate))
	if err != nil {
		return nil, err
	}
	bt.api = api
	return bt, nil
}

// Run starts long-polling and blocks until ctx is cancelled.
func (bt *Bot) Run(ctx context.Context) {
	bt.rootCtx = ctx
	bt.log.Info("telepath started", "version", bt.version, "agents", bt.reg.Names(), "default", bt.reg.DefaultName())
	bt.api.Start(ctx)
}

// handleUpdate is the single entry point; it enforces the chat allowlist on
// every update (messages AND callback queries) before anything else.
func (bt *Bot) handleUpdate(ctx context.Context, _ *bot.Bot, update *models.Update) {
	switch {
	case update.Message != nil:
		chatID := update.Message.Chat.ID
		if !bt.allowed[chatID] {
			bt.log.Warn("rejected message from non-allowlisted chat", "chat", chatID)
			return
		}
		bt.handleMessage(ctx, update.Message)
	case update.CallbackQuery != nil:
		cq := update.CallbackQuery
		var chatID int64
		if cq.Message.Message != nil {
			chatID = cq.Message.Message.Chat.ID
		}
		if !bt.allowed[chatID] {
			bt.log.Warn("rejected callback from non-allowlisted chat", "chat", chatID)
			return
		}
		bt.handleCallback(ctx, cq)
	}
}

// --- interactive approval --------------------------------------------------

// makeApprovalHook returns an ApprovalFunc that sends a Telegram prompt with
// Approve/Deny buttons and blocks on the user's tap (or the broker timeout).
func (bt *Bot) makeApprovalHook(chatID int64) agent.ApprovalFunc {
	return func(ctx context.Context, req agent.ApprovalRequest) agent.ApprovalDecision {
		id := newID()
		msgID, _ := bt.send(ctx, chatID, approvalPrompt(req), approvalKeyboard(id))

		bt.approvalMu.Lock()
		bt.approvalMsg[id] = approvalRef{chatID: chatID, msgID: msgID}
		bt.approvalMu.Unlock()

		decision := bt.broker.Await(ctx, id)

		// Single editor of the prompt message: show the final outcome.
		if msgID != 0 {
			bt.edit(ctx, chatID, msgID, approvalResolved(req, decision), nil)
		}
		bt.approvalMu.Lock()
		delete(bt.approvalMsg, id)
		bt.approvalMu.Unlock()

		return agent.ApprovalDecision{Allow: decision.Allow, Note: decision.Note}
	}
}

func newID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
