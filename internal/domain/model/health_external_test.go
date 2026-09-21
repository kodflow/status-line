package model_test

import (
	"testing"

	"github.com/florent/status-line/internal/domain/model"
)

func TestClassifyHealth(t *testing.T) {
	tests := []struct {
		name   string
		states []string
		want   model.ServiceHealth
	}{
		{name: "nothing to judge", states: nil, want: model.HealthUnknown},
		{name: "all operational", states: []string{"operational", "operational"}, want: model.HealthOK},
		{name: "maintenance is not an outage", states: []string{"under_maintenance", "operational"}, want: model.HealthOK},
		{name: "one degraded", states: []string{"degraded_performance", "operational"}, want: model.HealthDegraded},
		{name: "one partial outage", states: []string{"partial_outage", "operational"}, want: model.HealthDegraded},
		{name: "two degraded", states: []string{"degraded_performance", "partial_outage"}, want: model.HealthDown},
		{name: "one major outage", states: []string{"operational", "major_outage"}, want: model.HealthDown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := model.ClassifyHealth(tt.states); got != tt.want {
				t.Errorf("ClassifyHealth(%v) = %v, want %v", tt.states, got, tt.want)
			}
		})
	}
}
