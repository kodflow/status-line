// Package port defines domain interfaces (contracts).
package port

import "github.com/florent/status-line/internal/domain/model"

// TerminalProvider defines the interface for terminal information.
// Implementations should retrieve terminal dimensions.
type TerminalProvider interface {
	// Info returns the current terminal information.
	//
	// Returns:
	//   - model.TerminalInfo: terminal dimensions
	Info() model.TerminalInfo
}
