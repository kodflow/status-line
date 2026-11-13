// Package updater provides self-update functionality for the status-line binary.
package updater

import (
	"fmt"
	"time"
)

// statusError is a release lookup the API answered with something other than
// 200. retryAt is the date a 403 or 429 asked us to wait until, zero when the
// response carried no usable hint.
type statusError struct {
	code    int
	retryAt time.Time
}

// Error describes the refusal.
//
// Returns:
//   - string: the status code the API answered with
func (e *statusError) Error() string {
	// Name the status: it is what the state file records
	return fmt.Sprintf("unexpected status: %d", e.code)
}
