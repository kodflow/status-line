// Package diskcache stores small payloads in a private per-user directory so
// that rendering never blocks on the network.
package diskcache

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Cache permission constants.
const (
	// dirPerm keeps the cache directory private to its owner.
	dirPerm os.FileMode = 0700
	// filePerm keeps a cached payload private to its owner.
	filePerm os.FileMode = 0600
)

// Cache is one payload file on disk. The status line is redrawn constantly;
// a synchronous fetch would add its full latency to every single redraw.
type Cache struct {
	path string
}

// New resolves a per-user cache location.
//
// Params:
//   - dirPrefix: directory name prefix, completed with the user id
//   - fileName: payload file name inside that directory
//
// Returns:
//   - *Cache: cache rooted at a private per-user directory
func New(dirPrefix, fileName string) *Cache {
	base := os.Getenv("XDG_RUNTIME_DIR")
	// Fall back to the platform temp directory outside XDG systems
	if base == "" {
		base = os.TempDir()
	}
	dir := filepath.Join(base, dirPrefix+strconv.Itoa(os.Getuid()))
	return &Cache{path: filepath.Join(dir, fileName)}
}

// Path returns the payload file location.
//
// Returns:
//   - string: absolute payload path
func (c *Cache) Path() string {
	return c.path
}

// Read returns the cached payload and how old it is.
//
// Returns:
//   - []byte: cached payload
//   - time.Duration: age of the payload
//   - bool: false when no readable cache exists
func (c *Cache) Read() ([]byte, time.Duration, bool) {
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

// Write stores a freshly fetched payload.
//
// Params:
//   - data: raw payload to persist
//
// Returns:
//   - error: any error while writing
func (c *Cache) Write(data []byte) error {
	dir := filepath.Dir(c.path)
	name := filepath.Base(c.path)
	// Create the private directory on first use
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return err
	}
	// Write through a temporary file so a concurrent reader never sees a
	// half-written payload
	tmp, err := os.CreateTemp(dir, name+".*")
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
	if err := os.Chmod(tmpName, filePerm); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, c.path)
}
