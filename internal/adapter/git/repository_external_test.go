package git_test

import (
	"testing"

	"github.com/florent/status-line/internal/adapter/git"
)

func TestNewRepository(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates repository"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := git.NewRepository()
			if r == nil {
				t.Error("NewRepository() returned nil")
			}
		})
	}
}

func TestRepository_Status(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns status"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := git.NewRepository()
			status := r.Status()
			_ = status.IsInRepo() // Just verify no panic
		})
	}
}

func TestRepository_DiffStats(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns diff stats"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := git.NewRepository()
			changes := r.DiffStats()
			if changes.Added < 0 || changes.Removed < 0 {
				t.Errorf("DiffStats() = {%d, %d}, want non-negative", changes.Added, changes.Removed)
			}
		})
	}
}
