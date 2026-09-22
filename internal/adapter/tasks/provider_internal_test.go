package tasks

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestNewProviderLocatesTheStores(t *testing.T) {
	cfg := t.TempDir()
	t.Setenv(configDirEnv, cfg)
	t.Setenv(listIDEnv, "")
	p := NewProvider("896dc174-db17-4ec1")
	if want := filepath.Join(cfg, "kodflow", "sessions", "896dc174-db17-4ec1"); p.sessionDir != want {
		t.Errorf("sessionDir = %q, want %q", p.sessionDir, want)
	}
	if want := filepath.Join(cfg, "tasks", "session-896dc174"); p.builtinDir != want {
		t.Errorf("builtinDir = %q, want %q", p.builtinDir, want)
	}
	t.Setenv(listIDEnv, "my team/list")
	if got, want := NewProvider("896dc174").builtinDir, filepath.Join(cfg, "tasks", "my-team-list"); got != want {
		t.Errorf("builtinDir with override = %q, want %q", got, want)
	}
	t.Setenv(listIDEnv, "")
	if p := NewProvider(""); p.sessionDir != "" || p.builtinDir != "" {
		t.Errorf("provider without session located %q / %q", p.sessionDir, p.builtinDir)
	}
}

func TestTasksPrefersTheMainAgentsMCPList(t *testing.T) {
	dir := t.TempDir()
	p := &Provider{sessionDir: filepath.Join(dir, "s"), builtinDir: filepath.Join(dir, "b"), now: time.Now}
	writeFile(t, filepath.Join(p.builtinDir, "1.json"), `{"id":"1","subject":"builtin","status":"pending"}`)

	if got := p.Tasks(); got.Total() != 1 || got.Items[0].Subject != "builtin" {
		t.Fatalf("without an MCP list the built-in one is read, got %+v", got)
	}

	writeFile(t, filepath.Join(p.sessionDir, "tasks.json"), `{"tasks":[
		{"id":"10","agent":"main","subject":"ten","status":"pending"},
		{"id":"2","agent":"a1b2","subject":"sub","status":"in_progress"},
		{"id":"3","agent":"main","subject":"three","status":"in_progress"}]}`)
	list := p.Tasks()
	if list.Total() != 2 || list.Items[0].ID != "3" || list.Items[1].ID != "10" {
		t.Errorf("MCP list = %+v, want the main agent's tasks 3 then 10", list.Items)
	}
	if list.Current() != "three" {
		t.Errorf("Current() = %q, a subagent's task must not lead", list.Current())
	}

	writeFile(t, filepath.Join(p.sessionDir, "tasks.json"), `{"tasks":[{"id":"1","agent":"a1b2","subject":"sub","status":"pending"}]}`)
	if got := p.Tasks(); got.Items[0].Subject != "builtin" {
		t.Errorf("an MCP list with no main-agent task falls back, got %+v", got)
	}
}

func TestBuiltinTasks(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "10.json"), `{"id":"10","subject":"ten","status":"pending"}`)
	writeFile(t, filepath.Join(dir, "2.json"), `{"id":"2","subject":"two","status":"in_progress"}`)
	writeFile(t, filepath.Join(dir, "1.json"), `{"id":"1","subject":"one","status":"completed"}`)
	writeFile(t, filepath.Join(dir, "3.json"), `{"id":`)
	writeFile(t, filepath.Join(dir, ".lock"), ``)

	list := (&Provider{builtinDir: dir}).Tasks()
	if list.Total() != 3 {
		t.Fatalf("Total() = %d, want 3 (malformed file skipped)", list.Total())
	}
	if list.Items[0].ID != "1" || list.Items[1].ID != "2" || list.Items[2].ID != "10" {
		t.Errorf("order = %v, want numeric id order", list.Items)
	}
	if got := (&Provider{}).Tasks(); got.Total() != 0 {
		t.Errorf("unlocated provider returned %d tasks", got.Total())
	}
}

func TestSubagents(t *testing.T) {
	dir := t.TempDir()
	now := time.Unix(1_800_000_000, 0)
	p := &Provider{sessionDir: dir, now: func() time.Time { return now }}
	if got := p.Subagents(); got != 0 {
		t.Errorf("no registry: Subagents() = %d, want 0", got)
	}
	recent, old := now.Add(-time.Minute).Unix(), now.Add(-13*time.Hour).Unix()
	writeFile(t, filepath.Join(dir, "agents.json"), `{"agents":{
		"a":{"type":"Explore","started":`+itoa(recent)+`,"stopped":null},
		"b":{"type":"Plan","started":`+itoa(recent)+`,"stopped":`+itoa(recent)+`},
		"c":{"type":"Explore","started":`+itoa(recent)+`,"stopped":null},
		"d":{"type":"Explore","started":`+itoa(old)+`,"stopped":null}}}`)
	if got := p.Subagents(); got != 2 {
		t.Errorf("Subagents() = %d, want 2 (stopped and stale excluded)", got)
	}
	writeFile(t, filepath.Join(dir, "agents.json"), `{"agents":`)
	if got := p.Subagents(); got != 0 {
		t.Errorf("malformed registry: Subagents() = %d, want 0", got)
	}
	if got := (&Provider{}).Subagents(); got != 0 {
		t.Errorf("unlocated provider: Subagents() = %d, want 0", got)
	}
}

func itoa(v int64) string {
	return strconv.FormatInt(v, 10)
}

func TestTasksShowTheCurrentEpicOnly(t *testing.T) {
	dir := t.TempDir()
	p := &Provider{sessionDir: dir, now: time.Now}
	writeFile(t, filepath.Join(dir, "tasks.json"), `{"epics":{"main":{"id":2,"title":"SDK"}},"tasks":[
		{"id":"1","agent":"main","epic":1,"subject":"old subject","status":"completed"},
		{"id":"2","agent":"main","epic":2,"subject":"freeze goldens","status":"in_progress"},
		{"id":"3","agent":"main","epic":2,"subject":"write the SDK","status":"pending"}]}`)
	list := p.Tasks()
	if list.Total() != 2 || list.Items[0].Subject != "freeze goldens" {
		t.Errorf("current epic = %+v, want tasks 2 and 3 only", list.Items)
	}
	writeFile(t, filepath.Join(dir, "tasks.json"), `{"tasks":[{"id":"1","agent":"main","subject":"before epics","status":"pending"}]}`)
	if got := p.Tasks(); got.Total() != 1 {
		t.Errorf("a list with no epic yet is shown whole, got %+v", got.Items)
	}
}
