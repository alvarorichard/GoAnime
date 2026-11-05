package player

import "strings"

// mpvConfigEnv selects whether mpv loads the user's own configuration
// (~/.config/mpv, %APPDATA%\mpv): mpv.conf, input.conf, scripts, shaders and
// script-opts. Unset means yes. "off", "0", "false", "no" or "none" start mpv
// with --no-config instead, the isolated behaviour GoAnime had before.
const mpvConfigEnv = "GOANIME_MPV_CONFIG"

// mpvUserConfigEnabled reports whether GoAnime should let mpv load the user's
// configuration. getenv is a parameter so the decision is testable without
// touching the process environment.
func mpvUserConfigEnabled(getenv func(string) string) bool {
	switch strings.ToLower(strings.TrimSpace(getenv(mpvConfigEnv))) {
	case "off", "0", "false", "no", "none":
		return false
	default:
		return true
	}
}

// userConfigGuardArgs are the mpv options GoAnime re-asserts on the command
// line whenever the user's config is loaded. Command-line options beat
// mpv.conf, so these win over whatever the user wrote. Each one protects
// something GoAnime depends on:
//   - keep-open, idle, loop-file, loop-playlist: mpv must exit when the
//     episode ends, or the player menu, tracking and "next episode" never run.
//   - resume-playback: GoAnime resumes with its own saved position; mpv's
//     watch_later file must not seek a second time.
func userConfigGuardArgs() []string {
	return []string{
		"--keep-open=no",
		"--idle=no",
		"--loop-file=no",
		"--loop-playlist=no",
		"--resume-playback=no",
	}
}

// mpvConfigMarkers are the entries of an mpv config directory that mean the
// user configured something. Folders mpv fills on its own (watch_later, cache)
// are not listed: they exist on a stock install.
var mpvConfigMarkers = []string{
	"mpv.conf", "input.conf", "encoding-profiles.conf",
	"scripts", "script-opts", "script-modules", "shaders", "fonts", "lua-settings",
}

// mpvConfigDirs lists, in mpv's own lookup order, the directories where a user
// config can live. getenv and home are parameters so the list is testable for
// any platform; home is "" when unknown.
func mpvConfigDirs(goos string, getenv func(string) string, home string) []string {
	var dirs []string
	add := func(parts ...string) {
		dirs = append(dirs, strings.Join(parts, sepFor(goos)))
	}
	if v := getenv("MPV_HOME"); v != "" {
		dirs = append(dirs, v)
	}
	if goos == "windows" {
		if v := getenv("APPDATA"); v != "" {
			add(v, "mpv")
		}
		return dirs
	}
	if v := getenv("XDG_CONFIG_HOME"); v != "" {
		add(v, "mpv")
	} else if home != "" {
		add(home, ".config", "mpv")
	}
	if home != "" {
		add(home, ".mpv") // legacy location mpv still reads
	}
	return dirs
}

func sepFor(goos string) string {
	if goos == "windows" {
		return `\`
	}
	return "/"
}

// mpvHasUserConfig reports whether any config directory in dirs holds one of
// mpvConfigMarkers. exists answers for a full path, so tests need no disk.
func mpvHasUserConfig(dirs []string, goos string, exists func(string) bool) bool {
	for _, d := range dirs {
		for _, m := range mpvConfigMarkers {
			if exists(d + sepFor(goos) + m) {
				return true
			}
		}
	}
	return false
}

// shouldLoadMPVUserConfig decides whether playback starts with the user's mpv
// config. It needs both the opt-out to be unset and a config to exist: without
// one there is nothing to respect, and GoAnime keeps its isolated, tuned
// profile (--no-config, --profile=fast). present is only called when the
// opt-out leaves the question open.
func shouldLoadMPVUserConfig(getenv func(string) string, present func() bool) bool {
	return mpvUserConfigEnabled(getenv) && present()
}
