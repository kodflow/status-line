// Package model contains domain entities and value objects.
package model

import "time"

// Duration constants for burn rate calculation.
const (
	// sessionDuration is the 5-hour rate limit window from Anthropic API.
	sessionDuration time.Duration = 5 * time.Hour
	// weekDuration is the duration of one week for burn rate calculation.
	weekDuration time.Duration = 7 * 24 * time.Hour
)

// Usage represents API usage from Anthropic for a specific time window.
// It contains utilization percentage, reset time, and window duration
// for burn rate calculation. Used for both session (5h) and weekly (7d).
type Usage struct {
	Utilization    int
	ResetsAt       time.Time
	WindowDuration time.Duration
}

// UsageData holds both session and weekly usage from the Anthropic API.
type UsageData struct {
	Session Usage
	Weekly  Usage
}

// NewUsage creates a Usage with the given values and window duration.
//
// Params:
//   - utilization: usage percentage (0-100)
//   - resetsAt: time when usage resets
//   - windowDuration: total window duration for cursor calculation
//
// Returns:
//   - Usage: configured usage value object
func NewUsage(utilization int, resetsAt time.Time, windowDuration time.Duration) Usage {
	// Check for negative utilization
	if utilization < 0 {
		utilization = 0
	}
	// Check for utilization exceeding maximum
	if utilization > maxPercent {
		utilization = maxPercent
	}
	return Usage{
		Utilization:    utilization,
		ResetsAt:       resetsAt,
		WindowDuration: windowDuration,
	}
}

// NewSessionUsage creates a Usage for the 5-hour session window.
func NewSessionUsage(utilization int, resetsAt time.Time) Usage {
	return NewUsage(utilization, resetsAt, sessionDuration)
}

// NewWeeklyUsage creates a Usage for the 7-day weekly window.
func NewWeeklyUsage(utilization int, resetsAt time.Time) Usage {
	return NewUsage(utilization, resetsAt, weekDuration)
}

// CursorPosition calculates where the burn rate cursor should be.
// The cursor represents expected usage based on time elapsed in the window.
//
// Returns:
//   - int: cursor position as percentage (0-100)
func (u Usage) CursorPosition() int {
	// Guard against missing window duration
	if u.WindowDuration <= 0 {
		return 0
	}
	// Calculate time remaining until reset
	remaining := time.Until(u.ResetsAt)
	// Check if reset already passed
	if remaining <= 0 {
		return maxPercent
	}
	// Check if more than a full window remaining
	if remaining >= u.WindowDuration {
		return 0
	}
	// Calculate elapsed time in current window
	elapsed := u.WindowDuration - remaining
	// Calculate cursor position as percentage
	position := int(elapsed * time.Duration(maxPercent) / u.WindowDuration)
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
