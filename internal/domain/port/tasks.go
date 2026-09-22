// Package port defines domain interfaces (contracts).
package port

import "github.com/florent/status-line/internal/domain/model"

// TasksProvider reads the session's task board: the main agent's open epics
// and the running subagents.
type TasksProvider interface {
	// Board returns the open epics in display order, each with the subagents
	// started for it, and the count of subagents tied to none shown.
	Board() model.TaskBoard
}

// ActivityProvider reads whether the host is working on the session.
type ActivityProvider interface {
	// Working reports whether the session is busy on a turn right now.
	Working() bool
}
