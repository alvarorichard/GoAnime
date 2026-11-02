//go:build live

package goyabu_test

import (
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/cfsolver"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/goyabu"
)

// TestGoyabuThroughCfsolveLive walks search -> episodes -> stream against the
// real goyabu.io with the cfsolve binary clearing the Cloudflare gate. It skips
// when cfsolve is not installed (GOANIME_CFSOLVE_BIN / PATH).
func TestGoyabuThroughCfsolveLive(t *testing.T) {
	if !cfsolver.Register() {
		t.Skip("cfsolve is not installed; set GOANIME_CFSOLVE_BIN")
	}
	t.Cleanup(func() { netx.RegisterChallengeSolver(nil) })

	c := goyabu.NewGoyabuClient()
	results, err := c.SearchAnime("black clover")
	if err != nil || len(results) == 0 {
		t.Fatalf("search: %d results, err=%v", len(results), err)
	}
	t.Logf("search: %d results, first %q %s", len(results), results[0].Name, results[0].URL)

	var eps []models.Episode
	for _, r := range results {
		got, err := c.GetAnimeEpisodes(r.URL)
		t.Logf("episodes of %s: %d, err=%v", r.URL, len(got), err)
		if err == nil && len(got) > 0 && eps == nil {
			eps = got
		}
	}
	if len(eps) == 0 {
		t.Fatal("no result had episodes")
	}

	stream, err := c.GetEpisodeStreamURL(eps[0].URL)
	if err != nil || stream == "" {
		t.Fatalf("stream: %q, err=%v", stream, err)
	}
	t.Logf("stream: %.80s", stream)
}
