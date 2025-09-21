package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/version"
	"github.com/charmbracelet/huh"
)

const (
	GitHubOwner = "alvarorichard"
	GitHubRepo  = "GoAnime"
	GitHubAPI   = "https://api.github.com/repos/" + GitHubOwner + "/" + GitHubRepo
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
	// Get latest release from GitHub API
	resp, err := http.Get(GitHubAPI + "/releases/latest")
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
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&release); err != nil {
		return nil, false, fmt.Errorf("failed to decode release data: %w", err)
	}

	// Compare versions
	currentVersion := version.Version
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
	assetURL, assetName, err := findAssetForPlatform(release)
	if err != nil {
		return err
	}

	util.Info("Downloading update:", assetName)

	// Download the asset
	tempFile, err := downloadAsset(assetURL, assetName)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer func() {
		if removeErr := os.Remove(tempFile); removeErr != nil {
			util.Debug("Failed to remove temp file:", removeErr)
		}
	}()

	// Get current executable path
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	// Create backup of current executable
	backupFile := currentExe + ".backup"
	if err := copyFile(currentExe, backupFile); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	defer func() {
		if removeErr := os.Remove(backupFile); removeErr != nil {
			util.Debug("Failed to remove backup file:", removeErr)
		}
	}()

	// Replace current executable
	if err := replaceExecutable(currentExe, tempFile); err != nil {
		// Try to restore backup if replacement fails
		if _, backupErr := os.Stat(backupFile); backupErr == nil {
			if restoreErr := copyFile(backupFile, currentExe); restoreErr != nil {
				util.Warn("Failed to restore backup:", restoreErr)
			} else {
				util.Info("Successfully restored backup after failed update")
			}
		}

		// Check if this is the Windows deferred update case
		if strings.Contains(err.Error(), "update script created") {
			// This is not actually an error - the update will complete after restart
			util.Info("Update will complete when you restart the application")
			return nil
		}

		return fmt.Errorf("failed to replace executable: %w", err)
	}

	util.Info("Update completed successfully! Please restart the application.")
	return nil
}

// PromptForUpdate shows an interactive prompt asking user if they want to update
func PromptForUpdate(release *GitHubRelease) (bool, error) {
	var shouldUpdate bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("🚀 Update Available").
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

	if err := form.Run(); err != nil {
		return false, fmt.Errorf("failed to show update prompt: %w", err)
	}

	return shouldUpdate, nil
}

// CheckAndPromptUpdate is a convenience function that checks for updates and prompts user
func CheckAndPromptUpdate() error {
	util.Info("Checking for updates...")

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
		util.Info(fmt.Sprintf("🚀 New version available: %s (current: %s)",
			release.TagName, version.Version))
		util.Info("Run with --update flag to update")
	}
}

// Helper functions

func isVersionNewer(latest, current string) (bool, error) {
	latestParts := strings.Split(latest, ".")
	currentParts := strings.Split(current, ".")

	// Pad shorter version with zeros
	maxLen := len(latestParts)
	if len(currentParts) > maxLen {
		maxLen = len(currentParts)
	}

	for len(latestParts) < maxLen {
		latestParts = append(latestParts, "0")
	}
	for len(currentParts) < maxLen {
		currentParts = append(currentParts, "0")
	}

	// Compare each part
	for i := 0; i < maxLen; i++ {
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

func findAssetForPlatform(release *GitHubRelease) (string, string, error) {
	return findAssetForPlatformWithInfo(release, GetCurrentPlatform())
}

func findAssetForPlatformWithInfo(release *GitHubRelease, platform PlatformInfo) (string, string, error) {
	// Map platform names to expected asset names
	var expectedNames []string
	switch platform.OS {
	case "windows":
		expectedNames = []string{
			fmt.Sprintf("goanime-windows-%s.exe", platform.Arch),
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

	// Find matching asset
	for _, asset := range release.Assets {
		for _, expectedName := range expectedNames {
			if strings.EqualFold(asset.Name, expectedName) {
				return asset.BrowserDownloadURL, asset.Name, nil
			}
		}
	}

	return "", "", fmt.Errorf("no compatible asset found for %s/%s", platform.OS, platform.Arch)
}

// validateGitHubURLWithTestFlag validates URLs with optional test mode support
func validateGitHubURLWithTestFlag(urlStr string, allowTestMode bool) error {
	parsedURL, err := url.Parse(urlStr)
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

	hostAllowed := false
	for _, allowedHost := range allowedHosts {
		if parsedURL.Host == allowedHost {
			hostAllowed = true
			break
		}
	}

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
		// #nosec G302 - 0755 is appropriate for executable files
		if err := os.Chmod(currentExe, 0755); err != nil {
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

// createWindowsUpdateScript creates a batch script to replace the executable
// after the current process exits. This is used as a fallback when the
// executable cannot be moved while running.
func createWindowsUpdateScript(currentExe, newExe string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("createWindowsUpdateScript called on non-Windows platform")
	}

	// Create a batch script that will:
	// 1. Wait for the current process to exit
	// 2. Replace the executable
	// 3. Clean up the temporary files
	// 4. Restart the application (optional)

	scriptPath := filepath.Join(filepath.Dir(currentExe), "update_goanime.bat")

	// Create the batch script content
	scriptContent := fmt.Sprintf(`@echo off
echo Updating GoAnime...
timeout /t 2 /nobreak > nul
:WAIT
tasklist /FI "PID eq %d" 2>NUL | find /I /N "%d">NUL
if "%%ERRORLEVEL%%"=="0" (
    timeout /t 1 /nobreak > nul
    goto WAIT
)
echo Replacing executable...
copy /Y "%s" "%s" > nul
if exist "%s" del "%s"
echo Update completed!
del "%%~f0"
`, os.Getpid(), os.Getpid(), newExe, currentExe, newExe, newExe)

	// Write the script to file
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0600); err != nil {
		return fmt.Errorf("failed to create update script: %w", err)
	}

	util.Info("Created update script. The application will be updated after exit.")
	util.Info("Please close the application to complete the update.")

	// Validate script path before executing
	if err := validateFilePath(scriptPath); err != nil {
		// Clean up script if path is invalid
		if removeErr := os.Remove(scriptPath); removeErr != nil {
			util.Debug("Failed to remove invalid script:", removeErr)
		}
		return fmt.Errorf("script path validation failed: %w", err)
	}

	// Execute the script in the background
	// Note: We don't wait for it to complete as it needs to run after this process exits
	// #nosec G204 - scriptPath is validated above for safety
	if err := exec.Command("cmd", "/C", "start", "/B", scriptPath).Start(); err != nil {
		// Clean up script if we can't execute it
		if removeErr := os.Remove(scriptPath); removeErr != nil {
			util.Debug("Failed to remove script after execution failure:", removeErr)
		}
		return fmt.Errorf("failed to execute update script: %w", err)
	}

	return fmt.Errorf("update script created - please restart the application to complete the update")
}
