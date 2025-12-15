package mcp_test

import (
	"testing"

	"github.com/florent/status-line/internal/adapter/mcp"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name       string
		projectDir string
	}{
		{name: "with project dir", projectDir: "/workspace"},
		{name: "empty project dir", projectDir: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := mcp.NewProvider(tt.projectDir)
			if p == nil {
				t.Error("NewProvider() returned nil")
			}
		})
	}
}

func TestProvider_Servers(t *testing.T) {
	tests := []struct {
		name       string
		projectDir string
	}{
		{name: "nonexistent path", projectDir: "/nonexistent"},
		{name: "empty path", projectDir: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := mcp.NewProvider(tt.projectDir)
			servers := p.Servers()
			if servers == nil {
				t.Error("Servers() returned nil, want empty slice")
			}
		})
	}
}
