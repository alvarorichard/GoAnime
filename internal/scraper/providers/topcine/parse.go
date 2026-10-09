package topcine

import (
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	// Player paths seen on azullog.top (2026-10-09): /filme/<tmdb> for a movie,
	// /serie/<tmdb>/<season>/<episode> for an episode. Matched on any host so a
	// player domain rotation does not lose the TMDB id.
	moviePlayerRe  = regexp.MustCompile(`^/filme/(\d+)/?$`)
	seriesPlayerRe = regexp.MustCompile(`^/serie/(\d+)/(\d+)/(\d+)/?$`)

	// yearRe finds the release year in a card's "2008 • Série" line.
	yearRe = regexp.MustCompile(`\b(19|20)\d{2}\b`)

	// genericEpisodeRe matches episode names that say nothing the number does
	// not: "Episódio 3", "Episodio".
	genericEpisodeRe = regexp.MustCompile(`(?i)^epis[óo]dio\s*\d*$`)

	// posterSizeRe upgrades the 342px TMDB poster the search page uses.
	posterSizeRe = regexp.MustCompile(`/t/p/w\d+/`)
)

// parseSearchResults reads the search grid. Only movies and series are kept.
func parseSearchResults(doc *goquery.Document, base *url.URL) []Media {
	var out []Media
	seen := map[string]bool{}
	doc.Find("article.pagina-busca-card").Each(func(_ int, card *goquery.Selection) {
		link := card.Find("a[href]").First()
		href, err := base.Parse(strings.TrimSpace(link.AttrOr("href", "")))
		if err != nil {
			return
		}
		kind, ok := kindFromPath(href.Path)
		if !ok || seen[href.String()] {
			return
		}
		title := strings.Join(strings.Fields(card.Find("h2").First().Text()), " ")
		if title == "" {
			return
		}
		seen[href.String()] = true

		poster := strings.TrimSpace(card.Find("img").First().AttrOr("src", ""))
		out = append(out, Media{
			Title:  title,
			URL:    href.String(),
			Poster: posterSizeRe.ReplaceAllString(poster, "/t/p/w500/"),
			Year:   yearRe.FindString(card.Find("small").First().Text()),
			Kind:   kind,
		})
	})
	return out
}

// kindFromPath classifies a title page by its path.
func kindFromPath(path string) (MediaKind, bool) {
	switch {
	case strings.HasPrefix(path, "/filme/") && len(path) > len("/filme/"):
		return KindMovie, true
	case strings.HasPrefix(path, "/serie/") && len(path) > len("/serie/"):
		return KindSeries, true
	default:
		return "", false
	}
}

// playerURL parses an absolute player URL, rejecting anything else — the
// site's own relative /filme/<slug> links included.
func playerURL(raw string) (*url.URL, bool) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
		return nil, false
	}
	return u, true
}

// episodeTitle keeps a real episode name and drops the generic ones.
func episodeTitle(raw string) string {
	title := strings.Join(strings.Fields(raw), " ")
	if genericEpisodeRe.MatchString(title) {
		return ""
	}
	return title
}

// parseTitle reads a title page: a series lists one player per episode in its
// episode selects, a movie iframes a single player.
func parseTitle(doc *goquery.Document) (Title, bool) {
	seasons := map[int]*Season{}
	tmdb := 0
	doc.Find("option[data-player]").Each(func(_ int, opt *goquery.Selection) {
		u, ok := playerURL(opt.AttrOr("data-player", ""))
		if !ok {
			return
		}
		m := seriesPlayerRe.FindStringSubmatch(u.Path)
		if m == nil {
			return
		}
		id, _ := strconv.Atoi(m[1])
		season, _ := strconv.Atoi(m[2])
		episode, _ := strconv.Atoi(m[3])
		if id <= 0 || episode <= 0 {
			return
		}
		// The first player names the series; one for another id is not its.
		if tmdb == 0 {
			tmdb = id
		} else if id != tmdb {
			return
		}
		s := seasons[season]
		if s == nil {
			s = &Season{Number: season}
			seasons[season] = s
		}
		s.Episodes = append(s.Episodes, Episode{
			Number:    episode,
			Title:     episodeTitle(opt.AttrOr("data-titulo", "")),
			PlayerURL: u.String(),
		})
	})
	if len(seasons) > 0 {
		t := Title{Kind: KindSeries, TMDBID: tmdb}
		for _, s := range seasons {
			sort.SliceStable(s.Episodes, func(i, j int) bool { return s.Episodes[i].Number < s.Episodes[j].Number })
			t.Seasons = append(t.Seasons, *s)
		}
		sort.Slice(t.Seasons, func(i, j int) bool { return t.Seasons[i].Number < t.Seasons[j].Number })
		return t, true
	}

	var t Title
	found := false
	doc.Find("iframe[src]").EachWithBreak(func(_ int, f *goquery.Selection) bool {
		u, ok := playerURL(f.AttrOr("src", ""))
		if !ok {
			return true
		}
		m := moviePlayerRe.FindStringSubmatch(u.Path)
		if m == nil {
			return true
		}
		id, _ := strconv.Atoi(m[1])
		if id <= 0 {
			return true
		}
		t, found = Title{Kind: KindMovie, TMDBID: id, PlayerURL: u.String()}, true
		return false
	})
	return t, found
}
