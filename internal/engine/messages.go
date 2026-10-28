package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/telepath/telepath/internal/agent"
	"github.com/telepath/telepath/internal/transport"
)

// This file builds all engine-produced messages as transport.Doc values. The
// wording mirrors Telepath's original Telegram text 1:1 so behavior is
// unchanged; each transport renders the Doc in its own markup.

func (e *Engine) helpDoc() transport.Doc {
	d := transport.NewDoc()
	d.Line(transport.Text("🤖 "), transport.Bold("Telepath"), transport.Text(" — Telegram/Slack ↔ coding-agent bridge"))
	d.Blank()
	d.Line(transport.Bold("Commands"))
	d.Line(transport.Text("/help — this help"))
	d.Line(transport.Text("/agents — list agents and availability"))
	d.Line(transport.Text("/use <agent> — set this conversation's default agent"))
	d.Line(transport.Text("/perm <auto|allowlist|interactive> — permission mode"))
	d.Line(transport.Text("/branch <off|per-task> — branch workflow"))
	d.Line(transport.Text("/<agent> <task> — run with a specific agent"))
	d.Line(transport.Text("<message> — run with the default agent"))
	d.Line(transport.Text("/resume <text> — continue the last session"))
	d.Line(transport.Text("/branches — list telepath/* branches"))
	d.Line(transport.Text("/status — current task + mode summary"))
	d.Line(transport.Text("/cancel — cancel the running task"))
	d.Blank()
	d.Line(transport.Text("Enabled agents: "), transport.Bold(strings.Join(e.reg.Names(), ", ")))
	d.Line(transport.Text("Telepath never pushes, merges, or touches main/master — it leaves a reviewable branch."))
	return *d
}

func (e *Engine) agentsDoc(ctx context.Context) transport.Doc {
	d := transport.NewDoc()
	d.Line(transport.Text("🧩 "), transport.Bold("Agents"))
	def := e.reg.DefaultName()
	for _, ag := range e.reg.List() {
		caps := ag.Capabilities()
		avail := "❌ not installed"
		if ag.Available(ctx) {
			avail = "✅ installed"
		}
		marker := ""
		if ag.Name() == def {
			marker = " (default)"
		}
		d.Blank()
		d.Line(transport.Bold(ag.Name()), transport.Text(marker+" — "+avail))
		d.Line(transport.Text("  stream " + yesno(caps.StructuredStream) +
			" · resume " + yesno(caps.Resume) +
			" · interactive " + yesno(caps.InteractiveApproval) +
			" · file-edits " + yesno(caps.FileEditEvents)))
	}
	return *d
}

func yesno(b bool) string {
	if b {
		return "✅"
	}
	return "—"
}

func plain(s string) transport.Doc {
	d := transport.NewDoc()
	d.Line(transport.Text(s))
	return *d
}

func (e *Engine) branchesDoc(prefix string, branches []string) transport.Doc {
	d := transport.NewDoc()
	d.Line(transport.Text("🌿 "), transport.Bold("Telepath branches"))
	for _, name := range branches {
		d.Line(transport.Text("• "), transport.Code(name))
	}
	return *d
}

func (e *Engine) statusDoc(st statusSnapshot) transport.Doc {
	d := transport.NewDoc()
	d.Line(transport.Text("📊 "), transport.Bold("Status"))
	d.Line(transport.Text("Agent: "), transport.Bold(st.Agent))
	d.Line(transport.Text("Permission: "), transport.Bold(string(st.PermissionMode)))
	d.Line(transport.Text("Branch mode: "), transport.Bold(string(st.BranchMode)))
	if st.Running {
		if st.TaskBranch != "" {
			d.Line(transport.Text("Running: "), transport.Bold("yes"), transport.Text(" on 🌿 "), transport.Code(st.TaskBranch))
		} else {
			d.Line(transport.Text("Running: "), transport.Bold("yes"))
		}
		d.Line(transport.Text("Send /cancel to stop it."))
	} else {
		d.Line(transport.Text("Running: "), transport.Bold("no"))
	}
	return *d
}

type statusSnapshot struct {
	Agent          string
	PermissionMode agent.PermissionMode
	BranchMode     agent.BranchMode
	Running        bool
	TaskBranch     string
}

// errorDoc wraps a title line plus a preformatted error body.
func errorDoc(title, err string) transport.Doc {
	d := transport.NewDoc()
	d.Line(transport.Text(title))
	d.Pre(truncate(err, 1500))
	return *d
}

// --- small helpers mirrored from the original renderer ----------------------

func truncate(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "…"
}

func humanDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	return d.Round(time.Second).String()
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

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
