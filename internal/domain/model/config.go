// Package model contains domain entities and value objects.
package model

import (
	"os"
	"strings"
)

// IconConfig holds configuration for icon visibility per component.
// It controls which icons are displayed in the status line.
type IconConfig struct {
	OS    bool
	Path  bool
	Git   bool
	Model bool
}

// DefaultIconConfig returns the default configuration with all icons enabled.
//
// Returns:
//   - IconConfig: configuration with all icons enabled
func DefaultIconConfig() IconConfig {
	// Return config with all icons enabled
	return IconConfig{
		OS:    true,
		Path:  true,
		Git:   true,
		Model: true,
	}
}

// IconConfigFromEnv reads icon configuration from environment variables.
//
// Returns:
//   - IconConfig: configuration based on environment variables
func IconConfigFromEnv() IconConfig {
	config := DefaultIconConfig()

	// Check OS icon setting
	if val := os.Getenv("STATUSLINE_ICON_OS"); val != "" {
		config.OS = parseBool(val)
	}
	// Check Path icon setting
	if val := os.Getenv("STATUSLINE_ICON_PATH"); val != "" {
		config.Path = parseBool(val)
	}
	// Check Git icon setting
	if val := os.Getenv("STATUSLINE_ICON_GIT"); val != "" {
		config.Git = parseBool(val)
	}
	// Check Model icon setting
	if val := os.Getenv("STATUSLINE_ICON_MODEL"); val != "" {
		config.Model = parseBool(val)
	}

	// Return configured settings
	return config
}

// parseBool parses a string to boolean, defaulting to true.
//
// Params:
//   - s: string to parse
//
// Returns:
//   - bool: parsed boolean value
func parseBool(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	// Return false only for explicit false values
	return lower != "false" && lower != "0" && lower != "no"
}
