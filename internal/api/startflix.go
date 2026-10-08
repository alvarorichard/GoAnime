package api

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api/movie"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/startflix"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
)

// StartFlix flow: title page → video panel → season → Dublado/Legendado →
// episodes. Each episode's URL is the panel endpoint that lists its players, so
// a stream (or a batch download) resolves from the episode alone.
//
// The audio choice is made here, while listing, rather than per stream: the
// panel keeps a separate episode list per audio, and asking before the list is
// shown means a batch download never has to prompt from a worker goroutine.

// Seams, swapped by tests so the whole flow runs without network or a TTY.
var (
	sfxClientFn                    = startflix.Shared
	sfxSeasonPickFn seasonPickFunc = tui.Pick
	sfxPickFn                      = func(prompt string, labels []string) (int, error) {
		return tui.PickLabels(labels, tui.PickOptions{
			Breadcrumb:   "StartFlix > " + prompt,
			WindowTitle:  "GoAnime - StartFlix",
			ItemSingular: "option",
			ItemPlural:   "options",
		})
	}
)

const (
	sfxListBudget   = 30 * time.Second
	sfxStreamBudget = 45 * time.Second
)

// describeStartFlixErr turns a StartFlix failure into a plain-language message,
// keeping the cause reachable through Unwrap.
func describeStartFlixErr(err error) error {
	var noStream *startflix.NoStreamError
	switch {
	case err == nil:
		return nil
	case errors.As(err, &noStream) && len(noStream.Failures) == 0:
		return &friendlyError{cause: err, msg: "⚠️  StartFlix only offers this on servers GoAnime can't play yet (" +
			strings.Join(noStream.Unsupported, ", ") + "). Try another episode, or another title."}
	case errors.Is(err, startflix.ErrNoSupportedServer):
		return &friendlyError{cause: err, msg: "⚠️  None of StartFlix's servers worked for this right now. Try again in a moment."}
	case errors.Is(err, startflix.ErrNoPlayers):
		return &friendlyError{cause: err, msg: "⚠️  No video sources for this on StartFlix right now. Try another episode, or come back later."}
	case errors.Is(err, startflix.ErrNotOnPanel), errors.Is(err, startflix.ErrNoPanel):
		return &friendlyError{cause: err, msg: "⚠️  StartFlix lists this title but has no video for it. Try searching it on another source."}
	case errors.Is(err, context.DeadlineExceeded):
		return &friendlyError{cause: err, msg: "⚠️  StartFlix took too long to answer. Please try again."}
	default:
		return err
	}
}

// sfxAudioChoices remembers, per title page, the audio the user picked — so
// every season of a binge is asked once. Session-scoped on purpose: it is a
// playback preference, not something worth persisting to disk.
var (
	sfxAudioMu      sync.Mutex
	sfxAudioChoices = map[string]startflix.Audio{}
)

func rememberStartFlixAudio(titleURL string, a startflix.Audio) {
	sfxAudioMu.Lock()
	defer sfxAudioMu.Unlock()
	sfxAudioChoices[titleURL] = a
}

func recallStartFlixAudio(titleURL string) (startflix.Audio, bool) {
	sfxAudioMu.Lock()
	defer sfxAudioMu.Unlock()
	a, ok := sfxAudioChoices[titleURL]
	return a, ok
}

// startFlixPinnedAudioLang is the --audio value only when the user actually
// passed it. util.GlobalAudioLanguage is never empty in the real binary (it
// holds the flag's default), so reading it alone would take "pt-BR,pt,english"
// for a choice and never offer Legendado. When the flag was passed this package
// never writes the global (see GetStartFlixStreamURL), so it is still the
// user's value.
func startFlixPinnedAudioLang() string {
	if !util.GlobalAudioLanguageExplicit {
		return ""
	}
	return util.GetGlobalAudioLanguage()
}

// portugueseAudioCodes are the language codes that mean "the Portuguese dub".
var portugueseAudioCodes = map[string]bool{"por": true, "pob": true, "pt": true, "ptb": true}

// mpv matches --alang against whatever the manifest tags a track with, and
// hosts spell the same language differently, so every plausible alias is given.
const (
	portugueseALang = "por,pob,pt-BR,ptbr,pt,portuguese"
	originalALang   = "jpn,ja,japanese,eng,en,english,kor,ko,korean,spa,es,spanish"
)

// pinnedStartFlixAudio maps an explicit --audio onto the panel's lists:
// Portuguese means the dub, any other language means the original audio.
func pinnedStartFlixAudio() (startflix.Audio, bool) {
	pinned := strings.ToLower(strings.TrimSpace(startFlixPinnedAudioLang()))
	if pinned == "" {
		return "", false
	}
	first := strings.TrimSpace(strings.Split(pinned, ",")[0])
	if portugueseAudioCodes[first] || strings.HasPrefix(first, "pt") || strings.HasPrefix(first, "portug") {
		return startflix.AudioDubbed, true
	}
	return startflix.AudioSubtitled, true
}

// selectStartFlixAudio picks the episode list to show, asking only when the
// season genuinely has both.
func selectStartFlixAudio(titleURL string, season startflix.Season) startflix.Audio {
	audios := season.Audios()
	if len(audios) == 1 {
		return audios[0]
	}
	has := func(a startflix.Audio) bool { return len(season.Episodes(a)) > 0 }
	if a, ok := recallStartFlixAudio(titleURL); ok && has(a) {
		return a
	}
	if a, ok := pinnedStartFlixAudio(); ok && has(a) {
		return a
	}

	labels := []string{
		fmt.Sprintf("🎙️  Dublado (%s)", episodeCountLabel(len(season.Dubbed))),
		fmt.Sprintf("💬 Legendado (%s)", episodeCountLabel(len(season.Subtitled))),
	}
	choice := startflix.AudioDubbed
	if idx, err := sfxPickFn("Você quer assistir dublado ou legendado? ", labels); err == nil && idx == 1 {
		choice = startflix.AudioSubtitled
	} else if err != nil {
		util.Debug("StartFlix: audio picker unavailable, taking the dub", "err", err)
	}
	rememberStartFlixAudio(titleURL, choice)
	return choice
}

// startFlixSeasonItems renders the season picker rows.
func startFlixSeasonItems(seasons []startflix.Season) []tui.PickItem {
	items := make([]tui.PickItem, len(seasons))
	for i, s := range seasons {
		var parts []string
		if n := len(s.Dubbed); n > 0 {
			parts = append(parts, "Dublado: "+episodeCountLabel(n))
		}
		if n := len(s.Subtitled); n > 0 {
			parts = append(parts, "Legendado: "+episodeCountLabel(n))
		}
		items[i] = tui.PickItem{Label: seasonDisplayName(s.Key()), Details: strings.Join(parts, "  •  ")}
	}
	return items
}

func selectStartFlixSeason(media *models.Anime, seasons []startflix.Season) (startflix.Season, error) {
	keys := make([]string, len(seasons))
	for i, s := range seasons {
		keys[i] = s.Key()
	}
	key, err := selectSeasonWith(sfxSeasonPickFn, media, keys, func() []tui.PickItem { return startFlixSeasonItems(seasons) })
	if err != nil {
		return startflix.Season{}, err
	}
	for _, s := range seasons {
		if s.Key() == key {
			return s, nil
		}
	}
	return startflix.Season{}, fmt.Errorf("season %q not found", key)
}

// loadStartFlixPanel resolves the title's panel, records the ids it carries
// and looks up the official record those ids name.
func loadStartFlixPanel(c *startflix.Client, media *models.Anime) (startflix.Panel, error) {
	var panel startflix.Panel
	var err error
	runWithSpinner("Loading title...", func() {
		ctx, cancel := context.WithTimeout(context.Background(), sfxListBudget)
		defer cancel()
		panel, err = c.Panel(ctx, media.URL)
		if err == nil {
			adoptStartFlixIDs(media, panel)
			resolveStartFlixOfficial(media, panel)
		}
	})
	if err != nil {
		return startflix.Panel{}, err
	}
	// The panel is the authority on what the title is; the search page only
	// guessed from its URL.
	if panel.Kind == startflix.KindMovie {
		media.MediaType = models.MediaTypeMovie
	} else if media.MediaType == models.MediaTypeMovie {
		media.MediaType = models.MediaTypeTV
	}
	return panel, nil
}

// adoptStartFlixIDs records the panel's ids on the title. The panel is keyed
// by them (TMDB id for shows, IMDb id for movies), so they are the title's
// real ids. Enrichment at selection time searched by the Portuguese name and
// may have matched another title — "Coração Selvagem" is also how Brazil
// knows Lynch's "Wild at Heart" — so when the ids disagree, that guess and
// everything it brought in are dropped.
func adoptStartFlixIDs(media *models.Anime, panel startflix.Panel) {
	tmdbDiffers := panel.TMDBID > 0 && panel.TMDBID != media.TMDBID
	imdbDiffers := panel.IMDBID != "" && panel.IMDBID != media.IMDBID
	if !tmdbDiffers && !imdbDiffers {
		return
	}
	if media.TMDBID > 0 || media.IMDBID != "" {
		util.Debug("StartFlix panel ids replace a name-search match", "title", media.Name,
			"guessed_tmdb", media.TMDBID, "guessed_imdb", media.IMDBID, "tmdb", panel.TMDBID, "imdb", panel.IMDBID)
	}
	media.TMDBID, media.IMDBID = panel.TMDBID, panel.IMDBID
	media.TMDBDetails, media.Rating, media.Overview, media.Genres, media.Runtime = nil, 0, "", nil, 0
}

// sfxEnrichByIDFn looks up the official record by id; a seam for tests.
var sfxEnrichByIDFn = movie.EnrichByID

// resolveStartFlixOfficial names the title after its official English record
// — the one IMDb/TMDB file under the panel's ids — so downloads land as
// "Heart of the Beast (2026) {imdb-tt7526136}" rather than under the
// Portuguese "Coração Selvagem". SuperFlix got this for free, its search
// results already carried the ids; StartFlix only reveals them here. Best
// effort: without an answer, the StartFlix name stands.
func resolveStartFlixOfficial(media *models.Anime, panel startflix.Panel) {
	if err := sfxEnrichByIDFn(media, panel.Kind == startflix.KindMovie); err != nil {
		util.Debug("StartFlix official title unavailable", "title", media.Name,
			"tmdb", media.TMDBID, "imdb", media.IMDBID, "err", err)
		return
	}
	util.Debug("StartFlix official title", "title", media.Name, "official", media.OfficialTitle(), "year", media.Year)
}

func loadStartFlixSeasons(c *startflix.Client, panel startflix.Panel) ([]startflix.Season, error) {
	var seasons []startflix.Season
	var err error
	runWithSpinner("Loading seasons...", func() {
		ctx, cancel := context.WithTimeout(context.Background(), sfxListBudget)
		defer cancel()
		seasons, err = c.Seasons(ctx, panel)
	})
	return seasons, err
}

// GetStartFlixEpisodes lists a StartFlix title: the movie itself, or the
// episodes of the season and audio the user picks.
func GetStartFlixEpisodes(media *models.Anime) ([]models.Episode, error) {
	if media == nil || media.URL == "" {
		return nil, fmt.Errorf("no StartFlix page for this title")
	}
	c := sfxClientFn()

	panel, err := loadStartFlixPanel(c, media)
	if err != nil {
		return nil, describeStartFlixErr(err)
	}

	if panel.Kind == startflix.KindMovie {
		return []models.Episode{{
			Number: "1",
			Num:    1,
			URL:    panel.URL,
			Title:  models.TitleDetails{English: media.Name, Romaji: media.Name},
		}}, nil
	}

	seasons, err := loadStartFlixSeasons(c, panel)
	if err != nil {
		return nil, describeStartFlixErr(err)
	}

	season, err := selectStartFlixSeason(media, seasons)
	if err != nil {
		if errors.Is(err, tui.ErrPickBack) || errors.Is(err, tui.ErrPickCancelled) {
			return nil, ErrBackToSearch
		}
		return nil, fmt.Errorf("season selection cancelled: %w", err)
	}
	return startFlixSeasonEpisodes(media, panel, season), nil
}

// ErrStartFlixNotSeries is returned when a series operation meets a movie.
var ErrStartFlixNotSeries = errors.New("this StartFlix title is a movie, not a series")

// loadStartFlixSeries opens a title's panel and its seasons, refusing movies.
func loadStartFlixSeries(media *models.Anime) (startflix.Panel, []startflix.Season, error) {
	if media == nil || media.URL == "" {
		return startflix.Panel{}, nil, fmt.Errorf("no StartFlix page for this title")
	}
	c := sfxClientFn()
	panel, err := loadStartFlixPanel(c, media)
	if err != nil {
		return startflix.Panel{}, nil, describeStartFlixErr(err)
	}
	if panel.Kind == startflix.KindMovie {
		return startflix.Panel{}, nil, ErrStartFlixNotSeries
	}
	seasons, err := loadStartFlixSeasons(c, panel)
	if err != nil {
		return startflix.Panel{}, nil, describeStartFlixErr(err)
	}
	return panel, seasons, nil
}

// GetStartFlixSeasonNumbers lists a series' season numbers, in panel order,
// for downloads that name the seasons rather than pick one.
func GetStartFlixSeasonNumbers(media *models.Anime) ([]int, error) {
	_, seasons, err := loadStartFlixSeries(media)
	if err != nil {
		return nil, err
	}
	nums := make([]int, 0, len(seasons))
	for _, s := range seasons {
		nums = append(nums, s.Number)
	}
	return nums, nil
}

// GetStartFlixSeasonEpisodes lists one season of a series without the season
// picker, in the audio the user picks (asked once per title, then
// remembered), and records it as the title's current season.
func GetStartFlixSeasonEpisodes(media *models.Anime, seasonNum int) ([]models.Episode, error) {
	panel, seasons, err := loadStartFlixSeries(media)
	if err != nil {
		return nil, err
	}
	available := make([]string, 0, len(seasons))
	for _, s := range seasons {
		if s.Number == seasonNum {
			return startFlixSeasonEpisodes(media, panel, s), nil
		}
		available = append(available, s.Key())
	}
	return nil, fmt.Errorf("season %d is not on StartFlix for %q (seasons: %s)", seasonNum, media.Name, strings.Join(available, ", "))
}

// startFlixSeasonEpisodes builds a season's episode list in the audio the
// user picks, and records the season as the title's current one.
func startFlixSeasonEpisodes(media *models.Anime, panel startflix.Panel, season startflix.Season) []models.Episode {
	audio := selectStartFlixAudio(media.URL, season)
	media.CurrentSeason = season.Number

	list := season.Episodes(audio)
	episodes := make([]models.Episode, 0, len(list))
	for _, ep := range list {
		episodes = append(episodes, models.Episode{
			Number:   strconv.Itoa(ep.Number),
			Num:      ep.Number,
			URL:      panel.EpisodeURL(ep.ID),
			DataID:   ep.ID,
			SeasonID: season.Key(),
			Title:    models.TitleDetails{English: ep.Title, Romaji: ep.Title},
		})
	}
	util.Debug("StartFlix episodes loaded", "season", season.Number, "audio", audio, "count", len(episodes))
	return episodes
}

// isStartFlixPlayersURL reports whether an episode URL is already the panel
// endpoint that lists players.
func isStartFlixPlayersURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	return strings.HasPrefix(u.Path, "/episodio/") || strings.HasPrefix(u.Path, "/filme/")
}

// startFlixPlayersURL finds where an episode's players are listed. Episodes
// built by GetStartFlixEpisodes already carry it; one rebuilt from history may
// only know its season and number, so the panel is walked again.
func startFlixPlayersURL(c *startflix.Client, media *models.Anime, episode *models.Episode) (string, error) {
	if isStartFlixPlayersURL(episode.URL) {
		return episode.URL, nil
	}
	panel, err := loadStartFlixPanel(c, media)
	if err != nil {
		return "", err
	}
	if panel.Kind == startflix.KindMovie {
		return panel.URL, nil
	}
	seasons, err := loadStartFlixSeasons(c, panel)
	if err != nil {
		return "", err
	}
	want := episode.SeasonID
	if want == "" && media.CurrentSeason > 0 {
		want = strconv.Itoa(media.CurrentSeason)
	}
	num := episode.Num
	if num == 0 {
		num, _ = strconv.Atoi(episode.Number)
	}
	preferred, _ := recallStartFlixAudio(media.URL)
	for _, s := range seasons {
		if want != "" && s.Key() != want {
			continue
		}
		for _, a := range []startflix.Audio{preferred, startflix.AudioDubbed, startflix.AudioSubtitled} {
			for _, ep := range s.Episodes(a) {
				if a != "" && ep.Number == num {
					return panel.EpisodeURL(ep.ID), nil
				}
			}
		}
	}
	return "", fmt.Errorf("startflix: episode %s of season %q not found on the panel", episode.Number, want)
}

// audioLangForStartFlix tells mpv which track to prefer. A Dublado list may
// still be a dual-audio release, and a Legendado one usually is the original
// with burned-in subtitles; either way the user's list choice decides.
func audioLangForStartFlix(a startflix.Audio) string {
	if a == startflix.AudioSubtitled {
		return originalALang
	}
	return portugueseALang
}

// GetStartFlixStreamURL resolves an episode or movie to a playable URL and
// hands mpv/the downloader the referer, subtitles and audio preference.
func GetStartFlixStreamURL(media *models.Anime, episode *models.Episode, _ string) (string, error) {
	if media == nil || episode == nil {
		return "", fmt.Errorf("no StartFlix title or episode to play")
	}
	c := sfxClientFn()

	playersURL, err := startFlixPlayersURL(c, media, episode)
	if err != nil {
		return "", describeStartFlixErr(err)
	}

	var stream *startflix.Stream
	runWithSpinner("Loading stream...", func() {
		ctx, cancel := context.WithTimeout(context.Background(), sfxStreamBudget)
		defer cancel()
		stream, err = c.Stream(ctx, playersURL)
	})
	if err != nil {
		// No prefix: the player dispatch already wraps this as "failed to get
		// StartFlix stream URL", and the friendly text should follow it directly.
		return "", describeStartFlixErr(err)
	}

	if stream.Referer != "" {
		util.SetGlobalReferer(stream.Referer)
	}
	// Always set when the user pinned nothing: the preference is a process-wide
	// global, and leaving it alone would carry the last title's Legendado choice
	// into the next movie.
	if startFlixPinnedAudioLang() == "" {
		audio := startflix.AudioDubbed
		if a, ok := recallStartFlixAudio(media.URL); ok && media.MediaType != models.MediaTypeMovie {
			audio = a
		}
		util.SetGlobalAudioLanguage(audioLangForStartFlix(audio))
	}
	if len(stream.Subtitles) > 0 && !util.GlobalNoSubs {
		subs := make([]util.SubtitleInfo, 0, len(stream.Subtitles))
		for _, s := range stream.Subtitles {
			subs = append(subs, util.SubtitleInfo{URL: s.URL, Language: s.Language, Label: s.Label})
		}
		util.SetGlobalSubtitles(subs)
	}
	util.Debug("StartFlix stream URL obtained", "host", stream.Host, "url", stream.URL[:min(len(stream.URL), 80)])
	return stream.URL, nil
}
