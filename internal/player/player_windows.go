//go:build windows

package player

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/alvarorichard/Goanime/v2/internal/util"
)

func setProcessGroup(cmd *exec.Cmd) {
	// Windows has no process-group equivalent to Setpgid; this only logs
	// the resolved mpv command for troubleshooting. Route through the file
	// logger (not stdout) so it doesn't spill the full command line —
	// stream URL included — into the terminal during playback.
	util.Debugf("Setting process group for command: %s", cmd.String())
}

// findMPVPath searches for mpv executable in PATH and common installation directories on Windows.
// This function handles the case where mpv is installed via the GoAnime installer but
// the PATH environment variable hasn't been updated yet (common in Windows Sandbox).
func findMPVPath() (string, error) {
	l := mpvLocator{lookPath: exec.LookPath, stat: os.Stat}
	execPath, _ := os.Executable() // empty when unknown: the bundled locations are skipped
	candidates := mpvCandidatesWindows(execPath, os.Getenv)
	for i, c := range candidates {
		candidates[i] = filepath.Clean(c)
	}
	if path, ok := locateMPV(l, candidates, false); ok {
		util.Debugf("Found mpv at: %s", path)
		return path, nil
	}
	return "", fmt.Errorf("mpv not found in PATH or common installation directories. Please install mpv: https://mpv.io/installation/")
}
