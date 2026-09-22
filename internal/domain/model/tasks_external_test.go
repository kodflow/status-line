package model_test

import (
	"testing"

	"github.com/florent/status-line/internal/domain/model"
)

func TestTaskList(t *testing.T) {
	list := model.TaskList{Items: []model.TaskItem{
		{ID: "1", Subject: "a", Status: model.TaskCompleted},
		{ID: "2", Subject: "b", Status: model.TaskInProgress},
		{ID: "3", Subject: "c", Status: model.TaskInProgress},
		{ID: "4", Subject: "d", Status: model.TaskPending},
	}}
	if list.Total() != 4 || list.Done() != 1 || list.Active() != 2 {
		t.Errorf("Total/Done/Active = %d/%d/%d, want 4/1/2", list.Total(), list.Done(), list.Active())
	}
	if got := list.Current(); got != "b" {
		t.Errorf("Current() = %q, want the earliest started task", got)
	}
	if !list.IsActive() {
		t.Error("IsActive() = false with open tasks")
	}
	finished := model.TaskList{Items: []model.TaskItem{{ID: "1", Status: model.TaskCompleted}}}
	if finished.IsActive() || (model.TaskList{}).IsActive() {
		t.Error("a finished or empty list must not be active")
	}
}
