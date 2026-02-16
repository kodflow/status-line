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
	Cost          InputCost      `json:"cost"`
}

// InputCost contains cost and code change information from JSON.
// It tracks session costs and lines of code modified.
type InputCost struct {
	TotalCostUSD       float64 `json:"total_cost_usd"`
	TotalDurationMs    int64   `json:"total_duration_ms"`
	TotalAPIDurationMs int64   `json:"total_api_duration_ms"`
	TotalLinesAdded    int     `json:"total_lines_added"`
	TotalLinesRemoved  int     `json:"total_lines_removed"`
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
	TotalInputTokens    int      `json:"total_input_tokens"`
	TotalOutputTokens   int      `json:"total_output_tokens"`
	ContextWindowSize   int      `json:"context_window_size"`
	UsedPercentage      *float64 `json:"used_percentage"`
	RemainingPercentage *float64 `json:"remaining_percentage"`
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
// Prefers the pre-calculated used_percentage from Claude Code when available,
// as cumulative token totals exceed the context window after compaction.
//
// Returns:
//   - Progress: calculated progress based on context usage
func (i *Input) Progress() Progress {
	// Prefer pre-calculated percentage (reflects actual context window state)
	if i.ContextWindow.UsedPercentage != nil {
		percent := int(*i.ContextWindow.UsedPercentage)
		return Progress{Percent: min(percent, maxPercent)}
	}
	// Fallback to token-based calculation
	return NewProgress(i.TotalTokens(), i.ContextWindowSize())
}

// CodeChanges returns the lines added and removed.
//
// Returns:
//   - CodeChanges: lines added and removed in session
func (i *Input) CodeChanges() CodeChanges {
	// Return code changes from cost data
	return CodeChanges{
		Added:   i.Cost.TotalLinesAdded,
		Removed: i.Cost.TotalLinesRemoved,
	}
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
