// Package model contains domain entities and value objects.
package model

// MCPServer represents an MCP server configuration.
// It holds the server name and enabled status.
type MCPServer struct {
	Name    string
	Enabled bool
}

// MCPServers is a list of MCP server configurations.
// It represents all configured MCP servers.
type MCPServers []MCPServer
