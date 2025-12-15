package model_test

import (
	"testing"

	"github.com/florent/status-line/internal/domain/model"
)

func TestNewProgress(t *testing.T) {
	tests := []struct {
		name        string
		totalTokens int
		contextSize int
		wantPercent int
	}{
		{name: "zero context", totalTokens: 100, contextSize: 0, wantPercent: 0},
		{name: "50 percent", totalTokens: 100, contextSize: 200, wantPercent: 50},
		{name: "100 percent", totalTokens: 200, contextSize: 200, wantPercent: 100},
		{name: "over 100 capped", totalTokens: 300, contextSize: 200, wantPercent: 100},
		{name: "zero tokens", totalTokens: 0, contextSize: 200, wantPercent: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := model.NewProgress(tt.totalTokens, tt.contextSize)
			if p.Percent != tt.wantPercent {
				t.Errorf("NewProgress() Percent = %d, want %d", p.Percent, tt.wantPercent)
			}
		})
	}
}

func TestProgress_Level(t *testing.T) {
	tests := []struct {
		name    string
		percent int
		want    model.ProgressLevel
	}{
		{name: "low 0", percent: 0, want: model.LevelLow},
		{name: "low 49", percent: 49, want: model.LevelLow},
		{name: "medium 50", percent: 50, want: model.LevelMedium},
		{name: "medium 74", percent: 74, want: model.LevelMedium},
		{name: "high 75", percent: 75, want: model.LevelHigh},
		{name: "high 89", percent: 89, want: model.LevelHigh},
		{name: "critical 90", percent: 90, want: model.LevelCritical},
		{name: "critical 100", percent: 100, want: model.LevelCritical},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := model.Progress{Percent: tt.percent}
			if got := p.Level(); got != tt.want {
				t.Errorf("Level() = %v, want %v", got, tt.want)
			}
		})
	}
}
