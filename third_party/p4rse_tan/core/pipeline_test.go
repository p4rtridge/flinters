package core

import (
	"context"
	"errors"
	"testing"

	"github.com/p4rtridge/p4rse_tan/errs"
)

type stubPlugin struct {
	name    string
	reqs    []DataKey
	provs   []DataKey
	execute func(Context) error
	cleanup func(Context) error
}

func (s *stubPlugin) Name() string        { return s.name }
func (s *stubPlugin) Requires() []DataKey { return s.reqs }
func (s *stubPlugin) Provides() []DataKey { return s.provs }
func (s *stubPlugin) Cleanup(ctx Context) error {
	if s.cleanup != nil {
		return s.cleanup(ctx)
	}
	return nil
}
func (s *stubPlugin) Execute(ctx Context) error {
	if s.execute != nil {
		return s.execute(ctx)
	}
	return nil
}

func TestPipelineExecuteWrapsPluginError(t *testing.T) {
	cleanupOrder := make([]string, 0, 2)

	seedKey := DataKey{id: "seed"}
	seed := &stubPlugin{
		name:  "Seed",
		provs: []DataKey{seedKey},
		execute: func(ctx Context) error {
			ctx[seedKey] = 1
			return nil
		},
		cleanup: func(Context) error {
			cleanupOrder = append(cleanupOrder, "Seed")
			return nil
		},
	}

	executeErr := errors.New("boom")
	failing := &stubPlugin{
		name:    "FailingPlugin",
		reqs:    []DataKey{seedKey},
		execute: func(Context) error { return executeErr },
		cleanup: func(Context) error {
			cleanupOrder = append(cleanupOrder, "FailingPlugin")
			return nil
		},
	}

	pipeline := mustBuildPipeline(t, []Plugin{failing, seed})

	err := pipeline.Execute(context.Background(), make(Context))
	if err == nil {
		t.Fatal("expected error from failing plugin")
	}
	if !errors.Is(err, errs.ErrPluginExecution) {
		t.Fatalf("expected ErrPluginExecution, got %v", err)
	}
	if !errors.Is(err, executeErr) {
		t.Fatalf("expected wrapped execute error, got %v", err)
	}

	expectedCleanup := []string{"FailingPlugin", "Seed"}
	if len(cleanupOrder) != len(expectedCleanup) {
		t.Fatalf("expected %d cleanup calls, got %d", len(expectedCleanup), len(cleanupOrder))
	}
	for i, name := range expectedCleanup {
		if cleanupOrder[i] != name {
			t.Fatalf("expected cleanup order %v, got %v", expectedCleanup, cleanupOrder)
		}
	}
}

func TestPipelineExecuteAsyncError(t *testing.T) {
	seedKey := DataKey{id: "seed"}
	seed := &stubPlugin{
		name:  "Seed",
		provs: []DataKey{seedKey},
		execute: func(ctx Context) error {
			ctx[seedKey] = true
			return nil
		},
	}

	asyncErr := errors.New("async failure")
	asyncPlugin := &stubPlugin{
		name: "AsyncPlugin",
		reqs: []DataKey{seedKey},
		execute: func(ctx Context) error {
			eg, err := KeyErrGroup.Get(ctx)
			if err != nil {
				t.Fatalf("expected errgroup in context, got %v", err)
			}
			if eg == nil {
				t.Fatal("expected non-nil errgroup")
			}
			eg.Go(func() error {
				return asyncErr
			})
			return nil
		},
	}

	pipeline := mustBuildPipeline(t, []Plugin{asyncPlugin, seed})

	err := pipeline.Execute(context.Background(), make(Context))
	if err == nil {
		t.Fatal("expected async failure")
	}
	if !errors.Is(err, errs.ErrAsyncExecution) {
		t.Fatalf("expected ErrAsyncExecution, got %v", err)
	}
	if !errors.Is(err, asyncErr) {
		t.Fatalf("expected wrapped async error, got %v", err)
	}
}

func TestPipelineExecutePanicsWithoutParent(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when parent context is nil")
		}
	}()
	pipeline := &Pipeline{}
	pipeline.Execute(nil, make(Context))
}

func TestPipelinePluginsAccessor(t *testing.T) {
	plugins := []Plugin{
		&stubPlugin{name: "A"},
		&stubPlugin{name: "B"},
	}
	p := &Pipeline{executionOrder: plugins}
	ordered := p.Plugins()
	if len(ordered) != len(plugins) {
		t.Fatalf("expected %d plugins, got %d", len(plugins), len(ordered))
	}
	for i, plugin := range ordered {
		if plugin.Name() != plugins[i].Name() {
			t.Fatalf("expected plugin %s at position %d, got %s", plugins[i].Name(), i, plugin.Name())
		}
	}
}

func TestPipelineSetsContextKeys(t *testing.T) {
	state := make(Context)
	executeErr := errors.New("boom")
	plug := &stubPlugin{
		name:    "FailFast",
		execute: func(Context) error { return executeErr },
	}

	pipeline := mustBuildPipeline(t, []Plugin{plug})

	err := pipeline.Execute(context.Background(), state)
	if err == nil {
		t.Fatal("expected execution failure")
	}
	if _, err := KeyErrGroup.Get(state); err != nil {
		t.Fatalf("expected errgroup stored in context, got %v", err)
	}
	if _, err := KeyContext.Get(state); err != nil {
		t.Fatalf("expected pipeline context stored, got %v", err)
	}
}

func mustBuildPipeline(t *testing.T, plugins []Plugin) *Pipeline {
	t.Helper()
	pipeline, err := BuildPipeline(make(Context), plugins)
	if err != nil {
		t.Fatalf("failed to build pipeline: %v", err)
	}
	return pipeline
}
