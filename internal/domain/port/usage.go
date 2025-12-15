// Package port defines domain interfaces (contracts).
package port

import "github.com/florent/status-line/internal/domain/model"

// UsageProvider defines the interface for API usage information.
// Implementations should fetch usage data from Anthropic API.
type UsageProvider interface {
	// Usage returns the current weekly API usage.
	//
	// Returns:
	//   - model.Usage: utilization and reset time
	//   - error: any error during fetch
	Usage() (model.Usage, error)
}
