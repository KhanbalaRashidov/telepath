// Package gemini adapts the Gemini CLI. Gemini's non-interactive JSON output
// has been unstable across releases, so this adapter runs in plain text mode
// and maps each output line to an AssistantText event; file changes are
// surfaced by Telepath's end-of-run git diff. Gemini has no MCP permission
// hook, so interactive approval is not available (callers fall back to
// allowlist). Verified flags: `-p` (non-interactive prompt), `-m` (model),
// `--approval-mode=yolo|default`.
package gemini

import (
	"context"
	"strings"

	"github.com/telepath/telepath/internal/agent"
	"github.com/telepath/telepath/internal/permission"
)

const binaryName = "gemini"

// Gemini implements agent.Agent for the Gemini CLI.
type Gemini struct{ bin string }

// New creates the adapter. bin overrides the CLI path (from GEMINI_BIN).
func New(bin string) *Gemini { return &Gemini{bin: bin} }

func (g *Gemini) Name() string { return "gemini" }

func (g *Gemini) Available(ctx context.Context) bool {
	return agent.IsInstalled(binaryName, g.bin)
}

func (g *Gemini) Capabilities() agent.Capabilities {
	return agent.Capabilities{
		StructuredStream:    false,
		Sessions:            false,
		Resume:              false,
		InteractiveApproval: false,
		FileEditEvents:      false,
	}
}

func (g *Gemini) Run(ctx context.Context, t agent.Task) (<-chan agent.Event, error) {
	bin := agent.ResolveBinary(binaryName, g.bin)
	args := []string{"-p", t.Prompt}
	if t.Model != "" {
		args = append(args, "-m", t.Model)
	}
	args = append(args, permission.GeminiArgs(t.PermissionMode)...)

	return agent.StreamLines(ctx, t.ProjectDir, bin, args, lineToEvents), nil
}

func (g *Gemini) Resume(ctx context.Context, sessionID, msg string) (<-chan agent.Event, error) {
	return nil, agent.ErrResumeUnsupported
}

func lineToEvents(line string) []agent.Event {
	if s := strings.TrimSpace(line); s != "" {
		return []agent.Event{agent.AssistantText{Text: s}}
	}
	return nil
}
