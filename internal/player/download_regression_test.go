package player

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"charm.land/log/v2"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadDirectHTTPWithClientDownloadsMockVideoAndTracksProgress(t *testing.T) {
	payload := bytes.Repeat([]byte("goanime-video-payload"), 32*1024)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/episode.mp4" {
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	home := t.TempDir()
	t.Setenv("HOME", home)
	outPath := filepath.Join(home, "downloads", "episode.mp4")
	m := &model{}

	err := downloadDirectHTTPWithClient(server.URL+"/episode.mp4", outPath, m, server.Client())
	require.NoError(t, err)

	got, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Equal(t, payload, got)
	assert.Equal(t, int64(len(payload)), m.progressTotal())

	m.mu.Lock()
	received := m.received
	m.mu.Unlock()
	assert.Equal(t, int64(len(payload)), received)
}

func TestDownloadDirectHTTPWithClientReturnsHTTPStatusErrorFromMockCDN(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "missing object", http.StatusNotFound)
	}))
	defer server.Close()

	home := t.TempDir()
	t.Setenv("HOME", home)
	outPath := filepath.Join(home, "downloads", "episode.mp4")

	err := downloadDirectHTTPWithClient(server.URL+"/missing.mp4", outPath, &model{}, server.Client())
	require.Error(t, err)
	assert.True(t, isHTTPStatusError(err, http.StatusNotFound), "error should be recognized as HTTP 404: %v", err)

	_, statErr := os.Stat(outPath)
	assert.True(t, os.IsNotExist(statErr), "404 response must not create a completed file")
}

func TestAnimeFireFallbackUsesRealHTTPDownloaderAfter404(t *testing.T) {
	payload := bytes.Repeat([]byte("animefire-fallback-video"), 32*1024)

	var (
		mu       sync.Mutex
		requests []string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests = append(requests, r.URL.Path)
		mu.Unlock()

		switch r.URL.Path {
		case "/hd/20.mp4":
			http.Error(w, "404 Not Found", http.StatusNotFound)
		case "/fhd/20.mp4":
			w.Header().Set("Content-Type", "video/mp4")
			w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
			_, _ = w.Write(payload)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	home := t.TempDir()
	t.Setenv("HOME", home)
	outPath := filepath.Join(home, "downloads", "episode-20.mp4")
	progressModel := &model{}

	primaryURL := server.URL + "/hd/20.mp4"
	fallbackURL := server.URL + "/fhd/20.mp4"
	err := runAnimeFireDirectDownloadWithFallback(
		"https://animefire.io/video/jujutsu-kaisen-2nd-season-dublado/20",
		primaryURL,
		outPath,
		progressModel,
		func(url, path string, m *model) error {
			return downloadDirectHTTPWithClient(url, path, m, server.Client())
		},
		func(apiURL, failedURL string) (string, error) {
			assert.Equal(t, "https://animefire.io/video/jujutsu-kaisen-2nd-season-dublado/20", apiURL)
			assert.Equal(t, primaryURL, failedURL)
			return fallbackURL, nil
		},
	)
	require.NoError(t, err)

	got, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Equal(t, payload, got)
	assert.Equal(t, int64(len(payload)), progressModel.progressTotal())

	progressModel.mu.Lock()
	received := progressModel.received
	progressModel.mu.Unlock()
	assert.Equal(t, int64(len(payload)), received)

	mu.Lock()
	gotRequests := append([]string(nil), requests...)
	mu.Unlock()
	assert.Equal(t, []string{"/hd/20.mp4", "/fhd/20.mp4"}, gotRequests)
}

func TestHandleBatchDownloadRangeReturnsBatchErrorForAnimeFireNoStream(t *testing.T) {
	outputDir := t.TempDir()
	restore := installDownloadRangeTestState(outputDir)
	defer restore()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><h1>episode page without any playable source</h1></body></html>`))
	}))
	defer server.Close()

	SetAnimeName("JUJUTSU KAISEN Season 2", 2)
	SetExactMediaType(string(models.MediaTypeAnime))
	SetMediaMeta(&util.MediaMeta{Year: "2023", AnilistID: 145064, MalID: 51009})

	anime := &models.Anime{
		Name:      "JUJUTSU KAISEN Season 2",
		URL:       server.URL + "/anime/jujutsu-kaisen-2",
		Source:    "Animefire.io",
		MediaType: models.MediaTypeAnime,
	}
	episodes := []models.Episode{
		{Number: "1", Num: 1, URL: server.URL + "/episodio-1"},
	}

	err := HandleBatchDownloadRange(episodes, anime, 1, 1)
	require.Error(t, err)

	var batchErr batchDownloadError
	require.ErrorAs(t, err, &batchErr)
	require.Len(t, batchErr.Failures, 1)
	assert.Equal(t, 1, batchErr.Failures[0].Episode)
	assert.Contains(t, err.Error(), "1 episode failed")
	assert.Contains(t, err.Error(), "failed to resolve stream")
	assert.Contains(t, err.Error(), "no video source found in the page")

	var mp4s []string
	walkErr := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info != nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".mp4") {
			mp4s = append(mp4s, path)
		}
		return nil
	})
	require.NoError(t, walkErr)
	assert.Empty(t, mp4s, "a batch with no stream must not leave a fake completed mp4")
}

func installDownloadRangeTestState(outputDir string) func() {
	media := snapshotMedia()
	output := util.GlobalOutputDir
	quality := util.GlobalQuality
	subs := append([]util.SubtitleInfo(nil), util.GlobalSubtitles...)
	source := util.GlobalAnimeSource
	request := util.GlobalDownloadRequest
	logger := util.Logger

	util.GlobalOutputDir = outputDir
	util.GlobalQuality = "best"
	util.GlobalSubtitles = nil
	util.GlobalAnimeSource = ""
	util.GlobalDownloadRequest = nil
	util.Logger = log.NewWithOptions(io.Discard, log.Options{Prefix: "player-test"})

	return func() {
		util.GlobalOutputDir = output
		util.GlobalQuality = quality
		util.GlobalSubtitles = subs
		util.GlobalAnimeSource = source
		util.GlobalDownloadRequest = request
		util.Logger = logger

		gMedia.mu.Lock()
		gMedia.animeName = media.AnimeName
		gMedia.animeSeason = media.AnimeSeason
		gMedia.isMovieOrTV = media.IsMovieOrTV
		gMedia.mediaType = media.MediaType
		gMedia.animeURL = media.AnimeURL
		gMedia.seasonMap = media.SeasonMap
		gMedia.meta = media.Meta
		gMedia.mu.Unlock()
	}
}
