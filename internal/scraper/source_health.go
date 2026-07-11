// Package scraper implements per-source adapters and source diagnostics.
package scraper

import (
	"context"
	"fmt"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
)

// SourceHealthStatus is the result class for a provider health probe.
type SourceHealthStatus string

const (
	// SourceHealthHealthy means the provider returned parseable search results.
	SourceHealthHealthy SourceHealthStatus = "healthy"
	// SourceHealthSkipped means the provider is offline/blocked and CI should not fail.
	SourceHealthSkipped SourceHealthStatus = "skipped"
	// SourceHealthFailed means the provider responded but GoAnime likely needs a fix.
	SourceHealthFailed SourceHealthStatus = "failed"
)

// SourceHealthResult describes a single provider health probe.
type SourceHealthResult struct {
	Source      ScraperType
	SourceName  string
	Query       string
	Status      SourceHealthStatus
	Results     int
	Duration    time.Duration
	Diagnostic  *netx.SourceDiagnostic
	Description string
}

// DefaultHealthCheckQuery returns a stable query expected to produce results.
func DefaultHealthCheckQuery(source ScraperType) string {
	switch source {
	case SuperFlixType:
		return "dexter"
	default:
		return "naruto"
	}
}

// healthTargets returns the source types to probe, in deterministic order.
func healthTargets() []ScraperType {
	return []ScraperType{AllAnimeType, AnimefireType, GoyabuType, SuperFlixType}
}

// checkSourceHealthWith probes a single scraper (which may be nil) and classifies
// the outcome for CI/app diagnostics. It is the injection seam: tests pass a mock
// scraper, while the live entry points build a real adapter via NewAdapter.
func checkSourceHealthWith(ctx context.Context, source ScraperType, sc UnifiedScraper, query string) SourceHealthResult {
	sourceName := scraperDisplayName(source)
	if query == "" {
		query = DefaultHealthCheckQuery(source)
	}

	startedAt := time.Now()
	result := SourceHealthResult{
		Source:     source,
		SourceName: sourceName,
		Query:      query,
	}

	if sc == nil {
		diagnostic := netx.NewInternalBugError(sourceName, "health-check", "scraper not registered", nil)
		result.Status = SourceHealthFailed
		result.Diagnostic = netx.DiagnoseError(sourceName, "health-check", diagnostic)
		result.Duration = time.Since(startedAt)
		result.Description = result.Diagnostic.UserMessage()
		return result
	}

	type searchOutcome struct {
		results []*models.Anime
		err     error
	}

	done := make(chan searchOutcome, 1)
	go func() {
		results, err := sc.SearchAnime(query)
		done <- searchOutcome{results: results, err: err}
	}()

	select {
	case outcome := <-done:
		result.Duration = time.Since(startedAt)
		if outcome.err != nil {
			diagnostic := netx.DiagnoseError(sourceName, "health-check", outcome.err)
			result.Diagnostic = diagnostic
			result.Description = diagnostic.UserMessage()
			if diagnostic.ShouldSkipHealthCheck() {
				result.Status = SourceHealthSkipped
				return result
			}
			result.Status = SourceHealthFailed
			return result
		}

		result.Results = len(outcome.results)
		if len(outcome.results) == 0 {
			diagnostic := netx.DiagnoseError(sourceName, "health-check", netx.NewParserError(sourceName, "health-check", "known query returned zero results", nil))
			result.Status = SourceHealthFailed
			result.Diagnostic = diagnostic
			result.Description = diagnostic.UserMessage()
			return result
		}

		result.Status = SourceHealthHealthy
		result.Description = fmt.Sprintf("%s healthy: query %q returned %d result(s)", sourceName, query, len(outcome.results))
		return result

	case <-ctx.Done():
		result.Duration = time.Since(startedAt)
		diagnostic := netx.DiagnoseError(sourceName, "health-check", ctx.Err())
		result.Status = SourceHealthSkipped
		result.Diagnostic = diagnostic
		result.Description = diagnostic.UserMessage()
		return result
	}
}

// CheckSourceHealth builds the live adapter for source and probes it.
func CheckSourceHealth(ctx context.Context, source ScraperType, query string) SourceHealthResult {
	sc, _ := NewAdapter(source)
	return checkSourceHealthWith(ctx, source, sc, query)
}

// CheckAllSourcesHealth probes all live sources in deterministic order.
func CheckAllSourcesHealth(ctx context.Context) []SourceHealthResult {
	targets := healthTargets()
	results := make([]SourceHealthResult, 0, len(targets))
	for _, source := range targets {
		results = append(results, CheckSourceHealth(ctx, source, DefaultHealthCheckQuery(source)))
	}
	return results
}
