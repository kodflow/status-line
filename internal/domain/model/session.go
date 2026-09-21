// Package model contains domain entities and value objects.
package model

import "time"

// Effort levels reported by Claude Code, cheapest first.
const (
	// EffortLow is the cheapest reasoning effort.
	EffortLow string = "low"
	// EffortMedium is the default reasoning effort on most models.
	EffortMedium string = "medium"
	// EffortHigh is a raised reasoning effort.
	EffortHigh string = "high"
	// EffortXHigh is the effort above high.
	EffortXHigh string = "xhigh"
	// EffortMax is the highest reasoning effort.
	EffortMax string = "max"
)

// effortScale lists the known levels in order. Claude Code only reports the
// current level's name, never the scale, so the scale is pinned here from the
// documentation; a level it does not list is reported as unknown.
var effortScale = []string{EffortLow, EffortMedium, EffortHigh, EffortXHigh, EffortMax}

// EffortSteps returns how many levels the known scale has.
//
// Returns:
//   - int: number of known effort levels
func EffortSteps() int {
	// The scale length is the gauge length
	return len(effortScale)
}

// EffortRank returns the position of a level on the known scale.
//
// Params:
//   - level: effort level as reported
//
// Returns:
//   - int: 1 for the cheapest level up to EffortSteps for the highest
//   - bool: false when the level is not on the known scale
func EffortRank(level string) (int, bool) {
	// Find the level on the scale
	for idx, known := range effortScale {
		if known == level {
			return idx + 1, true
		}
	}
	return 0, false
}

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
