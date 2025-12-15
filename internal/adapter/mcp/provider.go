// Package mcp provides the MCP configuration adapter.
package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

// Config file names and paths.
const (
	// userConfigFileName is the user-level Claude config file.
	userConfigFileName string = ".claude.json"
	// projectMCPFileName is the project-level MCP config file.
	projectMCPFileName string = ".mcp.json"
	// managedMCPFileName is the enterprise managed MCP config file.
	managedMCPFileName string = "managed-mcp.json"
	// managedPathLinux is the Linux enterprise config directory.
	managedPathLinux string = "/etc/claude-code"
	// managedPathMacOS is the macOS enterprise config directory.
	managedPathMacOS string = "/Library/Application Support/ClaudeCode"
	// defaultMapCapacity is the default capacity for server tracking map.
	defaultMapCapacity int = 16
	// defaultSliceCapacity is the default capacity for server list.
	defaultSliceCapacity int = 8
)

// Compile-time interface implementation check.
var _ port.MCPProvider = (*Provider)(nil)

// Provider implements port.MCPProvider by reading Claude settings.
// It reads MCP server configurations from official Claude Code config files.
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
// Reads from all official config locations and merges results.
// Order: Enterprise → User → Local → Project (following precedence)
//
// Returns:
//   - model.MCPServers: list of MCP server configurations
func (p *Provider) Servers() model.MCPServers {
	// Track unique servers by name (last one wins per precedence)
	seen := make(map[string]bool, defaultMapCapacity)
	servers := make(model.MCPServers, 0, defaultSliceCapacity)

	// Read enterprise managed config (highest precedence)
	enterpriseServers := p.readManagedConfig()
	// Add enterprise servers to results
	for _, s := range enterpriseServers {
		// Mark server as seen and add to list
		seen[s.Name] = true
		servers = append(servers, s)
	}

	// Read user scope from ~/.claude.json mcpServers
	userServers := p.readUserConfig()
	// Add user servers not already seen
	for _, s := range userServers {
		// Skip if server already added from higher precedence
		if !seen[s.Name] {
			seen[s.Name] = true
			servers = append(servers, s)
		}
	}

	// Read local scope from ~/.claude.json projects[path].mcpServers
	localServers := p.readLocalConfig()
	// Add local servers not already seen
	for _, s := range localServers {
		// Skip if server already added from higher precedence
		if !seen[s.Name] {
			seen[s.Name] = true
			servers = append(servers, s)
		}
	}

	// Read project scope from {project}/.mcp.json
	projectServers := p.readProjectConfig()
	// Add project servers not already seen
	for _, s := range projectServers {
		// Skip if server already added from higher precedence
		if !seen[s.Name] {
			seen[s.Name] = true
			servers = append(servers, s)
		}
	}

	// Return combined servers
	return servers
}

// userConfigPath returns the path to user-level Claude config.
//
// Returns:
//   - string: path to ~/.claude.json
func (p *Provider) userConfigPath() string {
	home, err := os.UserHomeDir()
	// Check if home directory is accessible
	if err != nil {
		// Return empty path if home not found
		return ""
	}
	// Return user config path
	return filepath.Join(home, userConfigFileName)
}

// projectConfigPath returns the path to project MCP config.
//
// Returns:
//   - string: path to {project}/.mcp.json
func (p *Provider) projectConfigPath() string {
	// Check if project directory is set
	if p.projectDir == "" {
		// Return empty path if no project dir
		return ""
	}
	// Return project MCP config path
	return filepath.Join(p.projectDir, projectMCPFileName)
}

// managedConfigPath returns the path to enterprise managed MCP config.
//
// Returns:
//   - string: platform-specific path to managed-mcp.json
func (p *Provider) managedConfigPath() string {
	var basePath string
	// Select path based on platform
	switch runtime.GOOS {
	// macOS enterprise path
	case "darwin":
		basePath = managedPathMacOS
	// Linux enterprise path
	case "linux":
		basePath = managedPathLinux
	// Other platforms not supported
	default:
		// Windows uses C:\Program Files\ClaudeCode but skip for now
		return ""
	}
	// Return managed config path
	return filepath.Join(basePath, managedMCPFileName)
}

// readUserConfig reads MCP servers from user-level config.
// Looks for mcpServers at root of ~/.claude.json.
//
// Returns:
//   - model.MCPServers: list of MCP servers from user config
func (p *Provider) readUserConfig() model.MCPServers {
	path := p.userConfigPath()
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

	var config userConfigFile
	// Check if JSON is valid
	if err := json.Unmarshal(data, &config); err != nil {
		// Return empty list if parsing fails
		return model.MCPServers{}
	}

	// Return servers from root mcpServers
	return p.convertServers(config.MCPServers)
}

// readLocalConfig reads MCP servers from local scope config.
// Looks for servers in projects[projectDir].mcpServers of ~/.claude.json.
//
// Returns:
//   - model.MCPServers: list of MCP servers from local config
func (p *Provider) readLocalConfig() model.MCPServers {
	path := p.userConfigPath()
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

	var config userConfigFile
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

	// Return servers from project config
	return p.convertServers(projCfg.MCPServers)
}

// readProjectConfig reads MCP servers from project .mcp.json file.
//
// Returns:
//   - model.MCPServers: list of MCP servers from .mcp.json
func (p *Provider) readProjectConfig() model.MCPServers {
	path := p.projectConfigPath()
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

	var config mcpConfigFile
	// Check if JSON is valid
	if err := json.Unmarshal(data, &config); err != nil {
		// Return empty list if parsing fails
		return model.MCPServers{}
	}

	// Return servers from mcpServers
	return p.convertServers(config.MCPServers)
}

// readManagedConfig reads MCP servers from enterprise managed config.
//
// Returns:
//   - model.MCPServers: list of MCP servers from managed-mcp.json
func (p *Provider) readManagedConfig() model.MCPServers {
	path := p.managedConfigPath()
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

	var config mcpConfigFile
	// Check if JSON is valid
	if err := json.Unmarshal(data, &config); err != nil {
		// Return empty list if parsing fails
		return model.MCPServers{}
	}

	// Return servers from mcpServers
	return p.convertServers(config.MCPServers)
}

// convertServers converts a map of server configs to MCPServers slice.
//
// Params:
//   - servers: map of server name to config
//
// Returns:
//   - model.MCPServers: slice of MCP servers
func (p *Provider) convertServers(servers map[string]mcpServerConfig) model.MCPServers {
	// Check if servers map is empty
	if len(servers) == 0 {
		// Return empty list
		return model.MCPServers{}
	}

	// Preallocate with known capacity
	result := make(model.MCPServers, 0, len(servers))
	// Convert map to slice
	for name, serverConfig := range servers {
		server := model.MCPServer{
			Name:    name,
			Enabled: !serverConfig.Disabled,
		}
		// Append server to result
		result = append(result, server)
	}

	// Return converted servers
	return result
}
