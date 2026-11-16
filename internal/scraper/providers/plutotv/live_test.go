package plutotv

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// liveClient skips unless the run opted in: the real service answers per
// country, and CI must never depend on it. Opt in with GOANIME_LIVE=1.
func liveClient(t *testing.T) *Client {
	t.Helper()
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
	return NewClient()
}

// fetchLive reads a URL the way the player will: with its own User-Agent, not
// the scraper's. The CDN refuses requests that name none; mpv names one.
func fetchLive(t *testing.T, target string) (int, string) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, target, http.NoBody)
	require.NoError(t, err)
	req.Header.Set("User-Agent", "libmpv")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	require.NoError(t, err)
	return resp.StatusCode, string(body)
}

// TestLiveChain walks search → seasons → stream for a series and a movie, and
// then reads the playlist it was handed, because a URL that resolves but does
// not play is the one failure a parser test cannot see.
func TestLiveChain(t *testing.T) {
	c := liveClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()

	titles, err := c.Search(ctx, "death note")
	require.NoError(t, err)
	var series, movie *Title
	for i := range titles {
		switch titles[i].Kind {
		case KindSeries:
			if series == nil {
				series = &titles[i]
			}
		case KindMovie:
			if movie == nil {
				movie = &titles[i]
			}
		}
	}
	if series == nil {
		t.Skip("the catalog of this country lists no series for the query")
	}
	t.Logf("series: %s %s", series.Name, series.URL())

	show, err := c.Show(ctx, series.URL())
	require.NoError(t, err)
	require.NotEmpty(t, show.Seasons)
	first := show.Seasons[0].Episodes[0]
	t.Logf("episode: S%d E%d %s (%s)", first.Season, first.Number, first.Name, first.Duration)

	stream, err := c.Stream(ctx, EpisodeURL(show.ID, first.ID), "best")
	require.NoError(t, err)
	t.Logf("stream: adFree=%v %s", stream.AdFree, stream.URL)
	assert.True(t, stream.AdFree, "the title's own playlist should be found")
	assert.Contains(t, stream.URL, "/hls/0-end/", "the uninterrupted playlist, not the stitched one")

	status, body := fetchLive(t, stream.URL)
	require.Equal(t, http.StatusOK, status)
	require.True(t, strings.HasPrefix(body, "#EXTM3U"))
	assert.Contains(t, body, "#EXT-X-ENDLIST")
	assert.NotContains(t, body, "/ad%2F", "no ad creative in the title's playlist")
	assert.NotContains(t, body, "ad_bumper")

	// A segment answers too: the keys sit beside it, so this is the whole path.
	segment := ""
	for _, line := range strings.Split(body, "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			segment = line
			break
		}
	}
	require.NotEmpty(t, segment)
	base := stream.URL[:strings.LastIndex(stream.URL, "/")+1]
	status, _ = fetchLive(t, base+segment)
	assert.Equal(t, http.StatusOK, status, "the first segment is served")

	if movie == nil {
		return
	}
	t.Logf("movie: %s %s", movie.Name, movie.URL())
	stream, err = c.Stream(ctx, movie.URL(), "720p")
	require.NoError(t, err)
	t.Logf("movie stream: adFree=%v %s", stream.AdFree, stream.URL)
	status, _ = fetchLive(t, stream.URL)
	assert.Equal(t, http.StatusOK, status)
}

// TestLiveSearchIgnoresChannels pins the reason Search filters its answer: the
// same endpoint lists live channels and timelines, which cannot be played here.
func TestLiveSearchIgnoresChannels(t *testing.T) {
	c := liveClient(t)
	titles, err := c.Search(t.Context(), "naruto")
	require.NoError(t, err)
	require.NotEmpty(t, titles)
	for _, title := range titles {
		assert.Contains(t, []Kind{KindMovie, KindSeries}, title.Kind, title.Name)
		assert.Regexp(t, idRe, title.ID)
	}
}
