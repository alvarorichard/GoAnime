//go:build e2e

package e2e_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/alvarorichard/Goanime/internal/util"
)

func TestCommandWorkflowsEndToEndWithoutSideEffects(t *testing.T) {
	t.Run("direct playback search route", func(t *testing.T) {
		name, err := parseCommand(t, "Attack", "on", "Titan")
		if err != nil {
			t.Fatalf("FlagParser returned error: %v", err)
		}
		if name != "attack-on-titan" {
			t.Fatalf("parsed name = %q, want %q", name, "attack-on-titan")
		}
		if util.GlobalQuality != "best" {
			t.Fatalf("GlobalQuality = %q, want best", util.GlobalQuality)
		}
	})

	t.Run("update route", func(t *testing.T) {
		_, err := parseCommand(t, "--update")
		if !errors.Is(err, util.ErrUpdateRequested) {
			t.Fatalf("error = %v, want ErrUpdateRequested", err)
		}
	})

	t.Run("anime single episode download route", func(t *testing.T) {
		outDir := t.TempDir()
		name, err := parseCommand(t,
			"-d",
			"--source", "allanime",
			"--quality", "720p",
			"-o", outDir,
			"One", "Piece", "12",
		)
		if !errors.Is(err, util.ErrDownloadRequested) {
			t.Fatalf("error = %v, want ErrDownloadRequested", err)
		}
		if name != "one-piece" {
			t.Fatalf("parsed name = %q, want one-piece", name)
		}
		req := mustDownloadRequest(t)
		if req.AnimeName != "One Piece" || req.EpisodeNum != 12 || req.IsRange || req.IsAll {
			t.Fatalf("unexpected download request: %+v", req)
		}
		if req.Source != "allanime" || req.Quality != "720p" || req.OutputDir != outDir {
			t.Fatalf("download options not preserved: %+v", req)
		}
	})

	t.Run("anime smart range download route", func(t *testing.T) {
		_, err := parseCommand(t,
			"-d",
			"-r",
			"--source", "allanime",
			"--quality", "best",
			"--allanime-smart",
			"Vinland", "Saga", "1-4",
		)
		if !errors.Is(err, util.ErrDownloadRequested) {
			t.Fatalf("error = %v, want ErrDownloadRequested", err)
		}
		req := mustDownloadRequest(t)
		if req.AnimeName != "Vinland Saga" || !req.IsRange || req.StartEpisode != 1 || req.EndEpisode != 4 {
			t.Fatalf("unexpected range request: %+v", req)
		}
		if !req.AllAnimeSmart {
			t.Fatal("AllAnimeSmart was not enabled")
		}
	})

	t.Run("anime download all route", func(t *testing.T) {
		_, err := parseCommand(t, "-d", "-a", "One", "Piece")
		if !errors.Is(err, util.ErrDownloadRequested) {
			t.Fatalf("error = %v, want ErrDownloadRequested", err)
		}
		req := mustDownloadRequest(t)
		if req.AnimeName != "One Piece" || !req.IsAll {
			t.Fatalf("unexpected download-all request: %+v", req)
		}
	})

	t.Run("movie download route", func(t *testing.T) {
		_, err := parseCommand(t,
			"-dm",
			"--quality", "1080p",
			"--subs", "portuguese",
			"Spider", "Man",
		)
		if !errors.Is(err, util.ErrMovieDownloadRequested) {
			t.Fatalf("error = %v, want ErrMovieDownloadRequested", err)
		}
		req := mustDownloadRequest(t)
		if req.AnimeName != "Spider Man" || !req.IsMovie || req.IsTV || req.Quality != "1080p" || req.SubsLanguage != "portuguese" {
			t.Fatalf("unexpected movie request: %+v", req)
		}
	})

	t.Run("tv single episode download route", func(t *testing.T) {
		_, err := parseCommand(t, "-dm", "--type", "tv", "Breaking", "Bad", "1", "2")
		if !errors.Is(err, util.ErrMovieDownloadRequested) {
			t.Fatalf("error = %v, want ErrMovieDownloadRequested", err)
		}
		req := mustDownloadRequest(t)
		if req.AnimeName != "Breaking Bad" || !req.IsTV || req.SeasonNum != 1 || req.EpisodeNum != 2 || req.IsRange {
			t.Fatalf("unexpected tv request: %+v", req)
		}
	})

	t.Run("tv range download route", func(t *testing.T) {
		_, err := parseCommand(t, "-dm", "-r", "Breaking", "Bad", "2", "1-3")
		if !errors.Is(err, util.ErrMovieDownloadRequested) {
			t.Fatalf("error = %v, want ErrMovieDownloadRequested", err)
		}
		req := mustDownloadRequest(t)
		if req.AnimeName != "Breaking Bad" || !req.IsTV || !req.IsRange || req.SeasonNum != 2 || req.StartEpisode != 1 || req.EndEpisode != 3 {
			t.Fatalf("unexpected tv range request: %+v", req)
		}
	})

	t.Run("tv download all route", func(t *testing.T) {
		_, err := parseCommand(t, "-dm", "-a", "Breaking", "Bad")
		if !errors.Is(err, util.ErrMovieDownloadRequested) {
			t.Fatalf("error = %v, want ErrMovieDownloadRequested", err)
		}
		req := mustDownloadRequest(t)
		if req.AnimeName != "Breaking Bad" || !req.IsTV || !req.IsAll {
			t.Fatalf("unexpected tv download-all request: %+v", req)
		}
	})

	t.Run("playback media preferences route", func(t *testing.T) {
		name, err := parseCommand(t,
			"--type", "movie",
			"--subs", "portuguese",
			"--audio", "pt-BR,english",
			"--no-subs",
			"Matrix",
		)
		if err != nil {
			t.Fatalf("FlagParser returned error: %v", err)
		}
		if name != "matrix" {
			t.Fatalf("parsed name = %q, want matrix", name)
		}
		if util.GlobalMediaType != "movie" || util.GlobalSubsLanguage != "portuguese" || util.GlobalAudioLanguage != "pt-BR,english" || !util.GlobalNoSubs {
			t.Fatalf("media preferences not preserved: type=%q subs=%q audio=%q noSubs=%v",
				util.GlobalMediaType, util.GlobalSubsLanguage, util.GlobalAudioLanguage, util.GlobalNoSubs)
		}
	})

	t.Run("image upscale route", func(t *testing.T) {
		tmp := t.TempDir()
		input := filepath.Join(tmp, "frame.png")
		output := filepath.Join(tmp, "frame-upscaled.png")
		if err := os.WriteFile(input, []byte("placeholder"), 0o600); err != nil {
			t.Fatalf("failed to create input file: %v", err)
		}

		name, err := parseCommand(t,
			"--upscale",
			"--upscale-output", output,
			"--upscale-scale", "4",
			"--upscale-passes", "3",
			"--upscale-fast",
			"--upscale-workers", "2",
			input,
		)
		if !errors.Is(err, util.ErrUpscaleRequested) {
			t.Fatalf("error = %v, want ErrUpscaleRequested", err)
		}
		if name != input {
			t.Fatalf("parsed input = %q, want %q", name, input)
		}
		req := util.GlobalUpscaleRequest
		if req == nil {
			t.Fatal("GlobalUpscaleRequest is nil")
		}
		if req.InputPath != input || req.OutputPath != output || req.ScaleFactor != 4 || req.Passes != 3 || !req.FastMode || req.Workers != 2 {
			t.Fatalf("unexpected upscale request: %+v", req)
		}
	})
}

func parseCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()

	oldArgs := os.Args
	resetUtilGlobals()
	os.Args = append([]string{"goanime"}, args...)

	t.Cleanup(func() {
		os.Args = oldArgs
		resetUtilGlobals()
	})

	return util.FlagParser()
}

func mustDownloadRequest(t *testing.T) *util.DownloadRequest {
	t.Helper()

	if util.GlobalDownloadRequest == nil {
		t.Fatal("GlobalDownloadRequest is nil")
	}
	return util.GlobalDownloadRequest
}

func resetUtilGlobals() {
	util.IsDebug = false
	util.PerfEnabled = false
	util.GlobalSource = ""
	util.GlobalQuality = ""
	util.GlobalMediaType = ""
	util.GlobalSubsLanguage = ""
	util.GlobalAudioLanguage = ""
	util.GlobalSubtitles = nil
	util.GlobalNoSubs = false
	util.GlobalReferer = ""
	util.GlobalOutputDir = ""
	util.GlobalAnimeSource = ""
	util.GlobalDownloadRequest = nil
	util.GlobalUpscaleRequest = nil
}
