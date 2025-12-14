// Package main provides the entry point for the status-line CLI tool.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/florent/status-line/internal/adapter/git"
	"github.com/florent/status-line/internal/adapter/system"
	"github.com/florent/status-line/internal/adapter/terminal"
	"github.com/florent/status-line/internal/application"
	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/presentation/renderer"
)

func main() {
	input, err := readInput()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	svc := buildService()
	fmt.Print(svc.Generate(input))
}

func readInput() (*model.Input, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("reading stdin: %w", err)
	}

	var input model.Input
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}
	return &input, nil
}

func buildService() *application.StatusLineService {
	return application.NewStatusLineService(
		git.NewRepository(),
		system.NewProvider(),
		terminal.NewProvider(),
		renderer.NewPowerline(),
	)
}
