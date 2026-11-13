//go:build linux || darwin

// Package updater provides self-update functionality for the status-line binary.
package updater

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

const (
	// accessWrite is W_OK for access(2), which package syscall does not name.
	accessWrite uint32 = 0x2
	// lockPerm is the permission of the lock file.
	lockPerm os.FileMode = 0o600
)

// withLock runs fn under the updater's lock beside the binary, without
// waiting for it.
//
// Every check and apply runs under it, so two detached children can never
// download or rename at the same time; the one that loses simply returns.
// Closing the file releases the lock, even when fn panics.
//
// Params:
//   - dir: directory holding the binary
//   - fn: the work to serialise
//
// Returns:
//   - error: the lock is held elsewhere or could not be opened, or fn's error
func withLock(dir string, fn func() error) (err error) {
	file, err := os.OpenFile(filepath.Join(dir, lockFileName), os.O_CREATE|os.O_RDWR, lockPerm)
	// Without a lock file there is no way to serialise: refuse
	if err != nil {
		return fmt.Errorf("%w: %w", errLockUnusable, err)
	}
	defer func() {
		// Releasing the lock cannot fail in a way worth more than fn's error
		err = errors.Join(err, file.Close())
	}()
	// Non-blocking: a busy lock means another child is already at work
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return fmt.Errorf("taking lock: %w", err)
	}
	return fn()
}

// dirWritable says whether this process may create files in a directory.
//
// Params:
//   - dir: directory holding the binary
//
// Returns:
//   - bool: true when access(2) grants write permission
func dirWritable(dir string) bool {
	// No directory resolved: nothing can be written
	if dir == "" {
		return false
	}
	return syscall.Access(dir, accessWrite) == nil
}
