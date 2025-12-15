// Package mcp provides the MCP configuration adapter.
package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

// Settings file paths.
const (
	// globalSettingsDir is the Claude config directory name.
	globalSettingsDir string = ".claude"
	// settingsFileName is the settings file name.
	settingsFileName string = "settings.json"
)

// Compile-time interface implementation check.
var _ port.MCPProvider = (*Provider)(nil)

// Provider implements port.MCPProvider by reading Claude settings.
// It reads MCP server configurations from settings files.
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
//
// Returns:
//   - model.MCPServers: list of MCP server configurations
func (p *Provider) Servers() model.MCPServers {
	// Read global settings
	globalServers := p.readSettingsFile(p.globalSettingsPath())

	// Read project settings
	projectServers := p.readSettingsFile(p.projectSettingsPath())

	// Preallocate with known capacity
	servers := make(model.MCPServers, 0, len(globalServers)+len(projectServers))
	// Append global servers
	servers = append(servers, globalServers...)
	// Append project servers
	servers = append(servers, projectServers...)

	// Return combined servers
	return servers
}

// globalSettingsPath returns the path to global settings.
//
// Returns:
//   - string: path to global settings file
func (p *Provider) globalSettingsPath() string {
	home, err := os.UserHomeDir()
	// Check if home directory is accessible
	if err != nil {
		// Return empty path if home not found
		return ""
	}
	// Return global settings path
	return filepath.Join(home, globalSettingsDir, settingsFileName)
}

// projectSettingsPath returns the path to project settings.
//
// Returns:
//   - string: path to project settings file
func (p *Provider) projectSettingsPath() string {
	// Check if project directory is set
	if p.projectDir == "" {
		// Return empty path if no project dir
		return ""
	}
	// Return project settings path
	return filepath.Join(p.projectDir, globalSettingsDir, settingsFileName)
}

// readSettingsFile reads MCP servers from a settings file.
//
// Params:
//   - path: path to settings file
//
// Returns:
//   - model.MCPServers: list of MCP servers from file
func (p *Provider) readSettingsFile(path string) model.MCPServers {
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

	var settings settingsFile
	// Check if JSON is valid
	if err := json.Unmarshal(data, &settings); err != nil {
		// Return empty list if parsing fails
		return model.MCPServers{}
	}

	// Preallocate with known capacity
	servers := make(model.MCPServers, 0, len(settings.MCPServers))
	// Convert map to slice
	for name, config := range settings.MCPServers {
		server := model.MCPServer{
			Name:    name,
			Enabled: !config.Disabled,
		}
		// Append server to list
		servers = append(servers, server)
	}

	// Return parsed servers
	return servers
}
