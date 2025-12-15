// Package renderer provides status line rendering.
package renderer

import (
	"strings"

	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

// Numeric constants for base conversion and progress.
const (
	// base10 is the decimal base for integer to string conversion.
	base10 int = 10
	// percentComplete represents 100% completion.
	percentComplete int = 100
)

// Compile-time interface implementation check.
var _ port.Renderer = (*Powerline)(nil)

// Powerline implements port.Renderer with powerline style.
// It renders a status bar with segments and rounded corners.
type Powerline struct{}

// NewPowerline creates a new powerline renderer.
//
// Returns:
//   - *Powerline: new renderer instance
func NewPowerline() *Powerline {
	// Return empty struct as no state is needed
	return &Powerline{}
}

// Render generates the status line string.
//
// Params:
//   - data: all information needed for rendering
//
// Returns:
//   - string: formatted status line with ANSI codes
func (r *Powerline) Render(data model.StatusLineData) string {
	var sb strings.Builder

	// Render first line with system info
	r.renderLine1(&sb, data)
	sb.WriteString("\n")
	// Render second line with model pill
	r.renderLine2(&sb, data)
	sb.WriteString("\n")

	// Return complete status line
	return sb.String()
}

// renderLine1 renders the first line with OS, Model, Path, Git, and Changes segments.
//
// Params:
//   - sb: string builder to write to
//   - data: status line data
func (r *Powerline) renderLine1(sb *strings.Builder, data model.StatusLineData) {
	// Get model background color for OS segment transition (use FullName for color detection)
	modelBg, _, _ := GetModelColors(data.Model.FullName())

	// Render OS segment (transitions to Model segment)
	r.renderOSSegment(sb, data.System, data.Icons.OS, modelBg)

	// Render Model segment (transitions to Path segment)
	r.renderModelSegment(sb, data.Model, data.Icons.Model, data.Progress, BgBlue)

	// Determine what follows git segment (or path if no git)
	changesNextBg := ""
	// Determine next segment background color based on changes
	if data.Changes.HasAdded() {
		// Use green background for added lines
		changesNextBg = BgGreen
		// Check if only removed changes exist
	} else if data.Changes.HasRemoved() {
		// Use red background for removed lines only
		changesNextBg = BgRed
	}

	// Render path segment (pass changesNextBg for case when no git)
	r.renderPathSegment(sb, data.Dir, data.Git.IsInRepo(), data.Icons.Path, changesNextBg)

	// Render git segment if in repo
	r.renderGitSegment(sb, data.Git, data.Icons.Git, changesNextBg)
	// Render code changes if any
	r.renderChangesSegment(sb, data.Changes)
}

// renderLine2 renders the second line with dynamic pills (Taskwarrior, MCP).
//
// Params:
//   - sb: string builder to write to
//   - data: status line data
func (r *Powerline) renderLine2(sb *strings.Builder, data model.StatusLineData) {
	// Track if we've rendered anything
	hasContent := false

	// Render Taskwarrior pill if installed and has projects
	if data.Taskwarrior.Installed && data.Taskwarrior.HasProjects() {
		r.renderTaskwarriorPill(sb, data.Taskwarrior)
		hasContent = true
	}

	// Render MCP server pills if any
	if len(data.MCP) > 0 {
		// Add separator space if previous content exists
		if hasContent {
			sb.WriteString(" ")
		}
		r.renderMCPPills(sb, data.MCP)
	}
}

// renderOSSegment renders the operating system segment.
//
// Params:
//   - sb: string builder to write to
//   - sys: system information
//   - showIcon: whether to show the OS icon
//   - nextBg: background color of the next segment
func (r *Powerline) renderOSSegment(sb *strings.Builder, sys model.SystemInfo, showIcon bool, nextBg string) {
	// Write left rounded cap
	sb.WriteString(FgWhite + LeftRound + Reset)
	// Check if icon should be shown
	if showIcon {
		icon := GetOSIcon(sys.OS, sys.IsDocker)
		// Write icon with background
		sb.WriteString(BgWhite + FgBlack + Bold + " " + icon + " " + Reset)
	} else {
		// Write empty space without icon
		sb.WriteString(BgWhite + FgBlack + Bold + "  " + Reset)
	}
	// Write separator to next segment
	sb.WriteString(nextBg + FgWhite + SepRight + Reset)
}

// renderModelSegment renders the AI model segment with integrated progress bar.
//
// Params:
//   - sb: string builder to write to
//   - m: model information
//   - showIcon: whether to show the model icon
//   - progress: context window usage progress
//   - nextBg: background color of the next segment
func (r *Powerline) renderModelSegment(sb *strings.Builder, m model.ModelInfo, showIcon bool, progress model.Progress, nextBg string) {
	// Use FullName for color detection (includes version like "Opus 4.5")
	fullName := m.FullName()
	bgColor, fgColor, textColor := GetModelColors(fullName)
	bar := RenderProgressBar(progress, StyleHeavy)

	// Check if icon should be shown
	if showIcon {
		// Write model name with icon
		sb.WriteString(bgColor + textColor + Bold + " " + IconModel + " " + m.Name + " " + Reset)
	} else {
		// Write model name without icon
		sb.WriteString(bgColor + textColor + Bold + " " + m.Name + " " + Reset)
	}

	// Write progress bar and percentage (using model's text color with bold for consistency)
	sb.WriteString(bgColor + textColor + Bold + bar + " " + itoa(progress.Percent) + "% " + Reset)

	// Write separator to next segment
	sb.WriteString(nextBg + fgColor + SepRight + Reset)
}

// renderPathSegment renders the current directory segment.
//
// Params:
//   - sb: string builder to write to
//   - dir: directory path
//   - hasGit: whether git segment follows
//   - showIcon: whether to show the folder icon
//   - nextBg: background color of next segment if no git
func (r *Powerline) renderPathSegment(sb *strings.Builder, dir string, hasGit bool, showIcon bool, nextBg string) {
	truncated := TruncatePath(dir, defaultMaxPath)
	// Check if icon should be shown
	if showIcon {
		// Write path with folder icon and dark blue text
		sb.WriteString(BgBlue + FgBlueDark + Bold + " " + IconFolder + " " + truncated + " " + Reset)
	} else {
		// Write path without icon with dark blue text
		sb.WriteString(BgBlue + FgBlueDark + Bold + " " + truncated + " " + Reset)
	}

	// Determine separator style based on next segment
	if hasGit {
		// Write separator to git segment
		sb.WriteString(BgCyan + FgBlue + SepRight + Reset)
		// Check if next segment has a colored background
	} else if nextBg != "" {
		// Write separator to next colored segment
		sb.WriteString(nextBg + FgBlue + SepRight + Reset)
		// No following segment
	} else {
		// Write final separator
		sb.WriteString(FgBlue + SepRight + Reset)
	}
}

// renderGitSegment renders the git branch and status segment.
//
// Params:
//   - sb: string builder to write to
//   - git: git status information
//   - showIcon: whether to show the git branch icon
//   - nextBg: background color of next segment for separator
func (r *Powerline) renderGitSegment(sb *strings.Builder, git model.GitStatus, showIcon bool, nextBg string) {
	// Skip if not in a git repository
	if !git.IsInRepo() {
		// Return early if not in repo
		return
	}

	// Check if icon should be shown
	if showIcon {
		// Write branch with icon and dark cyan text
		sb.WriteString(BgCyan + FgCyanDark + Bold + " " + IconGitBranch + " " + git.Branch)
	} else {
		// Write branch without icon with dark cyan text
		sb.WriteString(BgCyan + FgCyanDark + Bold + " " + git.Branch)
	}

	// Add modified indicator if present
	if git.Modified > 0 {
		sb.WriteString(" !" + itoa(git.Modified))
	}
	// Add untracked indicator if present
	if git.Untracked > 0 {
		sb.WriteString(" ?" + itoa(git.Untracked))
	}

	// Write segment end with appropriate separator
	sb.WriteString(" " + Reset)
	// Check if next segment has background
	if nextBg != "" {
		// Write separator to next colored segment
		sb.WriteString(nextBg + FgCyan + SepRight + Reset)
	} else {
		// Write final separator
		sb.WriteString(FgCyan + SepRight + Reset)
	}
}

// renderChangesSegment renders the lines added/removed as powerline segments.
//
// Params:
//   - sb: string builder to write to
//   - changes: code changes information
func (r *Powerline) renderChangesSegment(sb *strings.Builder, changes model.CodeChanges) {
	// Skip if no changes
	if !changes.HasChanges() {
		// Return early if nothing to show
		return
	}

	// Render added segment if any
	if changes.HasAdded() {
		// Write added segment with dark green text on pale green background
		sb.WriteString(BgGreen + FgGreenText + Bold + " +" + itoa(changes.Added) + " " + Reset)

		// Determine separator destination
		if changes.HasRemoved() {
			// Separator to red segment
			sb.WriteString(BgRed + FgGreenSep + SepRight + Reset)
		} else {
			// Final separator
			sb.WriteString(FgGreenSep + SepRight + Reset)
		}
	}

	// Render removed segment if any
	if changes.HasRemoved() {
		// Write removed segment with dark red text on pale red background
		sb.WriteString(BgRed + FgRedText + Bold + " -" + itoa(changes.Removed) + " " + Reset)
		// Final separator
		sb.WriteString(FgRedSep + SepRight + Reset)
	}
}

// renderMCPPills renders MCP server pills.
//
// Params:
//   - sb: string builder to write to
//   - servers: list of MCP servers
func (r *Powerline) renderMCPPills(sb *strings.Builder, servers model.MCPServers) {
	// Skip if no servers
	if len(servers) == 0 {
		// Return early if nothing to show
		return
	}

	// Add space before MCP pills
	sb.WriteString(" ")

	// Render each server as a pill
	for idx, server := range servers {
		// Add space between pills
		if idx > 0 {
			sb.WriteString(" ")
		}
		// Render individual MCP pill
		r.renderMCPPill(sb, server)
	}
}

// renderMCPPill renders a single MCP server pill.
//
// Params:
//   - sb: string builder to write to
//   - server: MCP server information
func (r *Powerline) renderMCPPill(sb *strings.Builder, server model.MCPServer) {
	var bgColor, fgColor, textColor string

	// Select colors based on enabled status
	if server.Enabled {
		// Use enabled colors (pale bg, dark text)
		bgColor = BgMCPEnabled
		fgColor = FgMCPEnabled
		textColor = FgMCPEnabledText
	} else {
		// Use disabled gray colors (pale bg, dark text)
		bgColor = BgMCPDisabled
		fgColor = FgMCPDisabled
		textColor = FgMCPDisabledText
	}

	// Write left rounded cap
	sb.WriteString(fgColor + LeftRound + Reset)
	// Write server name
	sb.WriteString(bgColor + textColor + " " + server.Name + " " + Reset)
	// Write right rounded cap
	sb.WriteString(fgColor + RightRound + Reset)
}

// renderTaskwarriorPill renders Taskwarrior project pills.
//
// Params:
//   - sb: string builder to write to
//   - tw: Taskwarrior information
func (r *Powerline) renderTaskwarriorPill(sb *strings.Builder, tw model.TaskwarriorInfo) {
	// Skip if Taskwarrior is not installed or no projects
	if !tw.Installed || !tw.HasProjects() {
		// Return early if not installed or empty
		return
	}

	// Render each project as a pill
	for _, project := range tw.Projects {
		r.renderTaskwarriorProjectPill(sb, project)
	}
}

// renderTaskwarriorProjectPill renders a single project pill with progress.
//
// Params:
//   - sb: string builder to write to
//   - project: project information
func (r *Powerline) renderTaskwarriorProjectPill(sb *strings.Builder, project model.TaskwarriorProject) {
	// Add space before pill
	sb.WriteString(" ")

	// Create progress for the project
	progress := model.NewProgress(project.Completed, project.Total())
	// Use gray for incomplete, lavender for 100%
	progressColor := ColorGray
	// Check if project is complete
	if progress.Percent == percentComplete {
		// Use themed color for completed projects
		progressColor = FgTaskwarriorText
	}
	bar := RenderProgressBar(progress, StyleHeavy)

	// Write left rounded cap
	sb.WriteString(FgTaskwarrior + LeftRound + Reset)
	// Write icon and project name
	sb.WriteString(BgTaskwarrior + FgTaskwarriorText + Bold + " " + IconTaskwarrior + " " + project.Name + " " + Reset)

	// Write nested progress bar pill
	sb.WriteString(BgTaskwarrior + FgWhite + LeftRound + Reset)
	// Progress bar on white background with progress color
	sb.WriteString(BgWhite + progressColor + " " + bar + " " + Reset)
	// Add task count (completed/total)
	sb.WriteString(BgWhite + FgBlack + Bold + itoa(project.Completed) + "/" + itoa(project.Total()) + " " + Reset)
	// Write right cap
	sb.WriteString(FgWhite + RightRound + Reset)
}

// itoa converts an integer to string.
//
// Params:
//   - n: integer to convert
//
// Returns:
//   - string: string representation of the integer
func itoa(n int) string {
	// Handle zero case
	if n == 0 {
		// Return zero string
		return "0"
	}
	var digits []byte
	// Extract digits in reverse order
	for n > 0 {
		digits = append([]byte{byte('0' + n%base10)}, digits...)
		n /= base10
	}
	// Return converted string
	return string(digits)
}
