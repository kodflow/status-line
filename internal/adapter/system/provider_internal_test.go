package system

import (
	"testing"

	"github.com/florent/status-line/internal/domain/model"
)

func TestProvider_detectOS(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "detects OS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{}
			osType := p.detectOS()
			if osType < 0 || osType > model.OSUnknown {
				t.Errorf("detectOS() = %d, want valid OSType", osType)
			}
		})
	}
}

func TestProvider_isDocker(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "detects docker"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{}
			_ = p.isDocker() // Just verify no panic
		})
	}
}
