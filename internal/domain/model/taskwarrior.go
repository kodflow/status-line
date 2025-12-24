// Package model contains domain entities and value objects.
package model

// Percentage calculation constant.
const (
	// percentMultiplier is used for percentage calculation.
	percentMultiplier int = 100
)

// TaskwarriorInfo contains Taskwarrior task information.
// It holds project-level statistics and active session data.
type TaskwarriorInfo struct {
	Installed     bool
	ActiveProject *TaskwarriorProject  // Project with active session (Epic/Task workflow)
	Projects      []TaskwarriorProject // Legacy projects (simple pending/completed)
}

// HasProjects returns true if there are any projects.
//
// Returns:
//   - bool: true if projects exist
func (t TaskwarriorInfo) HasProjects() bool {
	// Check if active project or legacy projects exist
	return t.ActiveProject != nil || len(t.Projects) > 0
}

// HasActiveSession returns true if there is an active session.
//
// Returns:
//   - bool: true if active session exists
func (t TaskwarriorInfo) HasActiveSession() bool {
	// Check for active project with session
	return t.ActiveProject != nil && t.ActiveProject.HasSession()
}
