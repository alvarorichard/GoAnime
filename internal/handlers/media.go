// Package handlers provides HTTP handlers and flow controllers for media playback
package handlers

import (
	"fmt"
	"strings"

	"charm.land/huh/v2"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
)

// Testable hooks: tests may override these to bypass real TUI / network.
var (
	runFormFn    = tui.RunClean
	findFn       = tui.Find[string]
	findResultFn = tui.Find[*models.Anime]
)

// mediaSource is the subset of scraper.MediaManager that handlers depend on.
// It exists so tests can substitute a fake without spinning up network.
type mediaSource interface {
	SearchAnimeOnly(query string) ([]*models.Anime, error)
	SearchAll(query string) ([]*models.Anime, error)
	GetAnimeStreamURL(anime *models.Anime, episodeNum, quality, mode string) (string, map[string]string, error)
}

// MediaHandler handles media selection and playback operations
type MediaHandler struct {
	mediaManager mediaSource
	provider     string
	quality      string
	subsLanguage string
}

// NewMediaHandler creates a new MediaHandler
func NewMediaHandler() *MediaHandler {
	return &MediaHandler{
		mediaManager: scraper.NewMediaManager(),
		provider:     "Vidcloud",
		quality:      "best",
		subsLanguage: "english",
	}
}

// SetOptions sets playback options
func (mh *MediaHandler) SetOptions(provider, quality, subsLanguage string) {
	if provider != "" {
		mh.provider = provider
	}
	if quality != "" {
		mh.quality = quality
	}
	if subsLanguage != "" {
		mh.subsLanguage = subsLanguage
	}
}

// SearchMedia searches for media based on content type
func (mh *MediaHandler) SearchMedia(query string, contentType models.MediaType) ([]*models.Anime, error) {
	switch contentType {
	case models.MediaTypeAnime:
		return mh.mediaManager.SearchAnimeOnly(query)
	case models.MediaTypeMovie, models.MediaTypeTV:
		return nil, fmt.Errorf("movie/TV scrapers have been removed; only anime sources are supported")
	default:
		return mh.mediaManager.SearchAll(query)
	}
}

// SelectMediaType prompts user to select media type
func (mh *MediaHandler) SelectMediaType() (models.MediaType, error) {
	items := []string{"Anime", "Search All"}
	idx, err := findFn(items, func(i int) string {
		return items[i]
	}, fuzzyfinder.WithPromptString("Select content type: "))
	if err != nil {
		return "", err
	}

	switch idx {
	case 0:
		return models.MediaTypeAnime, nil
	default:
		return "", nil
	}
}

// SelectMedia prompts user to select from search results
func (mh *MediaHandler) SelectMedia(results []*models.Anime) (*models.Anime, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results to select from")
	}

	idx, err := findResultFn(results, func(i int) string {
		r := results[i]
		typeTag := ""
		switch r.MediaType {
		case models.MediaTypeMovie:
			typeTag = "[Movie]"
		case models.MediaTypeTV:
			typeTag = "[TV]"
		case models.MediaTypeAnime:
			typeTag = "[Anime]"
		}
		year := ""
		if r.Year != "" {
			year = fmt.Sprintf(" (%s)", r.Year)
		}
		return fmt.Sprintf("%s %s%s - %s", typeTag, r.Name, year, r.Source)
	}, fuzzyfinder.WithPromptString("Select media: "))
	if err != nil {
		return nil, err
	}

	return results[idx], nil
}

// GetAnimeStreamURL gets stream URL for anime content
func (mh *MediaHandler) GetAnimeStreamURL(anime *models.Anime, episodeNum string, mode string) (string, map[string]string, error) {
	return mh.mediaManager.GetAnimeStreamURL(anime, episodeNum, mh.quality, mode)
}

// InteractiveMediaFlow runs an interactive media selection and playback flow
func (mh *MediaHandler) InteractiveMediaFlow(query string) (*PlaybackInfo, error) {
	var contentType models.MediaType
	if query == "" {
		var err error
		contentType, err = mh.SelectMediaType()
		if err != nil {
			return nil, err
		}
	}

	if query == "" {
		var searchQuery string
		prompt := huh.NewInput().
			Title("Search").
			Value(&searchQuery)
		if err := runFormFn(prompt.Run); err != nil {
			return nil, err
		}
		query = searchQuery
	}

	results, err := mh.SearchMedia(query, contentType)
	if err != nil {
		return nil, err
	}

	util.Debug("Search results", "count", len(results))

	selected, err := mh.SelectMedia(results)
	if err != nil {
		return nil, err
	}

	playbackInfo := &PlaybackInfo{
		Title:     selected.Name,
		MediaType: selected.MediaType,
		Source:    selected.Source,
		ImageURL:  selected.ImageURL,
	}

	return mh.handleAnimePlayback(selected, playbackInfo)
}

func (mh *MediaHandler) handleAnimePlayback(anime *models.Anime, info *PlaybackInfo) (*PlaybackInfo, error) {
	var episodeNum string
	prompt := huh.NewInput().
		Title("Episode number").
		Value(&episodeNum).
		Validate(func(v string) error {
			if len(v) == 0 {
				return fmt.Errorf("episode number is required")
			}
			return nil
		})

	if err := runFormFn(prompt.Run); err != nil {
		return nil, err
	}
	if episodeNum == "" {
		episodeNum = "1"
	}

	modeItems := []string{"Sub (Subtitled)", "Dub (English Dubbed)"}
	modeIdx, err := findFn(modeItems, func(i int) string {
		return modeItems[i]
	}, fuzzyfinder.WithPromptString("Select audio: "))
	if err != nil {
		return nil, err
	}

	mode := "sub"
	if modeIdx == 1 {
		mode = "dub"
	}

	streamURL, metadata, err := mh.GetAnimeStreamURL(anime, episodeNum, mode)
	if err != nil {
		return nil, err
	}

	info.StreamURL = streamURL
	info.Episode = fmt.Sprintf("Episode %s", episodeNum)
	info.Metadata = metadata

	return info, nil
}

// PlaybackInfo contains all information needed for playback
type PlaybackInfo struct {
	Title      string
	MediaType  models.MediaType
	Source     string
	Season     string
	Episode    string
	EpisodeNum int
	StreamURL  string
	Quality    string
	Subtitles  []models.Subtitle
	Referer    string
	ImageURL   string
	Metadata   map[string]string
}

// extractIDFromURL extracts a trailing ID segment from a URL.
func extractIDFromURL(urlStr string) string {
	parts := strings.Split(urlStr, "-")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// unusedExtractIDForward keeps the helper exported within the package so future
// scrapers that adopt path-based IDs can reuse it without re-introducing the
// function.
var _ = extractIDFromURL
