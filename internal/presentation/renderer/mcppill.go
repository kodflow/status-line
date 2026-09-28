// Package renderer provides status line rendering.
package renderer

import (
	"os"
	"strings"

	"github.com/florent/status-line/internal/domain/model"
)

// MCP pill constants.
const (
	// mcpLineEnv names the variable choosing the line of the MCP pill.
	mcpLineEnv string = "STATUSLINE_MCP_LINE"
	// mcpLineTwo keeps the pill on line two.
	mcpLineTwo string = "2"
	// mcpOffMark opens the count of disabled servers.
	mcpOffMark string = "\u00b7"
)

// mcpOnLine2 is true when the MCP servers are asked onto line two, as a
// pill of their own; by default they sit in the OS segment of line one.
// Resolved once at startup; tests replace it.
var mcpOnLine2 = os.Getenv(mcpLineEnv) == mcpLineTwo

// mergeSGR folds several SGR escapes into one sequence.
//
// Two escapes that set a ground and then its ink leave the terminal, for the
// length of the second one, showing the new ground under the old ink. A frame
// the host only ever renders whole makes that window harmless — but a frame
// cut inside it paints the rest of the row that way, and the MCP indicator is
// the one place in the line that opens a ground of its own in the middle of a
// segment. One sequence has no such window: a terminal applies it whole or,
// cut, not at all. It is also shorter, which keeps the frame further from the
// pipe's atomic-write limit.
//
// Params:
//   - escapes: SGR escapes to fold, in the order they apply
//
// Returns:
//   - string: one SGR sequence carrying every parameter
func mergeSGR(escapes ...string) string {
	params := make([]string, 0, len(escapes))
	// Keep the parameters, drop each sequence's opener and final byte
	for _, esc := range escapes {
		params = append(params, strings.TrimSuffix(strings.TrimPrefix(esc, "\033["), "m"))
	}
	return "\033[" + strings.Join(params, ";") + "m"
}

// Openings of the MCP indicator, each folded into one sequence. They are also
// what tells a lit frame apart in a trace, which is why they are named rather
// than built where they are written.
var (
	// mcpLitOpen opens the lit chip: bold white on dark teal. Shared with
	// the line-two pill, which lights up in the same colours.
	mcpLitOpen = mergeSGR(BgMCPLabel, FgWhite, Bold)
	// mcpGlyphOpen opens the glyph at rest: dark teal on the OS white.
	mcpGlyphOpen = mergeSGR(BgWhite, FgMCPOnWhite, Bold)
	// mcpCountOpen opens the count at rest, in the OS ink.
	mcpCountOpen = mergeSGR(BgWhite, FgBlack, Bold)
	// mcpOffOpen opens the disabled suffix, muted on the OS white.
	mcpOffOpen = mergeSGR(BgWhite, FgMCPMutedOnWhite)
)

// ChipLit reports whether a rendered status line drew the MCP indicator lit.
//
// The lit chip is the only place in the line that opens a ground of its own
// mid-segment, so it is the only frame a byte trace has to hunt for. The
// predicate lives here because only this package knows the bytes it writes.
//
// Params:
//   - out: a rendered status line, both rows
//
// Returns:
//   - bool: true when a call was in flight as the line was drawn
func ChipLit(out string) bool {
	// Both the inline chip and the line-two pill light up with this opening
	return strings.Contains(out, mcpLitOpen)
}

// mcpSummary is everything the MCP pill says.
type mcpSummary struct {
	// on counts the enabled servers, undeclared ones being called included
	on int
	// off counts the disabled servers
	off int
	// busy is true while a call to any server is in flight
	busy bool
}

// summarizeMCP counts the servers and tells whether one is being called.
//
// Params:
//   - servers: servers to sum up
//
// Returns:
//   - mcpSummary: counts and call state
func summarizeMCP(servers model.MCPServers) mcpSummary {
	var s mcpSummary
	// One pass: count by state, note any call
	for _, srv := range servers {
		// Enabled and disabled are counted apart
		if srv.Enabled {
			s.on++
		} else {
			s.off++
		}
		s.busy = s.busy || srv.Busy
	}
	return s
}

// writeMCPInline sums the MCP servers up inside the OS segment, on its
// white ground: the MCP glyph in dark teal, the number of enabled servers
// in the OS ink, then, when some are disabled, a muted "·N" crossed out.
// While a call is in flight the glyph and the count become a chip, bold
// white on dark teal, taking exactly the cells they took at rest so the
// line does not shift. No server, nothing.
//
// The lit chip is the only ground in the whole line that is not its
// segment's own, so it is the only one a cut frame could leave open over the
// rest of the row. Its ground, ink and weight are therefore opened in one
// sequence and closed by a bare Reset placed immediately after the last
// payload byte: the stretch of bytes a cut can land in and leave teal behind
// is then exactly the payload plus that four-byte reset, which is as small as
// white-on-teal can be drawn. `TestACutNeverLeavesTealOpenBeyondTheChip`
// measures that stretch and fails if it grows.
//
// Params:
//   - sb: string builder to write to
//   - s: servers summed up
func writeMCPInline(sb *strings.Builder, s mcpSummary) {
	// Nothing configured draws nothing
	if s.on+s.off == 0 {
		return
	}
	// Lit: one chip for glyph and count; at rest: each in its own ink
	if s.busy {
		sb.WriteString(mcpLitOpen + Labelled(glyphs.MCP, itoa(s.on)) + Reset)
	} else {
		// The text set spells the glyph out; it still needs its space
		sb.WriteString(mcpGlyphOpen + glyphs.MCP + " " + Reset)
		sb.WriteString(mcpCountOpen + itoa(s.on) + Reset)
	}
	// The disabled servers are a discreet suffix, never a label of their own
	if s.off > 0 {
		sb.WriteString(mcpOffOpen + " " + mcpOffMark + StrikeMCP + itoa(s.off) + Reset)
	}
	sb.WriteString(BgWhite + " " + Reset)
}

// renderMCPPill renders the MCP servers as one small pill (line two, when
// STATUSLINE_MCP_LINE=2): the MCP glyph
// and the number of enabled servers (text glyphs: "MCP 7"), then, when some
// are disabled, a muted "·N" crossed out. While a call is in flight the
// whole pill takes the chip colours, bold white on dark teal; the disabled
// count stays, in pale teal. No server, no pill.
//
// Params:
//   - sb: string builder to write to
//   - servers: list of MCP servers
func (r *Powerline) renderMCPPill(sb *strings.Builder, servers model.MCPServers) {
	s := summarizeMCP(servers)
	// Nothing configured draws nothing
	if s.on+s.off == 0 {
		return
	}
	// At rest: dark ink on pale teal; lit: the chip colours
	bg, ink, cap, offInk := BgMCPEnabled, FgMCPEnabledText, FgMCPEnabled, FgMCPMuted
	if s.busy {
		bg, ink, cap, offInk = BgMCPLabel, FgWhite, FgMCPEnabledText, FgMCPEnabled
	}
	sb.WriteString(" " + cap + LeftRound + Reset)
	// Ground and ink in one sequence, as in the inline chip: a pill ground is
	// never the terminal's own, so a cut between the two would show through
	sb.WriteString(mergeSGR(bg, ink, Bold) + " " + Labelled(glyphs.MCP, itoa(s.on)) + Reset)
	// The disabled servers are a discreet suffix, never a label of their own
	if s.off > 0 {
		sb.WriteString(mergeSGR(bg, offInk) + " " + mcpOffMark + StrikeMCP + itoa(s.off) + Reset)
	}
	sb.WriteString(bg + " " + Reset + cap + RightRound + Reset)
}
