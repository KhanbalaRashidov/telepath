// Package logx configures structured logging for Telepath.
package logx

import (
	"log/slog"
	"os"
	"strings"
)

// New builds a text slog.Logger at the given level ("debug"|"info"|"warn"|
// "error"). Unknown levels default to info.
func New(level string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})
	return slog.New(h)
}
