// Package usage provides the Anthropic API usage adapter.
package usage

import (
	"encoding/json"
	"time"
)

// Cache layout and freshness constants.
const (
	// cacheTTL is how long a fetched payload stays authoritative.
	cacheTTL time.Duration = 60 * time.Second
	// cacheDirPrefix is the per-user cache directory name.
	cacheDirPrefix string = "status-line-usage-"
	// cacheFileName is the cached payload file name.
	cacheFileName string = "usage.json"
	// refreshEnv marks a process spawned solely to refresh the cache.
	refreshEnv string = "STATUSLINE_USAGE_REFRESH"
)

// isFresh reports whether an age is still within the TTL.
//
// Params:
//   - age: how old the cached payload is
//
// Returns:
//   - bool: true while the payload is authoritative
func isFresh(age time.Duration) bool {
	// A payload younger than the TTL needs no refresh
	return age < cacheTTL
}

// decode parses a cached or freshly fetched payload.
//
// Params:
//   - data: raw API payload
//
// Returns:
//   - usageResponse: decoded payload
//   - bool: false when the payload is not valid JSON
func decode(data []byte) (usageResponse, bool) {
	var parsed usageResponse
	// A payload that does not decode is treated as a miss
	if err := json.Unmarshal(data, &parsed); err != nil {
		return usageResponse{}, false
	}
	return parsed, true
}
