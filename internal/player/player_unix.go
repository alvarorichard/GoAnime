//go:build !windows

package player

import (
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"syscall"
)

func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// findMPVPath searches for mpv executable on Unix systems.
// It first tries the system PATH, then checks common installation locations
// including Homebrew paths on both macOS and Linux.
func findMPVPath() (string, error) {
	// First, try to find mpv in PATH
	if path, err := exec.LookPath("mpv"); err == nil {
		return path, nil
	}

	// Common paths where mpv might be installed
	var commonPaths []string

	// Get current user for home directory paths
	currentUser, _ := user.Current()
	homeDir := ""
	if currentUser != nil {
		homeDir = currentUser.HomeDir
	}

	if runtime.GOOS == "darwin" {
		// macOS: Homebrew Intel and Apple Silicon paths
		commonPaths = []string{
			"/opt/homebrew/bin/mpv",                    // Homebrew Apple Silicon
			"/usr/local/bin/mpv",                       // Homebrew Intel / MacPorts
			"/opt/local/bin/mpv",                       // MacPorts alternative
			"/Applications/mpv.app/Contents/MacOS/mpv", // mpv.app bundle
		}
	} else {
		// Linux: Check Homebrew and common package manager paths
		commonPaths = []string{
			"/home/linuxbrew/.linuxbrew/bin/mpv",      // Homebrew on Linux (system-wide)
			"/usr/bin/mpv",                            // Most Linux distros
			"/usr/local/bin/mpv",                      // Manual installation
			"/snap/bin/mpv",                           // Snap package
			"/var/lib/flatpak/exports/bin/io.mpv.Mpv", // Flatpak
		}

		// Add user-specific Homebrew path
		if homeDir != "" {
			commonPaths = append([]string{
				filepath.Join(homeDir, ".linuxbrew/bin/mpv"), // Homebrew per-user install
			}, commonPaths...)
		}
	}

	// Add user-specific Flatpak path for both platforms
	if homeDir != "" {
		commonPaths = append(commonPaths, filepath.Join(homeDir, ".local/share/flatpak/exports/bin/io.mpv.Mpv"))
	}

	// Check each path
	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			// Verify it's executable
			if info, err := os.Stat(path); err == nil && info.Mode()&0111 != 0 {
				return path, nil
			}
		}
	}

	return "", exec.ErrNotFound
}
