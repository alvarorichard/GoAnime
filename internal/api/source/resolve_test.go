package source

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
)

func TestResolve_ExplicitSource(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		wantKind SourceKind
	}{
		{"AllAnime", "AllAnime", AllAnime},
		{"AnimeFire via Animefire.io", "Animefire.io", AnimeFire},
		{"AnimeFire direct", "AnimeFire", AnimeFire},
		{"FlixHQ", "FlixHQ", FlixHQ},
		{"SFlix", "SFlix", SFlix},
		{"9Anime", "9Anime", NineAnime},
		{"AnimeDrive", "AnimeDrive", AnimeDrive},
		{"Goyabu", "Goyabu", Goyabu},
		{"SuperFlix", "SuperFlix", SuperFlix},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anime := &models.Anime{Source: tt.source}
			got := Resolve(anime)
			if got.Kind != tt.wantKind {
				t.Errorf("Resolve(Source=%q) = %s (%s), want %s", tt.source, got.Kind, got.Reason, tt.wantKind)
			}
		})
	}
}

func TestResolve_ExplicitSourceTrumpsURL(t *testing.T) {
	anime := &models.Anime{
		Source: "9Anime",
		URL:    "https://animefire.plus/something",
	}
	got := Resolve(anime)
	if got.Kind != NineAnime {
		t.Errorf("explicit Source should win over URL, got %s (%s)", got.Kind, got.Reason)
	}
}

func TestResolve_MediaType(t *testing.T) {
	tests := []struct {
		name      string
		mediaType models.MediaType
		wantKind  SourceKind
	}{
		{"movie", models.MediaTypeMovie, FlixHQ},
		{"tv", models.MediaTypeTV, FlixHQ},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anime := &models.Anime{MediaType: tt.mediaType}
			got := Resolve(anime)
			if got.Kind != tt.wantKind {
				t.Errorf("Resolve(MediaType=%s) = %s, want %s", tt.mediaType, got.Kind, tt.wantKind)
			}
		})
	}
}

func TestResolve_NameTags(t *testing.T) {
	tests := []struct {
		name     string
		animName string
		wantKind SourceKind
	}{
		{"english tag", "Naruto [English]", AllAnime},
		{"animefire tag", "Naruto [AnimeFire]", AnimeFire},
		{"animedrive tag", "Naruto [AnimeDrive]", AnimeDrive},
		{"goyabu tag", "Naruto [Goyabu]", Goyabu},
		{"superflix tag", "Naruto [SuperFlix]", SuperFlix},
		{"9anime tag", "Naruto [9Anime]", NineAnime},
		{"multilanguage tag", "Naruto [Multilanguage]", NineAnime},
		{"movie tag", "Inception [Movie]", FlixHQ},
		{"tv tag", "Breaking Bad [TV]", FlixHQ},
		{"flixhq tag", "Movie [FlixHQ]", FlixHQ},
		{"sflix tag", "Movie [SFlix]", SFlix},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anime := &models.Anime{Name: tt.animName}
			got := Resolve(anime)
			if got.Kind != tt.wantKind {
				t.Errorf("Resolve(Name=%q) = %s (%s), want %s", tt.animName, got.Kind, got.Reason, tt.wantKind)
			}
		})
	}
}

func TestResolve_URLPatterns(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantKind SourceKind
	}{
		{"animesdrive URL", "https://animesdrive.blog/naruto", AnimeDrive},
		{"animefire URL", "https://animefire.plus/naruto", AnimeFire},
		{"goyabu URL", "https://goyabu.to/naruto", Goyabu},
		{"allanime URL", "https://allanime.to/anime/abc", AllAnime},
		{"flixhq URL", "https://flixhq.to/movie/inception", FlixHQ},
		{"sflix URL", "https://sflix.to/movie/inception", SFlix},
		{"9anime URL", "https://9anime.to/watch/naruto", NineAnime},
		{"superflix URL", "https://superflix.to/naruto", SuperFlix},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anime := &models.Anime{URL: tt.url}
			got := Resolve(anime)
			if got.Kind != tt.wantKind {
				t.Errorf("Resolve(URL=%q) = %s (%s), want %s", tt.url, got.Kind, got.Reason, tt.wantKind)
			}
		})
	}
}

func TestResolve_AnimeDriveNotConfusedWithAllAnime(t *testing.T) {
	// This is the latent bug in the current codebase:
	// animesdrive URL must resolve to AnimeDrive, not AllAnime.
	anime := &models.Anime{URL: "https://animesdrive.blog/ep/naruto-1"}
	got := Resolve(anime)
	if got.Kind != AnimeDrive {
		t.Errorf("animesdrive should resolve to AnimeDrive, got %s (%s)", got.Kind, got.Reason)
	}
}

func TestResolve_ShortID(t *testing.T) {
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
			anime := &models.Anime{URL: tt.url}
			got := Resolve(anime)
			if got.Kind != tt.wantKind {
				t.Errorf("Resolve(URL=%q) = %s, want %s", tt.url, got.Kind, tt.wantKind)
			}
		})
	}
}

func TestResolve_NumericOnlyIsNotShortID(t *testing.T) {
	// Purely numeric strings are NOT AllAnime short IDs.
	anime := &models.Anime{URL: "8143"}
	got := Resolve(anime)
	if got.Kind == AllAnime && got.Reason == "short ID" {
		t.Error("purely numeric '8143' should not match as AllAnime short ID")
	}
}

func TestResolve_PTBRFallback(t *testing.T) {
	anime := &models.Anime{Name: "Naruto [PT-BR]"}
	got := Resolve(anime)
	if got.Kind != AnimeFire {
		t.Errorf("[PT-BR] tag without source should default to AnimeFire, got %s", got.Kind)
	}
}

func TestResolve_NilAnime(t *testing.T) {
	got := Resolve(nil)
	if got.Kind != Unknown {
		t.Errorf("nil anime should return Unknown, got %s", got.Kind)
	}
}

func TestResolve_EmptyAnime(t *testing.T) {
	got := Resolve(&models.Anime{})
	if got.Kind != Unknown {
		t.Errorf("empty anime should return Unknown, got %s", got.Kind)
	}
}

func TestResolve_BestEffortKind(t *testing.T) {
	r := ResolvedSource{Kind: Unknown, Reason: "test"}
	if r.BestEffortKind() != AllAnime {
		t.Errorf("BestEffortKind for Unknown should be AllAnime, got %s", r.BestEffortKind())
	}

	r2 := ResolvedSource{Kind: FlixHQ, Reason: "test"}
	if r2.BestEffortKind() != FlixHQ {
		t.Errorf("BestEffortKind for FlixHQ should be FlixHQ, got %s", r2.BestEffortKind())
	}
}

func TestResolveURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantKind SourceKind
	}{
		{"animefire", "https://animefire.plus/ep/naruto-1", AnimeFire},
		{"animesdrive", "https://animesdrive.blog/ep/naruto", AnimeDrive},
		{"goyabu", "https://goyabu.to/ep/naruto-1", Goyabu},
		{"allanime", "https://allanime.to/anime/hHjXnUTda", AllAnime},
		{"short ID", "hHjXnUTda", AllAnime},
		{"empty", "", Unknown},
		{"unknown domain", "https://example.com/video", Unknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveURL(tt.url)
			if got.Kind != tt.wantKind {
				t.Errorf("ResolveURL(%q) = %s (%s), want %s", tt.url, got.Kind, got.Reason, tt.wantKind)
			}
		})
	}
}

func TestIsAllAnimeShortID(t *testing.T) {
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
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false}, // 31 chars
		{"abc def", false},                         // space
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsAllAnimeShortID(tt.input); got != tt.want {
				t.Errorf("IsAllAnimeShortID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractAllAnimeID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hHjXnUTda", "hHjXnUTda"},
		{"https://allanime.to/anime/hHjXnUTda", "hHjXnUTda"},
		{"https://example.com/8143", "https://example.com/8143"}, // numeric-only not extracted
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ExtractAllAnimeID(tt.input); got != tt.want {
				t.Errorf("ExtractAllAnimeID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestScraperTypeFor(t *testing.T) {
	st, ok := ScraperTypeFor(AllAnime)
	if !ok {
		t.Fatal("ScraperTypeFor(AllAnime) should return true")
	}
	if st != 0 { // AllAnimeType = iota = 0
		t.Errorf("ScraperTypeFor(AllAnime) = %d, want 0", st)
	}

	_, ok = ScraperTypeFor(Unknown)
	if ok {
		t.Error("ScraperTypeFor(Unknown) should return false")
	}
}
