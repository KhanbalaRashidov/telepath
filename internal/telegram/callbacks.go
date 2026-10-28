package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/telepath/telepath/internal/agent"
	"github.com/telepath/telepath/internal/git"
	"github.com/telepath/telepath/internal/permission"
	"github.com/telepath/telepath/internal/session"
)

// handleCallback resolves inline-button taps: approval decisions and the
// post-task Keep/Delete-branch actions.
func (bt *Bot) handleCallback(ctx context.Context, cq *models.CallbackQuery) {
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
			bt.answer(ctx, cq.ID, "Malformed request")
			return
		}
		id, allow := parts[0], parts[1] == "a"
		note := "approved by user"
		if !allow {
			note = "denied by user"
		}
		if bt.broker.Resolve(id, allow, note) {
			if allow {
				bt.answer(ctx, cq.ID, "✅ Approved")
			} else {
				bt.answer(ctx, cq.ID, "⛔ Denied")
			}
		} else {
			bt.answer(ctx, cq.ID, "⌛ This request already expired.")
		}

	case data == cbBranchKeep:
		st := bt.store.Get(chatID)
		bt.answer(ctx, cq.ID, "🌿 Branch kept")
		if msgID != 0 {
			bt.edit(ctx, chatID, msgID, "🌿 Kept <code>"+escapeHTML(st.TaskBranch)+"</code>. Review, then merge/push it yourself.", nil)
		}

	case data == cbBranchDelete:
		bt.handleBranchDelete(ctx, chatID, msgID, cq.ID)

	default:
		bt.answer(ctx, cq.ID, "")
	}
}

func (bt *Bot) handleBranchDelete(ctx context.Context, chatID int64, msgID int, cqID string) {
	st := bt.store.Get(chatID)
	if st.TaskBranch == "" || st.BaseBranch == "" {
		bt.answer(ctx, cqID, "Nothing to restore")
		return
	}
	if st.Running {
		bt.answer(ctx, cqID, "A task is running; /cancel first")
		return
	}
	repo := git.Open(bt.cfg.ProjectsDir, bt.cfg.GitBin)
	ws := git.NewWorkspace(repo, bt.cfg.BranchPrefix, bt.cfg.BranchStashDirty)
	if err := ws.RestoreToBase(ctx, st.BaseBranch, st.TaskBranch); err != nil {
		bt.answer(ctx, cqID, "⚠️ Failed")
		bt.sendText(ctx, chatID, "⚠️ Could not restore base:\n"+preBlock(err.Error()))
		return
	}
	bt.answer(ctx, cqID, "🗑 Deleted & restored")
	if msgID != 0 {
		bt.edit(ctx, chatID, msgID, fmt.Sprintf("🗑 Deleted <code>%s</code> and restored <code>%s</code>.", escapeHTML(st.TaskBranch), escapeHTML(st.BaseBranch)), nil)
	}
	bt.store.Update(chatID, func(s *session.State) { s.TaskBranch = ""; s.BaseBranch = "" })
}

func (bt *Bot) answer(ctx context.Context, cqID, text string) {
	_, _ = bt.api.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: cqID,
		Text:            text,
	})
}

// --- approval rendering ----------------------------------------------------

func approvalPrompt(req agent.ApprovalRequest) string {
	var b strings.Builder
	b.WriteString(riskEmoji(req.Risk) + " <b>Approval needed</b>\n")
	b.WriteString("Tool: <code>" + escapeHTML(req.Tool) + "</code>\n")
	b.WriteString("Risk: <b>" + escapeHTML(req.Risk) + "</b>\n")
	if detail := approvalDetail(req); detail != "" {
		b.WriteString(preBlock(detail))
	}
	return b.String()
}

func approvalResolved(req agent.ApprovalRequest, d permission.Decision) string {
	head := "⛔ <b>Denied</b>"
	if d.Allow {
		head = "✅ <b>Approved</b>"
	}
	line := head + " — <code>" + escapeHTML(req.Tool) + "</code>"
	if d.Note != "" {
		line += "\n<i>" + escapeHTML(d.Note) + "</i>"
	}
	return line
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

func riskEmoji(risk string) string {
	switch risk {
	case "high":
		return "🔴"
	case "low":
		return "🟢"
	default:
		return "🟡"
	}
}

// --- final report ----------------------------------------------------------

func (bt *Bot) finalReport(ctx context.Context, chatID int64, statusID int, view *taskView, result *agent.Result, runErr error, ws *git.Workspace, prep *git.Prep, branchMode agent.BranchMode) {
	if runErr != nil {
		if ctx.Err() != nil {
			view.state = "cancelled"
		} else {
			view.state = "error"
		}
		if statusID != 0 {
			bt.edit(ctx, chatID, statusID, view.render(), nil)
		}
		title := "❌ <b>Task failed</b>"
		if view.state == "cancelled" {
			title = "🛑 <b>Task cancelled</b>"
		}
		bt.sendText(ctx, chatID, title+"\n"+preBlock(truncate(runErr.Error(), 1500)))
		if prep != nil {
			_, _ = bt.send(ctx, chatID, "The branch <code>"+escapeHTML(prep.Branch)+"</code> holds whatever changed. Keep it, or delete it and restore the base?", branchKeyboard())
		}
		return
	}

	view.state = "done"
	if statusID != 0 {
		bt.edit(ctx, chatID, statusID, view.render(), nil)
	}

	var b strings.Builder
	b.WriteString("✅ <b>Done</b>")
	if result != nil {
		if result.Duration > 0 {
			b.WriteString(" · " + humanDuration(result.Duration))
		}
		if result.CostUSD != nil {
			b.WriteString(fmt.Sprintf(" · $%.4f", *result.CostUSD))
		}
		if s := strings.TrimSpace(result.Summary); s != "" && s != "completed" {
			b.WriteString("\n\n" + escapeHTML(truncate(s, 1500)))
		}
	}

	if branchMode == agent.BranchPerTask && prep != nil {
		stat, full, derr := ws.RangeSummary(ctx, prep.BaseSHA)
		if derr == nil {
			if view.committed != "" {
				b.WriteString("\n\n🌿 <code>" + escapeHTML(prep.Branch) + "</code> · commit <code>" + escapeHTML(view.committed) + "</code>")
			} else {
				b.WriteString("\n\n🌿 <code>" + escapeHTML(prep.Branch) + "</code>")
			}
			if strings.TrimSpace(stat) != "" {
				b.WriteString("\n" + preBlock(stat))
			}
			_, _ = bt.send(ctx, chatID, b.String(), branchKeyboard())
			if strings.TrimSpace(full) != "" {
				bt.sendDoc(ctx, chatID, "task.diff", []byte(full), "📄 Full diff of <code>"+escapeHTML(prep.Branch)+"</code>")
			}
			return
		}
	}

	// Branch-off mode (or diff error): show the working-tree diff.
	bt.sendText(ctx, chatID, b.String())
	if full, err := ws.Repo().DiffWorktree(ctx); err == nil && strings.TrimSpace(full) != "" {
		bt.sendDoc(ctx, chatID, "changes.diff", []byte(full), "📄 Working-tree diff")
	}
}

// --- resume ----------------------------------------------------------------

func (bt *Bot) cmdResume(chatID int64, msg string) {
	ctx := bt.rootCtx
	st := bt.store.Get(chatID)
	if st.LastSessionID == "" {
		bt.sendText(ctx, chatID, "No previous session to resume in this chat.")
		return
	}
	ag, ok := bt.reg.Get(st.Agent)
	if !ok {
		bt.sendText(ctx, chatID, "The chat's agent is no longer available.")
		return
	}
	if !ag.Capabilities().Resume {
		bt.sendText(ctx, chatID, "Agent <b>"+escapeHTML(ag.Name())+"</b> does not support /resume.")
		return
	}
	if strings.TrimSpace(msg) == "" {
		bt.sendText(ctx, chatID, "Usage: /resume &lt;what to do next&gt;")
		return
	}
	if st.Running {
		bt.sendText(ctx, chatID, "⏳ A task is already running. Send /cancel first.")
		return
	}

	select {
	case bt.taskSlot <- struct{}{}:
	default:
		bt.sendText(ctx, chatID, "⏳ The workspace is busy. Try again shortly.")
		return
	}

	taskCtx, cancel := context.WithCancel(ctx)
	bt.store.Update(chatID, func(s *session.State) { s.Running = true; s.Cancel = cancel })

	go func() {
		defer func() {
			cancel()
			<-bt.taskSlot
			bt.store.Update(chatID, func(s *session.State) { s.Running = false; s.Cancel = nil })
		}()
		bt.runResume(taskCtx, chatID, ag, st.LastSessionID, msg)
	}()
}

func (bt *Bot) runResume(ctx context.Context, chatID int64, ag agent.Agent, sessionID, msg string) {
	events, err := ag.Resume(ctx, sessionID, msg)
	if err != nil {
		bt.sendText(ctx, chatID, "❌ Could not resume:\n"+preBlock(err.Error()))
		return
	}
	bt.sendText(ctx, chatID, "🔁 Resuming session <code>"+escapeHTML(shortSHA(sessionID))+"</code> (auto permission mode).")

	view := taskView{agent: ag.Name()}
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

	repo := git.Open(bt.cfg.ProjectsDir, bt.cfg.GitBin)
	ws := git.NewWorkspace(repo, bt.cfg.BranchPrefix, bt.cfg.BranchStashDirty)
	bt.finalReport(ctx, chatID, statusID, &view, result, runErr, ws, nil, agent.BranchOff)
}

// --- static text -----------------------------------------------------------

func helpText(agents []string) string {
	return "🤖 <b>Telepath</b> — Telegram ↔ coding-agent bridge\n\n" +
		"<b>Commands</b>\n" +
		"/help — this help\n" +
		"/agents — list agents and availability\n" +
		"/use &lt;agent&gt; — set this chat's default agent\n" +
		"/perm &lt;auto|allowlist|interactive&gt; — permission mode\n" +
		"/branch &lt;off|per-task&gt; — branch workflow\n" +
		"/&lt;agent&gt; &lt;task&gt; — run with a specific agent\n" +
		"&lt;message&gt; — run with the default agent\n" +
		"/resume &lt;text&gt; — continue the last session\n" +
		"/branches — list telepath/* branches\n" +
		"/status — current task + mode summary\n" +
		"/cancel — cancel the running task\n\n" +
		"Enabled agents: <b>" + escapeHTML(strings.Join(agents, ", ")) + "</b>\n" +
		"Telepath never pushes, merges, or touches main/master — it leaves a reviewable branch."
}

func (bt *Bot) agentsText(ctx context.Context) string {
	var b strings.Builder
	b.WriteString("🧩 <b>Agents</b>\n")
	def := bt.reg.DefaultName()
	for _, ag := range bt.reg.List() {
		caps := ag.Capabilities()
		avail := "❌ not installed"
		if ag.Available(ctx) {
			avail = "✅ installed"
		}
		marker := ""
		if ag.Name() == def {
			marker = " (default)"
		}
		b.WriteString("\n<b>" + escapeHTML(ag.Name()) + "</b>" + marker + " — " + avail + "\n")
		b.WriteString("  stream " + yesno(caps.StructuredStream) +
			" · resume " + yesno(caps.Resume) +
			" · interactive " + yesno(caps.InteractiveApproval) +
			" · file-edits " + yesno(caps.FileEditEvents) + "\n")
	}
	return b.String()
}

func yesno(b bool) string {
	if b {
		return "✅"
	}
	return "—"
}
