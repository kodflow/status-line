package application_test

import (
	"testing"

	"github.com/florent/status-line/internal/application"
	"github.com/florent/status-line/internal/domain/model"
)

type mockGitRepo struct{}

func (m *mockGitRepo) Status() model.GitStatus      { return model.GitStatus{Branch: "main"} }
func (m *mockGitRepo) DiffStats() model.CodeChanges { return model.CodeChanges{Added: 10, Removed: 5} }

type mockSystemProv struct{}

func (m *mockSystemProv) Info() model.SystemInfo { return model.SystemInfo{OS: model.OSLinux} }

type mockTerminalProv struct{}

func (m *mockTerminalProv) Info() model.TerminalInfo { return model.TerminalInfo{Width: 120} }

type mockMCPProv struct{}

func (m *mockMCPProv) Servers() model.MCPServers { return model.MCPServers{} }

type mockTaskwarriorProv struct{}

func (m *mockTaskwarriorProv) Info() model.TaskwarriorInfo {
	return model.TaskwarriorInfo{Installed: false}
}

type mockRenderer struct{}

func (m *mockRenderer) Render(data model.StatusLineData) string { return "mocked output" }

type mockInputProvider struct{}

func (m *mockInputProvider) ModelInfo() model.ModelInfo { return model.ModelInfo{Name: "Opus"} }
func (m *mockInputProvider) WorkingDir() string         { return "/workspace" }
func (m *mockInputProvider) Progress() model.Progress   { return model.Progress{Percent: 50} }

func TestNewStatusLineService(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates service"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := application.ServiceDeps{
				Git:         &mockGitRepo{},
				System:      &mockSystemProv{},
				Terminal:    &mockTerminalProv{},
				MCP:         &mockMCPProv{},
				Taskwarrior: &mockTaskwarriorProv{},
			}
			svc := application.NewStatusLineService(deps, &mockRenderer{})
			if svc == nil {
				t.Error("NewStatusLineService() returned nil")
			}
		})
	}
}

func TestStatusLineService_Generate(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "generates output", want: "mocked output"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := application.ServiceDeps{
				Git:         &mockGitRepo{},
				System:      &mockSystemProv{},
				Terminal:    &mockTerminalProv{},
				MCP:         &mockMCPProv{},
				Taskwarrior: &mockTaskwarriorProv{},
			}
			svc := application.NewStatusLineService(deps, &mockRenderer{})
			result := svc.Generate(&mockInputProvider{})
			if result != tt.want {
				t.Errorf("Generate() = %q, want %q", result, tt.want)
			}
		})
	}
}
