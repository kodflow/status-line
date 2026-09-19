package renderer

import (
	"fmt"
	"math"
	"testing"
)

func TestContextHeat_RisesWithConsumption(t *testing.T) {
	tests := []struct {
		name    string
		percent int
		want    string
	}{
		{name: "empty", percent: 0, want: FgHeatCalm},
		{name: "just under warm", percent: 39, want: FgHeatCalm},
		{name: "warm", percent: 40, want: FgHeatWarm},
		{name: "just under hot", percent: 59, want: FgHeatWarm},
		{name: "hot", percent: 60, want: FgHeatHot},
		{name: "just under compaction", percent: 89, want: FgHeatHot},
		{name: "at the compaction threshold", percent: 90, want: FgHeatCritical},
		{name: "full", percent: 100, want: FgHeatCritical},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContextHeat(tt.percent); got != tt.want {
				t.Errorf("ContextHeat(%d) = %q, want %q", tt.percent, got, tt.want)
			}
		})
	}
}

func TestContextHeat_NeverCools(t *testing.T) {
	rank := map[string]int{
		FgHeatCalm:     0,
		FgHeatWarm:     1,
		FgHeatHot:      2,
		FgHeatCritical: 3,
	}
	prev := 0
	// Walking the range, the colour must only ever climb: a step backwards
	// would read as the context emptying while it fills
	for p := 0; p <= 100; p++ {
		cur, ok := rank[ContextHeat(p)]
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
	inks := []string{FgHeatCalm, FgHeatWarm, FgHeatHot, FgHeatCritical}
	seen := make(map[string]bool, len(inks))
	for _, ink := range inks {
		// Two steps sharing an ink would make one of them invisible
		if seen[ink] {
			t.Errorf("ink %q is used by two different steps", ink)
		}
		seen[ink] = true
	}
}

// ansiToRGB resolves an xterm-256 colour escape to its RGB triple, foreground
// (38) or background (48) alike.
func ansiToRGB(t *testing.T, esc string) (r, g, b float64) {
	t.Helper()
	var kind, n int
	if _, err := fmt.Sscanf(esc, "\033[%d;5;%dm", &kind, &n); err != nil {
		t.Fatalf("cannot parse %q", esc)
	}
	if n >= 232 {
		v := float64(8 + (n-232)*10)
		return v, v, v
	}
	n -= 16
	lv := []float64{0, 95, 135, 175, 215, 255}
	return lv[n/36], lv[(n/6)%6], lv[n%6]
}

func relLum(r, g, b float64) float64 {
	conv := func(v float64) float64 {
		v /= 255
		if v <= 0.03928 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	return 0.2126*conv(r) + 0.7152*conv(g) + 0.0722*conv(b)
}

func TestContextHeat_MeetsContrastOnItsOwnGround(t *testing.T) {
	// The context segment is drawn on BgContext; an ink is only a warning if
	// it can be read against it. 4.5:1 is the WCAG AA floor for body text.
	const minRatio = 4.5
	bgR, bgG, bgB := ansiToRGB(t, BgContext)

	tests := []struct {
		name string
		ink  string
	}{
		{name: "calm", ink: FgHeatCalm},
		{name: "warm", ink: FgHeatWarm},
		{name: "hot", ink: FgHeatHot},
		{name: "critical", ink: FgHeatCritical},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr, fg, fb := ansiToRGB(t, tt.ink)
			lf, lb := relLum(fr, fg, fb), relLum(bgR, bgG, bgB)
			hi, lo := math.Max(lf, lb), math.Min(lf, lb)
			ratio := (hi + 0.05) / (lo + 0.05)
			if ratio < minRatio {
				t.Errorf("contrast %.2f:1 is below the %.1f:1 floor", ratio, minRatio)
			}
		})
	}
}

// contrast returns the WCAG ratio between two xterm-256 escapes.
//
// Params:
//   - fg: foreground escape
//   - bg: background escape
//
// Returns:
//   - float64: contrast ratio
func contrast(t *testing.T, fg, bg string) float64 {
	t.Helper()
	fr, fg2, fb := ansiToRGB(t, fg)
	br, bg2, bb := ansiToRGB(t, bg)
	lf, lb := relLum(fr, fg2, fb), relLum(br, bg2, bb)
	hi, lo := math.Max(lf, lb), math.Min(lf, lb)
	return (hi + 0.05) / (lo + 0.05)
}

func TestEverySegmentInkIsReadableOnItsOwnGround(t *testing.T) {
	// A status line is glanced at, never studied. Pastel-on-pastel was the
	// project's original palette and measured as low as 2.08:1 — below the 3:1
	// floor for large text, let alone the 4.5:1 one for body text.
	const minRatio = 4.5

	tests := []struct {
		name string
		bg   string
		fg   string
	}{
		{name: "opus", bg: BgOpus, fg: FgOpusDark},
		{name: "sonnet", bg: BgSonnet, fg: FgSonnetDark},
		{name: "haiku", bg: BgHaiku, fg: FgHaikuDark},
		{name: "fable", bg: BgFable, fg: FgFableDark},
		{name: "unknown model", bg: BgModelUnknown, fg: FgModelUnknownDark},
		{name: "path", bg: BgBlue, fg: FgBlueDark},
		{name: "git", bg: BgCyan, fg: FgCyanDark},
		{name: "lines added", bg: BgGreen, fg: FgGreenText},
		{name: "lines removed", bg: BgRed, fg: FgRedText},
		{name: "mcp pill", bg: BgMCPEnabled, fg: FgMCPEnabledText},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contrast(t, tt.fg, tt.bg); got < minRatio {
				t.Errorf("contrast %.2f:1 is below the %.1f:1 floor", got, minRatio)
			}
		})
	}
}
