// Package port defines domain interfaces (contracts).
package port

import "github.com/florent/status-line/internal/domain/model"

// UsageProvider defines the interface for API usage information.
// Implementations should fetch usage data from Anthropic API.
type UsageProvider interface {
	// Limits returns every quota the account exposes.
	//
	// Returns:
	//   - model.LimitSet: session, weekly, scoped and extra quotas
	//   - error: any error during fetch
	Limits() (model.LimitSet, error)
}
