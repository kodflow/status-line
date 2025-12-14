// Package terminal provides the terminal information adapter.
package terminal

import (
	"os"

	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
	"golang.org/x/term"
)

const (
	// defaultWidth is the fallback terminal width.
	defaultWidth int = 120
	// ttyPath is the path to the TTY device.
	ttyPath string = "/dev/tty"
)

// Compile-time interface implementation check.
var _ port.TerminalProvider = (*Provider)(nil)

// Provider implements port.TerminalProvider using terminal detection.
// It retrieves terminal dimensions from the system.
type Provider struct{}

// NewProvider creates a new terminal provider adapter.
//
// Returns:
//   - *Provider: new provider instance
func NewProvider() *Provider {
	// Return empty struct as no state is needed
	return &Provider{}
}

// Info returns the current terminal information.
//
// Returns:
//   - model.TerminalInfo: terminal dimensions
func (p *Provider) Info() model.TerminalInfo {
	// Return terminal info with width
	return model.TerminalInfo{
		Width: p.getWidth(),
	}
}

// getWidth returns the terminal width in columns.
//
// Returns:
//   - int: terminal width or default if unavailable
func (p *Provider) getWidth() int {
	tty, err := os.Open(ttyPath)
	// Check if tty is available
	if err == nil {
		defer tty.Close()
		// Try to get width from tty
		if width, _, err := term.GetSize(int(tty.Fd())); err == nil && width > 0 {
			// Return tty width
			return width
		}
	}

	// Fallback: try stdout
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
		// Return stdout width
		return width
	}

	// Return default width as last resort
	return defaultWidth
}
