package plutotv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
)

// Regression tests: what the real service answered on 2026-11-16, kept as
// fixtures under testdata. Each guards a mistake that was one edit away while
// the source was being written — or a shape the code was fitted to.

func fixture(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)
	return string(raw)
}

// The one search answer lists far more than titles: live channels and
// timelines come back beside them, and none can be played here.
func TestRegressionSearchOfARealAnswerKeepsOnlyTitles(t *testing.T) {
	f := newFakePluto(t)
	f.set("/v1/search", fixture(t, "search_2026_11_16.json"))
	titles, err := f.client().Search(t.Context(), "naruto")
	require.NoError(t, err)

	var series []string
	for _, title := range titles {
		assert.Contains(t, []Kind{KindMovie, KindSeries}, title.Kind, title.Name)
		assert.Regexp(t, idRe, title.ID)
		assert.NotEmpty(t, title.Name)
		if title.Kind == KindSeries {
			series = append(series, title.Name)
		}
	}
	assert.Len(t, titles, 14, "14 of the answer's 30 entries are titles; the other 16 are channels and timelines")
	assert.Contains(t, series, "Naruto: Exame Chuunin")
	assert.Contains(t, series, "Death Note")
	assert.NotContains(t, series, "Pluto TV Anime Ação", "a channel")
	assert.NotContains(t, series, "Boruto: Naruto Next Generations", "a timeline, listed under the same name as a series")
}

func TestRegressionSearchResultsBecomeResultsTheListCanTell(t *testing.T) {
	f := newFakePluto(t)
	f.set("/v1/search", fixture(t, "search_2026_11_16.json"))
	titles, err := f.client().Search(t.Context(), "naruto")
	require.NoError(t, err)
	for _, title := range titles {
		m := title.ToAnimeModel()
		assert.Equal(t, SourceName, m.Source)
		assert.True(t, strings.HasPrefix(m.Name, "[TV] [PT-BR] ") || strings.HasPrefix(m.Name, "[Movie] [PT-BR] "), m.Name)
		assert.True(t, strings.HasPrefix(m.URL, "https://pluto.tv/on-demand/"), m.URL)
		_, err := parsePageURL(m.URL)
		assert.NoError(t, err, "every result URL is one the package reads back")
	}
}

// The season list of a real series, three episodes of it.
func TestRegressionSeasonsOfARealAnswer(t *testing.T) {
	f := newFakePluto(t)
	f.serveShow(fixture(t, "seasons_2026_11_16.json"))
	show, err := f.client().Show(t.Context(), SeriesURL(fakeSeriesID))
	require.NoError(t, err)
	assert.Equal(t, "Death Note", show.Name)
	require.Len(t, show.Seasons, 1)
	eps := show.Seasons[0].Episodes
	require.Len(t, eps, 3)
	assert.Equal(t, "Renascimento", eps[0].Name)
	assert.Equal(t, 1, eps[0].Number)
	assert.Equal(t, fakeEpisodeID, eps[0].ID)
	assert.Equal(t, 1395*time.Second, eps[0].Duration, "originalContentDuration: the title's, without the breaks the service plans around it")
	assert.Equal(t, "Confronto", eps[1].Name)
	assert.Equal(t, []int{1, 2, 3}, []int{eps[0].Number, eps[1].Number, eps[2].Number})
}

func TestRegressionTheRealMasterReadsAsFiveRungsWithRelativeURIs(t *testing.T) {
	const masterURL = "https://cfd-v4-service-channel-stitcher-use1-1.prd.pluto.tv/v2/stitch/hls/episode/6206bbaa15c8610013683a28/master.m3u8"
	variants := parseMaster(fixture(t, "master_2026_11_16.m3u8"), masterURL)
	require.Len(t, variants, 5)
	var bandwidths []int
	for _, v := range variants {
		bandwidths = append(bandwidths, v.Bandwidth)
		assert.True(t, strings.HasPrefix(v.URL, "https://cfd-v4-service-channel-stitcher-use1-1.prd.pluto.tv/v2/stitch/hls/episode/6206bbaa15c8610013683a28/"), v.URL)
		assert.Zero(t, v.Height, "the real master names no resolution: the bitrate has to stand in")
	}
	assert.Equal(t, []int{1042180, 1577180, 2294080, 3321280, 640930}, bandwidths, "the master lists them out of order")
	assert.NotEmpty(t, variants[0].URL)
	assert.Contains(t, variants[0].URL, "/1042180/playlist.m3u8?")
	assert.NotContains(t, variants[0].URL, "audio", "the audio rendition is not a rung")
}

func TestRegressionTheRealMastersLadderIsOrderedByTheBudgetsAssumption(t *testing.T) {
	// bandwidthBudget was fitted to this ladder; if the service changes it,
	// the heights it maps to move and this says so.
	variants := parseMaster(fixture(t, "master_2026_11_16.m3u8"), "https://x.pluto.tv/m.m3u8")
	want := map[string]string{"best": "3321280", "720p": "3321280", "576p": "2294080", "480p": "1577180", "360p": "1042180", "240p": "640930", "worst": "640930"}
	for quality, bandwidth := range want {
		got := orderVariants(variants, quality)
		assert.Contains(t, got[0].URL, "/"+bandwidth+"/", quality)
	}
}

// The three stitched playlists below are real: a series episode and two
// movies, each cut into ranges with ad creatives between them and a bumper at
// the end. The title's own playlist is the directory they all point into.
func TestRegressionTheRealStitchedPlaylistsPointAtTheTitlesOwnPlaylist(t *testing.T) {
	cases := []struct {
		file string
		want string
	}{
		{"stitched_series_2026_11_16.m3u8", "https://siloh-ns1.plutotv.net/538_Viz_Media/clip/6206bba415c86100136839a7_Death_Note_BR_Renascimento_S1E1/720p/20220214_054014/hls/0-end/hls_300.m3u8"},
		{"stitched_movie_2026_11_16.m3u8", "https://siloh-ns1.plutotv.net/772_pluto/clip/6880f1fa265d8a2eacd80d7b_4318f538-3fb5-4c15-98cc-6add190e32aa/720pDRM/20250723_143020/hls/0-end/hls_600.m3u8"},
		{"stitched_movie_pt_2026_11_16.m3u8", "https://siloh-ns1.plutotv.net/c6009f_pluto/clip/611ea6b5d24f5e00135dce27_Rob_ZombieS_Halloween_2_BR/720p/20210819_114509/hls/0-end/hls_600.m3u8"},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			body := fixture(t, tc.file)
			require.Contains(t, body, "#EXT-X-DISCONTINUITY\n", "the fixture still has the splices it is meant to show")
			require.Contains(t, body, "ad%2Fcreative", "and the ads")
			got, ok := contentPlaylistURL(body)
			require.True(t, ok)
			assert.Equal(t, tc.want, got)
			assert.NotContains(t, got, "ad%2F")
			assert.NotContains(t, got, "bumper")
			assert.False(t, isProtected(body), "AES-128 and NONE only, whatever the directory is called")
		})
	}
}

// The service's closing bumper is a clip of its own, 0-end like the title. A
// reader that took "the clip that reaches the end" would have picked it.
func TestRegressionTheClosingBumperIsNotTheTitle(t *testing.T) {
	body := fixture(t, "stitched_series_2026_11_16.m3u8")
	require.Contains(t, body, "ptv_7423_ad_bumper", "the fixture carries the real bumper")
	got, _ := contentPlaylistURL(body)
	assert.NotContains(t, got, "bumper")
}

// The CDN answers 403 to a request that names no User-Agent (the first probe
// of the source got that). The client must always name one, and the fake CDN
// now refuses what the real one refused.
func TestRegressionTheClientAlwaysNamesAUserAgentOnTheMediaHost(t *testing.T) {
	f := newFakePluto(t)
	f.serveEpisode(fakeEpisodeID)
	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	require.True(t, stream.AdFree, "a 403 from the media host would have sent this to the stitched master")
	r := f.lastRequest(f.contentPath("hls_2400"))
	require.NotNil(t, r)
	assert.Equal(t, netx.UserAgent, r.Header.Get("User-Agent"))
}

// A rendition that is called "720pDRM" is not a protected one: the service
// keys it with AES-128, which every player reads (God of War, 2025).
func TestRegressionADrmNamedDirectoryPlays(t *testing.T) {
	body := fixture(t, "stitched_movie_2026_11_16.m3u8")
	require.Contains(t, body, "720pDRM")
	assert.False(t, isProtected(body))
}

// The service's master carries no RESOLUTION; reading heights off it would
// have made "720p" and "best" the same by accident.
func TestRegressionTheMasterNamesNoResolution(t *testing.T) {
	assert.NotContains(t, fixture(t, "master_2026_11_16.m3u8"), "RESOLUTION=")
}

// Fixtures must not carry the session that fetched them.
func TestRegressionFixturesCarryNoSession(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	require.NoError(t, err)
	require.NotEmpty(t, entries)
	for _, e := range entries {
		body := fixture(t, e.Name())
		assert.NotRegexp(t, `eyJ[A-Za-z0-9_-]{20,}`, body, "%s: a JWT", e.Name())
		assert.NotRegexp(t, `(sid|deviceId|clientID|sessionID)=[0-9a-f]{8}-[0-9a-f]{4}-`, body, "%s: a session id", e.Name())
		assert.NotRegexp(t, `\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`, strings.ReplaceAll(body, "0.0.0.0", ""), "%s: an address", e.Name())
	}
}
