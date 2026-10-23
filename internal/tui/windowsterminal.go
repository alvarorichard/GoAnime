package tui

import (
	"slices"
	"strings"
)

// windowsTerminalLaunch reports whether GoAnime should hand this process to a
// Windows Terminal window.
//
// Only a standalone launch qualifies: a classic conhost window (not a ConPTY
// host such as Windows Terminal, VS Code or WezTerm), with GoAnime as the only
// process attached to that console (no shell holding it open), and an
// interactive stdin/stdout. A shell in the same console means the user already
// chose a terminal, and pipes or CI runs have nobody to show a window to.
func windowsTerminalLaunch(goos string, env []string, interactive, classic, soleProcess bool) bool {
	if goos != "windows" || !interactive || !classic || !soleProcess {
		return false
	}
	// The child carries WT_SESSION from Windows Terminal. The extra marker
	// protects the relaunch if a host ever fails to set it.
	if envValue(env, "WT_SESSION") != "" || envValue(env, "GOANIME_WT_RELAUNCH") != "" {
		return false
	}
	return !windowsTerminalDisabled(env)
}

// windowsTerminalOptOuts are the GOANIME_TERMINAL values that keep the classic
// console. The launch table and its meta-tests iterate this list, so a new
// alias cannot be added without a scenario.
var windowsTerminalOptOuts = []string{"conhost", "classic", "off", "0", "false", "no"}

// windowsTerminalDisabled reports whether GOANIME_TERMINAL opts out of the
// Windows Terminal relaunch, for users who prefer the classic console (or run
// a host whose ConPTY detection is wrong).
func windowsTerminalDisabled(env []string) bool {
	return slices.Contains(windowsTerminalOptOuts, strings.ToLower(envValue(env, "GOANIME_TERMINAL")))
}

// windowsTerminalNotice identifies a handoff message. Tests use the codes to
// pin the cascade without matching prose, and the real bootstrap maps them to
// the lines printed on stderr.
type windowsTerminalNotice string

const (
	terminalNoticeMissing       windowsTerminalNotice = "missing"
	terminalNoticeAliasDisabled windowsTerminalNotice = "alias-disabled"
	terminalNoticeInstallFailed windowsTerminalNotice = "install-failed"
	terminalNoticeStartFailed   windowsTerminalNotice = "start-failed"
)

func windowsTerminalNoticeMessage(notice windowsTerminalNotice) string {
	switch notice {
	case terminalNoticeMissing:
		return "GoAnime: Windows Terminal is missing; installing it for this user..."
	case terminalNoticeAliasDisabled:
		return "GoAnime: Windows Terminal is installed but its wt.exe alias is disabled; using a private copy."
	case terminalNoticeInstallFailed:
		return "GoAnime: Windows Terminal could not be installed; continuing in the classic console. Set GOANIME_TERMINAL=conhost to skip this check."
	case terminalNoticeStartFailed:
		return "GoAnime: could not start Windows Terminal; continuing in the classic console."
	default:
		return "GoAnime: continuing in the classic console."
	}
}

// windowsTerminalHandoffSteps are the side effects of the terminal handoff,
// separated so the fallback cascade can be tested without a console, winget,
// network or filesystem.
type windowsTerminalHandoffSteps struct {
	find      func() (string, bool)
	elevated  func() bool
	winget    func() (string, bool)
	installed func(string) bool
	install   func(string) bool
	await     func() (string, bool)
	portable  func() (string, bool)
	start     func(string) bool
	notice    func(windowsTerminalNotice)
}

// windowsTerminalHandoff finds a terminal, installing one when necessary, and
// starts it with the caller's executable. The cascade is ordered from least to
// most invasive:
//
//  1. an installed wt.exe (including a cached private copy);
//  2. the Store package through winget (which updates itself);
//  3. the bundled unpackaged build, then a download for other architectures.
//
// An elevated process skips winget, because the Store package would be
// installed for the administrator running it rather than for the user who
// started setup. The private copy is still safe there: it is unpacked into the
// running account's own profile, which is what lets the handoff work in
// Windows Sandbox even when UAC is disabled. A package with a disabled alias
// skips the pointless reinstall, and a failed start retries once with a
// private copy.
func windowsTerminalHandoff(steps windowsTerminalHandoffSteps) bool {
	wt, found := steps.find()
	if !found {
		elevated := steps.elevated()
		steps.notice(terminalNoticeMissing)

		if !elevated {
			if winget, ok := steps.winget(); ok {
				if steps.installed(winget) {
					steps.notice(terminalNoticeAliasDisabled)
				} else if steps.install(winget) {
					wt, found = steps.await()
				}
			}
		}

		if !found {
			wt, found = steps.portable()
		}
		if !found {
			steps.notice(terminalNoticeInstallFailed)
			return false
		}
	}

	if steps.start(wt) {
		return true
	}
	// An app execution alias can be stale even when the file exists. Retry
	// with a private copy, but only when it is a different binary, or the
	// same failure would just repeat.
	if portable, ok := steps.portable(); ok && portable != wt {
		if steps.start(portable) {
			return true
		}
	}
	steps.notice(terminalNoticeStartFailed)
	return false
}

// windowsTerminalArgs builds the wt command line:
//
//	wt -w -1 new-tab -d <cwd> -- <exe> <args...>
//
// `-w -1` forces a new window even when the user configured Windows Terminal
// to reuse the last one, and `--` keeps option-like GoAnime arguments
// (--debug, -h, --update) out of wt's own parser. Semicolons separate wt
// commands and are escaped, since `--` does not protect them.
func windowsTerminalArgs(exe, cwd string, args []string) []string {
	out := []string{"-w", "-1", "new-tab"}
	if cwd != "" {
		out = append(out, "-d", escapeWindowsTerminalArg(cwd))
	}
	out = append(out, "--", escapeWindowsTerminalArg(exe))
	for _, arg := range args {
		out = append(out, escapeWindowsTerminalArg(arg))
	}
	return out
}

// escapeWindowsTerminalArg protects a literal semicolon, which wt parses as a
// command separator even after `--`; a backslash is wt's own escape.
func escapeWindowsTerminalArg(arg string) string {
	return strings.ReplaceAll(arg, ";", `\;`)
}

// envValue returns key from an environment block, trimmed, or "" when unset.
func envValue(env []string, key string) string {
	prefix := key + "="
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			return strings.TrimSpace(item[len(prefix):])
		}
	}
	return ""
}
