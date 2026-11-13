//go:build !(linux || darwin)

// Package updater provides self-update functionality for the status-line binary.
package updater

import "errors"

// errUnsupported says self-replacement is not supported on this platform.
var errUnsupported = errors.New("self-update not supported on this platform")

// withLock always fails where self-replacement is not supported, so no
// check or apply ever runs there.
//
// Params:
//   - dir: directory holding the binary
//   - fn: never run
//
// Returns:
//   - error: always errUnsupported
func withLock(_ string, _ func() error) error {
	// Nothing to lock where nothing is replaced
	return errUnsupported
}

// dirWritable reports false where self-replacement is not supported, so a
// frame never spawns a check there.
//
// Params:
//   - dir: directory holding the binary
//
// Returns:
//   - bool: always false
func dirWritable(_ string) bool {
	// Never spawn where nothing can be replaced
	return false
}
