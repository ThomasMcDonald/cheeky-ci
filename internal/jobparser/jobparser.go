package jobparser

import (
	"os"

	"github.com/goccy/go-yaml"
	"github.com/thomasmcdonald/cheeky-ci/internal/job"
)

// Parser Parse the yaml.
func Parser(yml []byte) (job.Spec, error) {
	var v job.Spec

	if err := yaml.Unmarshal(yml, &v); err != nil {
		return v, err
	}

	newSteps := make([]job.StepSpec, 0, len(v.Steps))

	for _, step := range v.Steps {
		if step.Path == "" {
			newSteps = append(newSteps, step)
			continue
		}

		steps, err := os.ReadFile(step.Path)

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
