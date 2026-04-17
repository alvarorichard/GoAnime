package naming

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Normal Title", "Normal Title"},
		{"Title: With Colons", "Title With Colons"},
		{"Title/With/Slashes", "TitleWithSlashes"},
		{"Title?With*Special<Chars>", "TitleWithSpecialChars"},
		{"  Spaced  Out  ", "Spaced Out"},
		{"Trailing...", "Trailing"},
		{"", "Unknown"},
		{"....", "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := SanitizeFilename(tt.input); got != tt.want {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCleanTitle(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Naruto [English]", "Naruto"},
		{"Naruto [PT-BR]", "Naruto"},
		{"Naruto [AllAnime] [English]", "Naruto"},
		{"Naruto [AnimeFire]", "Naruto"},
		{"Naruto [Movie]", "Naruto"},
		{"Naruto", "Naruto"},
		{"Attack on Titan [Multilanguage]", "Attack on Titan"},
		// If only tags, return original
		{"[English]", "[English]"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := CleanTitle(tt.input); got != tt.want {
				t.Errorf("CleanTitle(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSeriesDir(t *testing.T) {
	tests := []struct {
		name string
		info *MediaInfo
		want string
	}{
		{"with year", &MediaInfo{Title: "Naruto", Year: "2002"}, "Naruto (2002)"},
		{"without year", &MediaInfo{Title: "Naruto"}, "Naruto"},
		{"with tags", &MediaInfo{Title: "Naruto [English]", Year: "2002"}, "Naruto (2002)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SeriesDir(tt.info); got != tt.want {
				t.Errorf("SeriesDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSeasonDir(t *testing.T) {
	tests := []struct {
		season int
		want   string
	}{
		{0, "Season 00"},
		{-1, "Season 00"},
		{1, "Season 01"},
		{2, "Season 02"},
		{10, "Season 10"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := SeasonDir(tt.season); got != tt.want {
				t.Errorf("SeasonDir(%d) = %q, want %q", tt.season, got, tt.want)
			}
		})
	}
}

func TestEpisodeFilename(t *testing.T) {
	tests := []struct {
		name string
		info *MediaInfo
		want string
	}{
		{
			"standard episode",
			&MediaInfo{Title: "Naruto", Season: 1, Episode: 3, Extension: ".mkv"},
			"Naruto S01E03.mkv",
		},
		{
			"season 2",
			&MediaInfo{Title: "Attack on Titan", Season: 2, Episode: 12, Extension: ".mkv"},
			"Attack on Titan S02E12.mkv",
		},
		{
			"default extension",
			&MediaInfo{Title: "Naruto", Season: 1, Episode: 1},
			"Naruto S01E01.mkv",
		},
		{
			"mp4 extension",
			&MediaInfo{Title: "Naruto", Season: 1, Episode: 1, Extension: ".mp4"},
			"Naruto S01E01.mp4",
		},
		{
			"zero season defaults to 1",
			&MediaInfo{Title: "Naruto", Season: 0, Episode: 5},
			"Naruto S01E05.mkv",
		},
		{
			"zero episode defaults to 1",
			&MediaInfo{Title: "Naruto", Season: 1, Episode: 0},
			"Naruto S01E01.mkv",
		},
		{
			"movie",
			&MediaInfo{Title: "Spirited Away", Year: "2001", IsMovie: true, Extension: ".mkv"},
			"Spirited Away (2001).mkv",
		},
		{
			"movie without year",
			&MediaInfo{Title: "Spirited Away", IsMovie: true},
			"Spirited Away.mkv",
		},
		{
			"title with tags cleaned",
			&MediaInfo{Title: "Naruto [English]", Season: 1, Episode: 1},
			"Naruto S01E01.mkv",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EpisodeFilename(tt.info); got != tt.want {
				t.Errorf("EpisodeFilename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFullPath_Series(t *testing.T) {
	info := &MediaInfo{
		Title:   "Attack on Titan",
		Year:    "2013",
		Season:  2,
		Episode: 5,
	}
	got := FullPath(info)
	want := "Shows/Attack on Titan (2013)/Season 02/Attack on Titan S02E05.mkv"
	if got != want {
		t.Errorf("FullPath() = %q, want %q", got, want)
	}
}

func TestFullPath_Movie(t *testing.T) {
	info := &MediaInfo{
		Title:   "Spirited Away",
		Year:    "2001",
		IsMovie: true,
	}
	got := FullPath(info)
	want := "Movies/Spirited Away (2001)/Spirited Away (2001).mkv"
	if got != want {
		t.Errorf("FullPath() = %q, want %q", got, want)
	}
}

func TestFullPath_Specials(t *testing.T) {
	info := &MediaInfo{
		Title:   "Naruto",
		Year:    "2002",
		Season:  0,
		Episode: 1,
	}
	got := FullPath(info)
	// Season 0 in dir but S01E01 in filename (season defaults to 1 for filename)
	want := "Shows/Naruto (2002)/Season 00/Naruto S01E01.mkv"
	if got != want {
		t.Errorf("FullPath() = %q, want %q", got, want)
	}
}

func TestFromAnimeEpisode(t *testing.T) {
	tests := []struct {
		name    string
		anime   *models.Anime
		episode *models.Episode
		season  int
		want    *MediaInfo
	}{
		{
			name: "AniList English title",
			anime: &models.Anime{
				Name: "Naruto [English]",
				Year: "2002",
				Details: models.AniListDetails{
					Title: models.Title{English: "Naruto", Romaji: "NARUTO"},
				},
			},
			episode: &models.Episode{Num: 3},
			season:  1,
			want: &MediaInfo{
				Title:     "Naruto",
				Year:      "2002",
				Season:    1,
				Episode:   3,
				Extension: ".mkv",
			},
		},
		{
			name: "AniList Romaji fallback",
			anime: &models.Anime{
				Name: "Shingeki no Kyojin [PT-BR]",
				Year: "2013",
				Details: models.AniListDetails{
					Title: models.Title{Romaji: "Shingeki no Kyojin"},
				},
			},
			episode: &models.Episode{Number: "12"},
			season:  2,
			want: &MediaInfo{
				Title:     "Shingeki no Kyojin",
				Year:      "2013",
				Season:    2,
				Episode:   12,
				Extension: ".mkv",
			},
		},
		{
			name: "scraped name fallback",
			anime: &models.Anime{
				Name: "My Anime [AnimeFire]",
			},
			episode: &models.Episode{Num: 1},
			season:  1,
			want: &MediaInfo{
				Title:     "My Anime",
				Year:      "",
				Season:    1,
				Episode:   1,
				Extension: ".mkv",
			},
		},
		{
			name: "movie with TMDB year",
			anime: &models.Anime{
				Name:      "Spirited Away [Movie]",
				MediaType: models.MediaTypeMovie,
				TMDBDetails: &models.TMDBDetails{
					ReleaseDate: "2001-07-20",
				},
			},
			episode: nil,
			season:  0,
			want: &MediaInfo{
				Title:     "Spirited Away",
				Year:      "2001",
				Season:    1,
				Episode:   0,
				IsMovie:   true,
				Extension: ".mkv",
			},
		},
		{
			name: "CurrentSeason from anime",
			anime: &models.Anime{
				Name:          "Naruto",
				Year:          "2002",
				CurrentSeason: 3,
			},
			episode: &models.Episode{Num: 5},
			season:  0,
			want: &MediaInfo{
				Title:     "Naruto",
				Year:      "2002",
				Season:    3,
				Episode:   5,
				Extension: ".mkv",
			},
		},
		{
			name: "episode with string number",
			anime: &models.Anime{
				Name: "Test",
			},
			episode: &models.Episode{Number: "ep42"},
			season:  1,
			want: &MediaInfo{
				Title:     "Test",
				Year:      "",
				Season:    1,
				Episode:   42,
				Extension: ".mkv",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromAnimeEpisode(tt.anime, tt.episode, tt.season)
			if got.Title != tt.want.Title {
				t.Errorf("Title = %q, want %q", got.Title, tt.want.Title)
			}
			if got.Year != tt.want.Year {
				t.Errorf("Year = %q, want %q", got.Year, tt.want.Year)
			}
			if got.Season != tt.want.Season {
				t.Errorf("Season = %d, want %d", got.Season, tt.want.Season)
			}
			if got.Episode != tt.want.Episode {
				t.Errorf("Episode = %d, want %d", got.Episode, tt.want.Episode)
			}
			if got.IsMovie != tt.want.IsMovie {
				t.Errorf("IsMovie = %v, want %v", got.IsMovie, tt.want.IsMovie)
			}
		})
	}
}

func TestParseEpisodeNumber(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"1", 1},
		{"42", 42},
		{"ep12", 12},
		{"Episode 3", 3},
		{"S01E05", 1}, // first number found
		{"", 0},
		{"abc", 0},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := parseEpisodeNumber(tt.input); got != tt.want {
				t.Errorf("parseEpisodeNumber(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// TestFullPath_Integration tests the complete flow from anime model to file path
func TestFullPath_Integration(t *testing.T) {
	tests := []struct {
		name    string
		anime   *models.Anime
		episode *models.Episode
		season  int
		want    string
	}{
		{
			name: "Naruto Season 1 Episode 3",
			anime: &models.Anime{
				Name: "Naruto [English]",
				Year: "2002",
				Details: models.AniListDetails{
					Title: models.Title{English: "Naruto"},
				},
			},
			episode: &models.Episode{Num: 3},
			season:  1,
			want:    "Shows/Naruto (2002)/Season 01/Naruto S01E03.mkv",
		},
		{
			name: "Attack on Titan Season 2 Episode 12",
			anime: &models.Anime{
				Name: "Shingeki no Kyojin [PT-BR]",
				Year: "2013",
				Details: models.AniListDetails{
					Title: models.Title{English: "Attack on Titan", Romaji: "Shingeki no Kyojin"},
				},
			},
			episode: &models.Episode{Num: 12},
			season:  2,
			want:    "Shows/Attack on Titan (2013)/Season 02/Attack on Titan S02E12.mkv",
		},
		{
			name: "Movie download",
			anime: &models.Anime{
				Name:      "Spirited Away [Movie]",
				MediaType: models.MediaTypeMovie,
				TMDBDetails: &models.TMDBDetails{
					ReleaseDate: "2001-07-20",
				},
				Details: models.AniListDetails{
					Title: models.Title{English: "Spirited Away"},
				},
			},
			episode: nil,
			season:  0,
			want:    "Movies/Spirited Away (2001)/Spirited Away (2001).mkv",
		},
		{
			name: "Anime without AniList data",
			anime: &models.Anime{
				Name: "Some Anime [Goyabu]",
				Year: "2020",
			},
			episode: &models.Episode{Num: 7},
			season:  1,
			want:    "Shows/Some Anime (2020)/Season 01/Some Anime S01E07.mkv",
		},
		{
			name: "TV show from FlixHQ",
			anime: &models.Anime{
				Name:      "Breaking Bad [TV]",
				MediaType: models.MediaTypeTV,
				Year:      "2008",
				Details: models.AniListDetails{
					Title: models.Title{English: "Breaking Bad"},
				},
			},
			episode: &models.Episode{Num: 5},
			season:  3,
			want:    "Shows/Breaking Bad (2008)/Season 03/Breaking Bad S03E05.mkv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := FromAnimeEpisode(tt.anime, tt.episode, tt.season)
			got := FullPath(info)
			if got != tt.want {
				t.Errorf("\nFullPath() = %q\nwant       %q", got, tt.want)
			}
		})
	}
}
