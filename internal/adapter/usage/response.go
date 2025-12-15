// Package usage provides the Anthropic API usage adapter.
package usage

// usageResponse represents the API response structure.
// It contains the seven-day usage period from Anthropic API.
type usageResponse struct {
	SevenDay usagePeriod `json:"seven_day"`
}

// usagePeriod represents a usage period from API.
// It contains utilization percentage and reset timestamp.
type usagePeriod struct {
	Utilization float64 `json:"utilization"`
	ResetsAt    string  `json:"resets_at"`
}
