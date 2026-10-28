package scraper

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/startflix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Live shapes captured 2026-10-16. StartFlix searches its own catalog but the
// episode URLs resolve on the panel, so the stub serves both halves and the
// cascade proves search → stream through the real adapter glue.

// TestStartFlixAdapterCascade_SearchToStream drives a catalog search and then
// resolves an episode from that same stubbed world, pinning the metadata
// contract the player depends on (source, slashed referer, host).
func TestStartFlixAdapterCascade_SearchToStream(t *testing.T) {
	t.Parallel()

	adapter := newStartFlixTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/movie.mp4":
			w.Header().Set("Content-Type", "video/mp4")
			w.WriteHeader(http.StatusPartialContent)
			return
		case r.URL.Query().Has("s"):
			fmt.Fprint(w, `<div class="result-item"><article>
				<div class="image"><img src="https://image.tmdb.org/t/p/w92/x.jpg" /></div>
				<div class="details"><div class="title"><a href="https://www.startflix.test/filmes/show/">Show</a></div>
				<div class="meta"><span class="year">2024</span></div></div></article></div>`)
			return
		default:
			fmt.Fprint(w, `<button data-show-player="true" data-source="https://files.test/movie.mp4" data-type="jwplayer" data-id="1">Player #1</button>`)
		}
	})

	results, err := adapter.SearchAnime("show")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Show", results[0].Name)
	assert.Equal(t, "StartFlix", results[0].Source)

	stream, meta, err := adapter.GetStreamURL("https://painel.test/filme/tt100")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/movie.mp4", stream)
	assert.Equal(t, "startflix", meta["source"])
	assert.Equal(t, "https://painel.test/", meta["referer"])
	assert.Equal(t, "files.test", meta["host"])
}

// TestStartFlixAdapterCascade_EmptyEpisode Is the emptied-episode shape (live:
// /episodio/1 answers 200 with an empty chooser): the adapter surfaces
// ErrNoPlayers instead of handing back an empty stream.
func TestStartFlixAdapterCascade_EmptyEpisode(t *testing.T) {
	t.Parallel()

	adapter := newStartFlixTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<div class="choose-player"><h1>SELECIONE UM PLAYER</h1><div id="players"></div></div>`)
	})

	_, meta, err := adapter.GetStreamURL("https://painel.test/episodio/1")
	require.ErrorIs(t, err, startflix.ErrNoPlayers)
	assert.Nil(t, meta, "no metadata without a stream")
}
