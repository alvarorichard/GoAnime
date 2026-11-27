package providers

import (
	"context"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/providers/naming"
	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Miruro was added beside HiAnime when upstream ani-cli moved to it. These
// tests walk the registry wiring docs/ADDING_A_SOURCE.md enumerates, so a
// half-finished touchpoint fails here rather than at runtime.

func TestMiruro_RegisteredWithExpectedDescriptor(t *testing.T) {
	t.Parallel()
	s, ok := source.Registered(source.Miruro)
	require.True(t, ok, "Miruro must be registered by init()")

	d := s.Describe()
	assert.Equal(t, source.Miruro, d.Kind)
	assert.Equal(t, "https://miruro.cx", d.ProbeURL)
	assert.Equal(t, []string{"miruro.cx"}, d.URLMatchers)
	assert.Equal(t, []string{"[miruro]"}, d.Tags, "the [english] tag stays HiAnime's")
	assert.False(t, d.DefaultDisabled, "Miruro ships live")

	hi, ok := source.Registered(source.HiAnime)
	require.True(t, ok)
	assert.Less(t, d.Priority, hi.Describe().Priority,
		"Miruro must be tried before HiAnime, whose [english] tag would otherwise claim Miruro's titles")
}

func TestMiruro_DisplayNameRoundTrips(t *testing.T) {
	t.Parallel()
	name := sourceDisplayName(source.Miruro)
	assert.Equal(t, "Miruro", name)

	s, ok := source.Registered(source.Miruro)
	require.True(t, ok)
	assert.Contains(t, s.Describe().Explicit, name, "sourceDisplayName must appear in Descriptor.Explicit")

	_, resolved := source.Resolve(&models.Anime{Source: name})
	assert.Equal(t, source.Miruro, resolved.Kind, "reason: %s", resolved.Reason)
}

func TestMiruro_ResolvesFromEveryIdentityChannel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		anime *models.Anime
	}{
		{"explicit source", &models.Anime{Source: "Miruro"}},
		{"explicit host", &models.Anime{Source: "miruro.cx"}},
		{"anime URL", &models.Anime{URL: "https://miruro.cx/info/" + miruroTestID}},
		{"episode URL", &models.Anime{URL: "https://miruro.cx/watch/" + miruroTestID + "/cowboy-bebop?ep=1"}},
		{"name tag", &models.Anime{Name: "Cowboy Bebop [miruro]"}},
		// The reason for the priority: a Miruro title that lost its Source
		// still carries the [English] tag HiAnime owns.
		{"English-tagged with a Miruro URL", &models.Anime{Name: "[English] Cowboy Bebop", URL: "https://miruro.cx/info/" + miruroTestID}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			src, resolved := source.Resolve(tt.anime)
			require.Equal(t, source.Miruro, resolved.Kind, "reason: %s", resolved.Reason)
			require.NotNil(t, src)
		})
	}
}

func TestMiruro_DoesNotStealOtherSources(t *testing.T) {
	t.Parallel()
	tests := []struct {
		anime *models.Anime
		want  source.SourceKind
	}{
		{&models.Anime{Name: "[English] Cowboy Bebop", URL: "https://hianime.at/cowboy-bebop-42"}, source.HiAnime},
		{&models.Anime{Name: "[English] Cowboy Bebop"}, source.HiAnime},
		{&models.Anime{Source: "HiAnime", URL: "https://miruro.cx/info/" + miruroTestID}, source.HiAnime},
		{&models.Anime{URL: "https://animefire.io/animes/naruto"}, source.AnimeFire},
	}
	for _, tt := range tests {
		t.Run(string(tt.want)+" "+tt.anime.URL, func(t *testing.T) {
			t.Parallel()
			_, resolved := source.Resolve(tt.anime)
			assert.Equal(t, tt.want, resolved.Kind, "reason: %s", resolved.Reason)
		})
	}
}

func TestMiruro_MapsToItsScraperType(t *testing.T) {
	t.Parallel()
	st, ok := source.ScraperTypeFor(source.Miruro)
	require.True(t, ok, "kind.go scraperTypeMap entry missing")
	assert.Equal(t, scraper.MiruroType, st)

	kind, ok := source.KindFor(scraper.MiruroType)
	require.True(t, ok)
	assert.Equal(t, source.Miruro, kind)

	adapter, err := scraper.NewAdapter(st)
	require.NoError(t, err, "manager.go NewAdapter case missing")
	assert.Equal(t, scraper.MiruroType, adapter.GetType())
	_, contextual := adapter.(scraper.ContextualScraper)
	assert.True(t, contextual, "without ContextualScraper a cancelled search keeps its request alive")
}

func TestMiruro_TagsResultsAsEnglish(t *testing.T) {
	t.Parallel()
	results := []*models.Anime{{Name: "Cowboy Bebop", URL: "https://miruro.cx/info/" + miruroTestID}}
	tagResults(results, source.Miruro)

	assert.Equal(t, "Miruro", results[0].Source)
	assert.Contains(t, results[0].Name, "[English]")
	assert.Equal(t, "Cowboy Bebop", naming.CleanTitle("Cowboy Bebop [Miruro]"), "naming.tagPattern must strip [Miruro]")
	assert.Equal(t, "Cowboy Bebop", naming.CleanTitle(results[0].Name))
}

func TestMiruro_IsSearchableAndNotGated(t *testing.T) {
	t.Parallel()
	s, ok := source.Registered(source.Miruro)
	require.True(t, ok)
	_, searchable := s.(source.Searchable)
	assert.True(t, searchable, "a source without Search is silently dropped from the fan-out")
	assert.False(t, source.IsBrowserGated(s))
	assert.False(t, source.IsSeasoned(s), "Miruro is a flat anime catalog")
}

func TestMiruro_HonoursCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := &miruroProvider{}

	_, err := p.Search(ctx, "jojo")
	assert.ErrorIs(t, err, context.Canceled)
	_, err = p.FetchEpisodes(ctx, &models.Anime{})
	assert.ErrorIs(t, err, context.Canceled)
	_, err = p.FetchStream(ctx, &models.Episode{}, &models.Anime{}, "best")
	assert.ErrorIs(t, err, context.Canceled)
}

func TestMiruro_KillSwitchApplies(t *testing.T) {
	// Not parallel: mutates the process environment.
	s, ok := source.Registered(source.Miruro)
	require.True(t, ok)
	d := s.Describe()
	assert.True(t, source.IsEnabled(d))

	t.Setenv("GOANIME_DISABLED_SOURCES", " miruro ")
	assert.False(t, source.IsEnabled(d), "GOANIME_DISABLED_SOURCES must turn Miruro off")
}

// miruroPlainScraper speaks only UnifiedScraper, no context methods: the
// provider's fallback path for an adapter that loses the capability.
type miruroPlainScraper struct{ gotQuality any }

func (*miruroPlainScraper) SearchAnime(string, ...any) ([]*models.Anime, error) {
	return []*models.Anime{{Name: "Cowboy Bebop", URL: "https://miruro.cx/info/" + miruroTestID}}, nil
}

func (*miruroPlainScraper) GetAnimeEpisodes(string) ([]models.Episode, error) {
	return []models.Episode{{Number: "1", URL: "https://miruro.cx/watch/" + miruroTestID + "/x?ep=1"}}, nil
}

func (s *miruroPlainScraper) GetStreamURL(_ string, options ...any) (url string, metadata map[string]string, err error) {
	if len(options) > 0 {
		s.gotQuality = options[0]
	}
	return "https://cdn.example/master.m3u8", map[string]string{"referer": "https://r.example/", "origin": "https://o.example"}, nil
}

func (*miruroPlainScraper) GetType() scraper.ScraperType { return scraper.MiruroType }

func TestMiruro_PlainAdapterPathStillWorks(t *testing.T) {
	t.Parallel()
	fake := &miruroPlainScraper{}
	p := &miruroProvider{}
	p.adapter.UnifiedScraper = fake
	p.once.Do(func() {})
	ctx := context.Background()

	results, err := p.Search(ctx, "bebop")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Miruro", results[0].Source, "results are tagged on the plain path too")

	eps, err := p.FetchEpisodes(ctx, results[0])
	require.NoError(t, err)
	require.Len(t, eps, 1)

	stream, err := p.FetchStream(ctx, &eps[0], results[0], "720p")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/master.m3u8", stream.URL)
	assert.Equal(t, "https://o.example", stream.Origin)
	assert.Equal(t, "720p", fake.gotQuality, "the quality reaches the adapter")
}
