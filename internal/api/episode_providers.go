// Package api provides episode data fetching from multiple sources with fallback support.
// This enables robust episode information retrieval even when primary APIs are unavailable.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// EpisodeDataProvider defines an interface for fetching episode data from various sources
type EpisodeDataProvider interface {
	Name() string
	FetchEpisodeData(animeID int, episodeNo int, anime *models.Anime) error
}

// JikanProvider fetches episode data from Jikan (MyAnimeList) API
type JikanProvider struct{}

func (p *JikanProvider) Name() string {
	return "Jikan (MyAnimeList)"
}

func (p *JikanProvider) FetchEpisodeData(animeID int, episodeNo int, anime *models.Anime) error {
	if animeID <= 0 {
		return fmt.Errorf("invalid anime ID: %d", animeID)
	}

	url := fmt.Sprintf("https://api.jikan.moe/v4/anime/%d/episodes/%d", animeID, episodeNo)

	response, err := makeGetRequest(url, nil)
	if err != nil {
		return fmt.Errorf("jikan API request failed: %w", err)
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid response structure from Jikan")
	}

	populateEpisodeFromMap(anime, data)
	return nil
}

// AniListProvider fetches episode data from AniList GraphQL API
type AniListProvider struct{}

func (p *AniListProvider) Name() string {
	return "AniList"
}

func (p *AniListProvider) FetchEpisodeData(animeID int, episodeNo int, anime *models.Anime) error {
	// AniList uses its own ID system, so we need to use the AnilistID from the anime
	anilistID := anime.AnilistID
	if anilistID <= 0 {
		// Try to find by MAL ID if we have it
		if animeID > 0 {
			var err error
			anilistID, err = getAniListIDFromMAL(animeID)
			if err != nil {
				return fmt.Errorf("could not find AniList ID: %w", err)
			}
		} else {
			return fmt.Errorf("no valid AniList or MAL ID available")
		}
	}

	// AniList doesn't have per-episode endpoint, but we can get anime metadata
	// Note: We don't declare unused variables in GraphQL to avoid 400 errors
	query := `query ($id: Int) {
		Media(id: $id, type: ANIME) {
			id
			title { romaji english native }
			episodes
			duration
			description
			streamingEpisodes {
				title
				thumbnail
				url
			}
		}
	}`

	jsonData, err := json.Marshal(map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"id": anilistID,
		},
	})
	if err != nil {
		return fmt.Errorf("JSON marshal failed: %w", err)
	}

	resp, err := httpPost("https://graphql.anilist.co", jsonData)
	if err != nil {
		return fmt.Errorf("AniList request failed: %w", err)
	}
	defer safeClose(resp.Body, "AniList episode response")

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AniList returned: %s", resp.Status)
	}

	var result struct {
		Data struct {
			Media struct {
				ID    int `json:"id"`
				Title struct {
					Romaji  string `json:"romaji"`
					English string `json:"english"`
					Native  string `json:"native"`
				} `json:"title"`
				Episodes          int    `json:"episodes"`
				Duration          int    `json:"duration"`
				Description       string `json:"description"`
				StreamingEpisodes []struct {
					Title     string `json:"title"`
					Thumbnail string `json:"thumbnail"`
					URL       string `json:"url"`
				} `json:"streamingEpisodes"`
			} `json:"Media"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("JSON decode failed: %w", err)
	}

	if result.Data.Media.ID == 0 {
		return fmt.Errorf("anime not found on AniList")
	}

	// Populate episode data from AniList response
	if len(anime.Episodes) == 0 {
		anime.Episodes = make([]models.Episode, 1)
	}

	ep := &anime.Episodes[0]
	ep.Title.Romaji = result.Data.Media.Title.Romaji
	ep.Title.English = result.Data.Media.Title.English
	ep.Title.Japanese = result.Data.Media.Title.Native
	ep.Duration = result.Data.Media.Duration * 60 // AniList returns duration in minutes

	// Try to get episode-specific title from streaming episodes
	if episodeNo > 0 && episodeNo <= len(result.Data.Media.StreamingEpisodes) {
		streamingEp := result.Data.Media.StreamingEpisodes[episodeNo-1]
		if streamingEp.Title != "" {
			ep.Title.English = streamingEp.Title
		}
	}

	return nil
}

// KitsuProvider fetches episode data from Kitsu API
type KitsuProvider struct{}

func (p *KitsuProvider) Name() string {
	return "Kitsu"
}

func (p *KitsuProvider) FetchEpisodeData(animeID int, episodeNo int, anime *models.Anime) error {
	if animeID <= 0 {
		// Try to search by anime name if no MAL ID
		return p.fetchByAnimeName(anime, episodeNo)
	}

	// First, we need to find the Kitsu anime ID using the MAL ID
	kitsuAnimeID, err := getKitsuAnimeID(animeID)
	if err != nil {
		// Fallback: try searching by anime name
		return p.fetchByAnimeName(anime, episodeNo)
	}

	return p.fetchEpisodeByKitsuID(kitsuAnimeID, episodeNo, anime)
}

func (p *KitsuProvider) fetchByAnimeName(anime *models.Anime, episodeNo int) error {
	// Clean the anime name for search (remove source tags and dub indicators)
	searchName := CleanTitle(anime.Name)

	searchURL := fmt.Sprintf("https://kitsu.io/api/edge/anime?filter[text]=%s&page[limit]=1",
		strings.ReplaceAll(searchName, " ", "%20"))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create search request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.api+json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("kitsu search failed: %w", err)
	}
	defer safeClose(resp.Body, "Kitsu search response")

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("kitsu search returned: %s", resp.Status)
	}

	var searchResult struct {
		Data []struct {
			ID         string `json:"id"`
			Attributes struct {
				CanonicalTitle string `json:"canonicalTitle"`
				EpisodeCount   int    `json:"episodeCount"`
				EpisodeLength  int    `json:"episodeLength"`
				Synopsis       string `json:"synopsis"`
				Titles         struct {
					En   string `json:"en"`
					EnJp string `json:"en_jp"`
					JaJp string `json:"ja_jp"`
				} `json:"titles"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return fmt.Errorf("kitsu search decode failed: %w", err)
	}

	if len(searchResult.Data) == 0 {
		return fmt.Errorf("anime not found on Kitsu: %s", searchName)
	}

	// Use the first result and populate basic data
	kitsuAnime := searchResult.Data[0]

	if len(anime.Episodes) == 0 {
		anime.Episodes = make([]models.Episode, 1)
	}

	ep := &anime.Episodes[0]
	ep.Title.English = kitsuAnime.Attributes.CanonicalTitle
	if kitsuAnime.Attributes.Titles.En != "" {
		ep.Title.English = kitsuAnime.Attributes.Titles.En
	}
	ep.Title.Romaji = kitsuAnime.Attributes.Titles.EnJp
	ep.Title.Japanese = kitsuAnime.Attributes.Titles.JaJp
	ep.Duration = kitsuAnime.Attributes.EpisodeLength * 60 // Convert minutes to seconds

	// Try to get episode-specific data if we have the Kitsu ID
	if kitsuAnime.ID != "" {
		_ = p.fetchEpisodeByKitsuID(kitsuAnime.ID, episodeNo, anime)
	}

	return nil
}

func (p *KitsuProvider) fetchEpisodeByKitsuID(kitsuAnimeID string, episodeNo int, anime *models.Anime) error {
	// Fetch episodes for this anime
	url := fmt.Sprintf("https://kitsu.io/api/edge/anime/%s/episodes?filter[number]=%d", kitsuAnimeID, episodeNo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.api+json")
	req.Header.Set("Content-Type", "application/vnd.api+json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("kitsu request failed: %w", err)
	}
	defer safeClose(resp.Body, "Kitsu episode response")

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("kitsu returned: %s", resp.Status)
	}

	var result struct {
		Data []struct {
			ID         string `json:"id"`
			Attributes struct {
				Synopsis string `json:"synopsis"`
				Titles   struct {
					EnJp string `json:"en_jp"`
					JaJp string `json:"ja_jp"`
				} `json:"titles"`
				CanonicalTitle string `json:"canonicalTitle"`
				Number         int    `json:"number"`
				Length         int    `json:"length"`
				Airdate        string `json:"airdate"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("JSON decode failed: %w", err)
	}

	if len(result.Data) == 0 {
		return fmt.Errorf("episode %d not found on Kitsu", episodeNo)
	}

	// Populate episode data from Kitsu response
	if len(anime.Episodes) == 0 {
		anime.Episodes = make([]models.Episode, 1)
	}

	ep := &anime.Episodes[0]
	kitsuEp := result.Data[0].Attributes

	ep.Title.English = kitsuEp.CanonicalTitle
	if ep.Title.English == "" {
		ep.Title.English = kitsuEp.Titles.EnJp
	}
	ep.Title.Japanese = kitsuEp.Titles.JaJp
	ep.Synopsis = kitsuEp.Synopsis
	ep.Duration = kitsuEp.Length * 60 // Kitsu returns length in minutes
	ep.Aired = kitsuEp.Airdate

	return nil
}

// getAniListIDFromMAL converts a MyAnimeList ID to an AniList ID
func getAniListIDFromMAL(malID int) (int, error) {
	query := `query ($malId: Int) {
		Media(idMal: $malId, type: ANIME) {
			id
		}
	}`

	jsonData, err := json.Marshal(map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"malId": malID,
		},
	})
	if err != nil {
		return 0, err
	}

	resp, err := httpPost("https://graphql.anilist.co", jsonData)
	if err != nil {
		return 0, err
	}
	defer safeClose(resp.Body, "AniList MAL lookup response")

	var result struct {
		Data struct {
			Media struct {
				ID int `json:"id"`
			} `json:"Media"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	if result.Data.Media.ID == 0 {
		return 0, fmt.Errorf("no AniList entry found for MAL ID %d", malID)
	}

	return result.Data.Media.ID, nil
}

// getKitsuAnimeID finds the Kitsu anime ID using the MyAnimeList ID
func getKitsuAnimeID(malID int) (string, error) {
	url := fmt.Sprintf("https://kitsu.io/api/edge/mappings?filter[externalSite]=myanimelist/anime&filter[externalId]=%d", malID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.api+json")
	req.Header.Set("Content-Type", "application/vnd.api+json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer safeClose(resp.Body, "Kitsu mapping response")

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("kitsu mapping lookup returned: %s", resp.Status)
	}

	var mappingResult struct {
		Data []struct {
			ID            string `json:"id"`
			Relationships struct {
				Item struct {
					Data struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"data"`
				} `json:"item"`
			} `json:"relationships"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&mappingResult); err != nil {
		return "", err
	}

	if len(mappingResult.Data) == 0 {
		return "", fmt.Errorf("no Kitsu mapping found for MAL ID %d", malID)
	}

	return mappingResult.Data[0].Relationships.Item.Data.ID, nil
}

// populateEpisodeFromMap populates episode data from a map (used by Jikan)
func populateEpisodeFromMap(anime *models.Anime, data map[string]interface{}) {
	if len(anime.Episodes) == 0 {
		anime.Episodes = make([]models.Episode, 1)
	}

	ep := &anime.Episodes[0]
	ep.Title.Romaji = getStringValue(data, "title_romanji")
	ep.Title.English = getStringValue(data, "title")
	ep.Title.Japanese = getStringValue(data, "title_japanese")
	ep.Aired = getStringValue(data, "aired")
	ep.Duration = getIntValue(data, "duration")
	ep.IsFiller = getBoolValue(data, "filler")
	ep.IsRecap = getBoolValue(data, "recap")
	ep.Synopsis = getStringValue(data, "synopsis")
}

// defaultProviders returns the ordered list of episode data providers
func defaultProviders() []EpisodeDataProvider {
	return []EpisodeDataProvider{
		&JikanProvider{},
		&AniListProvider{},
		&KitsuProvider{},
	}
}

// GetEpisodeDataWithFallback fetches episode data trying multiple providers
func GetEpisodeDataWithFallback(animeID int, episodeNo int, anime *models.Anime) error {
	providers := defaultProviders()
	var lastErr error
	var errors []string

	for _, provider := range providers {
		util.Debugf("Trying episode data provider: %s", provider.Name())

		err := provider.FetchEpisodeData(animeID, episodeNo, anime)
		if err == nil {
			util.Debugf("Successfully fetched episode data from %s", provider.Name())
			return nil
		}

		lastErr = err
		errMsg := fmt.Sprintf("%s: %v", provider.Name(), err)
		errors = append(errors, errMsg)
		util.Debugf("Provider %s failed: %v", provider.Name(), err)

		// Add a small delay to avoid rate limiting between providers
		time.Sleep(100 * time.Millisecond)
	}

	// All providers failed
	return fmt.Errorf("all episode data providers failed: %s (last error: %w)",
		strings.Join(errors, "; "), lastErr)
}
