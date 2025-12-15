package renderer_test

import (
	"testing"

	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/presentation/renderer"
)

func TestRenderProgressBar(t *testing.T) {
	tests := []struct {
		name    string
		percent int
		style   renderer.ProgressBarStyle
		wantLen int
	}{
		{name: "heavy 0%", percent: 0, style: renderer.StyleHeavy, wantLen: 10},
		{name: "heavy 50%", percent: 50, style: renderer.StyleHeavy, wantLen: 10},
		{name: "heavy 100%", percent: 100, style: renderer.StyleHeavy, wantLen: 10},
		{name: "block 50%", percent: 50, style: renderer.StyleBlock, wantLen: 10},
		{name: "braille 75%", percent: 75, style: renderer.StyleBraille, wantLen: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress := model.Progress{Percent: tt.percent}
			result := renderer.RenderProgressBar(progress, tt.style)
			if len([]rune(result)) != tt.wantLen {
				t.Errorf("RenderProgressBar() len = %d, want %d", len([]rune(result)), tt.wantLen)
			}
		})
	}
}
