// Package renderer provides status line rendering.
package renderer

import (
	"os"
	"strings"
)

// Colour depth selection.
const (
	// colorsEnv forces the colour depth: "truecolor" or "256". Unset, the
	// depth follows COLORTERM.
	colorsEnv string = "STATUSLINE_COLORS"
	// colorTermEnv is the de facto variable terminals set to announce 24-bit
	// colour.
	colorTermEnv string = "COLORTERM"
)

// trueColor reports whether 24-bit escapes are used, resolved once at startup.
var trueColor bool = trueColorFromEnv()

// trueColorFromEnv resolves the colour depth from the environment.
//
// A terminal that never announces 24-bit colour gets the 256-colour fallback:
// an unsupported 24-bit escape is dropped or misread, and the ink vanishes.
//
// Returns:
//   - bool: true when 24-bit colour is in use
func trueColorFromEnv() bool {
	// An explicit setting wins over detection
	switch strings.ToLower(os.Getenv(colorsEnv)) {
	// Forced 24-bit colour
	case "truecolor", "24bit":
		return true
	// Forced cube colours
	case "256":
		return false
	}
	// Otherwise trust what the terminal announces
	switch strings.ToLower(os.Getenv(colorTermEnv)) {
	// Both spellings are in use
	case "truecolor", "24bit":
		return true
	// Unset or anything else: stay on the cube
	default:
		return false
	}
}

// pickInk returns the ink for the active colour depth.
//
// Params:
//   - trueInk: 24-bit escape
//   - cubeInk: 256-colour fallback escape
//
// Returns:
//   - string: escape to render with
func pickInk(trueInk, cubeInk string) string {
	// Use the tuned ink only where the terminal can show it
	if trueColor {
		return trueInk
	}
	return cubeInk
}
