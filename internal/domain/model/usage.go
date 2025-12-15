// Package model contains domain entities and value objects.
package model

import "time"

// Duration constants for burn rate calculation.
const (
	// weekDuration is the duration of one week for burn rate calculation.
	weekDuration time.Duration = 7 * 24 * time.Hour
)

// Usage represents the weekly API usage from Anthropic.
// It contains utilization percentage and reset time for burn rate calculation.
type Usage struct {
	Utilization int
	ResetsAt    time.Time
}

// NewUsage creates a Usage with the given values.
//
// Params:
//   - utilization: usage percentage (0-100)
//   - resetsAt: time when usage resets
//
// Returns:
//   - Usage: configured usage value object
func NewUsage(utilization int, resetsAt time.Time) Usage {
	// Cap utilization to valid range
	if utilization < 0 {
		utilization = 0
	}
	if utilization > maxPercent {
		utilization = maxPercent
	}
	// Return configured usage
	return Usage{
		Utilization: utilization,
		ResetsAt:    resetsAt,
	}
}

// CursorPosition calculates where the burn rate cursor should be.
// The cursor represents expected usage based on time elapsed in the week.
//
// Returns:
//   - int: cursor position as percentage (0-100)
func (u Usage) CursorPosition() int {
	// Calculate time remaining until reset
	remaining := time.Until(u.ResetsAt)
	// Handle edge cases
	if remaining <= 0 {
		// Reset already passed, cursor at 100%
		return maxPercent
	}
	if remaining >= weekDuration {
		// More than a week remaining, cursor at 0%
		return 0
	}
	// Calculate elapsed time in current week
	elapsed := weekDuration - remaining
	// Calculate cursor position as percentage
	position := int(elapsed * time.Duration(maxPercent) / weekDuration)
	// Return calculated position
	return position
}

// IsOnTrack returns true if utilization is at or below expected usage.
//
// Returns:
//   - bool: true if consumption is sustainable
func (u Usage) IsOnTrack() bool {
	// Compare utilization to cursor position
	return u.Utilization <= u.CursorPosition()
}

// Progress returns a Progress value from the utilization.
//
// Returns:
//   - Progress: progress based on utilization percentage
func (u Usage) Progress() Progress {
	// Return progress with utilization as percent
	return Progress{Percent: u.Utilization}
}

// IsValid returns true if usage data was successfully fetched.
// Zero time indicates the data was not retrieved from API.
//
// Returns:
//   - bool: true if usage data is valid
func (u Usage) IsValid() bool {
	// Check if reset time is set (non-zero)
	return !u.ResetsAt.IsZero()
}
