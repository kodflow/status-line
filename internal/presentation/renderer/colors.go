// Package renderer provides status line rendering.
package renderer

import "strings"

// Color constants for terminal rendering using ANSI escape codes.
const (
	// Reset resets all terminal attributes.
	Reset string = "\033[0m"
	// Bold enables bold text.
	Bold string = "\033[1m"
	// BgBlue is the dark ground of the path segment.
	BgBlue string = "\033[48;5;236m"
	// BgCyan is the pale cyan background color.
	BgCyan string = "\033[48;5;116m"
	// BgWhite is the white background color.
	BgWhite string = "\033[48;5;255m"
	// FgBlueDark is the light ink on the path ground.
	FgBlueDark string = "\033[38;5;250m"
	// FgCyanDark is a darker cyan for text on cyan background.
	FgCyanDark string = "\033[38;5;23m"
	// BgHaiku is the powder blush background for Haiku.
	BgHaiku string = "\033[48;5;224m"
	// BgSonnet is the powder lavender background for Sonnet.
	BgSonnet string = "\033[48;5;189m"
	// BgOpus is the powder peach background for Opus.
	BgOpus string = "\033[48;5;223m"
	// BgFable is the mint background for Fable. The other families are all
	// warm — orange, violet, pink — so a cool hue tells Fable apart at once.
	BgFable string = "\033[48;5;194m"
	// BgModelUnknown is the neutral background for an unrecognised model. It is
	// deliberately not white: that is the OS segment's colour, and a model
	// sharing it merges with the segment before it, separator included.
	BgModelUnknown string = "\033[48;5;252m"
	// FgBlue is the path ground as a foreground, for separators.
	FgBlue string = "\033[38;5;236m"
	// FgCyan is the pale cyan foreground color for separators.
	FgCyan string = "\033[38;5;116m"
	// FgWhite is the white foreground color.
	FgWhite string = "\033[38;5;255m"
	// FgBlack is the black foreground color.
	FgBlack string = "\033[38;5;232m"
	// FgYellow is the yellow foreground color.
	FgYellow string = "\033[38;5;220m"
	// FgHaiku is the powder blush foreground for Haiku pill caps.
	FgHaiku string = "\033[38;5;224m"
	// FgSonnet is the powder lavender foreground for Sonnet pill caps.
	FgSonnet string = "\033[38;5;189m"
	// FgOpus is the powder peach foreground for Opus pill caps.
	FgOpus string = "\033[38;5;223m"
	// FgFable is the mint foreground for Fable pill caps.
	FgFable string = "\033[38;5;194m"
	// FgModelUnknown is the neutral foreground for an unrecognised model.
	FgModelUnknown string = "\033[38;5;252m"
	// FgModelUnknownDark is the grey ink on an unrecognised model. 5.39:1.
	FgModelUnknownDark string = "\033[38;5;239m"
	// BgGreen is the pale green background for lines added.
	BgGreen string = "\033[48;5;114m"
	// FgGreenText is the darker green for text on green background.
	FgGreenText string = "\033[38;5;22m"
	// FgGreenSep is the pale green foreground for separator.
	FgGreenSep string = "\033[38;5;114m"
	// BgRed is the pale red background for lines removed.
	BgRed string = "\033[48;5;174m"
	// FgRedText is the darker red for text on red background.
	FgRedText string = "\033[38;5;52m"
	// FgRedSep is the pale red foreground for separator.
	FgRedSep string = "\033[38;5;174m"
	// ColorGray is the gray foreground for incomplete progress.
	ColorGray string = "\033[38;5;245m"
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
	// BgMCPEnabled is the pale teal background for enabled MCP servers.
	BgMCPEnabled string = "\033[48;5;116m"
	// FgMCPEnabled is the pale teal foreground for enabled MCP pill caps.
	FgMCPEnabled string = "\033[38;5;116m"
	// FgMCPEnabledText is the dark teal for text on enabled MCP background.
	FgMCPEnabledText string = "\033[38;5;23m"
	// BgMCPDisabled is the pale gray background for disabled MCP servers.
	BgMCPDisabled string = "\033[48;5;250m"
	// FgMCPDisabled is the pale gray foreground for disabled MCP pill caps.
	FgMCPDisabled string = "\033[38;5;250m"
	// FgMCPDisabledText is the dark gray for text on disabled MCP background.
	FgMCPDisabledText string = "\033[38;5;240m"
	// BgWeekly is the pale gray background for weekly usage segment.
	BgWeekly string = "\033[48;5;252m"
	// FgWeekly is the pale gray foreground for weekly usage separator.
	FgWeekly string = "\033[38;5;252m"
	// FgWeeklyText is the dark gray text on weekly usage background.
	FgWeeklyText string = "\033[38;5;240m"
)

// Model inks. Each is its ground's hue, darkened until it reads at about
// 5.4:1 — legible without sitting on the pill like a block of ink. The text,
// the quota bars and the pace cursor all take it.
//
// The 256-colour cube has nothing in these hues between a deep ink at 7:1 and
// one under 4.5:1, so the tuned inks are 24-bit. Each has a 256-colour
// fallback: the nearest cube colour (CIELAB) that still clears 4.5:1.
const (
	// inkHaikuTrue is the muted red ink on the Haiku ground. 5.44:1.
	inkHaikuTrue string = "\033[38;2;155;52;52m"
	// inkHaiku256 is its cube fallback, #870000. 7.86:1.
	inkHaiku256 string = "\033[38;5;88m"
	// inkSonnetTrue is the muted indigo ink on the Sonnet ground. 5.43:1.
	inkSonnetTrue string = "\033[38;2;66;66;192m"
	// inkSonnet256 is its cube fallback, #5f00af. 7.15:1.
	inkSonnet256 string = "\033[38;5;55m"
	// inkOpusTrue is the muted brown ink on the Opus ground. 5.42:1.
	inkOpusTrue string = "\033[38;2;117;78;39m"
	// inkOpus256 is its cube fallback, #585858. The nearer olive (58) reads
	// as dirty green on peach, so the next nearest, a neutral grey, is used.
	// 5.28:1.
	inkOpus256 string = "\033[38;5;240m"
	// inkFableTrue is the muted green ink on the Fable ground. 5.45:1.
	inkFableTrue string = "\033[38;2;38;114;38m"
	// inkFable256 is its cube fallback, #005f00. 7.27:1.
	inkFable256 string = "\033[38;5;22m"
)

// Active model inks, resolved once at startup from the terminal's colour depth.
var (
	// FgHaikuDark is the ink on the Haiku background.
	FgHaikuDark string = pickInk(inkHaikuTrue, inkHaiku256)
	// FgSonnetDark is the ink on the Sonnet background.
	FgSonnetDark string = pickInk(inkSonnetTrue, inkSonnet256)
	// FgOpusDark is the ink on the Opus background.
	FgOpusDark string = pickInk(inkOpusTrue, inkOpus256)
	// FgFableDark is the ink on the Fable background.
	FgFableDark string = pickInk(inkFableTrue, inkFable256)
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
	// Mint colors for Fable model. Without an entry of its own Fable fell back
	// to white, which is the OS segment's colour: the two segments then merged
	// into one block with an invisible separator between them.
	case strings.Contains(nameLower, "fable"):
		// Return Fable colors
		return BgFable, FgFable, FgFableDark
	// Neutral colors for unknown models
	default:
		// Return default colors
		return BgModelUnknown, FgModelUnknown, FgModelUnknownDark
	}
}
