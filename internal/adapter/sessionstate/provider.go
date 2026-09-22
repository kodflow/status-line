// Package sessionstate reads whether the host is working on the session.
package sessionstate

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

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
}

// sessionFile is the part of a registry entry the status line reads.
type sessionFile struct {
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
// The registry holds one file per live session; the first one naming this
// session decides. Files mid-write or malformed are skipped, never fatal.
//
// Returns:
//   - bool: true when the session's entry says busy
func (p *Provider) Working() bool {
	// Without an id or a registry no entry can be this session's
	if p.dir == "" || p.sessionID == "" {
		return false
	}
	paths, err := filepath.Glob(filepath.Join(p.dir, "*.json"))
	if err != nil {
		return false
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
		return entry.Status == busyStatus
	}
	return false
}
