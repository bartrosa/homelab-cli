// Package appctx stores per-invocation state on context.Context.
package appctx

import (
	"context"

	"github.com/bartrosa/homelab-cli/internal/config"
	"github.com/bartrosa/homelab-cli/internal/ui"
)

type ctxKey struct{}

// Session holds config and UI flags for a lab invocation.
type Session struct {
	Config      *config.Config
	ConfigPath  string
	DryRun      bool
	NoColor     bool
	Styles      ui.Styles
	HomelabRoot string // flag override
}

// WithSession attaches a session to ctx.
func WithSession(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, ctxKey{}, s)
}

// FromContext returns the session or nil.
func FromContext(ctx context.Context) *Session {
	s, _ := ctx.Value(ctxKey{}).(*Session)
	return s
}

// MustSession returns the session or panics (CLI wiring should always set it).
func MustSession(ctx context.Context) *Session {
	s := FromContext(ctx)
	if s == nil {
		panic("appctx: missing session on context")
	}
	return s
}
