// Package model contains domain entities and value objects.
package model

import "time"

// Effort levels reported by Claude Code.
const (
	// EffortHigh is the highest reasoning effort.
	EffortHigh string = "high"
	// EffortMedium is the default reasoning effort.
	EffortMedium string = "medium"
	// EffortLow is the cheapest reasoning effort.
	EffortLow string = "low"
)

// InputEffort is the reasoning effort of the running session.
type InputEffort struct {
	Level string `json:"level"`
}

// InputThinking reports whether extended thinking is on.
type InputThinking struct {
	Enabled bool `json:"enabled"`
}

// InputOutputStyle names the active output style.
type InputOutputStyle struct {
	Name string `json:"name"`
}

// InputRateLimits is the quota block Claude Code pipes on stdin.
// Claude Code >= 2.1.140 fills it; a plan without a weekly cap simply omits
// SevenDay, so a nil pointer means "no such quota", never "quota at zero".
type InputRateLimits struct {
	FiveHour *InputRateLimit `json:"five_hour"`
	SevenDay *InputRateLimit `json:"seven_day"`
}

// InputRateLimit is one quota bucket from stdin.
type InputRateLimit struct {
	UsedPercentage *float64  `json:"used_percentage"`
	Utilization    *float64  `json:"utilization"`
	ResetsAt       Timestamp `json:"resets_at"`
}

// Percent returns the consumption of the bucket.
//
// Returns:
//   - int: consumption percentage
//   - bool: false when neither field was present
func (r *InputRateLimit) Percent() (int, bool) {
	// A nil bucket is an absent quota
	if r == nil {
		return 0, false
	}
	// Prefer the field Claude Code actually sends
	if r.UsedPercentage != nil {
		return int(*r.UsedPercentage), true
	}
	// Fall back to the API spelling of the same value
	if r.Utilization != nil {
		return int(*r.Utilization), true
	}
	return 0, false
}

// Limit converts the bucket into a domain limit.
//
// Params:
//   - kind: quota kind to tag the limit with
//   - label: short human label
//   - window: total window duration
//
// Returns:
//   - Limit: populated limit
//   - bool: false when the bucket carries no percentage
func (r *InputRateLimit) Limit(kind LimitKind, label string, window time.Duration) (Limit, bool) {
	percent, ok := r.Percent()
	// An absent percentage means the plan does not expose this quota
	if !ok {
		return Limit{}, false
	}
	return NewLimit(kind, label, percent, r.ResetsAt.Time, window, SourceStdin), true
}
