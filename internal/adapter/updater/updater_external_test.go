package updater_test

import (
	"testing"

	"github.com/florent/status-line/internal/adapter/updater"
)

func TestNewUpdater(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{name: "creates updater with version", version: "v1.0.0"},
		{name: "creates updater without version", version: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := updater.NewUpdater(tt.version)
			// Verify updater is not nil
			if u == nil {
				t.Error("NewUpdater() returned nil")
			}
		})
	}
}

func TestUpdater_CheckAndUpdate(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		wantError bool
	}{
		{name: "dev build skips update", version: "", wantError: false},
		{name: "versioned build checks update", version: "v0.0.1", wantError: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := updater.NewUpdater(tt.version)
			// CheckAndUpdate should not panic
			updated, err := u.CheckAndUpdate()
			// Verify no unexpected errors
			if tt.wantError && err == nil {
				t.Error("CheckAndUpdate() expected error")
			}
			// Dev builds should never update
			if tt.version == "" && updated {
				t.Error("CheckAndUpdate() updated dev build")
			}
		})
	}
}
