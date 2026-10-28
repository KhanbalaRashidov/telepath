package telegram

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// telegramMaxLen is Telegram's hard message length limit. Content that would
// exceed it (large diffs) is sent as a document instead.
const telegramMaxLen = 4096

// escapeHTML escapes the three characters significant to Telegram's HTML parse
// mode. This is the only escaping needed for <pre>/<b> payloads.
func escapeHTML(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}

// preBlock wraps text in an escaped <pre> code block.
func preBlock(s string) string {
	return "<pre>" + escapeHTML(s) + "</pre>"
}

// truncate shortens s to n runes with an ellipsis.
func truncate(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "…"
}

// taskView holds the fields rendered into the single live status message.
type taskView struct {
	agent     string
	branch    string
	action    string
	file      string
	cmd       string
	committed string
	state     string // "" | "done" | "error" | "cancelled"
}

// render builds the one-line status (escaped HTML).
func (v taskView) render() string {
	var parts []string
	parts = append(parts, "🤖 "+escapeHTML(v.agent))
	if v.branch != "" {
		parts = append(parts, "🌿 "+escapeHTML(truncate(v.branch, 40)))
	}
	switch v.state {
	case "done":
		parts = append(parts, "✅ done")
	case "error":
		parts = append(parts, "❌ error")
	case "cancelled":
		parts = append(parts, "🛑 cancelled")
	default:
		if v.file != "" {
			parts = append(parts, "✏️ "+escapeHTML(truncate(baseName(v.file), 32)))
		}
		if v.cmd != "" {
			parts = append(parts, "▶ "+escapeHTML(truncate(v.cmd, 40)))
		}
		if v.action != "" && v.file == "" && v.cmd == "" {
			parts = append(parts, "💭 "+escapeHTML(truncate(v.action, 60)))
		}
	}
	if v.committed != "" {
		parts = append(parts, "✅ committed "+escapeHTML(v.committed))
	}
	return strings.Join(parts, " · ")
}

func baseName(p string) string {
	p = strings.TrimRight(p, "/\\")
	if i := strings.LastIndexAny(p, "/\\"); i >= 0 {
		return p[i+1:]
	}
	return p
}

// fileEditMessage builds the message body for a FileEdit, and reports whether it
// must be sent as a document because it exceeds the length limit.
func fileEditMessage(path, diff string) (body string, asDocument bool) {
	header := "✏️ <b>" + escapeHTML(path) + "</b>\n"
	body = header + preBlock(diff)
	if len([]rune(body)) > telegramMaxLen {
		return header, true
	}
	return body, false
}

// --- Bot send/edit wrappers ------------------------------------------------

func (bt *Bot) send(ctx context.Context, chatID int64, html string, markup models.ReplyMarkup) (int, error) {
	m, err := bt.api.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:             chatID,
		Text:               html,
		ParseMode:          models.ParseModeHTML,
		ReplyMarkup:        markup,
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: ptrBool(true)},
	})
	if err != nil {
		bt.log.Warn("send message failed", "chat", chatID, "err", err)
		return 0, err
	}
	return m.ID, nil
}

func (bt *Bot) sendText(ctx context.Context, chatID int64, html string) {
	_, _ = bt.send(ctx, chatID, html, nil)
}

func (bt *Bot) edit(ctx context.Context, chatID int64, msgID int, html string, markup models.ReplyMarkup) {
	_, err := bt.api.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:             chatID,
		MessageID:          msgID,
		Text:               html,
		ParseMode:          models.ParseModeHTML,
		ReplyMarkup:        markup,
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: ptrBool(true)},
	})
	if err != nil {
		// "message is not modified" is benign; log others at debug.
		bt.log.Debug("edit message failed", "chat", chatID, "msg", msgID, "err", err)
	}
}

func (bt *Bot) sendDoc(ctx context.Context, chatID int64, filename string, data []byte, caption string) {
	_, err := bt.api.SendDocument(ctx, &bot.SendDocumentParams{
		ChatID:    chatID,
		Document:  &models.InputFileUpload{Filename: filename, Data: bytes.NewReader(data)},
		Caption:   caption,
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		bt.log.Warn("send document failed", "chat", chatID, "err", err)
		// Fall back to a short text note so the user is not left in the dark.
		bt.sendText(ctx, chatID, "⚠️ could not send document: "+escapeHTML(err.Error()))
	}
}

func ptrBool(b bool) *bool { return &b }

// commitMessage builds the task-branch commit message: a short summary plus the
// Telepath trailer naming the agent.
func commitMessage(prompt, agentName string) string {
	summary := strings.TrimSpace(strings.SplitN(prompt, "\n", 2)[0])
	summary = truncate(summary, 60)
	if summary == "" {
		summary = "Telepath task"
	}
	return fmt.Sprintf("%s\n\nvia Telepath (%s)", summary, agentName)
}

// humanDuration renders a duration compactly.
func humanDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	d = d.Round(time.Second)
	return d.String()
}
