package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/anroll"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

var (
	anrollClientFn = anroll.Shared
	anrollPickFn   = func(labels []string) (int, error) {
		return tui.PickLabels(labels, tui.PickOptions{
			Breadcrumb:   "AnimesROLL > Você quer assistir legendado ou dublado?",
			WindowTitle:  "GoAnime - AnimesROLL",
			ItemSingular: "option",
			ItemPlural:   "options",
		})
	}
)

const anrollListBudget = 90 * time.Second

// GetAnrollEpisodes selects the requested audio page and lists all episodes.
func GetAnrollEpisodes(media *models.Anime) ([]models.Episode, error) {
	if media == nil || strings.TrimSpace(media.URL) == "" {
		return nil, fmt.Errorf("no AnimesROLL page for this title")
	}
	variants := anroll.ParseVariants(media.URL)
	target := variants.Subtitled
	if target == "" {
		target = variants.Dubbed
	}
	if variants.Subtitled != "" && variants.Dubbed != "" {
		useDubbed := false
		if pinned, ok := pinnedStartFlixAudio(); ok {
			useDubbed = strings.EqualFold(string(pinned), "Dublado")
		} else if idx, err := anrollPickFn([]string{"Legendado", "Dublado"}); err == nil {
			useDubbed = idx == 1
		} else {
			util.Debug("AnimesROLL: audio picker unavailable, taking subtitles", "err", err)
		}
		if useDubbed {
			target = variants.Dubbed
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), anrollListBudget)
	defer cancel()
	return anrollClientFn().GetAnimeEpisodesContext(ctx, target)
}
