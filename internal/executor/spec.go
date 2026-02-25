package executor

import (
	"context"

	"github.com/ThomasMcDonald/cheeky-ci/internal/job"
)

type (
	ExecutorCapabilities struct {
		Architecture string
		Isolation    string
		MaxCPU       int
		MaxMemoryMB  int
	}
	Executor interface {
		CreateSandbox(ctx context.Context, spec job.Spec) (Sandbox, error)
		Name() string
		Capabilities() ExecutorCapabilities
	}

	Sandbox interface {
		RunStep(ctx context.Context, step job.StepSpec) job.StepResult

		Destroy(ctx context.Context) error
	}
)
