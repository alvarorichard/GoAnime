// Package api provides enhanced anime search and streaming capabilities
package api

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
)

// ErrBackToSearch is returned when user selects the back option to search again
var ErrBackToSearch = errors.New("back to search requested")

var searchSourceTypes = map[string]struct {
	scraperType scraper.ScraperType
	label       string
}{
	"allanime":   {scraperType: scraper.AllAnimeType, label: "AllAnime"},
	"animefire":  {scraperType: scraper.AnimefireType, label: "AnimeFire"},
	"animedrive": {scraperType: scraper.AnimeDriveType, label: "AnimeDrive"},
	"flixhq":     {scraperType: scraper.FlixHQType, label: "FlixHQ"},
	"movie":      {scraperType: scraper.FlixHQType, label: "FlixHQ"},
	"tv":         {scraperType: scraper.FlixHQType, label: "FlixHQ"},
	"9anime":     {scraperType: scraper.NineAnimeType, label: "9Anime"},
	"nineanime":  {scraperType: scraper.NineAnimeType, label: "9Anime"},
	"goyabu":     {scraperType: scraper.GoyabuType, label: "Goyabu"},
	"superflix":  {scraperType: scraper.SuperFlixType, label: "SuperFlix"},
}

// SearchAnimeEnhanced searches across one or more sources and returns a user-selected result.
func SearchAnimeEnhanced(name string, source string) (*models.Anime, error) {
	scraperManager := scraper.NewScraperManager()

	var scraperType *scraper.ScraperType
	isPTBR := false

	normalizedSource := strings.ToLower(strings.TrimSpace(source))

	switch {
	case normalizedSource == "ptbr" || normalizedSource == "pt-br":
		// Search only PT-BR sources (AnimeFire + Goyabu + SuperFlix) via dedicated method.
		isPTBR = true
		util.Debug("Searching all PT-BR sources (AnimeFire + Goyabu + SuperFlix)")
	case normalizedSource == "":
		// Default behavior: search all sources simultaneously (including FlixHQ).
		scraperType = nil
		util.Debug("Searching all sources", "query", name)
	default:
		if searchSource, ok := searchSourceTypes[normalizedSource]; ok {
			t := searchSource.scraperType
			scraperType = &t
			util.Debug("Searching specific source", "source", searchSource.label)
		} else {
			scraperType = nil
			util.Debug("Searching all sources", "query", name, "requested_source", source)
		}
	}

	util.Debug("Searching for anime/media", "query", name)
	var animes []*models.Anime
	var searchErr error
	tui.RunWithSpinner("Searching for anime...", func() {
		if isPTBR {
			animes, searchErr = scraperManager.SearchAnimePTBR(name)
		} else {
			animes, searchErr = scraperManager.SearchAnime(name, scraperType)
		}
	})
	if searchErr != nil {
		return nil, fmt.Errorf("failed to search: %w", searchErr)
	}

	if len(animes) == 0 {
		return nil, fmt.Errorf("no results found for: %s", name)
	}

	// Normalize source identification once so downstream flows share the same rules.
	for _, anime := range animes {
		if resolved, err := ResolveSource(anime); err == nil {
			resolved.Apply(anime)
		}
	}

	util.Debug("Search results summary", "total", len(animes))

	// Show source breakdown in debug only, without hardcoding every source.
	sourceCounts := make(map[string]int)
	for _, anime := range animes {
		resolved, err := ResolveSource(anime)
		if err != nil {
			continue
		}

		sourceCounts[resolved.Name]++
	}
	if len(sourceCounts) > 0 {
		sourceNames := make([]string, 0, len(sourceCounts))
		for sourceName := range sourceCounts {
			sourceNames = append(sourceNames, sourceName)
		}
		sort.Strings(sourceNames)
		for _, sourceName := range sourceNames {
			util.Debug("Source breakdown", "source", sourceName, "count", sourceCounts[sourceName])
		}
	}

	// Sort results by language priority: Portuguese first, then Multilanguage, Movies/TV, English, others.
	sort.SliceStable(animes, func(i, j int) bool {
		return languagePriority(animes[i].Name) < languagePriority(animes[j].Name)
	})

	backOption := &models.Anime{
		Name:   "← Back",
		URL:    "__back__",
		Source: "__back__",
	}

	animesWithBack := make([]*models.Anime, 0, len(animes)+1)
	animesWithBack = append(animesWithBack, backOption)
	animesWithBack = append(animesWithBack, animes...)

	var idx int
	var err error

	if util.IsDebug {
		idx, err = tui.Find(
			animesWithBack,
			func(i int) string {
				a := animesWithBack[i]
				displayName := a.Name
				if a.Year != "" && !strings.Contains(displayName, "("+a.Year+")") {
					displayName += " (" + a.Year + ")"
				}
				return displayName
			},
			fuzzyfinder.WithPromptString("Select the anime you want: "),
			fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
				if i < 0 || i >= len(animesWithBack) {
					return ""
				}
				anime := animesWithBack[i]
				if anime.Source == "__back__" {
					return "Go back to perform a new search"
				}

				preview := "Source: " + anime.Source + "\nURL: " + anime.URL
				if anime.ImageURL != "" {
					preview += "\nImage: " + anime.ImageURL
				}
				return preview
			}),
		)
	} else {
		idx, err = tui.Find(
			animesWithBack,
			func(i int) string {
				a := animesWithBack[i]
				displayName := a.Name
				if a.Year != "" && !strings.Contains(displayName, "("+a.Year+")") {
					displayName += " (" + a.Year + ")"
				}
				return displayName
			},
			fuzzyfinder.WithPromptString("Select the anime you want: "),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("anime selection cancelled: %w", err)
	}

	selectedAnime := animesWithBack[idx]
	if selectedAnime.Source == "__back__" {
		return nil, ErrBackToSearch
	}
	util.Debug("Anime selected", "name", selectedAnime.Name, "source", selectedAnime.Source)

	// Enrich with AniList data for images and metadata.
	if err := enrichAnimeData(selectedAnime); err != nil {
		util.Errorf("Error enriching anime data: %v", err)
	}

	return selectedAnime, nil
}

// GetAnimeEpisodesEnhanced fetches episodes using the provider-backed orchestration.
func GetAnimeEpisodesEnhanced(anime *models.Anime) ([]models.Episode, error) {
	resolved, resolveErr := ResolveSource(anime)
	if resolveErr != nil {
		return nil, resolveErr
	}

	return getEpisodesByResolvedSource(anime, resolved)
}

// GetEpisodeStreamURL resolves a stream URL using the provider-backed orchestration.
func GetEpisodeStreamURL(episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	resolved, resolveErr := ResolveSource(anime)
	if resolveErr != nil {
		return "", resolveErr
	}

	return getStreamURLByResolvedSource(anime, episode, quality, resolved)
}

// languagePriority returns a sort key for language-based ordering.
// Lower values sort first: Portuguese -> Multilanguage -> English -> Movies/TV -> Unknown.
func languagePriority(name string) int {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "[pt-br]") || strings.Contains(lower, "[portuguese]") || strings.Contains(lower, "[português]") {
		return 0
	}
	switch {
	case strings.HasPrefix(lower, "[multilanguage]"):
		return 1
	case strings.HasPrefix(lower, "[english]"):
		return 2
	case strings.HasPrefix(lower, "[movie]") || strings.HasPrefix(lower, "[tv]") || strings.HasPrefix(lower, "[movies/tv]"):
		return 3
	default:
		return 4
	}
}
