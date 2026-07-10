package providers

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// Aggregate search timing. searchAllTimeout is the hard ceiling for the whole
// fan-out; stragglerGrace bounds how long we keep waiting for the remaining
// sources once the first results are in — so a fast source isn't held hostage
// by a slow one. Mirrors the ScraperManager engine's budgets. Per-source
// breaker/timeout/tagging is handled inside each Source.Search (searchViaManager).
const (
	searchAllTimeout = 15 * time.Second
	stragglerGrace   = 5 * time.Second
)

// searchOne is a single source's search outcome in the fan-out.
type searchOne struct {
	kind    source.SourceKind
	results []*models.Anime
	err     error
}

// init wires SearchAll into the api package's search seam so
// api.SearchAnimeEnhanced dispatches through the Model B registry without
// importing providers (which would cycle).
func init() {
	api.SetSearchFetch(func(ctx context.Context, query string, kinds []source.SourceKind) ([]*models.Anime, error) {
		return SearchAll(ctx, query, kinds...)
	})
	api.SetEpisodesFetch(func(anime *models.Anime) ([]models.Episode, error) {
		return FetchEpisodes(context.Background(), anime)
	})
	api.SetStreamFetch(func(episode *models.Episode, anime *models.Anime, quality string) (string, error) {
		return FetchStreamURL(context.Background(), episode, anime, quality)
	})
}

// SearchAll fans out a query across the enabled, Searchable sources in the
// Model B registry — the registry-driven replacement for the ScraperManager's
// hardcoded adapter fan-out. When kinds is non-empty, only those sources are
// queried (specific-source search); otherwise every Searchable source runs.
//
// Each source's Search already applies the circuit breaker, kill-switch and
// language tagging (via searchViaManager); SearchAll adds only the concurrent
// collection with a straggler grace window so slow sources don't block fast
// ones. Per-source failures are logged and tolerated — a result is returned as
// long as at least one source succeeds.
func SearchAll(ctx context.Context, query string, kinds ...source.SourceKind) ([]*models.Anime, error) {
	want := map[source.SourceKind]bool{}
	for _, k := range kinds {
		want[k] = true
	}

	var searchers []source.Searchable
	var names []source.SourceKind
	for _, s := range source.ActiveSources() {
		sr, ok := s.(source.Searchable)
		if !ok {
			continue
		}
		kind := s.Describe().Kind
		if len(want) > 0 && !want[kind] {
			continue
		}
		searchers = append(searchers, sr)
		names = append(names, kind)
	}
	if len(searchers) == 0 {
		return nil, fmt.Errorf("no searchable source available for query %q", query)
	}

	ctx, cancel := context.WithTimeout(ctx, searchAllTimeout)
	defer cancel()

	resultChan := make(chan searchOne, len(searchers))
	var wg sync.WaitGroup
	for i, sr := range searchers {
		wg.Add(1)
		go func(sr source.Searchable, kind source.SourceKind) {
			defer wg.Done()
			res, err := sr.Search(ctx, query)
			resultChan <- searchOne{kind: kind, results: res, err: err}
		}(sr, names[i])
	}
	go func() { wg.Wait(); close(resultChan) }()

	var (
		all        []*models.Anime
		errs       []error
		graceTimer <-chan time.Time
	)
	for {
		select {
		case res, ok := <-resultChan:
			if !ok {
				return finishSearch(query, all, errs)
			}
			if res.err != nil {
				util.Debug("search source failed", "source", res.kind, "error", res.err)
				errs = append(errs, fmt.Errorf("%s: %w", res.kind, res.err))
				continue
			}
			if len(res.results) > 0 {
				all = append(all, res.results...)
				util.Debug("search results received", "source", res.kind, "count", len(res.results))
				if graceTimer == nil {
					graceTimer = time.After(stragglerGrace)
				}
			}
		case <-graceTimer:
			util.Debug("straggler grace elapsed; returning collected search results")
			return finishSearch(query, all, errs)
		case <-ctx.Done():
			util.Debug("search timeout reached; returning collected results")
			return finishSearch(query, all, errs)
		}
	}
}

func finishSearch(query string, all []*models.Anime, errs []error) ([]*models.Anime, error) {
	if len(all) == 0 {
		if len(errs) > 0 {
			return nil, fmt.Errorf("no results for %q (all sources failed): %w", query, errors.Join(errs...))
		}
		return nil, fmt.Errorf("no results found for: %s", query)
	}
	return all, nil
}

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

// FetchStreamURL resolves an episode's stream URL through the Model B registry
// — the declarative replacement for api.GetEpisodeStreamURL's per-source
// switch. Same resolution semantics as FetchEpisodes (kill-switch + best-effort
// fallback), then delegates to the resolved Source's FetchStreamURL, which
// warms up a browser-gated source first.
func FetchStreamURL(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	if anime == nil {
		return "", fmt.Errorf("cannot fetch stream for a nil anime")
	}

	src, resolved := source.Resolve(anime)
	if src == nil {
		if util.StrictSourceResolution() {
			return "", fmt.Errorf("unrecognized source for %q (%s); best-effort disabled by GOANIME_STRICT_SOURCE", anime.Name, resolved.Reason)
		}
		best, ok := source.Enabled(resolved.BestEffortKind())
		if !ok {
			return "", fmt.Errorf("no enabled source for %q (%s); it may be off via GOANIME_DISABLED_SOURCES", resolved.BestEffortKind(), resolved.Reason)
		}
		util.Warn("unrecognized source; resolving stream best-effort", "anime", anime.Name, "kind", best.Describe().Kind, "reason", resolved.Reason)
		src = best
	}

	if err := source.WarmUp(ctx, src); err != nil {
		return "", err
	}

	util.Debug("Fetching stream via registry", "kind", src.Describe().Kind, "reason", resolved.Reason)
	return src.FetchStreamURL(ctx, episode, anime, quality)
}
