package providers

import (
	"context"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/providers/naming"
	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The AniDB source was added because the AllAnime path went dark upstream.
// These tests walk the registry wiring the guide in docs/ADDING_A_SOURCE.md
// enumerates, so a half-finished touchpoint fails here rather than at runtime.

func TestAniDB_RegisteredWithExpectedDescriptor(t *testing.T) {
	t.Parallel()
	s, ok := source.Registered(source.AniDB)
	require.True(t, ok, "AniDB must be registered by init()")

	d := s.Describe()
	assert.Equal(t, source.AniDB, d.Kind)
	assert.Equal(t, "https://anidb.app", d.ProbeURL)
	assert.Contains(t, d.URLMatchers, "anidb.app")
	assert.Contains(t, d.Tags, "[anidb]")
	assert.False(t, d.DefaultDisabled, "AniDB ships live")
	assert.False(t, d.ShortID, "short IDs belong to AllAnime alone")

	assert.Greater(t, d.Priority, 40,
		"AniDB is the newest source and must sort after the established ones")
}

// TestAniDB_DisplayNameRoundTrips guards the trap called out in the guide: the
// name stamped on results must be accepted back by the descriptor, or a saved
// anime never resolves to its own source again.
func TestAniDB_DisplayNameRoundTrips(t *testing.T) {
	t.Parallel()
	name := sourceDisplayName(source.AniDB)
	assert.Equal(t, "AniDB", name)

	s, ok := source.Registered(source.AniDB)
	require.True(t, ok)
	assert.Contains(t, s.Describe().Explicit, name,
		"sourceDisplayName must appear in Descriptor.Explicit")

	_, resolved := source.Resolve(&models.Anime{Source: name})
	assert.Equal(t, source.AniDB, resolved.Kind, "reason: %s", resolved.Reason)
}

func TestAniDB_ResolvesFromEveryIdentityChannel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		anime *models.Anime
	}{
		{"explicit source", &models.Anime{Source: "AniDB"}},
		{"explicit legacy spelling", &models.Anime{Source: "anidb.app"}},
		{"anime URL", &models.Anime{URL: "https://anidb.app/anime/cowboy-bebop-42"}},
		{"episode URL", &models.Anime{URL: "https://anidb.app/episode/20049"}},
		{"name tag", &models.Anime{Name: "Cowboy Bebop [anidb]"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			src, resolved := source.Resolve(tt.anime)
			require.Equal(t, source.AniDB, resolved.Kind, "reason: %s", resolved.Reason)
			require.NotNil(t, src)
			assert.Equal(t, source.AniDB, src.Describe().Kind)
		})
	}
}

// TestAniDB_DoesNotStealOtherSources is the counterpart: adding a URL matcher
// must not capture anime that belong elsewhere.
func TestAniDB_DoesNotStealOtherSources(t *testing.T) {
	t.Parallel()
	tests := []struct {
		anime *models.Anime
		want  source.SourceKind
	}{
		{&models.Anime{URL: "https://animefire.io/animes/naruto"}, source.AnimeFire},
		{&models.Anime{URL: "https://goyabu.to/naruto"}, source.Goyabu},
		{&models.Anime{Source: "AllAnime"}, source.AllAnime},
		{&models.Anime{URL: "hHjXnUTda"}, source.AllAnime},
	}
	for _, tt := range tests {
		t.Run(string(tt.want), func(t *testing.T) {
			t.Parallel()
			_, resolved := source.Resolve(tt.anime)
			assert.Equal(t, tt.want, resolved.Kind, "reason: %s", resolved.Reason)
		})
	}
}

func TestAniDB_MapsToItsScraperType(t *testing.T) {
	t.Parallel()
	st, ok := source.ScraperTypeFor(source.AniDB)
	require.True(t, ok, "kind.go scraperTypeMap entry missing")
	assert.Equal(t, scraper.AniDBType, st)

	adapter, err := scraper.NewAdapter(st)
	require.NoError(t, err, "manager.go NewAdapter case missing")
	assert.Equal(t, scraper.AniDBType, adapter.GetType())
}

// TestAniDB_TagsResultsAsEnglish pins the tagging pass and, with it, that the
// tag it adds is one naming.CleanTitle knows how to strip again — skip that and
// AniList lookups are done on a title that still carries "[AniDB]".
func TestAniDB_TagsResultsAsEnglish(t *testing.T) {
	t.Parallel()
	results := []*models.Anime{{Name: "Cowboy Bebop", URL: "https://anidb.app/anime/cowboy-bebop-42"}}
	tagResults(results, source.AniDB)

	assert.Equal(t, "AniDB", results[0].Source)
	assert.Contains(t, results[0].Name, "[English]",
		"AniDB serves Japanese audio with English subs, not PT-BR")

	assert.Equal(t, "Cowboy Bebop", naming.CleanTitle("Cowboy Bebop [AniDB]"),
		"naming.tagPattern must strip the [AniDB] tag")
	assert.Equal(t, "Cowboy Bebop", naming.CleanTitle(results[0].Name))
}

func TestAniDB_IsSearchableAndNotGated(t *testing.T) {
	t.Parallel()
	s, ok := source.Registered(source.AniDB)
	require.True(t, ok)

	sr, isSearchable := s.(source.Searchable)
	require.True(t, isSearchable, "a source without Search is silently dropped from the fan-out")
	require.NotNil(t, sr)

	assert.False(t, source.IsBrowserGated(s), "AniDB is plain HTTP")
	assert.False(t, source.IsSeasoned(s), "AniDB is a flat anime catalog")
}

// TestAniDB_SearchHonoursContextCancellation keeps the fan-out contract: a
// cancelled search must return promptly without touching the network.
func TestAniDB_SearchHonoursContextCancellation(t *testing.T) {
	t.Parallel()
	s, ok := source.Registered(source.AniDB)
	require.True(t, ok)
	sr := s.(source.Searchable)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := sr.Search(ctx, "jojo")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

// TestAniDB_DeclaresTheContextualCapability pins the Model C wiring added so a
// per-source deadline can actually abort an in-flight request: the adapter must
// declare scraper.ContextualScraper, and the provider must use it.
func TestAniDB_DeclaresTheContextualCapability(t *testing.T) {
	t.Parallel()
	adapter, err := scraper.NewAdapter(scraper.AniDBType)
	require.NoError(t, err)

	_, ok := adapter.(scraper.ContextualScraper)
	assert.True(t, ok,
		"AniDBAdapter must implement ContextualScraper, or the provider silently "+
			"falls back to the non-cancelable path")
}

// TestAniDB_KillSwitchAppliesToTheNewSource checks the new kind participates in
// the generic enablement switch rather than needing its own wiring.
func TestAniDB_KillSwitchAppliesToTheNewSource(t *testing.T) {
	// Not parallel: mutates the process environment.
	s, ok := source.Registered(source.AniDB)
	require.True(t, ok)
	d := s.Describe()

	assert.True(t, source.IsEnabled(d), "AniDB ships enabled")

	t.Setenv("GOANIME_DISABLED_SOURCES", "AniDB")
	assert.False(t, source.IsEnabled(d), "GOANIME_DISABLED_SOURCES must turn AniDB off")

	t.Setenv("GOANIME_DISABLED_SOURCES", "  ANIDB , Goyabu ")
	assert.False(t, source.IsEnabled(d),
		"matching is case-insensitive and tolerates spaces inside the list")

	// Documented limitation, not a bug: util.canonSourceToken strips dots, so
	// the host spelling "anidb.app" normalises to "anidbapp" and does NOT match
	// the kind "AniDB". The switch keys on the SourceKind name.
	t.Setenv("GOANIME_DISABLED_SOURCES", "anidb.app")
	assert.True(t, source.IsEnabled(d),
		"the kill-switch keys on the SourceKind name, not on the host")
}
