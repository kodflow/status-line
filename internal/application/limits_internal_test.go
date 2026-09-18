package application

import (
	"testing"
	"time"

	"github.com/florent/status-line/internal/domain/model"
)

func stdinLimit(kind model.LimitKind, label string, percent int) model.Limit {
	return model.NewLimit(kind, label, percent, time.Now().Add(time.Hour), model.SessionWindow, model.SourceStdin)
}

func apiLimit(kind model.LimitKind, label string, percent int) model.Limit {
	return model.NewLimit(kind, label, percent, time.Now().Add(time.Hour), model.SessionWindow, model.SourceAPI)
}

func TestResolveLimits_StdinWinsOverAPI(t *testing.T) {
	stdin := model.LimitSet{Session: stdinLimit(model.KindSession, "session", 15)}
	api := model.LimitSet{Session: apiLimit(model.KindSession, "session", 99)}

	got := resolveLimits(stdin, api)

	// stdin is synchronous and current; the API is cached and therefore behind
	if got.Session.Percent != 15 {
		t.Errorf("Session.Percent = %d, want 15 (stdin value)", got.Session.Percent)
	}
	if got.Session.Source != model.SourceStdin {
		t.Errorf("Session.Source = %q, want %q", got.Session.Source, model.SourceStdin)
	}
}

func TestResolveLimits_APIFillsWhatStdinOmits(t *testing.T) {
	stdin := model.LimitSet{Session: stdinLimit(model.KindSession, "session", 15)}
	api := model.LimitSet{
		Weekly: apiLimit(model.KindWeekly, "weekly", 61),
		Scoped: []model.Limit{apiLimit(model.KindScoped, "fable", 13)},
		Extra:  model.ExtraUsage{Enabled: true, Percent: 25, Source: model.SourceAPI},
	}

	got := resolveLimits(stdin, api)

	// A weekly quota stdin does not carry must still come through
	if !got.Weekly.IsValid() || got.Weekly.Percent != 61 {
		t.Errorf("Weekly = %+v, want the API value", got.Weekly)
	}
	// Model-scoped quotas exist only in the API and must never be dropped
	if len(got.Scoped) != 1 || got.Scoped[0].Label != "fable" {
		t.Errorf("Scoped = %+v, want the API entry", got.Scoped)
	}
	// The credit balance exists only in the API
	if !got.Extra.IsValid() {
		t.Error("Extra must survive the merge")
	}
}

func TestResolveLimits_AbsentQuotaStaysAbsent(t *testing.T) {
	stdin := model.LimitSet{Session: stdinLimit(model.KindSession, "session", 15)}

	got := resolveLimits(stdin, model.LimitSet{})

	// A plan with no weekly cap must render nothing rather than a 0% bar, which
	// would claim the account has a weekly quota it has not touched
	if got.Weekly.IsValid() {
		t.Errorf("Weekly = %+v, want absent when neither source carries one", got.Weekly)
	}
	if len(got.Timed()) != 1 {
		t.Errorf("Timed() = %d quotas, want only the session one", len(got.Timed()))
	}
}

func TestResolveLimits_ContextIsNeverOverwritten(t *testing.T) {
	stdin := model.LimitSet{
		Context: model.NewLimit(model.KindContext, "ctx", 10, time.Time{}, 0, model.SourceStdin),
		Session: stdinLimit(model.KindSession, "session", 87),
	}
	api := model.LimitSet{Session: apiLimit(model.KindSession, "session", 87)}

	got := resolveLimits(stdin, api)

	// The context window measures the conversation, the session quota measures
	// the account. Letting one stand in for the other hides whichever is worse.
	if got.Context.Percent != 10 {
		t.Errorf("Context.Percent = %d, want 10 regardless of the session quota", got.Context.Percent)
	}
}

func TestResolveLimits_NoSourceAtAll(t *testing.T) {
	got := resolveLimits(model.LimitSet{}, model.LimitSet{})

	// With nothing to show the set must be empty, not fabricated
	if got.HasTimed() {
		t.Errorf("Timed() = %+v, want empty", got.Timed())
	}
}
