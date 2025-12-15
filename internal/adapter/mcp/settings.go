// Package mcp provides the MCP configuration adapter.
package mcp

// userConfigFile represents the ~/.claude.json structure.
// Contains user-level MCP servers and project-specific configs.
type userConfigFile struct {
	// MCPServers at root level (user scope)
	MCPServers map[string]mcpServerConfig `json:"mcpServers"`
	// Projects contains project-specific configurations (local scope)
	Projects map[string]projectConfig `json:"projects"`
}

// projectConfig represents a project's configuration in ~/.claude.json.
// It contains MCP server configurations for a specific project path.
type projectConfig struct {
	MCPServers map[string]mcpServerConfig `json:"mcpServers"`
}

// mcpConfigFile represents the .mcp.json or managed-mcp.json structure.
// Used for project-level and enterprise managed MCP configurations.
type mcpConfigFile struct {
	MCPServers map[string]mcpServerConfig `json:"mcpServers"`
}

// mcpServerConfig represents a single MCP server configuration.
type mcpServerConfig struct {
	// Type is the server type (e.g., "stdio", "sse")
	Type string `json:"type,omitempty"`
	// Command is the server command (e.g., "npx", "uvx")
	Command string `json:"command,omitempty"`
	// Args are the command arguments
	Args []string `json:"args,omitempty"`
	// URL is the server URL (for SSE type servers)
	URL string `json:"url,omitempty"`
	// Env contains environment variables
	Env map[string]string `json:"env,omitempty"`
	// Disabled indicates if the server is disabled
	Disabled bool `json:"disabled,omitempty"`
}
