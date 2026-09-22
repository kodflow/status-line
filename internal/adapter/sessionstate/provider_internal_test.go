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
	p := &Provider{dir: dir, sessionID: id}
	if p.Working() {
		t.Error("an empty registry reads as working")
	}

	writeEntry(t, dir, "1.json", `{"pid":1,"sessionId":"other","status":"busy"}`)
	writeEntry(t, dir, "2.json", `{"pid":2,"sessionId":"`)
	writeEntry(t, dir, "3.json", `{"pid":3,"sessionId":"other","status":"busy","note":"`+id+`"}`)
	writeEntry(t, dir, "4.key", id)
	if p.Working() {
		t.Error("no entry is this session's, yet it reads as working")
	}

	writeEntry(t, dir, "5.json", `{"pid":5,"sessionId":"`+id+`","status":"idle"}`)
	if p.Working() {
		t.Error("an idle session reads as working")
	}
	writeEntry(t, dir, "5.json", `{"pid":5,"sessionId":"`+id+`","status":"busy"}`)
	if !p.Working() {
		t.Error("a busy session reads as idle")
	}
	writeEntry(t, dir, "5.json", `{"pid":5,"sessionId":"`+id+`","status":"waiting"}`)
	if p.Working() {
		t.Error("a session waiting on the user reads as working")
	}

	if (&Provider{dir: dir}).Working() || (&Provider{sessionID: id}).Working() {
		t.Error("a provider without an id or a registry reads as working")
	}
}
