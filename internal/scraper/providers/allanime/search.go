// Package scraper provides web scraping functionality for anime sources
package allanime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// SearchAnime searches for anime using AllAnime API (based on Curd implementation)
func (c *AllAnimeClient) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	// Use the exact same GraphQL query as Curd
	searchGql := `query($search: SearchInput, $limit: Int, $page: Int, $translationType: VaildTranslationTypeEnumType, $countryOrigin: VaildCountryOriginEnumType) {
		shows(search: $search, limit: $limit, page: $page, translationType: $translationType, countryOrigin: $countryOrigin) {
			edges {
				_id
				name
				englishName
				availableEpisodes
				__typename
			}
		}
	}`

	// Prepare the GraphQL variables exactly like Curd
	variables := map[string]any{
		"search": map[string]any{
			"allowAdult":   false,
			"allowUnknown": false,
			"query":        query,
		},
		"limit":           40,
		"page":            1,
		"translationType": "sub",
		"countryOrigin":   "ALL",
	}

	// Marshal the variables to JSON
	variablesJSON, err := json.Marshal(variables)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal variables: %w", err)
	}

	// Build the POST request body with variables and query
	reqBody, err := json.Marshal(map[string]any{
		"variables": json.RawMessage(variablesJSON),
		"query":     searchGql,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.apiBase, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Referer", c.referer)

	resp, err := c.client.Do(req) // #nosec G704
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := netx.CheckHTTPStatus(resp, "AllAnime search"); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if err := netx.CheckHTMLResponse(resp, body, "AllAnime search"); err != nil {
		return nil, err
	}

	// Parse using a simple structure like Curd
	var response struct {
		Data struct {
			Shows struct {
				Edges []struct {
					ID                string `json:"_id"`
					Name              string `json:"name"`
					EnglishName       string `json:"englishName"`
					AvailableEpisodes any    `json:"availableEpisodes"`
				} `json:"edges"`
			} `json:"shows"`
		} `json:"data"`
	}

	if err := jsonx.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Collect results with episode counts for sorting
	type searchResult struct {
		anime   *models.Anime
		epCount int
	}
	var results []searchResult

	for _, edge := range response.Data.Shows.Edges {
		var episodesStr string
		var epCount int
		if episodes, ok := edge.AvailableEpisodes.(map[string]any); ok {
			if subEpisodes, ok := episodes["sub"].(float64); ok {
				epCount = int(subEpisodes)
				episodesStr = fmt.Sprintf("(%d episodes)", epCount)
			} else {
				episodesStr = "(Unknown episodes)"
			}
		}

		// Use English name if available, otherwise use default name
		displayName := edge.Name
		if edge.EnglishName != "" {
			displayName = edge.EnglishName
		}

		anime := &models.Anime{
			Name: fmt.Sprintf("%s %s", displayName, episodesStr),
			URL:  edge.ID, // For AllAnime, the "URL" is actually the anime ID
		}
		results = append(results, searchResult{anime: anime, epCount: epCount})
	}

	// Sort by episode count descending so the main series (most episodes) appears first
	sort.Slice(results, func(i, j int) bool {
		return results[i].epCount > results[j].epCount
	})

	animes := make([]*models.Anime, len(results))
	for i, r := range results {
		animes[i] = r.anime
	}

	return animes, nil
}
