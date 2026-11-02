package cfsolver

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
)

const testUA = "Mozilla/5.0 (iPhone; test)"

// fakeTool writes a stand-in for the cfsolve binary that runs script.
func fakeTool(t *testing.T, script string) *Solver {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "challenge")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"+script), 0o700); err != nil { //nolint:gosec // test helper script
		t.Fatal(err)
	}
	prof := filepath.Join(dir, "p.json")
	if err := os.WriteFile(prof, []byte(`{"transport":{"headers":{"User-Agent":"`+testUA+`"}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	return &Solver{Bin: bin, Profile: prof}
}

func TestSolveChallengeReturnsClearanceAndProfileUA(t *testing.T) {
	// $2 is the run directory, which must not exist yet; echo the page back too.
	s := fakeTool(t, `[ ! -e "$3" ] || exit 9
echo "solving $2" >&2
echo "cf_clearance=abc.def-123"
`)
	res, err := s.SolveChallenge(context.Background(), "https://goyabu.io/", time.Minute, netx.RevealNever)
	if err != nil {
		t.Fatal(err)
	}
	if res.UserAgent != testUA {
		t.Errorf("UA = %q, want the profile's %q", res.UserAgent, testUA)
	}
	if len(res.Cookies) != 1 || res.Cookies[0].Name != "cf_clearance" || res.Cookies[0].Value != "abc.def-123" {
		t.Fatalf("cookies = %+v", res.Cookies)
	}
	if res.Cookies[0].Domain != "goyabu.io" {
		t.Errorf("domain = %q", res.Cookies[0].Domain)
	}
}

func TestSolveChallengeFailures(t *testing.T) {
	for name, script := range map[string]string{
		"nonzero exit":    "echo boom >&2; exit 1",
		"token not clear": "echo 0.turnstile-token",
		"empty output":    "true",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := fakeTool(t, script).SolveChallenge(context.Background(), "https://goyabu.io/", time.Minute, netx.RevealNever); err == nil {
				t.Fatal("want an error")
			}
		})
	}
}

func TestSolveChallengeHonorsTimeout(t *testing.T) {
	s := fakeTool(t, "exec sleep 30")
	start := time.Now()
	_, err := s.SolveChallenge(context.Background(), "https://goyabu.io/", 300*time.Millisecond, netx.RevealNever)
	if err == nil || time.Since(start) > 10*time.Second {
		t.Fatalf("err=%v after %v; the solve must be killed at its timeout", err, time.Since(start))
	}
}

func TestDiscoverNeedsBinaryAndProfile(t *testing.T) {
	isolate(t)
	if _, ok := Discover(); ok {
		t.Fatal("discovered a solver with nothing installed")
	}

	s := fakeTool(t, "true")
	t.Setenv(envBin, s.Bin)
	t.Setenv(envProfile, filepath.Join(t.TempDir(), "missing.json"))
	if _, ok := Discover(); ok {
		t.Fatal("a missing profile must disable the solver")
	}
	t.Setenv(envProfile, s.Profile)
	got, ok := Discover()
	if !ok || got.Bin != s.Bin {
		t.Fatalf("Discover = %+v, %v", got, ok)
	}
}

func TestRegisterOnlyWhenInstalled(t *testing.T) {
	t.Cleanup(func() { netx.RegisterChallengeSolver(nil) })
	netx.RegisterChallengeSolver(nil)
	isolate(t)
	if Register() || netx.ChallengeSolverAvailable() {
		t.Fatal("registered without cfsolve installed")
	}
	s := fakeTool(t, "true")
	t.Setenv(envBin, s.Bin)
	t.Setenv(envProfile, s.Profile)
	if !Register() || !netx.ChallengeSolverAvailable() {
		t.Fatal("not registered although cfsolve is installed")
	}
}

// isolate hides every place a real cfsolve could be found.
func isolate(t *testing.T) {
	t.Helper()
	t.Setenv(envBin, "")
	t.Setenv(envProfile, "")
	t.Setenv("PATH", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("AppData", t.TempDir())
	t.Setenv("LocalAppData", t.TempDir())
}

func TestDiscoverFindsBinaryInConfigFolderWithEmbeddedProfile(t *testing.T) {
	isolate(t)
	cfg, err := os.UserConfigDir()
	if err != nil {
		t.Skip(err)
	}
	dir := filepath.Join(cfg, "goanime", "cfsolve")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	name := "challenge"
	if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"), 0o700); err != nil { //nolint:gosec // test helper
		t.Fatal(err)
	}
	got, ok := Discover()
	if !ok {
		t.Fatal("a binary in the config folder was not found")
	}
	if filepath.Base(got.Bin) != name {
		t.Errorf("Bin = %q", got.Bin)
	}
	// The profile came from the embedded copy, not from anything the user set up.
	ua, err := got.userAgent()
	if err != nil || !strings.Contains(ua, "iPhone") {
		t.Fatalf("embedded profile unusable: ua=%q err=%v", ua, err)
	}
	for _, f := range []string{"safari-26-ios/answers.json", "safari-26-ios/challenge-page-answers.json"} {
		if _, err := os.Stat(filepath.Join(filepath.Dir(got.Profile), f)); err != nil {
			t.Errorf("profile file %s was not unpacked: %v", f, err)
		}
	}
	again, _ := Discover()
	if again.Profile != got.Profile {
		t.Errorf("second unpack moved the profile: %s vs %s", again.Profile, got.Profile)
	}
}

func TestSolveTurnstilePassesTheWidgetAndReturnsTheToken(t *testing.T) {
	out := filepath.Join(t.TempDir(), "args")
	s := fakeTool(t, `printf '%s\n' "$@" > `+out+`
echo "progress" >&2
echo "0.the-token-value"
`)
	tok, err := s.SolveTurnstile(context.Background(), TurnstileRequest{
		PageURL: "https://site.example/filme/1", SiteKey: "0xKEY", Action: "embed_access", CData: "cd-1",
	}, time.Minute)
	if err != nil || tok != "0.the-token-value" {
		t.Fatalf("token = %q, err = %v", tok, err)
	}
	b, _ := os.ReadFile(out)
	got := string(b)
	for _, want := range []string{"--blank-page", "--solve", "--sitekey\n0xKEY", "--action\nembed_access", "--cdata\ncd-1", "https://site.example/filme/1"} {
		if !strings.Contains(got, want) {
			t.Errorf("argv lacks %q:\n%s", want, got)
		}
	}
}

func TestSolveTurnstileFailures(t *testing.T) {
	ctx := context.Background()
	if _, err := fakeTool(t, "true").SolveTurnstile(ctx, TurnstileRequest{PageURL: "https://x.example/"}, time.Minute); err == nil {
		t.Error("a missing sitekey must be refused before running anything")
	}
	r := TurnstileRequest{PageURL: "https://x.example/", SiteKey: "k"}
	for name, script := range map[string]string{"exit": "exit 3", "empty": "true", "clearance not token": "echo cf_clearance=abc"} {
		if _, err := fakeTool(t, script).SolveTurnstile(ctx, r, time.Minute); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

func TestParseClearance(t *testing.T) {
	if v, ok := parseClearance("noise\n  cf_clearance=xyz \n"); !ok || v != "xyz" {
		t.Errorf("got %q %v", v, ok)
	}
	if _, ok := parseClearance("cf_clearance=\n"); ok {
		t.Error("empty value accepted")
	}
	if !strings.Contains(lastLine("a\nb\n"), "b") {
		t.Error("lastLine")
	}
}
