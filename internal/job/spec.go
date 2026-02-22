package job

import "time"

type JobSpec struct {
	JobID     string            `yaml:"job_id"`
	Workspace string            `yaml:"workspace"`
	Env       map[string]string `yaml:"env"`
	Steps     []StepSpec        `yaml:"steps"`
	Timeout   time.Duration     `yaml:"timeout"`
}

type StepSpec struct {
	Name    string            `yaml: name"`
	Image   string            `yaml:"image"`
	Command []string          `yaml:"command"`
	Env     map[string]string `yaml:"env"`
	Workdir string            `yaml:"workdor"`
}
