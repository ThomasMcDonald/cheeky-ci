package jobparser

import (
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/thomasmcdonald/cheeky-ci/internal/job"
)

// Parser Parse the yaml.
func Parser(yml []byte) (job.Spec, error) {
	var v job.Spec

	if err := yaml.Unmarshal(yml, &v); err != nil {
		return v, err
	}

	for _, step := range v.Steps {
		if step.Path != "" {
			fmt.Println("Path to step found")
		}
	}

	return v, nil
}
