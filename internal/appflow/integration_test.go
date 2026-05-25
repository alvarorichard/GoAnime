package appflow

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Test infrastructure — package-level var swap helpers serialised on one mu.
// ---------------------------------------------------------------------------

// appflowOverrideMu serialises every package-level injection swap so race
// detector stays clean across concurrent tests.
var appflowOverrideMu sync.Mutex

// appflowOverrides is the bundle of injectable hooks tests may override.
// Pointer fields = "leave default if nil"; non-nil = swap for test.
type appflowOverrides struct {
	search        func(string, string) (*models.Anime, error)
	searchRetry   func(string, string) (*models.Anime, error)
	aniList       func(string) (*models.AniListResponse, error)
	sourceDetails func(*models.Anime) error
	getEpisodes   func(*models.Anime) ([]models.Episode, error)
	getEpisodesL  func(string) ([]models.Episode, error)
	runSpinner    func(string, func())
	promptForName func(string) (string, error)
}

// withOverrides applies the given overrides for the duration of the test
// and restores the previous values on cleanup. Holds appflowOverrideMu —
// tests that call this MUST NOT use t.Parallel().
func withOverrides(t *testing.T, o appflowOverrides) {
	t.Helper()
	appflowOverrideMu.Lock()

	prev := appflowOverrides{
		search:        searchEnhancedFn,
		searchRetry:   searchWithRetryFn,
		aniList:       aniListFetchFn,
		sourceDetails: sourceDetailsFetchFn,
		getEpisodes:   getAnimeEpisodesEnhancedFn,
		getEpisodesL:  getAnimeEpisodesLegacyFn,
		runSpinner:    runSpinnerFn,
		promptForName: promptForNameFn,
	}

	if o.search != nil {
		searchEnhancedFn = o.search
	}
	if o.searchRetry != nil {
		searchWithRetryFn = o.searchRetry
	}
	if o.aniList != nil {
		aniListFetchFn = o.aniList
	}
	if o.sourceDetails != nil {
		sourceDetailsFetchFn = o.sourceDetails
	}
	if o.getEpisodes != nil {
		getAnimeEpisodesEnhancedFn = o.getEpisodes
	}
	if o.getEpisodesL != nil {
		getAnimeEpisodesLegacyFn = o.getEpisodesL
	}
	if o.runSpinner != nil {
		runSpinnerFn = o.runSpinner
	}
	if o.promptForName != nil {
		promptForNameFn = o.promptForName
	}

	t.Cleanup(func() {
		searchEnhancedFn = prev.search
		searchWithRetryFn = prev.searchRetry
		aniListFetchFn = prev.aniList
		sourceDetailsFetchFn = prev.sourceDetails
		getAnimeEpisodesEnhancedFn = prev.getEpisodes
		getAnimeEpisodesLegacyFn = prev.getEpisodesL
		runSpinnerFn = prev.runSpinner
		promptForNameFn = prev.promptForName
		appflowOverrideMu.Unlock()
	})
}

// passthroughSpinner is the canonical test override for runSpinnerFn:
// runs the action synchronously without any TUI.
func passthroughSpinner(_ string, action func()) { action() }

// makeAniListResp builds a populated AniListResponse for enrichment tests.
func makeAniListResp(animeID, malID int, cover string) *models.AniListResponse {
	r := &models.AniListResponse{}
	r.Data.Media.ID = animeID
	r.Data.Media.IDMal = malID
	r.Data.Media.CoverImage.Large = cover
	r.Data.Media.Title.Romaji = "Mock Title"
	return r
}

// ---------------------------------------------------------------------------
// fetchAnimeDetailsCore — every branch covered with mocked dependencies.
// ---------------------------------------------------------------------------

func TestFetchAnimeDetailsCore_NilAnime(t *testing.T) {
	withOverrides(t, appflowOverrides{})
	assert.NotPanics(t, func() { fetchAnimeDetailsCore(nil) })
}

func TestFetchAnimeDetailsCore_SuperFlixSkipsAll(t *testing.T) {
	var aniCount, srcCount int32
	withOverrides(t, appflowOverrides{
		aniList: func(string) (*models.AniListResponse, error) {
			atomic.AddInt32(&aniCount, 1)
			return nil, nil
		},
		sourceDetails: func(*models.Anime) error {
			atomic.AddInt32(&srcCount, 1)
			return nil
		},
	})

	fetchAnimeDetailsCore(&models.Anime{Name: "Spirited Away", Source: "SuperFlix"})
	assert.Equal(t, int32(0), atomic.LoadInt32(&aniCount), "SuperFlix must skip AniList")
	assert.Equal(t, int32(0), atomic.LoadInt32(&srcCount), "SuperFlix must skip source details")
}

func TestFetchAnimeDetailsCore_SFlixCallsSourceOnly(t *testing.T) {
	var aniCount, srcCount int32
	withOverrides(t, appflowOverrides{
		aniList: func(string) (*models.AniListResponse, error) {
			atomic.AddInt32(&aniCount, 1)
			return nil, nil
		},
		sourceDetails: func(a *models.Anime) error {
			atomic.AddInt32(&srcCount, 1)
			assert.Equal(t, "SFlix", a.Source)
			return nil
		},
	})

	fetchAnimeDetailsCore(&models.Anime{Name: "Inception", Source: "SFlix"})
	assert.Equal(t, int32(0), atomic.LoadInt32(&aniCount), "SFlix must skip AniList")
	assert.Equal(t, int32(1), atomic.LoadInt32(&srcCount), "SFlix must call source details")
}

func TestFetchAnimeDetailsCore_MovieTypeCallsSourceOnly(t *testing.T) {
	var srcCount int32
	withOverrides(t, appflowOverrides{
		sourceDetails: func(*models.Anime) error {
			atomic.AddInt32(&srcCount, 1)
			return nil
		},
	})

	fetchAnimeDetailsCore(&models.Anime{Name: "Movie X", MediaType: models.MediaTypeMovie})
	assert.Equal(t, int32(1), atomic.LoadInt32(&srcCount))
}

func TestFetchAnimeDetailsCore_TVTypeCallsSourceOnly(t *testing.T) {
	var srcCount int32
	withOverrides(t, appflowOverrides{
		sourceDetails: func(*models.Anime) error {
			atomic.AddInt32(&srcCount, 1)
			return nil
		},
	})

	fetchAnimeDetailsCore(&models.Anime{Name: "TV Y", MediaType: models.MediaTypeTV})
	assert.Equal(t, int32(1), atomic.LoadInt32(&srcCount))
}

func TestFetchAnimeDetailsCore_SourceDetailsErrorIgnored(t *testing.T) {
	withOverrides(t, appflowOverrides{
		sourceDetails: func(*models.Anime) error {
			return errors.New("TMDB down")
		},
	})

	// Must not panic — error logged at debug, function returns normally.
	assert.NotPanics(t, func() {
		fetchAnimeDetailsCore(&models.Anime{Name: "Movie", Source: "SFlix"})
	})
}

func TestFetchAnimeDetailsCore_NeedsAniListOnly_Success(t *testing.T) {
	anime := &models.Anime{Name: "Bleach", Source: "AnimeFire"}
	wantResp := makeAniListResp(42, 1234, "https://cdn/large.jpg")
	withOverrides(t, appflowOverrides{
		aniList: func(name string) (*models.AniListResponse, error) {
			assert.Equal(t, "Bleach", name)
			return wantResp, nil
		},
	})

	fetchAnimeDetailsCore(anime)
	assert.Equal(t, 42, anime.AnilistID)
	assert.Equal(t, 1234, anime.MalID)
	assert.Equal(t, "https://cdn/large.jpg", anime.ImageURL)
	assert.Equal(t, "Mock Title", anime.Details.Title.Romaji)
}

func TestFetchAnimeDetailsCore_NeedsAniListOnly_FetchError(t *testing.T) {
	anime := &models.Anime{Name: "Bleach", Source: "AnimeFire"}
	withOverrides(t, appflowOverrides{
		aniList: func(string) (*models.AniListResponse, error) {
			return nil, errors.New("AniList 500")
		},
	})

	fetchAnimeDetailsCore(anime)
	assert.Equal(t, 0, anime.AnilistID, "error means no enrichment")
	assert.Equal(t, "", anime.ImageURL)
}

func TestFetchAnimeDetailsCore_NeedsAniListOnly_NoCoverDoesNotOverwrite(t *testing.T) {
	anime := &models.Anime{Name: "X", Source: "AnimeFire", ImageURL: "existing.jpg"}
	resp := makeAniListResp(1, 2, "") // empty cover
	withOverrides(t, appflowOverrides{
		aniList: func(string) (*models.AniListResponse, error) { return resp, nil },
	})

	fetchAnimeDetailsCore(anime)
	assert.Equal(t, "existing.jpg", anime.ImageURL, "empty cover must not overwrite existing")
}

func TestFetchAnimeDetailsCore_BothNeeded_ParallelEnrichment(t *testing.T) {
	anime := &models.Anime{
		Name:   "AllAnimeShow",
		Source: "AllAnime",
		URL:    "https://allanime.to/anime/abcdefghij1234567890",
	}
	resp := makeAniListResp(7, 99, "https://cover")

	var aniHit, srcHit int32
	withOverrides(t, appflowOverrides{
		aniList: func(string) (*models.AniListResponse, error) {
			atomic.AddInt32(&aniHit, 1)
			return resp, nil
		},
		sourceDetails: func(*models.Anime) error {
			atomic.AddInt32(&srcHit, 1)
			return nil
		},
	})

	fetchAnimeDetailsCore(anime)
	assert.Equal(t, int32(1), atomic.LoadInt32(&aniHit))
	assert.Equal(t, int32(1), atomic.LoadInt32(&srcHit))
	assert.Equal(t, 7, anime.AnilistID)
}

func TestFetchAnimeDetailsCore_BothNeeded_BothErrorNoCrash(t *testing.T) {
	anime := &models.Anime{
		Name:   "AllAnimeShow",
		Source: "AllAnime",
		URL:    "https://allanime.to/anime/abcdefghij1234567890",
	}
	withOverrides(t, appflowOverrides{
		aniList: func(string) (*models.AniListResponse, error) {
			return nil, errors.New("ani fail")
		},
		sourceDetails: func(*models.Anime) error {
			return errors.New("src fail")
		},
	})

	assert.NotPanics(t, func() { fetchAnimeDetailsCore(anime) })
	assert.Equal(t, 0, anime.AnilistID)
}

func TestFetchAnimeDetailsCore_DefaultBranch_AlreadyEnriched(t *testing.T) {
	anime := &models.Anime{
		Name:      "Done",
		Source:    "AnimeFire",
		AnilistID: 1,
		MalID:     2,
		ImageURL:  "x.jpg",
	}
	var aniHit, srcHit int32
	withOverrides(t, appflowOverrides{
		aniList: func(string) (*models.AniListResponse, error) {
			atomic.AddInt32(&aniHit, 1)
			return nil, nil
		},
		sourceDetails: func(*models.Anime) error {
			atomic.AddInt32(&srcHit, 1)
			return nil
		},
	})

	fetchAnimeDetailsCore(anime)
	assert.Equal(t, int32(0), atomic.LoadInt32(&aniHit), "already enriched → no AniList call")
	assert.Equal(t, int32(0), atomic.LoadInt32(&srcHit), "not AllAnime → no source call")
}

func TestFetchAnimeDetailsCore_DefaultBranch_AlreadyEnriched_AllAnimeStillFetchesSource(t *testing.T) {
	anime := &models.Anime{
		Name:      "Done",
		Source:    "AllAnime",
		URL:       "https://allanime.to/anime/abcdefghij1234567890",
		AnilistID: 1,
		MalID:     2,
		ImageURL:  "x.jpg",
	}
	var aniHit, srcHit int32
	withOverrides(t, appflowOverrides{
		aniList: func(string) (*models.AniListResponse, error) {
			atomic.AddInt32(&aniHit, 1)
			return nil, nil
		},
		sourceDetails: func(*models.Anime) error {
			atomic.AddInt32(&srcHit, 1)
			return nil
		},
	})

	fetchAnimeDetailsCore(anime)
	assert.Equal(t, int32(0), atomic.LoadInt32(&aniHit), "already enriched → no AniList call")
	assert.Equal(t, int32(1), atomic.LoadInt32(&srcHit), "AllAnime URL must trigger source details")
}

func TestFetchAnimeDetailsCore_DefaultBranch_SourceDetailsError(t *testing.T) {
	anime := &models.Anime{
		Name:      "Done",
		Source:    "AllAnime",
		URL:       "https://allanime.to/anime/abcdefghij1234567890",
		AnilistID: 1,
		MalID:     2,
		ImageURL:  "x.jpg",
	}
	withOverrides(t, appflowOverrides{
		sourceDetails: func(*models.Anime) error { return errors.New("allanime down") },
	})
	assert.NotPanics(t, func() { fetchAnimeDetailsCore(anime) })
}

// ---------------------------------------------------------------------------
// FetchAnimeDetails — exercises the full pipeline including the spinner
// wrapper (passthrough mock). Confirms runSpinnerFn is wired correctly.
// ---------------------------------------------------------------------------

func TestFetchAnimeDetails_FullPipeline_WithPassthroughSpinner(t *testing.T) {
	var spinnerCalls int32
	var spinnerTitle string
	withOverrides(t, appflowOverrides{
		runSpinner: func(title string, action func()) {
			atomic.AddInt32(&spinnerCalls, 1)
			spinnerTitle = title
			action()
		},
		aniList: func(string) (*models.AniListResponse, error) {
			return makeAniListResp(100, 200, "cv.jpg"), nil
		},
	})

	anime := &models.Anime{Name: "Test", Source: "AnimeFire"}
	FetchAnimeDetails(anime)

	assert.Equal(t, int32(1), atomic.LoadInt32(&spinnerCalls))
	assert.Equal(t, "Fetching anime details...", spinnerTitle)
	assert.Equal(t, 100, anime.AnilistID)
}

// ---------------------------------------------------------------------------
// GetAnimeEpisodes — every branch with mocked episode fetcher.
// ---------------------------------------------------------------------------

func TestGetAnimeEpisodes_NonSFlix_SpinnerSuccess(t *testing.T) {
	want := []models.Episode{{Num: 1, Number: "1"}, {Num: 2, Number: "2"}}
	var spinnerHit int32
	withOverrides(t, appflowOverrides{
		runSpinner: func(_ string, a func()) { atomic.AddInt32(&spinnerHit, 1); a() },
		getEpisodes: func(*models.Anime) ([]models.Episode, error) {
			return want, nil
		},
	})

	got, err := GetAnimeEpisodes(&models.Anime{Name: "X", Source: "AnimeFire"})
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, int32(1), atomic.LoadInt32(&spinnerHit), "non-SFlix must use spinner")
}

func TestGetAnimeEpisodes_SFlix_BypassesSpinner_Success(t *testing.T) {
	want := []models.Episode{{Num: 1, Number: "1"}}
	var spinnerHit int32
	withOverrides(t, appflowOverrides{
		runSpinner:  func(_ string, _ func()) { atomic.AddInt32(&spinnerHit, 1) },
		getEpisodes: func(*models.Anime) ([]models.Episode, error) { return want, nil },
	})

	got, err := GetAnimeEpisodes(&models.Anime{Name: "Y", Source: "SFlix"})
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, int32(0), atomic.LoadInt32(&spinnerHit), "SFlix must NOT use spinner")
}

func TestGetAnimeEpisodes_FetchErrorWrapped(t *testing.T) {
	withOverrides(t, appflowOverrides{
		runSpinner:  passthroughSpinner,
		getEpisodes: func(*models.Anime) ([]models.Episode, error) { return nil, errors.New("net down") },
	})

	_, err := GetAnimeEpisodes(&models.Anime{Name: "Z", Source: "AnimeFire"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch episodes")
	assert.Contains(t, err.Error(), "net down")
}

func TestGetAnimeEpisodes_EmptyListIsError(t *testing.T) {
	withOverrides(t, appflowOverrides{
		runSpinner:  passthroughSpinner,
		getEpisodes: func(*models.Anime) ([]models.Episode, error) { return []models.Episode{}, nil },
	})

	_, err := GetAnimeEpisodes(&models.Anime{Name: "Z", Source: "AnimeFire"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not have episodes")
}

func TestGetAnimeEpisodes_MovieType_BypassesSpinner_Success(t *testing.T) {
	want := []models.Episode{{Num: 1, Number: "1"}}
	withOverrides(t, appflowOverrides{
		runSpinner:  func(_ string, _ func()) { t.Fatalf("movie must skip spinner") },
		getEpisodes: func(*models.Anime) ([]models.Episode, error) { return want, nil },
	})

	got, err := GetAnimeEpisodes(&models.Anime{Name: "M", MediaType: models.MediaTypeMovie})
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

// ---------------------------------------------------------------------------
// GetAnimeEpisodesLegacy — URL-based legacy path.
// ---------------------------------------------------------------------------

func TestGetAnimeEpisodesLegacy_Success(t *testing.T) {
	want := []models.Episode{{Num: 1, Number: "1"}, {Num: 2, Number: "2"}}
	var gotURL string
	withOverrides(t, appflowOverrides{
		runSpinner:   passthroughSpinner,
		getEpisodesL: func(url string) ([]models.Episode, error) { gotURL = url; return want, nil },
	})

	got, err := GetAnimeEpisodesLegacy("https://ex/foo")
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, "https://ex/foo", gotURL, "URL must be forwarded verbatim")
}

func TestGetAnimeEpisodesLegacy_FetchErrorWrapped(t *testing.T) {
	withOverrides(t, appflowOverrides{
		runSpinner:   passthroughSpinner,
		getEpisodesL: func(string) ([]models.Episode, error) { return nil, errors.New("404") },
	})

	_, err := GetAnimeEpisodesLegacy("https://ex/foo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch episodes")
}

func TestGetAnimeEpisodesLegacy_EmptyListIsError(t *testing.T) {
	withOverrides(t, appflowOverrides{
		runSpinner:   passthroughSpinner,
		getEpisodesL: func(string) ([]models.Episode, error) { return nil, nil },
	})

	_, err := GetAnimeEpisodesLegacy("https://ex/foo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not have episodes")
}

// ---------------------------------------------------------------------------
// SearchAnimeWithRetry — full multi-iteration retry loop with scripted prompt.
// ---------------------------------------------------------------------------

func TestSearchAnimeWithRetry_RetriesUntilSuccess(t *testing.T) {
	want := &models.Anime{Name: "FoundIt"}
	var attempts int32
	withOverrides(t, appflowOverrides{
		searchRetry: func(name, _ string) (*models.Anime, error) {
			n := atomic.AddInt32(&attempts, 1)
			if n < 3 {
				return nil, errors.New("not yet")
			}
			assert.Equal(t, "third", name, "third attempt must use prompted name")
			return want, nil
		},
		promptForName: func(_ string) (string, error) {
			n := atomic.LoadInt32(&attempts)
			switch n {
			case 1:
				return "second", nil
			case 2:
				return "third", nil
			}
			return "", errors.New("unexpected prompt call")
		},
	})

	got, err := SearchAnimeWithRetry("initial")
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestSearchAnimeWithRetry_BackToSearchBranch(t *testing.T) {
	want := &models.Anime{Name: "FoundIt"}
	var attempts int32
	withOverrides(t, appflowOverrides{
		searchRetry: func(name, _ string) (*models.Anime, error) {
			n := atomic.AddInt32(&attempts, 1)
			if n == 1 {
				return nil, api.ErrBackToSearch
			}
			return want, nil
		},
		promptForName: func(string) (string, error) { return "newname", nil },
	})

	got, err := SearchAnimeWithRetry("first")
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
}

func TestSearchAnimeWithRetry_PromptCancelled(t *testing.T) {
	withOverrides(t, appflowOverrides{
		searchRetry: func(string, string) (*models.Anime, error) {
			return nil, errors.New("no result")
		},
		promptForName: func(string) (string, error) {
			return "", errors.New("search cancelled by user")
		},
	})

	_, err := SearchAnimeWithRetry("anything")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")
}

func TestSearchAnimeWithRetry_NilAnimeContinuesLoop(t *testing.T) {
	want := &models.Anime{Name: "Eventually"}
	var attempts int32
	withOverrides(t, appflowOverrides{
		searchRetry: func(string, string) (*models.Anime, error) {
			n := atomic.AddInt32(&attempts, 1)
			if n == 1 {
				return nil, nil // nil anime, no error → continues
			}
			return want, nil
		},
		promptForName: func(string) (string, error) { return "retry", nil },
	})

	got, err := SearchAnimeWithRetry("first")
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

// ---------------------------------------------------------------------------
// enrichFromAniList — pure helper, table-driven edge cases.
// ---------------------------------------------------------------------------

func TestEnrichFromAniList(t *testing.T) {
	cases := []struct {
		name        string
		resp        *models.AniListResponse
		respErr     error
		anime       *models.Anime
		wantAniID   int
		wantMalID   int
		wantImage   string
		wantUntouch bool
	}{
		{
			name:      "success_overwrites_image",
			resp:      makeAniListResp(10, 20, "newcv.jpg"),
			anime:     &models.Anime{Name: "x", ImageURL: "old.jpg"},
			wantAniID: 10, wantMalID: 20, wantImage: "newcv.jpg",
		},
		{
			name:      "success_empty_cover_keeps_existing",
			resp:      makeAniListResp(1, 2, ""),
			anime:     &models.Anime{Name: "x", ImageURL: "existing.jpg"},
			wantAniID: 1, wantMalID: 2, wantImage: "existing.jpg",
		},
		{
			name:        "error_leaves_anime_untouched",
			respErr:     errors.New("boom"),
			anime:       &models.Anime{Name: "x", AnilistID: 99, MalID: 88, ImageURL: "keep.jpg"},
			wantAniID:   99,
			wantMalID:   88,
			wantImage:   "keep.jpg",
			wantUntouch: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withOverrides(t, appflowOverrides{
				aniList: func(string) (*models.AniListResponse, error) {
					return tc.resp, tc.respErr
				},
			})
			enrichFromAniList(tc.anime)
			assert.Equal(t, tc.wantAniID, tc.anime.AnilistID)
			assert.Equal(t, tc.wantMalID, tc.anime.MalID)
			assert.Equal(t, tc.wantImage, tc.anime.ImageURL)
		})
	}
}

// ---------------------------------------------------------------------------
// defaultPromptForName — non-TTY path returns the expected sentinel error.
// (TTY path requires a real terminal — cannot test headlessly.)
// ---------------------------------------------------------------------------

func TestDefaultPromptForName_NoTTYReturnsCancelled(t *testing.T) {
	// Lock the mutex even though we don't swap — the package-level prompt is
	// shared state and the default impl reads no test-mutable globals; this
	// test only invokes it, but defending against concurrent swaps is cheap.
	appflowOverrideMu.Lock()
	defer appflowOverrideMu.Unlock()

	_, err := defaultPromptForName("foo")
	// Without a TTY this returns either "cancelled by user" or "cancelled: empty".
	// Both are valid sentinels for the no-TTY path; we just need a non-nil error.
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")
}

// ---------------------------------------------------------------------------
// defaultRunSpinner — symbol pin (requires Bubble Tea runtime).
// Production paths are integration-tested by leaving runSpinnerFn at default
// in the existing TTY-skipping tests; here we just verify the helper exists.
// ---------------------------------------------------------------------------

func TestDefaultRunSpinner_SymbolPin(t *testing.T) {
	t.Parallel()
	_ = defaultRunSpinner
}

// ---------------------------------------------------------------------------
// Defaults wired correctly (defensive — same pattern as playback package).
// ---------------------------------------------------------------------------

func TestPackageDefaults_AllWired(t *testing.T) {
	appflowOverrideMu.Lock()
	defer appflowOverrideMu.Unlock()

	require.NotNil(t, searchEnhancedFn)
	require.NotNil(t, searchWithRetryFn)
	require.NotNil(t, aniListFetchFn)
	require.NotNil(t, sourceDetailsFetchFn)
	require.NotNil(t, getAnimeEpisodesEnhancedFn)
	require.NotNil(t, getAnimeEpisodesLegacyFn)
	require.NotNil(t, runSpinnerFn)
	require.NotNil(t, promptForNameFn)
}

// touch fmt to silence linter if future test removal drops usage.
var _ = fmt.Sprintf
