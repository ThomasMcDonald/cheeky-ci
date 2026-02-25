package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/thomasmcdonald/cheeky-ci/internal/executor"
	"github.com/thomasmcdonald/cheeky-ci/internal/jobparser"
	"github.com/thomasmcdonald/cheeky-ci/internal/runner"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	filePath := flag.String("file", "", "YAML Job Specification file path")

	flag.Parse()

	if *filePath == "" {
		panic("YAML Job specification file path missing. usage: -file=")
	}

	data, err := os.ReadFile(*filePath)

	if err != nil {
		panic(fmt.Errorf("Failed to read YAML job Specification:L %v", err))
	}

	job, err := jobparser.Parser(data)

	if err != nil {
		panic(fmt.Errorf("Failed to parse YAML job specification: %v", err))
	}

	executor, err := executor.NewDockerExecutor()
	if err != nil {
		log.Fatalf("failed to create executor: %v", err)
	}

	runner := runner.NewAgentRunner(executor, job)

	if err := runner.Run(ctx); err != nil {
		log.Fatalf("runner failed: %v", err)
	}
}
