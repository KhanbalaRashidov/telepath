// Package engine is the transport-agnostic core of Telepath. It owns command
// dispatch and the task pipeline (git branch → agent run → normalized events →
// status/diffs → commit → report), driving everything through the transport
// abstraction. It is identical in behavior to the original Telegram bridge; the
// only structural change is that it is keyed by a transport-qualified
// conversation key instead of a bare Telegram chat id.
package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"log/slog"

	"github.com/telepath/telepath/internal/agent"
	"github.com/telepath/telepath/internal/config"
	"github.com/telepath/telepath/internal/git"
	"github.com/telepath/telepath/internal/permission"
	"github.com/telepath/telepath/internal/session"
	"github.com/telepath/telepath/internal/transport"
)

// Engine wires config, agents, state, and the permission broker to one or more
// transports.
type Engine struct {
	cfg     *config.Config
	reg     *agent.Registry
	store   session.Store
	broker  *permission.Broker
	log     *slog.Logger
	version string

	transports map[string]transport.Transport

	// taskSlot serializes tasks: all conversations share one ProjectsDir
	// working tree, so only one task may touch the repo at a time.
	taskSlot chan struct{}

	rootCtx context.Context
}

// New creates an engine.
func New(cfg *config.Config, reg *agent.Registry, store session.Store, broker *permission.Broker, log *slog.Logger, version string) *Engine {
	return &Engine{
		cfg:        cfg,
		reg:        reg,
		store:      store,
		broker:     broker,
		log:        log,
		version:    version,
		transports: make(map[string]transport.Transport),
		taskSlot:   make(chan struct{}, 1),
	}
}

// AddTransport registers a transport by name.
func (e *Engine) AddTransport(t transport.Transport) { e.transports[t.Name()] = t }

// Run starts every registered transport and blocks until ctx is cancelled.
func (e *Engine) Run(ctx context.Context) error {
	e.rootCtx = ctx
	names := make([]string, 0, len(e.transports))
	for n := range e.transports {
		names = append(names, n)
	}
	e.log.Info("telepath started", "version", e.version, "transports", names, "agents", e.reg.Names(), "default", e.reg.DefaultName())

	var wg sync.WaitGroup
	for _, t := range e.transports {
		wg.Add(1)
		go func(tr transport.Transport) {
			defer wg.Done()
			if err := tr.Start(ctx); err != nil && ctx.Err() == nil {
				e.log.Error("transport stopped with error", "transport", tr.Name(), "err", err)
			}
		}(t)
	}
	wg.Wait()
	return nil
}

func (e *Engine) transportFor(conv transport.ConversationRef) transport.Transport {
	return e.transports[conv.Transport]
}

func (e *Engine) reply(ctx context.Context, conv transport.ConversationRef, doc transport.Doc) {
	if t := e.transportFor(conv); t != nil {
		t.Reply(ctx, conv, doc)
	}
}

func keyOf(conv transport.ConversationRef) string {
	return session.Key(conv.Transport, conv.ConversationID, "")
}

// commandWords are the recognized non-agent commands.
var commandWords = map[string]bool{
	"help": true, "start": true, "agents": true, "use": true, "perm": true,
	"branch": true, "branches": true, "status": true, "cancel": true, "resume": true,
}

// Known reports whether word is a recognized command or an enabled agent name.
func (e *Engine) Known(word string) bool {
	word = strings.ToLower(strings.TrimSpace(word))
	if commandWords[word] {
		return true
	}
	_, ok := e.reg.Get(word)
	return ok
}

// HandleMessage dispatches a normalized inbound message. Text beginning with
// "/" is a command (or /<agent> task); anything else is a default-agent task.
func (e *Engine) HandleMessage(ctx context.Context, in transport.InboundMessage) {
	conv := in.Conversation
	key := keyOf(conv)
	text := strings.TrimSpace(in.Text)
	if text == "" {
		return
	}

	if !strings.HasPrefix(text, "/") {
		e.startTask(conv, key, e.store.Get(key).Agent, text)
		return
	}

	cmd, rest := splitCommand(text)
	switch cmd {
	case "help", "start":
		e.reply(ctx, conv, e.helpDoc())
	case "agents":
		e.reply(ctx, conv, e.agentsDoc(ctx))
	case "use":
		e.cmdUse(ctx, conv, key, rest)
	case "perm":
		e.cmdPerm(ctx, conv, key, rest)
	case "branch":
		e.cmdBranch(ctx, conv, key, rest)
	case "branches":
		e.cmdBranches(ctx, conv)
	case "status":
		e.cmdStatus(ctx, conv, key)
	case "cancel":
		e.cmdCancel(ctx, conv, key)
	case "resume":
		e.cmdResume(conv, key, rest)
	default:
		if _, ok := e.reg.Get(cmd); ok {
			if strings.TrimSpace(rest) == "" {
				e.reply(ctx, conv, plain("Usage: /"+cmd+" <task description>"))
				return
			}
			e.startTask(conv, key, cmd, rest)
			return
		}
		e.reply(ctx, conv, plain("Unknown command: "+cmd+"\nSend /help for usage."))
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

func (e *Engine) cmdUse(ctx context.Context, conv transport.ConversationRef, key, name string) {
	name = strings.ToLower(strings.TrimSpace(name))
	if _, ok := e.reg.Get(name); !ok {
		e.reply(ctx, conv, plain("Unknown agent "+name+". Enabled: "+strings.Join(e.reg.Names(), ", ")))
		return
	}
	e.store.Update(key, func(s *session.State) { s.Agent = name })
	d := transport.NewDoc()
	d.Line(transport.Text("✅ Default agent for this conversation is now "), transport.Bold(name), transport.Text("."))
	e.reply(ctx, conv, *d)
}

func (e *Engine) cmdPerm(ctx context.Context, conv transport.ConversationRef, key, mode string) {
	m := agent.PermissionMode(strings.ToLower(strings.TrimSpace(mode)))
	if !m.Valid() {
		e.reply(ctx, conv, plain("Usage: /perm <auto|allowlist|interactive>"))
		return
	}
	e.store.Update(key, func(s *session.State) { s.PermissionMode = m })
	d := transport.NewDoc()
	d.Line(transport.Text("✅ Permission mode for this conversation is now "), transport.Bold(string(m)), transport.Text("."))
	if m == agent.PermissionInteractive {
		ag, _ := e.reg.Get(e.store.Get(key).Agent)
		if ag != nil && !ag.Capabilities().InteractiveApproval {
			d.Line(transport.Text("⚠️ The current agent does not support interactive approval; it will fall back to "), transport.Bold("allowlist"), transport.Text("."))
		}
	}
	e.reply(ctx, conv, *d)
}

func (e *Engine) cmdBranch(ctx context.Context, conv transport.ConversationRef, key, mode string) {
	m := agent.BranchMode(strings.ToLower(strings.TrimSpace(mode)))
	if !m.Valid() {
		e.reply(ctx, conv, plain("Usage: /branch <off|per-task>"))
		return
	}
	e.store.Update(key, func(s *session.State) { s.BranchMode = m })
	d := transport.NewDoc()
	d.Line(transport.Text("✅ Branch mode for this conversation is now "), transport.Bold(string(m)), transport.Text("."))
	e.reply(ctx, conv, *d)
}

func (e *Engine) cmdBranches(ctx context.Context, conv transport.ConversationRef) {
	repo := git.Open(e.cfg.ProjectsDir, e.cfg.GitBin)
	if !repo.IsRepo(ctx) {
		e.reply(ctx, conv, plain("PROJECTS_DIR is not a git repository."))
		return
	}
	branches, err := repo.ListBranches(ctx, e.cfg.BranchPrefix)
	if err != nil {
		e.reply(ctx, conv, plain("⚠️ "+err.Error()))
		return
	}
	if len(branches) == 0 {
		d := transport.NewDoc()
		d.Line(transport.Text("No "), transport.Code(e.cfg.BranchPrefix+"*"), transport.Text(" branches."))
		e.reply(ctx, conv, *d)
		return
	}
	e.reply(ctx, conv, e.branchesDoc(e.cfg.BranchPrefix, branches))
}

func (e *Engine) cmdStatus(ctx context.Context, conv transport.ConversationRef, key string) {
	st := e.store.Get(key)
	e.reply(ctx, conv, e.statusDoc(statusSnapshot{
		Agent:          st.Agent,
		PermissionMode: st.PermissionMode,
		BranchMode:     st.BranchMode,
		Running:        st.Running,
		TaskBranch:     st.TaskBranch,
	}))
}

func (e *Engine) cmdCancel(ctx context.Context, conv transport.ConversationRef, key string) {
	st := e.store.Get(key)
	if !st.Running || st.Cancel == nil {
		e.reply(ctx, conv, plain("No task is running in this conversation."))
		return
	}
	st.Cancel()
	e.reply(ctx, conv, plain("🛑 Cancelling the running task…"))
}

// --- task execution --------------------------------------------------------

func (e *Engine) startTask(conv transport.ConversationRef, key, agentName, prompt string) {
	ctx := e.rootCtx
	t := e.transportFor(conv)
	if t == nil {
		return
	}
	st := e.store.Get(key)
	if st.Running {
		e.reply(ctx, conv, plain("⏳ A task is already running in this conversation. Send /cancel first."))
		return
	}
	ag, ok := e.reg.Get(agentName)
	if !ok {
		e.reply(ctx, conv, plain("Unknown agent "+agentName+"."))
		return
	}
	if !ag.Available(ctx) {
		d := transport.NewDoc()
		d.Line(transport.Text("Agent "), transport.Bold(agentName), transport.Text(" is not installed on this machine."))
		e.reply(ctx, conv, *d)
		return
	}

	// Acquire the global slot (all conversations share one working tree).
	select {
	case e.taskSlot <- struct{}{}:
	default:
		e.reply(ctx, conv, plain("⏳ Another task is using the workspace right now. Try again in a moment."))
		return
	}

	taskCtx, cancel := context.WithCancel(ctx)
	e.store.Update(key, func(s *session.State) {
		s.Running = true
		s.Cancel = cancel
		s.TaskBranch = ""
		s.BaseBranch = ""
	})

	go func() {
		defer func() {
			cancel()
			<-e.taskSlot
			e.store.Update(key, func(s *session.State) {
				s.Running = false
				s.Cancel = nil
			})
		}()
		e.runTask(taskCtx, conv, key, t.NewSession(conv), ag, prompt)
	}()
}

func (e *Engine) runTask(ctx context.Context, conv transport.ConversationRef, key string, sess transport.Session, ag agent.Agent, prompt string) {
	st := e.store.Get(key)
	agentName := ag.Name()

	mode := st.PermissionMode
	if mode == agent.PermissionInteractive && !ag.Capabilities().InteractiveApproval {
		mode = agent.PermissionAllowlist
		d := transport.NewDoc()
		d.Line(transport.Bold(agentName), transport.Text(" has no interactive approval; using "), transport.Bold("allowlist"), transport.Text(" for this task."))
		// prefix emoji on its own leading span to match original "⚠️ <b>agent</b> ..."
		d.Blocks[0].Spans = append([]transport.Span{transport.Text("⚠️ ")}, d.Blocks[0].Spans...)
		sess.Notice(ctx, *d)
	}

	repo := git.Open(e.cfg.ProjectsDir, e.cfg.GitBin)
	ws := git.NewWorkspace(repo, e.cfg.BranchPrefix, e.cfg.BranchStashDirty)
	branchMode := st.BranchMode

	var prep *git.Prep
	if branchMode == agent.BranchPerTask {
		p, err := ws.Prepare(ctx, prompt, time.Now())
		if err != nil {
			sess.Notice(ctx, errorDoc("❌ Could not prepare branch:", err.Error()))
			return
		}
		prep = p
		e.store.Update(key, func(s *session.State) {
			s.TaskBranch = p.Branch
			s.BaseBranch = p.Base
		})
		d := transport.NewDoc()
		d.Line(transport.Text("🌿 Working on "), transport.Code(p.Branch))
		d.Line(transport.Text("(base: "), transport.Code(p.Base), transport.Text(")"))
		sess.Notice(ctx, *d)
	}

	var hook agent.ApprovalFunc
	if mode == agent.PermissionInteractive {
		hook = sess.Approval
	}
	task := agent.Task{
		Prompt:         prompt,
		ProjectDir:     e.cfg.ProjectsDir,
		PermissionMode: mode,
		ApprovalHook:   hook,
		MaxTurns:       e.cfg.MaxTurns,
		Model:          e.cfg.Model,
	}

	events, err := ag.Run(ctx, task)
	if err != nil {
		sess.Notice(ctx, errorDoc("❌ Could not start the agent:", err.Error()))
		return
	}

	view := transport.StatusView{Agent: agentName, Starting: true}
	if prep != nil {
		view.Branch = prep.Branch
	}
	sess.Status(ctx, view)
	view.Starting = false

	result, runErr := e.consume(ctx, sess, &view, events)

	if result != nil && result.SessionID != "" {
		e.store.Update(key, func(s *session.State) { s.LastSessionID = result.SessionID })
	}

	if branchMode == agent.BranchPerTask && prep != nil && runErr == nil {
		if e.cfg.BranchCommit {
			sha, committed, cerr := ws.CommitAll(ctx, commitMessage(prompt, agentName))
			switch {
			case cerr != nil:
				sess.Notice(ctx, errorDoc("⚠️ Commit failed:", cerr.Error()))
			case committed:
				view.Committed = shortSHA(sha)
			default:
				sess.Notice(ctx, plain("ℹ️ The agent made no changes to commit."))
			}
		} else {
			d := transport.NewDoc()
			d.Line(transport.Text("ℹ️ Changes left "), transport.Bold("uncommitted"), transport.Text(" on "), transport.Code(prep.Branch), transport.Text(" (BRANCH_COMMIT=false; the working tree is shared, so isolation is weaker)."))
			sess.Notice(ctx, *d)
		}
	}

	e.finalReport(ctx, sess, &view, result, runErr, ws, prep, branchMode)
}

// consume reads the event stream, maintaining the live status and emitting
// per-file-edit messages. It returns the terminal Result and/or error.
func (e *Engine) consume(ctx context.Context, sess transport.Session, view *transport.StatusView, events <-chan agent.Event) (*agent.Result, error) {
	var result *agent.Result
	var runErr error

	for ev := range events {
		switch ev := ev.(type) {
		case agent.AssistantText:
			view.Action = ev.Text
			view.File, view.Cmd = "", ""
			sess.Status(ctx, *view)
		case agent.ToolUse:
			view.Action = strings.TrimSpace(ev.Name + " " + ev.InputSummary)
			sess.Status(ctx, *view)
		case agent.Command:
			view.Cmd = ev.Cmd
			view.Action = ""
			sess.Status(ctx, *view)
		case agent.FileEdit:
			view.File = ev.Path
			view.Action = ""
			sess.Status(ctx, *view)
			sess.FileEdit(ctx, ev.Path, ev.UnifiedDiff)
		case agent.ToolResult:
			if !ev.OK {
				view.Action = "⚠️ " + ev.Summary
				sess.Status(ctx, *view)
			}
		case agent.Progress:
			// Partial tokens are intentionally not streamed to avoid flooding.
		case agent.Result:
			r := ev
			result = &r
		case agent.Error:
			runErr = ev.Err
		}
	}
	if runErr == nil && ctx.Err() != nil {
		runErr = ctx.Err()
	}
	return result, runErr
}

func (e *Engine) finalReport(ctx context.Context, sess transport.Session, view *transport.StatusView, result *agent.Result, runErr error, ws *git.Workspace, prep *git.Prep, branchMode agent.BranchMode) {
	if runErr != nil {
		if ctx.Err() != nil {
			view.State = "cancelled"
		} else {
			view.State = "error"
		}
		title := "❌ Task failed"
		if view.State == "cancelled" {
			title = "🛑 Task cancelled"
		}
		body := transport.NewDoc()
		body.Line(boldPrefixTitle(title)...)
		body.Pre(truncate(runErr.Error(), 1500))
		rv := transport.ResultView{Status: *view, Body: *body}
		if prep != nil {
			prompt := transport.NewDoc()
			prompt.Line(transport.Text("The branch "), transport.Code(prep.Branch), transport.Text(" holds whatever changed. Keep it, or delete it and restore the base?"))
			rv.Prompt = prompt
			rv.Actions = &transport.BranchActions{Branch: prep.Branch, Base: prep.Base}
		}
		sess.Result(ctx, rv)
		return
	}

	view.State = "done"
	body := transport.NewDoc()
	done := []transport.Span{transport.Text("✅ "), transport.Bold("Done")}
	suffix := ""
	if result != nil {
		if result.Duration > 0 {
			suffix += " · " + humanDuration(result.Duration)
		}
		if result.CostUSD != nil {
			suffix += " · $" + formatCost(*result.CostUSD)
		}
	}
	if suffix != "" {
		done = append(done, transport.Text(suffix))
	}
	body.Line(done...)
	if result != nil {
		if s := strings.TrimSpace(result.Summary); s != "" && s != "completed" {
			body.Blank()
			body.Line(transport.Text(truncate(s, 1500)))
		}
	}

	rv := transport.ResultView{Status: *view}

	if branchMode == agent.BranchPerTask && prep != nil {
		stat, full, derr := ws.RangeSummary(ctx, prep.BaseSHA)
		if derr == nil {
			body.Blank()
			if view.Committed != "" {
				body.Line(transport.Text("🌿 "), transport.Code(prep.Branch), transport.Text(" · commit "), transport.Code(view.Committed))
			} else {
				body.Line(transport.Text("🌿 "), transport.Code(prep.Branch))
			}
			if strings.TrimSpace(stat) != "" {
				body.Pre(stat)
			}
			rv.Body = *body
			rv.Actions = &transport.BranchActions{Branch: prep.Branch, Base: prep.Base}
			if strings.TrimSpace(full) != "" {
				cap := transport.NewDoc()
				cap.Line(transport.Text("📄 Full diff of "), transport.Code(prep.Branch))
				rv.Diff = &transport.DiffDoc{Name: "task.diff", Caption: *cap, Content: full}
			}
			sess.Result(ctx, rv)
			return
		}
	}

	rv.Body = *body
	if full, err := ws.Repo().DiffWorktree(ctx); err == nil && strings.TrimSpace(full) != "" {
		rv.Diff = &transport.DiffDoc{Name: "changes.diff", Caption: plain("📄 Working-tree diff"), Content: full}
	}
	sess.Result(ctx, rv)
}

// boldPrefixTitle splits a "emoji Title" string into [Text(emoji+space), Bold(rest)].
func boldPrefixTitle(title string) []transport.Span {
	if i := strings.IndexByte(title, ' '); i >= 0 {
		return []transport.Span{transport.Text(title[:i+1]), transport.Bold(title[i+1:])}
	}
	return []transport.Span{transport.Bold(title)}
}

func formatCost(c float64) string { return fmt.Sprintf("%.4f", c) }
