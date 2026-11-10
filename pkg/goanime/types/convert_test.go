package types

import (
	"fmt"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fullInternalAnime() *models.Anime {
	return &models.Anime{
		Name: "Naruto", URL: "https://x/naruto", ImageURL: "https://x/n.jpg", Source: "Animefire.io",
		AnilistID: 20, MalID: 21,
		Episodes: []models.Episode{{Number: "1", Num: 1, URL: "e1"}, {Number: "2", Num: 2, URL: "e2"}},
		Details: models.AniListDetails{
			ID: 20, IDMal: 21, Description: "ninja", Genres: []string{"Action", "Adventure"},
			AverageScore: 79, Episodes: 220, Status: "FINISHED",
			Title:      models.Title{Romaji: "Naruto", English: "Naruto EN"},
			CoverImage: models.CoverImages{Large: "L", Medium: "M"},
		},
	}
}

func TestFromInternalAnimeNil(t *testing.T) {
	assert.Nil(t, FromInternalAnime(nil))
	assert.Nil(t, FromInternalEpisode(nil))
}

func TestFromInternalAnimeCopiesEveryField(t *testing.T) {
	a := FromInternalAnime(fullInternalAnime())
	require.NotNil(t, a)
	assert.Equal(t, "Naruto", a.Name)
	assert.Equal(t, "https://x/naruto", a.URL)
	assert.Equal(t, "https://x/n.jpg", a.ImageURL)
	assert.Equal(t, "Animefire.io", a.Source)
	assert.Equal(t, 20, a.AnilistID)
	assert.Equal(t, 21, a.MalID)
	require.Len(t, a.Episodes, 2)
	assert.Equal(t, "e2", a.Episodes[1].URL)

	d := a.Details
	require.NotNil(t, d)
	assert.Equal(t, 20, d.ID)
	assert.Equal(t, 21, d.IDMal)
	assert.Equal(t, "ninja", d.Description)
	assert.Equal(t, []string{"Action", "Adventure"}, d.Genres)
	assert.Equal(t, 79, d.AverageScore)
	assert.Equal(t, 220, d.Episodes)
	assert.Equal(t, "FINISHED", d.Status)
	assert.Equal(t, &Title{Romaji: "Naruto", English: "Naruto EN"}, d.Title)
	assert.Equal(t, &CoverImages{Large: "L", Medium: "M"}, d.CoverImage)
}

func TestFromInternalAnimeOptionalSections(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*models.Anime)
		wantTitle bool
		wantCover bool
		wantEps   bool
	}{
		{"all present", func(*models.Anime) {}, true, true, true},
		{"no episodes", func(a *models.Anime) { a.Episodes = nil }, true, true, false},
		{"no title", func(a *models.Anime) { a.Details.Title = models.Title{} }, false, true, true},
		{"romaji only", func(a *models.Anime) { a.Details.Title = models.Title{Romaji: "R"} }, true, true, true},
		{"english only", func(a *models.Anime) { a.Details.Title = models.Title{English: "E"} }, true, true, true},
		{"native only is not a title", func(a *models.Anime) { a.Details.Title = models.Title{Native: "ナ"} }, false, true, true},
		{"no cover", func(a *models.Anime) { a.Details.CoverImage = models.CoverImages{} }, true, false, true},
		{"large cover only", func(a *models.Anime) { a.Details.CoverImage = models.CoverImages{Large: "L"} }, true, true, true},
		{"medium cover only", func(a *models.Anime) { a.Details.CoverImage = models.CoverImages{Medium: "M"} }, true, true, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := fullInternalAnime()
			tc.mutate(in)
			out := FromInternalAnime(in)
			assert.Equal(t, tc.wantTitle, out.Details.Title != nil, "title")
			assert.Equal(t, tc.wantCover, out.Details.CoverImage != nil, "cover")
			assert.Equal(t, tc.wantEps, len(out.Episodes) > 0, "episodes")
			assert.NotNil(t, out.Details, "details are always present")
		})
	}
}

func TestFromInternalAnimeListKeepsOrderAndNils(t *testing.T) {
	assert.Empty(t, FromInternalAnimeList(nil))
	assert.NotNil(t, FromInternalAnimeList(nil))

	in := []*models.Anime{{Name: "A"}, nil, {Name: "C"}}
	out := FromInternalAnimeList(in)
	require.Len(t, out, 3)
	assert.Equal(t, "A", out[0].Name)
	assert.Nil(t, out[1], "a nil entry stays nil, in place")
	assert.Equal(t, "C", out[2].Name)
}

func TestFromInternalEpisodeCopiesEveryField(t *testing.T) {
	in := &models.Episode{
		Number: "1.5", Num: 2, URL: "u", Aired: "2020-01-02", Duration: 1400,
		IsFiller: true, IsRecap: true, Synopsis: "s",
		Title:     models.TitleDetails{Romaji: "r", English: "e", Japanese: "j"},
		SkipTimes: models.SkipTimes{Op: models.Skip{Start: 10, End: 90}, Ed: models.Skip{Start: 1300, End: 1400}},
	}
	e := FromInternalEpisode(in)
	require.NotNil(t, e)
	assert.Equal(t, "1.5", e.Number)
	assert.Equal(t, 2, e.Num)
	assert.Equal(t, "u", e.URL)
	assert.Equal(t, "2020-01-02", e.Aired)
	assert.Equal(t, 1400, e.Duration)
	assert.True(t, e.IsFiller)
	assert.True(t, e.IsRecap)
	assert.Equal(t, "s", e.Synopsis)
	assert.Equal(t, &TitleDetails{Romaji: "r", English: "e", Japanese: "j"}, e.Title)
	assert.Equal(t, &SkipTimes{Op: &SkipTime{Start: 10, End: 90}, Ed: &SkipTime{Start: 1300, End: 1400}}, e.SkipTimes)
}

func TestFromInternalEpisodeFlagsAreIndependent(t *testing.T) {
	e := FromInternalEpisode(&models.Episode{IsFiller: true})
	assert.True(t, e.IsFiller)
	assert.False(t, e.IsRecap)
	e = FromInternalEpisode(&models.Episode{IsRecap: true})
	assert.False(t, e.IsFiller)
	assert.True(t, e.IsRecap)
}

func TestFromInternalEpisodeTitleAppearsWithAnyLanguage(t *testing.T) {
	for name, title := range map[string]models.TitleDetails{
		"romaji": {Romaji: "r"}, "english": {English: "e"}, "japanese": {Japanese: "j"},
	} {
		e := FromInternalEpisode(&models.Episode{Title: title})
		assert.NotNil(t, e.Title, name)
	}
	assert.Nil(t, FromInternalEpisode(&models.Episode{}).Title, "no title at all")
}

func TestFromInternalEpisodeSkipTimesAppearWithAnyNonZeroBoundary(t *testing.T) {
	for name, st := range map[string]models.SkipTimes{
		"op start": {Op: models.Skip{Start: 1}},
		"op end":   {Op: models.Skip{End: 1}},
		"ed start": {Ed: models.Skip{Start: 1}},
		"ed end":   {Ed: models.Skip{End: 1}},
	} {
		e := FromInternalEpisode(&models.Episode{SkipTimes: st})
		require.NotNil(t, e.SkipTimes, name)
		require.NotNil(t, e.SkipTimes.Op, name)
		require.NotNil(t, e.SkipTimes.Ed, name)
	}
	assert.Nil(t, FromInternalEpisode(&models.Episode{}).SkipTimes, "all zero means no skip data")
}

func TestFromInternalEpisodeListKeepsOrderAndDoesNotAlias(t *testing.T) {
	in := []models.Episode{{Number: "1", URL: "a"}, {Number: "2", URL: "b"}, {Number: "3", URL: "c"}}
	out := FromInternalEpisodeList(in)
	require.Len(t, out, 3)
	for i, want := range []string{"a", "b", "c"} {
		assert.Equal(t, want, out[i].URL)
	}
	assert.NotSame(t, out[0], out[1], "each entry is its own value (loop variable aliasing)")
	in[0].URL = "changed"
	assert.Equal(t, "a", out[0].URL, "the copy is independent of the source slice")
	assert.Empty(t, FromInternalEpisodeList(nil))
}

func TestSourceToScraperTypeAndString(t *testing.T) {
	assert.Equal(t, scraper.AnimefireType, SourceAnimeFire.ToScraperType())
	assert.Equal(t, scraper.AnimefireType, Source(99).ToScraperType(), "unknown falls back to the default source")
	assert.Equal(t, "Animefire.io", SourceAnimeFire.String())
	assert.Equal(t, "Unknown", Source(99).String())
}

func TestParseSourceAllSpellings(t *testing.T) {
	for _, in := range []string{"AnimeFire", "animefire", "fire", "Animefire.io", "animefire.io"} {
		got, err := ParseSource(in)
		require.NoError(t, err, in)
		assert.Equal(t, SourceAnimeFire, got, in)
	}
	for in, want := range map[string]Source{
		"Goyabu": SourceGoyabu, "goyabu": SourceGoyabu, "hianime": SourceHiAnime,
		"StartFlix": SourceStartFlix, "TopCine": SourceTopCine,
		"animesdigital": SourceAnimesDigital, "AnimeDrive": SourceAnimeDrive,
		"Pobreflix": SourcePobreflix, "AnimesROLL": SourceAnroll, "anroll": SourceAnroll,
		"Pomfy": SourcePomfy,
	} {
		got, err := ParseSource(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}
	for _, in := range []string{"", " ", "ANIMEFIRE", "allanime", "Fire "} {
		_, err := ParseSource(in)
		require.Error(t, err, "%q", in)
		assert.Contains(t, err.Error(), in)
	}
}

// The round trip clients depend on: the label SearchAnime stamps on results
// (Source.String) must parse back to the same source.
func TestSourceStringRoundTripsThroughParse(t *testing.T) {
	for _, src := range Sources() {
		got, err := ParseSource(src.String())
		require.NoError(t, err, src.String())
		assert.Equal(t, src, got)
	}
}

// Every source reaches its own scraper, and no two share one: a table row
// copied from its neighbour would make a source silently serve another's data.
func TestSourcesMapToDistinctScrapers(t *testing.T) {
	seen := map[scraper.ScraperType]Source{}
	names := map[string]Source{}
	for _, src := range Sources() {
		st := src.ToScraperType()
		if prev, dup := seen[st]; dup {
			t.Errorf("%v and %v both map to scraper type %v", prev, src, st)
		}
		seen[st] = src
		if prev, dup := names[src.String()]; dup {
			t.Errorf("%v and %v share the name %q", prev, src, src.String())
		}
		names[src.String()] = src
		assert.NotEqual(t, "Unknown", src.String())
	}
	assert.Equal(t, scraper.GoyabuType, SourceGoyabu.ToScraperType())
	assert.Equal(t, scraper.PomfyType, SourcePomfy.ToScraperType())
	assert.Equal(t, scraper.AnrollType, SourceAnroll.ToScraperType())
}

func TestSourceSupportsEpisodes(t *testing.T) {
	for _, src := range []Source{SourceAnimeFire, SourceGoyabu, SourceHiAnime, SourceStartFlix, SourceAnimesDigital, SourceAnimeDrive, SourceAnroll} {
		assert.True(t, src.SupportsEpisodes(), src.String())
	}
	for _, src := range []Source{SourceTopCine, SourcePobreflix, SourcePomfy, Source(-1), Source(99)} {
		assert.False(t, src.SupportsEpisodes(), src.String())
	}
}

func TestSourceIsCatalog(t *testing.T) {
	for _, src := range Sources() {
		assert.Equal(t, src == SourceStartFlix, src.IsCatalog(), src.String())
		assert.False(t, src.IsCatalog() && !src.SupportsEpisodes(), "%s: a catalogue lists episodes", src)
	}
	assert.False(t, Source(-1).IsCatalog())
	assert.False(t, Source(99).IsCatalog())
}

// Values are stored by callers: they must not move.
func TestSourceValuesAreStable(t *testing.T) {
	assert.Equal(t, 0, int(SourceAnimeFire))
	assert.Equal(t, 1, int(SourceGoyabu))
	assert.Equal(t, 9, int(SourcePomfy))
	assert.Len(t, Sources(), 10)
	assert.Equal(t, "Unknown", Source(-1).String())
}

// ---- metatests --------------------------------------------------------------

// animeConversionFailures judges any converter against the full-field contract.
func animeConversionFailures(rawConv func(*models.Anime) *Anime) (f []string) {
	// A converter that panics on some input is itself a failure of the contract.
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprintf("panic: %v", r))
		}
	}()
	conv := rawConv
	if conv(nil) != nil {
		f = append(f, "nil not preserved")
	}
	a := conv(fullInternalAnime())
	if a == nil || a.Details == nil {
		return append(f, "nil result")
	}
	for name, ok := range map[string]bool{
		"name": a.Name == "Naruto", "url": a.URL == "https://x/naruto", "image": a.ImageURL == "https://x/n.jpg",
		"source": a.Source == "Animefire.io", "anilist": a.AnilistID == 20, "mal": a.MalID == 21,
		"episodes":   len(a.Episodes) == 2 && a.Episodes[1].URL == "e2",
		"details id": a.Details.ID == 20, "details mal": a.Details.IDMal == 21,
		"description": a.Details.Description == "ninja", "genres": fmt.Sprint(a.Details.Genres) == "[Action Adventure]",
		"score": a.Details.AverageScore == 79, "eps": a.Details.Episodes == 220, "status": a.Details.Status == "FINISHED",
		"title":  a.Details.Title != nil && a.Details.Title.Romaji == "Naruto" && a.Details.Title.English == "Naruto EN",
		"covers": a.Details.CoverImage != nil && a.Details.CoverImage.Large == "L" && a.Details.CoverImage.Medium == "M",
	} {
		if !ok {
			f = append(f, name+" wrong")
		}
	}
	noTitle := fullInternalAnime()
	noTitle.Details.Title = models.Title{}
	if r := conv(noTitle); r.Details.Title != nil {
		f = append(f, "empty title materialised")
	}
	noCover := fullInternalAnime()
	noCover.Details.CoverImage = models.CoverImages{}
	if r := conv(noCover); r.Details.CoverImage != nil {
		f = append(f, "empty cover materialised")
	}
	return f
}

func TestFromInternalAnimeMeta(t *testing.T) {
	require.Empty(t, animeConversionFailures(FromInternalAnime))

	with := func(mut func(*Anime)) func(*models.Anime) *Anime {
		return func(in *models.Anime) *Anime {
			a := FromInternalAnime(in)
			if a != nil {
				mut(a)
			}
			return a
		}
	}
	mutants := map[string]func(*models.Anime) *Anime{
		"name lost":        with(func(a *Anime) { a.Name = "" }),
		"url lost":         with(func(a *Anime) { a.URL = "" }),
		"image lost":       with(func(a *Anime) { a.ImageURL = "" }),
		"source lost":      with(func(a *Anime) { a.Source = "" }),
		"anilist lost":     with(func(a *Anime) { a.AnilistID = 0 }),
		"mal swapped":      with(func(a *Anime) { a.MalID = a.AnilistID }),
		"episodes dropped": with(func(a *Anime) { a.Episodes = nil }),
		"episode order":    with(func(a *Anime) { a.Episodes[0], a.Episodes[1] = a.Episodes[1], a.Episodes[0] }),
		"details dropped":  with(func(a *Anime) { a.Details.Description = "" }),
		"genres dropped":   with(func(a *Anime) { a.Details.Genres = nil }),
		"score lost":       with(func(a *Anime) { a.Details.AverageScore = 0 }),
		"status lost":      with(func(a *Anime) { a.Details.Status = "" }),
		"title always":     with(func(a *Anime) { a.Details.Title = &Title{} }),
		"title swapped": with(func(a *Anime) {
			a.Details.Title.Romaji, a.Details.Title.English = a.Details.Title.English, a.Details.Title.Romaji
		}),
		"cover always":      with(func(a *Anime) { a.Details.CoverImage = &CoverImages{} }),
		"cover medium lost": with(func(a *Anime) { a.Details.CoverImage.Medium = "" }),
		"nil becomes empty": func(in *models.Anime) *Anime {
			if in == nil {
				return &Anime{}
			}
			return FromInternalAnime(in)
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, animeConversionFailures(m), "mutant %q survived", name)
	}
}

func TestFromInternalAnimeEpisodeCountBoundaries(t *testing.T) {
	none := fullInternalAnime()
	none.Episodes = nil
	assert.Nil(t, FromInternalAnime(none).Episodes, "no episodes stays nil, not an empty slice")

	one := fullInternalAnime()
	one.Episodes = []models.Episode{{Number: "1", URL: "only"}}
	got := FromInternalAnime(one).Episodes
	require.Len(t, got, 1, "a single episode is converted")
	assert.Equal(t, "only", got[0].URL)
}
