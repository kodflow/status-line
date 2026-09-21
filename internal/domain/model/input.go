// Package model contains domain entities and value objects.
package model

import "time"

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
	Model         InputModel       `json:"model"`
	Workspace     InputWorkspace   `json:"workspace"`
	ContextWindow InputContext     `json:"context_window"`
	Cost          InputCost        `json:"cost"`
	RateLimits    InputRateLimits  `json:"rate_limits"`
	Effort        InputEffort      `json:"effort"`
	Thinking      InputThinking    `json:"thinking"`
	OutputStyle   InputOutputStyle `json:"output_style"`
	SessionName   string           `json:"session_name"`
	Transcript    string           `json:"transcript_path"`
	SessionID     string           `json:"session_id"`
	Version       string           `json:"version"`
	FastMode      bool             `json:"fast_mode"`
	Exceeds200k   bool             `json:"exceeds_200k_tokens"`
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
// It holds the display name and identifier of the AI model being used.
type InputModel struct {
	DisplayName string `json:"display_name"`
	ID          string `json:"id"`
}

// InputWorkspace contains workspace information from JSON.
// It holds the current working directory path.
type InputWorkspace struct {
	CurrentDir string `json:"current_dir"`
	ProjectDir string `json:"project_dir"`
}

// InputContext contains context window information from JSON.
// It tracks input/output tokens and the maximum context size.
type InputContext struct {
	CurrentUsage        InputCurrentUsage `json:"current_usage"`
	TotalInputTokens    int               `json:"total_input_tokens"`
	TotalOutputTokens   int               `json:"total_output_tokens"`
	ContextWindowSize   int               `json:"context_window_size"`
	UsedPercentage      *float64          `json:"used_percentage"`
	RemainingPercentage *float64          `json:"remaining_percentage"`
}

// InputCurrentUsage is the live token composition of the context window.
// It reflects what is resident now, unlike the cumulative session totals.
type InputCurrentUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
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

// Session returns the session identifier, empty when absent.
//
// Returns:
//   - string: session id as reported by Claude Code
func (i *Input) Session() string {
	// Hand the raw value over; an absent field is an empty id
	return i.SessionID
}

// TranscriptPath returns the session transcript file, empty when absent.
//
// Returns:
//   - string: absolute path of the JSONL transcript
func (i *Input) TranscriptPath() string {
	// Hand the raw value over; an absent field is an empty path
	return i.Transcript
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

// StdinLimits returns the quotas Claude Code piped in on stdin.
// These are free and always current, unlike the API which costs a request.
// A plan without a weekly cap yields a set with only the session limit, and
// that absence is preserved rather than flattened into a zero percent bar.
//
// Returns:
//   - LimitSet: context window plus whichever rate limits were present
func (i *Input) StdinLimits() LimitSet {
	set := LimitSet{
		Context: NewLimit(KindContext, "ctx", i.Progress().Percent, time.Time{}, 0, SourceStdin),
	}

	// Adopt the five-hour bucket when the plan exposes one
	if session, ok := i.RateLimits.FiveHour.Limit(KindSession, "session", SessionWindow); ok {
		set.Session = session
	}
	// Adopt the seven-day bucket when the plan exposes one
	if weekly, ok := i.RateLimits.SevenDay.Limit(KindWeekly, "weekly", WeeklyWindow); ok {
		set.Weekly = weekly
	}

	return set
}

// EffortLevel returns the reasoning effort of the session.
//
// Returns:
//   - string: effort level, empty when not reported
func (i *Input) EffortLevel() string {
	// Return the level verbatim; the renderer maps it to a glyph
	return i.Effort.Level
}

// SessionCost returns the accumulated cost of the session in USD.
//
// Returns:
//   - float64: session cost
func (i *Input) SessionCost() float64 {
	// Return the cumulative cost reported by Claude Code
	return i.Cost.TotalCostUSD
}

// IsFastMode reports whether fast mode is enabled for the session.
//
// Returns:
//   - bool: true when fast mode is on
func (i *Input) IsFastMode() bool {
	// Return the flag as reported by Claude Code
	return i.FastMode
}

// SessionLabel returns the human name of the session.
//
// Returns:
//   - string: session name, empty when unnamed
func (i *Input) SessionLabel() string {
	// Return the name verbatim; truncation is the renderer's concern
	return i.SessionName
}
