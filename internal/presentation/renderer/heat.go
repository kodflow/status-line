// Package renderer provides status line rendering.
package renderer

import (
	"os"
	"strconv"
)

// Context heat configuration.
const (
	// compactEnv sets the percentage at which the context reads as critical.
	compactEnv string = "STATUSLINE_COMPACT_PCT"
	// defaultCompactPct is where the context turns red by default.
	//
	// Claude Code has no single auto-compaction threshold to align with: it
	// depends on the model and on the window the user picked, and
	// CLAUDE_AUTOCOMPACT_PCT_OVERRIDE can move it anywhere. The payload does not
	// carry it either. 90 matches the documented default for a 200k window;
	// anyone who moved theirs sets this to match.
	defaultCompactPct int = 90
	// warmPct is where the context starts warming up.
	warmPct int = 40
	// hotPct is where it reads as hot, short of critical.
	hotPct int = 60
)

// Context ground and heat inks.
//
// The segment sits on a dark grey rather than on white, and that is what lets
// the scale use the colours the eye reads as heat. On white, every warm hue
// bright enough to look like amber or orange fails the contrast floor — amber
// measures 2.88:1 there — so the scale had to fall back on burnt, muddy tones.
// Against #303030 the same scale runs vivid, and every step clears WCAG AA
// (4.5:1) with margin to spare.
//
// Every step clears WCAG AA (4.5:1) against the segment's own white ground.
// The obvious bright amber and orange do not: they measure 2.88:1 and 3.28:1,
// which is below the 3:1 floor for large text in the first case and only
// passes on boldness in the second. A warning nobody can read is not one.
const (
	// BgContext is the dark grey ground of the context segment.
	BgContext string = "\033[48;5;236m"
	// FgContext is that ground as a foreground, for the powerline separator.
	FgContext string = "\033[38;5;236m"
	// FgHeatCalm is the neutral ink of a context with room to spare. 6.95:1.
	FgHeatCalm string = "\033[38;5;250m"
	// FgHeatWarm is the gold ink of a filling context. 7.15:1.
	FgHeatWarm string = "\033[38;5;214m"
	// FgHeatHot is the peach ink of a context worth watching. 7.26:1.
	FgHeatHot string = "\033[38;5;215m"
	// FgHeatCritical is the red of a context about to be compacted. 5.70:1.
	FgHeatCritical string = "\033[38;5;210m"
)

// compactPct is the resolved critical threshold, read once at startup.
var compactPct = compactPctFromEnv()

// compactPctFromEnv resolves the critical threshold from the environment.
//
// Returns:
//   - int: percentage at which the context reads as critical
func compactPctFromEnv() int {
	raw := os.Getenv(compactEnv)
	// An unset variable keeps the default
	if raw == "" {
		return defaultCompactPct
	}
	parsed, err := strconv.Atoi(raw)
	// A malformed or out-of-range value keeps the default rather than
	// colouring the whole range red or never colouring it at all
	if err != nil || parsed <= 0 || parsed > percentComplete {
		return defaultCompactPct
	}
	return parsed
}

// ContextHeat returns the ink for a context percentage.
//
// The colour rises with the window so the state is legible before the figure
// is read, and the top step lands on the compaction threshold: red means the
// conversation is about to be rewritten, not merely that the bar looks full.
//
// Params:
//   - percent: context window consumption, 0-100
//
// Returns:
//   - string: ANSI foreground for that level
func ContextHeat(percent int) string {
	// Map the percentage onto the four steps, hottest first
	switch {
	// About to be compacted
	case percent >= compactPct:
		return FgHeatCritical
	// Worth watching
	case percent >= hotPct:
		return FgHeatHot
	// Filling up
	case percent >= warmPct:
		return FgHeatWarm
	// Room to spare
	default:
		return FgHeatCalm
	}
}
