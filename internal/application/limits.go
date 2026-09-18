// Package application contains application services.
package application

import "github.com/florent/status-line/internal/domain/model"

// resolveLimits merges the quotas Claude Code pipes on stdin with the ones the
// usage API exposes.
//
// stdin wins wherever both carry the same bucket: it is free, synchronous and
// always current, while the API is cached and therefore slightly behind. The
// API only contributes what stdin cannot know — model-scoped weekly quotas and
// the extra-usage credit balance — plus the buckets a given Claude Code build
// does not send.
//
// A bucket missing from both sources stays missing. Plans genuinely differ:
// one exposes a weekly cap, another does not, and rendering an absent quota as
// a zero-percent bar would be a lie about the account.
//
// Params:
//   - stdin: quotas parsed from the Claude Code payload
//   - api: quotas fetched from the usage API, possibly empty
//
// Returns:
//   - model.LimitSet: merged quotas
func resolveLimits(stdin, api model.LimitSet) model.LimitSet {
	merged := stdin

	// Take the session bucket from the API only when stdin omitted it
	if !merged.Session.IsValid() && api.Session.IsValid() {
		merged.Session = api.Session
	}
	// Take the weekly bucket from the API only when stdin omitted it
	if !merged.Weekly.IsValid() && api.Weekly.IsValid() {
		merged.Weekly = api.Weekly
	}

	// Model-scoped quotas exist only in the API
	merged.Scoped = api.Scoped
	// Credit balance exists only in the API
	merged.Extra = api.Extra

	return merged
}
