package logging

import (
	"context"
	"log/slog"
)

type ctxKey int

const loggerKey ctxKey = iota

// WithLogger attaches a slog.Logger to the context.
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// LoggerFromContext returns the logger from ctx or slog.Default if missing.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if v := ctx.Value(loggerKey); v != nil {
		if l, ok := v.(*slog.Logger); ok && l != nil {
			return l
		}
	}
	return slog.Default()
}
