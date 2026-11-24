// Package main provides the entry point for the status-line CLI tool.
package main

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/florent/status-line/internal/adapter/detach"
	"github.com/florent/status-line/internal/adapter/health"
	"github.com/florent/status-line/internal/adapter/trace"
	"github.com/florent/status-line/internal/adapter/updater"
	"github.com/florent/status-line/internal/adapter/usage"
	"github.com/florent/status-line/internal/domain/model"
)

// deps is everything run reaches outside itself, so a test can drive a frame
// without a network, a terminal or a real binary to replace.
type deps struct {
	// version is the compiled version, empty for a dev build.
	version string
	// stateDir is the real directory of the binary, empty when unresolved.
	stateDir string
	// checkState reads the cached update state.
	checkState func() updateState
	// spawn starts a detached child with a flag.
	spawn func(flag string)
	// newService builds the renderer pipeline for one payload.
	newService func(*model.Input) generator
	// traceWrite records a frame when tracing is on.
	traceWrite func(trace.Frame)
	// stderr receives panic stacks.
	stderr io.Writer
	// now is the clock.
	now func() time.Time
	// refreshUsage refreshes the usage cache for --refresh-usage.
	refreshUsage func() error
	// refreshHealth refreshes the health cache for --refresh-health.
	refreshHealth func() error
}

// productionDeps wires run to the real binary, the real state and the real
// providers.
//
// Returns:
//   - deps: production dependencies
func productionDeps() deps {
	dir := ""
	// The state lives beside the real binary, never beside a link to it
	if exe, err := updater.Executable(); err == nil {
		dir = filepath.Dir(exe)
	}
	return deps{
		version:    version,
		stateDir:   dir,
		checkState: func() updateState { return cachedState(dir, version, time.Now()) },
		spawn:      func(flag string) { detach.Spawn(flag, updatingEnv) },
		newService: func(in *model.Input) generator { return buildService(in) },
		traceWrite: trace.Write,
		stderr:     os.Stderr,
		now:        time.Now,
		refreshUsage: func() error {
			return usage.NewProvider().Refresh()
		},
		refreshHealth: func() error {
			return health.NewProvider().Refresh()
		},
	}
}

// cachedState reads the update state a frame needs, from the file the
// detached children write. It never touches the network.
//
// Params:
//   - dir: real directory of the binary
//   - v: compiled version
//   - now: current time
//
// Returns:
//   - updateState: the pill and whether a check is due
func cachedState(dir, v string, now time.Time) updateState {
	st := updater.ReadState(dir)
	return updateState{info: st.Pill(v, now), due: st.Due(now)}
}
