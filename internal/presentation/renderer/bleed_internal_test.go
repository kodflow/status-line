package renderer

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/florent/status-line/internal/domain/model"
)

// inkSnapshot holds the palette entries that depend on the colour depth.
type inkSnapshot struct {
	haiku, sonnet, opus, fable         string
	haikuTrack, sonnetTrack, opusTrack string
	fableTrack                         string
	epicActive, epicPulse, epicTrack   string
}

// takeInks records the depth-dependent palette as it stands.
func takeInks() inkSnapshot {
	return inkSnapshot{
		haiku: FgHaikuDark, sonnet: FgSonnetDark, opus: FgOpusDark, fable: FgFableDark,
		haikuTrack: fgHaikuTrack, sonnetTrack: fgSonnetTrack, opusTrack: fgOpusTrack,
		fableTrack: fgFableTrack,
		epicActive: FgEpicActive, epicPulse: FgEpicPulse, epicTrack: FgEpicTrack,
	}
}

// putInks restores a recorded palette.
func putInks(s inkSnapshot) {
	FgHaikuDark, FgSonnetDark, FgOpusDark, FgFableDark = s.haiku, s.sonnet, s.opus, s.fable
	fgHaikuTrack, fgSonnetTrack, fgOpusTrack, fgFableTrack = s.haikuTrack, s.sonnetTrack, s.opusTrack, s.fableTrack
	FgEpicActive, FgEpicPulse, FgEpicTrack = s.epicActive, s.epicPulse, s.epicTrack
}

// setColorDepth switches the palette between 24-bit and cube escapes the way
// startup does. The two depths emit escapes of different shapes — 38;2;r;g;b
// against 38;5;n, and a combined 1;38;2;… for the pulse — so both have to be
// walked before the parser's verdict means anything.
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
	FgEpicActive = pickInk(epicActiveTrue, epicActive256)
	FgEpicPulse = pickInk(epicPulseTrue, epicPulse256)
	FgEpicTrack = pickInk(epicTrackTrue, epicTrack256)
}

// namedMCP is one server list the indicator has to draw.
type namedMCP struct {
	name string
	list model.MCPServers
}

// mcpShapes lists every shape of the MCP indicator: nothing, one server, many,
// with and without disabled ones, each at rest, lit by a declared server being
// called, and lit by an undeclared key caught mid-call.
func mcpShapes() []namedMCP {
	shapes := []namedMCP{{name: "mcp:none", list: nil}}
	for _, counts := range [][2]int{{1, 0}, {7, 0}, {7, 2}, {0, 2}, {12, 3}} {
		on, off := counts[0], counts[1]
		label := "mcp:" + itoa(on) + "on" + itoa(off) + "off"
		lit := servers(on, off)
		lit[0].Busy = true
		shapes = append(shapes,
			namedMCP{name: label + "/rest", list: servers(on, off)},
			namedMCP{name: label + "/lit", list: lit},
			namedMCP{name: label + "/ghost", list: servers(on, off).WithBusy([]string{"ghost"})},
		)
	}
	return shapes
}

// namedBoard is one shape of line two, with the update notice that goes with it.
type namedBoard struct {
	name    string
	board   model.TaskBoard
	working bool
	update  model.UpdateInfo
	pulse   int64
}

// epicTasks builds a task list holding one of each status.
func epicTasks() model.TaskList {
	return model.TaskList{Items: []model.TaskItem{
		{ID: "1", Subject: "Poser l'invariant", Status: model.TaskCompleted},
		{ID: "2", Subject: "Traquer la fuite de fond", Status: model.TaskInProgress},
		{ID: "3", Subject: "Attendre la capture", Status: model.TaskWaiting},
		{ID: "4", Subject: "Ouvrir la PR", Status: model.TaskPending},
	}}
}

// boardShapes lists the shapes of line two: empty, one collapsed epic, the
// active epic expanded on each frame of the pulse, and the update notice.
func boardShapes() []namedBoard {
	collapsed := model.TaskBoard{Epics: []model.Epic{{ID: 6, Title: "Bug fond vert OS", Tasks: epicTasks()}}}
	expanded := model.TaskBoard{Epics: []model.Epic{
		{ID: 6, Title: "Bug fond vert OS", Active: true, Tasks: epicTasks(), Subagents: 2},
		{ID: 7, Title: "Papercuts", Tasks: epicTasks()},
	}}
	update := model.UpdateInfo{Available: true, Version: "v0.21.0"}
	return []namedBoard{
		{name: "line2:empty", pulse: 2},
		{name: "line2:collapsed", board: collapsed, pulse: 2},
		{name: "line2:expanded/pulse-on", board: expanded, working: true, pulse: 2},
		{name: "line2:expanded/pulse-off", board: expanded, working: true, pulse: 3},
		{name: "line2:update", board: collapsed, working: true, update: update, pulse: 2},
	}
}

// namedHealth is one state of Claude's services, or the health glyph hidden.
type namedHealth struct {
	name   string
	health model.ServiceHealth
	hide   bool
}

// healthShapes lists every state the health glyph can be in, the hidden one
// included: it sits immediately before the MCP indicator, so it decides what
// the chip inherits.
func healthShapes() []namedHealth {
	return []namedHealth{
		{name: "health:ok", health: model.HealthOK},
		{name: "health:degraded", health: model.HealthDegraded},
		{name: "health:down", health: model.HealthDown},
		{name: "health:unknown", health: model.HealthUnknown},
		{name: "health:hidden", health: model.HealthOK, hide: true},
	}
}

// TestNoGroundEscapesItsSegment walks every combination around the MCP chip
// and asserts that no background ever gets out of the segment or the pill that
// opened it, and that both lines end with a full reset.
//
// The chip is only lit while a call is in flight, which is why the reported
// bleed looked random: the matrix renders both states on purpose, under both
// glyph sets, both colour depths and every width the condenser walks.
func TestNoGroundEscapesItsSegment(t *testing.T) {
	savedGlyphs, savedLine2, savedDepth, savedHidden, savedClock := glyphs, mcpOnLine2, trueColor, hidden, clockNow
	savedInks := takeInks()
	t.Cleanup(func() {
		glyphs, mcpOnLine2, hidden, clockNow = savedGlyphs, savedLine2, savedHidden, savedClock
		trueColor = savedDepth
		putInks(savedInks)
	})

	for _, onLine2 := range []bool{false, true} {
		mcpOnLine2 = onLine2
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
				prefix := "mcpLine2=" + strconv.FormatBool(onLine2) + " " + set.name + " " + depth.name
				// A broken invariant repeats across the whole matrix: the
				// first case to break it is the one worth reading
				if !checkDataMatrix(t, prefix) {
					return
				}
			}
		}
	}
}

// checkDataMatrix renders every shape of the data under the environment in
// force and returns false as soon as one render breaks an invariant.
func checkDataMatrix(t *testing.T, prefix string) bool {
	t.Helper()
	shapes, boards := mcpShapes(), boardShapes()
	for _, health := range healthShapes() {
		hidden = map[string]bool{}
		// The hidden case goes through the same path the variable takes
		if health.hide {
			hidden = map[string]bool{hideHealth: true}
		}
		for _, mcp := range shapes {
			for _, subagents := range []int{0, 3} {
				for _, board := range boards {
					// Four widths, not every one: the condenser's own levels
					// are walked exhaustively by TestEveryFitLevelIsAWholeLine
					for _, width := range []int{0, 200, 100, 60} {
						data := busyLine(width)
						data.Health = health.health
						data.MCP = mcp.list
						data.Tasks = board.board
						data.Tasks.Unattributed = subagents
						data.Working = board.working
						data.Update = board.update
						clockNow = fixedClock(board.pulse)
						label := prefix + " " + health.name + " " + mcp.name +
							" agents:" + itoa(subagents) + " " + board.name + " COLUMNS=" + itoa(width)
						checkRendered(t, label, (&Powerline{}).Render(data))
						// Stop on the first case that breaks an invariant
						if t.Failed() {
							return false
						}
					}
				}
			}
		}
	}
	return true
}

// fixedClock freezes the pulse clock on one second.
func fixedClock(sec int64) func() time.Time {
	return func() time.Time { return time.Unix(sec, 0) }
}

// namedShape is one shape of line one: which segments the data draws at all.
type namedShape struct {
	name string
	make func(int) model.StatusLineData
}

// lineShapes lists the shapes line one takes when segments are missing. Which
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
// is missing altogether: no changes, no git, no path, no quota, no icon. The
// chip sits in the first segment of the chain, and what closes the chain is
// what has to hand the ground back at the other end of it.
func TestNoGroundEscapesAnyLineShape(t *testing.T) {
	savedGlyphs, savedLine2, savedDepth := glyphs, mcpOnLine2, trueColor
	savedInks := takeInks()
	t.Cleanup(func() {
		glyphs, mcpOnLine2 = savedGlyphs, savedLine2
		trueColor = savedDepth
		putInks(savedInks)
	})

	rest := servers(7, 2)
	lit := servers(7, 2)
	lit[0].Busy = true
	for _, onLine2 := range []bool{false, true} {
		mcpOnLine2 = onLine2
		for _, set := range []GlyphSet{nerdGlyphs, textGlyphs} {
			glyphs = set
			for _, depth := range []bool{true, false} {
				setColorDepth(depth)
				for _, shape := range lineShapes() {
					for _, mcp := range []namedMCP{{name: "mcp:none"}, {name: "mcp:rest", list: rest}, {name: "mcp:lit", list: lit}} {
						for _, width := range []int{0, 200, 140, 120, 100, 80, 60, 40} {
							data := shape.make(width)
							data.MCP = mcp.list
							label := "mcpLine2=" + strconv.FormatBool(onLine2) + " " +
								shape.name + " " + mcp.name + " COLUMNS=" + itoa(width)
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
}

// TestEveryFitLevelIsAWholeLine asserts that no condensing level, and no pass
// of the bisection, ever cuts a line: every level is a whole render that holds
// the same invariants, and the line the condenser returns is byte for byte the
// render of the level it says it chose.
func TestEveryFitLevelIsAWholeLine(t *testing.T) {
	withMCPLine(t, false)
	lit := servers(7, 2)
	lit[0].Busy = true
	for _, width := range []int{0, 200, 140, 120, 100, 80, 60, 40, 20, 1} {
		data := busyLine(width)
		data.MCP = lit
		data.Tasks.Unattributed = 3
		// Every level, not only the ones the bisection lands on
		for level := range fitLevels {
			var sb strings.Builder
			(&Powerline{}).renderLine1Fit(&sb, data, fitLevels[level])
			cells := checkLine(t, "level "+itoa(level), sb.String())
			checkLine1Grounds(t, "level "+itoa(level), cells)
		}
		line, level, _ := fitLine1(lineBudget(width), func(sb *strings.Builder, fit lineFit) {
			(&Powerline{}).renderLine1Fit(sb, data, fit)
		})
		var want strings.Builder
		(&Powerline{}).renderLine1Fit(&want, data, fitLevels[level])
		// The condenser picks a level and re-renders; it never trims bytes
		if line != want.String() {
			t.Errorf("COLUMNS=%d: the chosen line is not the whole render of level %d", width, level)
		}
	}
}

// TestLightingTheChipNeverMovesTheLine asserts the other half of the chip's
// contract: lit and at rest it takes exactly the same cells, so a call landing
// mid-frame never shifts the line under the reader's eyes. Checked inside the
// whole line, not on the indicator alone: the condenser picks a level from the
// width, so an indicator one cell wider could also change everything after it.
func TestLightingTheChipNeverMovesTheLine(t *testing.T) {
	savedGlyphs, savedLine2 := glyphs, mcpOnLine2
	t.Cleanup(func() { glyphs, mcpOnLine2 = savedGlyphs, savedLine2 })
	mcpOnLine2 = false
	for _, set := range []GlyphSet{nerdGlyphs, textGlyphs} {
		glyphs = set
		for _, counts := range [][2]int{{1, 0}, {7, 0}, {7, 2}, {0, 2}, {12, 3}} {
			on, off := counts[0], counts[1]
			lit := servers(on, off)
			lit[0].Busy = true
			for _, subagents := range []int{0, 3} {
				for _, width := range []int{0, 200, 140, 120, 100, 80, 60, 40} {
					rest, flare := busyLine(width), busyLine(width)
					rest.MCP, flare.MCP = servers(on, off), lit
					rest.Tasks.Unattributed, flare.Tasks.Unattributed = subagents, subagents
					restLine, _, _ := strings.Cut((&Powerline{}).Render(rest), "\n")
					litLine, _, _ := strings.Cut((&Powerline{}).Render(flare), "\n")
					// A call in flight may recolour cells, never add or drop one
					if VisibleWidth(restLine) != VisibleWidth(litLine) {
						t.Errorf("%don%doff agents:%d COLUMNS=%d: the line is %d cells at rest and %d lit",
							on, off, subagents, width, VisibleWidth(restLine), VisibleWidth(litLine))
					}
				}
			}
		}
	}
}

// TestACutNeverLeavesTealOpenBeyondTheChip measures how much of a frame a cut
// can land in and still leave the chip's teal ground open behind it.
//
// The invariants above prove a *whole* frame is well formed. They say nothing
// about a frame cut short — and the host renders each line up to whatever it
// got, carrying anything still open onto the next line. So for every byte
// offset in line one, this replays the prefix and asks what ground the
// terminal would be left holding.
//
// The answer must be: teal only while the cut falls inside the chip's own
// bytes — one contiguous stretch, exactly the chip's payload plus the
// four-byte reset that closes it. That is irreducible: white on teal cannot be
// drawn without teal being open over the glyph and the count. Anything more
// means a ground and its ink were opened by two escapes instead of one, which
// is the window this file exists to keep shut.
func TestACutNeverLeavesTealOpenBeyondTheChip(t *testing.T) {
	withMCPLine(t, false)
	savedGlyphs := glyphs
	t.Cleanup(func() { glyphs = savedGlyphs })
	// Ask the parser what the chip's ground looks like once applied, rather
	// than spelling its escape out a second time
	teal := replayPrefix(BgMCPLabel).bg
	for _, set := range []struct {
		name string
		set  GlyphSet
	}{
		{name: "glyphs:nerd", set: nerdGlyphs},
		{name: "glyphs:text", set: textGlyphs},
	} {
		glyphs = set.set
		lit := servers(7, 2)
		lit[0].Busy = true
		for _, width := range []int{0, 200, 120, 80, 60, 40} {
			data := busyLine(width)
			data.MCP = lit
			data.Tasks.Unattributed = 3
			line, _, _ := strings.Cut((&Powerline{}).Render(data), "\n")
			label := set.name + " COLUMNS=" + itoa(width)

			payload, drawn := chipPayload(line)
			// The narrowest levels give the indicator up altogether; then
			// there is no accent ground in the line to leave open at all
			if !drawn {
				// Teal in force with no single opening sequence to be found
				// means the ground was opened by something else — separate
				// escapes, which is exactly the window this test closes
				if open := tealOffsets(line, teal); open != 0 {
					t.Errorf("%s: a cut leaves teal open at %d offsets, yet the chip's one opening sequence is nowhere in the line: its ground is being opened apart from its ink",
						label, open)
				}
				continue
			}
			first, last, open := tealRun(line, teal)
			// The stretch is the payload the chip draws and its closing
			// reset, whatever level the chip is drawn at
			if want := len(payload) + len(Reset); open != want {
				t.Errorf("%s: a cut leaves teal open at %d offsets, want %d (payload %q is %d bytes + reset %d)",
					label, open, want, payload, len(payload), len(Reset))
			}
			// And it is one stretch, not scattered pieces of the line
			if last-first+1 != open {
				t.Errorf("%s: the offsets that leave teal open run from %d to %d but there are only %d of them",
					label, first, last, open)
			}
		}
	}
}

// chipPayload returns the bytes the lit chip draws between its single opening
// sequence and the reset that closes it, and whether the chip is lit at all.
func chipPayload(line string) (string, bool) {
	at := strings.Index(line, mcpLitOpen)
	// An unlit line, or one whose level gave the indicator up, has no chip
	if at < 0 {
		return "", false
	}
	rest := line[at+len(mcpLitOpen):]
	end := strings.Index(rest, Reset)
	// The chip always closes; a chip that did not would fail checkLine first
	if end < 0 {
		return "", false
	}
	return rest[:end], true
}

// tealOffsets counts the byte offsets at which a cut would leave the chip's
// ground open.
func tealOffsets(line, teal string) int {
	_, _, open := tealRun(line, teal)
	return open
}

// tealRun returns the first and last byte offset at which a cut leaves the
// chip's ground open, and how many such offsets there are.
func tealRun(line, teal string) (int, int, int) {
	first, last, open := -1, -1, 0
	// Every prefix of the line is a frame the host could have got
	for cut := range len(line) + 1 {
		// Only the chip's ground is an accent; every other ground a cut
		// leaves open is the segment's own
		if replayPrefix(line[:cut]).bg != teal {
			continue
		}
		open++
		if first < 0 {
			first = cut
		}
		last = cut
	}
	return first, last, open
}
