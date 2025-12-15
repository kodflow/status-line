// Package updater provides self-update functionality for the status-line binary.
package updater

// UpdateInfo contains information about available updates.
// Used to display update status in the status line.
type UpdateInfo struct {
	Available bool
	Version   string
}
