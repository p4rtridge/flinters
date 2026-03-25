package logger

import (
	"io"
	"log/slog"
	"os"

	"github.com/p4rtridge/p4rse_tan/core"
)

// Key re-exports the engine logger typed key for convenience.
var Key = core.KeyLogger

// FromContext extracts the logger injected into ctx.
// Returns a no-op logger (discards all output) when none was set,
// so callers never need to nil-check the result.
// Callers inject via: logger.Key.Set(ctx, logger.NewDefault())
func FromContext(ctx core.Context) *slog.Logger {
	log, err := Key.Get(ctx)
	if err != nil {
		return slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return log
}

// NewDefault returns a structured text logger writing to stderr at Info level.
// Also sets it as the process-wide slog default so engine code using
// slog.Default() (dag.go, pipeline.go) picks up the same handler.
func NewDefault() *slog.Logger {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	SetDefault(log)
	return log
}

// NewWithLevel returns a structured text logger writing to stderr at the given level.
func NewWithLevel(level slog.Level) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
}

// SetDefault sets log as the process-wide slog default.
// Engine code (dag.go, pipeline.go) uses slog.Default() directly — zap.L() style —
// so calling this ensures all logs share the same handler and level.
func SetDefault(log *slog.Logger) {
	slog.SetDefault(log)
}

// EnableDebug configures a debug-level logger as the process default and
// returns it ready for context injection:
//
//	ctx := make(core.Context)
//	logger.Key.Set(ctx, logger.EnableDebug())
//	pipeline.Execute(context.Background(), ctx)
func EnableDebug() *slog.Logger {
	log := NewWithLevel(slog.LevelDebug)
	SetDefault(log)
	return log
}

// NextLogThreshold calculates the next threshold for error warning logs
// to avoid flood (exponential up to 10000, then linear steps of 10000).
func NextLogThreshold(current int) int {
	if current < 10000 {
		return current * 10
	}
	return current + 10000
}
