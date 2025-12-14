// Package renderer provides status line rendering.
package renderer

import (
	"strings"

	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

// Numeric constants for base conversion.
const (
	// base10 is the decimal base for integer to string conversion.
	base10 int = 10
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

// renderLine1 renders the first line with OS, Path, and Git segments.
//
// Params:
//   - sb: string builder to write to
//   - data: status line data
func (r *Powerline) renderLine1(sb *strings.Builder, data model.StatusLineData) {
	// Render OS segment
	r.renderOSSegment(sb, data.System, data.Icons.OS)
	// Render path segment
	r.renderPathSegment(sb, data.Dir, data.Git.IsInRepo(), data.Icons.Path)
	// Render git segment if in repo
	r.renderGitSegment(sb, data.Git, data.Icons.Git)
}

// renderLine2 renders the second line with Model pill.
//
// Params:
//   - sb: string builder to write to
//   - data: status line data
func (r *Powerline) renderLine2(sb *strings.Builder, data model.StatusLineData) {
	// Render model pill
	r.renderModelPill(sb, data.Model, data.Icons.Model)
}

// renderOSSegment renders the operating system segment.
//
// Params:
//   - sb: string builder to write to
//   - sys: system information
//   - showIcon: whether to show the OS icon
func (r *Powerline) renderOSSegment(sb *strings.Builder, sys model.SystemInfo, showIcon bool) {
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
	sb.WriteString(BgBlue + FgWhite + SepRight + Reset)
}

// renderPathSegment renders the current directory segment.
//
// Params:
//   - sb: string builder to write to
//   - dir: directory path
//   - hasGit: whether git segment follows
//   - showIcon: whether to show the folder icon
func (r *Powerline) renderPathSegment(sb *strings.Builder, dir string, hasGit bool, showIcon bool) {
	truncated := TruncatePath(dir, defaultMaxPath)
	// Check if icon should be shown
	if showIcon {
		// Write path with folder icon
		sb.WriteString(BgBlue + FgBlack + Bold + " " + IconFolder + " " + truncated + " " + Reset)
	} else {
		// Write path without icon
		sb.WriteString(BgBlue + FgBlack + Bold + " " + truncated + " " + Reset)
	}

	// Check if git segment follows
	if hasGit {
		// Write separator to git segment
		sb.WriteString(BgCyan + FgBlue + SepRight + Reset)
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
func (r *Powerline) renderGitSegment(sb *strings.Builder, git model.GitStatus, showIcon bool) {
	// Skip if not in a git repository
	if !git.IsInRepo() {
		// Return early if not in repo
		return
	}

	// Check if icon should be shown
	if showIcon {
		// Write branch with icon
		sb.WriteString(BgCyan + FgBlack + Bold + " " + IconGitBranch + " " + git.Branch)
	} else {
		// Write branch without icon
		sb.WriteString(BgCyan + FgBlack + Bold + " " + git.Branch)
	}

	// Add modified indicator if present
	if git.Modified > 0 {
		sb.WriteString(" " + FgYellow + "!" + itoa(git.Modified) + FgBlack)
	}
	// Add untracked indicator if present
	if git.Untracked > 0 {
		sb.WriteString(" " + FgYellow + "?" + itoa(git.Untracked) + FgBlack)
	}

	// Write segment end
	sb.WriteString(" " + Reset + FgCyan + SepRight + Reset)
}

// renderModelPill renders the model name as a nested pill.
//
// Params:
//   - sb: string builder to write to
//   - m: model information
//   - showIcon: whether to show the model icon
func (r *Powerline) renderModelPill(sb *strings.Builder, m model.ModelInfo, showIcon bool) {
	bgColor, fgColor := GetModelColors(m.Name)

	// Write left rounded cap
	sb.WriteString(fgColor + LeftRound + Reset)

	// Check if icon should be shown
	if showIcon {
		// Write model name with icon
		sb.WriteString(bgColor + FgBlack + Bold + " " + IconModel + " " + m.Name + " " + Reset)
	} else {
		// Write model name without icon
		sb.WriteString(bgColor + FgBlack + Bold + " " + m.Name + " " + Reset)
	}

	// Check if version exists for nested pill
	if m.Version != "" {
		// Write nested version pill
		sb.WriteString(bgColor + FgWhite + LeftRound + Reset)
		sb.WriteString(BgWhite + FgBlack + Bold + " " + m.Version + " " + Reset)
		sb.WriteString(FgWhite + RightRound + Reset)
	} else {
		// Write simple right cap
		sb.WriteString(fgColor + RightRound + Reset)
	}
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
