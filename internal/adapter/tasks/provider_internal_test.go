package tasks

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTask(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestNewProviderLocatesTheList(t *testing.T) {
	cfg := t.TempDir()
	t.Setenv(configDirEnv, cfg)
	t.Setenv(listIDEnv, "")
	if got, want := NewProvider("896dc174-db17-4ec1").dir, filepath.Join(cfg, "tasks", "session-896dc174"); got != want {
		t.Errorf("dir = %q, want %q", got, want)
	}
	t.Setenv(listIDEnv, "my team/list")
	if got, want := NewProvider("896dc174").dir, filepath.Join(cfg, "tasks", "my-team-list"); got != want {
		t.Errorf("dir with override = %q, want %q", got, want)
	}
	t.Setenv(listIDEnv, "")
	if got := NewProvider("").dir; got != "" {
		t.Errorf("dir without session = %q, want empty", got)
	}
}

func TestTasks(t *testing.T) {
	dir := t.TempDir()
	writeTask(t, dir, "10.json", `{"id":"10","subject":"ten","status":"pending"}`)
	writeTask(t, dir, "2.json", `{"id":"2","subject":"two","status":"in_progress"}`)
	writeTask(t, dir, "1.json", `{"id":"1","subject":"one","status":"completed"}`)
	writeTask(t, dir, "3.json", `{"id":`)
	writeTask(t, dir, ".lock", ``)

	list := (&Provider{dir: dir}).Tasks()
	if list.Total() != 3 {
		t.Fatalf("Total() = %d, want 3 (malformed file skipped)", list.Total())
	}
	if list.Items[0].ID != "1" || list.Items[1].ID != "2" || list.Items[2].ID != "10" {
		t.Errorf("order = %v, want numeric id order", list.Items)
	}
	if list.Current() != "two" {
		t.Errorf("Current() = %q, want %q", list.Current(), "two")
	}
	if got := (&Provider{}).Tasks(); got.Total() != 0 {
		t.Errorf("unlocated provider returned %d tasks", got.Total())
	}
	if got := (&Provider{dir: filepath.Join(dir, "missing")}).Tasks(); got.Total() != 0 {
		t.Errorf("missing directory returned %d tasks", got.Total())
	}
}
