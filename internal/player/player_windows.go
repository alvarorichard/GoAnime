//go:build windows

package player

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/alvarorichard/Goanime/internal/util"
)

func setProcessGroup(cmd *exec.Cmd) {
	// mensage debug
	if util.IsDebug {
		fmt.Println("Setting process group for command:", cmd.String())
	}

}

// findMPVPath searches for mpv executable in PATH and common installation directories on Windows.
// This function handles the case where mpv is installed via the GoAnime installer but
// the PATH environment variable hasn't been updated yet (common in Windows Sandbox).
func findMPVPath() (string, error) {
	// First, try standard PATH lookup
	if mpvPath, err := exec.LookPath("mpv"); err == nil {
		return mpvPath, nil
	}

	// List of common MPV installation paths on Windows
	possiblePaths := []string{}

	// Get the directory where goanime.exe is located
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		// Check for bundled mpv in the same directory as goanime
		possiblePaths = append(possiblePaths,
			filepath.Join(execDir, "bin", "mpv.exe"),     // Installed via GoAnime installer
			filepath.Join(execDir, "mpv.exe"),            // Portable installation
		)
	}

	// Check Program Files directories
	programFiles := os.Getenv("ProgramFiles")
	programFilesX86 := os.Getenv("ProgramFiles(x86)")
	localAppData := os.Getenv("LOCALAPPDATA")

	if programFiles != "" {
		possiblePaths = append(possiblePaths,
			filepath.Join(programFiles, "GoAnime", "bin", "mpv.exe"),
			filepath.Join(programFiles, "mpv", "mpv.exe"),
			filepath.Join(programFiles, "mpv.net", "mpv.exe"),
		)
	}
	if programFilesX86 != "" {
		possiblePaths = append(possiblePaths,
			filepath.Join(programFilesX86, "GoAnime", "bin", "mpv.exe"),
			filepath.Join(programFilesX86, "mpv", "mpv.exe"),
		)
	}
	if localAppData != "" {
		possiblePaths = append(possiblePaths,
			filepath.Join(localAppData, "Programs", "GoAnime", "bin", "mpv.exe"),
		)
	}

	// Also check scoop and chocolatey installations
	userProfile := os.Getenv("USERPROFILE")
	if userProfile != "" {
		possiblePaths = append(possiblePaths,
			filepath.Join(userProfile, "scoop", "apps", "mpv", "current", "mpv.exe"),
			filepath.Join(userProfile, "scoop", "shims", "mpv.exe"),
		)
	}

	// Check each possible path
	for _, path := range possiblePaths {
		if util.IsDebug {
			fmt.Printf("[DEBUG] Checking for mpv at: %s\n", path)
		}
		if _, err := os.Stat(path); err == nil {
			if util.IsDebug {
				fmt.Printf("[DEBUG] Found mpv at: %s\n", path)
			}
			return path, nil
		}
	}

	return "", fmt.Errorf("mpv not found in PATH or common installation directories. Please install mpv: https://mpv.io/installation/")
}
