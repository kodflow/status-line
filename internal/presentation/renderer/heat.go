// Package renderer provides status line rendering.
package renderer

// Context segment colours.
//
// One ground, one ink, whatever the window holds. The percentage next to them
// already says how full it is, and a colour that changes underneath a figure
// that says the same thing is noise on a line read at a glance.
const (
	// BgContext is the blue ground of the context segment.
	BgContext string = "\033[48;5;111m"
	// FgContext is that ground as a foreground, for the powerline separator.
	FgContext string = "\033[38;5;111m"
	// FgContextInk is the deep blue ink on that ground. 4.85:1.
	FgContextInk string = "\033[38;5;20m"
)
