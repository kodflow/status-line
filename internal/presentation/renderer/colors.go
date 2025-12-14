// Package renderer provides status line rendering.
package renderer

import (
	"strings"

	"github.com/florent/status-line/internal/domain/model"
)

// ANSI escape codes for terminal formatting.
const (
	// Reset resets all terminal attributes.
	Reset string = "\033[0m"
	// Bold enables bold text.
	Bold string = "\033[1m"
)

// Background color constants using 256-color palette.
const (
	// BgBlue is the blue background color.
	BgBlue string = "\033[48;5;31m"
	// BgCyan is the cyan background color.
	BgCyan string = "\033[48;5;37m"
	// BgWhite is the white background color.
	BgWhite string = "\033[48;5;255m"
)

// Model-specific background colors.
const (
	// BgHaiku is the light pink background for Haiku.
	BgHaiku string = "\033[48;5;218m"
	// BgSonnet is the purple background for Sonnet.
	BgSonnet string = "\033[48;5;141m"
	// BgOpus is the gold/orange background for Opus.
	BgOpus string = "\033[48;5;214m"
)

// Foreground color constants using 256-color palette.
const (
	// FgBlue is the blue foreground color.
	FgBlue string = "\033[38;5;31m"
	// FgCyan is the cyan foreground color.
	FgCyan string = "\033[38;5;37m"
	// FgWhite is the white foreground color.
	FgWhite string = "\033[38;5;255m"
	// FgBlack is the black foreground color.
	FgBlack string = "\033[38;5;232m"
	// FgYellow is the yellow foreground color.
	FgYellow string = "\033[38;5;220m"
)

// Model-specific foreground colors for pill caps.
const (
	// FgHaiku is the light pink foreground for Haiku.
	FgHaiku string = "\033[38;5;218m"
	// FgSonnet is the purple foreground for Sonnet.
	FgSonnet string = "\033[38;5;141m"
	// FgOpus is the gold/orange foreground for Opus.
	FgOpus string = "\033[38;5;214m"
)

// Progress bar color constants.
const (
	// ColorGreen is the color for low usage.
	ColorGreen string = "\033[38;5;76m"
	// ColorYellow is the color for medium usage.
	ColorYellow string = "\033[38;5;220m"
	// ColorOrange is the color for high usage.
	ColorOrange string = "\033[38;5;208m"
	// ColorRed is the color for critical usage.
	ColorRed string = "\033[38;5;196m"
)

// GetProgressColor returns color for a progress level.
//
// Params:
//   - level: the progress severity level
//
// Returns:
//   - string: ANSI color code for the level
func GetProgressColor(level model.ProgressLevel) string {
	// Select color based on level
	switch level {
	// Green for low/safe usage
	case model.LevelLow:
		// Return green color
		return ColorGreen
	// Yellow for medium usage
	case model.LevelMedium:
		// Return yellow color
		return ColorYellow
	// Orange for high usage
	case model.LevelHigh:
		// Return orange color
		return ColorOrange
	// Red for critical usage
	default:
		// Return red color
		return ColorRed
	}
}

// GetModelColors returns colors for a model pill.
//
// Params:
//   - modelName: the display name of the AI model
//
// Returns:
//   - bgColor: background color for the pill
//   - fgColor: foreground color for the pill caps
func GetModelColors(modelName string) (bgColor, fgColor string) {
	nameLower := strings.ToLower(modelName)
	// Detect model type and return appropriate colors
	switch {
	// Pink colors for Haiku model
	case strings.Contains(nameLower, "haiku"):
		// Return Haiku colors
		return BgHaiku, FgHaiku
	// Purple colors for Sonnet model
	case strings.Contains(nameLower, "sonnet"):
		// Return Sonnet colors
		return BgSonnet, FgSonnet
	// Orange colors for Opus model
	case strings.Contains(nameLower, "opus"):
		// Return Opus colors
		return BgOpus, FgOpus
	// White colors for unknown models
	default:
		// Return default colors
		return BgWhite, FgWhite
	}
}
