package api

import (
	"context"
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/providers/superflix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sfStreamHarness struct {
	listCalls   int
	serverCalls int
	sniffCalls  int
	usedServer  string
}

func (h *sfStreamHarness) install(
	t *testing.T,
	servers []superflix.SuperFlixServer, listErr error,
	serverErr error,
) {
	t.Helper()
	pl, ps, pn := sfGetServersFn, sfStreamFromServerFn, sfSniffStreamFn
	t.Cleanup(func() { sfGetServersFn, sfStreamFromServerFn, sfSniffStreamFn = pl, ps, pn })

	sfGetServersFn = func(_ *superflix.SuperFlixClient, _ context.Context, _, _, _, _ string) ([]superflix.SuperFlixServer, *superflix.SuperFlixTokens, error) {
		h.listCalls++
		return servers, &superflix.SuperFlixTokens{ContentID: "127972"}, listErr
	}
	sfStreamFromServerFn = func(_ *superflix.SuperFlixClient, _ context.Context, _ *superflix.SuperFlixTokens, id string) (*superflix.SuperFlixStreamResult, error) {
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

// The server list is the ONLY path that lets the user choose a source and the
// audio, so it must be tried first — the sniff, which offers neither, is a
// fallback and nothing more.
func TestSuperFlixStream_PrefersTheServerList(t *testing.T) {
	stubServerPicker(t, func([]string) (int, error) { return 1, nil }) // pick the MP4

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
	stubServerPicker(t, func([]string) (int, error) {
		t.Error("nothing to pick from when the list failed")
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

// A server can be listed and still refuse to play. Dead-ending there would strand
// the user on a source they cannot re-pick, so we fall through to the sniff.
func TestSuperFlixStream_ChosenServerFailsFallsBack(t *testing.T) {
	stubServerPicker(t, func([]string) (int, error) { return 0, nil })

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
