package scraper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/superflix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuperFlixAdapter_SearchAnime(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<html><body>
			<div class="group/card">
				<img alt="Test Movie" src="https://image.tmdb.org/t/p/w500/test.jpg" />
				<button data-msg="Copiar TMDB" data-copy="100">TMDB</button>
				<button data-msg="Copiar IMDB" data-copy="tt1000000">IMDB</button>
				<button data-msg="Copiar Link" data-copy="http://x.com/filme/100">Link</button>
				<div class="mt-3">2020 | FILME</div>
			</div>
		</body></html>`)
	}))
	defer srv.Close()

	sfClient := superflix.NewClientForTest(srv.URL)
	adapter := &SuperFlixAdapter{client: sfClient}

	results, err := adapter.SearchAnime("test")
	require.NoError(t, err)
	require.Len(t, results, 1)

	anime := results[0]
	assert.Equal(t, "Test Movie", anime.Name)
	assert.Equal(t, "SuperFlix", anime.Source)
	assert.Equal(t, models.MediaTypeMovie, anime.MediaType)
	assert.Equal(t, 100, anime.TMDBID)
	assert.Equal(t, "tt1000000", anime.IMDBID)
	assert.Equal(t, "https://image.tmdb.org/t/p/w500/test.jpg", anime.ImageURL)
}

func TestSuperFlixAdapter_SearchAnime_Error(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	sfClient := superflix.NewClientForTest(srv.URL)
	adapter := &SuperFlixAdapter{client: sfClient}

	_, err := adapter.SearchAnime("test")
	require.Error(t, err)
}

func TestSuperFlixAdapter_GetType(t *testing.T) {
	t.Parallel()

	adapter := &SuperFlixAdapter{client: superflix.NewSuperFlixClient()}
	assert.Equal(t, SuperFlixType, adapter.GetType())
}

func TestSuperFlixAdapter_GetAnimeEpisodes_ReturnsError(t *testing.T) {
	t.Parallel()

	adapter := &SuperFlixAdapter{client: superflix.NewSuperFlixClient()}
	_, err := adapter.GetAnimeEpisodes("1405")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SuperFlix")
}

func TestSuperFlixAdapter_GetClient(t *testing.T) {
	t.Parallel()

	inner := superflix.NewSuperFlixClient()
	adapter := &SuperFlixAdapter{client: inner}
	assert.Equal(t, inner, adapter.GetClient())
}

// =============================================================================
// Unit Tests: SuperFlixAdapter.GetStreamURL with mock pipeline
// =============================================================================

func TestSuperFlixAdapter_GetStreamURL_FullMock(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/filme/100", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `var CSRF_TOKEN = "c"; var PAGE_TOKEN = "p"; var INITIAL_CONTENT_ID = 100; var CONTENT_TYPE = "filme";
		<title>Player | Test Movie</title>`)
	})
	mux.HandleFunc("/player/bootstrap", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"options":[{"ID":"sv1","name":"Main"}]}}`)
	})
	mux.HandleFunc("/player/source", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":{"video_url":"%s/video/testhash"}}`, srv.URL)
	})
	mux.HandleFunc("/video/testhash", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<html>
			var defaultAudio = ["Portuguese"];
			var playerjsSubtitle = "[PT-BR]https://subs.example.com/pt.vtt";
		</html>`)
	})
	mux.HandleFunc("/player/index.php", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"securedLink":"https://cdn.example.com/movie.m3u8","videoImage":""}`)
	})

	sfClient := superflix.NewClientForTest(srv.URL)
	adapter := &SuperFlixAdapter{client: sfClient}

	streamURL, metadata, err := adapter.GetStreamURL("100", "filme")
	require.NoError(t, err)

	assert.Equal(t, "https://cdn.example.com/movie.m3u8", streamURL)
	assert.Equal(t, "superflix", metadata["source"])
	assert.Equal(t, "Test Movie", metadata["title"])
	assert.Equal(t, "Portuguese", metadata["audio_lang"])
	assert.NotEmpty(t, metadata["subtitles"])
	assert.NotEmpty(t, metadata["subtitle_labels"])
}

func TestSuperFlixAdapter_GetStreamURL_SeriesWithSeasonEpisode(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/serie/1405/1/3", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `var CSRF_TOKEN = "c"; var PAGE_TOKEN = "p"; var INITIAL_CONTENT_ID = 1405; var CONTENT_TYPE = "serie";
		<title>Dexter S01E03</title>`)
	})
	mux.HandleFunc("/player/bootstrap", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"options":[{"ID":"sv1","name":"Main"}]}}`)
	})
	mux.HandleFunc("/player/source", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":{"video_url":"%s/video/epihash"}}`, srv.URL)
	})
	mux.HandleFunc("/video/epihash", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<html></html>`)
	})
	mux.HandleFunc("/player/index.php", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"securedLink":"https://cdn.example.com/dexter-s1e3.m3u8"}`)
	})

	sfClient := superflix.NewClientForTest(srv.URL)
	adapter := &SuperFlixAdapter{client: sfClient}

	streamURL, metadata, err := adapter.GetStreamURL("1405", "serie", "1", "3")
	require.NoError(t, err)

	assert.Equal(t, "https://cdn.example.com/dexter-s1e3.m3u8", streamURL)
	assert.Equal(t, "superflix", metadata["source"])
}

// =============================================================================
// Unit Tests: sortPTBRFirst
// =============================================================================

func TestSortPTBRFirst(t *testing.T) {
	t.Parallel()

	results := []*models.Anime{
		{Name: "[English] Show A"},
		{Name: "[PT-BR] Show B"},
		{Name: "[English] Show C"},
		{Name: "[PT-BR] Show D"},
	}

	sortPTBRFirst(results)

	// PT-BR should come first
	assert.Contains(t, results[0].Name, "[PT-BR]")
	assert.Contains(t, results[1].Name, "[PT-BR]")
	assert.Contains(t, results[2].Name, "[English]")
	assert.Contains(t, results[3].Name, "[English]")
}

func TestSortPTBRFirst_AllPTBR(t *testing.T) {
	t.Parallel()

	results := []*models.Anime{
		{Name: "[PT-BR] A"},
		{Name: "[PT-BR] B"},
	}

	sortPTBRFirst(results)

	// Order should be preserved
	assert.Equal(t, "[PT-BR] A", results[0].Name)
	assert.Equal(t, "[PT-BR] B", results[1].Name)
}

func TestSortPTBRFirst_NoPTBR(t *testing.T) {
	t.Parallel()

	results := []*models.Anime{
		{Name: "[English] A"},
		{Name: "[English] B"},
	}

	sortPTBRFirst(results)

	// Order should be preserved
	assert.Equal(t, "[English] A", results[0].Name)
	assert.Equal(t, "[English] B", results[1].Name)
}

// =============================================================================
// Regex Tests: Pre-compiled regex patterns
// =============================================================================
