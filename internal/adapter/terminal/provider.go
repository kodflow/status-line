// Package terminal provides the terminal information adapter.
package terminal

import (
	"os"
	"strconv"
	"strings"

	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

const (
	// DefaultWidth is the width assumed when COLUMNS is absent or invalid.
	DefaultWidth int = 120
	// columnsEnv is the variable the host sets to the width it gives the
	// status line command.
	columnsEnv string = "COLUMNS"
	// maxWidth bounds a plausible width; anything larger is a typo.
	maxWidth int = 10000
)

// Compile-time interface implementation check.
var _ port.TerminalProvider = (*Provider)(nil)

// Provider implements port.TerminalProvider from the environment.
//
// The status line command runs with its output piped to the host, so its
// own stdout is no terminal and /dev/tty would report the whole window
// rather than the room the host gives the line. The host says that room in
// COLUMNS; nothing is executed and no device is opened.
type Provider struct {
	getenv func(string) string
}

// NewProvider creates a new terminal provider adapter.
//
// Returns:
//   - *Provider: provider reading the process environment
func NewProvider() *Provider {
	// Read the real environment
	return &Provider{getenv: os.Getenv}
}

// Info returns the current terminal information.
//
// Returns:
//   - model.TerminalInfo: terminal dimensions
func (p *Provider) Info() model.TerminalInfo {
	getenv := p.getenv
	// A zero provider still reads the real environment
	if getenv == nil {
		getenv = os.Getenv
	}
	return model.TerminalInfo{Width: ParseWidth(getenv(columnsEnv))}
}

// ParseWidth reads a COLUMNS value.
//
// Params:
//   - value: raw COLUMNS value
//
// Returns:
//   - int: the width, DefaultWidth when the value is absent, not a
//     positive integer, or implausibly large
func ParseWidth(value string) int {
	width, err := strconv.Atoi(strings.TrimSpace(value))
	// Anything but a plausible positive count falls back to the default
	if err != nil || width <= 0 || width > maxWidth {
		return DefaultWidth
	}
	return width
}
