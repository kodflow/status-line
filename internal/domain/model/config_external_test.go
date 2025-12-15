package model_test

import (
	"testing"

	"github.com/florent/status-line/internal/domain/model"
)

func TestDefaultIconConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns all icons enabled"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := model.DefaultIconConfig()
			if !config.OS || !config.Path || !config.Git || !config.Model {
				t.Errorf("DefaultIconConfig() = %+v, want all true", config)
			}
		})
	}
}

func TestIconConfigFromEnv(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns config from env"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := model.IconConfigFromEnv()
			_ = config // Just verify no panic
		})
	}
}
