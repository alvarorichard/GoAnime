package superflix

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
	c.browserSolver = nil                                                               // use the plain-HTTP episode path against httptest
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
		{
			name:         "filters missing, null, and future air_dates",
			html:         fmt.Sprintf(`var ALL_EPISODES = {"1":[{"epi_num":"1","title":"Valid","air_date":"2020-01-15"},{"epi_num":"2","title":"Missing","air_date":""},{"epi_num":"3","title":"Null","air_date":"null"},{"epi_num":"4","title":"Future","air_date":"%s"}]};`, time.Now().Add(48*time.Hour).Format("2006-01-02")),
			expectKeys:   []string{"1"},
			expectCounts: map[string]int{"1": 1},
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

// TestSearchMedia_NormalizesHyphenatedQuery verifies that hyphenated CLI
// queries (e.g. "the-boys" produced by TreatingAnimeName) are converted to
// spaced queries before being sent to SuperFlix, which treats the dash as a
// literal character and would otherwise return zero results.
func TestSearchMedia_NormalizesHyphenatedQuery(t *testing.T) {
	t.Parallel()

	var receivedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.Query().Get("s")
		fmt.Fprint(w, `<html><body>
			<div class="group/card">
				<img alt="The Boys" src="https://image.tmdb.org/t/p/w342/p.jpg" />
				<button data-msg="Copiar TMDB" data-copy="76479">TMDB</button>
				<button data-msg="Copiar Link" data-copy="http://example.com/serie/76479">Link</button>
				<div class="mt-3">2019 | SÉRIE</div>
			</div>
		</body></html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	results, err := client.SearchMedia("the-boys")

	require.NoError(t, err)
	assert.Equal(t, "the boys", receivedQuery, "hyphens should be normalized to spaces")
	require.Len(t, results, 1)
	assert.Equal(t, "The Boys", results[0].Title)
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

	// Non-JSON, non-HTML body still surfaces a JSON decode error.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{not json`)
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

// =============================================================================
// Regression tests (added 2026-04-30)
//
// Context: SuperFlix migrates hosts via server-side 301 redirects
// (`superflixapi.rest` → `.online` → `.best`).
// Go's http.Client follows the redirect but
// downgrades the POST to a GET (dropping the body), so /player/bootstrap
// returned an HTML 404 page. The JSON decoder then surfaced the cryptic
// `invalid character '<' looking for beginning of value`, breaking playback.
// These tests pin (a) the canonical base URL and (b) that an HTML/non-2xx
// response from the player API produces a clear, actionable error rather
// than the cryptic JSON decode error.
// =============================================================================

func TestSuperFlixBase_PointsToLiveHost_2026_06_05(t *testing.T) {
	t.Parallel()
	// Pinning the canonical host. If this needs to change in the future,
	// also update internal/api/providers/metadata/metadata.go.
	// 2026-05-18: .online started 301-redirecting to .best, breaking POSTs.
	// 2026-06-05: .best now 301-redirects to .fit (same POST→GET downgrade),
	// and .fit gates the player page behind a Cloudflare Turnstile served on
	// a plain HTTP 200 — handled by cfFallbackTransport.shouldInspect.
	// 2026-06-18: .fit went dead (NXDOMAIN) and rotated to .cyou.
	// 2026-07-04: .cyou→.lifestyle→.pro; .lifestyle 301-redirects to .pro, so
	// we pin the real canonical host .pro (confirmed via the embed cfv token).
	assert.Equal(t, "https://superflixapi.pro", SuperFlixBase)
}

func TestBootstrap_HTMLResponseSurfacesActionableError_2026_04_30(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `<!DOCTYPE html><html><head><title>Not Found</title></head><body>404</body></html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	tokens := &SuperFlixTokens{CSRF: "a", PageToken: "b", ContentID: "1", ContentType: "filme"}
	_, err := client.Bootstrap(context.Background(), tokens)

	require.Error(t, err)
	// Must NOT leak the cryptic JSON decode error.
	assert.NotContains(t, err.Error(), "invalid character '<'")
	// Must surface the real cause: HTML body with status code in context.
	assert.Contains(t, err.Error(), "bootstrap")
	assert.Contains(t, err.Error(), "HTML")
	assert.Contains(t, err.Error(), "404")
}

func TestGetSourceURL_HTMLResponseSurfacesActionableError_2026_04_30(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `<html><body>blocked</body></html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	tokens := &SuperFlixTokens{CSRF: "a", PageToken: "b"}
	_, err := client.GetSourceURL(context.Background(), "vid", tokens)

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "invalid character '<'")
	assert.Contains(t, err.Error(), "source")
	assert.Contains(t, err.Error(), "HTML")
	assert.Contains(t, err.Error(), "403")
}

func TestGetVideoAPI_HTMLResponseSurfacesActionableError_2026_04_30(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<html><body>captcha</body></html>`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	_, _, err := client.GetVideoAPI(context.Background(), srv.URL, "hash", srv.URL+"/")

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "invalid character '<'")
	assert.Contains(t, err.Error(), "video API")
	assert.Contains(t, err.Error(), "HTML")
}

// Some upstream players (firevideoplayer.com behind llanfairpwllgwyngy.com)
// serve real JSON with `Content-Type: text/html; charset=utf-8`. Trusting the
// header alone would reject these valid responses. The body sniff is the
// source of truth.
func TestGetVideoAPI_AcceptsJSONBodyWithHTMLContentType_2026_04_30(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `{"hls":true,"securedLink":"https://example.com/master.m3u8","videoSource":"https://example.com/master.txt","videoImage":"https://example.com/thumb.jpg"}`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	streamURL, thumb, err := client.GetVideoAPI(context.Background(), srv.URL, "hash", srv.URL+"/")

	require.NoError(t, err)
	assert.Equal(t, "https://example.com/master.m3u8", streamURL)
	assert.Equal(t, "https://example.com/thumb.jpg", thumb)
}

func TestBootstrap_AcceptsJSONBodyWithHTMLContentType_2026_04_30(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `{"data":{"options":[{"ID":"sv1","name":"Server 1"}]}}`)
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	tokens := &SuperFlixTokens{CSRF: "a", PageToken: "b", ContentID: "1", ContentType: "filme"}
	servers, err := client.Bootstrap(context.Background(), tokens)

	require.NoError(t, err)
	require.Len(t, servers, 1)
	assert.Equal(t, "Server 1", servers[0].Name)
}

func TestEnsureJSONResponse_BlankBodyWithBadStatus_2026_04_30(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		// Empty body — JSON decode would also fail with EOF; ensure the
		// status-code path produces a useful error first.
	}))
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	tokens := &SuperFlixTokens{CSRF: "a", PageToken: "b", ContentID: "1", ContentType: "filme"}
	_, err := client.Bootstrap(context.Background(), tokens)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

// Regression test (added 2026-05-02)
//
// Episode air_date filter must not depend on the time-of-day in `now`. A
// previous version used `t.After(now.Add(24*time.Hour))`, which let
// tomorrow's episodes leak through whenever `now`'s UTC time-of-day was
// past 00:00 — i.e. for any caller in a timezone west of UTC, every
// evening reproduced a 1-day-of-leak window. Pin two boundary cases on
// a fixed clock so neither timezone nor wall-clock drift can hide a
// regression:
//  1. now=2026-05-02 03:30 UTC, ep.air_date=2026-05-03 → must be filtered
//  2. now=2026-05-02 03:30 UTC, ep.air_date=2026-05-02 → must be kept
func TestFilterEpisodesByAirDate_TomorrowIsNotKept_2026_05_02(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 2, 3, 30, 0, 0, time.UTC)
	episodes := map[string][]SuperFlixEpisode{
		"1": {
			{EpiNum: json.Number("1"), Title: "Today", AirDate: "2026-05-02"},
			{EpiNum: json.Number("2"), Title: "Tomorrow", AirDate: "2026-05-03"},
			{EpiNum: json.Number("3"), Title: "WayPast", AirDate: "2020-01-15"},
			{EpiNum: json.Number("4"), Title: "Empty", AirDate: ""},
			{EpiNum: json.Number("5"), Title: "Null", AirDate: "null"},
		},
	}

	got := filterEpisodesByAirDate(episodes, now)
	require.Len(t, got["1"], 2, "should keep only Today and WayPast")

	titles := []string{got["1"][0].Title, got["1"][1].Title}
	assert.Contains(t, titles, "Today")
	assert.Contains(t, titles, "WayPast")
	assert.NotContains(t, titles, "Tomorrow", "future air_date must be filtered regardless of UTC time-of-day")
}

// Regression test (added 2026-05-02)
//
// Same boundary as above, but exercised through a non-UTC `now` to prove
// the filter normalizes to UTC internally. BRT (UTC-3) at 23:30 local on
// 2026-05-01 == 02:30 UTC on 2026-05-02, so episodes with air_date
// 2026-05-03 are still tomorrow-in-UTC and must be filtered.
func TestFilterEpisodesByAirDate_NonUTCNow_2026_05_02(t *testing.T) {
	t.Parallel()

	brt := time.FixedZone("BRT", -3*60*60)
	now := time.Date(2026, 5, 1, 23, 30, 0, 0, brt) // == 2026-05-02 02:30 UTC
	episodes := map[string][]SuperFlixEpisode{
		"1": {
			{EpiNum: json.Number("1"), Title: "TodayUTC", AirDate: "2026-05-02"},
			{EpiNum: json.Number("2"), Title: "TomorrowUTC", AirDate: "2026-05-03"},
		},
	}

	got := filterEpisodesByAirDate(episodes, now)
	require.Len(t, got["1"], 1)
	assert.Equal(t, "TodayUTC", got["1"][0].Title)
}

// Regression test (added 2026-05-01)
//
// Naruto S2E5 on SuperFlix is a placeholder episode (`air_date: null`,
// title "Episódio 5"); /player/bootstrap returns `{"data":{"options":[]}}`
// for it. Before this fix the user saw a generic "no servers available"
// error that looked like a system bug. Pin three things:
//  1. ErrSuperFlixNoServers wraps the empty-options condition
//  2. The error message includes the player URL and contentid for triage
//  3. Real responses with servers still succeed (no false positives)
func TestGetStreamURL_EmptyBootstrapOptionsReturnsTypedError_2026_05_01(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/serie/46260/2/5", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<html><head><title>Player</title></head><body>
			<script>
			var CSRF_TOKEN = "csrf123";
			var PAGE_TOKEN = "pt456";
			var INITIAL_CONTENT_ID = 999914;
			var CONTENT_TYPE = "serie";
			</script>
		</body></html>`)
	})
	mux.HandleFunc("/player/bootstrap", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"options":[],"flags":{"mp4_active":false}}}`)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestSuperFlixClient(srv.URL)
	_, err := client.GetStreamURL(context.Background(), "serie", "46260", "2", "5")

	require.Error(t, err)
	require.ErrorIs(t, err, ErrSuperFlixNoServers,
		"empty bootstrap options must produce ErrSuperFlixNoServers so callers can distinguish content-unavailability from system errors")
	msg := err.Error()
	assert.Contains(t, msg, "/serie/46260/2/5", "error must include the player path for triage")
	assert.Contains(t, msg, "contentid=999914", "error must include the contentid that returned no servers")
}
