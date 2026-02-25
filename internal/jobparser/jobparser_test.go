package jobparser

import (
	"fmt"
	"os"
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
			yml, err := os.ReadFile(test.path)

			v, err := Parser(yml)

			if err != nil {
				t.Error(err)
				return
			}

			fmt.Println(v.JobID)
			if len(v.Steps) != 3 {

				// output error with YAML source

				fmt.Printf("it aint work")

				t.Fail()
			}

		})

	}

}
