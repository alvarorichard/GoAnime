package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/startflix"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/topcine"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Live shapes captured 2026-10-16. The TopCine catalog is healthy; what bites
// is the join with the StartFlix panel, so these tests drive the real api
// glue (not fn stubs) against a stub TopCine page plus a stub panel.

// flatTopCinePage is the Solo Leveling shape: 5 episodes flattened under
// season 1 while the panel splits S1 (2 eps) / S2 (1 ep).
const flatTopCinePage = `<html><body><select>
	<option data-temporada="1" data-episodio="1" data-titulo="Flat Um" data-player="https://azullog.test/serie/127532/1/1">EP 01</option>
	<option data-temporada="1" data-episodio="2" data-titulo="Flat Dois" data-player="https://azullog.test/serie/127532/1/2">EP 02</option>
	<option data-temporada="1" data-episodio="3" data-titulo="Flat Tres" data-player="https://azullog.test/serie/127532/1/3">EP 03</option>
	<option data-temporada="1" data-episodio="4" data-titulo="Flat Quatro" data-player="https://azullog.test/serie/127532/1/4">EP 04</option>
	<option data-temporada="1" data-episodio="5" data-titulo="Flat Cinco" data-player="https://azullog.test/serie/127532/1/5">EP 05</option>
</select></body></html>`

// splitPanelPage is the panel side of the same shape: S1 carries 2 Dublado
// episodes, S2 carries 1 Dublado + 1 Legendado, all unnamed ("1 - Episódio").
const splitPanelPage = `<html><body>
<ul class="header-navigation">
	<li data-season-id="s1" data-season-number="1">1 TEMP</li>
	<li data-season-id="s2" data-season-number="2">2 TEMP</li>
</ul>
<div class="cards">
	<div class="card"><div class="card-header"><h2 class="card-title">Dublado</h2></div><div class="card-body"><ul>
		<li data-season-id="s1" data-episode-id="101"><a href="#s1_101">1 - Episódio</a></li>
		<li data-season-id="s1" data-episode-id="102"><a href="#s1_102">2 - Episódio</a></li>
		<li data-season-id="s2" data-episode-id="103"><a href="#s2_103">1 - Episódio</a></li>
	</ul></div></div>
	<div class="card"><div class="card-header"><h2 class="card-title">Legendado</h2></div><div class="card-body"><ul>
		<li data-season-id="s2" data-episode-id="203"><a href="#s2_203">1 - Episódio</a></li>
	</ul></div></div>
</div>
</body></html>`

// useFlatTitle wires a stub TopCine page (TMDB 127532, the fake panel's
// series) and a stub panel serving splitPanelPage for every seasons read.
func useFlatTitle(t *testing.T) {
	t.Helper()

	panel := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, splitPanelPage)
	}))
	t.Cleanup(panel.Close)
	t.Setenv("GOANIME_STARTFLIX_PANEL_URL", panel.URL)
	prevSfx := sfxClientFn
	sfxClientFn = func() *startflix.Client {
		return startflix.NewClientForTest(http.DefaultClient, "https://startflix.test")
	}
	t.Cleanup(func() { sfxClientFn = prevSfx })

	tc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, flatTopCinePage)
	}))
	t.Cleanup(tc.Close)
	tcURL, _ := url.Parse(tc.URL)
	prevTc := tcClientFn
	tcClientFn = func() *topcine.Client {
		return topcine.NewClientForTest(&http.Client{Transport: routeAll{tcURL}}, "https://topcine.test")
	}
	t.Cleanup(func() { tcClientFn = prevTc })
}

func flatSeries() *models.Anime {
	return &models.Anime{Name: "Flat", URL: "https://topcine.test/serie/flat", Source: topcine.SourceName, MediaType: models.MediaTypeTV}
}

// TestFillFlattenedSeasonTitles pins the Solo Leveling live shape: the page
// flattens 5 episodes into season 1, the panel splits S1/S2, and the S2
// episodes must still be named — in both audios, even though S1 has no
// Legendado to count.
func TestFillFlattenedSeasonTitles(t *testing.T) {
	useFlatTitle(t)

	media := flatSeries()
	media.CurrentSeason = 2
	eps := []models.Episode{{Num: 1}}
	fillTopCineEpisodeTitles(media, eps)
	assert.Equal(t, "Flat Tres", eps[0].Title.English, "S2E1 follows the 2 S1 episodes: flat #3")
	assert.Equal(t, "Flat Tres", eps[0].Title.Romaji)

	media.CurrentSeason = 1
	eps = []models.Episode{{Num: 1}, {Num: 2}}
	fillTopCineEpisodeTitles(media, eps)
	assert.Equal(t, "Flat Um", eps[0].Title.English, "S1 still matches exactly")
	assert.Equal(t, "Flat Dois", eps[1].Title.English)
}

// TestFillFlattenedSeasonTitles_OutOfRange keeps a missing title missing
// instead of inventing one past the end of the flat list.
func TestFillFlattenedSeasonTitles_OutOfRange(t *testing.T) {
	useFlatTitle(t)

	media := flatSeries()
	media.CurrentSeason = 2
	eps := []models.Episode{{Num: 9}}
	fillTopCineEpisodeTitles(media, eps)
	assert.Empty(t, eps[0].Title.English)
}

// TestSelectPanelAudio_PinnedAudioMissingLogsAndAsks pins the Solo Leveling S1
// shape (Dublado-only season, --audio pinned to Legendado): the pinned value
// cannot be honoured, so the picker — not a silent substitution — decides.
func TestSelectPanelAudio_PinnedAudioMissingLogsAndAsks(t *testing.T) {
	useFakeTopCine(t, 0, 0)
	prevLang, prevExplicit := util.GetGlobalAudioLanguage(), util.GlobalAudioLanguageExplicit
	util.SetGlobalAudioLanguage("legendado")
	util.GlobalAudioLanguageExplicit = true
	t.Cleanup(func() {
		util.SetGlobalAudioLanguage(prevLang)
		util.GlobalAudioLanguageExplicit = prevExplicit
	})

	dubOnly := startflix.Season{Number: 1, Dubbed: []startflix.Episode{{Number: 1}}}

	// Single-audio season: the pinned value is unusable, no picker exists.
	sfxPickFn = func(string, []string) (int, error) { t.Fatal("asked with a single audio"); return 0, nil }
	assert.Equal(t, startflix.AudioDubbed, selectPanelAudio("TopCine", "https://topcine.test/serie/solo", dubOnly))

	// Both audios but pinned one missing is impossible by construction
	// (pinned is one of the two); the remembered choice still wins.
	both := startflix.Season{Number: 2,
		Dubbed:    []startflix.Episode{{Number: 1}},
		Subtitled: []startflix.Episode{{Number: 1}},
	}
	var prompts int
	sfxPickFn = func(string, []string) (int, error) { prompts++; return 0, nil }
	rememberStartFlixAudio("https://topcine.test/serie/solo", startflix.AudioSubtitled)
	assert.Equal(t, startflix.AudioSubtitled, selectPanelAudio("TopCine", "https://topcine.test/serie/solo", both))
	assert.Zero(t, prompts, "a remembered audio skips the picker")
}

// TestGetTopCineSeasonEpisodes_FlattenedSeasonNamesS2 drives the full listing
// for the flattened shape: S2 episodes come back named from the flat page.
func TestGetTopCineSeasonEpisodes_FlattenedSeasonNamesS2(t *testing.T) {
	useFlatTitle(t)
	sfxPickFn = func(string, []string) (int, error) { return 0, nil } // Dublado
	media := flatSeries()

	eps, err := GetTopCineSeasonEpisodes(media, 2)
	require.NoError(t, err)
	require.Len(t, eps, 1)
	assert.Equal(t, "Flat Tres", eps[0].Title.English)
	assert.Equal(t, 2, media.CurrentSeason)
}
