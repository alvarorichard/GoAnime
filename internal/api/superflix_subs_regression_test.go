package api

import (
	"context"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/superflix"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Subtitles that the stream ships must ALWAYS reach mpv.
//
// A previous version withheld them whenever the Portuguese dub was selected, on
// the theory that Portuguese subtitles over Portuguese audio only echo the
// dialogue. Nobody asked for that, and it broke real viewing — subtitles that had
// always been there stopped appearing. The flag also defaulted to "off", so a
// stream that exposed no audio-track list, or a user who had pinned --audio-lang,
// lost its subtitles silently too.
//
// This drives the real GetSuperFlixStreamURL over the seams and checks what
// actually lands in util.GlobalSubtitles.
func TestGetSuperFlixStreamURL_AlwaysLoadsSubtitles(t *testing.T) {
	withSubtitledStream := func(t *testing.T) {
		t.Helper()
		pl, pn := sfGetServersFn, sfSniffStreamFn
		t.Cleanup(func() { sfGetServersFn, sfSniffStreamFn = pl, pn })

		// No server list: exercise the fallback, where the audio prompt runs.
		sfGetServersFn = func(*superflix.SuperFlixClient, context.Context, string, string, string, string) ([]superflix.SuperFlixServer, *superflix.SuperFlixTokens, error) {
			return nil, nil, superflix.ErrSuperFlixRateLimited
		}
		sfSniffStreamFn = func(*superflix.SuperFlixClient, context.Context, string, string, string, string) (*superflix.SuperFlixStreamResult, error) {
			return &superflix.SuperFlixStreamResult{
				StreamURL:    "https://cdn/x.m3u8",
				DefaultAudio: []string{"por", "jpn"},
				Subtitles:    []superflix.SuperFlixSubtitle{{Lang: "Portuguese", URL: "https://subs/pt.vtt"}},
			}, nil
		}
	}

	media := &models.Anime{Name: "X", URL: "1", Source: "SuperFlix", MediaType: models.MediaTypeTV}
	ep := &models.Episode{Number: "1", Num: 1, URL: "1", SeasonID: "1"}

	t.Run("dubbed audio still gets the subtitles", func(t *testing.T) {
		withSubtitledStream(t)
		stubAudioPicker(t, func([]string) (int, error) { return 0, nil }) // Português (Dublado)
		util.ClearGlobalSubtitles()
		util.GlobalAudioLanguage = ""
		resetSuperFlixAudioChoices()

		_, err := GetSuperFlixStreamURL(media, ep, "best")
		require.NoError(t, err)

		require.Len(t, util.GlobalSubtitles, 1, "picking the dub must not throw the subtitles away")
		assert.Equal(t, "https://subs/pt.vtt", util.GlobalSubtitles[0].URL)
		assert.Contains(t, util.GlobalAudioLanguage, "por")
	})

	t.Run("subtitles survive a pinned --audio-lang", func(t *testing.T) {
		withSubtitledStream(t)
		stubAudioPicker(t, func([]string) (int, error) {
			t.Error("a pinned --audio-lang must not prompt")
			return 0, nil
		})
		util.ClearGlobalSubtitles()
		prev := util.GlobalAudioLanguage
		util.GlobalAudioLanguage = "jpn"
		t.Cleanup(func() { util.GlobalAudioLanguage = prev })
		resetSuperFlixAudioChoices()

		_, err := GetSuperFlixStreamURL(media, ep, "best")
		require.NoError(t, err)

		require.Len(t, util.GlobalSubtitles, 1, "no audio decision must not mean no subtitles")
	})

	t.Run("--no-subs is still honored", func(t *testing.T) {
		withSubtitledStream(t)
		stubAudioPicker(t, func([]string) (int, error) { return 0, nil })
		util.ClearGlobalSubtitles()
		util.GlobalAudioLanguage = ""
		resetSuperFlixAudioChoices()
		util.GlobalNoSubs = true
		t.Cleanup(func() { util.GlobalNoSubs = false })

		_, err := GetSuperFlixStreamURL(media, ep, "best")
		require.NoError(t, err)
		assert.Empty(t, util.GlobalSubtitles, "--no-subs remains the way to opt out")
	})
}
