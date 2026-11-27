//go:build !windows

package cfsolver

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// scripts/check-bundled-cfsolve.sh is what release.yml relies on to never
// publish a GoAnime that leaves a new user without a working solver. These
// tests drive it with a fake GoAnime (a shell script that "unpacks" a file the
// way the real one does) and fake solvers, so each way a release could go
// wrong is shown to fail the check.

const bundleCheckScript = "../../../scripts/check-bundled-cfsolve.sh"

// fakeSolverOK and fakeSolverBroken stand in for cfsolve: one starts, one
// exits with an error, as a binary missing a glibc symbol does.
const (
	fakeSolverOK     = "#!/bin/sh\n# solver build A\nexit 0\n"
	fakeSolverOther  = "#!/bin/sh\n# solver build B\nexit 0\n"
	fakeSolverBroken = "#!/bin/sh\nexit 3\n"
)

// fakeGoAnime writes a GoAnime stand-in that copies payload (if any) to the
// config folder the real one unpacks into, then exits with code.
func fakeGoAnime(t *testing.T, dir, payload string, code int) string {
	t.Helper()
	unpack := ""
	if payload != "" {
		unpack = fmt.Sprintf(`case "$(uname -s)" in
  Darwin) d="$HOME/Library/Application Support/goanime/cfsolve" ;;
  *) d="$HOME/.config/goanime/cfsolve" ;;
esac
mkdir -p "$d" && cp %q "$d/challenge" && chmod 700 "$d/challenge"
`, payload)
	}
	path := filepath.Join(dir, "goanime")
	script := "#!/bin/sh\n" + unpack + fmt.Sprintf("exit %d\n", code)
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil { // #nosec G306 -- a test executable
		t.Fatal(err)
	}
	return path
}

func writeFile(t *testing.T, path, content string, mode os.FileMode) string {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
	return path
}

// bundleCheckCase is one release candidate: what GoAnime unpacks, what was
// bundled, and whether the check must pass.
type bundleCheckCase struct {
	name     string
	bundled  string // content of the bundled solver; "" means no file
	unpacked string // content GoAnime unpacks; "" means it unpacks nothing
	exitCode int    // GoAnime's exit code
	noExec   bool   // GoAnime is not executable
	wantOK   bool
	wantMsg  string // what the output must say, so a CI log tells why
}

func bundleCheckCases() []bundleCheckCase {
	return []bundleCheckCase{
		{name: "a working bundle", bundled: fakeSolverOK, unpacked: fakeSolverOK, wantOK: true, wantMsg: "unpacked its solver and it starts"},
		{name: "GoAnime fails before unpacking", bundled: fakeSolverOK, exitCode: 1},
		{name: "GoAnime fails after unpacking", bundled: fakeSolverOK, unpacked: fakeSolverOK, exitCode: 1},
		{name: "nothing unpacked", bundled: fakeSolverOK, wantMsg: "did not unpack a solver"},
		{name: "another solver unpacked", bundled: fakeSolverOK, unpacked: fakeSolverOther, wantMsg: "not the bundled one"},
		{name: "the solver does not start", bundled: fakeSolverBroken, unpacked: fakeSolverBroken, wantMsg: "does not start"},
		{name: "no bundled solver", unpacked: fakeSolverOK, wantMsg: "no bundled solver"},
		{name: "GoAnime not executable", bundled: fakeSolverOK, unpacked: fakeSolverOK, noExec: true, wantMsg: "is not an executable"},
	}
}

// runBundleCheck runs script (a path to a copy of the check) on one case.
func runBundleCheck(t *testing.T, script string, c bundleCheckCase) (ok bool, out string) {
	t.Helper()
	dir := t.TempDir()
	payload := ""
	if c.unpacked != "" {
		payload = writeFile(t, filepath.Join(dir, "payload"), c.unpacked, 0o600)
	}
	goanime := fakeGoAnime(t, dir, payload, c.exitCode)
	if c.noExec {
		if err := os.Chmod(goanime, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	bundled := filepath.Join(dir, "bundled")
	if c.bundled != "" {
		writeFile(t, bundled, c.bundled, 0o600)
	}
	b, err := exec.Command("bash", script, goanime, bundled).CombinedOutput() // #nosec G204 -- the repo's own script
	return err == nil, string(b)
}

// bundleCheckFailures lists the cases script gets wrong; the real one must
// get none.
func bundleCheckFailures(t *testing.T, script string) []string {
	t.Helper()
	var bad []string
	for _, c := range bundleCheckCases() {
		ok, out := runBundleCheck(t, script, c)
		switch {
		case ok != c.wantOK:
			bad = append(bad, fmt.Sprintf("%s: passed=%v, want %v (%s)", c.name, ok, c.wantOK, strings.TrimSpace(out)))
		case !strings.Contains(out, c.wantMsg):
			bad = append(bad, fmt.Sprintf("%s: output %q does not say %q", c.name, strings.TrimSpace(out), c.wantMsg))
		}
	}
	return bad
}

func requireBash(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not installed")
	}
}

func TestBundleCheckScript(t *testing.T) {
	requireBash(t)
	for _, c := range bundleCheckCases() {
		t.Run(c.name, func(t *testing.T) {
			ok, out := runBundleCheck(t, bundleCheckScript, c)
			if ok != c.wantOK || !strings.Contains(out, c.wantMsg) {
				t.Fatalf("passed=%v, want %v; output must say %q:\n%s", ok, c.wantOK, c.wantMsg, out)
			}
		})
	}
}

func TestBundleCheckScriptUsage(t *testing.T) {
	requireBash(t)
	for _, args := range [][]string{{}, {"a"}, {"a", "b", "c"}, {"a", "b", "--in", "img"}, {"a", "b", "--in-image", ""}} {
		cmd := exec.Command("bash", append([]string{bundleCheckScript}, args...)...) // #nosec G204 -- the repo's own script
		err := cmd.Run()
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 2 {
			t.Errorf("args %q: got %v, want exit code 2", args, err)
		}
	}
}

// TestBundleCheckScriptMetatest tampers with the script, one check at a time,
// and requires the case table to catch every tampered copy.
func TestBundleCheckScriptMetatest(t *testing.T) {
	requireBash(t)
	if bad := bundleCheckFailures(t, bundleCheckScript); len(bad) != 0 {
		t.Fatalf("the real script breaks the contract: %v", bad)
	}
	src, err := os.ReadFile(bundleCheckScript)
	if err != nil {
		t.Fatal(err)
	}
	tampers := map[string][2]string{
		"GoAnime's exit code ignored":  {"set -euo pipefail", "set -uo pipefail"},
		"unpacked file not required":   {`[ -s "$solver" ] ||`, "true ||"},
		"another solver accepted":      {`cmp -s "$solver" "$bundled" ||`, "true ||"},
		"a solver that fails accepted": {`"$solver" --help >/dev/null ||`, "true ||"},
	}
	for name, tm := range tampers {
		if n := strings.Count(string(src), tm[0]); n != 1 {
			t.Fatalf("%s: anchor %q occurs %d times, want 1", name, tm[0], n)
		}
		path := filepath.Join(t.TempDir(), "check.sh")
		writeFile(t, path, strings.Replace(string(src), tm[0], tm[1], 1), 0o600)
		if bad := bundleCheckFailures(t, path); len(bad) == 0 {
			t.Errorf("tampered script %q passes the contract", name)
		}
	}
}
