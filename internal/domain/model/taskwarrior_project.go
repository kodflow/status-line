// Package model contains domain entities and value objects.
package model

// TaskwarriorStatus represents the status of an epic or task.
type TaskwarriorStatus string

// Task status constants.
const (
	// StatusTodo represents a pending task.
	StatusTodo TaskwarriorStatus = "TODO"
	// StatusWip represents a task in progress.
	StatusWip TaskwarriorStatus = "WIP"
	// StatusDone represents a completed task.
	StatusDone TaskwarriorStatus = "DONE"
)

// TaskwarriorMode represents the workflow mode.
type TaskwarriorMode string

// Workflow mode constants.
const (
	// ModePlan is the planning mode (no code edits).
	ModePlan TaskwarriorMode = "plan"
	// ModeBypass is the execution mode (code edits allowed).
	ModeBypass TaskwarriorMode = "bypass"
)

// TaskwarriorProject represents a project/epic with its task stats.
// It tracks pending and completed task counts for a single project.
// Supports both legacy mode (simple counts) and session mode (Epic/Task workflow).
type TaskwarriorProject struct {
	Name        string
	Pending     int
	Completed   int
	Mode        TaskwarriorMode   // Workflow mode (plan/bypass)
	Epics       []TaskwarriorEpic // Epic hierarchy with tasks
	CurrentEpic int               // Currently active epic number
	CurrentTask string            // Currently active task ID (e.g., "2.1")
}

// TaskwarriorEpic represents an epic containing multiple tasks.
type TaskwarriorEpic struct {
	ID         int
	Name       string
	Status     TaskwarriorStatus
	Tasks      []TaskwarriorTask
	DoneCount  int
	TotalCount int
}

// TaskwarriorTask represents a single task within an epic.
type TaskwarriorTask struct {
	ID       string // e.g., "1.1", "2.3"
	Name     string
	Status   TaskwarriorStatus
	Parallel bool
}

// Total returns total tasks in the project.
//
// Returns:
//   - int: pending + completed tasks
func (p TaskwarriorProject) Total() int {
	// If we have epics, calculate from them
	if len(p.Epics) > 0 {
		total := 0
		for _, epic := range p.Epics {
			total += epic.TotalCount
		}
		return total
	}
	// Sum pending and completed for legacy mode
	return p.Pending + p.Completed
}

// TotalDone returns completed tasks in the project.
//
// Returns:
//   - int: completed task count
func (p TaskwarriorProject) TotalDone() int {
	// If we have epics, calculate from them
	if len(p.Epics) > 0 {
		done := 0
		for _, epic := range p.Epics {
			done += epic.DoneCount
		}
		return done
	}
	// Return completed for legacy mode
	return p.Completed
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
	// Calculate percentage from done tasks
	return p.TotalDone() * percentMultiplier / total
}

// HasSession returns true if project has Epic/Task session data.
//
// Returns:
//   - bool: true if epics are defined
func (p TaskwarriorProject) HasSession() bool {
	// Check if epics exist
	return len(p.Epics) > 0
}

// IsPlanMode returns true if project is in planning mode.
//
// Returns:
//   - bool: true if mode is plan
func (p TaskwarriorProject) IsPlanMode() bool {
	// Check mode
	return p.Mode == ModePlan
}

// CurrentTaskName returns the name of the current WIP task.
//
// Returns:
//   - string: task name or empty string
func (p TaskwarriorProject) CurrentTaskName() string {
	// Find current task in epics
	for _, epic := range p.Epics {
		for _, task := range epic.Tasks {
			if task.Status == StatusWip {
				return task.Name
			}
		}
	}
	return ""
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
