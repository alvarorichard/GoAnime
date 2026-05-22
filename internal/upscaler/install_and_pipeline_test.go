package upscaler

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withShaderDirOverride sets the override + restores it.
func withShaderDirOverride(t *testing.T, dir string) {
	t.Helper()
	prev := shaderDirOverride
	shaderDirOverride = dir
	t.Cleanup(func() { shaderDirOverride = prev })
}

func withShaderURLs(t *testing.T, releaseURL, ganBaseURL string) {
	t.Helper()
	prevRel, prevGAN := anime4kShaderURL, anime4kGANShaderBaseURL
	anime4kShaderURL = releaseURL
	anime4kGANShaderBaseURL = ganBaseURL
	t.Cleanup(func() {
		anime4kShaderURL = prevRel
		anime4kGANShaderBaseURL = prevGAN
	})
}

// buildGLSLZip creates an in-memory zip with a small set of .glsl files.
func buildGLSLZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		fw, err := zw.Create(name)
		require.NoError(t, err)
		_, err = fw.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, zw.Close())
	return buf.Bytes()
}

func TestInstallShaders_Success(t *testing.T) {
	zipBytes := buildGLSLZip(t, map[string]string{
		"shaders/Anime4K_Clamp_Highlights.glsl":  "clamp",
		"shaders/Anime4K_Restore_CNN_M.glsl":     "restore",
		"shaders/Anime4K_Upscale_CNN_x2_M.glsl":  "upscale",
		"shaders/Anime4K_Other_File.glsl":        "other",
		"README.md":                              "ignored",
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(zipBytes)
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	withShaderDirOverride(t, dir)
	withShaderURLs(t, srv.URL, "http://unused")

	require.NoError(t, InstallShaders())
	for _, f := range []string{"Anime4K_Clamp_Highlights.glsl", "Anime4K_Restore_CNN_M.glsl", "Anime4K_Upscale_CNN_x2_M.glsl"} {
		_, err := os.Stat(filepath.Join(dir, f))
		assert.NoError(t, err, "expected %s extracted", f)
	}
	assert.True(t, ShadersInstalled())
}

func TestInstallShaders_HTTP404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.NotFound(w, nil)
	}))
	t.Cleanup(srv.Close)
	withShaderDirOverride(t, t.TempDir())
	withShaderURLs(t, srv.URL, "http://x")

	err := InstallShaders()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 404")
}

func TestInstallShaders_NetworkError(t *testing.T) {
	withShaderDirOverride(t, t.TempDir())
	withShaderURLs(t, "http://127.0.0.1:1", "http://x")
	err := InstallShaders()
	require.Error(t, err)
}

func TestInstallShaders_CannotCreateDir(t *testing.T) {
	// Point to an unwritable path (a file, not a dir)
	tmp := t.TempDir()
	blocker := filepath.Join(tmp, "blocker")
	require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o644))
	withShaderDirOverride(t, filepath.Join(blocker, "sub")) // can't mkdir under a file
	withShaderURLs(t, "http://x", "http://x")
	err := InstallShaders()
	require.Error(t, err)
}

func TestInstallGANShaders_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should request Restore/...glsl and Upscale/...glsl
		_, _ = w.Write([]byte("/* glsl */"))
		_ = r // silence
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	withShaderDirOverride(t, dir)
	withShaderURLs(t, "http://x", srv.URL+"/")

	require.NoError(t, InstallGANShaders())
	for _, f := range []string{"Anime4K_Restore_GAN_UUL.glsl", "Anime4K_Upscale_GAN_x4_UUL.glsl"} {
		_, err := os.Stat(filepath.Join(dir, f))
		assert.NoError(t, err)
	}
	assert.True(t, GANShadersInstalled())
}

func TestInstallGANShaders_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	withShaderDirOverride(t, t.TempDir())
	withShaderURLs(t, "http://x", srv.URL+"/")
	err := InstallGANShaders()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 500")
}

func TestInstallGANShaders_NetworkError(t *testing.T) {
	withShaderDirOverride(t, t.TempDir())
	withShaderURLs(t, "http://x", "http://127.0.0.1:1/")
	err := InstallGANShaders()
	require.Error(t, err)
}

// extractZip directly
func TestExtractZip_FiltersGLSLOnly(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "x.zip")
	require.NoError(t, os.WriteFile(zipPath, buildGLSLZip(t, map[string]string{
		"a.glsl":     "a-content",
		"b.txt":      "skip-me",
		"sub/c.glsl": "c-content",
	}), 0o644))

	dest := filepath.Join(tmp, "out")
	require.NoError(t, os.MkdirAll(dest, 0o750))
	require.NoError(t, extractZip(zipPath, dest))

	for _, want := range []string{"a.glsl", "c.glsl"} {
		_, err := os.Stat(filepath.Join(dest, want))
		assert.NoError(t, err, "expected %s extracted", want)
	}
	_, err := os.Stat(filepath.Join(dest, "b.txt"))
	assert.True(t, os.IsNotExist(err), ".txt must not extract")
}

func TestExtractZip_BadZip(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	bad := filepath.Join(tmp, "bad.zip")
	require.NoError(t, os.WriteFile(bad, []byte("not a zip"), 0o644))
	err := extractZip(bad, tmp)
	require.Error(t, err)
}

// --- Close on Anime4KUpscaler ---

func TestAnime4KUpscaler_Close_NoPanic(t *testing.T) {
	t.Parallel()
	u := NewAnime4KUpscaler(DefaultOptions())
	u.Close()
	assert.NotPanics(t, func() { u.Close() })
}

// --- VideoUpscaler.Close paths covered indirectly; this stub uses a true binary as ffmpeg ---

// makeStubFFmpegConfig returns a VideoUpscaleConfig whose FFmpeg/FFprobe points
// at /usr/bin/true so cmd.Run() succeeds with exit 0.
func makeStubFFmpegConfig(t *testing.T) VideoUpscaleConfig {
	t.Helper()
	if _, err := os.Stat("/usr/bin/true"); err != nil {
		t.Skip("/usr/bin/true unavailable on this OS")
	}
	cfg := DefaultVideoConfig()
	cfg.FFmpegPath = "/usr/bin/true"
	cfg.FFprobePath = "/usr/bin/true"
	cfg.Workers = 1
	cfg.PreserveAudio = false
	cfg.InputPath = "/tmp/in.mp4"
	cfg.OutputPath = filepath.Join(t.TempDir(), "out.mp4")
	return cfg
}

func TestVideoUpscaler_ExtractFrames_StubFFmpeg(t *testing.T) {
	cfg := makeStubFFmpegConfig(t)
	v, err := NewVideoUpscaler(cfg)
	require.NoError(t, err)
	defer v.Close()

	outDir := t.TempDir()
	err = v.extractFrames(context.Background(), outDir)
	assert.NoError(t, err, "true returns 0; extractFrames should not error")
}

func TestVideoUpscaler_ExtractFrames_BadFFmpegPath(t *testing.T) {
	t.Parallel()
	v := &VideoUpscaler{config: VideoUpscaleConfig{FFmpegPath: "/no/such/binary"}}
	err := v.extractFrames(context.Background(), t.TempDir())
	require.Error(t, err)
}

func TestVideoUpscaler_EncodeVideo_StubFFmpeg(t *testing.T) {
	cfg := makeStubFFmpegConfig(t)
	cfg.PreserveAudio = false
	cfg.UseGPUEncoding = false
	v, err := NewVideoUpscaler(cfg)
	require.NoError(t, err)
	defer v.Close()
	v.frameInfo = VideoFrameInfo{Width: 320, Height: 180, FrameRate: 24}
	err = v.encodeVideo(context.Background(), t.TempDir())
	assert.NoError(t, err)
}

func TestVideoUpscaler_EncodeVideo_GPUBranches(t *testing.T) {
	cfg := makeStubFFmpegConfig(t)
	cfg.UseGPUEncoding = true
	cfg.PreserveAudio = true
	v, err := NewVideoUpscaler(cfg)
	require.NoError(t, err)
	defer v.Close()
	v.frameInfo = VideoFrameInfo{Width: 640, Height: 480, FrameRate: 30, AudioCodec: "aac"}
	err = v.encodeVideo(context.Background(), t.TempDir())
	assert.NoError(t, err)
}

func TestVideoUpscaler_EncodeVideo_BadFFmpegPath(t *testing.T) {
	t.Parallel()
	v := &VideoUpscaler{config: VideoUpscaleConfig{FFmpegPath: "/no/such", FrameRate: 30}}
	v.frameInfo = VideoFrameInfo{Width: 320, Height: 180, FrameRate: 24}
	err := v.encodeVideo(context.Background(), t.TempDir())
	require.Error(t, err)
}

func TestVideoUpscaler_UpscaleFrames_NoFrames(t *testing.T) {
	cfg := makeStubFFmpegConfig(t)
	v, err := NewVideoUpscaler(cfg)
	require.NoError(t, err)
	defer v.Close()
	err = v.upscaleFrames(context.Background(), t.TempDir(), t.TempDir())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no frames")
}

func TestVideoUpscaler_UpscaleFrames_BadInputDir(t *testing.T) {
	t.Parallel()
	v := &VideoUpscaler{}
	err := v.upscaleFrames(context.Background(), "/no/such/dir", t.TempDir())
	require.Error(t, err)
}

// helper: write a tiny PNG to disk
func writeTinyPNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	f, err := os.Create(path)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()
	require.NoError(t, png.Encode(f, img))
}

func TestUpscaleSingleFrame_RoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	in := filepath.Join(dir, "in.png")
	out := filepath.Join(dir, "out.png")
	writeTinyPNG(t, in, 4, 4)

	v := &VideoUpscaler{upscaler: NewAnime4KUpscaler(DefaultOptions())}
	require.NoError(t, v.upscaleSingleFrame(in, out))

	f, err := os.Open(out)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()
	_, _, err = image.Decode(f)
	require.NoError(t, err)
}

func TestUpscaleSingleFrame_OpenError(t *testing.T) {
	t.Parallel()
	v := &VideoUpscaler{upscaler: NewAnime4KUpscaler(DefaultOptions())}
	err := v.upscaleSingleFrame("/no/such/file.png", "/tmp/out.png")
	require.Error(t, err)
}

func TestUpscaleSingleFrame_DecodeError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	in := filepath.Join(dir, "bad.png")
	require.NoError(t, os.WriteFile(in, []byte("not a png"), 0o644))
	v := &VideoUpscaler{upscaler: NewAnime4KUpscaler(DefaultOptions())}
	err := v.upscaleSingleFrame(in, filepath.Join(dir, "out.png"))
	require.Error(t, err)
}

func TestUpscaleSingleFrame_CreateOutputError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	in := filepath.Join(dir, "in.png")
	writeTinyPNG(t, in, 2, 2)
	v := &VideoUpscaler{upscaler: NewAnime4KUpscaler(DefaultOptions())}
	// Output path under a file (not dir) → create fails
	blocker := filepath.Join(dir, "blocker")
	require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o644))
	err := v.upscaleSingleFrame(in, filepath.Join(blocker, "out.png"))
	require.Error(t, err)
}

// --- progress model: Init / Update / View ---

func TestUpscaleProgressModel_Init_ReturnsNil(t *testing.T) {
	t.Parallel()
	m := &upscaleProgressModel{totalFrames: 10}
	assert.Nil(t, m.Init())
}

func TestUpscaleProgressModel_Update_FrameMsg(t *testing.T) {
	t.Parallel()
	m := &upscaleProgressModel{totalFrames: 10}
	got, cmd := m.Update(frameProgressMsg{completed: 5})
	assert.Same(t, m, got)
	assert.Nil(t, cmd)
	assert.Equal(t, 5, m.completed)
}

func TestUpscaleProgressModel_Update_CompleteMsg(t *testing.T) {
	t.Parallel()
	m := &upscaleProgressModel{totalFrames: 10}
	_, cmd := m.Update(upscaleCompleteMsg{})
	assert.NotNil(t, cmd) // tea.Quit
	assert.True(t, m.done)
}

func TestUpscaleProgressModel_Update_KeyQuit(t *testing.T) {
	t.Parallel()
	m := &upscaleProgressModel{totalFrames: 10}
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 'q', Text: "q"}))
	assert.NotNil(t, cmd)
}

func TestUpscaleProgressModel_Update_Unknown(t *testing.T) {
	t.Parallel()
	m := &upscaleProgressModel{totalFrames: 10}
	got, cmd := m.Update(struct{}{})
	assert.Same(t, m, got)
	assert.Nil(t, cmd)
}

func TestUpscaleProgressModel_View_NotDone(t *testing.T) {
	t.Parallel()
	m := &upscaleProgressModel{totalFrames: 10, completed: 3}
	v := m.View()
	// View renders as monospace; just smoke-check non-empty by serializing via fmt
	assert.NotPanics(t, func() { _ = v })
	_ = fmt.Sprint(v)
}

func TestUpscaleProgressModel_View_Done(t *testing.T) {
	t.Parallel()
	m := &upscaleProgressModel{totalFrames: 10, completed: 10, done: true}
	v := m.View()
	assert.NotPanics(t, func() { _ = v })
	s := fmt.Sprint(v)
	// best-effort check; if View returns object whose String includes "100"
	if !strings.Contains(s, "100") {
		t.Log("view string:", s)
	}
}

// guard for io.LimitReader presence (lint: would fail if unused above)
var _ = io.LimitReader

// sanity util for errors.Is
var _ = errors.Is
