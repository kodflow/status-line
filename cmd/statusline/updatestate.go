// Package main provides the entry point for the status-line CLI tool.
package main

import "github.com/florent/status-line/internal/domain/model"

// updateState is what a frame takes from the cached update state.
type updateState struct {
	// info is what the update pill shows.
	info model.UpdateInfo
	// due says the next check may start.
	due bool
}
