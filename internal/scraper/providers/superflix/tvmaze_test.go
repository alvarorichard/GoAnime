package superflix

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tvmazeStub serves the two TVmaze endpoints the lister hits: /lookup/shows and
// /shows/{id}/episodes. Past airdate so the air-date filter keeps the episode.
func tvmazeStub(t *testing.T, lookupBody, episodesBody string, lookupCode, epsCode int) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/lookup/shows":
			w.WriteHeader(lookupCode)
			_, _ = w.Write([]byte(lookupBody))
		case "/shows/15299/episodes":
			w.WriteHeader(epsCode)
			_, _ = w.Write([]byte(episodesBody))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

// TestGetEpisodesFromTVmaze covers the keyless episode lister: a happy path plus
// the failure modes that must fall back to the browser.
func TestGetEpisodesFromTVmaze(t *testing.T) {
	const okEps = `[
		{"season":1,"number":1,"name":"The Name of the Game","airdate":"2019-07-26"},
		{"season":1,"number":2,"name":"Cherry","airdate":"2019-07-26"},
		{"season":2,"number":1,"name":"The Big Ride","airdate":"2020-09-04"},
		{"season":0,"number":1,"name":"Special","airdate":"2019-01-01"},
		{"season":9,"number":1,"name":"Unaired","airdate":"2099-01-01"}
	]`

	tests := []struct {
		name       string
		imdb       string
		lookup     string
		eps        string
		lookupCode int
		epsCode    int
		wantErr    bool
		assert     func(t *testing.T, got map[string][]SuperFlixEpisode)
	}{
		{
			name: "happy path groups seasons and filters specials/future",
			imdb: "tt1190634", lookup: `{"id":15299,"name":"The Boys"}`, eps: okEps,
			lookupCode: 200, epsCode: 200,
			assert: func(t *testing.T, got map[string][]SuperFlixEpisode) {
				require.Len(t, got, 2, "season 0 (special) and season 9 (future) dropped")
				require.Len(t, got["1"], 2)
				assert.Equal(t, "1", got["1"][0].EpiNum.String())
				assert.Equal(t, "The Name of the Game", got["1"][0].Title)
				assert.Equal(t, "2019-07-26", got["1"][0].AirDate)
				require.Len(t, got["2"], 1)
				assert.Equal(t, "The Big Ride", got["2"][0].Title)
			},
		},
		{name: "empty imdb errors", imdb: "", wantErr: true},
		{
			name: "lookup 404 errors", imdb: "tt0000000",
			lookup: ``, eps: ``, lookupCode: 404, epsCode: 200, wantErr: true,
		},
		{
			name: "lookup with no id errors", imdb: "tt1190634",
			lookup: `{}`, eps: ``, lookupCode: 200, epsCode: 200, wantErr: true,
		},
		{
			name: "no usable episodes errors", imdb: "tt1190634",
			lookup: `{"id":15299}`, eps: `[{"season":0,"number":1,"name":"x","airdate":"2019-01-01"}]`,
			lookupCode: 200, epsCode: 200, wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Not parallel: each subtest swaps the package-global tvmazeBaseURL.
			if tt.imdb != "" {
				base := tvmazeStub(t, tt.lookup, tt.eps, tt.lookupCode, tt.epsCode)
				prev := tvmazeBaseURL
				tvmazeBaseURL = base
				t.Cleanup(func() { tvmazeBaseURL = prev })
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			got, err := GetEpisodesFromTVmaze(ctx, http.DefaultClient, tt.imdb)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			tt.assert(t, got)
		})
	}
}

// TestGetEpisodesFromTVmaze_NilClient verifies the function provides its own
// client when passed nil (it must not panic).
func TestGetEpisodesFromTVmaze_NilClient(t *testing.T) {
	base := tvmazeStub(t, `{"id":15299}`,
		`[{"season":1,"number":1,"name":"Ep","airdate":"2019-07-26"}]`, 200, 200)
	prev := tvmazeBaseURL
	tvmazeBaseURL = base
	t.Cleanup(func() { tvmazeBaseURL = prev })

	got, err := GetEpisodesFromTVmaze(context.Background(), nil, "tt1190634")
	require.NoError(t, err)
	assert.Len(t, got["1"], 1)
}

func TestGetEpisodesFromTVmaze_ServerError(t *testing.T) {
	base := tvmazeStub(t, `{"id":15299}`, ``, 200, 500)
	prev := tvmazeBaseURL
	tvmazeBaseURL = base
	t.Cleanup(func() { tvmazeBaseURL = prev })

	_, err := GetEpisodesFromTVmaze(context.Background(), http.DefaultClient, "tt1190634")
	require.Error(t, err)
	assert.Contains(t, fmt.Sprint(err), "500")
}
