package player

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlayerMenuOptions(t *testing.T) {
	t.Parallel()

	values := func(options []playerMenuOption) []string {
		out := make([]string, len(options))
		for i, o := range options {
			out[i] = o.Value
		}
		return out
	}

	series := playerMenuOptions(false, false, false)
	assert.Equal(t, []string{"next", "previous", "select", "change", "skip", "download_options", "quit"}, values(series))

	movie := playerMenuOptions(true, false, false)
	assert.Equal(t, []string{"next", "change", "download_options", "quit"}, values(movie))
	assert.Equal(t, "Replay movie", movie[0].Label)

	withTracks := playerMenuOptions(false, true, true)
	assert.Equal(t, []string{"next", "previous", "select", "change", "skip", "audio", "subtitle", "download_options", "quit"}, values(withTracks))

	// Every value is one handleUserInput acts on, and no two entries share
	// a key.
	handled := map[string]bool{"download_options": true, "next": true, "previous": true, "quit": true,
		"change": true, "select": true, "audio": true, "subtitle": true, "skip": true}
	for _, options := range [][]playerMenuOption{series, movie, withTracks} {
		keys := map[string]bool{}
		for _, o := range options {
			assert.True(t, handled[o.Value], o.Value)
			require.NotEmpty(t, o.Key, o.Label)
			assert.False(t, keys[o.Key], "key %q used twice", o.Key)
			keys[o.Key] = true
		}
		assert.Equal(t, "esc", options[len(options)-2].Key, "back is esc, which the picker already treats as back")
	}
}

func TestPlayerTrackChoices_NoSocket(t *testing.T) {
	t.Parallel()
	audio, subtitles := playerTrackChoices("")
	assert.False(t, audio)
	assert.False(t, subtitles)
}

func TestNowPlayingPanel(t *testing.T) {
	t.Parallel()

	panel := nowPlayingPanel("/tmp/mpv.sock", "Momotarou", 1, false)
	assert.Equal(t, "Now playing", panel.Label)
	assert.Equal(t, "Momotarou", panel.Title)
	assert.Equal(t, []string{"Episode 1 · 第1話"}, panel.Lines)
	assert.True(t, panel.Live)
	assert.NotNil(t, panel.Progress)

	movie := nowPlayingPanel("", "", 0, true)
	assert.Equal(t, "mpv", movie.Title)
	assert.Empty(t, movie.Lines)
	assert.Nil(t, movie.Progress, "no socket, nothing to ask")
	assert.False(t, movie.Live)
}

func TestTrackPickItem(t *testing.T) {
	t.Parallel()
	got := trackPickItem(2, []string{"jpn", "aac", "2ch"}, true)
	assert.Equal(t, tui.PickItem{Label: "jpn", Details: "Track 2 · aac · 2ch", Status: tui.PickCurrent}, got)

	got = trackPickItem(5, nil, false)
	assert.Equal(t, tui.PickItem{Label: "Track 5"}, got)
}

func TestQualityPickItems(t *testing.T) {
	t.Parallel()
	got := qualityPickItems([]string{"1080p", "720p (mirror 2)", "Auto"}, []int{1080, 720, 0})
	require.Len(t, got, 3)
	assert.InDelta(t, 1.0, got[0].Meter, 1e-9)
	assert.InDelta(t, 720.0/1080, got[1].Meter, 1e-9)
	assert.Equal(t, "720p (mirror 2)", got[1].Label)
	assert.Zero(t, got[2].Meter)
	assert.Equal(t, "resolution not reported", got[2].Details)

	none := qualityPickItems([]string{"Auto"}, nil)
	assert.Equal(t, "resolution not reported", none[0].Details)
}

func TestQualityFlagPanel(t *testing.T) {
	t.Parallel()
	panel := qualityFlagPanel()
	assert.Equal(t, "goanime --quality 720p", panel.Command)
	assert.Nil(t, panel.Progress)
}
