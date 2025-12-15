// Package updater provides self-update functionality for the status-line binary.
package updater

// releaseInfo represents GitHub release API response.
// Contains the tag name used for version comparison.
type releaseInfo struct {
	TagName string `json:"tag_name"`
}
