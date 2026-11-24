// Package main provides the entry point for the status-line CLI tool.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/florent/status-line/internal/adapter/activity"
	"github.com/florent/status-line/internal/adapter/git"
	"github.com/florent/status-line/internal/adapter/health"
	"github.com/florent/status-line/internal/adapter/mcp"
	"github.com/florent/status-line/internal/adapter/mcpcalls"
	"github.com/florent/status-line/internal/adapter/sessionstate"
	"github.com/florent/status-line/internal/adapter/system"
	"github.com/florent/status-line/internal/adapter/tasks"
	"github.com/florent/status-line/internal/adapter/terminal"
	"github.com/florent/status-line/internal/adapter/trace"
	"github.com/florent/status-line/internal/adapter/updater"
	"github.com/florent/status-line/internal/adapter/usage"
	"github.com/florent/status-line/internal/application"
	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/presentation/renderer"
)

const (
	// checkUpdateFlag runs the detached update check.
	checkUpdateFlag string = "--check-update"
	// forceCheckFlag runs the check by hand, even before next_attempt_at.
	forceCheckFlag string = "--check-update=force"
	// applyUpdateFlag runs the detached install of one release: --apply-update=<tag>.
	applyUpdateFlag string = "--apply-update="
	// updatingEnv marks a detached updater child, so it never spawns another.
	updatingEnv string = "STATUS_LINE_UPDATING"
)

// version is set at build time via ldflags.
// Empty value means development build (no auto-update).
var version string

// main is the entry point of the application.
//
// Returns:
//   - void: exits with the code run settled on
func main() {
	d := productionDeps()
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, &d))
}

// run renders one frame, or runs the one flag it was given.
//
// Flags are handled first, before stdin is read and before the frame's
// recover is armed: a flag never renders. The frame itself never touches the
// network. It reads the update state cached beside the binary, starts the
// detached check when that state says one is due, then renders.
//
// Whatever happens, the exit code follows what reached stdout: nothing at all
// is replaced by a plain fallback frame and exits 0, a whole frame exits 0,
// and a partial one exits 1 so the host drops it rather than draw it.
//
// Params:
//   - args: command line without the program name
//   - stdin: the host's JSON payload
//   - stdout: where the frame goes
//   - d: everything run reaches outside itself
//
// Returns:
//   - int: process exit code
func run(args []string, stdin io.Reader, stdout io.Writer, d *deps) (code int) {
	// A flag is a whole command of its own: no stdin, no frame
	if len(args) > 0 {
		// A detached child has nobody to report to: every flag exits 0
		if runFlag(args[0], stdout, d) {
			return 0
		}
	}

	out := &countingWriter{w: stdout}
	var frame string
	var update model.UpdateInfo
	// A panic anywhere in the frame settles on what already went out
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(d.stderr, "status-line: panic: %v\n%s", r, debug.Stack())
			code = settle(out, frame, d.version, update)
		}
	}()

	state, broken := readUpdateState(d)
	update = state.info
	// Start the check before rendering, and only when it is due: a broken
	// state is a stale one
	if (broken || state.due) && updater.Allowed(d.version, d.stateDir) {
		d.spawn(checkUpdateFlag)
	}

	// A malformed payload must still produce a status line: this process is
	// the shell prompt of a running session, and exiting non-zero replaces it
	// with a raw error on every single redraw
	input := readInput(stdin)
	started := d.now()
	frame = d.newService(input).GenerateWithUpdate(input, update)
	written, err := emit(out, frame)
	line1, _, _ := strings.Cut(frame, "\n")
	d.traceWrite(trace.Frame{
		Out:     frame,
		Written: written,
		Err:     err,
		Elapsed: d.now().Sub(started),
		ChipLit: renderer.ChipLit(frame),
		Budget:  renderer.LineBudget(terminal.NewProvider().Info().Width),
		Emitted: renderer.VisibleWidth(line1),
	})
	// A frame that did not go out whole settles on what did
	if err != nil {
		return settle(out, frame, d.version, update)
	}
	return 0
}

// runFlag runs a command-line flag.
//
// Params:
//   - arg: first command-line argument
//   - stdout: where --version prints
//   - d: version and stderr
//
// Returns:
//   - bool: true when arg was a flag and has been handled
func runFlag(arg string, stdout io.Writer, d *deps) bool {
	switch arg {
	// Print the version and exit
	case "-v", "--version":
		printVersion(stdout, d.version)
	// Refresh the usage cache out of band and exit
	case "--refresh-usage":
		// A refresh failure only leaves the cache stale
		if err := d.refreshUsage(); err != nil {
			fmt.Fprintln(d.stderr, "status-line: usage:", err)
		}
	// Refresh the service health cache out of band and exit
	case health.RefreshFlag:
		// A refresh failure only leaves the cache stale
		if err := d.refreshHealth(); err != nil {
			fmt.Fprintln(d.stderr, "status-line: health:", err)
		}
	// The updater's own flags, or a frame
	default:
		return runUpdateFlag(arg, d)
	}
	return true
}

// runUpdateFlag runs one of the updater's flags.
//
// Params:
//   - arg: first command-line argument
//   - d: version, clock and stderr
//
// Returns:
//   - bool: true when arg was an updater flag and has been handled
func runUpdateFlag(arg string, d *deps) bool {
	switch {
	// Detached check, applying a newer release in the same process
	case arg == checkUpdateFlag:
		runUpdater(d, func(cfg updater.Config) error { return updater.Check(cfg) })
	// Manual check that ignores next_attempt_at
	case arg == forceCheckFlag:
		runUpdater(d, func(cfg updater.Config) error {
			cfg.Force = true
			return updater.Check(cfg)
		})
	// Detached install of one named release
	case strings.HasPrefix(arg, applyUpdateFlag):
		tag := strings.TrimPrefix(arg, applyUpdateFlag)
		runUpdater(d, func(cfg updater.Config) error { return updater.Apply(cfg, tag) })
	// Anything else renders a frame
	default:
		return false
	}
	return true
}

// runUpdater runs one detached updater child against the real binary.
//
// A detached child has nobody to report to: every failure is recorded in the
// state file by the updater itself, and the child exits 0.
//
// Params:
//   - d: compiled version
//   - child: the check or the apply
func runUpdater(d *deps, child func(updater.Config) error) {
	// The opt-out and dev builds stop before anything else
	if os.Getenv(updater.DisableEnv) != "" || d.version == "" {
		return
	}
	exe, err := updater.Executable()
	// Without the real binary there is nothing to update
	if err != nil {
		return
	}
	// The updater records its own failures in the state file
	if err := child(updater.Config{ExecPath: exe, Version: d.version, Now: d.now}); err != nil {
		fmt.Fprintln(d.stderr, "status-line: update:", err)
	}
}

// readUpdateState reads the cached update state under its own recover.
//
// Params:
//   - d: the state reader
//
// Returns:
//   - updateState: what the frame shows and whether a check is due
//   - bool: true when reading panicked, which counts as a due check
func readUpdateState(d *deps) (state updateState, broken bool) {
	defer func() {
		// A state that cannot be read is a stale one
		if r := recover(); r != nil {
			fmt.Fprintf(d.stderr, "status-line: update state panic: %v\n", r)
			state, broken = updateState{}, true
		}
	}()
	return d.checkState(), false
}

// settle turns what reached stdout into an exit code.
//
// Params:
//   - out: the counting writer the frame went through
//   - frame: the frame, empty when rendering never finished
//   - v: compiled version, for the fallback text
//   - update: cached update state, for the fallback text
//
// Returns:
//   - int: 0 for a whole frame or a fallback, 1 for a partial frame
func settle(out *countingWriter, frame, v string, update model.UpdateInfo) int {
	// Nothing went out: a plain line beats a blank bar
	if out.n == 0 {
		// Even the fallback could not be written
		if _, err := io.WriteString(out.w, fallbackFrame(v, update)); err != nil {
			return 1
		}
		return 0
	}
	// The whole frame went out: whatever failed after it does not matter
	if frame != "" && out.n == len(frame) {
		return 0
	}
	// Part of a frame: the host must drop it, and no second text follows
	return 1
}

// fallbackFrame is the plain text shown when rendering failed before any byte
// went out. It carries no escape sequence, so nothing in it can bleed.
//
// Params:
//   - v: compiled version
//   - update: cached update state
//
// Returns:
//   - string: one line of plain text
func fallbackFrame(v string, update model.UpdateInfo) string {
	// A dev build has no version to name
	if v == "" {
		v = "dev"
	}
	line := "status-line " + v + " : erreur de rendu"
	// Say an update is pending only when a newer release is known
	if update.Version != "" {
		line += " \u00b7 mise \u00e0 jour en attente"
	}
	return line + "\n"
}

// emit hands one frame to the host and says whether all of it went out.
//
// A frame that went out in part must never be reported as a success. The host
// reads our stdout to end of file and draws it only when we exit zero, and it
// carries any attribute still open at the end of one line onto the next: a
// frame cut inside the lit MCP chip — the one ground in the line that is not
// its segment's own — would paint the rest of the bar teal until the next
// redraw. A short write with no error of its own is therefore an error here.
//
// Params:
//   - w: where the frame goes, os.Stdout in production
//   - out: the whole frame
//
// Returns:
//   - int: bytes actually written
//   - error: write error, or ErrShortWrite when the frame went out in part
func emit(w io.Writer, out string) (int, error) {
	written, err := io.WriteString(w, out)
	// A writer that takes fewer bytes without saying why still cut the frame
	if err == nil && written != len(out) {
		return written, io.ErrShortWrite
	}
	return written, err
}

// printVersion prints the version information.
// If version is empty (development build), it prints "dev".
//
// Params:
//   - w: where to print
//   - v: compiled version
func printVersion(w io.Writer, v string) {
	// A dev build has no version to name
	if v == "" {
		v = "dev"
	}
	fmt.Fprintln(w, "status-line", v)
}

// readInput reads and parses JSON input from stdin.
// An unreadable or malformed payload yields a zero input rather than an error:
// the renderer degrades to what it can still establish on its own, which is
// strictly better than leaving the session with no status line at all.
//
// Params:
//   - stdin: the host's payload
//
// Returns:
//   - *model.Input: parsed input data, zero valued on failure
func readInput(stdin io.Reader) *model.Input {
	var input model.Input

	data, err := io.ReadAll(stdin)
	// An unreadable stdin leaves the zero input in place
	if err != nil {
		return &input
	}
	// A malformed payload leaves the zero input in place
	if err := json.Unmarshal(data, &input); err != nil {
		return &input
	}
	return &input
}

// buildService creates and wires all dependencies for the status line service.
//
// Params:
//   - input: parsed stdin payload
//
// Returns:
//   - *application.StatusLineService: fully configured service instance
func buildService(input *model.Input) *application.StatusLineService {
	// MCP servers are configured for the session's own directory, while git
	// follows wherever the session is actually working
	sessionDir := input.WorkingDir()
	workDir := activity.Dir(input.TranscriptPath(), sessionDir)
	// Without a real directory git keeps the process working directory
	gitDir := workDir
	if !filepath.IsAbs(gitDir) {
		gitDir = ""
	}
	// The session registry names the host process, whose command line may
	// carry MCP servers; one scan serves both adapters
	session := sessionstate.NewProvider(input.Session())
	deps := application.ServiceDeps{
		Git:      git.NewRepository(gitDir),
		System:   system.NewProvider(),
		Terminal: terminal.NewProvider(),
		MCP:      mcp.NewProvider(sessionDir, session.PID),
		MCPCalls: mcpcalls.NewProvider(input.TranscriptPath(), input.Session()),
		Usage:    usage.NewProvider(),
		Health:   health.NewProvider(),
		Tasks:    tasks.NewProvider(input.Session()),
		Activity: session,
		WorkDir:  workDir,
		Stderr:   os.Stderr,
	}
	// Return service with all adapters injected
	return application.NewStatusLineService(deps, renderer.NewPowerline())
}
