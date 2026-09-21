// Package port defines domain interfaces (contracts).
package port

import "github.com/florent/status-line/internal/domain/model"

// HealthProvider reports the aggregate state of Claude's public services.
type HealthProvider interface {
	// Health returns the current level, HealthUnknown when none is known.
	Health() model.ServiceHealth
}
