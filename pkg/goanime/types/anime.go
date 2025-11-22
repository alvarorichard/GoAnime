// Package types provides public type definitions for the goanime library
package types

import (
	"github.com/alvarorichard/Goanime/internal/models"
)

// Anime represents an anime series with its metadata
type Anime struct {
	// Name is the title of the anime
	Name string
	// URL is the source-specific URL for this anime
	URL string
	// ImageURL is the cover/poster image URL
	ImageURL string
	// Episodes contains all available episodes (may be empty until GetAnimeEpisodes is called)
	Episodes []*Episode
	// AnilistID is the AniList database ID (if available)
	AnilistID int
	// MalID is the MyAnimeList database ID (if available)
	MalID int
	// Source identifies where this anime came from (e.g., "AllAnime", "AnimeFire")
	Source string
	// Details contains additional metadata from AniList
	Details *AniListDetails
}

// Episode represents a single episode of an anime
type Episode struct {
	// Number is the episode number as a string (e.g., "1", "1.5", "OVA")
	Number string
	// Num is the episode number as an integer
	Num int
	// URL is the source-specific URL for this episode
	URL string
	// Title contains the episode title in different languages
	Title *TitleDetails
	// Aired is the air date
	Aired string
	// Duration in seconds
	Duration int
	// IsFiller indicates if this is a filler episode
	IsFiller bool
	// IsRecap indicates if this is a recap episode
	IsRecap bool
	// Synopsis is the episode description
	Synopsis string
	// SkipTimes contains timestamps for skipping intros/outros
	SkipTimes *SkipTimes
}

// TitleDetails contains anime/episode titles in multiple languages
type TitleDetails struct {
	Romaji   string
	English  string
	Japanese string
}

// SkipTimes contains timestamps for skipping openings and endings
type SkipTimes struct {
	Op *SkipTime
	Ed *SkipTime
}

// SkipTime represents a time range to skip
type SkipTime struct {
	Start int
	End   int
}

// AniListDetails contains extended metadata from AniList
type AniListDetails struct {
	ID           int
	IDMal        int
	Title        *Title
	Description  string
	Genres       []string
	AverageScore int
	Episodes     int
	Status       string
	CoverImage   *CoverImages
}

// Title contains anime title in multiple languages
type Title struct {
	Romaji  string
	English string
}

// CoverImages contains URLs for different cover image sizes
type CoverImages struct {
	Large  string
	Medium string
}

// FromInternalAnime converts internal anime model to public type
func FromInternalAnime(internal *models.Anime) *Anime {
	if internal == nil {
		return nil
	}

	anime := &Anime{
		Name:      internal.Name,
		URL:       internal.URL,
		ImageURL:  internal.ImageURL,
		AnilistID: internal.AnilistID,
		MalID:     internal.MalID,
		Source:    internal.Source,
	}

	// Convert episodes
	if len(internal.Episodes) > 0 {
		anime.Episodes = make([]*Episode, len(internal.Episodes))
		for i, ep := range internal.Episodes {
			anime.Episodes[i] = FromInternalEpisode(&ep)
		}
	}

	// Convert details
	anime.Details = &AniListDetails{
		ID:           internal.Details.ID,
		IDMal:        internal.Details.IDMal,
		Description:  internal.Details.Description,
		Genres:       internal.Details.Genres,
		AverageScore: internal.Details.AverageScore,
		Episodes:     internal.Details.Episodes,
		Status:       internal.Details.Status,
	}

	if internal.Details.Title.Romaji != "" || internal.Details.Title.English != "" {
		anime.Details.Title = &Title{
			Romaji:  internal.Details.Title.Romaji,
			English: internal.Details.Title.English,
		}
	}

	if internal.Details.CoverImage.Large != "" || internal.Details.CoverImage.Medium != "" {
		anime.Details.CoverImage = &CoverImages{
			Large:  internal.Details.CoverImage.Large,
			Medium: internal.Details.CoverImage.Medium,
		}
	}

	return anime
}

// FromInternalAnimeList converts a slice of internal anime models to public types
func FromInternalAnimeList(internal []*models.Anime) []*Anime {
	result := make([]*Anime, len(internal))
	for i, a := range internal {
		result[i] = FromInternalAnime(a)
	}
	return result
}

// FromInternalEpisode converts internal episode model to public type
func FromInternalEpisode(internal *models.Episode) *Episode {
	if internal == nil {
		return nil
	}

	episode := &Episode{
		Number:   internal.Number,
		Num:      internal.Num,
		URL:      internal.URL,
		Aired:    internal.Aired,
		Duration: internal.Duration,
		IsFiller: internal.IsFiller,
		IsRecap:  internal.IsRecap,
		Synopsis: internal.Synopsis,
	}

	if internal.Title.Romaji != "" || internal.Title.English != "" || internal.Title.Japanese != "" {
		episode.Title = &TitleDetails{
			Romaji:   internal.Title.Romaji,
			English:  internal.Title.English,
			Japanese: internal.Title.Japanese,
		}
	}

	if internal.SkipTimes.Op.Start > 0 || internal.SkipTimes.Op.End > 0 ||
		internal.SkipTimes.Ed.Start > 0 || internal.SkipTimes.Ed.End > 0 {
		episode.SkipTimes = &SkipTimes{
			Op: &SkipTime{
				Start: internal.SkipTimes.Op.Start,
				End:   internal.SkipTimes.Op.End,
			},
			Ed: &SkipTime{
				Start: internal.SkipTimes.Ed.Start,
				End:   internal.SkipTimes.Ed.End,
			},
		}
	}

	return episode
}

// FromInternalEpisodeList converts a slice of internal episode models to public types
func FromInternalEpisodeList(internal []models.Episode) []*Episode {
	result := make([]*Episode, len(internal))
	for i := range internal {
		result[i] = FromInternalEpisode(&internal[i])
	}
	return result
}
