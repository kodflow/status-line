// Package port defines domain interfaces (contracts).
package port

import "github.com/florent/status-line/internal/domain/model"

// GitRepository defines the interface for git operations.
// Implementations should retrieve git status information.
type GitRepository interface {
	// Status retrieves the current git repository status.
	//
	// Returns:
	//   - model.GitStatus: branch and change information
	Status() model.GitStatus
}
