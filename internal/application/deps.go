// Package application contains application services.
package application

import (
	"io"

	"github.com/florent/status-line/internal/domain/port"
)

// ServiceDeps bundles dependencies for StatusLineService.
// It groups providers together to reduce constructor parameters.
type ServiceDeps struct {
	Git      port.GitRepository
	System   port.SystemProvider
	Terminal port.TerminalProvider
	Usage    port.UsageProvider
	Health   port.HealthProvider
	// WorkDir is the directory the session is working in, as inferred from
	// its recent tool calls; empty means the directory Claude Code reported.
	WorkDir string
	// Stderr receives the stack of a provider that panicked; nil means
	// os.Stderr.
	Stderr io.Writer
}
