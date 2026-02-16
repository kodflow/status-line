// Package usage provides the Anthropic API usage adapter.
package usage

// usageResponse represents the API response structure.
// It contains both five-hour session and seven-day weekly usage periods.
type usageResponse struct {
	FiveHour usagePeriod `json:"five_hour"`
	SevenDay usagePeriod `json:"seven_day"`
}

// usagePeriod represents a usage period from API.
// It contains utilization percentage and reset timestamp.
type usagePeriod struct {
	Utilization float64 `json:"utilization"`
	ResetsAt    string  `json:"resets_at"`
}
