// Package model contains domain entities and value objects.
package model

// MCPServer represents an MCP server configuration.
// It holds the server name, its enabled status and, for a server a plugin
// provides, the plugin's name.
type MCPServer struct {
	Name    string
	Enabled bool
	Plugin  string
}

// MCPServers is a list of MCP server configurations.
// It represents all configured MCP servers.
type MCPServers []MCPServer
