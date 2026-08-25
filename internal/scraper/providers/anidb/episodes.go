package anidb

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

// apiEpisode mirrors one entry of /api/frontend/anime/<id>/episodes.
type apiEpisode struct {
	ID      int    `json:"id"`
	Number  int    `json:"number"`
	Number2 *int   `json:"number2"`
	Filler  bool   `json:"filler"`
	Title   string `json:"title"`
}

type apiEpisodesResponse struct {
	Episodes []apiEpisode `json:"episodes"`
}

// GetAnimeEpisodes lists the episodes of an anime. animeURL is the permalink
// produced by SearchAnime; a bare numeric id is also accepted.
func (c *AniDBClient) GetAnimeEpisodes(ctx context.Context, animeURL string) ([]models.Episode, error) {
	id, err := AnimeID(animeURL)
	if err != nil {
		return nil, err
	}
	apiURL := fmt.Sprintf("%s/api/frontend/anime/%s/episodes", c.baseURL, id)
	util.Debug("AniDB episodes", "anime", id, "url", apiURL)

	var payload apiEpisodesResponse
	if err := c.getJSON(ctx, apiURL, "episodes", &payload); err != nil {
		return nil, err
	}
	if len(payload.Episodes) == 0 {
		return nil, netx.NewParserError(sourceLabel, "episodes",
			fmt.Sprintf("no episodes listed for anime %s", id), nil)
	}

	episodes := make([]models.Episode, 0, len(payload.Episodes))
	for _, e := range payload.Episodes {
		if e.ID == 0 {
			continue // an entry without an id cannot be streamed
		}
		num := e.Number
		episodes = append(episodes, models.Episode{
			Number:   strconv.Itoa(num),
			Num:      num,
			URL:      c.episodeURL(e.ID),
			DataID:   strconv.Itoa(e.ID),
			IsFiller: e.Filler,
			Title:    models.TitleDetails{English: e.Title},
		})
	}
	// Every entry was dropped for want of an id: the API answered, but nothing
	// here is streamable. Returning an empty list with a nil error would show
	// the user an empty episode picker and look like the anime simply has no
	// episodes. (Caught by the fixture mutation in meta_test.go.)
	if len(episodes) == 0 {
		return nil, netx.NewParserError(sourceLabel, "episodes",
			fmt.Sprintf("anime %s listed %d episodes but none carried a usable id", id, len(payload.Episodes)), nil)
	}

	sort.SliceStable(episodes, func(i, j int) bool { return episodes[i].Num < episodes[j].Num })
	return episodes, nil
}
