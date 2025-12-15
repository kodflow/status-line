// Package application contains application services.
package application

import "github.com/florent/status-line/internal/domain/port"

// ServiceDeps bundles dependencies for StatusLineService.
// It groups providers together to reduce constructor parameters.
type ServiceDeps struct {
	Git         port.GitRepository
	System      port.SystemProvider
	Terminal    port.TerminalProvider
	MCP         port.MCPProvider
	Taskwarrior port.TaskwarriorProvider
	Usage       port.UsageProvider
}
