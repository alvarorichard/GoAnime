package source

import (
	"github.com/alvarorichard/Goanime/internal/scraper"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// registerProductionLikeSources swaps the registry for fakes carrying the live
// descriptors (mirrored from the providers' Describe methods) so resolution
// scenarios run against production-like matching data. The providers package
// itself cannot be imported here (import cycle), so the real registrations are
// asserted by the resolution tests in internal/api/providers.
func registerProductionLikeSources(t *testing.T) {
	t.Helper()
	restore := SwapRegistryForTesting(
		newFake(AnimeFire, 10, func(d *Descriptor) {
			d.Explicit = []string{"Animefire.io", "AnimeFire"}
			d.Tags = []string{"[animefire]"}
			d.URLMatchers = []string{"animefire"}
		}),
		newFake(Goyabu, 20, func(d *Descriptor) {
			d.Explicit = []string{"Goyabu"}
			d.Tags = []string{"[goyabu]"}
			d.URLMatchers = []string{"goyabu"}
		}),
		newFake(SuperFlix, 30, func(d *Descriptor) {
			d.Explicit = []string{"SuperFlix"}
			d.Tags = []string{"[superflix]"}
			d.URLMatchers = []string{"superflix"}
		}),
		newFake(AniDB, 40, func(d *Descriptor) {
			d.Explicit = []string{"AniDB"}
			d.Tags = []string{"[english]"}
			d.URLMatchers = []string{"stubhost"}
		}),
	)
	t.Cleanup(restore)
}

func TestResolve_ExplicitSource(t *testing.T) {
	// Swaps the global registry — not parallel.
	registerProductionLikeSources(t)
	tests := []struct {
		name     string
		source   string
		wantKind SourceKind
	}{
		{"AniDB", "AniDB", AniDB},
		{"AnimeFire via Animefire.io", "Animefire.io", AnimeFire},
		{"AnimeFire direct", "AnimeFire", AnimeFire},
		{"Goyabu", "Goyabu", Goyabu},
		{"SuperFlix", "SuperFlix", SuperFlix},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, got := Resolve(&models.Anime{Source: tt.source})
			assert.Equal(t, tt.wantKind, got.Kind, "reason: %s", got.Reason)
			require.NotNil(t, src)
			assert.Equal(t, tt.wantKind, src.Describe().Kind)
		})
	}
}

func TestResolve_ExplicitSourceTrumpsURL(t *testing.T) {
	registerProductionLikeSources(t)
	anime := &models.Anime{
		Source: "Goyabu",
		URL:    "https://animefire.plus/something",
	}
	_, got := Resolve(anime)
	assert.Equal(t, Goyabu, got.Kind, "explicit Source should win over URL (reason: %s)", got.Reason)
}

func TestResolve_NameTags(t *testing.T) {
	registerProductionLikeSources(t)
	tests := []struct {
		name     string
		animName string
		wantKind SourceKind
	}{
		{"english tag", "Naruto [English]", AniDB},
		{"animefire tag", "Naruto [AnimeFire]", AnimeFire},
		{"goyabu tag", "Naruto [Goyabu]", Goyabu},
		{"superflix tag", "Naruto [SuperFlix]", SuperFlix},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := Resolve(&models.Anime{Name: tt.animName})
			assert.Equal(t, tt.wantKind, got.Kind, "reason: %s", got.Reason)
		})
	}
}

func TestResolve_URLPatterns(t *testing.T) {
	registerProductionLikeSources(t)
	tests := []struct {
		name     string
		url      string
		wantKind SourceKind
	}{
		{"animefire URL", "https://animefire.plus/naruto", AnimeFire},
		{"goyabu URL", "https://goyabu.to/naruto", Goyabu},
		{"removed allanime host resolves to nothing", "https://allanime.to/anime/abc", Unknown},
		{"superflix URL", "https://superflix.to/naruto", SuperFlix},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := Resolve(&models.Anime{URL: tt.url})
			assert.Equal(t, tt.wantKind, got.Kind, "reason: %s", got.Reason)
		})
	}
}

func TestResolve_PTBRFallback(t *testing.T) {
	registerProductionLikeSources(t)
	src, got := Resolve(&models.Anime{Name: "Naruto [PT-BR]"})
	assert.Equal(t, AnimeFire, got.Kind, "[PT-BR] tag without source should default to AnimeFire")
	require.NotNil(t, src)
	assert.Equal(t, AnimeFire, src.Describe().Kind)
}

func TestResolve_NilAnime(t *testing.T) {
	registerProductionLikeSources(t)
	src, got := Resolve(nil)
	assert.Equal(t, Unknown, got.Kind)
	assert.Nil(t, src)
}

func TestResolve_EmptyAnime(t *testing.T) {
	registerProductionLikeSources(t)
	src, got := Resolve(&models.Anime{})
	assert.Equal(t, Unknown, got.Kind)
	assert.Nil(t, src)
}

func TestResolve_BestEffortKind(t *testing.T) {
	t.Parallel()
	r := ResolvedSource{Kind: Unknown, Reason: "test"}
	assert.Equal(t, Unknown, r.BestEffortKind(),
		"Unknown must stay Unknown: the AllAnime best-effort guess was removed with the source")

	r2 := ResolvedSource{Kind: Goyabu, Reason: "test"}
	assert.Equal(t, Goyabu, r2.BestEffortKind())
}

func TestResolveURL_ProductionDescriptors(t *testing.T) {
	registerProductionLikeSources(t)
	tests := []struct {
		name     string
		url      string
		wantKind SourceKind
	}{
		{"animefire", "https://animefire.plus/ep/naruto-1", AnimeFire},
		{"goyabu", "https://goyabu.to/ep/naruto-1", Goyabu},
		{"removed allanime host is now unknown", "https://allanime.to/anime/hHjXnUTda", Unknown},
		{"empty", "", Unknown},
		{"unknown domain", "https://example.com/video", Unknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, got := ResolveURL(tt.url)
			assert.Equal(t, tt.wantKind, got.Kind, "reason: %s", got.Reason)
			assert.Equal(t, tt.wantKind == Unknown, src == nil)
		})
	}
}

func TestScraperTypeFor(t *testing.T) {
	t.Parallel()
	st, ok := ScraperTypeFor(AniDB)
	require.True(t, ok, "ScraperTypeFor(AniDB) should return true")
	assert.Equal(t, scraper.AniDBType, st)

	st, ok = ScraperTypeFor(AnimeFire)
	require.True(t, ok)
	assert.Equal(t, scraper.AnimefireType, st,
		"AnimefireType became the iota zero value when AllAnimeType was removed")

	_, ok = ScraperTypeFor(Unknown)
	assert.False(t, ok, "ScraperTypeFor(Unknown) should return false")
}
