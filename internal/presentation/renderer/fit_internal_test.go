package renderer

import (
	"os"
	"reflect"
	"regexp"
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

// stripSGR removes every SGR escape from a rendered line.
func stripSGR(s string) string {
	return regexp.MustCompile("\033\\[[0-9;]*m").ReplaceAllString(s, "")
}

// line1At renders line one of data at one fit level, stripped.
func line1At(data model.StatusLineData, fit lineFit) string {
	var sb strings.Builder
	(&Powerline{}).renderLine1Fit(&sb, data, fit)
	return stripSGR(sb.String())
}

// failSteps is the ladder the failure variant of the update pill walks, in
// the order its steps are taken.
var failSteps, _ = buildFitLevels(failedWeights(weightsFromEnv(os.Getenv(weightsEnv))))

// describeSteps names each step in the order it is taken.
func describeSteps(steps []fitStep) []string {
	out := make([]string, 0, len(steps))
	for _, st := range steps {
		out = append(out, st.Describe())
	}
	return out
}

func TestFitStepsFollowTheUsersOrder(t *testing.T) {
	want := []string{
		"context: icon + %",
		"scoped: label + % + countdown",
		"weekly: label + % + countdown",
		"session: % + countdown",
		"update: hidden",
		"path: 20 cells",
		"branch: 20 runes",
		"scoped: label + %",
		"weekly: label + %",
		"session: %",
		"path: last element",
		"branch: 12 runes",
		"scoped: initial + %",
		"weekly: initial + %",
		"changes: hidden",
		"path: hidden (inside a repository)",
		"model: name",
		"branch: 8 runes",
		"os: OS icon alone",
	}
	if got := describeSteps(fitSteps); !reflect.DeepEqual(got, want) {
		t.Errorf("steps =\n%q\nwant\n%q", got, want)
	}
	if len(fitLevels) != len(fitSteps)+1 || fitLevels[0] != (lineFit{}) {
		t.Errorf("fitLevels starts at the full line and adds one state per step")
	}
}

// TestHealthIsTheLastStep pins the one ordering the whole ladder is built
// around: the health light outlives every other step, on both ladders, and
// its weight keeps it there — above the highest rank any other step reaches.
func TestHealthIsTheLastStep(t *testing.T) {
	for name, steps := range map[string][]fitStep{"plain": fitSteps, "update failed": failSteps} {
		last := steps[len(steps)-1]
		if last.seg != segOS || last.level != 1 || !(lineFit{segOS: 1}).dropHealth() {
			t.Errorf("%s ladder: the last step is %s, want the health light", name, last.Describe())
		}
		// No other step touches the OS segment
		for _, st := range steps[:len(steps)-1] {
			if st.seg == segOS {
				t.Errorf("%s ladder: %s comes before the end", name, st.Describe())
			}
		}
	}
	// The weight is what keeps it last: above every other step's rank
	osWeight := condensePolicy[segOS].weight
	for id, pol := range condensePolicy {
		if segID(id) == segOS {
			continue
		}
		if top := (len(pol.levels)-2)*passStride + pol.weight; top >= osWeight {
			t.Errorf("%s reaches rank %d, not below the health light's %d", pol.name, top, osWeight)
		}
	}
	if top := max(condensePolicy[segUpdate].weight, updateFailedWeight); top >= osWeight {
		t.Errorf("the failed update pill reaches rank %d, not below the health light's %d", top, osWeight)
	}
}

// TestFailedUpdateHoldsOutPastTheFirstPass asserts the weight of the pill's
// failure variant: the plain pill gives way inside the first pass, the
// failure variant only after it, so the bars, the path and the branch shrink
// before a pill saying the bar no longer updates itself.
func TestFailedUpdateHoldsOutPastTheFirstPass(t *testing.T) {
	index := func(steps []fitStep) int {
		for i, st := range steps {
			if st.seg == segUpdate {
				return i
			}
		}
		return -1
	}
	if updateFailedWeight <= passStride {
		t.Errorf("updateFailedWeight = %d, want it above the first pass (%d)", updateFailedWeight, passStride)
	}
	plain, failed := index(fitSteps), index(failSteps)
	if plain < 0 || failed < 0 {
		t.Fatalf("the update pill is missing from a ladder: plain %d, failed %d", plain, failed)
	}
	// Every first-pass step comes before the failed pill: only the segments
	// held back by a weight above the first pass may still wait after it
	for _, st := range failSteps[failed+1:] {
		if st.level == 1 && condensePolicy[st.seg].weight < passStride {
			t.Errorf("%s still waits after the failed pill gave way", st.Describe())
		}
	}
	if failed <= plain {
		t.Errorf("the failed pill gives way at step %d, the plain one at %d: it must hold out longer", failed, plain)
	}
	// Each ladder is chosen from the data
	data := busyLine(0)
	data.Update = model.UpdateInfo{Available: true, Version: "v9.9.9", Failed: true}
	if got := levelsFor(data); &got[0] != &failLevels[0] {
		t.Error("a failed update does not walk the failure ladder")
	}
	data.Update.Failed = false
	if got := levelsFor(data); &got[0] != &fitLevels[0] {
		t.Error("a pending update does not walk the plain ladder")
	}
}

func TestUpdatePillGivesWay(t *testing.T) {
	data := busyLine(0)
	data.Update = model.UpdateInfo{Available: true, Version: "v9.9.9"}
	if line := line1At(data, lineFit{}); !strings.HasSuffix(line, "v9.9.9 "+RightRound) {
		t.Errorf("the pill closes the full line, got %q", line)
	}
	if line := line1At(data, lineFit{segUpdate: 1}); strings.Contains(line, "v9.9.9") {
		t.Errorf("the pill is gone at its hidden level, got %q", line)
	}
	// The pill counts as present only when it is drawn
	if !presentSegments(data)[segUpdate] || presentSegments(busyLine(0))[segUpdate] {
		t.Error("presentSegments misreports the update pill")
	}
}

func TestFitPassesLowerEverySegmentOnceBeforeTwice(t *testing.T) {
	// With every weight inside the first pass, all first steps come before
	// any second step, and within a pass the lighter segment goes first
	var weights [segCount]int
	for id := range weights {
		weights[id] = 90 - 10*id
	}
	steps, levels := buildFitLevels(weights)
	for i := 1; i < len(steps); i++ {
		prev, cur := steps[i-1], steps[i]
		if cur.level < prev.level {
			t.Errorf("step %d (%s) is a lower pass than step %d (%s)", i, cur.Describe(), i-1, prev.Describe())
		}
		if cur.level == prev.level && weights[cur.seg] < weights[prev.seg] {
			t.Errorf("step %d (%s) should come before %s: lower weight", i, cur.Describe(), prev.Describe())
		}
	}
	if steps[0].seg != segUpdate {
		t.Errorf("the lightest segment shrinks first, got %s", steps[0].Describe())
	}
	// Each state differs from the previous one by exactly one level
	for i := 1; i < len(levels); i++ {
		diff := 0
		for id := range levels[i] {
			diff += levels[i][id] - levels[i-1][id]
		}
		if diff != 1 {
			t.Errorf("state %d moves %d levels, want 1", i, diff)
		}
	}
}

func TestWeightsFromEnv(t *testing.T) {
	def := weightsFromEnv("")
	for id, pol := range condensePolicy {
		if def[id] != pol.weight {
			t.Errorf("%s: default weight %d, want %d", pol.name, def[id], pol.weight)
		}
	}
	got := weightsFromEnv(" context=95 , weekly=5,bogus=1,path=-1,branch=1000,session=x,model,changes=7=8,scoped=0")
	want := def
	want[segContext], want[segWeekly], want[segScoped] = 95, 5, 0
	if got != want {
		t.Errorf("weightsFromEnv = %v, want %v (bad entries ignored)", got, want)
	}
}

func TestWeightOverrideReordersTheSteps(t *testing.T) {
	w := weightsFromEnv("context=65,weekly=5")
	steps, _ := buildFitLevels(w)
	got := describeSteps(steps)[:6]
	want := []string{
		"weekly: label + % + countdown",
		"scoped: label + % + countdown",
		"session: % + countdown",
		"update: hidden",
		"path: 20 cells",
		"branch: 20 runes",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("first pass = %q, want %q", got, want)
	}
}

func TestSegmentLadders(t *testing.T) {
	data := busyLine(0)
	at := func(seg segID, level int) string {
		var fit lineFit
		fit[seg] = level
		return line1At(data, fit)
	}
	bar := func(line, after string) bool {
		// A bar sits right before the percentage: heavy or light rules
		i := strings.Index(line, after)
		return i > 0 && strings.ContainsAny(line[max(0, i-12):i], "\u2501\u2500")
	}
	tests := []struct {
		seg   segID
		level int
		check func(string) bool
		what  string
	}{
		{segContext, 0, func(l string) bool { return strings.Contains(l, nameContext) && bar(l, " 42%") }, "bar + label + %"},
		{segContext, 1, func(l string) bool {
			return !strings.Contains(l, nameContext) && strings.Contains(l, glyphs.Ctx+" 42%")
		}, "icon + %"},
		{segScoped, 0, func(l string) bool { return bar(l, " 48%") }, "bar"},
		{segScoped, 1, func(l string) bool { return strings.Contains(l, "Opus 48% "+glyphs.Reset) }, "label + % + countdown"},
		{segScoped, 2, func(l string) bool { return strings.Contains(l, "Opus 48% ") && strings.Count(l, glyphs.Reset) == 2 }, "label + %"},
		{segScoped, 3, func(l string) bool { return strings.Contains(l, " O 48%") }, "initial + %"},
		{segWeekly, 1, func(l string) bool { return strings.Contains(l, "Weekly 61% "+glyphs.Reset) }, "label + % + countdown"},
		{segWeekly, 2, func(l string) bool { return strings.Contains(l, "Weekly 61% ") && strings.Count(l, glyphs.Reset) == 2 }, "label + %"},
		{segWeekly, 3, func(l string) bool { return strings.Contains(l, " W 61%") }, "initial + %"},
		{segSession, 0, func(l string) bool { return bar(l, " 34%") }, "bar + % + countdown"},
		{segSession, 1, func(l string) bool { return strings.Contains(l, " 34% "+glyphs.Reset) && !bar(l, " 34%") }, "% + countdown"},
		{segSession, 2, func(l string) bool { return strings.Contains(l, " 34% ") && strings.Count(l, glyphs.Reset) == 2 }, "%"},
		{segPath, 0, func(l string) bool { return strings.Contains(l, "presentation/renderer") }, "default budget"},
		{segPath, 1, func(l string) bool { return strings.Contains(l, " .../renderer ") }, "20 cells"},
		{segPath, 3, func(l string) bool { return !strings.Contains(l, "renderer") && strings.Contains(l, "feat/") }, "hidden, branch kept"},
		{segBranch, 1, func(l string) bool { return strings.Contains(l, " feat/adaptive-line-\u2026 !3 ?1") }, "20 runes, counts kept"},
		{segBranch, 2, func(l string) bool { return strings.Contains(l, " feat/adapti\u2026 !3 ?1") }, "12 runes"},
		{segBranch, 3, func(l string) bool { return strings.Contains(l, " feat/ad\u2026 !3 ?1") }, "8 runes"},
		{segChanges, 1, func(l string) bool { return !strings.Contains(l, "+12") && !strings.Contains(l, "-3") }, "hidden"},
		{segModel, 1, func(l string) bool {
			return !strings.Contains(l, IconModel) && strings.Contains(l, " Opus 5 "+glyphs.EffortOn)
		}, "name, gauge kept"},
	}
	for _, tt := range tests {
		if line := at(tt.seg, tt.level); !tt.check(line) {
			t.Errorf("%s level %d: %s, got %q", condensePolicy[tt.seg].name, tt.level, tt.what, line)
		}
	}
	// No step along the default order widens the line (the bisection
	// relies on it); a step can be a no-op, as the path to its last element
	// when 20 cells already left only that
	first := VisibleWidth(line1At(data, fitLevels[0]))
	prev := first
	for i := 1; i < len(fitLevels); i++ {
		w := VisibleWidth(line1At(data, fitLevels[i]))
		if w > prev {
			t.Errorf("after %s the line is %d wide, wider than %d", fitSteps[i-1].Describe(), w, prev)
		}
		prev = w
	}
	if prev >= first-100 {
		t.Errorf("the whole ladder takes the busy line from %d to %d cells only", first, prev)
	}
}

func TestOSSegmentGivesWayInOrder(t *testing.T) {
	data := busyLine(0)
	osAt := func(level int) string {
		var fit lineFit
		fit[segOS] = level
		head, _, _ := strings.Cut(line1At(data, fit), SepRight)
		return head
	}
	// One step: the health light, last of the whole line
	want := []string{" " + glyphs.Health + " ", " "}
	for level, tail := range want {
		// The OS icon is never given up: it is what says the bar is the bar
		if got, expect := osAt(level), LeftRound+" "+IconLinux+tail; got != expect {
			t.Errorf("os level %d = %q, want %q", level, got, expect)
		}
	}
	// The ladder has exactly as many rungs as the policy declares
	if got := len(condensePolicy[segOS].levels); got != len(want) {
		t.Errorf("the policy declares %d levels, the ladder draws %d", got, len(want))
	}
}

func TestCondensed(t *testing.T) {
	if got := Condensed(busyLine(1000)); len(got) != 0 {
		t.Errorf("a wide terminal condenses nothing, got %q", got)
	}
	light := busyLine(80)
	light.Limits.Scoped = nil
	for _, entry := range Condensed(light) {
		if strings.HasPrefix(entry, "scoped:") {
			t.Errorf("an absent segment is not reported, got %q", entry)
		}
	}
	got := Condensed(busyLine(160))
	if len(got) == 0 || got[0] != "context: icon + %" {
		t.Errorf("at 160 the context shrinks first, got %q", got)
	}
}

func TestFitLine1PicksTheFirstLevelThatFits(t *testing.T) {
	// Level n draws 100-n cells: one cell per rung keeps every level
	// positive however many rungs the policy grows
	render := func(sb *strings.Builder, fit lineFit) {
		for level, f := range fitLevels {
			if f == fit {
				sb.WriteString("\033[1m" + strings.Repeat("x", 100-level) + "\033[0m")
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
		{budget: 85, level: 15},
		{budget: 82, level: 18},
		{budget: 1, level: len(fitLevels) - 1},
	}
	for _, tt := range tests {
		line, level, fits := fitLine1(tt.budget, fitLevels, render)
		if want := tt.budget != 1; fits != want {
			t.Errorf("budget %d: fits = %v, want %v", tt.budget, fits, want)
		}
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
	var short lineFit
	short[segWeekly], short[segScoped] = 3, 3
	long := lineFit{}
	if got := long.quotaLabel(weekly); got != QuotaLabel(weekly) {
		t.Errorf("full weekly label = %q", got)
	}
	if got, want := short.quotaLabel(weekly), Labelled(glyphs.Quota, "W"); got != want {
		t.Errorf("short weekly label = %q, want %q", got, want)
	}
	if got, want := short.quotaLabel(scoped), Labelled(glyphs.Quota, "F"); got != want {
		t.Errorf("short scoped label = %q, want %q", got, want)
	}
	short[segSession] = 2
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
		{name: "cols-100", width: 100},
		{name: "cols-80", width: 80},
		{name: "overflows-all", width: 20},
	} {
		data := busyLine(bc.width)
		// The update pill closes the line at every level that keeps it
		data.Update = model.UpdateInfo{Available: true, Version: "v9.9.9"}
		b.Run(bc.name, func(b *testing.B) {
			for b.Loop() {
				var sb strings.Builder
				(&Powerline{}).renderLine1(&sb, data)
			}
		})
	}
}

// TestLineOneNeverOverflowsTheHostsBox is the acceptance the ladder exists
// for. The host lays the status line out in a box of exactly lineBudget cells
// and truncates every line longer than that, on every redraw — so a line over
// budget is not a cosmetic overflow, it is a frame the host cuts each second.
//
// Below some width no ladder can help: the tightest level still has to draw
// the model, its quotas and the branch. So the assertion is in two halves —
// the line fits wherever the budget allows it, and where it cannot it is the
// tightest line the policy can produce, never one rung short of it.
func TestLineOneNeverOverflowsTheHostsBox(t *testing.T) {
	savedGlyphs := glyphs
	t.Cleanup(func() { glyphs = savedGlyphs })
	for _, set := range []struct {
		name string
		set  GlyphSet
	}{{name: "glyphs:nerd", set: nerdGlyphs}, {name: "glyphs:text", set: textGlyphs}} {
		glyphs = set.set
		for _, update := range updateShapes() {
			shape := func(width int) model.StatusLineData {
				data := busyLine(width)
				data.Update = update.update
				return data
			}
			// What the policy can reach when it has given up everything
			levels := levelsFor(shape(0))
			tightest := VisibleWidth(line1At(shape(0), levels[len(levels)-1]))
			floor := tightest + lineMargin
			label := set.name + " " + update.name
			// A loaded session at 80 columns overflowed once: the ladder
			// has to keep the floor under that
			if floor > 80 {
				t.Errorf("%s: the tightest line is %d cells, so nothing under %d columns fits", label, tightest, floor)
			}
			for width := 40; width <= 200; width++ {
				line1, _, _ := strings.Cut((&Powerline{}).Render(shape(width)), "\n")
				emitted, budget := VisibleWidth(line1), lineBudget(width)
				// Wide enough for the policy to fit: it must fit, which
				// is cut=no in the byte trace
				if width >= floor {
					if emitted > budget {
						t.Errorf("%s COLUMNS=%d: emitted %d cells for a budget of %d", label, width, emitted, budget)
						return
					}
					continue
				}
				// Too narrow for any level: the tightest one, exactly —
				// a wider line here would mean a rung left unused
				if emitted != tightest {
					t.Errorf("%s COLUMNS=%d: emitted %d cells, but the tightest level is %d",
						label, width, emitted, tightest)
					return
				}
			}
		}
	}
}

// measuredFloor is the narrowest terminal, in columns, the busy line fits in
// whole at its tightest level; CLAUDE.md quotes it.
const measuredFloor int = 71

// TestTheFloorIsMeasured pins the narrowest terminal the busy line fits in,
// as the documentation states it: a change in the ladder or in a segment
// moves it, and the documentation has to move with it.
func TestTheFloorIsMeasured(t *testing.T) {
	data := busyLine(0)
	tightest := VisibleWidth(line1At(data, fitLevels[len(fitLevels)-1]))
	if floor := tightest + lineMargin; floor != measuredFloor {
		t.Errorf("the busy line fits from %d columns, the documentation says %d", floor, measuredFloor)
	}
}
