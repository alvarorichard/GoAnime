package api

// Regression test for the SuperFlix → AniList leak (2026-07-01 debug log):
// SuperFlix tags western animation ("Os Simpsons") as MediaTypeAnime, so the
// old media-type-only guard in enrichAnimeData let it fall through to AniList
// — a query that cannot match TMDB-indexed content and, at the time, also hit
// a Cloudflare challenge (403 + ~6KB of HTML per attempt in the debug log).
// enrichAnimeData must skip AniList for the SuperFlix SOURCE outright,
// mirroring appflow.fetchAnimeDetailsCore.

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
)

func TestEnrichAnimeData_SuperFlixNeverQueriesAniList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("AniList must not be queried for SuperFlix content")
	}))
	t.Cleanup(srv.Close)
	withAniList(t, srv.URL)

	// The SuperFlix trap case: source SuperFlix but MediaType anime (western
	// animation classified as anime by the catalog).
	anime := &models.Anime{
		Name:      "[PT-BR] Os Simpsons",
		Source:    "SuperFlix",
		MediaType: models.MediaTypeAnime,
		URL:       "456",
	}
	// The movie/TV enrichment branch may fail without TMDB/OMDb keys; the only
	// contract under test is that the AniList endpoint is never touched.
	_ = enrichAnimeData(anime)
}
