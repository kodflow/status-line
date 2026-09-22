// Package sessionstate reads whether the host is working on the session.
package sessionstate

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/florent/status-line/internal/domain/port"
)

// Session registry constants.
const (
	// configDirEnv relocates the whole configuration directory.
	configDirEnv string = "CLAUDE_CONFIG_DIR"
	// busyStatus is the status the host writes while a turn is running.
	busyStatus string = "busy"
)

// Compile-time interface implementation check.
var _ port.ActivityProvider = (*Provider)(nil)

// Provider reads the host's session registry, <config>/sessions/<pid>.json:
// one small file per running session, naming its session id and status.
type Provider struct {
	dir       string
	sessionID string
	once      sync.Once
	entry     sessionFile
	found     bool
}

// sessionFile is the part of a registry entry the status line reads.
type sessionFile struct {
	PID       int    `json:"pid"`
	SessionID string `json:"sessionId"`
	Status    string `json:"status"`
}

// NewProvider locates the session registry.
//
// Params:
//   - sessionID: session identifier from stdin, may be empty
//
// Returns:
//   - *Provider: provider for that session, never working without an id
func NewProvider(sessionID string) *Provider {
	base := os.Getenv(configDirEnv)
	// Default to the per-user configuration directory
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return &Provider{}
		}
		base = filepath.Join(home, ".claude")
	}
	return &Provider{dir: filepath.Join(base, "sessions"), sessionID: sessionID}
}

// Working reports whether the session is busy on a turn.
//
// Returns:
//   - bool: true when the session's entry says busy
func (p *Provider) Working() bool {
	entry, ok := p.lookup()
	// Without an entry the session cannot be seen working
	if !ok {
		return false
	}
	return entry.Status == busyStatus
}

// PID returns the process id of the host running this session.
//
// Returns:
//   - int: pid named by the session's entry, 0 when there is none
func (p *Provider) PID() int {
	entry, ok := p.lookup()
	// No entry, or one without a usable pid, names no process
	if !ok || entry.PID <= 0 {
		return 0
	}
	return entry.PID
}

// lookup finds this session's registry entry, scanning the registry once.
//
// The registry holds one file per live session; the first one naming this
// session decides. Files mid-write or malformed are skipped, never fatal.
// Several adapters ask concurrently, hence the once.
//
// Returns:
//   - sessionFile: the session's entry
//   - bool: whether an entry was found
func (p *Provider) lookup() (sessionFile, bool) {
	p.once.Do(func() {
		p.entry, p.found = p.scan()
	})
	return p.entry, p.found
}

// scan reads the registry for this session's entry.
//
// Returns:
//   - sessionFile: the session's entry
//   - bool: whether an entry was found
func (p *Provider) scan() (sessionFile, bool) {
	// Without an id or a registry no entry can be this session's
	if p.dir == "" || p.sessionID == "" {
		return sessionFile{}, false
	}
	paths, err := filepath.Glob(filepath.Join(p.dir, "*.json"))
	// A broken pattern cannot happen with a fixed suffix; stay quiet anyway
	if err != nil {
		return sessionFile{}, false
	}
	needle := []byte(p.sessionID)
	// Stop at the first entry that is this session's
	for _, path := range paths {
		data, err := os.ReadFile(path)
		// Skip the other sessions without decoding them
		if err != nil || !bytes.Contains(data, needle) {
			continue
		}
		var entry sessionFile
		// A malformed entry, or one that only mentions the id, is not ours
		if json.Unmarshal(data, &entry) != nil || entry.SessionID != p.sessionID {
			continue
		}
		return entry, true
	}
	return sessionFile{}, false
}
