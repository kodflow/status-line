// Package model contains domain entities.
package model

// Task statuses as written by the task tools.
const (
	// TaskPending is a task not started yet.
	TaskPending string = "pending"
	// TaskInProgress is the task being worked on.
	TaskInProgress string = "in_progress"
	// TaskWaiting is a task blocked on the user: a decision, an approval.
	TaskWaiting string = "waiting"
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

// Waiting returns how many tasks wait on the user.
//
// Returns:
//   - int: number of tasks blocked on the user
func (l TaskList) Waiting() int {
	waiting := 0
	// Count the items blocked on the user
	for _, item := range l.Items {
		if item.Status == TaskWaiting {
			waiting++
		}
	}
	return waiting
}

// Headline returns the task the line names, and what state it is in.
//
// The line always names one: the task under way, else the first one waiting
// on the user, else the next one to do. A bar with no title leaves the reader
// guessing whether work stopped or is blocked.
//
// Returns:
//   - string: subject to show, empty for an empty or finished list
//   - string: its status (in_progress, waiting or pending)
func (l TaskList) Headline() (string, string) {
	// Prefer, in order, the state that says the most about what happens now
	for _, status := range []string{TaskInProgress, TaskWaiting, TaskPending} {
		for _, item := range l.Items {
			if item.Status == status {
				return item.Subject, status
			}
		}
	}
	return "", ""
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

// NoEpic is the epic id of tasks filed under no epic: those created before the
// first epic, or while none was active.
const NoEpic int = 0

// Epic is one subject the main agent works on: a titled list of its tasks.
type Epic struct {
	// ID identifies the epic; NoEpic gathers the tasks filed under none.
	ID int
	// Title names the epic, empty for NoEpic.
	Title string
	// Active marks the epic the main agent is focused on. When no epic is
	// active, the tasks filed under none are.
	Active bool
	// Tasks holds the epic's tasks, in creation order.
	Tasks TaskList
	// Subagents counts the running subagents started for this epic.
	Subagents int
}

// TaskBoard is what the session is working on: the main agent's open epics,
// in display order, and the running subagents tied to none of them.
type TaskBoard struct {
	// Epics lists the open epics, the active one first.
	Epics []Epic
	// Unattributed counts running subagents whose epic is not shown.
	Unattributed int
}

// IsEmpty reports whether the board has nothing to draw.
//
// Returns:
//   - bool: true when no epic is open and no subagent runs
func (b TaskBoard) IsEmpty() bool {
	// Both halves of the board are empty
	return len(b.Epics) == 0 && b.Unattributed == 0
}
