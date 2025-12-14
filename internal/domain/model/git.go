// Package model contains domain entities and value objects.
package model

// GitStatus represents the current state of a git repository.
// It contains branch information and change counts.
type GitStatus struct {
	Branch    string
	Modified  int
	Untracked int
}

// IsInRepo returns true if currently inside a git repository.
//
// Returns:
//   - bool: true if branch is set
func (s GitStatus) IsInRepo() bool {
	// Return true if branch name is not empty
	return s.Branch != ""
}
