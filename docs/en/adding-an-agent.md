> **Languages:** [English](../en/adding-an-agent.md) · [Русский](../ru/adding-an-agent.md) · [Türkçe](../tr/adding-an-agent.md) · [Azərbaycanca](../az/adding-an-agent.md)

# Adding an agent

Adding a new agent is a new package implementing one interface, plus a single
registry line. This page shows the contract and a minimal adapter.

## The `Agent` interface

```go
type Agent interface {
    Name() string
    Available(ctx context.Context) bool
    Capabilities() Capabilities
    Run(ctx context.Context, t Task) (<-chan Event, error)
    Resume(ctx context.Context, sessionID, msg string) (<-chan Event, error)
}

type Capabilities struct {
    StructuredStream    bool // emits a machine-readable event stream
    Sessions            bool // reports a session id
    Resume              bool // can continue a previous session
    InteractiveApproval bool // supports live per-action approval (e.g. MCP)
    FileEditEvents      bool // emits per-file edit events during the run
}

type Task struct {
    Prompt, ProjectDir string
    PermissionMode     PermissionMode
    ApprovalHook       ApprovalFunc // non-nil only in interactive mode
    MaxTurns           int
    Model              string
}
```

## Events to emit

Map the CLI's output to these normalized events (the only vocabulary the UI
understands):

`AssistantText`, `ToolUse`, `FileEdit{Path, UnifiedDiff}`, `Command`,
`ToolResult`, `ApprovalRequest`, `Progress`, `Result{Summary, Duration,
CostUSD, SessionID}`, `Error`.

A rich adapter emits the full set. A simple adapter emits `AssistantText` and
lets the git layer's end-of-run diff show the file changes.

## The text + end-of-run-diff fallback

If the CLI has no structured stream, you do **not** need to parse diffs. The
shared helper `agent.StreamLines` runs the command, maps each stdout line to
events through your function, and always finishes with a terminal `Result` (or
`Error`). Set `FileEditEvents: false`; file changes still appear because the git
layer diffs the branch at the end.

## Minimal adapter

```go
package myagent

import (
    "context"
    "strings"

    "github.com/telepath/telepath/internal/agent"
    "github.com/telepath/telepath/internal/permission"
)

const binaryName = "myagent"

type MyAgent struct{ bin string }

func New(bin string) *MyAgent { return &MyAgent{bin: bin} }

func (a *MyAgent) Name() string { return "myagent" }

func (a *MyAgent) Available(ctx context.Context) bool {
    return agent.IsInstalled(binaryName, a.bin)
}

func (a *MyAgent) Capabilities() agent.Capabilities {
    return agent.Capabilities{StructuredStream: false}
}

func (a *MyAgent) Run(ctx context.Context, t agent.Task) (<-chan agent.Event, error) {
    bin := agent.ResolveBinary(binaryName, a.bin)
    args := []string{"--message", t.Prompt}
    // Map the permission mode to real flags; keep the agent's own git disabled.
    _ = permission.AiderArgs // see internal/permission/policy.go for the pattern
    return agent.StreamLines(ctx, t.ProjectDir, bin, args, func(line string) []agent.Event {
        if s := strings.TrimSpace(line); s != "" {
            return []agent.Event{agent.AssistantText{Text: s}}
        }
        return nil
    }), nil
}

func (a *MyAgent) Resume(ctx context.Context, sessionID, msg string) (<-chan agent.Event, error) {
    return nil, agent.ErrResumeUnsupported
}
```

## Register it

Add one case in `cmd/telepath/main.go`:

```go
case "myagent":
    reg.Register(myagent.New(cfg.MyAgentBin))
```

and (optionally) a `MYAGENT_BIN` field in `internal/config` if you want a binary
override. Then add your agent to `ENABLED_AGENTS` and document it in the
[agents](agents.md) table.

## Interactive approval (optional)

Only implement interactive approval if your CLI exposes a permission hook (like
Claude's MCP permission-prompt tool). If it doesn't, leave
`InteractiveApproval: false`; Telepath will automatically fall back to
`allowlist` when a user selects interactive.

## Verify the real flags first

CLI flags change often. Before coding, run the tool's `--help` and read its docs,
and implement to the **real** current interface — don't trust guessed flags.

---

<!-- prev-next -->
[← Previous: Docker](./docker.md)  ·  [Next: Security →](./security.md)
