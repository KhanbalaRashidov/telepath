package telegram

import (
	"bytes"
	"context"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/telepath/telepath/internal/transport"
)

// escapeHTML escapes the three characters significant to Telegram's HTML parse
// mode. This is the only escaping needed for <pre>/<b>/<i>/<code> payloads.
func escapeHTML(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}

// renderHTML renders a transport.Doc as Telegram HTML.
func renderHTML(d transport.Doc) string {
	lines := make([]string, 0, len(d.Blocks))
	for _, b := range d.Blocks {
		switch {
		case b.Blank:
			lines = append(lines, "")
		case b.Pre != "":
			lines = append(lines, "<pre>"+escapeHTML(b.Pre)+"</pre>")
		default:
			var sb strings.Builder
			for _, sp := range b.Spans {
				txt := escapeHTML(sp.Text)
				switch {
				case sp.Bold:
					sb.WriteString("<b>" + txt + "</b>")
				case sp.Italic:
					sb.WriteString("<i>" + txt + "</i>")
				case sp.Code:
					sb.WriteString("<code>" + txt + "</code>")
				default:
					sb.WriteString(txt)
				}
			}
			lines = append(lines, sb.String())
		}
	}
	return strings.Join(lines, "\n")
}

// --- Telegram send/edit wrappers -------------------------------------------

func (t *Transport) send(ctx context.Context, chatID int64, html string, markup models.ReplyMarkup) (int, error) {
	m, err := t.api.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:             chatID,
		Text:               html,
		ParseMode:          models.ParseModeHTML,
		ReplyMarkup:        markup,
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: ptrBool(true)},
	})
	if err != nil {
		t.log.Warn("send message failed", "chat", chatID, "err", err)
		return 0, err
	}
	return m.ID, nil
}

func (t *Transport) sendText(ctx context.Context, chatID int64, html string) {
	_, _ = t.send(ctx, chatID, html, nil)
}

func (t *Transport) edit(ctx context.Context, chatID int64, msgID int, html string, markup models.ReplyMarkup) {
	_, err := t.api.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:             chatID,
		MessageID:          msgID,
		Text:               html,
		ParseMode:          models.ParseModeHTML,
		ReplyMarkup:        markup,
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: ptrBool(true)},
	})
	if err != nil {
		t.log.Debug("edit message failed", "chat", chatID, "msg", msgID, "err", err)
	}
}

func (t *Transport) sendDoc(ctx context.Context, chatID int64, filename string, data []byte, caption string) {
	_, err := t.api.SendDocument(ctx, &bot.SendDocumentParams{
		ChatID:    chatID,
		Document:  &models.InputFileUpload{Filename: filename, Data: bytes.NewReader(data)},
		Caption:   caption,
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		t.log.Warn("send document failed", "chat", chatID, "err", err)
		t.sendText(ctx, chatID, "⚠️ could not send document: "+escapeHTML(err.Error()))
	}
}

func ptrBool(b bool) *bool { return &b }
