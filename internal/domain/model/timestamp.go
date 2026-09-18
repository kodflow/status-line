// Package model contains domain entities and value objects.
package model

import (
	"bytes"
	"encoding/json"
	"strconv"
	"time"
)

// Timestamp decodes an instant written either as a Unix epoch in seconds
// (what Claude Code pipes on stdin) or as an RFC 3339 string (what the
// Anthropic usage API returns). Mixing the two silently yields a zero time,
// which the renderer would then read as "quota absent".
type Timestamp struct {
	time.Time
}

// UnmarshalJSON decodes both the numeric and the string representation.
//
// Params:
//   - data: raw JSON value
//
// Returns:
//   - error: nil; an unparseable value yields the zero instant
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	// A JSON null carries no instant
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		t.Time = time.Time{}
		return nil
	}

	// A quoted value is an RFC 3339 string
	if trimmed[0] == '"' {
		var raw string
		// An undecodable string leaves the zero instant in place
		if err := json.Unmarshal(trimmed, &raw); err != nil {
			t.Time = time.Time{}
			return nil
		}
		parsed, err := time.Parse(time.RFC3339, raw)
		// An unparseable layout leaves the zero instant in place
		if err != nil {
			t.Time = time.Time{}
			return nil
		}
		t.Time = parsed
		return nil
	}

	// An unquoted value is a Unix epoch in seconds
	epoch, err := strconv.ParseFloat(string(trimmed), 64)
	// A non-numeric value leaves the zero instant in place
	if err != nil {
		t.Time = time.Time{}
		return nil
	}
	// Epoch zero means "unset" upstream, not 1970
	if epoch <= 0 {
		t.Time = time.Time{}
		return nil
	}
	t.Time = time.Unix(int64(epoch), 0)
	return nil
}
