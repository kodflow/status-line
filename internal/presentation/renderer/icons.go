// Package renderer provides status line rendering.
package renderer

import "github.com/florent/status-line/internal/domain/model"

// Icon and separator constants for powerline rendering.
const (
	// SepRight is the right-pointing arrow separator.
	SepRight string = "\uE0B0"
	// SepLeft is the left-pointing arrow separator.
	SepLeft string = "\uE0B2"
	// LeftRound is the left rounded cap.
	LeftRound string = "\uE0B6"
	// RightRound is the right rounded cap.
	RightRound string = "\uE0B4"
	// IconLinux is the Linux penguin icon.
	IconLinux string = "\uf17c"
	// IconDocker is the Docker whale icon.
	IconDocker string = "\uf308"
	// IconApple is the Apple logo icon.
	IconApple string = "\uf179"
	// IconWindows is the Windows logo icon.
	IconWindows string = "\uf17a"
	// IconDesktop is the generic desktop icon.
	IconDesktop string = "\uf108"
	// IconFolder is the folder icon.
	IconFolder string = "\uf07b"
	// IconGitBranch is the git branch icon.
	IconGitBranch string = "\ue0a0"
	// IconModel is the microchip icon for AI models.
	IconModel string = "\uf2db"
	// IconTaskwarrior is the tasks list icon for Taskwarrior.
	IconTaskwarrior string = "\uf0ae"
	// IconUpdate is the download/update icon.
	IconUpdate string = "\uf019"
	// IconWeekly is the calendar/clock icon for weekly usage.
	IconWeekly string = "\uf073"
)

// GetOSIcon returns the appropriate icon for the OS type.
//
// Params:
//   - osType: the operating system type
//   - isDocker: whether running in Docker
//
// Returns:
//   - string: Nerd Font icon character
func GetOSIcon(osType model.OSType, isDocker bool) string {
	// Check for Docker on Linux first
	if osType == model.OSLinux && isDocker {
		// Return Docker icon for containers
		return IconDocker
	}
	// Select icon based on OS type
	switch osType {
	// Linux penguin icon
	case model.OSLinux:
		// Return Linux icon
		return IconLinux
	// Apple logo icon
	case model.OSDarwin:
		// Return Apple icon
		return IconApple
	// Windows logo icon
	case model.OSWindows:
		// Return Windows icon
		return IconWindows
	// Generic desktop icon for unknown OS
	default:
		// Return desktop icon
		return IconDesktop
	}
}
