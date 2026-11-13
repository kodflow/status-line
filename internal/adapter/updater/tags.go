// Package updater provides self-update functionality for the status-line binary.
package updater

import (
	"net/http"
	"regexp"
	"strconv"
	"time"
)

const (
	// maxRetryWait caps a rate-limit hint, so a malformed header can never
	// switch the check off for days.
	maxRetryWait time.Duration = 24 * time.Hour
)

// tagPattern is the only shape of release tag this binary ever installs.
var tagPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

// validTag says whether a tag names a plain release.
//
// Params:
//   - tag: release tag as published
//
// Returns:
//   - bool: true for vMAJOR.MINOR.PATCH and nothing else
func validTag(tag string) bool {
	// Anything else is refused, pre-releases included
	return tagPattern.MatchString(tag)
}

// newer says whether a release tag is more recent than a compiled version.
//
// Params:
//   - tag: release tag
//   - current: compiled version, empty for a dev build
//
// Returns:
//   - bool: true when tag is valid, current is set and tag is ahead of it
func newer(tag, current string) bool {
	// A dev build never updates, and an invalid tag is never newer
	if current == "" || !validTag(tag) {
		return false
	}
	return (&Updater{version: current}).isNewer(tag)
}

// retryAt reads how long a 403 or 429 asked us to wait.
//
// Retry-After is honoured first, in seconds or as an HTTP date; failing that,
// X-RateLimit-Reset gives the epoch second the quota comes back.
//
// Params:
//   - h: response headers
//   - now: current time
//
// Returns:
//   - time.Time: the date to wait until, zero without a usable hint
func retryAt(h http.Header, now time.Time) time.Time {
	at := retryAfter(h.Get("Retry-After"), now)
	// No Retry-After: the reset of the rate-limit window
	if at.IsZero() {
		at = rateLimitReset(h.Get("X-RateLimit-Reset"))
	}
	// No hint, or one already past: the caller's own backoff applies
	if at.IsZero() || !at.After(now) {
		return time.Time{}
	}
	// A hint further out than a day is distrusted
	if limit := now.Add(maxRetryWait); at.After(limit) {
		return limit
	}
	return at
}

// retryAfter parses a Retry-After value.
//
// Params:
//   - v: header value, seconds or an HTTP date
//   - now: current time
//
// Returns:
//   - time.Time: the date it names, zero when absent or malformed
func retryAfter(v string, now time.Time) time.Time {
	// A delay in seconds
	if secs, err := strconv.ParseInt(v, 10, 64); err == nil && secs >= 0 {
		return now.Add(time.Duration(secs) * time.Second)
	}
	// An HTTP date
	if date, err := http.ParseTime(v); err == nil {
		return date
	}
	return time.Time{}
}

// rateLimitReset parses an X-RateLimit-Reset value.
//
// Params:
//   - v: header value, an epoch second
//
// Returns:
//   - time.Time: the date it names, zero when absent or malformed
func rateLimitReset(v string) time.Time {
	epoch, err := strconv.ParseInt(v, 10, 64)
	// Absent, malformed or nonsensical: no hint
	if err != nil || epoch <= 0 {
		return time.Time{}
	}
	return time.Unix(epoch, 0)
}
