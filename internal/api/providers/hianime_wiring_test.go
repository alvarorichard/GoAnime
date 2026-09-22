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

// The HiAnime source was added because the AllAnime path went dark upstream.
// These tests walk the registry wiring the guide in docs/ADDING_A_SOURCE.md
// enumerates, so a half-finished touchpoint fails here rather than at runtime.

func TestHiAnime_RegisteredWithExpectedDescriptor(t *testing.T) {
	t.Parallel()
	s, ok := source.Registered(source.HiAnime)
	require.True(t, ok, "HiAnime must be registered by init()")

	d := s.Describe()
	assert.Equal(t, source.HiAnime, d.Kind)
	assert.Equal(t, "https://hianime.at", d.ProbeURL)
	assert.Contains(t, d.URLMatchers, "hianime.at")
	assert.Contains(t, d.Tags, "[hianime]")
	assert.False(t, d.DefaultDisabled, "HiAnime ships live")

	assert.Greater(t, d.Priority, 40,
		"HiAnime is the newest source and must sort after the established ones")
}

// TestHiAnime_DisplayNameRoundTrips guards the trap called out in the guide: the
// name stamped on results must be accepted back by the descriptor, or a saved
// anime never resolves to its own source again.
func TestHiAnime_DisplayNameRoundTrips(t *testing.T) {
	t.Parallel()
	name := sourceDisplayName(source.HiAnime)
	assert.Equal(t, "HiAnime", name)

	s, ok := source.Registered(source.HiAnime)
	require.True(t, ok)
	assert.Contains(t, s.Describe().Explicit, name,
		"sourceDisplayName must appear in Descriptor.Explicit")

	_, resolved := source.Resolve(&models.Anime{Source: name})
	assert.Equal(t, source.HiAnime, resolved.Kind, "reason: %s", resolved.Reason)
}

func TestHiAnime_ResolvesFromEveryIdentityChannel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		anime *models.Anime
	}{
		{"explicit source", &models.Anime{Source: "HiAnime"}},
		{"explicit legacy spelling", &models.Anime{Source: "hianime.at"}},
		{"anime URL", &models.Anime{URL: "https://hianime.at/anime/cowboy-bebop-42"}},
		{"episode URL", &models.Anime{URL: "https://hianime.at/episode/20049"}},
		{"name tag", &models.Anime{Name: "Cowboy Bebop [hianime]"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			src, resolved := source.Resolve(tt.anime)
			require.Equal(t, source.HiAnime, resolved.Kind, "reason: %s", resolved.Reason)
			require.NotNil(t, src)
			assert.Equal(t, source.HiAnime, src.Describe().Kind)
		})
	}
}

// TestHiAnime_DoesNotStealOtherSources is the counterpart: adding a URL matcher
// must not capture anime that belong elsewhere.
func TestHiAnime_DoesNotStealOtherSources(t *testing.T) {
	t.Parallel()
	tests := []struct {
		anime *models.Anime
		want  source.SourceKind
	}{
		{&models.Anime{URL: "https://animefire.io/animes/naruto"}, source.AnimeFire},
		{&models.Anime{URL: "https://goyabu.to/naruto"}, source.Goyabu},
		{&models.Anime{Source: "HiAnime"}, source.HiAnime},
	}
	for _, tt := range tests {
		t.Run(string(tt.want), func(t *testing.T) {
			t.Parallel()
			_, resolved := source.Resolve(tt.anime)
			assert.Equal(t, tt.want, resolved.Kind, "reason: %s", resolved.Reason)
		})
	}
}

func TestHiAnime_MapsToItsScraperType(t *testing.T) {
	t.Parallel()
	st, ok := source.ScraperTypeFor(source.HiAnime)
	require.True(t, ok, "kind.go scraperTypeMap entry missing")
	assert.Equal(t, scraper.HiAnimeType, st)

	adapter, err := scraper.NewAdapter(st)
	require.NoError(t, err, "manager.go NewAdapter case missing")
	assert.Equal(t, scraper.HiAnimeType, adapter.GetType())
}

// TestHiAnime_TagsResultsAsEnglish pins the tagging pass and, with it, that the
// tag it adds is one naming.CleanTitle knows how to strip again — skip that and
// AniList lookups are done on a title that still carries "[HiAnime]".
func TestHiAnime_TagsResultsAsEnglish(t *testing.T) {
	t.Parallel()
	results := []*models.Anime{{Name: "Cowboy Bebop", URL: "https://hianime.at/anime/cowboy-bebop-42"}}
	tagResults(results, source.HiAnime)

	assert.Equal(t, "HiAnime", results[0].Source)
	assert.Contains(t, results[0].Name, "[English]",
		"HiAnime serves Japanese audio with English subs, not PT-BR")

	assert.Equal(t, "Cowboy Bebop", naming.CleanTitle("Cowboy Bebop [HiAnime]"),
		"naming.tagPattern must strip the [HiAnime] tag")
	assert.Equal(t, "Cowboy Bebop", naming.CleanTitle(results[0].Name))
}

func TestHiAnime_IsSearchableAndNotGated(t *testing.T) {
	t.Parallel()
	s, ok := source.Registered(source.HiAnime)
	require.True(t, ok)

	sr, isSearchable := s.(source.Searchable)
	require.True(t, isSearchable, "a source without Search is silently dropped from the fan-out")
	require.NotNil(t, sr)

	assert.False(t, source.IsBrowserGated(s), "HiAnime is plain HTTP")
	assert.False(t, source.IsSeasoned(s), "HiAnime is a flat anime catalog")
}

// TestHiAnime_SearchHonoursContextCancellation keeps the fan-out contract: a
// cancelled search must return promptly without touching the network.
func TestHiAnime_SearchHonoursContextCancellation(t *testing.T) {
	t.Parallel()
	s, ok := source.Registered(source.HiAnime)
	require.True(t, ok)
	sr := s.(source.Searchable)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := sr.Search(ctx, "jojo")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

// TestHiAnime_DeclaresTheContextualCapability pins the Model C wiring added so a
// per-source deadline can actually abort an in-flight request: the adapter must
// declare scraper.ContextualScraper, and the provider must use it.
func TestHiAnime_DeclaresTheContextualCapability(t *testing.T) {
	t.Parallel()
	adapter, err := scraper.NewAdapter(scraper.HiAnimeType)
	require.NoError(t, err)

	_, ok := adapter.(scraper.ContextualScraper)
	assert.True(t, ok,
		"HiAnimeAdapter must implement ContextualScraper, or the provider silently "+
			"falls back to the non-cancelable path")
}

// TestHiAnime_KillSwitchAppliesToTheNewSource checks the new kind participates in
// the generic enablement switch rather than needing its own wiring.
func TestHiAnime_KillSwitchAppliesToTheNewSource(t *testing.T) {
	// Not parallel: mutates the process environment.
	s, ok := source.Registered(source.HiAnime)
	require.True(t, ok)
	d := s.Describe()

	assert.True(t, source.IsEnabled(d), "HiAnime ships enabled")

	t.Setenv("GOANIME_DISABLED_SOURCES", "HiAnime")
	assert.False(t, source.IsEnabled(d), "GOANIME_DISABLED_SOURCES must turn HiAnime off")

	t.Setenv("GOANIME_DISABLED_SOURCES", "  HIANIME , Goyabu ")
	assert.False(t, source.IsEnabled(d),
		"matching is case-insensitive and tolerates spaces inside the list")

	// Documented limitation, not a bug: util.canonSourceToken strips dots, so
	// the host spelling "hianime.at" normalises to "hianimeat" and does NOT match
	// the kind "HiAnime". The switch keys on the SourceKind name.
	t.Setenv("GOANIME_DISABLED_SOURCES", "hianime.at")
	assert.True(t, source.IsEnabled(d),
		"the kill-switch keys on the SourceKind name, not on the host")
}
