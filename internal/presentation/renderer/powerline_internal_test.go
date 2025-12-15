package renderer

import (
	"strings"
	"testing"

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

func TestPowerline_renderLine2(t *testing.T) {
	tests := []struct {
		name string
		data model.StatusLineData
	}{
		{
			name: "with taskwarrior",
			data: model.StatusLineData{
				Taskwarrior: model.TaskwarriorInfo{
					Installed: true,
					Projects:  []model.TaskwarriorProject{{Name: "test", Pending: 5, Completed: 3}},
				},
			},
		},
		{
			name: "empty",
			data: model.StatusLineData{
				Taskwarrior: model.TaskwarriorInfo{Installed: false},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Powerline{}
			var sb strings.Builder
			r.renderLine2(&sb, tt.data)
			_ = sb.String() // Just verify no panic
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

func TestPowerline_renderMCPPill(t *testing.T) {
	tests := []struct {
		name   string
		server model.MCPServer
	}{
		{name: "enabled", server: model.MCPServer{Name: "test", Enabled: true}},
		{name: "disabled", server: model.MCPServer{Name: "test", Enabled: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Powerline{}
			var sb strings.Builder
			r.renderMCPPill(&sb, tt.server)
			if sb.Len() == 0 {
				t.Error("renderMCPPill() produced empty output")
			}
		})
	}
}

func TestPowerline_renderTaskwarriorPill(t *testing.T) {
	tests := []struct {
		name string
		tw   model.TaskwarriorInfo
	}{
		{name: "installed with projects", tw: model.TaskwarriorInfo{Installed: true, Projects: []model.TaskwarriorProject{{Name: "test", Pending: 5}}}},
		{name: "not installed", tw: model.TaskwarriorInfo{Installed: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Powerline{}
			var sb strings.Builder
			r.renderTaskwarriorPill(&sb, tt.tw)
			_ = sb.String() // Just verify no panic
		})
	}
}

func TestPowerline_renderTaskwarriorProjectPill(t *testing.T) {
	tests := []struct {
		name    string
		project model.TaskwarriorProject
	}{
		{name: "with tasks", project: model.TaskwarriorProject{Name: "test", Pending: 5, Completed: 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Powerline{}
			var sb strings.Builder
			r.renderTaskwarriorProjectPill(&sb, tt.project)
			if sb.Len() == 0 {
				t.Error("renderTaskwarriorProjectPill() produced empty output")
			}
		})
	}
}
