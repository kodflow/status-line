// Package port defines domain interfaces.
package port

import "github.com/florent/status-line/internal/domain/model"

// InputProvider provides input data for status line generation.
// It abstracts the source of input data.
type InputProvider interface {
	// ModelInfo returns the AI model information.
	ModelInfo() model.ModelInfo
	// WorkingDir returns the current working directory.
	WorkingDir() string
	// Progress returns context usage progress (fallback when API unavailable).
	Progress() model.Progress
}
