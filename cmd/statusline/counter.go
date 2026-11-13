// Package main provides the entry point for the status-line CLI tool.
package main

import "io"

// countingWriter counts the bytes that actually reached stdout, so the exit
// code can follow what the host will read.
type countingWriter struct {
	w io.Writer
	n int
}

// Write forwards to the underlying writer and counts what it took.
//
// Params:
//   - p: bytes to write
//
// Returns:
//   - int: bytes written
//   - error: the underlying writer's error
func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += n
	return n, err
}
