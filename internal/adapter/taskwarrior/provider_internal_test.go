package taskwarrior

import "testing"

func TestProvider_shortProjectName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "hierarchical name", input: "Epic1.Feature.SubTask", want: "SubTask"},
		{name: "simple name", input: "MyProject", want: "MyProject"},
		{name: "single level hierarchy", input: "Parent.Child", want: "Child"},
		{name: "empty string", input: "", want: ""},
		{name: "name ending with dot", input: "Project.", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{}
			if got := p.shortProjectName(tt.input); got != tt.want {
				t.Errorf("shortProjectName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestProvider_isInstalled(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "checks installation"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{}
			_ = p.isInstalled() // Just verify no panic
		})
	}
}

func TestProvider_countTasks(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "counts tasks"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{}
			count := p.countTasks("status:pending")
			if count < 0 {
				t.Errorf("countTasks() = %d, want >= 0", count)
			}
		})
	}
}

func TestProvider_getProjects(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "gets projects"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{}
			projects := p.getProjects()
			if projects == nil {
				t.Error("getProjects() = nil, want non-nil")
			}
		})
	}
}

func TestProvider_listProjects(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "lists projects"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{}
			projects := p.listProjects()
			if projects == nil {
				t.Error("listProjects() = nil, want non-nil")
			}
		})
	}
}
