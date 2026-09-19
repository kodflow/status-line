package renderer_test

import (
	"testing"

	"github.com/florent/status-line/internal/presentation/renderer"
)

func TestContextGlyph_CoversTheWholeRange(t *testing.T) {
	tests := []struct {
		name    string
		percent int
	}{
		{name: "empty", percent: 0},
		{name: "just under a fifth", percent: 19},
		{name: "a fifth", percent: 20},
		{name: "half", percent: 50},
		{name: "four fifths", percent: 80},
		{name: "full", percent: 100},
		{name: "negative is clamped", percent: -5},
		{name: "over full is clamped", percent: 150},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Every percentage must map to a glyph: an out-of-range index would
			// panic on a status line that redraws constantly
			if got := renderer.ContextGlyph(tt.percent); got == "" {
				t.Errorf("ContextGlyph(%d) returned nothing", tt.percent)
			}
		})
	}
}

func TestContextGlyph_RisesWithConsumption(t *testing.T) {
	var seen []string
	prev := renderer.ContextGlyph(0)
	seen = append(seen, prev)

	// Walk the range and record each change: the glyph must only ever move
	// forward, never flip back to a lower level as the window fills
	for p := 1; p <= 100; p++ {
		cur := renderer.ContextGlyph(p)
		if cur == prev {
			continue
		}
		for _, old := range seen {
			if cur == old {
				t.Fatalf("ContextGlyph(%d) went back to a level already passed", p)
			}
		}
		seen = append(seen, cur)
		prev = cur
	}

	// The five published levels must all be reachable, otherwise the series
	// carries less information than it appears to
	if len(seen) != 5 {
		t.Errorf("walked %d distinct levels across 0-100%%, want 5", len(seen))
	}
}

func TestContextGlyph_ExtremesDiffer(t *testing.T) {
	// An empty window and a full one must not look the same
	if renderer.ContextGlyph(0) == renderer.ContextGlyph(100) {
		t.Error("an empty context renders the same glyph as a full one")
	}
}
