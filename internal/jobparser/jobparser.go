package jobparser

import (
	"github.com/goccy/go-yaml"
)

// Parser Parse the yaml. No reason for this to by Generic, thought I'd try something new.
func Parser[T any](yml []byte) (T, error) {
	var v T

	if err := yaml.Unmarshal(yml, &v); err != nil {
		return v, err
	}

	return v, nil
}
