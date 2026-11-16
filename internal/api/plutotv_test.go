package api

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/plutotv"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

const (
	ptSeriesID   = "6206bba315c8610013683974"
	ptMovieID    = "62b9ea9ec4eb8f00140cb633"
	ptEpisode1   = "6206bbaa15c8610013683a28"
	ptEpisode2   = "6206bbb015c8610013683d63"
	ptEpisode3   = "6206bbb015c8610013683d64"
	ptSessionTok = "tok"
)

// plutoAPIFake is a Pluto TV with one series (two seasons), one movie and two
// rungs, enough to drive the api layer end to end without the network.
type plutoAPIFake struct {
	srv     *httptest.Server
	mu      sync.Mutex
	boots   int
	hits    map[string]int
	seasons string
	// stitchedBody, when set, replaces the stitched playlist (for a title the
	// reader cannot make an ad-free playlist of).
	stitchedBody string
}

func newPlutoAPIFake(t *testing.T) *plutoAPIFake {
	t.Helper()
	f := &plutoAPIFake{hits: map[string]int{}}
	f.seasons = fmt.Sprintf(`{"_id": %q, "name": "Death Note", "seasons": [
	  {"number": 1, "episodes": [
	    {"_id": %q, "name": "Renascimento", "description": "d1", "number": 1, "season": 1, "originalContentDuration": 1395000},
	    {"_id": %q, "name": "Confronto", "description": "d2", "number": 2, "season": 1, "originalContentDuration": 1410000}]},
	  {"number": 2, "episodes": [
	    {"_id": %q, "name": "", "description": "", "number": 1, "season": 2, "originalContentDuration": 60000}]}]}`,
		ptSeriesID, ptEpisode1, ptEpisode2, ptEpisode3)
	f.srv = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.srv.Close)

	prev := PlutoTVClientFn
	PlutoTVClientFn = func() *plutotv.Client { return plutotv.NewClientForTest(f.srv.Client(), f.srv.URL) }
	t.Cleanup(func() { PlutoTVClientFn = prev })
	return f
}

func (f *plutoAPIFake) hit(path string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.hits[path]
}

func (f *plutoAPIFake) clip(rendition string) string {
	return "/media/clip/c/720p/hls/0-end/" + rendition + ".m3u8"
}

func (f *plutoAPIFake) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	f.hits[r.URL.Path]++
	f.mu.Unlock()
	path := r.URL.Path
	switch {
	case path == "/v4/start":
		f.mu.Lock()
		f.boots++
		f.mu.Unlock()
		_, _ = fmt.Fprintf(w, `{"sessionToken": %q, "stitcherParams": "appName=web&sid=S", "servers": {"search": %q, "vod": %q, "stitcher": %q}}`,
			ptSessionTok, f.srv.URL, f.srv.URL, f.srv.URL)
	case strings.HasSuffix(path, "/seasons"):
		_, _ = w.Write([]byte(f.seasons))
	case strings.HasSuffix(path, "/master.m3u8"):
		_, _ = w.Write([]byte("#EXTM3U\n" +
			"#EXT-X-STREAM-INF:BANDWIDTH=1042180\n1042180/playlist.m3u8\n" +
			"#EXT-X-STREAM-INF:BANDWIDTH=3321280\n3321280/playlist.m3u8\n"))
	case strings.HasSuffix(path, "/playlist.m3u8"):
		if f.stitchedBody != "" {
			_, _ = w.Write([]byte(f.stitchedBody))
			return
		}
		rendition := "hls_600"
		if strings.Contains(path, "/3321280/") {
			rendition = "hls_2400"
		}
		var b strings.Builder
		for _, rng := range []string{"0-100", "101-end"} {
			for i := 0; i < 8; i++ {
				fmt.Fprintf(&b, "#EXTINF:5,\n%s/media/clip/c/720p/hls/%s/%s-%05d.ts\n", f.srv.URL, rng, rendition, i)
			}
			b.WriteString("#EXT-X-DISCONTINUITY\n#EXTINF:5,\n" + f.srv.URL + "/v1/mp4/ad.ts\n#EXT-X-DISCONTINUITY\n")
		}
		_, _ = w.Write([]byte(b.String()))
	case strings.HasPrefix(path, "/media/"):
		// The title's own playlist, relative like the real one.
		rendition := strings.TrimSuffix(path[strings.LastIndex(path, "/")+1:], ".m3u8")
		_, _ = fmt.Fprintf(w, "#EXTM3U\n#EXT-X-VERSION:3\n#EXTINF:5.0,\n%s-00000.ts\n#EXT-X-ENDLIST\n", rendition)
	default:
		http.NotFound(w, r)
	}
}

func plutoSeries() *models.Anime {
	return &models.Anime{Name: "[TV] [PT-BR] Death Note", URL: plutotv.SeriesURL(ptSeriesID), Source: "PlutoTV", MediaType: models.MediaTypeTV}
}

func plutoMovie() *models.Anime {
	return &models.Anime{Name: "[Movie] [PT-BR] Em Nome do Pai", URL: plutotv.MovieURL(ptMovieID), Source: "PlutoTV", MediaType: models.MediaTypeMovie}
}

// seasonPicker replaces the shared season picker for one test. It records what
// it was shown and answers with `index` and `err`.
type seasonPicker struct {
	called int
	items  []tui.PickItem
	opts   tui.PickOptions
}

func installSeasonPicker(t *testing.T, index int, err error) *seasonPicker {
	t.Helper()
	p := &seasonPicker{}
	prev := sfxSeasonPickFn
	sfxSeasonPickFn = func(items []tui.PickItem, opts tui.PickOptions) (int, error) {
		p.called++
		p.items, p.opts = items, opts
		return index, err
	}
	t.Cleanup(func() { sfxSeasonPickFn = prev })
	return p
}

// ---- listing ---------------------------------------------------------------

func TestPlutoTVMovieIsOneEpisodeAndAsksNothing(t *testing.T) {
	f := newPlutoAPIFake(t)
	picker := installSeasonPicker(t, 0, nil)
	movie := plutoMovie()

	eps, err := GetPlutoTVEpisodes(movie)
	require.NoError(t, err)
	require.Len(t, eps, 1)
	assert.Equal(t, models.Episode{
		Number: "1", Num: 1, URL: movie.URL, SeasonID: "0",
		Title: models.TitleDetails{English: movie.Name, Romaji: movie.Name},
	}, eps[0])
	assert.Zero(t, picker.called)
	assert.Zero(t, f.boots, "a movie needs no catalog call to be listed")
	assert.Zero(t, movie.CurrentSeason)
}

func TestPlutoTVSeriesAsksForASeasonAndListsIt(t *testing.T) {
	newPlutoAPIFake(t)
	picker := installSeasonPicker(t, 1, nil) // the second season
	series := plutoSeries()

	eps, err := GetPlutoTVEpisodes(series)
	require.NoError(t, err)
	assert.Equal(t, 1, picker.called)
	require.Len(t, picker.items, 2)
	assert.Equal(t, "Season 1", picker.items[0].Label)
	assert.Equal(t, "2 episodes", picker.items[0].Details)
	assert.Equal(t, "Season 2", picker.items[1].Label)
	assert.Equal(t, "1 episode", picker.items[1].Details)
	assert.Equal(t, []string{"PlutoTV"}, picker.items[0].Tags)
	assert.Contains(t, picker.opts.Breadcrumb, "Death Note")

	require.Len(t, eps, 1)
	assert.Equal(t, 2, series.CurrentSeason, "the chosen season is remembered, the way the other catalogs do")
	assert.Equal(t, "Episode 1", eps[0].Title.English, "an unnamed episode is named by its number")
	assert.Equal(t, "2", eps[0].SeasonID)
}

func TestPlutoTVEpisodesCarryWhatThePlayerNeeds(t *testing.T) {
	newPlutoAPIFake(t)
	installSeasonPicker(t, 0, nil)
	eps, err := GetPlutoTVEpisodes(plutoSeries())
	require.NoError(t, err)
	require.Len(t, eps, 2)
	assert.Equal(t, models.Episode{
		Number: "1", Num: 1, URL: plutotv.EpisodeURL(ptSeriesID, ptEpisode1), DataID: ptEpisode1, SeasonID: "1",
		Synopsis: "d1", Duration: 1395, Title: models.TitleDetails{English: "Renascimento", Romaji: "Renascimento"},
	}, eps[0])
	assert.Equal(t, 1410, eps[1].Duration, "seconds")
	assert.Equal(t, 2, eps[1].Num)
}

func TestPlutoTVSingleSeasonSkipsThePicker(t *testing.T) {
	f := newPlutoAPIFake(t)
	f.seasons = fmt.Sprintf(`{"_id": %q, "name": "x", "seasons": [{"number": 4, "episodes": [{"_id": %q, "name": "a", "number": 1, "season": 4}]}]}`, ptSeriesID, ptEpisode1)
	picker := installSeasonPicker(t, 0, errors.New("must not be asked"))
	series := plutoSeries()
	eps, err := GetPlutoTVEpisodes(series)
	require.NoError(t, err)
	assert.Zero(t, picker.called)
	assert.Equal(t, 4, series.CurrentSeason)
	assert.Equal(t, "4", eps[0].SeasonID)
}

func TestPlutoTVBackAndCancelReturnToTheSearch(t *testing.T) {
	for name, pickErr := range map[string]error{"back": tui.ErrPickBack, "cancelled": tui.ErrPickCancelled, "wrapped back": fmt.Errorf("x: %w", tui.ErrPickBack)} {
		t.Run(name, func(t *testing.T) {
			newPlutoAPIFake(t)
			installSeasonPicker(t, 0, pickErr)
			series := plutoSeries()
			_, err := GetPlutoTVEpisodes(series)
			assert.ErrorIs(t, err, ErrBackToSearch)
			assert.Zero(t, series.CurrentSeason, "nothing is remembered of a season that was not chosen")
		})
	}
}

func TestPlutoTVAPickerFailureIsReported(t *testing.T) {
	newPlutoAPIFake(t)
	installSeasonPicker(t, 0, errors.New("terminal gone"))
	_, err := GetPlutoTVEpisodes(plutoSeries())
	require.ErrorContains(t, err, "season selection cancelled")
	require.ErrorContains(t, err, "terminal gone")
	assert.NotErrorIs(t, err, ErrBackToSearch)
}

func TestPlutoTVListingNeedsATitle(t *testing.T) {
	newPlutoAPIFake(t)
	for name, media := range map[string]*models.Anime{"nil": nil, "no URL": {Name: "x"}} {
		t.Run(name, func(t *testing.T) {
			_, err := GetPlutoTVEpisodes(media)
			require.Error(t, err)
			_, err = GetPlutoTVSeasonNumbers(media)
			require.Error(t, err)
			_, err = GetPlutoTVSeasonEpisodes(media, 1)
			require.Error(t, err)
		})
	}
}

func TestPlutoTVListingPassesACatalogFailureOn(t *testing.T) {
	f := newPlutoAPIFake(t)
	f.seasons = `{"_id": "` + ptSeriesID + `", "name": "x", "seasons": []}`
	picker := installSeasonPicker(t, 0, nil)
	_, err := GetPlutoTVEpisodes(plutoSeries())
	require.ErrorContains(t, err, "no episodes")
	assert.Zero(t, picker.called, "no picker over nothing")
	_, err = GetPlutoTVSeasonNumbers(plutoSeries())
	require.Error(t, err)
	_, err = GetPlutoTVSeasonEpisodes(plutoSeries(), 1)
	require.Error(t, err)
}

func TestPlutoTVSeasonNumbersAndEpisodesForDownloads(t *testing.T) {
	newPlutoAPIFake(t)
	picker := installSeasonPicker(t, 0, errors.New("a download never asks"))

	nums, err := GetPlutoTVSeasonNumbers(plutoSeries())
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2}, nums)

	series := plutoSeries()
	eps, err := GetPlutoTVSeasonEpisodes(series, 2)
	require.NoError(t, err)
	require.Len(t, eps, 1)
	assert.Equal(t, plutotv.EpisodeURL(ptSeriesID, ptEpisode3), eps[0].URL)
	assert.Equal(t, 2, series.CurrentSeason)
	assert.Zero(t, picker.called)

	_, err = GetPlutoTVSeasonEpisodes(plutoSeries(), 9)
	require.ErrorContains(t, err, "no season 9")
	require.ErrorContains(t, err, "Death Note")
}

// ---- stream ----------------------------------------------------------------

// servedPlaylist reads the playlist a resolved stream points at, the way the
// player will.
func servedPlaylist(t *testing.T, target string) string {
	t.Helper()
	require.True(t, strings.HasPrefix(target, "http://127.0.0.1:"), "served from loopback: %s", target)
	require.True(t, strings.HasSuffix(target, ".m3u8"), "named as a playlist, which is how the player and the downloader tell: %s", target)
	resp, err := http.Get(target) //nolint:noctx // a loopback URL made by the code under test
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/vnd.apple.mpegurl", resp.Header.Get("Content-Type"))
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(body)
}

func TestResolvePlutoTVStreamServesTheTitlesOwnPlaylistFromLoopback(t *testing.T) {
	f := newPlutoAPIFake(t)
	stream, err := ResolvePlutoTVStream(plutoSeries(), &models.Episode{URL: plutotv.EpisodeURL(ptSeriesID, ptEpisode1)}, "best")
	require.NoError(t, err)
	assert.NotEqual(t, f.srv.URL+f.clip("hls_2400"), stream.URL, "not the CDN's own URL: one of the CDNs refuses it")
	assert.True(t, util.IsLocalProxyURL(stream.URL), "registered, so the download paths let it through their loopback guard")

	playlist := servedPlaylist(t, stream.URL)
	dir := strings.TrimSuffix(f.srv.URL+f.clip("hls_2400"), "hls_2400.m3u8")
	assert.True(t, strings.HasPrefix(playlist, "#EXTM3U"))
	assert.Contains(t, playlist, "\n"+dir+"hls_2400-", "its segments are absolute, on the CDN")
	assert.NotContains(t, playlist, "\nhls_2400-", "none is left relative to the loopback server")

	assert.Equal(t, "https://pluto.tv/", stream.Referer, "the site's own, so no other source's referer is reused")
	assert.Empty(t, stream.Origin)
	assert.Empty(t, stream.Subtitles)
	assert.Empty(t, stream.AudioLang)
}

func TestResolvePlutoTVStreamHonoursTheQuality(t *testing.T) {
	f := newPlutoAPIFake(t)
	ep := &models.Episode{URL: plutotv.EpisodeURL(ptSeriesID, ptEpisode1)}
	for quality, rendition := range map[string]string{"best": "hls_2400", "worst": "hls_600", "720p": "hls_2400", "360p": "hls_600"} {
		stream, err := ResolvePlutoTVStream(plutoSeries(), ep, quality)
		require.NoError(t, err, quality)
		dir := strings.TrimSuffix(f.srv.URL+f.clip(rendition), rendition+".m3u8")
		assert.Contains(t, servedPlaylist(t, stream.URL), "\n"+dir+rendition+"-", quality)
	}
}

func TestResolvePlutoTVStreamPlaysAMovie(t *testing.T) {
	f := newPlutoAPIFake(t)
	stream, err := ResolvePlutoTVStream(plutoMovie(), &models.Episode{URL: plutotv.MovieURL(ptMovieID)}, "best")
	require.NoError(t, err)
	assert.Contains(t, servedPlaylist(t, stream.URL), "hls_2400-")
	assert.Equal(t, 1, f.hit("/v2/stitch/hls/episode/"+ptMovieID+"/master.m3u8"))
}

func TestResolvePlutoTVStreamPlaysTheCDNsURLWhenNothingCanBeServedLocally(t *testing.T) {
	f := newPlutoAPIFake(t)
	prev := plutoPlaylistURLFn
	t.Cleanup(func() { plutoPlaylistURLFn = prev })
	plutoPlaylistURLFn = func(string) (string, error) { return "", errors.New("no port to listen on") }

	stream, err := ResolvePlutoTVStream(plutoSeries(), &models.Episode{URL: plutotv.EpisodeURL(ptSeriesID, ptEpisode1)}, "best")
	require.NoError(t, err)
	assert.Equal(t, f.srv.URL+f.clip("hls_2400"), stream.URL, "it still plays wherever the CDN answers")
}

// ---- the loopback playlist server -------------------------------------------

func TestPlutoTVPlaylistServerServesOnlyWhatWasPublished(t *testing.T) {
	s := &plutoPlaylistServer{lists: make(map[string]string)}
	const playlist = "#EXTM3U\n#EXTINF:5,\nhttps://cdn/a.ts\n#EXT-X-ENDLIST\n"
	target, err := s.publish(playlist)
	require.NoError(t, err)
	assert.Equal(t, playlist, servedPlaylist(t, target))

	other, err := s.publish(playlist)
	require.NoError(t, err)
	assert.NotEqual(t, target, other, "each publication has an address of its own, which cannot be guessed")
	assert.Equal(t, s.base, target[:len(s.base)], "one listener serves them all")

	status := func(method, u string) (int, string) {
		req, err := http.NewRequestWithContext(t.Context(), method, u, http.NoBody)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		body, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(body)
	}
	id := strings.TrimSuffix(strings.TrimPrefix(target, s.base+"/plutotv/"), ".m3u8")
	for name, u := range map[string]string{
		"an id that was never published": s.base + "/plutotv/" + strings.Repeat("0", 32) + ".m3u8",
		"the id without the extension":   s.base + "/plutotv/" + id,
		"the id under another path":      s.base + "/other/" + id + ".m3u8",
		"the root":                       s.base + "/",
		"a path above":                   s.base + "/plutotv/../" + id + ".m3u8",
	} {
		code, body := status(http.MethodGet, u)
		assert.Equal(t, http.StatusNotFound, code, name)
		assert.NotContains(t, body, "EXTM3U", name)
	}
	code, body := status(http.MethodPost, target)
	assert.Equal(t, http.StatusMethodNotAllowed, code)
	assert.NotContains(t, body, "EXTM3U")
	code, body = status(http.MethodHead, target)
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, body)
}

func TestPlutoTVPlaylistServerRefusesWhatIsNotAPlaylist(t *testing.T) {
	s := &plutoPlaylistServer{lists: make(map[string]string)}
	for _, text := range []string{"", "<html>", "https://cdn/x.m3u8"} {
		_, err := s.publish(text)
		require.Error(t, err, "%q", text)
	}
	assert.Empty(t, s.base, "nothing listens until there is something to serve")
}

func TestPlutoTVPlaylistServerForgetsTheOldestPastItsBound(t *testing.T) {
	s := &plutoPlaylistServer{lists: make(map[string]string)}
	var urls []string
	for i := range maxPlutoPlaylists + 3 {
		u, err := s.publish(fmt.Sprintf("#EXTM3U\n# %d\n", i))
		require.NoError(t, err)
		urls = append(urls, u)
	}
	assert.Len(t, s.lists, maxPlutoPlaylists)
	assert.Len(t, s.order, maxPlutoPlaylists)
	for i, u := range urls {
		resp, err := http.Get(u) //nolint:noctx // loopback
		require.NoError(t, err)
		_ = resp.Body.Close()
		want := http.StatusOK
		if i < 3 {
			want = http.StatusNotFound
		}
		assert.Equal(t, want, resp.StatusCode, "playlist %d", i)
	}
}

func TestPlutoTVPlaylistServerIsSafeToPublishToFromManyResolutions(t *testing.T) {
	s := &plutoPlaylistServer{lists: make(map[string]string)}
	var wg sync.WaitGroup
	seen := make(chan string, 40)
	for range 40 {
		wg.Go(func() {
			u, err := s.publish("#EXTM3U\n")
			if assert.NoError(t, err) {
				seen <- u
			}
		})
	}
	wg.Wait()
	close(seen)
	unique := map[string]bool{}
	for u := range seen {
		unique[u] = true
		assert.True(t, strings.HasPrefix(u, s.base+"/plutotv/"), "one listener, started once")
	}
	assert.Len(t, unique, 40)
}

func TestResolvePlutoTVStreamStillPlaysATitleItCannotCleanUp(t *testing.T) {
	f := newPlutoAPIFake(t)
	f.stitchedBody = "#EXTM3U\n#EXTINF:5,\n" + f.srv.URL + "/media/clip/a/hls/0-100/hls_600-00000.ts\n#EXTINF:5,\n" + f.srv.URL + "/media/clip/b/hls/0-end/hls_600-00000.ts\n"
	for _, media := range []*models.Anime{plutoSeries(), nil} {
		stream, err := ResolvePlutoTVStream(media, &models.Episode{URL: plutotv.EpisodeURL(ptSeriesID, ptEpisode1)}, "best")
		require.NoError(t, err)
		assert.Contains(t, stream.URL, "/master.m3u8", "the service's own playlist, breaks and all, rather than nothing")
	}
}

func TestResolvePlutoTVStreamNeedsAnEpisode(t *testing.T) {
	f := newPlutoAPIFake(t)
	for name, ep := range map[string]*models.Episode{"nil": nil, "no URL": {Number: "1"}} {
		t.Run(name, func(t *testing.T) {
			_, err := ResolvePlutoTVStream(plutoSeries(), ep, "best")
			require.Error(t, err)
		})
	}
	assert.Zero(t, f.boots)
}

func TestResolvePlutoTVStreamPassesAServiceFailureOn(t *testing.T) {
	f := newPlutoAPIFake(t)
	_, err := ResolvePlutoTVStream(plutoSeries(), &models.Episode{URL: plutotv.SeriesURL(ptSeriesID)}, "best")
	require.ErrorContains(t, err, "names no episode")
	assert.Zero(t, f.boots)

	f.stitchedBody = "#EXTM3U\n#EXT-X-KEY:METHOD=SAMPLE-AES\n#EXTINF:5,\na.ts\n"
	_, err = ResolvePlutoTVStream(plutoSeries(), &models.Episode{URL: plutotv.EpisodeURL(ptSeriesID, ptEpisode1)}, "best")
	assert.ErrorIs(t, err, plutotv.ErrProtected)
}

func TestPlutoTVPlaylistServerReportsThatItCannotListen(t *testing.T) {
	s := &plutoPlaylistServer{lists: make(map[string]string), listen: func() (net.Listener, error) {
		return nil, errors.New("no free port")
	}}
	_, err := s.publish("#EXTM3U\n")
	require.ErrorContains(t, err, "local playlist server")
	require.ErrorContains(t, err, "no free port")
	assert.Empty(t, s.lists, "nothing is kept for an address that does not exist")

	s.listen = nil
	target, err := s.publish("#EXTM3U\n")
	require.NoError(t, err, "and it recovers once a port can be had")
	assert.Equal(t, "#EXTM3U\n", servedPlaylist(t, target))
}
