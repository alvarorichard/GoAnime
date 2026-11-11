package handlers

// Regression net for what the search home handler hands to the screen TODAY
// (before lists live beside Favorites): the SearchHomeOptions openSearchHome
// builds, the FAVORITES rows and their order, the damaged-file Notice and that
// it is told once, what each callback does, the scope tab picked from the
// --source flag and how the screen's result becomes a decision.
//
// Tests that touch process globals (the favorites store, tracking's CGO flag,
// util.GlobalSource, the seams) are NOT parallel on purpose, and restore what
// they changed. Where today's behavior is arguably wrong the assertion still
// freezes it with a "frozen:" comment; the next steps may change such an
// assertion ONLY on purpose.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/favorites"
	"github.com/alvarorichard/Goanime/v2/internal/suggest"
	"github.com/alvarorichard/Goanime/v2/internal/tracking"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// regHome gives a test its own favorites file and its own (empty) watch
// history, whatever the real HOME holds, and restores everything.
func regHome(t *testing.T) (*favorites.Store, string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("LOCALAPPDATA", home)
	require.NoError(t, tracking.CloseGlobalTracker())
	t.Cleanup(func() { _ = tracking.CloseGlobalTracker() })
	path := filepath.Join(t.TempDir(), "favorites.json")
	store := favorites.Open(path)
	prevStore := favoritesStore
	favoritesStore = func() *favorites.Store { return store }
	prevSource := util.GlobalSource
	util.GlobalSource = ""
	t.Cleanup(func() {
		favoritesStore = prevStore
		util.GlobalSource = prevSource
	})
	return store, path
}

func regCapture(t *testing.T, query string) tui.SearchHomeOptions {
	t.Helper()
	var got tui.SearchHomeOptions
	calls := 0
	res, err := openSearchHome(query, func(o tui.SearchHomeOptions) (tui.SearchHomeResult, error) {
		calls++
		got = o
		return tui.SearchHomeResult{Action: tui.HomeSearch, Query: "q", Scope: "s"}, nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, calls, "the screen is opened once")
	assert.Equal(t, tui.SearchHomeResult{Action: tui.HomeSearch, Query: "q", Scope: "s"}, res, "the screen's answer comes back untouched")
	return got
}

func regAt(minute int) time.Time {
	return time.Date(2024, 3, 1, 10, minute, 0, 0, time.UTC)
}

func TestRegressionHomeOptionsScalarsAndEmptyLists(t *testing.T) {
	regHome(t)

	got := regCapture(t, "frieren")

	assert.Equal(t, "frieren", got.Query, "the query is kept in the box")
	assert.Equal(t, 0, got.ScopeIndex)
	assert.Equal(t, tui.SourcesSummary(homeSourceRows(source.SourceStates())), got.SourcesLine)
	assert.Equal(t, tui.SourcesSummary(homeSources()), got.SourcesLine)
	// frozen: with no history the CONTINUE list is nil and RECENT is empty but not nil.
	assert.Empty(t, got.Continue)
	assert.NotNil(t, got.Recent)
	assert.Empty(t, got.Recent)
	// frozen: an empty (never-written) favorites file gives an empty, non-nil list.
	assert.NotNil(t, got.Favorites)
	assert.Empty(t, got.Favorites)
	assert.Equal(t, "", got.Notice)

	assert.Equal(t, "", regCapture(t, "").Query)
	assert.Equal(t, "  spaced  query ", regCapture(t, "  spaced  query ").Query, "the handler does not touch the query")
}

func TestRegressionHomeOptionsEveryCallbackIsWired(t *testing.T) {
	regHome(t)
	got := regCapture(t, "")

	assert.NotNil(t, got.DeleteEntry)
	assert.NotNil(t, got.DeleteRecent)
	assert.NotNil(t, got.ClearContinue)
	assert.NotNil(t, got.ClearRecent)
	assert.NotNil(t, got.Suggest)
	assert.NotNil(t, got.Sources)
	assert.NotNil(t, got.SetSource)
	assert.NotNil(t, got.DeleteFavorite)
	assert.NotNil(t, got.ClearFavorites)
}

func TestRegressionHomeFavoriteRowsOrderAndFields(t *testing.T) {
	store, _ := regHome(t)
	add := func(e favorites.Entry) {
		t.Helper()
		_, err := store.Toggle(e)
		require.NoError(t, err)
	}
	// Added out of date order on purpose; two share a date.
	add(favorites.Entry{Key: " Middle ", Title: "Middle", Year: "2010", Media: "TV", Source: "Pobreflix", AddedAt: regAt(2)})
	add(favorites.Entry{Key: "oldest", Title: "Oldest", AddedAt: regAt(1)})
	add(favorites.Entry{Key: "tie-1", Title: "Tie One", Year: "1999", Media: "Movie", Source: "TopCine", AddedAt: regAt(5)})
	add(favorites.Entry{Key: "tie-2", Title: "Tie Two", Media: "Anime", Source: "Animefire.io", AddedAt: regAt(5)})
	add(favorites.Entry{Key: "newest", Title: "Newest", Year: "2024", Media: "Anime", Source: "HiAnime", AddedAt: regAt(9)})

	got := regCapture(t, "")

	assert.Equal(t, []tui.HomeFavorite{
		{Key: "newest", Title: "Newest", Year: "2024", Media: "Anime", Source: "HiAnime"},
		{Key: "tie-1", Title: "Tie One", Year: "1999", Media: "Movie", Source: "TopCine"},
		{Key: "tie-2", Title: "Tie Two", Year: "", Media: "Anime", Source: "Animefire.io"},
		{Key: "middle", Title: "Middle", Year: "2010", Media: "TV", Source: "Pobreflix"},
		{Key: "oldest", Title: "Oldest"},
	}, got.Favorites, "newest first, ties in the order added, the key normalized, the source raw")
	assert.Equal(t, got.Favorites, homeFavorites())
}

func TestRegressionHomeFavoriteRowsSanitizeOnlyTheTitle(t *testing.T) {
	store, _ := regHome(t)
	add := func(e favorites.Entry) {
		t.Helper()
		_, err := store.Toggle(e)
		require.NoError(t, err)
	}
	add(favorites.Entry{Key: "a", Title: "  Two\tWords\n here \x1b[31mred\x1b[0m ", Year: " 2001 ", Media: "A\tB", Source: "S\x1b[1m", AddedAt: regAt(3)})
	add(favorites.Entry{Key: "blank", Title: " \t\n ", AddedAt: regAt(2)})
	add(favorites.Entry{Key: "ctrl", Title: "\x1b[2J\x07", AddedAt: regAt(1)})

	got := homeFavorites()

	// frozen: only the title is made one clean line; Year, Media and Source
	// travel as stored (the screen cleans them when it draws).
	assert.Equal(t, []tui.HomeFavorite{
		{Key: "a", Title: "Two Words here red", Year: " 2001 ", Media: "A\tB", Source: "S\x1b[1m"},
	}, got, "a favorite whose title is empty once cleaned is not handed over (it stays in the file)")
	has := store.Has("blank")
	assert.True(t, has)
	assert.True(t, store.Has("ctrl"))
}

func TestRegressionHomeNoCapAtTheHandler(t *testing.T) {
	store, _ := regHome(t)
	for i := range 40 {
		_, err := store.Toggle(favorites.Entry{Key: "k" + strconv.Itoa(i), Title: "T", AddedAt: regAt(0).Add(time.Duration(i) * time.Minute)})
		require.NoError(t, err)
	}
	got := regCapture(t, "")
	assert.Len(t, got.Favorites, 40, "the screen draws eight and counts the rest; the handler hands over all")
	assert.Equal(t, "k39", got.Favorites[0].Key)
}

func TestRegressionHomeNoticeForADamagedFileIsToldOnce(t *testing.T) {
	_, path := regHome(t)
	require.NoError(t, os.WriteFile(path, []byte("not json"), 0o600))

	first := regCapture(t, "")

	entries, err := os.ReadDir(filepath.Dir(path))
	require.NoError(t, err)
	var aside string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "favorites.json.corrupt.") {
			aside = e.Name()
		}
	}
	require.NotEmpty(t, aside, "the damaged file was set aside")
	assert.Regexp(t, regexp.MustCompile(`^favorites\.json\.corrupt\.\d{8}-\d{6}$`), aside)
	assert.Equal(t, "Your favorites file was damaged, so the list starts empty. A copy was kept as "+aside+".", first.Notice)
	assert.Empty(t, first.Favorites)

	second := regCapture(t, "")
	assert.Equal(t, "", second.Notice, "the notice is told once")
}

func TestRegressionHomeNoticeTextIsBuiltFromTheBaseNameOnly(t *testing.T) {
	regHome(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "deep", "favorites.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, []byte("[[["), 0o600))
	store := favorites.Open(path)
	favoritesStore = func() *favorites.Store { return store }

	_, err := store.List() // sets the damaged file aside
	require.NoError(t, err)

	got := favoritesNotice()
	assert.True(t, strings.HasPrefix(got, "Your favorites file was damaged, so the list starts empty. A copy was kept as favorites.json.corrupt."), got)
	assert.True(t, strings.HasSuffix(got, "."), got)
	assert.NotContains(t, got, string(filepath.Separator), "no directory in the sentence")
	assert.Equal(t, "", favoritesNotice(), "told once")
}

func TestRegressionHomeFavoritesUnavailableIsSilent(t *testing.T) {
	regHome(t)

	t.Run("a newer file gives no rows and no notice", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "favorites.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"version":2,"entries":[{"key":"a","title":"A"}]}`), 0o600))
		store := favorites.Open(path)
		favoritesStore = func() *favorites.Store { return store }

		got := regCapture(t, "")

		// frozen: a favorites file from a newer GoAnime shows an empty list
		// and says nothing about it; removing then fails with the newer-file error.
		assert.Nil(t, got.Favorites)
		assert.Equal(t, "", got.Notice)
		err := got.DeleteFavorite("a")
		require.Error(t, err)
		assert.ErrorIs(t, err, favorites.ErrNewerFile)
		assert.ErrorIs(t, got.ClearFavorites(), favorites.ErrNewerFile)
	})

	t.Run("no data directory", func(t *testing.T) {
		store := favorites.Open("")
		favoritesStore = func() *favorites.Store { return store }

		got := regCapture(t, "")

		assert.Nil(t, got.Favorites)
		assert.Equal(t, "", got.Notice)
		assert.EqualError(t, got.DeleteFavorite("a"), "favorites: no data directory")
		assert.EqualError(t, got.ClearFavorites(), "favorites: no data directory")
	})
}

func TestRegressionHomeFavoriteCallbacksActOnTheStore(t *testing.T) {
	store, _ := regHome(t)
	for i, k := range []string{"one", "two", "three"} {
		_, err := store.Toggle(favorites.Entry{Key: k, Title: k, AddedAt: regAt(i)})
		require.NoError(t, err)
	}
	got := regCapture(t, "")

	require.NoError(t, got.DeleteFavorite("  TWO "))
	keys := store.Keys()
	assert.Equal(t, map[string]bool{"one": true, "three": true}, keys, "removes the one key, normalized")
	require.NoError(t, got.DeleteFavorite("ghost"), "a key that is not there is not an error")
	assert.Len(t, store.Keys(), 2)

	require.NoError(t, got.ClearFavorites())
	assert.Empty(t, store.Keys())
	list, err := store.List()
	require.NoError(t, err)
	assert.Empty(t, list)

	assert.NoError(t, forgetFavorite("nothing"))
	assert.NoError(t, clearFavorites())
}

func TestRegressionHomeHistoryCallbacksDelegate(t *testing.T) {
	regHome(t)
	prevForget, prevCont, prevRecent := forgetSearchFn, clearContinueFn, clearRecentFn
	prevRows, prevDelete := continueRows, deleteTracked
	t.Cleanup(func() {
		forgetSearchFn, clearContinueFn, clearRecentFn = prevForget, prevCont, prevRecent
		continueRows, deleteTracked = prevRows, prevDelete
	})
	var forgotten []string
	boom := errors.New("boom")
	forgetSearchFn = func(q string) error { forgotten = append(forgotten, q); return boom }
	clearContinueFn = func() (int, error) { return 3, errors.New("continue failed") }
	clearRecentFn = func() (int, error) { return 4, errors.New("recent failed") }
	got := regCapture(t, "")

	assert.Same(t, boom, got.DeleteRecent("naruto"))
	assert.Equal(t, []string{"naruto"}, forgotten)
	assert.EqualError(t, got.ClearContinue(), "continue failed")
	assert.EqualError(t, got.ClearRecent(), "recent failed")
	clearContinueFn = func() (int, error) { return 0, nil }
	clearRecentFn = func() (int, error) { return 0, nil }
	assert.NoError(t, got.ClearContinue())
	assert.NoError(t, got.ClearRecent())

	// DeleteEntry removes every episode of the show the row stands for.
	continueRows = func() ([]tracking.Anime, error) {
		return []tracking.Anime{
			{AllanimeID: "u1:ep1", AnimeTitle: "Show", Source: "S"},
			{AllanimeID: "u1:ep2", AnimeTitle: "show", Source: "s"},
			{AllanimeID: "other:ep1", AnimeTitle: "Other", Source: "S"},
		}, nil
	}
	var deleted []string
	deleteTracked = func(k string) error { deleted = append(deleted, k); return nil }
	require.NoError(t, got.DeleteEntry("u1:ep2"))
	assert.Equal(t, []string{"u1:ep1", "u1:ep2"}, deleted)
	deleted = nil
	require.NoError(t, got.DeleteEntry("unknown"))
	assert.Equal(t, []string{"unknown"}, deleted, "a key that is in no row is deleted by itself")
}

func TestRegressionHomeSourcesCallbacks(t *testing.T) {
	regHome(t)
	prevPath, prevPersist := sourcePrefsPath, persistSourcePref
	t.Cleanup(func() { sourcePrefsPath, persistSourcePref = prevPath, prevPersist })
	got := regCapture(t, "")

	rows := got.Sources()
	assert.Equal(t, homeSourceRows(source.SourceStates()), rows)

	var gotPath, gotID string
	var gotOn bool
	sourcePrefsPath = func() (string, error) { return "/prefs/path", nil }
	persistSourcePref = func(path, id string, on bool) error {
		gotPath, gotID, gotOn = path, id, on
		return errors.New("disk full")
	}
	assert.EqualError(t, got.SetSource("goyabu", true), "disk full")
	assert.Equal(t, "/prefs/path", gotPath)
	assert.Equal(t, "goyabu", gotID)
	assert.True(t, gotOn)
	assert.EqualError(t, got.SetSource("x", false), "disk full")
	assert.False(t, gotOn)

	sourcePrefsPath = func() (string, error) { return "", errors.New("no home") }
	assert.EqualError(t, got.SetSource("goyabu", true), "no home")
}

func TestRegressionHomeSourceRowsAdaptStates(t *testing.T) {
	t.Parallel()
	rows := homeSourceRows([]source.State{
		{Kind: source.SourceKind("AnimeFire"), Note: "n1", On: true},
		{Kind: source.SourceKind("Goyabu"), Note: "n2", On: false, Locked: true},
	})
	assert.Equal(t, []tui.SourceRow{
		{ID: "AnimeFire", Name: "AnimeFire", Note: "n1", On: true},
		{ID: "Goyabu", Name: "Goyabu", Note: "n2", Locked: true},
	}, rows)
	assert.Empty(t, homeSourceRows(nil))
}

func TestRegressionHomeSuggestAdapter(t *testing.T) {
	t.Parallel()
	var gotQuery string
	var gotKinds []suggest.Kind
	ask := func(_ context.Context, q string, kinds []suggest.Kind) ([]suggest.Suggestion, error) {
		gotQuery, gotKinds = q, kinds
		return []suggest.Suggestion{{Title: "Frieren", Kind: suggest.KindAnime, Year: 2023}, {Title: "Heat"}}, nil
	}
	rows, err := homeSuggestWith(t.Context(), "fri", "ptbr", ask)
	require.NoError(t, err)
	assert.Equal(t, "fri", gotQuery)
	assert.Equal(t, []suggest.Kind{suggest.KindAnime}, gotKinds)
	require.Len(t, rows, 2)
	assert.Equal(t, "Frieren", rows[0].Title)
	assert.Equal(t, suggest.Suggestion{Title: "Frieren", Kind: suggest.KindAnime, Year: 2023}.Detail(), rows[0].Detail)
	assert.Equal(t, "Heat", rows[1].Title)

	_, _ = homeSuggestWith(t.Context(), "x", "movietv", ask)
	assert.Equal(t, []suggest.Kind{suggest.KindMovie, suggest.KindSeries}, gotKinds)
	_, _ = homeSuggestWith(t.Context(), "x", "", ask)
	assert.Nil(t, gotKinds)

	failing := func(context.Context, string, []suggest.Kind) ([]suggest.Suggestion, error) {
		return []suggest.Suggestion{{Title: "ignored"}}, errors.New("offline")
	}
	rows, err = homeSuggestWith(t.Context(), "x", "", failing)
	assert.EqualError(t, err, "offline")
	assert.Nil(t, rows)
}

// ---- scope tab from --source ---------------------------------------------------------------

func TestRegressionScopeIndexForSource(t *testing.T) {
	t.Parallel()
	// The tabs are ALL(0) PT-BR(1) ENGLISH(2) MOVIES/TV(3) FAVORITES(4).
	tests := []struct {
		src  string
		want int
	}{
		{"", 0}, {"   ", 0}, {"all", 0}, {"unknown", 0}, {"favorites", 0}, {"english", 0}, {"en", 0},
		{"hianime", 2}, {"HiAnime", 2}, {"  HIANIME  ", 2},
		{"ptbr", 1}, {"pt-br", 1}, {"PT-BR", 1}, {"animefire", 1}, {"AnimeFire", 1}, {"goyabu", 1}, {"animesdigital", 1}, {"anroll", 1}, {" Goyabu\t", 1},
		{"movietv", 3}, {"MovieTV", 3}, {"startflix", 3}, {"StartFlix", 3}, {"topcine", 3}, {"pobreflix", 3},
		// frozen: sources that exist but are not in the lists fall on ALL.
		{"animedrive", 0}, {"superflix", 0}, {"abyss", 0},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, scopeIndexForSource(tt.src), "%q", tt.src)
		})
	}
}

func TestRegressionHomeScopeIndexFollowsTheSourceFlag(t *testing.T) {
	regHome(t)
	for src, want := range map[string]int{"": 0, "hianime": 2, "goyabu": 1, "startflix": 3} {
		util.GlobalSource = src
		assert.Equal(t, want, regCapture(t, "").ScopeIndex, "--source %q", src)
	}
}

// ---- the screen's answer becomes a decision --------------------------------------------------

func TestRegressionHandleSearchHomeDecisions(t *testing.T) {
	regHome(t)
	prevTTY, prevRun := stdinIsTerminal, runSearchHome
	t.Cleanup(func() { stdinIsTerminal, runSearchHome = prevTTY, prevRun })
	stdinIsTerminal = func() bool { return true }
	entry := tui.HomeContinueEntry{Key: "k", AnimeTitle: "Show", EpisodeTitle: "Ep", EpisodeNum: 4, Position: 70, Duration: 100, Source: "S", MediaType: "anime"}
	var opened []tui.SearchHomeOptions
	answer := func(res tui.SearchHomeResult, err error) {
		runSearchHome = func(o tui.SearchHomeOptions) (tui.SearchHomeResult, error) {
			opened = append(opened, o)
			return res, err
		}
	}

	answer(tui.SearchHomeResult{Action: tui.HomeSearch, Query: "frieren", Scope: "ptbr"}, nil)
	d, err := HandleSearchHome()
	require.NoError(t, err)
	assert.Equal(t, SearchHomeDecision{Query: "frieren", Scope: "ptbr"}, d)
	require.Len(t, opened, 1)
	assert.Equal(t, "", opened[0].Query, "the first visit starts with an empty box")

	answer(tui.SearchHomeResult{Action: tui.HomeSearch, Query: "pokemon aventuras"}, nil)
	d, err = HandleSearchHome()
	require.NoError(t, err)
	assert.Equal(t, SearchHomeDecision{Query: "pokemon aventuras", Scope: ""}, d, "a favorite searches every scope")

	answer(tui.SearchHomeResult{Action: tui.HomeResume, Entry: entry, HasEntry: true, Scope: "hianime"}, nil)
	d, err = HandleSearchHome()
	require.NoError(t, err)
	require.NotNil(t, d.Entry)
	assert.Equal(t, entry, *d.Entry)
	assert.Equal(t, "Show", d.Title)
	assert.Equal(t, "hianime", d.Scope)
	assert.False(t, d.FromStart)
	assert.False(t, d.EpisodesOnly)
	assert.False(t, d.Cancelled)
	assert.Equal(t, "", d.Query)

	answer(tui.SearchHomeResult{Action: tui.HomeFromStart, Entry: entry, HasEntry: true, Scope: "movietv"}, nil)
	d, err = HandleSearchHome()
	require.NoError(t, err)
	require.NotNil(t, d.Entry)
	assert.True(t, d.FromStart)
	assert.False(t, d.EpisodesOnly)
	assert.Equal(t, "Show", d.Title)
	assert.Equal(t, "movietv", d.Scope)

	answer(tui.SearchHomeResult{Action: tui.HomeEpisodes, Entry: entry, HasEntry: true, Scope: ""}, nil)
	d, err = HandleSearchHome()
	require.NoError(t, err)
	require.NotNil(t, d.Entry)
	assert.True(t, d.EpisodesOnly)
	assert.False(t, d.FromStart)
	assert.Equal(t, "Show", d.Title)

	answer(tui.SearchHomeResult{}, tui.ErrSearchHomeCancelled)
	d, err = HandleSearchHome()
	require.NoError(t, err)
	assert.Equal(t, SearchHomeDecision{Cancelled: true}, d)

	boom := errors.New("terminal exploded")
	answer(tui.SearchHomeResult{}, boom)
	d, err = HandleSearchHome()
	assert.Same(t, boom, err)
	assert.Equal(t, SearchHomeDecision{}, d)
}

func TestRegressionOpenSearchHomeWithoutATerminal(t *testing.T) {
	regHome(t)
	prevTTY, prevRun := stdinIsTerminal, runSearchHome
	t.Cleanup(func() { stdinIsTerminal, runSearchHome = prevTTY, prevRun })
	stdinIsTerminal = func() bool { return false }
	called := false
	runSearchHome = func(tui.SearchHomeOptions) (tui.SearchHomeResult, error) {
		called = true
		return tui.SearchHomeResult{}, nil
	}

	res, err := OpenSearchHome("q")

	assert.EqualError(t, err, "no interactive terminal available for search")
	assert.Equal(t, tui.SearchHomeResult{}, res)
	assert.False(t, called)

	stdinIsTerminal = func() bool { return true }
	var gotQuery string
	runSearchHome = func(o tui.SearchHomeOptions) (tui.SearchHomeResult, error) {
		gotQuery = o.Query
		return tui.SearchHomeResult{Query: "x"}, nil
	}
	res, err = OpenSearchHome("kept")
	require.NoError(t, err)
	assert.Equal(t, "kept", gotQuery)
	assert.Equal(t, "x", res.Query)
}

func TestRegressionHomeLimits(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 30, maxHomeContinue)
	assert.Equal(t, 5, maxHomeRecent)
}

// ---- CONTINUE and RECENT, from a real history in a throwaway HOME --------------------------------

func regSeedHistory(t *testing.T) {
	t.Helper()
	if !tracking.IsCgoEnabled {
		t.Skip("tracking unavailable (CGO/SQLite not enabled in this build)")
	}
	tracker := tracking.NewLocalTracker(tracking.DBPath())
	require.NotNil(t, tracker)
	base := time.Now().Add(-48 * time.Hour).Truncate(time.Second)
	// 35 shows with one episode each; show i was watched i minutes after base.
	for i := range 35 {
		require.NoError(t, tracker.UpdateProgress(tracking.Anime{
			AllanimeID: "show" + strconv.Itoa(i) + ":ep1", EpisodeNumber: 1, PlaybackTime: 100, Duration: 1400,
			AnimeTitle: "Show " + strconv.Itoa(i), Title: "Episode One", Source: "HiAnime", Audio: "Sub", MediaType: "anime",
			LastUpdated: base.Add(time.Duration(i) * time.Minute),
		}))
	}
	for i, q := range []string{"a1", "b2", "c3", "d4", "e5", "f6", "g7"} {
		require.NoError(t, tracker.RecordRecentSearch(q, i))
	}
}

func TestRegressionHomeContinueAndRecentRows(t *testing.T) {
	regHome(t)
	regSeedHistory(t)

	got := regCapture(t, "")

	require.Len(t, got.Continue, 30, "CONTINUE is capped at 30 shows, newest first")
	assert.Equal(t, "show34:ep1", got.Continue[0].Key)
	assert.Equal(t, "show5:ep1", got.Continue[29].Key)
	first := got.Continue[0]
	assert.Equal(t, "Show 34", first.AnimeTitle)
	assert.Equal(t, "Episode One", first.EpisodeTitle)
	assert.Equal(t, 1, first.EpisodeNum)
	assert.Equal(t, 100, first.Position)
	assert.Equal(t, 1400, first.Duration)
	assert.Equal(t, "HiAnime", first.Source)
	assert.Equal(t, "Sub", first.Audio)
	assert.Equal(t, "anime", first.MediaType)
	assert.False(t, first.UpdatedAt.IsZero())

	require.Len(t, got.Recent, 5, "RECENT is capped at 5 searches, newest first")
	for _, r := range got.Recent {
		assert.NotEmpty(t, r.Query)
	}
	counts := map[string]int{}
	for _, r := range got.Recent {
		counts[r.Query] = r.Count
	}
	assert.Len(t, counts, 5)
	for q, c := range counts {
		i := strings.Index("a1b2c3d4e5f6g7", q) / 2
		assert.Equal(t, i, c, "the count of %s travels with it", q)
	}
}

func TestRegressionHomeRecentRowsAreCleanLines(t *testing.T) {
	regHome(t)
	if !tracking.IsCgoEnabled {
		t.Skip("tracking unavailable (CGO/SQLite not enabled in this build)")
	}
	tracker := tracking.NewLocalTracker(tracking.DBPath())
	require.NotNil(t, tracker)
	require.NoError(t, tracker.RecordRecentSearch("one\x1b[31m   piece", 3))
	got := regCapture(t, "")
	require.Len(t, got.Recent, 1)
	assert.Equal(t, tui.HomeRecentSearch{Query: "one piece", Count: 3}, got.Recent[0])
}

func TestRegressionGroupContinueEntries(t *testing.T) {
	t.Parallel()
	at := func(m int) time.Time { return time.Date(2024, 3, 1, 10, m, 0, 0, time.UTC) }
	rows := []tracking.Anime{
		{AllanimeID: "a:ep1", AnimeTitle: "Show A", Title: "Ep 1", EpisodeNumber: 1, PlaybackTime: 50, Duration: 1000, Source: "HiAnime", LastUpdated: at(1)},
		{AllanimeID: "a:ep2", AnimeTitle: " show a ", Title: "Ep 2", EpisodeNumber: 2, PlaybackTime: 70, Duration: 1000, Source: "hianime", LastUpdated: at(5)},
		{AllanimeID: "a:ep3", AnimeTitle: "Show A", Title: "Ep 3", EpisodeNumber: 3, PlaybackTime: 10, Duration: 1000, Source: "HiAnime", LastUpdated: at(3)},
		{AllanimeID: "never", AnimeTitle: "Never Played", EpisodeNumber: 1, PlaybackTime: 0, Duration: 1000, LastUpdated: at(9)},
		{AllanimeID: "done", AnimeTitle: "Finished", EpisodeNumber: 4, PlaybackTime: 900, Duration: 1000, LastUpdated: at(2)},
		{AllanimeID: "https://www.startflix.test/movie/heat:ep1", Title: "Heat", EpisodeNumber: 1, PlaybackTime: 30, Duration: 6000, LastUpdated: at(4), MediaType: "movie"},
		{AllanimeID: "old:ep5", Title: " Old\tShow ", EpisodeNumber: 5, PlaybackTime: 30, Duration: 600, LastUpdated: at(6)},
		{AllanimeID: "old:ep6", Title: "Old Show Again", EpisodeNumber: 6, PlaybackTime: 40, Duration: 600, LastUpdated: at(7)},
	}

	got := groupContinueEntries(rows)

	keys := make([]string, len(got))
	for i, e := range got {
		keys[i] = e.Key
	}
	assert.Equal(t, []string{"old:ep6", "a:ep2", "https://www.startflix.test/movie/heat:ep1", "done"}, keys,
		"one row per show, the latest episode wins, never-played rows are dropped, newest first")
	assert.Equal(t, "Old Show Again", got[0].AnimeTitle, "frozen: a row with no show name uses the episode's title")
	assert.Equal(t, "show a", got[1].AnimeTitle, "frozen: the winning row keeps its own spelling of the title")
	assert.Equal(t, "StartFlix", got[2].Source, "a row without a source recovers it from the URL")
	assert.Equal(t, "movie", got[2].MediaType)
	assert.Equal(t, "Heat", got[2].AnimeTitle)
	assert.Equal(t, "Heat", got[2].EpisodeTitle)
	assert.Equal(t, 900, got[3].Position)
}

func TestRegressionContinueTargetAndSourceFromURL(t *testing.T) {
	t.Parallel()
	e := func(num, pos, dur int, media string) tui.HomeContinueEntry {
		return tui.HomeContinueEntry{EpisodeNum: num, Position: pos, Duration: dur, MediaType: media}
	}
	for _, tt := range []struct {
		name      string
		entry     tui.HomeContinueEntry
		fromStart bool
		num, pos  int
		resume    bool
		start     bool
	}{
		{"resume", e(3, 100, 1000, "anime"), false, 3, 100, true, false},
		{"just under the last tenth", e(3, 899, 1000, "anime"), false, 3, 899, true, false},
		{"finished episode offers the next", e(3, 900, 1000, "anime"), false, 4, 0, false, false},
		{"finished movie replays", e(1, 950, 1000, " Movie "), false, 1, 0, false, true},
		{"from start", e(3, 100, 1000, "anime"), true, 3, 0, false, true},
		{"unknown duration resumes", e(3, 100, 0, "anime"), false, 3, 100, true, false},
	} {
		num, pos, resume, start := continueTarget(tt.entry, tt.fromStart)
		assert.Equal(t, []any{tt.num, tt.pos, tt.resume, tt.start}, []any{num, pos, resume, start}, tt.name)
	}
	for url, want := range map[string]string{
		"https://www.animefire.io/x:ep1": "Animefire.io",
		"https://goyabu.to/x":            "Goyabu",
		"https://hianime.to/watch/1":     "HiAnime",
		"https://animesdigital.org/a":    "AnimesDigital",
		"https://anroll.net/a":           "AnimesROLL",
		"https://www.startflix.test/s/x": "StartFlix",
		"https://topcine.net/a":          "TopCine",
		"https://pobreflixtv.club/a":     "Pobreflix",
		"https://animedrive.cc/a":        "AnimeDrive",
		"https://animesdrive.cc/a":       "AnimeDrive",
		"HTTPS://WWW.STARTFLIX.TEST/S/X": "StartFlix",
		"https://example.com/a":          "",
		"":                               "",
		"startflix.test/path":            "StartFlix",
	} {
		assert.Equal(t, want, sourceFromURL(url), url)
	}
	assert.Equal(t, "u:https://x/y", continueGroupKey(tracking.Anime{AllanimeID: "HTTPS://x/y:ep12"}))
	// frozen: a trailing blank keeps the episode suffix in the key.
	assert.Equal(t, "u:https://x/y:ep12", continueGroupKey(tracking.Anime{AllanimeID: " HTTPS://x/y:ep12 "}))
	assert.Equal(t, "u:x:epx", continueGroupKey(tracking.Anime{AllanimeID: "x:epx"}))
	assert.Equal(t, "t:show\x00hianime", continueGroupKey(tracking.Anime{AnimeTitle: " Show ", Source: " HiAnime "}))
}
