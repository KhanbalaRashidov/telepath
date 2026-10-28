package slack

import (
	"strings"

	slackgo "github.com/slack-go/slack"

	"github.com/telepath/telepath/internal/transport"
)

// slackSectionMax is Slack's ~3000-char limit for a section block's text.
const slackSectionMax = 2900

// escapeMrkdwn escapes the characters Slack treats specially in message text.
func escapeMrkdwn(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}

// renderMrkdwn renders a transport.Doc as Slack mrkdwn.
func renderMrkdwn(d transport.Doc) string {
	lines := make([]string, 0, len(d.Blocks))
	for _, b := range d.Blocks {
		switch {
		case b.Blank:
			lines = append(lines, "")
		case b.Pre != "":
			lines = append(lines, "```\n"+escapeMrkdwn(b.Pre)+"\n```")
		default:
			var sb strings.Builder
			for _, sp := range b.Spans {
				txt := escapeMrkdwn(sp.Text)
				switch {
				case sp.Bold:
					sb.WriteString("*" + txt + "*")
				case sp.Italic:
					sb.WriteString("_" + txt + "_")
				case sp.Code:
					sb.WriteString("`" + txt + "`")
				default:
					sb.WriteString(txt)
				}
			}
			lines = append(lines, sb.String())
		}
	}
	return strings.Join(lines, "\n")
}

func renderStatusMrkdwn(v transport.StatusView) string {
	return escapeMrkdwn(transport.RenderStatusText(v))
}

func truncateRunes(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "…"
}

func section(text string) *slackgo.SectionBlock {
	return slackgo.NewSectionBlock(
		slackgo.NewTextBlockObject(slackgo.MarkdownType, truncateRunes(text, slackSectionMax), false, false),
		nil, nil,
	)
}

func button(actionID, value, label string) *slackgo.ButtonBlockElement {
	return slackgo.NewButtonBlockElement(actionID, value, slackgo.NewTextBlockObject(slackgo.PlainTextType, label, true, false))
}

// approvalBlocks builds the Approve/Deny Block Kit message for an approval id.
func approvalBlocks(id, text string) []slackgo.Block {
	return []slackgo.Block{
		section(text),
		slackgo.NewActionBlock("", button("ap:"+id+":a", "a", "✅ Approve"), button("ap:"+id+":d", "d", "⛔ Deny")),
	}
}

// branchBlocks builds the post-task Keep / Delete-&-restore Block Kit message.
func branchBlocks(text string) []slackgo.Block {
	return []slackgo.Block{
		section(text),
		slackgo.NewActionBlock("", button("br:keep", "keep", "🌿 Keep branch"), button("br:del", "del", "🗑 Delete branch & restore base")),
	}
}
