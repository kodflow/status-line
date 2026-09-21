// Package usage provides the Anthropic API usage adapter.
package usage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/florent/status-line/internal/adapter/detach"
	"github.com/florent/status-line/internal/adapter/diskcache"
	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

// API and authentication constants.
const (
	// usageAPIURL is the Anthropic usage API endpoint.
	usageAPIURL string = "https://api.anthropic.com/api/oauth/usage"
	// httpTimeout bounds the only blocking fetch: the cold-cache one.
	httpTimeout time.Duration = 3 * time.Second
	// refreshFlag re-executes the binary in cache-refresh mode.
	refreshFlag string = "--refresh-usage"
	// keychainService is the macOS keychain service name.
	keychainService string = "Claude Code-credentials"
	// credentialsFileName is the credentials file name.
	credentialsFileName string = ".credentials.json"
	// claudeConfigDir is the Claude configuration directory.
	claudeConfigDir string = ".claude"
)

// Compile-time interface implementation check.
var _ port.UsageProvider = (*Provider)(nil)

// errDecode marks a payload that could not be parsed as usage data.
var errDecode = errors.New("usage api: undecodable payload")

// Provider implements port.UsageProvider using Anthropic API.
// It fetches weekly usage data from the OAuth usage endpoint.
type Provider struct {
	client *http.Client
	cache  *diskcache.Cache
}

// NewProvider creates a new usage provider adapter.
//
// Returns:
//   - *Provider: new provider instance
func NewProvider() *Provider {
	// Return provider with configured HTTP client
	return &Provider{
		client: &http.Client{Timeout: httpTimeout},
		cache:  diskcache.New(cacheDirPrefix, cacheFileName),
	}
}

// Limits returns every quota the account exposes.
// The cached payload is served immediately and a stale one triggers a
// detached refresh, so rendering never pays the network latency.
//
// Returns:
//   - model.LimitSet: session, weekly, scoped and extra quotas
//   - error: any error when no usable payload could be obtained
func (p *Provider) Limits() (model.LimitSet, error) {
	data, age, cached := p.cache.Read()

	// A fresh payload is authoritative; serve it without touching the network
	if cached && isFresh(age) {
		return decodeSet(data)
	}

	// A stale payload is still good enough to render; refresh behind the scenes
	if cached {
		p.refreshDetached()
		return decodeSet(data)
	}

	// Nothing cached: this is the only path that may block, and only once
	fresh, err := p.fetch()
	if err != nil {
		return model.LimitSet{}, err
	}
	// Persist for the next render; a cache write failure is not fatal
	_ = p.cache.Write(fresh)
	return decodeSet(fresh)
}

// Refresh fetches the API payload and stores it, ignoring the cache TTL.
// It is the entry point of the detached refresh process.
//
// Returns:
//   - error: any error during fetch or write
func (p *Provider) Refresh() error {
	fresh, err := p.fetch()
	// Leave the previous payload in place when the fetch fails
	if err != nil {
		return err
	}
	return p.cache.Write(fresh)
}

// refreshDetached refreshes the cache in a background process.
func (p *Provider) refreshDetached() {
	detach.Spawn(refreshFlag, refreshEnv)
}

// fetch performs the authenticated request against the usage endpoint.
//
// Returns:
//   - []byte: raw API payload
//   - error: any error during the request
func (p *Provider) fetch() ([]byte, error) {
	token, err := p.getToken()
	// Without a token the account simply has no API enrichment
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, usageAPIURL, nil)
	// A malformed request is a programming error, not a runtime state
	if err != nil {
		return nil, err
	}

	// Authenticate with the OAuth token
	req.Header.Set("Authorization", "Bearer "+token)
	// Opt into the usage beta the endpoint requires
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")

	resp, err := p.client.Do(req)
	// A network failure leaves the previous payload in place
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// A non-success status carries an error body, not usage data
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("usage api: status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// decodeSet turns a raw payload into the domain limit set.
//
// Params:
//   - data: raw API payload
//
// Returns:
//   - model.LimitSet: decoded quotas
//   - error: any error when the payload is not valid JSON
func decodeSet(data []byte) (model.LimitSet, error) {
	parsed, ok := decode(data)
	// A payload that does not decode must not masquerade as zero usage
	if !ok {
		return model.LimitSet{}, errDecode
	}
	return parsed.toLimitSet(), nil
}

// getToken retrieves the OAuth token from the appropriate source.
//
// Returns:
//   - string: OAuth access token
//   - error: any error during retrieval
func (p *Provider) getToken() (string, error) {
	// Try platform-specific methods first
	switch runtime.GOOS {
	// macOS uses keychain
	case "darwin":
		token, err := p.getTokenFromKeychain()
		// Check if token was successfully retrieved
		if err == nil && token != "" {
			// Return valid keychain token
			return token, nil
		}
	}

	// Fall back to credentials file
	return p.getTokenFromFile()
}

// getTokenFromKeychain retrieves the token from macOS keychain.
//
// Returns:
//   - string: OAuth access token
//   - error: any error during retrieval
func (p *Provider) getTokenFromKeychain() (string, error) {
	// Execute security command
	cmd := exec.Command("security", "find-generic-password", "-s", keychainService, "-w")
	output, err := cmd.Output()
	// Check if command failed
	if err != nil {
		// Return empty on error
		return "", err
	}

	// Parse JSON credentials
	var creds credentialsFile
	// Check if parsing failed
	if err := json.Unmarshal(output, &creds); err != nil {
		// Return empty on error
		return "", err
	}

	// Return access token
	return creds.ClaudeAiOauth.AccessToken, nil
}

// getTokenFromFile retrieves the token from credentials file.
//
// Returns:
//   - string: OAuth access token
//   - error: any error during retrieval
func (p *Provider) getTokenFromFile() (string, error) {
	// Get user home directory
	home, err := os.UserHomeDir()
	// Check if home lookup failed
	if err != nil {
		// Return empty on error
		return "", err
	}

	// Build credentials path
	credPath := filepath.Join(home, claudeConfigDir, credentialsFileName)
	// Read credentials file
	data, err := os.ReadFile(credPath)
	// Check if read failed
	if err != nil {
		// Return empty on error
		return "", err
	}

	// Parse JSON credentials
	var creds credentialsFile
	// Check if parsing failed
	if err := json.Unmarshal(data, &creds); err != nil {
		// Return empty on error
		return "", err
	}

	// Return access token
	return strings.TrimSpace(creds.ClaudeAiOauth.AccessToken), nil
}
