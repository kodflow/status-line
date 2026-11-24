package renderer

import (
	"strings"
	"testing"
	"time"

	"github.com/florent/status-line/internal/domain/model"
)

func TestPowerline_renderLine1(t *testing.T) {
	tests := []struct {
		name string
		data model.StatusLineData
	}{
		{
			name: "with git and changes",
			data: model.StatusLineData{
				Model:    model.ModelInfo{Name: "Opus", Version: "4.5"},
				Progress: model.Progress{Percent: 50},
				Icons:    model.IconConfig{OS: true, Model: true, Path: true, Git: true},
				Git:      model.GitStatus{Branch: "main", Modified: 1},
				System:   model.SystemInfo{OS: model.OSLinux},
				Dir:      "/workspace",
				Changes:  model.CodeChanges{Added: 10, Removed: 5},
			},
		},
		{
			name: "without git",
			data: model.StatusLineData{
				Model:  model.ModelInfo{Name: "Sonnet"},
				System: model.SystemInfo{OS: model.OSDarwin},
				Dir:    "/Users/test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Powerline{}
			var sb strings.Builder
			r.renderLine1(&sb, tt.data)
			if sb.Len() == 0 {
				t.Error("renderLine1() produced empty output")
			}
		})
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  string
	}{
		{name: "zero", input: 0, want: "0"},
		{name: "single digit", input: 5, want: "5"},
		{name: "multiple digits", input: 123, want: "123"},
		{name: "large number", input: 99999, want: "99999"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := itoa(tt.input); got != tt.want {
				t.Errorf("itoa(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// osSegment renders the OS segment at one fit level.
func osSegment(sb *strings.Builder, health model.ServiceHealth, fit lineFit) {
	(&Powerline{}).renderOSSegment(sb, &OSSegmentData{
		System:   model.SystemInfo{OS: model.OSLinux},
		ShowIcon: true,
		Health:   health,
		NextBg:   BgBlue,
		Fit:      fit,
	})
}

func TestPowerline_renderOSSegment(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "renders OS segment"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			osSegment(&sb, model.HealthUnknown, lineFit{})
			if sb.Len() == 0 {
				t.Error("renderOSSegment() produced empty output")
			}
		})
	}
}

func TestPowerline_renderModelSegment(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "renders model segment"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Powerline{}
			var sb strings.Builder
			data := &ModelSegmentData{
				Model:    model.ModelInfo{Name: "Opus"},
				ShowIcon: true,
				Progress: model.Progress{Percent: 50},
				Cursor:   nil,
				NextBg:   BgBlue,
			}
			r.renderModelSegment(&sb, data)
			if sb.Len() == 0 {
				t.Error("renderModelSegment() produced empty output")
			}
		})
	}
}

func TestPowerline_renderPathSegment(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "renders path segment"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Powerline{}
			var sb strings.Builder
			r.renderPathSegment(&sb, "/workspace", true, true, "", 0)
			if sb.Len() == 0 {
				t.Error("renderPathSegment() produced empty output")
			}
		})
	}
}

func TestPowerline_renderGitSegment(t *testing.T) {
	tests := []struct {
		name string
		git  model.GitStatus
	}{
		{name: "with branch", git: model.GitStatus{Branch: "main"}},
		{name: "not in repo", git: model.GitStatus{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Powerline{}
			var sb strings.Builder
			r.renderGitSegment(&sb, tt.git, true, "", 0)
			_ = sb.String() // Just verify no panic
		})
	}
}

func TestPowerline_renderChangesSegment(t *testing.T) {
	tests := []struct {
		name    string
		changes model.CodeChanges
	}{
		{name: "with changes", changes: model.CodeChanges{Added: 10, Removed: 5}},
		{name: "no changes", changes: model.CodeChanges{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Powerline{}
			var sb strings.Builder
			r.renderChangesSegment(&sb, tt.changes)
			_ = sb.String() // Just verify no panic
		})
	}
}

func TestPowerline_renderUpdatePill(t *testing.T) {
	tests := []struct {
		name   string
		update model.UpdateInfo
	}{
		{name: "with update available", update: model.UpdateInfo{Available: true, Version: "v1.0.0"}},
		{name: "no update", update: model.UpdateInfo{Available: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Powerline{}
			var sb strings.Builder
			r.renderUpdatePill(&sb, tt.update)
			// Should produce output only if update available
			if tt.update.Available && sb.Len() == 0 {
				t.Error("renderUpdatePill() produced empty output for available update")
			}
		})
	}
}

func TestPowerline_renderUpdatePillFailureVariant(t *testing.T) {
	r := &Powerline{}
	var lit, failed, failedNoTag strings.Builder
	r.renderUpdatePill(&lit, model.UpdateInfo{Available: true, Version: "v1.1.0"})
	r.renderUpdatePill(&failed, model.UpdateInfo{Available: true, Version: "v1.1.0", Failed: true})
	r.renderUpdatePill(&failedNoTag, model.UpdateInfo{Available: true, Failed: true})
	if !strings.Contains(lit.String(), IconUpdate) || strings.Contains(lit.String(), IconUpdateFailed) {
		t.Errorf("lit pill = %q, want the download glyph", lit.String())
	}
	if !strings.Contains(failed.String(), IconUpdateFailed+" v1.1.0") || !strings.Contains(failed.String(), BgRed) {
		t.Errorf("failed pill = %q, want the warning glyph on red with the release", failed.String())
	}
	if strings.Contains(failedNoTag.String(), " v") || !strings.Contains(failedNoTag.String(), IconUpdateFailed) {
		t.Errorf("failed pill without a release = %q", failedNoTag.String())
	}
}

func TestPowerline_renderModelSegmentCursorTakesTheInk(t *testing.T) {
	// The even-burn cursor used to be one orange for every model; on the pink
	// and lavender grounds it measured 2.2:1 and clashed in hue. It now takes
	// the pill's own ink, which is already held to the contrast floor.
	now := time.Now()
	for _, name := range []string{"Haiku 4.5", "Sonnet 5", "Opus 5", "Fable 5.1", "Mystery 1"} {
		t.Run(name, func(t *testing.T) {
			_, _, ink := GetModelColors(name)
			quota := model.Limit{
				Kind:     model.KindSession,
				Percent:  35,
				ResetsAt: now.Add(2 * time.Hour),
				Window:   5 * time.Hour,
			}
			r := &Powerline{}
			var sb strings.Builder
			r.renderModelSegment(&sb, &ModelSegmentData{
				Model:  model.ModelInfo{Name: name},
				Quotas: []model.Limit{quota},
				NextBg: BgBlue,
			})
			if !strings.Contains(sb.String(), ink+string(cursorChar)) {
				t.Errorf("cursor is not drawn in the model ink %q", ink)
			}
		})
	}
}

func TestPowerline_renderOSSegmentHealth(t *testing.T) {
	tests := []struct {
		name   string
		health model.ServiceHealth
		want   string
	}{
		{name: "operational", health: model.HealthOK, want: FgHealthOK + glyphs.Health},
		{name: "degraded", health: model.HealthDegraded, want: FgHealthDegraded + glyphs.Health},
		{name: "down", health: model.HealthDown, want: FgHealthDown + glyphs.Health},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			osSegment(&sb, tt.health, lineFit{})
			if !strings.Contains(sb.String(), tt.want) {
				t.Errorf("health glyph not drawn in its colour")
			}
		})
	}

	t.Run("unknown draws nothing", func(t *testing.T) {
		var sb strings.Builder
		osSegment(&sb, model.HealthUnknown, lineFit{})
		if strings.Contains(sb.String(), glyphs.Health) {
			t.Errorf("an unknown state must not be drawn")
		}
	})
}

func TestPowerline_renderGitSegmentWorktrees(t *testing.T) {
	var with, without strings.Builder
	(&Powerline{}).renderGitSegment(&with, model.GitStatus{Branch: "main", Worktrees: 2}, true, "", 0)
	(&Powerline{}).renderGitSegment(&without, model.GitStatus{Branch: "main"}, true, "", 0)
	if !strings.Contains(with.String(), glyphs.Worktree+" 2") {
		t.Errorf("worktree count missing from %q", with.String())
	}
	if strings.Contains(without.String(), glyphs.Worktree) {
		t.Errorf("worktree glyph drawn without any linked worktree")
	}
}

func TestPowerline_renderIsOneLine(t *testing.T) {
	for _, update := range []model.UpdateInfo{
		{},
		{Available: true, Version: "v9.9.9"},
		{Available: true, Version: "v9.9.9", Failed: true},
	} {
		for _, width := range []int{0, 200, 80, 40} {
			data := busyLine(width)
			data.Update = update
			out := (&Powerline{}).Render(data)
			// One line, closed by the one newline the host splits on
			if strings.Count(out, "\n") != 1 || !strings.HasSuffix(out, "\n") {
				t.Errorf("update %+v COLUMNS=%d: want exactly one line, got %q", update, width, out)
			}
		}
	}
}

func TestPowerline_updatePillEndsTheLine(t *testing.T) {
	data := busyLine(0)
	data.Update = model.UpdateInfo{Available: true, Version: "v9.9.9"}
	out := (&Powerline{}).Render(data)
	changes, pill := strings.Index(out, "-3"), strings.Index(out, "v9.9.9")
	// The pill follows the changes segment, and nothing follows the pill
	if changes < 0 || pill < changes {
		t.Errorf("want the update pill after the changes, got %q", out)
	}
	if !strings.HasSuffix(out, FgWhite+RightRound+Reset+"\n") {
		t.Errorf("want the pill's closing cap to end the line, got %q", out)
	}
	// Without a pending release the line ends on the changes segment
	data.Update = model.UpdateInfo{}
	if out := (&Powerline{}).Render(data); strings.Contains(out, IconUpdate) || strings.Contains(out, RightRound) {
		t.Errorf("no update draws no pill, got %q", out)
	}
}

func TestPowerline_osSegmentDrawsNoMoreThanHealth(t *testing.T) {
	var sb strings.Builder
	osSegment(&sb, model.HealthOK, lineFit{})
	want := FgWhite + LeftRound + Reset + BgWhite + FgBlack + Bold + " " + IconLinux + " " + Reset +
		BgWhite + FgHealthOK + glyphs.Health + " " + Reset + BgBlue + FgWhite + SepRight + Reset
	// The icon, the health light and the arrow: nothing else lives here now
	if sb.String() != want {
		t.Errorf("os segment = %q, want %q", sb.String(), want)
	}
}
