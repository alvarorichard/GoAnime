package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Helper: newTestSuperFlixClient creates a client pointing at the test server
// =============================================================================

func newTestSuperFlixClient(serverURL string) *SuperFlixClient {
	c := NewSuperFlixClient()
	c.baseURL = serverURL
	c.client = &http.Client{Timeout: 5 * time.Second, Transport: http.DefaultTransport} // bypass SSRF-safe transport for localhost
	c.maxRetries = 0
	c.retryDelay = 0
	return c
}

// =============================================================================
// Unit Tests: NewSuperFlixClient defaults
// =============================================================================

func TestNewSuperFlixClient_Defaults(t *testing.T) {
	t.Parallel()
	c := NewSuperFlixClient()

	assert.Equal(t, SuperFlixBase, c.baseURL)
	assert.Equal(t, SuperFlixUserAgent, c.userAgent)
	assert.Equal(t, 2, c.maxRetries)
	assert.Equal(t, 200*time.Millisecond, c.retryDelay)
	assert.NotNil(t, c.client)
}

// =============================================================================
// Unit Tests: ExtractTokens
// =============================================================================

func TestExtractTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		html     string
		expected *SuperFlixTokens
	}{
		{
			name: "all tokens present",
			html: `<script>
				var CSRF_TOKEN = "abc123csrf";
				var PAGE_TOKEN = "page_tok_456";
				var INITIAL_CONTENT_ID = 98765;
				var CONTENT_TYPE = "serie";
				<title>Player | Breaking Bad</title>
			</script>`,
			expected: &SuperFlixTokens{
				CSRF:        "abc123csrf",
				PageToken:   "page_tok_456",
				ContentID:   "98765",
				ContentType: "serie",
				Title:       "Breaking Bad",
			},
		},
		{
			name: "title without Player prefix",
			html: `<title>Dexter</title>`,
			expected: &SuperFlixTokens{
				Title: "Dexter",
			},
		},
		{
			name:     "empty HTML returns empty tokens",
			html:     "",
			expected: &SuperFlixTokens{},
		},
		{
			name:     "malformed HTML returns empty tokens",
			html:     `var CSRF_TOKEN = ;var PAGE_TOKEN = `,
			expected: &SuperFlixTokens{},
		},
		{
			name: "tokens with spaces around equals",
			html: `var CSRF_TOKEN  =  "spaced_csrf";
			       var PAGE_TOKEN  =  "spaced_page";
			       var INITIAL_CONTENT_ID  =  42;
			       var CONTENT_TYPE  =  "filme";`,
			expected: &SuperFlixTokens{
				CSRF:        "spaced_csrf",
				PageToken:   "spaced_page",
				ContentID:   "42",
				ContentType: "filme",
			},
		},
		{
			name: "partial tokens - only CSRF and PageToken",
			html: `var CSRF_TOKEN = "only_csrf";
			       var PAGE_TOKEN = "only_page";`,
			expected: &SuperFlixTokens{
				CSRF:      "only_csrf",
				PageToken: "only_page",
			},
		},
	}

	client := NewSuperFlixClient()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tokens := client.ExtractTokens(tc.html)
			assert.Equal(t, tc.expected.CSRF, tokens.CSRF)
			assert.Equal(t, tc.expected.PageToken, tokens.PageToken)
			assert.Equal(t, tc.expected.ContentID, tokens.ContentID)
			assert.Equal(t, tc.expected.ContentType, tokens.ContentType)
			assert.Equal(t, tc.expected.Title, tokens.Title)
		})
	}
}

// =============================================================================
// Unit Tests: ExtractEpisodes
// =============================================================================

func TestExtractEpisodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		html         string
		expectNil    bool
		expectError  bool
		expectKeys   []string
		expectCounts map[string]int
	}{
		{
			name:         "valid ALL_EPISODES JSON",
			html:         `var ALL_EPISODES = {"1":[{"epi_num":"1","title":"Pilot","air_date":"2008-01-20"},{"epi_num":"2","title":"Cat's in the Bag","air_date":"2008-01-27"}],"2":[{"epi_num":"1","title":"Seven Thirty-Seven","air_date":"2009-03-08"}]};`,
			expectKeys:   []string{"1", "2"},
			expectCounts: map[string]int{"1": 2, "2": 1},
		},
		{
			name:      "no ALL_EPISODES variable",
			html:      `<script>var OTHER_VAR = "something";</script>`,
			expectNil: true,
		},
		{
			name:        "malformed JSON in ALL_EPISODES",
			html:        `var ALL_EPISODES = {invalid json};`,
			expectError: true,
		},
		{
			name:      "empty episodes object",
			html:      `var ALL_EPISODES = {};`,
			expectNil: true, // regex requires at least one char between { and }
		},
	}

	client := NewSuperFlixClient()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := client.ExtractEpisodes(tc.html)

			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tc.expectNil {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			assert.Len(t, result, len(tc.expectKeys))
			for _, key := range tc.expectKeys {
				episodes, exists := result[key]
				assert.True(t, exists, "missing season key: %s", key)
				if expected, ok := tc.expectCounts[key]; ok {
					assert.Len(t, episodes, expected, "wrong episode count for season %s", key)
				}
			}
		})
	}
}

func TestExtractEpisodes_EpisodeFields(t *testing.T) {
	t.Parallel()

	html := `var ALL_EPISODES = {"1":[{"epi_num":"5","title":"Gray Matter","air_date":"2008-02-24"}]};`
	client := NewSuperFlixClient()
	result, err := client.ExtractEpisodes(html)
	require.NoError(t, err)
	require.Len(t, result["1"], 1)

	ep := result["1"][0]
	assert.Equal(t, json.Number("5"), ep.EpiNum)
	assert.Equal(t, "Gray Matter", ep.Title)
	assert.Equal(t, "2008-02-24", ep.AirDate)
}

// =============================================================================
// Unit Tests: ExtractPlayerExtras (subtitles & audio)
// =============================================================================

func TestExtractPlayerExtras(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		html           string
		expectAudio    []string
		expectSubCount int
		expectSubs     []SuperFlixSubtitle
	}{
		{
			name: "both audio and subtitles present",
			html: `var defaultAudio = ["Portuguese"];
			       var playerjsSubtitle = "[Portuguese]https://subs.example.com/pt.vtt,[English]https://subs.example.com/en.vtt";`,
			expectAudio:    []string{"Portuguese"},
			expectSubCount: 2,
			expectSubs: []SuperFlixSubtitle{
				{Lang: "Portuguese", URL: "https://subs.example.com/pt.vtt"},
				{Lang: "English", URL: "https://subs.example.com/en.vtt"},
			},
		},
		{
			name:           "only audio, no subtitles",
			html:           `var defaultAudio = ["Japanese","English"];`,
			expectAudio:    []string{"Japanese", "English"},
			expectSubCount: 0,
		},
		{
			name:           "only subtitles, no audio",
			html:           `var playerjsSubtitle = "[Spanish]https://subs.example.com/es.vtt";`,
			expectSubCount: 1,
			expectSubs: []SuperFlixSubtitle{
				{Lang: "Spanish", URL: "https://subs.example.com/es.vtt"},
			},
		},
		{
			name:           "neither audio nor subtitles",
			html:           `<script>var nothing = true;</script>`,
			expectSubCount: 0,
		},
		{
			name:           "empty audio array",
			html:           `var defaultAudio = [];`,
			expectAudio:    nil, // JSON Unmarshal of [] into []string yields nil
			expectSubCount: 0,
		},
	}

	client := NewSuperFlixClient()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			audio, subs := client.ExtractPlayerExtras(tc.html)

			if tc.expectAudio != nil {
				assert.Equal(t, tc.expectAudio, audio)
			} else {
				assert.Nil(t, audio)
			}

			assert.Len(t, subs, tc.expectSubCount)
			if tc.expectSubs != nil {
				for i, expected := range tc.expectSubs {
					assert.Equal(t, expected.Lang, subs[i].Lang)
					assert.Equal(t, expected.URL, subs[i].URL)
				}
			}
		})
	}
}

// =============================================================================
// Unit Tests: splitAndTrim helper
// =============================================================================

func TestSplitAndTrim(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		sep      string
		expected []string
	}{
		{
			name:     "normal pipe-separated",
			input:    "2006 | SÉRIE",
			sep:      "|",
			expected: []string{"2006", "SÉRIE"},
		},
		{
			name:     "extra whitespace",
			input:    "  2010  |  FILME  ",
			sep:      "|",
			expected: []string{"2010", "FILME"},
		},
		{
			name:     "empty parts are filtered",
			input:    "a || b",
			sep:      "|",
			expected: []string{"a", "b"},
		},
		{
			name:     "single value no separator",
			input:    "ANIME",
			sep:      "|",
			expected: []string{"ANIME"},
		},
		{
			name:     "empty string",
			input:    "",
			sep:      "|",
			expected: nil,
		},
		{
			name:     "only separators",
			input:    "|||",
			sep:      "|",
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := splitAndTrim(tc.input, tc.sep)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// =============================================================================
// Unit Tests: ToAnimeModel
// =============================================================================

func TestToAnimeModel_AllMediaTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		media        SuperFlixMedia
		expectType   models.MediaType
		expectSource string
	}{
		{
			name:         "filme -> MediaTypeMovie",
			media:        SuperFlixMedia{Title: "Inception", SFType: "filme", Type: "Filme"},
			expectType:   models.MediaTypeMovie,
			expectSource: "SuperFlix",
		},
		{
			name:         "serie -> MediaTypeTV",
			media:        SuperFlixMedia{Title: "Breaking Bad", SFType: "serie", Type: "Série"},
			expectType:   models.MediaTypeTV,
			expectSource: "SuperFlix",
		},
		{
			name:         "anime type -> MediaTypeAnime",
			media:        SuperFlixMedia{Title: "Dexter Lab", SFType: "serie", Type: "Anime"},
			expectType:   models.MediaTypeAnime,
			expectSource: "SuperFlix",
		},
		{
			name:         "dorama type -> MediaTypeAnime",
			media:        SuperFlixMedia{Title: "Squid Game", SFType: "serie", Type: "Dorama"},
			expectType:   models.MediaTypeAnime,
			expectSource: "SuperFlix",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			anime := tc.media.ToAnimeModel()
			assert.Equal(t, tc.expectType, anime.MediaType)
			assert.Equal(t, tc.expectSource, anime.Source)
			assert.Equal(t, tc.media.Title, anime.Name)
		})
	}
}

func TestToAnimeModel_TMDBIDParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tmdbID   string
		expected int
	}{
		{"valid ID", "1405", 1405},
		{"large ID", "999999", 999999},
		{"zero", "0", 0},
		{"non-numeric", "abc", 0},
		{"empty", "", 0},
		{"negative", "-1", -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			media := &SuperFlixMedia{Title: "Test", SFType: "serie", TMDBID: tc.tmdbID}
			anime := media.ToAnimeModel()
			assert.Equal(t, tc.expected, anime.TMDBID)
		})
	}
}

func TestToAnimeModel_IMDBIDPreserved(t *testing.T) {
	t.Parallel()

	media := &SuperFlixMedia{
		Title:  "Test",
		SFType: "filme",
		IMDBID: "tt1234567",
	}
	anime := media.ToAnimeModel()
	assert.Equal(t, "tt1234567", anime.IMDBID)
}

func TestToAnimeModel_URLIsTMDBID(t *testing.T) {
	t.Parallel()

	media := &SuperFlixMedia{
		Title:  "Test",
		SFType: "filme",
		TMDBID: "27205",
	}
	anime := media.ToAnimeModel()
	assert.Equal(t, "27205", anime.URL, "URL field should store TMDB ID")
}

func TestToAnimeModel_YearPreserved(t *testing.T) {
	t.Parallel()

	media := &SuperFlixMedia{
		Title:  "Test",
		SFType: "filme",
		Year:   "2024",
	}
	anime := media.ToAnimeModel()
	assert.Equal(t, "2024", anime.Year)
}

// =============================================================================
// HTTP Mock Tests: SearchMedia
// =============================================================================

func TestSearchMedia_Success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/pesquisar", r.URL.Path)
		assert.Equal(t, "test query", r.URL.Query().Get("s"))

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<html><body>
			<div class="group/card">
				<img alt="Test Show" src="https://image.tmdb.org/t/p/w342/poster.jpg" />
				<button data-msg="Copiar TMDB" data-copy="12345">TMDB</button>
				<button data-msg="Copiar IMDB" data-copy="tt9999999">IMDB</button>
				<button data-msg="Copiar Link" data-copy="http://example.com/serie/12345">Link</button>
				<div class="mt-3">PG-13 | 2024 | SÉRIE</div>
			</div>
		</body></html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	results, err := client.SearchMedia("test query")

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Test Show", results[0].Title)
	assert.Equal(t, "12345", results[0].TMDBID)
	assert.Equal(t, "tt9999999", results[0].IMDBID)
	assert.Equal(t, "serie", results[0].SFType)
	assert.Equal(t, "2024", results[0].Year)
}

func TestSearchMedia_EmptyResults(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<html><body><div class="no-results">Nenhum resultado</div></body></html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	results, err := client.SearchMedia("nonexistent_xyzzy_12345")

	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearchMedia_ServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	_, err := client.SearchMedia("test")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server returned")
}

func TestSearchMedia_InvalidHTML(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		// goquery can handle malformed HTML gracefully
		fmt.Fprint(w, `<html><body><div class="group/card"><h3>Broken`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	results, err := client.SearchMedia("broken")

	// goquery is lenient; it should still parse what it can
	require.NoError(t, err)
	// The card has no title extracted properly due to broken HTML - may or may not have results
	_ = results
}

func TestSearchMedia_Caching(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount.Add(1)
		fmt.Fprint(w, `<html><body>
			<div class="group/card">
				<img alt="Cached Show" src="https://image.tmdb.org/t/p/w500/cached.jpg" />
				<button data-msg="Copiar TMDB" data-copy="111">TMDB</button>
				<button data-msg="Copiar Link" data-copy="http://example.com/serie/111">Link</button>
				<div class="mt-3">2024 | SÉRIE</div>
			</div>
		</body></html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)

	// First call hits server
	results1, err1 := client.SearchMedia("cached test")
	require.NoError(t, err1)
	require.Len(t, results1, 1)

	// Second call should use cache (case-insensitive)
	results2, err2 := client.SearchMedia("CACHED TEST")
	require.NoError(t, err2)
	require.Len(t, results2, 1)

	// Server should only be called once
	assert.Equal(t, int32(1), callCount.Load(), "second call should hit cache, not server")
}

func TestSearchMedia_CacheCaseInsensitive(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount.Add(1)
		fmt.Fprint(w, `<html><body>
			<div class="group/card">
				<h3>Result</h3>
				<button data-msg="Copiar TMDB" data-copy="1">TMDB</button>
				<button data-msg="Copiar Link" data-copy="http://x.com/serie/1">Link</button>
				<div class="mt-3">2024 | SÉRIE</div>
			</div>
		</body></html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)

	_, _ = client.SearchMedia("  Naruto  ")
	_, _ = client.SearchMedia("naruto")

	assert.Equal(t, int32(1), callCount.Load())
}

func TestSearchMediaWithContext_Cancelled(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(5 * time.Second) // slow server
		fmt.Fprint(w, `<html><body></body></html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.SearchMediaWithContext(ctx, "slow query")
	require.Error(t, err)
}

func TestSearchMedia_DeduplicatesByTMDBID(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<html><body>
			<div class="group/card">
				<h3>Show A</h3>
				<button data-msg="Copiar TMDB" data-copy="555">TMDB</button>
				<button data-msg="Copiar Link" data-copy="http://x.com/serie/555">Link</button>
				<div class="mt-3">2024 | SÉRIE</div>
			</div>
			<div class="group/card">
				<h3>Show A Duplicate</h3>
				<button data-msg="Copiar TMDB" data-copy="555">TMDB</button>
				<button data-msg="Copiar Link" data-copy="http://x.com/serie/555">Link</button>
				<div class="mt-3">2024 | SÉRIE</div>
			</div>
		</body></html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	results, err := client.SearchMedia("dupes")

	require.NoError(t, err)
	assert.Len(t, results, 1, "duplicate TMDB IDs should be deduplicated")
	assert.Equal(t, "Show A", results[0].Title)
}

func TestSearchMedia_DeduplicatesByTitle(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<html><body>
			<div class="group/card">
				<h3>No TMDB Show</h3>
				<button data-msg="Copiar Link" data-copy="http://x.com/serie/1">Link</button>
				<div class="mt-3">2024 | SÉRIE</div>
			</div>
			<div class="group/card">
				<h3>No TMDB Show</h3>
				<button data-msg="Copiar Link" data-copy="http://x.com/serie/2">Link</button>
				<div class="mt-3">2024 | SÉRIE</div>
			</div>
		</body></html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	results, err := client.SearchMedia("dupes")

	require.NoError(t, err)
	assert.Len(t, results, 1, "cards without TMDB ID should dedup by title")
}

func TestSearchMedia_CardWithoutTitle_Skipped(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<html><body>
			<div class="group/card">
				<button data-msg="Copiar TMDB" data-copy="999">TMDB</button>
				<div class="mt-3">2024 | SÉRIE</div>
			</div>
		</body></html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	results, err := client.SearchMedia("empty")

	require.NoError(t, err)
	assert.Empty(t, results, "cards without title should be skipped")
}

func TestSearchMedia_TypeDetection(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<html><body>
			<div class="group/card">
				<h3>Movie</h3>
				<button data-msg="Copiar TMDB" data-copy="1">TMDB</button>
				<button data-msg="Copiar Link" data-copy="http://x.com/filme/1">Link</button>
				<div class="mt-3">2024 | FILME</div>
			</div>
			<div class="group/card">
				<h3>Series</h3>
				<button data-msg="Copiar TMDB" data-copy="2">TMDB</button>
				<button data-msg="Copiar Link" data-copy="http://x.com/serie/2">Link</button>
				<div class="mt-3">2024 | SÉRIE</div>
			</div>
		</body></html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	results, err := client.SearchMedia("types")

	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "filme", results[0].SFType)
	assert.Equal(t, "FILME", results[0].Type) // raw text from meta
	assert.Equal(t, "serie", results[1].SFType)
}

func TestSearchMedia_TypeFallback_NoMeta(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<html><body>
			<div class="group/card">
				<h3>Movie No Meta</h3>
				<button data-msg="Copiar TMDB" data-copy="1">TMDB</button>
				<button data-msg="Copiar Link" data-copy="http://x.com/filme/1">Link</button>
			</div>
			<div class="group/card">
				<h3>Serie No Meta</h3>
				<button data-msg="Copiar TMDB" data-copy="2">TMDB</button>
				<button data-msg="Copiar Link" data-copy="http://x.com/serie/2">Link</button>
			</div>
		</body></html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	results, err := client.SearchMedia("fallback")

	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "Filme", results[0].Type, "should fallback to Filme for /filme/ URL")
	assert.Equal(t, "Série", results[1].Type, "should fallback to Série for /serie/ URL")
}

// =============================================================================
// HTTP Mock Tests: GetPlayerPage
// =============================================================================

func TestGetPlayerPage_Movie(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/filme/27205", r.URL.Path)
		assert.Contains(t, r.Header.Get("User-Agent"), "Mozilla")
		fmt.Fprint(w, `<html>var CSRF_TOKEN = "test";</html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	html, err := client.GetPlayerPage(context.Background(), "filme", "27205", "", "")

	require.NoError(t, err)
	assert.Contains(t, html, "CSRF_TOKEN")
}

func TestGetPlayerPage_SeriesWithSeasonAndEpisode(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/serie/1405/2/5", r.URL.Path)
		fmt.Fprint(w, `<html>season 2 episode 5</html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	html, err := client.GetPlayerPage(context.Background(), "serie", "1405", "2", "5")

	require.NoError(t, err)
	assert.Contains(t, html, "season 2 episode 5")
}

func TestGetPlayerPage_Cancelled(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.GetPlayerPage(ctx, "filme", "1", "", "")
	require.Error(t, err)
}

// =============================================================================
// HTTP Mock Tests: Bootstrap
// =============================================================================

func TestBootstrap_Success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/player/bootstrap", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		assert.Equal(t, "XMLHttpRequest", r.Header.Get("X-Requested-With"))

		err := r.ParseForm()
		require.NoError(t, err)
		assert.Equal(t, "content123", r.FormValue("contentid"))
		assert.Equal(t, "serie", r.FormValue("type"))
		assert.Equal(t, "csrf_tok", r.FormValue("_token"))
		assert.Equal(t, "page_tok", r.FormValue("page_token"))

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"options":[{"ID":"sv1","name":"Server 1"},{"ID":2,"name":"Server 2"}]}}`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	tokens := &SuperFlixTokens{
		CSRF:        "csrf_tok",
		PageToken:   "page_tok",
		ContentID:   "content123",
		ContentType: "serie",
	}

	servers, err := client.Bootstrap(context.Background(), tokens)

	require.NoError(t, err)
	require.Len(t, servers, 2)
	assert.Equal(t, "Server 1", servers[0].Name)
	assert.Equal(t, "Server 2", servers[1].Name)
}

func TestBootstrap_EmptyServerList(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"options":[]}}`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	tokens := &SuperFlixTokens{CSRF: "a", PageToken: "b", ContentID: "1", ContentType: "filme"}
	servers, err := client.Bootstrap(context.Background(), tokens)

	require.NoError(t, err)
	assert.Empty(t, servers)
}

func TestBootstrap_InvalidJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `not json`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	tokens := &SuperFlixTokens{CSRF: "a", PageToken: "b", ContentID: "1", ContentType: "filme"}
	_, err := client.Bootstrap(context.Background(), tokens)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode bootstrap response")
}

// =============================================================================
// HTTP Mock Tests: GetSourceURL
// =============================================================================

func TestGetSourceURL_Success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/player/source", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		err := r.ParseForm()
		require.NoError(t, err)
		assert.Equal(t, "vid123", r.FormValue("video_id"))
		assert.Equal(t, "page_tok", r.FormValue("page_token"))

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"video_url":"https://redirect.example.com/goto"}}`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	tokens := &SuperFlixTokens{CSRF: "csrf", PageToken: "page_tok"}
	videoURL, err := client.GetSourceURL(context.Background(), "vid123", tokens)

	require.NoError(t, err)
	assert.Equal(t, "https://redirect.example.com/goto", videoURL)
}

func TestGetSourceURL_EmptyVideoURL(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"video_url":""}}`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	tokens := &SuperFlixTokens{CSRF: "a", PageToken: "b"}
	_, err := client.GetSourceURL(context.Background(), "vid", tokens)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no video URL")
}

func TestGetSourceURL_InvalidJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{bad json}`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	tokens := &SuperFlixTokens{CSRF: "a", PageToken: "b"}
	_, err := client.GetSourceURL(context.Background(), "vid", tokens)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode source response")
}

// =============================================================================
// HTTP Mock Tests: GetVideoAPI
// =============================================================================

func TestGetVideoAPI_SecuredLink(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/player/index.php")
		assert.Equal(t, "getVideo", r.URL.Query().Get("do"))
		assert.Equal(t, "hashABC", r.URL.Query().Get("data"))
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "XMLHttpRequest", r.Header.Get("X-Requested-With"))

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"securedLink":"https://cdn.example.com/stream.m3u8","videoImage":"https://img.example.com/thumb.jpg"}`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	streamURL, thumbURL, err := client.GetVideoAPI(context.Background(), srv.URL, "hashABC", srv.URL+"/video/hashABC")

	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/stream.m3u8", streamURL)
	assert.Equal(t, "https://img.example.com/thumb.jpg", thumbURL)
}

func TestGetVideoAPI_FallbackToVideoSource(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"videoSource":"https://fallback.example.com/video.mp4","videoImage":""}`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	streamURL, _, err := client.GetVideoAPI(context.Background(), srv.URL, "hash", srv.URL+"/")

	require.NoError(t, err)
	assert.Equal(t, "https://fallback.example.com/video.mp4", streamURL)
}

func TestGetVideoAPI_NoStreamURL(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"securedLink":"","videoSource":"","videoImage":""}`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	_, _, err := client.GetVideoAPI(context.Background(), srv.URL, "hash", srv.URL+"/")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no stream URL in video API response")
}

func TestGetVideoAPI_InvalidJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<html>Error</html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	_, _, err := client.GetVideoAPI(context.Background(), srv.URL, "hash", srv.URL+"/")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode video API response")
}

// =============================================================================
// HTTP Mock Tests: ResolveRedirect
// =============================================================================

func TestResolveRedirect_FollowsRedirect(t *testing.T) {
	t.Parallel()

	// Final destination server
	finalSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The actual player page
		fmt.Fprint(w, `<html>player content</html>`)
	}))
	defer finalSrv.Close()

	// Redirect server
	redirectSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", finalSrv.URL+"/video/abc123hash")
		w.WriteHeader(http.StatusFound)
	}))
	defer redirectSrv.Close()

	client := newTestSuperFlixClient(redirectSrv.URL)
	baseURL, videoHash, playerHTML, err := client.ResolveRedirect(context.Background(), redirectSrv.URL+"/redirect")

	require.NoError(t, err)
	assert.NotEmpty(t, baseURL)
	assert.Equal(t, "abc123hash", videoHash)
	assert.Contains(t, playerHTML, "player content")
}

func TestResolveRedirect_NoRedirect(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<html>direct page</html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	_, _, html, err := client.ResolveRedirect(context.Background(), srv.URL+"/video/directhash")

	require.NoError(t, err)
	assert.Contains(t, html, "direct page")
}

// =============================================================================
// HTTP Mock Tests: Full GetStreamURL pipeline
// =============================================================================

func TestGetStreamURL_FullPipeline(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Step 1: Player page with tokens
	mux.HandleFunc("/filme/27205", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, `<html>
			<script>
				var CSRF_TOKEN = "test_csrf";
				var PAGE_TOKEN = "test_page_token";
				var INITIAL_CONTENT_ID = 27205;
				var CONTENT_TYPE = "filme";
			</script>
			<title>Player | Inception</title>
		</html>`)
	})

	// Step 2: Bootstrap returns servers
	mux.HandleFunc("/player/bootstrap", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"options":[{"ID":"server1","name":"Primary"}]}}`)
	})

	// Step 3: Source returns redirect URL
	mux.HandleFunc("/player/source", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":{"video_url":"%s/video/hash123"}}`, srv.URL)
	})

	// Step 4: The "external player" page (redirect target)
	mux.HandleFunc("/video/hash123", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<html>
			<script>
				var defaultAudio = ["Portuguese"];
				var playerjsSubtitle = "[Portuguese]https://subs.example.com/pt.vtt";
			</script>
		</html>`)
	})

	// Step 5: Video API returns stream
	mux.HandleFunc("/player/index.php", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"securedLink": "https://cdn.example.com/inception.m3u8",
			"videoImage": "https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/inception_thumb.jpg"
		}`)
	})

	client := newTestSuperFlixClient(srv.URL)
	result, err := client.GetStreamURL(context.Background(), "filme", "27205", "", "")

	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/inception.m3u8", result.StreamURL)
	assert.Equal(t, "Inception", result.Title)
	assert.NotContains(t, result.Thumb, "cloudfront.net", "thumb must be normalized")
	assert.Equal(t, "https://image.tmdb.org/t/p/w500/inception_thumb.jpg", result.Thumb)
	assert.Equal(t, []string{"Portuguese"}, result.DefaultAudio)
	require.Len(t, result.Subtitles, 1)
	assert.Equal(t, "Portuguese", result.Subtitles[0].Lang)
	assert.Equal(t, "https://subs.example.com/pt.vtt", result.Subtitles[0].URL)
}

func TestGetStreamURL_MissingTokens(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<html>no tokens here</html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	_, err := client.GetStreamURL(context.Background(), "filme", "1", "", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to extract tokens")
}

func TestGetStreamURL_NoServers(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/filme/1", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `var CSRF_TOKEN = "c"; var PAGE_TOKEN = "p"; var INITIAL_CONTENT_ID = 1; var CONTENT_TYPE = "filme";`)
	})
	mux.HandleFunc("/player/bootstrap", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"options":[]}}`)
	})

	client := newTestSuperFlixClient(srv.URL)
	_, err := client.GetStreamURL(context.Background(), "filme", "1", "", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no servers available")
}

func TestGetStreamURL_FallbackServer(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/filme/1", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `var CSRF_TOKEN = "c"; var PAGE_TOKEN = "p"; var INITIAL_CONTENT_ID = 1; var CONTENT_TYPE = "filme";`)
	})
	mux.HandleFunc("/player/bootstrap", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Only fallback servers available
		fmt.Fprint(w, `{"data":{"options":[{"ID":"fallback1","name":"Fallback Server"}]}}`)
	})
	mux.HandleFunc("/player/source", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":{"video_url":"%s/video/fallback_hash"}}`, srv.URL)
	})
	mux.HandleFunc("/video/fallback_hash", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<html></html>`)
	})
	mux.HandleFunc("/player/index.php", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"securedLink":"https://cdn.example.com/fallback.m3u8"}`)
	})

	client := newTestSuperFlixClient(srv.URL)
	result, err := client.GetStreamURL(context.Background(), "filme", "1", "", "")

	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/fallback.m3u8", result.StreamURL)
}

func TestGetStreamURL_NumericServerID(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/filme/1", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `var CSRF_TOKEN = "c"; var PAGE_TOKEN = "p"; var INITIAL_CONTENT_ID = 1; var CONTENT_TYPE = "filme";`)
	})
	mux.HandleFunc("/player/bootstrap", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"options":[{"ID":42,"name":"Numeric Server"}]}}`)
	})
	mux.HandleFunc("/player/source", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		assert.Equal(t, "42", r.FormValue("video_id"))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":{"video_url":"%s/video/num_hash"}}`, srv.URL)
	})
	mux.HandleFunc("/video/num_hash", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<html></html>`)
	})
	mux.HandleFunc("/player/index.php", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"securedLink":"https://cdn.example.com/numeric.m3u8"}`)
	})

	client := newTestSuperFlixClient(srv.URL)
	result, err := client.GetStreamURL(context.Background(), "filme", "1", "", "")

	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/numeric.m3u8", result.StreamURL)
}

// =============================================================================
// HTTP Mock Tests: GetEpisodes
// =============================================================================

func TestGetEpisodes_Success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/serie/1405", r.URL.Path)
		fmt.Fprint(w, `<html><script>
			var ALL_EPISODES = {"1":[{"epi_num":"1","title":"Pilot","air_date":"2006-10-01"}],"2":[{"epi_num":"1","title":"S2E1","air_date":"2007-09-30"}]};
		</script></html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	episodes, err := client.GetEpisodes(context.Background(), "1405")

	require.NoError(t, err)
	require.Len(t, episodes, 2)
	assert.Len(t, episodes["1"], 1)
	assert.Len(t, episodes["2"], 1)
	assert.Equal(t, "Pilot", episodes["1"][0].Title)
}

func TestGetEpisodes_NoEpisodesVar(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<html>no episodes var</html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	episodes, err := client.GetEpisodes(context.Background(), "999")

	require.NoError(t, err)
	assert.Nil(t, episodes)
}

// =============================================================================
// Unit Tests: decorateRequest sets correct headers
// =============================================================================

func TestDecorateRequest_SetsHeaders(t *testing.T) {
	t.Parallel()

	var capturedHeaders http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header
		fmt.Fprint(w, `<html><body></body></html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	// Override to a real HTTP client that works with test server
	client.client = &http.Client{Timeout: 5 * time.Second}

	_, _ = client.SearchMedia("header_test")

	assert.Contains(t, capturedHeaders.Get("User-Agent"), "Mozilla")
	assert.Contains(t, capturedHeaders.Get("Accept"), "text/html")
	assert.Contains(t, capturedHeaders.Get("Accept-Language"), "pt-BR")
}

// =============================================================================
// Concurrency Tests
// =============================================================================

func TestSearchMedia_ConcurrentSafe(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(10 * time.Millisecond)
		fmt.Fprint(w, `<html><body>
			<div class="group/card">
				<h3>Concurrent Result</h3>
				<button data-msg="Copiar TMDB" data-copy="777">TMDB</button>
				<button data-msg="Copiar Link" data-copy="http://x.com/serie/777">Link</button>
				<div class="mt-3">2024 | SÉRIE</div>
			</div>
		</body></html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)

	var wg sync.WaitGroup
	errCh := make(chan error, 20)

	for i := range 20 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results, err := client.SearchMedia(fmt.Sprintf("query%d", idx))
			if err != nil {
				errCh <- fmt.Errorf("search %d failed: %w", idx, err)
				return
			}
			if len(results) != 1 {
				errCh <- fmt.Errorf("search %d: expected 1 result, got %d", idx, len(results))
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}

// =============================================================================
// Unit Tests: SuperFlixAdapter (UnifiedScraper interface)
// =============================================================================

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

	sfClient := newTestSuperFlixClient(srv.URL)
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

	sfClient := newTestSuperFlixClient(srv.URL)
	adapter := &SuperFlixAdapter{client: sfClient}

	_, err := adapter.SearchAnime("test")
	require.Error(t, err)
}

func TestSuperFlixAdapter_GetType(t *testing.T) {
	t.Parallel()

	adapter := &SuperFlixAdapter{client: NewSuperFlixClient()}
	assert.Equal(t, SuperFlixType, adapter.GetType())
}

func TestSuperFlixAdapter_GetAnimeEpisodes_ReturnsError(t *testing.T) {
	t.Parallel()

	adapter := &SuperFlixAdapter{client: NewSuperFlixClient()}
	_, err := adapter.GetAnimeEpisodes("1405")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SuperFlix")
}

func TestSuperFlixAdapter_GetClient(t *testing.T) {
	t.Parallel()

	inner := NewSuperFlixClient()
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

	sfClient := newTestSuperFlixClient(srv.URL)
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

	sfClient := newTestSuperFlixClient(srv.URL)
	adapter := &SuperFlixAdapter{client: sfClient}

	streamURL, metadata, err := adapter.GetStreamURL("1405", "serie", "1", "3")
	require.NoError(t, err)

	assert.Equal(t, "https://cdn.example.com/dexter-s1e3.m3u8", streamURL)
	assert.Equal(t, "superflix", metadata["source"])
}

// =============================================================================
// Unit Tests: tagResults for SuperFlix
// =============================================================================

func TestTagResults_SuperFlix_Movie(t *testing.T) {
	t.Parallel()

	manager := &ScraperManager{scrapers: make(map[ScraperType]UnifiedScraper)}
	results := []*models.Anime{
		{Name: "Inception", MediaType: models.MediaTypeMovie},
	}

	manager.tagResults(results, SuperFlixType)

	assert.Equal(t, "[Movie] [PT-BR] Inception", results[0].Name)
	assert.Equal(t, "SuperFlix", results[0].Source)
}

func TestTagResults_SuperFlix_TV(t *testing.T) {
	t.Parallel()

	manager := &ScraperManager{scrapers: make(map[ScraperType]UnifiedScraper)}
	results := []*models.Anime{
		{Name: "Breaking Bad", MediaType: models.MediaTypeTV},
	}

	manager.tagResults(results, SuperFlixType)

	assert.Equal(t, "[TV] [PT-BR] Breaking Bad", results[0].Name)
	assert.Equal(t, "SuperFlix", results[0].Source)
}

func TestTagResults_SuperFlix_Anime(t *testing.T) {
	t.Parallel()

	manager := &ScraperManager{scrapers: make(map[ScraperType]UnifiedScraper)}
	results := []*models.Anime{
		{Name: "Naruto", MediaType: models.MediaTypeAnime},
	}

	manager.tagResults(results, SuperFlixType)

	assert.Equal(t, "[PT-BR] Naruto", results[0].Name)
	assert.Equal(t, "SuperFlix", results[0].Source)
}

func TestTagResults_SuperFlix_NoDoubleTag(t *testing.T) {
	t.Parallel()

	manager := &ScraperManager{scrapers: make(map[ScraperType]UnifiedScraper)}
	results := []*models.Anime{
		{Name: "[PT-BR] Already Tagged", MediaType: models.MediaTypeAnime},
	}

	manager.tagResults(results, SuperFlixType)

	// Should not add another tag
	assert.Equal(t, "[PT-BR] Already Tagged", results[0].Name)
}

// =============================================================================
// Unit Tests: SSRF protection
// =============================================================================

func TestIsDisallowedIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ip         string
		disallowed bool
	}{
		{"loopback IPv4", "127.0.0.1", true},
		{"loopback IPv6", "::1", true},
		{"private 10.x", "10.0.0.1", true},
		{"private 192.168.x", "192.168.1.1", true},
		{"private 172.16.x", "172.16.0.1", true},
		{"multicast", "224.0.0.1", true},
		{"unspecified", "0.0.0.0", true},
		{"public IP", "8.8.8.8", false},
		{"public IP 2", "93.184.216.34", false},
		{"invalid IP", "notanip", true},
		{"empty", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.disallowed, isDisallowedIP(tc.ip))
		})
	}
}

// =============================================================================
// Unit Tests: Error helpers
// =============================================================================

func TestCheckHTTPStatus_Blocked(t *testing.T) {
	t.Parallel()

	blockedCodes := []int{http.StatusForbidden, http.StatusTooManyRequests, http.StatusServiceUnavailable}
	for _, code := range blockedCodes {
		t.Run(fmt.Sprintf("status_%d", code), func(t *testing.T) {
			t.Parallel()
			resp := &http.Response{StatusCode: code}
			err := checkHTTPStatus(resp, "test")
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrSourceUnavailable)
		})
	}
}

func TestCheckHTTPStatus_Success(t *testing.T) {
	t.Parallel()

	resp := &http.Response{StatusCode: http.StatusOK}
	err := checkHTTPStatus(resp, "test")
	assert.NoError(t, err)
}

func TestCheckHTTPStatus_OtherError(t *testing.T) {
	t.Parallel()

	resp := &http.Response{StatusCode: http.StatusNotFound}
	err := checkHTTPStatus(resp, "test")
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrSourceUnavailable, "404 is not a source-unavailable error")
}

func TestCheckHTMLResponse_JSONContentType(t *testing.T) {
	t.Parallel()

	resp := &http.Response{Header: http.Header{"Content-Type": []string{"application/json"}}}
	err := checkHTMLResponse(resp, []byte(`{"ok":true}`), "test")
	assert.NoError(t, err)
}

func TestCheckHTMLResponse_HTMLContentType(t *testing.T) {
	t.Parallel()

	resp := &http.Response{Header: http.Header{"Content-Type": []string{"text/html; charset=utf-8"}}}
	err := checkHTMLResponse(resp, []byte(`<html>`), "test")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSourceUnavailable)
}

func TestCheckHTMLResponse_HTMLBody(t *testing.T) {
	t.Parallel()

	resp := &http.Response{Header: http.Header{"Content-Type": []string{"application/octet-stream"}}}
	err := checkHTMLResponse(resp, []byte(`  <html>`), "test")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSourceUnavailable)
}

func TestValidateStreamURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		url       string
		expectErr bool
	}{
		{"valid HTTPS", "https://cdn.example.com/stream.m3u8", false},
		{"valid HTTP", "http://cdn.example.com/stream.m3u8", false},
		{"FTP scheme", "ftp://example.com/file", true},
		{"no scheme", "cdn.example.com/stream", true},
		{"empty", "", true},
		{"relative path", "/video/stream.m3u8", true},
		{"javascript scheme", "javascript:alert(1)", true},
		{"data URI", "data:text/html,<h1>hi</h1>", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := validateStreamURL(tc.url, "test")
			if tc.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidStreamURL)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, result)
			}
		})
	}
}

// =============================================================================
// Unit Tests: Scraper manager integration with SuperFlix
// =============================================================================

func TestScraperManager_SuperFlixRegistered(t *testing.T) {
	t.Parallel()

	// Create a manager with a mock SuperFlix to avoid real HTTP calls
	manager := &ScraperManager{
		scrapers: make(map[ScraperType]UnifiedScraper),
	}
	mockSF := &MockScraper{
		scraperType: SuperFlixType,
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{
				{Name: "Test Movie", MediaType: models.MediaTypeMovie},
			}, nil
		},
	}
	manager.scrapers[SuperFlixType] = mockSF

	scraper, err := manager.GetScraper(SuperFlixType)
	require.NoError(t, err)
	assert.Equal(t, SuperFlixType, scraper.GetType())
}

func TestScraperManager_SuperFlixDisplayName(t *testing.T) {
	t.Parallel()

	manager := &ScraperManager{scrapers: make(map[ScraperType]UnifiedScraper)}
	assert.Equal(t, "SuperFlix", manager.getScraperDisplayName(SuperFlixType))
}

func TestScraperManager_SuperFlixLanguageTag(t *testing.T) {
	t.Parallel()

	manager := &ScraperManager{scrapers: make(map[ScraperType]UnifiedScraper)}
	assert.Equal(t, "[PT-BR]", manager.getLanguageTag(SuperFlixType))
}

func TestScraperManager_SearchSpecificSuperFlix(t *testing.T) {
	t.Parallel()

	mockSF := &MockScraper{
		scraperType: SuperFlixType,
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{
				{Name: "Inception", MediaType: models.MediaTypeMovie},
				{Name: "Dexter", MediaType: models.MediaTypeTV},
				{Name: "Naruto", MediaType: models.MediaTypeAnime},
			}, nil
		},
	}

	manager := &ScraperManager{
		scrapers: map[ScraperType]UnifiedScraper{
			SuperFlixType: mockSF,
		},
	}

	scraperType := SuperFlixType
	results, err := manager.SearchAnime("test", &scraperType)

	require.NoError(t, err)
	require.Len(t, results, 3)

	// Verify tagging
	assert.Equal(t, "[Movie] [PT-BR] Inception", results[0].Name)
	assert.Equal(t, "[TV] [PT-BR] Dexter", results[1].Name)
	assert.Equal(t, "[PT-BR] Naruto", results[2].Name)

	// All should have SuperFlix source
	for _, r := range results {
		assert.Equal(t, "SuperFlix", r.Source)
	}
}

func TestScraperManager_SearchWithSuperFlixAmongOthers(t *testing.T) {
	t.Parallel()

	allAnimeMock := &MockScraper{
		scraperType: AllAnimeType,
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{
				{Name: "Naruto EN", URL: "aa-naruto"},
			}, nil
		},
	}

	superFlixMock := &MockScraper{
		scraperType: SuperFlixType,
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{
				{Name: "Naruto PT", MediaType: models.MediaTypeAnime},
			}, nil
		},
	}

	manager := &ScraperManager{
		scrapers: map[ScraperType]UnifiedScraper{
			AllAnimeType:  allAnimeMock,
			SuperFlixType: superFlixMock,
		},
	}

	results, err := manager.SearchAnime("naruto", nil)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Verify sources are present
	sources := make(map[string]bool)
	for _, r := range results {
		sources[r.Source] = true
	}
	assert.True(t, sources["AllAnime"])
	assert.True(t, sources["SuperFlix"])
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

func TestRegexPatterns(t *testing.T) {
	t.Parallel()

	t.Run("CSRF token regex", func(t *testing.T) {
		t.Parallel()
		match := sfCSRFTokenRe.FindStringSubmatch(`var CSRF_TOKEN = "abc123def";`)
		require.Len(t, match, 2)
		assert.Equal(t, "abc123def", match[1])
	})

	t.Run("PAGE token regex", func(t *testing.T) {
		t.Parallel()
		match := sfPageTokenRe.FindStringSubmatch(`var PAGE_TOKEN = "tok_xyz";`)
		require.Len(t, match, 2)
		assert.Equal(t, "tok_xyz", match[1])
	})

	t.Run("content ID regex", func(t *testing.T) {
		t.Parallel()
		match := sfContentIDRe.FindStringSubmatch(`var INITIAL_CONTENT_ID = 12345;`)
		require.Len(t, match, 2)
		assert.Equal(t, "12345", match[1])
	})

	t.Run("content type regex", func(t *testing.T) {
		t.Parallel()
		match := sfContentTypeRe.FindStringSubmatch(`var CONTENT_TYPE = "serie";`)
		require.Len(t, match, 2)
		assert.Equal(t, "serie", match[1])
	})

	t.Run("title regex with Player prefix", func(t *testing.T) {
		t.Parallel()
		match := sfTitleRe.FindStringSubmatch(`<title>Player | Breaking Bad</title>`)
		require.Len(t, match, 2)
		assert.Equal(t, "Breaking Bad", match[1])
	})

	t.Run("title regex without Player prefix", func(t *testing.T) {
		t.Parallel()
		match := sfTitleRe.FindStringSubmatch(`<title>Dexter</title>`)
		require.Len(t, match, 2)
		assert.Equal(t, "Dexter", match[1])
	})

	t.Run("subtitle part regex", func(t *testing.T) {
		t.Parallel()
		match := sfSubPartRe.FindStringSubmatch(`[Portuguese]https://subs.example.com/pt.vtt`)
		require.Len(t, match, 3)
		assert.Equal(t, "Portuguese", match[1])
		assert.Equal(t, "https://subs.example.com/pt.vtt", match[2])
	})
}
