// Package tasks reads the session task list written by the task tools.
package tasks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

// Task list location constants.
const (
	// listIDEnv overrides the list id, as it does for the task tools.
	listIDEnv string = "CLAUDE_CODE_TASK_LIST_ID"
	// configDirEnv relocates the whole configuration directory.
	configDirEnv string = "CLAUDE_CONFIG_DIR"
	// sessionPrefix and sessionIDLen form the default list id: the first
	// characters of the session id.
	sessionPrefix string = "session-"
	sessionIDLen  int    = 8
)

// unsafeID matches the characters the task tools replace in a list id.
var unsafeID = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

// Compile-time interface implementation check.
var _ port.TasksProvider = (*Provider)(nil)

// Provider reads one task list directory.
type Provider struct {
	dir string
}

// taskFile is the part of a task file this adapter reads.
type taskFile struct {
	ID      string `json:"id"`
	Subject string `json:"subject"`
	Status  string `json:"status"`
}

// NewProvider locates the task list of a session.
//
// The task tools keep one JSON file per task under
// <config>/tasks/<list id>/. The list id is CLAUDE_CODE_TASK_LIST_ID when
// set, and otherwise derives from the session id.
//
// Params:
//   - sessionID: session identifier from stdin, may be empty
//
// Returns:
//   - *Provider: provider for that list, reading nothing when it has no id
func NewProvider(sessionID string) *Provider {
	listID := os.Getenv(listIDEnv)
	// Fall back to the per-session list
	if listID == "" && sessionID != "" {
		listID = sessionPrefix + sessionID[:min(sessionIDLen, len(sessionID))]
	}
	// Without a list id there is no list to read
	if listID == "" {
		return &Provider{}
	}
	base := os.Getenv(configDirEnv)
	// Default to the per-user configuration directory
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return &Provider{}
		}
		base = filepath.Join(home, ".claude")
	}
	return &Provider{dir: filepath.Join(base, "tasks", unsafeID.ReplaceAllString(listID, "-"))}
}

// Tasks returns the task list, in creation order.
//
// Returns:
//   - model.TaskList: current list, empty when there is none
func (p *Provider) Tasks() model.TaskList {
	// No list located, nothing to read
	if p.dir == "" {
		return model.TaskList{}
	}
	paths, err := filepath.Glob(filepath.Join(p.dir, "*.json"))
	// An unreadable directory is an empty list
	if err != nil {
		return model.TaskList{}
	}
	items := make([]model.TaskItem, 0, len(paths))
	// Read each task, skipping any that is mid-write or malformed
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var task taskFile
		if err := json.Unmarshal(data, &task); err != nil || task.ID == "" {
			continue
		}
		items = append(items, model.TaskItem{ID: task.ID, Subject: task.Subject, Status: task.Status})
	}
	// Ids are sequence numbers: order them numerically, not as text
	sort.SliceStable(items, func(i, j int) bool {
		a, errA := strconv.Atoi(items[i].ID)
		b, errB := strconv.Atoi(items[j].ID)
		if errA != nil || errB != nil {
			return items[i].ID < items[j].ID
		}
		return a < b
	})
	return model.TaskList{Items: items}
}
