package telegram

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// handleCallback resolves inline-button taps: approval decisions (resolved
// against the shared permission broker) and the post-task Keep / Delete-&-
// restore branch actions (delegated to the engine).
func (t *Transport) handleCallback(ctx context.Context, cq *models.CallbackQuery) {
	data := cq.Data
	var chatID int64
	var msgID int
	if cq.Message.Message != nil {
		chatID = cq.Message.Message.Chat.ID
		msgID = cq.Message.Message.ID
	}

	switch {
	case strings.HasPrefix(data, cbApprovePrefix):
		rest := strings.TrimPrefix(data, cbApprovePrefix)
		parts := strings.Split(rest, ":")
		if len(parts) != 2 {
			t.answer(ctx, cq.ID, "Malformed request")
			return
		}
		id, allow := parts[0], parts[1] == "a"
		note := "approved by user"
		if !allow {
			note = "denied by user"
		}
		if t.broker.Resolve(id, allow, note) {
			if allow {
				t.answer(ctx, cq.ID, "✅ Approved")
			} else {
				t.answer(ctx, cq.ID, "⛔ Denied")
			}
		} else {
			t.answer(ctx, cq.ID, "⌛ This request already expired.")
		}

	case data == cbBranchKeep:
		res := t.router.HandleBranchAction(ctx, convRef(chatID), "keep")
		t.answer(ctx, cq.ID, res.Toast)
		if res.Replace != nil && msgID != 0 {
			t.edit(ctx, chatID, msgID, renderHTML(*res.Replace), nil)
		}

	case data == cbBranchDelete:
		res := t.router.HandleBranchAction(ctx, convRef(chatID), "delete")
		t.answer(ctx, cq.ID, res.Toast)
		if res.Replace != nil && msgID != 0 {
			t.edit(ctx, chatID, msgID, renderHTML(*res.Replace), nil)
		}

	default:
		t.answer(ctx, cq.ID, "")
	}
}

func (t *Transport) answer(ctx context.Context, cqID, text string) {
	_, _ = t.api.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: cqID,
		Text:            text,
	})
}

func newID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
