package health

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/florent/status-line/internal/adapter/diskcache"
	"github.com/florent/status-line/internal/domain/model"
)

const (
	allUp = `{"components":[
		{"name":"claude.ai","status":"operational"},
		{"name":"Claude Code","status":"operational"},
		{"name":"Claude for Government","status":"major_outage"}]}`
	oneDown = `{"components":[
		{"name":"claude.ai","status":"partial_outage"},
		{"name":"Claude Code","status":"operational"}]}`
	groupOnly = `{"components":[
		{"name":"APIs","status":"major_outage","group":true},
		{"name":"Claude API","status":"operational"}]}`
)

// newTestProvider builds a provider on a private cache with a counting refresh.
func newTestProvider(t *testing.T, url string) (*Provider, *int) {
	t.Helper()
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	calls := 0
	return &Provider{
		cache:   diskcache.New(cacheDirPrefix, cacheFileName),
		url:     url,
		refresh: func() { calls++ },
	}, &calls
}

func TestClassify(t *testing.T) {
	tests := []struct {
		name string
		data string
		want model.ServiceHealth
	}{
		{name: "government outage is ignored", data: allUp, want: model.HealthOK},
		{name: "one partial outage", data: oneDown, want: model.HealthDegraded},
		{name: "groups restate their members", data: groupOnly, want: model.HealthOK},
		{name: "undecodable", data: "<html>", want: model.HealthUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classify([]byte(tt.data)); got != tt.want {
				t.Errorf("classify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHealthNeverBlocksAndRefreshesWhenStale(t *testing.T) {
	p, calls := newTestProvider(t, "http://unused.invalid")

	// Cold cache: nothing drawn, one refresh asked for
	if got := p.Health(); got != model.HealthUnknown {
		t.Fatalf("cold Health() = %v, want HealthUnknown", got)
	}
	if *calls != 1 {
		t.Fatalf("cold refresh calls = %d, want 1", *calls)
	}

	// Fresh cache: served as is, no refresh
	if err := p.cache.Write([]byte(oneDown)); err != nil {
		t.Fatal(err)
	}
	if got := p.Health(); got != model.HealthDegraded {
		t.Errorf("fresh Health() = %v, want HealthDegraded", got)
	}
	if *calls != 1 {
		t.Errorf("fresh refresh calls = %d, want 1", *calls)
	}

	// Past the TTL: still drawn, refresh asked for
	past := time.Now().Add(-cacheTTL - time.Second)
	if err := os.Chtimes(p.cache.Path(), past, past); err != nil {
		t.Fatal(err)
	}
	if got := p.Health(); got != model.HealthDegraded {
		t.Errorf("stale Health() = %v, want HealthDegraded", got)
	}
	if *calls != 2 {
		t.Errorf("stale refresh calls = %d, want 2", *calls)
	}

	// Past the maximum age: no longer evidence
	old := time.Now().Add(-maxAge - time.Second)
	if err := os.Chtimes(p.cache.Path(), old, old); err != nil {
		t.Fatal(err)
	}
	if got := p.Health(); got != model.HealthUnknown {
		t.Errorf("expired Health() = %v, want HealthUnknown", got)
	}
}

func TestRefresh(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		wantErr bool
		want    model.ServiceHealth
	}{
		{name: "stores a valid summary", status: http.StatusOK, body: oneDown, want: model.HealthDegraded},
		{name: "keeps the cache on HTTP error", status: http.StatusBadGateway, body: oneDown, wantErr: true, want: model.HealthUnknown},
		{name: "keeps the cache on garbage", status: http.StatusOK, body: "<html>", wantErr: true, want: model.HealthUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()
			p, _ := newTestProvider(t, srv.URL)
			if err := p.Refresh(); (err != nil) != tt.wantErr {
				t.Fatalf("Refresh() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := p.Health(); got != tt.want {
				t.Errorf("Health() after refresh = %v, want %v", got, tt.want)
			}
		})
	}
}
