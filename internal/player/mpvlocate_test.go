package player

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type modeInfo struct {
	os.FileInfo
	dir  bool
	mode os.FileMode
}

func (m modeInfo) IsDir() bool       { return m.dir }
func (m modeInfo) Mode() os.FileMode { return m.mode }

// fakeSystem is a PATH and a file system in a map. A nil entry value means
// the path does not exist.
type fakeSystem struct {
	onPath string // what lookPath returns; empty = not on PATH
	files  map[string]modeInfo
	stats  []string // every path stat was asked about, in order
}

func (f *fakeSystem) locator() mpvLocator {
	return mpvLocator{
		lookPath: func(string) (string, error) {
			if f.onPath == "" {
				return "", errors.New("not found")
			}
			return f.onPath, nil
		},
		stat: func(name string) (os.FileInfo, error) {
			f.stats = append(f.stats, name)
			if fi, ok := f.files[name]; ok {
				return fi, nil
			}
			return nil, os.ErrNotExist
		},
	}
}

type locateFn func(l mpvLocator, candidates []string, needExecBit bool) (string, bool)

type locateCase struct {
	name     string
	onPath   string
	files    map[string]modeInfo
	execBit  bool
	want     string
	wantOK   bool
	noStats  bool // the disk must not be touched
	maxStats int  // when > 0, the most stat calls allowed
}

var exec755 = modeInfo{mode: 0o755}

func locateCases() []locateCase {
	return []locateCase{
		{name: "PATH wins without touching the disk", onPath: "/bin/mpv", files: map[string]modeInfo{"/a/mpv": exec755}, execBit: true, want: "/bin/mpv", wantOK: true, noStats: true},
		{name: "first existing candidate wins", files: map[string]modeInfo{"/b/mpv": exec755, "/c/mpv": exec755}, execBit: true, want: "/b/mpv", wantOK: true, maxStats: 2},
		{name: "nothing found", files: map[string]modeInfo{}, execBit: true, wantOK: false},
		{name: "a directory is skipped", files: map[string]modeInfo{"/a/mpv": {dir: true, mode: 0o755}, "/b/mpv": exec755}, execBit: true, want: "/b/mpv", wantOK: true},
		{name: "unix needs the exec bit", files: map[string]modeInfo{"/a/mpv": {mode: 0o644}, "/c/mpv": exec755}, execBit: true, want: "/c/mpv", wantOK: true},
		{name: "any single exec bit is enough", files: map[string]modeInfo{"/a/mpv": {mode: 0o001}}, execBit: true, want: "/a/mpv", wantOK: true},
		{name: "windows ignores the exec bit", files: map[string]modeInfo{"/a/mpv": {mode: 0o644}}, execBit: false, want: "/a/mpv", wantOK: true},
		{name: "windows still skips directories", files: map[string]modeInfo{"/a/mpv": {dir: true}}, execBit: false, wantOK: false},
		{name: "a lone non-executable file is not found", files: map[string]modeInfo{"/a/mpv": {mode: 0o644}}, execBit: true, wantOK: false},
	}
}

func locateFailures(fn locateFn) []string {
	var bad []string
	for _, c := range locateCases() {
		sys := &fakeSystem{onPath: c.onPath, files: c.files}
		got, ok := fn(sys.locator(), []string{"/a/mpv", "/b/mpv", "/c/mpv"}, c.execBit)
		if got != c.want || ok != c.wantOK {
			bad = append(bad, fmt.Sprintf("%s: got (%q,%v), want (%q,%v)", c.name, got, ok, c.want, c.wantOK))
		}
		if c.noStats && len(sys.stats) > 0 {
			bad = append(bad, fmt.Sprintf("%s: touched the disk %v", c.name, sys.stats))
		}
		if c.maxStats > 0 && len(sys.stats) > c.maxStats {
			bad = append(bad, fmt.Sprintf("%s: %d stats, want at most %d", c.name, len(sys.stats), c.maxStats))
		}
	}
	return bad
}

func TestLocateMPV(t *testing.T) {
	if bad := locateFailures(locateMPV); len(bad) > 0 {
		t.Fatal(strings.Join(bad, "\n"))
	}
}

func TestMeta_LocateMPV_CatchesBrokenImplementations(t *testing.T) {
	mutants := map[string]locateFn{
		"ignores PATH": func(l mpvLocator, c []string, x bool) (string, bool) {
			l.lookPath = func(string) (string, error) { return "", errors.New("no") }
			return locateMPV(l, c, x)
		},
		"candidates before PATH": func(l mpvLocator, c []string, x bool) (string, bool) {
			for _, p := range c {
				if _, err := l.stat(p); err == nil {
					return p, true
				}
			}
			return locateMPV(l, c, x)
		},
		"takes the last candidate": func(l mpvLocator, c []string, x bool) (string, bool) {
			r := make([]string, len(c))
			for i := range c {
				r[len(c)-1-i] = c[i]
			}
			return locateMPV(l, r, x)
		},
		"accepts directories": func(l mpvLocator, c []string, x bool) (string, bool) {
			inner := l.stat
			l.stat = func(n string) (os.FileInfo, error) {
				fi, err := inner(n)
				if err == nil {
					return modeInfo{mode: fi.Mode()}, nil
				}
				return fi, err
			}
			return locateMPV(l, c, x)
		},
		"never checks the exec bit": func(l mpvLocator, c []string, _ bool) (string, bool) {
			return locateMPV(l, c, false)
		},
		"always checks the exec bit": func(l mpvLocator, c []string, _ bool) (string, bool) {
			return locateMPV(l, c, true)
		},
		"requires all exec bits": func(l mpvLocator, c []string, x bool) (string, bool) {
			inner := l.stat
			l.stat = func(n string) (os.FileInfo, error) {
				fi, err := inner(n)
				if err == nil && fi.Mode()&0o111 != 0o111 {
					return modeInfo{mode: 0}, nil
				}
				return fi, err
			}
			return locateMPV(l, c, x)
		},
		"reports found for nothing": func(l mpvLocator, c []string, x bool) (string, bool) {
			p, _ := locateMPV(l, c, x)
			return p, true
		},
	}
	for name, fn := range mutants {
		t.Run(name, func(t *testing.T) {
			if len(locateFailures(fn)) == 0 {
				t.Fatalf("mutant %q survived", name)
			}
		})
	}
}

// ---- candidate lists ----

func j(parts ...string) string { return filepath.Join(parts...) }

type candFn func(goos, home string) []string

func unixCandFailures(fn candFn) []string {
	var bad []string
	eq := func(name string, got, want []string) {
		if strings.Join(got, "|") != strings.Join(want, "|") {
			bad = append(bad, fmt.Sprintf("%s:\n got  %q\n want %q", name, got, want))
		}
	}
	flatpak := j("/h", ".local", "share", "flatpak", "exports", "bin", "io.mpv.Mpv")
	eq("darwin", fn("darwin", "/h"), []string{
		"/opt/homebrew/bin/mpv", "/usr/local/bin/mpv", "/opt/local/bin/mpv",
		"/Applications/mpv.app/Contents/MacOS/mpv", flatpak})
	eq("darwin without home", fn("darwin", ""), []string{
		"/opt/homebrew/bin/mpv", "/usr/local/bin/mpv", "/opt/local/bin/mpv",
		"/Applications/mpv.app/Contents/MacOS/mpv"})
	eq("linux", fn("linux", "/h"), []string{
		j("/h", ".linuxbrew", "bin", "mpv"),
		"/home/linuxbrew/.linuxbrew/bin/mpv", "/usr/bin/mpv", "/usr/local/bin/mpv",
		"/snap/bin/mpv", "/var/lib/flatpak/exports/bin/io.mpv.Mpv", flatpak})
	eq("linux without home", fn("linux", ""), []string{
		"/home/linuxbrew/.linuxbrew/bin/mpv", "/usr/bin/mpv", "/usr/local/bin/mpv",
		"/snap/bin/mpv", "/var/lib/flatpak/exports/bin/io.mpv.Mpv"})
	eq("freebsd behaves like linux", fn("freebsd", ""), fn("linux", ""))
	return bad
}

func TestMPVCandidatesUnix(t *testing.T) {
	if bad := unixCandFailures(mpvCandidatesUnix); len(bad) > 0 {
		t.Fatal(strings.Join(bad, "\n"))
	}
}

func TestMeta_MPVCandidatesUnix_CatchesBrokenImplementations(t *testing.T) {
	mutants := map[string]candFn{
		"darwin uses linux paths": func(_, h string) []string { return mpvCandidatesUnix("linux", h) },
		"linux uses darwin paths": func(g, h string) []string {
			if g != "darwin" {
				return mpvCandidatesUnix("darwin", h)
			}
			return mpvCandidatesUnix(g, h)
		},
		"drops the per-user flatpak": func(g, h string) []string {
			c := mpvCandidatesUnix(g, h)
			if h != "" {
				return c[:len(c)-1]
			}
			return c
		},
		"homebrew per-user last": func(g, h string) []string {
			c := mpvCandidatesUnix(g, h)
			if g == "linux" && h != "" {
				return append(c[1:], c[0])
			}
			return c
		},
		"empty home still adds paths": func(g, h string) []string {
			if h == "" {
				return mpvCandidatesUnix(g, "x")
			}
			return mpvCandidatesUnix(g, h)
		},
		"nothing": func(string, string) []string { return nil },
	}
	for name, fn := range mutants {
		t.Run(name, func(t *testing.T) {
			if len(unixCandFailures(fn)) == 0 {
				t.Fatalf("mutant %q survived", name)
			}
		})
	}
}

type winCandFn func(exec string, env map[string]string) []string

func winCandFailures(fn winCandFn) []string {
	var bad []string
	eq := func(name string, got, want []string) {
		if strings.Join(got, "|") != strings.Join(want, "|") {
			bad = append(bad, fmt.Sprintf("%s:\n got  %q\n want %q", name, got, want))
		}
	}
	env := map[string]string{
		"ProgramFiles": "PF", "ProgramFiles(x86)": "PF86", "LOCALAPPDATA": "LA", "USERPROFILE": "UP",
	}
	eq("everything set", fn(j("X", "goanime.exe"), env), []string{
		j("X", "bin", "mpv.exe"), j("X", "mpv.exe"),
		j("PF", "GoAnime", "bin", "mpv.exe"), j("PF", "mpv", "mpv.exe"), j("PF", "mpv.net", "mpv.exe"),
		j("PF86", "GoAnime", "bin", "mpv.exe"), j("PF86", "mpv", "mpv.exe"),
		j("LA", "Programs", "GoAnime", "bin", "mpv.exe"),
		j("UP", "scoop", "apps", "mpv", "current", "mpv.exe"), j("UP", "scoop", "shims", "mpv.exe"),
	})
	eq("nothing known", fn("", map[string]string{}), nil)
	eq("only the executable", fn(j("X", "goanime.exe"), nil), []string{j("X", "bin", "mpv.exe"), j("X", "mpv.exe")})
	eq("only scoop", fn("", map[string]string{"USERPROFILE": "UP"}), []string{
		j("UP", "scoop", "apps", "mpv", "current", "mpv.exe"), j("UP", "scoop", "shims", "mpv.exe")})
	eq("only program files x86", fn("", map[string]string{"ProgramFiles(x86)": "PF86"}), []string{
		j("PF86", "GoAnime", "bin", "mpv.exe"), j("PF86", "mpv", "mpv.exe")})
	eq("only local app data", fn("", map[string]string{"LOCALAPPDATA": "LA"}), []string{
		j("LA", "Programs", "GoAnime", "bin", "mpv.exe")})
	return bad
}

func realWin(exec string, env map[string]string) []string {
	return mpvCandidatesWindows(exec, func(k string) string { return env[k] })
}

func TestMPVCandidatesWindows(t *testing.T) {
	if bad := winCandFailures(realWin); len(bad) > 0 {
		t.Fatal(strings.Join(bad, "\n"))
	}
}

func TestMeta_MPVCandidatesWindows_CatchesBrokenImplementations(t *testing.T) {
	mutants := map[string]winCandFn{
		"ignores the executable directory": func(_ string, env map[string]string) []string { return realWin("", env) },
		"ignores program files": func(e string, env map[string]string) []string {
			return realWin(e, without(env, "ProgramFiles"))
		},
		"ignores x86 program files": func(e string, env map[string]string) []string {
			return realWin(e, without(env, "ProgramFiles(x86)"))
		},
		"ignores local app data": func(e string, env map[string]string) []string {
			return realWin(e, without(env, "LOCALAPPDATA"))
		},
		"ignores scoop": func(e string, env map[string]string) []string {
			return realWin(e, without(env, "USERPROFILE"))
		},
		"portable before installer": func(e string, env map[string]string) []string {
			c := realWin(e, env)
			if e != "" {
				c[0], c[1] = c[1], c[0]
			}
			return c
		},
		"adds empty-variable paths": func(e string, env map[string]string) []string {
			c := realWin(e, env)
			return append(c, j("", "mpv", "mpv.exe"))
		},
	}
	for name, fn := range mutants {
		t.Run(name, func(t *testing.T) {
			if len(winCandFailures(fn)) == 0 {
				t.Fatalf("mutant %q survived", name)
			}
		})
	}
}

func without(env map[string]string, key string) map[string]string {
	out := map[string]string{}
	for k, v := range env {
		if k != key {
			out[k] = v
		}
	}
	return out
}

// findMPVPath is the thin real wrapper: with a PATH that has an executable
// named mpv it must return that file; with an empty PATH and no install it
// must fail rather than invent a path.
func TestFindMPVPath_UsesPATH(t *testing.T) {
	dir := t.TempDir()
	name := "mpv"
	if os.PathSeparator == '\\' {
		name = "mpv.exe"
	}
	bin := filepath.Join(dir, name)
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	got, err := findMPVPath()
	if err != nil || got != bin {
		t.Fatalf("findMPVPath() = (%q, %v), want (%q, nil)", got, err, bin)
	}
}

// ---- locateAllMPV ----------------------------------------------------------------

type locateAllFn func(l mpvLocator, candidates []string, needExecBit, foldCase bool) []string

type locateAllCase struct {
	name       string
	onPath     string
	files      map[string]modeInfo
	candidates []string
	execBit    bool
	foldCase   bool
	want       []string
}

func locateAllCases() []locateAllCase {
	abc := []string{"/a/mpv", "/b/mpv", "/c/mpv"}
	return []locateAllCase{
		{name: "PATH first, then every existing candidate", onPath: "/bin/mpv", files: map[string]modeInfo{"/a/mpv": exec755, "/c/mpv": exec755}, candidates: abc, execBit: true, want: []string{"/bin/mpv", "/a/mpv", "/c/mpv"}},
		{name: "no PATH hit lists the candidates in order", files: map[string]modeInfo{"/b/mpv": exec755, "/a/mpv": exec755}, candidates: abc, execBit: true, want: []string{"/a/mpv", "/b/mpv"}},
		{name: "nothing found is an empty list", files: map[string]modeInfo{}, candidates: abc, execBit: true, want: nil},
		{name: "the PATH hit is not repeated as a candidate", onPath: "/a/mpv", files: map[string]modeInfo{"/a/mpv": exec755, "/b/mpv": exec755}, candidates: abc, execBit: true, want: []string{"/a/mpv", "/b/mpv"}},
		{name: "a candidate listed twice appears once", files: map[string]modeInfo{"/a/mpv": exec755}, candidates: []string{"/a/mpv", "/a/mpv"}, execBit: true, want: []string{"/a/mpv"}},
		{name: "directories are skipped", files: map[string]modeInfo{"/a/mpv": {dir: true, mode: 0o755}, "/b/mpv": exec755}, candidates: abc, execBit: true, want: []string{"/b/mpv"}},
		{name: "unix needs the exec bit", files: map[string]modeInfo{"/a/mpv": {mode: 0o644}, "/b/mpv": exec755}, candidates: abc, execBit: true, want: []string{"/b/mpv"}},
		{name: "windows ignores the exec bit", files: map[string]modeInfo{"/a/mpv": {mode: 0o644}}, candidates: abc, execBit: false, want: []string{"/a/mpv"}},
		{name: "windows paths differing in case are one mpv", onPath: `C:\GoAnime\bin\mpv.exe`, files: map[string]modeInfo{`c:\goanime\bin\mpv.exe`: exec755}, candidates: []string{`c:\goanime\bin\mpv.exe`}, foldCase: true, want: []string{`C:\GoAnime\bin\mpv.exe`}},
		{name: "case matters where the file system cares", onPath: "/A/mpv", files: map[string]modeInfo{"/a/mpv": exec755}, candidates: []string{"/a/mpv"}, execBit: true, want: []string{"/A/mpv", "/a/mpv"}},
		{name: "a PATH stub does not hide the installer's mpv", onPath: `C:\Users\n\AppData\Local\Microsoft\WindowsApps\mpv.exe`, files: map[string]modeInfo{`C:\Program Files\GoAnime\bin\mpv.exe`: exec755}, candidates: []string{`C:\Program Files\GoAnime\bin\mpv.exe`}, want: []string{`C:\Users\n\AppData\Local\Microsoft\WindowsApps\mpv.exe`, `C:\Program Files\GoAnime\bin\mpv.exe`}},
	}
}

func locateAllFailures(fn locateAllFn) []string {
	var bad []string
	for _, c := range locateAllCases() {
		sys := &fakeSystem{onPath: c.onPath, files: c.files}
		got := fn(sys.locator(), c.candidates, c.execBit, c.foldCase)
		if fmt.Sprint(got) != fmt.Sprint(c.want) {
			bad = append(bad, fmt.Sprintf("%s: got %q, want %q", c.name, got, c.want))
		}
	}
	return bad
}

func TestLocateAllMPV(t *testing.T) {
	if bad := locateAllFailures(locateAllMPV); len(bad) > 0 {
		t.Fatal(strings.Join(bad, "\n"))
	}
}

// The first of them is what locateMPV answers, so starting mpv prefers what
// it always preferred.
func TestLocateAllMPVStartsWhereLocateMPVDoes(t *testing.T) {
	for _, c := range locateCases() {
		sys := &fakeSystem{onPath: c.onPath, files: c.files}
		cands := []string{"/a/mpv", "/b/mpv", "/c/mpv"}
		first, ok := locateMPV(sys.locator(), cands, c.execBit)
		all := locateAllMPV((&fakeSystem{onPath: c.onPath, files: c.files}).locator(), cands, c.execBit, false)
		if !ok {
			if len(all) != 0 {
				t.Errorf("%s: locateMPV found nothing but locateAllMPV listed %q", c.name, all)
			}
			continue
		}
		if len(all) == 0 || all[0] != first {
			t.Errorf("%s: locateAllMPV starts with %q, locateMPV answers %q", c.name, all, first)
		}
	}
}

func TestMeta_LocateAllMPV_CatchesBrokenImplementations(t *testing.T) {
	mutants := map[string]locateAllFn{
		"ignores PATH": func(l mpvLocator, c []string, x, f bool) []string {
			l.lookPath = func(string) (string, error) { return "", errors.New("no") }
			return locateAllMPV(l, c, x, f)
		},
		"stops at the first": func(l mpvLocator, c []string, x, f bool) []string {
			if all := locateAllMPV(l, c, x, f); len(all) > 1 {
				return all[:1]
			}
			return locateAllMPV(l, c, x, f)
		},
		"candidates before PATH": func(l mpvLocator, c []string, x, f bool) []string {
			all := locateAllMPV(l, c, x, f)
			if len(all) > 1 {
				return append(all[1:], all[0])
			}
			return all
		},
		"reversed": func(l mpvLocator, c []string, x, f bool) []string {
			all := locateAllMPV(l, c, x, f)
			for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
				all[i], all[j] = all[j], all[i]
			}
			return all
		},
		"repeats a path": func(l mpvLocator, c []string, x, f bool) []string {
			all := locateAllMPV(l, c, x, f)
			if len(all) > 0 {
				return append(all, all[0])
			}
			return all
		},
		"accepts directories": func(l mpvLocator, c []string, x, f bool) []string {
			inner := l.stat
			l.stat = func(n string) (os.FileInfo, error) {
				fi, err := inner(n)
				if err == nil {
					return modeInfo{mode: fi.Mode()}, nil
				}
				return fi, err
			}
			return locateAllMPV(l, c, x, f)
		},
		"never checks the exec bit": func(l mpvLocator, c []string, _, f bool) []string {
			return locateAllMPV(l, c, false, f)
		},
		"always checks the exec bit": func(l mpvLocator, c []string, _, f bool) []string {
			return locateAllMPV(l, c, true, f)
		},
		"never folds case": func(l mpvLocator, c []string, x, _ bool) []string {
			return locateAllMPV(l, c, x, false)
		},
		"always folds case": func(l mpvLocator, c []string, x, _ bool) []string {
			return locateAllMPV(l, c, x, true)
		},
	}
	for name, fn := range mutants {
		t.Run(name, func(t *testing.T) {
			if len(locateAllFailures(fn)) == 0 {
				t.Fatalf("mutant %q survived", name)
			}
		})
	}
}
