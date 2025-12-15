// Package port defines domain interfaces (contracts).
package port

import "github.com/florent/status-line/internal/domain/model"

// Renderer defines the interface for rendering the status line.
// Implementations should produce a formatted string for terminal output.
type Renderer interface {
	// Render generates the status line string from the provided data.
	//
	// Params:
	//   - data: all information needed for rendering
	//
	// Returns:
	//   - string: formatted status line with ANSI codes
	Render(data model.StatusLineData) string
}
