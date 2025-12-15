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
	gitRepo         port.GitRepository
	systemProv      port.SystemProvider
	terminalProv    port.TerminalProvider
	mcpProv         port.MCPProvider
	taskwarriorProv port.TaskwarriorProvider
	renderer        port.Renderer
}

// NewStatusLineService creates a new status line service.
//
// Params:
//   - gitRepo: git repository adapter
//   - systemProv: system information provider
//   - terminalProv: terminal information provider
//   - mcpProv: MCP configuration provider
//   - taskwarriorProv: Taskwarrior provider
//   - renderer: status line renderer
//
// Returns:
//   - *StatusLineService: configured service instance
func NewStatusLineService(
	gitRepo port.GitRepository,
	systemProv port.SystemProvider,
	terminalProv port.TerminalProvider,
	mcpProv port.MCPProvider,
	taskwarriorProv port.TaskwarriorProvider,
	renderer port.Renderer,
) *StatusLineService {
	// Return service with all dependencies injected
	return &StatusLineService{
		gitRepo:         gitRepo,
		systemProv:      systemProv,
		terminalProv:    terminalProv,
		mcpProv:         mcpProv,
		taskwarriorProv: taskwarriorProv,
		renderer:        renderer,
	}
}

// Generate creates the status line string from input.
//
// Params:
//   - input: parsed JSON input from Claude Code
//
// Returns:
//   - string: formatted status line ready for output
func (s *StatusLineService) Generate(input *model.Input) string {
	// Gather all data from various sources
	data := model.StatusLineData{
		Model:       input.ModelInfo(),
		Progress:    input.Progress(),
		Icons:       model.IconConfigFromEnv(),
		Git:         s.gitRepo.Status(),
		System:      s.systemProv.Info(),
		Terminal:    s.terminalProv.Info(),
		Dir:         input.WorkingDir(),
		Time:        time.Now().Format(timeFormat),
		Changes:     s.gitRepo.DiffStats(),
		MCP:         s.mcpProv.Servers(),
		Taskwarrior: s.taskwarriorProv.Info(),
	}

	// Delegate rendering to the renderer
	return s.renderer.Render(data)
}
