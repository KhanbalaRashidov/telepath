package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-telegram/bot/models"

	"github.com/telepath/telepath/internal/agent"
	"github.com/telepath/telepath/internal/git"
	"github.com/telepath/telepath/internal/session"
)

// handleMessage routes a text message to a command handler or a task run.
func (bt *Bot) handleMessage(ctx context.Context, msg *models.Message) {
	chatID := msg.Chat.ID
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return
	}

	if !strings.HasPrefix(text, "/") {
		bt.startTask(chatID, bt.store.Get(chatID).Agent, text)
		return
	}

	cmd, rest := splitCommand(text)
	switch cmd {
	case "help", "start":
		bt.sendText(ctx, chatID, helpText(bt.reg.Names()))
	case "agents":
		bt.sendText(ctx, chatID, bt.agentsText(ctx))
	case "use":
		bt.cmdUse(ctx, chatID, rest)
	case "perm":
		bt.cmdPerm(ctx, chatID, rest)
	case "branch":
		bt.cmdBranch(ctx, chatID, rest)
	case "branches":
		bt.cmdBranches(ctx, chatID)
	case "status":
		bt.cmdStatus(ctx, chatID)
	case "cancel":
		bt.cmdCancel(ctx, chatID)
	case "resume":
		bt.cmdResume(chatID, rest)
	default:
		if _, ok := bt.reg.Get(cmd); ok {
			if strings.TrimSpace(rest) == "" {
				bt.sendText(ctx, chatID, "Usage: /"+escapeHTML(cmd)+" &lt;task description&gt;")
				return
			}
			bt.startTask(chatID, cmd, rest)
			return
		}
		bt.sendText(ctx, chatID, "Unknown command: "+escapeHTML(cmd)+"\nSend /help for usage.")
	}
}

// splitCommand splits "/cmd@bot rest" into ("cmd", "rest").
func splitCommand(text string) (cmd, rest string) {
	text = strings.TrimPrefix(text, "/")
	if i := strings.IndexAny(text, " \t\n"); i >= 0 {
		cmd, rest = text[:i], strings.TrimSpace(text[i+1:])
	} else {
		cmd = text
	}
	if at := strings.IndexByte(cmd, '@'); at >= 0 {
		cmd = cmd[:at]
	}
	return strings.ToLower(cmd), rest
}

func (bt *Bot) cmdUse(ctx context.Context, chatID int64, name string) {
	name = strings.ToLower(strings.TrimSpace(name))
	if _, ok := bt.reg.Get(name); !ok {
		bt.sendText(ctx, chatID, "Unknown agent "+escapeHTML(name)+". Enabled: "+escapeHTML(strings.Join(bt.reg.Names(), ", ")))
		return
	}
	bt.store.Update(chatID, func(s *session.State) { s.Agent = name })
	bt.sendText(ctx, chatID, "✅ Default agent for this chat is now <b>"+escapeHTML(name)+"</b>.")
}

func (bt *Bot) cmdPerm(ctx context.Context, chatID int64, mode string) {
	m := agent.PermissionMode(strings.ToLower(strings.TrimSpace(mode)))
	if !m.Valid() {
		bt.sendText(ctx, chatID, "Usage: /perm &lt;auto|allowlist|interactive&gt;")
		return
	}
	bt.store.Update(chatID, func(s *session.State) { s.PermissionMode = m })
	note := ""
	if m == agent.PermissionInteractive {
		ag, _ := bt.reg.Get(bt.store.Get(chatID).Agent)
		if ag != nil && !ag.Capabilities().InteractiveApproval {
			note = "\n⚠️ The current agent does not support interactive approval; it will fall back to <b>allowlist</b>."
		}
	}
	bt.sendText(ctx, chatID, "✅ Permission mode for this chat is now <b>"+escapeHTML(string(m))+"</b>."+note)
}

func (bt *Bot) cmdBranch(ctx context.Context, chatID int64, mode string) {
	m := agent.BranchMode(strings.ToLower(strings.TrimSpace(mode)))
	if !m.Valid() {
		bt.sendText(ctx, chatID, "Usage: /branch &lt;off|per-task&gt;")
		return
	}
	bt.store.Update(chatID, func(s *session.State) { s.BranchMode = m })
	bt.sendText(ctx, chatID, "✅ Branch mode for this chat is now <b>"+escapeHTML(string(m))+"</b>.")
}

func (bt *Bot) cmdBranches(ctx context.Context, chatID int64) {
	repo := git.Open(bt.cfg.ProjectsDir, bt.cfg.GitBin)
	if !repo.IsRepo(ctx) {
		bt.sendText(ctx, chatID, "PROJECTS_DIR is not a git repository.")
		return
	}
	branches, err := repo.ListBranches(ctx, bt.cfg.BranchPrefix)
	if err != nil {
		bt.sendText(ctx, chatID, "⚠️ "+escapeHTML(err.Error()))
		return
	}
	if len(branches) == 0 {
		bt.sendText(ctx, chatID, "No <code>"+escapeHTML(bt.cfg.BranchPrefix)+"*</code> branches.")
		return
	}
	var b strings.Builder
	b.WriteString("🌿 <b>Telepath branches</b>\n")
	for _, name := range branches {
		b.WriteString("• <code>" + escapeHTML(name) + "</code>\n")
	}
	bt.sendText(ctx, chatID, b.String())
}

func (bt *Bot) cmdStatus(ctx context.Context, chatID int64) {
	st := bt.store.Get(chatID)
	var b strings.Builder
	b.WriteString("📊 <b>Status</b>\n")
	b.WriteString("Agent: <b>" + escapeHTML(st.Agent) + "</b>\n")
	b.WriteString("Permission: <b>" + escapeHTML(string(st.PermissionMode)) + "</b>\n")
	b.WriteString("Branch mode: <b>" + escapeHTML(string(st.BranchMode)) + "</b>\n")
	if st.Running {
		b.WriteString("Running: <b>yes</b>")
		if st.TaskBranch != "" {
			b.WriteString(" on 🌿 <code>" + escapeHTML(st.TaskBranch) + "</code>")
		}
		b.WriteString("\nSend /cancel to stop it.")
	} else {
		b.WriteString("Running: <b>no</b>")
	}
	bt.sendText(ctx, chatID, b.String())
}

func (bt *Bot) cmdCancel(ctx context.Context, chatID int64) {
	st := bt.store.Get(chatID)
	if !st.Running || st.Cancel == nil {
		bt.sendText(ctx, chatID, "No task is running in this chat.")
		return
	}
	st.Cancel()
	bt.sendText(ctx, chatID, "🛑 Cancelling the running task…")
}

// --- task execution --------------------------------------------------------

// startTask validates preconditions, acquires the global repo slot, and runs
// the task in a goroutine so the update handler returns promptly.
func (bt *Bot) startTask(chatID int64, agentName, prompt string) {
	ctx := bt.rootCtx
	st := bt.store.Get(chatID)
	if st.Running {
		bt.sendText(ctx, chatID, "⏳ A task is already running in this chat. Send /cancel first.")
		return
	}
	ag, ok := bt.reg.Get(agentName)
	if !ok {
		bt.sendText(ctx, chatID, "Unknown agent "+escapeHTML(agentName)+".")
		return
	}
	if !ag.Available(ctx) {
		bt.sendText(ctx, chatID, "Agent <b>"+escapeHTML(agentName)+"</b> is not installed on this machine.")
		return
	}

	// Acquire the global slot (all chats share one working tree).
	select {
	case bt.taskSlot <- struct{}{}:
	default:
		bt.sendText(ctx, chatID, "⏳ Another task is using the workspace right now. Try again in a moment.")
		return
	}

	taskCtx, cancel := context.WithCancel(ctx)
	bt.store.Update(chatID, func(s *session.State) {
		s.Running = true
		s.Cancel = cancel
		s.TaskBranch = ""
		s.BaseBranch = ""
	})

	go func() {
		defer func() {
			cancel()
			<-bt.taskSlot
			bt.store.Update(chatID, func(s *session.State) {
				s.Running = false
				s.Cancel = nil
			})
		}()
		bt.runTask(taskCtx, chatID, ag, prompt)
	}()
}

// runTask drives the full pipeline for one task.
func (bt *Bot) runTask(ctx context.Context, chatID int64, ag agent.Agent, prompt string) {
	st := bt.store.Get(chatID)
	agentName := ag.Name()

	// Resolve the effective permission mode (fall back if unsupported).
	mode := st.PermissionMode
	if mode == agent.PermissionInteractive && !ag.Capabilities().InteractiveApproval {
		mode = agent.PermissionAllowlist
		bt.sendText(ctx, chatID, "⚠️ <b>"+escapeHTML(agentName)+"</b> has no interactive approval; using <b>allowlist</b> for this task.")
	}

	// Prepare the git workspace / branch.
	repo := git.Open(bt.cfg.ProjectsDir, bt.cfg.GitBin)
	ws := git.NewWorkspace(repo, bt.cfg.BranchPrefix, bt.cfg.BranchStashDirty)
	branchMode := st.BranchMode

	var prep *git.Prep
	if branchMode == agent.BranchPerTask {
		p, err := ws.Prepare(ctx, prompt, time.Now())
		if err != nil {
			bt.sendText(ctx, chatID, "❌ Could not prepare branch:\n"+preBlock(err.Error()))
			return
		}
		prep = p
		bt.store.Update(chatID, func(s *session.State) {
			s.TaskBranch = p.Branch
			s.BaseBranch = p.Base
		})
		bt.sendText(ctx, chatID, fmt.Sprintf("🌿 Working on <code>%s</code>\n(base: <code>%s</code>)", escapeHTML(p.Branch), escapeHTML(p.Base)))
	}

	// Build the task.
	var hook agent.ApprovalFunc
	if mode == agent.PermissionInteractive {
		hook = bt.makeApprovalHook(chatID)
	}
	task := agent.Task{
		Prompt:         prompt,
		ProjectDir:     bt.cfg.ProjectsDir,
		PermissionMode: mode,
		ApprovalHook:   hook,
		MaxTurns:       bt.cfg.MaxTurns,
		Model:          bt.cfg.Model,
	}

	events, err := ag.Run(ctx, task)
	if err != nil {
		bt.sendText(ctx, chatID, "❌ Could not start the agent:\n"+preBlock(err.Error()))
		return
	}

	view := taskView{agent: agentName}
	if prep != nil {
		view.branch = prep.Branch
	}
	statusID, _ := bt.send(ctx, chatID, "⏳ "+view.render(), nil)
	th := newThrottle(1500*time.Millisecond, func(text string) {
		if statusID != 0 {
			bt.edit(ctx, chatID, statusID, text, nil)
		}
	})

	result, runErr := bt.consume(ctx, chatID, &view, th, events)
	th.Stop()

	if result != nil && result.SessionID != "" {
		bt.store.Update(chatID, func(s *session.State) { s.LastSessionID = result.SessionID })
	}

	// Commit on the task branch (success only).
	committedSHA := ""
	if branchMode == agent.BranchPerTask && prep != nil && runErr == nil {
		if bt.cfg.BranchCommit {
			sha, committed, cerr := ws.CommitAll(ctx, commitMessage(prompt, agentName))
			switch {
			case cerr != nil:
				bt.sendText(ctx, chatID, "⚠️ Commit failed:\n"+preBlock(cerr.Error()))
			case committed:
				committedSHA = shortSHA(sha)
				view.committed = committedSHA
			default:
				bt.sendText(ctx, chatID, "ℹ️ The agent made no changes to commit.")
			}
		} else {
			bt.sendText(ctx, chatID, "ℹ️ Changes left <b>uncommitted</b> on <code>"+escapeHTML(prep.Branch)+"</code> (BRANCH_COMMIT=false; the working tree is shared, so isolation is weaker).")
		}
	}

	bt.finalReport(ctx, chatID, statusID, &view, result, runErr, ws, prep, branchMode)
}

// consume reads the event stream, maintaining the live status and emitting
// per-file-edit messages. It returns the terminal Result and/or error.
func (bt *Bot) consume(ctx context.Context, chatID int64, view *taskView, th *throttle, events <-chan agent.Event) (*agent.Result, error) {
	var result *agent.Result
	var runErr error

	for ev := range events {
		switch e := ev.(type) {
		case agent.AssistantText:
			view.action = e.Text
			view.file, view.cmd = "", ""
			th.Update(view.render())
		case agent.ToolUse:
			view.action = strings.TrimSpace(e.Name + " " + e.InputSummary)
			th.Update(view.render())
		case agent.Command:
			view.cmd = e.Cmd
			view.action = ""
			th.Update(view.render())
			th.Flush()
		case agent.FileEdit:
			view.file = e.Path
			view.action = ""
			th.Update(view.render())
			th.Flush()
			bt.emitFileEdit(ctx, chatID, e)
		case agent.ToolResult:
			if !e.OK {
				view.action = "⚠️ " + e.Summary
				th.Update(view.render())
			}
		case agent.Progress:
			// Partial tokens are intentionally not streamed to avoid flooding;
			// the coalesced AssistantText already conveys "what I'm doing".
		case agent.Result:
			r := e
			result = &r
		case agent.Error:
			runErr = e.Err
		}
	}

	if runErr == nil && ctx.Err() != nil {
		runErr = ctx.Err()
	}
	return result, runErr
}

func (bt *Bot) emitFileEdit(ctx context.Context, chatID int64, e agent.FileEdit) {
	body, asDoc := fileEditMessage(e.Path, e.UnifiedDiff)
	if asDoc {
		bt.sendDoc(ctx, chatID, safeDiffName(e.Path), []byte(e.UnifiedDiff), "✏️ <b>"+escapeHTML(e.Path)+"</b>")
		return
	}
	bt.sendText(ctx, chatID, body)
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

func safeDiffName(path string) string {
	name := baseName(path)
	name = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			return r
		default:
			return '_'
		}
	}, name)
	if name == "" {
		name = "changes"
	}
	return name + ".diff"
}
