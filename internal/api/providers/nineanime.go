package providers

import (
	"fmt"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
)

// NineAnimeProvider handles episode fetching and stream resolution for 9Anime sources.
type NineAnimeProvider struct{}

func NewNineAnimeProvider() *NineAnimeProvider {
	return &NineAnimeProvider{}
}

func (p *NineAnimeProvider) Name() string { return "9Anime" }

func (p *NineAnimeProvider) HasSeasons() bool { return false }

func (p *NineAnimeProvider) FetchEpisodes(anime *models.Anime) ([]models.Episode, error) {
	nineAnimeClient := scraper.NewNineAnimeClient()

	animeID := anime.URL
	util.Debug("Getting 9Anime episodes", "animeID", animeID)

	episodes, err := nineAnimeClient.GetAnimeEpisodes(animeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get episodes from 9Anime: %w", err)
	}

	util.Debug("9Anime episodes loaded", "count", len(episodes))
	return episodes, nil
}

func (p *NineAnimeProvider) GetStreamURL(episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	util.ClearGlobalSubtitles()
	util.SetGlobalAnimeSource("9Anime")

	nineAnimeClient := scraper.NewNineAnimeClient()

	episodeID := episode.DataID
	if episodeID == "" {
		episodeID = episode.URL
	}

	util.Debug("Getting 9Anime stream", "episodeID", episodeID, "quality", quality)

	streamURL, metadata, err := nineAnimeClient.GetStreamURL(episodeID, "sub")
	if err != nil {
		return "", fmt.Errorf("failed to get stream URL from 9Anime: %w", err)
	}

	if referer, ok := metadata["referer"]; ok && referer != "" {
		util.SetGlobalReferer(referer)
	}

	p.storeSubtitlesFromMetadata(metadata)

	util.Debug("9Anime stream URL obtained", "url", streamURL[:min(len(streamURL), 80)])
	return streamURL, nil
}

// storeSubtitlesFromMetadata parses subtitle information from 9Anime stream metadata
// and stores them globally for mpv playback.
func (p *NineAnimeProvider) storeSubtitlesFromMetadata(metadata map[string]string) {
	subtitleURLs, ok := metadata["subtitles"]
	if !ok || subtitleURLs == "" || util.GlobalNoSubs {
		return
	}

	subURLs := strings.Split(subtitleURLs, ",")
	var subLabels []string
	if labels, ok := metadata["subtitle_labels"]; ok {
		subLabels = strings.Split(labels, ",")
	}

	var subInfos []util.SubtitleInfo
	for i, subURL := range subURLs {
		label := "Unknown"
		lang := "unknown"
		if i < len(subLabels) {
			label = subLabels[i]
			lang = labelToLanguageCode(label)
		}
		subInfos = append(subInfos, util.SubtitleInfo{
			URL:      subURL,
			Language: lang,
			Label:    label,
		})
	}
	util.SetGlobalSubtitles(subInfos)
	util.Debug("9Anime subtitles loaded", "count", len(subInfos))
}

var languageMap = map[string]string{
	"english":    "eng",
	"portuguese": "por",
	"spanish":    "spa",
	"japanese":   "jpn",
	"french":     "fre",
	"german":     "ger",
	"italian":    "ita",
	"arabic":     "ara",
}

func labelToLanguageCode(label string) string {
	lower := strings.ToLower(label)
	for keyword, code := range languageMap {
		if strings.Contains(lower, keyword) {
			return code
		}
	}
	return "unknown"
}
