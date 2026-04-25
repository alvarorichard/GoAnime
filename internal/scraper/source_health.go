// Package scraper implements provider search, stream extraction, and source diagnostics.
package scraper

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
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
	Diagnostic  *SourceDiagnostic
	Description string
}

// DefaultHealthCheckQuery returns a stable query expected to produce results.
func DefaultHealthCheckQuery(source ScraperType) string {
	switch source {
	case FlixHQType, SFlixType, SuperFlixType:
		return "dexter"
	default:
		return "naruto"
	}
}

// AvailableSources returns registered scraper types in deterministic order.
func (sm *ScraperManager) AvailableSources() []ScraperType {
	sources := make([]ScraperType, 0, len(sm.scrapers))
	for source := range sm.scrapers {
		sources = append(sources, source)
	}
	sort.Slice(sources, func(i, j int) bool {
		return sources[i] < sources[j]
	})
	return sources
}

// CheckSourceHealth probes one provider and classifies the result for CI/app diagnostics.
func (sm *ScraperManager) CheckSourceHealth(ctx context.Context, source ScraperType, query string) SourceHealthResult {
	sourceName := sm.getScraperDisplayName(source)
	if query == "" {
		query = DefaultHealthCheckQuery(source)
	}

	startedAt := time.Now()
	result := SourceHealthResult{
		Source:     source,
		SourceName: sourceName,
		Query:      query,
	}

	scraper, exists := sm.scrapers[source]
	if !exists {
		diagnostic := NewInternalBugError(sourceName, "health-check", "scraper not registered", nil)
		result.Status = SourceHealthFailed
		result.Diagnostic = DiagnoseError(sourceName, "health-check", diagnostic)
		result.Duration = time.Since(startedAt)
		result.Description = result.Diagnostic.UserMessage()
		return result
	}

	if diagnostic, retryAfter, open := sm.circuitOpenDiagnostic(source); open {
		diagnostic.Message = fmt.Sprintf("%s; retry after %s", diagnostic.Message, retryAfter.Round(time.Second))
		result.Status = SourceHealthSkipped
		result.Diagnostic = diagnostic
		result.Duration = time.Since(startedAt)
		result.Description = diagnostic.UserMessage()
		return result
	}

	type searchOutcome struct {
		results []*models.Anime
		err     error
	}

	done := make(chan searchOutcome, 1)
	go func() {
		results, err := scraper.SearchAnime(query)
		done <- searchOutcome{results: results, err: err}
	}()

	select {
	case outcome := <-done:
		result.Duration = time.Since(startedAt)
		if outcome.err != nil {
			diagnostic := DiagnoseError(sourceName, "health-check", outcome.err)
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
			diagnostic := DiagnoseError(sourceName, "health-check", NewParserError(sourceName, "health-check", "known query returned zero results", nil))
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
		diagnostic := DiagnoseError(sourceName, "health-check", ctx.Err())
		result.Status = SourceHealthSkipped
		result.Diagnostic = diagnostic
		result.Description = diagnostic.UserMessage()
		return result
	}
}

// CheckAllSourcesHealth probes all registered providers in deterministic order.
func (sm *ScraperManager) CheckAllSourcesHealth(ctx context.Context) []SourceHealthResult {
	sources := sm.AvailableSources()
	results := make([]SourceHealthResult, 0, len(sources))
	for _, source := range sources {
		results = append(results, sm.CheckSourceHealth(ctx, source, DefaultHealthCheckQuery(source)))
	}
	return results
}
