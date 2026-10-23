//go:build linux

package tui

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// GoAnime Glyphs is Noto Sans Mono CJK JP cut down to the characters GoAnime
// draws (fonts/build.py): 11 KB instead of the 16 MB a Japanese font takes.
// The font bytes live in japanese.go so the Windows installer shares them.
// Linux terminals use it for the glosses once fontconfig finds it.

// clockTicks is the kernel's USER_HZ, the unit of a process start time in
// /proc. Linux fixes it at 100 for user space on every architecture Go
// supports.
const clockTicks = 100

// glyphFontDir is the folder GoAnime installs its font into: one of its own
// inside the user's font directory, which fontconfig reads by default.
func glyphFontDir() (string, error) {
	if data := os.Getenv("XDG_DATA_HOME"); filepath.IsAbs(data) {
		return filepath.Join(data, "fonts", "goanime"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "fonts", "goanime"), nil
}

// japaneseFontProbe asks fontconfig the question about this machine's fonts.
func japaneseFontProbe() fontState { return fontconfigState() }

// installGlyphFont writes GoAnime Glyphs and its license into the user's
// font directory and says so for the debug log. The terminal that is open
// now loaded its fonts before, so the glosses appear once it is reopened.
func installGlyphFont() string {
	dir, err := glyphFontDir()
	if err != nil {
		return "No Japanese font, and no home folder to install GoAnime Glyphs into: " + err.Error()
	}
	if err := writeGlyphFont(dir); err != nil {
		return "No Japanese font, and installing GoAnime Glyphs failed: " + err.Error()
	}
	return fmt.Sprintf("No Japanese font: installed GoAnime Glyphs in %s. The Japanese labels appear once the terminal is reopened", dir)
}

// glyphFontState tells fonts that draw the glosses apart. Any font but
// GoAnime Glyphs is one the terminal already had. GoAnime Glyphs alone only
// counts once the terminal has started after it was installed: terminals
// load their fonts when they start, and before that it would draw boxes.
func glyphFontState(files []string) fontState {
	dir, err := glyphFontDir()
	if err != nil {
		return fontPresent
	}
	for _, file := range files {
		if filepath.Dir(file) != dir {
			return fontPresent
		}
	}
	info, err := os.Stat(filepath.Join(dir, glyphFontFile))
	if err != nil {
		return fontPresent
	}
	started, ok := terminalStart()
	return reloadedState(info.ModTime(), started, ok)
}

// reloadedState decides for GoAnime Glyphs alone: present once the terminal
// started after the install, pending before. An unknown start time counts
// as present, like every probe that cannot answer.
func reloadedState(installed, started time.Time, known bool) fontState {
	if !known || started.After(installed) {
		return fontPresent
	}
	return fontPending
}

// terminalStart is terminalStarted, replaced by tests that need a terminal
// opened before or after the install.
var terminalStart = terminalStarted

// terminalStarted reports when the terminal drawing this process started.
// It is the nearest ancestor outside this process's controlling terminal:
// the terminal emulator, or what stands in for it, such as the tmux server.
func terminalStarted() (time.Time, bool) {
	proc, err := os.OpenRoot("/proc")
	if err != nil {
		return time.Time{}, false
	}
	defer func() { _ = proc.Close() }()

	self, ok := readProcStat(proc, os.Getpid())
	if !ok || self.tty == 0 {
		return time.Time{}, false
	}
	pid := os.Getppid()
	for range 64 {
		if pid <= 1 {
			return time.Time{}, false
		}
		stat, ok := readProcStat(proc, pid)
		if !ok {
			return time.Time{}, false
		}
		if stat.tty != self.tty {
			boot, ok := bootTime(proc)
			if !ok {
				return time.Time{}, false
			}
			return boot.Add(ticksToDuration(stat.start)), true
		}
		pid = stat.ppid
	}
	return time.Time{}, false
}

// procStat holds the fields GoAnime reads from /proc/<pid>/stat.
type procStat struct {
	ppid, tty int
	start     int64 // clock ticks after boot
}

func readProcStat(proc *os.Root, pid int) (procStat, bool) {
	data, err := proc.ReadFile(strconv.Itoa(pid) + "/stat")
	if err != nil {
		return procStat{}, false
	}
	return parseProcStat(data)
}

// parseProcStat reads ppid (field 4), tty_nr (7) and starttime (22). The
// command name (2) is in parentheses and may hold spaces or parentheses, so
// the fields are counted from the last ")".
func parseProcStat(data []byte) (procStat, bool) {
	end := bytes.LastIndexByte(data, ')')
	if end < 0 {
		return procStat{}, false
	}
	fields := strings.Fields(string(data[end+1:]))
	if len(fields) < 20 {
		return procStat{}, false
	}
	ppid, err1 := strconv.Atoi(fields[1])
	tty, err2 := strconv.Atoi(fields[4])
	start, err3 := strconv.ParseInt(fields[19], 10, 64)
	if err1 != nil || err2 != nil || err3 != nil || start < 0 {
		return procStat{}, false
	}
	return procStat{ppid: ppid, tty: tty, start: start}, true
}

// ticksToDuration converts clock ticks to a duration. Whole seconds first:
// ticks times time.Second overflows after 2.9 years of uptime.
func ticksToDuration(ticks int64) time.Duration {
	return time.Duration(ticks/clockTicks)*time.Second + time.Duration(ticks%clockTicks)*(time.Second/clockTicks)
}

// bootTime reads the btime line of /proc/stat.
func bootTime(proc *os.Root) (time.Time, bool) {
	data, err := proc.ReadFile("stat")
	if err != nil {
		return time.Time{}, false
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		if value, ok := strings.CutPrefix(scanner.Text(), "btime "); ok {
			seconds, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
			if err != nil {
				return time.Time{}, false
			}
			return time.Unix(seconds, 0), true
		}
	}
	return time.Time{}, false
}
