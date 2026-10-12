package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/animesdigital"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
)

var (
	animesDigitalClientFn = animesdigital.Shared
	animesDigitalPickFn   = func(labels []string) (int, error) {
		return tui.PickLabels(labels, tui.PickOptions{
			Breadcrumb:   "AnimesDigital > Você quer assistir legendado ou dublado?",
			WindowTitle:  "GoAnime - AnimesDigital",
			ItemSingular: "option",
			ItemPlural:   "options",
		})
	}
)

const animesDigitalListBudget = 90 * time.Second

// GetAnimesDigitalEpisodes selects the title page for the desired audio and
// then lists all of its paginated episodes. Search pairs the two title pages;
// a title with only one audio never opens the picker.
func GetAnimesDigitalEpisodes(media *models.Anime) ([]models.Episode, error) {
	if media == nil || strings.TrimSpace(media.URL) == "" {
		return nil, fmt.Errorf("no AnimesDigital page for this title")
	}
	variants := animesdigital.ParseVariants(media.URL)
	target := variants.Subtitled
	if target == "" {
		target = variants.Dubbed
	}

	if variants.Subtitled != "" && variants.Dubbed != "" {
		useDubbed := false
		if pinned, ok := pinnedStartFlixAudio(); ok {
			useDubbed = strings.EqualFold(string(pinned), "Dublado")
		} else if idx, err := animesDigitalPickFn([]string{"💬 Legendado", "🎙️  Dublado"}); err == nil {
			useDubbed = idx == 1
		} else {
			util.Debug("AnimesDigital: audio picker unavailable, taking subtitles", "err", err)
		}
		if useDubbed {
			target = variants.Dubbed
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), animesDigitalListBudget)
	defer cancel()
	return animesDigitalClientFn().GetAnimeEpisodesContext(ctx, target)
}
