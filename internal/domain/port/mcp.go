// Package port defines domain interfaces (contracts).
package port

import "github.com/florent/status-line/internal/domain/model"

// MCPProvider defines the interface for reading MCP server configurations.
// Implementations should read from Claude settings files.
type MCPProvider interface {
	// Servers returns the list of configured MCP servers.
	//
	// Returns:
	//   - model.MCPServers: list of MCP server configurations
	Servers() model.MCPServers
}

// MCPCallsProvider reads which MCP servers are being called right now.
type MCPCallsProvider interface {
	// Busy returns the tool-name keys (mcp__<key>__<tool>) of the servers
	// with a call in flight or one that ended a moment ago.
	//
	// Returns:
	//   - []string: distinct keys, sorted
	Busy() []string
}
