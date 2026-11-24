// Package updater provides self-update functionality for the status-line binary.
package updater

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	// probeTimeout bounds the trial run of a downloaded binary.
	probeTimeout time.Duration = 5 * time.Second
	// probeMaxOutput bounds what the trial run may print.
	probeMaxOutput int = 4096
)

// probeVersion runs a downloaded binary with --version and requires it to
// name the release it was downloaded as.
//
// Params:
//   - path: the downloaded binary, already executable
//   - tag: the release it must report
//
// Returns:
//   - error: the binary did not start, did not finish in time, or named
//     another version
func probeVersion(path, tag string) error {
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, "--version")
	// The trial run must never start an updater of its own
	cmd.Env = append(os.Environ(), DisableEnv+"=1")
	var out bytes.Buffer
	cmd.Stdout = &out
	// A binary that cannot even start is never installed
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("trial run of %s: %w", tag, err)
	}
	got := strings.TrimSpace(out.String())
	// Keep the error readable whatever the binary printed
	if len(got) > probeMaxOutput {
		got = got[:probeMaxOutput]
	}
	// The release must say it is the release
	if got != "status-line "+tag {
		return fmt.Errorf("trial run of %s reported %q", tag, got)
	}
	return nil
}
