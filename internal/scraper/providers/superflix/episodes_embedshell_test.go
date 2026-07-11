package superflix

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// embedShellHTML is what SuperFlix now frequently answers /serie/<tmdb> with: a
// player shell carrying a signed iframe and nothing else. It has no ALL_EPISODES
// blob, no window.allEpisodes, and no data-episode-id anchors — so every parser
// comes back empty. Captured live from /serie/46260 (Naruto).
const embedShellHTML = `<html><head><title>Embed | Naruto</title></head><body>
<script>var __Y="x";</script>
<iframe src="https://superflixapi.pro/serie/46260?embed_expires=1815317791&amp;embed_sig=deadbeef" allowfullscreen></iframe>
</body></html>`

// stubSolver returns a canned page for every Solve call.
type stubSolver struct {
	html     string
	finalURL string
	calls    int
}

func (s *stubSolver) Solve(_ context.Context, _ string, _ time.Duration) (*CFSolveResult, error) {
	s.calls++
	return &CFSolveResult{HTML: s.html, FinalURL: s.finalURL}, nil
}

// TestGetEpisodes_EmbedShellReportsError pins the fix for the silent failure
// behind issue #184's "no seasons found": when the solved page carries no
// episode list, GetEpisodes used to return (nil, nil) — indistinguishable from a
// title that genuinely has no episodes, and surfaced as a bare, unactionable
// "no seasons found". It must report a scrape failure instead.
func TestGetEpisodes_EmbedShellReportsError(t *testing.T) {
	t.Parallel()

	solver := &stubSolver{html: embedShellHTML, finalURL: "https://superflixapi.pro/serie/46260"}
	c := NewSuperFlixClient()
	c.browserSolver = solver

	eps, err := c.GetEpisodes(context.Background(), "46260")

	require.Error(t, err, "an unparseable page must not look like an empty season list")
	assert.Empty(t, eps)
	assert.ErrorIs(t, err, ErrSuperFlixNoEpisodeList)
	// The diagnostic must name what defeated us, so a future rotation is obvious.
	assert.Contains(t, err.Error(), "Embed | Naruto")
	assert.Contains(t, err.Error(), "superflixapi.pro/serie/46260")
}

// A page that DOES carry episodes must still parse — the error above must not
// swallow the healthy path.
func TestGetEpisodes_FrontendPageStillParses(t *testing.T) {
	t.Parallel()

	const frontendHTML = `<html><head><title>Naruto</title></head><body>
	<a data-episode-id="1" data-season="1" data-episode="1" href="/episodio/x1">Ep 1</a>
	<a data-episode-id="2" data-season="1" data-episode="2" href="/episodio/x2">Ep 2</a>
	</body></html>`

	solver := &stubSolver{html: frontendHTML, finalURL: "https://superflixapi.pro/serie/46260"}
	c := NewSuperFlixClient()
	c.browserSolver = solver

	eps, err := c.GetEpisodes(context.Background(), "46260")
	require.NoError(t, err)
	require.Len(t, eps["1"], 2)
}

func TestErrSuperFlixNoEpisodeList_IsDistinctFromNoServers(t *testing.T) {
	t.Parallel()
	// The two signals mean different things (scrape break vs content availability)
	// and describeSuperFlixErr maps them to different user messages.
	assert.False(t, errors.Is(ErrSuperFlixNoEpisodeList, ErrSuperFlixNoServers))
}
