// Package model contains domain entities and value objects.
package model

import "strings"

// ExtraUsage is the monthly extra-usage credit balance.
// Amounts are minor units (cents) as reported by the API.
type ExtraUsage struct {
	Enabled    bool
	Percent    int
	UsedMinor  int
	LimitMinor int
	Currency   string
	Exponent   int
	Source     LimitSource
}

// IsValid reports whether the credit balance should be rendered.
//
// Returns:
//   - bool: true when credits are enabled and were actually fetched
func (e ExtraUsage) IsValid() bool {
	// Credits that are switched off carry no useful signal
	return e.Enabled && e.Source != SourceNone
}

// LimitSet is every quota that applies to the current session.
// Plans differ: any field but Context may legitimately be absent, and the
// renderer must hide what is absent rather than display a fabricated zero.
type LimitSet struct {
	Context Limit
	Session Limit
	Weekly  Limit
	Scoped  []Limit
	Extra   ExtraUsage
}

// Timed returns the time-boxed quotas that are present, most local first.
// The context window is excluded: it has no reset schedule.
//
// Returns:
//   - []Limit: valid session, weekly and model-scoped limits
func (s LimitSet) Timed() []Limit {
	limits := make([]Limit, 0, len(s.Scoped)+2)
	// Session first: the window that bites soonest
	if s.Session.IsValid() {
		limits = append(limits, s.Session)
	}
	// Weekly next: the plan-wide seven day quota
	if s.Weekly.IsValid() {
		limits = append(limits, s.Weekly)
	}
	// Model-scoped quotas last, in the order the API returned them
	for _, scoped := range s.Scoped {
		// Skip scoped entries that were never populated
		if !scoped.IsValid() {
			continue
		}
		limits = append(limits, scoped)
	}
	return limits
}

// HasTimed reports whether any time-boxed quota is available.
//
// Returns:
//   - bool: true when at least one session or weekly quota is present
func (s LimitSet) HasTimed() bool {
	// A single valid time-boxed quota is enough to render the dashboard
	return len(s.Timed()) > 0
}

// ScopedFor returns the model-scoped quotas that apply to the model in use.
//
// A quota restricted to one model family does not constrain a session running
// on another: showing the Fable quota while working in Opus is noise that
// crowds out the quotas that can actually stop the session. The API says as
// much by marking such an entry inactive.
//
// Params:
//   - modelName: full name of the model currently in use
//
// Returns:
//   - []Limit: scoped quotas matching that model
func (s LimitSet) ScopedFor(modelName string) []Limit {
	lowered := strings.ToLower(modelName)
	matching := make([]Limit, 0, len(s.Scoped))

	// Keep only the quotas whose scope names the running model
	for _, scoped := range s.Scoped {
		// Skip entries that were never populated
		if !scoped.IsValid() {
			continue
		}
		// Match on the label the API scoped the quota to
		if !strings.Contains(lowered, strings.ToLower(scoped.Label)) {
			continue
		}
		matching = append(matching, scoped)
	}

	return matching
}
