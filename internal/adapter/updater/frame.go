// Package updater provides self-update functionality for the status-line binary.
package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/florent/status-line/internal/domain/model"
)

const (
	// checkFailureGrace is how long a failing check stays quiet before the
	// pill shows it: one blip offline is not worth a warning.
	checkFailureGrace time.Duration = time.Hour
)

// Executable resolves the running binary through its symlinks: the state, the
// lock and the replacement all live beside the real file, never beside a link.
//
// Returns:
//   - string: path of the real binary
//   - error: path resolution error
func Executable() (string, error) {
	self, err := os.Executable()
	// Without a path there is nothing to update
	if err != nil {
		return "", fmt.Errorf("getting executable path: %w", err)
	}
	real, err := filepath.EvalSymlinks(self)
	// A dangling link leaves nothing to update either
	if err != nil {
		return "", fmt.Errorf("resolving symlinks: %w", err)
	}
	return real, nil
}

// Allowed says whether a frame may start a check at all.
//
// Params:
//   - version: compiled version, empty for a dev build
//   - dir: directory holding the binary
//
// Returns:
//   - bool: false under the opt-out, for a dev build, or when the binary's
//     directory is not writable
func Allowed(version, dir string) bool {
	// A child that could not replace anything is not worth starting
	return enabled(version) && dirWritable(dir)
}

// Due says whether the next check may start.
//
// Params:
//   - now: current time
//
// Returns:
//   - bool: true once now reaches NextAttemptAt
func (s *State) Due(now time.Time) bool {
	// A zero date, from a missing or broken file, is always due
	return !now.Before(s.NextAttemptAt)
}

// Pill derives the update pill from the cached state.
//
// It is lit while the cached tag is newer than the compiled version, so it
// goes out by itself once the new binary renders. Its failure variant shows a
// failed apply, or a check failing for more than an hour. The opt-out hides it.
//
// Params:
//   - version: compiled version
//   - now: current time
//
// Returns:
//   - model.UpdateInfo: what the pill shows
func (s *State) Pill(version string, now time.Time) model.UpdateInfo {
	// Switched off: no pill at all
	if os.Getenv(DisableEnv) != "" || version == "" {
		return model.UpdateInfo{}
	}
	available := newer(s.Tag, version)
	failed := (available && s.ApplyError != "") ||
		(s.CheckError != "" && !s.CheckFailingSince.IsZero() && now.Sub(s.CheckFailingSince) > checkFailureGrace)
	info := model.UpdateInfo{Available: available || failed, Failed: failed}
	// Name the release only when one is actually pending
	if available {
		info.Version = s.Tag
	}
	return info
}
