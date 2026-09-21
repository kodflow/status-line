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

func TestPowerline_renderOSSegment(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "renders OS segment"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Powerline{}
			var sb strings.Builder
			r.renderOSSegment(&sb, model.SystemInfo{OS: model.OSLinux}, true, model.HealthUnknown, BgBlue)
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
			r.renderPathSegment(&sb, "/workspace", true, true, "")
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
			r.renderGitSegment(&sb, tt.git, true, "")
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

func TestPowerline_renderMCPPills(t *testing.T) {
	tests := []struct {
		name    string
		servers model.MCPServers
	}{
		{name: "with servers", servers: model.MCPServers{{Name: "test", Enabled: true}}},
		{name: "empty", servers: model.MCPServers{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Powerline{}
			var sb strings.Builder
			r.renderMCPPills(&sb, tt.servers)
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
			(&Powerline{}).renderOSSegment(&sb, model.SystemInfo{OS: model.OSLinux}, true, tt.health, BgBlue)
			if !strings.Contains(sb.String(), tt.want) {
				t.Errorf("health glyph not drawn in its colour")
			}
		})
	}

	t.Run("unknown draws nothing", func(t *testing.T) {
		var sb strings.Builder
		(&Powerline{}).renderOSSegment(&sb, model.SystemInfo{OS: model.OSLinux}, true, model.HealthUnknown, BgBlue)
		if strings.Contains(sb.String(), glyphs.Health) {
			t.Errorf("an unknown state must not be drawn")
		}
	})
}

func TestPowerline_renderGitSegmentWorktrees(t *testing.T) {
	var with, without strings.Builder
	(&Powerline{}).renderGitSegment(&with, model.GitStatus{Branch: "main", Worktrees: 2}, true, "")
	(&Powerline{}).renderGitSegment(&without, model.GitStatus{Branch: "main"}, true, "")
	if !strings.Contains(with.String(), glyphs.Worktree+" 2") {
		t.Errorf("worktree count missing from %q", with.String())
	}
	if strings.Contains(without.String(), glyphs.Worktree) {
		t.Errorf("worktree glyph drawn without any linked worktree")
	}
}

func TestPowerline_renderLine2Tasks(t *testing.T) {
	list := model.TaskList{Items: []model.TaskItem{
		{ID: "1", Subject: "done", Status: model.TaskCompleted},
		{ID: "2", Subject: "a rather long title that must never be shortened", Status: model.TaskInProgress},
		{ID: "3", Subject: "later", Status: model.TaskPending},
	}}
	mcp := model.MCPServers{{Name: "github", Enabled: true}}

	var sb strings.Builder
	(&Powerline{}).renderLine2(&sb, model.StatusLineData{Tasks: list, MCP: mcp})
	out := sb.String()
	for _, want := range []string{
		FgTaskDone + glyphs.TaskDone + FgTaskActive + glyphs.TaskDone + FgTaskTodo + glyphs.TaskOpen,
		"1/3",
		"a rather long title that must never be shortened",
		"github",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("line 2 misses %q in %q", want, out)
		}
	}
	if strings.Index(out, "1/3") > strings.Index(out, "github") {
		t.Error("the task list must lead the line, before the MCP pills")
	}

	var done strings.Builder
	finished := model.TaskList{Items: []model.TaskItem{{ID: "1", Status: model.TaskCompleted}}}
	(&Powerline{}).renderLine2(&done, model.StatusLineData{Tasks: finished, MCP: mcp})
	if strings.Contains(done.String(), glyphs.Tasks) {
		t.Error("a finished list must not be drawn")
	}
}
