package jobparser

import (
	"os"
	"path/filepath"

	"github.com/ThomasMcDonald/cheeky-ci/internal/job"
	"github.com/goccy/go-yaml"
)

// Parser Parse the yaml.
func Parser(path string) (job.Spec, error) {
	var v job.Spec
	var dir = filepath.Dir(path)

	yml, err := os.ReadFile(path)

	if err != nil {
		return v, err
	}

	if err := yaml.Unmarshal(yml, &v); err != nil {
		return v, err
	}

	newSteps := make([]job.StepSpec, 0, len(v.Steps))

	for _, step := range v.Steps {
		if step.Path == "" {
			newSteps = append(newSteps, step)
			continue
		}

		steps, err := os.ReadFile(filepath.Join(dir, step.Path))

		if err != nil {
			return v, err
		}

		var subSteps job.SubJobSpec

		if err := yaml.Unmarshal(steps, &subSteps); err != nil {
			return v, err
		}

		newSteps = append(newSteps, subSteps.Steps...)

	}

	v.Steps = newSteps

	return v, nil
}
