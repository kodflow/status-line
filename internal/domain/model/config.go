// Package model contains domain entities and value objects.
package model

import (
	"os"
	"strings"
)

// IconConfig holds configuration for icon visibility per component.
type IconConfig struct {
	OS    bool
	Path  bool
	Git   bool
	Model bool
}

// DefaultIconConfig returns the default configuration with all icons enabled.
func DefaultIconConfig() IconConfig {
	return IconConfig{
		OS:    true,
		Path:  true,
		Git:   true,
		Model: true,
	}
}

// IconConfigFromEnv reads icon configuration from environment variables.
func IconConfigFromEnv() IconConfig {
	config := DefaultIconConfig()

	if val := os.Getenv("STATUSLINE_ICON_OS"); val != "" {
		config.OS = parseBool(val)
	}
	if val := os.Getenv("STATUSLINE_ICON_PATH"); val != "" {
		config.Path = parseBool(val)
	}
	if val := os.Getenv("STATUSLINE_ICON_GIT"); val != "" {
		config.Git = parseBool(val)
	}
	if val := os.Getenv("STATUSLINE_ICON_MODEL"); val != "" {
		config.Model = parseBool(val)
	}

	return config
}

func parseBool(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	return lower != "false" && lower != "0" && lower != "no"
}
