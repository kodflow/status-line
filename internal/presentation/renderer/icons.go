// Package renderer provides status line rendering.
package renderer

import "github.com/florent/status-line/internal/domain/model"

// Powerline separators.
const (
	SepRight   = "\uE0B0"
	SepLeft    = "\uE0B2"
	LeftRound  = "\uE0B6"
	RightRound = "\uE0B4"
)

// Nerd Font icons.
const (
	IconLinux     = "\uf17c"
	IconDocker    = "\uf308"
	IconApple     = "\uf179"
	IconWindows   = "\uf17a"
	IconDesktop   = "\uf108"
	IconFolder    = "\uf07b"
	IconGitBranch = "\ue0a0"
	IconModel     = "\uf2db" // Microchip
)

// GetOSIcon returns the appropriate icon for the OS type.
func GetOSIcon(osType model.OSType, isDocker bool) string {
	if osType == model.OSLinux && isDocker {
		return IconDocker
	}
	switch osType {
	case model.OSLinux:
		return IconLinux
	case model.OSDarwin:
		return IconApple
	case model.OSWindows:
		return IconWindows
	default:
		return IconDesktop
	}
}
