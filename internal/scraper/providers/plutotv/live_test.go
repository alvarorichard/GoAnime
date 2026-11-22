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
	assert.True(t, stream.AdFree, "the title alone should be found")

	// What plays is the playlist's text, not the URL: one of the CDNs behind
	// the media host refuses the playlist at its own URL, so which one answered
	// decides whether stream.URL is the title's own playlist or the stitched
	// master. Either way the text is the title alone, every URI absolute.
	body := stream.Playlist
	require.True(t, strings.HasPrefix(body, "#EXTM3U"))
	assert.Contains(t, body, "#EXT-X-ENDLIST")
	assert.NotContains(t, body, "/ad%2F", "no ad creative in the title's playlist")
	assert.NotContains(t, body, "_ad/", "nor an ad served as a clip")
	assert.NotContains(t, body, "ad_bumper")
	t.Logf("playlist: %d segments, %d range joins, own=%v", strings.Count(body, "#EXTINF"), strings.Count(body, "#EXT-X-DISCONTINUITY"), strings.Contains(stream.URL, "/hls/0-end/"))

	// A segment and its key answer too, to the player's own User-Agent: that is
	// the whole path, on whichever CDN the name resolves to here.
	segment, key := "", ""
	for _, line := range strings.Fields(body) {
		if m := tagURIRe.FindString(line); m != "" && key == "" {
			key = m[len(`URI="`) : len(m)-1]
		}
		if !strings.HasPrefix(line, "#") {
			segment = line
			break
		}
	}
	require.True(t, strings.HasPrefix(segment, "https://"), "segments are absolute: %s", segment)
	require.True(t, strings.HasPrefix(key, "https://"), "and so are key files: %s", key)
	status, _ := fetchLive(t, segment)
	assert.Equal(t, http.StatusOK, status, "the first segment is served")
	status, keyBody := fetchLive(t, key)
	assert.Equal(t, http.StatusOK, status, "its key is served")
	assert.Len(t, keyBody, 16, "an AES-128 key")

	if movie == nil {
		return
	}
	t.Logf("movie: %s %s", movie.Name, movie.URL())
	stream, err = c.Stream(ctx, movie.URL(), "720p")
	require.NoError(t, err)
	t.Logf("movie stream: adFree=%v %s", stream.AdFree, stream.URL)
	assert.True(t, stream.AdFree)
	assert.True(t, strings.HasPrefix(stream.Playlist, "#EXTM3U"), "a movie is handed its playlist the same way")
}

// TestLiveSearchIgnoresChannels pins the reason Search filters its answer: the
// same endpoint lists live channels and timelines, which cannot be played here.
func TestLiveSearchIgnoresChannels(t *testing.T) {
	c := liveClient(t)
	titles, err := c.Search(t.Context(), "naruto")
	require.NoError(t, err)
	if len(titles) == 0 {
		// The catalog is per country, and some list nothing on demand at all.
		t.Skip("the catalog of this country lists nothing for the query")
	}
	for _, title := range titles {
		assert.Contains(t, []Kind{KindMovie, KindSeries}, title.Kind, title.Name)
		assert.Regexp(t, idRe, title.ID)
	}
}

// TestLiveEveryMoviePlaysMoreThanAnAd walks the movies the catalog lists for a
// few queries and holds what Stream hands over to the one thing that went
// wrong in production (2026-11-21, "Jumper"): a title whose segment layout was
// not known was played as the five-second ad bumper at the end of its playlist.
// Whatever layout a title is in, its playlist must be most of a film, or be the
// stitched master (which has the whole film in it).
func TestLiveEveryMoviePlaysMoreThanAnAd(t *testing.T) {
	c := liveClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), 4*time.Minute)
	defer cancel()

	seen := map[string]bool{}
	checked := 0
	for _, query := range []string{"jumper", "matrix", "action", "terror", "comedia", "star", "dragon", "guerra"} {
		titles, err := c.Search(ctx, query)
		require.NoError(t, err)
		for _, title := range titles {
			if title.Kind != KindMovie || seen[title.ID] || checked >= 25 {
				continue
			}
			seen[title.ID] = true
			stream, err := c.Stream(ctx, title.URL(), "best")
			if err != nil {
				t.Logf("%s: %v (not playable here)", title.Name, err)
				continue
			}
			checked++
			if stream.Playlist == "" {
				assert.Contains(t, stream.URL, "/stitch/", "%s: with no playlist of its own, the stitched master", title.Name)
				continue
			}
			segments := strings.Count(stream.Playlist, "#EXTINF")
			total := seconds(stream.Playlist)
			t.Logf("%s: %d segments, %.0f min, adFree=%v", title.Name, segments, total/60, stream.AdFree)
			assert.Greater(t, total, 20*60.0, "%s: a film, not an ad (%d segments)", title.Name, segments)
			assert.NotRegexp(t, adMarkerRe, stream.Playlist, "%s: no ad in the title's playlist", title.Name)
			assertLiveTimeline(t, ctx, title.Name, stream)
		}
	}
	if checked == 0 {
		t.Skip("the catalog of this country lists no movie for these queries")
	}
}

// assertLiveTimeline holds a title that is served from here to the one thing
// the rebase is for: the first segment of every range, as the player gets it,
// starts where the playlist's durations say, so a seek finds what it looks for.
// A title the CDN plays as it is has nothing to hold.
func assertLiveTimeline(t *testing.T, ctx context.Context, name string, stream Stream) {
	t.Helper()
	plan := stream.rebase
	if plan == nil {
		return
	}
	var keys keyCache
	elapsed, first := 0.0, int64(-1)
	for i, s := range plan.segs {
		if i == 0 || s.group != plan.segs[i-1].group {
			data, err := plan.segment(ctx, i, &keys)
			require.NoError(t, err, "%s: segment %d", name, i)
			require.NoError(t, checkTS(data), "%s: segment %d", name, i)
			at, ok := firstTimestamp(data)
			require.True(t, ok, "%s: segment %d", name, i)
			if first < 0 {
				first = at
			}
			want := first + int64(elapsed*90000)
			assert.InDelta(t, float64(want), float64(at), 10*90000, "%s: range starting at segment %d", name, i)
		}
		elapsed += extinfSeconds(s.extinf)
	}
	t.Logf("%s: %d ranges on one timeline", name, plan.segs[len(plan.segs)-1].group+1)
}
