package movie

import "net/http"

// NewTMDBClientForTest builds a TMDB client over a caller-supplied client and
// base URL — the fake TMDB servers the provider packages unit-test against.
// It exists so the movie client's own parsing runs over recorded fixtures
// without network.
func NewTMDBClientForTest(hc *http.Client, baseURL, apiKey string) *TMDBClient {
	return &TMDBClient{
		client:    hc,
		apiKey:    apiKey,
		baseURL:   baseURL,
		imageBase: TMDBImageBaseURL,
	}
}
