package topcine

import (
	"errors"

	"github.com/alvarorichard/Goanime/internal/models"
)

// SourceName is the canonical Source string TopCine results carry.
const SourceName = "TopCine"

// Content signals: what the site offers, not a scraping fault.
var (
	// ErrNoPlayer is returned when a title page embeds no player, so it names
	// no TMDB id to play the title by.
	ErrNoPlayer = errors.New("topcine: title page has no player")
)

// MediaKind tells movies from series. It follows the site's own URL layout:
// /filme/<slug> and /serie/<slug>.
type MediaKind string

const (
	KindMovie  MediaKind = "movie"
	KindSeries MediaKind = "series"
)

// Media is one search result.
type Media struct {
	Title  string
	URL    string // title page, e.g. https://topcine3.site/serie/breaking-bad
	Poster string
	Year   string
	Kind   MediaKind
}

// ToAnimeModel converts a search result into the app-wide model. The title page
// URL is the identifier; the TMDB id is read from that page when it is opened.
func (m Media) ToAnimeModel() *models.Anime {
	anime := &models.Anime{
		Name:     m.Title,
		URL:      m.URL,
		ImageURL: m.Poster,
		Source:   SourceName,
		Year:     m.Year,
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
	Kind   MediaKind
	TMDBID int
	// PlayerURL is the movie's player; series carry one per episode.
	PlayerURL string
	Seasons   []Season
}

// Season is one entry of a series' season select.
type Season struct {
	Number   int
	Episodes []Episode
}

// Episode is one entry of a season's episode select.
type Episode struct {
	Number    int
	Title     string
	PlayerURL string // e.g. https://azullog.top/serie/1396/1/1
}

// EpisodeTitle returns the name TopCine gives an episode, "" when it lists
// none.
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

// Languages is the player's answer to "which audio tracks exist".
type Languages struct {
	Dubbed    bool
	Subtitled bool
	// Hosts are the hosts the tracks are served from, for diagnostics.
	Hosts []string
}
