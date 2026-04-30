package upscaler

import (
	"errors"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveFFmpegToolsPrefersPathLookups(t *testing.T) {
	originalLookPath := lookPath
	originalStatPath := statPath
	originalGetEnv := getEnv
	originalGOOS := goos
	t.Cleanup(func() {
		lookPath = originalLookPath
		statPath = originalStatPath
		getEnv = originalGetEnv
		goos = originalGOOS
	})

	goos = "darwin"
	getEnv = func(string) string { return "" }
	lookPath = func(file string) (string, error) {
		switch file {
		case "ffmpeg":
			return "/tmp/shims/ffmpeg", nil
		case "ffprobe":
			return "/tmp/shims/ffprobe", nil
		default:
			return "", errors.New("unexpected tool")
		}
	}
	statPath = func(path string) (fs.FileInfo, error) {
		t.Fatalf("stat fallback should not run when PATH already resolves %s", path)
		return nil, errors.New("unreachable")
	}

	ffmpegPath, ffprobePath := resolveFFmpegTools()
	assert.Equal(t, "/tmp/shims/ffmpeg", ffmpegPath)
	assert.Equal(t, "/tmp/shims/ffprobe", ffprobePath)
}

func TestResolveFFmpegToolsSkipsMacFallbackInE2E(t *testing.T) {
	originalLookPath := lookPath
	originalStatPath := statPath
	originalGetEnv := getEnv
	originalGOOS := goos
	t.Cleanup(func() {
		lookPath = originalLookPath
		statPath = originalStatPath
		getEnv = originalGetEnv
		goos = originalGOOS
	})

	goos = "darwin"
	getEnv = func(key string) string {
		if key == "GOANIME_E2E" {
			return "1"
		}
		return ""
	}
	lookPath = func(_ string) (string, error) {
		return "", errors.New("missing from PATH")
	}
	statPath = func(path string) (fs.FileInfo, error) {
		t.Fatalf("macOS fallback should be disabled in GOANIME_E2E, got stat on %s", path)
		return nil, errors.New("unreachable")
	}

	ffmpegPath, ffprobePath := resolveFFmpegTools()
	assert.Equal(t, "ffmpeg", ffmpegPath)
	assert.Equal(t, "ffprobe", ffprobePath)
}
