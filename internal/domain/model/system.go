// Package model contains domain entities and value objects.
package model

// OSType represents the operating system type.
type OSType int

// Operating system type constants.
const (
	OSLinux   OSType = iota
	OSDarwin
	OSWindows
	OSUnknown
)

// SystemInfo contains system information.
type SystemInfo struct {
	OS       OSType
	IsDocker bool
}
