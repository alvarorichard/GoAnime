package handlers

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/upscaler"
	"github.com/alvarorichard/Goanime/internal/util"
)

func TestHandleDownloadRequestDelegatesGlobalRequest(t *testing.T) {
	restoreDownloadHandlerState(t)

	wantReq := &util.DownloadRequest{AnimeName: "frieren", EpisodeNum: 7}
	util.GlobalDownloadRequest = wantReq

	var gotReq *util.DownloadRequest
	runAnimeDownload = func(req *util.DownloadRequest) error {
		gotReq = req
		return nil
	}

	if err := HandleDownloadRequest(); err != nil {
		t.Fatalf("HandleDownloadRequest() error = %v", err)
	}
	if gotReq != wantReq {
		t.Fatalf("download runner received %#v, want original request %#v", gotReq, wantReq)
	}
}

func TestHandleDownloadRequestRejectsNilRequest(t *testing.T) {
	restoreDownloadHandlerState(t)

	var called bool
	util.GlobalDownloadRequest = nil
	runAnimeDownload = func(req *util.DownloadRequest) error {
		called = true
		return nil
	}

	err := HandleDownloadRequest()
	if err == nil || !strings.Contains(err.Error(), "download request is nil") {
		t.Fatalf("HandleDownloadRequest() error = %v, want nil request error", err)
	}
	if called {
		t.Fatal("download runner must not be called when global request is nil")
	}
}

func TestHandleDownloadRequestWrapsRunnerError(t *testing.T) {
	restoreDownloadHandlerState(t)

	sentinel := errors.New("network blocked")
	util.GlobalDownloadRequest = &util.DownloadRequest{AnimeName: "frieren", EpisodeNum: 1}
	runAnimeDownload = func(req *util.DownloadRequest) error {
		return sentinel
	}

	err := HandleDownloadRequest()
	if !errors.Is(err, sentinel) {
		t.Fatalf("HandleDownloadRequest() error = %v, want wrapped sentinel", err)
	}
	if !strings.Contains(err.Error(), "download failed") {
		t.Fatalf("HandleDownloadRequest() error = %v, want context", err)
	}
}

func TestHandleMovieDownloadRequestDelegatesGlobalRequest(t *testing.T) {
	restoreDownloadHandlerState(t)

	wantReq := &util.DownloadRequest{AnimeName: "inception", IsMovie: true, Quality: "1080"}
	util.GlobalDownloadRequest = wantReq

	var gotReq *util.DownloadRequest
	runMovieDownload = func(req *util.DownloadRequest) error {
		gotReq = req
		return nil
	}

	if err := HandleMovieDownloadRequest(); err != nil {
		t.Fatalf("HandleMovieDownloadRequest() error = %v", err)
	}
	if gotReq != wantReq {
		t.Fatalf("movie download runner received %#v, want original request %#v", gotReq, wantReq)
	}
}

func TestHandleUpdateRequestDelegatesAndWrapsError(t *testing.T) {
	original := checkAndPromptUpdate
	t.Cleanup(func() { checkAndPromptUpdate = original })

	sentinel := errors.New("updater offline")
	checkAndPromptUpdate = func() error { return sentinel }

	err := HandleUpdateRequest()
	if !errors.Is(err, sentinel) {
		t.Fatalf("HandleUpdateRequest() error = %v, want wrapped sentinel", err)
	}
	if !strings.Contains(err.Error(), "update failed") {
		t.Fatalf("HandleUpdateRequest() error = %v, want context", err)
	}
}

func TestHandleUpscaleRequestImageDefaultsOutputAndOptions(t *testing.T) {
	restoreUpscaleHandlerState(t)

	inputPath := filepath.Join(t.TempDir(), "frame.PNG")
	util.GlobalUpscaleRequest = &util.UpscaleRequest{
		InputPath:        inputPath,
		ScaleFactor:      3,
		Passes:           5,
		StrengthColor:    0.25,
		StrengthGradient: 0.75,
	}

	validateFFmpeg = func() (string, error) {
		return "ffmpeg test", nil
	}

	var gotInput, gotOutput string
	var gotOpts upscaler.Anime4KOptions
	upscaleImageFile = func(inputPath, outputPath string, opts upscaler.Anime4KOptions) error {
		gotInput = inputPath
		gotOutput = outputPath
		gotOpts = opts
		return nil
	}

	if err := HandleUpscaleRequest(); err != nil {
		t.Fatalf("HandleUpscaleRequest() error = %v", err)
	}
	if gotInput != inputPath {
		t.Fatalf("upscale input = %q, want %q", gotInput, inputPath)
	}
	wantOutput := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + "_upscaled" + filepath.Ext(inputPath)
	if gotOutput != wantOutput {
		t.Fatalf("upscale output = %q, want %q", gotOutput, wantOutput)
	}
	if gotOpts.ScaleFactor != 3 || gotOpts.Passes != 5 || gotOpts.StrengthColor != 0.25 || gotOpts.StrengthGradient != 0.75 {
		t.Fatalf("upscale options = %+v, want request values propagated", gotOpts)
	}
}

func TestHandleUpscaleRequestStopsWhenFFmpegValidationFails(t *testing.T) {
	restoreUpscaleHandlerState(t)

	sentinel := errors.New("ffmpeg missing")
	util.GlobalUpscaleRequest = &util.UpscaleRequest{InputPath: "frame.png"}
	validateFFmpeg = func() (string, error) {
		return "", sentinel
	}
	upscaleImageFile = func(inputPath, outputPath string, opts upscaler.Anime4KOptions) error {
		t.Fatal("upscaleImageFile must not run when FFmpeg validation fails")
		return nil
	}

	err := HandleUpscaleRequest()
	if !errors.Is(err, sentinel) {
		t.Fatalf("HandleUpscaleRequest() error = %v, want wrapped sentinel", err)
	}
	if !strings.Contains(err.Error(), "FFmpeg validation failed") {
		t.Fatalf("HandleUpscaleRequest() error = %v, want validation context", err)
	}
}

func TestIsImageExtension(t *testing.T) {
	tests := []struct {
		ext  string
		want bool
	}{
		{ext: ".png", want: true},
		{ext: ".webp", want: true},
		{ext: ".mp4", want: false},
		{ext: "", want: false},
	}

	for _, tt := range tests {
		if got := isImageExtension(tt.ext); got != tt.want {
			t.Fatalf("isImageExtension(%q) = %v, want %v", tt.ext, got, tt.want)
		}
	}
}

func restoreDownloadHandlerState(t *testing.T) {
	t.Helper()

	originalRequest := util.GlobalDownloadRequest
	originalAnimeRunner := runAnimeDownload
	originalMovieRunner := runMovieDownload
	t.Cleanup(func() {
		util.GlobalDownloadRequest = originalRequest
		runAnimeDownload = originalAnimeRunner
		runMovieDownload = originalMovieRunner
	})
}

func restoreUpscaleHandlerState(t *testing.T) {
	t.Helper()

	originalRequest := util.GlobalUpscaleRequest
	originalValidate := validateFFmpeg
	originalUpscaleImage := upscaleImageFile
	t.Cleanup(func() {
		util.GlobalUpscaleRequest = originalRequest
		validateFFmpeg = originalValidate
		upscaleImageFile = originalUpscaleImage
	})
}
