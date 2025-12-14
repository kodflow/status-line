// Package system provides the system information adapter.
package system

import (
	"os"
	"runtime"
	"strings"

	"github.com/florent/status-line/internal/domain/model"
	"github.com/florent/status-line/internal/domain/port"
)

const (
	// dockerEnvPath is the path to the Docker environment file.
	dockerEnvPath string = "/.dockerenv"
	// cgroupPath is the path to the cgroup file for Docker detection.
	cgroupPath string = "/proc/1/cgroup"
	// dockerIdentifier is the string to search for in cgroup.
	dockerIdentifier string = "docker"
)

// Compile-time interface implementation check.
var _ port.SystemProvider = (*Provider)(nil)

// Provider implements port.SystemProvider using runtime detection.
// It detects the operating system type and Docker environment.
type Provider struct{}

// NewProvider creates a new system provider adapter.
//
// Returns:
//   - *Provider: new provider instance
func NewProvider() *Provider {
	// Return empty struct as no state is needed
	return &Provider{}
}

// Info returns the current system information.
//
// Returns:
//   - model.SystemInfo: OS type and Docker status
func (p *Provider) Info() model.SystemInfo {
	// Return aggregated system info
	return model.SystemInfo{
		OS:       p.detectOS(),
		IsDocker: p.isDocker(),
	}
}

// detectOS determines the current operating system.
//
// Returns:
//   - model.OSType: detected operating system type
func (p *Provider) detectOS() model.OSType {
	// Determine OS based on runtime.GOOS
	switch runtime.GOOS {
	// Handle Linux operating system
	case "linux":
		// Return Linux type
		return model.OSLinux
	// Handle macOS operating system
	case "darwin":
		// Return Darwin type
		return model.OSDarwin
	// Handle Windows operating system
	case "windows":
		// Return Windows type
		return model.OSWindows
	// Handle unknown operating system
	default:
		// Return unknown type
		return model.OSUnknown
	}
}

// isDocker returns true if running inside a Docker container.
//
// Returns:
//   - bool: true if in Docker environment
func (p *Provider) isDocker() bool {
	// Check for .dockerenv file
	if _, err := os.Stat(dockerEnvPath); err == nil {
		// Docker environment file exists
		return true
	}
	data, err := os.ReadFile(cgroupPath)
	// Check if cgroup contains docker string
	if err == nil && strings.Contains(string(data), dockerIdentifier) {
		// Running in Docker container
		return true
	}
	// Not running in Docker
	return false
}
