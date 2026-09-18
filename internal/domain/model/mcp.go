// Package model contains domain entities and value objects.
package model

// MCPServer represents an MCP server configuration.
// It holds the server name, its enabled status, and the config file it was
// declared in, so the status line can point at where to go and change it.
type MCPServer struct {
	Name    string
	Enabled bool
	Source  string
}

// MCPServers is a list of MCP server configurations.
// It represents all configured MCP servers.
type MCPServers []MCPServer
