// Package updater provides self-update functionality for the status-line binary.
// It checks GitHub releases for newer versions and downloads/replaces the binary.
package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Updater constants for GitHub API, versioning, and file operations.
const (
	// repoOwner is the GitHub repository owner.
	repoOwner string = "kodflow"
	// repoName is the GitHub repository name.
	repoName string = "status-line"
	// downloadURL is the release asset download URL pattern.
	downloadURL string = "https://github.com/%s/%s/releases/download/%s/%s"
	// checkTimeout bounds the release lookup: one small JSON document.
	checkTimeout time.Duration = 10 * time.Second
	// applyTimeout bounds the download of a whole binary and its checksum.
	applyTimeout time.Duration = 2 * time.Minute
	// semverMajorIdx is the index of major version component.
	semverMajorIdx int = 0
	// semverMinorIdx is the index of minor version component.
	semverMinorIdx int = 1
	// semverPatchIdx is the index of patch version component.
	semverPatchIdx int = 2
	// semverComponents is the number of semver components.
	semverComponents int = 3
	// executablePerm is the permission for executable files.
	executablePerm os.FileMode = 0755
	// checksumSuffix is appended to an asset name to reach its checksum file.
	checksumSuffix string = ".sha256"
	// DisableEnv switches the self-updater off entirely when non-empty.
	DisableEnv string = "STATUS_LINE_NO_SELF_UPDATE"
	// checksumMaxBytes bounds the checksum file read.
	checksumMaxBytes int64 = 1024
	// sha256HexLen is the length of a SHA-256 digest in hex.
	sha256HexLen int = 64
	// releaseMaxBytes bounds the release document read.
	releaseMaxBytes int64 = 1 << 20
	// tempPrefix names the partial downloads beside the binary; the sweep
	// removes the stale ones and nothing else carries this prefix.
	tempPrefix string = "status-line-update-"
)

// Updater handles self-update logic for status-line binary.
// It checks GitHub releases periodically and downloads newer versions.
type Updater struct {
	version string
	client  *http.Client
	// execPath is the binary to replace: the running one resolved through
	// its symlinks in production, a throwaway file in tests.
	execPath string
}

// Endpoints are variables so tests can point the updater at a local server:
// the alternative is reaching the real API, which would make these tests
// depend on the network and on whatever is currently published.
var (
	// apiURL is the GitHub endpoint naming the latest release.
	apiURL = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	// downloadURLFor builds the URL of a release asset.
	downloadURLFor = func(owner, repo, version, asset string) string {
		return fmt.Sprintf(downloadURL, owner, repo, version, asset)
	}
)

// newUpdater creates an updater with its own HTTP timeout.
//
// Params:
//   - version: current binary version (empty means dev build)
//   - execPath: binary to replace
//   - timeout: bound on every request this updater makes
//
// Returns:
//   - *Updater: configured updater instance
func newUpdater(version, execPath string, timeout time.Duration) *Updater {
	// Return configured updater
	return &Updater{
		version:  version,
		client:   &http.Client{Timeout: timeout},
		execPath: execPath,
	}
}

// getLatestVersion fetches the latest release tag from GitHub.
//
// A refusal carrying a rate-limit hint comes back as a *statusError, so the
// caller can wait as long as the API asked rather than its own backoff.
//
// Params:
//   - now: clock used to turn a relative Retry-After into a date
//
// Returns:
//   - string: latest version tag, validated
//   - error: any API error
func (u *Updater) getLatestVersion(now time.Time) (string, error) {
	resp, err := u.client.Get(apiURL)
	// Check for HTTP errors
	if err != nil {
		// Return HTTP error
		return "", fmt.Errorf("fetching release info: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		// Return status error, with the date the API asked us to wait until
		return "", &statusError{code: resp.StatusCode, retryAt: retryAt(resp.Header, now)}
	}

	var release releaseInfo
	// Check for JSON decode errors
	if err := json.NewDecoder(io.LimitReader(resp.Body, releaseMaxBytes)).Decode(&release); err != nil {
		// Return decode error
		return "", fmt.Errorf("parsing release info: %w", err)
	}
	// A tag that is not a plain release version is never installed
	if !validTag(release.TagName) {
		// Return validation error
		return "", fmt.Errorf("invalid release tag %q", release.TagName)
	}

	// Return parsed tag name
	return release.TagName, nil
}

// isNewer checks if the given version is newer than current.
//
// Params:
//   - latest: version to compare against
//
// Returns:
//   - bool: true if latest is newer
func (u *Updater) isNewer(latest string) bool {
	current := u.parseVersion(u.version)
	remote := u.parseVersion(latest)

	// Compare major version
	if remote[semverMajorIdx] > current[semverMajorIdx] {
		// Remote major is higher
		return true
	}
	// Check if major is equal
	if remote[semverMajorIdx] < current[semverMajorIdx] {
		// Current major is higher
		return false
	}

	// Compare minor version
	if remote[semverMinorIdx] > current[semverMinorIdx] {
		// Remote minor is higher
		return true
	}
	// Check if minor is equal
	if remote[semverMinorIdx] < current[semverMinorIdx] {
		// Current minor is higher
		return false
	}

	// Compare patch version
	return remote[semverPatchIdx] > current[semverPatchIdx]
}

// parseVersion parses a semver string into components.
//
// Params:
//   - v: version string (e.g., "v1.2.3")
//
// Returns:
//   - [3]int: major, minor, patch as integers
func (u *Updater) parseVersion(v string) [semverComponents]int {
	// Remove 'v' prefix
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")

	var result [semverComponents]int
	// Parse each component
	for i := 0; i < semverComponents && i < len(parts); i++ {
		// Ignore parse errors, default to 0
		result[i], _ = strconv.Atoi(parts[i])
	}
	// Return parsed components
	return result
}

// downloadAndReplace downloads the new binary and replaces the current one.
//
// Params:
//   - version: version to download
//
// Returns:
//   - error: any download or replacement error
func (u *Updater) downloadAndReplace(version string) error {
	// Get binary name for current platform
	binaryName := u.getBinaryName()
	url := downloadURLFor(repoOwner, repoName, version, binaryName)

	// Fetch the published checksum first: downloading the binary only to find
	// there is nothing to check it against wastes the transfer, and a missing
	// checksum must stop the update rather than wave it through.
	expected, err := u.fetchChecksum(version, binaryName)
	// Check for checksum retrieval errors
	if err != nil {
		// Return checksum error
		return fmt.Errorf("fetching checksum: %w", err)
	}

	// Download new binary
	resp, err := u.client.Get(url)
	// Check for HTTP errors
	if err != nil {
		// Return HTTP error
		return fmt.Errorf("downloading binary: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		// Return status error
		return fmt.Errorf("download failed: status %d", resp.StatusCode)
	}

	execPath := u.execPath

	// Create temporary file
	tmpFile, err := os.CreateTemp(filepath.Dir(execPath), tempPrefix+"*")
	// Check for temp file creation errors
	if err != nil {
		// Return temp file error
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Hash while writing: the bytes are streamed once, and the digest covers
	// exactly what landed on disk rather than what was meant to
	hasher := sha256.New()
	_, err = io.Copy(io.MultiWriter(tmpFile, hasher), resp.Body)
	tmpFile.Close()
	// Check for write errors
	if err != nil {
		os.Remove(tmpPath)
		// Return write error
		return fmt.Errorf("writing temp file: %w", err)
	}

	// Refuse anything that does not match the published digest. This binary
	// replaces itself in place, so an unverified download would be executed on
	// every redraw from here on.
	actual := hex.EncodeToString(hasher.Sum(nil))
	if !strings.EqualFold(actual, expected) {
		os.Remove(tmpPath)
		// Return checksum mismatch
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}

	// Make executable
	err = os.Chmod(tmpPath, executablePerm)
	// Check for chmod errors
	if err != nil {
		os.Remove(tmpPath)
		// Return chmod error
		return fmt.Errorf("setting permissions: %w", err)
	}

	// Replace old binary
	err = os.Rename(tmpPath, execPath)
	// Check for rename errors
	if err != nil {
		os.Remove(tmpPath)
		// Return rename error
		return fmt.Errorf("replacing binary: %w", err)
	}

	// Return success
	return nil
}

// getBinaryName returns the binary name for the current platform.
//
// Returns:
//   - string: binary name with platform suffix
func (u *Updater) getBinaryName() string {
	name := fmt.Sprintf("status-line-%s-%s", runtime.GOOS, runtime.GOARCH)
	// Add .exe extension for Windows
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	// Return platform-specific name
	return name
}

// fetchChecksum retrieves the published SHA-256 of a release asset.
//
// The checksum file is fetched from the same resolved release tag as the
// binary, so a tag moving between the two requests cannot pair a new binary
// with an old digest.
//
// Params:
//   - version: release tag being installed
//   - binaryName: asset name for the current platform
//
// Returns:
//   - string: expected hex digest
//   - error: any error while fetching or parsing it
func (u *Updater) fetchChecksum(version, binaryName string) (string, error) {
	url := downloadURLFor(repoOwner, repoName, version, binaryName+checksumSuffix)

	resp, err := u.client.Get(url)
	// A network failure leaves the running binary in place
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// A release without a checksum file is not one we can verify, and an
	// unverifiable update is refused rather than trusted
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(io.LimitReader(resp.Body, checksumMaxBytes))
	// An unreadable body carries no digest
	if err != nil {
		return "", err
	}

	// The file is "<hex>  <filename>", as sha256sum writes it
	digest := strings.TrimSpace(string(raw))
	if idx := strings.IndexAny(digest, " \t"); idx > 0 {
		digest = digest[:idx]
	}

	// A digest of the wrong length is not one, and comparing against it would
	// pass anything through
	if len(digest) != sha256HexLen {
		return "", fmt.Errorf("malformed checksum %q", digest)
	}
	return digest, nil
}
