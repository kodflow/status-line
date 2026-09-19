package renderer_test

import (
	"testing"

	"github.com/florent/status-line/internal/presentation/renderer"
)

func TestContextHeat_RisesWithConsumption(t *testing.T) {
	tests := []struct {
		name    string
		percent int
		want    string
	}{
		{name: "empty", percent: 0, want: renderer.FgHeatCalm},
		{name: "just under warm", percent: 39, want: renderer.FgHeatCalm},
		{name: "warm", percent: 40, want: renderer.FgHeatWarm},
		{name: "just under hot", percent: 59, want: renderer.FgHeatWarm},
		{name: "hot", percent: 60, want: renderer.FgHeatHot},
		{name: "just under compaction", percent: 89, want: renderer.FgHeatHot},
		{name: "at the compaction threshold", percent: 90, want: renderer.FgHeatCritical},
		{name: "full", percent: 100, want: renderer.FgHeatCritical},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := renderer.ContextHeat(tt.percent); got != tt.want {
				t.Errorf("ContextHeat(%d) = %q, want %q", tt.percent, got, tt.want)
			}
		})
	}
}

func TestContextHeat_NeverCools(t *testing.T) {
	rank := map[string]int{
		renderer.FgHeatCalm:     0,
		renderer.FgHeatWarm:     1,
		renderer.FgHeatHot:      2,
		renderer.FgHeatCritical: 3,
	}
	prev := 0
	// Walking the range, the colour must only ever climb: a step backwards
	// would read as the context emptying while it fills
	for p := 0; p <= 100; p++ {
		cur, ok := rank[renderer.ContextHeat(p)]
		if !ok {
			t.Fatalf("ContextHeat(%d) returned an ink outside the scale", p)
		}
		if cur < prev {
			t.Fatalf("ContextHeat(%d) cooled down from the previous step", p)
		}
		prev = cur
	}
	// And the scale must actually be traversed end to end
	if prev != 3 {
		t.Errorf("the scale topped out at step %d, want the critical step", prev)
	}
}

func TestContextHeat_DistinctInks(t *testing.T) {
	inks := []string{renderer.FgHeatCalm, renderer.FgHeatWarm, renderer.FgHeatHot, renderer.FgHeatCritical}
	seen := make(map[string]bool, len(inks))
	for _, ink := range inks {
		// Two steps sharing an ink would make one of them invisible
		if seen[ink] {
			t.Errorf("ink %q is used by two different steps", ink)
		}
		seen[ink] = true
	}
}
