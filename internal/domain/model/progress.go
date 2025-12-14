// Package model contains domain entities and value objects.
package model

const (
	thresholdMedium   int = 50
	thresholdHigh     int = 75
	thresholdCritical int = 90
	maxPercent        int = 100
)

// ProgressLevel represents the severity of context usage.
type ProgressLevel int

const (
	LevelLow      ProgressLevel = iota
	LevelMedium
	LevelHigh
	LevelCritical
)

// Progress represents context window usage.
type Progress struct {
	Percent int
}

// NewProgress creates a Progress from token counts.
func NewProgress(totalTokens, contextSize int) Progress {
	if contextSize == 0 {
		return Progress{Percent: 0}
	}
	percent := min(totalTokens*maxPercent/contextSize, maxPercent)
	return Progress{Percent: percent}
}

// Level returns the severity level based on usage percentage.
func (p Progress) Level() ProgressLevel {
	switch {
	case p.Percent < thresholdMedium:
		return LevelLow
	case p.Percent < thresholdHigh:
		return LevelMedium
	case p.Percent < thresholdCritical:
		return LevelHigh
	default:
		return LevelCritical
	}
}
