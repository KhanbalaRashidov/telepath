// Package codex adapts the OpenAI Codex CLI's non-interactive exec mode. Codex
// emits a JSONL event stream with `--json` (thread/turn/item lifecycle events),
// which this adapter parses on a best-effort basis into AssistantText and
// Command events; file changes are also surfaced by Telepath's end-of-run git
// diff. Codex has no MCP permission hook, so interactive approval falls back to
// allowlist. Verified: `codex exec --json --sandbox workspace-write` (the older
// `--full-auto` flag is deprecated).
package codex

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/telepath/telepath/internal/agent"
	"github.com/telepath/telepath/internal/permission"
)

const binaryName = "codex"

// Codex implements agent.Agent for the Codex CLI.
type Codex struct{ bin string }

// New creates the adapter. bin overrides the CLI path (from CODEX_BIN).
func New(bin string) *Codex { return &Codex{bin: bin} }

func (c *Codex) Name() string { return "codex" }

func (c *Codex) Available(ctx context.Context) bool {
	return agent.IsInstalled(binaryName, c.bin)
}

func (c *Codex) Capabilities() agent.Capabilities {
	return agent.Capabilities{
		StructuredStream:    true, // JSONL event stream
		Sessions:            false,
		Resume:              false,
		InteractiveApproval: false,
		FileEditEvents:      false, // edits are surfaced via git diff, not parsed
	}
}

func (c *Codex) Run(ctx context.Context, t agent.Task) (<-chan agent.Event, error) {
	bin := agent.ResolveBinary(binaryName, c.bin)
	args := []string{"exec", "--json"}
	args = append(args, permission.CodexArgs(t.PermissionMode)...)
	if t.Model != "" {
		args = append(args, "--model", t.Model)
	}
	args = append(args, t.Prompt)

	return agent.StreamLines(ctx, t.ProjectDir, bin, args, lineToEvents), nil
}

func (c *Codex) Resume(ctx context.Context, sessionID, msg string) (<-chan agent.Event, error) {
	return nil, agent.ErrResumeUnsupported
}

// lineToEvents parses one JSONL event line. The schema varies across versions,
// so this extracts text/command fields defensively and ignores the rest.
func lineToEvents(line string) []agent.Event {
	line = strings.TrimSpace(line)
	if line == "" || line[0] != '{' {
		if line != "" {
			return []agent.Event{agent.AssistantText{Text: line}}
		}
		return nil
	}
	var ev struct {
		Type string `json:"type"`
		Item struct {
			Type    string `json:"type"`
			Text    string `json:"text"`
			Command string `json:"command"`
		} `json:"item"`
		Msg struct {
			Message string `json:"message"`
		} `json:"msg"`
	}
	if err := json.Unmarshal([]byte(line), &ev); err != nil {
		return nil
	}
	switch {
	case ev.Item.Command != "":
		return []agent.Event{agent.Command{Cmd: ev.Item.Command}}
	case ev.Item.Text != "":
		return []agent.Event{agent.AssistantText{Text: ev.Item.Text}}
	case ev.Msg.Message != "":
		return []agent.Event{agent.AssistantText{Text: ev.Msg.Message}}
	default:
		return nil
	}
}
