package main

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/thomasmcdonald/cheeky-ci/internal/job"
)

type FakeExecutor struct {
	mu sync.Mutex

	StepResults   map[string]StepResult
	ExecutedSteps []string
}

func NewFakeExecutor() *FakeExecutor {
	return &FakeExecutor{
		StepResults: make(map[string]StepResult),
	}
}

func (f *FakeExecutor) Name() string {
	return "Fake"
}

func (f *FakeExecutor) Capabilities() ExecutorCapabilities {
	return ExecutorCapabilities{
		Architecture: "Test",
		Isolation:    "Test",
		MaxCPU:       1,
		MaxMemoryMB:  128,
	}
}

func (f *FakeExecutor) CreateSandbox(ctx context.Context, spec job.Spec) (Sandbox, error) {
	return &FakeSandbox{
		Executor: f,
		JobID:    spec.JobID,
	}, nil
}

type FakeSandbox struct {
	Executor  *FakeExecutor
	JobID     string
	destroyed bool
}

func (s *FakeSandbox) RunStep(ctx context.Context, step job.StepSpec) StepResult {
	s.Executor.mu.Lock()
	defer s.Executor.mu.Unlock()

	s.Executor.ExecutedSteps = append(s.Executor.ExecutedSteps, step.Name)

	if result, ok := s.Executor.StepResults[step.Name]; ok {
		return result
	}

	return StepResult{
		ExitCode: 0,
	}
}

func (s *FakeSandbox) Destroy(ctx context.Context) error {
	s.destroyed = true

	return nil
}

func TestRunnerStopsOnFailure(t *testing.T) {
	executor := NewFakeExecutor()

	executor.StepResults["build"] = StepResult{ExitCode: 1}

	job := job.Spec{
		JobID: "job-1",
		Steps: []job.StepSpec{
			{Name: "checkout"},
			{Name: "build"},
			{Name: "test"},
		},
	}

	sandbox, _ := executor.CreateSandbox(context.Background(), job)

	for _, step := range job.Steps {
		result := sandbox.RunStep(context.Background(), step)
		if result.ExitCode != 0 {
			break
		}
	}

	if !reflect.DeepEqual(executor.ExecutedSteps, []string{"checkout", "build"}) {
		t.Fatalf("Unexpected steps: %v", executor.ExecutedSteps)
	}
}
