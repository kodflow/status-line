// Package model contains domain entities and value objects.
package model

// Default values for input fields.
const (
	// defaultModelName is the fallback model name.
	defaultModelName string = "Claude"
	// defaultWorkingDir is the fallback working directory.
	defaultWorkingDir string = "~"
	// defaultContextWindowSize is the fallback context window size.
	defaultContextWindowSize int = 200000
)

// Input represents the JSON input from Claude Code.
// It contains all the information needed to render the status line.
type Input struct {
	Model         InputModel     `json:"model"`
	Workspace     InputWorkspace `json:"workspace"`
	ContextWindow InputContext   `json:"context_window"`
}

// InputModel contains model display information from JSON.
// It holds the display name of the AI model being used.
type InputModel struct {
	DisplayName string `json:"display_name"`
}

// InputWorkspace contains workspace information from JSON.
// It holds the current working directory path.
type InputWorkspace struct {
	CurrentDir string `json:"current_dir"`
}

// InputContext contains context window information from JSON.
// It tracks input/output tokens and the maximum context size.
type InputContext struct {
	TotalInputTokens  int `json:"total_input_tokens"`
	TotalOutputTokens int `json:"total_output_tokens"`
	ContextWindowSize int `json:"context_window_size"`
}

// ModelInfo returns parsed model information.
//
// Returns:
//   - ModelInfo: parsed model name and version
func (i *Input) ModelInfo() ModelInfo {
	name := i.Model.DisplayName
	// Use default if display name is empty
	if name == "" {
		name = defaultModelName
	}
	baseName, version := parseModelName(name)
	// Return parsed model info
	return ModelInfo{Name: baseName, Version: version}
}

// WorkingDir returns the working directory.
//
// Returns:
//   - string: current working directory or default
func (i *Input) WorkingDir() string {
	// Use default if directory is empty
	if i.Workspace.CurrentDir == "" {
		// Return fallback directory
		return defaultWorkingDir
	}
	// Return configured directory
	return i.Workspace.CurrentDir
}

// ContextWindowSize returns the context window size.
//
// Returns:
//   - int: context window size or default
func (i *Input) ContextWindowSize() int {
	// Use default if size is zero
	if i.ContextWindow.ContextWindowSize == 0 {
		// Return fallback size
		return defaultContextWindowSize
	}
	// Return configured size
	return i.ContextWindow.ContextWindowSize
}

// TotalTokens returns total tokens used.
//
// Returns:
//   - int: sum of input and output tokens
func (i *Input) TotalTokens() int {
	// Sum both token counts
	return i.ContextWindow.TotalInputTokens + i.ContextWindow.TotalOutputTokens
}

// Progress returns the context usage progress.
//
// Returns:
//   - Progress: calculated progress based on token usage
func (i *Input) Progress() Progress {
	// Create progress from token counts
	return NewProgress(i.TotalTokens(), i.ContextWindowSize())
}

// parseModelName splits a model name into base name and version.
//
// Params:
//   - name: full model name to parse
//
// Returns:
//   - baseName: model name without version
//   - version: version string or empty
func parseModelName(name string) (baseName, version string) {
	// Find space separator
	for idx, ch := range name {
		// Check for space and remaining characters
		if ch == ' ' && idx+1 < len(name) {
			// Return split name and version
			return name[:idx], name[idx+1:]
		}
	}
	// Return name only if no version found
	return name, ""
}
