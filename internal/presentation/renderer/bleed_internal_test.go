package renderer

import (
	"strings"
	"testing"

	"github.com/florent/status-line/internal/domain/model"
)

// inkSnapshot holds the palette entries that depend on the colour depth.
type inkSnapshot struct {
	haiku, sonnet, opus, fable         string
	haikuTrack, sonnetTrack, opusTrack string
	fableTrack                         string
}

// takeInks records the depth-dependent palette as it stands.
func takeInks() inkSnapshot {
	return inkSnapshot{
		haiku: FgHaikuDark, sonnet: FgSonnetDark, opus: FgOpusDark, fable: FgFableDark,
		haikuTrack: fgHaikuTrack, sonnetTrack: fgSonnetTrack, opusTrack: fgOpusTrack,
		fableTrack: fgFableTrack,
	}
}

// putInks restores a recorded palette.
func putInks(s inkSnapshot) {
	FgHaikuDark, FgSonnetDark, FgOpusDark, FgFableDark = s.haiku, s.sonnet, s.opus, s.fable
	fgHaikuTrack, fgSonnetTrack, fgOpusTrack, fgFableTrack = s.haikuTrack, s.sonnetTrack, s.opusTrack, s.fableTrack
}

// setColorDepth switches the palette between 24-bit and cube escapes the way
// startup does. The two depths emit escapes of different shapes — 38;2;r;g;b
// against 38;5;n — so both have to be walked before the parser's verdict
// means anything.
func setColorDepth(on bool) {
	trueColor = on
	FgHaikuDark = pickInk(inkHaikuTrue, inkHaiku256)
	FgSonnetDark = pickInk(inkSonnetTrue, inkSonnet256)
	FgOpusDark = pickInk(inkOpusTrue, inkOpus256)
	FgFableDark = pickInk(inkFableTrue, inkFable256)
	fgHaikuTrack = pickInk(trackHaikuTrue, trackHaiku256)
	fgSonnetTrack = pickInk(trackSonnetTrue, trackSonnet256)
	fgOpusTrack = pickInk(trackOpusTrue, trackOpus256)
	fgFableTrack = pickInk(trackFableTrue, trackFable256)
}

// namedUpdate is one state of the update pill that closes the line.
type namedUpdate struct {
	name   string
	update model.UpdateInfo
}

// updateShapes lists every state the update pill can be in: absent, a release
// pending, an apply that failed on it, and checks failing with no release
// known.
func updateShapes() []namedUpdate {
	return []namedUpdate{
		{name: "update:none"},
		{name: "update:pending", update: model.UpdateInfo{Available: true, Version: "v0.21.0"}},
		{name: "update:failed", update: model.UpdateInfo{Available: true, Version: "v0.21.0", Failed: true}},
		{name: "update:failing", update: model.UpdateInfo{Available: true, Failed: true}},
	}
}

// namedHealth is one state of Claude's services, or the health glyph hidden.
type namedHealth struct {
	name   string
	health model.ServiceHealth
	hide   bool
}

// healthShapes lists every state the health glyph can be in, the hidden one
// included: it is the last thing the OS segment draws before its arrow, so it
// decides what the arrow inherits.
func healthShapes() []namedHealth {
	return []namedHealth{
		{name: "health:ok", health: model.HealthOK},
		{name: "health:degraded", health: model.HealthDegraded},
		{name: "health:down", health: model.HealthDown},
		{name: "health:unknown", health: model.HealthUnknown},
		{name: "health:hidden", health: model.HealthOK, hide: true},
	}
}

// TestNoGroundEscapesItsSegment walks every state of the health light and of
// the update pill and asserts that no background ever gets out of the segment
// or the pill that opened it, and that the line ends with a full reset —
// under both glyph sets, both colour depths and every width the condenser
// walks.
func TestNoGroundEscapesItsSegment(t *testing.T) {
	savedGlyphs, savedDepth, savedHidden := glyphs, trueColor, hidden
	savedInks := takeInks()
	t.Cleanup(func() {
		glyphs, hidden = savedGlyphs, savedHidden
		trueColor = savedDepth
		putInks(savedInks)
	})

	for _, set := range []struct {
		name string
		set  GlyphSet
	}{{name: "glyphs:nerd", set: nerdGlyphs}, {name: "glyphs:text", set: textGlyphs}} {
		glyphs = set.set
		for _, depth := range []struct {
			name string
			on   bool
		}{{name: "colors:truecolor", on: true}, {name: "colors:256", on: false}} {
			setColorDepth(depth.on)
			// A broken invariant repeats across the whole matrix: the
			// first case to break it is the one worth reading
			if !checkDataMatrix(t, set.name+" "+depth.name) {
				return
			}
		}
	}
}

// checkDataMatrix renders every shape of the data under the environment in
// force and returns false as soon as one render breaks an invariant.
func checkDataMatrix(t *testing.T, prefix string) bool {
	t.Helper()
	for _, health := range healthShapes() {
		hidden = map[string]bool{}
		// The hidden case goes through the same path the variable takes
		if health.hide {
			hidden = map[string]bool{hideHealth: true}
		}
		for _, update := range updateShapes() {
			// A handful of widths, not every one: the condenser's own levels
			// are walked exhaustively by TestEveryFitLevelIsAWholeLine
			for _, width := range []int{0, 200, 100, 80, 60} {
				data := busyLine(width)
				data.Health = health.health
				data.Update = update.update
				label := prefix + " " + health.name + " " + update.name + " COLUMNS=" + itoa(width)
				checkRendered(t, label, (&Powerline{}).Render(data))
				// Stop on the first case that breaks an invariant
				if t.Failed() {
					return false
				}
			}
		}
	}
	return true
}

// namedShape is one shape of the line: which segments the data draws at all.
type namedShape struct {
	name string
	make func(int) model.StatusLineData
}

// lineShapes lists the shapes the line takes when segments are missing. Which
// segment closes the line decides which write has to hand the ground back, so
// a chain that ends early is a different line, not a shorter one.
func lineShapes() []namedShape {
	return []namedShape{
		{name: "shape:full", make: busyLine},
		{name: "shape:no-changes", make: func(w int) model.StatusLineData {
			data := busyLine(w)
			data.Changes = model.CodeChanges{}
			return data
		}},
		{name: "shape:no-git", make: func(w int) model.StatusLineData {
			data := busyLine(w)
			data.Git = model.GitStatus{}
			return data
		}},
		{name: "shape:no-git-no-changes", make: func(w int) model.StatusLineData {
			data := busyLine(w)
			data.Git, data.Changes = model.GitStatus{}, model.CodeChanges{}
			return data
		}},
		{name: "shape:os-and-model-only", make: func(w int) model.StatusLineData {
			data := busyLine(w)
			data.Git, data.Changes, data.Dir = model.GitStatus{}, model.CodeChanges{}, ""
			data.Limits = model.LimitSet{}
			return data
		}},
		{name: "shape:no-quotas", make: func(w int) model.StatusLineData {
			data := busyLine(w)
			data.Limits = model.LimitSet{}
			return data
		}},
		{name: "shape:no-icons-unknown-model", make: func(w int) model.StatusLineData {
			data := busyLine(w)
			data.Icons = model.IconConfig{}
			data.Model = model.ModelInfo{Name: "Weird", Version: "9"}
			data.Effort = ""
			return data
		}},
	}
}

// TestNoGroundEscapesAnyLineShape asserts the same invariants when a segment
// is missing altogether: no changes, no git, no path, no quota, no icon. What
// closes the chain is what has to hand the ground back to the terminal before
// the update pill, so a chain that ends early is a different line.
func TestNoGroundEscapesAnyLineShape(t *testing.T) {
	savedGlyphs, savedDepth := glyphs, trueColor
	savedInks := takeInks()
	t.Cleanup(func() {
		glyphs = savedGlyphs
		trueColor = savedDepth
		putInks(savedInks)
	})

	for _, set := range []GlyphSet{nerdGlyphs, textGlyphs} {
		glyphs = set
		for _, depth := range []bool{true, false} {
			setColorDepth(depth)
			for _, shape := range lineShapes() {
				for _, update := range updateShapes() {
					for _, width := range []int{0, 200, 140, 120, 100, 80, 60, 40} {
						data := shape.make(width)
						data.Update = update.update
						label := shape.name + " " + update.name + " COLUMNS=" + itoa(width)
						checkRendered(t, label, (&Powerline{}).Render(data))
						// Stop on the first case that breaks an invariant
						if t.Failed() {
							return
						}
					}
				}
			}
		}
	}
}

// TestEveryFitLevelIsAWholeLine asserts that no condensing level, and no pass
// of the bisection, ever cuts a line: every level is a whole render that holds
// the same invariants, and the line the condenser returns is byte for byte the
// render of the level it says it chose. Both ladders are walked: the plain
// one, and the one the failure variant of the update pill holds back.
func TestEveryFitLevelIsAWholeLine(t *testing.T) {
	for _, update := range updateShapes() {
		for _, width := range []int{0, 200, 140, 120, 100, 80, 60, 40, 20, 1} {
			data := busyLine(width)
			data.Update = update.update
			levels := levelsFor(data)
			// Every level, not only the ones the bisection lands on
			for level := range levels {
				var sb strings.Builder
				(&Powerline{}).renderLine1Fit(&sb, data, levels[level])
				label := update.name + " level " + itoa(level)
				checkLine1Grounds(t, label, checkLine(t, label, sb.String()))
			}
			line, level, _ := fitLine1(lineBudget(width), levels, func(sb *strings.Builder, fit lineFit) {
				(&Powerline{}).renderLine1Fit(sb, data, fit)
			})
			var want strings.Builder
			(&Powerline{}).renderLine1Fit(&want, data, levels[level])
			// The condenser picks a level and re-renders; it never trims bytes
			if line != want.String() {
				t.Errorf("%s COLUMNS=%d: the chosen line is not the whole render of level %d", update.name, width, level)
			}
		}
	}
}
