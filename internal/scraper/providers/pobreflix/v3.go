package pobreflix

import (
	"context"
	"encoding/json"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// The site was rebuilt again on 2026-11-14 and moved to pobreflixtvfree.cv. The
// old host answers every path with a 301 to that home page, so a title page URL
// from the old layout ends on the home page. The new site is server-rendered,
// with no embedded JSON and no #view mount:
//
//	/filmes/online/<slug>-<audio>-<tmdb>      <div class="player" data-src="https://redeflixapi.click/filme/<tmdb>">
//	/series/online/<slug>-<audio>-<tmdb>      <section class="seasons"><details><summary>Temporada N ...
//	                                          <ol class="episodes"><li><a href="/episodios/online/...">
//	/episodios/online/<slug>-<S>x<E>-<audio>-<tmdb>   <div class="player" data-src=".../serie/<tmdb>/<S>/<E>">
//
// The id that ends every slug is the TMDB id. The player is the RedeFlixApi
// page, which ResolveRedeflix opens without a solver. Search is
// /api/busca?q=<text>.

// v3EpisodeRe reads the slot an episode page's slug names: "dexter-1x2-dublado-1405".
var v3EpisodeRe = regexp.MustCompile(`-(\d+)x(\d+)-`)

// v3SeasonRe reads a season summary: "Temporada 3 12 episódios".
var v3SeasonRe = regexp.MustCompile(`(?i)temporada\s+(\d+)`)

// IsEpisodePage reports whether raw is the page of one episode, in the layout
// the site has now or the one it had before.
func IsEpisodePage(raw string) bool {
	path := pathOfURL(raw)
	return strings.Contains(path, "/episodios/online/") || strings.Contains(path, "/assistir/")
}

// v3Servers lists the players a page mounts: the iframe each .player box loads
// when the visitor presses play.
func v3Servers(doc *goquery.Document, audio Audio) []Server {
	var out []Server
	doc.Find("div.player[data-src]").Each(func(_ int, box *goquery.Selection) {
		src := strings.TrimSpace(attrOr(box, "data-src"))
		if src == "" {
			return
		}
		out = append(out, Server{URL: src, Name: "Player Principal", Audio: audio})
	})
	return out
}

// parseV3Title reads a title page of the current site. ok=false covers a page
// that is not a record of it, or one whose slug names no TMDB id.
func parseV3Title(doc *goquery.Document, pageURL string) (Title, bool) {
	kind, slug, audio, tmdb := titleRecordFromPath(pathOfURL(pageURL))
	if slug == "" || tmdb <= 0 || IsEpisodePage(pageURL) {
		return Title{}, false
	}
	servers := v3Servers(doc, audio)
	seasons := v3Seasons(doc, audio, pageURL)
	if len(servers) == 0 && len(seasons) == 0 {
		return Title{}, false
	}
	name := displayTitle(doc.Find("h1.page-title").First().Text())
	if name == "" {
		name = slugToTitle(slug)
	}
	t := Title{
		Kind:       kind,
		VideoID:    tmdb,
		TMDBID:     tmdb,
		InternalID: tmdb,
		Audio:      audio,
		Name:       name,
		Poster:     strings.TrimSpace(doc.Find(`meta[property="og:image"]`).First().AttrOr("content", "")),
		Slug:       slug,
		Servers:    servers,
		Seasons:    seasons,
	}
	if kind == KindSeries {
		t.Servers = nil
	}
	return t, true
}

// v3Seasons reads the season blocks of a series page.
func v3Seasons(doc *goquery.Document, audio Audio, pageURL string) []Season {
	var out []Season
	doc.Find("section.seasons details").Each(func(_ int, block *goquery.Selection) {
		m := v3SeasonRe.FindStringSubmatch(block.Find("summary").First().Text())
		if m == nil {
			return
		}
		number, _ := strconv.Atoi(m[1])
		if number <= 0 {
			return
		}
		season := Season{Number: number, Audio: audio}
		block.Find("ol.episodes li").Each(func(_ int, li *goquery.Selection) {
			href := strings.TrimSpace(attrOr(li.Find("a[href]").First(), "href"))
			num, _ := strconv.Atoi(strings.TrimSpace(li.Find(".ep-num").First().Text()))
			if href == "" || num <= 0 {
				return
			}
			season.Episodes = append(season.Episodes, Episode{
				Number: num,
				Title:  strings.Join(strings.Fields(li.Find("strong").First().Text()), " "),
				Audio:  audio,
				URL:    urlResolve(pageURL, href),
			})
		})
		if len(season.Episodes) > 0 {
			sortEpisodes(season.Episodes)
			out = append(out, season)
		}
	})
	sortSeasons(out)
	return out
}

// parseV3Episode reads an episode page of the current site: its slot comes from
// the slug and its player from the page.
func parseV3Episode(doc *goquery.Document, pageURL string) (EpisodeData, bool) {
	_, slug, audio, tmdb := titleRecordFromPath(pathOfURL(pageURL))
	if tmdb <= 0 || !IsEpisodePage(pageURL) {
		return EpisodeData{}, false
	}
	m := v3EpisodeRe.FindStringSubmatch(pathOfURL(pageURL))
	servers := v3Servers(doc, audio)
	if m == nil || len(servers) == 0 {
		return EpisodeData{}, false
	}
	season, _ := strconv.Atoi(m[1])
	number, _ := strconv.Atoi(m[2])
	name := strings.Join(strings.Fields(doc.Find("h2").First().Text()), " ")
	return EpisodeData{
		Series: Title{
			Kind: KindSeries, VideoID: tmdb, TMDBID: tmdb, InternalID: tmdb,
			Audio: audio, Slug: slug,
			Name: displayTitle(doc.Find("h1.page-title").First().Text()),
		},
		Season:  season,
		Number:  number,
		Name:    name,
		Servers: servers,
	}, true
}

// busca is one answer of /api/busca.
type busca struct {
	Title string `json:"title"`
	Year  int    `json:"year"`
	Type  string `json:"type"`
	URL   string `json:"url"`
}

// buscaSearch asks the site's own search, which answers with the records and
// their page paths in one request.
func (c *Client) buscaSearch(ctx context.Context, query string) ([]Media, error) {
	body, err := c.get(ctx, c.base()+"/api/busca?q="+url.QueryEscape(query), getOptions{
		accept:  "application/json",
		referer: c.base() + "/",
	})
	if err != nil {
		return nil, err
	}
	var found []busca
	if err := json.Unmarshal(body, &found); err != nil {
		util.Debug("Pobreflix busca answered something that is not JSON", "err", err)
		return nil, err
	}
	out := make([]Media, 0, len(found))
	for _, f := range found {
		kind, _, audio, id := titleRecordFromPath(pathOfURL(f.URL))
		if kind == "" || id <= 0 || strings.TrimSpace(f.Title) == "" {
			continue
		}
		out = append(out, Media{
			Title:   strings.Join(strings.Fields(f.Title), " "),
			URL:     c.absoluteURL(f.URL),
			Kind:    kind,
			Audio:   audio,
			VideoID: id,
		})
	}
	return out, nil
}
