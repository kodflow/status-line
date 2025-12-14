// Package port defines domain interfaces (contracts).
package port

import "github.com/florent/status-line/internal/domain/model"

// SystemProvider defines the interface for system information.
// Implementations should detect OS type and environment.
type SystemProvider interface {
	// Info returns the current system information.
	//
	// Returns:
	//   - model.SystemInfo: OS type and Docker status
	Info() model.SystemInfo
}
