// Package main provides the entry point for the status-line CLI tool.
package main

import (
	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

// generator renders one frame from a payload and the update pill.
type generator interface {
	// GenerateWithUpdate renders the whole frame.
	GenerateWithUpdate(input port.InputProvider, update model.UpdateInfo) string
}
