// Command widths renders line one of the sample state at several terminal
// widths, to see the adaptive condensing at work.
//
//	go run ./demo/widths            # coloured, then stripped
//	go run ./demo/widths 100 72     # chosen widths only
package main

import (
	"fmt"
	"os"
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
	widths := []int{200, 160, 120, 100, 80, 60}
	// Widths given on the command line replace the default set
	if len(os.Args) > 1 {
		widths = widths[:0]
		for _, arg := range os.Args[1:] {
			w, err := strconv.Atoi(arg)
			// Skip what is not a width
			if err != nil {
				continue
			}
			widths = append(widths, w)
		}
	}
	data := sample.Data()
	r := renderer.NewPowerline()
	for _, w := range widths {
		data.Terminal.Width = w
		start := time.Now()
		out := r.Render(data)
		took := time.Since(start)
		line1, _, _ := strings.Cut(out, "\n")
		fmt.Printf("── COLUMNS=%d  visible=%d  render=%s\n", w, renderer.VisibleWidth(line1), took)
		fmt.Println(line1)
		fmt.Println(ansi.ReplaceAllString(line1, ""))
	}
}
