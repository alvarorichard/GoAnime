package appflow

import (
	"errors"
	"os"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api"
	"github.com/alvarorichard/Goanime/v2/internal/api/providers"
	apisource "github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/term"
)

func tryTestFailure() *providers.SearchFailure {
	return &providers.SearchFailure{
		Query:    "namakra gatana",
		Searched: 6,
		Sources: []providers.SourceFailure{
			{Kind: apisource.TopCine, Reason: "is temporarily unavailable (HTTP 503)"},
		},
		AnsweredKinds: []apisource.SourceKind{
			apisource.AnimeFire, apisource.HiAnime, apisource.AnimesDigital,
			apisource.AnimeDrive, apisource.StartFlix,
		},
		Off: []models.SearchOffEntry{{Name: "Goyabu", Reason: "off by default"}},
	}
}

func TestTrySummary(t *testing.T) {
	t.Parallel()

	both := trySummary("namakra gatana",
		[]apisource.SourceKind{apisource.AnimeFire, apisource.HiAnime},
		[]providers.SourceFailure{{Kind: apisource.TopCine}})
	assert.Equal(t, `Nothing for "namakra gatana". 2 sources answered without a match, 1 could not be reached.`, both)

	one := trySummary("x", []apisource.SourceKind{apisource.AnimeFire}, nil)
	assert.Equal(t, `Nothing for "x". 1 source answered without a match.`, one)

	failed := trySummary("x", nil, []providers.SourceFailure{{}, {}})
	assert.Equal(t, `Nothing for "x". 2 could not be reached.`, failed)

	single := trySummary("x", nil, []providers.SourceFailure{{}})
	assert.Equal(t, `Nothing for "x". 1 could not be reached.`, single)

	bare := trySummary("x", nil, nil)
	assert.Equal(t, `Nothing for "x".`, bare)
}

func TestBuildTrySuggestions(t *testing.T) {
	withOverrides(t, appflowOverrides{
		aniList: func(string) (*models.AniListResponse, error) {
			r := &models.AniListResponse{}
			r.Data.Media.Title.English = "Namakura Gatana"
			return r, nil
		},
	})

	got := buildTrySuggestions("namakra gatana", "")
	require.Len(t, got, 3)
	assert.Equal(t, "Namakura Gatana", got[0].Label)
	assert.Equal(t, "did you mean · AniList", got[0].Hint)
	assert.Equal(t, "", got[0].Scope)
	assert.Equal(t, "gatana", got[1].Label)
	assert.Equal(t, "fewer words", got[1].Hint)
	assert.Equal(t, "namakra gatana", got[2].Label)
	assert.Equal(t, "English sources only", got[2].Hint)
	assert.Equal(t, "hianime", got[2].Scope)
}

func TestBuildTrySuggestionsSkips(t *testing.T) {
	withOverrides(t, appflowOverrides{
		aniList: func(string) (*models.AniListResponse, error) {
			return nil, errors.New("offline")
		},
	})

	got := buildTrySuggestions("gatana", "hianime")
	assert.Empty(t, got, "one word: no fewer-words; english scope: no english-only; offline: no did-you-mean")

	got = buildTrySuggestions("gatana", "")
	require.Len(t, got, 1, "english-only still helps a one-word query")
	assert.Equal(t, "gatana", got[0].Label)
	assert.Equal(t, "English sources only", got[0].Hint)
}

func TestBuildTrySuggestionsSameSpellingSkipsDidYouMean(t *testing.T) {
	withOverrides(t, appflowOverrides{
		aniList: func(string) (*models.AniListResponse, error) {
			r := &models.AniListResponse{}
			r.Data.Media.Title.Romaji = "Gatana"
			return r, nil
		},
	})

	got := buildTrySuggestions("gatana", "")
	for _, s := range got {
		assert.NotEqual(t, "did you mean · AniList", s.Hint)
	}
}

func TestBuildTryOptions(t *testing.T) {
	withOverrides(t, appflowOverrides{
		aniList: func(string) (*models.AniListResponse, error) {
			return nil, errors.New("offline")
		},
	})

	opts := buildTryOptions("namakra-gatana", "", tryTestFailure())

	assert.Equal(t, "namakra gatana", opts.Query, "slugs read spaced")
	assert.Contains(t, opts.Summary, `Nothing for "namakra gatana"`)
	assert.Contains(t, opts.Summary, "5 sources answered without a match, 1 could not be reached")
	require.Len(t, opts.Sources, 6)
	assert.Equal(t, "AnimeFire", opts.Sources[0].Name)
	assert.True(t, opts.Sources[0].Ok)
	assert.Equal(t, "TopCine", opts.Sources[5].Name)
	assert.False(t, opts.Sources[5].Ok)
	assert.Equal(t, "temporarily unavailable (HTTP 503)", opts.Sources[5].Detail)
	require.Len(t, opts.Off, 1)
	assert.Equal(t, "Goyabu", opts.Off[0].Name)
	require.Len(t, opts.Suggestions, 2, "no did-you-mean offline: fewer words + english")
	assert.NotEmpty(t, opts.Note)
	assert.NotEmpty(t, opts.SourcesLine)
}

func TestBuildTryOptionsNoResults(t *testing.T) {
	withOverrides(t, appflowOverrides{
		aniList: func(string) (*models.AniListResponse, error) {
			return nil, errors.New("offline")
		},
	})

	err := &providers.NoResultsInfo{
		Query:         "x",
		AnsweredKinds: []apisource.SourceKind{apisource.HiAnime},
	}
	opts := buildTryOptions("x", "hianime", err)
	assert.Contains(t, opts.Summary, "1 source answered without a match")
	require.Len(t, opts.Sources, 1)
	assert.True(t, opts.Sources[0].Ok)
}

func TestBuildTryOptionsGenericError(t *testing.T) {
	withOverrides(t, appflowOverrides{
		aniList: func(string) (*models.AniListResponse, error) {
			return nil, errors.New("offline")
		},
	})

	opts := buildTryOptions("x", "", errors.New("boom"))
	assert.Equal(t, `Nothing for "x".`, opts.Summary)
	assert.Empty(t, opts.Sources)
	assert.NotEmpty(t, opts.Suggestions, "suggestions still help")
}

func TestDidYouMean(t *testing.T) {
	withOverrides(t, appflowOverrides{
		aniList: func(string) (*models.AniListResponse, error) {
			r := &models.AniListResponse{}
			r.Data.Media.Title.English = "Namakura Gatana"
			return r, nil
		},
	})
	assert.Equal(t, "Namakura Gatana", didYouMean("namakra gatana"))
	assert.Equal(t, "", didYouMean("   "))
}

func TestTryAgainDecision(t *testing.T) {
	withOverrides(t, appflowOverrides{
		tryScreen: func(tui.TryOptions) (tui.TryDecision, error) {
			return tui.TryDecision{Query: "gatana", Scope: ""}, nil
		},
	})

	decision, err := tryAgain("namakra gatana", "", errors.New("boom"))
	require.NoError(t, err)
	assert.Equal(t, "gatana", decision.Query)
}

func TestTryAgainUnavailableFallsBack(t *testing.T) {
	withOverrides(t, appflowOverrides{
		tryScreen: func(tui.TryOptions) (tui.TryDecision, error) {
			return tui.TryDecision{}, errTryUnavailable
		},
	})

	_, err := tryAgain("x", "", errors.New("boom"))
	assert.ErrorIs(t, err, errTryUnavailable)
}

func TestTryAgainCancelAborts(t *testing.T) {
	withOverrides(t, appflowOverrides{
		tryScreen: func(tui.TryOptions) (tui.TryDecision, error) {
			return tui.TryDecision{}, tui.ErrSelectionCancelled
		},
	})

	_, err := tryAgain("x", "", errors.New("boom"))
	assert.ErrorIs(t, err, tui.ErrSelectionCancelled)
}

func TestTryAgainEmptyQueryAborts(t *testing.T) {
	withOverrides(t, appflowOverrides{
		tryScreen: func(tui.TryOptions) (tui.TryDecision, error) {
			return tui.TryDecision{Query: "   "}, nil
		},
	})

	_, err := tryAgain("x", "", errors.New("boom"))
	assert.ErrorIs(t, err, api.ErrSearchAborted)
}

func TestSearchAnimeWithRetry_TryScreenRetriesWithScope(t *testing.T) {
	want := &models.Anime{Name: "FoundIt"}
	var gotSource string
	calls := 0
	withOverrides(t, appflowOverrides{
		searchRetry: func(name, source string) (*models.Anime, error) {
			calls++
			if calls == 1 {
				return nil, tryTestFailure()
			}
			gotSource = source
			return want, nil
		},
		tryScreen: func(opts tui.TryOptions) (tui.TryDecision, error) {
			assert.Contains(t, opts.Summary, "5 sources answered without a match")
			return tui.TryDecision{Query: "gatana", Scope: "hianime"}, nil
		},
		aniList: func(string) (*models.AniListResponse, error) {
			return nil, errors.New("offline")
		},
	})

	got, err := SearchAnimeWithRetry("namakra gatana")
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, 2, calls)
	assert.Equal(t, "hianime", gotSource, "the TRY scope drives the retry")
}

func TestSearchAnimeWithRetry_TryCancelAborts(t *testing.T) {
	withOverrides(t, appflowOverrides{
		searchRetry: func(string, string) (*models.Anime, error) {
			return nil, tryTestFailure()
		},
		tryScreen: func(tui.TryOptions) (tui.TryDecision, error) {
			return tui.TryDecision{}, tui.ErrSelectionCancelled
		},
		aniList: func(string) (*models.AniListResponse, error) {
			return nil, errors.New("offline")
		},
	})

	_, err := SearchAnimeWithRetry("x")
	require.Error(t, err)
	assert.ErrorIs(t, err, tui.ErrSelectionCancelled)
}

func TestTrySourcesLine(t *testing.T) {
	assert.Contains(t, trySourcesLine(), "sources on")

	t.Setenv("GOANIME_DISABLED_SOURCES", "hianime")
	assert.Contains(t, trySourcesLine(), "HiAnime")
}

func TestBuildTrySuggestionsCarryScope(t *testing.T) {
	withOverrides(t, appflowOverrides{
		aniList: func(string) (*models.AniListResponse, error) {
			r := &models.AniListResponse{}
			r.Data.Media.Title.English = "X"
			return r, nil
		},
	})

	got := buildTrySuggestions("aa bb", "ptbr")
	require.Len(t, got, 3)
	assert.Equal(t, "ptbr", got[0].Scope, "query suggestions keep the loop scope")
	assert.Equal(t, "ptbr", got[1].Scope)
	assert.Equal(t, "hianime", got[2].Scope, "english-only always scopes english")
}

func TestDidYouMeanPrefersEnglish(t *testing.T) {
	calls := 0
	withOverrides(t, appflowOverrides{
		aniList: func(string) (*models.AniListResponse, error) {
			calls++
			r := &models.AniListResponse{}
			if calls == 1 {
				r.Data.Media.Title.English = "English Title"
				r.Data.Media.Title.Romaji = "Romaji Title"
			} else {
				r.Data.Media.Title.Romaji = "Romaji Title"
			}
			return r, nil
		},
	})
	assert.Equal(t, "English Title", didYouMean("x"))
	assert.Equal(t, "Romaji Title", didYouMean("x"))
}

func TestTryAgainNilErrorStillSuggests(t *testing.T) {
	withOverrides(t, appflowOverrides{
		tryScreen: func(opts tui.TryOptions) (tui.TryDecision, error) {
			assert.Contains(t, opts.Summary, `Nothing for "x"`)
			return tui.TryDecision{Query: "x", Scope: ""}, nil
		},
		aniList: func(string) (*models.AniListResponse, error) {
			return nil, errors.New("offline")
		},
	})

	decision, err := tryAgain("x", "", nil)
	require.NoError(t, err)
	assert.Equal(t, "x", decision.Query)
}

func TestSearchAnimeWithRetry_TryUnavailableUsesLegacyPrompt(t *testing.T) {
	withOverrides(t, appflowOverrides{
		searchRetry: func(string, string) (*models.Anime, error) {
			return nil, errors.New("transport down")
		},
		tryScreen: func(tui.TryOptions) (tui.TryDecision, error) {
			return tui.TryDecision{}, errTryUnavailable
		},
		promptForName: func(current string) (string, error) {
			return "", errors.New("legacy prompt reached and cancelled")
		},
	})

	_, err := SearchAnimeWithRetry("x")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "legacy prompt")
}

func TestSearchAnimeWithRetry_GenericErrorOpensTry(t *testing.T) {
	opened := false
	withOverrides(t, appflowOverrides{
		searchRetry: func(string, string) (*models.Anime, error) {
			return nil, errors.New("transport down")
		},
		tryScreen: func(opts tui.TryOptions) (tui.TryDecision, error) {
			opened = true
			assert.Empty(t, opts.Sources, "no per-source outcomes on a transport error")
			assert.NotEmpty(t, opts.Suggestions)
			return tui.TryDecision{}, tui.ErrSelectionCancelled
		},
		aniList: func(string) (*models.AniListResponse, error) {
			return nil, errors.New("offline")
		},
	})

	_, err := SearchAnimeWithRetry("x")
	require.Error(t, err)
	assert.True(t, opened, "generic failures get the TRY screen, not the bare prompt")
}

func TestSearchAnimeWithRetry_TryScopeSticks(t *testing.T) {
	want := &models.Anime{Name: "FoundIt"}
	var sources []string
	rounds := 0
	withOverrides(t, appflowOverrides{
		searchRetry: func(name, source string) (*models.Anime, error) {
			sources = append(sources, source)
			if len(sources) < 3 {
				return nil, tryTestFailure()
			}
			return want, nil
		},
		tryScreen: func(opts tui.TryOptions) (tui.TryDecision, error) {
			rounds++
			if rounds == 1 {
				assert.Equal(t, "", opts.CurrentScope)
				return tui.TryDecision{Query: "gatana", Scope: "hianime"}, nil
			}
			assert.Equal(t, "hianime", opts.CurrentScope, "the chosen scope sticks for later rounds")
			return tui.TryDecision{Query: "gatana", Scope: "hianime"}, nil
		},
		aniList: func(string) (*models.AniListResponse, error) {
			return nil, errors.New("offline")
		},
	})

	_, err := SearchAnimeWithRetry("x")
	require.NoError(t, err)
	assert.Equal(t, []string{"", "hianime", "hianime"}, sources)
}

func TestDefaultOpenTryScreen_NoTTYUnavailable(t *testing.T) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		t.Skip("needs a non-interactive stdin")
	}
	_, err := defaultOpenTryScreen(tui.TryOptions{})
	assert.ErrorIs(t, err, errTryUnavailable)
}

func TestBuildTryOptionsEmptyQuery(t *testing.T) {
	withOverrides(t, appflowOverrides{
		aniList: func(string) (*models.AniListResponse, error) {
			return nil, errors.New("offline")
		},
	})
	opts := buildTryOptions("", "", nil)
	assert.Equal(t, "", opts.Query)
	assert.Contains(t, opts.Summary, "Nothing")
}

func TestBackToSearchHome(t *testing.T) {
	prev := OpenSearchHomeFn
	t.Cleanup(func() { OpenSearchHomeFn = prev })

	t.Run("nil opener is unavailable", func(t *testing.T) {
		OpenSearchHomeFn = nil
		_, err := backToSearchHome("x")
		assert.ErrorIs(t, err, errTryUnavailable)
	})

	t.Run("cancel aborts", func(t *testing.T) {
		OpenSearchHomeFn = func(string) (tui.SearchHomeResult, error) {
			return tui.SearchHomeResult{}, tui.ErrSearchHomeCancelled
		}
		_, err := backToSearchHome("x")
		assert.ErrorIs(t, err, api.ErrSearchAborted)
	})

	t.Run("search decision", func(t *testing.T) {
		OpenSearchHomeFn = func(query string) (tui.SearchHomeResult, error) {
			assert.Equal(t, "kept", query, "the query is kept for editing")
			return tui.SearchHomeResult{Action: tui.HomeSearch, Query: "new", Scope: "ptbr", HasEntry: false}, nil
		}
		hb, err := backToSearchHome("kept")
		require.NoError(t, err)
		assert.True(t, hb.search)
		assert.Equal(t, "new", hb.query)
		assert.Equal(t, "ptbr", hb.scope)
		assert.Nil(t, hb.cont)
	})

	t.Run("continue decision", func(t *testing.T) {
		entry := tui.HomeContinueEntry{AnimeTitle: "Show", EpisodeNum: 3}
		OpenSearchHomeFn = func(string) (tui.SearchHomeResult, error) {
			return tui.SearchHomeResult{Action: tui.HomeResume, Entry: entry, HasEntry: true}, nil
		}
		hb, err := backToSearchHome("x")
		require.NoError(t, err)
		require.NotNil(t, hb.cont)
		assert.Equal(t, 3, hb.cont.Entry.EpisodeNum)
	})

	t.Run("empty query falls back", func(t *testing.T) {
		OpenSearchHomeFn = func(string) (tui.SearchHomeResult, error) {
			return tui.SearchHomeResult{Action: tui.HomeSearch, Query: "  "}, nil
		}
		_, err := backToSearchHome("x")
		assert.ErrorIs(t, err, errTryUnavailable)
	})

	t.Run("resume without entry falls back", func(t *testing.T) {
		OpenSearchHomeFn = func(string) (tui.SearchHomeResult, error) {
			return tui.SearchHomeResult{Action: tui.HomeResume}, nil
		}
		_, err := backToSearchHome("x")
		assert.ErrorIs(t, err, errTryUnavailable)
	})
}

func TestSearchAnimeWithRetry_BackOpensHome(t *testing.T) {
	prev := OpenSearchHomeFn
	t.Cleanup(func() { OpenSearchHomeFn = prev })

	want := &models.Anime{Name: "FoundIt"}
	calls := 0
	withOverrides(t, appflowOverrides{
		searchRetry: func(name, source string) (*models.Anime, error) {
			calls++
			if calls == 1 {
				return nil, api.ErrBackToSearch
			}
			assert.Equal(t, "edited", name)
			assert.Equal(t, "hianime", source)
			return want, nil
		},
	})
	OpenSearchHomeFn = func(query string) (tui.SearchHomeResult, error) {
		assert.Equal(t, "first", query)
		return tui.SearchHomeResult{Action: tui.HomeSearch, Query: "edited", Scope: "hianime"}, nil
	}

	got, err := SearchAnimeWithRetry("first")
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, 2, calls)
}

func TestSearchAnimeWithRetry_HomeContinuePends(t *testing.T) {
	prev := OpenSearchHomeFn
	t.Cleanup(func() {
		OpenSearchHomeFn = prev
		PendingHomeResult = nil
	})

	withOverrides(t, appflowOverrides{
		searchRetry: func(string, string) (*models.Anime, error) {
			return nil, api.ErrBackToSearch
		},
	})
	entry := tui.HomeContinueEntry{AnimeTitle: "Show", EpisodeNum: 3}
	OpenSearchHomeFn = func(string) (tui.SearchHomeResult, error) {
		return tui.SearchHomeResult{Action: tui.HomeResume, Entry: entry, HasEntry: true}, nil
	}

	anime, err := SearchAnimeWithRetry("x")
	require.NoError(t, err)
	assert.Nil(t, anime)
	require.NotNil(t, PendingHomeResult)
	assert.Equal(t, 3, PendingHomeResult.Entry.EpisodeNum)
}

func TestSearchAnimeWithRetry_HomeCancelAborts(t *testing.T) {
	prev := OpenSearchHomeFn
	t.Cleanup(func() { OpenSearchHomeFn = prev })

	withOverrides(t, appflowOverrides{
		searchRetry: func(string, string) (*models.Anime, error) {
			return nil, api.ErrBackToSearch
		},
	})
	OpenSearchHomeFn = func(string) (tui.SearchHomeResult, error) {
		return tui.SearchHomeResult{}, tui.ErrSearchHomeCancelled
	}

	_, err := SearchAnimeWithRetry("x")
	require.Error(t, err)
	assert.ErrorIs(t, err, api.ErrSearchAborted)
	assert.Nil(t, PendingHomeResult)
}

func TestChangeAnimeHome(t *testing.T) {
	prev := OpenSearchHomeFn
	t.Cleanup(func() {
		OpenSearchHomeFn = prev
		PendingHomeResult = nil
	})
	PendingHomeResult = nil

	t.Run("search decision searches with scope", func(t *testing.T) {
		wantAnime := &models.Anime{Name: "FoundIt"}
		wantEps := []models.Episode{{Number: "1"}}
		withOverrides(t, appflowOverrides{
			searchRetry: func(name, source string) (*models.Anime, error) {
				assert.Equal(t, "edited", name)
				assert.Equal(t, "ptbr", source)
				return wantAnime, nil
			},
			getEpisodes: func(anime *models.Anime) ([]models.Episode, error) {
				assert.Same(t, wantAnime, anime)
				return wantEps, nil
			},
		})
		OpenSearchHomeFn = func(query string) (tui.SearchHomeResult, error) {
			assert.Equal(t, "", query, "change-anime starts blank")
			return tui.SearchHomeResult{Action: tui.HomeSearch, Query: "edited", Scope: "ptbr"}, nil
		}

		anime, episodes, err := ChangeAnimeHome()
		require.NoError(t, err)
		assert.Same(t, wantAnime, anime)
		assert.Equal(t, wantEps, episodes)
		assert.Nil(t, PendingHomeResult)
	})

	t.Run("continue decision pends", func(t *testing.T) {
		entry := tui.HomeContinueEntry{AnimeTitle: "Show", EpisodeNum: 3}
		OpenSearchHomeFn = func(string) (tui.SearchHomeResult, error) {
			return tui.SearchHomeResult{Action: tui.HomeEpisodes, Entry: entry, HasEntry: true}, nil
		}

		_, _, err := ChangeAnimeHome()
		assert.ErrorIs(t, err, ErrHomeContinue)
		require.NotNil(t, PendingHomeResult)
		assert.Equal(t, 3, PendingHomeResult.Entry.EpisodeNum)
		PendingHomeResult = nil
	})

	t.Run("cancel aborts", func(t *testing.T) {
		OpenSearchHomeFn = func(string) (tui.SearchHomeResult, error) {
			return tui.SearchHomeResult{}, tui.ErrSearchHomeCancelled
		}

		_, _, err := ChangeAnimeHome()
		assert.ErrorIs(t, err, api.ErrSearchAborted)
		assert.Nil(t, PendingHomeResult)
	})

	t.Run("search error propagates", func(t *testing.T) {
		withOverrides(t, appflowOverrides{
			searchRetry: func(string, string) (*models.Anime, error) {
				return nil, api.ErrBackToSearch
			},
		})
		OpenSearchHomeFn = func(string) (tui.SearchHomeResult, error) {
			return tui.SearchHomeResult{Action: tui.HomeSearch, Query: "x", Scope: ""}, nil
		}

		_, _, err := ChangeAnimeHome()
		assert.ErrorIs(t, err, api.ErrBackToSearch)
	})

	t.Run("empty episodes errors", func(t *testing.T) {
		withOverrides(t, appflowOverrides{
			searchRetry: func(string, string) (*models.Anime, error) {
				return &models.Anime{Name: "X"}, nil
			},
			getEpisodes: func(*models.Anime) ([]models.Episode, error) {
				return nil, nil
			},
		})
		OpenSearchHomeFn = func(string) (tui.SearchHomeResult, error) {
			return tui.SearchHomeResult{Action: tui.HomeSearch, Query: "x", Scope: ""}, nil
		}

		_, _, err := ChangeAnimeHome()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no episodes")
	})

	t.Run("nil opener keeps legacy prompt", func(t *testing.T) {
		OpenSearchHomeFn = nil
		_, _, err := ChangeAnimeHome()
		require.Error(t, err, "huh errors without a terminal instead of hanging")
	})
}
