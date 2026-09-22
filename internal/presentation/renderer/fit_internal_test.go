package renderer

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/florent/status-line/internal/domain/model"
)

// busyLine is a real-looking session: every quota with its window, a deep
// path, a long branch and changes. The directory avoids the home prefix so
// the result does not depend on who runs the test.
func busyLine(width int) model.StatusLineData {
	now := time.Now()
	week := 7 * 24 * time.Hour
	return model.StatusLineData{
		Model:  model.ModelInfo{Name: "Opus", Version: "5"},
		Effort: "xhigh",
		Limits: model.LimitSet{
			Context: model.NewLimit(model.KindContext, "context", 42, time.Time{}, 0, model.SourceStdin),
			Session: model.NewLimit(model.KindSession, "session", 34, now.Add(2*time.Hour), 5*time.Hour, model.SourceStdin),
			Weekly:  model.NewLimit(model.KindWeekly, "weekly", 61, now.Add(76*time.Hour), week, model.SourceStdin),
			Scoped:  []model.Limit{model.NewLimit(model.KindScoped, "opus", 48, now.Add(76*time.Hour), week, model.SourceAPI)},
		},
		Icons:    model.IconConfig{OS: true, Model: true, Path: true, Git: true},
		System:   model.SystemInfo{OS: model.OSLinux},
		Health:   model.HealthOK,
		Dir:      "/srv/code/worktrees/status-line-adaptive/internal/presentation/renderer",
		Git:      model.GitStatus{Branch: "feat/adaptive-line-condense-to-terminal-width", Modified: 3, Untracked: 1},
		Changes:  model.CodeChanges{Added: 12, Removed: 3},
		Terminal: model.TerminalInfo{Width: width},
	}
}

// line1At renders line one of data at one fit level, stripped.
func line1At(data model.StatusLineData, fit lineFit) string {
	var sb strings.Builder
	(&Powerline{}).renderLine1Fit(&sb, data, fit)
	return stripSGR(sb.String())
}

func TestFitLevelsOrder(t *testing.T) {
	// Each level gives up exactly one more thing, in the user's order
	want := []lineFit{
		{},
		{dropCtxBar: true},
		{dropCtxBar: true, dropScopedBar: true},
		{dropCtxBar: true, dropScopedBar: true, dropWeeklyBar: true},
		{dropCtxBar: true, dropScopedBar: true, dropWeeklyBar: true, dropSessionBar: true},
	}
	all := want[4]
	all.pathMax = fitPathMax
	want = append(want, all)
	all.branchMax = fitBranchMax
	want = append(want, all)
	all.dropCountdowns = true
	want = append(want, all)
	all.pathMax, all.branchMax = tightPathMax, tightBranchMax
	want = append(want, all)
	all.shortNames = true
	want = append(want, all)
	all.dropChanges = true
	want = append(want, all)
	all.dropPath = true
	want = append(want, all)
	if !reflect.DeepEqual(fitLevels, want) {
		t.Fatalf("fitLevels =\n%+v\nwant\n%+v", fitLevels, want)
	}
}

func TestFitLine1PicksTheFirstLevelThatFits(t *testing.T) {
	// Level n draws 100-5n cells
	render := func(sb *strings.Builder, fit lineFit) {
		for level, f := range fitLevels {
			if f == fit {
				sb.WriteString("\033[1m" + strings.Repeat("x", 100-5*level) + "\033[0m")
				return
			}
		}
	}
	tests := []struct {
		budget, level int
	}{
		{budget: 0, level: 0},
		{budget: 100, level: 0},
		{budget: 99, level: 1},
		{budget: 85, level: 3},
		{budget: 71, level: 6},
		{budget: 1, level: len(fitLevels) - 1},
	}
	for _, tt := range tests {
		line, level := fitLine1(tt.budget, render)
		if level != tt.level {
			t.Errorf("budget %d: level %d, want %d", tt.budget, level, tt.level)
		}
		if got := VisibleWidth(line); tt.budget > 0 && level < len(fitLevels)-1 && got > tt.budget {
			t.Errorf("budget %d: line is %d wide", tt.budget, got)
		}
	}
}

func TestLineBudget(t *testing.T) {
	for width, want := range map[int]int{0: 0, -3: 0, 80: 80 - lineMargin, 3: 1, 120: 120 - lineMargin} {
		if got := lineBudget(width); got != want {
			t.Errorf("lineBudget(%d) = %d, want %d", width, got, want)
		}
	}
}

func TestLine1EachStep(t *testing.T) {
	data := busyLine(0)
	bar := func(line, after string) bool {
		// A bar sits right before the percentage: heavy or light rules
		i := strings.Index(line, after)
		return i > 0 && strings.ContainsAny(line[:i], "━─")
	}
	full := line1At(data, fitLevels[0])
	if !strings.Contains(full, nameContext) || !bar(full, " 42%") {
		t.Errorf("level 0 draws the context bar and name, got %q", full)
	}
	tests := []struct {
		level int
		check func(string) bool
		what  string
	}{
		{1, func(l string) bool {
			return !strings.Contains(l, nameContext) && strings.Contains(l, glyphs.Ctx+" 42%")
		}, "context is icon + percentage"},
		{1, func(l string) bool { return bar(l, "Opus 48%") || strings.Contains(l, "Opus ━") }, "the scoped bar is still there"},
		{2, func(l string) bool { return strings.Contains(l, "Opus 48%") }, "the scoped quota is label + percentage"},
		{2, func(l string) bool { return strings.Count(l, glyphs.Reset) == 3 }, "the countdowns stay"},
		{3, func(l string) bool { return strings.Contains(l, "Weekly 61%") }, "the weekly quota is label + percentage"},
		{3, func(l string) bool { return !strings.Contains(l, "Opus 5 ") || bar(l, "34%") }, "the session bar is still there"},
		{4, func(l string) bool { return strings.Contains(l, glyphs.EffortOn) && !strings.ContainsAny(l, "━─") }, "no bar is left"},
		{4, func(l string) bool { return strings.Contains(l, "presentation/renderer") }, "the path keeps its default budget"},
		{5, func(l string) bool { return strings.Contains(l, " .../renderer ") }, "the path is shortened"},
		{5, func(l string) bool { return strings.Contains(l, "terminal-width") }, "the branch is whole"},
		{6, func(l string) bool { return strings.Contains(l, " feat/adaptive-line-… ") }, "the branch keeps 20 runes"},
		{6, func(l string) bool { return strings.Count(l, glyphs.Reset) == 3 }, "the countdowns stay"},
		{7, func(l string) bool { return !strings.Contains(l, glyphs.Reset) }, "the countdowns are gone"},
		{8, func(l string) bool { return strings.Contains(l, " feat/adapti… ") }, "the branch keeps 12 runes"},
		{9, func(l string) bool { return strings.Contains(l, " W 61%") && strings.Contains(l, " O 48%") }, "quota names are initials"},
		{9, func(l string) bool { return strings.Contains(l, "+12") }, "the changes stay"},
		{10, func(l string) bool { return !strings.Contains(l, "+12") && !strings.Contains(l, "-3") }, "the changes are gone"},
		{10, func(l string) bool { return strings.Contains(l, "renderer") }, "the path stays"},
		{11, func(l string) bool { return !strings.Contains(l, "renderer") && strings.Contains(l, "feat/") }, "the path is gone, the branch stays"},
	}
	for _, tt := range tests {
		if line := line1At(data, fitLevels[tt.level]); !tt.check(line) {
			t.Errorf("level %d: %s, got %q", tt.level, tt.what, line)
		}
	}
	// Every level is narrower than the one before it
	prev := VisibleWidth(full)
	for level := 1; level < len(fitLevels); level++ {
		w := VisibleWidth(line1At(data, fitLevels[level]))
		if w >= prev {
			t.Errorf("level %d is %d wide, not narrower than %d", level, w, prev)
		}
		prev = w
	}
}

func TestRenderLine1FitsTheTerminal(t *testing.T) {
	for _, width := range []int{200, 160, 120, 100, 80} {
		var sb strings.Builder
		(&Powerline{}).renderLine1(&sb, busyLine(width))
		if got := VisibleWidth(sb.String()); got > width-lineMargin {
			t.Errorf("COLUMNS=%d: line one is %d wide, over %d", width, got, width-lineMargin)
		}
	}
	// Without a width the line is drawn whole
	var sb strings.Builder
	(&Powerline{}).renderLine1(&sb, busyLine(0))
	if got, want := stripSGR(sb.String()), line1At(busyLine(0), lineFit{}); got != want {
		t.Errorf("no width: got %q, want the full line %q", got, want)
	}
	// Room enough keeps everything
	sb.Reset()
	(&Powerline{}).renderLine1(&sb, busyLine(1000))
	if !strings.Contains(stripSGR(sb.String()), nameContext) {
		t.Errorf("a wide terminal keeps the full line, got %q", stripSGR(sb.String()))
	}
}

func TestRenderLine1PathlessHandsOverToGit(t *testing.T) {
	var sb strings.Builder
	(&Powerline{}).renderLine1Fit(&sb, busyLine(0), fitLevels[len(fitLevels)-1])
	if !strings.Contains(sb.String(), BgGit+FgContext+SepRight) {
		t.Errorf("without the path the context hands over to git, got %q", sb.String())
	}
	// Outside a repository the path is never given up: it is all there is
	data := busyLine(0)
	data.Git = model.GitStatus{}
	if line := line1At(data, fitLevels[len(fitLevels)-1]); !strings.Contains(line, "renderer") {
		t.Errorf("no repository: the path stays, got %q", line)
	}
}

func TestLine2IsNeverCondensed(t *testing.T) {
	data := busyLine(40)
	data.Tasks = model.TaskBoard{Epics: []model.Epic{sampleEpic()}}
	data.Working = true
	out := (&Powerline{}).Render(data)
	if !strings.Contains(out, "a rather long title that must never be shortened") {
		t.Errorf("line two keeps full task titles at any width, got %q", out)
	}
}

func TestTruncateBranch(t *testing.T) {
	tests := []struct {
		branch string
		max    int
		want   string
	}{
		{"main", 0, "main"},
		{"main", 4, "main"},
		{"feature", 4, "fea…"},
		{"feature", 1, "…"},
		{"féature-été", 5, "féat…"},
		{"feat/adaptive-line-condense", 12, "feat/adapti…"},
	}
	for _, tt := range tests {
		if got := TruncateBranch(tt.branch, tt.max); got != tt.want {
			t.Errorf("TruncateBranch(%q, %d) = %q, want %q", tt.branch, tt.max, got, tt.want)
		}
	}
}

func TestLineFitQuotaLabel(t *testing.T) {
	weekly := model.Limit{Kind: model.KindWeekly}
	scoped := model.Limit{Kind: model.KindScoped, Label: "fable"}
	session := model.Limit{Kind: model.KindSession}
	long, short := lineFit{}, lineFit{shortNames: true}
	if got := long.quotaLabel(weekly); got != QuotaLabel(weekly) {
		t.Errorf("full weekly label = %q", got)
	}
	if got, want := short.quotaLabel(weekly), Labelled(glyphs.Quota, "W"); got != want {
		t.Errorf("short weekly label = %q, want %q", got, want)
	}
	if got, want := short.quotaLabel(scoped), Labelled(glyphs.Quota, "F"); got != want {
		t.Errorf("short scoped label = %q, want %q", got, want)
	}
	if got := short.quotaLabel(session); got != QuotaLabel(session) {
		t.Errorf("session label is never shortened, got %q", got)
	}
	if got := short.quotaLabel(model.Limit{Kind: model.KindScoped}); got != QuotaLabel(model.Limit{Kind: model.KindScoped}) {
		t.Errorf("a nameless scoped quota keeps its label, got %q", got)
	}
}

func TestCompactLabelTextGlyphs(t *testing.T) {
	saved := glyphs
	t.Cleanup(func() { glyphs = saved })
	glyphs = textGlyphs
	ctx := model.Limit{Kind: model.KindContext}
	if got := compactLabel(ctx); got != nameContext {
		t.Errorf("without glyphs the context keeps its name, got %q", got)
	}
}

func BenchmarkRenderLine1(b *testing.B) {
	for _, bc := range []struct {
		name  string
		width int
	}{
		{name: "no-width", width: 0},
		{name: "fits-first", width: 1000},
		{name: "cols-120", width: 120},
		{name: "cols-80", width: 80},
		{name: "overflows-all", width: 20},
	} {
		data := busyLine(bc.width)
		b.Run(bc.name, func(b *testing.B) {
			for b.Loop() {
				var sb strings.Builder
				(&Powerline{}).renderLine1(&sb, data)
			}
		})
	}
}
