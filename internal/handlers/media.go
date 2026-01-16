// Package handlers provides HTTP handlers and flow controllers for media playback
package handlers

import (
	"fmt"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/manifoldco/promptui"
)

// MediaHandler handles media selection and playback operations
type MediaHandler struct {
	mediaManager *scraper.MediaManager
	provider     string
	quality      string
	subsLanguage string
}

// NewMediaHandler creates a new MediaHandler
func NewMediaHandler() *MediaHandler {
	return &MediaHandler{
		mediaManager: scraper.NewMediaManager(),
		provider:     "Vidcloud",
		quality:      "1080",
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
		media, err := mh.mediaManager.SearchMoviesAndTV(query)
		if err != nil {
			return nil, err
		}
		return scraper.ConvertFlixHQToAnime(media), nil
	default:
		return mh.mediaManager.SearchAll(query)
	}
}

// SelectMediaType prompts user to select media type
func (mh *MediaHandler) SelectMediaType() (models.MediaType, error) {
	prompt := promptui.Select{
		Label: "Select content type",
		Items: []string{"Anime", "Movies", "TV Shows", "Search All"},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	switch idx {
	case 0:
		return models.MediaTypeAnime, nil
	case 1:
		return models.MediaTypeMovie, nil
	case 2:
		return models.MediaTypeTV, nil
	default:
		return "", nil // Search all
	}
}

// SelectMedia prompts user to select from search results
func (mh *MediaHandler) SelectMedia(results []*models.Anime) (*models.Anime, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results to select from")
	}

	// Prepare display items
	var items []string
	for _, r := range results {
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
		items = append(items, fmt.Sprintf("%s %s%s - %s", typeTag, r.Name, year, r.Source))
	}

	prompt := promptui.Select{
		Label: "Select media",
		Items: items,
		Size:  15,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return results[idx], nil
}

// SelectSeason prompts user to select a TV season
func (mh *MediaHandler) SelectSeason(mediaID string) (*scraper.FlixHQSeason, error) {
	seasons, err := mh.mediaManager.GetTVSeasons(mediaID)
	if err != nil {
		return nil, err
	}

	if len(seasons) == 0 {
		return nil, fmt.Errorf("no seasons found")
	}

	var items []string
	for _, s := range seasons {
		items = append(items, s.Title)
	}

	prompt := promptui.Select{
		Label: "Select season",
		Items: items,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &seasons[idx], nil
}

// SelectEpisode prompts user to select a TV episode
func (mh *MediaHandler) SelectEpisode(seasonID string) (*scraper.FlixHQEpisode, error) {
	episodes, err := mh.mediaManager.GetTVEpisodes(seasonID)
	if err != nil {
		return nil, err
	}

	if len(episodes) == 0 {
		return nil, fmt.Errorf("no episodes found")
	}

	var items []string
	for _, e := range episodes {
		items = append(items, fmt.Sprintf("Episode %d: %s", e.Number, e.Title))
	}

	prompt := promptui.Select{
		Label: "Select episode",
		Items: items,
		Size:  15,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &episodes[idx], nil
}

// GetStreamInfo gets streaming information for selected media
func (mh *MediaHandler) GetStreamInfo(media *models.Anime, episode *scraper.FlixHQEpisode) (*scraper.FlixHQStreamInfo, error) {
	source := strings.ToLower(media.Source)

	if !strings.Contains(source, "flixhq") {
		return nil, fmt.Errorf("media source %s does not support FlixHQ streaming", media.Source)
	}

	// Extract media ID from URL
	mediaID := extractIDFromURL(media.URL)
	if mediaID == "" {
		return nil, fmt.Errorf("could not extract media ID from URL: %s", media.URL)
	}

	if media.MediaType == models.MediaTypeMovie {
		return mh.mediaManager.GetMovieStreamInfo(mediaID, mh.provider, mh.quality, mh.subsLanguage)
	}

	if episode == nil {
		return nil, fmt.Errorf("episode is required for TV shows")
	}

	return mh.mediaManager.GetTVEpisodeStreamInfo(episode.DataID, mh.provider, mh.quality, mh.subsLanguage)
}

// GetAnimeStreamURL gets stream URL for anime content
func (mh *MediaHandler) GetAnimeStreamURL(anime *models.Anime, episodeNum string, mode string) (string, map[string]string, error) {
	return mh.mediaManager.GetAnimeStreamURL(anime, episodeNum, mh.quality, mode)
}

// InteractiveMediaFlow runs an interactive media selection and playback flow
func (mh *MediaHandler) InteractiveMediaFlow(query string) (*PlaybackInfo, error) {
	// Select media type if not already searching
	var contentType models.MediaType
	if query == "" {
		var err error
		contentType, err = mh.SelectMediaType()
		if err != nil {
			return nil, err
		}
	}

	// Get search query if not provided
	if query == "" {
		searchPrompt := promptui.Prompt{
			Label: "Search",
		}
		var err error
		query, err = searchPrompt.Run()
		if err != nil {
			return nil, err
		}
	}

	// Search for media
	results, err := mh.SearchMedia(query, contentType)
	if err != nil {
		return nil, err
	}

	util.Debug("Search results", "count", len(results))

	// Select media
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

	// Handle based on media type and source
	if strings.Contains(strings.ToLower(selected.Source), "flixhq") {
		return mh.handleFlixHQPlayback(selected, playbackInfo)
	}

	// Handle anime sources
	return mh.handleAnimePlayback(selected, playbackInfo)
}

func (mh *MediaHandler) handleFlixHQPlayback(media *models.Anime, info *PlaybackInfo) (*PlaybackInfo, error) {
	mediaID := extractIDFromURL(media.URL)

	if media.MediaType == models.MediaTypeMovie {
		streamInfo, err := mh.mediaManager.GetMovieStreamInfo(mediaID, mh.provider, mh.quality, mh.subsLanguage)
		if err != nil {
			return nil, err
		}
		info.StreamURL = streamInfo.VideoURL
		info.Subtitles = convertSubtitles(streamInfo.Subtitles)
		return info, nil
	}

	// TV Show flow
	season, err := mh.SelectSeason(mediaID)
	if err != nil {
		return nil, err
	}
	info.Season = season.Title

	episode, err := mh.SelectEpisode(season.ID)
	if err != nil {
		return nil, err
	}
	info.Episode = episode.Title
	info.EpisodeNum = episode.Number

	streamInfo, err := mh.mediaManager.GetTVEpisodeStreamInfo(episode.DataID, mh.provider, mh.quality, mh.subsLanguage)
	if err != nil {
		return nil, err
	}
	info.StreamURL = streamInfo.VideoURL
	info.Subtitles = convertSubtitles(streamInfo.Subtitles)

	return info, nil
}

func (mh *MediaHandler) handleAnimePlayback(anime *models.Anime, info *PlaybackInfo) (*PlaybackInfo, error) {
	// For anime, we need to select an episode
	episodePrompt := promptui.Prompt{
		Label:   "Episode number",
		Default: "1",
	}

	episodeNum, err := episodePrompt.Run()
	if err != nil {
		return nil, err
	}

	modePrompt := promptui.Select{
		Label: "Select audio",
		Items: []string{"Sub (Subtitled)", "Dub (English Dubbed)"},
	}

	modeIdx, _, err := modePrompt.Run()
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
	Subtitles  []models.Subtitle
	Referer    string
	ImageURL   string
	Metadata   map[string]string
}

// Helper functions

func extractIDFromURL(urlStr string) string {
	// Extract ID from URL like /movie/watch-movie-name-12345 or /tv/watch-show-name-12345
	parts := strings.Split(urlStr, "-")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func convertSubtitles(flixSubs []scraper.FlixHQSubtitle) []models.Subtitle {
	var subs []models.Subtitle
	for _, fs := range flixSubs {
		subs = append(subs, models.Subtitle{
			URL:      fs.URL,
			Language: fs.Language,
			Label:    fs.Label,
		})
	}
	return subs
}
