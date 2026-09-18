// Package renderer provides status line rendering.
package renderer

import (
	"os"
	"strconv"
	"strings"
)

// Vertical spacing between the two status line rows.
const (
	// lineGapEnv sets how many blank lines separate the two rows.
	lineGapEnv string = "STATUSLINE_LINE_GAP"
	// defaultLineGap keeps the rows adjacent; vertical breathing room is better
	// bought from the terminal's own cell height than from a line of status bar.
	defaultLineGap int = 0
	// maxLineGap bounds the spacing so the prompt stays usable.
	maxLineGap int = 3
)

// LineGap returns the blank lines written between the two status line rows.
//
// A terminal cell cannot be split, so vertical breathing room is whole blank
// lines or nothing.
//
// Returns:
//   - string: newlines written between rows
func LineGap() string {
	gap := defaultLineGap
	// Honour a configured height when it parses and is in range
	if raw := os.Getenv(lineGapEnv); raw != "" {
		parsed, err := strconv.Atoi(raw)
		// A malformed or out-of-range value keeps the default
		if err == nil && parsed >= 0 && parsed <= maxLineGap {
			gap = parsed
		}
	}
	return strings.Repeat("\n", gap)
}
