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
		// Goyabu and SuperFlix ship off by default, so they are never offered
		// unless opted in: the user would not find them in the results.
		{name: "goyabu", current: "Goyabu", want: []string{"AnimeFire", "StartFlix"}},
		// AnimeFire's display label is not the bare kind, so the match has to
		// be prefix-based or the failing source is offered back to the user.
		{name: "animefire label", current: "Animefire.io", want: []string{"StartFlix"}},
		{name: "superflix", current: "SuperFlix", want: []string{"AnimeFire", "StartFlix"}},
		{name: "startflix", current: "StartFlix", want: []string{"AnimeFire"}},
		// TopCine plays through StartFlix's panel: neither rescues the other.
		{name: "topcine", current: "TopCine", want: []string{"AnimeFire"}},
		{name: "unknown source keeps every searched one", current: "", want: []string{"AnimeFire", "StartFlix"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, alternateSources(tt.current))
		})
	}
}

func TestAlternateSources_FollowsOptInAndKillSwitch(t *testing.T) {
	t.Setenv("GOANIME_ENABLED_SOURCES", "goyabu")
	assert.Equal(t, []string{"AnimeFire", "Goyabu"}, alternateSources("StartFlix"),
		"an opted-in source is offered again")

	t.Setenv("GOANIME_DISABLED_SOURCES", "animefire")
	assert.Equal(t, []string{"Goyabu"}, alternateSources("StartFlix"),
		"a killed source is never offered")
}

// The report must survive a nil anime and a blank episode label: it runs on an
// error path, and panicking there would replace a bad message with a crash.
func TestReportSourceVideoRemoved_HandlesMissingContext(t *testing.T) {
	assert.NotPanics(t, func() { reportSourceVideoRemoved(nil, "") })
	assert.NotPanics(t, func() {
		reportSourceVideoRemoved(&models.Anime{Name: "Bleach", Source: "Goyabu"}, "Episódio 157")
	})
}
