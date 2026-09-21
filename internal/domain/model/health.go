// Package model contains domain entities.
package model

// ServiceHealth is the aggregate state of Claude's public services.
type ServiceHealth int

// Service health levels, from unknown to down.
const (
	// HealthUnknown means no recent status is available; nothing is drawn.
	HealthUnknown ServiceHealth = iota
	// HealthOK means every counted component is operational.
	HealthOK
	// HealthDegraded means exactly one component is degraded.
	HealthDegraded
	// HealthDown means two degraded components or one major outage.
	HealthDown
)

// Component states as published by the status page.
const (
	// componentDegraded is a component running slower than normal.
	componentDegraded string = "degraded_performance"
	// componentPartial is a component partly unavailable.
	componentPartial string = "partial_outage"
	// componentMajor is a component fully unavailable.
	componentMajor string = "major_outage"
	// degradedForDown is how many degraded components make an outage.
	degradedForDown int = 2
)

// ClassifyHealth folds component states into one level.
//
// One degraded component is a warning; two at once, or a single major outage,
// is an outage. Maintenance is scheduled and announced, so it counts as neither.
//
// Params:
//   - states: status of each counted component
//
// Returns:
//   - ServiceHealth: aggregate level, HealthUnknown when there is nothing to judge
func ClassifyHealth(states []string) ServiceHealth {
	// Without components there is nothing to judge
	if len(states) == 0 {
		return HealthUnknown
	}
	degraded := 0
	// Count what is broken, and how badly
	for _, state := range states {
		switch state {
		// A full outage is enough on its own
		case componentMajor:
			return HealthDown
		// A slowdown or a partial outage is a warning
		case componentDegraded, componentPartial:
			degraded++
		}
	}
	// Map the count onto the three visible levels
	switch {
	// Two warnings at once are an outage
	case degraded >= degradedForDown:
		return HealthDown
	// One warning stays a warning
	case degraded == 1:
		return HealthDegraded
	// Everything else is operational
	default:
		return HealthOK
	}
}
