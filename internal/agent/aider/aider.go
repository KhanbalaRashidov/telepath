// Package aider adapts the Aider CLI. Aider's own git integration is disabled
// entirely (--no-auto-commits --no-git) so Telepath owns version control, and it
// runs non-interactively (--message --yes). Aider streams plain text, mapped to
// AssistantText events; file changes are surfaced by Telepath's end-of-run git
// diff. Aider has no MCP permission hook, so interactive approval is not
// available (callers fall back to allowlist).
package aider

import (
	"context"
	"strings"

	"github.com/telepath/telepath/internal/agent"
	"github.com/telepath/telepath/internal/permission"
)

const binaryName = "aider"

// Aider implements agent.Agent for the Aider CLI.
type Aider struct{ bin string }

// New creates the adapter. bin overrides the CLI path (from AIDER_BIN).
func New(bin string) *Aider { return &Aider{bin: bin} }

func (a *Aider) Name() string { return "aider" }

func (a *Aider) Available(ctx context.Context) bool {
	return agent.IsInstalled(binaryName, a.bin)
}

func (a *Aider) Capabilities() agent.Capabilities {
	return agent.Capabilities{
		StructuredStream:    false,
		Sessions:            false,
		Resume:              false,
		InteractiveApproval: false,
		FileEditEvents:      false,
	}
}

func (a *Aider) Run(ctx context.Context, t agent.Task) (<-chan agent.Event, error) {
	bin := agent.ResolveBinary(binaryName, a.bin)
	args := []string{"--message", t.Prompt, "--no-stream"}
	if t.Model != "" {
		args = append(args, "--model", t.Model)
	}
	args = append(args, permission.AiderArgs(t.PermissionMode)...)

	return agent.StreamLines(ctx, t.ProjectDir, bin, args, lineToEvents), nil
}

func (a *Aider) Resume(ctx context.Context, sessionID, msg string) (<-chan agent.Event, error) {
	return nil, agent.ErrResumeUnsupported
}

func lineToEvents(line string) []agent.Event {
	if s := strings.TrimSpace(line); s != "" {
		return []agent.Event{agent.AssistantText{Text: s}}
	}
	return nil
}
