//go:build e2e

package e2e_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	buildOnce   sync.Once
	binaryPath  string
	binaryError error
	buildDir    string
)

const e2eCommandTimeout = 30 * time.Second

func TestMain(m *testing.M) {
	code := m.Run()
	if buildDir != "" {
		_ = os.RemoveAll(buildDir)
	}
	os.Exit(code)
}

func goanimeBinary(t *testing.T) string {
	t.Helper()

	buildOnce.Do(func() {
		var err error
		buildDir, err = os.MkdirTemp("", "goanime-e2e-*")
		if err != nil {
			binaryError = err
			return
		}

		name := "goanime-e2e"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		binaryPath = filepath.Join(buildDir, name)

		cmd := exec.Command("go", "build", "-trimpath", "-o", binaryPath, "./cmd/goanime")
		cmd.Dir = repoRoot(t)
		cmd.Env = append(os.Environ(), "CGO_ENABLED=0")

		var output bytes.Buffer
		cmd.Stdout = &output
		cmd.Stderr = &output
		if err := cmd.Run(); err != nil {
			binaryError = &commandError{err: err, output: output.String()}
		}
	})

	if binaryError != nil {
		t.Fatalf("failed to build goanime binary: %v", binaryError)
	}
	return binaryPath
}

func repoRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

func runGoAnime(t *testing.T, args ...string) (int, string) {
	t.Helper()

	return runGoAnimeWithEnv(t, isolatedEnv(t), args...)
}

func runGoAnimeWithEnv(t *testing.T, env []string, args ...string) (int, string) {
	t.Helper()

	binary := goanimeBinary(t)

	ctx, cancel := context.WithTimeout(context.Background(), e2eCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = repoRoot(t)
	cmd.Env = env

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("goanime %q timed out; likely a regression in early validation\noutput:\n%s", strings.Join(args, " "), output.String())
	}
	if err == nil {
		return 0, output.String()
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode(), output.String()
	}
	t.Fatalf("failed to run goanime %q: %v\n%s", strings.Join(args, " "), err, output.String())
	return -1, ""
}

func isolatedEnv(t *testing.T) []string {
	t.Helper()

	home := t.TempDir()
	env := append(os.Environ(),
		"GOANIME_E2E=1",
		"HOME="+home,
		"USERPROFILE="+home,
	)
	return env
}

func assertExitCode(t *testing.T, got, want int, output string) {
	t.Helper()

	if got != want {
		t.Fatalf("exit code = %d, want %d\noutput:\n%s", got, want, output)
	}
}

func assertOutputContains(t *testing.T, output string, want ...string) {
	t.Helper()

	for _, fragment := range want {
		if !strings.Contains(output, fragment) {
			t.Fatalf("output does not contain %q\noutput:\n%s", fragment, output)
		}
	}
}

func TestCLIHelpAndVersionSmoke(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "help",
			args: []string{"--help"},
			want: []string{"GoAnime", "Usage:", "--update", "--upscale"},
		},
		{
			name: "short help",
			args: []string{"-h"},
			want: []string{"GoAnime", "Options:", "Examples:"},
		},
		{
			name: "version",
			args: []string{"--version"},
			want: []string{"GoAnime v", "tracking"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, output := runGoAnime(t, tt.args...)
			assertExitCode(t, code, 0, output)
			assertOutputContains(t, output, tt.want...)
		})
	}
}

func TestCLIRejectsInvalidPlaybackName(t *testing.T) {
	code, output := runGoAnime(t, "abc")

	if code == 0 {
		t.Fatalf("expected non-zero exit for short playback query\noutput:\n%s", output)
	}
	assertOutputContains(t, output, "anime name must have at least 4 characters")
}

func TestCLIRejectsInvalidDownloadArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "download requires arguments",
			args: []string{"-d"},
			want: "download mode requires anime name and episode number/range",
		},
		{
			name: "download range must be ordered",
			args: []string{"-d", "-r", "naruto", "5-1"},
			want: "start episode (5) cannot be greater than end episode (1)",
		},
		{
			name: "movie tv download requires season and episode",
			args: []string{"-dm", "--type", "tv", "dexter"},
			want: "TV episode download requires show name, season number, and episode number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, output := runGoAnime(t, tt.args...)
			if code == 0 {
				t.Fatalf("expected non-zero exit\noutput:\n%s", output)
			}
			assertOutputContains(t, output, tt.want)
		})
	}
}

func TestCLIRejectsInvalidUpscaleArguments(t *testing.T) {
	missingInput := filepath.Join(t.TempDir(), "definitely-missing.png")

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "upscale requires input",
			args: []string{"--upscale"},
			want: "upscale mode requires an input file path",
		},
		{
			name: "upscale input must exist",
			args: []string{"--upscale", missingInput},
			want: "input file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, output := runGoAnime(t, tt.args...)
			if code == 0 {
				t.Fatalf("expected non-zero exit\noutput:\n%s", output)
			}
			assertOutputContains(t, output, tt.want)
		})
	}
}

func TestCLISmokeCompletesQuickly(t *testing.T) {
	start := time.Now()
	code, output := runGoAnime(t, "--version")
	assertExitCode(t, code, 0, output)

	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("--version took %s, expected a fast no-network smoke path", elapsed)
	}
}

type commandError struct {
	err    error
	output string
}

func (e *commandError) Error() string {
	return e.err.Error() + "\n" + e.output
}
