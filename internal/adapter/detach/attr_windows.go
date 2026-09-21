//go:build windows

package detach

import "syscall"

// attr detaches the refresh process from the current console so it
// survives the exit of the status line process that spawned it.
//
// Returns:
//   - *syscall.SysProcAttr: attributes hiding the child console window
func attr() *syscall.SysProcAttr {
	// CREATE_NO_WINDOW keeps the refresh invisible to the user
	return &syscall.SysProcAttr{HideWindow: true}
}
