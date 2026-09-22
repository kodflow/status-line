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

// mcpOnLine2 is true when the pill is asked to stay on line two; by default
// it closes line one. Resolved once at startup; tests replace it.
var mcpOnLine2 = os.Getenv(mcpLineEnv) == mcpLineTwo

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

// renderMCPPill renders the MCP servers as one small pill: the MCP glyph
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
	sb.WriteString(bg + ink + Bold + " " + Labelled(glyphs.MCP, itoa(s.on)) + Reset)
	// The disabled servers are a discreet suffix, never a label of their own
	if s.off > 0 {
		sb.WriteString(bg + offInk + " " + mcpOffMark + StrikeMCP + itoa(s.off) + Reset)
	}
	sb.WriteString(bg + " " + Reset + cap + RightRound + Reset)
}

// mcpPill renders the MCP pill on its own.
//
// Params:
//   - servers: list of MCP servers
//
// Returns:
//   - string: the pill, empty when there is no server
func (r *Powerline) mcpPill(servers model.MCPServers) string {
	var sb strings.Builder
	r.renderMCPPill(&sb, servers)
	return sb.String()
}
