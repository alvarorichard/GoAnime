package startflix

import (
	"errors"
	"strconv"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
)

// SourceName is the canonical Source string StartFlix results carry.
const SourceName = "StartFlix"

// Content signals. They describe what the site offers, not a scraping fault, so
// callers should tell the user rather than retry.
var (
	// ErrNoPanel is returned when a title page embeds no video panel at all.
	ErrNoPanel = errors.New("startflix: title page has no video panel")

	// ErrNotOnPanel is returned when the panel answers "Conteúdo não encontrado"
	// for a title StartFlix lists — Naruto (TMDB 46260) was one on 2026-10-07.
	ErrNotOnPanel = errors.New("startflix: the video panel has no content for this title")

	// ErrNoPlayers is returned when an episode or movie lists no player at all.
	ErrNoPlayers = errors.New("startflix: no players listed for this content")

	// ErrNoSupportedServer is returned when players exist but none is on a host
	// this package can resolve. The error text names the hosts that were seen.
	ErrNoSupportedServer = errors.New("startflix: no supported server for this content")
)

// MediaKind tells movies from series. It follows the site's own URL layout:
// /filmes/<slug>/ and /series/<slug>/.
type MediaKind string

const (
	KindMovie  MediaKind = "movie"
	KindSeries MediaKind = "series"
)

// Media is one search result.
type Media struct {
	Title  string
	URL    string // title page, e.g. https://www.startflix.biz/series/naruto/
	Poster string
	Year   string
	Kind   MediaKind
}

// ToAnimeModel converts a search result into the app-wide model. The title page
// URL is the identifier: everything else (panel, TMDB/IMDb ids) is read from
// that page when the title is opened.
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

// Panel is the page a title's videos live on. StartFlix does not host video
// itself: each title page iframes a third-party panel keyed by TMDB id for
// series (/embed/<tmdb>) and by IMDb id for movies (/filme/<tt…>).
type Panel struct {
	URL    string
	Kind   MediaKind
	TMDBID int
	IMDBID string
}

// Origin returns the panel's scheme://host, which is where its /episodio/<id>
// endpoint lives.
func (p Panel) Origin() string {
	if i := strings.Index(p.URL, "://"); i >= 0 {
		if j := strings.IndexByte(p.URL[i+3:], '/'); j >= 0 {
			return p.URL[:i+3+j]
		}
	}
	return p.URL
}

// EpisodeURL is the endpoint listing one episode's players.
func (p Panel) EpisodeURL(episodeID string) string {
	return p.Origin() + "/episodio/" + episodeID
}

// Audio is the panel's own name for an episode list: it groups episodes into a
// "Dublado" card and a "Legendado" card, each with its own episode ids.
type Audio string

const (
	AudioDubbed    Audio = "Dublado"
	AudioSubtitled Audio = "Legendado"
)

// Episode is one entry in a season's Dublado or Legendado list.
type Episode struct {
	Number int
	Title  string
	ID     string // panel episode id, resolved through Panel.EpisodeURL
	Audio  Audio
}

// Season groups a season's episodes by audio. Either list may be empty: a
// season can be dubbed-only or subtitled-only.
type Season struct {
	Number    int
	ID        string
	Dubbed    []Episode
	Subtitled []Episode
}

// Key renders the season number the way the season picker keys it.
func (s Season) Key() string { return strconv.Itoa(s.Number) }

// Audios lists the audio tracks this season offers, dub first.
func (s Season) Audios() []Audio {
	var out []Audio
	if len(s.Dubbed) > 0 {
		out = append(out, AudioDubbed)
	}
	if len(s.Subtitled) > 0 {
		out = append(out, AudioSubtitled)
	}
	return out
}

// Episodes returns the list for the requested audio.
func (s Season) Episodes(a Audio) []Episode {
	if a == AudioSubtitled {
		return s.Subtitled
	}
	return s.Dubbed
}

// Player is one button in the panel's "SELECIONE UM PLAYER" list.
type Player struct {
	URL       string
	Type      string // iframe, jwplayer, videojs or plyr
	Subtitles string
	Label     string
	ID        string

	// referer is the origin of the page that listed this player, which is
	// what the panel's own <video> element would send.
	referer string
}

// IsDirectFile reports whether the panel plays this source in its own <video>
// element, i.e. the URL is the media itself rather than a host's embed page.
func (p Player) IsDirectFile() bool {
	switch strings.ToLower(p.Type) {
	case "jwplayer", "videojs", "plyr":
		return true
	default:
		return false
	}
}

// Subtitle is an external subtitle track.
type Subtitle struct {
	URL      string
	Language string
	Label    string
}

// Stream is a resolved, playable media URL.
type Stream struct {
	URL       string
	Referer   string
	Subtitles []Subtitle
	// Host is the player host that served it, for diagnostics.
	Host string
}
