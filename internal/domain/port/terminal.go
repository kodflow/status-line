// Package port defines domain interfaces (contracts).
package port

import "github.com/florent/status-line/internal/domain/model"

// TerminalProvider defines the interface for terminal information.
type TerminalProvider interface {
	Info() model.TerminalInfo
}
