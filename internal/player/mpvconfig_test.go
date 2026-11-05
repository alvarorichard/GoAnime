package player

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeEnv is a getenv mock: it serves a fixed map and records every key asked
// for, so tests can prove the code reads the documented variable and nothing
// else.
type fakeEnv struct {
	vars  map[string]string
	asked []string
}

func (f *fakeEnv) get(key string) string {
	f.asked = append(f.asked, key)
	return f.vars[key]
}

// userConfigCases is the behaviour table shared by the unit test and the
// metatest, so a mutant is judged against exactly what the real function is.
var userConfigCases = []struct {
	name  string
	value string
	want  bool
}{
	{"unset loads the user config", "", true},
	{"on", "on", true},
	{"1", "1", true},
	{"true", "true", true},
	{"unknown value keeps the default", "banana", true},
	{"off isolates", "off", false},
	{"OFF is case-insensitive", "OFF", false},
	{"zero isolates", "0", false},
	{"false isolates", "false", false},
	{"no isolates", "no", false},
	{"none isolates", "none", false},
	{"surrounding space is ignored", "  off\t", false},
	{"whitespace only is unset", "   ", true},
}

// userConfigFailures runs impl over userConfigCases and returns one line per
// disagreement. Empty means impl matches the contract.
func userConfigFailures(impl func(func(string) string) bool) []string {
	var failures []string
	for _, tc := range userConfigCases {
		env := &fakeEnv{vars: map[string]string{mpvConfigEnv: tc.value}}
		if got := impl(env.get); got != tc.want {
			failures = append(failures, fmt.Sprintf("%s: got %v want %v", tc.name, got, tc.want))
		}
	}
	return failures
}

func TestMPVUserConfigEnabled(t *testing.T) {
	t.Parallel()
	for _, tc := range userConfigCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			env := &fakeEnv{vars: map[string]string{mpvConfigEnv: tc.value}}
			assert.Equal(t, tc.want, mpvUserConfigEnabled(env.get))
			assert.Equal(t, []string{mpvConfigEnv}, env.asked, "reads only the documented variable, once")
		})
	}
}

func TestMPVConfigEnvName(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "GOANIME_MPV_CONFIG", mpvConfigEnv, "documented in the README; renaming breaks users")
}

// TestMPVUserConfigEnabledMeta is the metatest: it feeds the contract table
// deliberately broken implementations and requires every one to be caught. If a
// mutant passes, the table above is too weak to protect the real function.
func TestMPVUserConfigEnabledMeta(t *testing.T) {
	t.Parallel()

	require.Empty(t, userConfigFailures(mpvUserConfigEnabled), "the real function satisfies the contract")

	offWords := map[string]bool{"off": true, "0": true, "false": true, "no": true, "none": true}
	mutants := map[string]func(func(string) string) bool{
		"always on":  func(func(string) string) bool { return true },
		"always off": func(func(string) string) bool { return false },
		"inverted":   func(g func(string) string) bool { return !mpvUserConfigEnabled(g) },
		"case-sensitive": func(g func(string) string) bool {
			return !offWords[g(mpvConfigEnv)]
		},
		"no trim": func(g func(string) string) bool {
			return !offWords[strings.ToLower(g(mpvConfigEnv))]
		},
		"wrong variable": func(g func(string) string) bool {
			return mpvUserConfigEnabled(func(string) string { return g("GOANIME_MPV") })
		},
		"off word dropped": func(g func(string) string) bool {
			v := strings.ToLower(strings.TrimSpace(g(mpvConfigEnv)))
			return !(v == "off" || v == "0" || v == "false" || v == "no")
		},
		"unknown value isolates": func(g func(string) string) bool {
			v := strings.ToLower(strings.TrimSpace(g(mpvConfigEnv)))
			return v == "" || v == "on" || v == "1" || v == "true"
		},
	}
	for name, mutant := range mutants {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.NotEmpty(t, userConfigFailures(mutant), "mutant %q survived: the contract table does not detect it", name)
		})
	}
}

func TestUserConfigGuardArgs(t *testing.T) {
	t.Parallel()
	got := userConfigGuardArgs()
	assert.ElementsMatch(t, []string{
		"--keep-open=no", "--idle=no", "--loop-file=no", "--loop-playlist=no", "--resume-playback=no",
	}, got)

	// Every guard must survive the argument allowlist, or StartVideo drops it.
	assert.Equal(t, got, filterMPVArgs(got), "guards pass filterMPVArgs unchanged")

	// Fresh slice each call: a caller appending to it must not corrupt the next.
	got[0] = "--mutated"
	assert.NotEqual(t, "--mutated", userConfigGuardArgs()[0])
}

// TestUserConfigGuardArgsMeta mutates the guard list and proves the contract check notices.
func TestUserConfigGuardArgsMeta(t *testing.T) {
	t.Parallel()
	check := func(args []string) bool {
		want := map[string]bool{
			"--keep-open=no": true, "--idle=no": true, "--loop-file=no": true,
			"--loop-playlist=no": true, "--resume-playback=no": true,
		}
		if len(args) != len(want) {
			return false
		}
		seen := map[string]bool{}
		for _, a := range args {
			if !want[a] || seen[a] {
				return false
			}
			seen[a] = true
		}
		return true
	}
	require.True(t, check(userConfigGuardArgs()))

	real := userConfigGuardArgs()
	mutants := map[string][]string{
		"empty":           nil,
		"missing one":     real[1:],
		"keep-open yes":   append([]string{"--keep-open=yes"}, real[1:]...),
		"extra flag":      append(append([]string{}, real...), "--pause=yes"),
		"duplicate guard": append(append([]string{}, real[1:]...), real[1]),
	}
	for name, m := range mutants {
		assert.False(t, check(m), "mutant %q survived", name)
	}
}

func envOf(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestMPVConfigDirs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		goos string
		env  map[string]string
		home string
		want []string
	}{
		{"linux default", "linux", nil, "/home/u", []string{"/home/u/.config/mpv", "/home/u/.mpv"}},
		{"darwin default", "darwin", nil, "/Users/u", []string{"/Users/u/.config/mpv", "/Users/u/.mpv"}},
		{"XDG_CONFIG_HOME replaces ~/.config", "linux", map[string]string{"XDG_CONFIG_HOME": "/x"}, "/home/u", []string{"/x/mpv", "/home/u/.mpv"}},
		{"MPV_HOME comes first", "linux", map[string]string{"MPV_HOME": "/m"}, "/home/u", []string{"/m", "/home/u/.config/mpv", "/home/u/.mpv"}},
		{"windows uses APPDATA", "windows", map[string]string{"APPDATA": `C:\Users\u\AppData\Roaming`}, `C:\Users\u`, []string{`C:\Users\u\AppData\Roaming\mpv`}},
		{"windows without APPDATA has none", "windows", nil, `C:\Users\u`, nil},
		{"unknown home yields nothing", "linux", nil, "", nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, mpvConfigDirs(tc.goos, envOf(tc.env), tc.home))
		})
	}
}

// configDirsFailures runs impl over the cases above that matter most and
// returns the disagreements; the metatest feeds it broken versions.
func configDirsFailures(impl func(goos string, getenv func(string) string, home string) []string) []string {
	var f []string
	check := func(name string, got, want []string) {
		if fmt.Sprint(got) != fmt.Sprint(want) {
			f = append(f, fmt.Sprintf("%s: got %v want %v", name, got, want))
		}
	}
	check("linux", impl("linux", envOf(nil), "/h"), []string{"/h/.config/mpv", "/h/.mpv"})
	check("xdg", impl("linux", envOf(map[string]string{"XDG_CONFIG_HOME": "/x"}), "/h"), []string{"/x/mpv", "/h/.mpv"})
	check("mpv_home", impl("linux", envOf(map[string]string{"MPV_HOME": "/m"}), "/h"), []string{"/m", "/h/.config/mpv", "/h/.mpv"})
	check("windows", impl("windows", envOf(map[string]string{"APPDATA": `C:\A`}), `C:\u`), []string{`C:\A\mpv`})
	check("no home", impl("linux", envOf(nil), ""), nil)
	return f
}

func TestMPVConfigDirsMeta(t *testing.T) {
	t.Parallel()
	require.Empty(t, configDirsFailures(mpvConfigDirs))

	mutants := map[string]func(string, func(string) string, string) []string{
		"ignores MPV_HOME": func(g string, e func(string) string, h string) []string {
			return mpvConfigDirs(g, func(k string) string {
				if k == "MPV_HOME" {
					return ""
				}
				return e(k)
			}, h)
		},
		"ignores XDG_CONFIG_HOME": func(g string, e func(string) string, h string) []string {
			return mpvConfigDirs(g, func(k string) string {
				if k == "XDG_CONFIG_HOME" {
					return ""
				}
				return e(k)
			}, h)
		},
		"drops legacy dir": func(g string, e func(string) string, h string) []string {
			d := mpvConfigDirs(g, e, h)
			if g != "windows" && len(d) > 0 {
				return d[:len(d)-1]
			}
			return d
		},
		"unix separators on windows": func(g string, e func(string) string, h string) []string {
			return mpvConfigDirs("linux", e, h)
		},
		"empty": func(string, func(string) string, string) []string { return nil },
	}
	for name, m := range mutants {
		assert.NotEmpty(t, configDirsFailures(m), "mutant %q survived", name)
	}
}

func existsIn(paths ...string) func(string) bool {
	set := map[string]bool{}
	for _, p := range paths {
		set[p] = true
	}
	return func(p string) bool { return set[p] }
}

func TestMPVHasUserConfig(t *testing.T) {
	t.Parallel()
	dirs := []string{"/a/mpv", "/b"}
	tests := []struct {
		name   string
		exists func(string) bool
		want   bool
	}{
		{"nothing on disk", existsIn(), false},
		{"mpv.conf in first dir", existsIn("/a/mpv/mpv.conf"), true},
		{"scripts in second dir", existsIn("/b/scripts"), true},
		{"shaders count", existsIn("/a/mpv/shaders"), true},
		{"stock watch_later is not a config", existsIn("/a/mpv/watch_later", "/a/mpv/cache"), false},
		{"an empty config dir is not a config", existsIn("/a/mpv"), false},
		{"marker in an unlisted dir is ignored", existsIn("/c/mpv.conf"), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, mpvHasUserConfig(dirs, "linux", tc.exists))
		})
	}
	assert.False(t, mpvHasUserConfig(nil, "linux", existsIn("/a/mpv/mpv.conf")), "no dirs, no config")
	assert.True(t, mpvHasUserConfig([]string{`C:\A\mpv`}, "windows", existsIn(`C:\A\mpv\input.conf`)), "windows separator")
}

func hasConfigFailures(impl func([]string, string, func(string) bool) bool) []string {
	var f []string
	chk := func(name string, got, want bool) {
		if got != want {
			f = append(f, fmt.Sprintf("%s: got %v want %v", name, got, want))
		}
	}
	chk("empty", impl([]string{"/a"}, "linux", existsIn()), false)
	chk("mpv.conf", impl([]string{"/a"}, "linux", existsIn("/a/mpv.conf")), true)
	chk("input.conf", impl([]string{"/a"}, "linux", existsIn("/a/input.conf")), true)
	chk("scripts", impl([]string{"/a"}, "linux", existsIn("/a/scripts")), true)
	chk("last dir", impl([]string{"/a", "/b"}, "linux", existsIn("/b/mpv.conf")), true)
	chk("empty dir that exists", impl([]string{"/a"}, "linux", existsIn("/a")), false)
	chk("watch_later", impl([]string{"/a"}, "linux", existsIn("/a/watch_later")), false)
	chk("no dirs", impl(nil, "linux", existsIn("/a/mpv.conf")), false)
	chk("windows", impl([]string{`C:\a`}, "windows", existsIn(`C:\a\mpv.conf`)), true)
	return f
}

func TestMPVHasUserConfigMeta(t *testing.T) {
	t.Parallel()
	require.Empty(t, hasConfigFailures(mpvHasUserConfig))

	mutants := map[string]func([]string, string, func(string) bool) bool{
		"always true":  func([]string, string, func(string) bool) bool { return true },
		"always false": func([]string, string, func(string) bool) bool { return false },
		"first dir only": func(d []string, g string, e func(string) bool) bool {
			if len(d) > 1 {
				d = d[:1]
			}
			return mpvHasUserConfig(d, g, e)
		},
		"counts the directory itself": func(d []string, g string, e func(string) bool) bool {
			for _, x := range d {
				if e(x) {
					return true
				}
			}
			return mpvHasUserConfig(d, g, e)
		},
		"counts watch_later": func(d []string, g string, e func(string) bool) bool {
			for _, x := range d {
				if e(x + sepFor(g) + "watch_later") {
					return true
				}
			}
			return mpvHasUserConfig(d, g, e)
		},
		"unix separator always": func(d []string, _ string, e func(string) bool) bool {
			return mpvHasUserConfig(d, "linux", e)
		},
		"only mpv.conf": func(d []string, g string, e func(string) bool) bool {
			return mpvHasUserConfig(d, g, func(p string) bool { return strings.HasSuffix(p, "mpv.conf") && e(p) })
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, hasConfigFailures(m), "mutant %q survived", name)
	}
}

// presenceMock counts calls so the tests can prove present() is lazy.
type presenceMock struct {
	answer bool
	calls  int
}

func (p *presenceMock) fn() bool { p.calls++; return p.answer }

var shouldLoadCases = []struct {
	name       string
	envValue   string
	present    bool
	want       bool
	wantProbed bool
}{
	{"config present and not opted out", "", true, true, true},
	{"no config: keep the tuned isolated profile", "", false, false, true},
	{"opted out even with a config", "off", true, false, false},
	{"opted out and no config", "off", false, false, false},
}

func shouldLoadFailures(impl func(func(string) string, func() bool) bool) []string {
	var f []string
	for _, tc := range shouldLoadCases {
		m := &presenceMock{answer: tc.present}
		got := impl(envOf(map[string]string{mpvConfigEnv: tc.envValue}), m.fn)
		if got != tc.want {
			f = append(f, fmt.Sprintf("%s: got %v want %v", tc.name, got, tc.want))
		}
		if (m.calls > 0) != tc.wantProbed {
			f = append(f, fmt.Sprintf("%s: probed=%v want %v", tc.name, m.calls > 0, tc.wantProbed))
		}
	}
	return f
}

func TestShouldLoadMPVUserConfig(t *testing.T) {
	t.Parallel()
	assert.Empty(t, shouldLoadFailures(shouldLoadMPVUserConfig))
}

func TestShouldLoadMPVUserConfigMeta(t *testing.T) {
	t.Parallel()
	mutants := map[string]func(func(string) string, func() bool) bool{
		"ignores presence": func(g func(string) string, _ func() bool) bool { return mpvUserConfigEnabled(g) },
		"ignores opt-out":  func(_ func(string) string, p func() bool) bool { return p() },
		"or instead of and": func(g func(string) string, p func() bool) bool {
			return mpvUserConfigEnabled(g) || p()
		},
		"probes before the opt-out": func(g func(string) string, p func() bool) bool {
			present := p()
			return mpvUserConfigEnabled(g) && present
		},
		"always true": func(func(string) string, func() bool) bool { return true },
	}
	for name, m := range mutants {
		assert.NotEmpty(t, shouldLoadFailures(m), "mutant %q survived", name)
	}
}

func TestMPVUserConfigPresentRealProbe(t *testing.T) {
	// Not parallel: it redirects the process environment.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("MPV_HOME", "")
	t.Setenv("APPDATA", "")
	assert.False(t, mpvUserConfigPresent(), "empty home has no config")

	dir := t.TempDir()
	t.Setenv("MPV_HOME", dir)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mpv.conf"), []byte("osd-level=0\n"), 0o600))
	assert.True(t, mpvUserConfigPresent(), "MPV_HOME with an mpv.conf is a config")
}
