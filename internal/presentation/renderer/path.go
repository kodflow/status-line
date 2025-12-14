// Package renderer provides status line rendering.
package renderer

import (
	"os"
	"strings"
)

const (
	pathEllipsis   = "..."
	pathSeparator  = "/"
	defaultMaxPath = 30
)

// TruncatePath shortens a path by removing leading segments.
func TruncatePath(path string, maxLen int) string {
	if maxLen <= 0 {
		maxLen = defaultMaxPath
	}

	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(path, home) {
		path = "~" + strings.TrimPrefix(path, home)
	}

	if len(path) <= maxLen {
		return path
	}

	segments := strings.Split(path, pathSeparator)
	if len(segments) > 0 && segments[0] == "" {
		segments = segments[1:]
	}

	for len(segments) > 1 {
		segments = segments[1:]
		candidate := pathEllipsis + pathSeparator + strings.Join(segments, pathSeparator)
		if len(candidate) <= maxLen {
			return candidate
		}
	}

	if len(segments) == 1 {
		return pathEllipsis + pathSeparator + segments[0]
	}
	return path
}
