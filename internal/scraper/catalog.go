package scraper

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/startflix"
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
	// CatalogStream resolves the episode's stream. A non-empty audio that
	// does not play is ErrAudioUnavailable, never the other audio in silence.
	CatalogStream(ctx context.Context, episodeURL, audio string) (string, map[string]string, error)
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
				Title:    models.TitleDetails{English: e.Title, Romaji: e.Title},
			})
		}
		return eps, nil
	}
	return nil, fmt.Errorf("%w: season %d (seasons: %s)", ErrSeasonNotFound, season, strings.Join(have, ", "))
}

func (a *StartFlixAdapter) CatalogStream(ctx context.Context, episodeURL, audio string) (string, map[string]string, error) {
	want, err := ParseCatalogAudio(audio)
	if err != nil {
		return "", nil, err
	}
	players, err := a.client.Players(ctx, episodeURL)
	if err != nil {
		return "", nil, err
	}
	stream, err := a.client.ResolveStreamFor(ctx, players, want)
	if err != nil {
		return "", nil, err
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
