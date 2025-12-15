package terminal_test

import (
	"testing"

	"github.com/florent/status-line/internal/adapter/terminal"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates provider"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := terminal.NewProvider()
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
			p := terminal.NewProvider()
			info := p.Info()
			if info.Width <= 0 {
				t.Errorf("Info().Width = %d, want > 0", info.Width)
			}
		})
	}
}
