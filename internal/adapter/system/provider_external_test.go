package system_test

import (
	"testing"

	"github.com/florent/status-line/internal/adapter/system"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates provider"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := system.NewProvider()
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
			p := system.NewProvider()
			info := p.Info()
			_ = info.OS // Just verify no panic
		})
	}
}
