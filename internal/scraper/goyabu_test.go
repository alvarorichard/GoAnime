package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoyabuSearchAnimeAPIMapFormat(t *testing.T) {
	t.Parallel()

	// The Goyabu WP REST API returns results as a map keyed by post ID
	apiResponse := map[string]map[string]any{
		"41411": {
			"title": "Naruto Clássico Dublado",
			"url":   "/anime/naruto-classico-dublado-online-hd-3",
			"img":   "https://goyabu.io/wp-content/uploads/naruto.jpg",
			"audio": "ptBr",
			"year":  "2002",
		},
		"40740": {
			"title": "Naruto Shippuden Dublado",
			"url":   "/anime/naruto-shippuden-dublado-online-hd",
			"img":   "https://goyabu.io/wp-content/uploads/shippuden.jpg",
			"audio": "ptBr",
			"year":  "2007",
		},
	}
	apiJSON, _ := json.Marshal(apiResponse)

	mux := http.NewServeMux()
	// Nonce endpoint (homepage)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		_, _ = fmt.Fprint(w, `<script>var glosAP = {"nonce":"abc123def"};</script>`)
	})
	// Search API endpoint
	mux.HandleFunc("/wp-json/animeonline/search/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(apiJSON)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewGoyabuClient()
	client.baseURL = server.URL
	client.maxRetries = 0
	client.retryDelay = 0

	results, err := client.SearchAnime("naruto")
	require.NoError(t, err)
	require.Len(t, results, 2)

	// Verify both results have correct data
	titles := map[string]bool{}
	for _, r := range results {
		titles[r.Name] = true
		assert.NotEmpty(t, r.URL)
		assert.NotEmpty(t, r.ImageURL)
		assert.Contains(t, r.URL, server.URL)
	}
	assert.True(t, titles["Naruto Clássico Dublado"])
	assert.True(t, titles["Naruto Shippuden Dublado"])
}

func TestGoyabuSearchAnimeHTMLFallback(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Return page without nonce to force HTML fallback
		if r.URL.Path == "/" && r.URL.Query().Get("s") == "" {
			_, _ = fmt.Fprint(w, `<html><body>No nonce here</body></html>`)
			return
		}
		// HTML search fallback
		_, _ = fmt.Fprint(w, `<html><body>
			<article>
				<a href="/anime/naruto-classico">
					<h3>Naruto Clássico</h3>
					<img src="https://example.com/naruto.jpg" alt="Naruto">
				</a>
			</article>
		</body></html>`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewGoyabuClient()
	client.baseURL = server.URL
	client.maxRetries = 0
	client.retryDelay = 0

	results, err := client.SearchAnime("naruto")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Naruto Clássico", results[0].Name)
}

func TestGoyabuGetAnimeEpisodesFromJS(t *testing.T) {
	t.Parallel()

	// Simulate the real Goyabu page format with const allEpisodes
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><head><script>
			const allEpisodes = [{"id":41414,"episodio":"1","link":"\/41414","type":"episode","episode_name":"","audio":"ptBr","imagem":"","miniature":"","update":"2023-08-18T00:38:39+00:00","status":""},{"id":41415,"episodio":"2","link":"\/41415","type":"episode","episode_name":"","audio":"ptBr","imagem":"","miniature":"","update":"2023-08-18T00:38:39+00:00","status":""},{"id":41416,"episodio":"3","link":"\/41416","type":"episode","episode_name":"","audio":"ptBr","imagem":"","miniature":"","update":"2023-08-18T00:38:39+00:00","status":""}];
		</script></head><body></body></html>`)
	}))
	defer server.Close()

	client := NewGoyabuClient()
	client.baseURL = server.URL
	client.maxRetries = 0
	client.retryDelay = 0

	episodes, err := client.GetAnimeEpisodes(server.URL + "/anime/test")
	require.NoError(t, err)
	require.Len(t, episodes, 3)

	// Episodes should be sorted ascending
	assert.Equal(t, 1, episodes[0].Num)
	assert.Equal(t, 2, episodes[1].Num)
	assert.Equal(t, 3, episodes[2].Num)

	// URLs should use the post ID format
	assert.Contains(t, episodes[0].URL, "/?p=41414")
	assert.Contains(t, episodes[1].URL, "/?p=41415")
	assert.Contains(t, episodes[2].URL, "/?p=41416")
}

func TestGoyabuGetAnimeEpisodesJSUnquotedKeys(t *testing.T) {
	t.Parallel()

	// Test fallback: JS object notation with unquoted keys
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><script>
			var episodes = [{id:"69013",episodio:"1",thumb:""},{id:"69014",episodio:"2",thumb:""}];
		</script></html>`)
	}))
	defer server.Close()

	client := NewGoyabuClient()
	client.baseURL = server.URL
	client.maxRetries = 0
	client.retryDelay = 0

	episodes, err := client.GetAnimeEpisodes(server.URL + "/anime/test")
	require.NoError(t, err)
	require.Len(t, episodes, 2)
	assert.Equal(t, 1, episodes[0].Num)
	assert.Equal(t, 2, episodes[1].Num)
}

func TestGoyabuGetEpisodeStreamURLIframe(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body>
			<iframe src="https://www.blogger.com/video.g?token=TestToken123"></iframe>
		</body></html>`)
	}))
	defer server.Close()

	client := NewGoyabuClient()
	client.baseURL = server.URL
	client.maxRetries = 0
	client.retryDelay = 0

	streamURL, err := client.GetEpisodeStreamURL(server.URL + "/episode/1")
	require.NoError(t, err)
	assert.Contains(t, streamURL, "blogger.com/video.g?token=TestToken123")
}

func TestGoyabuGetEpisodeStreamURLDirect(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body>
			<video><source src="https://cdn.example.com/video.mp4"></source></video>
		</body></html>`)
	}))
	defer server.Close()

	client := NewGoyabuClient()
	client.baseURL = server.URL
	client.maxRetries = 0
	client.retryDelay = 0

	streamURL, err := client.GetEpisodeStreamURL(server.URL + "/episode/1")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/video.mp4", streamURL)
}

func TestGoyabuSearchAnimeNoResults(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" && r.URL.Query().Get("s") == "" {
			_, _ = fmt.Fprint(w, `<script>var glosAP = {"nonce":"abc123"};</script>`)
			return
		}
		// Empty HTML search page
		_, _ = fmt.Fprint(w, `<html><body></body></html>`)
	})
	mux.HandleFunc("/wp-json/animeonline/search/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Real Goyabu error response: mixed string values in the map
		_, _ = fmt.Fprint(w, `{"error":"no_posts","title":"Sem resultados"}`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewGoyabuClient()
	client.baseURL = server.URL
	client.maxRetries = 0
	client.retryDelay = 0

	results, err := client.SearchAnime("nonexistent")
	// Should handle the mixed-type error response gracefully and fall through to HTML fallback
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestGoyabuSearchHyphenNormalization(t *testing.T) {
	t.Parallel()

	var receivedKeyword string
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<script>var glosAP = {"nonce":"abc123"};</script>`)
	})
	mux.HandleFunc("/wp-json/animeonline/search/", func(w http.ResponseWriter, r *http.Request) {
		receivedKeyword = r.URL.Query().Get("keyword")
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"1":{"title":"Test Anime","url":"/anime/test","img":""}}`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewGoyabuClient()
	client.baseURL = server.URL
	client.maxRetries = 0
	client.retryDelay = 0

	results, err := client.SearchAnime("cavaleiros-do-zodiaco")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	// Verify hyphens were normalized to spaces before hitting the API
	assert.Equal(t, "cavaleiros do zodiaco", receivedKeyword)
}

func TestGoyabuGetEpisodeStreamURLPlayersData(t *testing.T) {
	t.Parallel()

	ajaxCalled := false
	mux := http.NewServeMux()
	mux.HandleFunc("/episode/1", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><head><script>
			var playersData = [{"name":"Blog","select":"blogger","idioma":"","url":"https://www.blogger.com/video.g?token=TestEmbed","blogger_token":"dGVzdHRva2VuMTIz"}];
		</script></head><body><div id="player"></div></body></html>`)
	})
	mux.HandleFunc("/wp-admin/admin-ajax.php", func(w http.ResponseWriter, r *http.Request) {
		ajaxCalled = true
		assert.Equal(t, "decode_blogger_video", r.FormValue("action"))
		assert.Equal(t, "dGVzdHRva2VuMTIz", r.FormValue("token"), "should send 'token' not 'blogger_token'")
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"success":true,"data":{"play":[{"src":"https://cdn.example.com/video-720.mp4","size":720,"type":"video/mp4"},{"src":"https://cdn.example.com/video-360.mp4","size":360,"type":"video/mp4"}]}}`)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewGoyabuClient()
	client.baseURL = server.URL
	client.maxRetries = 0
	client.retryDelay = 0

	streamURL, err := client.GetEpisodeStreamURL(server.URL + "/episode/1")
	require.NoError(t, err)
	assert.True(t, ajaxCalled, "AJAX endpoint should be called")
	assert.Equal(t, "https://cdn.example.com/video-720.mp4", streamURL, "should pick highest quality")
}

func TestGoyabuGetEpisodeStreamURLBloggerFallback(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/episode/1", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><head><script>
			var playersData = [{"name":"Blog","select":"blogger","url":"https://www.blogger.com/video.g?token=FallbackEmbed","blogger_token":"dGVzdA=="}];
		</script></head><body></body></html>`)
	})
	mux.HandleFunc("/wp-admin/admin-ajax.php", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"success":false,"data":{"message":"Nenhum vídeo encontrado."}}`)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewGoyabuClient()
	client.baseURL = server.URL
	client.maxRetries = 0
	client.retryDelay = 0

	streamURL, err := client.GetEpisodeStreamURL(server.URL + "/episode/1")
	require.NoError(t, err)
	assert.Equal(t, "https://www.blogger.com/video.g?token=FallbackEmbed", streamURL, "should fall back to Blogger embed URL")
}

func TestGoyabuParseEpisodesFromJS_NoEpisodes(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><p>No episodes here</p></body></html>`)
	}))
	defer server.Close()

	client := NewGoyabuClient()
	client.baseURL = server.URL
	client.maxRetries = 0
	client.retryDelay = 0

	episodes, err := client.GetAnimeEpisodes(server.URL + "/anime/test")
	require.NoError(t, err)
	assert.Empty(t, episodes)
}

func TestGoyabuSearchAnimeClassifiesBlockedHTMLAsSourceUnavailable(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/" && r.URL.Query().Get("s") == "":
			w.WriteHeader(http.StatusForbidden)
		case r.URL.Query().Get("s") != "":
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, `<html><head><title>Just a moment...</title></head><body><div id="cf-wrapper"></div></body></html>`)
		default:
			http.NotFound(w, r)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewGoyabuClient()
	client.baseURL = server.URL
	client.maxRetries = 0
	client.retryDelay = 0

	_, err := client.SearchAnime("naruto")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable), "expected ErrSourceUnavailable, got: %v", err)
}

func TestGoyabuGetEpisodeStreamURLBlockedPage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `<html><head><title>Just a moment...</title></head><body><div id="cf-wrapper"></div></body></html>`)
	}))
	defer server.Close()

	client := NewGoyabuClient()
	client.baseURL = server.URL
	client.maxRetries = 0
	client.retryDelay = 0

	_, err := client.GetEpisodeStreamURL(server.URL + "/episode/blocked")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable), "expected ErrSourceUnavailable, got: %v", err)
}

func TestGoyabuDecodeBloggerTokenClassifiesHTMLAsSourceUnavailable(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/wp-admin/admin-ajax.php", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `<html><body>blocked</body></html>`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewGoyabuClient()
	client.baseURL = server.URL
	client.maxRetries = 0
	client.retryDelay = 0

	_, err := client.decodeBloggerToken("token123")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable), "expected ErrSourceUnavailable, got: %v", err)
}
