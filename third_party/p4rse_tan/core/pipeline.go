package core

import (
	"context"
	"log/slog"

	"github.com/p4rtridge/p4rse_tan/errs"
	"golang.org/x/sync/errgroup"
)

// Pipeline holds a topologically resolved sequence of plugins.
type Pipeline struct {
	executionOrder []Plugin
}

// Plugins returns the resolved execution order for introspection.
func (p *Pipeline) Plugins() []Plugin {
	return p.executionOrder
}

// Execute runs all plugins in topological order, then waits for async work to complete.
// Cleanup is deferred in reverse order to guarantee safe resource teardown.
func (p *Pipeline) Execute(parent context.Context, state Context) error {
	if parent == nil {
		panic("core: pipeline Execute requires a non-nil parent context")
	}

	log := slog.Default()
	if state != nil {
		if contextual, err := KeyLogger.Get(state); err == nil && contextual != nil {
			log = contextual
		}
	}

	eg, egCtx := errgroup.WithContext(parent)
	KeyErrGroup.Set(state, eg)
	KeyContext.Set(state, egCtx)

	log.Info("pipeline starting", slog.Int("plugins", len(p.executionOrder)))

	defer func() {
		for i := len(p.executionOrder) - 1; i >= 0; i-- {
			plugin := p.executionOrder[i]
			log.Debug("[DEBUG] running cleanup", slog.String("plugin", plugin.Name()))

			if err := plugin.Cleanup(state); err != nil {
				log.Error("cleanup error",
					slog.String("plugin", plugin.Name()),
					slog.Any("error", err),
				)
			}
		}

		log.Debug("[DEBUG] pipeline cleanup complete")
	}()

	for _, plugin := range p.executionOrder {
		log.Debug("[DEBUG] executing plugin", slog.String("plugin", plugin.Name()))
		if err := plugin.Execute(state); err != nil {
			return errs.PluginExecution(plugin.Name(), err)
		}

		log.Debug("[DEBUG] plugin done", slog.String("plugin", plugin.Name()))
	}

	log.Debug("[DEBUG] waiting for async plugins")
	if err := eg.Wait(); err != nil {
		return errs.AsyncExecution(err)
	}

	log.Info("pipeline complete")
	return nil
}
