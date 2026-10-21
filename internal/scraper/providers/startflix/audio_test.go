package startflix

import (
	"context"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A movie panel names each player's audio in its label and mixes both audios
// on one list (2026-10-21, 24 of 24 movies sampled from startflix.biz/filmes:
// "1 Dublado" on Abyss, "2 Dublado" on Byse, "Legendado" on vidsrc.sh). These
// tests pin that the audio the user is watching in survives resolution: the
// Todo Poderoso report was a Dublado title that played in English.

// TestPlayerAudio reads the labels the live panel renders, whitespace and
// all.
func TestPlayerAudio(t *testing.T) {
	tests := []struct {
		label string
		want  Audio
	}{
		{"HD #1 Dublado", AudioDubbed},  // Todo Poderoso, tt0315327
		{"  2 Dublado", AudioDubbed},    // &nbsp; before the number
		{"1 Legendado", AudioSubtitled}, // panel_movie_2026_10_07.html
		{"  Legendado", AudioSubtitled}, // the vidsrc player the panel appends
		{"DUBLADO", AudioDubbed},
		{"Dublada", AudioDubbed},
		{"Player #1 HD", ""}, // an episode's buttons: the list says the audio
		{"Player #3", ""},
		{"", ""},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, Player{Label: tt.label}.Audio(), "label %q", tt.label)
	}
}

func TestListedAudios(t *testing.T) {
	dub := Player{Label: "1 Dublado"}
	sub := Player{Label: "Legendado"}
	plain := Player{Label: "Player #1 HD"}

	assert.Equal(t, []Audio{AudioDubbed}, ListedAudios([]Player{dub, dub}, AudioDubbed))
	assert.Equal(t, []Audio{AudioDubbed, AudioSubtitled}, ListedAudios([]Player{sub, dub}, AudioDubbed), "dub first, whatever the order")
	assert.Equal(t, []Audio{AudioSubtitled}, ListedAudios([]Player{sub}, AudioDubbed), "a Legendado-only movie offers no dub")
	assert.Equal(t, []Audio{AudioDubbed, AudioSubtitled}, ListedAudios([]Player{plain, sub}, AudioDubbed), "an unlabelled player is the list's audio")
	assert.Equal(t, []Audio{AudioSubtitled}, ListedAudios([]Player{plain}, AudioSubtitled))
	assert.Empty(t, ListedAudios([]Player{plain}, ""))
	assert.Empty(t, ListedAudios(nil, AudioDubbed))
}

// TestResolveStreamForPrefersWantedAudio: the wanted audio beats a taller
// stream in the other one, and the other audio is still handed back — named
// as such — when it is all that plays.
func TestResolveStreamForPrefersWantedAudio(t *testing.T) {
	byse := func(code, label string) Player {
		return Player{URL: "https://embedplaybyse.test/e/" + code + "/x", Type: "iframe", Label: label}
	}
	heights := map[string]int{"dub72000": 720, "leg10800": 1080, "dub10800": 1080}
	dub720 := byse("dub72000", "1 Dublado")
	leg1080 := byse("leg10800", "Legendado")
	deadDub := byse("missing1", "HD #1 Dublado")
	plain := byse("dub10800", "Player #1 HD")

	tests := []struct {
		name      string
		players   []Player
		want      Audio
		wantCode  string
		wantAudio Audio
	}{
		{"dub beats a taller Legendado", []Player{leg1080, dub720}, AudioDubbed, "dub72000", AudioDubbed},
		{"Legendado beats a shorter dub", []Player{dub720, leg1080}, AudioSubtitled, "leg10800", AudioSubtitled},
		{"no preference keeps the tallest", []Player{dub720, leg1080}, "", "leg10800", AudioSubtitled},
		{"dead dub hands back the Legendado, named", []Player{deadDub, leg1080}, AudioDubbed, "leg10800", AudioSubtitled},
		{"an unlabelled player is the wanted audio", []Player{leg1080, plain}, AudioDubbed, "dub10800", AudioDubbed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			c := newTestClient(t, qualityHosts(t, heights, &calls))
			s, err := c.ResolveStreamFor(context.Background(), tt.players, tt.want)
			require.NoError(t, err)
			assert.Contains(t, s.URL, "/"+tt.wantCode+"/")
			assert.Equal(t, tt.wantAudio, s.Audio)
		})
	}
}

// TestResolveStreamForYtDlpTriesWantedAudioFirst: yt-dlp's first answer wins,
// so a quick Legendado host must not outrun a slower Dublado one.
func TestResolveStreamForYtDlpTriesWantedAudioFirst(t *testing.T) {
	t.Setenv("GOANIME_YTDLP_FALLBACK_TEST", "1")
	old := ytdlpResolveFn
	t.Cleanup(func() { ytdlpResolveFn = old })

	dubAnswers := true
	var mu sync.Mutex
	var asked []string
	ytdlpResolveFn = func(ctx context.Context, playerURL, _ string) string {
		mu.Lock()
		asked = append(asked, playerURL)
		mu.Unlock()
		switch playerURL {
		case "https://vidsrc.test/embed/movie/tt1":
			return "https://cdn.test/leg.mp4"
		case "https://dood.test/e/dub":
			if !dubAnswers {
				return ""
			}
			select { // slower than the Legendado host
			case <-time.After(50 * time.Millisecond):
			case <-ctx.Done():
				return ""
			}
			return "https://cdn.test/dub.mp4"
		}
		return ""
	}
	players := []Player{
		{URL: "https://vidsrc.test/embed/movie/tt1", Type: "iframe", Label: "Legendado"},
		{URL: "https://dood.test/e/dub", Type: "iframe", Label: "HD #1 Dublado"},
	}
	c := NewClientForTest(nil, "https://www.startflix.test")

	s, err := c.ResolveStreamFor(context.Background(), players, AudioDubbed)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.test/dub.mp4", s.URL)
	assert.Equal(t, AudioDubbed, s.Audio)
	assert.False(t, slices.Contains(asked, "https://vidsrc.test/embed/movie/tt1"), "the Legendado round never ran: %v", asked)

	dubAnswers, asked = false, nil
	s, err = c.ResolveStreamFor(context.Background(), players, AudioDubbed)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.test/leg.mp4", s.URL)
	assert.Equal(t, AudioSubtitled, s.Audio, "the other audio comes back named, for the caller to ask about")
}
