package scraper

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/startflix"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// CatalogScraper is the optional capability (discovered by type assertion) of
// the movie and series catalogues, whose episodes hang off seasons and audio
// tracks. The CLI picks both through menus; this is the same listing without
// any: the caller names the season and the audio, and a choice that does not
// exist is an error that says what does.
type CatalogScraper interface {
	// CatalogSeasons lists a series' seasons with the audios each carries.
	// A movie has none: it returns an empty list.
	CatalogSeasons(ctx context.Context, titleURL string) ([]CatalogSeason, error)
	// CatalogEpisodes lists a title's episodes. A movie is one episode and
	// ignores season and audio; a series needs a season (ErrSeasonRequired
	// otherwise). An empty audio takes the dub when the season has one.
	CatalogEpisodes(ctx context.Context, titleURL string, season int, audio string) ([]models.Episode, error)
	// CatalogStream resolves an episode's stream (episode as CatalogEpisodes
	// listed it; titleURL is the title page, "" when the caller has none).
	// When the panel's servers fail, or do not carry the audio asked for,
	// the title is looked up by its ids on the independent vidsrc fallback,
	// which serves the original audio (Legendado). A non-empty audio that
	// does not play is ErrAudioUnavailable, never the other audio in silence.
	CatalogStream(ctx context.Context, titleURL string, ep models.Episode, audio string) (streamURL string, metadata map[string]string, err error)
}

// CatalogSeason is one season of a series.
type CatalogSeason struct {
	Number int
	Audios []string // "Dublado", "Legendado"
}

var (
	// ErrSeasonRequired is a series asked for without saying which season.
	ErrSeasonRequired = errors.New("this title is a series: choose a season")
	// ErrAudioUnavailable is an audio the title does not offer, or offers on
	// no server that plays.
	ErrAudioUnavailable = errors.New("the requested audio is not available")
	// ErrSeasonNotFound is a season the series does not have.
	ErrSeasonNotFound = errors.New("season not found")
)

// ParseCatalogAudio maps the spellings a caller may use to the panel's audio.
// "" means no preference.
func ParseCatalogAudio(s string) (startflix.Audio, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "":
		return "", nil
	case "dublado", "dub", "dubbed":
		return startflix.AudioDubbed, nil
	case "legendado", "sub", "subbed", "subtitled":
		return startflix.AudioSubtitled, nil
	}
	return "", fmt.Errorf("%w: %q (use Dublado or Legendado)", ErrAudioUnavailable, s)
}

func audioNames(as []startflix.Audio) []string {
	out := make([]string, len(as))
	for i, a := range as {
		out[i] = string(a)
	}
	return out
}

func (a *StartFlixAdapter) CatalogSeasons(ctx context.Context, titleURL string) ([]CatalogSeason, error) {
	panel, err := a.client.Panel(ctx, titleURL)
	if err != nil {
		return nil, err
	}
	if panel.Kind == startflix.KindMovie {
		return nil, nil
	}
	seasons, err := a.client.Seasons(ctx, panel)
	if err != nil {
		return nil, err
	}
	out := make([]CatalogSeason, len(seasons))
	for i, s := range seasons {
		out[i] = CatalogSeason{Number: s.Number, Audios: audioNames(s.Audios())}
	}
	return out, nil
}

func (a *StartFlixAdapter) CatalogEpisodes(ctx context.Context, titleURL string, season int, audio string) ([]models.Episode, error) {
	want, err := ParseCatalogAudio(audio)
	if err != nil {
		return nil, err
	}
	panel, err := a.client.Panel(ctx, titleURL)
	if err != nil {
		return nil, err
	}
	if panel.Kind == startflix.KindMovie {
		return []models.Episode{{Number: "1", Num: 1, URL: panel.URL}}, nil
	}
	if season <= 0 {
		return nil, ErrSeasonRequired
	}
	seasons, err := a.client.Seasons(ctx, panel)
	if err != nil {
		return nil, err
	}
	var have []string
	for _, s := range seasons {
		have = append(have, strconv.Itoa(s.Number))
		if s.Number != season {
			continue
		}
		pick := want
		if pick == "" {
			pick = startflix.AudioDubbed
			if len(s.Dubbed) == 0 {
				pick = startflix.AudioSubtitled
			}
		}
		list := s.Episodes(pick)
		if len(list) == 0 {
			return nil, fmt.Errorf("%w: season %d has no %s (it has: %s)", ErrAudioUnavailable,
				season, pick, strings.Join(audioNames(s.Audios()), ", "))
		}
		eps := make([]models.Episode, 0, len(list))
		for _, e := range list {
			eps = append(eps, models.Episode{
				Number:   strconv.Itoa(e.Number),
				Num:      e.Number,
				URL:      panel.EpisodeURL(e.ID),
				DataID:   e.ID,
				SeasonID: s.Key(),
				Audio:    string(pick),
				Title:    models.TitleDetails{English: e.Title, Romaji: e.Title},
			})
		}
		return eps, nil
	}
	return nil, fmt.Errorf("%w: season %d (seasons: %s)", ErrSeasonNotFound, season, strings.Join(have, ", "))
}

// catalogPhaseBudget bounds the panel's servers and, separately, the fallback:
// one hanging host must not leave the fallback a deadline it cannot meet.
const catalogPhaseBudget = 45 * time.Second

// catalogFallbackFn resolves a title by its ids on the vidsrc fallback; a seam
// for tests, as sfxVIDSRCFallbackFn is in the api package. Movies take the
// IMDb id (TMDB when the panel carries only that); series take the TMDB id
// with the season and episode number.
var catalogFallbackFn = func(ctx context.Context, panel startflix.Panel, season, episode int) (*startflix.Stream, error) {
	if panel.Kind == startflix.KindMovie {
		return startflix.MovieFallback(ctx, panel.IMDBID, panel.TMDBID)
	}
	return startflix.EpisodeFallback(ctx, panel.TMDBID, season, episode)
}

// vidsrcStream tries the fallback for the title. nil when the caller has no
// title page, the panel cannot be found, or the fallback misses.
func (a *StartFlixAdapter) vidsrcStream(ctx context.Context, titleURL string, ep models.Episode) *startflix.Stream {
	if titleURL == "" {
		util.Debug("vidsrc fallback skipped: no title page to find the ids on", "episode", ep.URL)
		return nil
	}
	fctx, cancel := context.WithTimeout(ctx, catalogPhaseBudget)
	defer cancel()
	panel, err := a.client.Panel(fctx, titleURL)
	if err != nil {
		util.Debug("vidsrc fallback skipped: no panel for the title", "title", titleURL, "err", err)
		return nil
	}
	season, _ := strconv.Atoi(ep.SeasonID)
	num := ep.Num
	if num == 0 {
		num, _ = strconv.Atoi(ep.Number)
	}
	// Season or episode 0 means "whatever the source names as Pilot": for a
	// series that would answer some other episode as the one asked for.
	if panel.Kind != startflix.KindMovie && (season <= 0 || num <= 0) {
		util.Debug("vidsrc fallback skipped: series episode without season or number", "season", season, "episode", num)
		return nil
	}
	fs, err := catalogFallbackFn(fctx, panel, season, num)
	if err != nil || fs == nil {
		util.Debug("vidsrc fallback missed", "title", titleURL, "season", season, "episode", num, "err", err)
		return nil
	}
	if fs.Audio == "" {
		fs.Audio = startflix.AudioSubtitled // the fallback serves the original audio
	}
	return fs
}

func (a *StartFlixAdapter) CatalogStream(ctx context.Context, titleURL string, ep models.Episode, audio string) (streamURL string, metadata map[string]string, err error) {
	want, err := ParseCatalogAudio(audio)
	if err != nil {
		return "", nil, err
	}
	if want == "" {
		// An episode listed in one audio plays in that audio: the panel adds
		// a player of the other audio to a list, and the tallest would win.
		want, _ = ParseCatalogAudio(ep.Audio)
	}
	pctx, cancel := context.WithTimeout(ctx, catalogPhaseBudget)
	players, perr := a.client.Players(pctx, ep.URL)
	var stream *startflix.Stream
	if perr == nil {
		stream, perr = a.client.ResolveStreamFor(pctx, players, want)
	}
	cancel()

	wrongAudio := perr == nil && want != "" && stream.Audio != "" && stream.Audio != want
	// The fallback only serves the original audio, so it is not asked for a dub.
	if (perr != nil || wrongAudio) && want != startflix.AudioDubbed {
		if ctx.Err() != nil {
			return "", nil, ctx.Err()
		}
		// It replaces a panel that failed, and a panel stream in the wrong
		// audio, when it carries the audio asked for.
		if fs := a.vidsrcStream(ctx, titleURL, ep); fs != nil && (want == "" || fs.Audio == want) {
			stream, perr = fs, nil
		}
	}
	if perr != nil {
		return "", nil, perr
	}
	if want != "" && stream.Audio != "" && stream.Audio != want {
		return "", nil, fmt.Errorf("%w: %s plays, %s does not (listed: %s)", ErrAudioUnavailable,
			stream.Audio, want, strings.Join(audioNames(startflix.ListedAudios(players, "")), ", "))
	}
	meta := startFlixStreamMetadata(stream)
	if stream.Audio != "" {
		meta["audio"] = string(stream.Audio)
	}
	return stream.URL, meta, nil
}
