package source

import (
	"context"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSource is a minimal Source for registry tests.
type fakeSource struct {
	desc Descriptor
}

func (f *fakeSource) Describe() Descriptor { return f.desc }
func (f *fakeSource) FetchEpisodes(_ context.Context, _ *models.Anime) ([]models.Episode, error) {
	return nil, nil
}
func (f *fakeSource) FetchStreamURL(_ context.Context, _ *models.Episode, _ *models.Anime, _ string) (string, error) {
	return "", nil
}

func newFake(kind SourceKind, priority int, mutate ...func(*Descriptor)) *fakeSource {
	d := Descriptor{Kind: kind, Priority: priority}
	for _, m := range mutate {
		m(&d)
	}
	return &fakeSource{desc: d}
}

func TestRegister(t *testing.T) {
	t.Run("registers and duplicate kind replaces", func(t *testing.T) {
		restore := SwapRegistryForTesting()
		t.Cleanup(restore)

		first := newFake("dup-kind", 1)
		second := newFake("dup-kind", 2)
		Register(first)
		Register(second)

		got, ok := Registered("dup-kind")
		require.True(t, ok)
		assert.Same(t, Source(second), got, "last registration for a Kind must win")
	})

	t.Run("nil source panics", func(t *testing.T) {
		assert.Panics(t, func() { Register(nil) })
	})

	t.Run("empty kind panics", func(t *testing.T) {
		assert.Panics(t, func() { Register(newFake("", 0)) })
	})
}

func TestRegistered(t *testing.T) {
	restore := SwapRegistryForTesting(newFake("known-kind", 1))
	t.Cleanup(restore)

	s, ok := Registered("known-kind")
	require.True(t, ok)
	assert.Equal(t, SourceKind("known-kind"), s.Describe().Kind)

	_, ok = Registered("absent-kind")
	assert.False(t, ok)
}

func TestRegisteredByPriority(t *testing.T) {
	restore := SwapRegistryForTesting(
		newFake("zz-low-prio", 5),
		newFake("aa-high-prio", 50),
		newFake("bb-tie", 10),
		newFake("ab-tie", 10),
	)
	t.Cleanup(restore)

	srcs := registeredByPriority()
	require.Len(t, srcs, 4)

	var kinds []SourceKind
	for _, s := range srcs {
		kinds = append(kinds, s.Describe().Kind)
	}
	// Priority ascending; Kind breaks the tie deterministically.
	assert.Equal(t, []SourceKind{"zz-low-prio", "ab-tie", "bb-tie", "aa-high-prio"}, kinds)
}

func TestDescriptorDefinition(t *testing.T) {
	t.Parallel()
	d := Descriptor{
		Kind:        SuperFlix,
		Priority:    30,
		Explicit:    []string{"SuperFlix"},
		Tags:        []string{"[superflix]"},
		URLMatchers: []string{"superflix"},
		MediaTypes:  []models.MediaType{models.MediaTypeMovie},
		ShortID:     true,
	}
	def := d.definition()
	assert.Equal(t, d.Kind, def.Kind)
	assert.Equal(t, d.Explicit, def.Explicit)
	assert.Equal(t, d.Tags, def.Tags)
	assert.Equal(t, d.URLMatchers, def.URLMatchers)
	assert.Equal(t, d.MediaTypes, def.MediaTypes)
	assert.Equal(t, d.ShortID, def.ShortID)
}

func TestResolveSource(t *testing.T) {
	animeFire := newFake(AnimeFire, 10, func(d *Descriptor) {
		d.Explicit = []string{"Animefire.io", "AnimeFire"}
		d.Tags = []string{"[animefire]"}
		d.URLMatchers = []string{"animefire"}
	})
	goyabu := newFake(Goyabu, 20, func(d *Descriptor) {
		d.Explicit = []string{"Goyabu"}
		d.Tags = []string{"[goyabu]"}
		d.URLMatchers = []string{"goyabu"}
	})
	allAnime := newFake(AllAnime, 40, func(d *Descriptor) {
		d.Explicit = []string{"AllAnime"}
		d.Tags = []string{"[english]"}
		d.URLMatchers = []string{"allanime"}
		d.ShortID = true
	})
	restore := SwapRegistryForTesting(animeFire, goyabu, allAnime)
	t.Cleanup(restore)

	tests := []struct {
		name     string
		anime    *models.Anime
		wantKind SourceKind
		wantSrc  Source // nil means expect nil Source
	}{
		{"nil anime", nil, Unknown, nil},
		{"empty anime", &models.Anime{}, Unknown, nil},
		{"explicit Source field", &models.Anime{Source: "Goyabu"}, Goyabu, goyabu},
		{"explicit wins over URL", &models.Anime{Source: "Goyabu", URL: "https://animefire.plus/x"}, Goyabu, goyabu},
		{"name tag", &models.Anime{Name: "Naruto [English]"}, AllAnime, allAnime},
		{"URL pattern", &models.Anime{URL: "https://animefire.plus/naruto"}, AnimeFire, animeFire},
		{"short ID", &models.Anime{URL: "hHjXnUTda"}, AllAnime, allAnime},
		{"PT-BR fallback to AnimeFire", &models.Anime{Name: "Naruto [PT-BR]"}, AnimeFire, animeFire},
		{"no match is Unknown", &models.Anime{Name: "X", URL: "https://example.com/v"}, Unknown, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, resolved := ResolveSource(tt.anime)
			assert.Equal(t, tt.wantKind, resolved.Kind, "reason: %s", resolved.Reason)
			if tt.wantSrc == nil {
				assert.Nil(t, src)
			} else {
				assert.Same(t, tt.wantSrc, src)
			}
		})
	}

	t.Run("lowest Priority wins on multiple matches", func(t *testing.T) {
		// Both match URL "shared"; priority decides.
		low := newFake("low-kind", 1, func(d *Descriptor) { d.URLMatchers = []string{"shared"} })
		high := newFake("high-kind", 99, func(d *Descriptor) { d.URLMatchers = []string{"shared"} })
		restore := SwapRegistryForTesting(high, low)
		t.Cleanup(restore)

		src, resolved := ResolveSource(&models.Anime{URL: "https://shared.example/x"})
		assert.Equal(t, SourceKind("low-kind"), resolved.Kind)
		assert.Same(t, Source(low), src)
	})

	t.Run("PT-BR without AnimeFire registered is Unknown", func(t *testing.T) {
		restore := SwapRegistryForTesting()
		t.Cleanup(restore)

		src, resolved := ResolveSource(&models.Anime{Name: "Naruto [PT-BR]"})
		assert.Nil(t, src)
		assert.Equal(t, Unknown, resolved.Kind)
	})
}

func TestResolveSourceURL(t *testing.T) {
	animeFire := newFake(AnimeFire, 10, func(d *Descriptor) { d.URLMatchers = []string{"animefire"} })
	allAnime := newFake(AllAnime, 40, func(d *Descriptor) {
		d.URLMatchers = []string{"allanime"}
		d.ShortID = true
	})
	restore := SwapRegistryForTesting(animeFire, allAnime)
	t.Cleanup(restore)

	tests := []struct {
		name     string
		url      string
		wantKind SourceKind
		wantNil  bool
	}{
		{"empty URL", "", Unknown, true},
		{"animefire URL", "https://animefire.plus/ep/naruto-1", AnimeFire, false},
		{"allanime URL", "https://allanime.to/anime/hHjXnUTda", AllAnime, false},
		{"short ID", "hHjXnUTda", AllAnime, false},
		{"unknown domain", "https://example.com/video", Unknown, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, resolved := ResolveSourceURL(tt.url)
			assert.Equal(t, tt.wantKind, resolved.Kind, "reason: %s", resolved.Reason)
			assert.Equal(t, tt.wantNil, src == nil)
		})
	}
}

func TestSwapRegistryForTesting(t *testing.T) {
	before := registeredByPriority()

	restore := SwapRegistryForTesting(newFake("swap-only-kind", 1))
	srcs := registeredByPriority()
	require.Len(t, srcs, 1)
	assert.Equal(t, SourceKind("swap-only-kind"), srcs[0].Describe().Kind)

	restore()
	after := registeredByPriority()
	assert.Equal(t, len(before), len(after), "restore must bring back the previous registry")
	_, ok := Registered("swap-only-kind")
	assert.False(t, ok)
}
