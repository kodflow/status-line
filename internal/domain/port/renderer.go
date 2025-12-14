// Package port defines domain interfaces (contracts).
package port

import "github.com/florent/status-line/internal/domain/model"

// Renderer defines the interface for rendering the status line.
type Renderer interface {
	Render(data model.StatusLineData) string
}
