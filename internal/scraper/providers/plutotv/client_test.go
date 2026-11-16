package plutotv

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
)

// ---- identifiers -----------------------------------------------------------

func TestParsePageURL(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		want    pageRef
		wantErr bool
	}{
		{"series", SeriesURL(oid('a')), pageRef{kind: KindSeries, id: oid('a')}, false},
		{"movie", MovieURL(oid('b')), pageRef{kind: KindMovie, id: oid('b')}, false},
		{"episode", EpisodeURL(oid('a'), oid('c')), pageRef{kind: KindSeries, id: oid('a'), episodeID: oid('c')}, false},
		{"trailing slash", SeriesURL(oid('a')) + "/", pageRef{kind: KindSeries, id: oid('a')}, false},
		{"www host", "https://www.pluto.tv/on-demand/movies/" + oid('b'), pageRef{kind: KindMovie, id: oid('b')}, false},
		{"upper-case host", "https://PLUTO.tv/on-demand/movies/" + oid('b'), pageRef{kind: KindMovie, id: oid('b')}, false},
		{"query is ignored", SeriesURL(oid('a')) + "?x=1", pageRef{kind: KindSeries, id: oid('a')}, false},
		{"http is refused", "http://pluto.tv/on-demand/movies/" + oid('b'), pageRef{}, true},
		{"another host", "https://pluto.tv.evil.example/on-demand/movies/" + oid('b'), pageRef{}, true},
		{"a look-alike host", "https://notpluto.tv/on-demand/movies/" + oid('b'), pageRef{}, true},
		{"credentials", "https://user:pw@pluto.tv/on-demand/movies/" + oid('b'), pageRef{}, true},
		{"short id", "https://pluto.tv/on-demand/movies/abc", pageRef{}, true},
		{"upper-case id", "https://pluto.tv/on-demand/movies/" + strings.ToUpper(oid('a')), pageRef{}, true},
		{"non-hex id", "https://pluto.tv/on-demand/movies/" + strings.Repeat("z", 24), pageRef{}, true},
		{"movie with an episode", "https://pluto.tv/on-demand/movies/" + oid('a') + "/episode/" + oid('b'), pageRef{}, true},
		{"unrelated path", "https://pluto.tv/live-tv/" + oid('a'), pageRef{}, true},
		{"empty", "", pageRef{}, true},
		{"not a URL", "%%%", pageRef{}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parsePageURL(tc.raw)
			if tc.wantErr {
				require.Error(t, err)
				var diag *netx.SourceDiagnostic
				assert.ErrorAs(t, err, &diag, "a refused URL is a classified error")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestPlayID(t *testing.T) {
	id, ok := pageRef{kind: KindMovie, id: oid('b')}.playID()
	assert.True(t, ok)
	assert.Equal(t, oid('b'), id, "a movie plays by its own id")

	id, ok = pageRef{kind: KindSeries, id: oid('a'), episodeID: oid('c')}.playID()
	assert.True(t, ok)
	assert.Equal(t, oid('c'), id, "an episode plays by the episode's id, not the series'")

	_, ok = pageRef{kind: KindSeries, id: oid('a')}.playID()
	assert.False(t, ok, "a series names nothing to play")
}

func TestIsMovieURL(t *testing.T) {
	assert.True(t, IsMovieURL(MovieURL(oid('b'))))
	assert.False(t, IsMovieURL(SeriesURL(oid('a'))))
	assert.False(t, IsMovieURL(EpisodeURL(oid('a'), oid('c'))))
	assert.False(t, IsMovieURL("https://example.com/"))
	assert.False(t, IsMovieURL(""))
}

func TestTitleURLFollowsTheKind(t *testing.T) {
	assert.Equal(t, MovieURL(oid('a')), Title{ID: oid('a'), Kind: KindMovie}.URL())
	assert.Equal(t, SeriesURL(oid('a')), Title{ID: oid('a'), Kind: KindSeries}.URL())
}

func TestTitleYear(t *testing.T) {
	cases := []struct {
		title Title
		want  string
	}{
		{Title{Kind: KindMovie, Slug: "god-of-war-2017-1-1-ptv2"}, "2017"},
		{Title{Kind: KindMovie, Slug: "rob-zombies-halloween-2-br-2008-1-1"}, "2008"},
		{Title{Kind: KindMovie, Slug: "vendetta-aka-blood-for-blood-2022-1-1-ptv1"}, "2022"},
		{Title{Kind: KindMovie, Slug: "no-year-here"}, ""},
		{Title{Kind: KindMovie, Slug: "year-1850-1-1"}, "" /* before 1900 */},
		{Title{Kind: KindMovie, Slug: ""}, ""},
		{Title{Kind: KindSeries, Slug: "death-note-2006-1-1"}, ""},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, tc.title.Year(), "%+v", tc.title)
	}
}

func TestLanguageTag(t *testing.T) {
	assert.Equal(t, "[PT-BR]", languageTag("PT"))
	assert.Equal(t, "[PT-BR]", languageTag(" pt "))
	assert.Equal(t, "[English]", languageTag("EN"))
	assert.Equal(t, "[Multilanguage]", languageTag("ES"))
	assert.Equal(t, "[Multilanguage]", languageTag(""))
}

func TestToAnimeModel(t *testing.T) {
	movie := Title{ID: oid('b'), Name: "Halloween II", Slug: "rob-zombies-halloween-2-br-2008-1-1", Language: "PT", Kind: KindMovie}.ToAnimeModel()
	assert.Equal(t, "[Movie] [PT-BR] Halloween II", movie.Name)
	assert.Equal(t, MovieURL(oid('b')), movie.URL)
	assert.Equal(t, SourceName, movie.Source)
	assert.Equal(t, "movie", string(movie.MediaType))
	assert.Equal(t, "2008", movie.Year)

	series := Title{ID: oid('a'), Name: "Death Note", Language: "EN", Kind: KindSeries}.ToAnimeModel()
	assert.Equal(t, "[TV] [English] Death Note", series.Name)
	assert.Equal(t, SeriesURL(oid('a')), series.URL)
	assert.Equal(t, "tv", string(series.MediaType))
	assert.Empty(t, series.Year)
}

func TestNewClientID(t *testing.T) {
	uuid := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	seen := map[string]bool{}
	for range 50 {
		id := newClientID()
		assert.Regexp(t, uuid, id)
		assert.False(t, seen[id], "an id is made per client, never repeated")
		seen[id] = true
	}
}

func TestSharedIsOneClient(t *testing.T) {
	assert.Same(t, Shared(), Shared())
}

// ---- trusted hosts ---------------------------------------------------------

func TestTrusted(t *testing.T) {
	c := NewClient()
	for _, raw := range []string{
		"https://pluto.tv/x",
		"https://boot.pluto.tv/v4/start",
		"https://service-vod.clusters.pluto.tv",
		"https://cfd-v4-service-channel-stitcher-use1-1.prd.pluto.tv",
		"https://siloh-ns1.plutotv.net/a/b.m3u8",
		"https://PLUTOTV.net/a",
	} {
		assert.True(t, c.trusted(raw), raw)
	}
	for _, raw := range []string{
		"http://pluto.tv/x",
		"https://pluto.tv.evil.example/x",
		"https://evilpluto.tv/x",
		"https://evil-plutotv.net/x",
		"https://pluto.tv@evil.example/x",
		"https://user@pluto.tv/x",
		"ftp://pluto.tv/x",
		"//pluto.tv/x",
		"pluto.tv/x",
		"https:///x",
		"",
		"%%%",
		"https://127.0.0.1/x",
	} {
		assert.False(t, c.trusted(raw), raw)
	}
}

func TestTheTestClientTrustsOnlyItsOwnServerBesidesPluto(t *testing.T) {
	f := newFakePluto(t)
	c := f.client()
	assert.True(t, c.trusted(f.url()+"/x"), "the fake")
	assert.True(t, c.trusted("https://pluto.tv/x"), "the real service stays trusted")
	assert.False(t, c.trusted("https://cdn.evil.example/x"), "anything else stays refused")
	assert.False(t, c.trusted("http://127.0.0.1:1/x"), "another local port is not the fake")
	assert.False(t, c.trusted("ftp://"+f.srv.Listener.Addr().String()+"/x"), "the fake is trusted over http(s) only")
}

// ---- session ---------------------------------------------------------------

func TestBootOpensAWebSessionWithAnIdentityOfItsOwn(t *testing.T) {
	f := newFakePluto(t)
	f.set("/v1/search", searchBody())
	c := f.client()
	_, err := c.Search(t.Context(), "x")
	require.NoError(t, err)

	q := f.lastRequest("/v4/start").URL.Query()
	assert.Equal(t, "web", q.Get("appName"))
	assert.Equal(t, "web", q.Get("deviceType"))
	assert.Equal(t, "false", q.Get("serverSideAds"), "the service is asked for no server-side ads")
	assert.Equal(t, c.clientID, q.Get("clientID"))
	assert.Regexp(t, `^[0-9a-f-]{36}$`, q.Get("clientID"))
	assert.NotEmpty(t, q.Get("appVersion"))
	assert.NotEmpty(t, q.Get("deviceVersion"))
}

func TestOneSessionServesManyCalls(t *testing.T) {
	f := newFakePluto(t)
	f.set("/v1/search", searchBody())
	c := f.client()
	for range 3 {
		_, err := c.Search(t.Context(), "x")
		require.NoError(t, err)
	}
	assert.Equal(t, 1, f.bootCount(), "the session is reused")
}

func TestASessionThatRanOutIsReplaced(t *testing.T) {
	f := newFakePluto(t)
	f.set("/v1/search", searchBody())
	c := f.client()
	_, err := c.Search(t.Context(), "x")
	require.NoError(t, err)

	c.mu.Lock()
	c.sess.expires = time.Now().Add(-time.Second)
	c.mu.Unlock()

	_, err = c.Search(t.Context(), "x")
	require.NoError(t, err)
	assert.Equal(t, 2, f.bootCount())
	assert.Equal(t, "Bearer tok-2", f.lastRequest("/v1/search").Header.Get("Authorization"), "the new token is the one used")
}

func TestTheSessionLifetimeFollowsTheServiceButNeverOutlastsTheCap(t *testing.T) {
	cases := []struct {
		name    string
		refresh int
		min     time.Duration
		max     time.Duration
	}{
		{"the service says nothing", 0, sessionLifetime - time.Minute, sessionLifetime + time.Second},
		{"a short refresh", 2, time.Second, 3 * time.Second},
		{"a long refresh is capped", 7200, sessionLifetime - time.Minute, sessionLifetime + time.Second},
		{"a negative refresh is ignored", -5, sessionLifetime - time.Minute, sessionLifetime + time.Second},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFakePluto(t)
			f.refresh = tc.refresh
			f.set("/v1/search", searchBody())
			c := f.client()
			_, err := c.Search(t.Context(), "x")
			require.NoError(t, err)
			c.mu.Lock()
			left := time.Until(c.sess.expires)
			c.mu.Unlock()
			assert.Greater(t, left, tc.min-time.Second)
			assert.LessOrEqual(t, left, tc.max)
		})
	}
}

func TestA401RenewsTheSessionOnceAndRetries(t *testing.T) {
	f := newFakePluto(t)
	f.set("/v1/search", searchBody(hit("series", oid('a'), "Death Note", "death-note-br", "PT")))
	f.reject = 1
	c := f.client()
	titles, err := c.Search(t.Context(), "death")
	require.NoError(t, err)
	require.Len(t, titles, 1)
	assert.Equal(t, 2, f.bootCount(), "the rejected session is replaced")
	assert.Equal(t, 2, f.hitCount("/v1/search"))
}

func TestASecond401IsAnErrorNotALoop(t *testing.T) {
	f := newFakePluto(t)
	f.set("/v1/search", searchBody())
	f.reject = 5
	_, err := f.client().Search(t.Context(), "x")
	require.Error(t, err)
	assert.Equal(t, 2, f.bootCount(), "one renewal, not more")
	assert.Equal(t, 2, f.hitCount("/v1/search"))
	var diag *netx.SourceDiagnostic
	assert.ErrorAs(t, err, &diag)
}

func TestBootFailures(t *testing.T) {
	t.Run("the service is down", func(t *testing.T) {
		f := newFakePluto(t)
		f.bootCode = http.StatusServiceUnavailable
		_, err := f.client().Search(t.Context(), "x")
		require.Error(t, err)
		assert.ErrorIs(t, err, netx.ErrSourceUnavailable)
	})
	t.Run("no token", func(t *testing.T) {
		f := newFakePluto(t)
		f.bootBody = func(f *fakePluto) string {
			return fmt.Sprintf(`{"sessionToken": "  ", "servers": {"search": %q, "vod": %q, "stitcher": %q}}`, f.url(), f.url(), f.url())
		}
		_, err := f.client().Search(t.Context(), "x")
		require.ErrorContains(t, err, "no token")
	})
	t.Run("malformed answer", func(t *testing.T) {
		f := newFakePluto(t)
		f.bootBody = func(*fakePluto) string { return "<html>maintenance</html>" }
		_, err := f.client().Search(t.Context(), "x")
		require.ErrorContains(t, err, "invalid API JSON")
	})
	for _, which := range []string{"search", "vod", "stitcher"} {
		t.Run("a foreign "+which+" host", func(t *testing.T) {
			f := newFakePluto(t)
			spy := newSpyHost(t, searchBody())
			f.bootBody = func(f *fakePluto) string {
				hosts := map[string]string{"search": f.url(), "vod": f.url(), "stitcher": f.url()}
				hosts[which] = spy.URL
				return fmt.Sprintf(`{"sessionToken": "tok", "servers": {"search": %q, "vod": %q, "stitcher": %q}}`, hosts["search"], hosts["vod"], hosts["stitcher"])
			}
			_, err := f.client().Search(t.Context(), "x")
			require.ErrorContains(t, err, "unexpected service host")
			assert.Zero(t, spy.hits.Load(), "the token is never sent to it")
			assert.Empty(t, spy.auths.Load())
		})
	}
}

func TestStitcherParamsAreUsedWithoutTheirEdgeSeparators(t *testing.T) {
	f := newFakePluto(t)
	f.bootBody = func(f *fakePluto) string {
		return fmt.Sprintf(`{"sessionToken": %q, "stitcherParams": "&appName=web&sid=%s&", "servers": {"search": %q, "vod": %q, "stitcher": %q}}`,
			f.token, fakeSID, f.url(), f.url(), f.url())
	}
	f.serveEpisode(fakeEpisodeID)
	_, err := f.client().Stream(t.Context(), EpisodeURL(fakeSeriesID, fakeEpisodeID), "best")
	require.NoError(t, err)
	raw := f.lastRequest(f.masterPath(fakeEpisodeID)).URL.RawQuery
	assert.NotContains(t, raw, "&&", "no empty parameter")
	assert.True(t, strings.HasPrefix(raw, "appName=web"), raw)
}

// ---- search ----------------------------------------------------------------

func TestSearchSendsWhatTheServiceExpects(t *testing.T) {
	f := newFakePluto(t)
	f.set("/v1/search", searchBody())
	_, err := f.client().Search(t.Context(), "  death note ")
	require.NoError(t, err)

	r := f.lastRequest("/v1/search")
	assert.Equal(t, "death note", r.URL.Query().Get("q"), "the query is trimmed")
	assert.Equal(t, "60", r.URL.Query().Get("limit"))
	assert.Equal(t, "Bearer tok-1", r.Header.Get("Authorization"))
	assert.Equal(t, netx.UserAgent, r.Header.Get("User-Agent"))
	assert.Equal(t, "https://pluto.tv", r.Header.Get("Origin"))
	assert.Equal(t, "https://pluto.tv/", r.Header.Get("Referer"))
	assert.Equal(t, "application/json", r.Header.Get("Accept"))
}

func TestSearchOfNothingAsksNothing(t *testing.T) {
	f := newFakePluto(t)
	for _, q := range []string{"", "   ", "\t\n"} {
		titles, err := f.client().Search(t.Context(), q)
		require.NoError(t, err)
		assert.Empty(t, titles)
	}
	assert.Zero(t, f.bootCount(), "not even a session is opened")
}

func TestSearchKeepsOnlyWhatCanBePlayed(t *testing.T) {
	f := newFakePluto(t)
	f.set("/v1/search", searchBody(
		hit("series", oid('a'), "Death Note", "death-note-br", "PT"),
		hit("channel", oid('b'), "Naruto", "naruto-ptv1", "PT"),
		hit("timeline", oid('c'), "One Piece", oid('c'), "PT"),
		hit("movie", oid('d'), "Halloween II", "rob-zombies-halloween-2-br-2008-1-1", "EN"),
		hit("episode", oid('e'), "An episode", "x", "PT"),
		hit("", oid('f'), "No type", "x", "PT"),
	))
	titles, err := f.client().Search(t.Context(), "x")
	require.NoError(t, err)
	require.Len(t, titles, 2)
	assert.Equal(t, Title{ID: oid('a'), Name: "Death Note", Slug: "death-note-br", Language: "PT", Kind: KindSeries}, titles[0])
	assert.Equal(t, KindMovie, titles[1].Kind)
}

func TestSearchDropsWhatItCannotAddress(t *testing.T) {
	f := newFakePluto(t)
	f.set("/v1/search", searchBody(
		hit("series", "not-an-id", "Bad id", "x", "PT"),
		hit("series", strings.ToUpper(oid('a')), "Upper id", "x", "PT"),
		hit("series", oid('b'), "   ", "x", "PT"),
		hit("series", oid('c'), " Kept ", "x", "PT"),
		hit("movie", oid('c'), "Same id again", "x", "PT"),
	))
	titles, err := f.client().Search(t.Context(), "x")
	require.NoError(t, err)
	require.Len(t, titles, 1)
	assert.Equal(t, "Kept", titles[0].Name, "names are trimmed and the first of an id wins")
}

func TestSearchKeepsTheServicesOrder(t *testing.T) {
	f := newFakePluto(t)
	f.set("/v1/search", searchBody(
		hit("movie", oid('3'), "Third", "x", "PT"),
		hit("series", oid('1'), "First", "x", "PT"),
		hit("movie", oid('2'), "Second", "x", "PT"),
	))
	titles, err := f.client().Search(t.Context(), "x")
	require.NoError(t, err)
	var names []string
	for _, title := range titles {
		names = append(names, title.Name)
	}
	assert.Equal(t, []string{"Third", "First", "Second"}, names)
}

func TestSearchErrors(t *testing.T) {
	t.Run("server error", func(t *testing.T) {
		f := newFakePluto(t)
		f.setRoute("/v1/search", http.StatusBadGateway, "oops")
		_, err := f.client().Search(t.Context(), "x")
		require.Error(t, err)
		assert.ErrorIs(t, err, netx.ErrSourceUnavailable)
	})
	t.Run("not JSON", func(t *testing.T) {
		f := newFakePluto(t)
		f.set("/v1/search", "<html>blocked</html>")
		_, err := f.client().Search(t.Context(), "x")
		require.ErrorContains(t, err, "invalid API JSON")
	})
	t.Run("cancelled", func(t *testing.T) {
		f := newFakePluto(t)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		_, err := f.client().Search(ctx, "x")
		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

// ---- seasons and episodes --------------------------------------------------

func (f *fakePluto) serveShow(body string) {
	f.set("/v4/vod/series/"+fakeSeriesID+"/seasons", body)
}

func TestShowOrdersSeasonsAndEpisodes(t *testing.T) {
	f := newFakePluto(t)
	f.serveShow(seasonsBody(fakeSeriesID, "Death Note",
		seasonOf(2, episodeOf(oid('5'), "S2E2", 2, 2, 1000), episodeOf(oid('4'), "S2E1", 1, 2, 1000)),
		seasonOf(1, episodeOf(oid('2'), "S1E2", 2, 1, 1000), episodeOf(oid('1'), "S1E1", 1, 1, 1000)),
	))
	show, err := f.client().Show(t.Context(), SeriesURL(fakeSeriesID))
	require.NoError(t, err)
	assert.Equal(t, fakeSeriesID, show.ID)
	assert.Equal(t, "Death Note", show.Name)
	require.Len(t, show.Seasons, 2)
	assert.Equal(t, 1, show.Seasons[0].Number)
	assert.Equal(t, 2, show.Seasons[1].Number)
	var names []string
	for _, s := range show.Seasons {
		for _, e := range s.Episodes {
			names = append(names, e.Name)
		}
	}
	assert.Equal(t, []string{"S1E1", "S1E2", "S2E1", "S2E2"}, names)
}

func TestShowFillsEveryFieldOfAnEpisode(t *testing.T) {
	f := newFakePluto(t)
	f.serveShow(seasonsBody(fakeSeriesID, "  Death Note ",
		seasonOf(3, episodeOf(oid('1'), " Renascimento ", 7, 3, 1395000))))
	show, err := f.client().Show(t.Context(), SeriesURL(fakeSeriesID))
	require.NoError(t, err)
	assert.Equal(t, "Death Note", show.Name, "the name is trimmed")
	e := show.Seasons[0].Episodes[0]
	assert.Equal(t, Episode{
		ID: oid('1'), Name: "Renascimento", Description: "about  Renascimento", Number: 7, Season: 3,
		Duration: 1395 * time.Second,
	}, e, "the duration is the content's own, in milliseconds on the wire")
}

func TestShowAsksForTheWholeSeriesAtOnce(t *testing.T) {
	f := newFakePluto(t)
	f.serveShow(seasonsBody(fakeSeriesID, "x", seasonOf(1, episodeOf(oid('1'), "a", 1, 1, 1))))
	_, err := f.client().Show(t.Context(), SeriesURL(fakeSeriesID))
	require.NoError(t, err)
	r := f.lastRequest("/v4/vod/series/" + fakeSeriesID + "/seasons")
	assert.Equal(t, "1000", r.URL.Query().Get("offset"))
	assert.Equal(t, "1", r.URL.Query().Get("page"))
	assert.Equal(t, "Bearer tok-1", r.Header.Get("Authorization"))
}

func TestShowNumbersEpisodesTheServiceLeftUnnumbered(t *testing.T) {
	cases := []struct {
		name    string
		numbers []int
		want    []int
	}{
		{"missing numbers", []int{0, 0, 0}, []int{1, 2, 3}},
		{"repeated numbers", []int{1, 1, 2}, []int{1, 2, 3}},
		{"a negative number", []int{-1, 2, 3}, []int{1, 2, 3}},
		{"sound numbers are kept, gaps included", []int{2, 5, 9}, []int{2, 5, 9}},
		{"out of order numbers are sorted", []int{3, 1, 2}, []int{1, 2, 3}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFakePluto(t)
			var eps []string
			for i, n := range tc.numbers {
				eps = append(eps, episodeOf(oid(byte('1'+i)), fmt.Sprintf("e%d", i), n, 1, 1))
			}
			f.serveShow(seasonsBody(fakeSeriesID, "x", seasonOf(1, eps...)))
			show, err := f.client().Show(t.Context(), SeriesURL(fakeSeriesID))
			require.NoError(t, err)
			var got []int
			for _, e := range show.Seasons[0].Episodes {
				got = append(got, e.Number)
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestShowSkipsWhatCannotBePlayed(t *testing.T) {
	f := newFakePluto(t)
	f.serveShow(seasonsBody(fakeSeriesID, "x",
		seasonOf(1,
			episodeOf("bad-id", "bad", 1, 1, 1),
			episodeOf(oid('1'), "kept", 2, 1, 1),
			episodeOf(oid('1'), "same id", 3, 1, 1),
		),
		seasonOf(2, episodeOf(oid('1'), "same id in another season", 1, 2, 1)),
		seasonOf(3), // a season announced and not aired
	))
	show, err := f.client().Show(t.Context(), SeriesURL(fakeSeriesID))
	require.NoError(t, err)
	require.Len(t, show.Seasons, 1, "an empty season is not offered, nor a season of repeats")
	require.Len(t, show.Seasons[0].Episodes, 1)
	assert.Equal(t, "kept", show.Seasons[0].Episodes[0].Name)
}

func TestShowRefusals(t *testing.T) {
	t.Run("a movie has no seasons", func(t *testing.T) {
		f := newFakePluto(t)
		_, err := f.client().Show(t.Context(), MovieURL(fakeMovieID))
		require.ErrorContains(t, err, "no seasons")
		assert.Zero(t, f.bootCount())
	})
	t.Run("a foreign URL", func(t *testing.T) {
		f := newFakePluto(t)
		_, err := f.client().Show(t.Context(), "https://example.com/")
		require.Error(t, err)
		assert.Zero(t, f.bootCount())
	})
	t.Run("an answer for another series", func(t *testing.T) {
		f := newFakePluto(t)
		f.serveShow(seasonsBody(oid('9'), "Other", seasonOf(1, episodeOf(oid('1'), "a", 1, 1, 1))))
		_, err := f.client().Show(t.Context(), SeriesURL(fakeSeriesID))
		require.ErrorContains(t, err, "different series")
	})
	t.Run("no episodes at all", func(t *testing.T) {
		f := newFakePluto(t)
		f.serveShow(seasonsBody(fakeSeriesID, "x"))
		_, err := f.client().Show(t.Context(), SeriesURL(fakeSeriesID))
		require.ErrorContains(t, err, "no episodes")
	})
	t.Run("not found", func(t *testing.T) {
		f := newFakePluto(t)
		_, err := f.client().Show(t.Context(), SeriesURL(fakeSeriesID))
		require.Error(t, err)
	})
}

// ---- stream ----------------------------------------------------------------

var episodeURL = EpisodeURL(fakeSeriesID, fakeEpisodeID)

func TestStreamReturnsTheTitlesOwnPlaylist(t *testing.T) {
	f := newFakePluto(t)
	f.serveEpisode(fakeEpisodeID)
	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	assert.True(t, stream.AdFree)
	assert.Equal(t, f.url()+f.contentPath("hls_2400"), stream.URL, "the best rung, the title alone")
	assert.NotContains(t, stream.URL, "/v2/stitch/", "not the stitched playlist")
}

func TestStreamAsksTheStitcherTheWayTheSiteDoes(t *testing.T) {
	f := newFakePluto(t)
	f.tokenFmt = "a+b/c=d-%d" // characters a query string would mangle
	f.serveEpisode(fakeEpisodeID)
	_, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err, "the token survives being put in a URL")

	q := f.lastRequest(f.masterPath(fakeEpisodeID)).URL.Query()
	assert.Equal(t, "a+b/c=d-1", q.Get("jwt"))
	assert.Equal(t, fakeSID, q.Get("sid"))
	assert.Equal(t, "true", q.Get("masterJWTPassthrough"))
	assert.Equal(t, "false", q.Get("serverSideAds"), "the stitcher params the service handed out are all sent")
	assert.Equal(t, 1, f.bootCount())
}

func TestStreamPicksTheRungQualityAsksFor(t *testing.T) {
	cases := []struct {
		quality string
		want    string
	}{
		{"best", "hls_2400"},
		{"BEST", "hls_2400"},
		{"", "hls_2400"},
		{"nonsense", "hls_2400"},
		{"worst", "hls_300"},
		{" Worst ", "hls_300"},
		{"1080p", "hls_2400"},
		{"720p", "hls_2400"},
		{"576p", "hls_1600"},
		{"480p", "hls_1000"},
		{"360p", "hls_600"},
		{"240p", "hls_300"},
		{"144p", "hls_300"},
		{"720", "hls_2400"},
	}
	for _, tc := range cases {
		t.Run(tc.quality, func(t *testing.T) {
			f := newFakePluto(t)
			f.serveEpisode(fakeEpisodeID)
			stream, err := f.client().Stream(t.Context(), episodeURL, tc.quality)
			require.NoError(t, err)
			assert.Equal(t, f.url()+f.contentPath(tc.want), stream.URL)
			assert.True(t, stream.AdFree)
		})
	}
}

func TestStreamPlaysAMovieByItsOwnID(t *testing.T) {
	f := newFakePluto(t)
	f.serveEpisode(fakeMovieID)
	stream, err := f.client().Stream(t.Context(), MovieURL(fakeMovieID), "best")
	require.NoError(t, err)
	assert.True(t, stream.AdFree)
	assert.NotNil(t, f.lastRequest(f.masterPath(fakeMovieID)))
}

func TestStreamPlaysAnEpisodeByTheEpisodesIDNotTheSeries(t *testing.T) {
	f := newFakePluto(t)
	f.serveEpisode(fakeEpisodeID)
	f.set(f.masterPath(fakeSeriesID), "not the one")
	_, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	assert.Zero(t, f.hitCount(f.masterPath(fakeSeriesID)))
}

func TestStreamOfASeriesIsRefused(t *testing.T) {
	f := newFakePluto(t)
	_, err := f.client().Stream(t.Context(), SeriesURL(fakeSeriesID), "best")
	require.ErrorContains(t, err, "names no episode")
	assert.Zero(t, f.bootCount())
}

func TestStreamOfAForeignURLIsRefused(t *testing.T) {
	f := newFakePluto(t)
	_, err := f.client().Stream(t.Context(), "https://example.com/x.m3u8", "best")
	require.Error(t, err)
	assert.Zero(t, f.bootCount())
}

func TestStreamFallsBackToTheStitchedPlaylistWhenTheTitleIsSeveralClips(t *testing.T) {
	f := newFakePluto(t)
	f.serveEpisode(fakeEpisodeID)
	a := f.stitched("hls_2400", 4, "0-100", "101-end")
	b := strings.ReplaceAll(f.stitched("hls_2400", 4, "0-100", "101-end"), "abc_Title", "other_Title")
	f.set(f.variantPath(fakeEpisodeID, 3321280), a+b)

	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	assert.False(t, stream.AdFree, "the result says the breaks are still in it")
	assert.Contains(t, stream.URL, "/v2/stitch/hls/episode/"+fakeEpisodeID+"/master.m3u8", "the stitched master, so it still plays")
	assert.Zero(t, f.hitCount(f.contentPath("hls_2400")), "no guess is made at a playlist")
}

func TestStreamTriesTheNextRungWhenOneCannotBeUsed(t *testing.T) {
	t.Run("its stitched playlist is gone", func(t *testing.T) {
		f := newFakePluto(t)
		f.serveEpisode(fakeEpisodeID)
		f.setRoute(f.variantPath(fakeEpisodeID, 3321280), http.StatusInternalServerError, "x")
		stream, err := f.client().Stream(t.Context(), episodeURL, "best")
		require.NoError(t, err)
		assert.Equal(t, f.url()+f.contentPath("hls_1600"), stream.URL)
	})
}

func TestStreamTriesAtMostThreeRungsAndThenPlaysTheStitchedMaster(t *testing.T) {
	f := newFakePluto(t)
	f.serveEpisode(fakeEpisodeID)
	for _, rung := range fakeLadder {
		f.setRoute(f.variantPath(fakeEpisodeID, rung.bandwidth), http.StatusInternalServerError, "x")
	}
	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	assert.False(t, stream.AdFree)
	assert.Contains(t, stream.URL, "master.m3u8")

	assert.Equal(t, 1, f.hitCount(f.variantPath(fakeEpisodeID, 3321280)), "best first")
	assert.Equal(t, 1, f.hitCount(f.variantPath(fakeEpisodeID, 2294080)))
	assert.Equal(t, 1, f.hitCount(f.variantPath(fakeEpisodeID, 1577180)))
	assert.Zero(t, f.hitCount(f.variantPath(fakeEpisodeID, 1042180)), "the fourth is not tried")
	assert.Zero(t, f.hitCount(f.variantPath(fakeEpisodeID, 640930)))
}

func TestStreamStopsWhenTheCallerGivesUp(t *testing.T) {
	f := newFakePluto(t)
	f.serveEpisode(fakeEpisodeID)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	for _, rung := range fakeLadder {
		f.setRoute(f.variantPath(fakeEpisodeID, rung.bandwidth), http.StatusInternalServerError, "x")
	}
	f.onHit = func(path string) {
		if strings.HasSuffix(path, "/playlist.m3u8") {
			cancel()
		}
	}
	_, err := f.client().Stream(ctx, episodeURL, "best")
	require.ErrorIs(t, err, context.Canceled, "not a stitched playlist handed back to a caller who left")
	total := 0
	for _, rung := range fakeLadder {
		total += f.hitCount(f.variantPath(fakeEpisodeID, rung.bandwidth))
	}
	assert.Equal(t, 1, total, "no other rung is tried")
}

func TestStreamStopsWhenTheCallerGivesUpOnTheContentPlaylist(t *testing.T) {
	f := newFakePluto(t)
	f.serveEpisode(fakeEpisodeID, 4) // one rung: nothing after it would notice the caller left
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	f.setRoute(f.contentPath("hls_2400"), http.StatusInternalServerError, "x")
	f.onHit = func(path string) {
		if strings.Contains(path, "/0-end/") {
			cancel()
		}
	}
	_, err := f.client().Stream(ctx, episodeURL, "best")
	require.ErrorIs(t, err, context.Canceled)
	assert.Zero(t, f.hitCount(f.contentPath("hls_1600")))
}

func TestStreamRefusesWhatIsProtected(t *testing.T) {
	protectedKey := "#EXT-X-KEY:METHOD=SAMPLE-AES,URI=\"skd://key\",KEYFORMAT=\"com.apple.streamingkeydelivery\"\n"
	t.Run("in the master", func(t *testing.T) {
		f := newFakePluto(t)
		f.serveEpisode(fakeEpisodeID)
		f.set(f.masterPath(fakeEpisodeID), strings.Replace(masterOf(0, 1, 2, 3, 4), "#EXT-X-VERSION:5\n", "#EXT-X-VERSION:5\n#EXT-X-SESSION-KEY:METHOD=SAMPLE-AES,URI=\"skd://key\"\n", 1))
		_, err := f.client().Stream(t.Context(), episodeURL, "best")
		require.ErrorIs(t, err, ErrProtected)
	})
	t.Run("in the stitched playlist", func(t *testing.T) {
		f := newFakePluto(t)
		f.serveEpisode(fakeEpisodeID)
		f.set(f.variantPath(fakeEpisodeID, 3321280), strings.Replace(f.stitched("hls_2400", 3, "0-100", "101-end"), "#EXT-X-MEDIA-SEQUENCE:0\n", "#EXT-X-MEDIA-SEQUENCE:0\n"+protectedKey, 1))
		_, err := f.client().Stream(t.Context(), episodeURL, "best")
		require.ErrorIs(t, err, ErrProtected)
	})
	t.Run("in the title's own playlist", func(t *testing.T) {
		f := newFakePluto(t)
		f.serveEpisode(fakeEpisodeID)
		f.set(f.contentPath("hls_2400"), strings.Replace(content("hls_2400", 5), "#EXT-X-MEDIA-SEQUENCE:0\n", "#EXT-X-MEDIA-SEQUENCE:0\n"+protectedKey, 1))
		_, err := f.client().Stream(t.Context(), episodeURL, "best")
		require.ErrorIs(t, err, ErrProtected, "no other rung is offered in its place")
		assert.Zero(t, f.hitCount(f.contentPath("hls_1600")))
	})
	t.Run("the refusal is not a parser fault", func(t *testing.T) {
		var diag *netx.SourceDiagnostic
		assert.False(t, errors.As(ErrProtected, &diag))
	})
}

func TestStreamDoesNotMistakeAKeyedPathForAProtectedPlaylist(t *testing.T) {
	// The service names its encoded renditions "720pDRM" while keying them with
	// plain AES-128 (2026-11-16); only the key method decides.
	f := newFakePluto(t)
	f.serveEpisode(fakeEpisodeID)
	f.set(f.variantPath(fakeEpisodeID, 3321280), strings.ReplaceAll(f.stitched("hls_2400", 3, "0-100", "101-end"), "720p/20220214", "720pDRM/20250723"))
	f.set(f.contentPath("hls_2400"), content("hls_2400", 5))
	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	assert.True(t, stream.AdFree)
}

func TestStreamErrors(t *testing.T) {
	t.Run("the master is missing", func(t *testing.T) {
		f := newFakePluto(t)
		_, err := f.client().Stream(t.Context(), episodeURL, "best")
		require.Error(t, err)
		var diag *netx.SourceDiagnostic
		assert.ErrorAs(t, err, &diag)
	})
	t.Run("the master lists no rendition", func(t *testing.T) {
		f := newFakePluto(t)
		f.set(f.masterPath(fakeEpisodeID), "#EXTM3U\n#EXT-X-VERSION:5\n")
		_, err := f.client().Stream(t.Context(), episodeURL, "best")
		require.ErrorContains(t, err, "no renditions")
	})
	t.Run("the master is too large", func(t *testing.T) {
		f := newFakePluto(t)
		f.set(f.masterPath(fakeEpisodeID), strings.Repeat("#", maxPlaylistBytes+1))
		_, err := f.client().Stream(t.Context(), episodeURL, "best")
		require.ErrorContains(t, err, "size limit")
	})
	t.Run("the master is exactly at the limit", func(t *testing.T) {
		f := newFakePluto(t)
		f.serveEpisode(fakeEpisodeID)
		f.set(f.masterPath(fakeEpisodeID), masterOf(4)+strings.Repeat("#", maxPlaylistBytes-len(masterOf(4))))
		_, err := f.client().Stream(t.Context(), episodeURL, "best")
		require.NoError(t, err)
	})
}

func TestStreamIsRenewedAfterA401(t *testing.T) {
	f := newFakePluto(t)
	f.serveEpisode(fakeEpisodeID)
	f.reject = 1
	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	assert.True(t, stream.AdFree)
	assert.Equal(t, 2, f.bootCount())
	assert.Equal(t, "tok-2", f.lastRequest(f.masterPath(fakeEpisodeID)).URL.Query().Get("jwt"), "the retry carries the new token")
}

func TestStreamNeverFollowsAHostThatIsNotPlutos(t *testing.T) {
	t.Run("a rendition on another host", func(t *testing.T) {
		f := newFakePluto(t)
		f.serveEpisode(fakeEpisodeID)
		spy := newSpyHost(t, f.stitched("hls_2400", 3, "0-100", "101-end")) // healthy, as a trap
		master := regexp.MustCompile(`(?m)^(\d+/playlist\.m3u8)`).ReplaceAllString(masterOf(0, 1, 2, 3, 4), spy.URL+"/$1")
		f.set(f.masterPath(fakeEpisodeID), master)

		stream, err := f.client().Stream(t.Context(), episodeURL, "best")
		require.NoError(t, err)
		assert.Zero(t, spy.hits.Load(), "no request is made to a host that is not Pluto's")
		assert.False(t, stream.AdFree)
		assert.Contains(t, stream.URL, f.masterPath(fakeEpisodeID), "the stitched master is Pluto's own")
	})
	t.Run("a title clip on another host", func(t *testing.T) {
		f := newFakePluto(t)
		f.serveEpisode(fakeEpisodeID)
		spy := newSpyHost(t, content("hls_2400", 5)) // a healthy media playlist, as a trap
		for _, rung := range fakeLadder {
			body := strings.ReplaceAll(f.stitched(rung.rendition, 3, "0-100", "101-end"), f.url()+fakeClip, spy.URL+"/clip/hls/")
			f.set(f.variantPath(fakeEpisodeID, rung.bandwidth), body)
		}
		stream, err := f.client().Stream(t.Context(), episodeURL, "best")
		require.NoError(t, err)
		assert.Zero(t, spy.hits.Load(), "a content playlist on a host outside Pluto is not even fetched")
		assert.False(t, stream.AdFree)
		assert.NotContains(t, stream.URL, spy.URL, "nor handed to the player")
	})
}

// ---- playlist reading ------------------------------------------------------

func TestParseMaster(t *testing.T) {
	const base = "https://stitch.pluto.tv/v2/stitch/hls/episode/abc/master.m3u8?jwt=T"
	t.Run("relative and absolute rungs", func(t *testing.T) {
		got := parseMaster("#EXTM3U\n"+
			"#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=1000,RESOLUTION=640x360\n1000/playlist.m3u8?sid=S\n"+
			"#EXT-X-STREAM-INF:BANDWIDTH=2000\n\n# a comment\nhttps://other.pluto.tv/2000.m3u8\n", base)
		assert.Equal(t, []variant{
			{URL: "https://stitch.pluto.tv/v2/stitch/hls/episode/abc/1000/playlist.m3u8?sid=S", Bandwidth: 1000, Height: 360},
			{URL: "https://other.pluto.tv/2000.m3u8", Bandwidth: 2000},
		}, got)
	})
	t.Run("windows line endings", func(t *testing.T) {
		got := parseMaster("#EXTM3U\r\n#EXT-X-STREAM-INF:BANDWIDTH=7\r\nr.m3u8\r\n", base)
		require.Len(t, got, 1)
		assert.Equal(t, 7, got[0].Bandwidth)
		assert.NotContains(t, got[0].URL, "\r")
	})
	t.Run("a rung without a URI is dropped", func(t *testing.T) {
		assert.Empty(t, parseMaster("#EXT-X-STREAM-INF:BANDWIDTH=7\n", base))
	})
	t.Run("a URI without a rung is not one", func(t *testing.T) {
		assert.Empty(t, parseMaster("#EXTM3U\nstray.m3u8\n", base))
	})
	t.Run("an audio rendition is not a rung", func(t *testing.T) {
		assert.Empty(t, parseMaster("#EXT-X-MEDIA:TYPE=AUDIO,BANDWIDTH=5,URI=\"a.m3u8\"\n", base))
	})
	t.Run("a bad base", func(t *testing.T) {
		assert.Nil(t, parseMaster("#EXT-X-STREAM-INF:BANDWIDTH=7\nr.m3u8\n", "%%%"))
	})
	t.Run("a rung that names no bandwidth", func(t *testing.T) {
		got := parseMaster("#EXT-X-STREAM-INF:PROGRAM-ID=1\nr.m3u8\n", base)
		require.Len(t, got, 1)
		assert.Zero(t, got[0].Bandwidth)
	})
}

func TestBandwidthBudget(t *testing.T) {
	cases := []struct {
		height int
		want   int
	}{
		{144, 800_000}, {240, 800_000}, {359, 800_000},
		{360, 1_300_000}, {479, 1_300_000},
		{480, 2_000_000}, {575, 2_000_000},
		{576, 2_800_000}, {719, 2_800_000},
		{720, int(^uint(0) >> 1)}, {1080, int(^uint(0) >> 1)},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, bandwidthBudget(tc.height), "%dp", tc.height)
	}
}

func TestOrderVariants(t *testing.T) {
	ladder := []variant{
		{URL: "a", Bandwidth: 500, Height: 240},
		{URL: "d", Bandwidth: 3000, Height: 720},
		{URL: "b", Bandwidth: 1000, Height: 360},
		{URL: "c", Bandwidth: 2000, Height: 480},
	}
	urls := func(vs []variant) string {
		var b strings.Builder
		for _, v := range vs {
			b.WriteString(v.URL)
		}
		return b.String()
	}
	cases := []struct{ quality, want string }{
		{"best", "dcba"},
		{"", "dcba"},
		{"worst", "abcd"},
		{"720p", "dcba"},
		{"1080p", "dcba"},
		{"480p", "cbad"},
		{"480", "cbad"},
		{"400p", "bacd"},
		{"360p", "bacd"},
		{"100p", "abcd"},
		{"WORST", "abcd"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, urls(orderVariants(ladder, tc.quality)), tc.quality)
	}
	assert.Equal(t, "a", ladder[0].URL, "the caller's slice is left as it was")

	t.Run("without resolutions the bitrate budget stands in", func(t *testing.T) {
		bare := []variant{{URL: "a", Bandwidth: 640930}, {URL: "b", Bandwidth: 1042180}, {URL: "c", Bandwidth: 1577180}, {URL: "d", Bandwidth: 2294080}, {URL: "e", Bandwidth: 3321280}}
		assert.Equal(t, "cbade", urls(orderVariants(bare, "480p")), "what fits, best first, then the nearest above first")
		assert.Equal(t, "c", urls(orderVariants(bare, "480p"))[:1])
		assert.Equal(t, "b", urls(orderVariants(bare, "360p"))[:1])
		assert.Equal(t, "a", urls(orderVariants(bare, "240p"))[:1])
		assert.Equal(t, "e", urls(orderVariants(bare, "720p"))[:1])
	})
	t.Run("ties keep the master's order", func(t *testing.T) {
		tied := []variant{{URL: "x", Bandwidth: 5}, {URL: "y", Bandwidth: 5}, {URL: "z", Bandwidth: 5}}
		assert.Equal(t, "xyz", urls(orderVariants(tied, "best")))
	})
	t.Run("nothing", func(t *testing.T) {
		assert.Empty(t, orderVariants(nil, "best"))
	})
}

func TestIsProtected(t *testing.T) {
	cases := []struct {
		name string
		body string
		want bool
	}{
		{"clear", "#EXTM3U\n#EXTINF:5,\na.ts\n", false},
		{"AES-128", "#EXT-X-KEY:METHOD=AES-128,URI=\"k\",IV=0x1\n", false},
		{"NONE", "#EXT-X-KEY:METHOD=NONE\n", false},
		{"none, in lower case", "#EXT-X-KEY:METHOD=none\n", false},
		{"SAMPLE-AES", "#EXT-X-KEY:METHOD=SAMPLE-AES,URI=\"skd://k\"\n", true},
		{"SAMPLE-AES in lower case", "#EXT-X-KEY:METHOD=sample-aes\n", true},
		{"SAMPLE-AES-CTR", "#EXT-X-KEY:METHOD=SAMPLE-AES-CTR\n", true},
		{"a session key", "#EXT-X-SESSION-KEY:METHOD=SAMPLE-AES\n", true},
		{"a clear key after a protected one", "#EXT-X-KEY:METHOD=NONE\n#EXT-X-KEY:METHOD=SAMPLE-AES\n", true},
		{"a protected key after clear ones", strings.Repeat("#EXT-X-KEY:METHOD=AES-128\n", 50) + "#EXT-X-KEY:METHOD=SAMPLE-AES\n", true},
		{"the word in a URI is not a method", "#EXT-X-KEY:METHOD=AES-128,URI=\"x/METHOD=SAMPLE-AES\"\n", false},
		{"the word in a segment name", "#EXTINF:5,\nhttps://x/720pDRM/METHOD=SAMPLE-AES.ts\n", false},
		{"a key that names a path with DRM", "#EXT-X-KEY:METHOD=AES-128,URI=\"https://x/720pDRM/k.key\"\n", false},
		{"a key tag with no method", "#EXT-X-KEY:URI=\"k\"\n", false},
		{"empty", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isProtected(tc.body))
		})
	}
}

// ---- the title's own playlist ----------------------------------------------

// clipSegments is `n` segment lines of one clip range.
func clipSegments(dir, rng, rendition string, n int) string {
	var b strings.Builder
	for i := range n {
		fmt.Fprintf(&b, "#EXTINF:5,\nhttps://siloh.plutotv.net/%s/hls/%s/%s-%05d.ts\n", dir, rng, rendition, i)
	}
	return b.String()
}

const adSegment = "#EXT-X-DISCONTINUITY\n#EXTINF:5,\nhttps://siloh.plutotv.net/v1/mp4/c(ts)/p(1_ad%2Fcreative%2Fx%2Fvideo_600.mp4)\n"

func TestContentPlaylistURL(t *testing.T) {
	const want = "https://siloh.plutotv.net/538/clip/T/hls/0-end/hls_600.m3u8"
	cases := []struct {
		name string
		body string
		ok   bool
	}{
		{"a title cut in three with ads between", clipSegments("538/clip/T", "0-100", "hls_600", 20) + adSegment + clipSegments("538/clip/T", "101-200", "hls_600", 20) + adSegment + clipSegments("538/clip/T", "201-end", "hls_600", 20), true},
		{"a title that was never cut", clipSegments("538/clip/T", "0-end", "hls_600", 20), true},
		{"a closing bumper of one segment", clipSegments("538/clip/T", "0-100", "hls_600", 20) + clipSegments("538/clip/T", "101-end", "hls_600", 20) + clipSegments("999/bumper", "0-end", "hls_300", 1), true},
		{"a bumper of the longest accepted length", clipSegments("538/clip/T", "0-end", "hls_600", 20) + clipSegments("999/bumper", "0-end", "hls_300", maxInterstitialSegments), true},
		{"a bumper one segment too long", clipSegments("538/clip/T", "0-end", "hls_600", 20) + clipSegments("999/bumper", "0-end", "hls_300", maxInterstitialSegments+1), false},
		{"two clips of the same weight", clipSegments("538/clip/T", "0-end", "hls_600", 20) + clipSegments("538/clip/U", "0-end", "hls_600", 20), false},
		{"a title that does not start at its start", clipSegments("538/clip/T", "50-end", "hls_600", 20), false},
		{"a title that stops before its end", clipSegments("538/clip/T", "0-100", "hls_600", 20), false},
		{"a title with a hole in the middle is not told apart from a whole one", clipSegments("538/clip/T", "0-100", "hls_600", 20) + clipSegments("538/clip/T", "300-end", "hls_600", 20), true},
		{"only ads", adSegment + adSegment, false},
		{"empty", "", false},
		{"a segment with a query string", "#EXTINF:5,\nhttps://siloh.plutotv.net/538/clip/T/hls/0-end/hls_600-00000.ts?x=1\n", false},
		{"a relative segment", "#EXTINF:5,\nhls/0-end/hls_600-00000.ts\n", false},
		{"windows line endings", strings.ReplaceAll(clipSegments("538/clip/T", "0-end", "hls_600", 5), "\n", "\r\n"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := contentPlaylistURL(tc.body)
			assert.Equal(t, tc.ok, ok)
			if tc.ok {
				assert.Equal(t, want, got)
			} else {
				assert.Empty(t, got)
			}
		})
	}
}

func TestContentPlaylistURLNamesTheRenditionTheSegmentsAre(t *testing.T) {
	for _, rendition := range []string{"hls_300", "hls_1000", "hls_2400"} {
		got, ok := contentPlaylistURL(clipSegments("a/clip/T", "0-end", rendition, 10))
		require.True(t, ok)
		assert.Equal(t, "https://siloh.plutotv.net/a/clip/T/hls/0-end/"+rendition+".m3u8", got)
	}
}

func TestContentPlaylistURLKeepsTheTitlesOwnDirectoryWhateverItIsCalled(t *testing.T) {
	got, ok := contentPlaylistURL(clipSegments("6880f1fa_4318-3fb5/720pDRM/20250723_143020", "0-end", "hls_600", 10))
	require.True(t, ok)
	assert.Equal(t, "https://siloh.plutotv.net/6880f1fa_4318-3fb5/720pDRM/20250723_143020/hls/0-end/hls_600.m3u8", got)
}

func TestCheckMediaPlaylist(t *testing.T) {
	f := newFakePluto(t)
	c := f.client()
	cases := []struct {
		name    string
		body    string
		status  int
		wantErr string
	}{
		{"a media playlist", content("hls_300", 3), 200, ""},
		{"a playlist with leading blank space", "\n  " + content("hls_300", 3), 200, ""},
		{"a page", "<html>x</html>", 200, "not a media playlist"},
		{"a master", masterOf(0), 200, "not a media playlist"},
		{"no media", "#EXTM3U\n", 200, "not a media playlist"},
		{"forbidden", "x", 403, "HTTP 403"},
		{"protected", "#EXTM3U\n#EXT-X-KEY:METHOD=SAMPLE-AES\n#EXTINF:5,\na.ts\n", 200, ErrProtected.Error()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f.setRoute("/media/p.m3u8", tc.status, tc.body)
			body, err := c.checkMediaPlaylist(t.Context(), f.url()+"/media/p.m3u8")
			if tc.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.body, body, "the playlist it checked is the one it hands back")
				return
			}
			require.ErrorContains(t, err, tc.wantErr)
			assert.Empty(t, body)
		})
	}
}

// ---- the fake itself -------------------------------------------------------

// The tests above lean on the fake's refusals: a token that is wrong, a media
// request that names no User-Agent. These hold it to them.
func TestTheFakeEnforcesWhatTheRealServiceDoes(t *testing.T) {
	f := newFakePluto(t)
	f.serveEpisode(fakeEpisodeID)
	f.set("/v1/search", searchBody())
	_, err := f.client().Search(t.Context(), "x") // opens a session: the token is tok-1
	require.NoError(t, err)

	get := func(target string, header http.Header) int {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, f.url()+target, http.NoBody)
		require.NoError(t, err)
		for k, v := range header {
			req.Header[k] = v
		}
		resp, err := f.srv.Client().Do(req)
		require.NoError(t, err)
		_ = resp.Body.Close()
		return resp.StatusCode
	}
	master := f.masterPath(fakeEpisodeID)
	good := "?sid=S&masterJWTPassthrough=true&jwt=tok-1"

	assert.Equal(t, 200, get("/v1/search", http.Header{"Authorization": {"Bearer tok-1"}}))
	assert.Equal(t, 401, get("/v1/search", nil), "no token")
	assert.Equal(t, 401, get("/v1/search", http.Header{"Authorization": {"Bearer tok-0"}}), "a stale token")
	assert.Equal(t, 401, get("/v1/search", http.Header{"Authorization": {"tok-1"}}), "not a Bearer")

	assert.Equal(t, 200, get(master+good, nil))
	assert.Equal(t, 401, get(master+"?sid=S&masterJWTPassthrough=true&jwt=tok-0", nil), "a stale jwt")
	assert.Equal(t, 400, get(master+"?masterJWTPassthrough=true&jwt=tok-1", nil), "no session id")
	assert.Equal(t, 400, get(master+"?sid=S&jwt=tok-1", nil), "no passthrough flag")

	// Go's client adds a User-Agent of its own; name none to see the 403.
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, f.url()+f.contentPath("hls_2400"), http.NoBody)
	require.NoError(t, err)
	req.Header.Set("User-Agent", "")
	resp, err := f.srv.Client().Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, 403, resp.StatusCode, "the CDN refuses a request that names no User-Agent")
	assert.Equal(t, 200, get(f.contentPath("hls_2400"), http.Header{"User-Agent": {"libmpv"}}))

	f.reject = 1
	assert.Equal(t, 401, get("/v1/search", http.Header{"Authorization": {"Bearer tok-1"}}), "a forced rejection")
	assert.Equal(t, 200, get("/v1/search", http.Header{"Authorization": {"Bearer tok-1"}}), "and only the one")
}

func TestTheFakesStitchedPlaylistHasTheShapeOfTheRealOne(t *testing.T) {
	f := newFakePluto(t)
	body := f.stitched("hls_600", 3, "0-100", "101-200", "201-end")
	got, ok := contentPlaylistURL(body)
	require.True(t, ok)
	assert.Equal(t, f.url()+f.contentPath("hls_600"), got)
	assert.Equal(t, 2, strings.Count(body, "ad%2Fcreative"), "an ad between each pair of ranges")
	assert.Equal(t, 5, strings.Count(body, "#EXT-X-DISCONTINUITY\n"), "the splices the real one marks")
	assert.Contains(t, body, "METHOD=NONE")
	assert.Contains(t, body, "METHOD=AES-128")
	assert.False(t, isProtected(body))
}
