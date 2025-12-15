// Package mcp provides the MCP configuration adapter.
package mcp

// settingsFile represents the Claude settings.json structure.
// It contains MCP server configurations.
type settingsFile struct {
	MCPServers map[string]mcpServerConfig `json:"mcpServers"`
}
