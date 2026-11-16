package api

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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

const (
	// plutoListBudget bounds a listing: one session and one catalog call.
	plutoListBudget = 25 * time.Second
	// plutoStreamBudget bounds a stream: the master, up to three renditions
	// and the content playlist.
	plutoStreamBudget = 30 * time.Second
)

// GetPlutoTVEpisodes lists what a Pluto TV title plays: the movie itself, or
// one season of a series, picked by the user when there are several.
func GetPlutoTVEpisodes(media *models.Anime) ([]models.Episode, error) {
	if media == nil || media.URL == "" {
		return nil, errors.New("no Pluto TV title to list")
	}
	if plutotv.IsMovieURL(media.URL) {
		return []models.Episode{plutoMovieEpisode(media)}, nil
	}
	show, err := plutoShow(media)
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
	show, err := plutoShow(media)
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
	show, err := plutoShow(media)
	if err != nil {
		return nil, err
	}
	return plutoSeasonEpisodes(media, show, season)
}

func plutoShow(media *models.Anime) (plutotv.Show, error) {
	ctx, cancel := context.WithTimeout(context.Background(), plutoListBudget)
	defer cancel()
	return PlutoTVClientFn().Show(ctx, media.URL)
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

// ResolvePlutoTVStream resolves a movie or episode to its playlist. It writes
// no playback state: Pluto TV's media needs no referer, so the Stream is just
// the URL.
func ResolvePlutoTVStream(media *models.Anime, episode *models.Episode, quality string) (source.Stream, error) {
	if episode == nil || episode.URL == "" {
		return source.Stream{}, errors.New("no Pluto TV episode to play")
	}
	ctx, cancel := context.WithTimeout(context.Background(), plutoStreamBudget)
	defer cancel()
	stream, err := PlutoTVClientFn().Stream(ctx, episode.URL, quality)
	if err != nil {
		return source.Stream{}, err
	}
	if !stream.AdFree && media != nil {
		util.Debug("Pluto TV title plays with the service's own breaks", "title", media.Name)
	}
	return source.Stream{URL: stream.URL}, nil
}
