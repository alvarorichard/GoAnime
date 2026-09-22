package playback

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
)

// Issue #203: every Bleach episode listed by Goyabu points at a Blogger token
// Google answers with RPC status 5 (NOT_FOUND) — the listing is alive, the
// media behind it is gone. The user saw only
//
//	WARN Failed to extract video URL: video unavailable on this source: …
//	     blogger video unavailable upstream: RPC status 5 (NOT_FOUND)
//
// and read it as "Goyabu simply doesn't open anything". alternateSources backs
// the replacement message, which names the sources still worth trying.

func TestAlternateSources_ExcludesTheFailingSource(t *testing.T) {
	for _, tt := range []struct {
		name    string
		current string
		want    []string
	}{
		{name: "goyabu", current: "Goyabu", want: []string{"AnimeFire", "SuperFlix"}},
		// AnimeFire's display label is not the bare kind, so the match has to
		// be prefix-based or the failing source is offered back to the user.
		{name: "animefire label", current: "Animefire.io", want: []string{"Goyabu", "SuperFlix"}},
		{name: "superflix", current: "SuperFlix", want: []string{"AnimeFire", "Goyabu"}},
		{name: "unknown source keeps them all", current: "", want: []string{"AnimeFire", "Goyabu", "SuperFlix"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, alternateSources(tt.current))
		})
	}
}

// The report must survive a nil anime and a blank episode label: it runs on an
// error path, and panicking there would replace a bad message with a crash.
func TestReportSourceVideoRemoved_HandlesMissingContext(t *testing.T) {
	assert.NotPanics(t, func() { reportSourceVideoRemoved(nil, "") })
	assert.NotPanics(t, func() {
		reportSourceVideoRemoved(&models.Anime{Name: "Bleach", Source: "Goyabu"}, "Episódio 157")
	})
}
