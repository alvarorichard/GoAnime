package api

import (
	"encoding/json"
	"errors"
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

// The server list answers both of the user's questions at once — which source,
// and dubbed or subtitled — so the picker must show both on every row.
func TestSuperFlixServerChoices(t *testing.T) {
	t.Parallel()

	// Declared subtitled-first on purpose: dubbed must still be promoted.
	servers := []superflix.SuperFlixServer{
		sfServer("9", leg, "Servidor 9", false),
		sfServer("159462", dub, "Servidor 159462", false),
		sfServer("native_media:233831", dub, "MP4 Dublado", true),
	}

	got := superFlixServerChoices(servers)
	require.Len(t, got, 3)

	assert.Equal(t, "🎙️  Dublado · Servidor 159462", got[0].Label)
	assert.Equal(t, "🎙️  Dublado · MP4 Dublado", got[1].Label)
	assert.Equal(t, "💬 Legendado · Servidor 9", got[2].Label)
}

func TestSuperFlixServerChoices_UnknownAudioStillOffered(t *testing.T) {
	t.Parallel()
	// A server we can't classify is still playable — hiding it would be worse
	// than labelling it honestly.
	got := superFlixServerChoices([]superflix.SuperFlixServer{sfServer("1", 0, "Servidor 1", false)})
	require.Len(t, got, 1)
	assert.Contains(t, got[0].Label, "Servidor 1")
	assert.Contains(t, got[0].Label, "desconhecido")
}

func stubServerPicker(t *testing.T, pick func([]string) (int, error)) (*int, *[]string) {
	t.Helper()
	calls := 0
	var lastLabels []string
	prev := sfServerPickFn
	sfServerPickFn = func(labels []string) (int, error) {
		calls++
		lastLabels = labels
		return pick(labels)
	}
	resetSuperFlixServerPrefs()
	t.Cleanup(func() {
		sfServerPickFn = prev
		resetSuperFlixServerPrefs()
	})
	return &calls, &lastLabels
}

// Tehran S1E1's real list: two servers, both dubbed. The user must be able to
// choose between them — and must not be re-asked on every episode.
func TestSelectSuperFlixServer_AsksOncePerTitle(t *testing.T) {
	calls, labels := stubServerPicker(t, func([]string) (int, error) { return 1, nil }) // pick the MP4

	ep1 := []superflix.SuperFlixServer{
		sfServer("159462", dub, "Servidor 159462", false),
		sfServer("native_media:233831", dub, "MP4 Dublado", true),
	}
	got, err := selectSuperFlixServer("103913", ep1)
	require.NoError(t, err)
	assert.Equal(t, "MP4 Dublado", got.Name)
	assert.Equal(t, []string{"🎙️  Dublado · Servidor 159462", "🎙️  Dublado · MP4 Dublado"}, *labels)

	// Next episode: the ids and names CHANGE (the name embeds the per-episode id),
	// so the preference has to survive on Type+IsFile — which it does.
	ep2 := []superflix.SuperFlixServer{
		sfServer("777888", dub, "Servidor 777888", false),
		sfServer("native_media:999", dub, "MP4 Dublado", true),
	}
	got2, err := selectSuperFlixServer("103913", ep2)
	require.NoError(t, err)
	assert.Equal(t, "native_media:999", got2.IDString(), "must honor the MP4 preference on the new episode")
	assert.Equal(t, 1, *calls, "a binge must not re-ask on every episode")
}

// If the exact kind the user picked is gone this episode, the audio preference
// (dublado/legendado) must still be honored rather than silently dropped.
func TestSelectSuperFlixServer_RelaxesToAudioWhenTheKindIsGone(t *testing.T) {
	calls, _ := stubServerPicker(t, func([]string) (int, error) { return 1, nil }) // pick the MP4 (dub)

	ep1 := []superflix.SuperFlixServer{
		sfServer("1", dub, "Servidor 1", false),
		sfServer("native_media:2", dub, "MP4 Dublado", true),
	}
	_, err := selectSuperFlixServer("t", ep1)
	require.NoError(t, err)

	// This episode has no MP4 mirror, but still one dubbed server and one subtitled.
	ep2 := []superflix.SuperFlixServer{
		sfServer("3", dub, "Servidor 3", false),
		sfServer("4", leg, "Servidor 4", false),
	}
	got, err := selectSuperFlixServer("t", ep2)
	require.NoError(t, err)
	assert.Equal(t, "Servidor 3", got.Name, "must stay dubbed rather than fall onto a subtitled server")
	assert.Equal(t, dub, got.Type)
	assert.Equal(t, 1, *calls)
}

func TestSelectSuperFlixServer_NoPromptForASingleServer(t *testing.T) {
	calls, _ := stubServerPicker(t, func([]string) (int, error) {
		t.Error("must not prompt when there is only one server")
		return 0, nil
	})

	got, err := selectSuperFlixServer("t", []superflix.SuperFlixServer{sfServer("1", dub, "Único", false)})
	require.NoError(t, err)
	assert.Equal(t, "Único", got.Name)
	assert.Zero(t, *calls)
}

// Without a TTY the picker fails. Playback must go on, defaulting to the first
// option — a dubbed server, since dubbed sorts first.
func TestSelectSuperFlixServer_FallsBackWhenPickerUnavailable(t *testing.T) {
	calls, _ := stubServerPicker(t, func([]string) (int, error) { return 0, errors.New("no tty") })

	got, err := selectSuperFlixServer("t", []superflix.SuperFlixServer{
		sfServer("9", leg, "Servidor 9", false),
		sfServer("1", dub, "Servidor 1", false),
	})
	require.NoError(t, err)
	assert.Equal(t, dub, got.Type, "the safe default for a PT-BR catalog is the dub")
	assert.Equal(t, 1, *calls)
}

// The whole point of tagging servers dublado/legendado: the choice the user
// already made must drive mpv, instead of asking them the same thing twice.
func TestAudioForServer(t *testing.T) {
	t.Parallel()

	// The real track list a SuperFlix HLS carries.
	tracks := []string{"por", "und", "eng", "spa", "kor", "jpn", "chi", "und"}

	t.Run("dubbed server plays Portuguese", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "por,pob,pt-BR,ptbr,pt,portuguese",
			audioForServer(sfServer("1", dub, "Servidor 1", false), tracks))
	})

	// A legendado server IS the original-audio source, so its default track is
	// already right. Guessing "the first non-Portuguese track" picks by manifest
	// order and gets it wrong exactly where it matters: Tehran's original audio is
	// Hebrew, carried as the undetermined ("und") track, while "eng" is just
	// another dub. Forcing a language there would hand the viewer English in the
	// name of avoiding Portuguese.
	t.Run("subtitled server keeps the stream's own audio", func(t *testing.T) {
		t.Parallel()
		assert.Empty(t, audioForServer(sfServer("2", leg, "Servidor 2", false), tracks),
			"must not guess the original language")
	})

	t.Run("dubbed server whose stream has no Portuguese track leaves mpv alone", func(t *testing.T) {
		t.Parallel()
		assert.Empty(t, audioForServer(sfServer("3", dub, "S", false), []string{"jpn", "eng"}),
			"forcing a track the manifest does not carry would just break audio")
	})

	// Subtitles are not this function's business — it must not be able to suppress
	// them by returning some "no subs" signal. An earlier version did exactly that,
	// and subtitles that had always worked stopped appearing whenever the dub was
	// picked.
	t.Run("says nothing about subtitles", func(t *testing.T) {
		t.Parallel()
		assert.Empty(t, audioForServer(sfServer("4", leg, "S", false), nil))
	})
}
