package renderer

import (
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/florent/status-line/internal/domain/model"
)

// capGlyphs are the powerline glyphs that divide one block of colour from the
// next: the arrow that hands a segment over and the two rounded pill caps.
// They are the only cells a line draws outside a ground of its own. The thin
// divider is not one of them: it divides two halves of a single segment and
// sits on that segment's ground.
var capGlyphs = map[rune]bool{
	firstRune(SepRight):   true,
	firstRune(SepLeft):    true,
	firstRune(LeftRound):  true,
	firstRune(RightRound): true,
}

// firstRune returns the single rune of a glyph constant.
func firstRune(s string) rune {
	r, _ := utf8.DecodeRuneInString(s)
	return r
}

// sgrState is what a terminal holds while it draws a cell. How a cell looks is
// decided by the escapes that came before it and by nothing else, so this is
// the only thing a test of colour bleeding needs to model.
type sgrState struct {
	bg     string
	fg     string
	bold   bool
	strike bool
}

// cleared reports whether the terminal holds no attribute at all.
func (s sgrState) cleared() bool { return s == sgrState{} }

// cell is one printed cell and the attributes in force on it.
type cell struct {
	r     rune
	state sgrState
}

// cellText returns the visible text of a stretch of cells, for error messages.
func cellText(cells []cell) string {
	var sb strings.Builder
	for _, c := range cells {
		sb.WriteRune(c.r)
	}
	return sb.String()
}

// parseCells replays a rendered line the way a terminal would: it returns the
// cells the line prints, each carrying the attributes in force on it, and the
// state the line leaves behind.
//
// An escape that is not a complete CSI sequence is reported rather than
// skipped: a line cut inside an escape is one of the ways a background gets
// out of its segment, and it must never come from this renderer.
func parseCells(t *testing.T, label, line string) ([]cell, sgrState) {
	t.Helper()
	var state sgrState
	// Escapes are most of a rendered line's bytes; a quarter of them is a
	// close enough guess at the cells to avoid growing the slice twice
	cells := make([]cell, 0, len(line)/4)
	for i := 0; i < len(line); {
		// Anything but an escape is a cell, drawn with the state so far
		if line[i] != escapeByte {
			r, size := utf8.DecodeRuneInString(line[i:])
			// Combining marks and control bytes take no cell of their own
			if RuneWidth(r) > 0 {
				cells = append(cells, cell{r: r, state: state})
			}
			i += size
			continue
		}
		seq := line[i:skipEscape(line, i)]
		final := seq[len(seq)-1]
		// A sequence without its CSI opener or its final byte is a cut escape
		if len(seq) < 3 || seq[1] != '[' || final < 0x40 || final > 0x7E {
			t.Errorf("%s: escape %q at byte %d is cut off", label, seq, i)
			return cells, state
		}
		// Only SGR sequences change how a cell looks
		if final == 'm' {
			state = applySGR(state, seq[2:len(seq)-1])
		}
		i += len(seq)
	}
	return cells, state
}

// applySGR applies the parameters of one SGR sequence to the terminal state.
func applySGR(state sgrState, body string) sgrState {
	// An SGR sequence with no parameter is a full reset
	if body == "" {
		return sgrState{}
	}
	params := strings.Split(body, ";")
	for i := 0; i < len(params); i++ {
		code, err := strconv.Atoi(params[i])
		// A parameter this renderer never emits is left to the terminal
		if err != nil {
			continue
		}
		switch {
		// Full reset
		case code == 0:
			state = sgrState{}
		// Weight and strike, and the codes that close them
		case code == 1:
			state.bold = true
		case code == 9:
			state.strike = true
		case code == 22:
			state.bold = false
		case code == 29:
			state.strike = false
		// Default colours
		case code == 39:
			state.fg = ""
		case code == 49:
			state.bg = ""
		// Extended colours carry their own arguments
		case code == 38, code == 48:
			value, next := extendedColor(params, i)
			// The selector says which channel the colour belongs to
			if code == 38 {
				state.fg = value
			} else {
				state.bg = value
			}
			i = next
		// The sixteen base colours name themselves
		case code >= 30 && code <= 37, code >= 90 && code <= 97:
			state.fg = params[i]
		case code >= 40 && code <= 47, code >= 100 && code <= 107:
			state.bg = params[i]
		}
	}
	return state
}

// extendedColor reads the argument of a 38 or 48 selector and returns the
// colour it names together with the index of its last parameter.
func extendedColor(params []string, i int) (string, int) {
	// A selector at the end of the sequence names no colour
	if i+1 >= len(params) {
		return "", i
	}
	switch params[i+1] {
	// One index into the 256-colour cube
	case "5":
		// The index must actually be there
		if i+2 < len(params) {
			return "5;" + params[i+2], i + 2
		}
	// Three 24-bit components
	case "2":
		// All three must actually be there
		if i+4 < len(params) {
			return "2;" + strings.Join(params[i+2:i+5], ";"), i + 4
		}
	}
	return "", i + 1
}

// checkLine asserts what every rendered line must hold, whatever the data: no
// escape is cut, the line closes every attribute it opened, a strike only ever
// crosses out digits, and no separator inherits weight from the block before
// it. It returns the cells so the caller can go on to the segment invariants.
func checkLine(t *testing.T, label, line string) []cell {
	t.Helper()
	cells, end := parseCells(t, label, line)
	// Whatever is still in force at the newline runs on into the next line
	if !end.cleared() {
		t.Errorf("%s: the line ends with %+v in force, not with a full reset", label, end)
	}
	for idx, c := range cells {
		// The strike escape is only closed by a full reset, so a leak of it
		// would cross out whatever the line draws next
		if c.state.strike && (c.r < '0' || c.r > '9') {
			t.Errorf("%s: cell %d %q is crossed out; only the disabled-server count is", label, idx, string(c.r))
		}
		// A separator is drawn between two blocks and belongs to neither
		if capGlyphs[c.r] && (c.state.bold || c.state.strike) {
			t.Errorf("%s: separator %q at cell %d inherits %+v from the block before it", label, string(c.r), idx, c.state)
		}
	}
	return cells
}

// checkLine1Grounds asserts that every segment of line one keeps one ground.
//
// A powerline segment is one block of colour: the arrow that closes it is
// drawn in the next segment's ground, and every cell between two arrows sits
// on the segment's own. An accent inside a segment — the MCP chip, lit while a
// call is in flight — is a run of cells in the middle of it. It may reach
// neither edge: if it did, the ground it replaced was never given back and the
// rest of the segment wears the accent instead. That is the bug this whole
// file exists for.
func checkLine1Grounds(t *testing.T, label string, cells []cell) {
	t.Helper()
	arrow := firstRune(SepRight)
	start := 0
	// Walk one past the end so the stretch after the last arrow is checked too
	for idx := 0; idx <= len(cells); idx++ {
		// A segment runs up to the arrow that hands over to the next one
		if idx < len(cells) && cells[idx].r != arrow {
			continue
		}
		checkSegmentGround(t, label, cells[start:idx])
		start = idx + 1
	}
}

// checkSegmentGround asserts that one segment of line one opens and closes on
// the same ground, that no cell of it falls back to the terminal's own, and
// that it carries at most one accent run, touching neither edge.
func checkSegmentGround(t *testing.T, label string, seg []cell) {
	t.Helper()
	// The rounded cap that opens line one belongs to no ground
	for len(seg) > 0 && capGlyphs[seg[0].r] {
		seg = seg[1:]
	}
	// Two arrows in a row leave no content to check
	if len(seg) == 0 {
		return
	}
	text := cellText(seg)
	ground := seg[0].state.bg
	// A segment with no ground at all has already lost its colour
	if ground == "" {
		t.Errorf("%s: segment %q opens on the terminal's own ground", label, text)
		return
	}
	// The ground the segment opened on is the one it must hand back
	if last := seg[len(seg)-1].state.bg; last != ground {
		t.Errorf("%s: segment %q opens on ground %q and ends on %q: the accent never gave the ground back", label, text, ground, last)
		return
	}
	runs, inRun := 0, false
	for _, c := range seg {
		// A cell with no ground shows the terminal's own colour through it
		if c.state.bg == "" {
			t.Errorf("%s: cell %q of segment %q sits on the terminal's own ground", label, string(c.r), text)
			return
		}
		// Count the accent runs rather than the accent cells
		if c.state.bg != ground {
			// A run only counts where it begins
			if !inRun {
				runs++
			}
			inRun = true
			continue
		}
		inRun = false
	}
	// One segment holds at most one accent: the MCP chip in the OS segment
	if runs > 1 {
		t.Errorf("%s: segment %q carries %d separate accent grounds, at most one is drawn", label, text, runs)
	}
}

// checkLine2Grounds asserts that line two's pills stay inside their caps.
//
// Line two is a row of pills on the terminal's own ground, separated by bare
// spaces written with no escape at all. A pill that leaks its ground past its
// closing cap therefore paints that gap, and everything after it, in its own
// colour — the same failure as an accent escaping a segment on line one.
func checkLine2Grounds(t *testing.T, label string, cells []cell) {
	t.Helper()
	inside := false
	for idx, c := range cells {
		// A cap is a glyph on the terminal's ground, coloured by its pill
		if capGlyphs[c.r] {
			// A cap drawn on a ground means the pill before it never closed
			if c.state.bg != "" {
				t.Errorf("%s: pill cap %q at cell %d is drawn on ground %q", label, string(c.r), idx, c.state.bg)
			}
			inside = c.r == firstRune(LeftRound)
			continue
		}
		// Inside a pill every cell carries the pill's ground
		if inside {
			// A pill cell with no ground has lost the pill's colour
			if c.state.bg == "" {
				t.Errorf("%s: cell %d %q inside a pill sits on the terminal's own ground", label, idx, string(c.r))
			}
			continue
		}
		// Between two pills there is nothing but an unstyled space
		if !c.state.cleared() || c.r != ' ' {
			t.Errorf("%s: cell %d %q between two pills carries %+v; the gap must be blank", label, idx, string(c.r), c.state)
		}
	}
}

// checkRendered runs every invariant over one full two-line render.
func checkRendered(t *testing.T, label, out string) {
	t.Helper()
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	for idx, line := range lines {
		cells := checkLine(t, label+" line"+itoa(idx+1), line)
		// Line one is a chained ribbon of segments; line two a row of pills
		if idx == 0 {
			checkLine1Grounds(t, label+" line1", cells)
		} else {
			checkLine2Grounds(t, label+" line"+itoa(idx+1), cells)
		}
	}
}

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
