package download

import (
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api"
	"github.com/alvarorichard/Goanime/v2/internal/models"
)

// The -dm dispatch tests: Pomfy titles run through the site's own season
// listing, the StartFlix family keeps the panel. The route is a label, so
// the tests are exact about the mapping without comparing func values.
func TestSeriesListingDispatchIsBySource(t *testing.T) {
	pomfyTitle := &models.Anime{Name: "Breaking Bad", Source: "Pomfy", MediaType: models.MediaTypeTV, TMDBID: 1396}
	sfxTitle := &models.Anime{Name: "Dexter", Source: "StartFlix", MediaType: models.MediaTypeTV}
	anonTitle := &models.Anime{Name: "Anon", MediaType: models.MediaTypeTV}

	cases := []struct {
		name  string
		title *models.Anime
		want  string
	}{
		{"pomfy title", pomfyTitle, "pomfy"},
		{"startflix title", sfxTitle, "panel"},
		{"unknown source", anonTitle, "panel"},
		{"nil title", nil, "panel"},
	}
	for _, tc := range cases {
		if got := seriesListingRoute(tc.title); got != tc.want {
			t.Errorf("%s: route = %q, want %q", tc.name, got, tc.want)
		}
	}
}

// TestDispatchHandsTheRightListing pins the mapping by outcome: the Pomfy
// route must hand out api.GetPomfySeasonNumbers (which refuses a TMDB-less
// title with its own message — no panel, no network), and the panel route
// keeps its function.
func TestDispatchHandsTheRightListing(t *testing.T) {
	pomfyTitle := &models.Anime{Name: "Breaking Bad", Source: "Pomfy"}
	panelTitle := &models.Anime{Name: "Dexter", Source: "StartFlix"}

	// The Pomfy listing refuses an id-less title right away — its error is
	// distinguishable from anything the panel flow says.
	_, pomfyErr := seriesSeasonsFor(pomfyTitle)(pomfyTitle)
	if pomfyErr == nil || !strings.Contains(pomfyErr.Error(), "TMDB id") {
		t.Errorf("Pomfy route listing said: %v, want the Pomfy id refusal", pomfyErr)
	}

	// Panel route keeps a function; its network behavior has its own tests.
	if fn := seriesEpisodesFor(panelTitle); fn == nil {
		t.Error("the panel route must keep a listing function")
	}
	// The wiring hands the api package's Pomfy functions out, not other fns.
	_, directErr := api.GetPomfySeasonNumbers(pomfyTitle)
	if _, routedErr := seriesSeasonsFor(pomfyTitle)(pomfyTitle); strings.Contains(sprintErr(routedErr), "TMDB id") != strings.Contains(sprintErr(directErr), "TMDB id") {
		t.Error("the Pomfy route must hand out api.GetPomfySeasonNumbers itself")
	}
}

func sprintErr(err error) string {
	if err == nil {
		return "<nil>"
	}
	return err.Error()
}
