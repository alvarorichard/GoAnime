package api

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/providers/superflix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubPrefetchSeams replaces the server-list and stream seams with counters and
// returns them, restoring the originals on cleanup. The stubs are safe to call
// from the prefetch goroutine after the test ends (they touch only their own
// captured state, never testing.T).
type prefetchSeamCalls struct {
	mu               sync.Mutex
	getServersEp     []string
	streamFromServer []string // "serverID episode" per call
}

func stubPrefetchSeams(t *testing.T, servers []superflix.SuperFlixServer) *prefetchSeamCalls {
	t.Helper()
	calls := &prefetchSeamCalls{}
	pl, ps := sfGetServersFn, sfStreamFromServerFn
	t.Cleanup(func() { sfGetServersFn, sfStreamFromServerFn = pl, ps })

	sfGetServersFn = func(_ *superflix.SuperFlixClient, _ context.Context, _, _, _, episode string) ([]superflix.SuperFlixServer, *superflix.SuperFlixTokens, error) {
		calls.mu.Lock()
		calls.getServersEp = append(calls.getServersEp, episode)
		calls.mu.Unlock()
		return servers, &superflix.SuperFlixTokens{ContentID: "1", PageToken: "tok"}, nil
	}
	sfStreamFromServerFn = func(_ *superflix.SuperFlixClient, _ context.Context, _ *superflix.SuperFlixTokens, serverID, _, _, _, episode string) (*superflix.SuperFlixStreamResult, error) {
		calls.mu.Lock()
		calls.streamFromServer = append(calls.streamFromServer, serverID+" "+episode)
		calls.mu.Unlock()
		return &superflix.SuperFlixStreamResult{StreamURL: "https://cdn/x.m3u8"}, nil
	}
	return calls
}

func sfTestServer(id string, audioType int, isFile bool) superflix.SuperFlixServer {
	return superflix.SuperFlixServer{
		ID:     json.RawMessage(`"` + id + `"`),
		Name:   "Servidor " + id,
		Type:   audioType,
		IsFile: isFile,
	}
}

// The warm-up must target episode N+1 and honor the remembered server
// preference silently — no picker, no persisted pick.
func TestMaybePrefetchNextSuperFlixEpisode_WarmsNextEpisode(t *testing.T) {
	dub := sfTestServer("111", superflix.SuperFlixAudioDubbed, false)
	leg := sfTestServer("222", superflix.SuperFlixAudioSubtitled, false)
	calls := stubPrefetchSeams(t, []superflix.SuperFlixServer{dub, leg})

	const tmdbID = "goanime-prefetch-test-424242"
	t.Cleanup(resetSuperFlixServerPrefs)
	resetSuperFlixServerPrefs()
	// The user picked Legendado on the episode that just played.
	rememberSuperFlixServer(tmdbID, leg)

	maybePrefetchNextSuperFlixEpisode(nil, tmdbID, "serie", "1", "3")
	sfPrefetchWG.Wait()

	calls.mu.Lock()
	defer calls.mu.Unlock()
	require.Equal(t, []string{"4"}, calls.getServersEp, "must warm exactly the NEXT episode")
	require.Equal(t, []string{"222 4"}, calls.streamFromServer,
		"must resolve through the remembered (legendado) server, silently")
}

// Guards: anything that is not a numbered series episode must not spawn a
// warm-up at all, and the kill-switch must be honored.
func TestMaybePrefetchNextSuperFlixEpisode_Guards(t *testing.T) {
	tests := []struct {
		name   string
		sfType string
		epNum  string
		env    string
	}{
		{"movie is skipped", "filme", "1", ""},
		{"non-numeric episode is skipped", "serie", "especial", ""},
		{"episode zero is skipped", "serie", "0", ""},
		{"kill-switch is honored", "serie", "3", "1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := stubPrefetchSeams(t, []superflix.SuperFlixServer{
				sfTestServer("111", superflix.SuperFlixAudioDubbed, false),
			})
			if tt.env != "" {
				t.Setenv("GOANIME_SF_NO_PREFETCH", tt.env)
			}

			maybePrefetchNextSuperFlixEpisode(nil, "goanime-prefetch-guard-test", tt.sfType, "1", tt.epNum)
			sfPrefetchWG.Wait()

			calls.mu.Lock()
			defer calls.mu.Unlock()
			assert.Empty(t, calls.getServersEp, "no warm-up may run for this input")
		})
	}
}

// A failing server list must die silently in the background — no panic, no
// stream call — and release the in-flight slot so a later attempt can run.
func TestMaybePrefetchNextSuperFlixEpisode_FailureIsSilentAndReleasesSlot(t *testing.T) {
	pl, ps := sfGetServersFn, sfStreamFromServerFn
	t.Cleanup(func() { sfGetServersFn, sfStreamFromServerFn = pl, ps })

	var mu sync.Mutex
	var attempts int
	sfGetServersFn = func(_ *superflix.SuperFlixClient, _ context.Context, _, _, _, _ string) ([]superflix.SuperFlixServer, *superflix.SuperFlixTokens, error) {
		mu.Lock()
		attempts++
		mu.Unlock()
		return nil, nil, superflix.ErrSuperFlixRateLimited
	}
	streamCalled := false
	sfStreamFromServerFn = func(_ *superflix.SuperFlixClient, _ context.Context, _ *superflix.SuperFlixTokens, _, _, _, _, _ string) (*superflix.SuperFlixStreamResult, error) {
		streamCalled = true
		return nil, nil
	}

	const tmdbID = "goanime-prefetch-fail-test"
	maybePrefetchNextSuperFlixEpisode(nil, tmdbID, "serie", "1", "7")
	sfPrefetchWG.Wait()
	// The in-flight slot must be free again: a second attempt runs.
	maybePrefetchNextSuperFlixEpisode(nil, tmdbID, "serie", "1", "7")
	sfPrefetchWG.Wait()

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 2, attempts, "failed warm-up must release the in-flight slot")
	assert.False(t, streamCalled, "no stream resolve after a failed server list")
}
