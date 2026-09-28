// Package trace records the exact bytes the status line hands the host.
//
// A colour that escapes its segment is either written wrong or cut short, and
// the two cannot be told apart from a screenshot. This appends every frame
// verbatim, with what the write to stdout actually took, so a bled frame can
// be read back byte for byte instead of described.
package trace

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Trace file settings.
const (
	// traceEnv names the file frames are appended to. Unset, nothing is
	// written and nothing is opened.
	traceEnv string = "STATUSLINE_TRACE"
	// maxTraceBytes caps the file so the trace can be left on for a day
	// without filling a disk: a frame is around 1.5 KB and the host redraws
	// every second, so a day of frames is about 130 MB.
	maxTraceBytes int64 = 192 << 20
	// recordSep opens every record. The payload holds newlines and escapes,
	// so a reader cannot split on either; it splits on this and trusts the
	// bytes= field for the payload length.
	recordSep string = "\x1e"
	// columnsEnv is the width the host tells us it has, worth recording: it
	// decides which condensing level the frame was drawn at.
	columnsEnv string = "COLUMNS"
)

// Frame is one status line handed to the host, and what became of the write.
type Frame struct {
	// Out is the bytes handed to stdout, verbatim.
	Out string
	// Written is how many of them the write actually took.
	Written int
	// Err is the write error, nil when the whole frame went out.
	Err error
	// Elapsed is how long the frame took to build.
	Elapsed time.Duration
	// ChipLit is true when the MCP indicator was drawn as a lit chip: the
	// one place in the line that opens a ground of its own mid-segment, and
	// so the only frame worth hunting for in a long trace.
	ChipLit bool
}

// Write appends one frame to the trace file, if one is configured.
//
// Every failure is swallowed: the trace is a diagnostic, and a status line
// that breaks because its own logging broke would be worse than the bug it
// is meant to catch.
//
// Params:
//   - f: the frame and the outcome of its write
func Write(f Frame) {
	path := os.Getenv(traceEnv)
	// No trace asked for: open nothing, write nothing
	if path == "" {
		return
	}
	// Append-only: one process runs per redraw, and a single write under
	// O_APPEND lands whole rather than interleaved with the previous one
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	// An unwritable path is not worth reporting to a status bar
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()
	// A trace left on for days must stop rather than fill the disk
	if info, statErr := file.Stat(); statErr == nil && info.Size() >= maxTraceBytes {
		return
	}
	_, _ = file.WriteString(record(f))
}

// record renders one frame as a trace record: a header naming everything that
// decided the frame, then the frame's own bytes, untouched.
//
// Params:
//   - f: the frame and the outcome of its write
//
// Returns:
//   - string: the record to append
func record(f Frame) string {
	var sb strings.Builder
	sb.Grow(len(f.Out) + 200)
	sb.WriteString(recordSep + "frame ts=" + time.Now().Format(time.RFC3339Nano))
	sb.WriteString(" bytes=" + strconv.Itoa(len(f.Out)))
	sb.WriteString(" wrote=" + strconv.Itoa(f.Written))
	sb.WriteString(" lines=" + strconv.Itoa(strings.Count(f.Out, "\n")))
	sb.WriteString(" cols=" + columns())
	sb.WriteString(" chip=" + chipState(f.ChipLit))
	sb.WriteString(" render=" + strconv.FormatInt(f.Elapsed.Microseconds(), 10) + "us")
	sb.WriteString(" err=" + writeError(f.Err) + "\n")
	sb.WriteString(f.Out)
	return sb.String()
}

// columns returns the width the host announced, or a dash when it announced
// none.
//
// Returns:
//   - string: the raw COLUMNS value, "-" when unset
func columns() string {
	cols := os.Getenv(columnsEnv)
	// An unset width is worth recording as such: it changes the render
	if cols == "" {
		return "-"
	}
	return cols
}

// chipState names whether the MCP indicator was lit.
//
// Params:
//   - lit: true while a call was in flight
//
// Returns:
//   - string: "lit" or "rest"
func chipState(lit bool) string {
	// A lit frame is the rare one; name it so a trace can be grepped for it
	if lit {
		return "lit"
	}
	return "rest"
}

// writeError renders the write error for the header.
//
// Params:
//   - err: error returned by the write to stdout
//
// Returns:
//   - string: the message, "-" when the write succeeded
func writeError(err error) string {
	// A clean write is the normal case and says nothing
	if err == nil {
		return "-"
	}
	// Newlines would break the one-line header
	return strings.ReplaceAll(err.Error(), "\n", " ")
}
