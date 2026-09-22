// Package tasks reads the session task list and the running subagents.
package tasks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

// Task list location constants.
const (
	// listIDEnv overrides the built-in list id, as it does for the task tools.
	listIDEnv string = "CLAUDE_CODE_TASK_LIST_ID"
	// configDirEnv relocates the whole configuration directory.
	configDirEnv string = "CLAUDE_CONFIG_DIR"
	// sessionPrefix and sessionIDLen form the default built-in list id: the
	// first characters of the session id.
	sessionPrefix string = "session-"
	sessionIDLen  int    = 8
	// mainAgent owns the tasks the main conversation creates.
	mainAgent string = "main"
	// staleAgent is how long a subagent may run without a stop event before it
	// is taken for one whose stop was lost with a crashed session.
	staleAgent time.Duration = 12 * time.Hour
)

// unsafeID matches the characters replaced in a list or session directory name.
var unsafeID = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

// Compile-time interface implementation check.
var _ port.TasksProvider = (*Provider)(nil)

// Provider reads the task list of one session from two places.
//
// The kodflow tasks MCP (kodflow-hooks) keeps every agent's tasks in
// <config>/kodflow/sessions/<session>/tasks.json, next to agents.json, the
// running subagents recorded by the SubagentStart/SubagentStop hooks. It is
// preferred: it tells the main agent's tasks apart. The built-in task tools'
// one-file-per-task list under <config>/tasks/<list>/ is the fallback.
type Provider struct {
	sessionDir string
	builtinDir string
	now        func() time.Time
}

// taskFile is one task, in either store.
type taskFile struct {
	ID      string `json:"id"`
	Agent   string `json:"agent"`
	Subject string `json:"subject"`
	Status  string `json:"status"`
	Epic    int    `json:"epic"`
}

// sessionTasks is the tasks MCP file.
type sessionTasks struct {
	Tasks []taskFile `json:"tasks"`
	// Epics holds each agent's current epic; tasks of earlier epics are
	// finished subjects and are not shown.
	Epics map[string]struct {
		ID int `json:"id"`
	} `json:"epics"`
}

// sessionAgents is the running-agents file.
type sessionAgents struct {
	Agents map[string]struct {
		Started int64  `json:"started"`
		Stopped *int64 `json:"stopped"`
	} `json:"agents"`
}

// NewProvider locates the task stores of a session.
//
// Params:
//   - sessionID: session identifier from stdin, may be empty
//
// Returns:
//   - *Provider: provider for that session, reading nothing without an id
func NewProvider(sessionID string) *Provider {
	base := os.Getenv(configDirEnv)
	// Default to the per-user configuration directory
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return &Provider{now: time.Now}
		}
		base = filepath.Join(home, ".claude")
	}
	p := &Provider{now: time.Now}
	// The kodflow stores are keyed by the whole session id
	if sessionID != "" {
		p.sessionDir = filepath.Join(base, "kodflow", "sessions", unsafeID.ReplaceAllString(sessionID, "-"))
	}
	listID := os.Getenv(listIDEnv)
	// The built-in list falls back to a per-session id
	if listID == "" && sessionID != "" {
		listID = sessionPrefix + sessionID[:min(sessionIDLen, len(sessionID))]
	}
	if listID != "" {
		p.builtinDir = filepath.Join(base, "tasks", unsafeID.ReplaceAllString(listID, "-"))
	}
	return p
}

// Tasks returns the main agent's task list, in creation order.
//
// Returns:
//   - model.TaskList: current list, empty when there is none
func (p *Provider) Tasks() model.TaskList {
	// The MCP list wins whenever the main agent has one
	if list := p.mcpTasks(); list.Total() > 0 {
		return list
	}
	return p.builtinTasks()
}

// Subagents returns how many subagents of the session are running.
//
// Returns:
//   - int: subagents started and not stopped, stale ones excluded
func (p *Provider) Subagents() int {
	// No session located, nothing to count
	if p.sessionDir == "" {
		return 0
	}
	data, err := os.ReadFile(filepath.Join(p.sessionDir, "agents.json"))
	// No registry yet means no subagent has started
	if err != nil {
		return 0
	}
	var reg sessionAgents
	// A registry mid-write reads as empty until the next redraw
	if err := json.Unmarshal(data, &reg); err != nil {
		return 0
	}
	cutoff := p.now().Add(-staleAgent).Unix()
	running := 0
	// Count the agents still running, forgetting any whose stop was lost
	for _, agent := range reg.Agents {
		if agent.Stopped == nil && agent.Started >= cutoff {
			running++
		}
	}
	return running
}

// mcpTasks reads the main agent's entries from the tasks MCP file.
//
// Returns:
//   - model.TaskList: the main agent's tasks, empty when the file is absent
func (p *Provider) mcpTasks() model.TaskList {
	// No session located, nothing to read
	if p.sessionDir == "" {
		return model.TaskList{}
	}
	data, err := os.ReadFile(filepath.Join(p.sessionDir, "tasks.json"))
	if err != nil {
		return model.TaskList{}
	}
	var file sessionTasks
	if err := json.Unmarshal(data, &file); err != nil {
		return model.TaskList{}
	}
	items := make([]model.TaskItem, 0, len(file.Tasks))
	// Subagents keep their own lists; only the main agent's is shown
	epic := file.Epics[mainAgent].ID
	for _, task := range file.Tasks {
		if task.Agent != "" && task.Agent != mainAgent {
			continue
		}
		// Only the current epic: an earlier subject is not this one's progress
		if task.Epic != epic {
			continue
		}
		items = append(items, model.TaskItem{ID: task.ID, Subject: task.Subject, Status: task.Status})
	}
	return model.TaskList{Items: sortByID(items)}
}

// builtinTasks reads the built-in tools' one-file-per-task list.
//
// Returns:
//   - model.TaskList: current list, empty when there is none
func (p *Provider) builtinTasks() model.TaskList {
	// No list located, nothing to read
	if p.builtinDir == "" {
		return model.TaskList{}
	}
	paths, err := filepath.Glob(filepath.Join(p.builtinDir, "*.json"))
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
	return model.TaskList{Items: sortByID(items)}
}

// sortByID orders tasks by their numeric id, falling back to text order.
//
// Params:
//   - items: tasks to order in place
//
// Returns:
//   - []model.TaskItem: the same slice, ordered
func sortByID(items []model.TaskItem) []model.TaskItem {
	// Ids are sequence numbers: order them numerically, not as text
	sort.SliceStable(items, func(i, j int) bool {
		a, errA := strconv.Atoi(items[i].ID)
		b, errB := strconv.Atoi(items[j].ID)
		if errA != nil || errB != nil {
			return items[i].ID < items[j].ID
		}
		return a < b
	})
	return items
}
