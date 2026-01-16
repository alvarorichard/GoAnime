// Package models contains data structures for media content
package models

import (
	"fmt"
	"strings"
)

// MediaType represents the type of media content
type MediaType string

const (
	MediaTypeAnime MediaType = "anime"
	MediaTypeMovie MediaType = "movie"
	MediaTypeTV    MediaType = "tv"
)

// Media represents any media content (anime, movie, or TV show)
// This is the unified type used across the application for polymorphic handling
type Media struct {
	// Common fields
	Name      string
	URL       string
	ImageURL  string
	Episodes  []Episode
	Source    string    // Identifies the source (AllAnime, AnimeFire, FlixHQ, etc.)
	MediaType MediaType // Type of media (anime, movie, tv)
	Year      string    // Release year
	Quality   string    // Video quality (if available)

	// Anime-specific fields (AniList)
	AnilistID int
	MalID     int
	Details   AniListDetails

	// Movie/TV-specific fields (TMDB/OMDb)
	TMDBID      int          // TMDB ID
	IMDBID      string       // IMDB ID
	TMDBDetails *TMDBDetails // Detailed TMDB information
	Rating      float64      // Rating (0-10)
	Overview    string       // Description/synopsis
	Genres      []string     // Genre list
	Runtime     int          // Runtime in minutes (for movies)
}

// Season represents a TV show season
type Season struct {
	ID       string
	Number   int
	Title    string
	Episodes []Episode
}

// Episode represents a single episode of any media type
type Episode struct {
	Number    string
	Num       int
	URL       string
	Title     TitleDetails
	Aired     string
	Duration  int
	IsFiller  bool
	IsRecap   bool
	Synopsis  string
	SkipTimes SkipTimes
	DataID    string // Used for FlixHQ episode identification
	SeasonID  string // Season identifier for TV shows
}

// TitleDetails contains title information in multiple languages
type TitleDetails struct {
	Romaji   string
	English  string
	Japanese string
}

// Subtitle represents a subtitle track for video playback
type Subtitle struct {
	URL      string
	Language string
	Label    string
	IsForced bool
}

// StreamInfo contains streaming information including video URL and subtitles
type StreamInfo struct {
	VideoURL   string
	Quality    string
	Subtitles  []Subtitle
	Referer    string
	SourceName string
	Headers    map[string]string
}

// IsAnime returns true if the media is anime content
func (m *Media) IsAnime() bool {
	return m.MediaType == MediaTypeAnime
}

// IsMovie returns true if the media is a movie
func (m *Media) IsMovie() bool {
	return m.MediaType == MediaTypeMovie
}

// IsTV returns true if the media is a TV show
func (m *Media) IsTV() bool {
	return m.MediaType == MediaTypeTV
}

// IsMovieOrTV returns true if the media is movie or TV (non-anime)
func (m *Media) IsMovieOrTV() bool {
	return m.MediaType == MediaTypeMovie || m.MediaType == MediaTypeTV
}

// GetDisplayName returns a formatted display name with year and type indicator
func (m *Media) GetDisplayName() string {
	name := m.Name
	if m.Year != "" {
		name += " (" + m.Year + ")"
	}
	return name
}

// GetRatingDisplay returns a formatted rating string
func (m *Media) GetRatingDisplay() string {
	if m.Rating > 0 {
		return fmt.Sprintf("★ %.1f", m.Rating)
	}
	return ""
}

// GetGenresDisplay returns genres as comma-separated string
func (m *Media) GetGenresDisplay() string {
	if len(m.Genres) == 0 {
		return ""
	}
	maxGenres := 3
	if len(m.Genres) < maxGenres {
		maxGenres = len(m.Genres)
	}
	return strings.Join(m.Genres[:maxGenres], ", ")
}

// GetRuntimeDisplay returns runtime in human-readable format
func (m *Media) GetRuntimeDisplay() string {
	if m.Runtime <= 0 {
		return ""
	}
	hours := m.Runtime / 60
	mins := m.Runtime % 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}
