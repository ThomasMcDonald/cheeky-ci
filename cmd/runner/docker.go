package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"runtime"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type DockerExecutor struct {
	client *client.Client
}

func NewDockerExecutor() (*DockerExecutor, error) {
	cli, err := client.New(client.FromEnv, client.WithAPIVersionFromEnv())
	if err != nil {
		return nil, err
	}

	return &DockerExecutor{
		client: cli,
	}, nil
}

func (d *DockerExecutor) Name() string {
	return "Docker"
}

func (d *DockerExecutor) Capabilities() ExecutorCapabilities {
	return ExecutorCapabilities{
		Architecture: runtime.GOARCH,
		Isolation:    "Container",
		MaxCPU:       runtime.NumCPU(),
		MaxMemoryMB:  0,
	}
}

func envMapToSlice(env map[string]string) []string {
	out := make([]string, 0, len(env))

	for k, v := range env {
		out = append(out, k+"="+v)
	}

	return out
}

func (d *DockerExecutor) CreateSandbox(ctx context.Context, spec JobSpec) (Sandbox, error) {
	image := spec.Steps[0].Image

	reader, err := d.client.ImagePull(
		ctx, image, client.ImagePullOptions{},
	)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return nil, err
	}

	err = reader.Close()
	if err != nil {
		return nil, err
	}

	containerName := "Job-" + spec.JobID

	fmt.Println(containerName)
	resp, err := d.client.ContainerCreate(
		ctx,
		client.ContainerCreateOptions{
			Config: &container.Config{
				Tty:        false,
				Env:        envMapToSlice(spec.Env),
				Cmd:        []string{"sleep", "infinity"},
				WorkingDir: "/workspace",
			},
			HostConfig: &container.HostConfig{
				Binds: []string{
					spec.Workspace + ":/workspace",
				},
			},
			Name:  containerName,
			Image: image,
		},
	)
	if err != nil {
		return nil, err
	}

	if _, err := d.client.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	return &DockerSandbox{
		client:      d.client,
		containerID: resp.ID,
	}, nil
}

type DockerSandbox struct {
	client      *client.Client
	containerID string
}

func (s *DockerSandbox) RunStep(ctx context.Context, step StepSpec) StepResult {
	execResp, err := s.client.ExecCreate(ctx, s.containerID, client.ExecCreateOptions{
		Cmd:          step.Command,
		Env:          envMapToSlice(step.Env),
		WorkingDir:   step.Workdir,
		AttachStdin:  true,
		AttachStdout: true,
	})
	if err != nil {
		return StepResult{ExitCode: -1, Error: err}
	}

	attach, err := s.client.ExecAttach(ctx, execResp.ID, client.ExecAttachOptions{})
	if err != nil {
		return StepResult{ExitCode: -1, Error: err}
	}

	defer attach.Close()

	var stdout, stderr bytes.Buffer

	_, err = stdcopy.StdCopy(&stdout, &stderr, attach.Reader)
	if err != nil {
		return StepResult{ExitCode: -1, Error: err}
	}

	inspect, err := s.client.ExecInspect(
		ctx,
		execResp.ID,
		client.ExecInspectOptions{},
	)
	if err != nil {
		return StepResult{
			ExitCode: -1, Error: err,
			Stdout: stdout.String(),
			Stderr: stderr.String(),
		}
	}

	return StepResult{
		ExitCode: inspect.ExitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}
}

func (s *DockerSandbox) Destroy(ctx context.Context) error {
	_, err := s.client.ContainerStop(ctx, s.containerID, client.ContainerStopOptions{})
	if err != nil {
		return err
	}

	_, err = s.client.ContainerRemove(ctx, s.containerID, client.ContainerRemoveOptions{
		Force: true,
	})
	if err != nil {
		return err
	}

	return nil
}
