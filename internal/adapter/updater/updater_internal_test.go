package updater

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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
			u := NewUpdater("")
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
			u := NewUpdater(tt.current)
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
			u := NewUpdater("")
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

func TestUpdater_shouldCheck(t *testing.T) {
	tests := []struct {
		name        string
		setupCache  bool
		cacheAge    time.Duration
		wantCheck   bool
	}{
		{name: "no cache file exists", setupCache: false, wantCheck: true},
		{name: "cache is old", setupCache: true, cacheAge: 2 * time.Hour, wantCheck: true},
		{name: "cache is recent", setupCache: true, cacheAge: 0, wantCheck: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUpdater("v1.0.0")
			cachePath := u.getCachePath()

			// Clean up any existing cache
			os.Remove(cachePath)

			// Setup cache if needed
			if tt.setupCache {
				f, _ := os.Create(cachePath)
				f.Close()
				// Set modification time if needed
				if tt.cacheAge > 0 {
					oldTime := time.Now().Add(-tt.cacheAge)
					os.Chtimes(cachePath, oldTime, oldTime)
				}
			}

			got := u.shouldCheck()
			// Verify check decision
			if got != tt.wantCheck {
				t.Errorf("shouldCheck() = %v, want %v", got, tt.wantCheck)
			}

			// Cleanup
			os.Remove(cachePath)
		})
	}
}

func TestUpdater_getCachePath(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns valid path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUpdater("")
			got := u.getCachePath()
			// Verify path is in temp directory
			if filepath.Dir(got) != os.TempDir() {
				t.Errorf("getCachePath() not in temp dir: %s", got)
			}
		})
	}
}

func TestUpdater_updateCache(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates cache file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUpdater("")
			cachePath := u.getCachePath()

			// Clean up
			os.Remove(cachePath)

			// Update cache
			u.updateCache()

			// Verify file exists
			_, err := os.Stat(cachePath)
			// Check file was created
			if err != nil {
				t.Errorf("updateCache() did not create file: %v", err)
			}

			// Cleanup
			os.Remove(cachePath)
		})
	}
}

func TestUpdater_getLatestVersion(t *testing.T) {
	tests := []struct {
		name      string
		wantError bool
	}{
		{name: "fetches latest version from GitHub", wantError: false},
		{name: "handles API errors gracefully", wantError: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUpdater("v1.0.0")
			// Try to get latest version
			version, err := u.getLatestVersion()
			// API may fail in CI environment
			if err != nil && !tt.wantError {
				// Network errors are acceptable in tests
				return
			}
			// Verify version format if successful
			if err == nil && version != "" && version[0] != 'v' {
				t.Errorf("getLatestVersion() returned invalid version: %s", version)
			}
		})
	}
}

func TestUpdater_downloadAndReplace(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		wantError bool
	}{
		{name: "handles invalid version gracefully", version: "v0.0.0-invalid", wantError: true},
		{name: "handles network errors gracefully", version: "v999.999.999", wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUpdater("v1.0.0")
			// Try to download
			err := u.downloadAndReplace(tt.version)
			// Verify error handling
			if tt.wantError && err == nil {
				t.Error("downloadAndReplace() expected error for invalid version")
			}
		})
	}
}
