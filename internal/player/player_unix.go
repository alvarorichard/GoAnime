//go:build !windows

package player

import (
	"os/exec"
	"syscall"
)

func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// findMPVPath searches for mpv executable on Unix systems.
// On Unix, mpv is typically installed system-wide and available in PATH.
func findMPVPath() (string, error) {
	return exec.LookPath("mpv")
}
