// Package mcp provides the MCP configuration adapter.
package mcp

// claudeConfigFile represents the ~/.claude/.claude.json structure.
// It contains project-specific MCP server configurations.
type claudeConfigFile struct {
	Projects map[string]projectConfig `json:"projects"`
}

// projectConfig represents a project's configuration in .claude.json.
// It contains MCP server configurations for a specific project.
type projectConfig struct {
	MCPServers map[string]mcpServerConfig `json:"mcpServers"`
}

// mcpJsonFile represents the {project}/.mcp.json structure.
// It contains MCP server configurations at project level.
type mcpJsonFile struct {
	MCPServers map[string]mcpServerConfig `json:"mcpServers"`
}
