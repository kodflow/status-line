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
	credentialsFileName string = "credentials.json"
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

// usageResponse represents the API response structure.
type usageResponse struct {
	SevenDay usagePeriod `json:"seven_day"`
}

// usagePeriod represents a usage period from API.
type usagePeriod struct {
	Utilization float64 `json:"utilization"`
	ResetsAt    string  `json:"resets_at"`
}

// credentialsFile represents the credentials JSON structure.
type credentialsFile struct {
	AccessToken string `json:"accessToken"`
}

// Usage returns the current weekly API usage.
//
// Returns:
//   - model.Usage: utilization and reset time
//   - error: any error during fetch
func (p *Provider) Usage() (model.Usage, error) {
	// Get OAuth token
	token, err := p.getToken()
	// Check if token retrieval failed
	if err != nil {
		// Return empty usage on error
		return model.Usage{}, err
	}

	// Create API request
	req, err := http.NewRequest(http.MethodGet, usageAPIURL, nil)
	// Check if request creation failed
	if err != nil {
		// Return empty usage on error
		return model.Usage{}, err
	}

	// Set authorization header
	req.Header.Set("Authorization", "Bearer "+token)
	// Set required beta header
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")

	// Execute request
	resp, err := p.client.Do(req)
	// Check if request failed
	if err != nil {
		// Return empty usage on error
		return model.Usage{}, err
	}
	// Ensure body is closed
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	// Check if read failed
	if err != nil {
		// Return empty usage on error
		return model.Usage{}, err
	}

	// Parse JSON response
	var usage usageResponse
	// Check if parsing failed
	if err := json.Unmarshal(body, &usage); err != nil {
		// Return empty usage on error
		return model.Usage{}, err
	}

	// Parse reset time
	resetsAt, err := time.Parse(time.RFC3339, usage.SevenDay.ResetsAt)
	// Check if time parsing failed
	if err != nil {
		// Use zero time on parse error
		resetsAt = time.Time{}
	}

	// Return parsed usage data
	return model.NewUsage(int(usage.SevenDay.Utilization), resetsAt), nil
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
		// Return token if found
		if err == nil && token != "" {
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
	return creds.AccessToken, nil
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
	return strings.TrimSpace(creds.AccessToken), nil
}
