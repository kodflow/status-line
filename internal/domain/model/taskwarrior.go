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

// TaskwarriorProject represents a project/epic with its task stats.
// It tracks pending and completed task counts for a single project.
type TaskwarriorProject struct {
	Name      string
	Pending   int
	Completed int
}

// Total returns total tasks in the project.
//
// Returns:
//   - int: pending + completed tasks
func (p TaskwarriorProject) Total() int {
	// Sum pending and completed
	return p.Pending + p.Completed
}

// Percent returns completion percentage.
//
// Returns:
//   - int: percentage 0-100
func (p TaskwarriorProject) Percent() int {
	total := p.Total()
	// Avoid division by zero
	if total == 0 {
		// Return zero for empty projects
		return 0
	}
	// Calculate percentage
	return p.Completed * percentMultiplier / total
}

// HasProjects returns true if there are any projects.
//
// Returns:
//   - bool: true if projects exist
func (t TaskwarriorInfo) HasProjects() bool {
	// Check if any projects
	return len(t.Projects) > 0
}
