// Package git provides the git repository adapter.
package git

import (
	"os/exec"
	"strings"

	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

const (
	// statusCodeUntracked is the git status code for untracked files.
	statusCodeUntracked string = "??"
	// minStatusLineLength is the minimum length for a valid status line.
	minStatusLineLength int = 2
	// minNumstatParts is the minimum fields in a numstat line (added, removed).
	minNumstatParts int = 2
	// base10 is the decimal base for parsing digits.
	base10 int = 10
)

// Compile-time interface implementation check.
var _ port.GitRepository = (*Repository)(nil)

// Repository implements port.GitRepository using git CLI commands.
// It retrieves git status information by executing shell commands.
type Repository struct{}

// NewRepository creates a new git repository adapter.
//
// Returns:
//   - *Repository: new repository instance
func NewRepository() *Repository {
	// Return empty struct as no state is needed
	return &Repository{}
}

// Status retrieves the current git status.
//
// Returns:
//   - model.GitStatus: branch and change information
func (r *Repository) Status() model.GitStatus {
	branch, err := r.getBranch()
	// Check if we're in a git repository
	if err != nil {
		// Return empty status if not in repo
		return model.GitStatus{}
	}

	modified, untracked := r.getChangeCounts()
	// Return populated status
	return model.GitStatus{
		Branch:    branch,
		Modified:  modified,
		Untracked: untracked,
	}
}

// getBranch retrieves the current branch name.
//
// Returns:
//   - string: branch name
//   - error: error if not in a git repository
func (r *Repository) getBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	// Check for git command errors
	if err != nil {
		// Return error if command failed
		return "", err
	}
	// Return trimmed branch name
	return strings.TrimSpace(string(output)), nil
}

// getChangeCounts counts modified and untracked files.
//
// Returns:
//   - modified: count of modified files
//   - untracked: count of untracked files
func (r *Repository) getChangeCounts() (modified, untracked int) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	// Check for git command errors
	if err != nil {
		// Return zero counts if command failed
		return 0, 0
	}

	// Iterate through status lines
	for line := range strings.SplitSeq(string(output), "\n") {
		// Skip lines that are too short
		if len(line) < minStatusLineLength {
			// Continue to next line
			continue
		}
		// Check if file is untracked
		if line[:minStatusLineLength] == statusCodeUntracked {
			untracked++
		} else {
			// Any other status means modified/staged
			modified++
		}
	}
	// Return calculated counts
	return modified, untracked
}

// DiffStats returns lines added and removed from git diff.
//
// Returns:
//   - model.CodeChanges: lines added and removed
func (r *Repository) DiffStats() model.CodeChanges {
	// Get diff stats for all changes (staged + unstaged)
	cmd := exec.Command("git", "diff", "--numstat", "HEAD")
	output, err := cmd.Output()
	// Check for git command errors
	if err != nil {
		// Return zero if command failed
		return model.CodeChanges{}
	}

	var added, removed int
	// Parse numstat output: "added removed filename"
	for line := range strings.SplitSeq(string(output), "\n") {
		// Skip empty lines
		if line == "" {
			continue
		}
		// Parse the line
		lineAdded, lineRemoved := parseNumstatLine(line)
		added += lineAdded
		removed += lineRemoved
	}

	// Return diff stats
	return model.CodeChanges{
		Added:   added,
		Removed: removed,
	}
}

// parseNumstatLine parses a single git diff --numstat line.
//
// Params:
//   - line: numstat line like "10 5 filename"
//
// Returns:
//   - added: lines added
//   - removed: lines removed
func parseNumstatLine(line string) (added, removed int) {
	parts := strings.Fields(line)
	// Need at least added and removed fields
	if len(parts) < minNumstatParts {
		// Return zero if line is malformed
		return 0, 0
	}
	// Parse added count (handle binary files showing "-")
	if parts[0] != "-" {
		// Convert string digits to integer
		for _, ch := range parts[0] {
			// Check if character is a digit
			if ch >= '0' && ch <= '9' {
				added = added*base10 + int(ch-'0')
			}
		}
	}
	// Parse removed count
	if parts[1] != "-" {
		// Convert string digits to integer
		for _, ch := range parts[1] {
			// Check if character is a digit
			if ch >= '0' && ch <= '9' {
				removed = removed*base10 + int(ch-'0')
			}
		}
	}
	// Return parsed counts
	return added, removed
}
