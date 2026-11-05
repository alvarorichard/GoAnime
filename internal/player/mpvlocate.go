package player

import (
	"os"
	"path/filepath"
)

// mpvLocator is what the mpv search needs from the system. The default set
// is the real one; tests replace each piece.
type mpvLocator struct {
	lookPath func(file string) (string, error)
	stat     func(name string) (os.FileInfo, error)
}

// locateMPV returns the first usable mpv: the PATH lookup wins, then each
// candidate in order. A candidate counts when it exists and is not a
// directory; on Unix it must also have an executable bit.
func locateMPV(l mpvLocator, candidates []string, needExecBit bool) (string, bool) {
	if p, err := l.lookPath("mpv"); err == nil {
		return p, true
	}
	for _, c := range candidates {
		info, err := l.stat(c)
		if err != nil || info.IsDir() {
			continue
		}
		if needExecBit && info.Mode()&0o111 == 0 {
			continue
		}
		return c, true
	}
	return "", false
}

// mpvCandidatesUnix lists the places mpv is installed on macOS and Linux
// when it is not on PATH: Homebrew, MacPorts, the app bundle, distro
// packages, Snap and Flatpak. home may be empty when it is unknown.
func mpvCandidatesUnix(goos, home string) []string {
	var paths []string
	if goos == "darwin" {
		paths = []string{
			"/opt/homebrew/bin/mpv",                    // Homebrew Apple Silicon
			"/usr/local/bin/mpv",                       // Homebrew Intel / MacPorts
			"/opt/local/bin/mpv",                       // MacPorts alternative
			"/Applications/mpv.app/Contents/MacOS/mpv", // mpv.app bundle
		}
	} else {
		paths = []string{
			"/home/linuxbrew/.linuxbrew/bin/mpv",      // Homebrew on Linux (system-wide)
			"/usr/bin/mpv",                            // Most Linux distros
			"/usr/local/bin/mpv",                      // Manual installation
			"/snap/bin/mpv",                           // Snap package
			"/var/lib/flatpak/exports/bin/io.mpv.Mpv", // Flatpak
		}
		if home != "" {
			paths = append([]string{filepath.Join(home, ".linuxbrew", "bin", "mpv")}, paths...)
		}
	}
	if home != "" {
		paths = append(paths, filepath.Join(home, ".local", "share", "flatpak", "exports", "bin", "io.mpv.Mpv"))
	}
	return paths
}

// mpvCandidatesWindows lists the places mpv is installed on Windows when it
// is not on PATH (for instance right after the GoAnime installer, before the
// PATH change reaches the session). execPath is goanime.exe's own path;
// getenv reads the environment. Empty variables add nothing.
func mpvCandidatesWindows(execPath string, getenv func(string) string) []string {
	var paths []string
	if execPath != "" {
		dir := filepath.Dir(execPath)
		paths = append(paths,
			filepath.Join(dir, "bin", "mpv.exe"), // installed via the GoAnime installer
			filepath.Join(dir, "mpv.exe"),        // portable installation
		)
	}
	if pf := getenv("ProgramFiles"); pf != "" {
		paths = append(paths,
			filepath.Join(pf, "GoAnime", "bin", "mpv.exe"),
			filepath.Join(pf, "mpv", "mpv.exe"),
			filepath.Join(pf, "mpv.net", "mpv.exe"),
		)
	}
	if pf86 := getenv("ProgramFiles(x86)"); pf86 != "" {
		paths = append(paths,
			filepath.Join(pf86, "GoAnime", "bin", "mpv.exe"),
			filepath.Join(pf86, "mpv", "mpv.exe"),
		)
	}
	if la := getenv("LOCALAPPDATA"); la != "" {
		paths = append(paths, filepath.Join(la, "Programs", "GoAnime", "bin", "mpv.exe"))
	}
	if up := getenv("USERPROFILE"); up != "" {
		paths = append(paths,
			filepath.Join(up, "scoop", "apps", "mpv", "current", "mpv.exe"),
			filepath.Join(up, "scoop", "shims", "mpv.exe"),
		)
	}
	return paths
}
