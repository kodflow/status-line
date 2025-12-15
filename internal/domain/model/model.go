// Package model contains domain entities and value objects.
package model

// ModelInfo contains AI model information.
// It holds the model name and version separately.
type ModelInfo struct {
	Name    string
	Version string
}

// FullName returns the complete model name.
//
// Returns:
//   - string: full model name with version
func (m ModelInfo) FullName() string {
	// Check if version is empty
	if m.Version == "" {
		// Return name only
		return m.Name
	}
	// Return name with version
	return m.Name + " " + m.Version
}
