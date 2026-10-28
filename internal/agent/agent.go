// Package agent defines the agent-agnostic contract that the rest of Telepath
// is built on: the Agent driver interface, the normalized Event stream that the
// Telegram UI consumes, and the Task/Capabilities/approval types.
//
// Nothing agent-specific (Claude/Gemini/Codex/Aider wire formats, CLI flags,
// MCP details) may leak through these types. Concrete adapters live in the
// sub-packages (claude, gemini, codex, aider) and translate a vendor CLI into
// this common vocabulary.
package agent

import (
	"context"
	"encoding/json"
	"time"
)

// PermissionMode selects how the agent's sensitive actions are gated.
type PermissionMode string

const (
	// PermissionAuto runs the agent fully autonomously with no prompts.
	PermissionAuto PermissionMode = "auto"
	// PermissionAllowlist pre-approves a safe tool set and denies the rest,
	// without any interactive prompting.
	PermissionAllowlist PermissionMode = "allowlist"
	// PermissionInteractive asks the user to approve/deny each sensitive
	// action live. Only available for agents whose Capabilities report
	// InteractiveApproval; otherwise callers must fall back to allowlist.
	PermissionInteractive PermissionMode = "interactive"
)

// Valid reports whether m is a recognized permission mode.
func (m PermissionMode) Valid() bool {
	switch m {
	case PermissionAuto, PermissionAllowlist, PermissionInteractive:
		return true
	default:
		return false
	}
}

// BranchMode selects the git branch workflow for a task.
type BranchMode string

const (
	// BranchOff runs on the current branch with no branch and no commit; the
	// UI gets the live stream plus an end-of-run diff (legacy behavior).
	BranchOff BranchMode = "off"
	// BranchPerTask runs each task on its own isolated branch and commits
	// there (the default "pro developer" flow).
	BranchPerTask BranchMode = "per-task"
)

// Valid reports whether m is a recognized branch mode.
func (m BranchMode) Valid() bool {
	switch m {
	case BranchOff, BranchPerTask:
		return true
	default:
		return false
	}
}

// Capabilities describes what an agent adapter can do, so downstream code can
// degrade gracefully for poorer agents.
type Capabilities struct {
	StructuredStream    bool // emits a machine-readable event stream (not just text)
	Sessions            bool // reports a session id that identifies the run
	Resume              bool // can continue a previous session
	InteractiveApproval bool // supports live per-action approval (e.g. MCP hook)
	FileEditEvents      bool // emits per-file edit events with diffs during the run
}

// ApprovalRequest is a single gated action awaiting a decision. It is also an
// Event so adapters may surface it on the stream for visibility, but the actual
// decision flows through Task.ApprovalHook.
type ApprovalRequest struct {
	ID    string         // unique id, assigned by the approval hook
	Tool  string         // tool/command name the agent wants to run
	Input map[string]any // raw tool input
	Risk  string         // coarse risk label: "low" | "medium" | "high"
}

func (ApprovalRequest) isEvent() {}

// ApprovalDecision is returned by an ApprovalFunc.
type ApprovalDecision struct {
	Allow bool
	Note  string // human-readable reason, surfaced to the agent on deny
	// UpdatedInput, if non-nil, replaces the tool input (allow case). Most
	// callers leave it nil to approve the input unchanged.
	UpdatedInput json.RawMessage
}

// ApprovalFunc decides a single ApprovalRequest. It blocks until the user
// responds or the implementation times out. It must respect ctx cancellation.
type ApprovalFunc func(ctx context.Context, req ApprovalRequest) ApprovalDecision

// Task is a unit of work handed to an agent.
type Task struct {
	Prompt         string
	ProjectDir     string // working directory the agent runs in (a git branch worktree)
	PermissionMode PermissionMode
	ApprovalHook   ApprovalFunc // non-nil only in interactive mode
	MaxTurns       int          // 0 = adapter/CLI default
	Model          string       // "" = adapter/CLI default
}

// Agent is a coding-agent driver.
type Agent interface {
	// Name is the stable identifier used in config and the /use command.
	Name() string
	// Available reports whether the underlying CLI is installed and runnable.
	Available(ctx context.Context) bool
	// Capabilities describes the adapter's feature set.
	Capabilities() Capabilities
	// Run executes the task and streams normalized events. The channel is
	// closed when the run finishes; a terminal Result or Error is always the
	// last meaningful event.
	Run(ctx context.Context, t Task) (<-chan Event, error)
	// Resume continues a prior session. Adapters without Capabilities.Resume
	// return ErrResumeUnsupported.
	Resume(ctx context.Context, sessionID, msg string) (<-chan Event, error)
}

// Event is the single type the UI consumes. Concrete events below implement it.
type Event interface{ isEvent() }

// AssistantText is a short "what I'm doing" line from the agent.
type AssistantText struct{ Text string }

// ToolUse is a non-file, non-command tool invocation (Read, Grep, WebFetch, …).
type ToolUse struct {
	Name         string
	InputSummary string
}

// FileEdit is the key event: a file the agent created or modified, with a
// unified diff built from the edit payload.
type FileEdit struct {
	Path        string
	UnifiedDiff string
}

// Command is a shell command the agent runs.
type Command struct{ Cmd string }

// ToolResult is the outcome of a tool invocation.
type ToolResult struct {
	OK      bool
	Summary string
}

// Progress is an optional partial-token stream chunk.
type Progress struct{ PartialText string }

// BranchInfo is emitted by the git layer when a task branch is prepared.
type BranchInfo struct {
	Name string
	Base string
}

// Committed is emitted by the git layer after a commit on the task branch.
type Committed struct {
	SHA     string
	Message string
}

// Result is the terminal success event.
type Result struct {
	Summary   string
	Duration  time.Duration
	CostUSD   *float64 // nil if the agent does not report cost
	SessionID string   // set if the agent reported one (for /resume)
}

// Error is the terminal failure event.
type Error struct{ Err error }

func (AssistantText) isEvent() {}
func (ToolUse) isEvent()       {}
func (FileEdit) isEvent()      {}
func (Command) isEvent()       {}
func (ToolResult) isEvent()    {}
func (Progress) isEvent()      {}
func (BranchInfo) isEvent()    {}
func (Committed) isEvent()     {}
func (Result) isEvent()        {}
func (Error) isEvent()         {}
