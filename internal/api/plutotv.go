package api

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/plutotv"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// Pluto TV is a movie/TV catalog read through its own JSON services (see the
// plutotv package). What the api layer adds is the shape the rest of GoAnime
// expects: a movie is one episode, and a series lists its seasons and asks
// which to open (the shared season picker).

// PlutoTVClientFn is the Pluto TV client; a seam for tests.
var PlutoTVClientFn = plutotv.Shared

// plutoPlayURLFn turns a resolved stream into the address to play; a seam.
var plutoPlayURLFn = plutotv.PlayURL

const (
	// plutoListBudget bounds a listing: one session and one catalog call.
	plutoListBudget = 25 * time.Second
	// plutoStreamBudget bounds a stream: the master, up to three renditions
	// and the content playlist.
	plutoStreamBudget = 30 * time.Second
	// plutoReferer is sent with the media requests, as the site's own player
	// does. Naming it also keeps the player from reusing the referer of
	// whatever source was played before.
	plutoReferer = "https://pluto.tv/"
	// plutoShowTTL is how long a series' listing is reused. One answer holds
	// every season, and a download of a whole series asks for it once per
	// season; the catalog does not change in the time that takes.
	plutoShowTTL = 2 * time.Minute
)

// GetPlutoTVEpisodes lists what a Pluto TV title plays: the movie itself, or
// one season of a series, picked by the user when there are several. ctx
// bounds the catalog request; the picker waits for the user regardless.
func GetPlutoTVEpisodes(ctx context.Context, media *models.Anime) ([]models.Episode, error) {
	if media == nil || media.URL == "" {
		return nil, errors.New("no Pluto TV title to list")
	}
	if plutotv.IsMovieURL(media.URL) {
		return []models.Episode{plutoMovieEpisode(media)}, nil
	}
	show, err := plutoShow(ctx, media)
	if err != nil {
		return nil, err
	}
	nums := make([]string, len(show.Seasons))
	items := make([]tui.PickItem, len(show.Seasons))
	for i, s := range show.Seasons {
		nums[i] = strconv.Itoa(s.Number)
		items[i] = tui.PickItem{
			Label:   seasonDisplayName(nums[i]),
			Details: episodeCountLabel(len(s.Episodes)),
			Tags:    []string{plutotv.SourceName},
		}
	}
	chosen, err := selectSeasonWith(sfxSeasonPickFn, media, nums, func() []tui.PickItem { return items })
	if err != nil {
		if errors.Is(err, tui.ErrPickBack) || errors.Is(err, tui.ErrPickCancelled) {
			return nil, ErrBackToSearch
		}
		return nil, fmt.Errorf("season selection cancelled: %w", err)
	}
	number, _ := strconv.Atoi(chosen)
	return plutoSeasonEpisodes(media, show, number)
}

// GetPlutoTVSeasonNumbers lists a series' season numbers, for downloads that
// name the seasons rather than pick one.
func GetPlutoTVSeasonNumbers(media *models.Anime) ([]int, error) {
	if media == nil || media.URL == "" {
		return nil, errors.New("no Pluto TV title")
	}
	show, err := plutoShow(context.Background(), media)
	if err != nil {
		return nil, err
	}
	nums := make([]int, len(show.Seasons))
	for i, s := range show.Seasons {
		nums[i] = s.Number
	}
	return nums, nil
}

// GetPlutoTVSeasonEpisodes lists one season of a series without the picker.
func GetPlutoTVSeasonEpisodes(media *models.Anime, season int) ([]models.Episode, error) {
	if media == nil || media.URL == "" {
		return nil, errors.New("no Pluto TV title")
	}
	show, err := plutoShow(context.Background(), media)
	if err != nil {
		return nil, err
	}
	return plutoSeasonEpisodes(media, show, season)
}

// plutoShows remembers the last listings by series URL (see plutoShowTTL).
var plutoShows = struct {
	mu    sync.Mutex
	byURL map[string]plutoShowEntry
	now   func() time.Time // a seam for tests
}{byURL: map[string]plutoShowEntry{}, now: time.Now}

type plutoShowEntry struct {
	show plutotv.Show
	at   time.Time
}

// ForgetPlutoTVListings drops the listings kept for reuse, so the next one is
// read from the service again.
func ForgetPlutoTVListings() {
	plutoShows.mu.Lock()
	defer plutoShows.mu.Unlock()
	clear(plutoShows.byURL)
}

// plutoShow lists a series, within ctx and the listing budget. A listing made
// moments ago is reused; a failed one is never kept.
func plutoShow(ctx context.Context, media *models.Anime) (plutotv.Show, error) {
	plutoShows.mu.Lock()
	entry, ok := plutoShows.byURL[media.URL]
	now := plutoShows.now()
	plutoShows.mu.Unlock()
	if ok && now.Sub(entry.at) < plutoShowTTL {
		return entry.show, nil
	}

	ctx, cancel := context.WithTimeout(ctx, plutoListBudget)
	defer cancel()
	show, err := PlutoTVClientFn().Show(ctx, media.URL)
	if err != nil {
		return plutotv.Show{}, err
	}
	plutoShows.mu.Lock()
	for url, e := range plutoShows.byURL {
		if now.Sub(e.at) >= plutoShowTTL {
			delete(plutoShows.byURL, url) // nothing stale is kept around
		}
	}
	plutoShows.byURL[media.URL] = plutoShowEntry{show: show, at: now}
	plutoShows.mu.Unlock()
	return show, nil
}

func plutoMovieEpisode(media *models.Anime) models.Episode {
	return models.Episode{
		Number:   "1",
		Num:      1,
		URL:      media.URL,
		SeasonID: "0",
		Title:    models.TitleDetails{English: media.Name, Romaji: media.Name},
	}
}

// plutoSeasonEpisodes builds one season's episode list and records the season
// as the title's current one, the way the other catalogs do.
func plutoSeasonEpisodes(media *models.Anime, show plutotv.Show, number int) ([]models.Episode, error) {
	for _, s := range show.Seasons {
		if s.Number != number {
			continue
		}
		media.CurrentSeason = number
		episodes := make([]models.Episode, 0, len(s.Episodes))
		for _, e := range s.Episodes {
			title := e.Name
			if title == "" {
				title = fmt.Sprintf("Episode %d", e.Number)
			}
			episodes = append(episodes, models.Episode{
				Number:   strconv.Itoa(e.Number),
				Num:      e.Number,
				URL:      plutotv.EpisodeURL(show.ID, e.ID),
				DataID:   e.ID,
				SeasonID: strconv.Itoa(number),
				Synopsis: e.Description,
				Duration: int(e.Duration / time.Second),
				Title:    models.TitleDetails{English: title, Romaji: title},
			})
		}
		util.Debug("Pluto TV episodes loaded", "series", show.ID, "season", number, "count", len(episodes))
		return episodes, nil
	}
	return nil, fmt.Errorf("%q has no season %d on Pluto TV", media.Name, number)
}

// ResolvePlutoTVStream resolves a movie or episode to what the player is
// given: the loopback address of the title's playlist (see plutotv.PlayURL),
// with the site's own referer for the media requests. It writes no playback
// state. ctx cancels the resolution; the stream budget bounds it.
func ResolvePlutoTVStream(ctx context.Context, media *models.Anime, episode *models.Episode, quality string) (source.Stream, error) {
	if episode == nil || episode.URL == "" {
		return source.Stream{}, errors.New("no Pluto TV episode to play")
	}
	ctx, cancel := context.WithTimeout(ctx, plutoStreamBudget)
	defer cancel()
	stream, err := PlutoTVClientFn().Stream(ctx, episode.URL, quality)
	if err != nil {
		return source.Stream{}, err
	}
	if !stream.AdFree && media != nil {
		util.Debug("Pluto TV title plays with the service's own breaks", "title", media.Name)
	}
	return source.Stream{URL: plutoPlayURLFn(stream), Referer: plutoReferer}, nil
}
