// Package model contains domain entities and value objects.
package model

// Progress threshold and limit constants.
const (
	// thresholdMedium is the percentage threshold for medium usage.
	thresholdMedium int = 50
	// thresholdHigh is the percentage threshold for high usage.
	thresholdHigh int = 75
	// thresholdCritical is the percentage threshold for critical usage.
	thresholdCritical int = 90
	// maxPercent is the maximum percentage value.
	maxPercent int = 100
)

// ProgressLevel represents the severity of context usage.
// It is used to determine the color of the progress indicator.
type ProgressLevel int

// Progress level constants for visual feedback.
const (
	// LevelLow represents safe usage (green).
	LevelLow ProgressLevel = iota
	// LevelMedium represents moderate usage (yellow).
	LevelMedium
	// LevelHigh represents high usage (orange).
	LevelHigh
	// LevelCritical represents critical usage (red).
	LevelCritical
)

// Progress represents context window usage.
// It tracks the percentage of context window capacity used.
type Progress struct {
	Percent int
}

// NewProgress creates a Progress from token counts.
//
// Params:
//   - totalTokens: total tokens used
//   - contextSize: maximum context window size
//
// Returns:
//   - Progress: calculated progress value object
func NewProgress(totalTokens, contextSize int) Progress {
	// Handle zero context size to avoid division by zero
	if contextSize == 0 {
		// Return zero progress for invalid input
		return Progress{Percent: 0}
	}
	// Calculate and cap percentage
	percent := min(totalTokens*maxPercent/contextSize, maxPercent)
	// Return calculated progress
	return Progress{Percent: percent}
}

// Level returns the severity level based on usage percentage.
//
// Returns:
//   - ProgressLevel: the severity level for the current usage
func (p Progress) Level() ProgressLevel {
	// Determine level based on threshold ranges
	switch {
	// Low level for usage below 50%
	case p.Percent < thresholdMedium:
		// Return low level
		return LevelLow
	// Medium level for usage between 50% and 75%
	case p.Percent < thresholdHigh:
		// Return medium level
		return LevelMedium
	// High level for usage between 75% and 90%
	case p.Percent < thresholdCritical:
		// Return high level
		return LevelHigh
	// Critical level for usage above 90%
	default:
		// Return critical level
		return LevelCritical
	}
}
