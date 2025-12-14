// Package port defines domain interfaces (contracts).
package port

import "github.com/florent/status-line/internal/domain/model"

// GitRepository defines the interface for git operations.
type GitRepository interface {
	Status() model.GitStatus
}
