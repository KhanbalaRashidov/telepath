// Package transport is the messaging-platform abstraction. The core
// (agent/permission/git + the normalized Event model) and the task engine
// depend only on these types, never on Telegram or Slack directly. Each
// platform (internal/telegram, internal/slack) implements Transport, owning its
// own inbound parsing, markup formatting, file upload, and button mechanics.
package transport

import (
	"context"

	"github.com/telepath/telepath/internal/agent"
)

// ConversationRef identifies where a message/task lives. ConversationID is the
// platform's channel/chat id; ThreadID is set by platforms that thread task
// output (Slack) and empty otherwise (Telegram).
type ConversationRef struct {
	Transport      string
	ConversationID string
	ThreadID       string
}

// InboundMessage is a platform message normalized for the engine.
type InboundMessage struct {
	Conversation ConversationRef
	UserID       string
	Text         string
}

// StatusView is the data behind the single live status line. Each transport
// formats it; RenderStatusText produces the shared one-line text.
type StatusView struct {
	Agent     string
	Branch    string
	Action    string
	File      string
	Cmd       string
	Committed string
	State     string // "" | "done" | "error" | "cancelled"
	Starting  bool   // true only for the first status post
}

// DiffDoc is a diff to attach as an uploaded file.
type DiffDoc struct {
	Name    string
	Caption Doc
	Content string
}

// BranchActions, when set on a ResultView, asks the transport to show the
// post-task "keep branch" / "delete branch & restore base" buttons.
type BranchActions struct {
	Branch string
	Base   string
}

// ResultView is the final task report.
//
// Rendering contract (preserves the original Telegram ordering):
//   - Status is the final status line the live status message is edited to.
//   - Body is the main report/error message.
//   - Diff, if set, is uploaded as a file afterwards.
//   - Actions, if set, shows the Keep / Delete-&-restore buttons. When Prompt
//     is nil the buttons attach to Body (success); when Prompt is set, Body is
//     posted first and the buttons attach to a separate Prompt message (error).
type ResultView struct {
	Status  StatusView
	Body    Doc
	Diff    *DiffDoc       // optional attached full diff
	Prompt  *Doc           // optional separate message that carries the buttons
	Actions *BranchActions // optional branch-cleanup buttons
}

// ActionResult is returned by the engine when a post-task branch button is
// tapped; the transport shows Toast and, if Replace is set, edits the message.
type ActionResult struct {
	Toast   string
	Replace *Doc
}

// Session is the per-task outbound driver the engine uses to render progress.
// Implementations are transport-specific; they own throttling and markup.
type Session interface {
	// Status creates (first call) or edits-in-place the single live status line.
	Status(ctx context.Context, v StatusView)
	// Notice posts a one-off message within the task.
	Notice(ctx context.Context, doc Doc)
	// FileEdit posts a file's diff (inline for small, uploaded file for large).
	FileEdit(ctx context.Context, path, unifiedDiff string)
	// Approval shows Approve/Deny and blocks until the user (or a timeout via
	// the broker) decides. This is wired as the agent Task's ApprovalHook.
	Approval(ctx context.Context, req agent.ApprovalRequest) agent.ApprovalDecision
	// Result posts the final report, attaching a diff and/or branch buttons.
	Result(ctx context.Context, r ResultView)
}

// Transport is a messaging platform.
type Transport interface {
	// Name is the stable transport id ("telegram", "slack").
	Name() string
	// Start begins receiving events and blocks until ctx is cancelled.
	Start(ctx context.Context) error
	// Reply sends a one-off message to a conversation (command replies).
	Reply(ctx context.Context, conv ConversationRef, doc Doc)
	// NewSession creates a per-task outbound session bound to a conversation.
	NewSession(conv ConversationRef) Session
}

// Router is implemented by the engine; transports call it with inbound events.
type Router interface {
	// HandleMessage dispatches a normalized inbound message (command or task).
	HandleMessage(ctx context.Context, in InboundMessage)
	// HandleBranchAction performs a post-task branch button action ("keep" or
	// "delete") for a conversation and returns what the transport should show.
	HandleBranchAction(ctx context.Context, conv ConversationRef, action string) ActionResult
	// Known reports whether a word is a recognized command or agent name, so a
	// transport can decide whether leading text is a command or a task prompt.
	Known(word string) bool
}

// RenderStatusText composes the shared one-line status string (emojis + " · "
// separators), identical across transports. Each transport escapes it for its
// own markup. This preserves the exact status wording regardless of platform.
func RenderStatusText(v StatusView) string {
	parts := make([]string, 0, 6)
	parts = append(parts, "🤖 "+v.Agent)
	if v.Branch != "" {
		parts = append(parts, "🌿 "+truncate(v.Branch, 40))
	}
	switch v.State {
	case "done":
		parts = append(parts, "✅ done")
	case "error":
		parts = append(parts, "❌ error")
	case "cancelled":
		parts = append(parts, "🛑 cancelled")
	default:
		if v.File != "" {
			parts = append(parts, "✏️ "+truncate(BaseName(v.File), 32))
		}
		if v.Cmd != "" {
			parts = append(parts, "▶ "+truncate(v.Cmd, 40))
		}
		if v.Action != "" && v.File == "" && v.Cmd == "" {
			parts = append(parts, "💭 "+truncate(v.Action, 60))
		}
	}
	if v.Committed != "" {
		parts = append(parts, "✅ committed "+v.Committed)
	}
	line := join(parts, " · ")
	if v.Starting {
		return "⏳ " + line
	}
	return line
}
