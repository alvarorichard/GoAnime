package player

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Subtitles resolved by a source have to reach mpv no matter which source that
// was.
//
// The old gate was a list of source names — movies/TV, SuperFlix, 9Anime — so
// any other source resolved its tracks, stored them in the globals, and had
// them dropped on the way to the command line. Nothing said so: playback
// started fine, just without subtitles. HiAnime ships an English and a
// Brazilian Portuguese track per episode and hit exactly that.
//
// These tests are not parallel: they mutate the process-wide subtitle globals.

func TestExternalSubtitleArgs_PassesTracksFromAnySource(t *testing.T) {
	t.Cleanup(util.ClearGlobalSubtitles)
	util.SetGlobalSubtitles([]util.SubtitleInfo{
		{URL: "https://cdn.example/pt.vtt", Label: "Portuguese (- Brazilian)"},
	})

	// is9Anime=false and a single track: no picker is reached, so this exercises
	// the gate itself rather than the prompt.
	got := externalSubtitleArgs(false)

	require.Len(t, got, 1, "a track resolved by a non-SuperFlix, non-9Anime source must still reach mpv")
	assert.Equal(t, "--sub-file=https://cdn.example/pt.vtt", got[0])
}

func TestExternalSubtitleArgs_NoTracksMeansNoArgs(t *testing.T) {
	t.Cleanup(util.ClearGlobalSubtitles)
	util.ClearGlobalSubtitles()

	assert.Empty(t, externalSubtitleArgs(false),
		"a source that resolved no subtitles must not add empty arguments")
}

func TestExternalSubtitleArgs_RespectsNoSubs(t *testing.T) {
	t.Cleanup(func() {
		util.ClearGlobalSubtitles()
		util.GlobalNoSubs = false
	})
	util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "https://cdn.example/en.vtt", Label: "English"}})
	util.GlobalNoSubs = true

	assert.Empty(t, externalSubtitleArgs(false), "--no-subs must still win")
}
