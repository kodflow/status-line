// Package model contains domain entities and value objects.
package model

import "time"

// Window durations for the rate limit buckets exposed by Claude Code.
const (
	// SessionWindow is the rolling 5-hour rate limit window.
	SessionWindow time.Duration = 5 * time.Hour
	// WeeklyWindow is the rolling 7-day rate limit window.
	WeeklyWindow time.Duration = 7 * 24 * time.Hour
)

// LimitKind identifies which quota a Limit measures.
// Plans expose different subsets, so every kind is optional at runtime.
type LimitKind string

// Limit kinds, ordered from the most local to the most global quota.
const (
	// KindContext is the context window of the current conversation.
	KindContext LimitKind = "context"
	// KindSession is the rolling 5-hour usage quota.
	KindSession LimitKind = "session"
	// KindWeekly is the rolling 7-day usage quota, all models combined.
	KindWeekly LimitKind = "weekly"
	// KindScoped is a 7-day quota restricted to one model family.
	KindScoped LimitKind = "weekly_scoped"
	// KindExtra is the monthly extra-usage credit balance.
	KindExtra LimitKind = "extra"
)

// LimitSource records where a Limit was read from.
// stdin is authoritative and free; api costs a request and may be stale.
type LimitSource string

// Limit sources, ordered by decreasing authority.
const (
	// SourceStdin is the payload Claude Code pipes in on every render.
	SourceStdin LimitSource = "stdin"
	// SourceAPI is the Anthropic OAuth usage endpoint.
	SourceAPI LimitSource = "api"
	// SourceNone marks an absent limit.
	SourceNone LimitSource = ""
)

// Limit is a single quota with its consumption and its refill schedule.
// A zero Limit means "this plan does not expose that quota" — which is a
// legitimate state, not an error, and must never be rendered as 0%.
type Limit struct {
	Kind     LimitKind
	Label    string
	Percent  int
	ResetsAt time.Time
	Window   time.Duration
	Active   bool
	Source   LimitSource
}

// NewLimit builds a Limit with its percentage clamped to 0-100.
//
// Params:
//   - kind: which quota this measures
//   - label: short human label used by the renderer
//   - percent: consumption percentage as reported upstream
//   - resetsAt: instant at which the quota refills
//   - window: total duration of the quota window
//   - source: where the values were read from
//
// Returns:
//   - Limit: normalised limit value object
func NewLimit(kind LimitKind, label string, percent int, resetsAt time.Time, window time.Duration, source LimitSource) Limit {
	// Clamp below the floor so a negative upstream value cannot invert the bar
	if percent < 0 {
		percent = 0
	}
	// Clamp above the ceiling so an overshoot renders as full, not overflowing
	if percent > maxPercent {
		percent = maxPercent
	}
	return Limit{
		Kind:     kind,
		Label:    label,
		Percent:  percent,
		ResetsAt: resetsAt,
		Window:   window,
		Active:   true,
		Source:   source,
	}
}

// IsValid reports whether the limit carries usable data.
// The context window has no reset instant, so it qualifies on source alone.
//
// Returns:
//   - bool: true when the limit can be rendered
func (l Limit) IsValid() bool {
	// A limit with no source was never populated
	if l.Source == SourceNone {
		return false
	}
	// The context window is timeless: having a source is enough
	if l.Kind == KindContext {
		return true
	}
	// Every other kind needs a reset instant to compute pace
	return !l.ResetsAt.IsZero()
}

// HasWindow reports whether pace arithmetic is meaningful for this limit.
//
// Returns:
//   - bool: true when both a window and a reset instant are known
func (l Limit) HasWindow() bool {
	// Pace needs a non-zero window and a known reset instant
	return l.Window > 0 && !l.ResetsAt.IsZero()
}

// Remaining returns the time left before the quota refills.
//
// Returns:
//   - time.Duration: time until reset, never negative
func (l Limit) Remaining() time.Duration {
	// A limit without a reset instant never expires
	if l.ResetsAt.IsZero() {
		return 0
	}
	remaining := time.Until(l.ResetsAt)
	// Clamp a reset instant already in the past to zero
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Elapsed returns how far into the window we are, as a 0-1 fraction.
//
// Returns:
//   - float64: fraction of the window already spent
func (l Limit) Elapsed() float64 {
	// Without a window there is no notion of elapsed time
	if !l.HasWindow() {
		return 0
	}
	remaining := l.Remaining()
	// A reset in the past means the window is fully spent
	if remaining <= 0 {
		return 1
	}
	// More than a full window left means the window has not started
	if remaining >= l.Window {
		return 0
	}
	return float64(l.Window-remaining) / float64(l.Window)
}

// CursorPosition returns the consumption a perfectly even burn would show now.
// It is the reference the actual percentage is judged against.
//
// Returns:
//   - int: expected consumption percentage (0-100)
func (l Limit) CursorPosition() int {
	// Return the break-even percentage for the elapsed fraction
	return int(l.Elapsed() * float64(maxPercent))
}

// Pace returns how far ahead or behind an even burn the consumption is.
// Negative means consuming slower than the clock, positive means faster.
//
// Returns:
//   - int: signed percentage points away from the even-burn reference
func (l Limit) Pace() int {
	// Compare actual consumption against the even-burn reference
	return l.Percent - l.CursorPosition()
}

// IsOnTrack reports whether the current burn rate lasts until the reset.
//
// Returns:
//   - bool: true when consumption is at or below the even-burn reference
func (l Limit) IsOnTrack() bool {
	// Consumption at or under the reference is sustainable
	return l.Pace() <= 0
}

// Projected extrapolates consumption at the end of the window from the
// current burn rate. Above 100 the quota runs out before it refills.
//
// Returns:
//   - int: projected end-of-window percentage, 0 when not computable
func (l Limit) Projected() int {
	elapsed := l.Elapsed()
	// Too early in the window for an extrapolation to mean anything
	if elapsed <= 0 {
		return 0
	}
	return int(float64(l.Percent) / elapsed)
}

// ExhaustsIn returns the time left before the quota is fully consumed at the
// current burn rate. It reports false when the quota outlasts its window.
//
// Returns:
//   - time.Duration: time until exhaustion
//   - bool: true when exhaustion happens before the reset
func (l Limit) ExhaustsIn() (time.Duration, bool) {
	elapsed := l.Elapsed()
	// No burn rate can be derived before the window starts
	if elapsed <= 0 || l.Percent <= 0 {
		return 0, false
	}
	spent := time.Duration(elapsed * float64(l.Window))
	// Time to burn one percentage point at the observed rate
	perPercent := spent / time.Duration(l.Percent)
	left := time.Duration(maxPercent-l.Percent) * perPercent
	// Exhaustion beyond the reset is not a real exhaustion
	if left >= l.Remaining() {
		return 0, false
	}
	return left, true
}

// Progress adapts the limit to the generic progress value object.
//
// Returns:
//   - Progress: progress carrying the limit percentage
func (l Limit) Progress() Progress {
	// Reuse the shared severity thresholds
	return Progress{Percent: l.Percent}
}
