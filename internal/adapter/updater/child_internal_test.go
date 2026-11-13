package updater

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

// release is a fake GitHub: the API, one asset and its checksum.
type release struct {
	tag       string
	status    int
	header    http.Header
	payload   []byte
	checksum  string
	apiCalls  atomic.Int32
	downloads atomic.Int32
	onAPI     func()
}

// serve starts the fake and points the updater at it for the test.
func (r *release) serve(t *testing.T) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/latest":
			r.apiCalls.Add(1)
			if r.onAPI != nil {
				r.onAPI()
			}
			for k, v := range r.header {
				w.Header()[k] = v
			}
			if r.status != 0 && r.status != http.StatusOK {
				w.WriteHeader(r.status)
				return
			}
			fmt.Fprintf(w, `{"tag_name":%q}`, r.tag)
		case "/asset" + checksumSuffix:
			fmt.Fprintf(w, "%s  asset\n", r.checksum)
		default:
			r.downloads.Add(1)
			_, _ = w.Write(r.payload)
		}
	}))
	t.Cleanup(srv.Close)
	pointAt(t, srv.URL)
}

// pointAt redirects the API and the downloads to base for the test.
func pointAt(t *testing.T, base string) {
	t.Helper()
	savedAPI, savedDL := apiURL, downloadURLFor
	apiURL = base + "/latest"
	downloadURLFor = func(_, _, _, asset string) string {
		if len(asset) > len(checksumSuffix) && asset[len(asset)-len(checksumSuffix):] == checksumSuffix {
			return base + "/asset" + checksumSuffix
		}
		return base + "/asset"
	}
	t.Cleanup(func() { apiURL, downloadURLFor = savedAPI, savedDL })
}

// clock is a settable fake time.
type clock struct{ t time.Time }

func (c *clock) now() time.Time          { return c.t }
func (c *clock) advance(d time.Duration) { c.t = c.t.Add(d) }

// install puts a throwaway binary in a fresh directory and returns a config
// that updates it.
func install(t *testing.T, version string, c *clock) Config {
	t.Helper()
	t.Setenv(DisableEnv, "")
	exe := filepath.Join(t.TempDir(), "status-line")
	if err := os.WriteFile(exe, []byte("original"), executablePerm); err != nil {
		t.Fatal(err)
	}
	return Config{ExecPath: exe, Version: version, Now: c.now}
}

func binary(t *testing.T, cfg Config) string {
	t.Helper()
	raw, err := os.ReadFile(cfg.ExecPath)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

var epoch = time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC)

func TestCheckRecordsNextAttemptBeforeTheRequest(t *testing.T) {
	c := &clock{t: epoch}
	cfg := install(t, "v1.0.0", c)
	r := &release{tag: "v1.0.0"}
	var during State
	r.onAPI = func() { during = ReadState(cfg.dir()) }
	r.serve(t)

	if err := Check(cfg); err != nil {
		t.Fatalf("Check() = %v", err)
	}
	// A child killed mid-request must already have pushed the next attempt out
	if want := epoch.Add(firstBackoff); !during.NextAttemptAt.Equal(want) {
		t.Errorf("next_attempt_at during the request = %v, want %v", during.NextAttemptAt, want)
	}
	st := ReadState(cfg.dir())
	if want := epoch.Add(checkInterval); !st.NextAttemptAt.Equal(want) {
		t.Errorf("next_attempt_at after success = %v, want %v", st.NextAttemptAt, want)
	}
	if !st.CheckedAt.Equal(epoch) || st.Tag != "v1.0.0" || st.CheckError != "" {
		t.Errorf("state after success = %+v", st)
	}
}

func TestCheckBacksOffWhileOffline(t *testing.T) {
	c := &clock{t: epoch}
	cfg := install(t, "v1.0.0", c)
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()
	pointAt(t, srv.URL)

	for i, wait := range []time.Duration{5, 10, 20, 40, 60, 60} {
		if err := Check(cfg); err == nil {
			t.Fatalf("Check() #%d succeeded offline", i)
		}
		st := ReadState(cfg.dir())
		if want := c.t.Add(wait * time.Minute); !st.NextAttemptAt.Equal(want) {
			t.Errorf("failure #%d: next_attempt_at = %v, want +%d min", i+1, st.NextAttemptAt.Sub(c.t), wait)
		}
		if st.CheckError == "" || !st.CheckFailingSince.Equal(epoch) {
			t.Errorf("failure #%d: state = %+v", i+1, st)
		}
		c.t = st.NextAttemptAt
	}
}

func TestCheckHonoursRateLimits(t *testing.T) {
	tests := []struct {
		name   string
		status int
		header http.Header
		want   time.Duration
	}{
		{name: "403 Retry-After seconds", status: http.StatusForbidden, header: http.Header{"Retry-After": {"120"}}, want: 2 * time.Minute},
		{name: "429 Retry-After date", status: http.StatusTooManyRequests, header: http.Header{"Retry-After": {epoch.Add(30 * time.Minute).Format(http.TimeFormat)}}, want: 30 * time.Minute},
		{name: "403 X-RateLimit-Reset", status: http.StatusForbidden, header: http.Header{"X-Ratelimit-Reset": {strconv.FormatInt(epoch.Add(45*time.Minute).Unix(), 10)}}, want: 45 * time.Minute},
		{name: "403 without a hint", status: http.StatusForbidden, want: firstBackoff},
		{name: "500 ignores Retry-After", status: http.StatusInternalServerError, header: http.Header{"Retry-After": {"7200"}}, want: firstBackoff},
		{name: "absurd hint capped", status: http.StatusTooManyRequests, header: http.Header{"Retry-After": {"9999999"}}, want: maxRetryWait},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &clock{t: epoch}
			cfg := install(t, "v1.0.0", c)
			(&release{status: tt.status, header: tt.header}).serve(t)
			if err := Check(cfg); err == nil {
				t.Fatal("Check() succeeded on a refusal")
			}
			if got := ReadState(cfg.dir()).NextAttemptAt.Sub(epoch); got != tt.want {
				t.Errorf("waits %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckRefusesAnInvalidTag(t *testing.T) {
	c := &clock{t: epoch}
	cfg := install(t, "v1.0.0", c)
	r := &release{tag: "v9.0.0-rc1", payload: []byte("x"), checksum: digestOf([]byte("x"))}
	r.serve(t)
	if err := Check(cfg); err == nil {
		t.Fatal("Check() accepted a pre-release tag")
	}
	st := ReadState(cfg.dir())
	if st.CheckError == "" || st.Tag != "" {
		t.Errorf("state = %+v, want a check error and no tag", st)
	}
	if r.downloads.Load() != 0 || binary(t, cfg) != "original" {
		t.Error("an invalid tag was downloaded")
	}
}

func TestCheckAppliesANewerReleaseInTheSameProcess(t *testing.T) {
	c := &clock{t: epoch}
	cfg := install(t, "v1.0.0", c)
	payload := []byte("v1.1.0 binary")
	r := &release{tag: "v1.1.0", payload: payload, checksum: digestOf(payload)}
	r.serve(t)

	// One child: no second spawn, the apply happens under the lock it holds
	if err := Check(cfg); err != nil {
		t.Fatalf("Check() = %v", err)
	}
	if got := binary(t, cfg); got != string(payload) {
		t.Errorf("binary = %q, want the new release", got)
	}
	if r.apiCalls.Load() != 1 || r.downloads.Load() != 1 {
		t.Errorf("api calls %d, downloads %d, want 1 and 1", r.apiCalls.Load(), r.downloads.Load())
	}
	st := ReadState(cfg.dir())
	// The old binary still renders the frames that follow: pill lit
	if info := st.Pill("v1.0.0", c.now()); !info.Available || info.Failed || info.Version != "v1.1.0" {
		t.Errorf("pill for the old binary = %+v, want lit on v1.1.0", info)
	}
	// The new binary catches up with the cache: pill out
	if info := st.Pill("v1.1.0", c.now()); info.Available {
		t.Errorf("pill for the new binary = %+v, want off", info)
	}
}

func TestCheckLeavesWhenTheLockIsHeld(t *testing.T) {
	c := &clock{t: epoch}
	cfg := install(t, "v1.0.0", c)
	r := &release{tag: "v1.1.0"}
	r.serve(t)
	held := withLock(cfg.dir(), func() error {
		if err := Check(cfg); err == nil {
			t.Error("Check() ran while another child held the lock")
		}
		if err := Apply(cfg, "v1.1.0"); err == nil {
			t.Error("Apply() ran while another child held the lock")
		}
		return nil
	})
	if held != nil {
		t.Fatal(held)
	}
	if r.apiCalls.Load() != 0 || r.downloads.Load() != 0 {
		t.Error("a child reached the network without the lock")
	}
}

func TestChildrenStopForOptOutAndDevBuilds(t *testing.T) {
	tests := []struct {
		name    string
		optOut  string
		version string
	}{
		{name: "opt-out of 1", optOut: "1", version: "v1.0.0"},
		{name: "opt-out of another value", optOut: "true", version: "v1.0.0"},
		{name: "dev build", version: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &clock{t: epoch}
			cfg := install(t, tt.version, c)
			t.Setenv(DisableEnv, tt.optOut)
			r := &release{tag: "v1.1.0", payload: []byte("x"), checksum: digestOf([]byte("x"))}
			r.serve(t)
			_ = Check(cfg)
			_ = Apply(cfg, "v1.1.0")
			if r.apiCalls.Load() != 0 || r.downloads.Load() != 0 {
				t.Error("a disabled child reached the network")
			}
			entries, _ := os.ReadDir(cfg.dir())
			if len(entries) != 1 {
				t.Errorf("a disabled child left files: %d entries", len(entries))
			}
		})
	}
}

func TestApplyRefusesWhatIsNotNewerOrNotATag(t *testing.T) {
	c := &clock{t: epoch}
	cfg := install(t, "v1.2.0", c)
	r := &release{payload: []byte("x"), checksum: digestOf([]byte("x"))}
	r.serve(t)

	if err := Apply(cfg, "v1.2.0"); err != nil {
		t.Errorf("Apply(same) = %v, want a quiet no-op", err)
	}
	if err := Apply(cfg, "v1.1.9"); err != nil {
		t.Errorf("Apply(older) = %v, want a quiet no-op", err)
	}
	if err := Apply(cfg, "../../etc"); !errors.Is(err, errInvalidTag) {
		t.Errorf("Apply(junk) = %v, want errInvalidTag", err)
	}
	if r.downloads.Load() != 0 || binary(t, cfg) != "original" {
		t.Error("a release that is not newer was downloaded")
	}
}

func TestApplyRecordsABadChecksum(t *testing.T) {
	c := &clock{t: epoch}
	cfg := install(t, "v1.0.0", c)
	r := &release{tag: "v1.1.0", payload: []byte("tampered"), checksum: digestOf([]byte("genuine"))}
	r.serve(t)

	if err := Check(cfg); err == nil {
		t.Fatal("Check() installed a payload whose checksum does not match")
	}
	if binary(t, cfg) != "original" {
		t.Error("binary replaced despite a checksum mismatch")
	}
	st := ReadState(cfg.dir())
	if st.ApplyError == "" {
		t.Fatalf("state = %+v, want an apply error", st)
	}
	if info := st.Pill("v1.0.0", c.now()); !info.Available || !info.Failed {
		t.Errorf("pill = %+v, want the failure variant", info)
	}

	// A good release later clears the error
	payload := []byte("genuine")
	r.payload, r.checksum = payload, digestOf(payload)
	if err := Apply(cfg, "v1.1.0"); err != nil {
		t.Fatalf("Apply() = %v", err)
	}
	if st := ReadState(cfg.dir()); st.ApplyError != "" {
		t.Errorf("apply error not cleared: %+v", st)
	}
}

func TestSweepSparesTheStateTemporary(t *testing.T) {
	dir := t.TempDir()
	old := epoch.Add(-2 * time.Hour)
	files := map[string]bool{
		"status-line-update-123":   false, // stale download: swept
		"status-line-update-456":   true,  // fresh download: kept
		stateFileName + ".tmp-789": true,  // state temporary: kept
		stateFileName:              true,
		lockFileName:               true,
		"status-line":              true,
	}
	for name := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		if name != "status-line-update-456" {
			_ = os.Chtimes(path, old, old)
		}
	}
	// A directory under the same prefix is never ours to remove
	sub := filepath.Join(dir, "status-line-update-dir")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	_ = os.Chtimes(sub, old, old)
	files["status-line-update-dir"] = true

	sweep(dir, epoch)
	for name, kept := range files {
		_, err := os.Lstat(filepath.Join(dir, name))
		if exists := err == nil; exists != kept {
			t.Errorf("%s: exists=%v, want %v", name, exists, kept)
		}
	}
}

func TestFramesSpawnAtMostOncePerWindow(t *testing.T) {
	tests := []struct {
		name   string
		optOut string
		srv    func(t *testing.T)
		want   int
	}{
		{name: "offline", srv: func(t *testing.T) {
			s := httptest.NewServer(http.NotFoundHandler())
			s.Close()
			pointAt(t, s.URL)
		}, want: 1},
		{name: "403", srv: func(t *testing.T) { (&release{status: http.StatusForbidden}).serve(t) }, want: 1},
		{name: "opt-out", optOut: "on", srv: func(t *testing.T) { (&release{tag: "v1.0.0"}).serve(t) }, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &clock{t: epoch}
			cfg := install(t, "v1.0.0", c)
			t.Setenv(DisableEnv, tt.optOut)
			tt.srv(t)
			spawns := 0
			// 100 frames at 1 Hz, each child run to completion before the next
			for range 100 {
				st := ReadState(cfg.dir())
				if Allowed(cfg.Version, cfg.dir()) && st.Due(c.now()) {
					spawns++
					_ = Check(cfg)
				}
				c.advance(time.Second)
			}
			if spawns != tt.want {
				t.Errorf("%d spawns in 100 s, want %d", spawns, tt.want)
			}
		})
	}
}

func TestPillShowsAPersistentCheckFailure(t *testing.T) {
	st := State{CheckError: "offline", CheckFailingSince: epoch}
	if info := st.Pill("v1.0.0", epoch.Add(59*time.Minute)); info.Available {
		t.Errorf("pill after 59 min of failures = %+v, want off", info)
	}
	if info := st.Pill("v1.0.0", epoch.Add(61*time.Minute)); !info.Available || !info.Failed {
		t.Errorf("pill after 61 min of failures = %+v, want the failure variant", info)
	}
	// An apply error about a release already installed is not shown
	st = State{Tag: "v1.0.0", ApplyError: "old"}
	if info := st.Pill("v1.0.0", epoch); info.Available {
		t.Errorf("pill = %+v, want off once the binary caught up", info)
	}
}

func TestAHalfWrittenStateIsStale(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, stateFileName), []byte(`{"tag":"v9.9.9","next_attempt_at":"2099-`), 0o600); err != nil {
		t.Fatal(err)
	}
	st := ReadState(dir)
	if st.Tag != "" || !st.Due(epoch) {
		t.Errorf("half-written state read as %+v", st)
	}
}
