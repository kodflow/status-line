// Package port defines domain interfaces (contracts).
package port

import "github.com/florent/status-line/internal/domain/model"

// SystemProvider defines the interface for system information.
type SystemProvider interface {
	Info() model.SystemInfo
}
