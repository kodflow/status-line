// Package health reads the public status page of Claude's services.
package health

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/florent/status-line/internal/adapter/detach"
	"github.com/florent/status-line/internal/adapter/diskcache"
	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

// Status page constants.
const (
	// summaryURL is the Statuspage summary of Claude's services.
	summaryURL string = "https://status.claude.com/api/v2/summary.json"
	// httpTimeout bounds the background fetch.
	httpTimeout time.Duration = 5 * time.Second
	// maxBodySize bounds the payload read; the summary is a few kilobytes.
	maxBodySize int64 = 1 << 20
	// cacheTTL is how long a fetched summary stays authoritative.
	cacheTTL time.Duration = 2 * time.Minute
	// maxAge is how old a summary may be and still be drawn: past it, a green
	// light would be a claim nobody has checked.
	maxAge time.Duration = 15 * time.Minute
	// cacheDirPrefix is the per-user cache directory name.
	cacheDirPrefix string = "status-line-health-"
	// cacheFileName is the cached summary file name.
	cacheFileName string = "summary.json"
	// RefreshFlag re-executes the binary in health-refresh mode.
	RefreshFlag string = "--refresh-health"
	// refreshEnv marks a process spawned solely to refresh the cache.
	refreshEnv string = "STATUSLINE_HEALTH_REFRESH"
	// excludedComponent names the component left out of the count: it serves
	// a separate public-sector deployment, not this account.
	excludedComponent string = "government"
)

// Compile-time interface implementation check.
var _ port.HealthProvider = (*Provider)(nil)

// Provider serves the service health from a disk cache refreshed out of band.
type Provider struct {
	cache   *diskcache.Cache
	url     string
	refresh func()
}

// summary is the part of the Statuspage payload this adapter reads.
type summary struct {
	Components []component `json:"components"`
}

// component is one service on the status page.
type component struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Group  bool   `json:"group"`
}

// NewProvider creates a provider reading Claude's status page.
//
// Returns:
//   - *Provider: provider backed by the per-user cache
func NewProvider() *Provider {
	// Refresh by re-executing this binary in the background
	return &Provider{
		cache:   diskcache.New(cacheDirPrefix, cacheFileName),
		url:     summaryURL,
		refresh: func() { detach.Spawn(RefreshFlag, refreshEnv) },
	}
}

// Health returns the aggregate service level without ever touching the network.
// A stale or missing summary triggers a background refresh; the first render
// of a session therefore shows nothing rather than waiting for the page.
//
// Returns:
//   - model.ServiceHealth: current level, HealthUnknown when none is recent
func (p *Provider) Health() model.ServiceHealth {
	data, age, cached := p.cache.Read()
	// Ask for a newer summary whenever this one is past its TTL
	if !cached || age >= cacheTTL {
		p.refresh()
	}
	// A missing or long-stale summary is not evidence of anything
	if !cached || age >= maxAge {
		return model.HealthUnknown
	}
	return classify(data)
}

// Refresh fetches the summary and stores it, ignoring the cache TTL.
// It is the entry point of the detached refresh process.
//
// Returns:
//   - error: any error during fetch, validation or write
func (p *Provider) Refresh() error {
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(p.url)
	// An unreachable page leaves the previous summary in place
	if err != nil {
		return fmt.Errorf("fetch status summary: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	// Only a successful response may replace the cache
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch status summary: HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	// A truncated body leaves the previous summary in place
	if err != nil {
		return fmt.Errorf("read status summary: %w", err)
	}
	// Never cache something that would not decode on the next render
	if err := json.Unmarshal(data, &summary{}); err != nil {
		return fmt.Errorf("decode status summary: %w", err)
	}
	return p.cache.Write(data)
}

// classify turns a cached summary into a health level.
//
// Params:
//   - data: raw Statuspage summary
//
// Returns:
//   - model.ServiceHealth: aggregate level, HealthUnknown when undecodable
func classify(data []byte) model.ServiceHealth {
	var parsed summary
	// An undecodable summary is not evidence of anything
	if err := json.Unmarshal(data, &parsed); err != nil {
		return model.HealthUnknown
	}
	states := make([]string, 0, len(parsed.Components))
	// Keep the individual services this account depends on
	for _, comp := range parsed.Components {
		// A group only restates the state of its members
		if comp.Group {
			continue
		}
		// The public-sector deployment does not serve this account
		if strings.Contains(strings.ToLower(comp.Name), excludedComponent) {
			continue
		}
		states = append(states, comp.Status)
	}
	return model.ClassifyHealth(states)
}
