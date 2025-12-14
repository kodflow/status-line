// Package git provides the git repository adapter.
package git

import (
	"os/exec"
	"strings"

	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

const (
	statusCodeUntracked = "??"
	minStatusLineLength = 2
)

var _ port.GitRepository = (*Repository)(nil)

// Repository implements port.GitRepository.
type Repository struct{}

// NewRepository creates a new git repository adapter.
func NewRepository() *Repository {
	return &Repository{}
}

// Status retrieves the current git status.
func (r *Repository) Status() model.GitStatus {
	branch, err := r.getBranch()
	if err != nil {
		return model.GitStatus{}
	}

	modified, untracked := r.getChangeCounts()
	return model.GitStatus{
		Branch:    branch,
		Modified:  modified,
		Untracked: untracked,
	}
}

func (r *Repository) getBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (r *Repository) getChangeCounts() (modified, untracked int) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	for line := range strings.SplitSeq(string(output), "\n") {
		if len(line) < minStatusLineLength {
			continue
		}
		if line[:minStatusLineLength] == statusCodeUntracked {
			untracked++
		} else {
			modified++
		}
	}
	return modified, untracked
}
