package model_test

import (
	"testing"
	"time"

	"github.com/florent/status-line/internal/domain/model"
)

func scoped(label string) model.Limit {
	return model.NewLimit(model.KindScoped, label, 13, time.Now().Add(20*time.Hour), model.WeeklyWindow, model.SourceAPI)
}

func TestLimitSet_ScopedFor(t *testing.T) {
	set := model.LimitSet{Scoped: []model.Limit{scoped("fable"), scoped("opus")}}

	tests := []struct {
		name      string
		modelName string
		want      []string
	}{
		{name: "running the scoped model", modelName: "Fable 5.1", want: []string{"fable"}},
		{name: "case does not matter", modelName: "FABLE 5.1", want: []string{"fable"}},
		{name: "another scoped model", modelName: "Opus 5 (1M context)", want: []string{"opus"}},
		{name: "unrelated model hides both", modelName: "Sonnet 5", want: nil},
		{name: "empty model name hides both", modelName: "", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := set.ScopedFor(tt.modelName)
			if len(got) != len(tt.want) {
				t.Fatalf("ScopedFor(%q) returned %d quotas, want %d", tt.modelName, len(got), len(tt.want))
			}
			for idx, label := range tt.want {
				if got[idx].Label != label {
					t.Errorf("ScopedFor(%q)[%d] = %q, want %q", tt.modelName, idx, got[idx].Label, label)
				}
			}
		})
	}
}

func TestLimitSet_ScopedFor_SkipsInvalid(t *testing.T) {
	set := model.LimitSet{Scoped: []model.Limit{{Kind: model.KindScoped, Label: "fable"}}}

	// An entry with no source was never populated and must not render
	if got := set.ScopedFor("Fable 5.1"); len(got) != 0 {
		t.Errorf("ScopedFor() returned %d quotas from an unpopulated entry, want 0", len(got))
	}
}
