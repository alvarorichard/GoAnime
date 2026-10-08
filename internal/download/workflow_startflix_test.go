package download

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/api/providers/metadata"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/util"
)

// batchCall is one call into the player's batch downloads, with the season
// the player would file the episodes under at that moment.
type batchCall struct {
	prompting  bool // a batch that asks the user for a range (none may)
	start, end int
	season     int
	episodes   []int
}

// stubStartFlixWorkflow runs HandleDownloadRequest against a StartFlix series
// whose listing — the season and audio pickers — lands on season 3, with the
// player's batch downloads mocked. Not parallel: it swaps package seams.
func stubStartFlixWorkflow(t *testing.T, listed []int) *[]batchCall {
	t.Helper()
	anime := &models.Anime{Name: "Dark", URL: "https://www.startflix.test/series/dark/", Source: "StartFlix", MediaType: models.MediaTypeTV}
	calls := &[]batchCall{}
	record := func(eps []models.Episode, prompting bool, start, end int) {
		nums := make([]int, 0, len(eps))
		for _, e := range eps {
			nums = append(nums, e.Num)
		}
		*calls = append(*calls, batchCall{prompting: prompting, start: start, end: end, season: player.GetAnimeSeason(), episodes: nums})
	}

	prevSearch, prevEnrich := workflowSearchFn, workflowEnrichFn
	prevFetch, prevLegacy, prevAll, prevRange := workflowFetchEpisodesFn, workflowLegacyEpisodesFn, workflowDownloadAllFn, workflowBatchRangeFn
	prevMeta := player.GetMediaMeta()
	workflowSearchFn = func(string) (*models.Anime, error) { return anime, nil }
	workflowEnrichFn = func(context.Context, *models.Anime) ([]metadata.SeasonMapping, error) { return nil, nil }
	workflowFetchEpisodesFn = func(_ context.Context, a *models.Anime) ([]models.Episode, error) {
		a.CurrentSeason = 3 // what the StartFlix season picker records
		eps := make([]models.Episode, 0, len(listed))
		for _, n := range listed {
			eps = append(eps, models.Episode{Num: n, SeasonID: "3"})
		}
		return eps, nil
	}
	workflowLegacyEpisodesFn = func(string) ([]models.Episode, error) {
		return nil, errors.New("legacy listing: no AnimeFire episode list on a StartFlix page")
	}
	workflowDownloadAllFn = func(eps []models.Episode, _ *models.Anime) error {
		record(eps, false, eps[0].Num, eps[len(eps)-1].Num)
		return nil
	}
	workflowBatchRangeFn = func(eps []models.Episode, _ *models.Anime, start, end int) error {
		record(eps, false, start, end)
		return nil
	}
	t.Cleanup(func() {
		workflowSearchFn, workflowEnrichFn = prevSearch, prevEnrich
		workflowFetchEpisodesFn, workflowLegacyEpisodesFn, workflowDownloadAllFn, workflowBatchRangeFn = prevFetch, prevLegacy, prevAll, prevRange
		player.SetMediaMeta(prevMeta)
	})
	return calls
}

// TestDownloadAll_DownloadsEveryEpisodeWithoutAskingARange: "-d -a" means
// every episode. It used to open HandleBatchDownload's start/end form — and
// without a terminal fall through to a legacy listing StartFlix has no page for.
func TestDownloadAll_DownloadsEveryEpisodeWithoutAskingARange(t *testing.T) {
	calls := stubStartFlixWorkflow(t, []int{1, 2, 3})
	if err := HandleDownloadRequest(&util.DownloadRequest{AnimeName: "dark", IsAll: true}); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 1 || (*calls)[0].prompting {
		t.Fatalf("batch calls = %+v, want one download of the whole list without a range prompt", *calls)
	}
	if c := (*calls)[0]; c.start != 1 || c.end != 3 {
		t.Errorf("downloaded %d-%d, want 1-3", c.start, c.end)
	}
}

// TestDownloadRange_FilesEpisodesUnderThePickedSeason: StartFlix numbers
// episodes per season; the season the picker chose must name the files.
// They used to be filed under season 1 (the flag's default), so S03E01 was
// written as S01E01, on top of the real one.
func TestDownloadRange_FilesEpisodesUnderThePickedSeason(t *testing.T) {
	calls := stubStartFlixWorkflow(t, []int{1, 2, 3})
	if err := HandleDownloadRequest(&util.DownloadRequest{AnimeName: "dark", IsRange: true, StartEpisode: 1, EndEpisode: 2}); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 1 {
		t.Fatalf("batch calls = %+v", *calls)
	}
	if got := (*calls)[0].season; got != 3 {
		t.Errorf("episodes filed under season %d, want the picked season 3", got)
	}
}

// TestDownloadSingleEpisode_UsesTheSourceListing: "-d name N" listed episodes
// through the legacy AnimeFire-page parser only, which has nothing to read on
// a StartFlix page, so a single StartFlix episode could never be downloaded.
func TestDownloadSingleEpisode_UsesTheSourceListing(t *testing.T) {
	calls := stubStartFlixWorkflow(t, []int{1, 2, 3})
	if err := HandleDownloadRequest(&util.DownloadRequest{AnimeName: "dark", EpisodeNum: 2}); err != nil {
		t.Fatalf("single StartFlix episode: %v", err)
	}
	if len(*calls) != 1 {
		t.Fatalf("batch calls = %+v", *calls)
	}
	if c := (*calls)[0]; c.prompting || c.start != 2 || c.end != 2 || c.season != 3 {
		t.Errorf("call = %+v, want episode 2 alone, filed under season 3", c)
	}
}

// dmFixture mocks StartFlix for the -dm downloads: a search that returns
// title, a series with seasons 1, 2 and 4 (episodes 1-3 each, season 4 only
// 1 and 3), and the player's batch downloads, recording what they were
// asked for. Not parallel: it swaps package seams.
func dmFixture(t *testing.T, title *models.Anime) (*[]batchCall, *[]string) {
	t.Helper()
	calls := &[]batchCall{}
	var searched []string
	record := func(eps []models.Episode, start, end int) {
		nums := make([]int, 0, len(eps))
		for _, e := range eps {
			nums = append(nums, e.Num)
		}
		*calls = append(*calls, batchCall{start: start, end: end, season: player.GetAnimeSeason(), episodes: nums})
	}
	seasonEps := map[int][]int{1: {1, 2, 3}, 2: {1, 2, 3}, 4: {1, 3}}

	prevSearch, prevSeasons, prevEps := movieSearchFn, seriesSeasonsFn, seriesEpisodesFn
	prevFetch, prevAll, prevRange := workflowFetchEpisodesFn, workflowDownloadAllFn, workflowBatchRangeFn
	prevMeta := player.GetMediaMeta()
	movieSearchFn = func(name string) (*models.Anime, error) { searched = append(searched, name); return title, nil }
	seriesSeasonsFn = func(a *models.Anime) ([]int, error) {
		if a.MediaType == models.MediaTypeMovie {
			return nil, api.ErrStartFlixNotSeries
		}
		return []int{1, 2, 4}, nil
	}
	seriesEpisodesFn = func(a *models.Anime, n int) ([]models.Episode, error) {
		if a.MediaType == models.MediaTypeMovie {
			return nil, api.ErrStartFlixNotSeries
		}
		nums, ok := seasonEps[n]
		if !ok {
			return nil, fmt.Errorf("season %d is not on StartFlix", n)
		}
		a.CurrentSeason = n
		eps := make([]models.Episode, 0, len(nums))
		for _, x := range nums {
			eps = append(eps, models.Episode{Num: x})
		}
		return eps, nil
	}
	workflowFetchEpisodesFn = func(_ context.Context, a *models.Anime) ([]models.Episode, error) {
		return []models.Episode{{Num: 1, Number: "1"}}, nil // the movie's single entry
	}
	workflowDownloadAllFn = func(eps []models.Episode, _ *models.Anime) error {
		record(eps, eps[0].Num, eps[len(eps)-1].Num)
		return player.ErrUserQuit // what a finished batch returns
	}
	workflowBatchRangeFn = func(eps []models.Episode, _ *models.Anime, start, end int) error {
		record(eps, start, end)
		return player.ErrUserQuit
	}
	t.Cleanup(func() {
		movieSearchFn, seriesSeasonsFn, seriesEpisodesFn = prevSearch, prevSeasons, prevEps
		workflowFetchEpisodesFn, workflowDownloadAllFn, workflowBatchRangeFn = prevFetch, prevAll, prevRange
		player.SetMediaMeta(prevMeta)
	})
	return calls, &searched
}

func series() *models.Anime {
	return &models.Anime{Name: "Dark", URL: "https://www.startflix.test/series/dark/", Source: "StartFlix", MediaType: models.MediaTypeTV}
}

// TestDM_Movie: "-dm Movie" was a stub that always failed; it now searches
// StartFlix and downloads the movie as one episode, filed as a movie.
func TestDM_Movie(t *testing.T) {
	movie := &models.Anime{Name: "Heart of the Beast", URL: "https://www.startflix.test/filmes/x/", Source: "StartFlix", MediaType: models.MediaTypeMovie}
	calls, searched := dmFixture(t, movie)
	if err := HandleMovieDownloadRequest(&util.DownloadRequest{AnimeName: "coracao selvagem", IsMovie: true}); err != nil {
		t.Fatal(err)
	}
	if len(*searched) != 1 || (*searched)[0] != "coracao selvagem" {
		t.Errorf("searched %v", *searched)
	}
	if len(*calls) != 1 || (*calls)[0].start != 1 || (*calls)[0].end != 1 {
		t.Fatalf("batch calls = %+v, want the movie's single entry", *calls)
	}
	if got := player.GetExactMediaType(); got != string(models.MediaTypeMovie) {
		t.Errorf("filed as %q, want movie", got)
	}
}

func TestDM_MovieThatIsASeries(t *testing.T) {
	calls, _ := dmFixture(t, series())
	err := HandleMovieDownloadRequest(&util.DownloadRequest{AnimeName: "dark", IsMovie: true})
	if err == nil || !strings.Contains(err.Error(), "--type tv") {
		t.Errorf("err = %v, want a pointer to the series forms", err)
	}
	if len(*calls) != 0 {
		t.Errorf("downloaded %+v", *calls)
	}
}

// TestDM_SeriesEpisodeAndRange: the season comes from the command and names
// the files; the range is the command's.
func TestDM_SeriesEpisodeAndRange(t *testing.T) {
	calls, _ := dmFixture(t, series())
	if err := HandleMovieDownloadRequest(&util.DownloadRequest{AnimeName: "dark", IsTV: true, SeasonNum: 2, EpisodeNum: 3}); err != nil {
		t.Fatal(err)
	}
	if err := HandleMovieDownloadRequest(&util.DownloadRequest{AnimeName: "dark", IsTV: true, IsRange: true, SeasonNum: 4, StartEpisode: 1, EndEpisode: 3}); err != nil {
		t.Fatal(err)
	}
	want := []batchCall{
		{start: 3, end: 3, season: 2, episodes: []int{1, 2, 3}},
		{start: 1, end: 3, season: 4, episodes: []int{1, 3}},
	}
	if len(*calls) != len(want) {
		t.Fatalf("batch calls = %+v", *calls)
	}
	for i, w := range want {
		if c := (*calls)[i]; c.start != w.start || c.end != w.end || c.season != w.season || len(c.episodes) != len(w.episodes) {
			t.Errorf("call %d = %+v, want %+v", i, c, w)
		}
	}
}

func TestDM_SeriesMissingSeason(t *testing.T) {
	calls, _ := dmFixture(t, series())
	err := HandleMovieDownloadRequest(&util.DownloadRequest{AnimeName: "dark", IsTV: true, SeasonNum: 3, EpisodeNum: 1})
	if err == nil || !strings.Contains(err.Error(), "season 3") {
		t.Errorf("err = %v, want the missing season named", err)
	}
	if len(*calls) != 0 {
		t.Errorf("downloaded %+v", *calls)
	}
}

// TestDM_SeriesAllSeasons: "-dm -a" downloads every season, each filed under
// its own number, every listed episode — gaps (season 4 has no episode 2)
// are not a range to ask about or a failure.
func TestDM_SeriesAllSeasons(t *testing.T) {
	calls, _ := dmFixture(t, series())
	if err := HandleMovieDownloadRequest(&util.DownloadRequest{AnimeName: "dark", IsTV: true, IsAll: true}); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 3 {
		t.Fatalf("batch calls = %+v, want one per season", *calls)
	}
	for i, season := range []int{1, 2, 4} {
		if c := (*calls)[i]; c.season != season {
			t.Errorf("call %d filed under season %d, want %d", i, c.season, season)
		}
	}
	if c := (*calls)[2]; c.start != 1 || c.end != 3 || len(c.episodes) != 2 {
		t.Errorf("season 4 = %+v, want episodes 1 and 3 spanning 1-3", c)
	}
}

func TestDM_SeriesAllSeasonsKeepsGoingPastAFailedSeason(t *testing.T) {
	calls, _ := dmFixture(t, series())
	failing := workflowDownloadAllFn
	workflowDownloadAllFn = func(eps []models.Episode, a *models.Anime) error {
		_ = failing(eps, a)
		if a.CurrentSeason == 2 {
			return errors.New("2 episodes failed to download")
		}
		return player.ErrUserQuit
	}
	err := HandleMovieDownloadRequest(&util.DownloadRequest{AnimeName: "dark", IsTV: true, IsAll: true})
	if err == nil || !strings.Contains(err.Error(), "season 2") {
		t.Errorf("err = %v, want season 2's failure reported", err)
	}
	if len(*calls) != 3 {
		t.Errorf("seasons attempted = %d, want all 3", len(*calls))
	}
}
