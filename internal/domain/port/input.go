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
	// Progress returns context window usage.
	Progress() model.Progress
	// StdinLimits returns the quotas Claude Code piped in on stdin.
	StdinLimits() model.LimitSet
	// EffortLevel returns the reasoning effort of the session.
	EffortLevel() string
	// SessionCost returns the accumulated session cost in USD.
	SessionCost() float64
	// IsFastMode reports whether fast mode is enabled.
	IsFastMode() bool
	// SessionLabel returns the human name of the session.
	SessionLabel() string
}
