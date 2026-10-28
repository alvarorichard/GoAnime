package api

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/api/movie"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/pobreflix"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/startflix"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/topcine"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// ytdlpOnPath reports whether the yt-dlp binary is available for the
// generic server fallback. When it is missing, the unsupported-servers
// message says so instead of reading as a dead end.
func ytdlpOnPath() bool {
	_, err := exec.LookPath("yt-dlp")
	return err == nil
}

// StartFlix flow: title page → video panel → season → Dublado/Legendado →
// episodes. Each episode's URL is the panel endpoint that lists its players, so
// a stream (or a batch download) resolves from the episode alone.
//
// TopCine titles take the same flow from the panel on: their page names the
// TMDB id the panel is keyed by (see topcine.go), so only the first step
// differs.
//
// The audio choice is made here, while listing, rather than per stream: the
// panel keeps a separate episode list per audio, and asking before the list is
// shown means a batch download never has to prompt from a worker goroutine.

// Seams, swapped by tests so the whole flow runs without network or a TTY.
var (
	sfxClientFn                    = startflix.Shared
	sfxSeasonPickFn seasonPickFunc = tui.Pick
	// sfxPickFn's prompt opens with the source's breadcrumb, "StartFlix > …".
	sfxPickFn = func(prompt string, labels []string) (int, error) {
		source, _, _ := strings.Cut(prompt, " > ")
		return tui.PickLabels(labels, tui.PickOptions{
			Breadcrumb:   prompt,
			WindowTitle:  "GoAnime - " + source,
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
	return describePanelErr(startflix.SourceName, err)
}

// describePanelErr is describeStartFlixErr for any source whose titles play
// through the StartFlix panel, naming that source.
func describePanelErr(source string, err error) error {
	var noStream *startflix.NoStreamError
	switch {
	case err == nil:
		return nil
	case errors.As(err, &noStream) && len(noStream.Failures) == 0:
		msg := source + " only offers this on servers GoAnime can't play yet (" +
			strings.Join(noStream.Unsupported, ", ") + "). Try another episode, or another title."
		if !ytdlpOnPath() {
			msg += " Installing yt-dlp unlocks more servers."
		}
		return &friendlyError{cause: err, msg: msg}
	case errors.Is(err, startflix.ErrNoSupportedServer):
		return &friendlyError{cause: err, msg: "None of " + source + "'s servers worked for this right now. Try again in a moment."}
	case errors.Is(err, startflix.ErrNoPlayers):
		return &friendlyError{cause: err, msg: "No video sources for this on " + source + " right now. Try another episode, or come back later."}
	case errors.Is(err, startflix.ErrNotOnPanel), errors.Is(err, startflix.ErrNoPanel):
		return &friendlyError{cause: err, msg: source + " lists this title but has no video for it. Try another result below."}
	case errors.Is(err, context.DeadlineExceeded):
		return &friendlyError{cause: err, msg: source + " took too long to answer. Please try again."}
	default:
		return err
	}
}

// describeMediaErr describes a panel failure for the source the title came
// from.
func describeMediaErr(media *models.Anime, err error) error {
	if isTopCine(media) {
		return describeTopCineErr(media, err)
	}
	if isPobreflix(media) {
		return describePobreflixErr(media, err)
	}
	return describeStartFlixErr(err)
}

// panelSourceName is the source a panel title is listed on, for messages.
func panelSourceName(media *models.Anime) string {
	if isTopCine(media) {
		return topcine.SourceName
	}
	if isPobreflix(media) {
		return pobreflix.SourceName
	}
	return startflix.SourceName
}

// sfxAudioChoices remembers, per title page, the audio the user picked — so
// every season of a binge is asked once. Session-scoped on purpose: it is a
// playback preference, not something worth persisting to disk.
var (
	sfxAudioMu      sync.Mutex
	sfxAudioChoices = map[string]startflix.Audio{}
)

// sfxDeadTitles remembers title pages whose video panel answered "no content"
// (ErrNotOnPanel) or that embed no panel at all (ErrNoPanel) — e.g. StartFlix
// still lists "Velozes e Furiosos – Espiões do Asfalto" while its panel at
// /embed/95594 answers 404 "Conteúdo não encontrado". The listing is stale
// upstream and will not heal on retry, so the title is hidden from later
// searches in the same session instead of being offered again. Session-scoped
// like the audio choice above: the panel may come back, and a restart must see
// the title again.
var (
	sfxDeadMu     sync.RWMutex
	sfxDeadTitles = map[string]bool{}
)

// markPanelDead records a title page as unplayable: the catalog lists it but
// its panel carries no video.
func markPanelDead(titleURL string) {
	if titleURL == "" {
		return
	}
	sfxDeadMu.Lock()
	defer sfxDeadMu.Unlock()
	if !sfxDeadTitles[titleURL] {
		sfxDeadTitles[titleURL] = true
		util.Debug("Panel title marked dead, hiding from later searches", "url", titleURL)
	}
}

// isPanelDead reports whether a title page was marked dead this session.
func isPanelDead(titleURL string) bool {
	if titleURL == "" {
		return false
	}
	sfxDeadMu.RLock()
	defer sfxDeadMu.RUnlock()
	return sfxDeadTitles[titleURL]
}

// isPanelDeadErr reports whether err means "listed but no video": the panel
// has no content for the title, the title embeds no panel at all, or the
// panel lists the title only on hosts GoAnime (including the yt-dlp
// fallback) cannot play. The last case is also stale upstream — e.g. an old
// movie whose five player hosts are all long-dead DNS — and retrying it in
// the same session cannot heal, so the title is hidden like a dead panel.
// Session-scoped: a restart (possibly with yt-dlp installed or a new
// resolver released) shows it again.
func isPanelDeadErr(err error) bool {
	if errors.Is(err, startflix.ErrNotOnPanel) || errors.Is(err, startflix.ErrNoPanel) {
		return true
	}
	var noStream *startflix.NoStreamError
	if errors.As(err, &noStream) && len(noStream.Unsupported) > 0 && len(noStream.Failures) == 0 {
		return true
	}
	return false
}

// markPanelDeadErr marks the title page dead when err is a panel-missing
// failure. It unwraps friendlyError layers (describeMediaErr) so callers pass
// the error they already have.
//
// Only StartFlix's own listing hides: its stale pages are what this map
// exists for — "listed here, panel answers no content" describes a page the
// panel never gained content for. Pobreflix and TopCine instead ride the panel
// by id (see topcine.go): their pages are third-party catalog entries that
// keep naming the same TMDB id no matter what the panel serves today. The
// panel's player records change upstream (its base is even env-repointable,
// GOANIME_STARTFLIX_PANEL_URL), so hiding their search result over a panel
// failure would take a still-real title out of search for the session —
// "Todo Poderoso (2003)" on Pobreflix is exactly that: every server the panel
// lists for it is dead, and the title must stay visible and re-triable. The
// describe functions already explain the failure in plain language.
func markPanelDeadErr(media *models.Anime, err error) {
	if isPobreflix(media) || isTopCine(media) {
		util.Debug("Panel failure on a proxy-catalog title, keeping it in the search results",
			"source", media.Source, "url", media.URL, "err", err)
		return
	}
	if isPanelDeadErr(err) {
		markPanelDead(media.URL)
	}
}

// filterPanelDead drops titles marked dead this session from a search result.
// When every result is dead the original list is kept: an empty picker reads
// as "nothing matched" when the truth is "everything matched is gone", and
// hiding the reason helps nobody.
func filterPanelDead(animes []*models.Anime) []*models.Anime {
	sfxDeadMu.RLock()
	empty := len(sfxDeadTitles) == 0
	sfxDeadMu.RUnlock()
	if empty {
		return animes
	}
	kept := make([]*models.Anime, 0, len(animes))
	for _, a := range animes {
		if a != nil && isPanelDead(a.URL) {
			util.Debug("Hiding dead panel title from search results", "title", a.Name, "url", a.URL)
			continue
		}
		kept = append(kept, a)
	}
	if len(kept) == 0 {
		return animes
	}
	return kept
}

// sfxListedAudio remembers, per title page and season, the audio list the
// episodes were built from — also when the season had a single audio and no
// choice was asked — so a stream knows what the user is watching in.
// sfxAudioSwaps remembers the titles whose user accepted another audio after
// theirs failed, so a binge is asked once. Both session-scoped, like the
// audio choice.
var (
	sfxListedAudio = map[string]startflix.Audio{}
	sfxAudioSwaps  = map[string]startflix.Audio{}
)

func listedAudioKey(titleURL, seasonKey string) string { return titleURL + "#" + seasonKey }

func rememberListedAudio(titleURL, seasonKey string, a startflix.Audio) {
	sfxAudioMu.Lock()
	defer sfxAudioMu.Unlock()
	sfxListedAudio[listedAudioKey(titleURL, seasonKey)] = a
}

func recallListedAudio(titleURL, seasonKey string) (startflix.Audio, bool) {
	sfxAudioMu.Lock()
	defer sfxAudioMu.Unlock()
	a, ok := sfxListedAudio[listedAudioKey(titleURL, seasonKey)]
	return a, ok
}

func rememberAudioSwap(titleURL string, a startflix.Audio) {
	sfxAudioMu.Lock()
	defer sfxAudioMu.Unlock()
	sfxAudioSwaps[titleURL] = a
}

func recallAudioSwap(titleURL string) (startflix.Audio, bool) {
	sfxAudioMu.Lock()
	defer sfxAudioMu.Unlock()
	a, ok := sfxAudioSwaps[titleURL]
	return a, ok
}

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
	return selectPanelAudio(startflix.SourceName, titleURL, season)
}

// selectPanelAudio is selectStartFlixAudio with the picker titled after the
// source the title came from.
func selectPanelAudio(source, titleURL string, season startflix.Season) startflix.Audio {
	audios := season.Audios()
	if len(audios) == 1 {
		// Live panels are often single-audio (Dexter, Frieren and Round 6 are
		// Dublado-only): the substitution is silent by design, so it is logged
		// with the counts instead of looking like a free choice downstream.
		util.Debug("Panel season has a single audio, taking it without asking",
			"source", source, "season", season.Number, "audio", audios[0],
			"dubbed", len(season.Dubbed), "subtitled", len(season.Subtitled))
		return audios[0]
	}
	has := func(a startflix.Audio) bool { return len(season.Episodes(a)) > 0 }
	if a, ok := recallStartFlixAudio(titleURL); ok && has(a) {
		return a
	}
	if a, ok := pinnedStartFlixAudio(); ok {
		if has(a) {
			return a
		}
		// A pinned --audio the panel does not carry for this season (Solo
		// Leveling S1 has no Legendado) must not silently become the other
		// audio in batch runs: log it so the download manifest is explainable.
		util.Debug("Pinned audio is not on this panel season, falling through to the picker",
			"source", source, "season", season.Number, "pinned", a,
			"dubbed", len(season.Dubbed), "subtitled", len(season.Subtitled))
	}

	labels := []string{
		fmt.Sprintf("Dublado (%s)", episodeCountLabel(len(season.Dubbed))),
		fmt.Sprintf("Legendado (%s)", episodeCountLabel(len(season.Subtitled))),
	}
	choice := startflix.AudioDubbed
	if idx, err := sfxPickFn(source+" > Você quer assistir dublado ou legendado? ", labels); err == nil && idx == 1 {
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
		panel, err = panelForMedia(ctx, c, media)
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

// panelForMedia finds a title's panel: from its StartFlix title page, or by
// the TMDB id a TopCine/Pobreflix page names.
func panelForMedia(ctx context.Context, c *startflix.Client, media *models.Anime) (startflix.Panel, error) {
	if isTopCine(media) {
		return topCinePanel(ctx, media)
	}
	if isPobreflix(media) {
		return pobreflixPanel(ctx, media)
	}
	return c.Panel(ctx, media.URL)
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
		return nil, fmt.Errorf("no %s page for this title", panelSourceName(media))
	}
	c := sfxClientFn()

	panel, err := loadStartFlixPanel(c, media)
	if err != nil {
		markPanelDeadErr(media, err)
		return nil, describeMediaErr(media, err)
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
		markPanelDeadErr(media, err)
		return nil, describeMediaErr(media, err)
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

// ErrStartFlixNotSeries is returned when a series operation meets a movie
// (StartFlix or TopCine).
var ErrStartFlixNotSeries = errors.New("this title is a movie, not a series")

// loadStartFlixSeries opens a title's panel and its seasons, refusing movies.
func loadStartFlixSeries(media *models.Anime) (startflix.Panel, []startflix.Season, error) {
	if media == nil || media.URL == "" {
		return startflix.Panel{}, nil, fmt.Errorf("no %s page for this title", panelSourceName(media))
	}
	c := sfxClientFn()
	panel, err := loadStartFlixPanel(c, media)
	if err != nil {
		markPanelDeadErr(media, err)
		return startflix.Panel{}, nil, describeMediaErr(media, err)
	}
	if panel.Kind == startflix.KindMovie {
		return startflix.Panel{}, nil, ErrStartFlixNotSeries
	}
	seasons, err := loadStartFlixSeasons(c, panel)
	if err != nil {
		markPanelDeadErr(media, err)
		return startflix.Panel{}, nil, describeMediaErr(media, err)
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
	return nil, fmt.Errorf("season %d is not on %s for %q (seasons: %s)", seasonNum, panelSourceName(media), media.Name, strings.Join(available, ", "))
}

// startFlixSeasonEpisodes builds a season's episode list in the audio the
// user picks, and records the season as the title's current one.
func startFlixSeasonEpisodes(media *models.Anime, panel startflix.Panel, season startflix.Season) []models.Episode {
	audio := selectPanelAudio(panelSourceName(media), media.URL, season)
	media.CurrentSeason = season.Number
	rememberListedAudio(media.URL, season.Key(), audio)

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
	if isTopCine(media) {
		fillTopCineEpisodeTitles(media, episodes)
	}
	if isPobreflix(media) {
		fillPobreflixEpisodeTitles(media, episodes)
	}
	util.Debug("Panel episodes loaded", "source", media.Source, "season", season.Number, "audio", audio, "count", len(episodes))
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

// isMoviePanelURL reports whether a players URL is a movie's panel
// (/filme/<imdb>) rather than one episode's endpoint.
func isMoviePanelURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Host != "" && strings.HasPrefix(u.Path, "/filme/")
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
// with burned-in subtitles; either way the audio the stream carries decides.
func audioLangForStartFlix(a startflix.Audio) string {
	if a == startflix.AudioSubtitled {
		return originalALang
	}
	return portugueseALang
}

// wantedStreamAudio is the audio the user is watching a title in: for a
// series, the list the episode came from (or the one picked for the title);
// otherwise the pinned --audio, and the dub when nothing says — what a movie
// has always been played in.
func wantedStreamAudio(media *models.Anime, episode *models.Episode) startflix.Audio {
	if media.MediaType != models.MediaTypeMovie {
		season := episode.SeasonID
		if season == "" && media.CurrentSeason > 0 {
			season = strconv.Itoa(media.CurrentSeason)
		}
		if a, ok := recallListedAudio(media.URL, season); ok {
			return a
		}
		if a, ok := recallStartFlixAudio(media.URL); ok {
			return a
		}
	}
	if a, ok := pinnedStartFlixAudio(); ok {
		return a
	}
	return startflix.AudioDubbed
}

// ErrAudioUnavailable is returned when no server GoAnime can play carries the
// audio the user is watching in, and the user did not take the other audio
// that does play.
var ErrAudioUnavailable = errors.New("the chosen audio is not on any server GoAnime can play")

// audioSwap is the question asked before a stream in another audio plays.
type audioSwap struct {
	Source string
	Title  string
	Want   startflix.Audio
	Have   startflix.Audio
	// Why explains, one line each, why the wanted audio does not play.
	Why []string
}

// sfxAudioSwapFn asks whether to play the other audio; a seam for tests.
// Without a terminal tui.Pick fails, and the swap is refused.
var sfxAudioSwapFn = func(swap audioSwap) (bool, error) {
	idx, err := tui.Pick([]tui.PickItem{
		{Label: audioSwapPlayLabel(swap.Have)},
		{Label: "Back"},
	}, tui.PickOptions{
		Breadcrumb:   swap.Source + " > " + string(swap.Want) + " unavailable",
		WindowTitle:  "GoAnime - " + swap.Source,
		ItemSingular: "option",
		ItemPlural:   "options",
		Panel: &tui.PickPanel{
			Label: string(swap.Want) + " unavailable",
			Gloss: "音声",
			Title: tui.SingleLine(swap.Title),
			Lines: swap.Why,
		},
	})
	if err != nil {
		return false, err
	}
	return idx == 0, nil
}

func audioSwapPlayLabel(have startflix.Audio) string {
	if have == startflix.AudioSubtitled {
		return "Play the original audio"
	}
	return "Play the Portuguese dub"
}

// acceptAudioSwap decides whether a stream in another audio than the user's
// may play. Never silently: Todo Poderoso (2003) on TopCine is listed
// Dublado, every Dublado server the panel names is dead, and the vidsrc
// fallback plays its English original — with mpv told to look for a
// Portuguese track it does not have. So the user is asked, once per title.
// The one silent case is a listing that proves the wanted audio absent (a
// Legendado-only movie): there was never a choice, as with a single-audio
// season.
func acceptAudioSwap(media *models.Anime, want, have startflix.Audio, offered []startflix.Audio, stream *startflix.Stream, panelErr error) error {
	if len(offered) > 0 && !slices.Contains(offered, want) {
		util.Debug("Panel lists this title in another audio only, taking it without asking",
			"title", media.Name, "want", want, "have", have, "offered", offered)
		return nil
	}
	if a, ok := recallAudioSwap(media.URL); ok && a == have {
		return nil
	}
	tcLine := topCineAudioLine(media, want)
	if tcLine == "" {
		tcLine = pobreflixGateLine(media.URL)
	}
	swap := audioSwap{
		Source: panelSourceName(media),
		Title:  media.Name,
		Want:   want,
		Have:   have,
		Why:    audioSwapReasons(media, want, have, stream, panelErr, tcLine),
	}
	ok, err := sfxAudioSwapFn(swap)
	if err == nil && ok {
		util.Debug("User took another audio", "title", media.Name, "want", want, "have", have)
		rememberAudioSwap(media.URL, have)
		return nil
	}
	cause := fmt.Errorf("%w: %s plays, %s does not", ErrAudioUnavailable, have, want)
	msg := fmt.Sprintf("No server GoAnime can play has this title %s right now, so the %s was not played in its place.",
		want, audioDescription(have))
	if err != nil && !errors.Is(err, tui.ErrPickBack) && !errors.Is(err, tui.ErrPickCancelled) {
		util.Debug("Audio swap question unavailable, refusing the swap", "err", err)
		msg += " Pass --audio " + audioFlagFor(have) + " to play it anyway."
	}
	if tcLine != "" {
		msg += " " + tcLine
	}
	return &friendlyError{cause: cause, msg: msg}
}

// audioSwapReasons explains the swap: what failed, and what plays instead.
func audioSwapReasons(media *models.Anime, want, have startflix.Audio, stream *startflix.Stream, panelErr error, tcLine string) []string {
	source := panelSourceName(media)
	line := fmt.Sprintf("%s lists this title %s, but none of the %s servers GoAnime plays it from worked.", source, want, want)
	var noStream *startflix.NoStreamError
	if errors.As(panelErr, &noStream) && len(noStream.Unsupported) > 0 {
		line += " Listed: " + strings.Join(noStream.Unsupported, ", ") + "."
	}
	why := []string{line}
	if tcLine != "" {
		why = append(why, tcLine)
	}
	what := "What plays is the " + audioDescription(have)
	if have == startflix.AudioSubtitled {
		if len(stream.Subtitles) > 0 {
			what += ", with Portuguese subtitles"
		} else {
			what += ", without Portuguese subtitles"
		}
	}
	return append(why, what+".")
}

func audioDescription(a startflix.Audio) string {
	if a == startflix.AudioSubtitled {
		return "original audio"
	}
	return "Portuguese dub"
}

// audioFlagFor is the --audio value that asks for a in advance.
func audioFlagFor(a startflix.Audio) string {
	if a == startflix.AudioSubtitled {
		return "eng"
	}
	return "pt-BR"
}

// GetStartFlixStreamURL resolves an episode or movie to a playable URL and
// hands mpv/the downloader the referer, subtitles and audio preference.

// sfxVIDSRCFallbackFn resolves a panel title by its ids on the independent
// vidsrc embed API after the panel's players failed. A seam for tests:
// production reaches the real fallback through the startflix provider.
// Movies take IMDb → TMDB; series take the TMDB id plus the season the user
// picked (media.CurrentSeason, set when the episode list was built) and the
// episode's number.
var sfxVIDSRCFallbackFn = func(ctx context.Context, media *models.Anime, episode *models.Episode) (*startflix.Stream, error) {
	if media.MediaType == models.MediaTypeMovie {
		return startflix.MovieFallback(ctx, media.IMDBID, media.TMDBID)
	}
	season := media.CurrentSeason
	if season <= 0 && episode != nil {
		if n, err := strconv.Atoi(episode.SeasonID); err == nil {
			season = n
		}
	}
	return startflix.EpisodeFallback(ctx, media.TMDBID, season, sfxEpisodeNumber(episode))
}

// sfxEpisodeNumber reads one episode's number, 0 when indeterminate (the
// fallback then plays whatever the source names as Pilot).
func sfxEpisodeNumber(episode *models.Episode) int {
	if episode == nil {
		return 0
	}
	if episode.Num > 0 {
		return episode.Num
	}
	n, _ := strconv.Atoi(episode.Number)
	return n
}

func GetStartFlixStreamURL(media *models.Anime, episode *models.Episode, _ string) (string, error) {
	if media == nil || episode == nil {
		return "", fmt.Errorf("no %s title or episode to play", panelSourceName(media))
	}
	c := sfxClientFn()

	playersURL, err := startFlixPlayersURL(c, media, episode)
	if err != nil {
		markPanelDeadErr(media, err)
		return "", describeMediaErr(media, err)
	}

	want := wantedStreamAudio(media, episode)
	var stream *startflix.Stream
	var offered []startflix.Audio
	panelExhausted := false
	runWithSpinner("Loading stream...", func() {
		ctx, cancel := context.WithTimeout(context.Background(), sfxStreamBudget)
		defer cancel()
		var players []startflix.Player
		if players, err = c.Players(ctx, playersURL); err == nil {
			offered = startflix.ListedAudios(players, want)
			stream, err = c.ResolveStreamFor(ctx, players, want)
		}
		if err != nil {
			panelExhausted = true
		}
	})
	panelErr := err
	if err != nil && panelExhausted {
		// The panel could not serve this title's players. The fallback
		// resolves the same title by its ids on an independent embed API,
		// so a stale panel record no longer means a dead title.
		//
		// It gets its own budget, not what is left of the spinner's: the
		// panel path self-contains a yt-dlp sweep of every listed host, and
		// one hanging host can burn most of that budget before the fallback
		// gets a turn — the fallback then dies on a deadline it never had a
		// chance to meet. Fresh time, only spent when the panel missed.
		fsCtx, fsc := context.WithTimeout(context.Background(), sfxStreamBudget)
		defer fsc()
		if fs, ferr := sfxVIDSRCFallbackFn(fsCtx, media, episode); ferr == nil && fs != nil {
			// The fallback serves the original audio; a stream that does not
			// say so is taken as that, never as the dub.
			if fs.Audio == "" {
				fs.Audio = startflix.AudioSubtitled
			}
			util.Debug("VIDSRC fallback resolved the title", "host", fs.Host, "audio", fs.Audio)
			stream, err = fs, nil
		}
	}
	if err != nil {
		// A movie's players are listed on its panel itself, so "no content"
		// there is the whole title gone; for an episode it is one episode.
		// The hiding stays behind the fallback attempt: a title the fallback
		// saved must stay listed.
		if isMoviePanelURL(playersURL) {
			markPanelDeadErr(media, err)
		}
		// No prefix: the player dispatch already wraps this as "failed to get
		// StartFlix stream URL", and the friendly text should follow it directly.
		return "", describeMediaErr(media, err)
	}

	have := stream.Audio
	if have == "" {
		have = want
	}
	if have != want {
		if err := acceptAudioSwap(media, want, have, offered, stream, panelErr); err != nil {
			return "", err
		}
	}

	if stream.Referer != "" {
		util.SetGlobalReferer(stream.Referer)
	}
	// Always set when the user pinned nothing: the preference is a process-wide
	// global, and leaving it alone would carry the last title's Legendado choice
	// into the next movie. It follows the audio the stream carries, not the one
	// wanted: told to prefer Portuguese, mpv falls back to a lone English track
	// without a word.
	if startFlixPinnedAudioLang() == "" {
		util.SetGlobalAudioLanguage(audioLangForStartFlix(have))
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
