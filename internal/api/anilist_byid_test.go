package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Not parallel: these swap aniListEndpoint and touch the AniList latch and
// cache, all package-level. Each test uses ids of its own so the cache does
// not carry an answer from one into another.

// aniListRequest is what a test server reads out of one AniList POST.
type aniListRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

func (r aniListRequest) byID() bool { return strings.Contains(r.Query, "Media(id: $id") }

func (r aniListRequest) id() int {
	n, _ := r.Variables["id"].(float64)
	return int(n)
}

// aniListServer answers by-id queries with byID and title searches with
// search, counting both.
type aniListServer struct {
	srv           *httptest.Server
	idCalls       atomic.Int32
	searchCalls   atomic.Int32
	lastRequested atomic.Int64
	badShape      atomic.Bool
}

func newAniListServer(t *testing.T, byID, search func(req aniListRequest) (int, []byte)) *aniListServer {
	t.Helper()
	s := &aniListServer{}
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var req aniListRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			s.badShape.Store(true)
		}
		status, body := http.StatusOK, []byte(`{}`)
		if req.byID() {
			s.idCalls.Add(1)
			s.lastRequested.Store(int64(req.id()))
			if !strings.Contains(req.Query, "idMal") || !strings.Contains(req.Query, "coverImage") {
				s.badShape.Store(true)
			}
			status, body = byID(req)
		} else {
			s.searchCalls.Add(1)
			if search != nil {
				status, body = search(req)
			}
		}
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(s.srv.Close)
	withAniList(t, s.srv.URL)
	return s
}

// byIDScenario is one AniList answer and what FetchAnimeFromAniListByID must
// make of it. Shared by the unit test and the metatest.
type byIDScenario struct {
	name      string
	id        int
	answer    func(id int) (int, []byte)
	wantID    int   // 0: an error is expected
	wantErr   error // a sentinel the error must carry, if any
	wantCalls int32
}

func byIDScenarios(base int) []byIDScenario {
	found := func(id int) (int, []byte) { return http.StatusOK, aniListSuccessBody(id, id+1000, "Found") }
	return []byIDScenario{
		{name: "found", id: base + 1, answer: found, wantID: base + 1, wantCalls: 1},
		{name: "another entry answered", id: base + 2, answer: func(id int) (int, []byte) {
			return http.StatusOK, aniListSuccessBody(id+1, 0, "Other")
		}, wantCalls: 1},
		{name: "no such entry", id: base + 3, answer: func(int) (int, []byte) {
			return http.StatusOK, []byte(`{"data":{"Media":null}}`)
		}, wantCalls: 1},
		{name: "server error", id: base + 4, answer: func(int) (int, []byte) {
			return http.StatusInternalServerError, []byte(`{"errors":[{"message":"boom"}]}`)
		}, wantCalls: 1},
		{name: "switched off upstream", id: base + 5, answer: func(int) (int, []byte) {
			return http.StatusForbidden, []byte(aniListDisabledBody)
		}, wantErr: ErrAniListAPIDisabled, wantCalls: 1},
		{name: "not JSON", id: base + 6, answer: func(int) (int, []byte) {
			return http.StatusOK, []byte(`<html>`)
		}, wantCalls: 1},
		{name: "no id", id: 0, answer: found, wantCalls: 0},
	}
}

// byIDFailures runs fetch against every scenario and lists what is wrong.
func byIDFailures(t *testing.T, base int, fetch func(int) (*models.AniListResponse, error)) (f []string) {
	t.Helper()
	for _, sc := range byIDScenarios(base) {
		resetAniListDisabledForTest()
		s := newAniListServer(t, func(req aniListRequest) (int, []byte) { return sc.answer(req.id()) }, nil)
		got, err := fetch(sc.id)
		switch {
		case sc.wantID != 0 && (err != nil || got == nil || got.Data.Media.ID != sc.wantID):
			f = append(f, fmt.Sprintf("%s: got %v, %v", sc.name, got, err))
		case sc.wantID == 0 && err == nil:
			f = append(f, sc.name+": no error")
		case sc.wantErr != nil && !errors.Is(err, sc.wantErr):
			f = append(f, fmt.Sprintf("%s: error %v does not name %v", sc.name, err, sc.wantErr))
		}
		if calls := s.idCalls.Load(); calls != sc.wantCalls {
			f = append(f, fmt.Sprintf("%s: %d requests, want %d", sc.name, calls, sc.wantCalls))
		}
		if sc.wantCalls > 0 && s.lastRequested.Load() != int64(sc.id) {
			f = append(f, fmt.Sprintf("%s: asked for id %d", sc.name, s.lastRequested.Load()))
		}
		if s.searchCalls.Load() != 0 || s.badShape.Load() {
			f = append(f, sc.name+": searched by title or sent a malformed query")
		}
	}
	resetAniListDisabledForTest()
	return f
}

func TestFetchAnimeFromAniListByID(t *testing.T) {
	assert.Empty(t, byIDFailures(t, 910000, FetchAnimeFromAniListByID))
}

func TestFetchAnimeFromAniListByIDTampersAreCaught(t *testing.T) {
	orig := FetchAnimeFromAniListByID
	tampers := map[string]func(int) (*models.AniListResponse, error){
		"errors swallowed": func(id int) (*models.AniListResponse, error) {
			got, err := orig(id)
			if err != nil {
				return &models.AniListResponse{}, nil //nolint:nilerr // deliberate: this tamper swallows the error so the contract must notice
			}
			return got, nil
		},
		"the id is ignored": func(int) (*models.AniListResponse, error) { return orig(1) },
		"off by one":        func(id int) (*models.AniListResponse, error) { return orig(id + 1) },
		"the outage is not named": func(id int) (*models.AniListResponse, error) {
			got, err := orig(id)
			if errors.Is(err, ErrAniListAPIDisabled) {
				return nil, errors.New("AniList returned: 403 Forbidden")
			}
			return got, err
		},
		"asked twice": func(id int) (*models.AniListResponse, error) {
			_, _ = orig(id)
			return orig(id)
		},
	}
	for name, tamper := range tampers {
		resetAniListDisabledForTest()
		// A fresh id range per tamper: a cached answer would hide the request.
		base := 920000 + 100*len(name)
		assert.NotEmpty(t, byIDFailures(t, base, tamper), "tamper %q survived", name)
	}
}

func TestFetchAnimeFromAniListByIDUsesItsCacheAndTheLatch(t *testing.T) {
	resetAniListDisabledForTest()
	t.Cleanup(resetAniListDisabledForTest)
	s := newAniListServer(t, func(req aniListRequest) (int, []byte) {
		return http.StatusOK, aniListSuccessBody(req.id(), 7, "Cached")
	}, nil)

	for range 2 {
		got, err := FetchAnimeFromAniListByID(930001)
		require.NoError(t, err)
		assert.Equal(t, 930001, got.Data.Media.ID)
	}
	assert.Equal(t, int32(1), s.idCalls.Load(), "the second lookup is answered from the cache")

	// A cache entry that names another entry is not an answer for this id.
	poisoned := 930003
	util.GetAniListCache().Set("anilist:id:930003", aniListSuccessBody(1, 1, "Somebody Else"))
	got, err := FetchAnimeFromAniListByID(poisoned)
	require.NoError(t, err)
	assert.Equal(t, poisoned, got.Data.Media.ID)
	assert.Equal(t, int32(2), s.idCalls.Load(), "the mismatched cache entry is asked again")

	noteAniListDisabled()
	_, err = FetchAnimeFromAniListByID(930002)
	assert.ErrorIs(t, err, ErrAniListAPIDisabled)
	assert.Equal(t, int32(2), s.idCalls.Load(), "a switched-off API is not asked again")
}

func TestFetchAnimeFromAniListByIDLatchesAnAnnouncedOutage(t *testing.T) {
	resetAniListDisabledForTest()
	t.Cleanup(resetAniListDisabledForTest)
	newAniListServer(t, func(aniListRequest) (int, []byte) { return http.StatusForbidden, []byte(aniListDisabledBody) }, nil)
	_, err := FetchAnimeFromAniListByID(930010)
	require.ErrorIs(t, err, ErrAniListAPIDisabled)
	assert.True(t, aniListIsDisabled(), "the rest of the session must not ask AniList again")
}

func TestFetchAniListForAnime(t *testing.T) {
	t.Cleanup(resetAniListDisabledForTest)
	search := func(aniListRequest) (int, []byte) { return http.StatusOK, aniListSuccessBody(5, 6, "By Title") }

	t.Run("an anime with an id is looked up by it", func(t *testing.T) {
		resetAniListDisabledForTest()
		s := newAniListServer(t, func(req aniListRequest) (int, []byte) {
			return http.StatusOK, aniListSuccessBody(req.id(), 52991, "Sousou no Frieren")
		}, search)
		got, err := fetchAniListForAnime(&models.Anime{Name: "Frieren", AnilistID: 940001})
		require.NoError(t, err)
		assert.Equal(t, 940001, got.Data.Media.ID)
		assert.Equal(t, int32(0), s.searchCalls.Load(), "the title is not searched when the id answered")
	})
	t.Run("a failed id lookup falls back to the title", func(t *testing.T) {
		resetAniListDisabledForTest()
		s := newAniListServer(t, func(aniListRequest) (int, []byte) { return http.StatusInternalServerError, nil }, search)
		got, err := fetchAniListForAnime(&models.Anime{Name: "Unique Title 940002", AnilistID: 940002})
		require.NoError(t, err)
		assert.Equal(t, 5, got.Data.Media.ID)
		assert.Equal(t, int32(1), s.idCalls.Load())
		assert.Positive(t, s.searchCalls.Load())
	})
	t.Run("a switched-off API is not searched either", func(t *testing.T) {
		resetAniListDisabledForTest()
		s := newAniListServer(t, func(aniListRequest) (int, []byte) { return http.StatusForbidden, []byte(aniListDisabledBody) }, search)
		_, err := fetchAniListForAnime(&models.Anime{Name: "Unique Title 940003", AnilistID: 940003})
		assert.ErrorIs(t, err, ErrAniListAPIDisabled)
		assert.Equal(t, int32(0), s.searchCalls.Load())
	})
	t.Run("an anime without an id is searched by title", func(t *testing.T) {
		resetAniListDisabledForTest()
		s := newAniListServer(t, func(aniListRequest) (int, []byte) { return http.StatusOK, nil }, search)
		got, err := fetchAniListForAnime(&models.Anime{Name: "Unique Title 940004"})
		require.NoError(t, err)
		assert.Equal(t, 5, got.Data.Media.ID)
		assert.Equal(t, int32(0), s.idCalls.Load())
	})
}

// enrichAnimeData keeps the entry the source named: a title search could
// have landed on another season.
func TestEnrichAnimeDataKeepsTheSourcesAniListEntry(t *testing.T) {
	resetAniListDisabledForTest()
	t.Cleanup(resetAniListDisabledForTest)
	newAniListServer(t, func(req aniListRequest) (int, []byte) {
		return http.StatusOK, aniListSuccessBody(req.id(), 59978, "Sousou no Frieren 2nd Season")
	}, func(aniListRequest) (int, []byte) {
		return http.StatusOK, aniListSuccessBody(154587, 52991, "Sousou no Frieren")
	})
	anime := &models.Anime{Name: "[English] Frieren: Beyond Journey's End Season 2", Source: "Miruro", AnilistID: 950001}
	require.NoError(t, enrichAnimeData(anime))
	assert.Equal(t, 950001, anime.AnilistID)
	assert.Equal(t, 59978, anime.MalID)
	assert.Equal(t, "Sousou no Frieren 2nd Season", anime.Details.Title.Romaji)
}
