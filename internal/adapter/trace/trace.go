// Package trace records the exact bytes the status line hands the host.
//
// A colour that escapes its segment is either written wrong or cut short, and
// the two cannot be told apart from a screenshot. This appends a frame
// verbatim, with what the write to stdout actually took, so a bled frame can
// be read back byte for byte instead of described.
//
// Only a frame that can prove something is kept — see Frame.worthKeeping. The
// bar redraws every second and is almost always right; a trace of those
// frames is a disk full of the bar working.
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
	// maxTraceBytes is the backstop under the filter: with only the frames
	// that prove something kept, a day costs kilobytes, but a session that
	// somehow made every frame interesting must still stop rather than fill
	// a disk.
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
	// Budget is the cells the condenser was told line one could take. It is
	// also the width the host lays the status line out in, so a frame wider
	// than this is one the host truncates.
	Budget int
	// Emitted is the cells line one actually takes. Line two is never
	// condensed, so only line one has a budget to be judged against.
	Emitted int
}

// overflows reports whether line one came out wider than the host's box.
//
// Returns:
//   - bool: true when the host has to truncate this frame
func (f Frame) overflows() bool {
	// An unknown width is no constraint, and neither is an unmeasured line
	return f.Budget > 0 && f.Emitted > f.Budget
}

// cutShort reports whether the frame went out in part, or not at all.
//
// Returns:
//   - bool: true when the write errored or took fewer bytes than the frame
func (f Frame) cutShort() bool {
	// A writer that took fewer bytes without an error of its own still cut it
	return f.Err != nil || f.Written != len(f.Out)
}

// worthKeeping reports whether the frame can prove anything.
//
// Almost every frame is the bar being right, and a trace of those proves
// nothing while filling a disk: 152 367 records over fifteen hours on the
// workstation, 192 MB, the cap reached — and nine of them carried a lit chip.
// Only four kinds of frame can settle where a colour that escaped its segment
// came from:
//
//   - the MCP chip lit, at any rung of the OS ladder: the one ground in the
//     line that is not its segment's own, so the only one a cut could leave
//     open over the rest of the row;
//   - a write that went out in part, which is the cut being ours;
//   - a write that errored, for the same reason;
//   - a line wider than the host's box, which is the host cutting it.
//
// Returns:
//   - bool: true when the frame is worth the disk it would take
func (f Frame) worthKeeping() bool {
	return f.ChipLit || f.cutShort() || f.overflows()
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
	// A conforming frame at rest proves nothing: not even open the file
	if !f.worthKeeping() {
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
	// The two numbers this whole hunt turned on: what the condenser aimed at,
	// and what line one actually came out as
	sb.WriteString(" budget=" + strconv.Itoa(f.Budget))
	sb.WriteString(" emitted=" + strconv.Itoa(f.Emitted))
	sb.WriteString(" cut=" + yesNo(f.overflows()))
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

// yesNo renders a flag for the header.
//
// Params:
//   - on: the flag
//
// Returns:
//   - string: "yes" or "no"
func yesNo(on bool) string {
	// A grep for cut=yes is the first thing anyone will run on a trace
	if on {
		return "yes"
	}
	return "no"
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
