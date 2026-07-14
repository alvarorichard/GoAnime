package api

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/providers/superflix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sfServer(id string, typ int, name string, isFile bool) superflix.SuperFlixServer {
	raw, _ := json.Marshal(id)
	return superflix.SuperFlixServer{ID: raw, Type: typ, Name: name, IsFile: isFile}
}

const (
	dub = superflix.SuperFlixAudioDubbed
	leg = superflix.SuperFlixAudioSubtitled
)

// promptRecord captures one prompt the flow showed the user.
type promptRecord struct {
	prompt string
	labels []string
}

// stubSFPicker installs the single selection seam and records every prompt.
// choose(prompt, labels) returns the selected index. It resets the per-title
// memory so each test starts clean. Not parallel — it swaps package state.
func stubSFPicker(t *testing.T, choose func(prompt string, labels []string) (int, error)) *[]promptRecord {
	t.Helper()
	var seen []promptRecord
	prev := sfPickFn
	sfPickFn = func(prompt string, labels []string) (int, error) {
		seen = append(seen, promptRecord{prompt: prompt, labels: labels})
		return choose(prompt, labels)
	}
	resetSuperFlixServerPrefs()
	resetSuperFlixAudioChoices()
	t.Cleanup(func() {
		sfPickFn = prev
		resetSuperFlixServerPrefs()
		resetSuperFlixAudioChoices()
	})
	return &seen
}

// alwaysFirst answers every prompt with option 0.
func alwaysFirst(string, []string) (int, error) { return 0, nil }

// isAudioPrompt / isServerPrompt classify a prompt so a cascade test can answer
// each step distinctly.
func isAudioPrompt(p string) bool  { return strings.Contains(p, "dublado ou legendado") }
func isServerPrompt(p string) bool { return strings.Contains(p, "servidor") }

// -----------------------------------------------------------------------------
// Pure helpers
// -----------------------------------------------------------------------------

func TestOrderedServers(t *testing.T) {
	t.Parallel()
	// Declared out of order on purpose: subtitled first, file before streaming.
	in := []superflix.SuperFlixServer{
		sfServer("s", leg, "sub-stream", false),
		sfServer("df", dub, "dub-file", true),
		sfServer("ds", dub, "dub-stream", false),
	}
	got := orderedServers(in)
	// Dubbed first; within dubbed, streaming before file; subtitled last.
	assert.Equal(t, []string{"dub-stream", "dub-file", "sub-stream"},
		[]string{got[0].Name, got[1].Name, got[2].Name})
	// Input must be left untouched.
	assert.Equal(t, "sub-stream", in[0].Name)
}

func TestDistinctAudioKinds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   []superflix.SuperFlixServer
		want []int
	}{
		{"dubbed only", []superflix.SuperFlixServer{sfServer("1", dub, "a", false), sfServer("2", dub, "b", false)}, []int{dub}},
		{"both, subtitled declared first, dubbed ranks first", []superflix.SuperFlixServer{sfServer("1", leg, "a", false), sfServer("2", dub, "b", false)}, []int{dub, leg}},
		{"unknown sorts last", []superflix.SuperFlixServer{sfServer("1", 0, "a", false), sfServer("2", dub, "b", false)}, []int{dub, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, distinctAudioKinds(tt.in))
		})
	}
}

func TestServerRowLabel(t *testing.T) {
	t.Parallel()
	// Opaque per-episode ids are hidden; the user sees a position number and the
	// one distinction that matters — direct MP4 vs streaming.
	assert.Equal(t, "▶  Servidor 1", serverRowLabel(sfServer("159462", dub, "Servidor 159462", false), 1))
	assert.Equal(t, "⬇  Servidor 2  ·  MP4 (download direto)", serverRowLabel(sfServer("native_media:1", dub, "MP4 Dublado", true), 2))
}

func TestAudioKindLabels(t *testing.T) {
	t.Parallel()
	assert.Contains(t, audioKindLabel(dub), "Dublado")
	assert.Contains(t, audioKindLabel(dub), "português")
	assert.Contains(t, audioKindLabel(leg), "Legendado")
	assert.Contains(t, audioKindLabel(leg), "original")
	assert.Equal(t, "Dublado", audioKindName(dub))
	assert.Equal(t, "Legendado", audioKindName(leg))
}

func TestNarrowByMemory(t *testing.T) {
	t.Parallel()
	servers := []superflix.SuperFlixServer{
		sfServer("1", dub, "dub-stream", false),
		sfServer("2", dub, "dub-file", true),
		sfServer("3", leg, "sub-stream", false),
	}

	t.Run("exact match on audio + file flavor", func(t *testing.T) {
		t.Parallel()
		got := narrowByMemory(servers, sfServerPref{Type: dub, IsFile: true})
		require.Len(t, got, 1)
		assert.Equal(t, "dub-file", got[0].Name)
	})

	t.Run("relaxes to the audio when the flavor is gone", func(t *testing.T) {
		t.Parallel()
		// Remembered a subtitled FILE, but this episode has only a subtitled stream.
		got := narrowByMemory(servers, sfServerPref{Type: leg, IsFile: true})
		require.Len(t, got, 1)
		assert.Equal(t, "sub-stream", got[0].Name)
	})

	t.Run("stale memory falls back to the whole list", func(t *testing.T) {
		t.Parallel()
		// Remembered an audio type that is not offered this episode at all.
		got := narrowByMemory(servers, sfServerPref{Type: 99})
		assert.Len(t, got, 3)
	})
}

// -----------------------------------------------------------------------------
// selectSuperFlixServer — cascade (two-step) behavior
// -----------------------------------------------------------------------------

// Tehran S1E1: two servers, both dubbed. There is no dub/sub decision to make, so
// the audio step must be SKIPPED and only the server step shown — with numbered,
// human labels (not the opaque site ids).
func TestSelectSuperFlixServer_DubbedOnly_SkipsAudioStep(t *testing.T) {
	seen := stubSFPicker(t, alwaysFirst)

	servers := []superflix.SuperFlixServer{
		sfServer("159462", dub, "Servidor 159462", false),
		sfServer("native_media:233831", dub, "MP4 Dublado", true),
	}
	got, err := selectSuperFlixServer("103913", servers)
	require.NoError(t, err)
	assert.Equal(t, dub, got.Type)

	require.Len(t, *seen, 1, "only the server step should run for a dubbed-only title")
	assert.True(t, isServerPrompt((*seen)[0].prompt))
	assert.Contains(t, (*seen)[0].prompt, "Dublado", "the server prompt should say which audio it is for")
	assert.Equal(t, []string{"▶  Servidor 1", "⬇  Servidor 2  ·  MP4 (download direto)"}, (*seen)[0].labels)
}

// One dubbed + one subtitled: the audio step is the only real decision, so the
// server step must be skipped.
func TestSelectSuperFlixServer_OnePerKind_OnlyAudioStep(t *testing.T) {
	seen := stubSFPicker(t, func(prompt string, _ []string) (int, error) {
		if isAudioPrompt(prompt) {
			return 1, nil // Legendado
		}
		t.Errorf("unexpected second prompt: %q", prompt)
		return 0, nil
	})

	servers := []superflix.SuperFlixServer{
		sfServer("1", dub, "Servidor 1", false),
		sfServer("2", leg, "Servidor 2", false),
	}
	got, err := selectSuperFlixServer("t", servers)
	require.NoError(t, err)

	assert.Equal(t, leg, got.Type, "picking Legendado must yield the subtitled server")
	require.Len(t, *seen, 1)
	assert.True(t, isAudioPrompt((*seen)[0].prompt))
	assert.Equal(t, []string{audioKindLabel(dub), audioKindLabel(leg)}, (*seen)[0].labels)
}

// Two dubbed + two subtitled: the full cascade — audio THEN server — and the two
// answers must compose to the exactly-right source.
func TestSelectSuperFlixServer_FullCascade(t *testing.T) {
	seen := stubSFPicker(t, func(prompt string, _ []string) (int, error) {
		switch {
		case isAudioPrompt(prompt):
			return 1, nil // Legendado
		case isServerPrompt(prompt):
			return 1, nil // second subtitled server
		}
		return 0, nil
	})

	servers := []superflix.SuperFlixServer{
		sfServer("d1", dub, "dub-stream", false),
		sfServer("d2", dub, "dub-file", true),
		sfServer("s1", leg, "sub-stream", false),
		sfServer("s2", leg, "sub-file", true),
	}
	got, err := selectSuperFlixServer("t", servers)
	require.NoError(t, err)

	assert.Equal(t, "s2", got.IDString(), "Legendado then second server must land on the second subtitled source")
	require.Len(t, *seen, 2, "two real decisions → two prompts")
	assert.True(t, isAudioPrompt((*seen)[0].prompt), "audio step must come first")
	assert.True(t, isServerPrompt((*seen)[1].prompt), "server step must come second")
	// The server step must be scoped to the chosen audio (only the two subtitled ones).
	assert.Len(t, (*seen)[1].labels, 2)
	assert.Contains(t, (*seen)[1].prompt, "Legendado")
}

func TestSelectSuperFlixServer_SingleServer_NoPrompt(t *testing.T) {
	seen := stubSFPicker(t, func(prompt string, _ []string) (int, error) {
		t.Errorf("must not prompt for a single server: %q", prompt)
		return 0, nil
	})

	got, err := selectSuperFlixServer("t", []superflix.SuperFlixServer{sfServer("1", dub, "Único", false)})
	require.NoError(t, err)
	assert.Equal(t, "Único", got.Name)
	assert.Empty(t, *seen)
}

// A binge is asked once. The first episode drives both prompts; later episodes —
// whose server ids/names change — resolve silently from memory.
func TestSelectSuperFlixServer_AsksOncePerTitle(t *testing.T) {
	seen := stubSFPicker(t, func(prompt string, _ []string) (int, error) {
		if isServerPrompt(prompt) {
			return 1, nil // the MP4
		}
		return 0, nil
	})

	ep1 := []superflix.SuperFlixServer{
		sfServer("159462", dub, "Servidor 159462", false),
		sfServer("native_media:233831", dub, "MP4 Dublado", true),
	}
	got, err := selectSuperFlixServer("103913", ep1)
	require.NoError(t, err)
	assert.True(t, got.IsFile, "picked the MP4")

	// Next episode: ids and names differ, but the (dubbed, file) preference holds.
	ep2 := []superflix.SuperFlixServer{
		sfServer("777888", dub, "Servidor 777888", false),
		sfServer("native_media:999", dub, "MP4 Dublado", true),
	}
	got2, err := selectSuperFlixServer("103913", ep2)
	require.NoError(t, err)
	assert.Equal(t, "native_media:999", got2.IDString())
	assert.Len(t, *seen, 1, "a binge must not re-prompt on every episode")
}

// If the exact flavor is gone next episode, the audio preference must still hold
// rather than silently flipping the viewer to the other audio.
func TestSelectSuperFlixServer_RelaxesToAudioWhenFlavorGone(t *testing.T) {
	seen := stubSFPicker(t, func(prompt string, _ []string) (int, error) {
		if isServerPrompt(prompt) {
			return 1, nil // the dubbed MP4
		}
		return 0, nil
	})

	ep1 := []superflix.SuperFlixServer{
		sfServer("1", dub, "dub-stream", false),
		sfServer("native_media:2", dub, "dub-file", true),
	}
	_, err := selectSuperFlixServer("t", ep1)
	require.NoError(t, err)

	// No MP4 this episode, but still a dubbed stream and a subtitled one.
	ep2 := []superflix.SuperFlixServer{
		sfServer("3", dub, "dub-stream-2", false),
		sfServer("4", leg, "sub", false),
	}
	got, err := selectSuperFlixServer("t", ep2)
	require.NoError(t, err)
	assert.Equal(t, dub, got.Type, "must stay dubbed, not fall onto the subtitled server")
	assert.Len(t, *seen, 1, "the relaxed match is unambiguous, so no re-prompt")
}

// Without a TTY the picker errors. Playback must go on, defaulting to the first
// option — a dubbed streaming server, the safe PT-BR default.
func TestSelectSuperFlixServer_FallsBackWhenPickerUnavailable(t *testing.T) {
	stubSFPicker(t, func(string, []string) (int, error) { return 0, errors.New("no tty") })

	got, err := selectSuperFlixServer("t", []superflix.SuperFlixServer{
		sfServer("9", leg, "sub", false),
		sfServer("native_media:1", dub, "dub-file", true),
		sfServer("2", dub, "dub-stream", false),
	})
	require.NoError(t, err)
	assert.Equal(t, dub, got.Type, "the safe default is the dub")
	assert.False(t, got.IsFile, "and a plain streaming server, not the MP4 file")
}

func TestSelectSuperFlixServer_EmptyIsError(t *testing.T) {
	stubSFPicker(t, alwaysFirst)
	_, err := selectSuperFlixServer("t", nil)
	require.Error(t, err)
}

// -----------------------------------------------------------------------------
// audioForServer
// -----------------------------------------------------------------------------

func TestAudioForServer(t *testing.T) {
	t.Parallel()
	tracks := []string{"por", "und", "eng", "spa", "kor", "jpn", "chi", "und"}

	t.Run("dubbed server plays Portuguese", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "por,pob,pt-BR,ptbr,pt,portuguese",
			audioForServer(sfServer("1", dub, "S", false), tracks))
	})

	// A legendado server IS the original-audio source; guessing "first
	// non-Portuguese track" hands Tehran's viewer English (original is Hebrew, the
	// "und" track we drop), so we force nothing.
	t.Run("subtitled server keeps the stream's own audio", func(t *testing.T) {
		t.Parallel()
		assert.Empty(t, audioForServer(sfServer("2", leg, "S", false), tracks))
	})

	t.Run("dubbed server with no Portuguese track leaves mpv alone", func(t *testing.T) {
		t.Parallel()
		assert.Empty(t, audioForServer(sfServer("3", dub, "S", false), []string{"jpn", "eng"}))
	})

	// It must never suppress subtitles — an earlier version did, breaking them.
	t.Run("says nothing about subtitles", func(t *testing.T) {
		t.Parallel()
		assert.Empty(t, audioForServer(sfServer("4", leg, "S", false), nil))
	})
}

// After a server-list play, a later CACHED play of the same title yields no server
// (chosen == nil) and so falls to the audio-from-tracks picker. Without recording
// the server's audio, that picker would prompt again on the re-watch. This pins
// that the choice is remembered and replayed silently.
func TestRememberServerAudioChoice_CachedRepeatDoesNotReprompt(t *testing.T) {
	tracks := []string{"por", "jpn"}

	t.Run("dubbed server → cached repeat replays Portuguese, no prompt", func(t *testing.T) {
		stubSFPicker(t, func(prompt string, _ []string) (int, error) {
			t.Errorf("a remembered title must not re-prompt: %q", prompt)
			return 0, nil
		})
		rememberServerAudioChoice("t", sfServer("1", dub, "S", false))

		opt, ok := selectSuperFlixAudio("t", tracks, true)
		require.True(t, ok)
		assert.True(t, opt.Dubbed)
		assert.Equal(t, "por,pob,pt-BR,ptbr,pt,portuguese", mpvAudioLanguage(opt))
	})

	t.Run("subtitled server → cached repeat keeps the stream's own audio, no prompt", func(t *testing.T) {
		stubSFPicker(t, func(prompt string, _ []string) (int, error) {
			t.Errorf("a remembered title must not re-prompt: %q", prompt)
			return 0, nil
		})
		rememberServerAudioChoice("t", sfServer("2", leg, "S", false))

		opt, ok := selectSuperFlixAudio("t", tracks, true)
		require.True(t, ok)
		assert.False(t, opt.Dubbed)
		assert.Empty(t, mpvAudioLanguage(opt), "a subtitled choice forces no language")
	})
}
