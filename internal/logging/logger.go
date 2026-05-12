// Package logging configures slog handlers for the lab CLI.
package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// New constructs a slog.Logger writing to w. If w is nil, stderr is used.
func New(w io.Writer, level, format string, noColor bool) *slog.Logger {
	if w == nil {
		w = os.Stderr
	}
	_ = noColor // reserved for future text styling; slog handlers are plain by default.
	h := handlerFor(w, level, format)
	return slog.New(h)
}

func handlerFor(w io.Writer, level, format string) slog.Handler {
	lvl := parseLevel(level)
	opts := &slog.HandlerOptions{Level: lvl}

	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return slog.NewJSONHandler(w, opts)
	default:
		return slog.NewTextHandler(w, opts)
	}
}

func parseLevel(s string) slog.Leveler {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
