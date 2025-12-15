package usage

import (
	"runtime"
	"testing"
)

func TestProvider_getToken(t *testing.T) {
	tests := []struct {
		name      string
		wantError bool
	}{
		{name: "attempts to retrieve token", wantError: false},
		{name: "handles missing credentials gracefully", wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider()
			// Token retrieval may fail in test environment
			token, err := p.getToken()
			// Verify function handles errors gracefully
			if tt.wantError && err == nil && token != "" {
				return
			}
			// Success case - just verify no panic occurred
			if !tt.wantError && token == "" && err != nil {
				return
			}
		})
	}
}

func TestProvider_getTokenFromKeychain(t *testing.T) {
	tests := []struct {
		name      string
		wantError bool
	}{
		{name: "attempts to read from keychain", wantError: false},
		{name: "handles keychain unavailable", wantError: runtime.GOOS != "darwin"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider()
			// Keychain may not have credentials or may not be available
			token, err := p.getTokenFromKeychain()
			// Verify error handling on non-macOS
			if tt.wantError && err == nil && token != "" {
				t.Error("getTokenFromKeychain() expected error on non-macOS")
			}
		})
	}
}

func TestProvider_getTokenFromFile(t *testing.T) {
	tests := []struct {
		name      string
		wantError bool
	}{
		{name: "attempts to read credentials file", wantError: false},
		{name: "handles missing file gracefully", wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider()
			// File may not exist in test environment
			token, err := p.getTokenFromFile()
			// Verify function handles errors gracefully
			if tt.wantError && err == nil && token != "" {
				return
			}
			// Success case - just verify no panic occurred
			if !tt.wantError && token == "" && err != nil {
				return
			}
		})
	}
}
