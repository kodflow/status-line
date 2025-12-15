// Package renderer provides status line rendering.
package renderer

import "github.com/florent/status-line/internal/domain/model"

// Progress bar width constant.
const (
	// progressBarWidth is the number of characters in the progress bar.
	progressBarWidth int = 10
)

// ProgressBarStyle represents the visual style of progress bars.
type ProgressBarStyle int

// Progress bar style constants.
const (
	// StyleBraille uses braille pattern characters.
	StyleBraille ProgressBarStyle = iota
	// StyleBlock uses Unicode block characters (pip/hyperfine style).
	StyleBlock
	// StyleHeavy uses heavy horizontal line characters.
	StyleHeavy
)

// Braille progress bar characters (empty to full).
// Uses vertical braille patterns for smooth progression.
var brailleChars = []rune{
	' ',      // 0/8 - empty
	'\u2801', // ⠁ 1/8
	'\u2803', // ⠃ 2/8
	'\u2807', // ⠇ 3/8
	'\u2847', // ⡇ 4/8
	'\u28C7', // ⣇ 5/8
	'\u28E7', // ⣧ 6/8
	'\u28F7', // ⣷ 7/8
	'\u28FF', // ⣿ 8/8 - full
}

// Block progress bar characters (empty to full).
// Uses horizontal block elements for pip/hyperfine style.
var blockChars = []rune{
	' ',      // 0/8 - empty
	'\u258F', // ▏ 1/8
	'\u258E', // ▎ 2/8
	'\u258D', // ▍ 3/8
	'\u258C', // ▌ 4/8
	'\u258B', // ▋ 5/8
	'\u258A', // ▊ 6/8
	'\u2589', // ▉ 7/8
	'\u2588', // █ 8/8 - full
}

// Heavy horizontal progress bar characters.
const (
	// heavyFull is the filled heavy horizontal character.
	heavyFull rune = '\u2501' // ━
	// heavyEmpty is the empty light horizontal character.
	heavyEmpty rune = '\u2500' // ─
)

// RenderProgressBar generates a progress bar string.
//
// Params:
//   - progress: the progress value (0-100%)
//   - style: the visual style to use
//
// Returns:
//   - string: rendered progress bar
func RenderProgressBar(progress model.Progress, style ProgressBarStyle) string {
	// Handle heavy style separately (no partial characters)
	if style == StyleHeavy {
		// Calculate filled characters
		filled := progress.Percent * progressBarWidth / 100
		result := make([]rune, progressBarWidth)
		// Build progress bar character by character
		for i := range progressBarWidth {
			// Determine if filled or empty
			if i < filled {
				result[i] = heavyFull
			} else {
				result[i] = heavyEmpty
			}
		}
		// Return completed progress bar
		return string(result)
	}

	chars := blockChars
	// Select character set based on style
	if style == StyleBraille {
		// Use braille characters
		chars = brailleChars
	}

	// Calculate filled portion (0-80 for 10 chars * 8 steps)
	totalSteps := progressBarWidth * 8
	filledSteps := progress.Percent * totalSteps / 100

	result := make([]rune, progressBarWidth)
	// Build progress bar character by character
	for i := range progressBarWidth {
		charSteps := filledSteps - (i * 8)
		// Determine character based on remaining steps
		switch {
		// Full character
		case charSteps >= 8:
			result[i] = chars[8]
		// Partial character
		case charSteps > 0:
			result[i] = chars[charSteps]
		// Empty character
		default:
			result[i] = chars[0]
		}
	}

	// Return completed progress bar
	return string(result)
}
