// Package renderer provides status line rendering.
package renderer

import (
	"strings"

	"github.com/florent/status-line/internal/domain/model"
)

// MCP pill constants.
const (
	// mcpLabel names the MCP pill.
	mcpLabel string = "MCP"
	// mcpOffLabel names the count of disabled servers.
	mcpOffLabel string = "off"
	// mcpUnknownPrefix marks a server called but declared by no scope.
	mcpUnknownPrefix string = "+"
)

// mcpItemKind tells how one entry of the MCP pill is drawn at rest.
type mcpItemKind int

// Entry kinds of the MCP pill.
const (
	// mcpItemSource is an enabled-server count for one scope.
	mcpItemSource mcpItemKind = iota
	// mcpItemUnknown is a server nobody declared, named after a "+".
	mcpItemUnknown
	// mcpItemOff is the count of disabled servers, muted and crossed out.
	mcpItemOff
)

// mcpItem is one entry of the MCP pill: "cli 5", "+name" or "off 2".
type mcpItem struct {
	text string
	kind mcpItemKind
	busy bool
}

// mcpItems sums the servers up per scope, in precedence order.
//
// Enabled servers are counted per scope, a scope with none is left out; a
// server no scope declared (seen only through a call) is named on its own
// after a "+"; disabled servers, whatever their scope, make one "off N" at
// the end. An entry is busy when a server it counts is being called, so the
// light lands on the scope that owns the call.
//
// Params:
//   - servers: servers to sum up
//
// Returns:
//   - []mcpItem: entries in drawing order, empty when there is no server
func mcpItems(servers model.MCPServers) []mcpItem {
	var (
		counts  [len(model.MCPSources)]int
		busy    [len(model.MCPSources)]bool
		unknown []mcpItem
		off     int
		offBusy bool
	)
	// File each server under its scope, or aside
	for _, srv := range servers {
		// Disabled servers only count as a whole
		if !srv.Enabled {
			off++
			offBusy = offBusy || srv.Busy
			continue
		}
		idx := mcpSourceIndex(srv.Source)
		// A server outside every scope is named rather than counted
		if idx < 0 {
			unknown = append(unknown, mcpItem{text: mcpUnknownPrefix + srv.Name, kind: mcpItemUnknown, busy: srv.Busy})
			continue
		}
		counts[idx]++
		busy[idx] = busy[idx] || srv.Busy
	}

	items := make([]mcpItem, 0, len(counts)+len(unknown)+1)
	// Scopes in precedence order, the empty ones left out
	for idx, n := range counts {
		// A scope that declares nothing enabled says nothing
		if n == 0 {
			continue
		}
		items = append(items, mcpItem{text: string(model.MCPSources[idx]) + " " + itoa(n), kind: mcpItemSource, busy: busy[idx]})
	}
	items = append(items, unknown...)
	// The disabled servers close the list
	if off > 0 {
		items = append(items, mcpItem{text: mcpOffLabel + " " + itoa(off), kind: mcpItemOff, busy: offBusy})
	}
	return items
}

// mcpSourceIndex finds a scope in precedence order.
//
// Params:
//   - src: scope to find
//
// Returns:
//   - int: index in model.MCPSources, -1 for an unknown scope
func mcpSourceIndex(src model.MCPSource) int {
	// A handful of scopes: a linear scan is the cheapest lookup
	for idx, known := range model.MCPSources {
		// Match the scope by its name
		if known == src {
			return idx
		}
	}
	return -1
}

// renderMCPPill renders every MCP server in one two-part pill.
//
// Left, the bold white label on dark teal; an arrow hands over to the light
// teal body: enabled servers counted per scope ("cli 5 · user 1"), a server
// no scope declares named after a "+", then "off N" muted and crossed out
// for the disabled ones, all divided by a middle dot. The entry that owns a
// server being called lights up as a dark teal chip with bold white ink,
// the label's own colours. No server, no pill.
//
// Params:
//   - sb: string builder to write to
//   - servers: list of MCP servers
func (r *Powerline) renderMCPPill(sb *strings.Builder, servers model.MCPServers) {
	items := mcpItems(servers)
	// Nothing configured draws nothing
	if len(items) == 0 {
		return
	}

	sb.WriteString(" " + FgMCPEnabledText + LeftRound + Reset)
	sb.WriteString(BgMCPLabel + FgWhite + Bold + " " + mcpLabel + " " + Reset)
	sb.WriteString(BgMCPEnabled + FgMCPEnabledText + glyphs.MCPArrow)
	// Each entry in its own style, the arrow opening the list
	for idx, item := range items {
		ink := FgMCPEnabledText
		// The disabled count is muted, its divider too
		if item.kind == mcpItemOff {
			ink = FgMCPMuted
		}
		// The arrow opens the list, a dot divides the rest
		if idx == 0 {
			sb.WriteString(" ")
		} else {
			sb.WriteString(ink + mcpSeparator)
		}
		writeMCPItem(sb, item, ink)
	}
	sb.WriteString(" " + Reset)
	sb.WriteString(FgMCPEnabled + RightRound + Reset)
}

// writeMCPItem writes one entry inside the pill's body.
//
// Params:
//   - sb: string builder to write to
//   - item: entry to write
//   - ink: ink of the entry at rest
func writeMCPItem(sb *strings.Builder, item mcpItem, ink string) {
	// An entry owning a call pops out as a chip in the label's colours
	if item.busy {
		sb.WriteString(BgMCPLabel + FgWhite + Bold + item.text + Reset + BgMCPEnabled)
		return
	}
	// The disabled count is crossed out as well as muted
	if item.kind == mcpItemOff {
		ink += StrikeMCP
	}
	sb.WriteString(ink + item.text + Reset + BgMCPEnabled)
}
