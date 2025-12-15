// Package model contains domain entities and value objects.
package model

// OSType represents the operating system type.
// It is used to determine the appropriate icon to display.
type OSType int

// Operating system type constants.
const (
	// OSLinux represents Linux-based systems.
	OSLinux OSType = iota
	// OSDarwin represents macOS systems.
	OSDarwin
	// OSWindows represents Windows systems.
	OSWindows
	// OSUnknown represents unknown or unsupported systems.
	OSUnknown
)

// SystemInfo contains system information.
// It holds OS type and Docker status.
type SystemInfo struct {
	OS       OSType
	IsDocker bool
}
