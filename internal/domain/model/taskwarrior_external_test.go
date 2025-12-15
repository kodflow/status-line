package model_test

import (
	"testing"

	"github.com/florent/status-line/internal/domain/model"
)

func TestTaskwarriorInfo_HasProjects(t *testing.T) {
	tests := []struct {
		name string
		info model.TaskwarriorInfo
		want bool
	}{
		{name: "no projects", info: model.TaskwarriorInfo{Projects: nil}, want: false},
		{name: "empty slice", info: model.TaskwarriorInfo{Projects: []model.TaskwarriorProject{}}, want: false},
		{name: "with projects", info: model.TaskwarriorInfo{Projects: []model.TaskwarriorProject{{Name: "test"}}}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.info.HasProjects(); got != tt.want {
				t.Errorf("HasProjects() = %v, want %v", got, tt.want)
			}
		})
	}
}
