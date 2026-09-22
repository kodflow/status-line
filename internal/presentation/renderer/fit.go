// Package renderer provides status line rendering.
package renderer

import (
	"strings"

	"github.com/florent/status-line/internal/domain/model"
)

// Adaptive line constants.
const (
	// lineMargin is the number of columns kept free at the right of line
	// one. The host pads the status line and draws its own notices beside
	// it; a line that ends exactly at the edge wraps on the slightest
	// disagreement about a glyph's width.
	lineMargin int = 4
	// fitPathMax is the path budget once the path has to give way.
	fitPathMax int = 20
	// fitBranchMax is the branch budget, in runes, once it has to give way.
	fitBranchMax int = 20
	// tightPathMax leaves the path its last element only: TruncatePath
	// never cuts below ".../<last>".
	tightPathMax int = 1
	// tightBranchMax is the branch budget of the tight levels.
	tightBranchMax int = 12
	// tightestBranchMax is the branch budget of the last level.
	tightestBranchMax int = 8
	// branchEllipsis ends a shortened branch name.
	branchEllipsis string = "…"
)

// lineFit says how much line one gives up to fit the terminal. The zero
// value draws everything.
type lineFit struct {
	// dropCtxBar draws the context window as its icon and percentage.
	dropCtxBar bool
	// dropScopedBar draws a model-scoped quota as label and percentage.
	dropScopedBar bool
	// dropWeeklyBar draws the weekly quota as label and percentage.
	dropWeeklyBar bool
	// dropSessionBar draws the session quota as its percentage.
	dropSessionBar bool
	// pathMax is the path budget, 0 for the default one.
	pathMax int
	// branchMax is the branch budget in runes, 0 for no limit.
	branchMax int
	// dropCountdowns leaves out the time to each refill.
	dropCountdowns bool
	// shortNames names the weekly and scoped quotas by their initial.
	shortNames bool
	// dropChanges leaves out the added/removed lines, which the git
	// counters already hint at.
	dropChanges bool
	// dropPath leaves out the directory; the branch still says where.
	dropPath bool
	// dropModelIcon leaves the model its name alone.
	dropModelIcon bool
}

// fitLevels are the degradation steps, each one giving up what the
// previous gave up and one thing more, in the order the user ranked them:
// the bars first (context, scoped, weekly, session), then the path, the
// branch and the countdowns. A full line is about 240 cells, and those
// steps bring it to about 120; the last ones, beyond that list, are what
// still has to go for 80 columns: the path to its last element and the
// branch to 12 runes, quota names to their initial, the changes, the path,
// the model icon, and the branch down to 8 runes.
var fitLevels = buildFitLevels()

// buildFitLevels accumulates the degradation steps.
//
// Returns:
//   - []lineFit: levels from the full line to the tightest one
func buildFitLevels() []lineFit {
	steps := []func(*lineFit){
		func(f *lineFit) { f.dropCtxBar = true },
		func(f *lineFit) { f.dropScopedBar = true },
		func(f *lineFit) { f.dropWeeklyBar = true },
		func(f *lineFit) { f.dropSessionBar = true },
		func(f *lineFit) { f.pathMax = fitPathMax },
		func(f *lineFit) { f.branchMax = fitBranchMax },
		func(f *lineFit) { f.dropCountdowns = true },
		func(f *lineFit) { f.pathMax, f.branchMax = tightPathMax, tightBranchMax },
		func(f *lineFit) { f.shortNames = true },
		func(f *lineFit) { f.dropChanges = true },
		func(f *lineFit) { f.dropPath = true },
		func(f *lineFit) { f.dropModelIcon = true },
		func(f *lineFit) { f.branchMax = tightestBranchMax },
	}
	levels := make([]lineFit, 0, len(steps)+1)
	var fit lineFit
	levels = append(levels, fit)
	// Each level is the previous one plus one step
	for _, step := range steps {
		step(&fit)
		levels = append(levels, fit)
	}
	return levels
}

// dropBar reports whether a quota of this kind loses its bar.
//
// Params:
//   - kind: quota kind
//
// Returns:
//   - bool: true when the bar is left out
func (f lineFit) dropBar(kind model.LimitKind) bool {
	// Each kind has its own step
	switch kind {
	// The context window goes first
	case model.KindContext:
		return f.dropCtxBar
	// Then the quota scoped to the model in use
	case model.KindScoped:
		return f.dropScopedBar
	// Then the plan-wide weekly quota
	case model.KindWeekly:
		return f.dropWeeklyBar
	// The session quota beside the model name goes last
	case model.KindSession:
		return f.dropSessionBar
	// Any other kind keeps its bar
	default:
		return false
	}
}

// lineBudget returns the columns line one may take.
//
// Params:
//   - width: terminal width, 0 or less when unknown
//
// Returns:
//   - int: budget in cells, 0 when nothing is to be fitted
func lineBudget(width int) int {
	// An unknown width is no constraint
	if width <= 0 {
		return 0
	}
	return max(width-lineMargin, 1)
}

// fitLine1 renders line one at the first level that fits the budget.
//
// Every level only gives up more than the one before it, so the widths
// never grow along the levels and the first level that fits is found by
// bisection: the full line first (the usual case, one render), then at
// most four more renders over the remaining levels.
//
// Params:
//   - budget: cells line one may take, 0 for no limit
//   - render: draws line one with the given fit
//
// Returns:
//   - string: the rendered line
//   - int: the level chosen, len(fitLevels)-1 when even the last overflows
//   - bool: false when even the last level overflows
func fitLine1(budget int, render func(*strings.Builder, lineFit)) (string, int, bool) {
	draw := func(level int) string {
		var sb strings.Builder
		render(&sb, fitLevels[level])
		return sb.String()
	}
	full := draw(0)
	// No budget, or a full line that fits, needs no search
	if budget <= 0 || VisibleWidth(full) <= budget {
		return full, 0, true
	}
	last := len(fitLevels) - 1
	best, bestLine := last, ""
	lo, hi := 1, last
	// Narrow down to the first level that fits
	for lo <= hi {
		mid := (lo + hi) / 2
		line := draw(mid)
		// A fit moves the search left, an overflow right
		if VisibleWidth(line) <= budget {
			best, bestLine, hi = mid, line, mid-1
		} else {
			lo = mid + 1
		}
	}
	// Nothing fits: keep the tightest line, the host wraps what is left
	if bestLine == "" {
		return draw(last), last, false
	}
	return bestLine, best, true
}

// quotaLabel names a quota inside the model segment.
//
// Params:
//   - limit: quota to name
//
// Returns:
//   - string: full label, or its initial once names have to give way
func (f lineFit) quotaLabel(limit model.Limit) string {
	// The full label reads best whenever there is room for it
	if !f.shortNames {
		return QuotaLabel(limit)
	}
	name := nameWeekly
	// A scoped quota is named by its model family
	if limit.Kind == model.KindScoped {
		name = Capitalise(limit.Label)
	}
	// Nothing to shorten: keep whatever label it has
	if limit.Kind != model.KindWeekly && limit.Kind != model.KindScoped || name == "" {
		return QuotaLabel(limit)
	}
	return Labelled(glyphs.Quota, string([]rune(name)[:1]))
}

// TruncateBranch shortens a branch name to max runes, keeping its start.
//
// Params:
//   - branch: branch name
//   - maxRunes: budget in runes, 0 or less for no limit
//
// Returns:
//   - string: the name, or its start followed by an ellipsis
func TruncateBranch(branch string, maxRunes int) string {
	runes := []rune(branch)
	// No budget, or a name that already fits, stays whole
	if maxRunes <= 0 || len(runes) <= maxRunes {
		return branch
	}
	// The ellipsis takes one of the runes
	if maxRunes == 1 {
		return branchEllipsis
	}
	return string(runes[:maxRunes-1]) + branchEllipsis
}
