package trace

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// sampleFrame is a one-line frame opening a segment's ground the way the
// renderer writes it, so the record's payload can be checked byte for byte.
const sampleFrame string = "\033[48;5;255;38;5;232;1m \uf17c \033[0m the bar\n"

func TestWriteWithoutATraceFileDoesNothing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(traceEnv, "")
	// Worth keeping, so only the unset variable can stop it being written
	Write(Frame{Out: sampleFrame, Written: len(sampleFrame), Budget: 76, Emitted: 80})
	entries, err := os.ReadDir(dir)
	// Nothing configured must open nothing at all
	if err != nil || len(entries) != 0 {
		t.Errorf("an unset %s wrote something: %v, %v", traceEnv, entries, err)
	}
}

func TestWriteAppendsTheFrameVerbatim(t *testing.T) {
	path := filepath.Join(t.TempDir(), "frames.trace")
	t.Setenv(traceEnv, path)
	t.Setenv(columnsEnv, "213")
	// Kept because line one overflows; the second because its write broke
	Write(Frame{Out: sampleFrame, Written: len(sampleFrame), Elapsed: 37 * time.Millisecond, Budget: 76, Emitted: 80})
	Write(Frame{Out: "cut", Written: 1, Err: errors.New("broken\npipe"), Budget: 76, Emitted: 40})

	raw, err := os.ReadFile(path)
	// The file must exist and hold both records
	if err != nil {
		t.Fatalf("reading the trace: %v", err)
	}
	records := strings.Split(string(raw), recordSep)[1:]
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2: %q", len(records), raw)
	}

	head, payload, _ := strings.Cut(records[0], "\n")
	// The payload is the frame, untouched: that is the whole point
	if payload != sampleFrame {
		t.Errorf("payload = %q, want the frame verbatim %q", payload, sampleFrame)
	}
	// bytes= is what a reader splits the payload on, so it must be exact
	if got := field(t, head, "bytes"); got != strconv.Itoa(len(sampleFrame)) {
		t.Errorf("bytes = %s, want %d", got, len(sampleFrame))
	}
	for name, want := range map[string]string{
		"wrote":   strconv.Itoa(len(sampleFrame)),
		"lines":   "1",
		"cols":    "213",
		"err":     "-",
		"budget":  "76",
		"emitted": "80",
		// 80 cells emitted against a 76-cell budget is a frame the host truncates
		"cut": "yes",
	} {
		if got := field(t, head, name); got != want {
			t.Errorf("%s = %s, want %s", name, got, want)
		}
	}
	// A render time is worth nothing without its unit
	if got := field(t, head, "render"); got != "37000us" {
		t.Errorf("render = %s, want 37000us", got)
	}

	cut, _, _ := strings.Cut(records[1], "\n")
	// A cut frame is the record worth finding: it must say so
	if got := field(t, cut, "wrote"); got != "1" {
		t.Errorf("a cut frame reports wrote = %s, want 1", got)
	}
	// A newline in the error would break the one-line header
	if got := field(t, cut, "err"); got != "broken" || strings.Contains(cut, "\n") {
		t.Errorf("the error must stay on the header line, got %q", cut)
	}
	// A line that fits its budget is a line the host never truncates
	if got := field(t, cut, "cut"); got != "no" {
		t.Errorf("a line inside its budget reports cut = %s, want no", got)
	}
}

func TestWriteStopsAtTheCap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "frames.trace")
	t.Setenv(traceEnv, path)
	// A trace left on for days stops rather than filling the disk
	if err := os.WriteFile(path, make([]byte, maxTraceBytes), 0o600); err != nil {
		t.Fatalf("seeding the trace: %v", err)
	}
	// A frame the filter would drop anyway would pass this for the wrong
	// reason: the cap has to hold against one that is worth keeping
	Write(Frame{Out: sampleFrame, Written: len(sampleFrame), Budget: 76, Emitted: 80})
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() != maxTraceBytes {
		t.Errorf("the trace grew past its cap: %d, want %d", info.Size(), maxTraceBytes)
	}
}

func TestWriteSurvivesAnUnwritablePath(t *testing.T) {
	t.Setenv(traceEnv, filepath.Join(t.TempDir(), "no", "such", "dir", "frames.trace"))
	// Worth keeping, so the open is actually attempted and actually fails.
	// A status line that broke because its own logging broke would be worse
	// than the bug the logging is meant to catch.
	Write(Frame{Out: sampleFrame, Written: len(sampleFrame), Budget: 76, Emitted: 80})
}

// field reads one name=value pair out of a record header.
func field(t *testing.T, head, name string) string {
	t.Helper()
	m := regexp.MustCompile(name + `=(\S+)`).FindStringSubmatch(head)
	if m == nil {
		t.Fatalf("no %s= in %q", name, head)
	}
	return m[1]
}

func TestWriteKeepsOnlyWhatProvesSomething(t *testing.T) {
	broken := errors.New("broken pipe")
	tests := []struct {
		name  string
		frame Frame
		keep  bool
	}{
		{
			name:  "a conforming frame at rest proves nothing",
			frame: Frame{Out: sampleFrame, Written: len(sampleFrame), Budget: 340, Emitted: 176},
		},
		{
			name:  "a write that went out in part is the cut being ours",
			frame: Frame{Out: sampleFrame, Written: 24, Budget: 340, Emitted: 176},
			keep:  true,
		},
		{
			name:  "a write that errored, for the same reason",
			frame: Frame{Out: sampleFrame, Written: len(sampleFrame), Err: broken, Budget: 340, Emitted: 176},
			keep:  true,
		},
		{
			name:  "a line wider than the host's box is the host cutting it",
			frame: Frame{Out: sampleFrame, Written: len(sampleFrame), Budget: 76, Emitted: 80},
			keep:  true,
		},
		{
			name:  "an unknown width is no constraint, so no proof",
			frame: Frame{Out: sampleFrame, Written: len(sampleFrame), Budget: 0, Emitted: 176},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "frames.trace")
			t.Setenv(traceEnv, path)
			Write(tt.frame)
			raw, err := os.ReadFile(path)
			// A frame not worth keeping must not even create the file
			if !tt.keep {
				if err == nil {
					t.Errorf("wrote %d bytes for a frame that proves nothing", len(raw))
				}
				return
			}
			if err != nil {
				t.Fatalf("nothing written for a frame that proves something: %v", err)
			}
			// Kept, it is the whole record: header then the bytes verbatim
			_, payload, found := strings.Cut(string(raw), "\n")
			if !found || payload != tt.frame.Out {
				t.Errorf("payload = %q, want the frame verbatim", payload)
			}
		})
	}
}

func TestWriteKeepsTheHeaderFormatUnchanged(t *testing.T) {
	path := filepath.Join(t.TempDir(), "frames.trace")
	t.Setenv(traceEnv, path)
	t.Setenv(columnsEnv, "344")
	Write(Frame{Out: sampleFrame, Written: 24, Elapsed: 37 * time.Millisecond,
		Budget: 340, Emitted: 176})
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the trace: %v", err)
	}
	head, _, _ := strings.Cut(strings.TrimPrefix(string(raw), recordSep), "\n")
	// The fields and their order are part of the contract, not an
	// implementation detail: chip= is gone with the MCP indicator, and every
	// other field keeps its place so a reader of an older trace still parses
	// this one
	want := `^frame ts=\S+ bytes=\d+ wrote=\d+ lines=\d+ cols=\d+ budget=\d+ emitted=\d+ ` +
		`cut=(yes|no) render=\d+us err=\S+$`
	if !regexp.MustCompile(want).MatchString(head) {
		t.Errorf("header = %q, want it to match %s", head, want)
	}
}
