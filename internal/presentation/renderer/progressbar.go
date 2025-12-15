// Package renderer provides status line rendering.
package renderer

import "github.com/florent/status-line/internal/domain/model"

// Progress bar dimension constants.
const (
	// progressBarWidth is the number of characters in the progress bar.
	progressBarWidth int = 10
	// stepsPerChar is the number of granular steps per character (for smooth rendering).
	stepsPerChar int = 8
	// charSetSize is the number of characters in braille/block sets (0-8 inclusive).
	charSetSize int = 9
	// percentMax is the maximum percentage value.
	percentMax int = 100
)

// Heavy horizontal progress bar characters.
const (
	// heavyFull is the filled heavy horizontal character.
	heavyFull rune = '\u2501' // ━
	// heavyEmpty is the empty light horizontal character.
	heavyEmpty rune = '\u2500' // ─
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

// Progress bar character sets.
var (
	// brailleChars contains braille progress bar characters (empty to full).
	// Uses vertical braille patterns for smooth progression.
	brailleChars [charSetSize]rune = [charSetSize]rune{
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

	// blockChars contains block progress bar characters (empty to full).
	// Uses horizontal block elements for pip/hyperfine style.
	blockChars [charSetSize]rune = [charSetSize]rune{
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
		// Render heavy style progress bar
		return renderHeavyBar(progress)
	}

	// Render granular style (braille or block)
	return renderGranularBar(progress, style)
}

// renderHeavyBar renders a progress bar using heavy horizontal characters.
//
// Params:
//   - progress: the progress value
//
// Returns:
//   - string: rendered progress bar
func renderHeavyBar(progress model.Progress) string {
	// Calculate filled characters
	filled := progress.Percent * progressBarWidth / percentMax
	var result [progressBarWidth]rune
	// Build progress bar character by character
	for i := range progressBarWidth {
		// Determine if filled or empty
		if i < filled {
			result[i] = heavyFull
		} else {
			// Use empty character
			result[i] = heavyEmpty
		}
	}
	// Return completed progress bar
	return string(result[:])
}

// renderGranularBar renders a progress bar using braille or block characters.
//
// Params:
//   - progress: the progress value
//   - style: StyleBraille or StyleBlock
//
// Returns:
//   - string: rendered progress bar
func renderGranularBar(progress model.Progress, style ProgressBarStyle) string {
	chars := blockChars
	// Select character set based on style
	if style == StyleBraille {
		// Use braille characters
		chars = brailleChars
	}

	// Calculate filled portion
	totalSteps := progressBarWidth * stepsPerChar
	filledSteps := progress.Percent * totalSteps / percentMax

	var result [progressBarWidth]rune
	// Build progress bar character by character
	for i := range progressBarWidth {
		charSteps := filledSteps - (i * stepsPerChar)
		// Determine character based on remaining steps
		switch {
		// Full character
		case charSteps >= stepsPerChar:
			result[i] = chars[stepsPerChar]
		// Partial character
		case charSteps > 0:
			result[i] = chars[charSteps]
		// Empty character
		default:
			result[i] = chars[0]
		}
	}

	// Return completed progress bar
	return string(result[:])
}
