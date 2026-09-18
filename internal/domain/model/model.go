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

// ShortName returns the model name with only its version number.
// Display names carry parenthesised qualifiers such as "(1M context)" which
// restate what the context pill already shows in tokens, and which would
// crowd out the segments that carry information the line has nowhere else.
//
// Returns:
//   - string: model name with its bare version
func (m ModelInfo) ShortName() string {
	// A model without a version is already short
	if m.Version == "" {
		return m.Name
	}
	version := m.Version
	// Cut the version at its first qualifier
	for idx, ch := range version {
		// A space or a parenthesis opens the qualifier
		if ch == ' ' || ch == '(' {
			version = version[:idx]
			break
		}
	}
	// A version reduced to nothing leaves the name alone
	if version == "" {
		return m.Name
	}
	return m.Name + " " + version
}
