package api

import (
	"context"
	"math"
	"os"

	apisource "github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"golang.org/x/term"
)

// SearchLiveFunc streams a source fan-out: the plan first, then one outcome
// per source as it answers, returning what SearchAll would have. Set by the
// providers package's init; nil in tests and in builds without it.
type SearchLiveFunc func(ctx context.Context, query string, kinds []apisource.SourceKind, emit func(models.SearchEvent)) ([]*models.Anime, error)

// searchLiveFetchFn is the live fan-out seam. Nil keeps the spinner.
var searchLiveFetchFn SearchLiveFunc

// SetSearchLiveFetch installs the registry-backed live search dispatch.
// Called from providers.init().
func SetSearchLiveFetch(f SearchLiveFunc) { searchLiveFetchFn = f }

type liveOutcome struct {
	animes []*models.Anime
	err    error
}

// searchLiveRunnerFn runs the SEARCHING screen; swapped in tests for a
// headless runner.
var searchLiveRunnerFn = tui.SearchLiveResults

// canShowLiveSearch reports whether the SEARCHING screen can open: the live
// seam is wired and stdin is a terminal. Anything else keeps the spinner.
func canShowLiveSearch() bool {
	if searchLiveFetchFn == nil {
		return false
	}
	fd := os.Stdin.Fd()
	return fd <= math.MaxInt && term.IsTerminal(int(fd))
}

// runLiveSearch runs the fan-out behind the SEARCHING screen (tela 2). It
// returns either the title enter opened early (chosen) or, once every source
// answered or the budget ran out, the collected results for the RESULTS
// screen. Cancelling or going back surfaces the screen's error.
func runLiveSearch(name string, kinds []apisource.SourceKind) (animes []*models.Anime, chosen *models.Anime, err error) {
	events := make(chan models.SearchEvent, 64)
	finished := make(chan liveOutcome, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Capture the seams up front: the fetch goroutine must keep working
	// with the hook it started with, even if SearchAnimeAuto swaps the
	// global for a direct reopen meanwhile.
	fetch := searchLiveFetchFn
	run := searchLiveRunnerFn
	go func() {
		a, e := fetch(ctx, name, kinds, func(ev models.SearchEvent) {
			select {
			case events <- ev:
			default:
			}
		})
		finished <- liveOutcome{a, e}
	}()

	res, runErr := run(tui.SearchLiveOptions{Query: name}, events)
	if runErr != nil {
		return nil, nil, runErr
	}
	if res.Chosen != nil {
		return nil, res.Chosen, nil
	}
	out := <-finished
	return out.animes, nil, out.err
}
