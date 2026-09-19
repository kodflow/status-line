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
)

// Context ground and inks.
//
// The context sits on the blue the path used to have, and the path takes the
// dark grey. That swap costs the heat scale: gold, peach and orange all fail
// the contrast floor against #87afff, and a warning nobody can read is not one.
//
// What survives is the step that matters. Deep red clears 6.46:1 on the blue,
// so the segment stays quiet until the window is about to be compacted, and
// then says so. A gradient nobody could read bought nothing the percentage did
// not already give.
const (
	// BgContext is the blue ground of the context segment.
	BgContext string = "[48;5;111m"
	// FgContext is that ground as a foreground, for the powerline separator.
	FgContext string = "[38;5;111m"
	// FgHeatCalm is the deep blue ink of a context below the threshold. 4.85:1.
	FgHeatCalm string = "[38;5;20m"
	// FgHeatCritical is the deep red of a context about to be compacted. 6.46:1.
	FgHeatCritical string = "[38;5;52m"
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
// Red lands on the compaction threshold: it means the conversation is about to
// be rewritten, not merely that the bar looks full.
//
// Params:
//   - percent: context window consumption, 0-100
//
// Returns:
//   - string: ANSI foreground for that level
func ContextHeat(percent int) string {
	// One step, on the only threshold that changes what happens next
	if percent >= compactPct {
		return FgHeatCritical
	}
	return FgHeatCalm
}
