package startflix

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVIDSRCBasesDefaultIncludeSh pins the fallback walk order: the primary,
// its mirror, and vidsrc.sh — the panel's own front, family-checked live on
// 2026-10-21 (same vs_src protocol, one title end to end in about two
// seconds). When a front goes away this fails loudly instead of the walk
// silently losing a base.
func TestVIDSRCBasesDefaultIncludeSh(t *testing.T) {
	t.Setenv("GOANIME_VIDSRC_URL", "")
	bases := vidsrcBasesFn()
	assert.Equal(t, []string{
		"https://vidsrc.io",
		"https://vsembed.ru",
		"https://vidsrc.sh",
	}, bases)
}

// TestVIDSRCPanelPlayerMapping pins which panel player URLs belong to the
// vidsrc family: vidsrc.sh embed paths key movies by IMDb id and series by
// TMDB id, with the optional season/episode tail; the ?ds_lang query does
// not matter. Anything else stays on yt-dlp.
func TestVIDSRCPanelPlayerMapping(t *testing.T) {
	for _, raw := range []string{
		// Path-keyed shapes (movie panel lives by tt id, others by TMDB id).
		"https://vidsrc.sh/embed/movie/tt7526136?ds_lang=pt",
		"https://vidsrc.sh/embed/movie/tt0315327",
		"https://vidsrc.sh/embed/movie/310",
		"https://vidsrc.sh/embed/tv/194583",
		"https://vidsrc.sh/embed/tv/194583/2/3",
		"https://vidsrc.sh/embed/series/194583/1/2",
		// Query-keyed shapes — the panel's live series embed keys on the
		// query alone (seen live 2026-10-21).
		"https://vidsrc.sh/embed/tv?tmdb=194583&season=1&episode=1&ds_lang=pt",
		"https://vidsrc.sh/embed/movie?imdb=tt0315327",
	} {
		u, err := url.Parse(raw)
		require.NoError(t, err)
		assert.True(t, isVIDSRCPanelPlayer(u), raw)
	}
	for _, raw := range []string{
		"https://vidsrcnotice.example/embed/movie/tt0315327", // another host entirely
		"https://vidsrc.sh/embed/movie/",                     // no id anywhere
		"https://vidsrc.sh/embed/movie?ds_lang=pt",           // query carries no id either
		"https://vidsrc.sh/embed/movie/tt123abc",             // malformed id
		"https://vidsrc.sh/embed/movie/tt0315327/season/2",   // non-numeric tail
		"https://vidsrc.sh/embed/player/movie/tt0315327",     // an inner render, not a panel embed
		"https://vidsrc.sh/",                                 // the front page, no path
	} {
		u, err := url.Parse(raw)
		require.NoError(t, err)
		assert.False(t, isVIDSRCPanelPlayer(u), raw)
	}
}

// TestVIDSRCPanelPlayerMovie resolves the panel's vidsrc.sh embed end to end
// through the fake: the URL's path names the id, the chain keys vs_src on
// it, and the first master is stamped and probed — the same walk
// MovieFallback runs, just entered through the player URL.
func TestVIDSRCPanelPlayerMovie(t *testing.T) {
	playerPage := fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tok1")
	f := &fakeVIDSRC{
		kind:         KindMovie,
		playerPages:  []string{playerPage},
		plainEnvJSON: `{"status_code":"200","data":{"stream_urls":["SRV/pl/panel/master.m3u8"]}}`,
	}
	base := vidsrcHTTPClientForTest(t, f.start(t))

	u, err := url.Parse("https://vidsrc.sh/embed/movie/tt0315327?ds_lang=pt")
	require.NoError(t, err)
	require.True(t, isVIDSRCPanelPlayer(u))

	stream, err := resolveVIDSRCPanelPlayer(context.Background(), u)
	require.NoError(t, err)

	assert.Equal(t, base.String()+"/pl/panel/master.m3u8?token=testtoken", stream.URL)
	assert.Equal(t, 1080, stream.Height)
	assert.Equal(t, 1, f.masterHits)
	assert.Equal(t, 1, f.playerRenders, "one render, one single-use token")
	assert.Equal(t, AudioSubtitled, stream.Audio, "the chain's own stream is the original audio, never the dub")
}

// TestVIDSRCPanelPlayerSeriesTail pins the series key: a panel URL naming a
// season/episode tail carries both into the stream-data call, and the
// defaults probe never happens.
func TestVIDSRCPanelPlayerSeriesTail(t *testing.T) {
	playerPage := fakeTVConfig("SRV/data/api.php?type=tv&tmdb=1429", "tok1")
	f := &fakeVIDSRC{
		kind:         KindSeries,
		id:           "1429",
		playerPages:  []string{playerPage},
		plainEnvJSON: `{"status_code":"200","data":{"stream_urls":["SRV/pl/s2e3/master.m3u8"]}}`,
	}
	base := vidsrcHTTPClientForTest(t, f.start(t))

	u, err := url.Parse("https://vidsrc.sh/embed/tv/1429/2/3")
	require.NoError(t, err)

	stream, err := resolveVIDSRCPanelPlayer(context.Background(), u)
	require.NoError(t, err)

	assert.Equal(t, base.String()+"/pl/s2e3/master.m3u8?token=testtoken", stream.URL)
	require.Len(t, f.dataFetched, 1, "no defaults probe when the URL names the episode")
	q := f.dataFetched[0]
	assert.Contains(t, q, "season=2")
	assert.Contains(t, q, "episode=3")
	assert.Contains(t, q, "stream_urls=")
	assert.Contains(t, q, "api_token=tok1")
}

// TestVIDSRCPanelPlayerSeriesDefaults covers a panel URL without a tail: the
// chain probes the series map and plays the source's own default episode,
// exactly as the standalone fallback does.
func TestVIDSRCPanelPlayerSeriesDefaults(t *testing.T) {
	playerPage := fakeTVConfig("SRV/data/api.php?type=tv&tmdb=1429", "tok1")
	f := &fakeVIDSRC{
		kind:         KindSeries,
		id:           "1429",
		playerPages:  []string{playerPage},
		plainEnvJSON: `{"status_code":"200","data":{"stream_urls":["SRV/pl/s1e2/master.m3u8"]}}`,
		tvMapJSON:    `{"status_code":"200","data":{"season":1,"episode":"2","eps":{"1":["1","2","3"]}}}`,
	}
	base := vidsrcHTTPClientForTest(t, f.start(t))

	u, err := url.Parse("https://vidsrc.sh/embed/tv/1429")
	require.NoError(t, err)

	stream, err := resolveVIDSRCPanelPlayer(context.Background(), u)
	require.NoError(t, err)

	assert.Equal(t, base.String()+"/pl/s1e2/master.m3u8?token=testtoken", stream.URL)
	require.Len(t, f.dataFetched, 1, "only the envelope call is recorded; the map probe is not")
	q := f.dataFetched[0]
	assert.Contains(t, q, "season=1")
	assert.Contains(t, q, "episode=2")
}

// TestVIDSRCPanelPlayerSeriesQueryShape covers the panel's live series embed
// shape: ids and season/episode ride the query, no path segments at all.
func TestVIDSRCPanelPlayerSeriesQueryShape(t *testing.T) {
	playerPage := fakeTVConfig("SRV/data/api.php?type=tv&tmdb=1429", "tok1")
	f := &fakeVIDSRC{
		kind:         KindSeries,
		id:           "1429",
		playerPages:  []string{playerPage},
		plainEnvJSON: `{"status_code":"200","data":{"stream_urls":["SRV/pl/s1e1/master.m3u8"]}}`,
	}
	base := vidsrcHTTPClientForTest(t, f.start(t))

	u, err := url.Parse("https://vidsrc.sh/embed/tv?tmdb=1429&season=1&episode=1&ds_lang=pt")
	require.NoError(t, err)
	require.True(t, isVIDSRCPanelPlayer(u))

	stream, err := resolveVIDSRCPanelPlayer(context.Background(), u)
	require.NoError(t, err)

	assert.Equal(t, base.String()+"/pl/s1e1/master.m3u8?token=testtoken", stream.URL)
	require.Len(t, f.dataFetched, 1, "no defaults probe when the query names the episode")
	q := f.dataFetched[0]
	assert.Contains(t, q, "season=1")
	assert.Contains(t, q, "episode=1")
}

// TestVIDSRCPanelPlayerMovieQueryShape covers the query-keyed movie embed:
// ?imdb= keys the title, the chain keeps its original-audio stream.
func TestVIDSRCPanelPlayerMovieQueryShape(t *testing.T) {
	playerPage := fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tok1")
	f := &fakeVIDSRC{
		kind:         KindMovie,
		playerPages:  []string{playerPage},
		plainEnvJSON: `{"status_code":"200","data":{"stream_urls":["SRV/pl/q/master.m3u8"]}}`,
	}
	base := vidsrcHTTPClientForTest(t, f.start(t))

	u, err := url.Parse("https://vidsrc.sh/embed/movie?imdb=tt0315327")
	require.NoError(t, err)
	require.True(t, isVIDSRCPanelPlayer(u))

	stream, err := resolveVIDSRCPanelPlayer(context.Background(), u)
	require.NoError(t, err)

	assert.Equal(t, base.String()+"/pl/q/master.m3u8?token=testtoken", stream.URL)
	assert.Equal(t, 1, f.masterHits)
}

// TestResolveStreamForRoutesVIDSRCPanelPlayer wires the dispatch: a panel
// player on vidsrc.sh resolves through the vidsrc machinery during the
// player sweep, not through the yt-dlp round.
func TestResolveStreamForRoutesVIDSRCPanelPlayer(t *testing.T) {
	playerPage := fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tok1")
	f := &fakeVIDSRC{
		kind:         KindMovie,
		playerPages:  []string{playerPage},
		plainEnvJSON: `{"status_code":"200","data":{"stream_urls":["SRV/pl/wired/master.m3u8"]}}`,
	}
	vidsrcHTTPClientForTest(t, f.start(t))

	c := NewClientForTest(nil, "https://www.startflix.test")
	stream, err := c.ResolveStreamFor(context.Background(),
		[]Player{{URL: "https://vidsrc.sh/embed/movie/tt0315327"}}, AudioDubbed)
	require.NoError(t, err)

	assert.Contains(t, stream.URL, "/pl/wired/master.m3u8")
	assert.Equal(t, AudioDubbed, stream.Audio, "an unlabeled player carries the wanted audio")
}
