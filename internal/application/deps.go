// Package application contains application services.
package application

import "github.com/florent/status-line/internal/domain/port"

// ServiceDeps bundles dependencies for StatusLineService.
// It groups providers together to reduce constructor parameters.
type ServiceDeps struct {
	Git      port.GitRepository
	System   port.SystemProvider
	Terminal port.TerminalProvider
	MCP      port.MCPProvider
	Usage    port.UsageProvider
	Health   port.HealthProvider
	Tasks    port.TasksProvider
	// WorkDir is the directory the session is working in, as inferred from
	// its recent tool calls; empty means the directory Claude Code reported.
	WorkDir string
}
