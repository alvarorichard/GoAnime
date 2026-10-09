package download

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/api/providers"
	apisource "github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// Not parallel: these swap package seams through dmFixture.

func topCineSeries() *models.Anime {
	return &models.Anime{Name: "Dexter: Ressurreição", URL: "https://topcine3.site/serie/dexter-ressurreicao", Source: "TopCine", MediaType: models.MediaTypeTV}
}

// TestDM_SearchesEveryMovieTVSource: the default -dm search fans out over
// StartFlix and TopCine together, so the user picks from both.
func TestDM_SearchesEveryMovieTVSource(t *testing.T) {
	var got []apisource.SourceKind
	api.SetSearchFetch(func(_ context.Context, _ string, kinds []apisource.SourceKind) ([]*models.Anime, error) {
		got = kinds
		return nil, nil // no results: the search fails before any picker opens
	})
	t.Cleanup(func() {
		api.SetSearchFetch(func(ctx context.Context, query string, kinds []apisource.SourceKind) ([]*models.Anime, error) {
			return providers.SearchAll(ctx, query, kinds...)
		})
	})

	_, err := movieSearchFn("dexter")
	if !errors.Is(err, api.ErrNoResults) {
		t.Fatalf("err = %v, want ErrNoResults from the stubbed fan-out", err)
	}
	want := []apisource.SourceKind{apisource.StartFlix, apisource.TopCine}
	if !slices.Equal(got, want) {
		t.Errorf("searched %v, want %v", got, want)
	}
}

func TestDM_TopCineSeriesEpisode(t *testing.T) {
	calls, _ := dmFixture(t, topCineSeries())
	if err := HandleMovieDownloadRequest(&util.DownloadRequest{AnimeName: "dexter", IsTV: true, SeasonNum: 2, EpisodeNum: 1}); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 1 || (*calls)[0].season != 2 || (*calls)[0].start != 1 {
		t.Errorf("batch calls = %+v, want season 2 episode 1", *calls)
	}
}

func TestDM_TopCineMessagesNameTheSource(t *testing.T) {
	t.Run("series asked as a movie", func(t *testing.T) {
		dmFixture(t, topCineSeries())
		err := HandleMovieDownloadRequest(&util.DownloadRequest{AnimeName: "dexter", IsMovie: true})
		if err == nil || !strings.Contains(err.Error(), "is a series on TopCine") {
			t.Errorf("err = %v, want TopCine named", err)
		}
	})
	t.Run("movie asked as a series", func(t *testing.T) {
		movie := &models.Anime{Name: "Zona Zero", URL: "https://topcine3.site/filme/zona-zero", Source: "TopCine", MediaType: models.MediaTypeMovie}
		dmFixture(t, movie)
		err := HandleMovieDownloadRequest(&util.DownloadRequest{AnimeName: "zona zero", IsTV: true, IsAll: true})
		if err == nil || !strings.Contains(err.Error(), "is a movie on TopCine") {
			t.Errorf("err = %v, want TopCine named", err)
		}
	})
}

func TestSourceOr(t *testing.T) {
	t.Parallel()
	if got := sourceOr(&models.Anime{Source: "TopCine"}); got != "TopCine" {
		t.Errorf("sourceOr = %q", got)
	}
	if got := sourceOr(&models.Anime{}); got != "StartFlix" {
		t.Errorf("sourceOr of an unlabelled title = %q, want the StartFlix default", got)
	}
}
