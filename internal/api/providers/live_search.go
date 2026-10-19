package providers

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

// panelSourceName is how a source reads on the SEARCHING screen's panel:
// "AnimeFire", not the "Animefire.io" spelling stamped onto results.
func panelSourceName(kind source.SourceKind) string {
	if kind == source.AnimeFire {
		return "AnimeFire"
	}
	return sourceDisplayName(kind)
}

// planLiveSearch splits the wanted sources into the fan-out runners and the
// ones sitting out, with the reason each sitter shows on the SEARCHING
// screen's SOURCES panel (tela 2).
func planLiveSearch(kinds []source.SourceKind) (searchers []activeSearcher, off []models.SearchOffEntry, searched []string) {
	wanted := kinds
	if len(wanted) == 0 {
		wanted = source.Kinds()
	}
	for _, kind := range wanted {
		s, ok := source.Registered(kind)
		if !ok {
			continue
		}
		desc := s.Describe()
		display := panelSourceName(kind)
		switch {
		case desc.DefaultDisabled && !util.SourceForceEnabled(string(kind)):
			off = append(off, models.SearchOffEntry{Name: display, Reason: "off by default"})
		case !source.IsEnabled(desc):
			off = append(off, models.SearchOffEntry{Name: display, Reason: "disabled"})
		default:
			sr, ok := s.(source.Searchable)
			if !ok {
				continue
			}
			// Skip a source whose circuit breaker is open (R5), like
			// SearchAll does, but say so on the panel instead of only the log.
			if diag, retry, open := searchBreaker.OpenDiagnostic(string(kind), sourceDisplayName(kind)); open {
				util.Warn("search source skipped (circuit open)", "source", kind, "retry_after", retry.Round(time.Second), "diagnostic", diag.UserMessage())
				off = append(off, models.SearchOffEntry{Name: display, Reason: "cooling down"})
				continue
			}
			searchers = append(searchers, activeSearcher{sr: sr, kind: kind, probeURL: desc.ProbeURL})
			searched = append(searched, display)
		}
	}
	return searchers, off, searched
}

// SearchLive fans out like SearchAll but streams the fan-out to the SEARCHING
// screen: the plan first, then one outcome per source as it answers. Unlike
// SearchAll it waits for every source (up to the 15s ceiling) instead of
// leaving on the straggler grace: the screen, not the clock, decides when
// partial results are enough (enter opens one now).
func SearchLive(ctx context.Context, query string, kinds []source.SourceKind, emit func(models.SearchEvent)) ([]*models.Anime, error) {
	query = normalizeSearchQuery(query)
	searchers, off, searched := planLiveSearch(kinds)
	emit(models.SearchEvent{Plan: true, Searched: searched, Off: off, GiveUp: searchAllTimeout})
	if len(searchers) == 0 {
		return nil, fmt.Errorf("no searchable source available for query %q", query)
	}

	ctx, cancel := context.WithTimeout(ctx, searchAllTimeout)
	defer cancel()

	resultChan := make(chan searchOne, len(searchers))
	var wg sync.WaitGroup
	start := time.Now()
	for _, a := range searchers {
		wg.Add(1)
		go func(a activeSearcher) {
			defer wg.Done()
			resultChan <- searchOneWithTimeout(ctx, a, query)
		}(a)
	}
	go func() { wg.Wait(); close(resultChan) }()

	var (
		all      []*models.Anime
		failures []SourceFailure
	)
	pending := make(map[source.SourceKind]bool, len(searchers))
	for _, a := range searchers {
		pending[a.kind] = true
	}
	for {
		select {
		case res, ok := <-resultChan:
			if !ok {
				return finishSearch(query, searchers, off, all, failures)
			}
			delete(pending, res.kind)
			display := panelSourceName(res.kind)
			if res.err != nil {
				if f, ok := parentContextFailure(ctx, res.kind, res.err); ok {
					failures = append(failures, f)
					emit(models.SearchEvent{
						Name:    display,
						State:   models.SearchFailed,
						Elapsed: time.Since(start),
						Reason:  f.Reason,
					})
					continue
				}
				diag := netx.DiagnoseError(display, "search", res.err)
				if searchBreaker.RecordFailure(string(res.kind), diag) {
					util.Warn("search source circuit opened", "source", res.kind, "diagnostic", diag.UserMessage())
				}
				util.Debug("search source failed", "source", res.kind, "error", res.err)
				reason, limited := describeFailure(diag)
				failures = append(failures, SourceFailure{
					Kind:        res.kind,
					Reason:      reason,
					RateLimited: limited,
					Err:         fmt.Errorf("%s: %w", res.kind, res.err),
				})
				emit(models.SearchEvent{
					Name: display, State: models.SearchFailed,
					Elapsed: time.Since(start), Reason: strings.TrimPrefix(reason, "is "),
				})
				continue
			}
			searchBreaker.RecordSuccess(string(res.kind))
			if len(res.results) > 0 {
				all = append(all, res.results...)
				util.Debug("search results received", "source", res.kind, "count", len(res.results))
			}
			emit(models.SearchEvent{
				Name: display, State: models.SearchDone,
				Count: len(res.results), Elapsed: time.Since(start),
			})
		case <-ctx.Done():
			util.Debug("search context ended; returning collected results", "reason", ctx.Err())
			for _, a := range searchers {
				if !pending[a.kind] {
					continue
				}
				f, _ := parentContextFailure(ctx, a.kind, ctx.Err())
				failures = append(failures, f)
				emit(models.SearchEvent{Name: panelSourceName(a.kind), State: models.SearchFailed, Elapsed: time.Since(start), Reason: f.Reason})
			}
			return finishSearch(query, searchers, off, all, failures)
		}
	}
}
