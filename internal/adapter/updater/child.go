// Package updater provides self-update functionality for the status-line binary.
package updater

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// checkInterval is the wait after a successful check.
	checkInterval time.Duration = time.Hour
	// firstBackoff is the wait after a first failed check, doubled on each
	// failure in a row up to maxBackoff.
	firstBackoff time.Duration = 5 * time.Minute
	// maxBackoff caps the wait after failed checks.
	maxBackoff time.Duration = time.Hour
	// staleDownload is the age past which a partial download is swept.
	staleDownload time.Duration = time.Hour
	// lockFileName serialises every check and apply beside the binary.
	lockFileName string = ".status-line-update.lock"
)

// errInvalidTag is an apply asked for something that is not a release tag.
var errInvalidTag = errors.New("invalid release tag")

// enabled says whether this build may update itself at all.
//
// Params:
//   - version: compiled version, empty for a dev build
//
// Returns:
//   - bool: false under the opt-out or for a dev build
func enabled(version string) bool {
	// Any value switches it off, not only "1"
	return os.Getenv(DisableEnv) == "" && version != ""
}

// failureBackoff is the wait after a given number of failed checks in a row.
//
// Params:
//   - failures: failures before this one
//
// Returns:
//   - time.Duration: 5 min doubled per failure, capped at 1 h
func failureBackoff(failures int) time.Duration {
	wait := firstBackoff
	// Double until the cap; a loop never overflows the shift
	for i := 0; i < failures && wait < maxBackoff; i++ {
		wait *= 2
	}
	return min(wait, maxBackoff)
}

// Check runs the detached --check-update child.
//
// It records the next attempt before touching the network, so a child killed
// mid-request still holds every frame off for the failure backoff. When the
// latest release is newer it applies it in the same process, under the lock
// it already holds, rather than spawning another child.
//
// Params:
//   - cfg: binary, version and clock to work with
//
// Returns:
//   - error: why nothing was checked or applied, nil when up to date
func Check(cfg Config) error {
	// The opt-out and dev builds stop here, before any lock or file
	if !enabled(cfg.Version) {
		return nil
	}
	// Another child already at work makes this one leave at once
	return withLock(cfg.dir(), func() error { return checkLocked(cfg) })
}

// checkLocked runs the check; the caller holds the lock.
//
// Params:
//   - cfg: binary, version and clock to work with
//
// Returns:
//   - error: why nothing was checked or applied, nil when up to date
func checkLocked(cfg Config) error {
	dir := cfg.dir()
	st := ReadState(dir)
	// Assume the worst before the request: a child killed during it must
	// still push the next attempt out
	st.NextAttemptAt = cfg.now().Add(failureBackoff(st.Failures))
	if err := writeState(dir, &st); err != nil {
		// A state that cannot be recorded would let every frame retry
		return err
	}

	tag, err := newUpdater(cfg.Version, cfg.ExecPath, checkTimeout).getLatestVersion(cfg.now())
	now := cfg.now()
	// A failed check backs off and is remembered for the pill
	if err != nil {
		recordCheckFailure(&st, err, now)
		return errors.Join(err, writeState(dir, &st))
	}
	st.CheckedAt, st.Tag, st.NextAttemptAt = now, tag, now.Add(checkInterval)
	st.CheckError, st.CheckFailingSince, st.Failures = "", time.Time{}, 0
	// An apply error is about a release we no longer need
	if !newer(tag, cfg.Version) {
		st.ApplyError = ""
	}
	if err := writeState(dir, &st); err != nil {
		return err
	}
	// Up to date: nothing to apply
	if !newer(tag, cfg.Version) {
		return nil
	}
	return applyLocked(cfg, tag)
}

// recordCheckFailure folds one failed check into the state.
//
// Params:
//   - st: state to update
//   - err: why the check failed
//   - now: current time
func recordCheckFailure(st *State, err error, now time.Time) {
	st.CheckError = err.Error()
	// Remember when the streak started, for the failure pill
	if st.CheckFailingSince.IsZero() {
		st.CheckFailingSince = now
	}
	st.NextAttemptAt = now.Add(failureBackoff(st.Failures))
	st.Failures++
	var refused *statusError
	// A rate limit says how long to wait: honour it over our own backoff
	if errors.As(err, &refused) && !refused.retryAt.IsZero() &&
		(refused.code == http.StatusForbidden || refused.code == http.StatusTooManyRequests) {
		st.NextAttemptAt = refused.retryAt
	}
}

// Apply runs the detached --apply-update=<tag> child: a standalone entry
// point that takes the lock itself.
//
// Params:
//   - cfg: binary, version and clock to work with
//   - tag: release to install
//
// Returns:
//   - error: why nothing was applied
func Apply(cfg Config, tag string) error {
	// The opt-out and dev builds stop here, before any lock or file
	if !enabled(cfg.Version) {
		return nil
	}
	// Another child already at work makes this one leave at once
	return withLock(cfg.dir(), func() error { return applyLocked(cfg, tag) })
}

// applyLocked installs a release; the caller holds the lock.
//
// Params:
//   - cfg: binary, version and clock to work with
//   - tag: release to install
//
// Returns:
//   - error: why nothing was applied
func applyLocked(cfg Config, tag string) error {
	// Revalidate: the tag may come from a command line
	if !validTag(tag) {
		return fmt.Errorf("%w: %q", errInvalidTag, tag)
	}
	// Never downgrade, never reinstall
	if !newer(tag, cfg.Version) {
		return nil
	}
	dir := cfg.dir()
	sweep(dir, cfg.now())
	err := newUpdater(cfg.Version, cfg.ExecPath, applyTimeout).downloadAndReplace(tag)
	st := ReadState(dir)
	st.ApplyError = ""
	// Remember the failure for the pill, clear it on success
	if err != nil {
		st.ApplyError = err.Error()
	}
	return errors.Join(err, writeState(dir, &st))
}

// sweep removes the partial downloads an interrupted apply left behind.
//
// Only regular files named status-line-update-* older than an hour go: a
// younger one may belong to an apply still running elsewhere, and the state
// file's own temporary starts with a dot, so it never matches.
//
// Params:
//   - dir: directory holding the binary
//   - now: current time
func sweep(dir string, now time.Time) {
	entries, err := os.ReadDir(dir)
	// An unreadable directory has nothing we can sweep
	if err != nil {
		return
	}
	for _, e := range entries {
		// Only our own partial downloads, never a link or a directory
		if !strings.HasPrefix(e.Name(), tempPrefix) || !e.Type().IsRegular() {
			continue
		}
		info, err := e.Info()
		// A file gone meanwhile, or still fresh, is left alone
		if err != nil || now.Sub(info.ModTime()) < staleDownload {
			continue
		}
		// A file that cannot be removed is retried at the next apply
		os.Remove(filepath.Join(dir, e.Name()))
	}
}
