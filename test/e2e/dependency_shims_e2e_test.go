//go:build e2e

package e2e_test

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLIUpscaleReportsMissingFFmpegWithIsolatedPath(t *testing.T) {
	tmp := t.TempDir()
	input := filepath.Join(tmp, "frame.png")
	writeTestPNG(t, input)

	env := isolatedEnvWithPath(t, t.TempDir())
	code, output := runGoAnimeWithEnv(t, env, "--upscale", input)

	if code == 0 {
		t.Fatalf("expected non-zero exit when ffmpeg is missing\noutput:\n%s", output)
	}
	assertOutputContains(t, output, "FFmpeg validation failed", "FFmpeg not found")
}

func TestCLIUpscaleUsesFFmpegShimFromIsolatedPath(t *testing.T) {
	tmp := t.TempDir()
	input := filepath.Join(tmp, "frame.png")
	outputPath := filepath.Join(tmp, "frame-upscaled.png")
	writeTestPNG(t, input)

	shimDir := t.TempDir()
	markerPath := filepath.Join(tmp, "ffmpeg-invocations.txt")
	writeFFmpegShim(t, shimDir, markerPath)

	env := isolatedEnvWithPath(t, shimDir)
	code, output := runGoAnimeWithEnv(t, env,
		"--upscale",
		"--upscale-output", outputPath,
		"--upscale-passes", "1",
		"--upscale-fast",
		input,
	)
	assertExitCode(t, code, 0, output)

	marker, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatalf("expected ffmpeg shim to be invoked: %v\noutput:\n%s", err, output)
	}
	if !strings.Contains(string(marker), "-version") {
		t.Fatalf("ffmpeg shim invocations = %q, want -version", string(marker))
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected upscaled output at %s: %v\noutput:\n%s", outputPath, err, output)
	}
}

func isolatedEnvWithPath(t *testing.T, path string) []string {
	t.Helper()

	env := isolatedEnv(t)
	env = replaceEnv(env, "PATH", path)
	env = replaceEnv(env, "Path", path)
	return env
}

func replaceEnv(env []string, key, value string) []string {
	prefix := key + "="
	replaced := false
	for i, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			env[i] = prefix + value
			replaced = true
			break
		}
	}
	if !replaced {
		env = append(env, prefix+value)
	}
	return env
}

func writeTestPNG(t *testing.T, path string) {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 0x20, G: 0x40, B: 0x80, A: 0xff})
	img.Set(1, 0, color.RGBA{R: 0xf0, G: 0xc0, B: 0x40, A: 0xff})
	img.Set(0, 1, color.RGBA{R: 0x30, G: 0x90, B: 0x60, A: 0xff})
	img.Set(1, 1, color.RGBA{R: 0xe0, G: 0x40, B: 0x50, A: 0xff})

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("failed to close test image: %v", err)
		}
	}()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("failed to encode test image: %v", err)
	}
}

func writeFFmpegShim(t *testing.T, dir, markerPath string) {
	t.Helper()

	if runtime.GOOS == "windows" {
		script := "@echo off\r\n" +
			"echo %*>>\"" + markerPath + "\"\r\n" +
			"echo ffmpeg version e2e-shim\r\n" +
			"exit /b 0\r\n"
		if err := os.WriteFile(filepath.Join(dir, "ffmpeg.bat"), []byte(script), 0o700); err != nil {
			t.Fatalf("failed to write ffmpeg shim: %v", err)
		}
		return
	}

	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$*\" >> '" + markerPath + "'\n" +
		"printf '%s\\n' 'ffmpeg version e2e-shim'\n" +
		"exit 0\n"
	if err := os.WriteFile(filepath.Join(dir, "ffmpeg"), []byte(script), 0o700); err != nil {
		t.Fatalf("failed to write ffmpeg shim: %v", err)
	}
}
