package renderer_test

import (
	"testing"

	"github.com/florent/status-line/internal/presentation/renderer"
)

func TestGetModelColors(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		wantBg    string
		wantFg    string
		wantText  string
	}{
		{name: "haiku model", modelName: "Haiku 3.5", wantBg: renderer.BgHaiku, wantFg: renderer.FgHaiku, wantText: renderer.FgHaikuDark},
		{name: "sonnet model", modelName: "Sonnet 3.5", wantBg: renderer.BgSonnet, wantFg: renderer.FgSonnet, wantText: renderer.FgSonnetDark},
		{name: "opus model", modelName: "Opus 4.5", wantBg: renderer.BgOpus, wantFg: renderer.FgOpus, wantText: renderer.FgOpusDark},
		{name: "unknown model", modelName: "Unknown", wantBg: renderer.BgWhite, wantFg: renderer.FgWhite, wantText: renderer.FgBlack},
		{name: "case insensitive", modelName: "OPUS", wantBg: renderer.BgOpus, wantFg: renderer.FgOpus, wantText: renderer.FgOpusDark},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bg, fg, text := renderer.GetModelColors(tt.modelName)
			if bg != tt.wantBg || fg != tt.wantFg || text != tt.wantText {
				t.Errorf("GetModelColors(%q) = (%q, %q, %q), want (%q, %q, %q)", tt.modelName, bg, fg, text, tt.wantBg, tt.wantFg, tt.wantText)
			}
		})
	}
}
