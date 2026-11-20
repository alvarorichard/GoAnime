//go:build !windows

package player

import (
	"os"
	"os/exec"
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
	home, _ := os.UserHomeDir() // empty when unknown: per-user locations are skipped
	l := mpvLocator{lookPath: exec.LookPath, stat: os.Stat}
	if path, ok := locateMPV(l, mpvCandidatesUnix(runtime.GOOS, home), true); ok {
		return path, nil
	}
	return "", exec.ErrNotFound
}

// findAllMPVPaths lists every mpv found, the one on PATH first; see locateAllMPV.
func findAllMPVPaths() []string {
	home, _ := os.UserHomeDir() // empty when unknown: per-user locations are skipped
	l := mpvLocator{lookPath: exec.LookPath, stat: os.Stat}
	return locateAllMPV(l, mpvCandidatesUnix(runtime.GOOS, home), true, false)
}
