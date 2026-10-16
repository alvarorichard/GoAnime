package appflow

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"charm.land/huh/v2"
	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/api/providers"
	apisource "github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
	"golang.org/x/term"
)

// errTryUnavailable means the TRY screen cannot open (no terminal): the
// retry loop falls back to the legacy report-and-prompt path.
var errTryUnavailable = errors.New("try screen unavailable")

// openTryScreenFn opens the no-results screen (tela 4). Tests inject a
// scripted decision; without a terminal the default reports unavailable.
func defaultOpenTryScreen(opts tui.TryOptions) (tui.TryDecision, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return tui.TryDecision{}, errTryUnavailable
	}
	decision, err := tui.TrySearch(opts)
	if err != nil {
		if errors.Is(err, tui.ErrSelectionCancelled) {
			return tui.TryDecision{}, fmt.Errorf("%w: %w", api.ErrSearchAborted, err)
		}
		return tui.TryDecision{}, err
	}
	return decision, nil
}

// tryAgain opens the TRY screen for a failed search. It reports whether the
// loop should retry: false with errTryUnavailable falls back to the legacy
// prompt, any other error aborts the search.
func tryAgain(query, loopSource string, searchErr error) (tui.TryDecision, error) {
	opts := buildTryOptions(query, loopSource, searchErr)
	decision, err := openTryScreenFn(opts)
	if err != nil {
		return tui.TryDecision{}, err
	}
	if strings.TrimSpace(decision.Query) == "" {
		return tui.TryDecision{}, fmt.Errorf("empty retry query: %w", api.ErrSearchAborted)
	}
	return decision, nil
}

// buildTryOptions assembles the no-results screen: the summary, per-source
// outcomes and the retry suggestions.
func buildTryOptions(query, loopSource string, searchErr error) tui.TryOptions {
	spaced := strings.Join(strings.Fields(strings.ReplaceAll(query, "-", " ")), " ")
	if spaced == "" {
		spaced = query
	}
	opts := tui.TryOptions{
		Query:        spaced,
		CurrentScope: loopSource,
		SourcesLine:  trySourcesLine(),
		Note:         "Titles often differ between sources: romaji, English, and the Portuguese name on the PT-BR sites.",
	}

	var answered []apisource.SourceKind
	var failed []providers.SourceFailure
	var off []models.SearchOffEntry
	var noResults *providers.NoResultsInfo
	var failure *providers.SearchFailure
	switch {
	case errors.As(searchErr, &failure):
		answered, failed, off = failure.AnsweredKinds, failure.Sources, failure.Off
	case errors.As(searchErr, &noResults):
		answered, off = noResults.AnsweredKinds, noResults.Off
	case searchErr != nil:
		util.Debugf("Try screen without source outcomes: %v", searchErr)
	}

	opts.Summary = trySummary(spaced, answered, failed)
	for _, kind := range answered {
		opts.Sources = append(opts.Sources, tui.TrySource{Name: string(kind), Ok: true})
	}
	for _, f := range failed {
		opts.Sources = append(opts.Sources, tui.TrySource{
			Name:   string(f.Kind),
			Detail: strings.TrimPrefix(f.Reason, "is "),
		})
	}
	opts.Off = off
	opts.Suggestions = buildTrySuggestions(spaced, loopSource)
	return opts
}

// trySummary reads "Nothing for "q". 5 sources answered without a match, 1
// could not be reached."
func trySummary(query string, answered []apisource.SourceKind, failed []providers.SourceFailure) string {
	var parts []string
	if n := len(answered); n > 0 {
		if n == 1 {
			parts = append(parts, "1 source answered without a match")
		} else {
			parts = append(parts, fmt.Sprintf("%d sources answered without a match", n))
		}
	}
	if n := len(failed); n > 0 {
		if n == 1 {
			parts = append(parts, "1 could not be reached")
		} else {
			parts = append(parts, fmt.Sprintf("%d could not be reached", n))
		}
	}
	summary := fmt.Sprintf("Nothing for %q.", query)
	if len(parts) > 0 {
		summary += " " + strings.Join(parts, ", ") + "."
	}
	return summary
}

// buildTrySuggestions offers a spelling fix, fewer words and English-only,
// in that order (tela 4).
func buildTrySuggestions(query, loopSource string) []tui.TrySuggestion {
	var out []tui.TrySuggestion
	if title := didYouMean(query); title != "" {
		out = append(out, tui.TrySuggestion{
			Label: title, Hint: "did you mean · AniList", Query: title, Scope: loopSource,
		})
	}
	if words := strings.Fields(query); len(words) >= 2 {
		shorter := strings.Join(words[1:], " ")
		out = append(out, tui.TrySuggestion{
			Label: shorter, Hint: "fewer words", Query: shorter, Scope: loopSource,
		})
	}
	if !strings.EqualFold(strings.TrimSpace(loopSource), "hianime") {
		out = append(out, tui.TrySuggestion{
			Label: query, Hint: "English sources only", Query: query, Scope: "hianime",
		})
	}
	return out
}

// didYouMean asks AniList for the title it would have searched: a misspelled
// query comes back canonical ("namakra gatana" suggests "Namakura Gatana").
// Best-effort: anything inconclusive yields no suggestion.
func didYouMean(query string) string {
	if strings.TrimSpace(query) == "" {
		return ""
	}
	info, err := aniListFetchFn(query)
	if err != nil || info == nil {
		util.Debugf("AniList suggestion unavailable: %v", err)
		return ""
	}
	title := strings.TrimSpace(info.Data.Media.Title.English)
	if title == "" {
		title = strings.TrimSpace(info.Data.Media.Title.Romaji)
	}
	if title == "" || strings.EqualFold(title, strings.TrimSpace(query)) {
		return ""
	}
	return title
}

// homeBack is what going back lands on: a new search or a CONTINUE row.
type homeBack struct {
	search       bool
	query, scope string
	cont         *tui.SearchHomeResult
}

// backToSearchHome returns to the SEARCH home screen with the query kept
// ("edit search", tela 3). Without an opener it reports errTryUnavailable
// and the loop keeps the legacy prompt.
func backToSearchHome(query string) (homeBack, error) {
	if OpenSearchHomeFn == nil {
		return homeBack{}, errTryUnavailable
	}
	res, err := OpenSearchHomeFn(query)
	if err != nil {
		if errors.Is(err, tui.ErrSearchHomeCancelled) {
			return homeBack{}, fmt.Errorf("%w: %w", api.ErrSearchAborted, err)
		}
		return homeBack{}, err
	}
	switch res.Action {
	case tui.HomeResume, tui.HomeFromStart, tui.HomeEpisodes:
		if !res.HasEntry {
			return homeBack{}, errTryUnavailable
		}
		return homeBack{cont: &res}, nil
	default:
		if strings.TrimSpace(res.Query) == "" {
			return homeBack{}, errTryUnavailable
		}
		return homeBack{search: true, query: res.Query, scope: res.Scope}, nil
	}
}

// ErrHomeContinue signals a change-anime that became a CONTINUE reopen:
// the decision travels via PendingHomeResult for the caller to dispatch.
var ErrHomeContinue = errors.New("change anime via continue")

// ChangeAnimeHome changes the title through the SEARCH home screen instead
// of the legacy prompt: type a query or pick CONTINUE/RECENT. A search runs
// the full retry flow and returns its anime with episodes; a CONTINUE row
// pends for the caller and reports ErrHomeContinue. Without an opener it
// keeps the legacy prompt, which errors without a terminal.
func ChangeAnimeHome() (*models.Anime, []models.Episode, error) {
	if OpenSearchHomeFn == nil {
		return legacyChangeAnime()
	}
	res, err := OpenSearchHomeFn("")
	if err != nil {
		if errors.Is(err, tui.ErrSearchHomeCancelled) {
			return nil, nil, fmt.Errorf("%w: %w", api.ErrSearchAborted, err)
		}
		return nil, nil, err
	}
	switch res.Action {
	case tui.HomeResume, tui.HomeFromStart, tui.HomeEpisodes:
		if !res.HasEntry {
			return nil, nil, fmt.Errorf("home continue without an entry")
		}
		PendingHomeResult = &res
		return nil, nil, ErrHomeContinue
	default:
		if strings.TrimSpace(res.Query) == "" {
			return nil, nil, fmt.Errorf("empty change-anime query")
		}
		anime, err := searchWithRetryFn(res.Query, res.Scope)
		if err != nil {
			return nil, nil, err
		}
		episodes, err := getAnimeEpisodesEnhancedFn(anime)
		if err != nil {
			return nil, nil, err
		}
		if len(episodes) == 0 {
			return nil, nil, fmt.Errorf("no episodes for %q", anime.Name)
		}
		return anime, episodes, nil
	}
}

// legacyChangeAnime is the pre-home prompt: a bare huh input, kept for
// binaries without a home opener. It errors without a terminal instead of
// hanging.
func legacyChangeAnime() (*models.Anime, []models.Episode, error) {
	const maxRetries = 3

	// Fail fast without a TTY: huh blocks on Windows console input headless
	// (ReadConsole) while it errors on Linux. Without this, headless CI on
	// windows-latest hangs the full go-test timeout instead of erroring.
	if !tui.HasTerminal() {
		return nil, nil, fmt.Errorf("no terminal: stdin is not a TTY")
	}

	for i := range maxRetries {
		var animeName string

		prompt := huh.NewInput().
			Title("Change Anime").
			Description("Enter the name of the anime you want to watch:").
			Value(&animeName).
			Validate(func(v string) error {
				if len(v) < 2 {
					return fmt.Errorf("anime name must be at least 2 characters")
				}
				return nil
			})

		if err := tui.RunClean(prompt.WithTheme(tui.FormTheme()).Run); err != nil {
			return nil, nil, err
		}

		// Use the enhanced API to search for anime
		anime, err := api.SearchAnimeEnhanced(animeName, "")
		if err != nil || anime == nil {
			if i < maxRetries-1 {
				util.Errorf("No anime found with the name: %s", animeName)
				util.Infof("Please try again with a different search term. (Attempt %d/%d)", i+2, maxRetries)
				continue
			}
			return nil, nil, fmt.Errorf("failed to find anime after %d attempts", maxRetries)
		}

		// Get episodes for the new anime via the Model B registry.
		episodes, err := getAnimeEpisodesEnhancedFn(anime)
		if err != nil {
			if i < maxRetries-1 {
				util.Errorf("Failed to get episodes for: %s", anime.Name)
				util.Infof("Please try searching for a different anime. (Attempt %d/%d)", i+2, maxRetries)
				continue
			}
			return nil, nil, fmt.Errorf("failed to get episodes after %d attempts", maxRetries)
		}

		return anime, episodes, nil
	}

	return nil, nil, fmt.Errorf("failed to change anime after %d attempts", maxRetries)
}

// trySourcesLine reads "6 sources on · Goyabu off" at the right of the
// header. Same line as the search home screen (see handlers.sourcesLine).
func trySourcesLine() string {
	active := apisource.ActiveSources()
	disabled := apisource.DisabledSources()
	sep := " · "
	if tui.Plain() {
		sep = " - "
	}
	line := fmt.Sprintf("%d sources on", len(active))
	if len(disabled) > 0 {
		names := make([]string, 0, len(disabled))
		for _, kind := range disabled {
			names = append(names, string(kind))
		}
		line += sep + strings.Join(names, ", ") + " off"
	}
	return line
}
