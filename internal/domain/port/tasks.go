// Package port defines domain interfaces (contracts).
package port

import "github.com/florent/status-line/internal/domain/model"

// TasksProvider reads the session task list and its running subagents.
type TasksProvider interface {
	// Tasks returns the main agent's list, empty when the session has none.
	Tasks() model.TaskList
	// Subagents returns how many subagents of the session are running.
	Subagents() int
}
