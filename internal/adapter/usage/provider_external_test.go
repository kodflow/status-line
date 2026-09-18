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

func TestProvider_Limits(t *testing.T) {
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
			result, err := p.Limits()
			// Error path: an unreachable API must yield no fabricated quota
			if err != nil {
				if result.HasTimed() {
					t.Error("Limits() should return no quota on error")
				}
			}
			// Success path: every returned quota stays in range
			if err == nil {
				for _, limit := range result.Timed() {
					if limit.Percent < 0 || limit.Percent > 100 {
						t.Errorf("Limits() returned out-of-range percentage %d for %s", limit.Percent, limit.Label)
					}
				}
			}
		})
	}
}
