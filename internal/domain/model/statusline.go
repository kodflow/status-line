// Package model contains domain entities and value objects.
package model

// StatusLineData contains all data needed to render the status line.
// It aggregates information from all sources for rendering.
type StatusLineData struct {
	Model         ModelInfo
	Progress      Progress
	Limits        LimitSet
	Session       Limit
	Usage         Limit
	Icons         IconConfig
	Git           GitStatus
	System        SystemInfo
	Terminal      TerminalInfo
	Dir           string
	Time          string
	ContextTokens int
	ContextSize   int
	Cost          float64
	Effort        string
	FastMode      bool
	SessionName   string
	Changes       CodeChanges
	MCP           MCPServers
	Update        UpdateInfo
}

// UpdateInfo contains information about available updates.
// Used to display update notification in the status line.
type UpdateInfo struct {
	Available bool
	Version   string
}
