// Package config loads Telepath configuration from (in precedence order) CLI
// flags, environment variables, an optional .env file, then built-in defaults.
// It validates everything on startup with actionable error messages.
package config

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/telepath/telepath/internal/agent"
)

// Config is the fully-resolved runtime configuration.
type Config struct {
	Transport      string // telegram | slack | both
	BotToken       string // Telegram bot token
	AllowedChatIDs []int64

	// Slack settings (used when Transport includes slack).
	SlackBotToken          string
	SlackAppToken          string
	AllowedSlackUserIDs    []string
	AllowedSlackChannelIDs []string

	ProjectsDir      string
	DefaultAgent     string
	EnabledAgents    []string
	PermissionMode   agent.PermissionMode
	ApprovalTimeout  time.Duration
	BranchMode       agent.BranchMode
	BranchPrefix     string
	BranchCommit     bool
	BranchStashDirty bool
	MaxTurns         int
	Model            string
	ClaudeBin        string
	GeminiBin        string
	CodexBin         string
	AiderBin         string
	GitBin           string
	LogLevel         string
}

// Result carries the outcome of Load, including the early-exit flags.
type Result struct {
	Config      *Config
	ShowVersion bool
	ShowHelp    bool
	Usage       func()
}

func defaults() *Config {
	return &Config{
		Transport:        "telegram",
		ProjectsDir:      ".",
		DefaultAgent:     "claude",
		EnabledAgents:    []string{"claude", "gemini", "codex", "aider"},
		PermissionMode:   agent.PermissionAuto,
		ApprovalTimeout:  120 * time.Second,
		BranchMode:       agent.BranchPerTask,
		BranchPrefix:     "telepath/",
		BranchCommit:     true,
		BranchStashDirty: false,
		LogLevel:         "info",
	}
}

// Load parses args, resolves every source with correct precedence, and returns
// the result. It does not validate; call Config.Validate separately so --help
// and --version work without a complete configuration.
func Load(args []string) (*Result, error) {
	cfg := defaults()

	fs := flag.NewFlagSet("telepath", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // we print usage ourselves

	var (
		fTransport  = fs.String("transport", "", "telegram|slack|both (env TRANSPORT)")
		fSlackBot   = fs.String("slack-bot-token", "", "Slack bot token xoxb- (env SLACK_BOT_TOKEN)")
		fSlackApp   = fs.String("slack-app-token", "", "Slack app-level token xapp- (env SLACK_APP_TOKEN)")
		fSlackUsers = fs.String("allowed-slack-user-ids", "", "comma-separated allowed Slack user IDs (env ALLOWED_SLACK_USER_IDS)")
		fSlackChans = fs.String("allowed-slack-channel-ids", "", "comma-separated allowed Slack channel IDs (env ALLOWED_SLACK_CHANNEL_IDS)")
		fToken      = fs.String("telegram-bot-token", "", "Telegram bot token (env TELEGRAM_BOT_TOKEN)")
		fChats      = fs.String("allowed-chat-ids", "", "comma-separated allowed chat IDs (env ALLOWED_CHAT_IDS)")
		fProjects   = fs.String("projects-dir", "", "projects directory the agent works in (env PROJECTS_DIR)")
		fDefault    = fs.String("default-agent", "", "default agent name (env DEFAULT_AGENT)")
		fEnabled    = fs.String("enabled-agents", "", "comma-separated enabled agents (env ENABLED_AGENTS)")
		fPerm       = fs.String("permission-mode", "", "auto|allowlist|interactive (env PERMISSION_MODE)")
		fApproval   = fs.Int("approval-timeout-seconds", 0, "interactive approval timeout (env APPROVAL_TIMEOUT_SECONDS)")
		fBranchMode = fs.String("branch-mode", "", "off|per-task (env BRANCH_MODE)")
		fPrefix     = fs.String("branch-prefix", "", "task branch prefix (env BRANCH_PREFIX)")
		fCommit     = fs.String("branch-commit", "", "commit on the task branch, true|false (env BRANCH_COMMIT)")
		fStash      = fs.String("branch-stash-dirty", "", "stash a dirty tree instead of refusing, true|false (env BRANCH_STASH_DIRTY)")
		fMaxTurns   = fs.Int("max-turns", 0, "max agent turns, 0=default (env MAX_TURNS)")
		fModel      = fs.String("model", "", "model override (env MODEL)")
		fClaudeBin  = fs.String("claude-bin", "", "path to claude binary (env CLAUDE_BIN)")
		fGeminiBin  = fs.String("gemini-bin", "", "path to gemini binary (env GEMINI_BIN)")
		fCodexBin   = fs.String("codex-bin", "", "path to codex binary (env CODEX_BIN)")
		fAiderBin   = fs.String("aider-bin", "", "path to aider binary (env AIDER_BIN)")
		fGitBin     = fs.String("git-bin", "", "path to git binary (env GIT_BIN)")
		fLog        = fs.String("log-level", "", "debug|info|warn|error (env LOG_LEVEL)")
		fEnvFile    = fs.String("env-file", ".env", "path to .env file")
		fVersion    = fs.Bool("version", false, "print version and exit")
		fHelp       = fs.Bool("help", false, "print help and exit")
	)

	usage := func() { printUsage(fs) }

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return &Result{Config: cfg, ShowHelp: true, Usage: usage}, nil
		}
		return nil, err
	}

	if *fVersion {
		return &Result{Config: cfg, ShowVersion: true, Usage: usage}, nil
	}
	if *fHelp {
		return &Result{Config: cfg, ShowHelp: true, Usage: usage}, nil
	}

	// .env is the lowest-priority source after built-in defaults: it only fills
	// values that are not already present in the real environment.
	fileEnv, err := parseEnvFile(*fEnvFile)
	if err != nil {
		return nil, err
	}
	get := func(key string) string {
		if v, ok := os.LookupEnv(key); ok {
			return v
		}
		return fileEnv[key]
	}

	// Apply env / .env layer over defaults.
	if v := get("TRANSPORT"); v != "" {
		cfg.Transport = strings.ToLower(strings.TrimSpace(v))
	}
	if v := get("TELEGRAM_BOT_TOKEN"); v != "" {
		cfg.BotToken = v
	}
	if v := get("SLACK_BOT_TOKEN"); v != "" {
		cfg.SlackBotToken = v
	}
	if v := get("SLACK_APP_TOKEN"); v != "" {
		cfg.SlackAppToken = v
	}
	if v := get("ALLOWED_SLACK_USER_IDS"); v != "" {
		cfg.AllowedSlackUserIDs = splitCSV(v)
	}
	if v := get("ALLOWED_SLACK_CHANNEL_IDS"); v != "" {
		cfg.AllowedSlackChannelIDs = splitCSV(v)
	}
	if v := get("ALLOWED_CHAT_IDS"); v != "" {
		ids, err := parseChatIDs(v)
		if err != nil {
			return nil, err
		}
		cfg.AllowedChatIDs = ids
	}
	if v := get("PROJECTS_DIR"); v != "" {
		cfg.ProjectsDir = v
	}
	if v := get("DEFAULT_AGENT"); v != "" {
		cfg.DefaultAgent = v
	}
	if v := get("ENABLED_AGENTS"); v != "" {
		cfg.EnabledAgents = splitCSV(v)
	}
	if v := get("PERMISSION_MODE"); v != "" {
		cfg.PermissionMode = agent.PermissionMode(v)
	}
	if v := get("APPROVAL_TIMEOUT_SECONDS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("APPROVAL_TIMEOUT_SECONDS: %w", err)
		}
		cfg.ApprovalTimeout = time.Duration(n) * time.Second
	}
	if v := get("BRANCH_MODE"); v != "" {
		cfg.BranchMode = agent.BranchMode(v)
	}
	if v := get("BRANCH_PREFIX"); v != "" {
		cfg.BranchPrefix = v
	}
	if v := get("BRANCH_COMMIT"); v != "" {
		cfg.BranchCommit = parseBool(v, cfg.BranchCommit)
	}
	if v := get("BRANCH_STASH_DIRTY"); v != "" {
		cfg.BranchStashDirty = parseBool(v, cfg.BranchStashDirty)
	}
	if v := get("MAX_TURNS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("MAX_TURNS: %w", err)
		}
		cfg.MaxTurns = n
	}
	if v := get("MODEL"); v != "" {
		cfg.Model = v
	}
	if v := get("CLAUDE_BIN"); v != "" {
		cfg.ClaudeBin = v
	}
	if v := get("GEMINI_BIN"); v != "" {
		cfg.GeminiBin = v
	}
	if v := get("CODEX_BIN"); v != "" {
		cfg.CodexBin = v
	}
	if v := get("AIDER_BIN"); v != "" {
		cfg.AiderBin = v
	}
	if v := get("GIT_BIN"); v != "" {
		cfg.GitBin = v
	}
	if v := get("LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}

	// Flags are highest priority: apply only those explicitly set.
	set := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })
	if set["transport"] {
		cfg.Transport = strings.ToLower(strings.TrimSpace(*fTransport))
	}
	if set["telegram-bot-token"] {
		cfg.BotToken = *fToken
	}
	if set["slack-bot-token"] {
		cfg.SlackBotToken = *fSlackBot
	}
	if set["slack-app-token"] {
		cfg.SlackAppToken = *fSlackApp
	}
	if set["allowed-slack-user-ids"] {
		cfg.AllowedSlackUserIDs = splitCSV(*fSlackUsers)
	}
	if set["allowed-slack-channel-ids"] {
		cfg.AllowedSlackChannelIDs = splitCSV(*fSlackChans)
	}
	if set["allowed-chat-ids"] {
		ids, err := parseChatIDs(*fChats)
		if err != nil {
			return nil, err
		}
		cfg.AllowedChatIDs = ids
	}
	if set["projects-dir"] {
		cfg.ProjectsDir = *fProjects
	}
	if set["default-agent"] {
		cfg.DefaultAgent = *fDefault
	}
	if set["enabled-agents"] {
		cfg.EnabledAgents = splitCSV(*fEnabled)
	}
	if set["permission-mode"] {
		cfg.PermissionMode = agent.PermissionMode(*fPerm)
	}
	if set["approval-timeout-seconds"] {
		cfg.ApprovalTimeout = time.Duration(*fApproval) * time.Second
	}
	if set["branch-mode"] {
		cfg.BranchMode = agent.BranchMode(*fBranchMode)
	}
	if set["branch-prefix"] {
		cfg.BranchPrefix = *fPrefix
	}
	if set["branch-commit"] {
		cfg.BranchCommit = parseBool(*fCommit, cfg.BranchCommit)
	}
	if set["branch-stash-dirty"] {
		cfg.BranchStashDirty = parseBool(*fStash, cfg.BranchStashDirty)
	}
	if set["max-turns"] {
		cfg.MaxTurns = *fMaxTurns
	}
	if set["model"] {
		cfg.Model = *fModel
	}
	if set["claude-bin"] {
		cfg.ClaudeBin = *fClaudeBin
	}
	if set["gemini-bin"] {
		cfg.GeminiBin = *fGeminiBin
	}
	if set["codex-bin"] {
		cfg.CodexBin = *fCodexBin
	}
	if set["aider-bin"] {
		cfg.AiderBin = *fAiderBin
	}
	if set["git-bin"] {
		cfg.GitBin = *fGitBin
	}
	if set["log-level"] {
		cfg.LogLevel = *fLog
	}

	return &Result{Config: cfg, Usage: usage}, nil
}

// TelegramEnabled reports whether the Telegram transport should run.
func (c *Config) TelegramEnabled() bool {
	return c.Transport == "telegram" || c.Transport == "both"
}

// SlackEnabled reports whether the Slack transport should run.
func (c *Config) SlackEnabled() bool {
	return c.Transport == "slack" || c.Transport == "both"
}

// Validate checks the resolved config and returns the first problem found.
func (c *Config) Validate() error {
	if !c.TelegramEnabled() && !c.SlackEnabled() {
		return fmt.Errorf("TRANSPORT %q is invalid (want telegram|slack|both)", c.Transport)
	}
	// Required tokens are validated only for the enabled transports.
	if c.TelegramEnabled() {
		if strings.TrimSpace(c.BotToken) == "" {
			return errors.New("TELEGRAM_BOT_TOKEN is required when TRANSPORT includes telegram (set it in the environment, .env, or --telegram-bot-token)")
		}
		if len(c.AllowedChatIDs) == 0 {
			return errors.New("ALLOWED_CHAT_IDS is required: Telepath refuses to serve any Telegram chat until you allowlist at least one chat ID")
		}
	}
	if c.SlackEnabled() {
		if strings.TrimSpace(c.SlackBotToken) == "" {
			return errors.New("SLACK_BOT_TOKEN (xoxb-) is required when TRANSPORT includes slack")
		}
		if strings.TrimSpace(c.SlackAppToken) == "" {
			return errors.New("SLACK_APP_TOKEN (xapp-) is required when TRANSPORT includes slack (Socket Mode app-level token)")
		}
	}
	if !c.PermissionMode.Valid() {
		return fmt.Errorf("PERMISSION_MODE %q is invalid (want auto|allowlist|interactive)", c.PermissionMode)
	}
	if !c.BranchMode.Valid() {
		return fmt.Errorf("BRANCH_MODE %q is invalid (want off|per-task)", c.BranchMode)
	}
	if len(c.EnabledAgents) == 0 {
		return errors.New("ENABLED_AGENTS must list at least one agent")
	}
	if !contains(c.EnabledAgents, c.DefaultAgent) {
		return fmt.Errorf("DEFAULT_AGENT %q is not in ENABLED_AGENTS %v", c.DefaultAgent, c.EnabledAgents)
	}
	if c.ProjectsDir == "" {
		return errors.New("PROJECTS_DIR must be set")
	}
	if info, err := os.Stat(c.ProjectsDir); err != nil {
		return fmt.Errorf("PROJECTS_DIR %q is not accessible: %w", c.ProjectsDir, err)
	} else if !info.IsDir() {
		return fmt.Errorf("PROJECTS_DIR %q is not a directory", c.ProjectsDir)
	}
	return nil
}

// --- helpers ---------------------------------------------------------------

func parseEnvFile(path string) (map[string]string, error) {
	out := map[string]string{}
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return out, nil // optional file
		}
		return nil, fmt.Errorf("open env file %q: %w", path, err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		eq := strings.IndexByte(line, '=')
		if eq < 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		val = unquote(val)
		if key != "" {
			out[key] = val
		}
	}
	return out, sc.Err()
}

func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func parseChatIDs(s string) ([]int64, error) {
	var ids []int64
	for _, part := range splitCSV(s) {
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid chat id %q in ALLOWED_CHAT_IDS: %w", part, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func parseBool(s string, def bool) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}

func contains(ss []string, v string) bool {
	for _, s := range ss {
		if s == v {
			return true
		}
	}
	return false
}

func printUsage(fs *flag.FlagSet) {
	out := os.Stdout
	fmt.Fprintln(out, "telepath — a multi-agent Telegram ↔ coding-agent bridge")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Usage: telepath [flags]")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Configuration is read from flags > environment > .env > defaults.")
	fmt.Fprintln(out, "Every flag has an environment-variable equivalent (shown in its help).")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Flags:")
	fs.SetOutput(out)
	fs.PrintDefaults()
}
