package core

import (
	"log/slog"

	"github.com/p4rtridge/p4rse_tan/errs"
)

// BuildPipeline resolves plugin dependencies via topological sort and returns
// an executable Pipeline. ctx is used solely for logger injection.
func BuildPipeline(ctx Context, plugins []Plugin) (*Pipeline, error) {
	if ctx == nil {
		panic("core: BuildPipeline requires a non-nil context for logging purposes")
	}

	log := slog.Default()
	if contextual, err := KeyLogger.Get(ctx); err == nil && contextual != nil {
		log = contextual
	}

	available := make(map[DataKey]bool)

	remaining := make([]Plugin, len(plugins))
	copy(remaining, plugins)

	var order []Plugin

	for len(remaining) > 0 {
		progress := false
		var nextRemaining []Plugin

		for _, p := range remaining {
			// Check if all requirements are met.
			canRun := true
			var missingReq DataKey
			for _, req := range p.Requires() {
				if !available[req] {
					canRun = false
					missingReq = req
					break
				}
			}

			if canRun {
				order = append(order, p)
				progress = true
				log.Debug("[DEBUG] plugin scheduled",
					slog.String("plugin", p.Name()),
					slog.Int("position", len(order)),
				)

				for _, prov := range p.Provides() {
					available[prov] = true
				}
			} else {
				nextRemaining = append(nextRemaining, p)
				log.Debug("[DEBUG] plugin blocked: dependency not yet available",
					slog.String("plugin", p.Name()),
					slog.String("waiting_for", missingReq.UntypedKey()),
				)
			}
		}

		if !progress {
			return nil, detectDAGError(remaining, available)
		}

		remaining = nextRemaining
	}

	log.Debug("[DEBUG] pipeline built", slog.Int("plugins", len(order)))
	return &Pipeline{executionOrder: order}, nil
}

// detectDAGError classifies a stalled resolution as either a missing dependency
// or a circular dependency.
func detectDAGError(remaining []Plugin, available map[DataKey]bool) error {
	remainingProvides := make(map[DataKey]bool)
	for _, rp := range remaining {
		for _, prov := range rp.Provides() {
			remainingProvides[prov] = true
		}
	}

	for _, p := range remaining {
		for _, req := range p.Requires() {
			if available[req] {
				continue
			}
			if remainingProvides[req] {
				return errs.CircularDependency(p.Name(), req.UntypedKey())
			}
			return errs.MissingRequirement(p.Name(), req.UntypedKey())
		}
	}

	return errs.ErrCircularDependency
}
