package api

import (
	"context"
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/superflix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sfStreamHarness struct {
	cacheCalls  int
	listCalls   int
	serverCalls int
	sniffCalls  int
	usedServer  string
	// cached, when non-nil, makes the cache fast-path hit.
	cached *superflix.SuperFlixStreamResult
}

func (h *sfStreamHarness) install(
	t *testing.T,
	servers []superflix.SuperFlixServer, listErr error,
	serverErr error,
) {
	t.Helper()
	pc, pl, ps, pn := sfCachedStreamFn, sfGetServersFn, sfStreamFromServerFn, sfSniffStreamFn
	t.Cleanup(func() {
		sfCachedStreamFn, sfGetServersFn, sfStreamFromServerFn, sfSniffStreamFn = pc, pl, ps, pn
	})

	sfCachedStreamFn = func(_ *superflix.SuperFlixClient, _, _, _, _ string) (*superflix.SuperFlixStreamResult, bool) {
		h.cacheCalls++
		return h.cached, h.cached != nil
	}
	sfGetServersFn = func(_ *superflix.SuperFlixClient, _ context.Context, _, _, _, _ string) ([]superflix.SuperFlixServer, *superflix.SuperFlixTokens, error) {
		h.listCalls++
		return servers, &superflix.SuperFlixTokens{ContentID: "127972"}, listErr
	}
	sfStreamFromServerFn = func(_ *superflix.SuperFlixClient, _ context.Context, _ *superflix.SuperFlixTokens, id, _, _, _, _ string) (*superflix.SuperFlixStreamResult, error) {
		h.serverCalls++
		h.usedServer = id
		if serverErr != nil {
			return nil, serverErr
		}
		return &superflix.SuperFlixStreamResult{StreamURL: "https://cdn/chosen.m3u8"}, nil
	}
	sfSniffStreamFn = func(_ *superflix.SuperFlixClient, _ context.Context, _, _, _, _ string) (*superflix.SuperFlixStreamResult, error) {
		h.sniffCalls++
		return &superflix.SuperFlixStreamResult{StreamURL: "https://cdn/sniffed.m3u8"}, nil
	}
}

// The cache is the FAST path: a previously-resolved episode replays with no
// server list, no sniff, and no browser. It must be tried before anything else —
// that is the whole point of it, turning a re-watch/resume into a ~1s open.
func TestSuperFlixStream_CacheFastPathSkipsEverything(t *testing.T) {
	stubSFPicker(t, func(prompt string, _ []string) (int, error) {
		t.Errorf("a cached play must not prompt: %q", prompt)
		return 0, nil
	})

	h := &sfStreamHarness{cached: &superflix.SuperFlixStreamResult{StreamURL: "https://cdn/cached.m3u8"}}
	h.install(t, []superflix.SuperFlixServer{sfServer("1", dub, "S", false)}, nil, nil)

	res, chosen, err := superFlixStream(nil, "103913", "serie", "1", "1")
	require.NoError(t, err)

	assert.Equal(t, "https://cdn/cached.m3u8", res.StreamURL)
	assert.Nil(t, chosen, "a cached play returns no server; the choice is honored from memory")
	assert.Equal(t, 1, h.cacheCalls)
	assert.Zero(t, h.listCalls, "the server list must not be fetched on a cache hit")
	assert.Zero(t, h.serverCalls)
	assert.Zero(t, h.sniffCalls, "and no browser sniff either")
}

// On a cache MISS the normal resolution must proceed — the fast path must not
// swallow first plays.
func TestSuperFlixStream_CacheMissProceedsToServerList(t *testing.T) {
	stubSFPicker(t, func(_ string, _ []string) (int, error) { return 0, nil })

	h := &sfStreamHarness{} // cached == nil → miss
	h.install(t, []superflix.SuperFlixServer{sfServer("1", dub, "S", false)}, nil, nil)

	res, _, err := superFlixStream(nil, "103913", "serie", "1", "1")
	require.NoError(t, err)

	assert.Equal(t, "https://cdn/chosen.m3u8", res.StreamURL)
	assert.Equal(t, 1, h.cacheCalls, "the cache is always checked first")
	assert.Equal(t, 1, h.listCalls, "a miss falls through to the server list")
}

// The server list is the ONLY path that lets the user choose a source and the
// audio, so it must be tried before the sniff — the sniff, which offers neither,
// is a fallback and nothing more.
func TestSuperFlixStream_PrefersTheServerList(t *testing.T) {
	stubSFPicker(t, func(_ string, _ []string) (int, error) { return 1, nil }) // second server

	h := &sfStreamHarness{}
	h.install(t, []superflix.SuperFlixServer{
		sfServer("159462", dub, "Servidor 159462", false),
		sfServer("native_media:233831", dub, "MP4 Dublado", true),
	}, nil, nil)

	res, chosen, err := superFlixStream(nil, "103913", "serie", "1", "1")
	require.NoError(t, err)

	assert.Equal(t, "https://cdn/chosen.m3u8", res.StreamURL)
	require.NotNil(t, chosen, "the caller must learn which server was used, to derive the audio from it")
	assert.Equal(t, "MP4 Dublado", chosen.Name)
	assert.Equal(t, "native_media:233831", h.usedServer, "must stream from the server the user picked")
	assert.Zero(t, h.sniffCalls, "the sniff must not run when the user could choose")
}

// When the list is unreachable (the site's shell, a rate limit…), playback must
// still happen — just without a choice.
func TestSuperFlixStream_FallsBackToSniff(t *testing.T) {
	stubSFPicker(t, func(prompt string, _ []string) (int, error) {
		t.Errorf("nothing to pick from when the list failed: %q", prompt)
		return 0, nil
	})

	h := &sfStreamHarness{}
	h.install(t, nil, superflix.ErrSuperFlixRateLimited, nil)

	res, chosen, err := superFlixStream(nil, "103913", "serie", "1", "1")
	require.NoError(t, err)

	assert.Equal(t, "https://cdn/sniffed.m3u8", res.StreamURL)
	assert.Nil(t, chosen, "a nil server tells the caller it must ask about the audio itself")
	assert.Equal(t, 1, h.sniffCalls)
}

func TestSuperFlixStream_RestrictedServerListDoesNotRepeatTheSniff(t *testing.T) {
	h := &sfStreamHarness{}
	h.install(t, nil, superflix.ErrSuperFlixRestricted, nil)

	res, chosen, err := superFlixStream(nil, "121390", "filme", "", "")

	require.ErrorIs(t, err, superflix.ErrSuperFlixRestricted)
	assert.Nil(t, res)
	assert.Nil(t, chosen)
	assert.Equal(t, 1, h.listCalls)
	assert.Zero(t, h.sniffCalls, "the terminal restricted shell must not be solved twice")
}

// A server can be listed and still refuse to play. Dead-ending there would strand
// the user on a source they cannot re-pick, so we fall through to the sniff.
func TestSuperFlixStream_ChosenServerFailsFallsBack(t *testing.T) {
	stubSFPicker(t, func(_ string, _ []string) (int, error) { return 0, nil })

	h := &sfStreamHarness{}
	h.install(t, []superflix.SuperFlixServer{
		sfServer("1", dub, "Servidor 1", false),
		sfServer("2", leg, "Servidor 2", false),
	}, nil, errors.New("server refused"))

	res, chosen, err := superFlixStream(nil, "103913", "serie", "1", "1")
	require.NoError(t, err)

	assert.Equal(t, "https://cdn/sniffed.m3u8", res.StreamURL)
	assert.Nil(t, chosen)
	assert.Equal(t, 1, h.serverCalls)
	assert.Equal(t, 1, h.sniffCalls)
}

// installReleaseSpy stubs the browser-release seam and returns a call counter.
func installReleaseSpy(t *testing.T) *int {
	t.Helper()
	prev := sfReleaseBrowserFn
	n := 0
	sfReleaseBrowserFn = func() { n++ }
	t.Cleanup(func() { sfReleaseBrowserFn = prev })
	return &n
}

// The solver window must be closed after EVERY resolve path so it never lingers
// through playback: cache fast-path, server-list success, sniff fallback, and the
// error path. This cascade drives GetSuperFlixStreamURL through each and asserts
// the release fires exactly once.
func TestGetSuperFlixStreamURL_ReleasesBrowserOnEveryPath(t *testing.T) {
	movie := &models.Anime{Source: "SuperFlix", MediaType: models.MediaTypeMovie, URL: "1"}
	ep := &models.Episode{URL: "1"}

	t.Run("cache fast-path", func(t *testing.T) {
		released := installReleaseSpy(t)
		h := &sfStreamHarness{cached: &superflix.SuperFlixStreamResult{StreamURL: "https://cdn/cached.m3u8"}}
		h.install(t, nil, nil, nil)

		url, err := GetSuperFlixStreamURL(movie, ep, "best")
		require.NoError(t, err)
		assert.Equal(t, "https://cdn/cached.m3u8", url)
		assert.Zero(t, h.listCalls, "cache hit must not touch the server list")
		assert.Equal(t, 1, *released, "the window must be released after a cache hit too")
	})

	t.Run("server-list success", func(t *testing.T) {
		released := installReleaseSpy(t)
		stubSFPicker(t, func(string, []string) (int, error) { return 0, nil })
		h := &sfStreamHarness{}
		h.install(t, []superflix.SuperFlixServer{sfServer("9", dub, "S", false)}, nil, nil)

		url, err := GetSuperFlixStreamURL(movie, ep, "best")
		require.NoError(t, err)
		assert.Equal(t, "https://cdn/chosen.m3u8", url)
		assert.Equal(t, 1, *released)
	})

	t.Run("sniff fallback", func(t *testing.T) {
		released := installReleaseSpy(t)
		h := &sfStreamHarness{}
		h.install(t, nil, superflix.ErrSuperFlixRateLimited, nil)

		url, err := GetSuperFlixStreamURL(movie, ep, "best")
		require.NoError(t, err)
		assert.Equal(t, "https://cdn/sniffed.m3u8", url)
		assert.Equal(t, 1, *released, "even the sniff path must release the window")
	})

	t.Run("error path still releases", func(t *testing.T) {
		released := installReleaseSpy(t)
		pn := sfSniffStreamFn
		t.Cleanup(func() { sfSniffStreamFn = pn })
		h := &sfStreamHarness{}
		h.install(t, nil, superflix.ErrSuperFlixRateLimited, nil)
		// Make the sniff fail too, so the whole resolve errors.
		sfSniffStreamFn = func(*superflix.SuperFlixClient, context.Context, string, string, string, string) (*superflix.SuperFlixStreamResult, error) {
			return nil, errors.New("boom")
		}

		_, err := GetSuperFlixStreamURL(movie, ep, "best")
		require.Error(t, err)
		assert.Equal(t, 1, *released, "a failed resolve must not leave the window open")
	})
}
