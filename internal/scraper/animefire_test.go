package scraper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnimefireSearchRetriesOnFailure(t *testing.T) {
	t.Parallel()

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}

		_, _ = fmt.Fprint(w, `
        <html>
            <body>
                <div class="row ml-1 mr-1">
                    <a href="/anime/1">Naruto</a>
                </div>
            </body>
        </html>
        `)
	}))
	defer server.Close()

	client := NewAnimefireClient()
	client.baseURL = server.URL
	client.maxRetries = 2
	client.retryDelay = 0

	results, err := client.SearchAnime("naruto")
	require.NoError(t, err)
	require.Len(t, results, 1)

	assert.Equal(t, "Naruto", results[0].Name)
	assert.Equal(t, server.URL+"/anime/1", results[0].URL)
}

func TestAnimefireSearchReturnsEmptySliceWhenNoMatch(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `<html><body><div class="nothing-here"></div></body></html>`)
	}))
	defer server.Close()

	client := NewAnimefireClient()
	client.baseURL = server.URL
	client.maxRetries = 1
	client.retryDelay = 0

	results, err := client.SearchAnime("unknown")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestAnimefireSearchDetectsChallengePage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `
        <html>
            <head><title>Just a moment...</title></head>
            <body>
                <div id="cf-wrapper">Blocked</div>
            </body>
        </html>
        `)
	}))
	defer server.Close()

	client := NewAnimefireClient()
	client.baseURL = server.URL
	client.maxRetries = 1
	client.retryDelay = 0

	_, err := client.SearchAnime("naruto")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "challenge")
}
