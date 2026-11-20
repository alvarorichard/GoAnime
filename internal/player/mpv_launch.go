package player

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// mpvEarlyExitError is an mpv that died before its IPC socket appeared. code is
// its exit code; mpv uses 1 for "could not initialise" (a rejected option, a
// video output that will not open).
type mpvEarlyExitError struct {
	code int
	err  error
}

func (e *mpvEarlyExitError) Error() string { return e.err.Error() }
func (e *mpvEarlyExitError) Unwrap() error { return e.err }

// mpvStartError is an mpv that could not be started at all: the file is not
// there, or is a stub the system refuses to run.
type mpvStartError struct{ err error }

func (e *mpvStartError) Error() string { return e.err.Error() }
func (e *mpvStartError) Unwrap() error { return e.err }

// failedToLaunch reports whether err means this mpv did not come up, as
// opposed to coming up and then not answering in time. Another mpv may do
// better in the first case; in the second, waiting for it as well would only
// double a long wait.
func failedToLaunch(err error) bool {
	var early *mpvEarlyExitError
	var start *mpvStartError
	return errors.As(err, &early) || errors.As(err, &start)
}

// tuningMPVArgs are the options GoAnime passes to make playback smoother:
// cache sizes, the video output and hardware decoding, the upscaling shaders.
// None is needed to play; an mpv build that rejects one still plays without it.
var tuningMPVArgs = []string{
	"--cache=", "--demuxer-max-bytes=", "--demuxer-readahead-secs=",
	"--audio-display=", "--hwdec=", "--vo=", "--gpu-context=", "--gpu-api=",
	"--profile=", "--video-latency-hacks=",
	"--glsl-shader=", "--glsl-shaders=", "--gpu-shader-cache-dir=",
}

// essentialMPVArgs splits args into the ones mpv needs to play what it was
// asked (headers, subtitles, audio, title, start position, the ytdl options)
// and the tuning options it only plays better with.
func essentialMPVArgs(args []string) (kept, dropped []string) {
	for _, a := range args {
		tuning := false
		for _, prefix := range tuningMPVArgs {
			if strings.HasPrefix(a, prefix) {
				tuning = true
				break
			}
		}
		if tuning {
			dropped = append(dropped, a)
		} else {
			kept = append(kept, a)
		}
	}
	return kept, dropped
}

// mpvChoice remembers what worked on this machine, so a launch that needed a
// fallback pays for it once and not on every episode.
var mpvChoice struct {
	sync.Mutex
	working string          // the mpv that last came up
	minimal map[string]bool // mpv builds that only start without the tuning options
}

func resetMPVChoice() {
	mpvChoice.Lock()
	defer mpvChoice.Unlock()
	mpvChoice.working, mpvChoice.minimal = "", nil
}

func rememberWorkingMPV(path string) {
	mpvChoice.Lock()
	defer mpvChoice.Unlock()
	mpvChoice.working = path
}

func rememberMinimalMPV(path string) {
	mpvChoice.Lock()
	defer mpvChoice.Unlock()
	if mpvChoice.minimal == nil {
		mpvChoice.minimal = map[string]bool{}
	}
	mpvChoice.minimal[path] = true
}

func mpvNeedsMinimal(path string) bool {
	mpvChoice.Lock()
	defer mpvChoice.Unlock()
	return mpvChoice.minimal[path]
}

// orderedMPV puts the mpv that last came up first and keeps the rest in the
// order they were found.
func orderedMPV(paths []string) []string {
	mpvChoice.Lock()
	working := mpvChoice.working
	mpvChoice.Unlock()
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if p == working {
			out = append(out, p)
		}
	}
	for _, p := range paths {
		if p != working {
			out = append(out, p)
		}
	}
	return out
}

// mpvSearch is every mpv found, looked up once.
var mpvSearch struct {
	once  sync.Once
	paths []string
}

// mpvPaths returns the mpv installations to try, the one the system finds
// first leading.
func mpvPaths() ([]string, error) {
	// Verify MPV is installed (cached after first lookup to avoid repeated filesystem searches)
	cachedMPVPathOnce.Do(func() {
		cachedMPVPath, cachedMPVPathErr = findMPVPath()
	})
	if cachedMPVPathErr != nil {
		return nil, fmt.Errorf("mpv not found: %w\nPlease install mpv: https://mpv.io/installation/", cachedMPVPathErr)
	}
	mpvSearch.once.Do(func() { mpvSearch.paths = findAllMPVPaths() })
	return leadWith(cachedMPVPath, mpvSearch.paths), nil
}

// leadWith puts first at the head of paths, once. The two lookups can disagree
// (the PATH changed between them, or only the first one saw an mpv); the one
// that was found and cached stays first.
func leadWith(first string, paths []string) []string {
	out := make([]string, 0, len(paths)+1)
	out = append(out, first)
	for _, p := range paths {
		if p != first {
			out = append(out, p)
		}
	}
	return out
}

// mpvAttempt is one mpv that did not come up, and why.
type mpvAttempt struct {
	path string
	err  error
}

// startVideoAcross opens link in the first mpv of paths that comes up.
//
// The mpv on PATH is not always one that can run: a Windows Store app alias
// (WindowsApps\mpv.exe) is on PATH ahead of the mpv the GoAnime installer
// ships, and exits at once. Every installation is tried in turn, and the one
// that works is remembered. An mpv that comes up and then does not answer in
// time ends the search: another would not be quicker.
//
// Only when none comes up is an mpv that died at initialisation tried again
// with just the options playback needs; see startMinimal.
func startVideoAcross(paths []string, link string, args []string) (string, error) {
	// Sanitize media target (URL or local file path)
	safeLink, err := sanitizeMediaTarget(link)
	if err != nil {
		return "", fmt.Errorf("invalid media target: %w", err)
	}
	var failed []mpvAttempt
	for _, path := range orderedMPV(paths) {
		useArgs := args
		if mpvNeedsMinimal(path) {
			useArgs, _ = essentialMPVArgs(args)
		}
		socket, err := launchMPV(path, safeLink, useArgs)
		if err == nil {
			rememberWorkingMPV(path)
			if len(failed) > 0 {
				util.Warnf("mpv at %s would not start; playing with %s instead.", failed[0].path, path)
			}
			return socket, nil
		}
		if !failedToLaunch(err) {
			return "", err
		}
		failed = append(failed, mpvAttempt{path: path, err: err})
	}
	if socket, ok := startMinimal(failed, safeLink, args); ok {
		return socket, nil
	}
	return "", describeMPVFailures(failed)
}

// startMinimal tries again, with only the options playback needs, each mpv
// that died at initialisation (exit code 1): a build that rejects one of the
// tuning options still plays without it. The first to come up is remembered
// as needing that, and what was left out is said. It notes on failed what the
// retry said, for the error the user gets.
func startMinimal(failed []mpvAttempt, safeLink string, args []string) (string, bool) {
	kept, dropped := essentialMPVArgs(args)
	if len(dropped) == 0 {
		return "", false
	}
	for i := range failed {
		var early *mpvEarlyExitError
		if !errors.As(failed[i].err, &early) || early.code != 1 || mpvNeedsMinimal(failed[i].path) {
			continue
		}
		path := failed[i].path
		util.Debugf("mpv %s exited with code 1; retrying without GoAnime's tuning options %q", path, dropped)
		socket, err := launchMPV(path, safeLink, kept)
		if err == nil {
			rememberMinimalMPV(path)
			rememberWorkingMPV(path)
			util.Warnf("mpv would not start with GoAnime's video options, so it was started with its own defaults (left out: %s).", strings.Join(dropped, " "))
			return socket, true
		}
		failed[i].err = fmt.Errorf("%w\n\nTrying again without GoAnime's video options (%s) did not help: %v",
			failed[i].err, strings.Join(dropped, " "), firstLine(err.Error()))
	}
	return "", false
}

// describeMPVFailures reports why no mpv came up: the first one's own error,
// then a line for each other that was tried. An mpv that is a Microsoft Store
// alias gets the one thing that can be done about it.
func describeMPVFailures(failed []mpvAttempt) error {
	first := failed[0]
	var extra strings.Builder
	for _, a := range failed[1:] {
		fmt.Fprintf(&extra, "\nAlso tried %s: %s", a.path, firstLine(a.err.Error()))
	}
	if isStoreAlias(first.path) {
		fmt.Fprintf(&extra, "\nHint: %s is a Microsoft Store app alias. If that is the one failing, turn it off under Settings > Apps > Advanced app settings > App execution aliases, or install mpv from https://mpv.io/installation/ (the GoAnime installer includes one).", first.path)
	}
	if extra.Len() == 0 {
		return first.err
	}
	return fmt.Errorf("%w%s", first.err, extra.String())
}

// isStoreAlias reports whether path is under the folder Windows keeps app
// execution aliases in.
func isStoreAlias(path string) bool {
	return strings.Contains(strings.ToLower(strings.ReplaceAll(path, `\`, "/")), "/microsoft/windowsapps/")
}

// firstLine is s up to its first line break.
func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return strings.TrimSpace(line)
}
