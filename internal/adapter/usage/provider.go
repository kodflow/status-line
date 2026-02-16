// Package usage provides the Anthropic API usage adapter.
package usage

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

// API and authentication constants.
const (
	// usageAPIURL is the Anthropic usage API endpoint.
	usageAPIURL string = "https://api.anthropic.com/api/oauth/usage"
	// httpTimeout is the timeout for API requests.
	httpTimeout time.Duration = 5 * time.Second
	// keychainService is the macOS keychain service name.
	keychainService string = "Claude Code-credentials"
	// credentialsFileName is the credentials file name.
	credentialsFileName string = ".credentials.json"
	// claudeConfigDir is the Claude configuration directory.
	claudeConfigDir string = ".claude"
)

// Compile-time interface implementation check.
var _ port.UsageProvider = (*Provider)(nil)

// Provider implements port.UsageProvider using Anthropic API.
// It fetches weekly usage data from the OAuth usage endpoint.
type Provider struct {
	client *http.Client
}

// NewProvider creates a new usage provider adapter.
//
// Returns:
//   - *Provider: new provider instance
func NewProvider() *Provider {
	// Return provider with configured HTTP client
	return &Provider{
		client: &http.Client{Timeout: httpTimeout},
	}
}


// Usage returns both session (5h) and weekly (7d) API usage.
//
// Returns:
//   - model.UsageData: session and weekly utilization and reset times
//   - error: any error during fetch
func (p *Provider) Usage() (model.UsageData, error) {
	// Get OAuth token
	token, err := p.getToken()
	if err != nil {
		return model.UsageData{}, err
	}

	// Create API request
	req, err := http.NewRequest(http.MethodGet, usageAPIURL, nil)
	if err != nil {
		return model.UsageData{}, err
	}

	// Set authorization header
	req.Header.Set("Authorization", "Bearer "+token)
	// Set required beta header
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")

	// Execute request
	resp, err := p.client.Do(req)
	if err != nil {
		return model.UsageData{}, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.UsageData{}, err
	}

	// Parse JSON response
	var usage usageResponse
	if err := json.Unmarshal(body, &usage); err != nil {
		return model.UsageData{}, err
	}

	// Parse session (five_hour) reset time
	sessionResetsAt, err := time.Parse(time.RFC3339, usage.FiveHour.ResetsAt)
	if err != nil {
		sessionResetsAt = time.Time{}
	}

	// Parse weekly (seven_day) reset time
	weeklyResetsAt, err := time.Parse(time.RFC3339, usage.SevenDay.ResetsAt)
	if err != nil {
		weeklyResetsAt = time.Time{}
	}

	return model.UsageData{
		Session: model.NewSessionUsage(int(usage.FiveHour.Utilization), sessionResetsAt),
		Weekly:  model.NewWeeklyUsage(int(usage.SevenDay.Utilization), weeklyResetsAt),
	}, nil
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
