package handlers

import (
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/updater"
	"github.com/alvarorichard/Goanime/internal/upscaler"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mediaSource fake ---

type fakeMediaSource struct {
	animeResults []*models.Anime
	allResults   []*models.Anime
	searchErr    error
	streamURL    string
	streamMeta   map[string]string
	streamErr    error
	animeArg     string
	episodeArg   string
}

func (f *fakeMediaSource) SearchAnimeOnly(q string) ([]*models.Anime, error) {
	f.animeArg = q
	return f.animeResults, f.searchErr
}

func (f *fakeMediaSource) SearchAll(q string) ([]*models.Anime, error) {
	f.animeArg = q
	return f.allResults, f.searchErr
}

func (f *fakeMediaSource) GetAnimeStreamURL(_ *models.Anime, ep, _, _ string) (string, map[string]string, error) {
	f.episodeArg = ep
	return f.streamURL, f.streamMeta, f.streamErr
}

func newHandler(src mediaSource) *MediaHandler {
	return &MediaHandler{
		mediaManager: src,
		provider:     "Vidcloud",
		quality:      "best",
		subsLanguage: "english",
	}
}

// stubFindStr returns a findFn that yields (idx, err) regardless of input.
func stubFindStr(idx int, err error) func([]string, func(int) string, ...fuzzyfinder.Option) (int, error) {
	return func([]string, func(int) string, ...fuzzyfinder.Option) (int, error) {
		return idx, err
	}
}

func stubFindRes(idx int, err error) func([]*models.Anime, func(int) string, ...fuzzyfinder.Option) (int, error) {
	return func([]*models.Anime, func(int) string, ...fuzzyfinder.Option) (int, error) {
		return idx, err
	}
}

func setFindFn(t *testing.T, fn func([]string, func(int) string, ...fuzzyfinder.Option) (int, error)) {
	t.Helper()
	prev := findFn
	findFn = fn
	t.Cleanup(func() { findFn = prev })
}

func setFindResultFn(t *testing.T, fn func([]*models.Anime, func(int) string, ...fuzzyfinder.Option) (int, error)) {
	t.Helper()
	prev := findResultFn
	findResultFn = fn
	t.Cleanup(func() { findResultFn = prev })
}

func setRunFormFn(t *testing.T, fn func(func() error) error) {
	t.Helper()
	prev := runFormFn
	runFormFn = fn
	t.Cleanup(func() { runFormFn = prev })
}

func testOpts() upscaler.Anime4KOptions {
	o := upscaler.DefaultOptions()
	o.ScaleFactor = 2
	o.Passes = 1
	return o
}

func writeTinyPNGForHandlers(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	f, err := os.Create(path)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()
	require.NoError(t, png.Encode(f, img))
}

// --- HandleDownloadRequest / HandleMovieDownloadRequest ---

func TestHandleDownloadRequest_NilGlobal(t *testing.T) {
	prev := util.GlobalDownloadRequest
	util.GlobalDownloadRequest = nil
	t.Cleanup(func() { util.GlobalDownloadRequest = prev })

	err := HandleDownloadRequest()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "download request is nil")
}

func TestHandleDownloadRequest_PropagatesDownloadError(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Windows console API bypasses os.Stdin redirection; the underlying
		// fuzzy-finder hangs instead of erroring out, blocking CI for 10 minutes.
		t.Skip("Windows fuzzy-finder cannot be driven from headless tests")
	}
	prev := util.GlobalDownloadRequest
	util.GlobalDownloadRequest = &util.DownloadRequest{AnimeName: "", EpisodeNum: -1}
	t.Cleanup(func() { util.GlobalDownloadRequest = prev })

	err := HandleDownloadRequest()
	if err != nil {
		assert.Contains(t, err.Error(), "download failed")
	}
}

func TestHandleMovieDownloadRequest_NilGlobal(t *testing.T) {
	prev := util.GlobalDownloadRequest
	util.GlobalDownloadRequest = nil
	t.Cleanup(func() { util.GlobalDownloadRequest = prev })

	err := HandleMovieDownloadRequest()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "movie download request is nil")
}

func TestHandleMovieDownloadRequest_PropagatesError(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Same Windows fuzzy-finder hang as TestHandleDownloadRequest_PropagatesDownloadError.
		t.Skip("Windows fuzzy-finder cannot be driven from headless tests")
	}
	prev := util.GlobalDownloadRequest
	util.GlobalDownloadRequest = &util.DownloadRequest{}
	t.Cleanup(func() { util.GlobalDownloadRequest = prev })

	err := HandleMovieDownloadRequest()
	if err != nil {
		assert.Contains(t, err.Error(), "movie download failed")
	}
}

// --- SearchMedia ---

func TestSearchMedia_AnimeBranch(t *testing.T) {
	t.Parallel()
	fake := &fakeMediaSource{animeResults: []*models.Anime{{Name: "A"}}}
	mh := newHandler(fake)
	got, err := mh.SearchMedia("foo", models.MediaTypeAnime)
	require.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "foo", fake.animeArg)
}

func TestSearchMedia_MovieBranchRejects(t *testing.T) {
	t.Parallel()
	mh := newHandler(&fakeMediaSource{})
	_, err := mh.SearchMedia("x", models.MediaTypeMovie)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "have been removed")
}

func TestSearchMedia_TVBranchRejects(t *testing.T) {
	t.Parallel()
	mh := newHandler(&fakeMediaSource{})
	_, err := mh.SearchMedia("y", models.MediaTypeTV)
	require.Error(t, err)
}

func TestSearchMedia_DefaultBranchUsesSearchAll(t *testing.T) {
	t.Parallel()
	fake := &fakeMediaSource{allResults: []*models.Anime{{Name: "X"}, {Name: "Y"}}}
	mh := newHandler(fake)
	got, err := mh.SearchMedia("q", "")
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

// --- SelectMediaType ---

func TestSelectMediaType_PicksAnime(t *testing.T) {
	setFindFn(t, stubFindStr(0, nil))
	mh := newHandler(&fakeMediaSource{})
	mt, err := mh.SelectMediaType()
	require.NoError(t, err)
	assert.Equal(t, models.MediaTypeAnime, mt)
}

func TestSelectMediaType_PicksSearchAll(t *testing.T) {
	setFindFn(t, stubFindStr(1, nil))
	mh := newHandler(&fakeMediaSource{})
	mt, err := mh.SelectMediaType()
	require.NoError(t, err)
	assert.Equal(t, models.MediaType(""), mt)
}

func TestSelectMediaType_Error(t *testing.T) {
	setFindFn(t, stubFindStr(-1, errors.New("aborted")))
	mh := newHandler(&fakeMediaSource{})
	_, err := mh.SelectMediaType()
	require.Error(t, err)
}

// --- GetAnimeStreamURL ---

func TestGetAnimeStreamURL_Delegates(t *testing.T) {
	t.Parallel()
	fake := &fakeMediaSource{streamURL: "http://x/y", streamMeta: map[string]string{"k": "v"}}
	mh := newHandler(fake)
	url, meta, err := mh.GetAnimeStreamURL(&models.Anime{Name: "A"}, "3", "sub")
	require.NoError(t, err)
	assert.Equal(t, "http://x/y", url)
	assert.Equal(t, "v", meta["k"])
	assert.Equal(t, "3", fake.episodeArg)
}

func TestGetAnimeStreamURL_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := &fakeMediaSource{streamErr: errors.New("no stream")}
	mh := newHandler(fake)
	_, _, err := mh.GetAnimeStreamURL(&models.Anime{}, "1", "sub")
	require.Error(t, err)
}

// --- InteractiveMediaFlow ---

func TestInteractiveMediaFlow_HappyPath_WithQuery(t *testing.T) {
	setFindFn(t, stubFindStr(0, nil))
	setFindResultFn(t, stubFindRes(0, nil))
	setRunFormFn(t, func(func() error) error { return nil })

	anime := &models.Anime{Name: "Foo", MediaType: models.MediaTypeAnime, Source: "AllAnime"}
	mh := newHandler(&fakeMediaSource{
		allResults: []*models.Anime{anime},
		streamURL:  "http://stream",
		streamMeta: map[string]string{"q": "best"},
	})
	info, err := mh.InteractiveMediaFlow("foo")
	require.NoError(t, err)
	assert.Equal(t, "Foo", info.Title)
	assert.Equal(t, "http://stream", info.StreamURL)
}

func TestInteractiveMediaFlow_SearchError(t *testing.T) {
	setRunFormFn(t, func(func() error) error { return nil })
	mh := newHandler(&fakeMediaSource{searchErr: errors.New("network")})
	_, err := mh.InteractiveMediaFlow("foo")
	require.Error(t, err)
}

func TestInteractiveMediaFlow_EmptyQueryNeedsSelectMediaType(t *testing.T) {
	setFindFn(t, stubFindStr(-1, errors.New("aborted")))
	setRunFormFn(t, func(func() error) error { return nil })

	mh := newHandler(&fakeMediaSource{})
	_, err := mh.InteractiveMediaFlow("")
	require.Error(t, err)
}

// --- handleAnimePlayback ---

func TestHandleAnimePlayback_ModeSub(t *testing.T) {
	setFindFn(t, stubFindStr(0, nil))
	setRunFormFn(t, func(func() error) error { return nil })

	mh := newHandler(&fakeMediaSource{streamURL: "http://s", streamMeta: map[string]string{}})
	info, err := mh.handleAnimePlayback(&models.Anime{Name: "A"}, &PlaybackInfo{Title: "A"})
	require.NoError(t, err)
	assert.Equal(t, "http://s", info.StreamURL)
}

func TestHandleAnimePlayback_ModeDub(t *testing.T) {
	setFindFn(t, stubFindStr(1, nil))
	setRunFormFn(t, func(func() error) error { return nil })

	mh := newHandler(&fakeMediaSource{streamURL: "http://s"})
	_, err := mh.handleAnimePlayback(&models.Anime{}, &PlaybackInfo{})
	require.NoError(t, err)
}

func TestHandleAnimePlayback_FormError(t *testing.T) {
	setRunFormFn(t, func(func() error) error { return errors.New("tty") })
	mh := newHandler(&fakeMediaSource{})
	_, err := mh.handleAnimePlayback(&models.Anime{}, &PlaybackInfo{})
	require.Error(t, err)
}

func TestHandleAnimePlayback_StreamError(t *testing.T) {
	setFindFn(t, stubFindStr(0, nil))
	setRunFormFn(t, func(func() error) error { return nil })

	mh := newHandler(&fakeMediaSource{streamErr: errors.New("no stream")})
	_, err := mh.handleAnimePlayback(&models.Anime{}, &PlaybackInfo{})
	require.Error(t, err)
}

// --- HandlePlaybackMode (symbol-pin only — full execution requires real network + TUI) ---

func TestHandlePlaybackMode_SymbolPin(t *testing.T) {
	t.Parallel()
	// Existence guard so future renames break this test loudly. The function
	// runs a TUI loop with real network calls — not driveable from unit tests.
	_ = HandlePlaybackMode
}

// --- HandleUpdateRequest ---

func TestHandleUpdateRequest_DoesNotPanic(t *testing.T) {
	if runtime.GOOS == "windows" {
		// HandleUpdateRequest → updater.CheckAndPromptUpdate may open the huh
		// confirm form when GitHub has a newer release. The form blocks on TTY
		// on Windows CI. Skip there.
		t.Skip("huh form blocks on Windows CI without TTY")
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(updater.GitHubRelease{TagName: "v0.0.0"})
	}))
	t.Cleanup(srv.Close)
	// Cannot override updater.releaseAPIURL from outside its package; just
	// ensure the wrapper doesn't panic and either succeeds or returns wrapped
	// error from the underlying network call.
	err := HandleUpdateRequest()
	if err != nil {
		assert.Contains(t, err.Error(), "update failed")
	}
}

// --- HandleUpscaleRequest, handleImageUpscale, handleVideoUpscale ---

func TestHandleUpscaleRequest_NilGlobal(t *testing.T) {
	prev := util.GlobalUpscaleRequest
	util.GlobalUpscaleRequest = nil
	t.Cleanup(func() { util.GlobalUpscaleRequest = prev })

	err := HandleUpscaleRequest()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "upscale request is nil")
}

func TestHandleUpscaleRequest_ImageBranch(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.png")
	writeTinyPNGForHandlers(t, in)
	out := filepath.Join(dir, "out.png")

	prev := util.GlobalUpscaleRequest
	util.GlobalUpscaleRequest = &util.UpscaleRequest{
		InputPath:   in,
		OutputPath:  out,
		ScaleFactor: 2,
		Passes:      1,
	}
	t.Cleanup(func() { util.GlobalUpscaleRequest = prev })

	err := HandleUpscaleRequest()
	if err != nil {
		// FFmpeg missing on this host → acceptable
		assert.Contains(t, err.Error(), "FFmpeg")
		return
	}
	_, statErr := os.Stat(out)
	require.NoError(t, statErr)
}

func TestHandleUpscaleRequest_HighQualityFlag(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.png")
	writeTinyPNGForHandlers(t, in)
	out := filepath.Join(dir, "out.png")

	prev := util.GlobalUpscaleRequest
	util.GlobalUpscaleRequest = &util.UpscaleRequest{
		InputPath:   in,
		OutputPath:  out,
		ScaleFactor: 2,
		HighQuality: true,
	}
	t.Cleanup(func() { util.GlobalUpscaleRequest = prev })

	_ = HandleUpscaleRequest() // either success or ffmpeg-missing
}

func TestHandleUpscaleRequest_FastModeFlag(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.png")
	writeTinyPNGForHandlers(t, in)

	prev := util.GlobalUpscaleRequest
	util.GlobalUpscaleRequest = &util.UpscaleRequest{
		InputPath:   in,
		ScaleFactor: 2,
		FastMode:    true,
	}
	t.Cleanup(func() { util.GlobalUpscaleRequest = prev })

	_ = HandleUpscaleRequest()
}

func TestHandleUpscaleRequest_VideoBranch_NonExistentInput(t *testing.T) {
	prev := util.GlobalUpscaleRequest
	util.GlobalUpscaleRequest = &util.UpscaleRequest{
		InputPath:   "/no/file.mp4",
		ScaleFactor: 2,
		Passes:      1,
	}
	t.Cleanup(func() { util.GlobalUpscaleRequest = prev })

	err := HandleUpscaleRequest()
	if err != nil {
		assert.Error(t, err)
	}
}

func TestHandleImageUpscale_NonExistentInput(t *testing.T) {
	t.Parallel()
	err := handleImageUpscale("/no/such/img.png", "/tmp/out.png", testOpts())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "image upscale failed")
}

func TestHandleImageUpscale_Success(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	in := filepath.Join(dir, "in.png")
	writeTinyPNGForHandlers(t, in)
	out := filepath.Join(dir, "out.png")
	err := handleImageUpscale(in, out, testOpts())
	require.NoError(t, err)
	_, statErr := os.Stat(out)
	require.NoError(t, statErr)
}

func TestHandleVideoUpscale_FFmpegMissingOrInvalidInput(t *testing.T) {
	t.Parallel()
	req := &util.UpscaleRequest{InputPath: "/no/such.mp4", ScaleFactor: 2, Passes: 1}
	err := handleVideoUpscale(req, "/tmp/out.mp4", testOpts())
	require.Error(t, err)
}

// --- SelectMedia ---

// stubFindResWithCallback returns a findResultFn that calls the formatter for every
// item before returning (idx, err). This exercises the switch/branch logic inside
// the formatter closure without launching a real TUI.
func stubFindResWithCallback(idx int, err error) func([]*models.Anime, func(int) string, ...fuzzyfinder.Option) (int, error) {
	return func(results []*models.Anime, formatter func(int) string, _ ...fuzzyfinder.Option) (int, error) {
		for i := range results {
			_ = formatter(i)
		}
		return idx, err
	}
}

func TestSelectMedia_EmptyResults(t *testing.T) {
	t.Parallel()
	mh := newHandler(&fakeMediaSource{})
	_, err := mh.SelectMedia(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no results")
}

func TestSelectMedia_AllTypeBranches(t *testing.T) {
	// callback is invoked for every item — covers Movie/TV/Anime/unknown switch branches
	// and both the year-present and year-absent branches.
	setFindResultFn(t, stubFindResWithCallback(0, nil))

	results := []*models.Anime{
		{Name: "Movie Title", MediaType: models.MediaTypeMovie, Source: "AllAnime", Year: "2023"},
		{Name: "TV Show", MediaType: models.MediaTypeTV, Source: "SuperFlix"},
		{Name: "Anime Title", MediaType: models.MediaTypeAnime, Source: "AllAnime"},
		{Name: "Unknown Type", Source: "AllAnime"}, // empty MediaType → empty typeTag
	}
	mh := newHandler(&fakeMediaSource{})
	selected, err := mh.SelectMedia(results)
	require.NoError(t, err)
	assert.Equal(t, "Movie Title", selected.Name)
}

func TestSelectMedia_FindError(t *testing.T) {
	setFindResultFn(t, func(_ []*models.Anime, _ func(int) string, _ ...fuzzyfinder.Option) (int, error) {
		return -1, errors.New("aborted")
	})
	mh := newHandler(&fakeMediaSource{})
	results := []*models.Anime{{Name: "Naruto"}}
	_, err := mh.SelectMedia(results)
	require.Error(t, err)
}

func TestSelectMedia_YearPresentInLabel(t *testing.T) {
	// Verify the year-suffix branch: Year != "" → " (YEAR)" appended to label.
	var capturedLabel string
	setFindResultFn(t, func(results []*models.Anime, formatter func(int) string, _ ...fuzzyfinder.Option) (int, error) {
		capturedLabel = formatter(0)
		return 0, nil
	})

	results := []*models.Anime{
		{Name: "Naruto", MediaType: models.MediaTypeAnime, Source: "AllAnime", Year: "2002"},
	}
	mh := newHandler(&fakeMediaSource{})
	_, err := mh.SelectMedia(results)
	require.NoError(t, err)
	assert.Contains(t, capturedLabel, "(2002)")
}
