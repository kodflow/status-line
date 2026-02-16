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

// renderLine1 renders the first line with OS, Model, Weekly, Path, Git, and Changes segments.
//
// Params:
//   - sb: string builder to write to
//   - data: status line data
func (r *Powerline) renderLine1(sb *strings.Builder, data model.StatusLineData) {
	// Get model background color for OS segment transition (use FullName for color detection)
	modelBg, _, _ := GetModelColors(data.Model.FullName())

	// Render OS segment (transitions to Model segment)
	r.renderOSSegment(sb, data.System, data.Icons.OS, modelBg)

	// Determine what follows the model segment
	hasWeekly := data.Usage.IsValid()
	modelNextBg := BgBlue
	if hasWeekly {
		modelNextBg = BgWeekly
	}

	// Render Model segment with session context progress (no cursor)
	modelData := &ModelSegmentData{
		Model:    data.Model,
		ShowIcon: data.Icons.Model,
		Progress: data.Progress,
		Cursor:   nil,
		NextBg:   modelNextBg,
	}
	r.renderModelSegment(sb, modelData)

	// Render weekly usage segment if API data is available (auto-hide)
	if hasWeekly {
		r.renderWeeklySegment(sb, data.Usage)
	}

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

// renderLine2 renders the second line with dynamic pills (Taskwarrior, MCP, Update).
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
		hasContent = true
	}

	// Render update pill if update is available
	if data.Update.Available {
		// Add separator space if previous content exists
		if hasContent {
			sb.WriteString(" ")
		}
		r.renderUpdatePill(sb, data.Update)
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
//   - data: model segment rendering data
func (r *Powerline) renderModelSegment(sb *strings.Builder, data *ModelSegmentData) {
	// Use FullName for color detection (includes version like "Opus 4.5")
	fullName := data.Model.FullName()
	bgColor, fgColor, textColor := GetModelColors(fullName)

	// Render progress bar (with cursor if usage data is valid)
	var bar string
	// Check if we have valid cursor data
	if data.Cursor != nil && data.Cursor.IsValid() {
		// Render with burn-rate cursor
		bar = RenderProgressBarWithCursor(data.Progress, data.Cursor.CursorPosition(), FgCursorOrange, bgColor+textColor+Bold)
	} else {
		// Render without cursor (no API data available)
		bar = RenderProgressBar(data.Progress, StyleHeavy)
	}

	// Check if icon should be shown
	if data.ShowIcon {
		// Write model name with icon
		sb.WriteString(bgColor + textColor + Bold + " " + IconModel + " " + data.Model.Name + " " + Reset)
	} else {
		// Write model name without icon
		sb.WriteString(bgColor + textColor + Bold + " " + data.Model.Name + " " + Reset)
	}

	// Write progress bar and percentage (using model's text color with bold for consistency)
	sb.WriteString(bgColor + textColor + Bold + bar + " " + itoa(data.Progress.Percent) + "% " + Reset)

	// Write separator to next segment
	sb.WriteString(data.NextBg + fgColor + SepRight + Reset)
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

// renderWeeklySegment renders the weekly API usage segment with burn-rate cursor.
// Auto-hidden when Usage.IsValid() is false (no API credentials).
//
// Params:
//   - sb: string builder to write to
//   - usage: weekly usage data
func (r *Powerline) renderWeeklySegment(sb *strings.Builder, usage model.Usage) {
	progress := usage.Progress()
	bar := RenderProgressBarWithCursor(progress, usage.CursorPosition(), FgCursorOrange, BgWeekly+FgWeeklyText+Bold)

	// Write segment content
	sb.WriteString(BgWeekly + FgWeeklyText + Bold + " " + IconWeekly + " " + bar + " " + itoa(progress.Percent) + "% " + Reset)

	// Write separator to path segment
	sb.WriteString(BgBlue + FgWeekly + SepRight + Reset)
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

	// Render active project with session (segmented bar)
	if tw.ActiveProject != nil && tw.ActiveProject.HasSession() {
		r.renderTaskwarriorSessionPill(sb, tw.ActiveProject)
		return
	}

	// Render each legacy project as a pill
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

// renderTaskwarriorSessionPill renders a project with Epic/Task session data.
// Format: 📋 feat-auth ▐━━━━│━━●─│────▌ 41% │ ▶ E2:T2 "AuthService"
//
// Params:
//   - sb: string builder to write to
//   - project: project with session data
func (r *Powerline) renderTaskwarriorSessionPill(sb *strings.Builder, project *model.TaskwarriorProject) {
	// Add space before pill
	sb.WriteString(" ")

	// Write left rounded cap
	sb.WriteString(FgTaskwarrior + LeftRound + Reset)

	// Write icon and project name
	sb.WriteString(BgTaskwarrior + FgTaskwarriorText + Bold + " " + IconTaskwarrior + " " + project.Name + " " + Reset)

	// Render segmented progress bar
	bar := r.renderSegmentedProgressBar(project.Epics)
	sb.WriteString(BgTaskwarrior + bar + Reset)

	// Write percentage
	sb.WriteString(BgTaskwarrior + FgTaskwarriorText + Bold + " " + itoa(project.Percent()) + "% " + Reset)

	// Render mode/task indicator
	if project.IsPlanMode() {
		// Show PLAN MODE indicator
		sb.WriteString(BgTaskwarrior + FgGraySep + "│ " + Reset)
		sb.WriteString(BgTaskwarrior + ColorGray + "🔍 PLAN" + Reset)
	} else if project.CurrentTask != "" {
		// Show current task indicator
		sb.WriteString(BgTaskwarrior + FgGraySep + "│ " + Reset)
		sb.WriteString(BgTaskwarrior + FgCyanTask + Bold + "▶ E" + itoa(project.CurrentEpic) + ":" + project.CurrentTask + Reset)
		// Show task name if available
		if taskName := project.CurrentTaskName(); taskName != "" {
			sb.WriteString(BgTaskwarrior + ColorGray + " \"" + taskName + "\"" + Reset)
		}
	}

	// Write right rounded cap
	sb.WriteString(BgTaskwarrior + " " + Reset)
	sb.WriteString(FgTaskwarrior + RightRound + Reset)
}

// renderSegmentedProgressBar renders a progress bar segmented by epics.
// Format: ▐━━━━│━━●─│────▌
//
// Params:
//   - epics: list of epics with task data
//
// Returns:
//   - string: rendered segmented progress bar with ANSI codes
func (r *Powerline) renderSegmentedProgressBar(epics []model.TaskwarriorEpic) string {
	var sb strings.Builder

	// Left border
	sb.WriteString(FgGraySep + "▐" + Reset)

	// Render each epic segment
	for i, epic := range epics {
		// Add separator between epics
		if i > 0 {
			sb.WriteString(FgGraySep + "│" + Reset)
		}
		// Render epic bar
		sb.WriteString(r.renderEpicBar(&epic))
	}

	// Right border
	sb.WriteString(FgGraySep + "▌" + Reset)

	return sb.String()
}

// epicBarMinWidth is the minimum width for an epic bar segment.
const epicBarMinWidth = 2

// epicBarMaxWidth is the maximum width for an epic bar segment.
const epicBarMaxWidth = 8

// renderEpicBar renders a single epic's progress bar segment.
//
// Params:
//   - epic: epic with task data
//
// Returns:
//   - string: rendered epic bar segment
func (r *Powerline) renderEpicBar(epic *model.TaskwarriorEpic) string {
	// Calculate width (min 2, max 8)
	width := epic.TotalCount
	if width < epicBarMinWidth {
		width = epicBarMinWidth
	}
	if width > epicBarMaxWidth {
		width = epicBarMaxWidth
	}

	// Calculate proportions
	doneWidth := 0
	wipWidth := 0
	todoWidth := width

	if epic.TotalCount > 0 {
		doneWidth = epic.DoneCount * width / epic.TotalCount
		// Check for WIP task
		for _, task := range epic.Tasks {
			if task.Status == model.StatusWip {
				wipWidth = 1
				break
			}
		}
		// Adjust done width if WIP present
		if wipWidth > 0 && doneWidth > 0 {
			doneWidth--
		}
		todoWidth = width - doneWidth - wipWidth
	}

	var sb strings.Builder

	// Done characters (heavy line, green)
	if doneWidth > 0 {
		sb.WriteString(FgGreenDone)
		for range doneWidth {
			sb.WriteRune('━')
		}
		sb.WriteString(Reset)
	}

	// WIP character (cursor, yellow)
	if wipWidth > 0 {
		sb.WriteString(FgYellowWip + "●" + Reset)
	}

	// Todo characters (light line, gray)
	if todoWidth > 0 {
		sb.WriteString(FgGrayTodo)
		for range todoWidth {
			sb.WriteRune('─')
		}
		sb.WriteString(Reset)
	}

	return sb.String()
}

// renderUpdatePill renders the update notification pill.
//
// Params:
//   - sb: string builder to write to
//   - update: update information
func (r *Powerline) renderUpdatePill(sb *strings.Builder, update model.UpdateInfo) {
	// Skip if no update available
	if !update.Available {
		// Return early if nothing to show
		return
	}

	// Add space before pill
	sb.WriteString(" ")

	// Write left rounded cap (white)
	sb.WriteString(FgWhite + LeftRound + Reset)
	// Write update icon and version on white background
	sb.WriteString(BgWhite + FgBlack + Bold + " " + IconUpdate + " " + update.Version + " " + Reset)
	// Write right rounded cap
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
