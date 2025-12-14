// Package model contains domain entities and value objects.
package model

// ModelInfo contains AI model information.
type ModelInfo struct {
	Name    string
	Version string
}

// FullName returns the complete model name.
func (m ModelInfo) FullName() string {
	if m.Version == "" {
		return m.Name
	}
	return m.Name + " " + m.Version
}
