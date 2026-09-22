// Package main provides the entry point for the status-line CLI tool.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/florent/status-line/internal/adapter/activity"
	"github.com/florent/status-line/internal/adapter/git"
	"github.com/florent/status-line/internal/adapter/health"
	"github.com/florent/status-line/internal/adapter/mcp"
	"github.com/florent/status-line/internal/adapter/sessionstate"
	"github.com/florent/status-line/internal/adapter/system"
	"github.com/florent/status-line/internal/adapter/tasks"
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
	// Handle flags before reading stdin
	if len(os.Args) > 1 {
		arg := os.Args[1]
		// Print the version and exit
		if arg == "-v" || arg == "--version" {
			printVersion()
			return
		}
		// Refresh the usage cache out of band and exit
		if arg == "--refresh-usage" {
			// A refresh failure only leaves the cache stale
			_ = usage.NewProvider().Refresh()
			return
		}
		// Refresh the service health cache out of band and exit
		if arg == health.RefreshFlag {
			// A refresh failure only leaves the cache stale
			_ = health.NewProvider().Refresh()
			return
		}
	}

	// A malformed payload must still produce a status line: this process is
	// the shell prompt of a running session, and exiting non-zero replaces it
	// with a raw error on every single redraw
	input := readInput()

	// Check for updates (returns info about available update)
	updateInfo := checkForUpdate()

	// Generate and output status line with update notification
	svc := buildService(input)
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
// An unreadable or malformed payload yields a zero input rather than an error:
// the renderer degrades to what it can still establish on its own, which is
// strictly better than leaving the session with no status line at all.
//
// Returns:
//   - *model.Input: parsed input data, zero valued on failure
func readInput() *model.Input {
	var input model.Input

	data, err := io.ReadAll(os.Stdin)
	// An unreadable stdin leaves the zero input in place
	if err != nil {
		return &input
	}
	// A malformed payload leaves the zero input in place
	if err := json.Unmarshal(data, &input); err != nil {
		return &input
	}
	return &input
}

// buildService creates and wires all dependencies for the status line service.
//
// Params:
//   - input: parsed stdin payload
//
// Returns:
//   - *application.StatusLineService: fully configured service instance
func buildService(input *model.Input) *application.StatusLineService {
	// MCP servers are configured for the session's own directory, while git
	// follows wherever the session is actually working
	sessionDir := input.WorkingDir()
	workDir := activity.Dir(input.TranscriptPath(), sessionDir)
	// Without a real directory git keeps the process working directory
	gitDir := workDir
	if !filepath.IsAbs(gitDir) {
		gitDir = ""
	}
	// The session registry names the host process, whose command line may
	// carry MCP servers; one scan serves both adapters
	session := sessionstate.NewProvider(input.Session())
	deps := application.ServiceDeps{
		Git:      git.NewRepository(gitDir),
		System:   system.NewProvider(),
		Terminal: terminal.NewProvider(),
		MCP:      mcp.NewProvider(sessionDir, session.PID),
		Usage:    usage.NewProvider(),
		Health:   health.NewProvider(),
		Tasks:    tasks.NewProvider(input.Session()),
		Activity: session,
		WorkDir:  workDir,
	}
	// Return service with all adapters injected
	return application.NewStatusLineService(deps, renderer.NewPowerline())
}
