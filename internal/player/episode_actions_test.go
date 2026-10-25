package player

import (
	"os"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/term"
)

func TestEpisodeMenuRequest(t *testing.T) {
	ClearEpisodeMenu()
	_, ok := TakeEpisodeMenu()
	assert.False(t, ok, "nothing pending")

	RequestEpisodeMenu(EpisodeMenuRequest{ShowMenu: true})
	got, ok := TakeEpisodeMenu()
	require.True(t, ok)
	assert.True(t, got.ShowMenu)

	_, ok = TakeEpisodeMenu()
	assert.False(t, ok, "consumed once")

	RequestEpisodeMenu(EpisodeMenuRequest{DownloadCode: 2, RangeStart: 5})
	ClearEpisodeMenu()
	_, ok = TakeEpisodeMenu()
	assert.False(t, ok, "cleared before use")
}

func TestEpisodeListContext(t *testing.T) {
	SetEpisodeListContext("Mukashi Banashi", "Animefire.io", "Legendado")
	ctx := snapshotEpisodeListContext()
	assert.Equal(t, "Mukashi Banashi", ctx.title)
	assert.Equal(t, "Animefire.io", ctx.source)
	assert.Equal(t, "Legendado", ctx.audio)
	SetEpisodeListContext("", "", "")

	assert.Equal(t, "AnimeFire", displayListSource("Animefire.io"))
	assert.Equal(t, "AnimeFire", displayListSource("AnimeFire"))
	assert.Equal(t, "HiAnime", displayListSource("HiAnime"))
	assert.Equal(t, "", displayListSource(""))
}

func TestSelectEpisodeActions(t *testing.T) {
	eps := []models.Episode{
		{Number: "1", Num: 1, URL: "http://ep/1", Title: models.TitleDetails{Romaji: "Start"}},
		{Number: "2", Num: 2, URL: "http://ep/2"},
	}
	watched := map[int]episodeWatch{0: {position: 372, duration: 580}}

	pickAction := func(action string, idx int) episodePickFunc {
		return func([]tui.PickItem, tui.PickOptions) (int, string, error) {
			return idx, action, nil
		}
	}

	t.Run("r forces from start", func(t *testing.T) {
		ClearResumeOverride()
		ClearEpisodeMenu()
		url, num, err := selectEpisodeWithPicker(pickAction("r", 0), eps, watched)
		require.NoError(t, err)
		assert.Equal(t, "http://ep/1", url)
		assert.Equal(t, "1", num)
		position, fromStart, ok := takeResumeOverride()
		require.True(t, ok)
		assert.True(t, fromStart)
		assert.Equal(t, 0, position)
	})

	t.Run("m requests the menu", func(t *testing.T) {
		ClearEpisodeMenu()
		_, _, err := selectEpisodeWithPicker(pickAction("m", 1), eps, watched)
		require.NoError(t, err)
		req, ok := TakeEpisodeMenu()
		require.True(t, ok)
		assert.True(t, req.ShowMenu)
		assert.Equal(t, 0, req.DownloadCode)
	})

	t.Run("d auto-confirms single", func(t *testing.T) {
		ClearEpisodeMenu()
		_, _, err := selectEpisodeWithPicker(pickAction("d", 1), eps, watched)
		require.NoError(t, err)
		req, ok := TakeEpisodeMenu()
		require.True(t, ok)
		assert.Equal(t, 1, req.DownloadCode)
	})

	t.Run("D presets the range start", func(t *testing.T) {
		ClearEpisodeMenu()
		_, _, err := selectEpisodeWithPicker(pickAction("D", 1), eps, watched)
		require.NoError(t, err)
		req, ok := TakeEpisodeMenu()
		require.True(t, ok)
		assert.Equal(t, 2, req.DownloadCode)
		assert.Equal(t, 2, req.RangeStart)
	})

	t.Run("A auto-confirms all", func(t *testing.T) {
		ClearEpisodeMenu()
		_, _, err := selectEpisodeWithPicker(pickAction("A", 0), eps, watched)
		require.NoError(t, err)
		req, ok := TakeEpisodeMenu()
		require.True(t, ok)
		assert.Equal(t, 5, req.DownloadCode)
	})

	t.Run("unknown action errors", func(t *testing.T) {
		ClearEpisodeMenu()
		_, _, err := selectEpisodeWithPicker(pickAction("z", 0), eps, watched)
		require.Error(t, err)
		_, ok := TakeEpisodeMenu()
		assert.False(t, ok, "nothing requested on error")
	})

	t.Run("action with bad index errors", func(t *testing.T) {
		_, _, err := selectEpisodeWithPicker(pickAction("d", 9), eps, watched)
		require.Error(t, err)
	})

	t.Run("enter resumes a half-watched episode", func(t *testing.T) {
		ClearResumeOverride()
		ClearEpisodeMenu()
		_, _, err := selectEpisodeWithPicker(func([]tui.PickItem, tui.PickOptions) (int, string, error) {
			return 0, "", nil
		}, eps, watched)
		require.NoError(t, err)
		position, fromStart, ok := takeResumeOverride()
		require.True(t, ok, "enter on ◐ resumes without asking")
		assert.False(t, fromStart)
		assert.Equal(t, 372, position)
	})

	t.Run("enter on unwatched asks normally", func(t *testing.T) {
		ClearResumeOverride()
		_, _, err := selectEpisodeWithPicker(func([]tui.PickItem, tui.PickOptions) (int, string, error) {
			return 1, "", nil
		}, eps, watched)
		require.NoError(t, err)
		_, _, ok := takeResumeOverride()
		assert.False(t, ok, "no position to force")
	})
}

func TestEpisodePickOptionsContract(t *testing.T) {
	// Not parallel: episodeListContext is process-wide, and the picker
	// contract test below reads it. Parallel tests would race on it.
	SetEpisodeListContext("Mukashi Banashi", "Animefire.io", "Legendado")
	t.Cleanup(func() { SetEpisodeListContext("", "", "") })

	eps := []models.Episode{{Number: "1", URL: "u1"}, {Number: "2", URL: "u2"}}
	watched := map[int]episodeWatch{0: {position: 590, duration: 600}}
	opts := episodePickOptions(eps, watched)

	assert.Equal(t, "Search > Mukashi Banashi > Episodes", opts.Breadcrumb)
	assert.Equal(t, "AnimeFire · Legendado", opts.Context)
	assert.Equal(t, "resume", opts.Action)
	assert.Equal(t, "results", opts.BackDesc)
	assert.True(t, opts.HelpHint)
	assert.Equal(t, "type a number to jump, / to search titles", opts.FilterPlaceholder)
	assert.Equal(t, "episodes · 1 watched", opts.ItemPlural)
	require.Len(t, opts.Actions, 5)
	assert.Equal(t, "r", opts.Actions[0].Key)
	assert.Equal(t, "all 2 episodes", opts.Actions[3].Label)
	assert.Equal(t, "Download", opts.ActionsLabel)
}

func TestDownloadMenuOption(t *testing.T) {

	t.Run("plays directly by default", func(t *testing.T) {
		ClearEpisodeMenu()
		assert.Equal(t, 4, downloadMenuOption(5, 12, "Show"))
	})

	t.Run("auto-confirms episodes-screen downloads", func(t *testing.T) {
		ClearEpisodeMenu()
		RequestEpisodeMenu(EpisodeMenuRequest{DownloadCode: 5})
		assert.Equal(t, 5, downloadMenuOption(5, 12, "Show"))
		_, ok := TakeEpisodeMenu()
		assert.False(t, ok, "consumed")
	})

	t.Run("menu opens without TTY as play", func(t *testing.T) {
		ClearEpisodeMenu()
		RequestEpisodeMenu(EpisodeMenuRequest{ShowMenu: true})
		// No TTY in tests: askForDownload falls back to play online.
		assert.Equal(t, 4, downloadMenuOption(5, 12, "Show"))
	})
}

func TestPromptRangeEndNoTTY(t *testing.T) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		t.Skip("needs a non-interactive stdin")
	}
	_, err := promptRangeEnd(5)
	require.Error(t, err)
}
