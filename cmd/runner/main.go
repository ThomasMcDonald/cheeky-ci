package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
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
		CreateSandbox(ctx context.Context, spec JobSpec) (Sandbox, error)
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
		RunStep(ctx context.Context, step StepSpec) StepResult

		Destroy(ctx context.Context) error
	}

	JobSpec struct {
		JobID     string
		Workspace string
		Env       map[string]string
		Steps     []StepSpec
		Timeout   time.Duration
	}

	StepSpec struct {
		Name    string
		Image   string
		Command []string
		Env     map[string]string
		Workdir string
	}

	AgentRunner struct {
		executor Executor
		logger   *log.Logger
		stopCh   chan struct{}
	}
)

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

func dummyJob() JobSpec {
	return JobSpec{
		JobID:     "job-dummy-001",
		Workspace: "/tmp/ci-workspace-job-dummy-001",
		Env: map[string]string{
			"CI":         "true",
			"JOB_ID":     "job-dummy-001",
			"RUNNER_ENV": "test",
		},
		Timeout: 60 * time.Second,
		Steps: []StepSpec{
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

const META_DIR_PATH = "/var/lib/cheeky-ci-runner"

type RunnerMeta struct {
	Token string `json:"token"`
}

func main() {
	if _, ok := os.LookupEnv("ORCHESTRATOR_HOST"); !ok {
		panic("ORCHESTRATOR_HOST not set")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	filePath := filepath.Join(META_DIR_PATH, "meta.json")

	var state RunnerMeta
	file, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Meta File missing, registering runner with REGISTRATION_TOKEN")

		reader := bufio.NewReader(os.Stdin)

		fmt.Println("Please enter the registration token")

		text, err := reader.ReadString('\n')
		if err != nil {
			panic("Registering runner: Failed to read response body")
		}

		text = strings.ReplaceAll(text, "\n", "")

		fmt.Println(text)

		state.Token = text
		stateStr, err := json.Marshal(state)
		if err != nil {
			panic(err)
		}

		err = os.Mkdir(META_DIR_PATH, 0755)
		if err != nil {
			panic(err)
		}

		f, err := os.Create(filePath)
		if err != nil {
			panic(err)
		}

		defer func() {
			err := f.Close()
			if err != nil {
				panic(err)
			}
		}()

		_, err = f.Write(stateStr)
		if err != nil {
			panic(err)
		}
	} else {

		err = json.Unmarshal(file, &state)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println(state.Token)

	// Check var/lib/{name}/runner to see if RUNNER_TOKEN exists
	//
	// if no RUNNER_TOKEN
	// Check Env REGISTRATION_TOKEN variable
	// IF No panic
	// if yes call os.GetEnv('orchestrator_host')/runner/register
	// call /runner/{id}/healthcheck

	return
	executor, err := NewDockerExecutor()
	if err != nil {
		log.Fatalf("failed to create executor: %v", err)
	}

	runner := NewAgentRunner(executor)

	if err := runner.Run(ctx); err != nil {
		log.Fatalf("runner failed: %v", err)
	}
}
