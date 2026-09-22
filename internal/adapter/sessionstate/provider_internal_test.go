package sessionstate

import (
	"os"
	"path/filepath"
	"testing"
)

func writeEntry(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestNewProviderLocatesTheRegistry(t *testing.T) {
	cfg := t.TempDir()
	t.Setenv(configDirEnv, cfg)
	if got, want := NewProvider("abc").dir, filepath.Join(cfg, "sessions"); got != want {
		t.Errorf("dir = %q, want %q", got, want)
	}
}

func TestWorking(t *testing.T) {
	dir := t.TempDir()
	const id = "896dc174-db17-4ec1-960b-5f879c31cc32"
	p := func() *Provider { return &Provider{dir: dir, sessionID: id} }
	if p().Working() {
		t.Error("an empty registry reads as working")
	}

	writeEntry(t, dir, "1.json", `{"pid":1,"sessionId":"other","status":"busy"}`)
	writeEntry(t, dir, "2.json", `{"pid":2,"sessionId":"`)
	writeEntry(t, dir, "3.json", `{"pid":3,"sessionId":"other","status":"busy","note":"`+id+`"}`)
	writeEntry(t, dir, "4.key", id)
	if p().Working() {
		t.Error("no entry is this session's, yet it reads as working")
	}

	writeEntry(t, dir, "5.json", `{"pid":5,"sessionId":"`+id+`","status":"idle"}`)
	if p().Working() {
		t.Error("an idle session reads as working")
	}
	writeEntry(t, dir, "5.json", `{"pid":5,"sessionId":"`+id+`","status":"busy"}`)
	if !p().Working() {
		t.Error("a busy session reads as idle")
	}
	writeEntry(t, dir, "5.json", `{"pid":5,"sessionId":"`+id+`","status":"waiting"}`)
	if p().Working() {
		t.Error("a session waiting on the user reads as working")
	}

	if (&Provider{dir: dir}).Working() || (&Provider{sessionID: id}).Working() {
		t.Error("a provider without an id or a registry reads as working")
	}
}

func TestPIDAndSingleScan(t *testing.T) {
	dir := t.TempDir()
	const id = "896dc174-db17-4ec1-960b-5f879c31cc32"
	if got := (&Provider{dir: dir, sessionID: id}).PID(); got != 0 {
		t.Errorf("PID() without an entry = %d, want 0", got)
	}
	writeEntry(t, dir, "1.json", `{"pid":1,"sessionId":"other","status":"busy"}`)
	writeEntry(t, dir, "5752.json", `{"pid":5752,"sessionId":"`+id+`","status":"busy"}`)
	p := &Provider{dir: dir, sessionID: id}
	if got := p.PID(); got != 5752 {
		t.Errorf("PID() = %d, want 5752", got)
	}
	// The registry is read once: a later change is not seen by this run
	writeEntry(t, dir, "5752.json", `{"pid":5752,"sessionId":"`+id+`","status":"idle"}`)
	if !p.Working() {
		t.Error("Working() rescanned the registry instead of reusing the entry")
	}
	writeEntry(t, dir, "9.json", `{"pid":-3,"sessionId":"neg","status":"busy"}`)
	if got := (&Provider{dir: dir, sessionID: "neg"}).PID(); got != 0 {
		t.Errorf("PID() of a negative pid = %d, want 0", got)
	}
}
