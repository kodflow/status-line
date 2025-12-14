// Package renderer provides status line rendering.
package renderer

import (
	"strings"

	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

var _ port.Renderer = (*Powerline)(nil)

// Powerline implements port.Renderer with powerline style.
type Powerline struct{}

// NewPowerline creates a new powerline renderer.
func NewPowerline() *Powerline {
	return &Powerline{}
}

// Render generates the status line string.
func (r *Powerline) Render(data model.StatusLineData) string {
	var sb strings.Builder

	r.renderLine1(&sb, data)
	sb.WriteString("\n")
	r.renderLine2(&sb, data)
	sb.WriteString("\n")

	return sb.String()
}

func (r *Powerline) renderLine1(sb *strings.Builder, data model.StatusLineData) {
	r.renderOSSegment(sb, data.System, data.Icons.OS)
	r.renderPathSegment(sb, data.Dir, data.Git.IsInRepo(), data.Icons.Path)
	r.renderGitSegment(sb, data.Git, data.Icons.Git)
}

func (r *Powerline) renderLine2(sb *strings.Builder, data model.StatusLineData) {
	r.renderModelPill(sb, data.Model, data.Icons.Model)
}

func (r *Powerline) renderOSSegment(sb *strings.Builder, sys model.SystemInfo, showIcon bool) {
	sb.WriteString(FgWhite + LeftRound + Reset)
	if showIcon {
		icon := GetOSIcon(sys.OS, sys.IsDocker)
		sb.WriteString(BgWhite + FgBlack + Bold + " " + icon + " " + Reset)
	} else {
		sb.WriteString(BgWhite + FgBlack + Bold + "  " + Reset)
	}
	sb.WriteString(BgBlue + FgWhite + SepRight + Reset)
}

func (r *Powerline) renderPathSegment(sb *strings.Builder, dir string, hasGit bool, showIcon bool) {
	truncated := TruncatePath(dir, defaultMaxPath)
	if showIcon {
		sb.WriteString(BgBlue + FgBlack + Bold + " " + IconFolder + " " + truncated + " " + Reset)
	} else {
		sb.WriteString(BgBlue + FgBlack + Bold + " " + truncated + " " + Reset)
	}

	if hasGit {
		sb.WriteString(BgCyan + FgBlue + SepRight + Reset)
	} else {
		sb.WriteString(FgBlue + SepRight + Reset)
	}
}

func (r *Powerline) renderGitSegment(sb *strings.Builder, git model.GitStatus, showIcon bool) {
	if !git.IsInRepo() {
		return
	}

	if showIcon {
		sb.WriteString(BgCyan + FgBlack + Bold + " " + IconGitBranch + " " + git.Branch)
	} else {
		sb.WriteString(BgCyan + FgBlack + Bold + " " + git.Branch)
	}

	if git.Modified > 0 {
		sb.WriteString(" " + FgYellow + "!" + itoa(git.Modified) + FgBlack)
	}
	if git.Untracked > 0 {
		sb.WriteString(" " + FgYellow + "?" + itoa(git.Untracked) + FgBlack)
	}

	sb.WriteString(" " + Reset + FgCyan + SepRight + Reset)
}

func (r *Powerline) renderModelPill(sb *strings.Builder, m model.ModelInfo, showIcon bool) {
	bgColor, fgColor := GetModelColors(m.Name)

	sb.WriteString(fgColor + LeftRound + Reset)

	if showIcon {
		sb.WriteString(bgColor + FgBlack + Bold + " " + IconModel + " " + m.Name + " " + Reset)
	} else {
		sb.WriteString(bgColor + FgBlack + Bold + " " + m.Name + " " + Reset)
	}

	if m.Version != "" {
		sb.WriteString(bgColor + FgWhite + LeftRound + Reset)
		sb.WriteString(BgWhite + FgBlack + Bold + " " + m.Version + " " + Reset)
		sb.WriteString(FgWhite + RightRound + Reset)
	} else {
		sb.WriteString(fgColor + RightRound + Reset)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
