// Package model contains domain entities.
package model

// Task statuses as written by the task tools.
const (
	// TaskPending is a task not started yet.
	TaskPending string = "pending"
	// TaskInProgress is the task being worked on.
	TaskInProgress string = "in_progress"
	// TaskCompleted is a finished task.
	TaskCompleted string = "completed"
)

// TaskItem is one entry of the session task list.
type TaskItem struct {
	ID      string
	Subject string
	Status  string
}

// TaskList is the session task list, in creation order.
type TaskList struct {
	Items []TaskItem
}

// Total returns how many tasks the list holds.
//
// Returns:
//   - int: number of tasks
func (l TaskList) Total() int {
	// Every item counts, whatever its status
	return len(l.Items)
}

// Done returns how many tasks are completed.
//
// Returns:
//   - int: number of completed tasks
func (l TaskList) Done() int {
	done := 0
	// Count the finished items
	for _, item := range l.Items {
		if item.Status == TaskCompleted {
			done++
		}
	}
	return done
}

// Active returns how many tasks are in progress.
//
// Returns:
//   - int: number of started, unfinished tasks
func (l TaskList) Active() int {
	active := 0
	// Count the started items
	for _, item := range l.Items {
		if item.Status == TaskInProgress {
			active++
		}
	}
	return active
}

// Current returns the subject of the task being worked on.
//
// Returns:
//   - string: first in-progress subject, empty when none is started
func (l TaskList) Current() string {
	// The earliest started task is the one in front
	for _, item := range l.Items {
		if item.Status == TaskInProgress {
			return item.Subject
		}
	}
	return ""
}

// IsActive reports whether the list is worth showing: it exists and is not
// finished. A completed list is history, not work in progress.
//
// Returns:
//   - bool: true while at least one task remains open
func (l TaskList) IsActive() bool {
	// Something is still open
	return l.Total() > 0 && l.Done() < l.Total()
}
