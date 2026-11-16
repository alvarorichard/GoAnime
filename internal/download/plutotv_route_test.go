package download

import (
	"context"
	"errors"
	"reflect"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alvarorichard/Goanime/v2/internal/api"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/plutotv"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// funcName is the name of a function value, to tell which one a route chose.
func funcName(fn any) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}

// -dm lists a series' seasons through the route its source names: Pluto TV's
// own JSON, not the StartFlix panel the rest of the family goes through.
func TestPlutoTVSeriesListThroughPlutoTV(t *testing.T) {
	title := &models.Anime{Source: "PlutoTV", MediaType: models.MediaTypeTV}
	assert.Equal(t, "plutotv", seriesListingRoute(title))
	assert.Equal(t, funcName(api.GetPlutoTVSeasonNumbers), funcName(seriesSeasonsFor(title)))
	assert.Equal(t, funcName(api.GetPlutoTVSeasonEpisodes), funcName(seriesEpisodesFor(title)))
}

func TestTheOtherRoutesAreWhereTheyWere(t *testing.T) {
	pomfy := &models.Anime{Source: "Pomfy"}
	assert.Equal(t, "pomfy", seriesListingRoute(pomfy))
	assert.Equal(t, funcName(api.GetPomfySeasonNumbers), funcName(seriesSeasonsFor(pomfy)))
	assert.Equal(t, funcName(api.GetPomfySeasonEpisodes), funcName(seriesEpisodesFor(pomfy)))

	for _, source := range []string{"StartFlix", "TopCine", "Pobreflix", "", "plutotv"} {
		title := &models.Anime{Source: source}
		assert.Equal(t, "panel", seriesListingRoute(title), "%q: the route is by the exact source name", source)
		assert.Equal(t, funcName(seriesSeasonsFn), funcName(seriesSeasonsFor(title)))
		assert.Equal(t, funcName(seriesEpisodesFn), funcName(seriesEpisodesFor(title)))
	}
	assert.Equal(t, "panel", seriesListingRoute(nil))
}

// -dm --source plutotv downloads from Pluto TV: the search is narrowed to the
// source that was named instead of mixing in the other movie/TV catalogs.
func TestDashDMSearchesOnlyTheSourceItWasGiven(t *testing.T) {
	for source, want := range map[string]string{
		"plutotv": "plutotv", "PlutoTV": "plutotv", "pluto": "pluto",
		"startflix": "startflix", "topcine": "topcine", "pobreflix": "pobreflix", "pomfy": "pomfy",
	} {
		assert.Equal(t, want, movieSearchScope(source), "%q", source)
		assert.Equal(t, want, movieSearchScope(" "+source+"\t"), "%q, as typed with spaces around it", source)
	}
	for _, source := range []string{"", "movietv", "hianime", "animefire", "ptbr", "english", "plutotv2"} {
		assert.Equal(t, api.MovieTVSources, movieSearchScope(source), "%q is not a movie/TV source: -dm searches them all", source)
	}
}

func TestDashDMOfPlutoTVSearchesPlutoTVAndDownloadsItsEpisode(t *testing.T) {
	series := &models.Anime{Name: "Death Note", Source: "PlutoTV", URL: plutotv.SeriesURL("6206bba315c8610013683974"), MediaType: models.MediaTypeTV}
	prevScoped, prevAll, prevRange := movieScopedSearchFn, movieSearchFn, workflowBatchRangeFn
	t.Cleanup(func() { movieScopedSearchFn, movieSearchFn, workflowBatchRangeFn = prevScoped, prevAll, prevRange })
	var scopes []string
	movieScopedSearchFn = func(name, scope string) (*models.Anime, error) {
		scopes = append(scopes, name+"@"+scope)
		return series, nil
	}
	movieSearchFn = func(string) (*models.Anime, error) {
		t.Error("--source plutotv must not search the whole movie/TV family")
		return nil, errors.New("unexpected")
	}
	var from, to, listed int
	workflowBatchRangeFn = func(eps []models.Episode, title *models.Anime, start, end int) error {
		from, to, listed = start, end, len(eps)
		assert.Same(t, series, title)
		return nil
	}
	useServiceFor(t, newPlutoService(t))

	req := &util.DownloadRequest{AnimeName: "death note", Source: "plutotv", IsTV: true, SeasonNum: 1, EpisodeNum: 2}
	require.NoError(t, HandleMovieDownloadRequest(req))
	assert.Equal(t, []string{"death note@plutotv"}, scopes)
	assert.Equal(t, [2]int{2, 2}, [2]int{from, to})
	assert.Positive(t, listed, "the season was listed from Pluto TV's own answer")
	assert.Equal(t, 1, series.CurrentSeason)
}

func TestDashDMWithoutASourceSearchesTheWholeFamily(t *testing.T) {
	prevScoped, prevAll := movieScopedSearchFn, movieSearchFn
	t.Cleanup(func() { movieScopedSearchFn, movieSearchFn = prevScoped, prevAll })
	movieScopedSearchFn = func(string, string) (*models.Anime, error) {
		t.Error("no source was named")
		return nil, nil
	}
	searched := 0
	movieSearchFn = func(string) (*models.Anime, error) { searched++; return nil, errors.New("stop here") }
	for _, source := range []string{"", "hianime"} {
		err := HandleMovieDownloadRequest(&util.DownloadRequest{AnimeName: "x", Source: source})
		require.ErrorContains(t, err, "stop here")
	}
	assert.Equal(t, 2, searched)
}

// The -dm parser leaves the request's Source empty: the flag lives in
// util.GlobalSource, and that is where the search reads it. A request that does
// name a source wins over the flag.
func TestDashDMReadsTheSourceFlag(t *testing.T) {
	prevScoped, prevAll, prevFlag := movieScopedSearchFn, movieSearchFn, util.GlobalSource
	t.Cleanup(func() { movieScopedSearchFn, movieSearchFn, util.GlobalSource = prevScoped, prevAll, prevFlag })
	var scopes []string
	movieScopedSearchFn = func(_, scope string) (*models.Anime, error) {
		scopes = append(scopes, scope)
		return nil, errors.New("stop here")
	}
	movieSearchFn = func(string) (*models.Anime, error) {
		scopes = append(scopes, "family")
		return nil, errors.New("stop here")
	}

	util.GlobalSource = "plutotv"
	_, _ = searchMovieTitle(&util.DownloadRequest{AnimeName: "x"})
	_, _ = searchMovieTitle(&util.DownloadRequest{AnimeName: "x", Source: "pomfy"})
	util.GlobalSource = ""
	_, _ = searchMovieTitle(&util.DownloadRequest{AnimeName: "x"})
	util.GlobalSource = "hianime"
	_, _ = searchMovieTitle(&util.DownloadRequest{AnimeName: "x"})
	assert.Equal(t, []string{"plutotv", "pomfy", "family", "family"}, scopes)
}

// -dm of a Pluto TV series asked for as a movie says so at once: listing it
// first would open its season picker in a command meant to run unattended.
func TestDashDMOfAPlutoTVSeriesAsAMovieSaysSoWithoutListingIt(t *testing.T) {
	series := &models.Anime{Name: "Death Note", Source: "PlutoTV", URL: plutotv.SeriesURL("6206bba315c8610013683974"), MediaType: models.MediaTypeTV}
	prevScoped, prevFetch := movieScopedSearchFn, workflowFetchEpisodesFn
	t.Cleanup(func() { movieScopedSearchFn, workflowFetchEpisodesFn = prevScoped, prevFetch })
	movieScopedSearchFn = func(string, string) (*models.Anime, error) { return series, nil }
	workflowFetchEpisodesFn = func(context.Context, *models.Anime) ([]models.Episode, error) {
		t.Error("the listing, and with it the season picker, must not be reached")
		return nil, nil
	}

	err := HandleMovieDownloadRequest(&util.DownloadRequest{AnimeName: "death note", Source: "plutotv"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"Death Note" is a series on PlutoTV`)
	assert.Contains(t, err.Error(), `-dm --type tv "Death Note" <season> <episode>`)
}

// -dm --type tv of a Pluto TV movie gets the same hint the other sources give,
// not the client's own "a movie has no seasons".
func TestDashDMOfAPlutoTVMovieAsASeriesSaysHowToDownloadIt(t *testing.T) {
	movie := &models.Anime{Name: "A Espia", Source: "PlutoTV", URL: plutotv.MovieURL("673f4126321c1900133853da"), MediaType: models.MediaTypeMovie}
	prevScoped := movieScopedSearchFn
	t.Cleanup(func() { movieScopedSearchFn = prevScoped })
	movieScopedSearchFn = func(string, string) (*models.Anime, error) { return movie, nil }
	useServiceFor(t, newPlutoService(t))

	for name, req := range map[string]*util.DownloadRequest{
		"one episode": {AnimeName: "a espia", Source: "plutotv", IsTV: true, SeasonNum: 1, EpisodeNum: 1},
		"all seasons": {AnimeName: "a espia", Source: "plutotv", IsTV: true, IsAll: true},
	} {
		err := HandleMovieDownloadRequest(req)
		require.Error(t, err, name)
		assert.Contains(t, err.Error(), `"A Espia" is a movie on PlutoTV: download it with -dm "A Espia"`, name)
	}
}

func TestSeriesErrOfAnotherSourceIsWhatItWas(t *testing.T) {
	startflix := &models.Anime{Name: "X", Source: "StartFlix", URL: "https://startflix.test/x"}
	assert.Contains(t, seriesErr(startflix, api.ErrStartFlixNotSeries).Error(), `"X" is a movie on StartFlix`)
	assert.Contains(t, seriesErr(startflix, errors.New("boom")).Error(), `failed to list "X": boom`)
	plutoSeries := &models.Anime{Name: "Y", Source: "PlutoTV", URL: plutotv.SeriesURL("6206bba315c8610013683974")}
	assert.Contains(t, seriesErr(plutoSeries, errors.New("boom")).Error(), `failed to list "Y": boom`, "a Pluto TV series that fails to list is a failure, not a movie")
	assert.Contains(t, seriesNotMovieErr(&models.Anime{Name: "Z"}).Error(), `"Z" is a series on StartFlix`, "the source's name, or the family's first when it has none")
}

// -dm of a Pluto TV movie goes on to download it: the "is it a series" check
// that spares a series its season picker must not stop a movie.
func TestDashDMOfPlutoTVMovieDownloadsIt(t *testing.T) {
	movie := &models.Anime{Name: "A Espia", Source: "PlutoTV", URL: plutotv.MovieURL("673f4126321c1900133853da"), MediaType: models.MediaTypeMovie}
	prevScoped, prevRange := movieScopedSearchFn, workflowBatchRangeFn
	t.Cleanup(func() { movieScopedSearchFn, workflowBatchRangeFn = prevScoped, prevRange })
	movieScopedSearchFn = func(string, string) (*models.Anime, error) { return movie, nil }
	var got []models.Episode
	var from, to int
	workflowBatchRangeFn = func(eps []models.Episode, title *models.Anime, start, end int) error {
		got, from, to = eps, start, end
		assert.Same(t, movie, title)
		return nil
	}

	require.NoError(t, HandleMovieDownloadRequest(&util.DownloadRequest{AnimeName: "a espia", Source: "plutotv"}))
	require.Len(t, got, 1, "the movie's one entry")
	assert.Equal(t, movie.URL, got[0].URL)
	assert.Equal(t, [2]int{1, 1}, [2]int{from, to})
}
