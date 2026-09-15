package api

import (
	"context"
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchAnimeEnhancedCore_ResultScreenCascade(t *testing.T) {
	t.Parallel()

	t.Run("search sort select enrich", func(t *testing.T) {
		t.Parallel()
		english := &models.Anime{Name: "Frieren [English]", Source: "AniDB", URL: "frieren"}
		portuguese := &models.Anime{Name: "Frieren [PT-BR]", URL: "https://animefire.example/frieren", Year: "2023"}
		providerResults := []*models.Anime{english, nil, portuguese}
		var selectedInput []*models.Anime
		var gotKinds []source.SourceKind
		search := func(_ context.Context, query string, kinds []source.SourceKind) ([]*models.Anime, error) {
			assert.Equal(t, "frieren", query)
			gotKinds = kinds
			return providerResults, nil
		}
		selectAnime := func(animes []*models.Anime) (*models.Anime, error) {
			selectedInput = animes
			return animes[0], nil
		}
		enrich := func(anime *models.Anime) error {
			anime.ImageURL = "mock://cover"
			return nil
		}

		selected, err := searchAnimeEnhanced("frieren", " animefire ", search, selectAnime, enrich)

		require.NoError(t, err)
		require.Len(t, selectedInput, 2)
		assert.Same(t, portuguese, selectedInput[0])
		assert.Equal(t, "Animefire.io", portuguese.Source)
		assert.Equal(t, []source.SourceKind{source.AnimeFire}, gotKinds)
		assert.Same(t, portuguese, selected)
		assert.Equal(t, "mock://cover", selected.ImageURL)
		assert.Equal(t, []*models.Anime{english, nil, portuguese}, providerResults, "sorting must not reorder provider-owned slices")
	})

	t.Run("source selector maps exact kinds", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			src  string
			want []source.SourceKind
		}{
			{src: "anidb", want: []source.SourceKind{source.AniDB}},
			{src: "AnimeFire", want: []source.SourceKind{source.AnimeFire}},
			{src: " goyabu ", want: []source.SourceKind{source.Goyabu}},
			{src: "superflix", want: []source.SourceKind{source.SuperFlix}},
			{src: "ptbr", want: []source.SourceKind{source.AnimeFire, source.Goyabu, source.SuperFlix}},
			{src: "pt-br", want: []source.SourceKind{source.AnimeFire, source.Goyabu, source.SuperFlix}},
			{src: "unknown", want: nil},
			{src: "", want: nil},
		}
		for _, tt := range tests {
			t.Run(tt.src, func(t *testing.T) {
				anime := &models.Anime{Name: "Frieren", Source: "Existing"}
				var got []source.SourceKind
				search := func(_ context.Context, _ string, kinds []source.SourceKind) ([]*models.Anime, error) {
					got = kinds
					return []*models.Anime{anime}, nil
				}
				selected, err := searchAnimeEnhanced("frieren", tt.src, search, func([]*models.Anime) (*models.Anime, error) {
					return anime, nil
				}, nil)
				require.NoError(t, err)
				assert.Same(t, anime, selected)
				assert.Equal(t, tt.want, got)
			})
		}
	})

	t.Run("source fallback is strict and case insensitive", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name string
			src  string
			url  string
			want string
		}{
			{name: "explicit anidb source", src: "anidb", url: "opaque", want: "AniDB"},
			{name: "explicit goyabu source", src: "goyabu", url: "opaque", want: "Goyabu"},
			{name: "explicit superflix numeric id", src: "superflix", url: "8143", want: "SuperFlix"},
			{name: "anidb URL", url: "https://ANIDB.app/anime/frieren-1", want: "AniDB"},
			{name: "animefire URL", url: "HTTPS://ANIMEFIRE.PLUS/frieren", want: "Animefire.io"},
			{name: "goyabu URL", url: "https://GOYABU.example/frieren", want: "Goyabu"},
			{name: "sflix URL", url: "https://SFLIX.example/tv/1", want: "SuperFlix"},
			{name: "bare numeric matches nothing", url: "8143", want: ""},
			{name: "bare punctuation matches nothing", url: "abc/def", want: ""},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				anime := &models.Anime{Name: "Frieren", URL: tt.url}
				selected, err := searchAnimeEnhanced("frieren", tt.src, func(context.Context, string, []source.SourceKind) ([]*models.Anime, error) {
					return []*models.Anime{anime}, nil
				}, func([]*models.Anime) (*models.Anime, error) { return anime, nil }, nil)
				require.NoError(t, err)
				assert.Same(t, anime, selected)
				assert.Equal(t, tt.want, anime.Source)
			})
		}
	})

	t.Run("selection failures short circuit enrichment", func(t *testing.T) {
		t.Parallel()
		anime := &models.Anime{Name: "Frieren", Source: "AnimeFire"}
		search := func(context.Context, string, []source.SourceKind) ([]*models.Anime, error) {
			return []*models.Anime{anime}, nil
		}
		for _, tt := range []struct {
			name      string
			selection *models.Anime
			err       error
			want      error
			contains  string
		}{
			{name: "back", err: tui.ErrSelectionBack, want: ErrBackToSearch},
			{name: "cancel", err: tui.ErrSelectionCancelled, want: tui.ErrSelectionCancelled},
			{name: "arbitrary", err: context.Canceled, want: context.Canceled},
			{name: "nil without error", contains: "selection returned nil"},
		} {
			t.Run(tt.name, func(t *testing.T) {
				enrichCalled := false
				selected, err := searchAnimeEnhanced("frieren", "", search, func([]*models.Anime) (*models.Anime, error) {
					return tt.selection, tt.err
				}, func(*models.Anime) error {
					enrichCalled = true
					return nil
				})
				assert.Nil(t, selected)
				require.Error(t, err)
				if tt.want != nil {
					assert.ErrorIs(t, err, tt.want)
				}
				if tt.contains != "" {
					assert.Contains(t, err.Error(), tt.contains)
				}
				assert.False(t, enrichCalled)
			})
		}
	})

	t.Run("missing selector is controlled", func(t *testing.T) {
		t.Parallel()
		anime := &models.Anime{Name: "Frieren", Source: "AnimeFire"}
		selected, err := searchAnimeEnhanced("frieren", "", func(context.Context, string, []source.SourceKind) ([]*models.Anime, error) {
			return []*models.Anime{anime}, nil
		}, nil, nil)
		assert.Nil(t, selected)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "selection not configured")
	})

	t.Run("search failures short circuit selection", func(t *testing.T) {
		t.Parallel()
		for _, tt := range []struct {
			name   string
			search SearchFetchFunc
			want   error
		}{
			{name: "missing dependency", search: nil},
			{name: "context cancelled", search: func(context.Context, string, []source.SourceKind) ([]*models.Anime, error) {
				return nil, context.Canceled
			}, want: context.Canceled},
			{name: "nil results", search: func(context.Context, string, []source.SourceKind) ([]*models.Anime, error) { return nil, nil }},
			{name: "only nil entries", search: func(context.Context, string, []source.SourceKind) ([]*models.Anime, error) {
				return []*models.Anime{nil, nil}, nil
			}},
		} {
			t.Run(tt.name, func(t *testing.T) {
				selectCalled := false
				selected, err := searchAnimeEnhanced("frieren", "", tt.search, func([]*models.Anime) (*models.Anime, error) {
					selectCalled = true
					return nil, nil
				}, nil)
				assert.Nil(t, selected)
				require.Error(t, err)
				if tt.want != nil {
					assert.ErrorIs(t, err, tt.want)
				}
				assert.False(t, selectCalled)
			})
		}
	})

	t.Run("enrichment is optional and best effort", func(t *testing.T) {
		t.Parallel()
		for _, enrich := range []func(*models.Anime) error{
			nil,
			func(*models.Anime) error { return errors.New("metadata offline") },
		} {
			anime := &models.Anime{Name: "Frieren", Source: "AnimeFire"}
			selected, err := searchAnimeEnhanced("frieren", "", func(context.Context, string, []source.SourceKind) ([]*models.Anime, error) {
				return []*models.Anime{anime}, nil
			}, func([]*models.Anime) (*models.Anime, error) { return anime, nil }, enrich)
			require.NoError(t, err)
			assert.Same(t, anime, selected)
		}
	})
}
