// Package movie — the official record of a title whose ids are known.
package movie

import (
	"errors"
	"fmt"
	"sync"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// Seams, swapped by tests so the lookup runs without network.
var (
	tmdbClientFn     = func() tmdbAPI { return NewTMDBClient() }
	imdbClientFn     = func() imdbAPI { return NewIMDbClient() }
	wikidataClientFn = func() wikidataAPI { return NewWikidataClient() }
)

type tmdbAPI interface {
	IsConfigured() bool
	FindByIMDBID(imdbID string) (*models.TMDBMedia, error)
	GetMovieDetails(movieID int) (*models.TMDBDetails, error)
	GetTVDetails(tvID int) (*models.TMDBDetails, error)
}

type imdbAPI interface {
	TitleByID(imdbID string) (*IMDbTitle, error)
}

type wikidataAPI interface {
	Lookup(tmdbID int, imdbID string, movie bool) (*WikidataRecord, error)
}

// officialByID caches what each id lookup settled, failures included: a
// title's episodes each pass through here, and the answer for a given pair
// of ids does not change within a session.
var officialByID sync.Map // officialKey -> officialResult

type officialKey struct {
	movie  bool
	tmdbID int
	imdbID string
}

type officialResult struct {
	details *models.TMDBDetails
	err     error
}

// ErrNoOfficialRecord means no metadata source had a record for the ids.
var ErrNoOfficialRecord = errors.New("no official record for this title")

// EnrichByID fills media with the official English record of the movie or
// show its ids name — TMDB id, IMDb id, or both — replacing whatever a search
// by a localized name put there. This is what names downloads the way media
// servers expect: "Heart of the Beast (2026)", not the Brazilian "Coração
// Selvagem".
//
// Nothing here needs an API key or an account. IMDb's own title lookup names
// the title by its IMDb id; Wikidata links a TMDB id to its IMDb id (StartFlix
// names shows only by TMDB id), fills in the TMDB id a movie's folder carries,
// and stands in with its English label if IMDb does not answer. TMDB is asked
// first only when a TMDB_API_KEY happens to be configured. Best effort: on
// error media is left as it was.
func EnrichByID(media *models.Media, movie bool) error {
	if media == nil || (media.TMDBID <= 0 && media.IMDBID == "") {
		return ErrNoOfficialRecord
	}
	key := officialKey{movie: movie, tmdbID: media.TMDBID, imdbID: media.IMDBID}
	if v, ok := officialByID.Load(key); ok {
		r := v.(officialResult)
		if r.err != nil {
			return r.err
		}
		applyOfficial(media, r.details, movie)
		return nil
	}
	details, err := lookupOfficial(media.TMDBID, media.IMDBID, movie)
	officialByID.Store(key, officialResult{details: details, err: err})
	if err != nil {
		return err
	}
	applyOfficial(media, details, movie)
	util.Debug("Official record by id", "title", officialTitle(details), "year", officialYear(details),
		"tmdb", details.ID, "imdb", details.IMDBID)
	return nil
}

// lookupOfficial finds the record by id: TMDB's when a key is configured,
// otherwise IMDb's title with Wikidata linking the ids.
func lookupOfficial(tmdbID int, imdbID string, movie bool) (*models.TMDBDetails, error) {
	var wd *WikidataRecord
	wikidata := func() *WikidataRecord {
		if wd == nil {
			rec, err := wikidataClientFn().Lookup(tmdbID, imdbID, movie)
			if err != nil {
				util.Debug("Wikidata lookup failed", "tmdb", tmdbID, "imdb", imdbID, "err", err)
				rec = &WikidataRecord{}
			}
			wd = rec
		}
		return wd
	}

	if d := lookupTMDB(tmdbID, imdbID, movie); d != nil {
		if d.IMDBID == "" { // TV details carry no imdb_id
			d.IMDBID = imdbID
			if d.IMDBID == "" {
				d.IMDBID = wikidata().IMDBID
			}
		}
		return d, nil
	}

	if imdbID == "" || tmdbID <= 0 {
		rec := wikidata()
		if imdbID == "" {
			imdbID = rec.IMDBID
		}
		if tmdbID <= 0 {
			tmdbID = rec.TMDBID
		}
	}

	var title, year string
	if imdbID != "" {
		if t, err := imdbClientFn().TitleByID(imdbID); err == nil {
			title, year = t.Title, t.YearString()
		} else {
			util.Debug("IMDb title lookup failed, trying Wikidata's label", "imdb", imdbID, "err", err)
		}
	}
	if title == "" {
		title = wikidata().Label
	}
	if year == "" {
		year = wikidata().Year
	}
	if title == "" {
		return nil, fmt.Errorf("%w (tmdb %d, imdb %q)", ErrNoOfficialRecord, tmdbID, imdbID)
	}

	// Only the year of the release date is known, which is all its readers
	// (OfficialTitle, year extraction) take.
	d := &models.TMDBDetails{ID: tmdbID, IMDBID: imdbID}
	if movie {
		d.Title, d.ReleaseDate = title, year
	} else {
		d.Name, d.FirstAirDate = title, year
	}
	return d, nil
}

// lookupTMDB returns TMDB's record when a TMDB key is configured, else nil.
func lookupTMDB(tmdbID int, imdbID string, movie bool) *models.TMDBDetails {
	tmdb := tmdbClientFn()
	if !tmdb.IsConfigured() {
		return nil
	}
	if tmdbID <= 0 && imdbID != "" {
		if hit, err := tmdb.FindByIMDBID(imdbID); err == nil && hit.ID > 0 && (hit.MediaType == "movie") == movie {
			tmdbID = hit.ID
		}
	}
	if tmdbID <= 0 {
		return nil
	}
	var d *models.TMDBDetails
	var err error
	if movie {
		d, err = tmdb.GetMovieDetails(tmdbID)
	} else {
		d, err = tmdb.GetTVDetails(tmdbID)
	}
	if err != nil || d == nil || officialTitle(d) == "" {
		util.Debug("TMDB details by id failed, using IMDb", "tmdb", tmdbID, "err", err)
		return nil
	}
	if d.ID <= 0 {
		d.ID = tmdbID
	}
	return d
}

// applyOfficial writes the record onto media. The record is authoritative, so
// it replaces fields a name search may have filled with another title's data.
func applyOfficial(media *models.Media, d *models.TMDBDetails, movie bool) {
	cp := *d
	media.TMDBDetails = &cp
	if d.ID > 0 {
		media.TMDBID = d.ID
	}
	if d.IMDBID != "" {
		media.IMDBID = d.IMDBID
	}
	if y := officialYear(d); y != "" {
		media.Year = y
	}
	if d.Overview != "" {
		media.Overview = d.Overview
	}
	if d.VoteAverage > 0 {
		media.Rating = d.VoteAverage
	}
	if movie && d.Runtime > 0 {
		media.Runtime = d.Runtime
	}
	if len(d.Genres) > 0 {
		genres := make([]string, 0, len(d.Genres))
		for _, g := range d.Genres {
			genres = append(genres, g.Name)
		}
		media.Genres = genres
	}
}

func officialTitle(d *models.TMDBDetails) string {
	if d.Title != "" {
		return d.Title
	}
	return d.Name
}

func officialYear(d *models.TMDBDetails) string {
	for _, date := range []string{d.ReleaseDate, d.FirstAirDate} {
		if y := leadingYear(date); y != "" {
			return y
		}
	}
	return ""
}

// leadingYear returns the leading four-digit year of s: "2026", "2023–",
// "2011–2019" (series runs) and "2026-09-25T00:00:00Z" all start with it.
func leadingYear(s string) string {
	if len(s) < 4 {
		return ""
	}
	for i := range 4 {
		if s[i] < '0' || s[i] > '9' {
			return ""
		}
	}
	return s[:4]
}
