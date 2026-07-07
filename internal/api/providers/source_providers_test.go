package providers

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEpisodeNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   *models.Episode
		want string
	}{
		{"nil episode", nil, ""},
		{"empty episode", &models.Episode{}, ""},
		{"Number string set", &models.Episode{Number: "12"}, "12"},
		{"Num int set", &models.Episode{Num: 7}, "7"},
		{"Number wins over Num", &models.Episode{Number: "abc", Num: 5}, "abc"},
		{"Num zero falls through", &models.Episode{Num: 0}, ""},
		{"Num negative ignored", &models.Episode{Num: -1}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := EpisodeNumber(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAllAnimeProvider_KindAndHasSeasons(t *testing.T) {
	t.Parallel()
	p, err := ForKind(source.AllAnime)
	require.NoError(t, err)
	assert.Equal(t, source.AllAnime, p.Kind())
	assert.False(t, p.HasSeasons())
}

func TestAnimeFireProvider_KindAndHasSeasons(t *testing.T) {
	t.Parallel()
	p, err := ForKind(source.AnimeFire)
	require.NoError(t, err)
	assert.Equal(t, source.AnimeFire, p.Kind())
	assert.False(t, p.HasSeasons())
}

func TestGoyabuProvider_KindAndHasSeasons(t *testing.T) {
	t.Parallel()
	p, err := ForKind(source.Goyabu)
	require.NoError(t, err)
	assert.Equal(t, source.Goyabu, p.Kind())
	assert.False(t, p.HasSeasons())
}

func TestSuperFlixProvider_KindAndHasSeasons(t *testing.T) {
	t.Parallel()
	p, err := ForKind(source.SuperFlix)
	require.NoError(t, err)
	assert.Equal(t, source.SuperFlix, p.Kind())
	assert.True(t, p.HasSeasons())
}

func TestAllAnimeProvider_Describe(t *testing.T) {
	t.Parallel()
	d := (&allAnimeProvider{}).Describe()
	assert.Equal(t, source.AllAnime, d.Kind)
	assert.Equal(t, 40, d.Priority)
	assert.Equal(t, []string{"AllAnime"}, d.Explicit)
	assert.Equal(t, []string{"[english]"}, d.Tags)
	assert.Equal(t, []string{"allanime"}, d.URLMatchers)
	assert.True(t, d.ShortID)
}

func TestAnimeFireProvider_Describe(t *testing.T) {
	t.Parallel()
	d := (&animeFireProvider{}).Describe()
	assert.Equal(t, source.AnimeFire, d.Kind)
	assert.Equal(t, 10, d.Priority)
	assert.Equal(t, []string{"Animefire.io", "AnimeFire"}, d.Explicit)
	assert.Equal(t, []string{"[animefire]"}, d.Tags)
	assert.Equal(t, []string{"animefire"}, d.URLMatchers)
	assert.False(t, d.ShortID)
}

func TestGoyabuProvider_Describe(t *testing.T) {
	t.Parallel()
	d := (&goyabuProvider{}).Describe()
	assert.Equal(t, source.Goyabu, d.Kind)
	assert.Equal(t, 20, d.Priority)
	assert.Equal(t, []string{"Goyabu"}, d.Explicit)
	assert.Equal(t, []string{"[goyabu]"}, d.Tags)
	assert.Equal(t, []string{"goyabu"}, d.URLMatchers)
	assert.False(t, d.ShortID)
}

func TestSuperFlixProvider_Describe(t *testing.T) {
	t.Parallel()
	d := (&superFlixProvider{}).Describe()
	assert.Equal(t, source.SuperFlix, d.Kind)
	assert.Equal(t, 30, d.Priority)
	assert.Equal(t, []string{"SuperFlix"}, d.Explicit)
	assert.Equal(t, []string{"[superflix]"}, d.Tags)
	assert.Equal(t, []string{"superflix"}, d.URLMatchers)
	assert.False(t, d.ShortID)
}

func TestAllAnimeProvider_Manager(t *testing.T) {
	t.Parallel()
	sm := scraper.NewScraperManagerForTest()
	assert.Same(t, sm, (&allAnimeProvider{sm: sm}).manager(), "injected manager must win")
	assert.NotNil(t, (&allAnimeProvider{}).manager(), "nil sm must fall back to the global singleton")
}

func TestAnimeFireProvider_Manager(t *testing.T) {
	t.Parallel()
	sm := scraper.NewScraperManagerForTest()
	assert.Same(t, sm, (&animeFireProvider{sm: sm}).manager(), "injected manager must win")
	assert.NotNil(t, (&animeFireProvider{}).manager(), "nil sm must fall back to the global singleton")
}

func TestGoyabuProvider_Manager(t *testing.T) {
	t.Parallel()
	sm := scraper.NewScraperManagerForTest()
	assert.Same(t, sm, (&goyabuProvider{sm: sm}).manager(), "injected manager must win")
	assert.NotNil(t, (&goyabuProvider{}).manager(), "nil sm must fall back to the global singleton")
}

func TestSuperFlixProvider_Manager(t *testing.T) {
	t.Parallel()
	sm := scraper.NewScraperManagerForTest()
	assert.Same(t, sm, (&superFlixProvider{sm: sm}).manager(), "injected manager must win")
	assert.NotNil(t, (&superFlixProvider{}).manager(), "nil sm must fall back to the global singleton")
}

// TestSourceRegistry_LiveSourcesRegistered verifies init() populated the
// Model B registry with every live source.
func TestSourceRegistry_LiveSourcesRegistered(t *testing.T) {
	t.Parallel()
	for _, kind := range []source.SourceKind{source.AllAnime, source.AnimeFire, source.Goyabu, source.SuperFlix} {
		s, ok := source.Registered(kind)
		require.True(t, ok, "source %s must be registered", kind)
		assert.Equal(t, kind, s.Describe().Kind)
	}
}

// TestResolve_LiveRegistry resolves against the REAL registry populated by
// this package's init() — it pins the production descriptors' matching
// behavior end to end (the source-package tests use mirrored fakes).
func TestResolve_LiveRegistry(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		anime    *models.Anime
		wantKind source.SourceKind
	}{
		{"nil anime", nil, source.Unknown},
		{"empty anime", &models.Anime{}, source.Unknown},
		{"explicit AllAnime", &models.Anime{Source: "AllAnime"}, source.AllAnime},
		{"explicit AnimeFire legacy", &models.Anime{Source: "Animefire.io"}, source.AnimeFire},
		{"explicit Goyabu", &models.Anime{Source: "Goyabu"}, source.Goyabu},
		{"explicit SuperFlix", &models.Anime{Source: "SuperFlix"}, source.SuperFlix},
		{"explicit wins over URL", &models.Anime{Source: "Goyabu", URL: "https://animefire.plus/x"}, source.Goyabu},
		{"english tag", &models.Anime{Name: "Naruto [English]"}, source.AllAnime},
		{"animefire tag", &models.Anime{Name: "Naruto [AnimeFire]"}, source.AnimeFire},
		{"goyabu URL", &models.Anime{URL: "https://goyabu.to/naruto"}, source.Goyabu},
		{"superflix URL", &models.Anime{URL: "https://superflix.to/naruto"}, source.SuperFlix},
		{"short ID", &models.Anime{URL: "hHjXnUTda"}, source.AllAnime},
		{"PT-BR fallback", &models.Anime{Name: "Naruto [PT-BR]"}, source.AnimeFire},
		{"unknown", &models.Anime{Name: "X", URL: "https://example.com/v"}, source.Unknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			src, resolved := source.Resolve(tt.anime)
			assert.Equal(t, tt.wantKind, resolved.Kind, "reason: %s", resolved.Reason)
			if tt.wantKind != source.Unknown {
				require.NotNil(t, src)
				assert.Equal(t, tt.wantKind, src.Describe().Kind)
			}
		})
	}
}

// TestResolveURL_LiveRegistry mirrors TestResolve_LiveRegistry for URL-only
// resolution against the real registered descriptors.
func TestResolveURL_LiveRegistry(t *testing.T) {
	t.Parallel()
	tests := []struct {
		url      string
		wantKind source.SourceKind
	}{
		{"", source.Unknown},
		{"https://animefire.plus/ep/naruto-1", source.AnimeFire},
		{"https://goyabu.to/ep/naruto-1", source.Goyabu},
		{"https://allanime.to/anime/hHjXnUTda", source.AllAnime},
		{"https://superflix.to/naruto", source.SuperFlix},
		{"hHjXnUTda", source.AllAnime},
		{"https://example.com/video", source.Unknown},
	}
	for _, tt := range tests {
		t.Run("url="+tt.url, func(t *testing.T) {
			t.Parallel()
			_, resolved := source.ResolveURL(tt.url)
			assert.Equal(t, tt.wantKind, resolved.Kind, "reason: %s", resolved.Reason)
		})
	}
}
