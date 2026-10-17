package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/pobreflix"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/startflix"
	"github.com/alvarorichard/Goanime/internal/util"
)

// Pobreflix flow. The catalog is plain HTTP — search over the sitemap corpus,
// title pages, seasons and Portuguese episode names — but its player is
// assembled client-side by an obfuscated bundle, so the stream URL is not in
// any HTML. Every page names the title's TMDB id, though, and the StartFlix
// panel is keyed by that same id — by IMDb id for movies, which the
// official-record lookup links keylessly — so a Pobreflix title plays through
// the panel's player hosts. From the panel on, the flow is StartFlix's
// (startflix.go): season picker, Dublado/Legendado, episode URLs that resolve
// on their own for batch downloads.

// pbClientFn is the Pobreflix client; a seam for tests.
var pbClientFn = pobreflix.Shared

// ErrPobreflixNoIMDb is returned when a Pobreflix movie's TMDB id links to no
// IMDb id, which the video panel keys movies by.
var ErrPobreflixNoIMDb = errors.New("no IMDb id is linked to this Pobreflix movie")

func isPobreflix(media *models.Anime) bool {
	return media != nil && media.Source == pobreflix.SourceName
}

// pobreflixPanel opens a Pobreflix title page and returns the panel its TMDB
// id names. The page is the authority on the id: enrichment may have guessed
// another title from the Portuguese name, so a disagreeing guess is dropped
// along with what it brought in (as adoptStartFlixIDs does).
func pobreflixPanel(ctx context.Context, media *models.Anime) (startflix.Panel, error) {
	title, err := pbClientFn().Title(ctx, media.URL)
	if err != nil {
		return startflix.Panel{}, err
	}
	if title.TMDBID != media.TMDBID {
		if media.TMDBID > 0 {
			util.Debug("Pobreflix TMDB id replaces a name-search match", "title", media.Name,
				"guessed_tmdb", media.TMDBID, "tmdb", title.TMDBID)
		}
		media.TMDBID, media.IMDBID = title.TMDBID, ""
		media.TMDBDetails, media.Rating, media.Overview, media.Genres, media.Runtime = nil, 0, "", nil, 0
	}
	if title.Kind == pobreflix.KindSeries {
		return startflix.SeriesPanel(title.TMDBID), nil
	}
	if media.IMDBID == "" {
		err := sfxEnrichByIDFn(media, true)
		if media.IMDBID == "" {
			if err != nil {
				return startflix.Panel{}, fmt.Errorf("%w (tmdb %d): %w", ErrPobreflixNoIMDb, title.TMDBID, err)
			}
			return startflix.Panel{}, fmt.Errorf("%w (tmdb %d)", ErrPobreflixNoIMDb, title.TMDBID)
		}
	}
	return startflix.MoviePanel(media.IMDBID), nil
}

// describePobreflixErr turns a Pobreflix failure into a plain-language
// message, keeping the cause reachable through Unwrap.
func describePobreflixErr(media *models.Anime, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrPobreflixNoIMDb):
		return &friendlyError{cause: err, msg: "⚠️  GoAnime couldn't link this Pobreflix movie to the servers it plays from. Try searching it on StartFlix."}
	case errors.Is(err, startflix.ErrNotOnPanel), errors.Is(err, startflix.ErrNoPanel):
		return &friendlyError{cause: err, msg: "⚠️  Pobreflix has this title only on servers that don't carry it. Try searching it on StartFlix or another source."}
	default:
		return describePanelErr(pobreflix.SourceName, err)
	}
}

// fillPobreflixEpisodeTitles names the episodes the panel lists without a
// name, from Pobreflix's own episode list. Best effort: the title page is
// already cached in the client, and the page lists the rendered season's
// episodes with Portuguese names.
func fillPobreflixEpisodeTitles(media *models.Anime, episodes []models.Episode) {
	if len(episodes) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), tcCheckBudget)
	defer cancel()
	title, err := pbClientFn().Title(ctx, media.URL)
	if err != nil {
		return
	}
	for i := range episodes {
		if episodes[i].Title.English != "" {
			continue
		}
		if name := title.EpisodeTitle(media.CurrentSeason, episodes[i].Num); name != "" {
			episodes[i].Title.English, episodes[i].Title.Romaji = name, name
		}
	}
}

// GetPobreflixEpisodes lists a Pobreflix title: the movie itself, or the
// episodes of the season and audio the user picks.
func GetPobreflixEpisodes(media *models.Anime) ([]models.Episode, error) {
	if media == nil || media.URL == "" {
		return nil, fmt.Errorf("no Pobreflix page for this title")
	}
	return GetStartFlixEpisodes(media)
}

// GetPobreflixSeasonNumbers lists a series' season numbers, for downloads that
// name the seasons rather than pick one.
func GetPobreflixSeasonNumbers(media *models.Anime) ([]int, error) {
	return GetStartFlixSeasonNumbers(media)
}

// GetPobreflixSeasonEpisodes lists one season of a series without the season
// picker; see GetStartFlixSeasonEpisodes.
func GetPobreflixSeasonEpisodes(media *models.Anime, seasonNum int) ([]models.Episode, error) {
	return GetStartFlixSeasonEpisodes(media, seasonNum)
}

// GetPobreflixStreamURL resolves an episode or movie to a playable URL and
// hands mpv/the downloader the referer, subtitles and audio preference.
func GetPobreflixStreamURL(media *models.Anime, episode *models.Episode, quality string) (string, error) {
	return GetStartFlixStreamURL(media, episode, quality)
}
