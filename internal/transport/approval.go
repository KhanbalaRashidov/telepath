package transport

import (
	"fmt"

	"github.com/telepath/telepath/internal/agent"
)

// ApprovalPromptDoc builds the shared approval-request message (identical
// content across transports; each renders it in its own markup).
func ApprovalPromptDoc(req agent.ApprovalRequest) Doc {
	d := NewDoc()
	d.Line(Text(RiskEmoji(req.Risk)+" "), Bold("Approval needed"))
	d.Line(Text("Tool: "), Code(req.Tool))
	d.Line(Text("Risk: "), Bold(req.Risk))
	if detail := approvalDetail(req); detail != "" {
		d.Pre(detail)
	}
	return *d
}

// ApprovalResolvedDoc builds the message the prompt is edited into once decided.
func ApprovalResolvedDoc(req agent.ApprovalRequest, allow bool, note string) Doc {
	d := NewDoc()
	prefix, word := "⛔ ", "Denied"
	if allow {
		prefix, word = "✅ ", "Approved"
	}
	d.Line(Text(prefix), Bold(word), Text(" — "), Code(req.Tool))
	if note != "" {
		d.Line(Italic(note))
	}
	return *d
}

// RiskEmoji maps a coarse risk label to an indicator.
func RiskEmoji(risk string) string {
	switch risk {
	case "high":
		return "🔴"
	case "low":
		return "🟢"
	default:
		return "🟡"
	}
}

func approvalDetail(req agent.ApprovalRequest) string {
	if cmd, ok := req.Input["command"].(string); ok && cmd != "" {
		return cmd
	}
	for _, k := range []string{"file_path", "path", "url", "pattern"} {
		if v, ok := req.Input[k]; ok {
			return fmt.Sprintf("%s = %v", k, v)
		}
	}
	return ""
}
