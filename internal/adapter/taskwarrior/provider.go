// Package taskwarrior provides the Taskwarrior adapter.
package taskwarrior

import (
	"os/exec"
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

	// Get all projects with stats
	projects := p.getProjects()

	// Return task information
	return model.TaskwarriorInfo{
		Installed: true,
		Projects:  projects,
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
		return projects[i].Name < projects[j].Name
	})

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
	// Return empty if command fails
	if err != nil {
		return nil
	}

	// Parse project names
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	projects := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines
		if line != "" {
			projects = append(projects, line)
		}
	}

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
	// Find last dot
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		return name[idx+1:]
	}
	// Return full name if no dot
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
	// Return zero if command fails
	if err != nil {
		return 0
	}

	// Parse count from output
	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	// Return zero if parsing fails
	if err != nil {
		return 0
	}

	// Return task count
	return count
}
