// Package usage provides the Anthropic API usage adapter.
package usage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Cache layout and freshness constants.
const (
	// cacheTTL is how long a fetched payload stays authoritative.
	cacheTTL time.Duration = 60 * time.Second
	// cacheDirPerm keeps the cache directory private to its owner.
	cacheDirPerm os.FileMode = 0700
	// cacheFilePerm keeps the cached payload private to its owner.
	cacheFilePerm os.FileMode = 0600
	// cacheDirPrefix is the per-user cache directory name.
	cacheDirPrefix string = "status-line-usage-"
	// cacheFileName is the cached payload file name.
	cacheFileName string = "usage.json"
	// refreshEnv marks a process spawned solely to refresh the cache.
	refreshEnv string = "STATUSLINE_USAGE_REFRESH"
)

// cache stores the last API payload on disk so that rendering never blocks
// on the network. The status line is redrawn constantly; a synchronous fetch
// would add its full latency to every single redraw.
type cache struct {
	path string
}

// newCache resolves the per-user cache location.
//
// Returns:
//   - *cache: cache rooted at a private per-user directory
func newCache() *cache {
	base := os.Getenv("XDG_RUNTIME_DIR")
	// Fall back to the platform temp directory outside XDG systems
	if base == "" {
		base = os.TempDir()
	}
	dir := filepath.Join(base, cacheDirPrefix+strconv.Itoa(os.Getuid()))
	return &cache{path: filepath.Join(dir, cacheFileName)}
}

// read returns the cached payload and how old it is.
//
// Returns:
//   - []byte: cached payload
//   - time.Duration: age of the payload
//   - bool: false when no readable cache exists
func (c *cache) read() ([]byte, time.Duration, bool) {
	info, err := os.Stat(c.path)
	// A missing or unreadable cache is simply a miss
	if err != nil {
		return nil, 0, false
	}
	data, err := os.ReadFile(c.path)
	// A truncated or unreadable file is a miss rather than an error
	if err != nil || len(data) == 0 {
		return nil, 0, false
	}
	return data, time.Since(info.ModTime()), true
}

// write stores a freshly fetched payload.
//
// Params:
//   - data: raw API payload to persist
//
// Returns:
//   - error: any error while writing
func (c *cache) write(data []byte) error {
	dir := filepath.Dir(c.path)
	// Create the private directory on first use
	if err := os.MkdirAll(dir, cacheDirPerm); err != nil {
		return err
	}
	// Write through a temporary file so a concurrent reader never sees a
	// half-written payload
	tmp, err := os.CreateTemp(dir, cacheFileName+".*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// Write the payload, closing and cleaning up on failure
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	// Close before renaming so the content is flushed
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	// Restrict the payload to its owner before publishing it
	if err := os.Chmod(tmpName, cacheFilePerm); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, c.path)
}

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
