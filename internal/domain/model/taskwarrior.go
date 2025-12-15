// Package model contains domain entities and value objects.
package model

// Percentage calculation constant.
const (
	// percentMultiplier is used for percentage calculation.
	percentMultiplier int = 100
)

// TaskwarriorInfo contains Taskwarrior task information.
// It holds project-level statistics.
type TaskwarriorInfo struct {
	Installed bool
	Projects  []TaskwarriorProject
}

// HasProjects returns true if there are any projects.
//
// Returns:
//   - bool: true if projects exist
func (t TaskwarriorInfo) HasProjects() bool {
	// Check if any projects
	return len(t.Projects) > 0
}
