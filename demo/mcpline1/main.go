// Command mcpline1 shows where the short MCP pill could sit on line one.
//
// Nothing here is wired into the status line: it renders the real line one
// and the real MCP pill of a sample session with the product renderer,
// then splices the pill into line one at five candidate places. Line one is
// rendered for the width left once the pill is in, so the adaptive
// condensing makes room for it the way it would if it were wired.
//
//	go run ./demo/mcpline1              # 160 and 80 columns, idle
//	go run ./demo/mcpline1 -busy        # a call to a cli server in flight
//	go run ./demo/mcpline1 120 100      # chosen widths
package main

import (
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/florent/status-line/demo/sample"
	"github.com/florent/status-line/internal/domain/model"
	r "github.com/florent/status-line/internal/presentation/renderer"
)

// ansi matches the escapes the renderer writes.
var ansi = regexp.MustCompile("\033\\[[0-9;]*m")

// bgBefore matches a background escape ending a string.
var bgBefore = regexp.MustCompile("\033\\[48;[0-9;]*m$")

// placement is one candidate spot for the pill.
type placement struct {
	name string
	// piece is what the spot inserts, given the pill's body
	piece func(body string) string
	// splice puts the piece into line one
	splice func(line, piece string) string
}

// main renders every placement at every width.
func main() {
	busy := flag.Bool("busy", false, "light the cli scope as if one of its servers were being called")
	flag.Parse()
	widths := []int{160, 80}
	// Widths given on the command line replace the default pair
	if flag.NArg() > 0 {
		widths = widths[:0]
		for _, arg := range flag.Args() {
			// Skip what is not a width
			if w, err := strconv.Atoi(arg); err == nil {
				widths = append(widths, w)
			}
		}
	}

	data := sample.Data()
	// A call in flight lights the scope that owns the server
	if *busy {
		data.MCP = data.MCP.WithBusy([]string{"github"})
	}
	pill := mcpPill(data)
	body := pillBody(pill)
	places := placements(data, pill)

	fmt.Println("MCP pill (line two today):")
	fmt.Println(pill)
	fmt.Println()
	for _, w := range widths {
		fmt.Printf("════════ COLUMNS=%d ════════\n", w)
		for _, pl := range places {
			piece := pl.piece(body)
			// Leave line one the room the pill takes
			data.Terminal.Width = w - r.VisibleWidth(piece)
			line := pl.splice(line1(data), piece)
			note := "fits"
			// Say by how much a placement misses the terminal
			if over := r.VisibleWidth(line) - w; over > 0 {
				note = "overflows by " + strconv.Itoa(over)
			}
			fmt.Printf("%s  (%d cells, %s)\n%s\n%s\n\n", pl.name, r.VisibleWidth(line), note, line, ansi.ReplaceAllString(line, ""))
		}
	}
}

// placements lists the candidate spots for the pill on line one.
func placements(data model.StatusLineData, pill string) []placement {
	_, modelFg, _ := r.GetModelColors(data.Model.FullName())
	return []placement{
		{
			name:   "1. inside the OS segment, after the health glyph",
			piece:  func(b string) string { return onGround(b, r.BgWhite, r.FgMCPEnabledText) },
			splice: func(l, p string) string { return insertBeforeSep(l, r.FgWhite, p, false) },
		},
		{
			name:   "2. own segment between the model and the context",
			piece:  segment,
			splice: func(l, p string) string { return insertBeforeSep(l, modelFg, p, true) },
		},
		{
			name:   "3. own segment after the path and git segments",
			piece:  segment,
			splice: func(l, p string) string { return insertBeforeSep(l, r.FgGit, p, true) },
		},
		{
			name:   "4. merged into the context segment",
			piece:  func(b string) string { return onGround(b, r.BgContext, r.FgContextInk) },
			splice: mergeIntoContext,
		},
		{
			name:   "5. at the very end, as the line-two pill",
			piece:  func(string) string { return pill },
			splice: func(l, p string) string { return l + p },
		},
	}
}

// line1 renders line one with the product renderer.
func line1(data model.StatusLineData) string {
	out := r.NewPowerline().Render(data)
	first, _, _ := strings.Cut(out, "\n")
	return first
}

// mcpPill renders the real MCP pill: line two of a state with nothing else.
func mcpPill(data model.StatusLineData) string {
	only := model.StatusLineData{MCP: data.MCP}
	out := r.NewPowerline().Render(only)
	lines := strings.Split(out, "\n")
	return lines[len(lines)-2]
}

// pillBody cuts the light teal body out of the pill: " cli 5 · … ".
func pillBody(pill string) string {
	open := r.BgMCPEnabled + r.FgMCPEnabledText + r.SepRight
	closing := r.Reset + r.FgMCPEnabled + r.RightRound + r.Reset
	start := strings.Index(pill, open) + len(open)
	end := strings.LastIndex(pill, closing)
	return r.BgMCPEnabled + pill[start:end] + r.Reset
}

// onGround moves the body onto another ground and ink, labelled "MCP".
func onGround(body, bg, ink string) string {
	b := strings.ReplaceAll(body, r.BgMCPEnabled, bg)
	b = strings.ReplaceAll(b, r.FgMCPEnabledText, ink)
	return bg + ink + r.SepThinRight + r.Bold + " MCP" + r.Reset + b
}

// segment makes the body a powerline segment of its own, label included.
func segment(body string) string {
	return r.BgMCPEnabled + r.FgMCPEnabledText + r.Bold + " MCP" + r.Reset + body
}

// insertBeforeSep inserts the piece where the segment whose separator is
// drawn in sepFg ends. As a segment of its own, the piece takes over the
// separator and hands over to what followed.
func insertBeforeSep(line, sepFg, piece string, own bool) string {
	sep := sepFg + r.SepRight + r.Reset
	idx := strings.Index(line, sep)
	// No such segment on this line: leave it as it is
	if idx < 0 {
		return line
	}
	start, nextBg := idx, bgBefore.FindString(line[:idx])
	start -= len(nextBg)
	// Inside the segment: the piece goes before its separator
	if !own {
		return line[:start] + piece + line[start:]
	}
	return line[:start] + r.BgMCPEnabled + sep + piece + nextBg + r.FgMCPEnabled + r.SepRight + r.Reset + line[idx+len(sep):]
}

// mergeIntoContext inserts the piece at the end of the context segment.
func mergeIntoContext(line, piece string) string {
	sep := r.FgContext + r.SepRight + r.Reset
	idx := strings.Index(line, sep)
	// No context on this line: leave it as it is
	if idx < 0 {
		return line
	}
	start := idx - len(bgBefore.FindString(line[:idx]))
	// Step back over the segment's closing space
	closing := r.BgContext + " " + r.Reset
	if strings.HasSuffix(line[:start], closing) {
		start -= len(closing)
	}
	return line[:start] + r.BgContext + " " + r.Reset + piece + line[start:]
}
