package jobparser

import (
	"github.com/goccy/go-yaml"
)

// Parser Parse the yaml
func Parser[T any](yml string) (T, error) {
	var v T

	if err := yaml.Unmarshal([]byte(yml), &v); err != nil {
		return v, err
	}

	return v, nil
}
