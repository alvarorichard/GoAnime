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
	// ErrNoTitle is returned when a title page carries no video mount, so it
	// names no TMDB id to play the title by.
	ErrNoTitle = errors.New("pobreflix: title page has no video mount")
)

// MediaKind tells movies from series. It follows the site's own URL layout:
// /filmes/online/<slug>-<id>/ for a movie, /series/online/<slug>-<id>/ for a
// series, /episodios/online/<slug>-<s>x<e>-<audio>-<id>/ for an episode.
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
	URL   string // title page, e.g. https://www.pobreflixtv.sale/filmes/online/verity-dublado-84506/
	Kind  MediaKind
	Audio Audio
	// VideoID is the record id the URL ends in; the sitemap corpus carries it
	// and the title page repeats it as data-video-id.
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
	Audio   Audio
	// Titles is the page's own display title ("O Apanhador de Sonhos"),
	// drawn from the record so a search-corpora slug ("o apanhador de
	// sonhos") can be replaced with the site's accented form.
	Name   string
	Poster string
	// Seasons carry the series' episode names; a movie lists none. The
	// playback flow rides the StartFlix panel — seasons here only name
	// the panel's episodes.
	Seasons []Season
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
	URL    string // the episode's own record page, for diagnostics
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
