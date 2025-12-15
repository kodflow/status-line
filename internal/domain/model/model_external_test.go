package model_test

import (
	"testing"

	"github.com/florent/status-line/internal/domain/model"
)

func TestModelInfo_FullName(t *testing.T) {
	tests := []struct {
		name  string
		info  model.ModelInfo
		want  string
	}{
		{name: "name only", info: model.ModelInfo{Name: "Claude"}, want: "Claude"},
		{name: "name with version", info: model.ModelInfo{Name: "Opus", Version: "4.5"}, want: "Opus 4.5"},
		{name: "empty name", info: model.ModelInfo{Name: "", Version: "1.0"}, want: " 1.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.info.FullName(); got != tt.want {
				t.Errorf("FullName() = %q, want %q", got, tt.want)
			}
		})
	}
}
