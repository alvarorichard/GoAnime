package startflix

import (
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

var (
	// panelPathRe matches the two panel layouts. The TMDB id is digits only,
	// which is also what keeps a YouTube trailer (youtube.com/embed/<11 chars>)
	// from being taken for the series panel.
	panelPathRe = regexp.MustCompile(`^/(embed|filme)/(\d+|tt\d+)/?$`)

	// Episode labels seen on the panel (2026-10-07): "1 - Episódio",
	// "1 - Velhos Conhecimentos" and "Episódio 1".
	episodeLeadingNumRe = regexp.MustCompile(`^\s*(\d+)\s*[-–:.]`)
	episodeWordNumRe    = regexp.MustCompile(`(?i)epis[óo]dio\s*(\d+)`)
	episodeTitleRe      = regexp.MustCompile(`^\s*\d+\s*[-–:.]\s*`)

	// posterSizeRe upgrades the 92px TMDB thumbnail the search page uses.
	posterSizeRe = regexp.MustCompile(`/t/p/w\d+/`)

	// notFoundMarker is the panel's whole body for a title it does not carry.
	notFoundMarker = "Conteúdo não encontrado"
)

// parseSearchResults reads the DooPlay search page. Only movies and series are
// kept; anything else the theme might list (episodes, people) is skipped.
func parseSearchResults(doc *goquery.Document) []Media {
	var out []Media
	seen := map[string]bool{}
	doc.Find("div.result-item article").Each(func(_ int, item *goquery.Selection) {
		link := item.Find(".details .title a").First()
		href, _ := link.Attr("href")
		title := strings.TrimSpace(link.Text())
		if href == "" || title == "" || seen[href] {
			return
		}
		kind, ok := kindFromURL(href)
		if !ok {
			return
		}
		seen[href] = true

		poster, _ := item.Find(".image img").First().Attr("src")
		out = append(out, Media{
			Title:  title,
			URL:    href,
			Poster: posterSizeRe.ReplaceAllString(poster, "/t/p/w500/"),
			Year:   strings.TrimSpace(item.Find(".meta .year").First().Text()),
			Kind:   kind,
		})
	})
	return out
}

// kindFromURL classifies a title page by its path.
func kindFromURL(raw string) (MediaKind, bool) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}
	switch {
	case strings.HasPrefix(u.Path, "/filmes/"):
		return KindMovie, true
	case strings.HasPrefix(u.Path, "/series/"):
		return KindSeries, true
	default:
		return "", false
	}
}

// parsePanel finds the video panel a title page iframes.
func parsePanel(doc *goquery.Document) (Panel, bool) {
	var panel Panel
	found := false
	doc.Find("iframe[src]").EachWithBreak(func(_ int, f *goquery.Selection) bool {
		src, _ := f.Attr("src")
		p, ok := panelFromURL(src)
		if ok {
			panel, found = p, true
		}
		return !ok
	})
	return panel, found
}

func panelFromURL(raw string) (Panel, bool) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
		return Panel{}, false
	}
	m := panelPathRe.FindStringSubmatch(u.Path)
	if m == nil {
		return Panel{}, false
	}
	p := Panel{URL: u.Scheme + "://" + u.Host + "/" + m[1] + "/" + m[2]}
	if m[1] == "filme" {
		if !strings.HasPrefix(m[2], "tt") {
			return Panel{}, false
		}
		p.Kind = KindMovie
		p.IMDBID = m[2]
		return p, true
	}
	p.Kind = KindSeries
	id, err := strconv.Atoi(m[2])
	if err != nil {
		return Panel{}, false
	}
	p.TMDBID = id
	return p, true
}

// parseSeasons reads a series panel: the season tabs, then one card per audio
// whose rows carry (season id, episode id).
func parseSeasons(doc *goquery.Document) []Season {
	numberByID := map[string]int{}
	var order []string
	doc.Find(".header-navigation li[data-season-id]").Each(func(_ int, li *goquery.Selection) {
		id, _ := li.Attr("data-season-id")
		n, err := strconv.Atoi(strings.TrimSpace(li.AttrOr("data-season-number", "")))
		if id == "" || err != nil {
			return
		}
		if _, dup := numberByID[id]; !dup {
			order = append(order, id)
		}
		numberByID[id] = n
	})

	seasons := map[string]*Season{}
	for _, id := range order {
		seasons[id] = &Season{Number: numberByID[id], ID: id}
	}

	doc.Find(".cards .card").Each(func(_ int, card *goquery.Selection) {
		audio, ok := audioFromCardTitle(card.Find(".card-title").First().Text())
		if !ok {
			return
		}
		position := map[string]int{}
		card.Find("li[data-episode-id]").Each(func(_ int, li *goquery.Selection) {
			seasonID, _ := li.Attr("data-season-id")
			episodeID, _ := li.Attr("data-episode-id")
			s, known := seasons[seasonID]
			if !known || episodeID == "" {
				return
			}
			position[seasonID]++
			label := strings.Join(strings.Fields(li.Find("a").First().Text()), " ")
			ep := Episode{
				Number: episodeNumber(label, position[seasonID]),
				Title:  episodeTitle(label),
				ID:     episodeID,
				Audio:  audio,
			}
			if audio == AudioDubbed {
				s.Dubbed = append(s.Dubbed, ep)
			} else {
				s.Subtitled = append(s.Subtitled, ep)
			}
		})
	})

	out := make([]Season, 0, len(seasons))
	for _, id := range order {
		s := seasons[id]
		if len(s.Dubbed) == 0 && len(s.Subtitled) == 0 {
			continue
		}
		sortEpisodes(s.Dubbed)
		sortEpisodes(s.Subtitled)
		out = append(out, *s)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Number < out[j].Number })
	return out
}

func audioFromCardTitle(text string) (Audio, bool) {
	switch strings.ToLower(strings.TrimSpace(text)) {
	case "dublado":
		return AudioDubbed, true
	case "legendado":
		return AudioSubtitled, true
	default:
		return "", false
	}
}

// episodeNumber reads the number out of a row label. The panel lists episodes
// interleaved across seasons but in order within one, so the row's position in
// its season is a sound fallback for a label that carries no number.
func episodeNumber(label string, position int) int {
	for _, re := range []*regexp.Regexp{episodeLeadingNumRe, episodeWordNumRe} {
		if m := re.FindStringSubmatch(label); m != nil {
			if n, err := strconv.Atoi(m[1]); err == nil {
				return n
			}
		}
	}
	return position
}

// episodeTitle keeps a real episode name and drops the generic ones, which say
// nothing the episode number does not.
func episodeTitle(label string) string {
	title := strings.TrimSpace(episodeTitleRe.ReplaceAllString(label, ""))
	if title == "" || episodeWordNumRe.MatchString(title) || strings.EqualFold(title, "episódio") || strings.EqualFold(title, "episodio") {
		return ""
	}
	return title
}

func sortEpisodes(eps []Episode) {
	sort.SliceStable(eps, func(i, j int) bool { return eps[i].Number < eps[j].Number })
}

// inlinePlayerRe finds the play({...}) call an episode with a single player
// renders instead of buttons — Breaking Bad S01E01 on 2026-10-09.
var inlinePlayerRe = regexp.MustCompile(`(?s)\bplay\((\{.*?\})\);?\s*(?:</script>|$)`)

// inlinePlayer is the record that call passes.
type inlinePlayer struct {
	Title     string  `json:"title"`
	Source    string  `json:"source"`
	Subtitles *string `json:"subtitles"`
	Player    string  `json:"player"`
	Type      string  `json:"type"`
	ID        int     `json:"id"`
}

// parsePlayers reads the player buttons from an episode fragment or a movie
// panel. Both render the same markup — except when there is a single player,
// which the panel starts at once from a script; see parseInlinePlayers.
func parsePlayers(doc *goquery.Document) []Player {
	if out := parsePlayerButtons(doc); len(out) > 0 {
		return out
	}
	return parseInlinePlayers(doc)
}

// parseInlinePlayers reads the play({...}) calls in the page's scripts.
func parseInlinePlayers(doc *goquery.Document) []Player {
	var out []Player
	doc.Find("script").Each(func(_ int, script *goquery.Selection) {
		for _, m := range inlinePlayerRe.FindAllStringSubmatch(script.Text()+"</script>", -1) {
			var rec inlinePlayer
			if err := jsonx.Unmarshal([]byte(m[1]), &rec); err != nil || strings.TrimSpace(rec.Source) == "" {
				continue
			}
			kind := rec.Type
			if kind == "" {
				kind = rec.Player
			}
			p := Player{
				URL:   strings.TrimSpace(rec.Source),
				Type:  strings.TrimSpace(kind),
				Label: strings.Join(strings.Fields(rec.Title), " "),
			}
			if rec.Subtitles != nil {
				p.Subtitles = strings.TrimSpace(*rec.Subtitles)
			}
			if rec.ID > 0 {
				p.ID = strconv.Itoa(rec.ID)
			}
			out = append(out, p)
		}
	})
	return out
}

// parsePlayerButtons reads the "SELECIONE UM PLAYER" buttons.
func parsePlayerButtons(doc *goquery.Document) []Player {
	var out []Player
	doc.Find("[data-show-player][data-source]").Each(func(_ int, b *goquery.Selection) {
		src := strings.TrimSpace(b.AttrOr("data-source", ""))
		if src == "" {
			return
		}
		out = append(out, Player{
			URL:       src,
			Type:      strings.TrimSpace(b.AttrOr("data-type", "")),
			Subtitles: strings.TrimSpace(b.AttrOr("data-subtitles", "")),
			Label:     strings.Join(strings.Fields(b.Text()), " "),
			ID:        strings.TrimSpace(b.AttrOr("data-id", "")),
		})
	})
	return out
}

// isPanelNotFound reports the panel's "no content" answer.
func isPanelNotFound(body []byte) bool {
	return len(body) < 512 && strings.Contains(string(body), notFoundMarker)
}
