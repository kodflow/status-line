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

func TestTaskwarriorProject_HasSession(t *testing.T) {
	tests := []struct {
		name    string
		project model.TaskwarriorProject
		want    bool
	}{
		{name: "no epics", project: model.TaskwarriorProject{}, want: false},
		{name: "with epics", project: model.TaskwarriorProject{Epics: []model.TaskwarriorEpic{{ID: 1}}}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.project.HasSession(); got != tt.want {
				t.Errorf("HasSession() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskwarriorProject_IsPlanMode(t *testing.T) {
	tests := []struct {
		name    string
		project model.TaskwarriorProject
		want    bool
	}{
		{name: "plan mode", project: model.TaskwarriorProject{Mode: model.ModePlan}, want: true},
		{name: "bypass mode", project: model.TaskwarriorProject{Mode: model.ModeBypass}, want: false},
		{name: "empty mode", project: model.TaskwarriorProject{}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.project.IsPlanMode(); got != tt.want {
				t.Errorf("IsPlanMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskwarriorProject_TotalWithEpics(t *testing.T) {
	tests := []struct {
		name    string
		project model.TaskwarriorProject
		want    int
	}{
		{
			name: "with epics",
			project: model.TaskwarriorProject{
				Epics: []model.TaskwarriorEpic{
					{TotalCount: 4},
					{TotalCount: 3},
				},
			},
			want: 7,
		},
		{
			name:    "legacy mode",
			project: model.TaskwarriorProject{Pending: 5, Completed: 3},
			want:    8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.project.Total(); got != tt.want {
				t.Errorf("Total() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTaskwarriorProject_TotalDone(t *testing.T) {
	tests := []struct {
		name    string
		project model.TaskwarriorProject
		want    int
	}{
		{
			name: "with epics",
			project: model.TaskwarriorProject{
				Epics: []model.TaskwarriorEpic{
					{DoneCount: 4},
					{DoneCount: 2},
				},
			},
			want: 6,
		},
		{
			name:    "legacy mode",
			project: model.TaskwarriorProject{Pending: 5, Completed: 3},
			want:    3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.project.TotalDone(); got != tt.want {
				t.Errorf("TotalDone() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTaskwarriorProject_CurrentTaskName(t *testing.T) {
	tests := []struct {
		name    string
		project model.TaskwarriorProject
		want    string
	}{
		{name: "no epics", project: model.TaskwarriorProject{}, want: ""},
		{
			name: "with wip task",
			project: model.TaskwarriorProject{
				Epics: []model.TaskwarriorEpic{
					{
						Tasks: []model.TaskwarriorTask{
							{ID: "1.1", Name: "Task1", Status: model.StatusDone},
							{ID: "1.2", Name: "Current", Status: model.StatusWip},
						},
					},
				},
			},
			want: "Current",
		},
		{
			name: "no wip task",
			project: model.TaskwarriorProject{
				Epics: []model.TaskwarriorEpic{
					{
						Tasks: []model.TaskwarriorTask{
							{ID: "1.1", Name: "Task1", Status: model.StatusDone},
						},
					},
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.project.CurrentTaskName(); got != tt.want {
				t.Errorf("CurrentTaskName() = %q, want %q", got, tt.want)
			}
		})
	}
}
