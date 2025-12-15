package model_test

import (
	"testing"

	"github.com/florent/status-line/internal/domain/model"
)

func TestInput_ModelInfo(t *testing.T) {
	tests := []struct {
		name        string
		input       model.Input
		wantName    string
		wantVersion string
	}{
		{
			name:        "empty display name uses default",
			input:       model.Input{},
			wantName:    "Claude",
			wantVersion: "",
		},
		{
			name:        "name with version",
			input:       model.Input{Model: model.InputModel{DisplayName: "Opus 4.5"}},
			wantName:    "Opus",
			wantVersion: "4.5",
		},
		{
			name:        "name without version",
			input:       model.Input{Model: model.InputModel{DisplayName: "Sonnet"}},
			wantName:    "Sonnet",
			wantVersion: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := tt.input.ModelInfo()
			if info.Name != tt.wantName || info.Version != tt.wantVersion {
				t.Errorf("ModelInfo() = {%q, %q}, want {%q, %q}", info.Name, info.Version, tt.wantName, tt.wantVersion)
			}
		})
	}
}

func TestInput_WorkingDir(t *testing.T) {
	tests := []struct {
		name  string
		input model.Input
		want  string
	}{
		{name: "empty uses default", input: model.Input{}, want: "~"},
		{name: "custom dir", input: model.Input{Workspace: model.InputWorkspace{CurrentDir: "/workspace"}}, want: "/workspace"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.WorkingDir(); got != tt.want {
				t.Errorf("WorkingDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInput_ContextWindowSize(t *testing.T) {
	tests := []struct {
		name  string
		input model.Input
		want  int
	}{
		{name: "zero uses default", input: model.Input{}, want: 200000},
		{name: "custom size", input: model.Input{ContextWindow: model.InputContext{ContextWindowSize: 100000}}, want: 100000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.ContextWindowSize(); got != tt.want {
				t.Errorf("ContextWindowSize() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestInput_TotalTokens(t *testing.T) {
	tests := []struct {
		name  string
		input model.Input
		want  int
	}{
		{name: "zero tokens", input: model.Input{}, want: 0},
		{name: "sum of tokens", input: model.Input{ContextWindow: model.InputContext{TotalInputTokens: 100, TotalOutputTokens: 50}}, want: 150},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.TotalTokens(); got != tt.want {
				t.Errorf("TotalTokens() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestInput_Progress(t *testing.T) {
	tests := []struct {
		name        string
		input       model.Input
		wantPercent int
	}{
		{name: "zero tokens", input: model.Input{ContextWindow: model.InputContext{ContextWindowSize: 200000}}, wantPercent: 0},
		{name: "50 percent", input: model.Input{ContextWindow: model.InputContext{TotalInputTokens: 100000, ContextWindowSize: 200000}}, wantPercent: 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.input.Progress()
			if p.Percent != tt.wantPercent {
				t.Errorf("Progress() = %d%%, want %d%%", p.Percent, tt.wantPercent)
			}
		})
	}
}

func TestInput_CodeChanges(t *testing.T) {
	tests := []struct {
		name        string
		input       model.Input
		wantAdded   int
		wantRemoved int
	}{
		{name: "zero changes", input: model.Input{}, wantAdded: 0, wantRemoved: 0},
		{name: "with changes", input: model.Input{Cost: model.InputCost{TotalLinesAdded: 10, TotalLinesRemoved: 5}}, wantAdded: 10, wantRemoved: 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.input.CodeChanges()
			if c.Added != tt.wantAdded || c.Removed != tt.wantRemoved {
				t.Errorf("CodeChanges() = {%d, %d}, want {%d, %d}", c.Added, c.Removed, tt.wantAdded, tt.wantRemoved)
			}
		})
	}
}
