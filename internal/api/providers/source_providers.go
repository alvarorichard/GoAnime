package providers

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
)

type scraperLookup interface {
	GetScraper(scraperType scraper.ScraperType) (scraper.UnifiedScraper, error)
}

// EpisodeNumber extracts the episode number string from an Episode model.
// Returns "" if indeterminate — caller must decide how to handle.
func EpisodeNumber(ep *models.Episode) string {
	if ep == nil {
		return ""
	}
	if ep.Number != "" {
		return ep.Number
	}
	if ep.Num > 0 {
		return fmt.Sprintf("%d", ep.Num)
	}
	return ""
}

// --- AllAnime Provider ---

type allAnimeProvider struct {
	sm scraperLookup
}

func init() {
	RegisterProvider(source.AllAnime, func(sm *scraper.ScraperManager) Provider {
		return &allAnimeProvider{sm: sm}
	})
}

func (p *allAnimeProvider) Kind() source.SourceKind { return source.AllAnime }
func (p *allAnimeProvider) HasSeasons() bool        { return false }

func (p *allAnimeProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.sm.GetScraper(scraper.AllAnimeType)
	if err != nil {
		return nil, err
	}
	animeID := source.ExtractAllAnimeID(anime.URL)
	return adapter.GetAnimeEpisodes(animeID)
}

func (p *allAnimeProvider) FetchStreamURL(_ context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	adapter, err := p.sm.GetScraper(scraper.AllAnimeType)
	if err != nil {
		return "", err
	}
	animeID := source.ExtractAllAnimeID(anime.URL)
	epNum := EpisodeNumber(episode)
	if quality == "" {
		quality = "best"
	}
	url, _, err := adapter.GetStreamURL(animeID, epNum, quality)
	if err != nil {
		return "", fmt.Errorf("allAnime stream: %w", err)
	}
	return url, nil
}

// --- AnimeFire Provider ---

type animeFireProvider struct {
	sm scraperLookup
}

func init() {
	RegisterProvider(source.AnimeFire, func(sm *scraper.ScraperManager) Provider {
		return &animeFireProvider{sm: sm}
	})
}

func (p *animeFireProvider) Kind() source.SourceKind { return source.AnimeFire }
func (p *animeFireProvider) HasSeasons() bool        { return false }

func (p *animeFireProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.sm.GetScraper(scraper.AnimefireType)
	if err != nil {
		return nil, err
	}
	return adapter.GetAnimeEpisodes(anime.URL)
}

func (p *animeFireProvider) FetchStreamURL(_ context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	adapter, err := p.sm.GetScraper(scraper.AnimefireType)
	if err != nil {
		return "", err
	}
	url, _, err := adapter.GetStreamURL(episode.URL)
	if err != nil {
		return "", fmt.Errorf("animeFire stream: %w", err)
	}
	return url, nil
}

// --- FlixHQ Provider ---

type flixHQProvider struct {
	sm scraperLookup
}

func init() {
	RegisterProvider(source.FlixHQ, func(sm *scraper.ScraperManager) Provider {
		return &flixHQProvider{sm: sm}
	})
}

func (p *flixHQProvider) Kind() source.SourceKind { return source.FlixHQ }
func (p *flixHQProvider) HasSeasons() bool        { return true }

func (p *flixHQProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	mediaID := scraper.ExtractMediaID(anime.URL)
	if mediaID == "" {
		return nil, fmt.Errorf("could not extract media ID from URL: %s", anime.URL)
	}

	if anime.MediaType == models.MediaTypeMovie {
		return []models.Episode{{
			Number: "1",
			Num:    1,
			URL:    mediaID,
			Title: models.TitleDetails{
				English: anime.Name,
				Romaji:  anime.Name,
			},
		}}, nil
	}

	client, err := p.client()
	if err != nil {
		return nil, err
	}

	var seasons []scraper.FlixHQSeason
	var seasonsErr error
	tui.RunWithSpinner("Loading seasons...", func() {
		seasons, seasonsErr = client.GetSeasons(mediaID)
	})
	if seasonsErr != nil {
		return nil, fmt.Errorf("failed to get seasons: %w", seasonsErr)
	}
	if len(seasons) == 0 {
		return nil, fmt.Errorf("no seasons found for TV show")
	}

	seasonIdx, err := chooseIndex(seasons, func(i int) string {
		return seasons[i].Title
	}, "Select season: ")
	if err != nil {
		return nil, fmt.Errorf("season selection cancelled: %w", err)
	}

	selectedSeason := seasons[seasonIdx]
	anime.CurrentSeason = selectedSeason.Number

	var flixEpisodes []scraper.FlixHQEpisode
	var episodesErr error
	tui.RunWithSpinner("Loading episodes...", func() {
		flixEpisodes, episodesErr = client.GetEpisodes(selectedSeason.ID)
	})
	if episodesErr != nil {
		return nil, fmt.Errorf("failed to get episodes: %w", episodesErr)
	}

	episodes := make([]models.Episode, 0, len(flixEpisodes))
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

	return episodes, nil
}

func (p *flixHQProvider) FetchStreamURL(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	client, err := p.client()
	if err != nil {
		return "", err
	}

	if anime != nil && anime.URL != "" {
		client.SetMediaPath(scraper.ExtractMediaPath(anime.URL))
	}

	providerName := "Vidcloud"
	subsLanguage := preferredSubsLanguage()
	if quality == "" {
		quality = "auto"
	}

	var (
		episodeID  string
		embedLink  string
		streamInfo *scraper.FlixHQStreamInfo
		streamErr  error
	)

	tui.RunWithSpinner("Loading stream...", func() {
		if anime.MediaType == models.MediaTypeMovie {
			mediaID := episode.URL
			if mediaID == "" {
				mediaID = scraper.ExtractMediaID(anime.URL)
			}
			episodeID, streamErr = client.GetMovieServerID(mediaID, providerName)
		} else {
			dataID := episode.URL
			if dataID == "" {
				dataID = episode.DataID
			}
			episodeID, streamErr = client.GetEpisodeServerID(dataID, providerName)
		}
		if streamErr != nil {
			return
		}

		embedLink, streamErr = client.GetEmbedLink(episodeID)
		if streamErr != nil {
			return
		}

		reqCtx, cancel := contextWithTimeout(ctx, 60*time.Second)
		defer cancel()
		streamInfo, streamErr = client.ExtractStreamInfoWithContext(reqCtx, embedLink, quality, subsLanguage)
	})

	if streamErr != nil {
		return "", fmt.Errorf("flixHQ stream: %w", streamErr)
	}
	if streamInfo == nil {
		return "", fmt.Errorf("flixHQ stream: no stream info returned")
	}

	if len(streamInfo.Qualities) > 1 {
		selectedQuality, selectErr := selectFlixHQQualityOption(client, streamInfo.Qualities)
		if selectErr == nil && selectedQuality.URL != "" {
			streamInfo.VideoURL = selectedQuality.URL
			streamInfo.Quality = string(selectedQuality.Quality)
			streamInfo.IsM3U8 = selectedQuality.IsM3U8
		}
	}

	applyFlixHQPlaybackState(streamInfo)

	if streamInfo.VideoURL == "" {
		return "", fmt.Errorf("flixHQ stream: empty stream URL")
	}
	return streamInfo.VideoURL, nil
}

func (p *flixHQProvider) client() (*scraper.FlixHQClient, error) {
	if p.sm == nil {
		return scraper.NewFlixHQClient(), nil
	}

	adapter, err := p.sm.GetScraper(scraper.FlixHQType)
	if err != nil {
		return nil, err
	}

	if accessor, ok := adapter.(interface{ GetClient() *scraper.FlixHQClient }); ok {
		return accessor.GetClient(), nil
	}

	return scraper.NewFlixHQClient(), nil
}

type sflixProvider struct {
	sm scraperLookup
}

func init() {
	RegisterProvider(source.SFlix, func(sm *scraper.ScraperManager) Provider {
		return &sflixProvider{sm: sm}
	})
}

func (p *sflixProvider) Kind() source.SourceKind { return source.SFlix }
func (p *sflixProvider) HasSeasons() bool        { return true }

func (p *sflixProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	mediaID := scraper.ExtractMediaID(anime.URL)
	if mediaID == "" {
		return nil, fmt.Errorf("could not extract media ID from URL: %s", anime.URL)
	}

	if anime.MediaType == models.MediaTypeMovie {
		return []models.Episode{{
			Number: "1",
			Num:    1,
			URL:    mediaID,
			Title: models.TitleDetails{
				English: anime.Name,
				Romaji:  anime.Name,
			},
		}}, nil
	}

	client, err := p.client()
	if err != nil {
		return nil, err
	}

	var seasons []scraper.SFlixSeason
	var seasonsErr error
	tui.RunWithSpinner("Loading seasons...", func() {
		seasons, seasonsErr = client.GetSeasons(mediaID)
	})
	if seasonsErr != nil {
		return nil, fmt.Errorf("failed to get seasons: %w", seasonsErr)
	}
	if len(seasons) == 0 {
		return nil, fmt.Errorf("no seasons found for TV show")
	}

	seasonIdx, err := chooseIndex(seasons, func(i int) string {
		return seasons[i].Title
	}, "Select season: ")
	if err != nil {
		return nil, fmt.Errorf("season selection cancelled: %w", err)
	}

	selectedSeason := seasons[seasonIdx]
	anime.CurrentSeason = selectedSeason.Number

	var sflixEpisodes []scraper.SFlixEpisode
	var episodesErr error
	tui.RunWithSpinner("Loading episodes...", func() {
		sflixEpisodes, episodesErr = client.GetEpisodes(selectedSeason.ID)
	})
	if episodesErr != nil {
		return nil, fmt.Errorf("failed to get episodes: %w", episodesErr)
	}

	episodes := make([]models.Episode, 0, len(sflixEpisodes))
	for _, ep := range sflixEpisodes {
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

	return episodes, nil
}

func (p *sflixProvider) FetchStreamURL(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	client, err := p.client()
	if err != nil {
		return "", err
	}

	if anime != nil && anime.URL != "" {
		client.SetMediaPath(scraper.ExtractMediaPath(anime.URL))
	}

	providerName := "Vidcloud"
	subsLanguage := preferredSubsLanguage()
	if quality == "" {
		quality = "auto"
	}

	var (
		episodeID  string
		embedLink  string
		streamInfo *scraper.SFlixStreamInfo
		streamErr  error
	)

	tui.RunWithSpinner("Loading stream...", func() {
		if anime.MediaType == models.MediaTypeMovie {
			mediaID := episode.URL
			if mediaID == "" {
				mediaID = scraper.ExtractMediaID(anime.URL)
			}
			episodeID, streamErr = client.GetMovieServerID(mediaID, providerName)
		} else {
			dataID := episode.URL
			if dataID == "" {
				dataID = episode.DataID
			}
			episodeID, streamErr = client.GetEpisodeServerID(dataID, providerName)
		}
		if streamErr != nil {
			return
		}

		embedLink, streamErr = client.GetEmbedLink(episodeID)
		if streamErr != nil {
			return
		}

		reqCtx, cancel := contextWithTimeout(ctx, 60*time.Second)
		defer cancel()
		streamInfo, streamErr = client.ExtractStreamInfoWithContext(reqCtx, embedLink, quality, subsLanguage)
	})

	if streamErr != nil {
		return "", fmt.Errorf("sflix stream: %w", streamErr)
	}
	if streamInfo == nil {
		return "", fmt.Errorf("sflix stream: no stream info returned")
	}

	if len(streamInfo.Qualities) > 1 {
		selectedQuality, selectErr := selectSFlixQualityOption(client, streamInfo.Qualities)
		if selectErr == nil && selectedQuality.URL != "" {
			streamInfo.VideoURL = selectedQuality.URL
			streamInfo.Quality = string(selectedQuality.Quality)
			streamInfo.IsM3U8 = selectedQuality.IsM3U8
		}
	}

	applySFlixPlaybackState(streamInfo)

	if streamInfo.VideoURL == "" {
		return "", fmt.Errorf("sflix stream: empty stream URL")
	}
	return streamInfo.VideoURL, nil
}

func (p *sflixProvider) client() (*scraper.SFlixClient, error) {
	if p.sm == nil {
		return scraper.NewSFlixClient(), nil
	}

	adapter, err := p.sm.GetScraper(scraper.SFlixType)
	if err != nil {
		return nil, err
	}

	if accessor, ok := adapter.(interface{ GetClient() *scraper.SFlixClient }); ok {
		return accessor.GetClient(), nil
	}

	return scraper.NewSFlixClient(), nil
}

// --- NineAnime Provider ---

type nineAnimeProvider struct {
	sm scraperLookup
}

func init() {
	RegisterProvider(source.NineAnime, func(sm *scraper.ScraperManager) Provider {
		return &nineAnimeProvider{sm: sm}
	})
}

func (p *nineAnimeProvider) Kind() source.SourceKind { return source.NineAnime }
func (p *nineAnimeProvider) HasSeasons() bool        { return false }

func (p *nineAnimeProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.sm.GetScraper(scraper.NineAnimeType)
	if err != nil {
		return nil, err
	}
	return adapter.GetAnimeEpisodes(anime.URL)
}

func (p *nineAnimeProvider) FetchStreamURL(_ context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	adapter, err := p.sm.GetScraper(scraper.NineAnimeType)
	if err != nil {
		return "", err
	}
	url, metadata, err := adapter.GetStreamURL(episode.URL)
	if err != nil {
		return "", fmt.Errorf("9anime stream: %w", err)
	}
	applyNineAnimePlaybackMetadata(metadata)
	return url, nil
}

// --- SuperFlix Provider ---

type superFlixProvider struct {
	sm scraperLookup
}

func init() {
	RegisterProvider(source.SuperFlix, func(sm *scraper.ScraperManager) Provider {
		return &superFlixProvider{sm: sm}
	})
}

func (p *superFlixProvider) Kind() source.SourceKind { return source.SuperFlix }
func (p *superFlixProvider) HasSeasons() bool        { return true }

func (p *superFlixProvider) FetchEpisodes(ctx context.Context, anime *models.Anime) ([]models.Episode, error) {
	tmdbID := anime.URL
	if tmdbID == "" {
		return nil, fmt.Errorf("no TMDB ID found for SuperFlix content")
	}

	if anime.MediaType == models.MediaTypeMovie {
		return []models.Episode{{
			Number: "1",
			Num:    1,
			URL:    tmdbID,
			Title: models.TitleDetails{
				English: anime.Name,
				Romaji:  anime.Name,
			},
		}}, nil
	}

	client, err := p.client()
	if err != nil {
		return nil, err
	}

	var (
		allEpisodes map[string][]scraper.SuperFlixEpisode
		episodesErr error
	)
	tui.RunWithSpinner("Loading seasons...", func() {
		reqCtx, cancel := contextWithTimeout(ctx, 30*time.Second)
		defer cancel()
		allEpisodes, episodesErr = client.GetEpisodes(reqCtx, tmdbID)
	})
	if episodesErr != nil {
		return nil, fmt.Errorf("failed to get episodes: %w", episodesErr)
	}
	if len(allEpisodes) == 0 {
		return nil, fmt.Errorf("no seasons found")
	}

	seasonNums := make([]string, 0, len(allEpisodes))
	for seasonNum := range allEpisodes {
		seasonNums = append(seasonNums, seasonNum)
	}
	sort.Strings(seasonNums)

	seasonLabels := make([]string, 0, len(seasonNums))
	for _, seasonNum := range seasonNums {
		seasonLabels = append(seasonLabels, fmt.Sprintf("Season %s (%d episodes)", seasonNum, len(allEpisodes[seasonNum])))
	}

	seasonIdx, err := chooseIndex(seasonNums, func(i int) string {
		return seasonLabels[i]
	}, "Select season: ")
	if err != nil {
		return nil, fmt.Errorf("season selection cancelled: %w", err)
	}

	selectedSeason := seasonNums[seasonIdx]
	selectedEpisodes := allEpisodes[selectedSeason]

	episodes := make([]models.Episode, 0, len(selectedEpisodes))
	for _, ep := range selectedEpisodes {
		epNum := ep.EpiNum.String()
		num := 0
		if parsedNum, parseErr := ep.EpiNum.Int64(); parseErr == nil {
			num = int(parsedNum)
		}

		episodes = append(episodes, models.Episode{
			Number:   epNum,
			Num:      num,
			URL:      tmdbID,
			SeasonID: selectedSeason,
			Title: models.TitleDetails{
				English: ep.Title,
				Romaji:  ep.Title,
			},
			Aired: ep.AirDate,
		})
	}

	if _, parseErr := fmt.Sscanf(selectedSeason, "%d", &anime.CurrentSeason); parseErr != nil {
		anime.CurrentSeason = 0
	}

	return episodes, nil
}

func (p *superFlixProvider) FetchStreamURL(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	client, err := p.client()
	if err != nil {
		return "", err
	}

	tmdbID := episode.URL
	if tmdbID == "" {
		tmdbID = anime.URL
	}

	mediaType := "serie"
	season := episode.SeasonID
	epNum := EpisodeNumber(episode)
	if anime.MediaType == models.MediaTypeMovie {
		mediaType = "filme"
		season = ""
		epNum = ""
	} else if season == "" && anime.CurrentSeason > 0 {
		season = fmt.Sprintf("%d", anime.CurrentSeason)
	}

	var (
		result    *scraper.SuperFlixStreamResult
		streamErr error
	)
	tui.RunWithSpinner("Loading stream...", func() {
		reqCtx, cancel := contextWithTimeout(ctx, 60*time.Second)
		defer cancel()
		result, streamErr = client.GetStreamURL(reqCtx, mediaType, tmdbID, season, epNum)
	})
	if streamErr != nil {
		return "", fmt.Errorf("superFlix stream: %w", streamErr)
	}
	if result == nil || result.StreamURL == "" {
		return "", fmt.Errorf("superFlix stream: empty stream URL")
	}

	applySuperFlixPlaybackResult(anime, result)
	return result.StreamURL, nil
}

func (p *superFlixProvider) client() (*scraper.SuperFlixClient, error) {
	if p.sm == nil {
		return scraper.NewSuperFlixClient(), nil
	}

	adapter, err := p.sm.GetScraper(scraper.SuperFlixType)
	if err != nil {
		return nil, err
	}

	if accessor, ok := adapter.(interface {
		GetClient() *scraper.SuperFlixClient
	}); ok {
		return accessor.GetClient(), nil
	}

	return scraper.NewSuperFlixClient(), nil
}

// --- AnimeDrive Provider ---

type animeDriveProvider struct {
	sm scraperLookup
}

func init() {
	RegisterProvider(source.AnimeDrive, func(sm *scraper.ScraperManager) Provider {
		return &animeDriveProvider{sm: sm}
	})
}

func (p *animeDriveProvider) Kind() source.SourceKind { return source.AnimeDrive }
func (p *animeDriveProvider) HasSeasons() bool        { return false }

func (p *animeDriveProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.sm.GetScraper(scraper.AnimeDriveType)
	if err != nil {
		return nil, err
	}
	return adapter.GetAnimeEpisodes(anime.URL)
}

func (p *animeDriveProvider) FetchStreamURL(_ context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	adapter, err := p.sm.GetScraper(scraper.AnimeDriveType)
	if err != nil {
		return "", err
	}
	url, _, err := adapter.GetStreamURL(episode.URL, "auto")
	if err != nil {
		return "", fmt.Errorf("animeDrive stream: %w", err)
	}
	return url, nil
}

func applyNineAnimePlaybackMetadata(metadata map[string]string) {
	if len(metadata) == 0 {
		return
	}

	if referer := strings.TrimSpace(metadata["referer"]); referer != "" {
		util.SetGlobalReferer(referer)
	}

	if util.SubtitlesDisabled() {
		return
	}

	rawSubtitles := strings.TrimSpace(metadata["subtitles"])
	if rawSubtitles == "" {
		return
	}

	subURLs := strings.Split(rawSubtitles, ",")
	subLabels := strings.Split(metadata["subtitle_labels"], ",")

	subInfos := make([]util.SubtitleInfo, 0, len(subURLs))
	for i, rawURL := range subURLs {
		subURL := strings.TrimSpace(rawURL)
		if subURL == "" {
			continue
		}

		label := "Unknown"
		if i < len(subLabels) {
			if trimmed := strings.TrimSpace(subLabels[i]); trimmed != "" {
				label = trimmed
			}
		}

		subInfos = append(subInfos, util.SubtitleInfo{
			URL:      subURL,
			Language: nineAnimeSubtitleLanguage(label),
			Label:    label,
		})
	}

	if len(subInfos) > 0 {
		util.SetGlobalSubtitles(subInfos)
	}
}

func nineAnimeSubtitleLanguage(label string) string {
	lower := strings.ToLower(label)

	switch {
	case strings.Contains(lower, "english"):
		return "eng"
	case strings.Contains(lower, "portuguese"):
		return "por"
	case strings.Contains(lower, "spanish"):
		return "spa"
	case strings.Contains(lower, "japanese"):
		return "jpn"
	case strings.Contains(lower, "french"):
		return "fre"
	case strings.Contains(lower, "german"):
		return "ger"
	case strings.Contains(lower, "italian"):
		return "ita"
	case strings.Contains(lower, "arabic"):
		return "ara"
	default:
		return "unknown"
	}
}
