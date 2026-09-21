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
			r.renderOSSegment(&sb, model.SystemInfo{OS: model.OSLinux}, true, BgBlue)
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
