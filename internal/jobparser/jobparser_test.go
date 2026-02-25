package jobparser

import (
	"testing"
)

type testCase struct {
	path string
}

func TestParser(t *testing.T) {
	testCases := map[string]testCase{
		"Regular File": {
			path: "testdata/base.yaml",
		},
		"External Step File": {
			path: "testdata/externalSteps.yaml",
		},
	}

	for name, test := range testCases {

		t.Run(name, func(t *testing.T) {

			_, err := Parser(test.path)

			if err != nil {
				t.Error(err)
				return
			}

		})

	}

}
