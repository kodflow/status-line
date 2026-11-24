// Package updater provides self-update functionality for the status-line binary.
package updater

import (
	"path/filepath"
	"time"
)

// Config is what a detached check or apply works on. Everything it touches is
// named here, so tests run the children against a throwaway directory.
type Config struct {
	// ExecPath is the binary to replace, resolved through its symlinks; the
	// state and the lock live in its directory.
	ExecPath string
	// Version is the compiled version, empty for a dev build.
	Version string
	// Now is the clock, time.Now when nil.
	Now func() time.Time
	// Force runs a check even before next_attempt_at, for a manual
	// --check-update=force; frames never set it.
	Force bool
}

// dir is the directory holding the binary, its state and its lock.
//
// Returns:
//   - string: directory of ExecPath
func (c Config) dir() string {
	// Everything lives beside the binary it updates
	return filepath.Dir(c.ExecPath)
}

// now reads the configured clock.
//
// Returns:
//   - time.Time: current time
func (c Config) now() time.Time {
	// Production leaves the clock unset
	if c.Now == nil {
		return time.Now()
	}
	return c.Now()
}
