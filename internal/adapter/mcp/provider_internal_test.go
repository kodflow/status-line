package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProvider_userConfigPath(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{projectDir: "/workspace"}
			path := p.userConfigPath()
			if path != "" && !filepath.IsAbs(path) {
				t.Errorf("userConfigPath() = %q, want absolute path", path)
			}
		})
	}
}

func TestProvider_projectConfigPaths(t *testing.T) {
	tests := []struct {
		name       string
		projectDir string
		wantNil    bool
		wantLen    int
	}{
		{name: "with project dir", projectDir: "/workspace", wantNil: false, wantLen: 2},
		{name: "empty project dir", projectDir: "", wantNil: true, wantLen: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{projectDir: tt.projectDir}
			paths := p.projectConfigPaths()
			if tt.wantNil && paths != nil {
				t.Errorf("projectConfigPaths() = %v, want nil", paths)
			}
			if !tt.wantNil && len(paths) != tt.wantLen {
				t.Errorf("projectConfigPaths() len = %d, want %d", len(paths), tt.wantLen)
			}
		})
	}
}

func TestProvider_managedConfigPath(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns platform-specific path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{projectDir: "/workspace"}
			path := p.managedConfigPath()
			// Path can be empty on Windows, otherwise should be absolute
			if path != "" && !filepath.IsAbs(path) {
				t.Errorf("managedConfigPath() = %q, want absolute path", path)
			}
		})
	}
}

func TestProvider_readUserConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns empty for nonexistent file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{projectDir: "/nonexistent/project"}
			servers := p.readUserConfig()
			if servers == nil {
				t.Error("readUserConfig() = nil, want non-nil")
			}
		})
	}
}

func TestProvider_readLocalConfig(t *testing.T) {
	tests := []struct {
		name       string
		projectDir string
	}{
		{name: "empty project dir", projectDir: ""},
		{name: "nonexistent project", projectDir: "/nonexistent/project"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{projectDir: tt.projectDir}
			servers := p.readLocalConfig()
			if servers == nil {
				t.Error("readLocalConfig() = nil, want non-nil")
			}
		})
	}
}

func TestProvider_readProjectConfig(t *testing.T) {
	tests := []struct {
		name       string
		projectDir string
	}{
		{name: "empty project dir", projectDir: ""},
		{name: "nonexistent project", projectDir: "/nonexistent/project"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{projectDir: tt.projectDir}
			servers := p.readProjectConfig()
			if servers == nil {
				t.Error("readProjectConfig() = nil, want non-nil")
			}
		})
	}
}

func TestProvider_readManagedConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns empty for nonexistent file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{projectDir: "/workspace"}
			servers := p.readManagedConfig()
			if servers == nil {
				t.Error("readManagedConfig() = nil, want non-nil")
			}
		})
	}
}

func TestProvider_readProjectConfig_ValidJSON(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantLen int
	}{
		{name: "valid mcp config", content: `{"mcpServers": {"test-server": {"disabled": false}}}`, wantLen: 1},
		{name: "empty mcpServers", content: `{"mcpServers": {}}`, wantLen: 0},
		{name: "disabled server", content: `{"mcpServers": {"test": {"disabled": true}}}`, wantLen: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			mcpPath := filepath.Join(tmpDir, ".mcp.json")
			if err := os.WriteFile(mcpPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}
			p := &Provider{projectDir: tmpDir}
			servers := p.readProjectConfig()
			if len(servers) != tt.wantLen {
				t.Errorf("readProjectConfig() len = %d, want %d", len(servers), tt.wantLen)
			}
		})
	}
}

func TestProvider_readProjectConfig_FallbackUndotted(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		wantLen  int
	}{
		{name: "dotted .mcp.json", fileName: ".mcp.json", wantLen: 1},
		{name: "undotted mcp.json", fileName: "mcp.json", wantLen: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			content := `{"mcpServers": {"test-server": {"disabled": false}}}`
			mcpPath := filepath.Join(tmpDir, tt.fileName)
			if err := os.WriteFile(mcpPath, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}
			p := &Provider{projectDir: tmpDir}
			servers := p.readProjectConfig()
			if len(servers) != tt.wantLen {
				t.Errorf("readProjectConfig() len = %d, want %d", len(servers), tt.wantLen)
			}
		})
	}
}

func TestProvider_convertServers(t *testing.T) {
	tests := []struct {
		name    string
		servers map[string]mcpServerConfig
		wantLen int
	}{
		{name: "nil map", servers: nil, wantLen: 0},
		{name: "empty map", servers: map[string]mcpServerConfig{}, wantLen: 0},
		{name: "one server", servers: map[string]mcpServerConfig{"test": {}}, wantLen: 1},
		{name: "multiple servers", servers: map[string]mcpServerConfig{"a": {}, "b": {}, "c": {}}, wantLen: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{}
			result := p.convertServers(tt.servers)
			if len(result) != tt.wantLen {
				t.Errorf("convertServers() len = %d, want %d", len(result), tt.wantLen)
			}
		})
	}
}

func TestProvider_convertServers_EnabledState(t *testing.T) {
	tests := []struct {
		name        string
		disabled    bool
		wantEnabled bool
	}{
		{name: "enabled server", disabled: false, wantEnabled: true},
		{name: "disabled server", disabled: true, wantEnabled: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{}
			servers := map[string]mcpServerConfig{
				"test": {Disabled: tt.disabled},
			}
			result := p.convertServers(servers)
			if len(result) != 1 {
				t.Fatalf("convertServers() len = %d, want 1", len(result))
			}
			if result[0].Enabled != tt.wantEnabled {
				t.Errorf("server.Enabled = %v, want %v", result[0].Enabled, tt.wantEnabled)
			}
		})
	}
}
