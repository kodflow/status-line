// Package terminal provides the terminal information adapter.
package terminal

import (
	"os"

	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
	"golang.org/x/term"
)

const (
	defaultWidth = 120
	ttyPath      = "/dev/tty"
)

var _ port.TerminalProvider = (*Provider)(nil)

// Provider implements port.TerminalProvider.
type Provider struct{}

// NewProvider creates a new terminal provider adapter.
func NewProvider() *Provider {
	return &Provider{}
}

// Info returns the current terminal information.
func (p *Provider) Info() model.TerminalInfo {
	return model.TerminalInfo{
		Width: p.getWidth(),
	}
}

func (p *Provider) getWidth() int {
	tty, err := os.Open(ttyPath)
	if err == nil {
		defer tty.Close()
		if width, _, err := term.GetSize(int(tty.Fd())); err == nil && width > 0 {
			return width
		}
	}

	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
		return width
	}

	return defaultWidth
}
