package hianime

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

// GetAnimeEpisodes lists the episodes of an anime. animeURL is the permalink
// produced by SearchAnime; a bare numeric id is also accepted.
func (c *HiAnimeClient) GetAnimeEpisodes(ctx context.Context, animeURL string) ([]models.Episode, error) {
	id, err := AnimeID(animeURL)
	if err != nil {
		return nil, err
	}
	slug := animeSlug(animeURL)
	listURL := c.restURL("episode/list/" + id)
	util.Debug("HiAnime episodes", "anime", id, "url", listURL)

	fragment, err := c.getFragment(ctx, listURL, "episodes", c.baseURL+"/")
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(fragment))
	if err != nil {
		return nil, netx.NewParserError(sourceLabel, "episodes", "failed to parse episode list", err)
	}

	var episodes []models.Episode
	var listed int
	doc.Find("a.ep-item").Each(func(_ int, s *goquery.Selection) {
		listed++
		epID, convErr := strconv.Atoi(strings.TrimSpace(s.AttrOr("data-id", "")))
		if convErr != nil || epID == 0 {
			return // an entry without an id cannot be streamed
		}
		num, convErr := strconv.Atoi(strings.TrimSpace(s.AttrOr("data-number", "")))
		if convErr != nil {
			return
		}
		episodes = append(episodes, models.Episode{
			Number:   strconv.Itoa(num),
			Num:      num,
			URL:      c.episodeURL(slug, id, epID),
			DataID:   strconv.Itoa(epID),
			IsFiller: s.HasClass("ssl-item-filler"),
			Title:    models.TitleDetails{English: episodeTitle(s)},
		})
	})

	if listed == 0 {
		return nil, netx.NewParserError(sourceLabel, "episodes",
			"the episode list carried no entries (layout changed?)", nil)
	}
	// Every entry was dropped for want of an id or a number: the endpoint
	// answered, but nothing here is streamable. Returning an empty list with a
	// nil error would show the user an empty episode picker and look like the
	// anime simply has no episodes.
	if len(episodes) == 0 {
		return nil, netx.NewParserError(sourceLabel, "episodes",
			"anime "+id+" listed "+strconv.Itoa(listed)+" episodes but none carried a usable id", nil)
	}

	sort.SliceStable(episodes, func(i, j int) bool { return episodes[i].Num < episodes[j].Num })
	return episodes, nil
}

// episodeTitle reads an entry's human title, which the template renders both as
// an attribute and as padded text inside .ep-name.
func episodeTitle(s *goquery.Selection) string {
	if t := strings.TrimSpace(s.AttrOr("title", "")); t != "" {
		return t
	}
	name := s.Find(".ep-name").First()
	if t := strings.TrimSpace(name.AttrOr("title", "")); t != "" {
		return t
	}
	return strings.TrimSpace(name.Text())
}

// animeSlug recovers the slug half of a permalink so episode URLs stay
// human-openable. A bare id has none, and the site redirects a wrong slug to
// the right one anyway, so an empty slug is not worth failing over.
func animeSlug(animeURL string) string {
	if m := animeHrefRe.FindStringSubmatch(strings.TrimSpace(animeURL)); m != nil {
		return m[1]
	}
	return "anime"
}
