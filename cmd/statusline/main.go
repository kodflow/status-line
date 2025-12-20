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
	// Handle version flag before anything else
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "-v" || arg == "--version" {
			printVersion()
			return
		}
	}

	input, err := readInput()
	// Check for input reading errors
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	// Check for updates (returns info about available update)
	updateInfo := checkForUpdate()

	// Generate and output status line with update notification
	svc := buildService(input.WorkingDir())
	fmt.Print(svc.GenerateWithUpdate(input, updateInfo))

	// Download update if available (after output is displayed)
	downloadUpdate(updateInfo)
}

// checkForUpdate checks if an update is available.
// Returns update info for display in status line.
//
// Returns:
//   - model.UpdateInfo: information about available update
func checkForUpdate() model.UpdateInfo {
	u := updater.NewUpdater(version)
	info := u.CheckForUpdate()
	// Convert updater.UpdateInfo to model.UpdateInfo
	return model.UpdateInfo{
		Available: info.Available,
		Version:   info.Version,
	}
}

// downloadUpdate downloads and applies the update if available.
// Silently ignores errors to avoid disrupting normal operation.
//
// Params:
//   - info: update information from checkForUpdate
func downloadUpdate(info model.UpdateInfo) {
	// Skip if no update available
	if !info.Available {
		// No update to download
		return
	}
	u := updater.NewUpdater(version)
	// Ignore errors - update is best-effort
	_ = u.DownloadUpdate(info.Version)
}

// printVersion prints the version information and exits.
// If version is empty (development build), it prints "dev".
func printVersion() {
	v := version
	if v == "" {
		v = "dev"
	}
	fmt.Println("status-line", v)
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
