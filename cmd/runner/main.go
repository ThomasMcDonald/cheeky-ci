package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/thomasmcdonald/cheeky-ci/internal/job"
	"github.com/thomasmcdonald/cheeky-ci/internal/jobparser"
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
func NewAgentRunner(executor Executor, job job.Spec) *AgentRunner {
	return &AgentRunner{
		executor: executor,
		job:      job,
		logger:   log.New(os.Stdout, "[runner] ", log.LstdFlags),
		stopCh:   make(chan struct{}),
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	filePath := flag.String("file", "", "YAML Job Specification file path")

	flag.Parse()

	if *filePath == "" {
		panic("YAML Job specification file path missing. usage: -file=")
	}

	data, err := os.ReadFile(*filePath)

	if err != nil {
		panic(fmt.Errorf("Failed to read YAML job Specification:L %v", err))
	}

	job, err := jobparser.Parser[job.Spec](data)

	if err != nil {
		panic(fmt.Errorf("Failed to parse YAML job specification: %v", err))
	}

	executor, err := NewDockerExecutor()
	if err != nil {
		log.Fatalf("failed to create executor: %v", err)
	}

	runner := NewAgentRunner(executor, job)

	if err := runner.Run(ctx); err != nil {
		log.Fatalf("runner failed: %v", err)
	}
}
