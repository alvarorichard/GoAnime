package source

import (
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
)

// cleanPlayback starts a test from, and leaves it at, an empty playback
// state. Publish writes process-wide state, so these tests are not parallel.
func cleanPlayback(t *testing.T) {
	t.Helper()
	reset := func() {
		util.ClearGlobalReferer()
		util.ClearGlobalOrigin()
		util.ClearGlobalSubtitles()
		util.SetGlobalAudioLanguage("")
		util.GlobalNoSubs = false
	}
	reset()
	t.Cleanup(reset)
}

func TestStreamPublishWritesEachField(t *testing.T) {
	cleanPlayback(t)

	Stream{
		URL:       "https://cdn.example/master.m3u8",
		Referer:   "https://ref.example/",
		Origin:    "https://origin.example",
		AudioLang: "por",
		Subtitles: []Subtitle{{URL: "https://s.example/a.vtt", Language: "pt", Label: "Portuguese"}},
	}.Publish()

	assert.Equal(t, "https://ref.example/", util.GetGlobalReferer())
	assert.Equal(t, "https://origin.example", util.GetGlobalOrigin())
	assert.Equal(t, "por", util.GetGlobalAudioLanguage())
	assert.Equal(t, []util.SubtitleInfo{{URL: "https://s.example/a.vtt", Language: "pt", Label: "Portuguese"}}, util.GetGlobalSubtitles(),
		"URL, Language and Label land in their own fields")
}

func TestStreamPublishLeavesWhatTheStreamDoesNotSay(t *testing.T) {
	cleanPlayback(t)
	util.SetGlobalReferer("https://kept.example/")
	util.SetGlobalOrigin("https://kept.example")
	util.SetGlobalAudioLanguage("eng")
	util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "kept", Language: "en", Label: "kept"}})

	Stream{URL: "https://cdn.example/x.mp4"}.Publish()

	assert.Equal(t, "https://kept.example/", util.GetGlobalReferer(), "an empty referer clears nothing")
	assert.Equal(t, "https://kept.example", util.GetGlobalOrigin(), "an empty origin clears nothing")
	assert.Equal(t, "eng", util.GetGlobalAudioLanguage(), "an empty audio hint clears nothing")
	assert.Len(t, util.GetGlobalSubtitles(), 1, "no subtitles in the stream clears nothing")
}

func TestStreamPublishFieldsAreIndependent(t *testing.T) {
	// Each field alone moves only its own global.
	cases := map[string]struct {
		stream Stream
		check  func(t *testing.T)
	}{
		"referer": {Stream{Referer: "r"}, func(t *testing.T) {
			assert.Equal(t, "r", util.GetGlobalReferer())
			assert.Empty(t, util.GetGlobalOrigin())
			assert.Empty(t, util.GetGlobalAudioLanguage())
			assert.Empty(t, util.GetGlobalSubtitles())
		}},
		"origin": {Stream{Origin: "o"}, func(t *testing.T) {
			assert.Equal(t, "o", util.GetGlobalOrigin())
			assert.Empty(t, util.GetGlobalReferer())
			assert.Empty(t, util.GetGlobalAudioLanguage())
			assert.Empty(t, util.GetGlobalSubtitles())
		}},
		"audio": {Stream{AudioLang: "por"}, func(t *testing.T) {
			assert.Equal(t, "por", util.GetGlobalAudioLanguage())
			assert.Empty(t, util.GetGlobalReferer())
			assert.Empty(t, util.GetGlobalOrigin())
			assert.Empty(t, util.GetGlobalSubtitles())
		}},
		"subtitles": {Stream{Subtitles: []Subtitle{{URL: "u"}}}, func(t *testing.T) {
			assert.Len(t, util.GetGlobalSubtitles(), 1)
			assert.Empty(t, util.GetGlobalReferer())
			assert.Empty(t, util.GetGlobalOrigin())
			assert.Empty(t, util.GetGlobalAudioLanguage())
		}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cleanPlayback(t)
			tc.stream.Publish()
			tc.check(t)
		})
	}
}

func TestStreamPublishHonoursNoSubs(t *testing.T) {
	cleanPlayback(t)
	util.GlobalNoSubs = true

	Stream{Subtitles: []Subtitle{{URL: "u", Language: "pt", Label: "pt"}}, Referer: "r"}.Publish()

	assert.Empty(t, util.GetGlobalSubtitles(), "--no-subs wins over the stream's tracks")
	assert.Equal(t, "r", util.GetGlobalReferer(), "the rest of the context is still published")
}

func TestStreamPublishKeepsEveryTrackInOrder(t *testing.T) {
	cleanPlayback(t)
	Stream{Subtitles: []Subtitle{{URL: "1", Language: "pt"}, {URL: "2", Language: "en"}, {URL: "3", Language: "es"}}}.Publish()

	got := util.GetGlobalSubtitles()
	assert.Equal(t, []string{"1", "2", "3"}, []string{got[0].URL, got[1].URL, got[2].URL})
}

func TestStreamPublishCopiesTheSubtitles(t *testing.T) {
	cleanPlayback(t)
	subs := []Subtitle{{URL: "a"}}
	Stream{Subtitles: subs}.Publish()

	subs[0].URL = "changed after publishing"

	assert.Equal(t, "a", util.GetGlobalSubtitles()[0].URL, "the player's tracks do not alias the source's slice")
}
