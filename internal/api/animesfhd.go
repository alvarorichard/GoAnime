package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/animesfhd"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
)

// GetAnimesFHDEpisodes selects an exact season/audio version before listing it.
func GetAnimesFHDEpisodes(ctx context.Context, anime *models.Anime) ([]models.Episode, error) {
	wantAudio := ""
	if audio, ok := pinnedStartFlixAudio(); ok {
		wantAudio = string(audio)
	}
	pick := func(labels []string) (int, error) {
		return tui.PickLabels(labels, tui.PickOptions{
			Breadcrumb: "AnimesFHD > Temporada e áudio", WindowTitle: "GoAnime - AnimesFHD",
			ItemSingular: "version", ItemPlural: "versions",
		})
	}
	return getAnimesFHDEpisodes(ctx, anime, animesfhd.NewAnimesFHDClient(), pick, wantAudio)
}

func getAnimesFHDEpisodes(ctx context.Context, anime *models.Anime, client *animesfhd.Client, pick func([]string) (int, error), wantAudio string) ([]models.Episode, error) {
	if anime == nil || anime.URL == "" {
		return nil, fmt.Errorf("no AnimesFHD page for this title")
	}
	requestCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	versions, err := client.GetVersionsContext(requestCtx, anime.URL)
	cancel()
	if err != nil {
		return nil, err
	}
	selected := animesfhd.Version{URL: anime.URL}
	if wantAudio != "" {
		filtered := make([]animesfhd.Version, 0, len(versions))
		for _, v := range versions {
			if strings.EqualFold(v.Audio, wantAudio) {
				filtered = append(filtered, v)
			}
		}
		// The pin is a preference, as on AnimesROLL and AnimesDigital: it
		// narrows the list when it can and never hides the only audio a
		// title has.
		if len(filtered) > 0 {
			versions = filtered
		}
	}
	if len(versions) > 0 {
		index := 0
		if len(versions) > 1 {
			labels := make([]string, len(versions))
			for i, v := range versions {
				labels[i] = v.Label
			}
			index, err = pick(labels)
			if err != nil {
				return nil, err
			}
		}
		if index < 0 || index >= len(versions) {
			return nil, fmt.Errorf("invalid AnimesFHD version selection")
		}
		selected = versions[index]
	}
	// User input consumes no HTTP budget. Keep the parent's cancellation while
	// giving the selected version a fresh, bounded pagination window.
	requestCtx, cancel = context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	episodes, err := client.GetAnimeEpisodesContext(requestCtx, selected.URL)
	if err != nil {
		return nil, err
	}
	for i := range episodes {
		if selected.Audio != "" {
			episodes[i].Audio = selected.Audio
		}
	}
	anime.URL = selected.URL
	anime.CurrentSeason = selected.Season
	return episodes, nil
}
