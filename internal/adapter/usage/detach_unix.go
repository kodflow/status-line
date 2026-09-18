//go:build !windows

// Package usage provides the Anthropic API usage adapter.
package usage

import "syscall"

// detachedAttr detaches the refresh process from the current session so it
// survives the exit of the status line process that spawned it.
//
// Returns:
//   - *syscall.SysProcAttr: attributes placing the child in its own session
func detachedAttr() *syscall.SysProcAttr {
	// Setsid gives the child its own session and detaches it from the terminal
	return &syscall.SysProcAttr{Setsid: true}
}
