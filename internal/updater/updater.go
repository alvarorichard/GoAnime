package updater

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"charm.land/huh/v2"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/alvarorichard/Goanime/v2/internal/util/jsonx"
	"github.com/alvarorichard/Goanime/v2/internal/version"
)

// maxJSONResponseBytes caps how much of an HTTP response jsonx.Decode will read
// before failing with jsonx.ErrTooLarge. The decoders this replaced
// (json.NewDecoder(resp.Body)) had no bound at all, so a hostile or broken
// upstream could stream until the process ran out of memory.
const maxJSONResponseBytes = 10 << 20 // 10 MiB

// checksumsAssetName is the file every release publishes with the SHA-256
// digest of each of its assets, in sha256sum format.
const checksumsAssetName = "checksums-sha256.txt"

// maxChecksumsBytes caps how much of the checksums file is read. The real file
// is under 1 KiB; anything near this size is not a checksums file.
const maxChecksumsBytes = 1 << 20 // 1 MiB

// Suffixes of the files the updater keeps next to the running executable.
// They are all replaced or removed before the update finishes, except for the
// backup, which survives until the new binary has started successfully.
const (
	updateBackupSuffix = ".backup"
	updateStagedSuffix = ".new"
	updateMarkerSuffix = ".update-pending"
	updateScriptSuffix = ".update.bat"
)

const (
	GitHubOwner = "alvarorichard"
	GitHubRepo  = "GoAnime"
	GitHubAPI   = "https://api.github.com/repos/" + GitHubOwner + "/" + GitHubRepo
)

// errUpdateDeferred means the replacement was handed to a helper process and
// will only happen after this process exits. It is not a failure and must not
// trigger a rollback: rolling back while the helper waits would overwrite the
// new binary as soon as the process ends.
var errUpdateDeferred = errors.New("update deferred until restart")

// Injectable hooks (overridden by tests).
var (
	releaseAPIURL       = GitHubAPI + "/releases/latest"
	findAssetFn         = findAssetForPlatform
	downloadFn          = downloadAsset
	fetchChecksumsFn    = fetchChecksums
	verifyDownloadFn    = verifyDownload
	osExecutableFn      = os.Executable
	replaceExecutableFn = replaceExecutable
	runForm             = tui.RunClean
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// CheckForUpdates checks if a new version is available on GitHub
func CheckForUpdates() (*GitHubRelease, bool, error) {
	return checkForUpdatesFromURL(releaseAPIURL, version.Version)
}

// checkForUpdatesFromURL is the internal implementation that accepts a custom
// API URL and current version string. This enables testing the full update
// check flow with a mock HTTP server.
func checkForUpdatesFromURL(apiURL, currentVer string) (*GitHubRelease, bool, error) {
	// Get latest release from GitHub API
	httpClient := &http.Client{Timeout: 60 * time.Second}
	resp, err := httpClient.Get(apiURL) // #nosec G107 -- URL is validated by caller or is a constant trusted GitHub API endpoint
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			util.Debug("Failed to close response body:", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := jsonx.Decode(resp.Body, maxJSONResponseBytes, &release); err != nil {
		return nil, false, fmt.Errorf("failed to decode release data: %w", err)
	}

	// Compare versions - strip "v" prefix from both to normalize
	currentVersion := strings.TrimPrefix(currentVer, "v")
	latestVersion := strings.TrimPrefix(release.TagName, "v")

	isNewer, err := isVersionNewer(latestVersion, currentVersion)
	if err != nil {
		return nil, false, fmt.Errorf("failed to compare versions: %w", err)
	}

	return &release, isNewer, nil
}

// PerformUpdate downloads and installs the latest version
func PerformUpdate(release *GitHubRelease) error {
	// Find the appropriate asset for current platform
	assetURL, assetName, err := findAssetFn(release)
	if err != nil {
		return err
	}

	util.Infof("Downloading update: %s", assetName)

	// Download the asset
	tempFile, err := downloadFn(assetURL, assetName)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer func() {
		if removeErr := os.Remove(tempFile); removeErr != nil {
			util.Debug("Failed to remove temp file:", removeErr)
		}
	}()

	// Verify before anything else touches the file, including zip extraction:
	// nothing from an unverified download may reach the executable's path.
	util.Info("Verifying the download against " + checksumsAssetName + "...")
	if err := verifyDownloadFn(release, tempFile, assetName); err != nil {
		return fmt.Errorf("update verification failed, nothing was installed: %w", err)
	}
	util.Info("Checksum verified")

	updateFile := tempFile
	cleanupUpdateFile := func() {}

	if runtime.GOOS == "windows" && strings.HasSuffix(strings.ToLower(assetName), ".zip") {
		util.Info("Extracting executable from Windows zip package...")

		extractedExe, cleanup, extractErr := extractExecutableFromZipAsset(tempFile)
		if extractErr != nil {
			return fmt.Errorf("failed to extract executable from update package: %w", extractErr)
		}

		updateFile = extractedExe
		cleanupUpdateFile = cleanup
	}
	defer cleanupUpdateFile()

	// Get current executable path
	currentExe, err := osExecutableFn()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	// Hash the update now. On Windows the helper uses the same digest to
	// verify the staged copy and the installed file; the marker left behind
	// lets the next start prove that this executable is the one we installed.
	newHash, err := fileSHA256(updateFile)
	if err != nil {
		return fmt.Errorf("failed to hash update: %w", err)
	}

	// Create a backup of the current executable. It is deliberately not
	// removed on success: CleanupAfterUpdate deletes it only after the new
	// binary has actually started, so a crash on startup still leaves a
	// working copy next to the executable.
	backupFile := updateBackupPath(currentExe)
	if err := copyFile(currentExe, backupFile); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Replace current executable
	if err := replaceExecutableFn(currentExe, updateFile); err != nil {
		// The Windows helper will finish after this process exits. Rolling
		// back here would overwrite the staged file as soon as we terminated.
		if errors.Is(err, errUpdateDeferred) {
			util.Info("Update will complete when you restart the application")
			return nil
		}

		// Try to restore the backup if replacement fails. If the restore
		// itself fails, keep the backup: it is the only working copy left.
		if _, backupErr := os.Stat(backupFile); backupErr == nil {
			if restoreErr := copyFile(backupFile, currentExe); restoreErr != nil {
				util.Warn("Failed to restore backup; it was kept at", backupFile+":", restoreErr)
			} else {
				util.Info("Successfully restored backup after failed update")
				if removeErr := os.Remove(backupFile); removeErr != nil {
					util.Debug("Failed to remove backup after restore:", removeErr)
				}
			}
		}

		return fmt.Errorf("failed to replace executable: %w", err)
	}

	// Replacement is in place. Record its digest so the next start can clean
	// up the backup and staged files, and prove the installed binary runs.
	if err := writeUpdateMarker(currentExe, newHash); err != nil {
		util.Warn("Update installed, but the post-update marker could not be written:", err)
	}

	util.Info("Update completed successfully! Please restart the application.")
	return nil
}

func extractExecutableFromZipAsset(zipPath string) (executablePath string, cleanup func(), err error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to open zip file: %w", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			util.Debug("Failed to close zip reader:", closeErr)
		}
	}()

	isPortableExe := func(name string) bool {
		lower := strings.ToLower(name)
		if !strings.HasSuffix(lower, ".exe") {
			return false
		}
		if strings.Contains(lower, "installer") {
			return false
		}
		return strings.HasPrefix(lower, "goanime")
	}

	isFallbackExe := func(name string) bool {
		lower := strings.ToLower(name)
		return strings.HasSuffix(lower, ".exe") && !strings.Contains(lower, "installer")
	}

	var selected *zip.File
	for _, f := range reader.File {
		if f.FileInfo().IsDir() {
			continue
		}

		entryName := filepath.Base(f.Name)
		if isPortableExe(entryName) {
			selected = f
			break
		}
	}

	if selected == nil {
		for _, f := range reader.File {
			if f.FileInfo().IsDir() {
				continue
			}

			entryName := filepath.Base(f.Name)
			if isFallbackExe(entryName) {
				selected = f
				break
			}
		}
	}

	if selected == nil {
		return "", nil, fmt.Errorf("no portable executable found in zip asset")
	}

	tempDir, err := os.MkdirTemp("", "goanime-update-extract-")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create extraction directory: %w", err)
	}

	cleanup = func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			util.Debug("Failed to remove extraction directory:", removeErr)
		}
	}

	entryReader, err := selected.Open()
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to open executable entry: %w", err)
	}
	defer func() {
		if closeErr := entryReader.Close(); closeErr != nil {
			util.Debug("Failed to close zip entry reader:", closeErr)
		}
	}()

	exeName := filepath.Base(selected.Name)
	if exeName == "" || exeName == "." || exeName == ".." || strings.ContainsAny(exeName, `/\`) {
		cleanup()
		return "", nil, fmt.Errorf("invalid executable name in zip: %q", exeName)
	}

	// Use os.Root to scope file creation under tempDir, preventing any path traversal.
	root, err := os.OpenRoot(tempDir)
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to open extraction root: %w", err)
	}
	defer func() {
		if closeErr := root.Close(); closeErr != nil {
			util.Debug("Failed to close extraction root:", closeErr)
		}
	}()

	outFile, err := root.Create(exeName)
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to create extracted executable: %w", err)
	}

	// Limit copy to 512 MiB to prevent zip-bomb DoS.
	const maxUpdateSize = 512 << 20
	if _, err := io.Copy(outFile, io.LimitReader(entryReader, maxUpdateSize)); err != nil {
		if closeErr := outFile.Close(); closeErr != nil {
			util.Debug("Failed to close extracted executable after copy error:", closeErr)
		}
		cleanup()
		return "", nil, fmt.Errorf("failed to extract executable: %w", err)
	}

	extractedPath := filepath.Join(tempDir, exeName)

	if closeErr := outFile.Close(); closeErr != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to close extracted executable: %w", closeErr)
	}

	return extractedPath, cleanup, nil
}

// PromptForUpdate shows an interactive prompt asking user if they want to update
func PromptForUpdate(release *GitHubRelease) (bool, error) {
	var shouldUpdate bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Update Available").
				Description(fmt.Sprintf("A new version of GoAnime is available!\n\n"+
					"Current version: %s\n"+
					"Latest version: %s\n\n"+
					"Release notes:\n%s",
					version.Version,
					release.TagName,
					truncateText(release.Body, 300))),

			huh.NewConfirm().
				Title("Would you like to update now?").
				Value(&shouldUpdate),
		),
	)

	if err := runForm(form.WithTheme(tui.FormTheme()).Run); err != nil {
		return false, fmt.Errorf("failed to show update prompt: %w", err)
	}

	return shouldUpdate, nil
}

// CheckAndPromptUpdate is a convenience function that checks for updates and prompts user
func CheckAndPromptUpdate() error {
	release, hasUpdate, err := CheckForUpdates()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !hasUpdate {
		util.Info("You are running the latest version!")
		return nil
	}

	shouldUpdate, err := PromptForUpdate(release)
	if err != nil {
		return err
	}

	if shouldUpdate {
		return PerformUpdate(release)
	}

	util.Info("Update cancelled by user")
	return nil
}

// CheckForUpdatesQuietly checks for updates without user interaction
func CheckForUpdatesQuietly() {
	release, hasUpdate, err := CheckForUpdates()
	if err != nil {
		util.Debug("Failed to check for updates:", err)
		return
	}

	if hasUpdate {
		util.Info(fmt.Sprintf("New version available: %s (current: %s)",
			release.TagName, version.Version))
		util.Info("Run with --update flag to update")
	}
}

// Helper functions

func isVersionNewer(latest, current string) (bool, error) {
	// Normalize: strip any "v" prefix that might have been left
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	latestParts := strings.Split(latest, ".")
	currentParts := strings.Split(current, ".")

	// Pad shorter version with zeros
	maxLen := max(len(currentParts), len(latestParts))

	for len(latestParts) < maxLen {
		latestParts = append(latestParts, "0")
	}
	for len(currentParts) < maxLen {
		currentParts = append(currentParts, "0")
	}

	// Compare each part
	for i := range maxLen {
		latestNum, err := strconv.Atoi(latestParts[i])
		if err != nil {
			return false, fmt.Errorf("invalid version format in latest: %s", latest)
		}

		currentNum, err := strconv.Atoi(currentParts[i])
		if err != nil {
			return false, fmt.Errorf("invalid version format in current: %s", current)
		}

		if latestNum > currentNum {
			return true, nil
		} else if latestNum < currentNum {
			return false, nil
		}
	}

	return false, nil // Versions are equal
}

// PlatformInfo holds platform-specific information
type PlatformInfo struct {
	OS   string
	Arch string
}

// GetCurrentPlatform returns the current platform information
func GetCurrentPlatform() PlatformInfo {
	return PlatformInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

func findAssetForPlatform(release *GitHubRelease) (downloadURL, filename string, err error) {
	return findAssetForPlatformWithInfo(release, GetCurrentPlatform())
}

func findAssetForPlatformWithInfo(release *GitHubRelease, platform PlatformInfo) (downloadURL, filename string, err error) {
	// Log available assets for debugging
	util.Debug("Looking for assets for platform:", platform.OS+"/"+platform.Arch)
	util.Debug("Available assets in release:")
	for _, asset := range release.Assets {
		util.Debug("  -", asset.Name)
	}

	// Map platform names to expected asset names
	var expectedNames []string
	switch platform.OS {
	case "windows":
		expectedNames = []string{
			fmt.Sprintf("goanime-windows-%s.exe", platform.Arch),
			fmt.Sprintf("goanime-windows-%s.zip", platform.Arch), // Also check zip files
			"goanime-windows.exe",
			"goanime.exe",
		}
	case "darwin":
		expectedNames = []string{
			fmt.Sprintf("goanime-darwin-%s", platform.Arch),
			fmt.Sprintf("goanime-macos-%s", platform.Arch),
			"goanime-darwin-universal", // Universal binary (explicit)
			"goanime-darwin",           // Universal binary (generic) or fallback
			"goanime-macos",            // Alternative generic name
		}
	case "linux":
		expectedNames = []string{
			fmt.Sprintf("goanime-linux-%s", platform.Arch),
			"goanime-linux",
			"goanime",
		}
	default:
		return "", "", fmt.Errorf("unsupported platform: %s", platform.OS)
	}

	util.Debug("Looking for expected names:", strings.Join(expectedNames, ", "))

	// Find matching asset
	for _, asset := range release.Assets {
		for _, expectedName := range expectedNames {
			if strings.EqualFold(asset.Name, expectedName) {
				util.Debug("Found matching asset:", asset.Name)
				return asset.BrowserDownloadURL, asset.Name, nil
			}
		}
	}

	return "", "", fmt.Errorf("no compatible asset found for %s/%s", platform.OS, platform.Arch)
}

// validateGitHubURLWithTestFlag validates URLs with optional test mode support
func validateGitHubURLWithTestFlag(urlStr string, allowTestMode bool) error {
	parsedURL, err := neturl.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Allow localhost and 127.0.0.1 in test mode
	if allowTestMode && (parsedURL.Host == "localhost" ||
		strings.HasPrefix(parsedURL.Host, "127.0.0.1") ||
		strings.HasPrefix(parsedURL.Host, "localhost:")) {
		return nil
	}

	// Only allow GitHub domains
	allowedHosts := []string{
		"github.com",
		"api.github.com",
		"objects.githubusercontent.com",
		"github-releases.githubusercontent.com",
	}

	hostAllowed := slices.Contains(allowedHosts, parsedURL.Host)

	if !hostAllowed {
		return fmt.Errorf("URL host %s not allowed", parsedURL.Host)
	}

	// Ensure HTTPS (except in test mode)
	if !allowTestMode && parsedURL.Scheme != "https" {
		return fmt.Errorf("only HTTPS URLs are allowed")
	}

	return nil
}

// validateFilePath validates that a file path is safe and doesn't contain directory traversal
func validateFilePath(path string) error {
	// Clean the path to resolve any .. or . components
	cleanPath := filepath.Clean(path)

	// Check for directory traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path contains directory traversal: %s", path)
	}

	// Ensure the path is absolute or within expected directories
	if !filepath.IsAbs(cleanPath) {
		// For relative paths, ensure they don't start with ..
		if strings.HasPrefix(cleanPath, "..") {
			return fmt.Errorf("relative path traversal detected: %s", path)
		}
	}

	return nil
}

// safeTempFile creates a validated temporary file path
func safeTempFile(filename string) (string, error) {
	// Sanitize filename
	filename = filepath.Base(filename) // Remove any path components
	if filename == "" || filename == "." || filename == ".." {
		return "", fmt.Errorf("invalid filename: %s", filename)
	}

	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, "goanime-update-"+filename)

	// Validate the resulting path
	if err := validateFilePath(tempFile); err != nil {
		return "", fmt.Errorf("temp file path validation failed: %w", err)
	}

	return tempFile, nil
}

func downloadAsset(url, filename string) (string, error) {
	return downloadAssetWithTestFlag(url, filename, false)
}

// downloadAssetWithTestFlag downloads an asset with optional test mode support
func downloadAssetWithTestFlag(url, filename string, allowTestMode bool) (string, error) {
	// Validate URL before making request
	if err := validateGitHubURLWithTestFlag(url, allowTestMode); err != nil {
		return "", fmt.Errorf("URL validation failed: %w", err)
	}

	// #nosec G107 - URL is validated above to ensure it's from trusted GitHub domains
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			util.Debug("Failed to close response body:", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temporary file with validation
	tempFile, err := safeTempFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create safe temp file: %w", err)
	}

	// #nosec G304 - tempFile is validated by safeTempFile function above
	out, err := os.Create(tempFile)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			util.Debug("Failed to close temp file:", closeErr)
		}
	}()

	// Copy with progress indication
	util.Info("Downloading...")
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		if removeErr := os.Remove(tempFile); removeErr != nil {
			util.Debug("Failed to remove temp file after error:", removeErr)
		}
		return "", err
	}

	return tempFile, nil
}

// verifyDownload checks a downloaded release asset against the release's
// checksums-sha256.txt. A release that does not publish the file, or a file
// with no entry for the asset, is refused: installing an unverified executable
// is never the fallback.
func verifyDownload(release *GitHubRelease, path, assetName string) error {
	url, err := findChecksumsAsset(release)
	if err != nil {
		return err
	}

	data, err := fetchChecksumsFn(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", checksumsAssetName, err)
	}

	sums, err := parseChecksums(data)
	if err != nil {
		return err
	}

	return verifyAssetChecksum(path, assetName, sums)
}

// findChecksumsAsset returns the download URL of the release's checksums file.
func findChecksumsAsset(release *GitHubRelease) (string, error) {
	for _, asset := range release.Assets {
		if strings.EqualFold(asset.Name, checksumsAssetName) {
			return asset.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("release %s does not publish %s", release.TagName, checksumsAssetName)
}

func fetchChecksums(url string) ([]byte, error) {
	return fetchChecksumsWithTestFlag(url, false)
}

// fetchChecksumsWithTestFlag downloads the checksums file under the same URL
// rules as downloadAssetWithTestFlag, reading at most maxChecksumsBytes.
func fetchChecksumsWithTestFlag(url string, allowTestMode bool) ([]byte, error) {
	if err := validateGitHubURLWithTestFlag(url, allowTestMode); err != nil {
		return nil, fmt.Errorf("URL validation failed: %w", err)
	}

	httpClient := &http.Client{Timeout: 60 * time.Second}
	resp, err := httpClient.Get(url) // #nosec G107 -- URL is validated above to ensure it's from trusted GitHub domains
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			util.Debug("Failed to close response body:", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxChecksumsBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxChecksumsBytes {
		return nil, fmt.Errorf("%s is larger than %d bytes", checksumsAssetName, maxChecksumsBytes)
	}
	return data, nil
}

// parseChecksums reads sha256sum output, one "<digest>  <name>" per line (or
// "<digest> *<name>" in binary mode), into a map from file name to lowercase
// hex digest. Any malformed line fails the whole file rather than being
// skipped, since a damaged checksums file proves nothing.
func parseChecksums(data []byte) (map[string]string, error) {
	sums := make(map[string]string)
	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		digest, name, ok := strings.Cut(line, " ")
		name = strings.TrimLeft(name, " *")
		if _, hexErr := hex.DecodeString(digest); !ok || name == "" ||
			len(digest) != sha256.Size*2 || hexErr != nil {
			return nil, fmt.Errorf("malformed line %d in %s", i+1, checksumsAssetName)
		}

		digest = strings.ToLower(digest)
		if prev, seen := sums[name]; seen && prev != digest {
			return nil, fmt.Errorf("%s lists conflicting checksums for %s", checksumsAssetName, name)
		}
		sums[name] = digest
	}

	if len(sums) == 0 {
		return nil, fmt.Errorf("%s lists no checksums", checksumsAssetName)
	}
	return sums, nil
}

// verifyAssetChecksum checks the file at path against the digest sums lists for
// assetName.
func verifyAssetChecksum(path, assetName string, sums map[string]string) error {
	want, ok := sums[assetName]
	if !ok {
		return fmt.Errorf("%s has no entry for %s", checksumsAssetName, assetName)
	}

	got, err := fileSHA256(path)
	if err != nil {
		return fmt.Errorf("failed to hash %s: %w", assetName, err)
	}

	if got != want {
		return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", assetName, want, got)
	}
	return nil
}

// fileSHA256 returns the lowercase hex SHA-256 digest of the file at path.
func fileSHA256(path string) (string, error) {
	if err := validateFilePath(path); err != nil {
		return "", err
	}

	// #nosec G304 - path is validated above
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			util.Debug("Failed to close file:", closeErr)
		}
	}()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func updateBackupPath(currentExe string) string {
	return currentExe + updateBackupSuffix
}

func updateStagedPath(currentExe string) string {
	return currentExe + updateStagedSuffix
}

func updateMarkerPath(currentExe string) string {
	return currentExe + updateMarkerSuffix
}

func updateScriptPath(currentExe string) string {
	return currentExe + updateScriptSuffix
}

// writeUpdateMarker records the digest of the newly installed executable.
// CleanupAfterUpdate only deletes the backup when the running executable
// still matches this digest, which means the update actually booted.
func writeUpdateMarker(currentExe, digest string) error {
	markerPath := updateMarkerPath(currentExe)
	if err := os.WriteFile(markerPath, []byte(strings.ToLower(digest)+"\n"), 0o600); err != nil {
		return fmt.Errorf("failed to write %s: %w", markerPath, err)
	}
	return nil
}

func parseUpdateMarker(data []byte) (string, error) {
	digest := strings.ToLower(strings.TrimSpace(string(data)))
	if len(digest) != sha256.Size*2 {
		return "", fmt.Errorf("marker does not contain a SHA-256 digest")
	}
	if _, err := hex.DecodeString(digest); err != nil {
		return "", fmt.Errorf("marker contains an invalid SHA-256 digest: %w", err)
	}
	return digest, nil
}

// CleanupAfterUpdate removes the backup, staged executable and helper script
// left by a completed update. It is safe to call on every start: without a
// marker it does nothing, and it only deletes the backup when the running
// executable matches the digest recorded by the updater. If anything does not
// match, the backup is deliberately kept so the user can restore it.
func CleanupAfterUpdate() {
	currentExe, err := osExecutableFn()
	if err != nil {
		util.Debug("Could not locate the executable for update cleanup:", err)
		return
	}
	if err := cleanupUpdateArtifacts(currentExe); err != nil {
		util.Warn("Could not finish update cleanup:", err)
	}
}

func cleanupUpdateArtifacts(currentExe string) error {
	markerPath := updateMarkerPath(currentExe)
	// #nosec G304 - markerPath is derived from the running executable, not user input.
	data, err := os.ReadFile(markerPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read update marker: %w", err)
	}

	expected, err := parseUpdateMarker(data)
	if err != nil {
		// A marker we cannot verify must never authorize deleting the backup.
		removeErr := os.Remove(markerPath)
		if removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return fmt.Errorf("invalid update marker (backup kept at %s); failed to remove it: %w",
				updateBackupPath(currentExe), removeErr)
		}
		return fmt.Errorf("invalid update marker (backup kept at %s): %w",
			updateBackupPath(currentExe), err)
	}

	current, err := fileSHA256(currentExe)
	if err != nil {
		return fmt.Errorf("failed to hash the running executable: %w", err)
	}
	if !strings.EqualFold(current, expected) {
		// The marker belongs to another build: the update was rolled back or
		// the binary was replaced by hand. Keep the backup, drop the marker so
		// the warning is not repeated on every launch.
		if removeErr := os.Remove(markerPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return fmt.Errorf("running executable does not match the update marker (backup kept at %s); failed to remove the marker: %w",
				updateBackupPath(currentExe), removeErr)
		}
		return fmt.Errorf("running executable does not match the update marker; backup kept at %s",
			updateBackupPath(currentExe))
	}

	// Remove the backup first. If it cannot be deleted, keep the marker so the
	// next start retries instead of leaving an unmarked backup behind.
	if err := os.Remove(updateBackupPath(currentExe)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to remove the update backup: %w", err)
	}
	if err := os.Remove(updateStagedPath(currentExe)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to remove the staged update: %w", err)
	}
	if err := os.Remove(updateScriptPath(currentExe)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to remove the update helper script: %w", err)
	}
	if err := os.Remove(markerPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to remove the update marker: %w", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	// Validate source and destination paths
	if err := validateFilePath(src); err != nil {
		return fmt.Errorf("source path validation failed: %w", err)
	}
	if err := validateFilePath(dst); err != nil {
		return fmt.Errorf("destination path validation failed: %w", err)
	}

	// #nosec G304 - src path is validated above
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := sourceFile.Close(); closeErr != nil {
			util.Debug("Failed to close source file:", closeErr)
		}
	}()

	// #nosec G304 - dst path is validated above
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := destFile.Close(); closeErr != nil {
			util.Debug("Failed to close destination file:", closeErr)
		}
	}()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

func replaceExecutable(currentExe, newExe string) error {
	// On Unix systems (Linux, macOS), we can move the running executable
	// and place the new one in its location. The running process continues
	// to use the moved executable until it exits.
	if runtime.GOOS != "windows" {
		// Generate a unique temporary name for the old executable
		tempName := currentExe + ".old." + fmt.Sprintf("%d", os.Getpid())

		// Move current executable to temporary location
		// This works even while the executable is running on Unix systems
		if err := os.Rename(currentExe, tempName); err != nil {
			return fmt.Errorf("failed to move current executable: %w", err)
		}

		// Schedule cleanup of the old executable
		defer func() {
			if removeErr := os.Remove(tempName); removeErr != nil {
				util.Debug("Failed to remove old executable:", removeErr)
			}
		}()

		// Copy new executable to the original location
		if err := copyFile(newExe, currentExe); err != nil {
			// Try to restore the original executable if copy fails
			if restoreErr := os.Rename(tempName, currentExe); restoreErr != nil {
				util.Debug("Failed to restore original executable:", restoreErr)
			}
			return fmt.Errorf("failed to copy new executable: %w", err)
		}

		// Make sure the new executable has proper permissions
		// #nosec G302 - 0o755 is appropriate for executable files
		if err := os.Chmod(currentExe, 0o755); err != nil {
			util.Debug("Failed to set executable permissions:", err)
			// Don't fail the update for permission issues
		}

		return nil
	}

	// Windows handling - more complex due to file locking
	// We need to use a different approach for Windows
	tempName := currentExe + ".old"

	// Try to rename the current executable
	// This may fail if the executable is in use
	if err := os.Rename(currentExe, tempName); err != nil {
		// If rename fails, try a different approach
		// Create a batch script that will replace the executable after this process exits
		return createWindowsUpdateScript(currentExe, newExe)
	}

	defer func() {
		if removeErr := os.Remove(tempName); removeErr != nil {
			util.Debug("Failed to remove old executable:", removeErr)
		}
	}()

	// Copy new executable to current location
	if err := copyFile(newExe, currentExe); err != nil {
		// Try to restore the original executable if copy fails
		if restoreErr := os.Rename(tempName, currentExe); restoreErr != nil {
			util.Debug("Failed to restore original executable:", restoreErr)
		}
		return fmt.Errorf("failed to copy new executable: %w", err)
	}

	return nil
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// windowsUpdatePlan describes the files the post-exit helper needs. Paths are
// passed through environment variables, not interpolated into the batch text,
// so spaces, ampersands and parentheses in the install path stay intact.
type windowsUpdatePlan struct {
	currentExe string
	stagedExe  string
	backupFile string
	markerFile string
	scriptPath string
	digest     string
}

func removeUpdateArtifact(path string) {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		util.Debug("Failed to remove update file:", err)
	}
}

func prepareWindowsUpdate(currentExe, newExe string) (windowsUpdatePlan, error) {
	plan := windowsUpdatePlan{
		currentExe: currentExe,
		stagedExe:  updateStagedPath(currentExe),
		backupFile: updateBackupPath(currentExe),
		markerFile: updateMarkerPath(currentExe),
		scriptPath: updateScriptPath(currentExe),
	}

	digest, err := fileSHA256(newExe)
	if err != nil {
		return plan, fmt.Errorf("failed to hash the downloaded executable: %w", err)
	}
	plan.digest = digest

	if _, err := os.Stat(currentExe); err != nil {
		return plan, fmt.Errorf("current executable is unavailable: %w", err)
	}

	// Stage next to the target, not in the temp directory: PerformUpdate
	// removes the extraction directory before the helper gets to run.
	if err := copyFile(newExe, plan.stagedExe); err != nil {
		return plan, fmt.Errorf("failed to stage the update: %w", err)
	}

	stagedDigest, err := fileSHA256(plan.stagedExe)
	if err != nil {
		removeUpdateArtifact(plan.stagedExe)
		return plan, fmt.Errorf("failed to hash the staged update: %w", err)
	}
	if !strings.EqualFold(stagedDigest, digest) {
		removeUpdateArtifact(plan.stagedExe)
		return plan, fmt.Errorf("staged update does not match the downloaded executable")
	}

	return plan, nil
}

func (p windowsUpdatePlan) environment() []string {
	return []string{
		"GOANIME_UPDATE_TARGET=" + p.currentExe,
		"GOANIME_UPDATE_STAGED=" + p.stagedExe,
		"GOANIME_UPDATE_BACKUP=" + p.backupFile,
		"GOANIME_UPDATE_MARKER=" + p.markerFile,
		"GOANIME_UPDATE_SHA256=" + p.digest,
		"GOANIME_UPDATE_PID=" + strconv.Itoa(os.Getpid()),
	}
}

// windowsUpdateScript is the post-exit helper. It waits for the updater
// process to disappear, verifies the staged file, copies it over the target,
// verifies the result and restores the backup if anything fails. Only after a
// verified copy does it write the marker that authorizes the next start to
// delete the backup.
func windowsUpdateScript() string {
	return `@echo off
setlocal EnableExtensions DisableDelayedExpansion

set "TARGET=%GOANIME_UPDATE_TARGET%"
set "STAGED=%GOANIME_UPDATE_STAGED%"
set "BACKUP=%GOANIME_UPDATE_BACKUP%"
set "MARKER=%GOANIME_UPDATE_MARKER%"
set "EXPECTED=%GOANIME_UPDATE_SHA256%"
set "WAITPID=%GOANIME_UPDATE_PID%"

if not defined TARGET goto :fail
if not defined STAGED goto :fail
if not defined BACKUP goto :fail
if not defined MARKER goto :fail
if not defined EXPECTED goto :fail
if not defined WAITPID goto :fail

:wait
tasklist /FI "PID eq %WAITPID%" 2>nul | find /I "%WAITPID%" >nul
if not errorlevel 1 (
    ping -n 2 127.0.0.1 >nul
    goto wait
)

call :hash "%STAGED%"
if /I not "%HASH%"=="%EXPECTED%" goto :staged_mismatch

if not exist "%TARGET%" goto :fail
if not exist "%BACKUP%" (
    copy /Y "%TARGET%" "%BACKUP%" >nul 2>&1
    if errorlevel 1 goto :fail
)

set /a ATTEMPTS=0
:replace
copy /Y "%STAGED%" "%TARGET%" >nul 2>&1
if not errorlevel 1 goto :verify
set /a ATTEMPTS+=1
if %ATTEMPTS% GEQ 15 goto :restore
ping -n 2 127.0.0.1 >nul
goto replace

:verify
call :hash "%TARGET%"
if /I "%HASH%"=="%EXPECTED%" goto :success

:restore
copy /Y "%BACKUP%" "%TARGET%" >nul 2>&1
del /F /Q "%STAGED%" >nul 2>&1
del /F /Q "%MARKER%" >nul 2>&1
goto :fail

:success
> "%MARKER%" echo %EXPECTED%
del /F /Q "%STAGED%" >nul 2>&1
call :scheduleSelfDelete
exit /b 0

:staged_mismatch
del /F /Q "%STAGED%" >nul 2>&1
goto :fail

:fail
call :scheduleSelfDelete
exit /b 1

:scheduleSelfDelete
rem Deleting the running batch file before exit makes cmd lose the exit
rem code (it returns 0 even after "exit /b 1"). Let a detached child wait
rem for this process to end and remove the script afterwards.
set "SELF=%~f0"
start "" /B cmd /C "ping -n 2 127.0.0.1 >nul & del /F /Q "%SELF%" >nul 2>&1"
exit /b 0

:hash
set "HASH="
for /f "skip=1 delims=" %%H in ('certutil -hashfile "%~1" SHA256 2^>nul') do (
    if not defined HASH set "HASH=%%H"
)
set "HASH=%HASH: =%"
exit /b 0
`
}

// createWindowsUpdateScript stages the new binary and starts a helper that
// replaces the running executable after this process exits.
func createWindowsUpdateScript(currentExe, newExe string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("createWindowsUpdateScript called on non-Windows platform")
	}

	plan, err := prepareWindowsUpdate(currentExe, newExe)
	if err != nil {
		return err
	}

	if err := os.WriteFile(plan.scriptPath, []byte(windowsUpdateScript()), 0o600); err != nil {
		removeUpdateArtifact(plan.stagedExe)
		return fmt.Errorf("failed to create update script: %w", err)
	}

	// Validate script path before executing
	if err := validateFilePath(plan.scriptPath); err != nil {
		removeUpdateArtifact(plan.scriptPath)
		removeUpdateArtifact(plan.stagedExe)
		return fmt.Errorf("script path validation failed: %w", err)
	}

	// Execute the script in the background. It waits for this PID, so it must
	// not be waited for here; the process has to exit for the copy to start.
	// #nosec G204 - scriptPath is validated above for safety
	cmd := exec.Command("cmd", "/C", "start", "", "/B", plan.scriptPath)
	cmd.Env = append(os.Environ(), plan.environment()...)
	if err := cmd.Start(); err != nil {
		removeUpdateArtifact(plan.scriptPath)
		removeUpdateArtifact(plan.stagedExe)
		return fmt.Errorf("failed to execute update script: %w", err)
	}

	util.Info("Created update script. The application will be updated after exit.")
	util.Info("Please close the application to complete the update.")

	return fmt.Errorf("%w: helper %s", errUpdateDeferred, plan.scriptPath)
}
