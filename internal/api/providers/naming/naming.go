// Package naming provides Jellyfin/Plex-compatible file and directory naming
// for downloaded media. All sources use this to generate correct paths.
//
// Output format follows Jellyfin/Plex conventions:
//
//	Shows/
//	  Series Name (2020)/
//	    Season 01/
//	      Series Name S01E01.mkv
//	      Series Name S01E02.mkv
//	    Season 02/
//	      Series Name S02E01.mkv
//
//	Movies/
//	  Movie Name (2020)/
//	    Movie Name (2020).mkv
package naming

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/alvarorichard/Goanime/internal/models"
)

// MediaInfo holds the metadata needed to generate Jellyfin/Plex-compatible paths.
// Populated by metadata enrichment (AniList, TMDB, IMDB) or inferred from scraper data.
type MediaInfo struct {
	// Title is the canonical series/movie title (English preferred, romaji fallback).
	Title string

	// Year is the release year (e.g. "2020"). May be empty if unknown.
	Year string

	// Season number (1-based). 0 means unknown/not applicable.
	Season int

	// Episode number (1-based). 0 means unknown.
	Episode int

	// EpisodeTitle is optional (used in NFO, not in filename).
	EpisodeTitle string

	// IsMovie is true for movies (different directory structure).
	IsMovie bool

	// Extension is the file extension including dot (e.g. ".mkv").
	Extension string
}

// sanitize removes characters not allowed in filenames across OS.
var unsafeChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

// SanitizeFilename removes characters unsafe for filenames on all platforms.
func SanitizeFilename(name string) string {
	clean := unsafeChars.ReplaceAllString(name, "")
	clean = strings.TrimSpace(clean)
	// Collapse multiple spaces
	spaceRun := regexp.MustCompile(`\s{2,}`)
	clean = spaceRun.ReplaceAllString(clean, " ")
	// Remove trailing dots (Windows)
	clean = strings.TrimRight(clean, ".")
	if clean == "" {
		clean = "Unknown"
	}
	return clean
}

// CleanTitle removes source tags like [English], [PT-BR], [AllAnime] etc. from a title.
func CleanTitle(title string) string {
	tagPattern := regexp.MustCompile(`\s*\[(?:English|PT-BR|Portuguese|Multilanguage|AllAnime|AnimeFire|AnimeDrive|Goyabu|SuperFlix|FlixHQ|SFlix|9Anime|Movie|TV)\]`)
	clean := tagPattern.ReplaceAllString(title, "")
	clean = strings.TrimSpace(clean)
	if clean == "" {
		return title
	}
	return clean
}

// SeriesDir returns the top-level series directory name.
// Format: "Series Name (2020)" or "Series Name" if year unknown.
func SeriesDir(info *MediaInfo) string {
	title := SanitizeFilename(CleanTitle(info.Title))
	if info.Year != "" {
		return fmt.Sprintf("%s (%s)", title, info.Year)
	}
	return title
}

// SeasonDir returns the season subdirectory name.
// Format: "Season 01", "Season 02", etc. Season 0 is "Specials".
func SeasonDir(season int) string {
	if season <= 0 {
		return "Season 00"
	}
	return fmt.Sprintf("Season %02d", season)
}

// EpisodeFilename returns a Jellyfin/Plex-compatible episode filename.
// Format: "Series Name S01E01.mkv"
func EpisodeFilename(info *MediaInfo) string {
	title := SanitizeFilename(CleanTitle(info.Title))
	ext := info.Extension
	if ext == "" {
		ext = ".mkv"
	}

	if info.IsMovie {
		if info.Year != "" {
			return fmt.Sprintf("%s (%s)%s", title, info.Year, ext)
		}
		return title + ext
	}

	season := info.Season
	if season <= 0 {
		season = 1
	}
	episode := info.Episode
	if episode <= 0 {
		episode = 1
	}
	return fmt.Sprintf("%s S%02dE%02d%s", title, season, episode, ext)
}

// MovieFilename returns a Jellyfin/Plex-compatible movie filename.
// Format: "Movie Name (2020).mkv"
func MovieFilename(info *MediaInfo) string {
	info.IsMovie = true
	return EpisodeFilename(info)
}

// FullPath returns the complete relative path for a media file.
//
// For series: "Shows/Series Name (2020)/Season 01/Series Name S01E01.mkv"
// For movies: "Movies/Movie Name (2020)/Movie Name (2020).mkv"
func FullPath(info *MediaInfo) string {
	if info.IsMovie {
		dir := SeriesDir(info)
		filename := MovieFilename(info)
		return filepath.Join("Movies", dir, filename)
	}

	dir := SeriesDir(info)
	seasonDir := SeasonDir(info.Season)
	filename := EpisodeFilename(info)
	return filepath.Join("Shows", dir, seasonDir, filename)
}

// FromAnimeEpisode creates a MediaInfo from an Anime and Episode model.
// Uses AniList data if available, falls back to scraped data.
func FromAnimeEpisode(anime *models.Anime, episode *models.Episode, season int) *MediaInfo {
	info := &MediaInfo{
		Extension: ".mkv",
	}

	// Title: prefer AniList English > Romaji > scraped Name
	info.Title = bestTitle(anime)

	// Year
	info.Year = extractYear(anime)

	// Season
	info.Season = season
	if info.Season <= 0 && anime.CurrentSeason > 0 {
		info.Season = anime.CurrentSeason
	}
	if info.Season <= 0 {
		info.Season = 1
	}

	// Episode number
	if episode != nil {
		info.Episode = episode.Num
		if info.Episode <= 0 {
			info.Episode = parseEpisodeNumber(episode.Number)
		}
		if episode.Title.English != "" {
			info.EpisodeTitle = episode.Title.English
		} else if episode.Title.Romaji != "" {
			info.EpisodeTitle = episode.Title.Romaji
		}
	}

	// Movie detection
	info.IsMovie = anime.MediaType == models.MediaTypeMovie

	return info
}

// bestTitle extracts the best available title from an Anime model.
func bestTitle(anime *models.Anime) string {
	if anime.Details.Title.English != "" {
		return anime.Details.Title.English
	}
	if anime.Details.Title.Romaji != "" {
		return anime.Details.Title.Romaji
	}
	// Fallback: clean the scraped name
	return CleanTitle(anime.Name)
}

// extractYear gets the year from Anime metadata.
func extractYear(anime *models.Anime) string {
	if anime.Year != "" {
		return anime.Year
	}
	// Try TMDB
	if anime.TMDBDetails != nil {
		if anime.TMDBDetails.FirstAirDate != "" && len(anime.TMDBDetails.FirstAirDate) >= 4 {
			return anime.TMDBDetails.FirstAirDate[:4]
		}
		if anime.TMDBDetails.ReleaseDate != "" && len(anime.TMDBDetails.ReleaseDate) >= 4 {
			return anime.TMDBDetails.ReleaseDate[:4]
		}
	}
	return ""
}

// parseEpisodeNumber extracts a number from an episode number string.
func parseEpisodeNumber(s string) int {
	if s == "" {
		return 0
	}
	n := 0
	for _, r := range s {
		if unicode.IsDigit(r) {
			n = n*10 + int(r-'0')
		} else if n > 0 {
			break
		}
	}
	return n
}
