package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProvider_claudeConfigPath(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{projectDir: "/workspace"}
			path := p.claudeConfigPath()
			if path != "" && !filepath.IsAbs(path) {
				t.Errorf("claudeConfigPath() = %q, want absolute path", path)
			}
		})
	}
}

func TestProvider_mcpConfigPath(t *testing.T) {
	tests := []struct {
		name       string
		projectDir string
		wantEmpty  bool
	}{
		{name: "with project dir", projectDir: "/workspace", wantEmpty: false},
		{name: "empty project dir", projectDir: "", wantEmpty: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{projectDir: tt.projectDir}
			path := p.mcpConfigPath()
			if tt.wantEmpty && path != "" {
				t.Errorf("mcpConfigPath() = %q, want empty", path)
			}
			if !tt.wantEmpty && path == "" {
				t.Error("mcpConfigPath() = empty, want non-empty")
			}
		})
	}
}

func TestProvider_readClaudeConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns empty for nonexistent file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{projectDir: "/nonexistent/project"}
			servers := p.readClaudeConfig()
			if servers == nil {
				t.Error("readClaudeConfig() = nil, want non-nil")
			}
		})
	}
}

func TestProvider_readMCPConfig(t *testing.T) {
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
			servers := p.readMCPConfig()
			if servers == nil {
				t.Error("readMCPConfig() = nil, want non-nil")
			}
		})
	}
}

func TestProvider_readMCPConfig_ValidJSON(t *testing.T) {
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
			servers := p.readMCPConfig()
			if len(servers) != tt.wantLen {
				t.Errorf("readMCPConfig() len = %d, want %d", len(servers), tt.wantLen)
			}
		})
	}
}
