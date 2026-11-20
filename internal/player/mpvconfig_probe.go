package player

import (
	"os"
	"runtime"
)

// mpvUserConfigPresent probes the real machine for a user mpv config.
func mpvUserConfigPresent() bool {
	home, _ := os.UserHomeDir()
	dirs := mpvConfigDirs(runtime.GOOS, os.Getenv, home)
	return mpvHasUserConfig(dirs, runtime.GOOS, func(p string) bool {
		_, err := os.Stat(p)
		return err == nil
	})
}
