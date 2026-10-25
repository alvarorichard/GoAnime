package handlers

import (
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/tracking"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func continueTestRows() []tracking.Anime {
	now := time.Now()
	return []tracking.Anime{
		{AllanimeID: "https://animefire.io/x:ep5", AnimeTitle: "Mukashi Banashi", Title: "Urashima Tarou", Source: "Animefire.io", EpisodeNumber: 5, PlaybackTime: 372, Duration: 580, MediaType: "anime", LastUpdated: now.Add(-time.Hour)},
		{AllanimeID: "https://animefire.io/x:ep4", AnimeTitle: "Mukashi Banashi", Title: "Mikanbune", Source: "Animefire.io", EpisodeNumber: 4, PlaybackTime: 570, Duration: 580, MediaType: "anime", LastUpdated: now.Add(-48 * time.Hour)},
		{AllanimeID: "https://startflix.test/m:ep1", AnimeTitle: "Dark", Title: "Dark", Source: "StartFlix", EpisodeNumber: 1, PlaybackTime: 100, Duration: 3000, MediaType: "movie", LastUpdated: now.Add(-2 * time.Hour)},
		{AllanimeID: "https://animefire.io/y:ep1", AnimeTitle: "Kuro Nyago", Title: "Kuro Nyago", Source: "Animefire.io", EpisodeNumber: 1, PlaybackTime: 0, Duration: 600, MediaType: "anime", LastUpdated: now.Add(-3 * time.Hour)},
	}
}

func TestGroupContinueEntries(t *testing.T) {
	t.Parallel()

	entries := groupContinueEntries(continueTestRows())

	require.Len(t, entries, 2, "same-show episodes collapse; unwatched noise drops")
	assert.Equal(t, "Mukashi Banashi", entries[0].AnimeTitle, "newest first")
	assert.Equal(t, 5, entries[0].EpisodeNum, "latest episode wins")
	assert.Equal(t, "Urashima Tarou", entries[0].EpisodeTitle)
	assert.Equal(t, "https://animefire.io/x:ep5", entries[0].Key)
	assert.Equal(t, "Dark", entries[1].AnimeTitle)
	assert.Equal(t, "movie", entries[1].MediaType)
}

func TestGroupContinueEntriesOldRows(t *testing.T) {
	t.Parallel()
	now := time.Now()

	entries := groupContinueEntries([]tracking.Anime{
		{AllanimeID: "https://animefire.io/x:ep5", Title: "Ep Five", EpisodeNumber: 5, PlaybackTime: 10, Duration: 100, LastUpdated: now},
		{AllanimeID: "https://animefire.io/x:ep4", Title: "Ep Four", EpisodeNumber: 4, PlaybackTime: 10, Duration: 100, LastUpdated: now.Add(-time.Hour)},
		{AllanimeID: "https://other.io/z:ep1", Title: "Other", EpisodeNumber: 1, PlaybackTime: 10, Duration: 100, LastUpdated: now},
	})

	require.Len(t, entries, 2, "rows without titles group by episode URL")
	assert.Equal(t, 5, entries[0].EpisodeNum)
}

func TestSourceFromURL(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "StartFlix", sourceFromURL("https://www.startflix.test/series/dark:ep1"))
	assert.Equal(t, "Animefire.io", sourceFromURL("https://animefire.io/video/x:ep5"))
	assert.Equal(t, "HiAnime", sourceFromURL("https://hianime.to/watch/y"))
	assert.Equal(t, "", sourceFromURL("https://unknown.example/x"))
	assert.Equal(t, "", sourceFromURL(""))
}

func TestGroupContinueEntriesRecoversSource(t *testing.T) {
	t.Parallel()
	now := time.Now()
	entries := groupContinueEntries([]tracking.Anime{
		{AllanimeID: "https://www.startflix.test/f/velozes:ep1", Title: "Velozes e Furiosos", EpisodeNumber: 1, PlaybackTime: 100, Duration: 3000, MediaType: "movie", LastUpdated: now},
	})
	require.Len(t, entries, 1)
	assert.Equal(t, "StartFlix", entries[0].Source, "old movie rows recover their catalog from the URL")
}

func TestContinueTarget(t *testing.T) {
	t.Parallel()

	mid := tui.HomeContinueEntry{AnimeTitle: "Show", EpisodeNum: 5, Position: 372, Duration: 580, MediaType: "anime"}
	ep, pos, resume, start := continueTarget(mid, false)
	assert.Equal(t, 5, ep)
	assert.Equal(t, 372, pos)
	assert.True(t, resume)
	assert.False(t, start)

	done := tui.HomeContinueEntry{AnimeTitle: "Show", EpisodeNum: 5, Position: 590, Duration: 600, MediaType: "anime"}
	ep, pos, resume, start = continueTarget(done, false)
	assert.Equal(t, 6, ep, "a finished episode offers the next one, not the credits")
	assert.Equal(t, 0, pos)
	assert.False(t, resume)
	assert.False(t, start)

	movie := tui.HomeContinueEntry{AnimeTitle: "Dark", EpisodeNum: 1, Position: 2900, Duration: 3000, MediaType: "movie"}
	ep, _, resume, start = continueTarget(movie, false)
	assert.Equal(t, 1, ep, "a finished movie replays itself, not episode 2")
	assert.False(t, resume)
	assert.True(t, start)

	ep, _, _, start = continueTarget(mid, true)
	assert.Equal(t, 5, ep)
	assert.True(t, start, "r starts over")

	midMovie := tui.HomeContinueEntry{AnimeTitle: "Dark", EpisodeNum: 1, Position: 100, Duration: 3000}
	ep, pos, resume, _ = continueTarget(midMovie, false)
	assert.Equal(t, 1, ep)
	assert.Equal(t, 100, pos)
	assert.True(t, resume)
}

func TestPickContinueVersionPrefersMediaType(t *testing.T) {
	t.Parallel()
	anime := &models.Anime{Name: "Dark", Source: "Animefire.io", MediaType: models.MediaTypeAnime, URL: "https://animefire.io/dark"}
	movie := &models.Anime{Name: "Dark", Source: "StartFlix", MediaType: models.MediaTypeMovie, URL: "https://startflix.test/dark"}

	pick := pickContinueVersion("", "movie")
	assert.Same(t, movie, pick([]*models.Anime{anime, movie}),
		"a tracked movie must not reopen as the anime of the same name")

	pick = pickContinueVersion("", "anime")
	assert.Same(t, anime, pick([]*models.Anime{anime, movie}))

	pick = pickContinueVersion("StartFlix", "anime")
	assert.Same(t, movie, pick([]*models.Anime{anime, movie}), "source still wins over media type")

	pick = pickContinueVersion("", "")
	assert.Same(t, anime, pick([]*models.Anime{anime, movie}), "no hints keeps closest match first")
}

func FuzzContinueGroupKey(f *testing.F) {
	for _, seed := range []string{"", "k", "https://x.io/a:ep5", "https://x.io/a:ep", "Título [PT-BR]"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, id string) {
		a := tracking.Anime{AllanimeID: id, AnimeTitle: id, LastUpdated: time.Now()}
		k1 := continueGroupKey(a)
		k2 := continueGroupKey(a)
		if k1 == "" || k1 != k2 {
			t.Fatalf("key must be stable and non-empty: %q", k1)
		}
	})
}

func FuzzSourceFromURL(f *testing.F) {
	for _, seed := range []string{"", "foo", "https://www.startflix.biz/x:ep1", "HTTP://ANIMEFIRE.ONE/Y", "://", "www.goyabu.gg/"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		_ = sourceFromURL(raw)
		_ = sourceFromURL(raw)
	})
}

func FuzzCleanContinueTitle(f *testing.F) {
	for _, seed := range []string{"", "[]", "[PT-BR] X (Dublado)", "[[[", "X (Legendado) (Dublado)"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, name string) {
		got := cleanContinueTitle(&tui.HomeContinueEntry{AnimeTitle: name})
		if strings.ContainsAny(got, "\n\r\x00") {
			t.Fatalf("control leak: %q", got)
		}
	})
}
