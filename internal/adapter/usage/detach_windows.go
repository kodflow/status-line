//go:build windows

// Package usage provides the Anthropic API usage adapter.
package usage

import "syscall"

// detachedAttr detaches the refresh process from the current console so it
// survives the exit of the status line process that spawned it.
//
// Returns:
//   - *syscall.SysProcAttr: attributes hiding the child console window
func detachedAttr() *syscall.SysProcAttr {
	// CREATE_NO_WINDOW keeps the refresh invisible to the user
	return &syscall.SysProcAttr{HideWindow: true}
}
