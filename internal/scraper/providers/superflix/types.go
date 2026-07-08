package superflix

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
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
	ID   json.RawMessage `json:"ID"`
	Name string          `json:"name"`
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
