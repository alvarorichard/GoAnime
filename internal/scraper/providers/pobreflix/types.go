package pobreflix

import (
	"errors"
	"regexp"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
)

// SourceName is the canonical Source string Pobreflix results carry.
const SourceName = "Pobreflix"

// Content signals: what the site offers, not a scraping fault.
var (
	// ErrNoTitle is returned when a title page carries neither the current
	// site's embedded JSON nor the old theme's video mount, so it names no
	// TMDB id to play the title by.
	ErrNoTitle = errors.New("pobreflix: title page has no video mount")

	// ErrNoEpisode is returned when an episode page carries no embedded
	// episode data. The current site always embeds it; an older theme does
	// not, and the api layer falls back to the StartFlix panel there.
	ErrNoEpisode = errors.New("pobreflix: episode page has no embedded episode data")
)

// MediaKind tells movies from series. It follows both URL layouts the site has
// used: the old /filmes/online/<slug>-<id>/ and /series/online/<slug>-<id>/
// records, and the current /filme/<slug>.html and /serie/<slug>.html pages.
type MediaKind string

const (
	KindMovie  MediaKind = "movie"
	KindSeries MediaKind = "series"
)

// Audio is the audio track a record names in its slug. A page covers one
// record; the record is the title in one audio.
type Audio string

const (
	AudioDubbed    Audio = "Dublado"
	AudioSubtitled Audio = "Legendado"
	AudioUnknown   Audio = ""
)

// Media is one search result.
type Media struct {
	Title string
	URL   string // title page, e.g. https://pobreflixtv.gay/filme/verity-dublado.html
	Kind  MediaKind
	Audio Audio
	// VideoID is the site's own record id when the URL carries one. Old
	// records end in it; current slugs end in a random suffix or the TMDB id,
	// so it is 0 for most current results and is only a recency tie-break.
	VideoID int
}

// ToAnimeModel converts a search result into the app-wide model. The title
// page URL is the identifier; the TMDB id is read from that page when it is
// opened (the sitemap corpus cannot know it).
func (m Media) ToAnimeModel() *models.Anime {
	anime := &models.Anime{
		Name:   m.Title,
		URL:    m.URL,
		Source: SourceName,
		Audio:  models.AudioLabel(m.Audio == AudioDubbed, m.Audio == AudioSubtitled),
	}
	if m.Kind == KindMovie {
		anime.MediaType = models.MediaTypeMovie
	} else {
		anime.MediaType = models.MediaTypeTV
	}
	return anime
}

// Title is what a title page says about the title.
type Title struct {
	Kind    MediaKind
	VideoID int
	TMDBID  int
	// InternalID is the site's own record id. The site's /api stream endpoint
	// keys movies by it; the TMDB id remains the cross-source key.
	InternalID int
	Audio      Audio
	// Titles is the page's own display title ("O Apanhador de Sonhos"),
	// drawn from the record so a search-corpora slug ("o apanhador de
	// sonhos") can be replaced with the site's accented form.
	Name   string
	Poster string
	// Slug is the record's current-site slug (/filme/<slug>.html,
	// /serie/<slug>.html). It is what episode URLs are built from.
	Slug string
	// Servers are the playable files a movie page embeds. Series pages list
	// their servers per episode instead.
	Servers []Server
	// Seasons carry the series' episode names and episode-page URLs; a movie
	// lists none.
	Seasons []Season
}

// Server is one playable file the site offers for a title or episode.
type Server struct {
	URL      string
	Download string
	Quality  string
	Audio    Audio
	Name     string
}

// EpisodeData is one episode page: the series it belongs to, its slot, and
// the playable files the page embeds.
type EpisodeData struct {
	Series  Title
	Season  int
	Number  int
	Name    string
	Servers []Server
}

// Season is one entry of a series' season select.
type Season struct {
	Number   int
	Audio    Audio // the season option's audio grouping
	Episodes []Episode
}

// Episode is one entry of a season's episode list.
type Episode struct {
	Number int
	Title  string
	Audio  Audio  // the audio the embedded episode record belongs to
	URL    string // the episode's own record page
}

// HasDirect reports whether the page embeds playable files. Current pages do:
// a movie lists its servers, a series lists its episodes (whose pages list the
// servers). Old-theme pages embed neither and play through the StartFlix panel.
func (t Title) HasDirect() bool {
	if t.Kind == KindMovie {
		return len(t.Servers) > 0
	}
	return len(t.Seasons) > 0
}

// Season returns the season with the requested number.
func (t Title) Season(number int) (Season, bool) {
	for _, s := range t.Seasons {
		if s.Number == number {
			return s, true
		}
	}
	return Season{}, false
}

// EpisodePath is the current site's episode-page path for a series title:
// /assistir/<slug>-t<season>-ep<episode>.html. It returns "" when the title
// carries no slug.
func (t Title) EpisodePath(season, episode int) string {
	if t.Slug == "" || season <= 0 || episode <= 0 {
		return ""
	}
	return "/assistir/" + t.Slug + "-t" + itoa(season) + "-ep" + itoa(episode) + ".html"
}

// EpisodeTitle returns the name the page gives an episode of the current
// season, "" when it lists none.
func (t Title) EpisodeTitle(season, episode int) string {
	for _, s := range t.Seasons {
		if s.Number != season {
			continue
		}
		for _, e := range s.Episodes {
			if e.Number == episode {
				return e.Title
			}
		}
	}
	return ""
}

// displayTitle cleans a page title: the h1 reads "Assistir Verity Online
// Gratis (Filme HD)"; an episode names its slot ("… 1X1 Online") — the slot
// goes, and the "Online (Gratis)" tail with it. The record's plain name is
// what remains.
func displayTitle(raw string) string {
	name := strings.Join(strings.Fields(raw), " ")
	name = strings.TrimPrefix(name, "Assistir ")
	for _, suffix := range []string{" Online Gratis", " Online Grátis", " Online"} {
		if strings.HasSuffix(name, suffix) {
			name = strings.TrimSuffix(name, suffix)
			break
		}
	}
	name = episodeMarkerRe.ReplaceAllString(name, "")
	return strings.TrimSpace(name)
}

// episodeMarkerRe strips an episode page's "1X1" slot off a title tail.
var episodeMarkerRe = regexp.MustCompile(`(?i)\s+\d+X\d+\s*$`)
