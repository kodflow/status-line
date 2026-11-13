// Package updater provides self-update functionality for the status-line binary.
package updater

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	// stateFileName is the cached outcome of the last check, beside the binary.
	stateFileName string = ".status-line-update.json"
	// stateTempPrefix names the state being written. It starts with a dot,
	// so the sweep of stale downloads (tempPrefix) never matches it.
	stateTempPrefix string = stateFileName + ".tmp-"
	// stateMaxBytes bounds the state read: anything larger is not ours.
	stateMaxBytes int64 = 64 << 10
	// statePerm is the permission of the state file.
	statePerm os.FileMode = 0o644
)

// State is what the detached check and apply leave for the frames to read.
//
// The frame never touches the network: it only reads this file, decides from
// NextAttemptAt whether to start a check, and derives the update pill from it.
type State struct {
	// CheckedAt is the last time the API gave a decisive answer.
	CheckedAt time.Time `json:"checked_at,omitzero"`
	// Tag is the latest release tag that answer named.
	Tag string `json:"tag,omitempty"`
	// NextAttemptAt is the earliest time a frame may start another check.
	NextAttemptAt time.Time `json:"next_attempt_at,omitzero"`
	// CheckError is the reason the last check failed, empty after a success.
	CheckError string `json:"check_error,omitempty"`
	// CheckFailingSince is the first failure after the last success.
	CheckFailingSince time.Time `json:"check_failing_since,omitzero"`
	// Failures counts the checks failed in a row, driving the backoff.
	Failures int `json:"failures,omitempty"`
	// ApplyError is the reason the last apply failed, empty after a success.
	ApplyError string `json:"apply_error,omitempty"`
}

// ReadState reads the cached state beside the binary.
//
// A missing, unreadable or half-written file is a stale state: its zero value
// says a check is due and no update is known.
//
// Params:
//   - dir: directory holding the binary
//
// Returns:
//   - State: the cached state, zero on any failure
func ReadState(dir string) State {
	var st State
	// No directory resolved: nothing to read
	if dir == "" {
		return st
	}
	path := filepath.Join(dir, stateFileName)
	info, err := os.Stat(path)
	// A missing file is a check that never ran; an oversized one is not ours
	if err != nil || info.Size() > stateMaxBytes {
		return st
	}
	raw, err := os.ReadFile(path)
	// An unreadable file counts as stale
	if err != nil {
		return st
	}
	// A malformed file counts as stale
	if err := json.Unmarshal(raw, &st); err != nil {
		return State{}
	}
	return st
}

// writeState replaces the state file atomically: a frame reading it at any
// moment sees either the previous state or the new one, never half of it.
//
// The temporary is named after this process: only the lock holder writes, so
// no two writers can share it, and its leading dot keeps it out of the sweep.
//
// Params:
//   - dir: directory holding the binary
//   - st: state to record
//
// Returns:
//   - error: any write or rename error
func writeState(dir string, st *State) error {
	raw, err := json.Marshal(st)
	// A state that cannot be encoded is not written
	if err != nil {
		return fmt.Errorf("encoding state: %w", err)
	}
	tmpPath := filepath.Join(dir, stateTempPrefix+strconv.Itoa(os.Getpid()))
	// A short write never replaces the previous state
	if err := os.WriteFile(tmpPath, raw, statePerm); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("writing state: %w", err)
	}
	// Swap it in whole
	if err := os.Rename(tmpPath, filepath.Join(dir, stateFileName)); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replacing state: %w", err)
	}
	return nil
}
