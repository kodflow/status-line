package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// serveRelease stands in for the GitHub release endpoint, serving one asset and
// whatever checksum the test wants paired with it.
func serveRelease(t *testing.T, payload []byte, checksum string, checksumStatus int) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, checksumSuffix) {
			w.WriteHeader(checksumStatus)
			// An error status carries no body worth writing
			if checksumStatus == http.StatusOK {
				fmt.Fprintf(w, "%s  asset\n", checksum)
			}
			return
		}
		w.Write(payload)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func digestOf(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func TestFetchChecksum(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		status  int
		want    string
		wantErr bool
	}{
		{name: "sha256sum format", body: digestOf([]byte("x")) + "  status-line-linux-amd64", status: http.StatusOK, want: digestOf([]byte("x"))},
		{name: "bare digest", body: digestOf([]byte("y")), status: http.StatusOK, want: digestOf([]byte("y"))},
		{name: "surrounding whitespace", body: "  " + digestOf([]byte("z")) + "  asset\n", status: http.StatusOK, want: digestOf([]byte("z"))},
		{name: "missing checksum file", body: "", status: http.StatusNotFound, wantErr: true},
		{name: "truncated digest", body: "abc123  asset", status: http.StatusOK, wantErr: true},
		{name: "empty body", body: "", status: http.StatusOK, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.body)
			}))
			defer srv.Close()

			u := NewUpdater("v1.0.0")
			saved := downloadURLFor
			downloadURLFor = func(_, _, _, _ string) string { return srv.URL }
			defer func() { downloadURLFor = saved }()

			got, err := u.fetchChecksum("v1.0.0", "asset")
			if tt.wantErr {
				// An unverifiable release must not yield a digest to compare against
				if err == nil {
					t.Fatalf("fetchChecksum() = %q, want an error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("fetchChecksum() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("fetchChecksum() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDownloadAndReplace_RejectsMismatch(t *testing.T) {
	payload := []byte("tampered binary")
	srv := serveRelease(t, payload, digestOf([]byte("the real binary")), http.StatusOK)

	saved := downloadURLFor
	downloadURLFor = func(_, _, _, asset string) string { return srv.URL + "/" + asset }
	defer func() { downloadURLFor = saved }()

	// Stand in for the running executable so a successful swap would be visible
	dir := t.TempDir()
	exe := filepath.Join(dir, "status-line")
	if err := os.WriteFile(exe, []byte("original"), executablePerm); err != nil {
		t.Fatal(err)
	}

	u := NewUpdater("v1.0.0")
	u.execPath = exe

	err := u.downloadAndReplace("v1.0.1")
	// A payload that does not match its published digest must never be installed
	if err == nil {
		t.Fatal("downloadAndReplace() accepted a payload whose checksum does not match")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Errorf("error = %v, want a checksum mismatch", err)
	}

	// The running binary must be left exactly as it was
	after, readErr := os.ReadFile(exe)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(after) != "original" {
		t.Errorf("executable was replaced with %q", after)
	}

	// And no temporary file may be left behind next to it
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "status-line-update-") {
			t.Errorf("temporary file %q left behind", e.Name())
		}
	}
}

func TestDownloadAndReplace_AcceptsMatch(t *testing.T) {
	payload := []byte("the real binary")
	srv := serveRelease(t, payload, digestOf(payload), http.StatusOK)

	saved := downloadURLFor
	downloadURLFor = func(_, _, _, asset string) string { return srv.URL + "/" + asset }
	defer func() { downloadURLFor = saved }()

	dir := t.TempDir()
	exe := filepath.Join(dir, "status-line")
	if err := os.WriteFile(exe, []byte("original"), executablePerm); err != nil {
		t.Fatal(err)
	}

	u := NewUpdater("v1.0.0")
	u.execPath = exe

	// A payload matching its digest is the one case that should go through
	if err := u.downloadAndReplace("v1.0.1"); err != nil {
		t.Fatalf("downloadAndReplace() error = %v", err)
	}
	after, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(payload) {
		t.Errorf("executable = %q, want the downloaded payload", after)
	}
}

func TestCheckForUpdate_HonoursOptOut(t *testing.T) {
	t.Setenv(disableEnv, "1")

	// A managed image verifies this binary at build time; replacing it at
	// runtime would void that guarantee, so the switch must win outright
	if info := NewUpdater("v1.0.0").CheckForUpdate(); info.Available {
		t.Error("CheckForUpdate() reported an update while self-update is disabled")
	}
}
