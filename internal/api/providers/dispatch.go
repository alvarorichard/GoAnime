package providers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

// searchBreaker is the registry-wide per-source circuit breaker for search.
// It replaces the ScraperManager's breaker (which was keyed by ScraperType);
// here it is keyed by SourceKind. Package-level so its state persists across
// searches within a process.
var searchBreaker = netx.NewCircuitBreaker()

// Aggregate search timing. searchAllTimeout is the hard ceiling for the whole
// fan-out; stragglerGrace bounds how long we keep waiting for the remaining
// sources once the first results are in — so a fast source isn't held hostage
// by a slow one. Mirrors the ScraperManager engine's budgets.
const (
	searchAllTimeout = 15 * time.Second
	stragglerGrace   = 5 * time.Second
	// originProbeBudget bounds the disambiguation HEAD issued after a per-source
	// search deadline (see searchOneWithTimeout / netx.EnrichTimeoutWithProbe).
	originProbeBudget = 3 * time.Second
)

// perSourceSearchTimeout caps a single source's search so a wedged adapter
// (the underlying clients still issue non-cancelable http.NewRequest calls)
// cannot hold the fan-out open until searchAllTimeout and, more importantly,
// trips that source's circuit breaker with an accurate, probe-enriched
// diagnostic instead of a generic aggregate timeout. A var, not a const, so
// tests can shorten it.
var perSourceSearchTimeout = 12 * time.Second

// searchOne is a single source's search outcome in the fan-out.
type searchOne struct {
	kind    source.SourceKind
	results []*models.Anime
	err     error
}

// activeSearcher is a source selected for the fan-out: its Searchable behavior,
// its kind (for breaker keying and display), and its optional homepage probe
// URL (empty for opaque/browser-gated sources).
type activeSearcher struct {
	sr       source.Searchable
	kind     source.SourceKind
	probeURL string
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

	var searchers []activeSearcher
	for _, s := range source.ActiveSources() {
		sr, ok := s.(source.Searchable)
		if !ok {
			continue
		}
		desc := s.Describe()
		kind := desc.Kind
		if len(want) > 0 && !want[kind] {
			continue
		}
		// Skip a source whose circuit breaker is open (R5): it has been
		// failing, so don't hammer it — it auto-recovers after the cooldown.
		if diag, retry, open := searchBreaker.OpenDiagnostic(string(kind), sourceDisplayName(kind)); open {
			util.Warn("search source skipped (circuit open)", "source", kind, "retry_after", retry.Round(time.Second), "diagnostic", diag.UserMessage())
			continue
		}
		searchers = append(searchers, activeSearcher{sr: sr, kind: kind, probeURL: desc.ProbeURL})
	}
	if len(searchers) == 0 {
		return nil, fmt.Errorf("no searchable source available for query %q", query)
	}

	ctx, cancel := context.WithTimeout(ctx, searchAllTimeout)
	defer cancel()

	resultChan := make(chan searchOne, len(searchers))
	var wg sync.WaitGroup
	for _, a := range searchers {
		wg.Add(1)
		go func(a activeSearcher) {
			defer wg.Done()
			resultChan <- searchOneWithTimeout(ctx, a, query)
		}(a)
	}
	go func() { wg.Wait(); close(resultChan) }()

	var (
		all        []*models.Anime
		failures   []SourceFailure
		graceTimer <-chan time.Time
	)
	for {
		select {
		case res, ok := <-resultChan:
			if !ok {
				return finishSearch(query, all, failures)
			}
			if res.err != nil {
				// Feed the breaker so a repeatedly-failing source opens.
				diag := netx.DiagnoseError(sourceDisplayName(res.kind), "search", res.err)
				if searchBreaker.RecordFailure(string(res.kind), diag) {
					util.Warn("search source circuit opened", "source", res.kind, "diagnostic", diag.UserMessage())
				}
				util.Debug("search source failed", "source", res.kind, "error", res.err)
				// Record the failure as DATA. The raw error stays attached for
				// errors.Is and the debug log; the short reason is what a person
				// sees. Flattening both into one string is how the message grew
				// into three repetitions of the same fact.
				reason, limited := describeFailure(diag, res.err)
				failures = append(failures, SourceFailure{
					Kind:        res.kind,
					Reason:      reason,
					RateLimited: limited,
					Err:         fmt.Errorf("%s: %w", res.kind, res.err),
				})
				continue
			}
			searchBreaker.RecordSuccess(string(res.kind))
			if len(res.results) > 0 {
				all = append(all, res.results...)
				util.Debug("search results received", "source", res.kind, "count", len(res.results))
				if graceTimer == nil {
					graceTimer = time.After(stragglerGrace)
				}
			}
		case <-graceTimer:
			util.Debug("straggler grace elapsed; returning collected search results")
			return finishSearch(query, all, failures)
		case <-ctx.Done():
			util.Debug("search timeout reached; returning collected results")
			return finishSearch(query, all, failures)
		}
	}
}

// searchOneWithTimeout runs a single source's Search under its own deadline,
// derived from the fan-out context so the aggregate ceiling still applies.
//
// The underlying adapter clients still issue non-cancelable http.NewRequest
// calls, so a wedged source's goroutine may outlive this call — we abandon it
// (the buffered result channel absorbs a late send) rather than block. On
// timeout we synthesize a per-source error and, when the source exposes a
// homepage ProbeURL, upgrade it via netx.EnrichTimeoutWithProbe: a quick HEAD
// distinguishes "the site's origin is down (5xx / Cloudflare)" from an opaque
// hang, so the breaker opens with an actionable diagnostic.
func searchOneWithTimeout(parent context.Context, a activeSearcher, query string) searchOne {
	sctx, cancel := context.WithTimeout(parent, perSourceSearchTimeout)
	defer cancel()

	type outcome struct {
		results []*models.Anime
		err     error
	}
	done := make(chan outcome, 1)
	go func() {
		res, err := a.sr.Search(sctx, query)
		done <- outcome{results: res, err: err}
	}()

	select {
	case o := <-done:
		return searchOne{kind: a.kind, results: o.results, err: o.err}
	case <-sctx.Done():
		base := fmt.Errorf("%s search timed out after %v", sourceDisplayName(a.kind), perSourceSearchTimeout)
		err := netx.EnrichTimeoutWithProbe(parent, sourceDisplayName(a.kind), "search", a.probeURL, base, originProbeBudget)
		return searchOne{kind: a.kind, err: err}
	}
}

// reportPartialFailure says which sources did not answer a search that still
// returned something.
//
// A partial failure used to be invisible: the results were handed back and the
// per-source reasons dropped on the floor, leaving them only in the debug log.
// So a search where three of four sources were broken — AnimeFire's site
// rewritten out from under its parser, SuperFlix rate limiting the network,
// AniDB down — looked exactly like a search that had only ever had one source,
// and was reported as "it is only searching Goyabu".
//
// Warn, not Error: the search worked. But a user comparing what they got
// against what they expected deserves to know the catalogue was smaller than
// usual, and why.
func reportPartialFailure(failures []SourceFailure) {
	if len(failures) == 0 {
		return
	}
	util.Warnf("%d source(s) did not answer; results may be incomplete:", len(failures))
	for _, f := range failures {
		util.Warnf("  %s %s", f.Kind, f.Reason)
	}
}

// finishSearch turns the fan-out's outcome into one result or one error.
//
// When everything failed, the message leads with the per-source DIAGNOSTICS
// rather than the raw errors. A user staring at
//
//	no results for "matrix" (all sources failed): SuperFlix: server returned:
//	429 Too Many Requests
//
// cannot tell whether GoAnime is broken, the title does not exist, or the host
// is refusing them — and the third is the only one they can do something about
// (wait). The raw errors stay in the chain for errors.Is/As and the debug log.
func finishSearch(query string, all []*models.Anime, failures []SourceFailure) ([]*models.Anime, error) {
	if len(all) > 0 {
		reportPartialFailure(failures)
		return all, nil
	}
	if len(failures) == 0 {
		return nil, fmt.Errorf("no results found for: %s", query)
	}
	return nil, &SearchFailure{Query: query, Sources: failures}
}

// FetchEpisodes lists an anime's episodes through the Model B registry — the
// declarative replacement for api.GetAnimeEpisodesEnhanced's per-source switch.
// It resolves the source once (honoring the S1 kill-switch and best-effort
// fallback), normalizes anime.Source, then delegates to the resolved Source's
// FetchEpisodes.
//
// Behavior is equivalent to the legacy switch: HiAnime/AnimeFire/Goyabu list
// via their adapters; SuperFlix runs its season picker; an unrecognized source
// reports Unknown rather than guessing at a source.
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
