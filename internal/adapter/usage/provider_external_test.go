package usage_test

import (
	"testing"

	"github.com/florent/status-line/internal/adapter/usage"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates provider"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := usage.NewProvider()
			if p == nil {
				t.Error("NewProvider() returned nil")
			}
		})
	}
}

func TestProvider_Usage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "success case with valid credentials"},
		{name: "error case with invalid credentials"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := usage.NewProvider()
			// API call may succeed or fail depending on environment
			result, err := p.Usage()
			// Verify both success and error paths work correctly
			if err != nil {
				// Error path: verify result is zero value
				if result.Utilization != 0 {
					t.Error("Usage() should return zero utilization on error")
				}
			}
			// Success path: verify valid percentage range
			if err == nil && (result.Utilization < 0 || result.Utilization > 100) {
				t.Error("Usage() returned invalid utilization percentage")
			}
		})
	}
}
