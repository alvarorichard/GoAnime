package superflix

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// ErrSuperFlixNoServers is returned when /player/bootstrap responds with an
// empty options list. This is a content-availability signal from SuperFlix
// (the upstream JS shows a "not yet released" screen in the same case), not
// a system or scraping error — callers should surface it to the user as

// SuperFlixTokens holds the tokens extracted from a SuperFlix player page
type SuperFlixTokens struct {
	CSRF        string
	PageToken   string
	ContentID   string
	ContentType string
	Title       string
}

// SuperFlixServer represents a streaming server option
type SuperFlixServer struct {
	// ID is either a number (159462) or a string ("native_media:233831"), so it
	// stays raw until the caller needs it as text.
	ID   json.RawMessage `json:"ID"`
	Name string          `json:"name"`

	// Type is the audio track this server carries, and it is what the site's own
	// player uses to build its "Dublado" / "Legendado" tabs:
	//   1 = Dublado (Portuguese dub)
	//   2 = Legendado (original audio, Portuguese subtitles)
	// A title can offer several servers of the same type, and some episodes offer
	// only one type — Tehran S1E1, for example, is dubbed-only.
	Type int `json:"type"`

	// IsFile marks a direct MP4 rather than a streaming server. It is stable
	// across episodes (unlike a server's name, which embeds its per-episode id),
	// so it is what makes remembering the user's pick possible.
	IsFile bool `json:"is_file"`

	// CanDownload is advisory metadata from the site; kept so a future download
	// path can honor it.
	CanDownload bool `json:"can_download"`
}

// Audio track types as the SuperFlix player labels them.
const (
	SuperFlixAudioDubbed    = 1 // Dublado
	SuperFlixAudioSubtitled = 2 // Legendado
)

// IDString renders the raw ID as text, handling both the numeric and the string
// form SuperFlix mixes in the same list.
func (s SuperFlixServer) IDString() string {
	var str string
	if err := jsonx.Unmarshal(s.ID, &str); err == nil {
		return str
	}
	var num json.Number
	if err := jsonx.Unmarshal(s.ID, &num); err == nil {
		return num.String()
	}
	return ""
}

// IsFallback reports whether SuperFlix is offering a placeholder rather than a
// real source. Its own player filters these out before showing the server list.
func (s SuperFlixServer) IsFallback() bool {
	return strings.HasPrefix(strings.ToLower(s.IDString()), "fallback")
}

// SuperFlixSubtitle represents a subtitle track
type SuperFlixSubtitle struct {
	Lang string
	URL  string
}

// SuperFlixStreamResult holds the final stream extraction result
type SuperFlixStreamResult struct {
	StreamURL    string
	Title        string
	Referer      string
	Subtitles    []SuperFlixSubtitle
	DefaultAudio []string
	Thumb        string
}

// SuperFlixEpisode represents a single episode in a season
type SuperFlixEpisode struct {
	EpiNum  json.Number `json:"epi_num"`
	Title   string      `json:"title"`
	AirDate string      `json:"air_date"`
}

// SuperFlixMedia represents a search result from SuperFlix
type SuperFlixMedia struct {
	Title    string
	Year     string
	Type     string // "Filme", "Série", etc.
	SFType   string // "filme" or "serie"
	TMDBID   string
	IMDBID   string
	ImageURL string // Cover image URL from search results
}

// ToAnimeModel converts SuperFlixMedia to models.Anime for compatibility
func (m *SuperFlixMedia) ToAnimeModel() *models.Anime {
	anime := &models.Anime{
		Name:     m.Title,
		URL:      m.TMDBID, // Store TMDB ID as URL identifier
		Source:   "SuperFlix",
		Year:     m.Year,
		ImageURL: m.ImageURL,
	}

	lowerType := strings.ToLower(m.Type)
	switch {
	case m.SFType == "filme":
		anime.MediaType = models.MediaTypeMovie
	case lowerType == "anime" || lowerType == "dorama":
		anime.MediaType = models.MediaTypeAnime
	default:
		anime.MediaType = models.MediaTypeTV
	}

	if m.IMDBID != "" {
		anime.IMDBID = m.IMDBID
	}

	// Parse TMDB ID for direct API lookups during enrichment
	if m.TMDBID != "" {
		if id, err := strconv.Atoi(m.TMDBID); err == nil {
			anime.TMDBID = id
		}
	}

	util.Debug("SuperFlix ToAnimeModel", "title", m.Title, "tmdbID", m.TMDBID, "imageURL", anime.ImageURL)

	return anime
}
