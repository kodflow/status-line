// Package mcp provides the MCP configuration adapter.
package mcp

// mcpServerConfig represents a single MCP server config.
// It contains the server type, command, and disabled status.
type mcpServerConfig struct {
	Type     string `json:"type"`
	Command  string `json:"command"`
	URL      string `json:"url"`
	Disabled bool   `json:"disabled"`
}
