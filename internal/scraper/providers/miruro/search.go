package miruro

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// searchItem is one result of the search page's "items" list. The site
// builds it from AniList, which is why the AniList and MAL ids come with it.
type searchItem struct {
	ID    string `json:"id"`
	Title struct {
		Romaji  string `json:"romaji"`
		English string `json:"english"`
		Native  string `json:"native"`
	} `json:"title"`
	Format        string `json:"format"` // TV, MOVIE, OVA, ONA, SPECIAL, TV_SHORT
	Status        string `json:"status"` // FINISHED, RELEASING, NOT_YET_RELEASED, ...
	SeasonYear    *int   `json:"season_year"`
	EpisodeCounts struct {
		Sub *int `json:"sub"`
		Dub *int `json:"dub"`
	} `json:"episode_counts"`
	ExternalIDs map[string][]string `json:"external_ids"`
	CoverURL    string              `json:"cover_url"`
}

type searchPage struct {
	Items []searchItem `json:"items"`
}

// SearchAnime searches miruro.cx and returns the first page of results.
func (c *MiruroClient) SearchAnime(ctx context.Context, query string) ([]*models.Anime, error) {
	if strings.TrimSpace(query) == "" {
		return nil, netx.NewParserError(sourceLabel, "search", "empty query", nil)
	}
	util.Debug("Miruro search", "query", query)

	resp, err := c.getData(ctx, "/search", url.Values{"q": {query}}, "search")
	if err != nil {
		return nil, err
	}
	node, ok := resp.nodeWith("items")
	if !ok {
		return nil, netx.NewParserError(sourceLabel, "search", "the page data has no result list (layout changed?)", nil)
	}
	var page searchPage
	if err := decodeInto(node, &page); err != nil {
		return nil, netx.NewParserError(sourceLabel, "search", "unreadable result list", err)
	}

	results := c.toResults(page.Items)
	if len(results) == 0 {
		util.Debug("Miruro search returned no results", "query", query)
	}
	return results, nil
}

// toResults turns the result list into models.
//
// A title AniList marks NOT_YET_RELEASED is dropped: its watch page lists no
// episodes and no sources, so all it could do is lead to "no episodes yet".
// The per-audio episode counts are not used to filter: they lag the sources
// (a title counted with no subtitled episodes still served subtitled ones).
func (c *MiruroClient) toResults(items []searchItem) []*models.Anime {
	seen := make(map[string]struct{}, len(items))
	out := make([]*models.Anime, 0, len(items))
	for i := range items {
		it := &items[i]
		if it.ID == "" || !idRe.MatchString(it.ID) {
			continue
		}
		if _, dup := seen[it.ID]; dup {
			continue
		}
		if strings.EqualFold(it.Status, "NOT_YET_RELEASED") {
			continue
		}
		name := firstNonEmpty(it.Title.English, it.Title.Romaji, it.Title.Native)
		if name == "" {
			continue
		}
		seen[it.ID] = struct{}{}

		anime := &models.Anime{
			Name:      name,
			URL:       c.animeURL(it.ID),
			ImageURL:  strings.TrimSpace(it.CoverURL),
			Source:    sourceLabel,
			MediaType: models.MediaTypeAnime,
			Audio:     models.AudioLabel(positive(it.EpisodeCounts.Dub), positive(it.EpisodeCounts.Sub)),
			AnilistID: firstID(it.ExternalIDs["anilist"]),
			MalID:     firstID(it.ExternalIDs["mal"]),
		}
		if it.SeasonYear != nil && *it.SeasonYear > 0 {
			anime.Year = strconv.Itoa(*it.SeasonYear)
		}
		anime.Details.Title.English = strings.TrimSpace(it.Title.English)
		anime.Details.Title.Romaji = strings.TrimSpace(it.Title.Romaji)
		out = append(out, anime)
	}
	return out
}

func positive(n *int) bool { return n != nil && *n > 0 }

// firstID reads the first id of an external_ids list ("154587"), 0 when
// there is none.
func firstID(ids []string) int {
	for _, s := range ids {
		if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil && n > 0 {
			return n
		}
	}
	return 0
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}
