package api

import (
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveSource(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		anime          models.Anime
		wantKind       SourceKind
		wantName       string
		wantReasonLike string
		wantErrLike    string
	}{
		{
			name: "AllAnime by short ID",
			anime: models.Anime{
				Name: "Naruto",
				URL:  "naruto123abc",
			},
			wantKind:       SourceAllAnime,
			wantName:       "AllAnime",
			wantReasonLike: "short ID",
		},
		{
			name: "PT-BR tag plus URL resolves Goyabu",
			anime: models.Anime{
				Name: "[PT-BR] Naruto",
				URL:  "https://goyabu.to/anime/naruto",
			},
			wantKind:       SourceGoyabu,
			wantName:       "Goyabu",
			wantReasonLike: "PT-BR",
		},
		{
			name: "Goyabu by host",
			anime: models.Anime{
				Name: "Naruto",
				URL:  "https://goyabu.to/anime/naruto",
			},
			wantKind:       SourceGoyabu,
			wantName:       "Goyabu",
			wantReasonLike: "URL",
		},
		{
			name: "FlixHQ by media type",
			anime: models.Anime{
				Name:      "Inception",
				MediaType: models.MediaTypeMovie,
			},
			wantKind:       SourceFlixHQ,
			wantName:       "FlixHQ",
			wantReasonLike: "media type",
		},
		{
			name: "9Anime explicit source",
			anime: models.Anime{
				Name:   "[Multilanguage] Naruto",
				URL:    "8143",
				Source: "9Anime",
			},
			wantKind:       SourceNineAnime,
			wantName:       "9Anime",
			wantReasonLike: "explicit",
		},
		{
			name: "ambiguous PT-BR without URL fails",
			anime: models.Anime{
				Name: "[PT-BR] Naruto",
				URL:  "naruto",
			},
			wantErrLike: "could not resolve PT-BR source",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resolved, err := ResolveSource(&tc.anime)
			if tc.wantErrLike != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrLike)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantKind, resolved.Kind)
			assert.Equal(t, tc.wantName, resolved.Name)
			assert.True(t, strings.Contains(strings.ToLower(resolved.Reason), strings.ToLower(tc.wantReasonLike)))

			resolved.Apply(&tc.anime)
			assert.Equal(t, tc.wantName, tc.anime.Source)
		})
	}
}

func TestDefaultSourceProvidersCoverMigratedKinds(t *testing.T) {
	t.Parallel()

	kinds := []SourceKind{
		SourceAllAnime,
		SourceAnimefire,
		SourceAnimeDrive,
		SourceGoyabu,
	}

	for _, kind := range kinds {
		kind := kind
		t.Run(string(kind), func(t *testing.T) {
			t.Parallel()

			provider, ok := sourceProviderFor(kind)
			require.True(t, ok)
			assert.Equal(t, kind, provider.Kind())
		})
	}
}
