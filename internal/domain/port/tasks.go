// Package port defines domain interfaces (contracts).
package port

import "github.com/florent/status-line/internal/domain/model"

// TasksProvider reads the session task list.
type TasksProvider interface {
	// Tasks returns the current list, empty when the session has none.
	Tasks() model.TaskList
}
