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
