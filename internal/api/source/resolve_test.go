package source

import (
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
		newFake(AllAnime, 40, func(d *Descriptor) {
			d.Explicit = []string{"AllAnime"}
			d.Tags = []string{"[english]"}
			d.URLMatchers = []string{"allanime"}
			d.ShortID = true
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
		{"AllAnime", "AllAnime", AllAnime},
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
		{"english tag", "Naruto [English]", AllAnime},
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
		{"allanime URL", "https://allanime.to/anime/abc", AllAnime},
		{"superflix URL", "https://superflix.to/naruto", SuperFlix},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := Resolve(&models.Anime{URL: tt.url})
			assert.Equal(t, tt.wantKind, got.Kind, "reason: %s", got.Reason)
		})
	}
}

func TestResolve_ShortID(t *testing.T) {
	registerProductionLikeSources(t)
	tests := []struct {
		name     string
		url      string
		wantKind SourceKind
	}{
		{"alphanumeric short ID", "hHjXnUTda", AllAnime},
		{"mixed short ID", "abc123XYZ", AllAnime},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := Resolve(&models.Anime{URL: tt.url})
			assert.Equal(t, tt.wantKind, got.Kind)
		})
	}
}

func TestResolve_NumericOnlyIsNotShortID(t *testing.T) {
	registerProductionLikeSources(t)
	_, got := Resolve(&models.Anime{URL: "8143"})
	if got.Kind == AllAnime && got.Reason == "short ID" {
		t.Error("purely numeric '8143' should not match as AllAnime short ID")
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
	assert.Equal(t, AllAnime, r.BestEffortKind(), "BestEffortKind for Unknown should be AllAnime")

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
		{"allanime", "https://allanime.to/anime/hHjXnUTda", AllAnime},
		{"short ID", "hHjXnUTda", AllAnime},
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

func TestIsAllAnimeShortID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  bool
	}{
		{"hHjXnUTda", true},
		{"abc123XYZ", true},
		{"a", true},
		{"8143", false},
		{"", false},
		{"https://example.com", false},
		{"a/b", false},
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false},
		{"abc def", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, IsAllAnimeShortID(tt.input))
		})
	}
}

func TestExtractAllAnimeID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"hHjXnUTda", "hHjXnUTda"},
		{"https://allanime.to/anime/hHjXnUTda", "hHjXnUTda"},
		{"https://example.com/8143", "https://example.com/8143"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, ExtractAllAnimeID(tt.input))
		})
	}
}

func TestScraperTypeFor(t *testing.T) {
	t.Parallel()
	st, ok := ScraperTypeFor(AllAnime)
	require.True(t, ok, "ScraperTypeFor(AllAnime) should return true")
	assert.Equal(t, 0, int(st))

	_, ok = ScraperTypeFor(Unknown)
	assert.False(t, ok, "ScraperTypeFor(Unknown) should return false")
}
