// Package model contains domain entities and value objects.
package model

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

// NewTaskwarriorProject creates a new TaskwarriorProject instance.
//
// Params:
//   - name: project name
//   - pending: pending task count
//   - completed: completed task count
//
// Returns:
//   - TaskwarriorProject: new project value object
func NewTaskwarriorProject(name string, pending, completed int) TaskwarriorProject {
	// Return initialized struct
	return TaskwarriorProject{Name: name, Pending: pending, Completed: completed}
}
