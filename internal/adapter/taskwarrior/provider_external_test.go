package taskwarrior_test

import (
	"testing"

	"github.com/florent/status-line/internal/adapter/taskwarrior"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates provider"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := taskwarrior.NewProvider()
			if p == nil {
				t.Error("NewProvider() returned nil")
			}
		})
	}
}

func TestProvider_Info(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns info"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := taskwarrior.NewProvider()
			info := p.Info()
			_ = info.HasProjects() // Just verify no panic
		})
	}
}
