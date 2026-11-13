package main

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/florent/status-line/internal/adapter/trace"
	"github.com/florent/status-line/internal/adapter/updater"
	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

// frameText is a frame with escapes, as the renderer emits.
const frameText = "\033[48;5;255m line one \033[0m\n\033[48;5;23m line two \033[0m"

// fakeGen renders a fixed frame, or panics.
type fakeGen struct {
	out   string
	panic bool
	log   *[]string
}

func (g fakeGen) GenerateWithUpdate(_ port.InputProvider, _ model.UpdateInfo) string {
	if g.log != nil {
		*g.log = append(*g.log, "render")
	}
	if g.panic {
		panic("render exploded")
	}
	return g.out
}

// refuseStdin fails the test if anything reads it.
type refuseStdin struct{ t *testing.T }

func (r refuseStdin) Read(_ []byte) (int, error) {
	r.t.Error("stdin was read")
	return 0, errors.New("refused")
}

// onceFailWriter refuses its first write without taking a byte, then accepts.
type onceFailWriter struct {
	failed bool
	got    strings.Builder
}

func (w *onceFailWriter) Write(p []byte) (int, error) {
	if !w.failed {
		w.failed = true
		return 0, errors.New("first write refused")
	}
	return w.got.Write(p)
}

// testDeps builds deps with a writable state dir, a fixed frame and a
// recorded spawn.
func testDeps(t *testing.T, log *[]string) *deps {
	t.Helper()
	t.Setenv(updater.DisableEnv, "")
	return &deps{
		version:    "v1.0.0",
		stateDir:   t.TempDir(),
		checkState: func() updateState { return updateState{} },
		spawn:      func(flag string) { *log = append(*log, "spawn "+flag) },
		newService: func(*model.Input) generator { return fakeGen{out: frameText, log: log} },
		traceWrite: func(trace.Frame) {},
		stderr:     &strings.Builder{},
		now:        time.Now,
	}
}

func TestFlagsNeverReadStdin(t *testing.T) {
	for _, flag := range []string{"--version", "-v", "--check-update", "--apply-update=v9.9.9"} {
		t.Run(flag, func(t *testing.T) {
			var log []string
			d := testDeps(t, &log)
			// The updater flags stop at the opt-out before any network
			t.Setenv(updater.DisableEnv, "1")
			var out strings.Builder
			if code := run([]string{flag}, refuseStdin{t}, &out, d); code != 0 {
				t.Errorf("exit %d, want 0", code)
			}
			if len(log) != 0 {
				t.Errorf("a flag rendered or spawned: %v", log)
			}
			if strings.HasPrefix(flag, "-v") || flag == "--version" {
				if out.String() != "status-line v1.0.0\n" {
					t.Errorf("--version printed %q", out.String())
				}
			} else if out.Len() != 0 {
				t.Errorf("%s printed %q", flag, out.String())
			}
		})
	}
}

func TestSpawnComesBeforeRenderAndOnlyWhenDue(t *testing.T) {
	tests := []struct {
		name    string
		state   func() updateState
		optOut  string
		version string
		noDir   bool
		want    []string
	}{
		{name: "due", state: func() updateState { return updateState{due: true} }, version: "v1.0.0",
			want: []string{"spawn --check-update", "render"}},
		{name: "not due", state: func() updateState { return updateState{} }, version: "v1.0.0",
			want: []string{"render"}},
		{name: "broken state counts as due", state: func() updateState { panic("state exploded") }, version: "v1.0.0",
			want: []string{"spawn --check-update", "render"}},
		{name: "opt-out of any value", state: func() updateState { return updateState{due: true} }, optOut: "yes", version: "v1.0.0",
			want: []string{"render"}},
		{name: "dev build", state: func() updateState { return updateState{due: true} }, version: "",
			want: []string{"render"}},
		{name: "no writable directory", state: func() updateState { return updateState{due: true} }, version: "v1.0.0", noDir: true,
			want: []string{"render"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var log []string
			d := testDeps(t, &log)
			t.Setenv(updater.DisableEnv, tt.optOut)
			d.checkState, d.version = tt.state, tt.version
			if tt.noDir {
				d.stateDir = ""
			}
			var out strings.Builder
			if code := run(nil, strings.NewReader("{}"), &out, d); code != 0 {
				t.Errorf("exit %d, want 0", code)
			}
			if strings.Join(log, ",") != strings.Join(tt.want, ",") {
				t.Errorf("calls = %v, want %v", log, tt.want)
			}
			if out.String() != frameText {
				t.Errorf("frame = %q", out.String())
			}
		})
	}
}

func TestPanicsSettleOnWhatWentOut(t *testing.T) {
	t.Run("render panic gives the plain fallback", func(t *testing.T) {
		var log []string
		d := testDeps(t, &log)
		d.newService = func(*model.Input) generator { return fakeGen{panic: true} }
		d.checkState = func() updateState {
			return updateState{info: model.UpdateInfo{Available: true, Version: "v1.1.0"}}
		}
		var out strings.Builder
		if code := run(nil, strings.NewReader("{}"), &out, d); code != 0 {
			t.Errorf("exit %d, want 0", code)
		}
		got := out.String()
		if got != "status-line v1.0.0 : erreur de rendu · mise à jour en attente\n" {
			t.Errorf("fallback = %q", got)
		}
		if strings.Contains(got, "\033") {
			t.Error("the fallback carries an escape sequence")
		}
		if !strings.Contains(d.stderr.(*strings.Builder).String(), "render exploded") {
			t.Error("the panic did not reach stderr")
		}
	})
	t.Run("fallback names no update when none is known", func(t *testing.T) {
		var log []string
		d := testDeps(t, &log)
		d.newService = func(*model.Input) generator { panic("service exploded") }
		var out strings.Builder
		if code := run(nil, strings.NewReader("{}"), &out, d); code != 0 {
			t.Errorf("exit %d, want 0", code)
		}
		if out.String() != "status-line v1.0.0 : erreur de rendu\n" {
			t.Errorf("fallback = %q", out.String())
		}
	})
	t.Run("trace panic after a whole frame exits 0", func(t *testing.T) {
		var log []string
		d := testDeps(t, &log)
		d.traceWrite = func(trace.Frame) { panic("trace exploded") }
		var out strings.Builder
		if code := run(nil, strings.NewReader("{}"), &out, d); code != 0 {
			t.Errorf("exit %d, want 0", code)
		}
		if out.String() != frameText {
			t.Errorf("frame = %q, want it whole and alone", out.String())
		}
	})
}

func TestTheExitCodeFollowsTheBytesThatWentOut(t *testing.T) {
	t.Run("nothing went out: fallback, exit 0", func(t *testing.T) {
		var log []string
		w := &onceFailWriter{}
		if code := run(nil, strings.NewReader("{}"), w, testDeps(t, &log)); code != 0 {
			t.Errorf("exit %d, want 0", code)
		}
		if !strings.HasPrefix(w.got.String(), "status-line v1.0.0 : erreur de rendu") {
			t.Errorf("wrote %q, want the fallback", w.got.String())
		}
	})
	t.Run("whole frame: exit 0", func(t *testing.T) {
		var log []string
		w := &shortWriter{limit: len(frameText)}
		if code := run(nil, strings.NewReader("{}"), w, testDeps(t, &log)); code != 0 {
			t.Errorf("exit %d, want 0", code)
		}
	})
	t.Run("partial frame: exit 1 and no second text", func(t *testing.T) {
		var log []string
		w := &shortWriter{limit: 7, extra: true}
		if code := run(nil, strings.NewReader("{}"), w, testDeps(t, &log)); code != 1 {
			t.Errorf("exit %d, want 1", code)
		}
		if w.got.String() != frameText[:7] {
			t.Errorf("wrote %q, want only the first 7 bytes", w.got.String())
		}
	})
}

// countingTransport records every request that reaches the default transport.
type countingTransport struct{ n atomic.Int32 }

func (c *countingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	c.n.Add(1)
	return nil, errors.New("no network in a frame")
}

func TestAFrameNeverTouchesTheNetwork(t *testing.T) {
	saved := http.DefaultTransport
	tr := &countingTransport{}
	http.DefaultTransport = tr
	defer func() { http.DefaultTransport = saved }()

	var log []string
	d := testDeps(t, &log)
	// A cached newer tag: the frame lights the pill from the file alone
	state := `{"tag":"v1.1.0","checked_at":"2026-11-13T12:00:00Z","next_attempt_at":"2026-11-13T13:00:00Z"}`
	if err := os.WriteFile(filepath.Join(d.stateDir, ".status-line-update.json"), []byte(state), 0o600); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 13, 12, 30, 0, 0, time.UTC)
	d.checkState = func() updateState { return cachedState(d.stateDir, d.version, now) }
	var got model.UpdateInfo
	d.newService = func(*model.Input) generator { return recordGen{got: &got} }

	var out strings.Builder
	if code := run(nil, strings.NewReader("{}"), &out, d); code != 0 {
		t.Errorf("exit %d, want 0", code)
	}
	if tr.n.Load() != 0 {
		t.Errorf("%d requests left the frame", tr.n.Load())
	}
	if len(log) != 0 {
		t.Errorf("spawned before next_attempt_at: %v", log)
	}
	if !got.Available || got.Version != "v1.1.0" {
		t.Errorf("pill = %+v, want lit on v1.1.0", got)
	}
}

// recordGen keeps the update info it was asked to render.
type recordGen struct{ got *model.UpdateInfo }

func (g recordGen) GenerateWithUpdate(_ port.InputProvider, u model.UpdateInfo) string {
	*g.got = u
	return frameText
}
