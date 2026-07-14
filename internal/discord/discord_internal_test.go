package discord

import (
	"sync"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAbs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   int
		want int
	}{
		{"positive", 5, 5},
		{"negative", -5, 5},
		{"zero", 0, 0},
		{"min int + 1 stays positive", -1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, abs(tt.in))
		})
	}
}

func TestCleanMediaTags(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"strip Movies/TV", "[Movies/TV] Inception", "Inception"},
		{"strip Movie", "[Movie] Inception", "Inception"},
		{"strip TV", "[TV] Breaking Bad", "Breaking Bad"},
		{"strip English", "[English] Naruto", "Naruto"},
		{"strip PT-BR", "[PT-BR] Naruto", "Naruto"},
		{"strip Portuguese", "[Portuguese] Naruto", "Naruto"},
		{"strip Português", "[Português] Naruto", "Naruto"},
		{"strip multiple tags", "[Movie] [PT-BR] Inception", "Inception"},
		{"trims surrounding whitespace", "  [Movie] Foo  ", "Foo"},
		{"untouched when no tag", "Plain Name", "Plain Name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, cleanMediaTags(tt.in))
		})
	}
}

func TestFormatTime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		seconds int
		want    string
	}{
		{"zero", 0, "0:00"},
		{"under minute", 5, "0:05"},
		{"exact minute", 60, "1:00"},
		{"minutes and seconds", 65, "1:05"},
		{"over hour", 3725, "1:02:05"},
		{"two-digit padding seconds", 70, "1:10"},
		{"two-digit padding minutes in hour", 3600, "1:00:00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, FormatTime(tt.seconds))
		})
	}
}

func mkRPU(a *models.Anime) *RichPresenceUpdater {
	paused := false
	return &RichPresenceUpdater{
		anime:      a,
		isPaused:   &paused,
		animeMutex: &sync.Mutex{},
	}
}

func TestGetTitle_MovieOrTV_StripsTags(t *testing.T) {
	t.Parallel()
	a := &models.Anime{Name: "[Movie] Inception"}
	rpu := mkRPU(a)
	assert.Equal(t, "Inception", rpu.getTitle(true))
}

func TestGetTitle_AnimeRomaji(t *testing.T) {
	t.Parallel()
	a := &models.Anime{
		Name:    "Naruto",
		Details: models.AniListDetails{Title: models.Title{Romaji: "Naruto", English: "Naruto English"}},
	}
	rpu := mkRPU(a)
	assert.Equal(t, "Naruto", rpu.getTitle(false))
}

func TestGetTitle_AnimeFallsBackToEnglish(t *testing.T) {
	t.Parallel()
	a := &models.Anime{
		Name:    "Bleach",
		Details: models.AniListDetails{Title: models.Title{English: "Bleach EN"}},
	}
	rpu := mkRPU(a)
	assert.Equal(t, "Bleach EN", rpu.getTitle(false))
}

func TestGetTitle_AnimeStripsLeadingBracketTag(t *testing.T) {
	t.Parallel()
	a := &models.Anime{Name: "[English] One Piece"}
	rpu := mkRPU(a)
	got := rpu.getTitle(false)
	assert.Equal(t, "One Piece", got)
}

func TestBuildButtons_MovieIMDBOnly(t *testing.T) {
	t.Parallel()
	a := &models.Anime{IMDBID: "tt1234567"}
	rpu := mkRPU(a)
	btns := rpu.buildButtons(true)
	require.Len(t, btns, 1)
	assert.Contains(t, btns[0].Url, "imdb.com/title/tt1234567")
	assert.Equal(t, "View on IMDB", btns[0].Label)
}

func TestBuildButtons_MovieIMDBPlusTMDB(t *testing.T) {
	t.Parallel()
	a := &models.Anime{IMDBID: "tt1", TMDBID: 27205, MediaType: "movie"}
	rpu := mkRPU(a)
	btns := rpu.buildButtons(true)
	require.Len(t, btns, 2)
	assert.Contains(t, btns[1].Url, "themoviedb.org/movie/27205")
}

func TestBuildButtons_TV_TMDBPath(t *testing.T) {
	t.Parallel()
	a := &models.Anime{TMDBID: 1405, MediaType: "tv"}
	rpu := mkRPU(a)
	btns := rpu.buildButtons(true)
	require.Len(t, btns, 1)
	assert.Contains(t, btns[0].Url, "themoviedb.org/tv/1405")
}

func TestBuildButtons_Anime_AniListAndMAL(t *testing.T) {
	t.Parallel()
	a := &models.Anime{AnilistID: 21, MalID: 21}
	rpu := mkRPU(a)
	btns := rpu.buildButtons(false)
	require.Len(t, btns, 2)
	assert.Contains(t, btns[0].Url, "anilist.co/anime/21")
	assert.Contains(t, btns[1].Url, "myanimelist.net/anime/21")
}

func TestBuildButtons_AnimeWithNoIDs(t *testing.T) {
	t.Parallel()
	a := &models.Anime{}
	rpu := mkRPU(a)
	btns := rpu.buildButtons(false)
	assert.Empty(t, btns)
}

func TestBuildButtons_MaxTwo(t *testing.T) {
	t.Parallel()
	// All four IDs set → still limited to 2
	a := &models.Anime{IMDBID: "tt1", TMDBID: 1, MediaType: "movie", AnilistID: 2, MalID: 3}
	rpu := mkRPU(a)
	btns := rpu.buildButtons(true)
	assert.LessOrEqual(t, len(btns), 2)
}

func TestNewRichPresenceUpdater_FieldsSet(t *testing.T) {
	t.Parallel()
	a := &models.Anime{Name: "x"}
	paused := false
	mu := &sync.Mutex{}
	called := false
	mpv := func(_ string, _ []any) (any, error) { called = true; return nil, nil }
	rpu := NewRichPresenceUpdater(a, &paused, mu, 5*time.Second, 1500*time.Second, "mpv_test.sock", mpv)
	require.NotNil(t, rpu)
	assert.Same(t, a, rpu.GetAnime())
	assert.Same(t, &paused, rpu.GetIsPaused())
	assert.Same(t, mu, rpu.GetAnimeMutex())
	assert.Equal(t, 5*time.Second, rpu.GetUpdateFreq())
	assert.Equal(t, 1500*time.Second, rpu.GetEpisodeDuration())
	assert.Equal(t, "mpv_test.sock", rpu.GetSocketPath())
	assert.False(t, rpu.IsEpisodeStarted())
	assert.False(t, called)
}

func TestRichPresenceUpdater_SocketPathSetter(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	rpu.SetSocketPath("/x/y")
	assert.Equal(t, "/x/y", rpu.GetSocketPath())
}

func TestRichPresenceUpdater_EpisodeStartedSetter(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	assert.False(t, rpu.IsEpisodeStarted())
	rpu.SetEpisodeStarted(true)
	assert.True(t, rpu.IsEpisodeStarted())
}

func TestRichPresenceUpdater_EpisodeDurationSetter(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	rpu.SetEpisodeDuration(42 * time.Second)
	assert.Equal(t, 42*time.Second, rpu.GetEpisodeDuration())
}

func TestIsClientLoggedIn_DefaultFalse(t *testing.T) {
	// Singleton state — protect via clientMutex check; must not be parallel.
	clientMutex.Lock()
	current := isLoggedIn
	clientMutex.Unlock()
	if current {
		t.Skip("client globally logged in from previous test; cannot assert default state")
	}
	assert.False(t, IsClientLoggedIn())
}
