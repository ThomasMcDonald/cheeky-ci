package jobparser

import (
	"fmt"
	"testing"

	"github.com/thomasmcdonald/cheeky-ci/internal/job"
)

type v struct {
	A int
	B string
}

func TestParser(t *testing.T) {
	yml := `
job_id: job-dummy-001
workspace: /tmp/ci-workspace-job-dummy-001
timeout: 60s

env:
  CI: "true"
  JOB_ID: "job-dummy-001"
  RUNNER_ENV: "test"

steps:
  - name: checkout
    image: alpine:3.19
    workdir: /workspace
    command:
      - sh
      - -c
      - echo hello > hello.txt

  - name: build
    image: alpine:3.19
    workdir: /workspace
    command:
      - cat
      - hello.txt

  - name: test
    image: alpine:3.19
    workdir: /workspace
    command:
      - sh
      - -c
      - echo failing && exit 1
`

	v, err := Parser[job.JobSpec](yml)

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
