// Package taskwarrior provides the Taskwarrior adapter.
package taskwarrior

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

// Command constants.
const (
	// taskBinary is the Taskwarrior binary name.
	taskBinary string = "task"
	// sessionDir is the directory containing session files.
	sessionDir string = "/workspace/.claude/sessions"
	// maxTaskNameLen is the maximum length for task name display.
	maxTaskNameLen int = 15
)

// Compile-time interface implementation check.
var _ port.TaskwarriorProvider = (*Provider)(nil)

// Provider implements port.TaskwarriorProvider by running task commands.
// It detects Taskwarrior installation and reads project stats.
type Provider struct{}

// NewProvider creates a new Taskwarrior provider adapter.
//
// Returns:
//   - *Provider: new provider instance
func NewProvider() *Provider {
	// Return empty struct as no state is needed
	return &Provider{}
}

// Info returns Taskwarrior task information.
//
// Returns:
//   - model.TaskwarriorInfo: project stats and installation status
func (p *Provider) Info() model.TaskwarriorInfo {
	// Check if task binary exists
	if !p.isInstalled() {
		// Return empty info if not installed
		return model.TaskwarriorInfo{Installed: false}
	}

	// Try to find active session first
	activeProject := p.findActiveSession()

	// Get all legacy projects with stats
	projects := p.getProjects()

	// Return task information
	return model.TaskwarriorInfo{
		Installed:     true,
		ActiveProject: activeProject,
		Projects:      projects,
	}
}

// getProjects returns all projects with their task counts.
//
// Returns:
//   - []model.TaskwarriorProject: list of projects with stats
func (p *Provider) getProjects() []model.TaskwarriorProject {
	// Get list of project names
	projectNames := p.listProjects()
	projects := make([]model.TaskwarriorProject, 0, len(projectNames))

	// Get stats for each project
	for _, name := range projectNames {
		pending := p.countTasks("project:"+name, "status:pending")
		// Skip projects with no pending tasks
		if pending == 0 {
			continue
		}
		completed := p.countTasks("project:"+name, "status:completed")

		projects = append(projects, model.TaskwarriorProject{
			Name:      p.shortProjectName(name),
			Pending:   pending,
			Completed: completed,
		})
	}

	// Sort by name for consistent display
	sort.Slice(projects, func(i, j int) bool {
		// Compare names alphabetically
		return projects[i].Name < projects[j].Name
	})

	// Return sorted project list
	return projects
}

// listProjects returns all unique project names.
//
// Returns:
//   - []string: list of project names
func (p *Provider) listProjects() []string {
	// Run task _unique project command
	cmd := exec.Command(taskBinary, "rc.confirmation=off", "rc.verbose=nothing", "_unique", "project")
	output, err := cmd.Output()
	// Check if command succeeded
	if err != nil {
		// Return empty slice if command fails
		return []string{}
	}

	// Parse project names
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	projects := make([]string, 0, len(lines))
	// Iterate through output lines
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines
		if line != "" {
			projects = append(projects, line)
		}
	}

	// Return parsed project names
	return projects
}

// shortProjectName extracts the last part of a project name.
//
// Params:
//   - name: full project name like "Epic1.Auth"
//
// Returns:
//   - string: short name like "Auth"
func (p *Provider) shortProjectName(name string) string {
	// Find last dot separator in hierarchical name
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		// Return substring after last dot
		return name[idx+1:]
	}
	// Return full name if no hierarchy
	return name
}

// isInstalled checks if Taskwarrior is installed.
//
// Returns:
//   - bool: true if task binary is in PATH
func (p *Provider) isInstalled() bool {
	_, err := exec.LookPath(taskBinary)
	// Return true if binary found
	return err == nil
}

// countTasks returns the count of tasks matching filters.
//
// Params:
//   - filters: Taskwarrior filter expressions
//
// Returns:
//   - int: number of matching tasks
func (p *Provider) countTasks(filters ...string) int {
	// Build command arguments
	args := []string{"rc.confirmation=off", "rc.verbose=nothing"}
	args = append(args, filters...)
	args = append(args, "count")

	// Run task count command with filters
	cmd := exec.Command(taskBinary, args...)
	output, err := cmd.Output()
	// Check if command succeeded
	if err != nil {
		// Return zero if command fails
		return 0
	}

	// Parse count from output
	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	// Check if parsing succeeded
	if err != nil {
		// Return zero if parsing fails
		return 0
	}

	// Return task count
	return count
}

// sessionJSON represents the JSON structure of a session file.
type sessionJSON struct {
	Project     string          `json:"project"`
	Mode        string          `json:"mode"`
	CurrentEpic int             `json:"current_epic"`
	CurrentTask string          `json:"current_task"`
	Epics       []sessionEpic   `json:"epics"`
}

// sessionEpic represents an epic in the session JSON.
type sessionEpic struct {
	ID     int           `json:"id"`
	Name   string        `json:"name"`
	Status string        `json:"status"`
	Tasks  []sessionTask `json:"tasks"`
}

// sessionTask represents a task in the session JSON.
type sessionTask struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Parallel bool   `json:"parallel"`
}

// findActiveSession finds and parses the most recent session file.
//
// Returns:
//   - *model.TaskwarriorProject: active project or nil
func (p *Provider) findActiveSession() *model.TaskwarriorProject {
	// Check if session directory exists
	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		// Return nil if directory not found
		return nil
	}

	// Find the most recent .json file
	var newestFile string
	var newestTime int64
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Unix() > newestTime {
			newestTime = info.ModTime().Unix()
			newestFile = entry.Name()
		}
	}

	// Return nil if no session file found
	if newestFile == "" {
		return nil
	}

	// Read and parse session file
	return p.parseSessionFile(filepath.Join(sessionDir, newestFile))
}

// parseSessionFile reads and parses a session JSON file.
//
// Params:
//   - path: path to the session file
//
// Returns:
//   - *model.TaskwarriorProject: parsed project or nil
func (p *Provider) parseSessionFile(path string) *model.TaskwarriorProject {
	// Read file content
	data, err := os.ReadFile(path)
	if err != nil {
		// Return nil if file read fails
		return nil
	}

	// Parse JSON
	var session sessionJSON
	if err := json.Unmarshal(data, &session); err != nil {
		// Return nil if JSON parsing fails
		return nil
	}

	// Convert to domain model
	return p.convertSession(&session)
}

// convertSession converts a sessionJSON to a TaskwarriorProject.
//
// Params:
//   - session: parsed session JSON
//
// Returns:
//   - *model.TaskwarriorProject: converted project
func (p *Provider) convertSession(session *sessionJSON) *model.TaskwarriorProject {
	// Convert mode
	mode := model.ModeBypass
	if session.Mode == "plan" {
		mode = model.ModePlan
	}

	// Convert epics
	epics := make([]model.TaskwarriorEpic, 0, len(session.Epics))
	for _, e := range session.Epics {
		epic := p.convertEpic(&e)
		epics = append(epics, epic)
	}

	// Build project
	return &model.TaskwarriorProject{
		Name:        session.Project,
		Mode:        mode,
		Epics:       epics,
		CurrentEpic: session.CurrentEpic,
		CurrentTask: session.CurrentTask,
	}
}

// convertEpic converts a sessionEpic to a TaskwarriorEpic.
//
// Params:
//   - epic: session epic data
//
// Returns:
//   - model.TaskwarriorEpic: converted epic
func (p *Provider) convertEpic(epic *sessionEpic) model.TaskwarriorEpic {
	// Convert status
	status := p.convertStatus(epic.Status)

	// Convert tasks and count done
	tasks := make([]model.TaskwarriorTask, 0, len(epic.Tasks))
	doneCount := 0
	for _, t := range epic.Tasks {
		task := p.convertTask(&t)
		tasks = append(tasks, task)
		if task.Status == model.StatusDone {
			doneCount++
		}
	}

	// Build epic
	return model.TaskwarriorEpic{
		ID:         epic.ID,
		Name:       epic.Name,
		Status:     status,
		Tasks:      tasks,
		DoneCount:  doneCount,
		TotalCount: len(tasks),
	}
}

// convertTask converts a sessionTask to a TaskwarriorTask.
//
// Params:
//   - task: session task data
//
// Returns:
//   - model.TaskwarriorTask: converted task
func (p *Provider) convertTask(task *sessionTask) model.TaskwarriorTask {
	// Truncate name if too long
	name := task.Name
	if len(name) > maxTaskNameLen {
		name = name[:maxTaskNameLen-1] + "…"
	}

	// Build task
	return model.TaskwarriorTask{
		ID:       task.ID,
		Name:     name,
		Status:   p.convertStatus(task.Status),
		Parallel: task.Parallel,
	}
}

// convertStatus converts a string status to TaskwarriorStatus.
//
// Params:
//   - status: status string
//
// Returns:
//   - model.TaskwarriorStatus: converted status
func (p *Provider) convertStatus(status string) model.TaskwarriorStatus {
	// Map string to status enum
	switch strings.ToUpper(status) {
	case "DONE":
		return model.StatusDone
	case "WIP":
		return model.StatusWip
	default:
		return model.StatusTodo
	}
}
