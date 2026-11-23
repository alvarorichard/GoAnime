package api

import (
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/alvarorichard/Goanime/v2/internal/api/movie"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// Some catalogs name no release year (PlutoTV lists series with an id, a slug
// and a name), while others list a remake or revival under the same title with
// its year: iCarly (2007) and iCarly (2021) from TopCine and StartFlix, and a
// yearless iCarly from PlutoTV. The results screen groups a work by title,
// kind and year, and a copy with no year fits every group of its title, so the
// yearless iCarly landed on whichever came first — the 2021 revival, when it
// is the 2007 original that PlutoTV carries.
//
// This asks IMDb, which needs no key, what the title is, and gives a yearless
// copy the year of the original: the earliest IMDb knows for that exact title
// and kind. A catalog with no year is a library of old shows far more often
// than a revival's. It runs only when the results already disagree about the
// year (two or more years for one title and kind) and a copy has none, so a
// search for anything else costs nothing; and a failed lookup leaves the copy
// as it was.

const (
	// ambiguousYearBudget bounds the IMDb lookups of one search.
	ambiguousYearBudget = 6 * time.Second
	// maxAmbiguousLookups caps the titles one search asks IMDb about.
	maxAmbiguousLookups = 3
)

// imdbTitlesFn asks IMDb for the titles named like a query; a seam for tests.
var imdbTitlesFn = func(name string) ([]movie.IMDbTitle, error) {
	return movie.NewIMDbClient().SearchTitles(name)
}

var (
	imdbTitlesMu    sync.Mutex
	imdbTitlesCache = map[string][]movie.IMDbTitle{}
)

// cachedIMDbTitles is imdbTitlesFn, remembered per name for the process: the
// search retry loop searches the same words again and again.
func cachedIMDbTitles(name string) ([]movie.IMDbTitle, error) {
	key := yearTitleKey(name)
	imdbTitlesMu.Lock()
	if titles, ok := imdbTitlesCache[key]; ok {
		imdbTitlesMu.Unlock()
		return titles, nil
	}
	imdbTitlesMu.Unlock()
	titles, err := imdbTitlesFn(name)
	if err != nil {
		return nil, err
	}
	imdbTitlesMu.Lock()
	imdbTitlesCache[key] = titles
	imdbTitlesMu.Unlock()
	return titles, nil
}

// yearTitleKey folds a title into what two catalogs' spellings share: tags,
// a trailing year and audio words off, lower case, letters and digits only.
func yearTitleKey(name string) string {
	var sb strings.Builder
	for _, r := range strings.ToLower(yearLookupName(name)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// yearLookupName is a result's title as IMDb would spell it, without the
// scraper's tags, year and audio words.
func yearLookupName(name string) string {
	name = movie.CleanMediaName(name)
	for _, audio := range []string{"(dublado)", "(legendado)"} {
		if i := strings.Index(strings.ToLower(name), audio); i >= 0 {
			name = name[:i] + name[i+len(audio):]
		}
	}
	return strings.Join(strings.Fields(name), " ")
}

// imdbKindsFor is the IMDb type ids that count as the same kind of title.
func imdbKindsFor(mediaType models.MediaType) map[string]bool {
	if mediaType == models.MediaTypeMovie {
		return map[string]bool{"movie": true, "tvMovie": true}
	}
	return map[string]bool{"tvSeries": true, "tvMiniSeries": true}
}

// yearGroup is the results that share a title and a kind.
type yearGroup struct {
	mediaType models.MediaType
	name      string // a cleaned title to ask IMDb with
	years     map[string]bool
	yearless  []*models.Anime
}

// resolveAmbiguousYears gives each yearless movie or series the year of the
// original, when the other results name more than one year for its title.
func resolveAmbiguousYears(results []*models.Anime) {
	groups := map[string]*yearGroup{}
	var order []string
	for _, r := range results {
		if r == nil || (r.MediaType != models.MediaTypeMovie && r.MediaType != models.MediaTypeTV) {
			continue
		}
		key := string(r.MediaType) + "|" + yearTitleKey(r.Name)
		if yearTitleKey(r.Name) == "" {
			continue
		}
		g := groups[key]
		if g == nil {
			g = &yearGroup{mediaType: r.MediaType, name: yearLookupName(r.Name), years: map[string]bool{}}
			groups[key] = g
			order = append(order, key)
		}
		if r.Year != "" {
			g.years[r.Year] = true
		} else {
			g.yearless = append(g.yearless, r)
		}
	}

	var todo []*yearGroup
	for _, key := range order {
		if g := groups[key]; len(g.yearless) > 0 && len(g.years) >= 2 {
			todo = append(todo, g)
		}
	}
	if len(todo) > maxAmbiguousLookups {
		todo = todo[:maxAmbiguousLookups]
	}
	if len(todo) == 0 {
		return
	}

	type answer struct {
		titles []movie.IMDbTitle
		err    error
	}
	answers := make([]chan answer, len(todo))
	for i, g := range todo {
		answers[i] = make(chan answer, 1)
		go func() {
			titles, err := cachedIMDbTitles(g.name)
			answers[i] <- answer{titles, err}
		}()
	}
	deadline := time.After(ambiguousYearBudget)
	for i, g := range todo {
		select {
		case a := <-answers[i]:
			if a.err != nil {
				util.Debug("IMDb year lookup failed", "title", g.name, "err", a.err)
				continue
			}
			year := originalYear(a.titles, g)
			if year == "" {
				util.Debug("IMDb knows no original for this title", "title", g.name)
				continue
			}
			for _, r := range g.yearless {
				util.Debug("A yearless copy takes its title's original year", "title", r.Name, "source", r.Source, "year", year)
				r.Year = year
			}
		case <-deadline:
			util.Debug("IMDb year lookups ran out of time", "title", g.name)
			return
		}
	}
}

// originalYear is the earliest year IMDb lists for exactly this title and
// kind, as text, or "".
func originalYear(titles []movie.IMDbTitle, g *yearGroup) string {
	want := yearTitleKey(g.name)
	kinds := imdbKindsFor(g.mediaType)
	best := 0
	for _, t := range titles {
		if t.Year <= 0 || !kinds[t.Kind] || yearTitleKey(t.Title) != want {
			continue
		}
		if best == 0 || t.Year < best {
			best = t.Year
		}
	}
	if best == 0 {
		return ""
	}
	return (&movie.IMDbTitle{Year: best}).YearString()
}
