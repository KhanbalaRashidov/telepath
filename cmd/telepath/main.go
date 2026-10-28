// Command telepath is a self-hosted bridge from Telegram to a coding agent
// running on the user's own machine. See README.md and ARCHITECTURE.md.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/telepath/telepath/internal/agent"
	"github.com/telepath/telepath/internal/agent/aider"
	"github.com/telepath/telepath/internal/agent/claude"
	"github.com/telepath/telepath/internal/agent/codex"
	"github.com/telepath/telepath/internal/agent/gemini"
	"github.com/telepath/telepath/internal/config"
	"github.com/telepath/telepath/internal/logx"
	"github.com/telepath/telepath/internal/permission"
	"github.com/telepath/telepath/internal/session"
	"github.com/telepath/telepath/internal/telegram"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "telepath: "+err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	res, err := config.Load(args)
	if err != nil {
		return fmt.Errorf("configuration: %w", err)
	}
	if res.ShowVersion {
		fmt.Println("telepath " + version)
		return nil
	}
	if res.ShowHelp {
		res.Usage()
		return nil
	}

	cfg := res.Config
	if err := cfg.Validate(); err != nil {
		return err
	}

	log := logx.New(cfg.LogLevel)

	reg := agent.NewRegistry()
	for _, name := range cfg.EnabledAgents {
		switch name {
		case "claude":
			reg.Register(claude.New(cfg.ClaudeBin))
		case "gemini":
			reg.Register(gemini.New(cfg.GeminiBin))
		case "codex":
			reg.Register(codex.New(cfg.CodexBin))
		case "aider":
			reg.Register(aider.New(cfg.AiderBin))
		default:
			log.Warn("ignoring unknown agent in ENABLED_AGENTS", "name", name)
		}
	}
	if len(reg.Names()) == 0 {
		return fmt.Errorf("no known agents enabled (ENABLED_AGENTS=%v)", cfg.EnabledAgents)
	}
	if err := reg.SetDefault(cfg.DefaultAgent); err != nil {
		return err
	}

	store := session.NewMemStore(session.Defaults{
		Agent:          cfg.DefaultAgent,
		PermissionMode: cfg.PermissionMode,
		BranchMode:     cfg.BranchMode,
	})
	broker := permission.NewBroker(cfg.ApprovalTimeout)

	b, err := telegram.New(cfg, reg, store, broker, log, version)
	if err != nil {
		return fmt.Errorf("telegram init: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	b.Run(ctx)
	log.Info("telepath stopped")
	return nil
}
