package tui

import "strings"

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

// windowsTerminalDisabled reports whether GOANIME_TERMINAL opts out of the
// Windows Terminal relaunch, for users who prefer the classic console (or run
// a host whose ConPTY detection is wrong).
func windowsTerminalDisabled(env []string) bool {
	switch strings.ToLower(envValue(env, "GOANIME_TERMINAL")) {
	case "conhost", "classic", "off", "0", "false", "no":
		return true
	}
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
