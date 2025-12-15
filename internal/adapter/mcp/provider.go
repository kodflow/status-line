// Package mcp provides the MCP configuration adapter.
package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

// Config file paths and names.
const (
	// claudeDir is the Claude config directory name.
	claudeDir string = ".claude"
	// claudeConfigFileName is the main Claude config file name.
	claudeConfigFileName string = ".claude.json"
	// mcpConfigFileName is the project MCP config file name.
	mcpConfigFileName string = ".mcp.json"
)

// Compile-time interface implementation check.
var _ port.MCPProvider = (*Provider)(nil)

// Provider implements port.MCPProvider by reading Claude settings.
// It reads MCP server configurations from Claude config files.
type Provider struct {
	projectDir string
}

// NewProvider creates a new MCP provider adapter.
//
// Params:
//   - projectDir: the project directory path
//
// Returns:
//   - *Provider: new provider instance
func NewProvider(projectDir string) *Provider {
	// Return provider with project directory
	return &Provider{projectDir: projectDir}
}

// Servers returns the list of configured MCP servers.
// Reads from multiple config locations and merges results.
//
// Returns:
//   - model.MCPServers: list of MCP server configurations
func (p *Provider) Servers() model.MCPServers {
	// Read from ~/.claude/.claude.json (project-specific config)
	claudeServers := p.readClaudeConfig()

	// Read from {project}/.mcp.json
	mcpServers := p.readMCPConfig()

	// Preallocate with known capacity
	servers := make(model.MCPServers, 0, len(claudeServers)+len(mcpServers))
	// Append Claude config servers
	servers = append(servers, claudeServers...)
	// Append MCP config servers
	servers = append(servers, mcpServers...)

	// Return combined servers
	return servers
}

// claudeConfigPath returns the path to Claude config file.
//
// Returns:
//   - string: path to ~/.claude/.claude.json
func (p *Provider) claudeConfigPath() string {
	home, err := os.UserHomeDir()
	// Check if home directory is accessible
	if err != nil {
		// Return empty path if home not found
		return ""
	}
	// Return Claude config path
	return filepath.Join(home, claudeDir, claudeConfigFileName)
}

// mcpConfigPath returns the path to project MCP config.
//
// Returns:
//   - string: path to {project}/.mcp.json
func (p *Provider) mcpConfigPath() string {
	// Check if project directory is set
	if p.projectDir == "" {
		// Return empty path if no project dir
		return ""
	}
	// Return MCP config path
	return filepath.Join(p.projectDir, mcpConfigFileName)
}

// readClaudeConfig reads MCP servers from Claude config file.
// Looks for servers in projects[projectDir].mcpServers.
//
// Returns:
//   - model.MCPServers: list of MCP servers from Claude config
func (p *Provider) readClaudeConfig() model.MCPServers {
	path := p.claudeConfigPath()
	// Check if path is provided
	if path == "" {
		// Return empty list for empty path
		return model.MCPServers{}
	}

	data, err := os.ReadFile(path)
	// Check if file is readable
	if err != nil {
		// Return empty list if file not accessible
		return model.MCPServers{}
	}

	var config claudeConfigFile
	// Check if JSON is valid
	if err := json.Unmarshal(data, &config); err != nil {
		// Return empty list if parsing fails
		return model.MCPServers{}
	}

	// Look for project-specific config
	projCfg, exists := config.Projects[p.projectDir]
	// Check if project exists in config
	if !exists {
		// Return empty list if project not found
		return model.MCPServers{}
	}

	// Preallocate with known capacity
	servers := make(model.MCPServers, 0, len(projCfg.MCPServers))
	// Convert map to slice
	for name, serverConfig := range projCfg.MCPServers {
		server := model.MCPServer{
			Name:    name,
			Enabled: !serverConfig.Disabled,
		}
		// Append server to list
		servers = append(servers, server)
	}

	// Return parsed servers
	return servers
}

// readMCPConfig reads MCP servers from project .mcp.json file.
//
// Returns:
//   - model.MCPServers: list of MCP servers from .mcp.json
func (p *Provider) readMCPConfig() model.MCPServers {
	path := p.mcpConfigPath()
	// Check if path is provided
	if path == "" {
		// Return empty list for empty path
		return model.MCPServers{}
	}

	data, err := os.ReadFile(path)
	// Check if file is readable
	if err != nil {
		// Return empty list if file not accessible
		return model.MCPServers{}
	}

	var config mcpJsonFile
	// Check if JSON is valid
	if err := json.Unmarshal(data, &config); err != nil {
		// Return empty list if parsing fails
		return model.MCPServers{}
	}

	// Preallocate with known capacity
	servers := make(model.MCPServers, 0, len(config.MCPServers))
	// Convert map to slice
	for name, serverConfig := range config.MCPServers {
		server := model.MCPServer{
			Name:    name,
			Enabled: !serverConfig.Disabled,
		}
		// Append server to list
		servers = append(servers, server)
	}

	// Return parsed servers
	return servers
}
