package renderer

import (
	"testing"

	"github.com/florent/status-line/internal/domain/model"
)

func TestRenderHeavyBar(t *testing.T) {
	tests := []struct {
		name    string
		percent int
		wantLen int
	}{
		{name: "0 percent", percent: 0, wantLen: progressBarWidth},
		{name: "50 percent", percent: 50, wantLen: progressBarWidth},
		{name: "100 percent", percent: 100, wantLen: progressBarWidth},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress := model.Progress{Percent: tt.percent}
			result := renderHeavyBar(progress)
			if len([]rune(result)) != tt.wantLen {
				t.Errorf("renderHeavyBar() len = %d, want %d", len([]rune(result)), tt.wantLen)
			}
		})
	}
}

func TestRenderGranularBar(t *testing.T) {
	tests := []struct {
		name    string
		percent int
		style   ProgressBarStyle
		wantLen int
	}{
		{name: "braille 50%", percent: 50, style: StyleBraille, wantLen: progressBarWidth},
		{name: "block 75%", percent: 75, style: StyleBlock, wantLen: progressBarWidth},
		{name: "braille 0%", percent: 0, style: StyleBraille, wantLen: progressBarWidth},
		{name: "block 100%", percent: 100, style: StyleBlock, wantLen: progressBarWidth},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress := model.Progress{Percent: tt.percent}
			result := renderGranularBar(progress, tt.style)
			if len([]rune(result)) != tt.wantLen {
				t.Errorf("renderGranularBar() len = %d, want %d", len([]rune(result)), tt.wantLen)
			}
		})
	}
}
