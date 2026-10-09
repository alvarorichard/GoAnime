package api

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/startflix"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/topcine"
	"github.com/alvarorichard/Goanime/internal/util"
)

// TopCine flow. The catalog is plain HTTP, but the one host its player serves
// video from (1take.top) answers only after a Cloudflare Turnstile, which
// GoAnime does not solve (checked 2026-10-09). Every TopCine page names its
// title's TMDB id, though, and the StartFlix panel is keyed by that same id —
// by IMDb id for movies, which the official-record lookup links keylessly — so
// a TopCine title plays through the panel's player hosts. From the panel on,
// the flow is StartFlix's (startflix.go): season picker, Dublado/Legendado,
// episode URLs that resolve on their own for batch downloads.

// tcClientFn is the TopCine client; a seam for tests.
var tcClientFn = topcine.Shared

// tcCheckBudget bounds the player check made only to explain a failure.
const tcCheckBudget = 8 * time.Second

// ErrTopCineNoIMDb is returned when a TopCine movie's TMDB id links to no IMDb
// id, which the video panel keys movies by.
var ErrTopCineNoIMDb = errors.New("no IMDb id is linked to this TopCine movie")

func isTopCine(media *models.Anime) bool {
	return media != nil && media.Source == topcine.SourceName
}

// topCinePanel opens a TopCine title page and returns the panel its TMDB id
// names. The page is the authority on the id: enrichment may have guessed
// another title from the Portuguese name, so a disagreeing guess is dropped
// along with what it brought in (as adoptStartFlixIDs does).
func topCinePanel(ctx context.Context, media *models.Anime) (startflix.Panel, error) {
	title, err := tcClientFn().Title(ctx, media.URL)
	if err != nil {
		return startflix.Panel{}, err
	}
	if title.TMDBID != media.TMDBID {
		if media.TMDBID > 0 {
			util.Debug("TopCine TMDB id replaces a name-search match", "title", media.Name,
				"guessed_tmdb", media.TMDBID, "tmdb", title.TMDBID)
		}
		media.TMDBID, media.IMDBID = title.TMDBID, ""
		media.TMDBDetails, media.Rating, media.Overview, media.Genres, media.Runtime = nil, 0, "", nil, 0
	}
	if title.Kind == topcine.KindSeries {
		return startflix.SeriesPanel(title.TMDBID), nil
	}
	if media.IMDBID == "" {
		err := sfxEnrichByIDFn(media, true)
		if media.IMDBID == "" {
			if err != nil {
				return startflix.Panel{}, fmt.Errorf("%w (tmdb %d): %w", ErrTopCineNoIMDb, title.TMDBID, err)
			}
			return startflix.Panel{}, fmt.Errorf("%w (tmdb %d)", ErrTopCineNoIMDb, title.TMDBID)
		}
	}
	return startflix.MoviePanel(media.IMDBID), nil
}

// describeTopCineErr turns a TopCine failure into a plain-language message,
// keeping the cause reachable through Unwrap.
func describeTopCineErr(media *models.Anime, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, topcine.ErrNoPlayer):
		return &friendlyError{cause: err, msg: "⚠️  TopCine has no player for this title yet. Try searching it on another source."}
	case errors.Is(err, ErrTopCineNoIMDb):
		return &friendlyError{cause: err, msg: "⚠️  GoAnime couldn't link this TopCine movie to the servers it plays from. Try searching it on StartFlix."}
	case errors.Is(err, startflix.ErrNotOnPanel), errors.Is(err, startflix.ErrNoPanel):
		return &friendlyError{cause: err, msg: topCineUnplayableMessage(media)}
	default:
		return describePanelErr(topcine.SourceName, err)
	}
}

// tcLanguagesFn asks a TopCine player which tracks it has; a seam for tests.
var tcLanguagesFn = func(ctx context.Context, playerURL string) (topcine.Languages, error) {
	return tcClientFn().Languages(ctx, playerURL)
}

// topCineUnplayableMessage explains a title the panel does not carry. It asks
// TopCine's own player whether it has the video at all, so the user learns
// whether waiting can help.
func topCineUnplayableMessage(media *models.Anime) string {
	const fallback = "⚠️  TopCine has this title only on a server that asks for a browser check GoAnime can't pass. Try searching it on another source."
	var langs topcine.Languages
	var err error
	runWithSpinner("Checking TopCine's player...", func() {
		ctx, cancel := context.WithTimeout(context.Background(), tcCheckBudget)
		defer cancel()
		var title topcine.Title
		if title, err = tcClientFn().Title(ctx, media.URL); err != nil {
			return
		}
		player := topCineAnyPlayer(title, media.CurrentSeason)
		if player == "" {
			err = topcine.ErrNoPlayer
			return
		}
		langs, err = tcLanguagesFn(ctx, player)
	})
	switch {
	case err != nil:
		util.Debug("TopCine player check failed", "title", media.Name, "err", err)
		return fallback
	case !langs.Dubbed && !langs.Subtitled:
		return "⚠️  TopCine has no video for this title yet. Try again later, or search it on another source."
	case len(langs.Hosts) > 0:
		return "⚠️  TopCine plays this title only on " + strings.Join(langs.Hosts, ", ") +
			", which asks for a browser check GoAnime can't pass. Try searching it on another source."
	default:
		return fallback
	}
}

// topCineAnyPlayer picks a player to ask about the title: the movie's, or an
// episode's of the current season (the first season when none is current).
func topCineAnyPlayer(title topcine.Title, season int) string {
	if title.PlayerURL != "" {
		return title.PlayerURL
	}
	for _, s := range title.Seasons {
		if len(s.Episodes) > 0 && (season == 0 || s.Number == season) {
			return s.Episodes[0].PlayerURL
		}
	}
	if len(title.Seasons) > 0 && len(title.Seasons[0].Episodes) > 0 {
		return title.Seasons[0].Episodes[0].PlayerURL
	}
	return ""
}

// fillTopCineEpisodeTitles names the episodes the panel lists without a name,
// from TopCine's own episode list. Best effort: the page is already cached.
func fillTopCineEpisodeTitles(media *models.Anime, episodes []models.Episode) {
	if len(episodes) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), tcCheckBudget)
	defer cancel()
	title, err := tcClientFn().Title(ctx, media.URL)
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

// GetTopCineEpisodes lists a TopCine title: the movie itself, or the episodes
// of the season and audio the user picks.
func GetTopCineEpisodes(media *models.Anime) ([]models.Episode, error) {
	if media == nil || media.URL == "" {
		return nil, fmt.Errorf("no TopCine page for this title")
	}
	return GetStartFlixEpisodes(media)
}

// GetTopCineSeasonNumbers lists a series' season numbers, for downloads that
// name the seasons rather than pick one.
func GetTopCineSeasonNumbers(media *models.Anime) ([]int, error) {
	return GetStartFlixSeasonNumbers(media)
}

// GetTopCineSeasonEpisodes lists one season of a series without the season
// picker; see GetStartFlixSeasonEpisodes.
func GetTopCineSeasonEpisodes(media *models.Anime, seasonNum int) ([]models.Episode, error) {
	return GetStartFlixSeasonEpisodes(media, seasonNum)
}

// GetTopCineStreamURL resolves an episode or movie to a playable URL and hands
// mpv/the downloader the referer, subtitles and audio preference.
func GetTopCineStreamURL(media *models.Anime, episode *models.Episode, quality string) (string, error) {
	return GetStartFlixStreamURL(media, episode, quality)
}
