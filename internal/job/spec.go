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

// StepSpec outlines a step to run within a Runner
type StepSpec struct {
	Name    string            `yaml:"name"`
	Image   string            `yaml:"image"`
	Command []string          `yaml:"command"`
	Env     map[string]string `yaml:"env"`
	Workdir string            `yaml:"workdor"`
}
