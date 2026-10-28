//go:build windows

package tui

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/term"
)

var procGetConsoleProcessList = modkernel32.NewProc("GetConsoleProcessList")

// RelaunchInWindowsTerminal hands a standalone launch to Windows Terminal.
//
// A double-click, a Start menu shortcut or the installer's "Launch GoAnime"
// checkbox starts the executable in conhost, which has no font fallback and
// limited VT support; the TUI then degrades to ASCII. When that is the only
// process on the console, this starts `wt.exe` with the same executable,
// directory and arguments, and reports true so main returns without opening a
// second copy of the UI.
//
// The relaunch never happens from a shell or another ConPTY host: the user
// already picked a terminal there. It also never happens without a terminal
// (pipes, CI, scripts), and GOANIME_TERMINAL=conhost opts out.
func RelaunchInWindowsTerminal() bool {
	if !windowsTerminalLaunch(runtime.GOOS, os.Environ(), interactiveTerminal(), classicConsole(), soleConsoleProcess()) {
		return false
	}

	exe, err := os.Executable()
	if err != nil {
		return false
	}
	cwd, _ := os.Getwd()
	args := os.Args[1:]

	return windowsTerminalHandoff(windowsTerminalHandoffSteps{
		find:      windowsTerminalPath,
		elevated:  currentProcessElevated,
		winget:    wingetPath,
		installed: windowsTerminalInstalled,
		install:   installWindowsTerminal,
		await: func() (string, bool) {
			return waitForWindowsTerminal(10 * time.Second)
		},
		portable: installPortableWindowsTerminal,
		start: func(wt string) bool {
			return startWindowsTerminal(wt, exe, cwd, args)
		},
		notice: func(notice windowsTerminalNotice) {
			fmt.Fprintln(os.Stderr, windowsTerminalNoticeMessage(notice))
		},
	})
}

// startWindowsTerminal launches wt with the current executable and arguments.
// wt detaches once it has handed the command to the terminal, so the handle is
// released instead of waiting for a window that outlives this process.
func startWindowsTerminal(wt, exe, cwd string, args []string) bool {
	// #nosec G204 -- wt comes from the per-user WindowsApps alias directory or
	// the app's own install directory, and the executable and arguments are
	// the ones this process already received.
	cmd := exec.Command(wt, windowsTerminalArgs(exe, cwd, args)...)
	cmd.Env = append(os.Environ(), "GOANIME_WT_RELAUNCH=1")
	if err := cmd.Start(); err != nil {
		return false
	}
	_ = cmd.Process.Release()
	return true
}

// interactiveTerminal reports whether both ends of the TUI are a terminal. A
// console can be attached to one end only when a script redirects the other.
func interactiveTerminal() bool {
	if !HasTerminal() {
		return false
	}
	fd := os.Stdout.Fd()
	return fd <= math.MaxInt && term.IsTerminal(int(fd))
}

// soleConsoleProcess reports whether GoAnime is the only process attached to
// this console. A shell keeps the console alive, so a `goanime` typed in cmd
// or PowerShell is not standalone and must stay exactly where it was typed.
func soleConsoleProcess() bool {
	var pids [2]uint32
	// #nosec G103 -- a synchronous Win32 call fills this bounded PID array.
	n, _, _ := procGetConsoleProcessList.Call(uintptr(unsafe.Pointer(&pids[0])), uintptr(len(pids)))
	return n == 1
}

// windowsTerminalPath finds the wt.exe app execution alias or a cached private
// copy. The known alias directory comes first so a stale or planted wt.exe
// earlier on PATH cannot capture the relaunch; LookPath then covers a terminal
// that inherited PATH.
func windowsTerminalPath() (string, bool) {
	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		path := filepath.Join(local, "Microsoft", "WindowsApps", "wt.exe")
		// #nosec G703 -- LOCALAPPDATA is the user's profile root and the
		// joined path is the fixed Windows Terminal alias.
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, true
		}
	}
	if path, ok := portableWindowsTerminalPath(); ok {
		return path, true
	}
	if path, err := exec.LookPath("wt.exe"); err == nil {
		return path, true
	}
	return "", false
}

// windowsTerminalInstalled asks winget whether the package is installed. The
// alias can be disabled while the package stays installed, and installing it
// again on every standalone launch would only delay the fallback.
func windowsTerminalInstalled(winget string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// #nosec G204 -- winget comes from wingetPath and the arguments are
	// constants.
	out, err := exec.CommandContext(ctx, winget,
		"list",
		"--id", "Microsoft.WindowsTerminal",
		"--exact",
		"--accept-source-agreements",
		"--disable-interactivity",
	).Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(out)), "microsoft.windowsterminal")
}

// waitForWindowsTerminal gives a just-installed package time to register its
// app execution alias before falling back to conhost.
func waitForWindowsTerminal(timeout time.Duration) (string, bool) {
	deadline := time.Now().Add(timeout)
	for {
		if path, ok := windowsTerminalPath(); ok {
			return path, true
		}
		if time.Now().After(deadline) {
			return "", false
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// installWindowsTerminal installs Windows Terminal through winget, the package
// manager full Windows images ship. It streams winget's own output so a slow
// download does not look like a frozen console. Any failure returns false and
// the caller falls back to a portable copy.
func installWindowsTerminal(winget string) bool {
	fmt.Fprintln(os.Stderr, "GoAnime: installing Windows Terminal with winget...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	// #nosec G204 -- winget is resolved from PATH or the fixed WindowsApps
	// alias directory, and the arguments are constants.
	cmd := exec.CommandContext(ctx, winget,
		"install",
		"--id", "Microsoft.WindowsTerminal",
		"--exact",
		"--accept-package-agreements",
		"--accept-source-agreements",
		"--silent",
		"--disable-interactivity",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "GoAnime: winget could not install Windows Terminal: %v\n", err)
		return false
	}
	return true
}

// wingetPath finds the winget app execution alias, which lives in the same
// per-user WindowsApps directory as wt.exe.
func wingetPath() (string, bool) {
	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		path := filepath.Join(local, "Microsoft", "WindowsApps", "winget.exe")
		// #nosec G703 -- LOCALAPPDATA is the user's profile root and the
		// joined path is the fixed winget alias.
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, true
		}
	}
	if path, err := exec.LookPath("winget.exe"); err == nil {
		return path, true
	}
	return "", false
}

// currentProcessElevated reports whether this process runs with an
// administrator token.
func currentProcessElevated() bool {
	return windows.GetCurrentProcessToken().IsElevated()
}
