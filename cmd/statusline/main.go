// Package main provides the entry point for the status-line CLI tool.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/florent/status-line/internal/adapter/git"
	"github.com/florent/status-line/internal/adapter/mcp"
	"github.com/florent/status-line/internal/adapter/system"
	"github.com/florent/status-line/internal/adapter/taskwarrior"
	"github.com/florent/status-line/internal/adapter/terminal"
	"github.com/florent/status-line/internal/adapter/updater"
	"github.com/florent/status-line/internal/adapter/usage"
	"github.com/florent/status-line/internal/application"
	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/presentation/renderer"
)

// version is set at build time via ldflags.
// Empty value means development build (no auto-update).
var version string

// main is the entry point of the application.
// It reads JSON input from stdin, builds the service, and outputs the status line.
//
// Returns:
//   - void: exits with code 1 on error
func main() {
	input, err := readInput()
	// Check for input reading errors
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	svc := buildService(input.WorkingDir())
	fmt.Print(svc.Generate(input))

	// Check for updates after output (non-blocking to user)
	checkForUpdates()
}

// checkForUpdates performs update check in background.
// Silently ignores errors to avoid disrupting normal operation.
func checkForUpdates() {
	u := updater.NewUpdater(version)
	// Ignore errors - update is best-effort
	_, _ = u.CheckAndUpdate()
}

// readInput reads and parses JSON input from stdin.
//
// Returns:
//   - *model.Input: parsed input data
//   - error: reading or parsing error if any
func readInput() (*model.Input, error) {
	data, err := io.ReadAll(os.Stdin)
	// Check for stdin read errors
	if err != nil {
		// Return wrapped error for context
		return nil, fmt.Errorf("reading stdin: %w", err)
	}

	var input model.Input
	// Check for JSON parsing errors
	if err := json.Unmarshal(data, &input); err != nil {
		// Return wrapped error for context
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}
	// Return successfully parsed input
	return &input, nil
}

// buildService creates and wires all dependencies for the status line service.
//
// Params:
//   - projectDir: the project directory for MCP config lookup
//
// Returns:
//   - *application.StatusLineService: fully configured service instance
func buildService(projectDir string) *application.StatusLineService {
	deps := application.ServiceDeps{
		Git:         git.NewRepository(),
		System:      system.NewProvider(),
		Terminal:    terminal.NewProvider(),
		MCP:         mcp.NewProvider(projectDir),
		Taskwarrior: taskwarrior.NewProvider(),
		Usage:       usage.NewProvider(),
	}
	// Return service with all adapters injected
	return application.NewStatusLineService(deps, renderer.NewPowerline())
}
