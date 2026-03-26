package providers

import (
	"fmt"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/manifoldco/promptui"
)

// FlixHQProvider handles episode fetching and stream resolution for FlixHQ movie/TV sources.
type FlixHQProvider struct{}

func NewFlixHQProvider() *FlixHQProvider {
	return &FlixHQProvider{}
}

func (p *FlixHQProvider) Name() string { return "FlixHQ" }

func (p *FlixHQProvider) HasSeasons() bool { return true }

func (p *FlixHQProvider) FetchEpisodes(anime *models.Anime) ([]models.Episode, error) {
	flixhqClient := scraper.NewFlixHQClient()

	mediaID := ExtractMediaIDFromURL(anime.URL)
	if mediaID == "" {
		return nil, fmt.Errorf("could not extract media ID from URL: %s", anime.URL)
	}

	util.Debug("Getting FlixHQ content", "mediaType", anime.MediaType, "mediaID", mediaID)

	if anime.MediaType == models.MediaTypeMovie {
		util.Debug("FlixHQ: Processing movie")
		return []models.Episode{
			{
				Number: "1",
				Num:    1,
				URL:    mediaID,
				Title: models.TitleDetails{
					English: anime.Name,
					Romaji:  anime.Name,
				},
			},
		}, nil
	}

	util.Debug("FlixHQ: Processing TV show, getting seasons")

	var seasons []scraper.FlixHQSeason
	var seasonsErr error
	util.RunWithSpinner("Loading seasons...", func() {
		seasons, seasonsErr = flixhqClient.GetSeasons(mediaID)
	})
	if seasonsErr != nil {
		return nil, fmt.Errorf("failed to get seasons: %w", seasonsErr)
	}
	if len(seasons) == 0 {
		return nil, fmt.Errorf("no seasons found for TV show")
	}

	seasonNames := make([]string, len(seasons))
	for i, s := range seasons {
		seasonNames[i] = s.Title
	}

	seasonIdx, err := fuzzyfinder.Find(
		seasonNames,
		func(i int) string { return seasonNames[i] },
		fuzzyfinder.WithPromptString("Select season: "),
	)
	if err != nil {
		return nil, fmt.Errorf("season selection cancelled: %w", err)
	}

	selectedSeason := seasons[seasonIdx]
	util.Debug("Selected season", "season", selectedSeason.Title, "id", selectedSeason.ID)

	fmt.Print("\033[2K\033[1A\033[2K\r")

	var flixEpisodes []scraper.FlixHQEpisode
	var episodesErr error
	util.RunWithSpinner("Loading episodes...", func() {
		flixEpisodes, episodesErr = flixhqClient.GetEpisodes(selectedSeason.ID)
	})
	if episodesErr != nil {
		return nil, fmt.Errorf("failed to get episodes: %w", episodesErr)
	}

	var episodes []models.Episode
	for _, ep := range flixEpisodes {
		episodes = append(episodes, models.Episode{
			Number: fmt.Sprintf("%d", ep.Number),
			Num:    ep.Number,
			URL:    ep.DataID,
			Title: models.TitleDetails{
				English: ep.Title,
				Romaji:  ep.Title,
			},
			DataID:   ep.DataID,
			SeasonID: selectedSeason.ID,
		})
	}

	util.Debug("FlixHQ episodes loaded", "count", len(episodes))
	return episodes, nil
}

func (p *FlixHQProvider) GetStreamURL(episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	util.ClearGlobalSubtitles()
	util.SetGlobalAnimeSource("FlixHQ")

	flixhqClient := scraper.NewFlixHQClient()
	if anime.URL != "" {
		flixhqClient.SetMediaPath(scraper.ExtractMediaPath(anime.URL))
	}
	provider := "Vidcloud"
	subsLanguage := util.GlobalSubsLanguage
	if subsLanguage == "" {
		subsLanguage = "english"
	}

	streamInfo, err := p.resolveStream(flixhqClient, anime, episode, provider, subsLanguage)
	if err != nil {
		return "", err
	}

	p.storeSubtitlesGlobally(streamInfo)

	if streamInfo.Referer != "" {
		util.SetGlobalReferer(streamInfo.Referer)
	}

	return streamInfo.VideoURL, nil
}

// --- Stream resolution (encapsulates the server ID -> embed -> extract -> quality pipeline) ---

func (p *FlixHQProvider) resolveStream(
	client *scraper.FlixHQClient,
	anime *models.Anime,
	episode *models.Episode,
	provider, subsLanguage string,
) (*scraper.FlixHQStreamInfo, error) {

	var streamInfo *scraper.FlixHQStreamInfo
	var episodeID string
	var embedLink string
	var streamErr error

	if anime.MediaType == models.MediaTypeMovie {
		mediaID := episode.URL
		util.Debug("Getting movie stream", "mediaID", mediaID)

		util.RunWithSpinner("Loading movie stream...", func() {
			episodeID, streamErr = client.GetMovieServerID(mediaID, provider)
			if streamErr != nil {
				return
			}
			embedLink, streamErr = client.GetEmbedLink(episodeID)
			if streamErr != nil {
				return
			}
			streamInfo, streamErr = client.ExtractStreamInfo(embedLink, "auto", subsLanguage)
		})

		if streamErr != nil {
			return nil, fmt.Errorf("failed to get movie stream: %w", streamErr)
		}
		if streamInfo == nil {
			return nil, fmt.Errorf("failed to get movie stream: no stream info returned")
		}
	} else {
		dataID := episode.URL
		util.Debug("Getting TV episode stream", "dataID", dataID)

		util.RunWithSpinner("Loading episode stream...", func() {
			episodeID, streamErr = client.GetEpisodeServerID(dataID, provider)
			if streamErr != nil {
				return
			}
			embedLink, streamErr = client.GetEmbedLink(episodeID)
			if streamErr != nil {
				return
			}
			streamInfo, streamErr = client.ExtractStreamInfo(embedLink, "auto", subsLanguage)
		})

		if streamErr != nil {
			return nil, fmt.Errorf("failed to get episode stream: %w", streamErr)
		}
		if streamInfo == nil {
			return nil, fmt.Errorf("failed to get episode stream: no stream info returned")
		}
	}

	if len(streamInfo.Qualities) > 1 {
		selectedQuality, selectErr := p.selectQualityOptions(streamInfo.Qualities)
		if selectErr == nil && selectedQuality.URL != "" {
			streamInfo.VideoURL = selectedQuality.URL
			streamInfo.Quality = string(selectedQuality.Quality)
			streamInfo.IsM3U8 = selectedQuality.IsM3U8
		}
	}

	return streamInfo, nil
}

func (p *FlixHQProvider) selectQualityOptions(qualities []scraper.FlixHQQualityOption) (scraper.FlixHQQualityOption, error) {
	if len(qualities) == 0 {
		return scraper.FlixHQQualityOption{Quality: scraper.QualityAuto}, fmt.Errorf("no qualities available")
	}
	if len(qualities) == 1 {
		return qualities[0], nil
	}

	client := scraper.NewFlixHQClient()
	var items []string
	for _, q := range qualities {
		items = append(items, client.QualityToLabel(q.Quality))
	}

	prompt := promptui.Select{
		Label: "Select video quality",
		Items: items,
		Size:  10,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return qualities[0], err
	}
	return qualities[idx], nil
}

func (p *FlixHQProvider) storeSubtitlesGlobally(streamInfo *scraper.FlixHQStreamInfo) {
	if len(streamInfo.Subtitles) == 0 || util.GlobalNoSubs {
		return
	}
	var subInfos []util.SubtitleInfo
	for _, sub := range streamInfo.Subtitles {
		subInfos = append(subInfos, util.SubtitleInfo{
			URL:      sub.URL,
			Language: sub.Language,
			Label:    sub.Label,
		})
	}
	util.SetGlobalSubtitles(subInfos)
}

// ExtractMediaIDFromURL extracts the media ID from a FlixHQ URL.
// URL format: https://flixhq.to/movie/watch-movie-name-12345 or /movie/watch-movie-name-12345
func ExtractMediaIDFromURL(urlStr string) string {
	parts := strings.Split(urlStr, "-")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
