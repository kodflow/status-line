// Package model contains domain entities and value objects.
package model

import "strings"

// MCPSource names the configuration scope that declared an MCP server.
type MCPSource string

// Configuration scopes, strongest first. The values are the short lowercase
// names the status line prints.
const (
	// MCPSourceUnknown is a server nobody declared: a call to it was seen
	// in the transcript but no configuration names it.
	MCPSourceUnknown MCPSource = ""
	// MCPSourceManaged is the enterprise managed-mcp.json.
	MCPSourceManaged MCPSource = "managed"
	// MCPSourceCLI is the host's --mcp-config command line.
	MCPSourceCLI MCPSource = "cli"
	// MCPSourceLocal is projects[<dir>].mcpServers in the global config.
	MCPSourceLocal MCPSource = "local"
	// MCPSourceProject is the project's .mcp.json.
	MCPSourceProject MCPSource = "project"
	// MCPSourceUser is mcpServers in the global config.
	MCPSourceUser MCPSource = "user"
	// MCPSourcePlugin is an enabled plugin's .mcp.json.
	MCPSourcePlugin MCPSource = "plugin"
)

// MCPSources lists the known scopes in precedence order, strongest first.
var MCPSources = [...]MCPSource{
	MCPSourceManaged, MCPSourceCLI, MCPSourceLocal,
	MCPSourceProject, MCPSourceUser, MCPSourcePlugin,
}

// MCPServer represents an MCP server configuration.
// It holds the server name, its enabled status, for a server a plugin
// provides the plugin's name, the scope that declared it, and whether a
// call to it is under way.
type MCPServer struct {
	Name    string
	Enabled bool
	Plugin  string
	Source  MCPSource
	Busy    bool
}

// MCPServers is a list of MCP server configurations.
// It represents all configured MCP servers.
type MCPServers []MCPServer

// WithSource tags every server with the scope that declared it.
//
// Params:
//   - src: scope to record
//
// Returns:
//   - MCPServers: the same slice, tagged in place
func (s MCPServers) WithSource(src MCPSource) MCPServers {
	// Tag in place: each source slice is built fresh for one call
	for i := range s {
		s[i].Source = src
	}
	return s
}

// pluginToolPrefix opens the tool-name key of a plugin-provided server:
// plugin_<plugin>_<server>.
const pluginToolPrefix string = "plugin_"

// WithBusy marks the servers a call is under way to.
//
// keys are server names as tool names spell them (mcp__<key>__<tool>): the
// server name with every character outside [A-Za-z0-9_-] replaced by "_",
// and plugin_<plugin>_<server> for a plugin's server. A key matching no
// known server is still shown, as an enabled busy server of its own.
//
// Params:
//   - keys: tool-name keys of the servers being called
//
// Returns:
//   - MCPServers: a copy with Busy set, unknown servers appended
func (s MCPServers) WithBusy(keys []string) MCPServers {
	out := make(MCPServers, len(s), len(s)+len(keys))
	copy(out, s)
	// Resolve each key to a known server, or add it
	for _, key := range keys {
		idx := out.indexOfKey(key)
		// An unknown key becomes a server of its own
		if idx < 0 {
			out = append(out, MCPServer{Name: bareName(key), Enabled: true})
			idx = len(out) - 1
		}
		out[idx].Busy = true
	}
	return out
}

// indexOfKey finds the server a tool-name key designates.
//
// Params:
//   - key: tool-name key
//
// Returns:
//   - int: index of the server, -1 when none matches
func (s MCPServers) indexOfKey(key string) int {
	bare := bareName(key)
	fallback := -1
	// An exact match wins over a name recovered from a plugin key
	for i, srv := range s {
		name := ToolKey(srv.Name)
		// The plugin form is exact: plugin and server both match
		if name == key || (srv.Plugin != "" && pluginToolPrefix+ToolKey(srv.Plugin)+"_"+name == key) {
			return i
		}
		// Keep the first bare-name match as the fallback
		if fallback < 0 && (name == bare || srv.Name == bare) {
			fallback = i
		}
	}
	return fallback
}

// bareName strips the plugin part of a plugin key.
//
// Plugin names are kebab-case, so the plugin ends at the first underscore
// after the prefix; a server name may itself hold underscores.
//
// Params:
//   - key: tool-name key
//
// Returns:
//   - string: the server name
func bareName(key string) string {
	rest, isPlugin := strings.CutPrefix(key, pluginToolPrefix)
	// Only plugin keys carry a plugin part
	if !isPlugin {
		return key
	}
	_, server, found := strings.Cut(rest, "_")
	// A malformed plugin key is kept whole
	if !found || server == "" {
		return key
	}
	return server
}

// ToolKey spells a name the way tool names do: every character outside
// [A-Za-z0-9_-] becomes "_".
//
// Params:
//   - name: server or plugin name
//
// Returns:
//   - string: normalised name
func ToolKey(name string) string {
	return strings.Map(func(r rune) rune {
		// Letters, digits, underscore and hyphen survive as they are
		if r == '_' || r == '-' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, name)
}
