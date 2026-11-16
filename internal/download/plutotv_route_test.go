package download

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alvarorichard/Goanime/v2/internal/api"
	"github.com/alvarorichard/Goanime/v2/internal/models"
)

// funcName is the name of a function value, to tell which one a route chose.
func funcName(fn any) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}

// -dm lists a series' seasons through the route its source names: Pluto TV's
// own JSON, not the StartFlix panel the rest of the family goes through.
func TestPlutoTVSeriesListThroughPlutoTV(t *testing.T) {
	title := &models.Anime{Source: "PlutoTV", MediaType: models.MediaTypeTV}
	assert.Equal(t, "plutotv", seriesListingRoute(title))
	assert.Equal(t, funcName(api.GetPlutoTVSeasonNumbers), funcName(seriesSeasonsFor(title)))
	assert.Equal(t, funcName(api.GetPlutoTVSeasonEpisodes), funcName(seriesEpisodesFor(title)))
}

func TestTheOtherRoutesAreWhereTheyWere(t *testing.T) {
	pomfy := &models.Anime{Source: "Pomfy"}
	assert.Equal(t, "pomfy", seriesListingRoute(pomfy))
	assert.Equal(t, funcName(api.GetPomfySeasonNumbers), funcName(seriesSeasonsFor(pomfy)))
	assert.Equal(t, funcName(api.GetPomfySeasonEpisodes), funcName(seriesEpisodesFor(pomfy)))

	for _, source := range []string{"StartFlix", "TopCine", "Pobreflix", "", "plutotv"} {
		title := &models.Anime{Source: source}
		assert.Equal(t, "panel", seriesListingRoute(title), "%q: the route is by the exact source name", source)
		assert.Equal(t, funcName(seriesSeasonsFn), funcName(seriesSeasonsFor(title)))
		assert.Equal(t, funcName(seriesEpisodesFn), funcName(seriesEpisodesFor(title)))
	}
	assert.Equal(t, "panel", seriesListingRoute(nil))
}
