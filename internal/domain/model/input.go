// Package model contains domain entities and value objects.
package model

const (
	defaultModelName         = "Claude"
	defaultWorkingDir        = "~"
	defaultContextWindowSize = 200000
)

// Input represents the JSON input from Claude Code.
type Input struct {
	Model         InputModel     `json:"model"`
	Workspace     InputWorkspace `json:"workspace"`
	ContextWindow InputContext   `json:"context_window"`
}

// InputModel contains model display information from JSON.
type InputModel struct {
	DisplayName string `json:"display_name"`
}

// InputWorkspace contains workspace information from JSON.
type InputWorkspace struct {
	CurrentDir string `json:"current_dir"`
}

// InputContext contains context window information from JSON.
type InputContext struct {
	TotalInputTokens  int `json:"total_input_tokens"`
	TotalOutputTokens int `json:"total_output_tokens"`
	ContextWindowSize int `json:"context_window_size"`
}

// ModelInfo returns parsed model information.
func (i *Input) ModelInfo() ModelInfo {
	name := i.Model.DisplayName
	if name == "" {
		name = defaultModelName
	}
	baseName, version := parseModelName(name)
	return ModelInfo{Name: baseName, Version: version}
}

// WorkingDir returns the working directory.
func (i *Input) WorkingDir() string {
	if i.Workspace.CurrentDir == "" {
		return defaultWorkingDir
	}
	return i.Workspace.CurrentDir
}

// ContextWindowSize returns the context window size.
func (i *Input) ContextWindowSize() int {
	if i.ContextWindow.ContextWindowSize == 0 {
		return defaultContextWindowSize
	}
	return i.ContextWindow.ContextWindowSize
}

// TotalTokens returns total tokens used.
func (i *Input) TotalTokens() int {
	return i.ContextWindow.TotalInputTokens + i.ContextWindow.TotalOutputTokens
}

// Progress returns the context usage progress.
func (i *Input) Progress() Progress {
	return NewProgress(i.TotalTokens(), i.ContextWindowSize())
}

func parseModelName(name string) (baseName, version string) {
	for idx, ch := range name {
		if ch == ' ' && idx+1 < len(name) {
			return name[:idx], name[idx+1:]
		}
	}
	return name, ""
}
