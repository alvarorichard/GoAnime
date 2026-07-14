package api

import (
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuperFlixAudioOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		codes      []string
		wantCodes  []string
		wantLabels []string
	}{
		{
			// The real track list SuperFlix declares for Dexter S1E1.
			name:      "real stream: und dropped, dedup, Portuguese first",
			codes:     []string{"por", "und", "eng", "spa", "kor", "jpn", "chi", "und"},
			wantCodes: []string{"por", "eng", "spa", "kor", "jpn", "chi"},
			wantLabels: []string{
				"🎙️  Português (Dublado)",
				"💬 Inglês (Legendado)",
				"💬 Espanhol (Legendado)",
				"💬 Coreano (Legendado)",
				"💬 Japonês (Legendado)",
				"💬 Chinês (Legendado)",
			},
		},
		{
			name:       "Portuguese is promoted even when declared last",
			codes:      []string{"jpn", "eng", "por"},
			wantCodes:  []string{"por", "jpn", "eng"},
			wantLabels: []string{"🎙️  Português (Dublado)", "💬 Japonês (Legendado)", "💬 Inglês (Legendado)"},
		},
		{
			name:       "single track",
			codes:      []string{"jpn"},
			wantCodes:  []string{"jpn"},
			wantLabels: []string{"💬 Japonês (Legendado)"},
		},
		{
			// An unknown code is shown, not hidden: a track the user can still pick
			// beats a track we silently drop.
			name:       "unknown code survives, uppercased",
			codes:      []string{"por", "xyz"},
			wantCodes:  []string{"por", "xyz"},
			wantLabels: []string{"🎙️  Português (Dublado)", "💬 XYZ (Legendado)"},
		},
		{"only und is no choice at all", []string{"und", "und"}, nil, nil},
		{"empty", nil, nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := superFlixAudioOptions(tt.codes, true)

			codes := make([]string, 0, len(got))
			labels := make([]string, 0, len(got))
			for _, o := range got {
				codes = append(codes, o.Code)
				labels = append(labels, o.Label)
			}
			if tt.wantCodes == nil {
				assert.Empty(t, got)
				return
			}
			assert.Equal(t, tt.wantCodes, codes)
			assert.Equal(t, tt.wantLabels, labels)

			// Exactly the Portuguese track may be flagged as the dub.
			for _, o := range got {
				assert.Equal(t, portugueseAudioCodes[o.Code], o.Dubbed, "Dubbed flag for %q", o.Code)
			}
		})
	}
}

// "Legendado" promises subtitles, and SuperFlix does not always ship them.
// Mushoku Tensei's stream carries the SAME seven audio tracks as Dexter's but no
// subtitle file at all (verified live), so calling its Japanese track "Legendado"
// would sell the viewer an episode they cannot follow. The label must tell the
// truth instead.
func TestSuperFlixAudioOptions_NoSubtitlesIsNotLegendado(t *testing.T) {
	t.Parallel()

	got := superFlixAudioOptions([]string{"por", "jpn"}, false)
	require.Len(t, got, 2)

	assert.Equal(t, "🎙️  Português (Dublado)", got[0].Label, "the dub is unaffected — it needs no subtitles")
	assert.Equal(t, "🔊 Japonês (áudio original — sem legendas)", got[1].Label)
	assert.NotContains(t, got[1].Label, "Legendado", "must not promise subtitles that do not exist")
}

func TestMPVAudioLanguage(t *testing.T) {
	t.Parallel()
	// mpv matches --alang against whatever the manifest tags the track with, and
	// SuperFlix's sources spell the same language differently, so every plausible
	// alias must be offered — with the chosen language first.
	assert.Equal(t, "por,pob,pt-BR,ptbr,pt,portuguese", mpvAudioLanguage(audioOption{Code: "por"}))
	assert.Equal(t, "jpn,ja,jp,japanese", mpvAudioLanguage(audioOption{Code: "jpn"}))
	// Unknown codes fall back to themselves rather than to nothing.
	assert.Equal(t, "xyz", mpvAudioLanguage(audioOption{Code: "xyz"}))
}

// stubAudioPicker drives the audio-from-tracks fallback prompt through the shared
// selection seam, exposing a call counter so the "ask once per title" invariants
// stay easy to state.
func stubAudioPicker(t *testing.T, pick func([]string) (int, error)) *int {
	t.Helper()
	calls := 0
	stubSFPicker(t, func(_ string, labels []string) (int, error) {
		calls++
		return pick(labels)
	})
	return &calls
}

func TestSelectSuperFlixAudio_AsksOncePerTitle(t *testing.T) {
	// Swaps package state — not parallel.
	calls := stubAudioPicker(t, func(labels []string) (int, error) {
		require.Len(t, labels, 2)
		return 1, nil // pick the Japanese (legendado) track
	})

	codes := []string{"por", "jpn"}

	first, ok := selectSuperFlixAudio("1405", codes, true)
	require.True(t, ok)
	assert.Equal(t, "jpn", first.Code)
	assert.False(t, first.Dubbed)

	// Bingeing the same title must NOT re-ask on every episode.
	second, ok := selectSuperFlixAudio("1405", codes, true)
	require.True(t, ok)
	assert.Equal(t, "jpn", second.Code)
	assert.Equal(t, 1, *calls, "the picker must run once per title, not once per episode")

	// A different title is a fresh decision.
	_, ok = selectSuperFlixAudio("46260", codes, true)
	require.True(t, ok)
	assert.Equal(t, 2, *calls, "a new title must ask again")
}

func TestSelectSuperFlixAudio_NoPromptWhenThereIsNoChoice(t *testing.T) {
	calls := stubAudioPicker(t, func([]string) (int, error) {
		t.Error("must not prompt when there is nothing to choose between")
		return 0, nil
	})

	// One real track (und is not a language).
	got, ok := selectSuperFlixAudio("1", []string{"jpn", "und"}, true)
	require.True(t, ok)
	assert.Equal(t, "jpn", got.Code)
	assert.Zero(t, *calls)

	// No tracks at all: nothing to act on.
	_, ok = selectSuperFlixAudio("2", nil, true)
	assert.False(t, ok)
	assert.Zero(t, *calls)
}

// Without a TTY the picker fails. That must not break playback: fall back to the
// first option — Portuguese, the safe default for a PT-BR catalog.
func TestSelectSuperFlixAudio_FallsBackWhenPickerUnavailable(t *testing.T) {
	calls := stubAudioPicker(t, func([]string) (int, error) {
		return 0, errors.New("no tty")
	})

	got, ok := selectSuperFlixAudio("1405", []string{"jpn", "por"}, true)
	require.True(t, ok)
	assert.Equal(t, "por", got.Code, "must default to the dub, not to whatever came first in the stream")
	assert.True(t, got.Dubbed)
	assert.Equal(t, 1, *calls)
}

// An explicit --audio-lang means the user already answered; asking again would be
// rude, and overriding them would be worse.
func TestSelectSuperFlixAudio_RespectsUserPinnedFlag(t *testing.T) {
	calls := stubAudioPicker(t, func([]string) (int, error) {
		t.Error("must not prompt when --audio-lang was given")
		return 0, nil
	})

	prev := util.GlobalAudioLanguage
	util.GlobalAudioLanguage = "jpn"
	t.Cleanup(func() { util.GlobalAudioLanguage = prev })
	resetSuperFlixAudioChoices() // re-snapshot the flag

	_, ok := selectSuperFlixAudio("1405", []string{"por", "jpn"}, true)
	assert.False(t, ok, "the caller must leave GlobalAudioLanguage untouched")
	assert.Zero(t, *calls)
}

// The regression this guards: we WRITE GlobalAudioLanguage after the first
// episode. If that write were mistaken for a user flag, episode 2 would skip the
// picker and — worse — the next title would silently inherit episode 1's choice.
func TestSelectSuperFlixAudio_OurOwnWriteIsNotMistakenForAUserFlag(t *testing.T) {
	calls := stubAudioPicker(t, func([]string) (int, error) { return 1, nil })

	prev := util.GlobalAudioLanguage
	util.GlobalAudioLanguage = "" // user pinned nothing
	t.Cleanup(func() { util.GlobalAudioLanguage = prev })
	resetSuperFlixAudioChoices()

	got, ok := selectSuperFlixAudio("1405", []string{"por", "jpn"}, true)
	require.True(t, ok)
	// Simulate what GetSuperFlixStreamURL does next.
	util.GlobalAudioLanguage = mpvAudioLanguage(got)

	// A DIFFERENT title must still get its own prompt.
	got2, ok := selectSuperFlixAudio("46260", []string{"por", "jpn"}, true)
	require.True(t, ok, "our own write must not look like a user preference")
	assert.Equal(t, "jpn", got2.Code)
	assert.Equal(t, 2, *calls, "the second title must be asked, not inherit the first")
}
