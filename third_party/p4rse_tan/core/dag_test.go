package core

import (
	"context"
	"errors"
	"testing"

	"github.com/p4rtridge/p4rse_tan/errs"
)

// mockPlugin implements Plugin for testing
type mockPlugin struct {
	name     string
	reqs     []DataKey
	provs    []DataKey
	runFunc  func(Context) error
	executed bool
}

func (m *mockPlugin) Name() string              { return m.name }
func (m *mockPlugin) Requires() []DataKey       { return m.reqs }
func (m *mockPlugin) Provides() []DataKey       { return m.provs }
func (m *mockPlugin) Cleanup(ctx Context) error { return nil }
func (m *mockPlugin) Execute(ctx Context) error {
	m.executed = true
	if m.runFunc != nil {
		return m.runFunc(ctx)
	}
	for _, p := range m.provs {
		ctx[p] = "mock_value_for_" + p.id
	}
	return nil
}

func TestBuildPipelinePanicsWithoutContext(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when BuildPipeline context is nil")
		}
	}()
	BuildPipeline(nil, nil)
}

func TestDAGResolution(t *testing.T) {
	pluginA := &mockPlugin{
		name:  "PluginA",
		reqs:  []DataKey{{id: "initial_data"}},
		provs: []DataKey{{id: "data_a"}},
	}
	pluginB := &mockPlugin{
		name:  "PluginB",
		reqs:  []DataKey{{id: "data_a"}},
		provs: []DataKey{{id: "data_b"}},
	}
	pluginC := &mockPlugin{
		name:  "PluginC",
		reqs:  []DataKey{{id: "data_a"}, {id: "data_b"}},
		provs: []DataKey{{id: "data_c"}},
	}
	pluginInit := &mockPlugin{
		name:  "Init",
		provs: []DataKey{{id: "initial_data"}},
	}

	plugins := []Plugin{pluginC, pluginB, pluginA, pluginInit} // Shuffle order

	pipeline, err := BuildPipeline(make(Context), plugins)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	order := pipeline.Plugins()
	if len(pipeline.executionOrder) != 4 {
		t.Fatalf("expected 4 plugins in pipeline, got %d", len(pipeline.executionOrder))
	}

	// Order must be Init -> A -> B -> C
	expected := []string{"Init", "PluginA", "PluginB", "PluginC"}
	for i, p := range order {
		if p.Name() != expected[i] {
			var actualNames []string
			for _, op := range order {
				actualNames = append(actualNames, op.Name())
			}
			t.Errorf("expected order %v, got %v", expected, actualNames)
			break
		}
	}
}

func TestDAGMissingRequirement(t *testing.T) {
	pluginA := &mockPlugin{
		name:  "PluginA",
		reqs:  []DataKey{{id: "missing_data"}},
		provs: []DataKey{{id: "data_a"}},
	}
	pluginInit := &mockPlugin{
		name:  "Init",
		provs: []DataKey{{id: "initial_data"}},
	}

	_, err := BuildPipeline(make(Context), []Plugin{pluginA, pluginInit})
	if err == nil {
		t.Fatal("expected error for missing requirement, got nil")
	}
	if !errors.Is(err, errs.ErrMissingRequirement) {
		t.Fatalf("expected ErrMissingRequirement, got %v", err)
	}
}

func TestDAGCircularDependency(t *testing.T) {
	pluginA := &mockPlugin{
		name:  "PluginA",
		reqs:  []DataKey{{id: "data_b"}},
		provs: []DataKey{{id: "data_a"}},
	}
	pluginB := &mockPlugin{
		name:  "PluginB",
		reqs:  []DataKey{{id: "data_a"}},
		provs: []DataKey{{id: "data_b"}},
	}

	_, err := BuildPipeline(make(Context), []Plugin{pluginA, pluginB})
	if err == nil {
		t.Fatal("expected error for circular dependency, got nil")
	}
	if !errors.Is(err, errs.ErrCircularDependency) {
		t.Fatalf("expected ErrCircularDependency, got %v", err)
	}
}

func TestDetectDAGErrorPartialAvailability(t *testing.T) {
	availableKey := DataKey{id: "available"}
	missingKey := DataKey{id: "missing"}
	plugin := &mockPlugin{
		name: "Partial",
		reqs: []DataKey{availableKey, missingKey},
	}
	available := map[DataKey]bool{availableKey: true}
	err := detectDAGError([]Plugin{plugin}, available)
	if err == nil {
		t.Fatal("expected missing requirement error")
	}
	if !errors.Is(err, errs.ErrMissingRequirement) {
		t.Fatalf("expected ErrMissingRequirement, got %v", err)
	}
}

func TestDAGExecution(t *testing.T) {
	pluginA := &mockPlugin{
		name:  "PluginA",
		reqs:  []DataKey{{id: "initial"}},
		provs: []DataKey{{id: "a"}},
		runFunc: func(ctx Context) error {
			ctx[DataKey{id: "a"}] = ctx[DataKey{id: "initial"}].(int) * 2
			return nil
		},
	}
	pluginB := &mockPlugin{
		name:  "PluginB",
		reqs:  []DataKey{{id: "a"}},
		provs: []DataKey{{id: "b"}},
		runFunc: func(ctx Context) error {
			ctx[DataKey{id: "b"}] = ctx[DataKey{id: "a"}].(int) + 3
			return nil
		},
	}
	pluginInit := &mockPlugin{
		name:  "Init",
		provs: []DataKey{{id: "initial"}},
		runFunc: func(ctx Context) error {
			ctx[DataKey{id: "initial"}] = 5
			return nil
		},
	}

	pipeline, _ := BuildPipeline(make(Context), []Plugin{pluginB, pluginA, pluginInit})

	ctx := make(Context)
	err := pipeline.Execute(context.Background(), ctx)
	if err != nil {
		t.Fatalf("expected no execution error, got %v", err)
	}

	if ctx[DataKey{id: "b"}] != 13 {
		t.Errorf("expected ctx['b'] to be 13, got %v", ctx[DataKey{id: "b"}])
	}
}
