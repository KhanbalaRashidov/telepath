// Package claude is the reference Telepath adapter: it drives the Claude Code
// CLI in headless streaming mode and maps its stream-json output to the
// normalized agent.Event model. It fully supports structured streaming, file
// edit events, session resume, and live interactive approval via an in-process
// MCP permission-prompt server.
package claude

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/telepath/telepath/internal/agent"
	"github.com/telepath/telepath/internal/permission"
)

// binaryName is the default CLI binary looked up on PATH.
const binaryName = "claude"

// Claude implements agent.Agent for the Claude Code CLI.
type Claude struct {
	bin string

	mu   sync.Mutex
	dirs map[string]string // sessionID -> project dir, for Resume
}

// New creates the adapter. bin overrides the CLI path (from CLAUDE_BIN).
func New(bin string) *Claude {
	return &Claude{bin: bin, dirs: make(map[string]string)}
}

func (c *Claude) Name() string { return "claude" }

func (c *Claude) Available(ctx context.Context) bool {
	return agent.IsInstalled(binaryName, c.bin)
}

func (c *Claude) Capabilities() agent.Capabilities {
	return agent.Capabilities{
		StructuredStream:    true,
		Sessions:            true,
		Resume:              true,
		InteractiveApproval: true,
		FileEditEvents:      true,
	}
}

func (c *Claude) remember(sessionID, dir string) {
	if sessionID == "" {
		return
	}
	c.mu.Lock()
	c.dirs[sessionID] = dir
	c.mu.Unlock()
}

func (c *Claude) recall(sessionID string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.dirs[sessionID]
}

// Run executes the task and streams events.
func (c *Claude) Run(ctx context.Context, t agent.Task) (<-chan agent.Event, error) {
	bin := agent.ResolveBinary(binaryName, c.bin)

	args := []string{
		"-p", t.Prompt,
		"--output-format", "stream-json",
		"--verbose",
		"--include-partial-messages",
	}
	if t.MaxTurns > 0 {
		args = append(args, "--max-turns", itoa(t.MaxTurns))
	}
	if t.Model != "" {
		args = append(args, "--model", t.Model)
	}

	// Permission wiring.
	var mcp *permission.MCPServer
	var mcpConfigPath string
	mode := t.PermissionMode
	if mode == agent.PermissionInteractive && t.ApprovalHook == nil {
		// No hook supplied: degrade to allowlist so we never silently auto-run.
		mode = agent.PermissionAllowlist
	}
	if mode == agent.PermissionInteractive {
		srv, err := permission.StartMCPServer(t.ApprovalHook)
		if err != nil {
			return nil, err
		}
		cfgPath, err := writeTempFile("telepath-mcp-*.json", srv.MCPConfigJSON())
		if err != nil {
			_ = srv.Close()
			return nil, err
		}
		mcp = srv
		mcpConfigPath = cfgPath
		args = append(args, permission.ClaudeArgs(agent.PermissionInteractive)...)
		args = append(args, "--permission-prompt-tool", srv.ToolRef(), "--mcp-config", cfgPath)
	} else {
		args = append(args, permission.ClaudeArgs(mode)...)
	}

	out := make(chan agent.Event, 64)
	go c.stream(ctx, bin, t.ProjectDir, args, out, mcp, mcpConfigPath, true)
	return out, nil
}

// Resume continues a prior session. The project directory is recalled from the
// original Run (the Agent interface does not carry it). Resume runs in auto
// permission mode.
func (c *Claude) Resume(ctx context.Context, sessionID, msg string) (<-chan agent.Event, error) {
	bin := agent.ResolveBinary(binaryName, c.bin)
	dir := c.recall(sessionID)
	if dir == "" {
		dir = "."
	}
	args := []string{
		"--resume", sessionID,
		"-p", msg,
		"--output-format", "stream-json",
		"--verbose",
		"--include-partial-messages",
	}
	args = append(args, permission.ClaudeArgs(agent.PermissionAuto)...)

	out := make(chan agent.Event, 64)
	go c.stream(ctx, bin, dir, args, out, nil, "", false)
	return out, nil
}

// stream runs the CLI, pumps events, guarantees a terminal event, and cleans up.
func (c *Claude) stream(ctx context.Context, bin, dir string, args []string, out chan agent.Event, mcp *permission.MCPServer, mcpConfigPath string, remember bool) {
	defer close(out)
	if mcp != nil {
		defer mcp.Close()
	}
	if mcpConfigPath != "" {
		defer os.Remove(mcpConfigPath)
	}

	parser := NewParser()
	sawTerminal := false
	lineFn := func(line string) []agent.Event {
		evs := parser.Parse(line)
		for _, e := range evs {
			switch e.(type) {
			case agent.Result, agent.Error:
				sawTerminal = true
			}
		}
		return evs
	}

	cmd := agent.NewCommand(ctx, dir, bin, args...)
	stderr, err := agent.Pump(cmd, lineFn, out)

	if remember {
		c.remember(parser.SessionID(), dir)
	}

	if !sawTerminal {
		if err != nil {
			out <- agent.Error{Err: runError(err, stderr)}
		} else {
			out <- agent.Result{Summary: "completed", SessionID: parser.SessionID()}
		}
	}
}

func writeTempFile(pattern string, data []byte) (string, error) {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return "", err
	}
	return f.Name(), nil
}

func itoa(n int) string { return strconv.Itoa(n) }

func durationFromMS(ms int64) time.Duration {
	return time.Duration(ms) * time.Millisecond
}

// runError combines the process error with any stderr output for context.
func runError(err error, stderr string) error {
	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return err
	}
	if len(stderr) > 500 {
		stderr = stderr[len(stderr)-500:]
	}
	return fmt.Errorf("%w: %s", err, stderr)
}
