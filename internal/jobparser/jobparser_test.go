package jobparser

import (
	"fmt"
	"os"
	"testing"

	"github.com/thomasmcdonald/cheeky-ci/internal/job"
)

func TestParser(t *testing.T) {
	yml, err := os.ReadFile("testdata/dummy.yaml")

	v, err := Parser[job.Spec](yml)

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
}
