package providers

import (
	"context"
	"fmt"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// FetchEpisodes lists an anime's episodes through the Model B registry — the
// declarative replacement for api.GetAnimeEpisodesEnhanced's per-source switch.
// It resolves the source once (honoring the S1 kill-switch and best-effort
// fallback), normalizes anime.Source, then delegates to the resolved Source's
// FetchEpisodes.
//
// Behavior is equivalent to the legacy switch: AllAnime/AnimeFire/Goyabu list
// via their adapters; SuperFlix runs its season picker; an unrecognized source
// falls back to best-effort AllAnime (unless GOANIME_STRICT_SOURCE disables it).
func FetchEpisodes(ctx context.Context, anime *models.Anime) ([]models.Episode, error) {
	if anime == nil {
		return nil, fmt.Errorf("cannot fetch episodes for a nil anime")
	}

	src, resolved := source.Resolve(anime)
	if src == nil {
		if util.StrictSourceResolution() {
			return nil, fmt.Errorf("unrecognized source for %q (%s); best-effort disabled by GOANIME_STRICT_SOURCE", anime.Name, resolved.Reason)
		}
		best, ok := source.Enabled(resolved.BestEffortKind())
		if !ok {
			return nil, fmt.Errorf("no enabled source for %q (%s); it may be off via GOANIME_DISABLED_SOURCES", resolved.BestEffortKind(), resolved.Reason)
		}
		util.Warn("unrecognized source; listing episodes best-effort", "anime", anime.Name, "kind", best.Describe().Kind, "reason", resolved.Reason)
		src = best
	}

	// Canonicalize the source field to the dispatched kind when it was empty,
	// mirroring the legacy switch's detect-and-set behavior. An explicitly set
	// source is left untouched (e.g. the scraper's "Animefire.io" spelling).
	if anime.Source == "" {
		anime.Source = string(src.Describe().Kind)
	}

	util.Debug("Fetching episodes via registry", "kind", src.Describe().Kind, "reason", resolved.Reason)
	return src.FetchEpisodes(ctx, anime)
}
