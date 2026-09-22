// Command widths renders the sample state at several terminal widths, to
// see the adaptive condensing of line one at work. Line two is printed when
// it carries something (the MCP pill, with STATUSLINE_MCP_LINE=2).
//
//	go run ./demo/widths            # coloured, then stripped
//	go run ./demo/widths 100 72     # chosen widths only
//	go run ./demo/widths -light 80  # no scoped quota, branch "main"
package main

import (
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/florent/status-line/demo/sample"
	"github.com/florent/status-line/internal/presentation/renderer"
)

// ansi matches the escapes the renderer writes.
var ansi = regexp.MustCompile("\033\\[[0-9;]*m")

// main renders the sample at each width.
func main() {
	light := flag.Bool("light", false, "a lighter session: no scoped quota, branch main")
	flag.Parse()
	widths := []int{200, 160, 120, 100, 80, 60}
	// Widths given on the command line replace the default set
	if flag.NArg() > 0 {
		widths = widths[:0]
		for _, arg := range flag.Args() {
			w, err := strconv.Atoi(arg)
			// Skip what is not a width
			if err != nil {
				continue
			}
			widths = append(widths, w)
		}
	}
	data := sample.Data()
	// A lighter session, for comparison
	if *light {
		data.Limits.Scoped = nil
		data.Git.Branch = "main"
	}
	r := renderer.NewPowerline()
	for _, w := range widths {
		data.Terminal.Width = w
		start := time.Now()
		out := r.Render(data)
		took := time.Since(start)
		line1, rest, _ := strings.Cut(out, "\n")
		line2 := strings.TrimRight(rest, "\n")
		fmt.Printf("── COLUMNS=%d  visible=%d  render=%s\n", w, renderer.VisibleWidth(line1), took)
		fmt.Println(line1)
		fmt.Println(ansi.ReplaceAllString(line1, ""))
		// Say which segments gave way, and how far
		if shrunk := renderer.Condensed(data); len(shrunk) > 0 {
			fmt.Println("   condensed: " + strings.Join(shrunk, " | "))
		}
		// Line two only carries something when the MCP pill is asked there
		if line2 != "" {
			fmt.Println("   line 2: " + ansi.ReplaceAllString(line2, ""))
		}
	}
}
