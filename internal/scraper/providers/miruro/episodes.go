package miruro

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// watchEpisode is one entry of the watch page's episode list.
type watchEpisode struct {
	Number      *float64 `json:"number"`
	Title       string   `json:"title"`
	AirDate     string   `json:"airDate"`
	Description string   `json:"description"`
	Filler      bool     `json:"filler"`
}

// watchPage is the part of a watch page's data this scraper reads. Both
// fields arrive as promises (see devalue.go); a title that has not aired
// lists no episodes and has null sources.
type watchPage struct {
	EpisodesData *struct {
		Episodes []watchEpisode `json:"episodes"`
	} `json:"episodesData"`
	SourcesData *sourcesData `json:"sourcesData"`
}

// fetchWatchPage reads the watch page of episode ep. The slug may be empty:
// the site redirects to the canonical one, and the returned path carries it.
func (c *Client) fetchWatchPage(ctx context.Context, id, slug, ep, layer string) (watchPage, string, error) {
	resp, err := c.getData(ctx, watchPath(id, slug), url.Values{"ep": {ep}}, layer)
	if err != nil {
		return watchPage{}, "", err
	}
	node, ok := resp.nodeWith("episodesData")
	if !ok {
		return watchPage{}, "", netx.NewParserError(sourceLabel, layer, "the page data has no episode list (layout changed?)", nil)
	}
	var page watchPage
	if err := decodeInto(node, &page); err != nil {
		return watchPage{}, "", netx.NewParserError(sourceLabel, layer, "unreadable watch page", err)
	}
	return page, resp.Path, nil
}

// GetAnimeEpisodes lists the episodes of an anime. animeURL is the /info URL
// SearchAnime produced; a /watch URL or a bare id is also accepted.
//
// The list comes from the watch page of the first episode, which carries the
// whole list (One Piece's 1180 entries included) and, through its redirect,
// the slug the episode URLs are built with.
func (c *Client) GetAnimeEpisodes(ctx context.Context, animeURL string) ([]models.Episode, error) {
	id, err := AnimeID(animeURL)
	if err != nil {
		return nil, err
	}
	util.Debug("Miruro episodes", "anime", id)

	page, path, err := c.fetchWatchPage(ctx, id, slugFromPath(animeURL), "1", "episodes")
	if err != nil {
		return nil, err
	}
	if page.EpisodesData == nil {
		return nil, netx.NewParserError(sourceLabel, "episodes", "the episode list is missing (layout changed?)", nil)
	}
	listed := page.EpisodesData.Episodes
	if len(listed) == 0 {
		// The site's own answer for a title that has not aired: an empty
		// list, next to null sources.
		return nil, fmt.Errorf("%s anime %s: %w", sourceLabel, id, netx.ErrNoEpisodes)
	}

	slug := slugFromPath(path)
	seen := make(map[string]struct{}, len(listed))
	episodes := make([]models.Episode, 0, len(listed))
	for _, e := range listed {
		if e.Number == nil || *e.Number < 0 {
			continue // an entry without a number cannot be asked for
		}
		number := strconv.FormatFloat(*e.Number, 'f', -1, 64)
		if _, dup := seen[number]; dup {
			continue
		}
		seen[number] = struct{}{}
		episodes = append(episodes, models.Episode{
			Number:   number,
			Num:      int(*e.Number),
			URL:      c.episodeURL(id, slug, number),
			Title:    models.TitleDetails{English: strings.TrimSpace(e.Title)},
			Aired:    strings.TrimSpace(e.AirDate),
			Synopsis: strings.TrimSpace(e.Description),
			IsFiller: e.Filler,
		})
	}
	if len(episodes) == 0 {
		return nil, netx.NewParserError(sourceLabel, "episodes",
			"anime "+id+" listed "+strconv.Itoa(len(listed))+" episodes but none carried a number", nil)
	}

	sort.SliceStable(episodes, func(i, j int) bool {
		a, _ := strconv.ParseFloat(episodes[i].Number, 64)
		b, _ := strconv.ParseFloat(episodes[j].Number, 64)
		return a < b
	})
	return episodes, nil
}
