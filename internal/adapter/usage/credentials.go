// Package usage provides the Anthropic API usage adapter.
package usage

// credentialsFile represents the credentials JSON structure.
// It wraps the OAuth credentials from Claude Code configuration.
type credentialsFile struct {
	ClaudeAiOauth oauthCredentials `json:"claudeAiOauth"`
}

// oauthCredentials represents the nested OAuth credentials.
// It contains the access token for API authentication.
type oauthCredentials struct {
	AccessToken string `json:"accessToken"`
}
