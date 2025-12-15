package model_test

import (
	"testing"

	"github.com/florent/status-line/internal/domain/model"
)

func TestGitStatus_IsInRepo(t *testing.T) {
	tests := []struct {
		name   string
		status model.GitStatus
		want   bool
	}{
		{name: "empty branch", status: model.GitStatus{Branch: ""}, want: false},
		{name: "with branch", status: model.GitStatus{Branch: "main"}, want: true},
		{name: "feature branch", status: model.GitStatus{Branch: "feature/test"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsInRepo(); got != tt.want {
				t.Errorf("IsInRepo() = %v, want %v", got, tt.want)
			}
		})
	}
}
