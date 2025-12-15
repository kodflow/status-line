// Package renderer provides status line rendering.
package renderer

import "github.com/florent/status-line/internal/domain/model"

// Progress bar constants.
const (
	// progressBarWidth is the number of characters in the progress bar.
	progressBarWidth int = 10
	// stepsPerChar is the number of granular steps per character.
	stepsPerChar int = 8
	// charSetSize is the number of characters in braille/block sets.
	charSetSize int = 9
	// percentMax is the maximum percentage value.
	percentMax int = 100
	// heavyFull is the filled heavy horizontal character.
	heavyFull rune = '\u2501'
	// heavyEmpty is the empty light horizontal character.
	heavyEmpty rune = '\u2500'
	// noCursor indicates no cursor should be rendered.
	noCursor int = -1
)

// ProgressBarStyle represents the visual style of progress bars.
type ProgressBarStyle int

// Progress bar style constants.
const (
	// StyleBraille uses braille pattern characters.
	StyleBraille ProgressBarStyle = iota
	// StyleBlock uses Unicode block characters.
	StyleBlock
	// StyleHeavy uses heavy horizontal line characters.
	StyleHeavy
)

// Progress bar character sets.
var (
	// brailleChars contains braille progress bar characters.
	brailleChars [charSetSize]rune = [charSetSize]rune{
		' ', '\u2801', '\u2803', '\u2807', '\u2847', '\u28C7', '\u28E7', '\u28F7', '\u28FF',
	}
	// blockChars contains block progress bar characters.
	blockChars [charSetSize]rune = [charSetSize]rune{
		' ', '\u258F', '\u258E', '\u258D', '\u258C', '\u258B', '\u258A', '\u2589', '\u2588',
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
	// Handle heavy style separately
	if style == StyleHeavy {
		// Render heavy style progress bar without cursor
		return renderHeavyBar(progress, noCursor)
	}
	// Render granular style
	return renderGranularBar(progress, style)
}

// RenderProgressBarWithCursor generates a progress bar with a burn-rate cursor.
//
// Params:
//   - progress: the progress value (0-100%)
//   - cursorPos: cursor position as percentage (0-100), or -1 for no cursor
//   - cursorColor: ANSI color code for the cursor
//   - bgColor: background color to restore after cursor
//
// Returns:
//   - string: rendered progress bar with colored cursor
func RenderProgressBarWithCursor(progress model.Progress, cursorPos int, cursorColor, bgColor string) string {
	// Calculate cursor character index
	cursorIdx := cursorPos * progressBarWidth / percentMax
	// Render bar with cursor
	return renderHeavyBarWithCursor(progress, cursorIdx, cursorColor, bgColor)
}

// renderHeavyBar renders a progress bar using heavy horizontal characters.
//
// Params:
//   - progress: the progress value
//   - cursorIdx: index for cursor character, or -1 for none
//
// Returns:
//   - string: rendered progress bar
func renderHeavyBar(progress model.Progress, cursorIdx int) string {
	// Calculate filled characters
	filled := progress.Percent * progressBarWidth / percentMax
	var result [progressBarWidth]rune
	// Build progress bar
	for i := range progressBarWidth {
		// Determine if filled or empty
		if i < filled {
			result[i] = heavyFull
		} else {
			// Use empty character for unfilled portion
			result[i] = heavyEmpty
		}
	}
	// Return completed progress bar
	return string(result[:])
}

// renderHeavyBarWithCursor renders a progress bar with a colored cursor.
//
// Params:
//   - progress: the progress value
//   - cursorIdx: index for cursor character (0 to progressBarWidth-1)
//   - cursorColor: ANSI foreground color for the cursor
//   - bgColor: background color to restore after cursor
//
// Returns:
//   - string: rendered progress bar with ANSI color codes for cursor
func renderHeavyBarWithCursor(progress model.Progress, cursorIdx int, cursorColor, bgColor string) string {
	// Calculate filled characters
	filled := progress.Percent * progressBarWidth / percentMax
	// Build result with color codes
	var result []byte
	// Build progress bar
	for i := range progressBarWidth {
		// Check if this is the cursor position
		if i == cursorIdx {
			// Write cursor with special color
			result = append(result, cursorColor...)
			result = append(result, string(heavyFull)...)
			result = append(result, Reset...)
			result = append(result, bgColor...)
		} else if i < filled {
			// Write filled character
			result = append(result, string(heavyFull)...)
		} else {
			// Write empty character
			result = append(result, string(heavyEmpty)...)
		}
	}
	// Return completed progress bar
	return string(result)
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
	// Select character set
	if style == StyleBraille {
		chars = brailleChars
	}

	// Calculate filled portion
	totalSteps := progressBarWidth * stepsPerChar
	filledSteps := progress.Percent * totalSteps / percentMax

	var result [progressBarWidth]rune
	// Build progress bar
	for i := range progressBarWidth {
		charSteps := filledSteps - (i * stepsPerChar)
		// Determine character
		switch {
		// Fully filled character position
		case charSteps >= stepsPerChar:
			result[i] = chars[stepsPerChar]
		// Partially filled character position
		case charSteps > 0:
			result[i] = chars[charSteps]
		// Empty character position
		default:
			result[i] = chars[0]
		}
	}
	// Return completed progress bar
	return string(result[:])
}
