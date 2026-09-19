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

// Context heat inks, from calm to critical, readable on the white ground.
const (
	// FgHeatCalm is the neutral ink of a context with room to spare.
	FgHeatCalm string = "\033[38;5;240m"
	// FgHeatWarm is the gold ink of a filling context.
	FgHeatWarm string = "\033[38;5;136m"
	// FgHeatHot is the orange ink of a context worth watching.
	FgHeatHot string = "\033[38;5;166m"
	// FgHeatCritical is the red ink of a context about to be compacted.
	FgHeatCritical string = "\033[38;5;88m"
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
