package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProvider_globalSettingsPath(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{projectDir: "/workspace"}
			path := p.globalSettingsPath()
			if path != "" && !filepath.IsAbs(path) {
				t.Errorf("globalSettingsPath() = %q, want absolute path", path)
			}
		})
	}
}

func TestProvider_projectSettingsPath(t *testing.T) {
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
			path := p.projectSettingsPath()
			if tt.wantEmpty && path != "" {
				t.Errorf("projectSettingsPath() = %q, want empty", path)
			}
			if !tt.wantEmpty && path == "" {
				t.Error("projectSettingsPath() = empty, want non-empty")
			}
		})
	}
}

func TestProvider_readSettingsFile(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "empty path", path: ""},
		{name: "nonexistent file", path: "/nonexistent/path/settings.json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{}
			servers := p.readSettingsFile(tt.path)
			if servers == nil {
				t.Error("readSettingsFile() = nil, want non-nil")
			}
		})
	}
}

func TestProvider_readSettingsFile_ValidJSON(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantLen int
	}{
		{name: "valid settings", content: `{"mcpServers": {"test-server": {"disabled": false}}}`, wantLen: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			settingsPath := filepath.Join(tmpDir, "settings.json")
			if err := os.WriteFile(settingsPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}
			p := &Provider{}
			servers := p.readSettingsFile(settingsPath)
			if len(servers) != tt.wantLen {
				t.Errorf("readSettingsFile() len = %d, want %d", len(servers), tt.wantLen)
			}
		})
	}
}
