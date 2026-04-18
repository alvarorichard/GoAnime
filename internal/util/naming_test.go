package util

import (
	"fmt"
	"testing"
)

func TestSanitizeForFilename(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"Black Clover (Dublado) 7.27 A14", "Black Clover (Dublado)"},
		{"Naruto Shippuuden 8.50 A12", "Naruto Shippuuden"},
		{"One Piece 9.12 L", "One Piece"},
		{"Demon Slayer", "Demon Slayer"},
		{"Attack on Titan (Legendado) 9.00 AL", "Attack on Titan (Legendado)"},
		{"Jujutsu Kaisen 2nd Season 8.60 A14", "Jujutsu Kaisen 2nd Season"},
		{"Solo Leveling 8.21 A14", "Solo Leveling"},
		{"My Hero Academia 7.50 L", "My Hero Academia"},
		{"Bleach: Thousand-Year Blood War", "Bleach Thousand-Year Blood War"},
		{"[Movies/TV] Dexter", "Dexter"},
		{"[Movie] 2 Fast 2 Furious", "2 Fast 2 Furious"},
		{"[TV] Breaking Bad", "Breaking Bad"},
		{"[English] Naruto", "Naruto"},
		{"[Portuguese] One Piece", "One Piece"},
		{"[PT-BR] Dragon Ball Super Dublado", "Dragon Ball Super Dublado"},
		// 9anime-specific patterns
		{"[Multilanguage] Boruto Naruto Next Generations (HD SUB DUB Ep 293/293)", "Boruto Naruto Next Generations"},
		{"Naruto (SUB DUB Ep 220/220)", "Naruto"},
		{"One Piece (HD SUB Ep 1100/1100)", "One Piece"},
		{"Dragon Ball Super (Multilanguage DUB Ep 131)", "Dragon Ball Super"},
		{"[HD] Bleach", "Bleach"},
		{"[Multi Subs] Attack on Titan", "Attack on Titan"},
		{"Demon Slayer (SUB DUB Ep 44/44)", "Demon Slayer"},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got := SanitizeForFilename(tc.in)
			if got != tc.want {
				t.Errorf("SanitizeForFilename(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestFormatPlexEpisodePath(t *testing.T) {
	// Without metadata — backward compatible
	path := FormatPlexEpisodePath("/media/anime", "Black Clover (Dublado) 7.27 A14", 1, 3)
	want := "/media/anime/Black Clover (Dublado)/Season 01/Black Clover (Dublado) - S01E03.mp4"
	if path != want {
		t.Errorf("FormatPlexEpisodePath (no meta) = %q, want %q", path, want)
	}
}

func TestFormatPlexEpisodePathWithMeta(t *testing.T) {
	meta := &MediaMeta{
		Year:   "2017",
		TMDBID: 73223,
		IMDBID: "tt6771578",
	}
	path := FormatPlexEpisodePath("/media/anime", "Black Clover", 1, 3, meta)
	want := "/media/anime/Black Clover (2017) {tmdb-73223} {imdb-tt6771578}/Season 01/Black Clover (2017) - S01E03.mp4"
	if path != want {
		t.Errorf("FormatPlexEpisodePath (with meta) = %q, want %q", path, want)
	}
}

func TestPlexEpisodeFilename(t *testing.T) {
	fn := PlexEpisodeFilename("Naruto Shippuuden 8.50 A12", 2, 15)
	want := "Naruto Shippuuden - S02E15.mp4"
	if fn != want {
		t.Errorf("PlexEpisodeFilename = %q, want %q", fn, want)
	}
}

func TestBuildMediaFolderName(t *testing.T) {
	tests := []struct {
		name string
		meta *MediaMeta
		want string
	}{
		{
			name: "Dois Homens e Meio",
			meta: &MediaMeta{Year: "2003", TMDBID: 2691, IMDBID: "tt0369179"},
			want: "Dois Homens e Meio (2003) {tmdb-2691} {imdb-tt0369179}",
		},
		{
			name: "Attack on Titan",
			meta: &MediaMeta{Year: "2013", TMDBID: 1429, AnilistID: 16498, MalID: 16498},
			want: "Attack on Titan (2013) {tmdb-1429} {anilist-16498} {mal-16498}",
		},
		{
			name: "Naruto",
			meta: nil,
			want: "Naruto",
		},
		{
			name: "Naruto",
			meta: &MediaMeta{},
			want: "Naruto",
		},
		{
			name: "Solo Leveling",
			meta: &MediaMeta{Year: "2024", AnilistID: 151807},
			want: "Solo Leveling (2024) {anilist-151807}",
		},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			got := BuildMediaFolderName(tc.name, tc.meta)
			if got != tc.want {
				t.Errorf("BuildMediaFolderName(%q) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestBuildMediaFileName(t *testing.T) {
	tests := []struct {
		name string
		meta *MediaMeta
		want string
	}{
		{"Movie", &MediaMeta{Year: "2024"}, "Movie (2024)"},
		{"Movie", nil, "Movie"},
		{"Movie", &MediaMeta{TMDBID: 123}, "Movie"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			got := BuildMediaFileName(tc.name, tc.meta)
			if got != tc.want {
				t.Errorf("BuildMediaFileName(%q) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestFormatPlexMoviePath(t *testing.T) {
	meta := &MediaMeta{Year: "2024", TMDBID: 12345, IMDBID: "tt1234567"}
	path := FormatPlexMoviePath("/media/movies", "The Movie", "2024", meta)
	want := "/media/movies/The Movie (2024) {tmdb-12345} {imdb-tt1234567}/The Movie (2024).mp4"
	if path != want {
		t.Errorf("FormatPlexMoviePath = %q, want %q", path, want)
	}
}

func TestFormatPlexMoviePathNoMeta(t *testing.T) {
	path := FormatPlexMoviePath("/media/movies", "The Movie", "2024")
	want := "/media/movies/The Movie (2024)/The Movie (2024).mp4"
	if path != want {
		t.Errorf("FormatPlexMoviePath (no meta) = %q, want %q", path, want)
	}
}

func TestOfficialTitlePreferredOverScraperName(t *testing.T) {
	// When OfficialTitle is set, it should be used instead of the scraper name
	meta := &MediaMeta{
		OfficialTitle: "Two and a Half Men",
		Year:          "2003",
		TMDBID:        2691,
		IMDBID:        "tt0369179",
	}

	// BuildMediaFolderName should use OfficialTitle
	folder := BuildMediaFolderName("Dois Homens e Meio", meta)
	want := "Two and a Half Men (2003) {tmdb-2691} {imdb-tt0369179}"
	if folder != want {
		t.Errorf("BuildMediaFolderName with OfficialTitle = %q, want %q", folder, want)
	}

	// BuildMediaFileName should use OfficialTitle
	fileName := BuildMediaFileName("Dois Homens e Meio", meta)
	wantFile := "Two and a Half Men (2003)"
	if fileName != wantFile {
		t.Errorf("BuildMediaFileName with OfficialTitle = %q, want %q", fileName, wantFile)
	}

	// FormatPlexEpisodePath should use OfficialTitle
	path := FormatPlexEpisodePath("/media/tv", "Dois Homens e Meio", 5, 3, meta)
	wantPath := "/media/tv/Two and a Half Men (2003) {tmdb-2691} {imdb-tt0369179}/Season 05/Two and a Half Men (2003) - S05E03.mp4"
	if path != wantPath {
		t.Errorf("FormatPlexEpisodePath with OfficialTitle = %q, want %q", path, wantPath)
	}

	// FormatPlexMoviePath should use OfficialTitle
	moviePath := FormatPlexMoviePath("/media/movies", "Dois Homens e Meio", "2003", meta)
	wantMovie := "/media/movies/Two and a Half Men (2003) {tmdb-2691} {imdb-tt0369179}/Two and a Half Men (2003).mp4"
	if moviePath != wantMovie {
		t.Errorf("FormatPlexMoviePath with OfficialTitle = %q, want %q", moviePath, wantMovie)
	}

	// PlexEpisodeFilename should use OfficialTitle
	epFile := PlexEpisodeFilename("Dois Homens e Meio", 5, 3, meta)
	wantEpFile := "Two and a Half Men (2003) - S05E03.mp4"
	if epFile != wantEpFile {
		t.Errorf("PlexEpisodeFilename with OfficialTitle = %q, want %q", epFile, wantEpFile)
	}
}

func TestOfficialTitleFallbackToScraperName(t *testing.T) {
	// When OfficialTitle is empty, should fall back to scraper name
	meta := &MediaMeta{
		Year:   "2017",
		TMDBID: 73223,
	}

	folder := BuildMediaFolderName("Black Clover", meta)
	want := "Black Clover (2017) {tmdb-73223}"
	if folder != want {
		t.Errorf("BuildMediaFolderName without OfficialTitle = %q, want %q", folder, want)
	}
}

func TestOfficialTitleAnimeWithAniList(t *testing.T) {
	meta := &MediaMeta{
		OfficialTitle: "Attack on Titan",
		Year:          "2013",
		TMDBID:        1429,
		AnilistID:     16498,
		MalID:         16498,
	}

	path := FormatPlexEpisodePath("/media/anime", "Shingeki no Kyojin", 1, 1, meta)
	want := "/media/anime/Attack on Titan (2013) {tmdb-1429} {anilist-16498} {mal-16498}/Season 01/Attack on Titan (2013) - S01E01.mp4"
	if path != want {
		t.Errorf("FormatPlexEpisodePath with anime OfficialTitle = %q, want %q", path, want)
	}
}

func ExampleSanitizeForFilename() {
	fmt.Println(SanitizeForFilename("Black Clover (Dublado) 7.27 A14"))
	// Output: Black Clover (Dublado)
}
