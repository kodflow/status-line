// Package model contains domain entities and value objects.
package model

// GitStatus represents the current state of a git repository.
type GitStatus struct {
	Branch    string
	Modified  int
	Untracked int
}

// IsInRepo returns true if currently inside a git repository.
func (s GitStatus) IsInRepo() bool {
	return s.Branch != ""
}
