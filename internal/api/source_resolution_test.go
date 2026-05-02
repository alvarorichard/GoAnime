package api

import (
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
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

func TestSourceKindHelpers(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		kind                 SourceKind
		wantProviderBacked   bool
		wantScraperType      scraper.ScraperType
		wantScraperTypeFound bool
	}{
		{kind: SourceAllAnime, wantProviderBacked: true, wantScraperType: scraper.AllAnimeType, wantScraperTypeFound: true},
		{kind: SourceAnimefire, wantProviderBacked: true, wantScraperType: scraper.AnimefireType, wantScraperTypeFound: true},
		{kind: SourceAnimeDrive, wantProviderBacked: true, wantScraperType: scraper.AnimeDriveType, wantScraperTypeFound: true},
		{kind: SourceGoyabu, wantProviderBacked: true, wantScraperType: scraper.GoyabuType, wantScraperTypeFound: true},
		{kind: SourceNineAnime, wantProviderBacked: false, wantScraperType: scraper.NineAnimeType, wantScraperTypeFound: true},
		{kind: SourceFlixHQ, wantProviderBacked: false, wantScraperType: scraper.FlixHQType, wantScraperTypeFound: true},
		{kind: SourceSuperFlix, wantProviderBacked: false, wantScraperType: scraper.SuperFlixType, wantScraperTypeFound: true},
		{kind: SourceUnknown, wantProviderBacked: false, wantScraperTypeFound: false},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(string(tc.kind), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.wantProviderBacked, tc.kind.IsProviderBacked())

			gotType, ok := tc.kind.ScraperType()
			assert.Equal(t, tc.wantScraperTypeFound, ok)
			if tc.wantScraperTypeFound {
				assert.Equal(t, tc.wantScraperType, gotType)
			}

			resolved := ResolvedSource{Kind: tc.kind, Name: string(tc.kind)}
			assert.Equal(t, tc.wantProviderBacked, resolved.IsProviderBacked())
		})
	}
}

func TestResolvedSourceApplyNoOp(t *testing.T) {
	t.Parallel()

	anime := &models.Anime{Source: "keep"}
	ResolvedSource{}.Apply(anime)
	assert.Equal(t, "keep", anime.Source)

	ResolvedSource{Name: "AllAnime"}.Apply(nil)
}

func TestResolveSourceNilAnime(t *testing.T) {
	t.Parallel()

	_, err := ResolveSource(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil anime")
}

func TestAllAnimeIDHelpers(t *testing.T) {
	t.Parallel()

	isShortIDCases := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "short id with letters and numbers", value: "naruto123abc", want: true},
		{name: "url is not short id", value: "https://allanime.to/anime/naruto123abc", want: false},
		{name: "numeric value is not short id", value: "123456", want: false},
		{name: "too short", value: "abc", want: false},
		{name: "empty", value: "", want: false},
	}

	for _, tc := range isShortIDCases {
		tc := tc
		t.Run("IsAllAnimeShortID/"+tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, IsAllAnimeShortID(tc.value))
		})
	}

	extractCases := []struct {
		name  string
		value string
		want  string
	}{
		{name: "short id preserved", value: "naruto123abc", want: "naruto123abc"},
		{name: "allanime url extracts short id", value: "https://allanime.to/anime/naruto123abc/watch", want: "naruto123abc"},
		{name: "blank returns blank", value: " ", want: ""},
		{name: "non all-anime url preserved", value: "https://example.com/watch/naruto", want: "https://example.com/watch/naruto"},
	}

	for _, tc := range extractCases {
		tc := tc
		t.Run("ExtractAllAnimeID/"+tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, ExtractAllAnimeID(tc.value))
		})
	}
}
