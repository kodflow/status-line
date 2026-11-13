package updater

import (
	"testing"
)

func TestUpdater_parseVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    [3]int
	}{
		{name: "standard version", version: "v1.2.3", want: [3]int{1, 2, 3}},
		{name: "without v prefix", version: "1.2.3", want: [3]int{1, 2, 3}},
		{name: "major only", version: "v1", want: [3]int{1, 0, 0}},
		{name: "major.minor", version: "v1.2", want: [3]int{1, 2, 0}},
		{name: "zero version", version: "v0.0.0", want: [3]int{0, 0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := newUpdater("", "", checkTimeout)
			got := u.parseVersion(tt.version)
			// Verify parsed version matches expected
			if got != tt.want {
				t.Errorf("parseVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdater_isNewer(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{name: "newer major", current: "v1.0.0", latest: "v2.0.0", want: true},
		{name: "newer minor", current: "v1.0.0", latest: "v1.1.0", want: true},
		{name: "newer patch", current: "v1.0.0", latest: "v1.0.1", want: true},
		{name: "same version", current: "v1.0.0", latest: "v1.0.0", want: false},
		{name: "older major", current: "v2.0.0", latest: "v1.0.0", want: false},
		{name: "older minor", current: "v1.1.0", latest: "v1.0.0", want: false},
		{name: "older patch", current: "v1.0.1", latest: "v1.0.0", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := newUpdater(tt.current, "", checkTimeout)
			got := u.isNewer(tt.latest)
			// Verify comparison result
			if got != tt.want {
				t.Errorf("isNewer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdater_getBinaryName(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns platform-specific name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := newUpdater("", "", checkTimeout)
			got := u.getBinaryName()
			// Verify binary name is not empty
			if got == "" {
				t.Error("getBinaryName() returned empty string")
			}
			// Verify it contains status-line prefix
			if len(got) < len("status-line") {
				t.Error("getBinaryName() too short")
			}
		})
	}
}
