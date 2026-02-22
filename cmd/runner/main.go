package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/thomasmcdonald/cheeky-ci/internal/job"
)

type (
	Runner interface {
		Run(ctx context.Context) error
		Shutdown(ctx context.Context) error
	}

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

	StepResult struct {
		ExitCode int
		Stdout   string
		Stderr   string
		Error    error
	}

	Sandbox interface {
		RunStep(ctx context.Context, step job.StepSpec) StepResult

		Destroy(ctx context.Context) error
	}

	AgentRunner struct {
		executor Executor
		logger   *log.Logger
		stopCh   chan struct{}
	}
)

// Run cmment
func (r *AgentRunner) Run(ctx context.Context) error {
	r.logger.Printf("Starting runner using executor%s", r.executor.Name())

	job := dummyJob()

	r.logger.Printf("Received job %s", job.JobID)

	sandbox, err := r.executor.CreateSandbox(ctx, job)
	if err != nil {
		return err
	}

	defer func() {
		err := sandbox.Destroy(ctx)
		if err != nil {
			r.logger.Printf("Destroy failed: %v", err)
		}
	}()

	for _, step := range job.Steps {
		r.logger.Printf("running step: %s", step.Name)

		stepCtx, cancel := context.WithTimeout(ctx, job.Timeout)
		result := sandbox.RunStep(stepCtx, step)
		cancel()

		r.logger.Printf("[%s] stdout: %s", step.Name, result.Stdout)
		r.logger.Printf("[%s] stderr: %s", step.Name, result.Stderr)
		if result.ExitCode != 0 {
			r.logger.Printf("step %s failed (exit=%d)", step.Name, result.ExitCode)
			return fmt.Errorf("job %s failed", job.JobID)
		}

		r.logger.Printf("[%s] Succeeded", step.Name)
	}

	r.logger.Printf("Job %s completed successfully", job.JobID)

	return nil
}

func (r *AgentRunner) Shutdown(ctx context.Context) error {
	r.logger.Println("runner shutting down")
	close(r.stopCh)
	return nil
}

func NewAgentRunner(executor Executor) *AgentRunner {
	return &AgentRunner{
		executor: executor,
		logger:   log.New(os.Stdout, "[runner] ", log.LstdFlags),
		stopCh:   make(chan struct{}),
	}
}

func dummyJob() job.Spec {
	return job.Spec{
		JobID:     "job-dummy-001",
		Workspace: "/tmp/ci-workspace-job-dummy-001",
		Env: map[string]string{
			"CI":         "true",
			"JOB_ID":     "job-dummy-001",
			"RUNNER_ENV": "test",
		},
		Timeout: 60 * time.Second,
		Steps: []job.StepSpec{
			{
				Name:    "checkout",
				Image:   "alpine:3.19",
				Command: []string{"sh", "-c", "echo hello > hello.txt"},
				Workdir: "/workspace",
			},
			{
				Name:    "build",
				Image:   "alpine:3.19",
				Command: []string{"cat", "hello.txt"},
				Workdir: "/workspace",
			},
			{
				Name:    "test",
				Image:   "alpine:3.19",
				Command: []string{"sh", "-c", "echo failing && exit 1"},
				Workdir: "/workspace",
			},
		},
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	executor, err := NewDockerExecutor()
	if err != nil {
		log.Fatalf("failed to create executor: %v", err)
	}

	runner := NewAgentRunner(executor)

	if err := runner.Run(ctx); err != nil {
		log.Fatalf("runner failed: %v", err)
	}
}
