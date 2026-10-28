// Package telegram implements the Telegram transport: a long-polling bot that
// turns Telegram updates into the engine's normalized inbound messages, and
// renders the engine's outbound Docs/status/diffs/approvals as Telegram HTML
// messages with inline-keyboard buttons. All task/command logic lives in the
// transport-agnostic engine; this package only does Telegram I/O and formatting.
package telegram

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/telepath/telepath/internal/permission"
	"github.com/telepath/telepath/internal/transport"
)

const transportName = "telegram"

// Transport is the Telegram implementation of transport.Transport.
type Transport struct {
	api     *bot.Bot
	log     *slog.Logger
	router  transport.Router
	broker  *permission.Broker
	allowed map[int64]bool
}

// New builds the Telegram transport. It registers the update handler but does
// not start polling until Start is called.
func New(token string, allowedChatIDs []int64, router transport.Router, broker *permission.Broker, log *slog.Logger) (*Transport, error) {
	t := &Transport{
		log:     log,
		router:  router,
		broker:  broker,
		allowed: make(map[int64]bool, len(allowedChatIDs)),
	}
	for _, id := range allowedChatIDs {
		t.allowed[id] = true
	}
	api, err := bot.New(token, bot.WithDefaultHandler(t.handleUpdate))
	if err != nil {
		return nil, err
	}
	t.api = api
	return t, nil
}

func (t *Transport) Name() string { return transportName }

// Start begins long-polling and blocks until ctx is cancelled.
func (t *Transport) Start(ctx context.Context) error {
	t.api.Start(ctx)
	return nil
}

// convRef builds a ConversationRef for a Telegram chat.
func convRef(chatID int64) transport.ConversationRef {
	return transport.ConversationRef{Transport: transportName, ConversationID: strconv.FormatInt(chatID, 10)}
}

// Reply renders a Doc and sends it to the conversation.
func (t *Transport) Reply(ctx context.Context, conv transport.ConversationRef, doc transport.Doc) {
	chatID, err := strconv.ParseInt(conv.ConversationID, 10, 64)
	if err != nil {
		return
	}
	t.sendText(ctx, chatID, renderHTML(doc))
}

// NewSession creates a per-task outbound session for a conversation.
func (t *Transport) NewSession(conv transport.ConversationRef) transport.Session {
	chatID, _ := strconv.ParseInt(conv.ConversationID, 10, 64)
	return &tgSession{t: t, conv: conv, chatID: chatID}
}

// handleUpdate enforces the chat allowlist on every update (messages AND
// callback queries) before routing.
func (t *Transport) handleUpdate(ctx context.Context, _ *bot.Bot, update *models.Update) {
	switch {
	case update.Message != nil:
		chatID := update.Message.Chat.ID
		if !t.allowed[chatID] {
			t.log.Warn("rejected message from non-allowlisted chat", "chat", chatID)
			return
		}
		var userID string
		if update.Message.From != nil {
			userID = strconv.FormatInt(update.Message.From.ID, 10)
		}
		t.router.HandleMessage(ctx, transport.InboundMessage{
			Conversation: convRef(chatID),
			UserID:       userID,
			Text:         update.Message.Text,
		})
	case update.CallbackQuery != nil:
		cq := update.CallbackQuery
		var chatID int64
		if cq.Message.Message != nil {
			chatID = cq.Message.Message.Chat.ID
		}
		if !t.allowed[chatID] {
			t.log.Warn("rejected callback from non-allowlisted chat", "chat", chatID)
			return
		}
		t.handleCallback(ctx, cq)
	}
}
