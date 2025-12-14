// Package application contains application services.
package application

import (
	"time"

	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

const timeFormat = "15:04:05"

// StatusLineService orchestrates the status line generation.
type StatusLineService struct {
	gitRepo      port.GitRepository
	systemProv   port.SystemProvider
	terminalProv port.TerminalProvider
	renderer     port.Renderer
}

// NewStatusLineService creates a new status line service.
func NewStatusLineService(
	gitRepo port.GitRepository,
	systemProv port.SystemProvider,
	terminalProv port.TerminalProvider,
	renderer port.Renderer,
) *StatusLineService {
	return &StatusLineService{
		gitRepo:      gitRepo,
		systemProv:   systemProv,
		terminalProv: terminalProv,
		renderer:     renderer,
	}
}

// Generate creates the status line string from input.
func (s *StatusLineService) Generate(input *model.Input) string {
	data := model.StatusLineData{
		Model:    input.ModelInfo(),
		Progress: input.Progress(),
		Icons:    model.IconConfigFromEnv(),
		Git:      s.gitRepo.Status(),
		System:   s.systemProv.Info(),
		Terminal: s.terminalProv.Info(),
		Dir:      input.WorkingDir(),
		Time:     time.Now().Format(timeFormat),
	}

	return s.renderer.Render(data)
}
