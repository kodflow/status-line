// Package renderer provides status line rendering.
package renderer

// The machinery the colour-bleed tests are built on: an SGR state machine that
// replays a rendered line the way a terminal would, and the invariants every
// line must hold. The tests that drive it live in bleed_internal_test.go.

import (
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
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
		// An extended colour is the only parameter carrying arguments of its
		// own, so it is the only one that moves the cursor along
		if code == 38 || code == 48 {
			value, next := extendedColor(params, i)
			state, i = withColor(state, code, value), next
			continue
		}
		state = applyParam(state, code, params[i])
	}
	return state
}

// withColor puts an extended colour on the channel its selector names.
func withColor(state sgrState, selector int, value string) sgrState {
	// 38 is the foreground, 48 the background
	if selector == 38 {
		state.fg = value
	} else {
		state.bg = value
	}
	return state
}

// applyParam applies one SGR parameter that carries no argument of its own.
func applyParam(state sgrState, code int, raw string) sgrState {
	switch {
	// Full reset
	case code == 0:
		return sgrState{}
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
	// The sixteen base colours name themselves
	case code >= 30 && code <= 37, code >= 90 && code <= 97:
		state.fg = raw
	case code >= 40 && code <= 47, code >= 100 && code <= 107:
		state.bg = raw
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

// replayPrefix returns the attributes a terminal is left holding after the
// first cut bytes of a line — the state a cut frame would bleed.
func replayPrefix(prefix string) sgrState {
	var state sgrState
	for i := 0; i < len(prefix); {
		// Anything but an escape prints and changes nothing
		if prefix[i] != escapeByte {
			_, size := utf8.DecodeRuneInString(prefix[i:])
			i += size
			continue
		}
		end := skipEscape(prefix, i)
		seq := prefix[i:end]
		// A sequence the cut left incomplete is never applied
		if end > len(prefix) || len(seq) < 3 || seq[len(seq)-1] != 'm' {
			break
		}
		state = applySGR(state, seq[2:len(seq)-1])
		i = end
	}
	return state
}
