package util

import "testing"

func resetRuntimeState() {
	SetGlobalSource("")
	SetGlobalQuality("")
	SetGlobalMediaType("")
	SetPreferredSubtitleLanguage("")
	SetPreferredAudioLanguage("")
	SetSubtitlesDisabled(false)
	SetGlobalOutputDir("")
	ClearGlobalDownloadRequest()
	ClearGlobalUpscaleRequest()
}

func TestCurrentSessionConfig(t *testing.T) {
	t.Cleanup(resetRuntimeState)

	SetGlobalSource("goyabu")
	SetGlobalQuality("720p")
	SetGlobalMediaType("anime")
	SetPreferredSubtitleLanguage("por")
	SetPreferredAudioLanguage("jpn")
	SetSubtitlesDisabled(true)
	SetGlobalOutputDir("C:\\downloads")

	cfg := CurrentSessionConfig()
	if cfg.Source != "goyabu" || cfg.Quality != "720p" || cfg.MediaType != "anime" {
		t.Fatalf("unexpected config snapshot: %+v", cfg)
	}
	if cfg.SubsLanguage != "por" || cfg.AudioLanguage != "jpn" {
		t.Fatalf("unexpected language snapshot: %+v", cfg)
	}
	if !cfg.NoSubs || cfg.OutputDir != "C:\\downloads" {
		t.Fatalf("unexpected output/no-subs snapshot: %+v", cfg)
	}
}

func TestCurrentDownloadRequestReturnsCopy(t *testing.T) {
	t.Cleanup(resetRuntimeState)

	SetGlobalDownloadRequest(&DownloadRequest{
		AnimeName:     "Naruto",
		SeasonNum:     2,
		EpisodeNum:    7,
		AllAnimeSmart: true,
	})

	req := CurrentDownloadRequest()
	if req == nil {
		t.Fatal("CurrentDownloadRequest() returned nil")
	}
	req.AnimeName = "Mutated"

	current := CurrentDownloadRequest()
	if current.AnimeName != "Naruto" {
		t.Fatalf("stored request was mutated through snapshot: %+v", current)
	}
}

func TestCurrentUpscaleRequestReturnsCopy(t *testing.T) {
	t.Cleanup(resetRuntimeState)

	SetGlobalUpscaleRequest(&UpscaleRequest{
		InputPath:   "input.mp4",
		OutputPath:  "output.mp4",
		ScaleFactor: 2,
	})

	req := CurrentUpscaleRequest()
	if req == nil {
		t.Fatal("CurrentUpscaleRequest() returned nil")
	}
	req.OutputPath = "mutated.mp4"

	current := CurrentUpscaleRequest()
	if current.OutputPath != "output.mp4" {
		t.Fatalf("stored upscale request was mutated through snapshot: %+v", current)
	}
}
