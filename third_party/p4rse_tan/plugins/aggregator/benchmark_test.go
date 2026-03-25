package aggregator

import (
	"bytes"
	"strconv"
	"testing"
	"unsafe"

	"github.com/p4rtridge/p4rse_tan/core"
)

type mockRecord struct {
	data map[string][]byte
}

func (m mockRecord) Get(key string) ([]byte, bool) {
	val, ok := m.data[key]
	return val, ok
}

func makeStream(b *testing.B) func(yield func(core.Record) bool) {
	return func(yield func(core.Record) bool) {
		rec := mockRecord{data: map[string][]byte{
			"campaign_id": []byte("camp1"),
			"clicks":      []byte("100.5"),
			"spend":       []byte("50.25"),
		}}
		for i := 0; i < b.N; i++ {
			if !yield(rec) {
				break
			}
		}
	}
}

// benchmarkOld uses map[DataKey]core.Context (boxed float64)
func benchmarkOld(b *testing.B) {
	p := &Plugin{
		GroupKey: "campaign_id",
		SumKeys:  []string{"clicks", "spend"},
	}

	aggregates := make(map[string]core.Context)
	stream := makeStream(b)

	b.ResetTimer()
	for record := range stream {
		groupValBytes, ok := record.Get(p.GroupKey)
		if !ok {
			continue
		}
		groupKey := string(bytes.TrimSpace(groupValBytes))

		groupCtx, exists := aggregates[groupKey]
		if !exists {
			groupCtx = make(core.Context)
			core.NewTypedKey[string](p.GroupKey).Set(groupCtx, groupKey)
			for _, col := range p.SumKeys {
				core.NewTypedKey[float64](col).Set(groupCtx, float64(0))
			}
			aggregates[groupKey] = groupCtx
		}

		for _, colName := range p.SumKeys {
			valBytes, _ := record.Get(colName)
			valBytes = bytes.TrimSpace(valBytes)
			if len(valBytes) == 0 {
				valBytes = []byte("0")
			}
			valStr := unsafe.String(unsafe.SliceData(valBytes), len(valBytes))
			val, _ := strconv.ParseFloat(valStr, 64)
			key := core.NewTypedKey[float64](colName)
			currentVal, _ := key.Get(groupCtx)
			key.Set(groupCtx, currentVal+val)
		}
	}
}

// benchmarkNew uses struct-based groupState — zero float boxing
func benchmarkNew(b *testing.B) {
	p := &Plugin{
		GroupKey: "campaign_id",
		SumKeys:  []string{"clicks", "spend"},
	}

	var sumKeys []core.TypedKey[float64]
	for _, colName := range p.SumKeys {
		sumKeys = append(sumKeys, core.NewTypedKey[float64](colName))
	}

	type groupState struct {
		sums []float64
	}
	stateMap := make(map[string]*groupState)
	stream := makeStream(b)

	b.ResetTimer()
	for record := range stream {
		groupValBytes, ok := record.Get(p.GroupKey)
		if !ok {
			continue
		}
		groupKey := string(bytes.TrimSpace(groupValBytes))

		state, exists := stateMap[groupKey]
		if !exists {
			state = &groupState{sums: make([]float64, len(p.SumKeys))}
			stateMap[groupKey] = state
		}

		for i, colName := range p.SumKeys {
			valBytes, _ := record.Get(colName)
			valBytes = bytes.TrimSpace(valBytes)
			if len(valBytes) == 0 {
				valBytes = []byte("0")
			}
			valStr := unsafe.String(unsafe.SliceData(valBytes), len(valBytes))
			val, _ := strconv.ParseFloat(valStr, 64)
			state.sums[i] += val
		}
	}

	// Final mapping
	aggregates := make(map[string]core.Context)
	for groupKey, state := range stateMap {
		groupCtx := make(core.Context)
		core.NewTypedKey[string](p.GroupKey).Set(groupCtx, groupKey)
		for i, key := range sumKeys {
			key.Set(groupCtx, state.sums[i])
		}
		aggregates[groupKey] = groupCtx
	}
}

func BenchmarkAggregator_Old(b *testing.B) {
	benchmarkOld(b)
}

func BenchmarkAggregator_New(b *testing.B) {
	benchmarkNew(b)
}
