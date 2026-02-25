Implementation of CI infra.


`runner -file= ./internal/jobparser/testdata/externalSteps.yaml`

Sub step file paths are relative to the parent file. in the below example steps is in `./internal/jobparser/testdata/`

parent file
```
...otherConfig
steps:
  - name: checkout
    image: alpine:3.19
    workdir: /workspace
    command:
      - sh
      - -c
      - echo hello > hello.txt

  - name: build
    path: ./steps.yaml
```