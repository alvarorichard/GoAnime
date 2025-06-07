package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
		// Try to restore backup
		if _, backupErr := os.Stat(backupFile); backupErr == nil {
			if restoreErr := copyFile(backupFile, currentExe); restoreErr != nil {
				util.Warn("Failed to restore backup:", restoreErr)
			}
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
			"goanime-darwin",
			"goanime-macos",
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

func downloadAsset(url, filename string) (string, error) {
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

	// Create temporary file
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, "goanime-update-"+filename)

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
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := sourceFile.Close(); closeErr != nil {
			util.Debug("Failed to close source file:", closeErr)
		}
	}()

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
	// On Windows, we might need to rename the current executable first
	if runtime.GOOS == "windows" {
		tempName := currentExe + ".old"
		if err := os.Rename(currentExe, tempName); err != nil {
			return err
		}
		defer func() {
			if removeErr := os.Remove(tempName); removeErr != nil {
				util.Debug("Failed to remove old executable:", removeErr)
			}
		}()
	}

	// Copy new executable to current location
	if err := copyFile(newExe, currentExe); err != nil {
		return err
	}

	// Make sure it's executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(currentExe, 0755); err != nil {
			return err
		}
	}

	return nil
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
