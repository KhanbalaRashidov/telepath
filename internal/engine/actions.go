package engine

import (
	"context"

	"github.com/telepath/telepath/internal/agent"
	"github.com/telepath/telepath/internal/git"
	"github.com/telepath/telepath/internal/session"
	"github.com/telepath/telepath/internal/transport"
)

// HandleBranchAction performs a post-task branch button action for a
// conversation. Mirrors the original Telegram Keep / Delete-&-restore handlers.
func (e *Engine) HandleBranchAction(ctx context.Context, conv transport.ConversationRef, action string) transport.ActionResult {
	key := keyOf(conv)
	st := e.store.Get(key)

	switch action {
	case "keep":
		d := transport.NewDoc()
		d.Line(transport.Text("🌿 Kept "), transport.Code(st.TaskBranch), transport.Text(". Review, then merge/push it yourself."))
		return transport.ActionResult{Toast: "🌿 Branch kept", Replace: d}

	case "delete", "del":
		if st.TaskBranch == "" || st.BaseBranch == "" {
			return transport.ActionResult{Toast: "Nothing to restore"}
		}
		if st.Running {
			return transport.ActionResult{Toast: "A task is running; /cancel first"}
		}
		repo := git.Open(e.cfg.ProjectsDir, e.cfg.GitBin)
		ws := git.NewWorkspace(repo, e.cfg.BranchPrefix, e.cfg.BranchStashDirty)
		if err := ws.RestoreToBase(ctx, st.BaseBranch, st.TaskBranch); err != nil {
			return transport.ActionResult{Toast: "⚠️ Failed", Replace: ptrDoc(errorDoc("⚠️ Could not restore base:", err.Error()))}
		}
		d := transport.NewDoc()
		d.Line(transport.Text("🗑 Deleted "), transport.Code(st.TaskBranch), transport.Text(" and restored "), transport.Code(st.BaseBranch), transport.Text("."))
		e.store.Update(key, func(s *session.State) { s.TaskBranch = ""; s.BaseBranch = "" })
		return transport.ActionResult{Toast: "🗑 Deleted & restored", Replace: d}

	default:
		return transport.ActionResult{}
	}
}

func ptrDoc(d transport.Doc) *transport.Doc { return &d }

// --- resume ----------------------------------------------------------------

func (e *Engine) cmdResume(conv transport.ConversationRef, key, msg string) {
	ctx := e.rootCtx
	t := e.transportFor(conv)
	if t == nil {
		return
	}
	st := e.store.Get(key)
	if st.LastSessionID == "" {
		e.reply(ctx, conv, plain("No previous session to resume in this conversation."))
		return
	}
	ag, ok := e.reg.Get(st.Agent)
	if !ok {
		e.reply(ctx, conv, plain("The conversation's agent is no longer available."))
		return
	}
	if !ag.Capabilities().Resume {
		d := transport.NewDoc()
		d.Line(transport.Text("Agent "), transport.Bold(ag.Name()), transport.Text(" does not support /resume."))
		e.reply(ctx, conv, *d)
		return
	}
	if msg == "" {
		e.reply(ctx, conv, plain("Usage: /resume <what to do next>"))
		return
	}
	if st.Running {
		e.reply(ctx, conv, plain("⏳ A task is already running. Send /cancel first."))
		return
	}

	select {
	case e.taskSlot <- struct{}{}:
	default:
		e.reply(ctx, conv, plain("⏳ The workspace is busy. Try again shortly."))
		return
	}

	taskCtx, cancel := context.WithCancel(ctx)
	e.store.Update(key, func(s *session.State) { s.Running = true; s.Cancel = cancel })

	go func() {
		defer func() {
			cancel()
			<-e.taskSlot
			e.store.Update(key, func(s *session.State) { s.Running = false; s.Cancel = nil })
		}()
		e.runResume(taskCtx, key, t.NewSession(conv), ag, st.LastSessionID, msg)
	}()
}

func (e *Engine) runResume(ctx context.Context, key string, sess transport.Session, ag agent.Agent, sessionID, msg string) {
	events, err := ag.Resume(ctx, sessionID, msg)
	if err != nil {
		sess.Notice(ctx, errorDoc("❌ Could not resume:", err.Error()))
		return
	}
	d := transport.NewDoc()
	d.Line(transport.Text("🔁 Resuming session "), transport.Code(shortSHA(sessionID)), transport.Text(" (auto permission mode)."))
	sess.Notice(ctx, *d)

	view := transport.StatusView{Agent: ag.Name(), Starting: true}
	sess.Status(ctx, view)
	view.Starting = false

	result, runErr := e.consume(ctx, sess, &view, events)

	if result != nil && result.SessionID != "" {
		e.store.Update(key, func(s *session.State) { s.LastSessionID = result.SessionID })
	}

	repo := git.Open(e.cfg.ProjectsDir, e.cfg.GitBin)
	ws := git.NewWorkspace(repo, e.cfg.BranchPrefix, e.cfg.BranchStashDirty)
	e.finalReport(ctx, sess, &view, result, runErr, ws, nil, agent.BranchOff)
}
