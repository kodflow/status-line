// Package application contains application services.
package application

import (
	"time"

	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

// timeFormat defines the format for displaying time.
const timeFormat string = "15:04:05"

// StatusLineService orchestrates the status line generation.
// It coordinates between adapters and the renderer to produce output.
type StatusLineService struct {
	deps     ServiceDeps
	renderer port.Renderer
}

// NewStatusLineService creates a new status line service.
//
// Params:
//   - deps: bundled provider dependencies
//   - renderer: status line renderer
//
// Returns:
//   - *StatusLineService: configured service instance
func NewStatusLineService(deps ServiceDeps, renderer port.Renderer) *StatusLineService {
	// Return service with all dependencies injected
	return &StatusLineService{
		deps:     deps,
		renderer: renderer,
	}
}

// Generate creates the status line string from input.
//
// Params:
//   - input: input provider for status line data
//
// Returns:
//   - string: formatted status line ready for output
func (s *StatusLineService) Generate(input port.InputProvider) string {
	// Generate without update info
	return s.GenerateWithUpdate(input, model.UpdateInfo{})
}

// GenerateWithUpdate creates the status line string with update notification.
//
// Params:
//   - input: input provider for status line data
//   - update: update information to display
//
// Returns:
//   - string: formatted status line ready for output
func (s *StatusLineService) GenerateWithUpdate(input port.InputProvider, update model.UpdateInfo) string {
	// Fetch usage data (ignore error, use zero value on failure)
	usageData, _ := s.deps.Usage.Usage()

	// Determine progress: prefer session API (real rate limit), fallback to context window
	progress := input.Progress()
	if usageData.Session.IsValid() {
		progress = usageData.Session.Progress()
	}

	// Gather all data from various sources
	data := model.StatusLineData{
		Model:       input.ModelInfo(),
		Progress:    progress,
		Session:     usageData.Session,
		Usage:       usageData.Weekly,
		Icons:       model.IconConfigFromEnv(),
		Git:         s.deps.Git.Status(),
		System:      s.deps.System.Info(),
		Terminal:    s.deps.Terminal.Info(),
		Dir:         input.WorkingDir(),
		Time:        time.Now().Format(timeFormat),
		Changes:     s.deps.Git.DiffStats(),
		MCP:         s.deps.MCP.Servers(),
		Taskwarrior: s.deps.Taskwarrior.Info(),
		Update:      update,
	}

	// Delegate rendering to the renderer
	return s.renderer.Render(data)
}
