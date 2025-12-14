// Package renderer provides status line rendering.
package renderer

import (
	"strings"

	"github.com/florent/status-line/internal/domain/model"
)

// ANSI escape codes.
const (
	Reset = "\033[0m"
	Bold  = "\033[1m"

	BgBlue  = "\033[48;5;31m"
	BgCyan  = "\033[48;5;37m"
	BgWhite = "\033[48;5;255m"

	BgHaiku  = "\033[48;5;218m"
	BgSonnet = "\033[48;5;141m"
	BgOpus   = "\033[48;5;214m"

	FgBlue   = "\033[38;5;31m"
	FgCyan   = "\033[38;5;37m"
	FgWhite  = "\033[38;5;255m"
	FgBlack  = "\033[38;5;232m"
	FgYellow = "\033[38;5;220m"

	FgHaiku  = "\033[38;5;218m"
	FgSonnet = "\033[38;5;141m"
	FgOpus   = "\033[38;5;214m"

	ColorGreen  = "\033[38;5;76m"
	ColorYellow = "\033[38;5;220m"
	ColorOrange = "\033[38;5;208m"
	ColorRed    = "\033[38;5;196m"
)

// GetProgressColor returns color for a progress level.
func GetProgressColor(level model.ProgressLevel) string {
	switch level {
	case model.LevelLow:
		return ColorGreen
	case model.LevelMedium:
		return ColorYellow
	case model.LevelHigh:
		return ColorOrange
	default:
		return ColorRed
	}
}

// GetModelColors returns colors for a model pill.
func GetModelColors(modelName string) (bgColor, fgColor string) {
	nameLower := strings.ToLower(modelName)
	switch {
	case strings.Contains(nameLower, "haiku"):
		return BgHaiku, FgHaiku
	case strings.Contains(nameLower, "sonnet"):
		return BgSonnet, FgSonnet
	case strings.Contains(nameLower, "opus"):
		return BgOpus, FgOpus
	default:
		return BgWhite, FgWhite
	}
}
