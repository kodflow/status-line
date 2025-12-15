// Package port defines domain interfaces (contracts).
package port

import "github.com/florent/status-line/internal/domain/model"

// TaskwarriorProvider defines the interface for reading Taskwarrior data.
// Implementations should detect installation and read task counts.
type TaskwarriorProvider interface {
	// Info returns Taskwarrior task information.
	//
	// Returns:
	//   - model.TaskwarriorInfo: task counts and installation status
	Info() model.TaskwarriorInfo
}
