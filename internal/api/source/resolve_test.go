package source

import (
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
)

func TestResolve(t *testing.T) {
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
			wantKind:       AllAnime,
			wantName:       "AllAnime",
			wantReasonLike: "short ID",
		},
		{
			name: "explicit source wins over URL",
			anime: models.Anime{
				Name:   "Naruto",
				URL:    "https://animefire.plus/naruto",
				Source: "Goyabu",
			},
			wantKind:       Goyabu,
			wantName:       "Goyabu",
			wantReasonLike: "explicit",
		},
		{
			name: "AnimeFire by URL without PT-BR tag",
			anime: models.Anime{
				Name: "Naruto",
				URL:  "https://animefire.plus/animes/naruto",
			},
			wantKind:       AnimeFire,
			wantName:       "Animefire.io",
			wantReasonLike: "URL",
		},
		{
			name: "PT-BR tag plus URL resolves Goyabu",
			anime: models.Anime{
				Name: "[PT-BR] Naruto",
				URL:  "https://goyabu.to/anime/naruto",
			},
			wantKind:       Goyabu,
			wantName:       "Goyabu",
			wantReasonLike: "PT-BR",
		},
		{
			name: "AnimeDrive remains distinct from AnimeFire",
			anime: models.Anime{
				Name: "[PT-BR] Naruto",
				URL:  "https://animesdrive.blog/anime/naruto",
			},
			wantKind:       AnimeDrive,
			wantName:       "AnimeDrive",
			wantReasonLike: "PT-BR",
		},
		{
			name: "FlixHQ by media type",
			anime: models.Anime{
				Name:      "Inception",
				MediaType: models.MediaTypeMovie,
			},
			wantKind:       FlixHQ,
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
			wantKind:       NineAnime,
			wantName:       "9Anime",
			wantReasonLike: "explicit",
		},
		{
			name: "explicit source fuzzy match remains declarative",
			anime: models.Anime{
				Name:   "Naruto",
				URL:    "https://animefire.plus/animes/naruto",
				Source: "AnimeFire Legacy",
			},
			wantKind:       AnimeFire,
			wantName:       "Animefire.io",
			wantReasonLike: "explicit",
		},
		{
			name: "PT-BR tag plus AnimeFire URL resolves AnimeFire",
			anime: models.Anime{
				Name: "[PT-BR] Naruto",
				URL:  "https://animefire.plus/animes/naruto",
			},
			wantKind:       AnimeFire,
			wantName:       "Animefire.io",
			wantReasonLike: "PT-BR",
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

			resolved, err := Resolve(&tc.anime)
			if tc.wantErrLike != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErrLike)
				}
				if !strings.Contains(err.Error(), tc.wantErrLike) {
					t.Fatalf("error %q does not contain %q", err.Error(), tc.wantErrLike)
				}
				return
			}

			if err != nil {
				t.Fatalf("Resolve returned unexpected error: %v", err)
			}

			if resolved.Kind != tc.wantKind {
				t.Fatalf("Resolve kind = %s, want %s", resolved.Kind, tc.wantKind)
			}
			if resolved.Name != tc.wantName {
				t.Fatalf("Resolve name = %s, want %s", resolved.Name, tc.wantName)
			}
			if !strings.Contains(strings.ToLower(resolved.Reason), strings.ToLower(tc.wantReasonLike)) {
				t.Fatalf("Resolve reason = %q, want substring %q", resolved.Reason, tc.wantReasonLike)
			}

			resolved.Apply(&tc.anime)
			if tc.anime.Source != tc.wantName {
				t.Fatalf("Apply set Source = %q, want %q", tc.anime.Source, tc.wantName)
			}
		})
	}
}

func TestResolveURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		url      string
		wantKind SourceKind
	}{
		{"animefire", "https://animefire.plus/ep/naruto-1", AnimeFire},
		{"animesdrive", "https://animesdrive.blog/ep/naruto", AnimeDrive},
		{"goyabu", "https://goyabu.to/ep/naruto-1", Goyabu},
		{"allanime", "https://allanime.to/anime/hHjXnUTda", AllAnime},
		{"superflix", "https://superflixapi.rest/serie/123", SuperFlix},
		{"short ID", "hHjXnUTda", AllAnime},
		{"empty", "", Unknown},
		{"unknown domain", "https://example.com/video", Unknown},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ResolveURL(tt.url)
			if got.Kind != tt.wantKind {
				t.Fatalf("ResolveURL(%q) = %s, want %s", tt.url, got.Kind, tt.wantKind)
			}
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
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false},
		{"abc def", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			if got := IsAllAnimeShortID(tt.input); got != tt.want {
				t.Fatalf("IsAllAnimeShortID(%q) = %v, want %v", tt.input, got, tt.want)
			}
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
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			if got := ExtractAllAnimeID(tt.input); got != tt.want {
				t.Fatalf("ExtractAllAnimeID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestScraperTypeFor(t *testing.T) {
	t.Parallel()

	st, ok := ScraperTypeFor(AllAnime)
	if !ok {
		t.Fatal("ScraperTypeFor(AllAnime) should return true")
	}
	if st != 0 {
		t.Fatalf("ScraperTypeFor(AllAnime) = %d, want 0", st)
	}

	if _, ok := ScraperTypeFor(Unknown); ok {
		t.Fatal("ScraperTypeFor(Unknown) should return false")
	}
}
