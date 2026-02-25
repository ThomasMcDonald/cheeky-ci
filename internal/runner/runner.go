package runner

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ThomasMcDonald/cheeky-ci/internal/executor"
	"github.com/ThomasMcDonald/cheeky-ci/internal/job"
)

type (
	Runner interface {
		Run(ctx context.Context) error
		Shutdown(ctx context.Context) error
	}

	AgentRunner struct {
		executor executor.Executor
		job      job.Spec
		logger   *log.Logger
		stopCh   chan struct{}
	}
)

// Run cmment
func (r *AgentRunner) Run(ctx context.Context) error {
	r.logger.Printf("Starting runner using executor%s", r.executor.Name())

	job := r.job

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

// Shutdown Shutdown the agent runner by closing the channel
func (r *AgentRunner) Shutdown(ctx context.Context) error {
	r.logger.Println("runner shutting down")
	close(r.stopCh)
	return nil
}

// NewAgentRunner Create an Agent Runner instance
func NewAgentRunner(executor executor.Executor, job job.Spec) *AgentRunner {
	return &AgentRunner{
		executor: executor,
		job:      job,
		logger:   log.New(os.Stdout, "[runner] ", log.LstdFlags),
		stopCh:   make(chan struct{}),
	}
}
