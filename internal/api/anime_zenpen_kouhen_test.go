package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Regression test for the Zenpen / Kouhen mismatch bug.
//
// Before the fix, CleanTitle stripped everything after " - ", collapsing
//
//	"[PT-BR] Jujutsu Kaisen: Shimetsu Kaiyuu - Zenpen"   (AniList id 172463, 前編)
//
// down to "Jujutsu Kaisen: Shimetsu Kaiyuu", which AniList resolves to the
// sibling entry
//
//	"Jujutsu Kaisen: Shimetsu Kaiyuu - Kouhen"           (AniList id 209895, 後編)
//
// causing every downstream metadata consumer (Discord RPC, AniSkip, file
// naming, tracking) to bind to the wrong anime.
//
// These tests use an httptest.Server that mimics AniList's actual behaviour
// (proven via live curl during the fix) and assert end-to-end that the
// pipeline now selects the correct entry for every Zenpen/Kouhen input,
// while still stripping legitimate PT-BR noise after a hyphen.
// ============================================================================

// aniListMockEntry is a tiny subset of an AniList Media object — enough for
// FetchAnimeFromAniListWithURL to populate AniListResponse.
type aniListMockEntry struct {
	id     int
	idMal  int
	romaji string
}

// jujutsuMockDataset reproduces the exact subset of AniList we hit during
// the bug. The "bare prefix" entry encodes AniList's real behaviour — when
// the buggy CleanTitle output (no Zenpen/Kouhen suffix) is used as the
// search term, AniList returns the Kouhen entry. Verified live before
// freezing the dataset here.
var jujutsuMockDataset = map[string]aniListMockEntry{
	"jujutsu kaisen: shimetsu kaiyuu - zenpen": {
		id:     172463,
		idMal:  57658,
		romaji: "Jujutsu Kaisen: Shimetsu Kaiyuu - Zenpen",
	},
	"jujutsu kaisen: shimetsu kaiyuu - kouhen": {
		id:     209895,
		idMal:  0,
		romaji: "Jujutsu Kaisen: Shimetsu Kaiyuu - Kouhen",
	},
	// Bare prefix — what the OLD broken CleanTitle produced. AniList resolves
	// this to Kouhen in production, which is the data-corruption path.
	"jujutsu kaisen: shimetsu kaiyuu": {
		id:     209895,
		idMal:  0,
		romaji: "Jujutsu Kaisen: Shimetsu Kaiyuu - Kouhen",
	},
	// A control title used by the noise-still-stripped sanity test.
	"one piece": {
		id:     21,
		idMal:  21,
		romaji: "One Piece",
	},
}

// startMockAniListServer spins up an httptest.Server speaking AniList's
// GraphQL contract for the dataset above. Returns the server and a teardown
// helper that restores the original endpoint URL.
func startMockAniListServer(t *testing.T) (server *httptest.Server, cleanup func()) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "expected POST", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			Query     string `json:"query"`
			Variables struct {
				Search string `json:"search"`
			} `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}

		key := strings.ToLower(strings.TrimSpace(payload.Variables.Search))
		entry, found := jujutsuMockDataset[key]

		w.Header().Set("Content-Type", "application/json")
		if !found {
			// AniList returns 200 + null Media for unknown searches; mirror that.
			_, _ = w.Write([]byte(`{"data":{"Media":null}}`))
			return
		}

		resp := map[string]any{
			"data": map[string]any{
				"Media": map[string]any{
					"id":    entry.id,
					"idMal": entry.idMal,
					"title": map[string]any{
						"romaji":  entry.romaji,
						"english": nil,
						"native":  nil,
					},
					"coverImage": map[string]any{"large": ""},
					"synonyms":   []string{},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))

	originalEndpoint := aniListEndpoint
	aniListEndpoint = srv.URL

	teardown := func() {
		aniListEndpoint = originalEndpoint
		srv.Close()
	}
	return srv, teardown
}

// clearAniListCacheKey scrubs the AniList cache for a specific cleaned title
// so a previous subtest's hit cannot mask the current one.
func clearAniListCacheKey(cleanedTitle string) {
	cache := util.GetAniListCache()
	cache.Set("anilist:"+strings.ToLower(cleanedTitle), []byte(`{}`))
}

// TestZenpenKouhenBugFix is the full end-to-end regression. It exercises
// CleanTitle + the AniList query path through a mocked GraphQL server and
// asserts the correct AniList ID is selected for every variant of the bug.
func TestZenpenKouhenBugFix(t *testing.T) {
	_, teardown := startMockAniListServer(t)
	defer teardown()

	t.Run("preserves Zenpen and resolves to AniList id 172463", func(t *testing.T) {
		input := "[PT-BR] Jujutsu Kaisen: Shimetsu Kaiyuu - Zenpen"

		cleaned := CleanTitle(input)
		require.Equal(t, "Jujutsu Kaisen: Shimetsu Kaiyuu - Zenpen", cleaned,
			"CleanTitle must NOT strip the Zenpen suffix")

		clearAniListCacheKey(cleaned)
		resp, err := FetchAnimeFromAniListWithURL(input, "")
		require.NoError(t, err)

		assert.Equal(t, 172463, resp.Data.Media.ID,
			"must bind to Zenpen (172463), not Kouhen (209895)")
		assert.Equal(t, 57658, resp.Data.Media.IDMal)
		assert.Equal(t, "Jujutsu Kaisen: Shimetsu Kaiyuu - Zenpen", resp.Data.Media.Title.Romaji)
	})

	t.Run("preserves Kouhen and resolves to AniList id 209895", func(t *testing.T) {
		input := "[PT-BR] Jujutsu Kaisen: Shimetsu Kaiyuu - Kouhen"

		cleaned := CleanTitle(input)
		require.Equal(t, "Jujutsu Kaisen: Shimetsu Kaiyuu - Kouhen", cleaned,
			"CleanTitle must NOT strip the Kouhen suffix")

		clearAniListCacheKey(cleaned)
		resp, err := FetchAnimeFromAniListWithURL(input, "")
		require.NoError(t, err)

		assert.Equal(t, 209895, resp.Data.Media.ID, "must bind to Kouhen (209895)")
		assert.Equal(t, "Jujutsu Kaisen: Shimetsu Kaiyuu - Kouhen", resp.Data.Media.Title.Romaji)
	})

	// Demonstrates what the ORIGINAL bug looked like end-to-end: feed the
	// pre-stripped title (what the old CleanTitle emitted) directly into the
	// AniList path and observe the wrong-anime binding. This locks in the
	// proof that the fix is load-bearing — if anyone ever reverts CleanTitle,
	// this test still passes (AniList legitimately resolves the bare prefix
	// to Kouhen) while the two tests above start failing.
	t.Run("buggy stripped title would have bound to Kouhen (proves the bug)", func(t *testing.T) {
		preStripped := "Jujutsu Kaisen: Shimetsu Kaiyuu"

		clearAniListCacheKey(preStripped)
		resp, err := FetchAnimeFromAniListWithURL(preStripped, "")
		require.NoError(t, err)

		assert.Equal(t, 209895, resp.Data.Media.ID,
			"the buggy stripped title resolves to Kouhen — this is the data corruption the fix prevents")
		assert.NotEqual(t, 172463, resp.Data.Media.ID,
			"sanity: the buggy stripped title must NOT resolve to Zenpen")
	})

	t.Run("still strips PT-BR noise after hyphen so legitimate titles match", func(t *testing.T) {
		input := "[PT-BR] One Piece - Dublado"

		cleaned := CleanTitle(input)
		require.Equal(t, "One Piece", cleaned,
			"CleanTitle must still strip ' - Dublado' as noise")

		clearAniListCacheKey(cleaned)
		resp, err := FetchAnimeFromAniListWithURL(input, "")
		require.NoError(t, err)

		assert.Equal(t, 21, resp.Data.Media.ID, "must resolve One Piece by stripped title")
	})
}
