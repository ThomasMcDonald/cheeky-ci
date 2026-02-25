package job

import "time"

// Spec Top level struct to run in a runner.
type Spec struct {
	JobID     string            `yaml:"job_id"`
	Workspace string            `yaml:"workspace"`
	Env       map[string]string `yaml:"env"`
	Steps     []StepSpec        `yaml:"steps"`
	Timeout   time.Duration     `yaml:"timeout"`
}

// SubJobSpec contains only steps within the file
type SubJobSpec struct {
	Steps []StepSpec `yaml:"steps"`
}

// StepSpec outlines a step to run within a Runner
type StepSpec struct {
	Name    string            `yaml:"name"`
	Path    string            `yaml:"path"`
	Image   string            `yaml:"image"`
	Command []string          `yaml:"command"`
	Env     map[string]string `yaml:"env"`
	Workdir string            `yaml:"workdor"`
}

// StepResult the result of the step
type StepResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Error    error
}
