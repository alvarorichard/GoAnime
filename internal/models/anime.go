// Package models contains anime-specific data structures
package models

// Anime represents an anime series with its metadata
// For backwards compatibility, this is an alias to Media
type Anime = Media

// AniListResponse represents the response from AniList GraphQL API
type AniListResponse struct {
	Data struct {
		Media AniListDetails `json:"Media"`
	} `json:"data"`
}

// AniListDetails contains detailed anime information from AniList
type AniListDetails struct {
	ID           int         `json:"id"`
	IDMal        int         `json:"idMal"`
	Title        Title       `json:"title"`
	Description  string      `json:"description"`
	Genres       []string    `json:"genres"`
	AverageScore int         `json:"averageScore"`
	Episodes     int         `json:"episodes"`
	Status       string      `json:"status"`
	CoverImage   CoverImages `json:"coverImage"`
	Synonyms     []string    `json:"synonyms"`
}

// Title contains anime title in multiple languages
type Title struct {
	Romaji  string `json:"romaji"`
	English string `json:"english"`
	Native  string `json:"native"`
}

// CoverImages contains anime cover image URLs
type CoverImages struct {
	Large  string `json:"large"`
	Medium string `json:"medium"`
}
