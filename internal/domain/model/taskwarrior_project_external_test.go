package model_test

import (
	"testing"

	"github.com/florent/status-line/internal/domain/model"
)

func TestTaskwarriorProject_Total(t *testing.T) {
	tests := []struct {
		name    string
		project model.TaskwarriorProject
		want    int
	}{
		{name: "zero tasks", project: model.TaskwarriorProject{Pending: 0, Completed: 0}, want: 0},
		{name: "only pending", project: model.TaskwarriorProject{Pending: 5, Completed: 0}, want: 5},
		{name: "only completed", project: model.TaskwarriorProject{Pending: 0, Completed: 3}, want: 3},
		{name: "mixed", project: model.TaskwarriorProject{Pending: 5, Completed: 3}, want: 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.project.Total(); got != tt.want {
				t.Errorf("Total() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTaskwarriorProject_Percent(t *testing.T) {
	tests := []struct {
		name    string
		project model.TaskwarriorProject
		want    int
	}{
		{name: "zero total", project: model.TaskwarriorProject{Pending: 0, Completed: 0}, want: 0},
		{name: "50 percent", project: model.TaskwarriorProject{Pending: 5, Completed: 5}, want: 50},
		{name: "100 percent", project: model.TaskwarriorProject{Pending: 0, Completed: 10}, want: 100},
		{name: "0 percent", project: model.TaskwarriorProject{Pending: 10, Completed: 0}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.project.Percent(); got != tt.want {
				t.Errorf("Percent() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNewTaskwarriorProject(t *testing.T) {
	tests := []struct {
		name      string
		projName  string
		pending   int
		completed int
	}{
		{name: "basic project", projName: "test", pending: 5, completed: 3},
		{name: "empty name", projName: "", pending: 0, completed: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := model.NewTaskwarriorProject(tt.projName, tt.pending, tt.completed)
			if p.Name != tt.projName || p.Pending != tt.pending || p.Completed != tt.completed {
				t.Errorf("NewTaskwarriorProject() = %+v, want {Name:%s Pending:%d Completed:%d}", p, tt.projName, tt.pending, tt.completed)
			}
		})
	}
}
