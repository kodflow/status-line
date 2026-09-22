// Package sample builds a realistic status line state for the demo tools.
package sample

import (
	"os"
	"path/filepath"
	"time"

	"github.com/florent/status-line/internal/domain/model"
)

// Data returns a busy session: Opus 5 at effort xhigh, session, weekly and
// Opus-scoped quotas with their windows, 42 % of the context, a long path,
// a long branch and +12 -3 of changes.
//
// Returns:
//   - model.StatusLineData: state to render
func Data() model.StatusLineData {
	now := time.Now()
	home, _ := os.UserHomeDir()
	week := 7 * 24 * time.Hour
	return model.StatusLineData{
		Model:  model.ModelInfo{Name: "Opus", Version: "5"},
		Effort: "xhigh",
		Limits: model.LimitSet{
			Context: model.NewLimit(model.KindContext, "context", 42, time.Time{}, 0, model.SourceStdin),
			Session: model.NewLimit(model.KindSession, "session", 34, now.Add(2*time.Hour+10*time.Minute), 5*time.Hour, model.SourceStdin),
			Weekly:  model.NewLimit(model.KindWeekly, "weekly", 61, now.Add(3*24*time.Hour+4*time.Hour), week, model.SourceStdin),
			Scoped: []model.Limit{
				model.NewLimit(model.KindScoped, "opus", 48, now.Add(3*24*time.Hour+4*time.Hour), week, model.SourceAPI),
			},
		},
		Icons:   model.IconConfig{OS: true, Model: true, Path: true, Git: true},
		System:  model.SystemInfo{OS: model.OSLinux},
		Health:  model.HealthOK,
		Dir:     filepath.Join(home, "Documents", "worktrees", "status-line-adaptive", "internal", "presentation", "renderer"),
		Git:     model.GitStatus{Branch: "feat/adaptive-line-condense-to-terminal-width", Modified: 3, Untracked: 1},
		Changes: model.CodeChanges{Added: 12, Removed: 3},
		MCP: model.MCPServers{
			{Name: "github", Enabled: true, Source: model.MCPSourceCLI},
			{Name: "gitlab", Enabled: true, Source: model.MCPSourceCLI},
			{Name: "context7", Enabled: true, Source: model.MCPSourceCLI},
			{Name: "playwright", Enabled: true, Source: model.MCPSourceCLI},
			{Name: "codacy", Enabled: true, Source: model.MCPSourceCLI},
			{Name: "GitKraken", Enabled: true, Source: model.MCPSourceUser},
			{Name: "tasks", Enabled: true, Source: model.MCPSourcePlugin, Plugin: "kodflow-hooks"},
			{Name: "legacy", Enabled: false, Source: model.MCPSourceUser},
		},
	}
}
