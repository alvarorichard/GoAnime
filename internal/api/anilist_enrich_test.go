package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withAniList swaps aniListEndpoint for tests and restores it on cleanup.
// Must NOT be used with t.Parallel() — shares a package-level var.
func withAniList(t *testing.T, u string) {
	t.Helper()
	prev := aniListEndpoint
	aniListEndpoint = u
	t.Cleanup(func() { aniListEndpoint = prev })
}

func aniListSuccessBody(id int, malID int, romaji string) []byte {
	resp := map[string]any{
		"data": map[string]any{
			"Media": map[string]any{
				"id":    id,
				"idMal": malID,
				"title": map[string]any{
					"romaji":  romaji,
					"english": romaji + " EN",
					"native":  romaji + " JP",
				},
				"coverImage": map[string]any{
					"large": "https://cdn.anilist.co/cover.jpg",
				},
				"synonyms": []string{},
			},
		},
	}
	data, _ := json.Marshal(resp)
	return data
}

func TestFetchAnimeFromAniListWithURL_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(aniListSuccessBody(123, 456, "Shingeki no Kyojin"))
	}))
	t.Cleanup(srv.Close)
	withAniList(t, srv.URL)

	result, err := FetchAnimeFromAniListWithURL("Attack on Titan", "https://goyabu.io/anime/shingeki-no-kyojin")
	require.NoError(t, err)
	assert.Equal(t, 123, result.Data.Media.ID)
	assert.Equal(t, 456, result.Data.Media.IDMal)
	assert.Equal(t, "Shingeki no Kyojin", result.Data.Media.Title.Romaji)
}

func TestFetchAnimeFromAniListWithURL_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	t.Cleanup(srv.Close)
	withAniList(t, srv.URL)

	_, err := FetchAnimeFromAniListWithURL("SomeAnimeP18Error", "")
	require.Error(t, err)
}

func TestFetchAnimeFromAniListWithURL_ZeroIDErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"data":{"Media":{"id":0}}}`)
	}))
	t.Cleanup(srv.Close)
	withAniList(t, srv.URL)

	_, err := FetchAnimeFromAniListWithURL("SomeAnimeP18ZeroID", "")
	require.Error(t, err)
}

func TestFetchAnimeFromAniListWithURL_CacheHit(t *testing.T) {
	var callCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(aniListSuccessBody(789, 101, "CacheTestAnime"))
	}))
	t.Cleanup(srv.Close)
	withAniList(t, srv.URL)

	// First call hits the server and populates cache
	r1, err := FetchAnimeFromAniListWithURL("CacheTestAnime P18 Unique", "")
	require.NoError(t, err)
	assert.Equal(t, 789, r1.Data.Media.ID)

	// Second call should use cache (server not called again)
	r2, err := FetchAnimeFromAniListWithURL("CacheTestAnime P18 Unique", "")
	require.NoError(t, err)
	assert.Equal(t, 789, r2.Data.Media.ID)
	assert.Equal(t, 1, callCount, "cache should prevent second server call")
}

func TestFetchAnimeFromAniListWithURL_BadJSONErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{not valid json`)
	}))
	t.Cleanup(srv.Close)
	withAniList(t, srv.URL)

	_, err := FetchAnimeFromAniListWithURL("SomeAnimeP18BadJSON", "")
	require.Error(t, err)
}

func TestEnrichAnimeData_AniListSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(aniListSuccessBody(999, 111, "Naruto"))
	}))
	t.Cleanup(srv.Close)
	withAniList(t, srv.URL)

	anime := &models.Anime{
		Name: "Naruto P18 Unique Test",
		URL:  "https://goyabu.io/anime/naruto-shippuuden",
	}
	err := enrichAnimeData(anime)
	require.NoError(t, err)
	assert.Equal(t, 999, anime.AnilistID)
	assert.Equal(t, 111, anime.MalID)
	assert.Equal(t, "https://cdn.anilist.co/cover.jpg", anime.ImageURL)
}

func TestEnrichAnimeData_AniListFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	t.Cleanup(srv.Close)
	withAniList(t, srv.URL)

	anime := &models.Anime{
		Name: "SomeAnimeP18FailTest",
		URL:  "",
	}
	err := enrichAnimeData(anime)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AniList enrichment failed")
}

func TestAniListPost_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		_, _ = fmt.Fprint(w, `{"ok":true}`)
	}))
	t.Cleanup(srv.Close)

	resp, body, err := aniListPost(srv.URL, []byte(`{"test":1}`))
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, string(body), `"ok":true`)
}

func TestAniListPost_InvalidURLErrors(t *testing.T) {
	t.Parallel()
	_, _, err := aniListPost("\x00invalid", nil)
	require.Error(t, err)
}

func TestMakeGetRequest_NotFoundStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	t.Cleanup(srv.Close)
	prev := httpClient
	httpClient = &http.Client{}
	t.Cleanup(func() { httpClient = prev })

	_, err := makeGetRequest(srv.URL, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status")
}

func TestMakeGetRequest_WithCustomHeaders(t *testing.T) {
	var gotReferer string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReferer = r.Header.Get("Referer")
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	}))
	t.Cleanup(srv.Close)
	prev := httpClient
	httpClient = &http.Client{}
	t.Cleanup(func() { httpClient = prev })

	_, err := makeGetRequest(srv.URL, map[string]string{"Referer": "https://test.example.com"})
	require.NoError(t, err)
	assert.Equal(t, "https://test.example.com", gotReferer)
}

func TestValidateExternalURL_HostnameResolvesToLoopback(t *testing.T) {
	t.Parallel()
	// "localhost" resolves to 127.0.0.1 which is disallowed
	err := ValidateExternalURL("http://localhost/path")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "disallowed")
}

func TestValidateExternalURL_NXDomainErrors(t *testing.T) {
	t.Parallel()
	// .invalid TLD is reserved by RFC 2606 and never resolves
	err := ValidateExternalURL("http://no-such-host.invalid.example/path")
	require.Error(t, err)
}
