package movie

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The EnrichByID tests swap package-level seams and share its cache, so none
// of them run in parallel.

type fakeTMDB struct {
	configured bool
	find       map[string]*models.TMDBMedia
	movies     map[int]*models.TMDBDetails
	shows      map[int]*models.TMDBDetails
	calls      int
}

func (f *fakeTMDB) IsConfigured() bool { return f.configured }

func (f *fakeTMDB) FindByIMDBID(imdbID string) (*models.TMDBMedia, error) {
	f.calls++
	if m, ok := f.find[imdbID]; ok {
		return m, nil
	}
	return nil, errors.New("not found")
}

func (f *fakeTMDB) GetMovieDetails(id int) (*models.TMDBDetails, error) {
	f.calls++
	if d, ok := f.movies[id]; ok {
		cp := *d
		return &cp, nil
	}
	return nil, errors.New("TMDB API returned status: 404")
}

func (f *fakeTMDB) GetTVDetails(id int) (*models.TMDBDetails, error) {
	f.calls++
	if d, ok := f.shows[id]; ok {
		cp := *d
		return &cp, nil
	}
	return nil, errors.New("TMDB API returned status: 404")
}

type fakeIMDb struct {
	titles map[string]*IMDbTitle
	calls  int
}

func (f *fakeIMDb) TitleByID(id string) (*IMDbTitle, error) {
	f.calls++
	if t, ok := f.titles[id]; ok {
		return t, nil
	}
	return nil, ErrIMDbTitleNotFound
}

// fakeWikidata answers Lookup from records keyed by either id.
type fakeWikidata struct {
	records []*WikidataRecord
	calls   int
}

func (f *fakeWikidata) Lookup(tmdbID int, imdbID string, _ bool) (*WikidataRecord, error) {
	f.calls++
	for _, r := range f.records {
		if (imdbID != "" && r.IMDBID == imdbID) || (imdbID == "" && tmdbID > 0 && r.TMDBID == tmdbID) {
			cp := *r
			return &cp, nil
		}
	}
	return nil, ErrNoWikidataLink
}

// useOfficialFakes points the lookup at fakes and empties the cache.
func useOfficialFakes(t *testing.T, tmdb *fakeTMDB, imdb *fakeIMDb, wd *fakeWikidata) {
	t.Helper()
	prevT, prevI, prevW := tmdbClientFn, imdbClientFn, wikidataClientFn
	tmdbClientFn = func() tmdbAPI { return tmdb }
	imdbClientFn = func() imdbAPI { return imdb }
	wikidataClientFn = func() wikidataAPI { return wd }
	officialByID.Clear()
	t.Cleanup(func() {
		tmdbClientFn, imdbClientFn, wikidataClientFn = prevT, prevI, prevW
		officialByID.Clear()
	})
}

// The live records these fakes mirror, checked on 2026-10-08: StartFlix's
// "Coração Selvagem" is tt7526136, which IMDb names "Heart of the Beast"
// (2026) and Wikidata links to TMDB 1263337; TMDB TV 194583 is tt18546730,
// "The Walking Dead: Dead City" (2023).
var (
	heartOfTheBeast   = &IMDbTitle{ID: "tt7526136", Title: "Heart of the Beast", Year: 2026, Kind: "movie"}
	heartOfTheBeastWD = &WikidataRecord{IMDBID: "tt7526136", TMDBID: 1263337, Label: "Heart of the Beast", Year: "2026"}
	deadCity          = &IMDbTitle{ID: "tt18546730", Title: "The Walking Dead: Dead City", Year: 2023, Kind: "tvSeries"}
	deadCityWD        = &WikidataRecord{IMDBID: "tt18546730", TMDBID: 194583, Label: "The Walking Dead: Dead City", Year: "2023"}
)

// TestEnrichByID_MovieKeyless is the StartFlix movie case with no key of any
// kind: IMDb names the title by the panel's IMDb id, and Wikidata supplies
// the TMDB id for the folder.
func TestEnrichByID_MovieKeyless(t *testing.T) {
	tmdb := &fakeTMDB{}
	imdb := &fakeIMDb{titles: map[string]*IMDbTitle{"tt7526136": heartOfTheBeast}}
	wd := &fakeWikidata{records: []*WikidataRecord{heartOfTheBeastWD}}
	useOfficialFakes(t, tmdb, imdb, wd)
	media := &models.Media{Name: "Coração Selvagem", IMDBID: "tt7526136", MediaType: models.MediaTypeMovie}

	require.NoError(t, EnrichByID(media, true))
	assert.Equal(t, "Heart of the Beast", media.OfficialTitle())
	assert.Equal(t, "Heart of the Beast", media.TMDBDetails.Title, "movies carry the title as Title")
	assert.Equal(t, "2026", media.Year)
	assert.Equal(t, 1263337, media.TMDBID)
	assert.Equal(t, "tt7526136", media.IMDBID)
	assert.Zero(t, tmdb.calls, "TMDB is not asked without a key")
}

// TestEnrichByID_ShowKeyless is the StartFlix series case: the panel names
// the show only by TMDB id, so Wikidata supplies the IMDb id IMDb needs.
func TestEnrichByID_ShowKeyless(t *testing.T) {
	imdb := &fakeIMDb{titles: map[string]*IMDbTitle{"tt18546730": deadCity}}
	wd := &fakeWikidata{records: []*WikidataRecord{deadCityWD}}
	useOfficialFakes(t, &fakeTMDB{}, imdb, wd)
	media := &models.Media{Name: "The Walking Dead: Dead City", TMDBID: 194583, MediaType: models.MediaTypeTV}

	require.NoError(t, EnrichByID(media, false))
	assert.Equal(t, "The Walking Dead: Dead City", media.OfficialTitle())
	assert.Equal(t, "The Walking Dead: Dead City", media.TMDBDetails.Name, "shows carry the title as Name")
	assert.Equal(t, "2023", media.Year)
	assert.Equal(t, "tt18546730", media.IMDBID)
	assert.Equal(t, 194583, media.TMDBID)
}

// TestEnrichByID_IMDbDownUsesWikidataLabel: IMDb's lookup is unofficial, so
// when it fails Wikidata's English label and year name the title instead.
func TestEnrichByID_IMDbDownUsesWikidataLabel(t *testing.T) {
	imdb := &fakeIMDb{}
	wd := &fakeWikidata{records: []*WikidataRecord{heartOfTheBeastWD}}
	useOfficialFakes(t, &fakeTMDB{}, imdb, wd)
	media := &models.Media{Name: "Coração Selvagem", IMDBID: "tt7526136"}

	require.NoError(t, EnrichByID(media, true))
	assert.Equal(t, 1, imdb.calls)
	assert.Equal(t, "Heart of the Beast", media.OfficialTitle())
	assert.Equal(t, "2026", media.Year)
}

// TestEnrichByID_BothIDsKnownSkipsWikidata: with both ids in hand and IMDb
// answering, Wikidata is not asked at all.
func TestEnrichByID_BothIDsKnownSkipsWikidata(t *testing.T) {
	imdb := &fakeIMDb{titles: map[string]*IMDbTitle{"tt7526136": heartOfTheBeast}}
	wd := &fakeWikidata{}
	useOfficialFakes(t, &fakeTMDB{}, imdb, wd)
	media := &models.Media{TMDBID: 1263337, IMDBID: "tt7526136"}

	require.NoError(t, EnrichByID(media, true))
	assert.Equal(t, "Heart of the Beast", media.OfficialTitle())
	assert.Zero(t, wd.calls)
}

func TestEnrichByID_MovieWithTMDBKey(t *testing.T) {
	tmdb := &fakeTMDB{
		configured: true,
		find:       map[string]*models.TMDBMedia{"tt7526136": {ID: 1263337, MediaType: "movie"}},
		movies:     map[int]*models.TMDBDetails{1263337: {ID: 1263337, IMDBID: "tt7526136", Title: "Heart of the Beast", ReleaseDate: "2026-09-25"}},
	}
	imdb, wd := &fakeIMDb{}, &fakeWikidata{}
	useOfficialFakes(t, tmdb, imdb, wd)
	media := &models.Media{Name: "Coração Selvagem", IMDBID: "tt7526136", MediaType: models.MediaTypeMovie}

	require.NoError(t, EnrichByID(media, true))
	assert.Equal(t, "Heart of the Beast", media.OfficialTitle())
	assert.Equal(t, "2026", media.Year)
	assert.Equal(t, 1263337, media.TMDBID)
	assert.Zero(t, imdb.calls+wd.calls, "TMDB answered on its own")
}

// TestEnrichByID_ShowWithTMDBKey: TMDB's TV details have no imdb_id, so
// Wikidata fills it in for the folder.
func TestEnrichByID_ShowWithTMDBKey(t *testing.T) {
	tmdb := &fakeTMDB{
		configured: true,
		shows:      map[int]*models.TMDBDetails{2691: {ID: 2691, Name: "Two and a Half Men", FirstAirDate: "2003-09-22"}},
	}
	wd := &fakeWikidata{records: []*WikidataRecord{{IMDBID: "tt0369179", TMDBID: 2691}}}
	useOfficialFakes(t, tmdb, &fakeIMDb{}, wd)
	media := &models.Media{Name: "Dois Homens e Meio", TMDBID: 2691, MediaType: models.MediaTypeTV}

	require.NoError(t, EnrichByID(media, false))
	assert.Equal(t, "Two and a Half Men", media.OfficialTitle())
	assert.Equal(t, "2003", media.Year)
	assert.Equal(t, "tt0369179", media.IMDBID)
}

// TestEnrichByID_FindOfWrongKindIsIgnored: an IMDb id TMDB files as a show is
// not taken for a movie lookup; the keyless path answers instead.
func TestEnrichByID_FindOfWrongKindIsIgnored(t *testing.T) {
	tmdb := &fakeTMDB{configured: true, find: map[string]*models.TMDBMedia{"tt7526136": {ID: 55, MediaType: "tv"}}}
	imdb := &fakeIMDb{titles: map[string]*IMDbTitle{"tt7526136": heartOfTheBeast}}
	useOfficialFakes(t, tmdb, imdb, &fakeWikidata{})
	media := &models.Media{IMDBID: "tt7526136"}

	require.NoError(t, EnrichByID(media, true))
	assert.Equal(t, "Heart of the Beast", media.OfficialTitle())
	assert.NotEqual(t, 55, media.TMDBID)
}

func TestEnrichByID_TMDBFailureFallsBackToKeyless(t *testing.T) {
	imdb := &fakeIMDb{titles: map[string]*IMDbTitle{"tt18546730": deadCity}}
	wd := &fakeWikidata{records: []*WikidataRecord{deadCityWD}}
	useOfficialFakes(t, &fakeTMDB{configured: true}, imdb, wd)
	media := &models.Media{TMDBID: 194583}

	require.NoError(t, EnrichByID(media, false))
	assert.Equal(t, "The Walking Dead: Dead City", media.OfficialTitle())
}

// TestEnrichByID_ReplacesNameSearchRecord: the record by id is authoritative,
// so it overwrites what a search by the localized name had filled in.
func TestEnrichByID_ReplacesNameSearchRecord(t *testing.T) {
	imdb := &fakeIMDb{titles: map[string]*IMDbTitle{"tt7526136": heartOfTheBeast}}
	wd := &fakeWikidata{records: []*WikidataRecord{heartOfTheBeastWD}}
	useOfficialFakes(t, &fakeTMDB{}, imdb, wd)
	media := &models.Media{
		Name: "Coração Selvagem", IMDBID: "tt7526136", Year: "1990",
		TMDBDetails: &models.TMDBDetails{Title: "Wild at Heart", ReleaseDate: "1990-08-17"},
	}

	require.NoError(t, EnrichByID(media, true))
	assert.Equal(t, "Heart of the Beast", media.OfficialTitle())
	assert.Equal(t, "2026", media.Year)
}

// TestEnrichByID_CachesByID: a title's episodes each pass through the lookup;
// only the first one reaches the network, and a miss is remembered too.
func TestEnrichByID_CachesByID(t *testing.T) {
	imdb := &fakeIMDb{titles: map[string]*IMDbTitle{"tt7526136": heartOfTheBeast}}
	wd := &fakeWikidata{records: []*WikidataRecord{heartOfTheBeastWD}}
	useOfficialFakes(t, &fakeTMDB{}, imdb, wd)

	for range 3 {
		media := &models.Media{IMDBID: "tt7526136"}
		require.NoError(t, EnrichByID(media, true))
		assert.Equal(t, "Heart of the Beast", media.OfficialTitle())
	}
	assert.Equal(t, 1, imdb.calls)

	for range 3 {
		err := EnrichByID(&models.Media{IMDBID: "tt9999999"}, true)
		assert.ErrorIs(t, err, ErrNoOfficialRecord)
	}
	assert.Equal(t, 2, imdb.calls, "a miss is not retried for every episode")

	// Each media gets its own copy of the record.
	a, b := &models.Media{IMDBID: "tt7526136"}, &models.Media{IMDBID: "tt7526136"}
	require.NoError(t, EnrichByID(a, true))
	require.NoError(t, EnrichByID(b, true))
	a.TMDBDetails.Title = "changed"
	assert.Equal(t, "Heart of the Beast", b.OfficialTitle())
}

func TestEnrichByID_Unresolvable(t *testing.T) {
	useOfficialFakes(t, &fakeTMDB{}, &fakeIMDb{}, &fakeWikidata{})

	media := &models.Media{Name: "Sem Id"}
	assert.ErrorIs(t, EnrichByID(media, true), ErrNoOfficialRecord)
	assert.ErrorIs(t, EnrichByID(nil, true), ErrNoOfficialRecord)

	// A show neither Wikidata nor IMDb knows keeps its StartFlix name.
	media = &models.Media{Name: "Série", TMDBID: 1}
	assert.ErrorIs(t, EnrichByID(media, false), ErrNoOfficialRecord)
	assert.Nil(t, media.TMDBDetails)
	assert.Equal(t, "Série", media.OfficialTitle())
}

func TestLeadingYear(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{
		"2026": "2026", "2023–": "2023", "2011–2019": "2011", "2026-09-25T00:00:00Z": "2026",
		"N/A": "", "": "", "-202": "", "20x6": "",
	} {
		assert.Equal(t, want, leadingYear(in), in)
	}
}

// TestIMDbClient queries a local stand-in for IMDb's suggestion endpoint:
// the id shard path, an identifying User-Agent, English requested, and only
// the entry for the exact id accepted from the suggestion list.
func TestIMDbClient(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, netx.APIUserAgent, r.Header.Get("User-Agent"))
		assert.Contains(t, r.Header.Get("Accept-Language"), "en-US")
		switch r.URL.Path {
		case "/t/tt7526136.json":
			_, _ = w.Write([]byte(`{"d":[{"id":"tt7526136","l":"Heart of the Beast","q":"feature","qid":"movie","y":2026}],"q":"tt7526136","v":1}`))
		case "/t/tt0000001.json":
			// Suggestions that are not the asked-for id.
			_, _ = w.Write([]byte(`{"d":[{"id":"tt0000002","l":"Something Else","y":1999}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	c := &IMDbClient{client: srv.Client(), baseURL: srv.URL}

	got, err := c.TitleByID("tt7526136")
	require.NoError(t, err)
	assert.Equal(t, "Heart of the Beast", got.Title)
	assert.Equal(t, "2026", got.YearString())
	assert.Equal(t, "movie", got.Kind)

	_, err = c.TitleByID("tt0000001")
	assert.ErrorIs(t, err, ErrIMDbTitleNotFound)

	_, err = c.TitleByID("tt7777777")
	assert.Error(t, err, "a 404 is an error")

	_, err = c.TitleByID("../../x")
	assert.Error(t, err, "a malformed id never reaches the URL")
}

// TestWikidataClient queries a local SPARQL stand-in: the anchor property for
// each id and kind, a policy-compliant User-Agent, the earliest year, and
// only well-formed ids accepted from the answer.
func TestWikidataClient(t *testing.T) {
	t.Parallel()
	var (
		mu      sync.Mutex
		queries []string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, netx.APIUserAgent, r.Header.Get("User-Agent"))
		q := r.URL.Query().Get("query")
		mu.Lock()
		queries = append(queries, q)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/sparql-results+json")
		switch {
		case strings.Contains(q, `?item wdt:P4983 "194583"`):
			_, _ = w.Write([]byte(`{"results":{"bindings":[
				{"imdb":{"value":"nm0000001"},"tmdb":{"value":"194583"},"label":{"value":"The Walking Dead: Dead City"},"date":{"value":"2023-06-18T00:00:00Z"}},
				{"imdb":{"value":"tt18546730"},"tmdb":{"value":"194583"},"date":{"value":"2023-06-15T00:00:00Z"}}]}}`))
		case strings.Contains(q, `?item wdt:P345 "tt7526136"`) && strings.Contains(q, "wdt:P4947 ?tmdb"):
			_, _ = w.Write([]byte(`{"results":{"bindings":[
				{"imdb":{"value":"tt7526136"},"tmdb":{"value":"1263337"},"label":{"value":"Heart of the Beast"},"date":{"value":"2026-09-25T00:00:00Z"}},
				{"imdb":{"value":"tt7526136"},"tmdb":{"value":"1263337"},"label":{"value":"Heart of the Beast"},"date":{"value":"2026-09-24T00:00:00Z"}}]}}`))
		default:
			_, _ = w.Write([]byte(`{"results":{"bindings":[]}}`))
		}
	}))
	t.Cleanup(srv.Close)
	c := &WikidataClient{client: srv.Client(), endpoint: srv.URL}

	show, err := c.Lookup(194583, "", false)
	require.NoError(t, err)
	assert.Equal(t, &WikidataRecord{IMDBID: "tt18546730", TMDBID: 194583, Label: "The Walking Dead: Dead City", Year: "2023"}, show,
		"a non-title id is skipped and the earliest date wins")

	film, err := c.Lookup(0, "tt7526136", true)
	require.NoError(t, err)
	assert.Equal(t, &WikidataRecord{IMDBID: "tt7526136", TMDBID: 1263337, Label: "Heart of the Beast", Year: "2026"}, film)

	_, err = c.Lookup(194583, "", true)
	assert.ErrorIs(t, err, ErrNoWikidataLink, "a TV id is not looked up as a movie id")

	_, err = c.Lookup(0, `tt1" } ; DROP`, true)
	assert.Error(t, err, "a malformed IMDb id never reaches the query")
	_, err = c.Lookup(0, "", true)
	assert.Error(t, err)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, queries, 3)
	assert.Contains(t, queries[2], "wdt:P4947", "movies use the TMDB movie property")
}
