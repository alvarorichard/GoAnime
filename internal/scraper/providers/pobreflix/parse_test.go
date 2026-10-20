package pobreflix

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixtureDoc reads a testdata file as a goquery document. The fixtures are
// the sites' real pages, captured 2026-10-16: the day the scraping contract
// was written down, kept as the regression baseline.
func fixtureDoc(t *testing.T, name string) *goquery.Document {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err, "fixture %s must be present", name)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(raw)))
	require.NoError(t, err)
	return doc
}

const (
	movieURL   = "https://www.pobreflixtv.sale/filmes/online/o-apanhador-de-sonhos-dublado-84752/"
	seriesURL  = "https://www.pobreflixtv.sale/series/online/ben-hur-a-serie-dublado-83633/"
	episodeURL = "https://www.pobreflixtv.sale/episodios/online/ben-hur-a-serie-1x1-dublado-83634/"
)

// TestParseTitle_MovieFixture: the movie page names the record id and the
// TMDB id the panel keys playback by — the whole point of opening it.
func TestParseTitle_MovieFixture(t *testing.T) {
	t.Parallel()
	doc := fixtureDoc(t, "movie.html")
	title, ok := parseTitle(doc, movieURL)
	require.True(t, ok, "the movie fixture must parse")

	assert.Equal(t, KindMovie, title.Kind)
	assert.Equal(t, 84752, title.VideoID, "data-video-id")
	assert.Equal(t, 793002, title.TMDBID, "data-tmdb-id")
	assert.Equal(t, AudioDubbed, title.Audio, "the record slug says dublado")
	assert.Equal(t, "O Apanhador de Sonhos", title.Name, "h2.one is the clean PT record name")
	assert.Empty(t, title.Seasons, "a movie lists no seasons")
	// The poster is the record's TMDB image; upgrade rules live with the
	// api layer — the parser carries what the page says.
	assert.Contains(t, title.Poster, "image.tmdb.org", "og:image is the TMDB poster")
}

// TestParseTitle_SeriesFixture: the series page lists its seasons (with the
// audio each season option belongs to) and Portuguese episode names.
func TestParseTitle_SeriesFixture(t *testing.T) {
	t.Parallel()
	doc := fixtureDoc(t, "series.html")
	title, ok := parseTitle(doc, seriesURL)
	require.True(t, ok, "the series fixture must parse")

	assert.Equal(t, KindSeries, title.Kind)
	assert.Equal(t, 83633, title.VideoID)
	assert.Equal(t, 288458, title.TMDBID, "Ben-Hur: A Série's TMDB id")
	assert.Equal(t, AudioDubbed, title.Audio)
	assert.Equal(t, "Ben-Hur: A Série", title.Name)
	require.NotEmpty(t, title.Seasons, "the rendered season block must list seasons")

	season := title.Seasons[0]
	assert.Equal(t, 1, season.Number)
	assert.Equal(t, AudioDubbed, season.Audio, "the season option tags its audio")
	require.NotEmpty(t, season.Episodes)

	first := season.Episodes[0]
	assert.Equal(t, 1, first.Number)
	assert.Equal(t, "O Judeu e o Romano", first.Title, "the PT episode name is on the card")
	assert.True(t, strings.Contains(first.URL, "/episodios/online/ben-hur-a-serie-1x1-dublado-"), "the card links its own record")

	// Sorted, more than one episode, and each with a name to fill the panel.
	assert.GreaterOrEqual(t, len(season.Episodes), 4)
	seen := map[int]bool{}
	for _, ep := range season.Episodes {
		assert.False(t, seen[ep.Number], "no duplicate episode numbers")
		seen[ep.Number] = true
	}
}

// TestParseTitle_EpisodeFixture: an episode page carries the same ids plus
// its slot (temp/ep) — the builder for history-driven single-episode opens.
func TestParseTitle_EpisodeFixture(t *testing.T) {
	t.Parallel()
	doc := fixtureDoc(t, "episode.html")
	title, ok := parseTitle(doc, episodeURL)
	require.True(t, ok, "the episode fixture must parse")

	assert.Equal(t, KindSeries, title.Kind, "an episode record is the series' content")
	assert.Equal(t, 83634, title.VideoID, "its own record id")
	assert.Equal(t, 288458, title.TMDBID, "same TMDB id as the series")
	assert.Equal(t, AudioDubbed, title.Audio)
	assert.NotEqual(t, "Ben-Hur: A Série 1X1", title.Name, "the 1X1 slot must not leak into the name")
}

// TestParseTitle_RejectsPagesWithoutMount: a reskin that drops the #view
// element must fail loudly (ErrNoTitle in the client), never return a Title
// whose TMDB id is zero — the api layer plays by that id.
func TestParseTitle_RejectsPagesWithoutMount(t *testing.T) {
	t.Parallel()
	_, ok := parseTitle(goqueryT(t, "<html><body><p>nothing here</p></body></html>"), movieURL)
	assert.False(t, ok, "a page without #view cannot parse")

	_, ok = parseTitle(goqueryT(t, `<html><body><div id="content">no view</div></body></html>`), movieURL)
	assert.False(t, ok, "no mount, no title")

	// A mount without the TMDB id names nothing to play by.
	_, ok = parseTitle(goqueryT(t, `<html><body><div id="view" data-video-id="1"></div></body></html>`), movieURL)
	assert.False(t, ok, "no TMDB id, no playable key")
}

func goqueryT(t *testing.T, html string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)
	return doc
}

// TestParseRecordPath: the URL grammar every search result rides on. These
// are the site's own examples, at least one captured per shape the corpus
// feeds in.
func TestParseRecordPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path     string
		kind     MediaKind
		slug     string
		audio    Audio
		videoID  int
		removeok bool
	}{
		{path: "/filmes/online/verity-dublado-84506/", kind: KindMovie, slug: "verity", audio: AudioDubbed, videoID: 84506},
		{path: "/filmes/online/a-tristeza-legendado-84772/", kind: KindMovie, slug: "a-tristeza", audio: AudioSubtitled, videoID: 84772},
		{path: "/series/online/ben-hur-a-serie-dublado-83633/", kind: KindSeries, slug: "ben-hur-a-serie", audio: AudioDubbed, videoID: 83633},
		{path: "/series/online/little-birds-dublado-84762/", kind: KindSeries, slug: "little-birds", audio: AudioDubbed, videoID: 84762},
		// Dual-audio records carry their own marker.
		{path: "/filmes/online/o-terminal-dual-dublado-84749/", kind: KindMovie, slug: "o-terminal", audio: AudioDubbed, videoID: 84749},
		// Records without an audio marker still parse; only the audio is
		// unknown.
		{path: "/filmes/online/furie-84110/", kind: KindMovie, slug: "furie", audio: AudioUnknown, videoID: 84110},
		// Episode pages never enter the corpus.
		{path: "/episodios/online/ben-hur-a-serie-1x8-dublado-84788/", removeok: true},
		// Category and junk shapes fail closed.
		{path: "/filmes/", removeok: true},
		{path: "/series/online/", removeok: true},
		{path: "/filmes/online/no-id-here/", removeok: true},
		{path: "", removeok: true},
	}
	for _, tt := range tests {
		kind, slug, audio, id, ok := parseRecordPath(tt.path)
		if tt.removeok {
			assert.False(t, ok, "%q must not parse as a corpus record", tt.path)
			continue
		}
		require.True(t, ok, "%q must parse", tt.path)
		assert.Equal(t, tt.kind, kind, "%q kind", tt.path)
		assert.Equal(t, tt.slug, slug, "%q base slug", tt.path)
		assert.Equal(t, tt.audio, audio, "%q audio", tt.path)
		assert.Equal(t, tt.videoID, id, "%q record id", tt.path)
	}
}

// TestSlugTokensAndTitle: folding must survive punctuation-heavy slugs
// ("ben-hur-a-serie", "morgan-2-temporada"), and the display form must title
// all words, so "todo poderoso" reads as a person typed it.
func TestSlugTokensAndTitle(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "o apanhador de sonhos", slugTokens("o-apanhador-de-sonhos"))
	assert.Equal(t, "ben hur a serie", slugTokens("ben-hur-a-serie"))
	assert.Equal(t, "morgan 2 temporada", slugTokens("morgan-2-temporada"))
	assert.Equal(t, "", slugTokens("---"))
	assert.Equal(t, "Todo Poderoso", slugToTitle("todo-poderoso"))
	assert.Equal(t, "Ben Hur A Serie", slugToTitle("ben-hur-a-serie"))
}

// TestDisplayTitle: the h1 shapes observed across the catalog, including the
// episode page's "1X1" slot.
func TestDisplayTitle(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "Verity", displayTitle("Assistir Verity Online Gratis"))
	assert.Equal(t, "O Apanhador de Sonhos", displayTitle("Assistir O Apanhador de Sonhos Online"))
	assert.Equal(t, "Ben-Hur: A Série", displayTitle("Assistir Ben-Hur: A Série 1X1 Online"))
	assert.Equal(t, "", displayTitle("   "))
}

// TestQueryFold: the query side must meet the slugs' accent-free forms.
func TestQueryFold(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "acao vinganca", slugTokens(queryFoldRe.Replace("Ação Vingança")))
	assert.Equal(t, "todo poderoso", slugTokens(queryFoldRe.Replace("TôDo Pöderoso")))
}
