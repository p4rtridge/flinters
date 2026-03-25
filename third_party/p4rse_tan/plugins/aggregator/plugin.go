package aggregator

import (
	"bytes"
	"fmt"
	"log/slog"
	"strconv"
	"unsafe"

	"github.com/p4rtridge/p4rse_tan/core"
	"github.com/p4rtridge/p4rse_tan/logger"
)

var (
	KeyAggregates = core.NewTypedKey[map[string]core.Context]("aggregates")
)

type Plugin struct {
	GroupKey string
	SumKeys  []string
}

func (p *Plugin) Name() string { return "AggregatorPlugin" }

func (p *Plugin) Requires() []core.DataKey {
	return []core.DataKey{core.KeyRecordStream.Key}
}

func (p *Plugin) Provides() []core.DataKey {
	return []core.DataKey{KeyAggregates.Key}
}

func (p *Plugin) Cleanup(ctx core.Context) error { return nil }

func (p *Plugin) Execute(ctx core.Context) error {
	log := logger.FromContext(ctx).With(slog.String("plugin", p.Name()))

	log.Debug("[DEBUG] starting aggregation",
		slog.String("group_key", p.GroupKey),
		slog.Any("sum_keys", p.SumKeys),
	)

	if p.GroupKey == "" {
		return aggErr("group key is required", nil)
	}

	if len(p.SumKeys) == 0 {
		return aggErr("at least one sum key is required", nil)
	}

	stream, err := core.KeyRecordStream.Get(ctx)
	if err != nil {
		return aggErr("record stream unavailable", err)
	}

	var sumKeys []core.TypedKey[float64]
	for _, colName := range p.SumKeys {
		sumKeys = append(sumKeys, core.NewTypedKey[float64](colName))
	}

	type groupState struct {
		sums []float64
	}
	stateMap := make(map[string]*groupState)

	var totalParseErrors int
	nextErrorLog := 1

	for record := range stream {
		groupValBytes, ok := record.Get(p.GroupKey)
		if !ok {
			continue // skip row missing group key entirely
		}
		groupKey := string(bytes.TrimSpace(groupValBytes))

		state, exists := stateMap[groupKey]
		if !exists {
			state = &groupState{sums: make([]float64, len(p.SumKeys))}
			stateMap[groupKey] = state
		}

		var parseErrors int
		for i, colName := range p.SumKeys {
			valBytes, ok := record.Get(colName)
			if !ok {
				continue
			}

			valBytes = bytes.TrimSpace(valBytes)
			if len(valBytes) == 0 {
				valBytes = []byte("0")
			}

			valStr := unsafe.String(unsafe.SliceData(valBytes), len(valBytes))

			val, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				parseErrors++
				continue
			}

			state.sums[i] += val
		}

		if parseErrors > 0 {
			totalParseErrors += parseErrors

			if totalParseErrors >= nextErrorLog {
				log.Warn(
					"failed to parse some numeric fields",
					slog.Int("total_errors", totalParseErrors),
					slog.Int("row_errors", parseErrors),
					slog.String("group", groupKey),
				)

				nextErrorLog = logger.NextLogThreshold(nextErrorLog)
			}
		}
	}

	log.Debug("[DEBUG] aggregation complete", slog.Int("groups", len(stateMap)))

	aggregates := make(map[string]core.Context)
	for groupKey, state := range stateMap {
		groupCtx := make(core.Context)
		core.NewTypedKey[string](p.GroupKey).Set(groupCtx, groupKey)
		for i, key := range sumKeys {
			key.Set(groupCtx, state.sums[i])
		}
		aggregates[groupKey] = groupCtx
	}

	log.Info("aggregation done", slog.Int("groups", len(aggregates)))
	KeyAggregates.Set(ctx, aggregates)

	return nil
}

func aggErr(reason string, cause error) error {
	if cause == nil {
		return fmt.Errorf("aggregator %s", reason)
	}
	return fmt.Errorf("aggregator %s: %w", reason, cause)
}
