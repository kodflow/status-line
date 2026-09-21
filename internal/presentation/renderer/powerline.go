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

	r.renderCompact(&sb, data)

	// Return complete status line
	return sb.String()
}

// renderCompact renders the two-line shape: everything about the session on
// the first line, the ambient pills on the second.
//
// Params:
//   - sb: string builder to write to
//   - data: status line data
func (r *Powerline) renderCompact(sb *strings.Builder, data model.StatusLineData) {
	// Line one carries everything the session is: identity, quotas, repository.
	// Line two carries only what is ambient — the MCP servers and an update
	// notice — as the original status line did.
	r.renderLine1(sb, data)
	sb.WriteString("\n" + LineGap())
	r.renderLine2(sb, data)
	sb.WriteString("\n")
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
	r.renderOSSegment(sb, data.System, data.Icons.OS, data.Health, modelBg)

	// Build the quota chain first: each segment needs to know the colour of the
	// one that follows it to draw its separator
	segments := quotaSegments(data)

	// The model segment carries the account quotas that apply to this model
	modelNextBg := BgBlue
	// Hand over to the first remaining segment when there is one
	if len(segments) > 0 {
		modelNextBg = segments[0].bg
	}
	modelData := &ModelSegmentData{
		Model:    data.Model,
		ShowIcon: data.Icons.Model,
		Progress: data.Progress,
		Cursor:   nil,
		NextBg:   modelNextBg,
		Effort:   data.Effort,
		FastMode: data.FastMode,
		Quotas:   modelQuotas(data),
	}
	r.renderModelSegment(sb, modelData)

	// Chain every quota, each handing over to the next and the last to the path
	for idx, seg := range segments {
		nextBg := BgBlue
		// Hand over to the next quota when there is one
		if idx+1 < len(segments) {
			nextBg = segments[idx+1].bg
		}
		renderQuotaSegment(sb, seg, nextBg)
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

// renderLine2 renders the second line with dynamic pills (MCP, Update).
//
// Params:
//   - sb: string builder to write to
//   - data: status line data
func (r *Powerline) renderLine2(sb *strings.Builder, data model.StatusLineData) {
	// Track if we've rendered anything
	hasContent := false

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
//   - health: state of Claude's services, drawn beside the icon when known
//   - nextBg: background color of the next segment
func (r *Powerline) renderOSSegment(sb *strings.Builder, sys model.SystemInfo, showIcon bool, health model.ServiceHealth, nextBg string) {
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
	// The health glyph sits inside the same white ground, coloured by state
	if color := healthColor(health); color != "" && !isHidden(hideHealth) {
		sb.WriteString(BgWhite + color + glyphs.Health + " " + Reset)
	}
	// Write separator to next segment
	sb.WriteString(nextBg + FgWhite + SepRight + Reset)
}

// healthColor returns the glyph colour for a service health level.
//
// Params:
//   - health: aggregate state of Claude's services
//
// Returns:
//   - string: foreground escape, empty when nothing should be drawn
func healthColor(health model.ServiceHealth) string {
	// Map each known level onto its colour
	switch health {
	// Every counted service is operational
	case model.HealthOK:
		return FgHealthOK
	// One service is degraded
	case model.HealthDegraded:
		return FgHealthDegraded
	// Two degraded services or one major outage
	case model.HealthDown:
		return FgHealthDown
	// An unknown state is not drawn: a light nobody checked would mislead
	default:
		return ""
	}
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

	// Check if icon should be shown
	if data.ShowIcon {
		// Write model name with icon
		sb.WriteString(bgColor + textColor + Bold + " " + IconModel + " " + data.Model.ShortName() + Reset)
	} else {
		// Write model name without icon
		sb.WriteString(bgColor + textColor + Bold + " " + data.Model.ShortName() + Reset)
	}

	// Append the fast-mode marker when it is on
	if data.FastMode {
		sb.WriteString(bgColor + textColor + Bold + " " + glyphs.Fast + Reset)
	}

	// Draw the account quotas alongside the model they apply to. The session
	// window, the weekly window and any quota scoped to this model all limit
	// the same thing — what this account may spend on this model — so they
	// read as one group and are drawn as one, divided by a thin rule rather
	// than by a new segment.
	for idx, quota := range data.Quotas {
		cursor := noCursor
		// Place the even-burn cursor only when the window makes it meaningful
		if quota.HasWindow() {
			cursor = quota.CursorPosition()
		}
		bar := RenderProgressBarWidth(quota.Progress(), cursor, segBarWidth, textColor, bgColor+textColor)

		// The first quota is the model's own session budget and runs straight
		// on from its name; the rest are divided by a thin rule and named
		if idx == 0 {
			sb.WriteString(bgColor + textColor + " " + Reset)
		} else {
			sb.WriteString(bgColor + textColor + " " + glyphs.Divider + Reset)
			sb.WriteString(bgColor + textColor + Bold + " " + QuotaLabel(quota) + " " + Reset)
		}
		sb.WriteString(bgColor + textColor + bar + Bold + " " + itoa(quota.Percent) + "%" + Reset)
		// Append the countdown to the refill, which the bar cannot say
		if quota.HasWindow() {
			sb.WriteString(bgColor + textColor + " " + glyphs.Reset + FormatDuration(quota.Remaining()) + Reset)
		}
	}

	sb.WriteString(bgColor + " " + Reset)

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
	// Add the linked worktrees, where parallel work is going on
	if git.Worktrees > 0 {
		sb.WriteString(" " + glyphs.Worktree + " " + itoa(git.Worktrees))
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
func (r *Powerline) renderWeeklySegment(sb *strings.Builder, usage model.Limit) {
	progress := usage.Progress()
	bar := RenderProgressBarWithCursor(progress, usage.CursorPosition(), FgWeeklyText, BgWeekly+FgWeeklyText+Bold)

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
