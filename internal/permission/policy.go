// Package permission turns a permission mode into concrete agent configuration,
// brokers live approval requests, and hosts the in-process MCP server that
// Claude Code calls to ask for a decision.
package permission

import (
	"strings"

	"github.com/telepath/telepath/internal/agent"
)

// DefaultAllowedTools is the safe tool set pre-approved in "allowlist" mode.
// Syntax matches Claude Code rules: a bare tool name, or Tool(specifier).
// Read-only and build/test commands are allowed; everything else is denied.
func DefaultAllowedTools() []string {
	return []string{
		"Read",
		"Edit",
		"Write",
		"MultiEdit",
		"Glob",
		"Grep",
		"LS",
		"NotebookEdit",
		"TodoWrite",
		"Bash(git status:*)",
		"Bash(git diff:*)",
		"Bash(git log:*)",
		"Bash(git add:*)",
		"Bash(go build:*)",
		"Bash(go test:*)",
		"Bash(go vet:*)",
		"Bash(go run:*)",
		"Bash(npm test:*)",
		"Bash(npm run build:*)",
		"Bash(ls:*)",
		"Bash(cat:*)",
		"Bash(pwd)",
	}
}

// DefaultDisallowedTools is the explicitly denied set in "allowlist" mode. These
// take precedence and block destructive or exfiltrating actions outright.
func DefaultDisallowedTools() []string {
	return []string{
		"Bash(rm:*)",
		"Bash(git push:*)",
		"Bash(git reset:*)",
		"Bash(git checkout:*)",
		"Bash(curl:*)",
		"Bash(wget:*)",
		"Bash(sudo:*)",
		"WebFetch",
	}
}

// ClaudeArgs returns the permission-related CLI flags for Claude Code for the
// given mode. Interactive wiring (the MCP flags) is added separately by the
// adapter because it depends on a running MCP server.
func ClaudeArgs(mode agent.PermissionMode) []string {
	switch mode {
	case agent.PermissionAuto:
		return []string{"--permission-mode", "bypassPermissions"}
	case agent.PermissionAllowlist:
		args := []string{"--permission-mode", "default"}
		args = append(args, "--allowedTools", strings.Join(DefaultAllowedTools(), " "))
		args = append(args, "--disallowedTools", strings.Join(DefaultDisallowedTools(), " "))
		return args
	case agent.PermissionInteractive:
		// Base mode is "default"; the adapter appends --permission-prompt-tool
		// and --mcp-config so gated tools route to our approval server.
		return []string{"--permission-mode", "default"}
	default:
		return nil
	}
}

// GeminiArgs maps a mode to Gemini CLI approval flags. Gemini has no MCP
// approval hook, so interactive degrades to the same gated behavior as
// allowlist (default approval mode, which gates sensitive actions).
func GeminiArgs(mode agent.PermissionMode) []string {
	switch mode {
	case agent.PermissionAuto:
		return []string{"--approval-mode=yolo"}
	default:
		return []string{"--approval-mode=default"}
	}
}

// CodexArgs maps a mode to Codex CLI sandbox flags. --full-auto is deprecated;
// auto uses a workspace-write sandbox with automatic approval.
func CodexArgs(mode agent.PermissionMode) []string {
	switch mode {
	case agent.PermissionAuto:
		return []string{"--sandbox", "workspace-write", "--ask-for-approval", "never"}
	default:
		return []string{"--sandbox", "workspace-write", "--ask-for-approval", "on-request"}
	}
}

// AiderArgs maps a mode to Aider flags. Aider always runs non-interactively
// here; auto/allowlist both auto-confirm because Aider has no fine-grained
// gating. Git is always disabled so Telepath owns version control.
func AiderArgs(mode agent.PermissionMode) []string {
	// --yes auto-confirms; --no-auto-commits and --no-git keep Aider away from
	// git entirely.
	return []string{"--yes", "--no-auto-commits", "--no-git"}
}

// AssessRisk gives a coarse risk label for an approval request, used purely to
// colour the Telegram prompt. It never changes the decision.
func AssessRisk(tool string, input map[string]any) string {
	switch tool {
	case "Read", "Glob", "Grep", "LS", "NotebookRead", "TodoWrite":
		return "low"
	case "Edit", "Write", "MultiEdit", "NotebookEdit":
		return "medium"
	case "Bash":
		cmd, _ := input["command"].(string)
		c := strings.ToLower(cmd)
		for _, danger := range []string{"rm ", "rm -", "git push", "git reset", "sudo", "curl", "wget", "mkfs", "dd ", ":(){", "> /dev", "format "} {
			if strings.Contains(c, danger) {
				return "high"
			}
		}
		return "medium"
	default:
		return "medium"
	}
}
