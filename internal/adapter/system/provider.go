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
	dockerEnvPath    = "/.dockerenv"
	cgroupPath       = "/proc/1/cgroup"
	dockerIdentifier = "docker"
)

var _ port.SystemProvider = (*Provider)(nil)

// Provider implements port.SystemProvider.
type Provider struct{}

// NewProvider creates a new system provider adapter.
func NewProvider() *Provider {
	return &Provider{}
}

// Info returns the current system information.
func (p *Provider) Info() model.SystemInfo {
	return model.SystemInfo{
		OS:       p.detectOS(),
		IsDocker: p.isDocker(),
	}
}

func (p *Provider) detectOS() model.OSType {
	switch runtime.GOOS {
	case "linux":
		return model.OSLinux
	case "darwin":
		return model.OSDarwin
	case "windows":
		return model.OSWindows
	default:
		return model.OSUnknown
	}
}

func (p *Provider) isDocker() bool {
	if _, err := os.Stat(dockerEnvPath); err == nil {
		return true
	}
	data, err := os.ReadFile(cgroupPath)
	if err == nil && strings.Contains(string(data), dockerIdentifier) {
		return true
	}
	return false
}
