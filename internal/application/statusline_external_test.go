package application_test

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/florent/status-line/internal/application"
	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

type mockGitRepo struct{}

func (m *mockGitRepo) Status() model.GitStatus      { return model.GitStatus{Branch: "main"} }
func (m *mockGitRepo) DiffStats() model.CodeChanges { return model.CodeChanges{Added: 10, Removed: 5} }

type mockSystemProv struct{}

func (m *mockSystemProv) Info() model.SystemInfo { return model.SystemInfo{OS: model.OSLinux} }

type mockTerminalProv struct{}

func (m *mockTerminalProv) Info() model.TerminalInfo { return model.TerminalInfo{Width: 120} }

type mockUsageProv struct{}

func (m *mockUsageProv) Limits() (model.LimitSet, error) {
	return model.LimitSet{}, nil
}

type mockRenderer struct{}

func (m *mockRenderer) Render(data model.StatusLineData) string { return "mocked output" }

type mockInputProvider struct{}

func (m *mockInputProvider) ModelInfo() model.ModelInfo { return model.ModelInfo{Name: "Opus"} }
func (m *mockInputProvider) WorkingDir() string         { return "/workspace" }
func (m *mockInputProvider) Progress() model.Progress   { return model.Progress{Percent: 50} }
func (m *mockInputProvider) EffortLevel() string        { return model.EffortHigh }
func (m *mockInputProvider) ContextTokens() int         { return 100000 }
func (m *mockInputProvider) ContextWindowSize() int     { return 200000 }
func (m *mockInputProvider) SessionCost() float64       { return 1.23 }
func (m *mockInputProvider) IsFastMode() bool           { return false }
func (m *mockInputProvider) SessionLabel() string       { return "test session" }
func (m *mockInputProvider) StdinLimits() model.LimitSet {
	return model.LimitSet{
		Context: model.NewLimit(model.KindContext, "ctx", 50, time.Time{}, 0, model.SourceStdin),
	}
}

func TestNewStatusLineService(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates service"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := application.ServiceDeps{
				Git:      &mockGitRepo{},
				System:   &mockSystemProv{},
				Terminal: &mockTerminalProv{},
				Usage:    &mockUsageProv{},
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
				Git:      &mockGitRepo{},
				System:   &mockSystemProv{},
				Terminal: &mockTerminalProv{},
				Usage:    &mockUsageProv{},
			}
			svc := application.NewStatusLineService(deps, &mockRenderer{})
			result := svc.Generate(&mockInputProvider{})
			if result != tt.want {
				t.Errorf("Generate() = %q, want %q", result, tt.want)
			}
		})
	}
}

type mockHealthProv struct{ level model.ServiceHealth }

func (m *mockHealthProv) Health() model.ServiceHealth { return m.level }

type capturingRenderer struct{ data model.StatusLineData }

func (c *capturingRenderer) Render(data model.StatusLineData) string {
	c.data = data
	return ""
}

func TestGenerate_WorkDirAndHealth(t *testing.T) {
	tests := []struct {
		name       string
		workDir    string
		health     port.HealthProvider
		wantDir    string
		wantHealth model.ServiceHealth
	}{
		{name: "inferred directory wins", workDir: "/elsewhere", health: &mockHealthProv{level: model.HealthDegraded},
			wantDir: "/elsewhere", wantHealth: model.HealthDegraded},
		{name: "reported directory is the fallback", workDir: "", health: nil,
			wantDir: "/workspace", wantHealth: model.HealthUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rend := &capturingRenderer{}
			deps := application.ServiceDeps{
				Git:      &mockGitRepo{},
				System:   &mockSystemProv{},
				Terminal: &mockTerminalProv{},
				Usage:    &mockUsageProv{},
				Health:   tt.health,
				WorkDir:  tt.workDir,
			}
			application.NewStatusLineService(deps, rend).Generate(&mockInputProvider{})
			if rend.data.Dir != tt.wantDir {
				t.Errorf("Dir = %q, want %q", rend.data.Dir, tt.wantDir)
			}
			if rend.data.Health != tt.wantHealth {
				t.Errorf("Health = %v, want %v", rend.data.Health, tt.wantHealth)
			}
		})
	}
}

type panickingHealthProv struct{}

func (p *panickingHealthProv) Health() model.ServiceHealth { panic("health exploded") }

func TestGenerate_ProviderPanicStillRenders(t *testing.T) {
	var stderr strings.Builder
	rend := &capturingRenderer{}
	deps := application.ServiceDeps{
		Git:      &mockGitRepo{},
		System:   &mockSystemProv{},
		Terminal: &mockTerminalProv{},
		Usage:    &mockUsageProv{},
		Health:   &panickingHealthProv{},
		Stderr:   &stderr,
	}
	application.NewStatusLineService(deps, rend).Generate(&mockInputProvider{})
	// The panicking provider's value stays zero, the others still arrive
	if rend.data.Git.Branch != "main" {
		t.Errorf("git branch = %q, want the other providers to still render", rend.data.Git.Branch)
	}
	if rend.data.Health != model.HealthUnknown {
		t.Errorf("health = %v, want zero after a panic", rend.data.Health)
	}
	if !strings.Contains(stderr.String(), "health exploded") || !strings.Contains(stderr.String(), "goroutine") {
		t.Errorf("stderr = %q, want the panic and its stack", stderr.String())
	}
}

type panickingGitRepo struct{}

func (p *panickingGitRepo) Status() model.GitStatus      { panic("status exploded") }
func (p *panickingGitRepo) DiffStats() model.CodeChanges { panic("diff exploded") }

// lockedWriter serialises writes: several providers report their panic at
// once, each from its own goroutine.
type lockedWriter struct {
	mu sync.Mutex
	sb strings.Builder
}

func (w *lockedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.sb.Write(p)
}

func (w *lockedWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.sb.String()
}

func TestGenerate_EveryProviderRecoversOnItsOwn(t *testing.T) {
	stderr := &lockedWriter{}
	rend := &capturingRenderer{}
	deps := application.ServiceDeps{
		Git:      &panickingGitRepo{},
		System:   &mockSystemProv{},
		Terminal: &mockTerminalProv{},
		Usage:    &mockUsageProv{},
		Health:   &panickingHealthProv{},
		Stderr:   stderr,
	}
	application.NewStatusLineService(deps, rend).Generate(&mockInputProvider{})
	// Three closures panic: each recover is its own, so all three are
	// reported and the providers that did not panic still arrive
	for _, msg := range []string{"status exploded", "diff exploded", "health exploded"} {
		if !strings.Contains(stderr.String(), msg) {
			t.Errorf("stderr misses %q: %q", msg, stderr.String())
		}
	}
	if rend.data.System.OS != model.OSLinux || rend.data.Terminal.Width != 120 {
		t.Errorf("system = %+v, terminal = %+v, want them despite the panics", rend.data.System, rend.data.Terminal)
	}
}
