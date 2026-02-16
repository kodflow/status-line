// Package renderer provides status line rendering.
package renderer

import "strings"

// Color constants for terminal rendering using ANSI escape codes.
const (
	// Reset resets all terminal attributes.
	Reset string = "\033[0m"
	// Bold enables bold text.
	Bold string = "\033[1m"
	// BgBlue is the pale blue background color.
	BgBlue string = "\033[48;5;111m"
	// BgCyan is the pale cyan background color.
	BgCyan string = "\033[48;5;116m"
	// BgWhite is the white background color.
	BgWhite string = "\033[48;5;255m"
	// FgBlueDark is a darker blue for text on blue background.
	FgBlueDark string = "\033[38;5;25m"
	// FgCyanDark is a darker cyan for text on cyan background.
	FgCyanDark string = "\033[38;5;30m"
	// BgHaiku is the pale pink background for Haiku.
	BgHaiku string = "\033[48;5;218m"
	// BgSonnet is the pale purple background for Sonnet.
	BgSonnet string = "\033[48;5;183m"
	// BgOpus is the pale orange background for Opus.
	BgOpus string = "\033[48;5;222m"
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
	// FgHaiku is the pale pink foreground for Haiku pill caps.
	FgHaiku string = "\033[38;5;218m"
	// FgSonnet is the pale purple foreground for Sonnet pill caps.
	FgSonnet string = "\033[38;5;183m"
	// FgOpus is the pale orange foreground for Opus pill caps.
	FgOpus string = "\033[38;5;222m"
	// FgHaikuDark is a darker pink for text on Haiku background.
	FgHaikuDark string = "\033[38;5;168m"
	// FgSonnetDark is a darker purple for text on Sonnet background.
	FgSonnetDark string = "\033[38;5;97m"
	// FgOpusDark is a darker orange for text on Opus background.
	FgOpusDark string = "\033[38;5;172m"
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
	// BgTaskwarrior is the pale lavender background for Taskwarrior pill.
	BgTaskwarrior string = "\033[48;5;147m"
	// FgTaskwarrior is the lavender foreground for Taskwarrior pill caps.
	FgTaskwarrior string = "\033[38;5;147m"
	// FgTaskwarriorText is the dark indigo for text on Taskwarrior background.
	FgTaskwarriorText string = "\033[38;5;55m"
	// ColorGray is the gray foreground for incomplete progress.
	ColorGray string = "\033[38;5;245m"
	// FgCursorOrange is the dark orange foreground for burn-rate cursor.
	FgCursorOrange string = "\033[38;5;166m"
	// FgGreenDone is the green foreground for done tasks in segmented bar.
	FgGreenDone string = "\033[38;5;34m"
	// FgYellowWip is the yellow/orange foreground for WIP indicator.
	FgYellowWip string = "\033[38;5;214m"
	// FgGrayTodo is the dark gray foreground for todo tasks.
	FgGrayTodo string = "\033[38;5;240m"
	// FgGraySep is the separator color between epics.
	FgGraySep string = "\033[38;5;245m"
	// FgCyanTask is the cyan foreground for current task indicator.
	FgCyanTask string = "\033[38;5;44m"
	// BgWeekly is the pale gray background for weekly usage segment.
	BgWeekly string = "\033[48;5;252m"
	// FgWeekly is the pale gray foreground for weekly usage separator.
	FgWeekly string = "\033[38;5;252m"
	// FgWeeklyText is the dark gray text on weekly usage background.
	FgWeeklyText string = "\033[38;5;240m"
)

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
