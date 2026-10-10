package playback

import (
	"errors"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// useMovieData swaps the Jikan lookup for fn. Not parallel: getMovieData is a
// package-level seam.
func useMovieData(t *testing.T, fn func(int, *models.Anime) error) {
	t.Helper()
	prev := getMovieData
	getMovieData = fn
	t.Cleanup(func() { getMovieData = prev })
}

// An unreachable Jikan used to hold the screen blank until its request timed
// out. fetchMovieData must hand control back at once, however long the lookup
// takes.
func TestFetchMovieData_DoesNotWaitForTheLookup(t *testing.T) {
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	useMovieData(t, func(int, *models.Anime) error {
		<-release
		return nil
	})

	done := make(chan struct{})
	go func() {
		fetchMovieData(&models.Anime{MalID: 6654})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("fetchMovieData blocked on the metadata lookup")
	}
}

func TestFetchMovieData_FailureSendsNothing(t *testing.T) {
	called := make(chan struct{})
	useMovieData(t, func(int, *models.Anime) error {
		defer close(called)
		return errors.New("jikan: i/o timeout")
	})

	got := fetchMovieData(&models.Anime{MalID: 6654})
	<-called
	select {
	case eps := <-got:
		t.Fatalf("a failed lookup delivered episodes: %+v", eps)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestFetchMovieData_DeliversWithoutTouchingTheAnime(t *testing.T) {
	useMovieData(t, func(id int, a *models.Anime) error {
		assert.Equal(t, 6654, id)
		a.Episodes[0].Synopsis = "A samurai buys a blunt sword."
		return nil
	})

	anime := &models.Anime{MalID: 6654, Episodes: []models.Episode{{Number: "1"}}}
	select {
	case eps := <-fetchMovieData(anime):
		require.Len(t, eps, 1)
		assert.Equal(t, "A samurai buys a blunt sword.", eps[0].Synopsis)
	case <-time.After(time.Second):
		t.Fatal("a successful lookup delivered nothing")
	}
	assert.Empty(t, anime.Episodes[0].Synopsis, "the lookup wrote into the caller's anime")
}

func TestFetchMovieData_SkipsTitlesWithoutJikanData(t *testing.T) {
	useMovieData(t, func(int, *models.Anime) error {
		t.Error("Jikan was queried for a title it has no data for")
		return nil
	})

	for name, anime := range map[string]*models.Anime{
		"no MAL id": {},
		"movie":     {MalID: 1, MediaType: models.MediaTypeMovie},
		"tv":        {MalID: 1, MediaType: models.MediaTypeTV},
	} {
		select {
		case <-fetchMovieData(anime):
			t.Errorf("%s: episodes delivered", name)
		case <-time.After(20 * time.Millisecond):
		}
	}
}
