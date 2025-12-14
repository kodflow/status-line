// Package model contains domain entities and value objects.
package model

// CodeChanges represents lines of code added and removed.
// It tracks code modifications in the session.
type CodeChanges struct {
	Added   int
	Removed int
}

// HasChanges returns true if there are any code changes.
//
// Returns:
//   - bool: true if added or removed is greater than zero
func (c CodeChanges) HasChanges() bool {
	// Check if either count is positive
	return c.Added > 0 || c.Removed > 0
}

// HasAdded returns true if lines were added.
//
// Returns:
//   - bool: true if added is greater than zero
func (c CodeChanges) HasAdded() bool {
	// Check if added count is positive
	return c.Added > 0
}

// HasRemoved returns true if lines were removed.
//
// Returns:
//   - bool: true if removed is greater than zero
func (c CodeChanges) HasRemoved() bool {
	// Check if removed count is positive
	return c.Removed > 0
}
