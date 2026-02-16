// Package port defines domain interfaces (contracts).
package port

import "github.com/florent/status-line/internal/domain/model"

// UsageProvider defines the interface for API usage information.
// Implementations should fetch usage data from Anthropic API.
type UsageProvider interface {
	// Usage returns session (5h) and weekly (7d) API usage.
	//
	// Returns:
	//   - model.UsageData: session and weekly utilization and reset times
	//   - error: any error during fetch
	Usage() (model.UsageData, error)
}
