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

// Background color constants using 256-color palette (pale tones).
const (
	// BgBlue is the pale blue background color.
	BgBlue string = "\033[48;5;111m"
	// BgCyan is the pale cyan background color.
	BgCyan string = "\033[48;5;116m"
	// BgWhite is the white background color.
	BgWhite string = "\033[48;5;255m"
)

// Darker foreground variants for text on colored backgrounds.
const (
	// FgBlueDark is a darker blue for text on blue background.
	FgBlueDark string = "\033[38;5;25m"
	// FgCyanDark is a darker cyan for text on cyan background.
	FgCyanDark string = "\033[38;5;30m"
)

// Model-specific background colors (pale tones).
const (
	// BgHaiku is the pale pink background for Haiku.
	BgHaiku string = "\033[48;5;218m"
	// BgSonnet is the pale purple background for Sonnet.
	BgSonnet string = "\033[48;5;183m"
	// BgOpus is the pale orange background for Opus.
	BgOpus string = "\033[48;5;222m"
)

// Foreground color constants using 256-color palette.
const (
	// FgBlue is the pale blue foreground color for separators.
	FgBlue string = "\033[38;5;111m"
	// FgCyan is the pale cyan foreground color for separators.
	FgCyan string = "\033[38;5;116m"
	// FgWhite is the white foreground color.
	FgWhite string = "\033[38;5;255m"
	// FgBlack is the black foreground color.
	FgBlack string = "\033[38;5;232m"
	// FgYellow is the yellow foreground color.
	FgYellow string = "\033[38;5;220m"
)

// Model-specific foreground colors for pill caps (pale tones).
const (
	// FgHaiku is the pale pink foreground for Haiku pill caps.
	FgHaiku string = "\033[38;5;218m"
	// FgSonnet is the pale purple foreground for Sonnet pill caps.
	FgSonnet string = "\033[38;5;183m"
	// FgOpus is the pale orange foreground for Opus pill caps.
	FgOpus string = "\033[38;5;222m"
)

// Darker foreground variants for model text.
const (
	// FgHaikuDark is a darker pink for text on Haiku background.
	FgHaikuDark string = "\033[38;5;168m"
	// FgSonnetDark is a darker purple for text on Sonnet background.
	FgSonnetDark string = "\033[38;5;97m"
	// FgOpusDark is a darker orange for text on Opus background.
	FgOpusDark string = "\033[38;5;172m"
)

// Progress bar color constants (darker for white background).
const (
	// ColorGreen is the dark green foreground for low usage.
	ColorGreen string = "\033[38;5;28m"
	// ColorYellow is the dark yellow/olive foreground for medium usage.
	ColorYellow string = "\033[38;5;136m"
	// ColorOrange is the dark orange foreground for high usage.
	ColorOrange string = "\033[38;5;166m"
	// ColorRed is the dark red foreground for critical usage.
	ColorRed string = "\033[38;5;124m"
)

// Code changes segment colors (pale backgrounds, darker text).
const (
	// BgGreen is the pale green background for lines added.
	BgGreen string = "\033[48;5;114m"
	// FgGreenText is the darker green for text on green background.
	FgGreenText string = "\033[38;5;28m"
	// FgGreenSep is the pale green foreground for separator.
	FgGreenSep string = "\033[38;5;114m"
	// BgRed is the pale red background for lines removed.
	BgRed string = "\033[48;5;174m"
	// FgRedText is the darker red for text on red background.
	FgRedText string = "\033[38;5;124m"
	// FgRedSep is the pale red foreground for separator.
	FgRedSep string = "\033[38;5;174m"
)

// MCP server color constants (pale backgrounds, darker text).
const (
	// BgMCPEnabled is the pale teal background for enabled MCP servers.
	BgMCPEnabled string = "\033[48;5;116m"
	// FgMCPEnabled is the pale teal foreground for enabled MCP pill caps.
	FgMCPEnabled string = "\033[38;5;116m"
	// FgMCPEnabledText is the dark teal for text on enabled MCP background.
	FgMCPEnabledText string = "\033[38;5;30m"
	// BgMCPDisabled is the pale gray background for disabled MCP servers.
	BgMCPDisabled string = "\033[48;5;250m"
	// FgMCPDisabled is the pale gray foreground for disabled MCP pill caps.
	FgMCPDisabled string = "\033[38;5;250m"
	// FgMCPDisabledText is the dark gray for text on disabled MCP background.
	FgMCPDisabledText string = "\033[38;5;240m"
)

// Taskwarrior color constants (lavender/violet theme - distinct from Sonnet).
const (
	// BgTaskwarrior is the pale lavender background for Taskwarrior pill.
	BgTaskwarrior string = "\033[48;5;147m"
	// FgTaskwarrior is the lavender foreground for Taskwarrior pill caps.
	FgTaskwarrior string = "\033[38;5;147m"
	// FgTaskwarriorText is the dark indigo for text on Taskwarrior background.
	FgTaskwarriorText string = "\033[38;5;55m"
)

// Taskwarrior progress bar colors.
const (
	// ColorGray is the gray foreground for incomplete progress.
	ColorGray string = "\033[38;5;245m"
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
//   - textColor: darker foreground color for text on the pill
func GetModelColors(modelName string) (bgColor, fgColor, textColor string) {
	nameLower := strings.ToLower(modelName)
	// Detect model type and return appropriate colors
	switch {
	// Pink colors for Haiku model
	case strings.Contains(nameLower, "haiku"):
		// Return Haiku colors
		return BgHaiku, FgHaiku, FgHaikuDark
	// Purple colors for Sonnet model
	case strings.Contains(nameLower, "sonnet"):
		// Return Sonnet colors
		return BgSonnet, FgSonnet, FgSonnetDark
	// Orange colors for Opus model
	case strings.Contains(nameLower, "opus"):
		// Return Opus colors
		return BgOpus, FgOpus, FgOpusDark
	// White colors for unknown models
	default:
		// Return default colors
		return BgWhite, FgWhite, FgBlack
	}
}
