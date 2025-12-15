// Package model contains domain entities and value objects.
package model

// StatusLineData contains all data needed to render the status line.
// It aggregates information from all sources for rendering.
type StatusLineData struct {
	Model       ModelInfo
	Progress    Progress
	Icons       IconConfig
	Git         GitStatus
	System      SystemInfo
	Terminal    TerminalInfo
	Dir         string
	Time        string
	Changes     CodeChanges
	MCP         MCPServers
	Taskwarrior TaskwarriorInfo
}
