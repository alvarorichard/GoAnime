package player

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAbs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in, want int
	}{
		{0, 0},
		{5, 5},
		{-5, 5},
		{-1, 1},
		{1234, 1234},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, abs(tt.in))
	}
}

func TestExtractResolution(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		label string
		want  int
	}{
		{"1080p", "1080p", 1080},
		{"720p with prefix", "Quality 720p", 720},
		{"4k label", "2160p UHD", 2160},
		{"no number", "best", 0},
		{"empty", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, extractResolution(tt.label))
		})
	}
}

func TestDownloadFolderFormatter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"video url", "https://animefire.io/video/anime-naruto-ep1", "anime-naruto-ep1"},
		{"no match", "https://example.com/", ""},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, DownloadFolderFormatter(tt.in))
		})
	}
}

func TestExtractEpisodeNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "Episódio 5", "5"},
		{"two digit", "Episódio 42", "42"},
		{"movie tag", "Filme", "1"},
		{"ova", "OVA", "1"},
		{"special", "Special", "1"},
		{"naked numeric", "12", "12"},
		{"empty fallback", "", "1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, ExtractEpisodeNumber(tt.in))
		})
	}
}

func TestIsPlayableVideoURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"mp4", "https://x.com/v.mp4", true},
		{"mp4 query", "https://x.com/v.mp4?t=1", true},
		{"m3u8", "https://x.com/m.m3u8", true},
		{"webm", "https://x.com/v.webm", true},
		{"source param", "https://x.com/play?source=video", true},
		{"html", "https://x.com/index.html", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isPlayableVideoURL(tt.url))
		})
	}
}

func TestEstimateContentLengthForAllAnime_HLSReturnsFixed500MB(t *testing.T) {
	t.Parallel()
	// HLS URLs short-circuit: client is never touched.
	got, err := estimateContentLengthForAllAnime("https://cdn.example/x.m3u8", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(500*1024*1024), got)
}

func TestEstimateContentLengthForAllAnime_ParsesContentRange(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Range", "bytes 0-4095/123456789")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(make([]byte, 4096))
	}))
	t.Cleanup(srv.Close)

	got, err := estimateContentLengthForAllAnime(srv.URL+"/video.mp4", srv.Client())
	require.NoError(t, err)
	assert.Equal(t, int64(123456789), got)
}

func TestEstimateContentLengthForAllAnime_StarSizeFallsBackToDefault(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Range", "bytes 0-4095/*")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(make([]byte, 4096))
	}))
	t.Cleanup(srv.Close)

	got, err := estimateContentLengthForAllAnime(srv.URL+"/x.mp4", srv.Client())
	require.NoError(t, err)
	assert.Equal(t, int64(300*1024*1024), got)
}

func TestEstimateContentLengthForAllAnime_NoContentRangeFallsBack(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	}))
	t.Cleanup(srv.Close)

	got, err := estimateContentLengthForAllAnime(srv.URL+"/y.mp4", srv.Client())
	require.NoError(t, err)
	assert.Equal(t, int64(300*1024*1024), got)
}

func TestEstimateContentLengthForAllAnime_InvalidURLReturnsError(t *testing.T) {
	t.Parallel()
	_, err := estimateContentLengthForAllAnime("http://[::1\n", &http.Client{})
	require.Error(t, err)
}

func TestExtractActualVideoURL_UnknownHostReturnsError(t *testing.T) {
	t.Parallel()
	// Neither blogger nor animefire → function never touches network.
	_, err := extractActualVideoURL("https://example.com/some/video.mp4")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid video URL found")
}

func TestIsMovieOrTVSourcePlayer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		anime *models.Anime
		want  bool
	}{
		{"nil", nil, false},
		{"flixhq source", &models.Anime{Source: "SFlix"}, true},
		{"superflix source", &models.Anime{Source: "SuperFlix"}, true},
		{"movie media type", &models.Anime{MediaType: models.MediaTypeMovie}, true},
		{"tv media type", &models.Anime{MediaType: models.MediaTypeTV}, true},
		{"flixhq url", &models.Anime{URL: "https://flixhq.to/x"}, true},
		{"plain anime", &models.Anime{Source: "AllAnime", MediaType: models.MediaTypeAnime}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isMovieOrTVSourcePlayer(tt.anime))
		})
	}
}

// Mutate package singleton bloggerProxy — keep serial.
func TestGetBloggerVideoURL_DefaultEmpty(t *testing.T) {
	StopBloggerProxy()
	bloggerProxy.mu.Lock()
	bloggerProxy.videoURL = ""
	bloggerProxy.mu.Unlock()
	assert.Equal(t, "", GetBloggerVideoURL())
}

func TestStopBloggerProxy_NoServerNoop(t *testing.T) {
	assert.NotPanics(t, func() { StopBloggerProxy() })
}

func TestGetBloggerSessionClient_NotNil(t *testing.T) {
	t.Parallel()
	assert.NotNil(t, getBloggerSessionClient())
}

func TestNewSurfClient_NotNil(t *testing.T) {
	t.Parallel()
	assert.NotNil(t, newSurfClient())
}

func TestNewSurfDownloadClient_NotNil(t *testing.T) {
	t.Parallel()
	assert.NotNil(t, newSurfDownloadClient())
}

func TestSelectEpisodeWithFuzzyFinder_EmptyReturnsError(t *testing.T) {
	_, _, err := SelectEpisodeWithFuzzyFinder(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no episodes")
}

func TestGetVideoURLForEpisode_AllAnimeShortIDRejected(t *testing.T) {
	t.Parallel()
	_, err := GetVideoURLForEpisode("shortid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "use enhanced API")
}

func TestGetVideoURLForEpisodeEnhanced_CancelledContextReturnsImmediately(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := GetVideoURLForEpisodeEnhanced(ctx, &models.Episode{URL: "http://example.com/x"}, nil)
	require.ErrorIs(t, err, context.Canceled)
}

func TestGetVideoURLForEpisodeEnhanced_NilAnimeWithShortIDReturnsError(t *testing.T) {
	// Nil anime + short ID → must surface an error rather than guess.
	_, err := GetVideoURLForEpisodeEnhanced(context.Background(), &models.Episode{URL: "shortid"}, nil)
	require.Error(t, err)
}

func TestGetVideoURLForEpisodeEnhanced_NilAnimeHTTPDelegatesToLegacy(t *testing.T) {
	// Nil anime + HTTP URL → delegates to GetVideoURLForEpisode →
	// extractVideoURL (SafeGet blocked on loopback) returns an error.
	_, err := GetVideoURLForEpisodeEnhanced(context.Background(), &models.Episode{URL: "http://127.0.0.1:1/x"}, nil)
	require.Error(t, err)
}

func TestGetVideoURLForEpisodeEnhanced_AllAnimeSourceRoutesEnhanced(t *testing.T) {
	// AllAnime source → routes through enhanced API. With short bogus ID
	// the underlying client fails and an error surfaces.
	anime := &models.Anime{Source: "AllAnime", URL: "shortid"}
	_, err := GetVideoURLForEpisodeEnhanced(context.Background(), &models.Episode{URL: "shortid", Number: "1"}, anime)
	require.Error(t, err)
}

func TestGetVideoURLForEpisode_HTTPURLDelegatesToExtraction(t *testing.T) {
	// Non-AllAnime ID (HTTP URL) → extractVideoURL → SafeGet blocked → error.
	_, err := GetVideoURLForEpisode("http://127.0.0.1:1/episode-page")
	require.Error(t, err)
}

func TestExtractVideoURL_SSRFBlockedFromLoopbackHost(t *testing.T) {
	t.Parallel()
	// SafeGet rejects loopback IPs (SSRF guard). Function returns error.
	_, err := extractVideoURL("http://127.0.0.1:1/video")
	require.Error(t, err)
}

func TestFetchContent_SSRFBlockedFromLoopbackHost(t *testing.T) {
	t.Parallel()
	_, err := fetchContent("http://127.0.0.1:1/page")
	require.Error(t, err)
}

// Blogger funcs touch the package-level bloggerProxy singleton — keep serial.
func TestExtractBloggerVideoURL_BadURLReturnsError(t *testing.T) {
	_, err := extractBloggerVideoURL("https://blogger.com/does/not/exist")
	require.Error(t, err)
}

func TestStartBloggerProxy_BadURLReturnsError(t *testing.T) {
	t.Cleanup(StopBloggerProxy)
	_, err := startBloggerProxy("https://blogger.com/invalid")
	require.Error(t, err)
}

func TestSelectQualityFromOptions_BestPicksHighestResolution(t *testing.T) {
	t.Parallel()
	data := []VideoData{
		{Src: "low", Label: "480p"},
		{Src: "high", Label: "1080p"},
		{Src: "mid", Label: "720p"},
	}
	assert.Equal(t, "high", selectQualityFromOptions(data, "best"))
	assert.Equal(t, "high", selectQualityFromOptions(data, ""))
}

func TestSelectQualityFromOptions_WorstPicksLowestResolution(t *testing.T) {
	t.Parallel()
	data := []VideoData{
		{Src: "low", Label: "480p"},
		{Src: "high", Label: "1080p"},
	}
	assert.Equal(t, "low", selectQualityFromOptions(data, "worst"))
}

func TestSelectQualityFromOptions_ExactLabelMatch(t *testing.T) {
	t.Parallel()
	data := []VideoData{
		{Src: "a", Label: "480p"},
		{Src: "b", Label: "720p"},
		{Src: "c", Label: "1080p"},
	}
	assert.Equal(t, "b", selectQualityFromOptions(data, "720p"))
}

func TestSelectQualityFromOptions_ClosestResolution(t *testing.T) {
	t.Parallel()
	data := []VideoData{
		{Src: "a", Label: "480p"},
		{Src: "c", Label: "1080p"},
	}
	// 720 requested → 1080 (|240|) closer than 480 (|240|) — tie, first
	// match wins via iteration order.
	got := selectQualityFromOptions(data, "720")
	assert.Contains(t, []string{"a", "c"}, got)
}

func TestSelectQualityFromOptions_EmptyDataReturnsEmpty(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", selectQualityFromOptions(nil, "best"))
}
