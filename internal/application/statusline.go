// Package application contains application services.
package application

import (
	"fmt"
	"os"
	"runtime/debug"
	"sync"
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
	// Every provider shells out to something: git, the keychain,
	// the filesystem. Run in sequence their latencies add up on a process that
	// is re-executed on every redraw, so gather them concurrently instead and
	// pay only the slowest one.
	var (
		wg          sync.WaitGroup
		apiLimits   model.LimitSet
		gitStatus   model.GitStatus
		gitChanges  model.CodeChanges
		systemInfo  model.SystemInfo
		terminalNfo model.TerminalInfo
		health      model.ServiceHealth
	)

	stderr := s.deps.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}
	gather := func(fn func()) {
		wg.Go(func() {
			// A provider that panics on its own goroutine would take the
			// whole process down before the frame is written: its value
			// stays zero and the rest of the bar still renders
			defer func() {
				// Report the panic and let the other providers finish
				if r := recover(); r != nil {
					fmt.Fprintf(stderr, "status-line: provider panic: %v\n%s", r, debug.Stack())
				}
			}()
			fn()
		})
	}

	// An unreachable API only costs the enrichment it would have added
	gather(func() { apiLimits, _ = s.deps.Usage.Limits() })
	gather(func() { gitStatus = s.deps.Git.Status() })
	gather(func() { gitChanges = s.deps.Git.DiffStats() })
	gather(func() { systemInfo = s.deps.System.Info() })
	gather(func() { terminalNfo = s.deps.Terminal.Info() })
	// Service health is optional: without a provider nothing is drawn
	if s.deps.Health != nil {
		gather(func() { health = s.deps.Health.Health() })
	}
	wg.Wait()

	// Show where the session is working rather than where it was started
	dir := s.deps.WorkDir
	if dir == "" {
		dir = input.WorkingDir()
	}

	// Merge stdin and API quotas without ever letting a rate limit overwrite
	// the context window: they measure different things and both must stay
	// readable at a glance
	limits := resolveLimits(input.StdinLimits(), apiLimits)

	// Gather all data from various sources
	data := model.StatusLineData{
		Model:       input.ModelInfo(),
		Progress:    limits.Context.Progress(),
		Limits:      limits,
		Session:     limits.Session,
		Usage:       limits.Weekly,
		Icons:       model.IconConfigFromEnv(),
		Git:         gitStatus,
		System:      systemInfo,
		Terminal:    terminalNfo,
		Dir:         dir,
		Time:        time.Now().Format(timeFormat),
		Changes:     gitChanges,
		Update:      update,
		Effort:      input.EffortLevel(),
		Cost:        input.SessionCost(),
		FastMode:    input.IsFastMode(),
		SessionName: input.SessionLabel(),
		Health:      health,
	}

	// Delegate rendering to the renderer
	return s.renderer.Render(data)
}
