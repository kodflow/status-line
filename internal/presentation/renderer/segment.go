// Package renderer provides status line rendering.
package renderer

import "github.com/florent/status-line/internal/domain/model"

// CursorProvider defines the interface for burn-rate cursor position.
// It abstracts the cursor position calculation from concrete types.
type CursorProvider interface {
	// CursorPosition returns the cursor position as percentage (0-100).
	CursorPosition() int
	// IsValid returns true if cursor data is available.
	IsValid() bool
}

// ModelSegmentData groups data needed to render the model segment.
// It reduces the number of parameters for renderModelSegment.
type ModelSegmentData struct {
	Model    model.ModelInfo
	ShowIcon bool
	Progress model.Progress
	Cursor   CursorProvider
	NextBg   string
}
