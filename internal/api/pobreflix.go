package api

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/pobreflix"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/startflix"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// Pobreflix flow. The catalog is plain HTTP: the site's /api search answers
// with the record's real name and slug, and the sitemap corpus is the
// fallback. Current watch pages embed their playable MP4 servers directly, so
// playback no longer needs a browser or the StartFlix panel:
//
//   - a movie page embeds servers.dub/leg[].url (public R2 MP4s);
//   - a series page embeds its episodes; each /assistir/…-t<season>-ep<n>.html
//     page embeds servers.dub/leg[].download, which redirects to the final MP4.
//
// Old-theme pages carry no embedded servers. Those titles keep the previous
// path: the page's TMDB id opens the StartFlix panel, which lists its players.

// pbClientFn is the Pobreflix client; a seam for tests.
var pbClientFn = pobreflix.Shared

// ErrPobreflixNoIMDb is returned when a Pobreflix movie's TMDB id links to no
// IMDb id, which the video panel keys movies by. It remains for the old-theme
// panel fallback.
var ErrPobreflixNoIMDb = errors.New("no IMDb id is linked to this Pobreflix movie")

func isPobreflix(media *models.Anime) bool {
	return media != nil && media.Source == pobreflix.SourceName
}

// loadPobreflixTitle opens a title page and makes its ids the title's. The
// page is the authority: enrichment may have guessed another title from the
// Portuguese name, so a disagreeing guess is dropped along with what it
// brought in (as adoptStartFlixIDs does).
// pbGatedHosts remembers, per title page, the hosts whose player page the
// title's servers all point at. acceptAudioSwap reads it to say why the dub
// the site lists did not play.
var (
	pbGatedMu    sync.Mutex
	pbGatedHosts = map[string][]string{}
)

func rememberPobreflixGate(titleURL string, hosts []string) {
	pbGatedMu.Lock()
	defer pbGatedMu.Unlock()
	if len(hosts) == 0 {
		delete(pbGatedHosts, titleURL)
		return
	}
	pbGatedHosts[titleURL] = hosts
}

// pobreflixGateLine explains a missing audio when Pobreflix's only servers for
// the title are player pages behind a browser check. "" otherwise.
func pobreflixGateLine(titleURL string) string {
	pbGatedMu.Lock()
	hosts := pbGatedHosts[titleURL]
	pbGatedMu.Unlock()
	if len(hosts) == 0 {
		return ""
	}
	return "Pobreflix's own player (" + strings.Join(hosts, ", ") +
		") asks for a browser check GoAnime can't pass."
}

// errPobreflixGated means every server the page lists is a player page, not a
// file. The caller falls back to the panel instead of playing a challenge page.
var errPobreflixGated = errors.New("pobreflix: every listed server is a player page behind a browser check")

// pobreflixFiles keeps the servers that are files. When the page lists servers
// but none is a file, it records why and returns errPobreflixGated.
func pobreflixFiles(titleURL string, servers []pobreflix.Server) ([]pobreflix.Server, error) {
	files := pobreflix.DirectServers(servers)
	if len(files) > 0 {
		rememberPobreflixGate(titleURL, nil)
		return files, nil
	}
	if hosts := pobreflix.GatedHosts(servers); len(hosts) > 0 {
		rememberPobreflixGate(titleURL, hosts)
		return nil, errPobreflixGated
	}
	return servers, nil
}

// pobreflixGateResolvable reports whether the page lists a player page that
// can be opened: with a registered Turnstile solver any of them, without one
// only a player on a host other than the Turnstile one (the RedeFlixApi
// player needs no solver). When neither holds this is false, so the previous
// panel fallback runs without a single extra request.
func pobreflixGateResolvable(servers []pobreflix.Server) bool {
	if len(pobreflix.GatedServers(servers)) == 0 {
		return false
	}
	return netx.TurnstileSolverAvailable() || pbClientFn().HasSolverFreePlayer(servers)
}

// pobreflixCanPlayDirectly reports whether the title can bypass the panel:
// either it embeds a file, or its player page can be cleared.
func pobreflixCanPlayDirectly(title pobreflix.Title) bool {
	return title.HasDirect() || pobreflixGateResolvable(title.Servers)
}

// pobreflixFilesHaveAudio reports whether the embedded files already carry the
// wanted audio. An unlabelled file counts, as the direct chooser takes it
// silently.
func pobreflixFilesHaveAudio(servers []pobreflix.Server, want pobreflix.Audio) bool {
	for _, s := range servers {
		if s.Direct() && (s.Audio == want || s.Audio == pobreflix.AudioUnknown) {
			return true
		}
	}
	return false
}

func loadPobreflixTitle(ctx context.Context, media *models.Anime) (pobreflix.Title, error) {
	title, err := pbClientFn().Title(ctx, media.URL)
	if err != nil {
		return pobreflix.Title{}, err
	}
	if title.TMDBID != media.TMDBID {
		if media.TMDBID > 0 {
			util.Debug("Pobreflix TMDB id replaces a name-search match", "title", media.Name,
				"guessed_tmdb", media.TMDBID, "tmdb", title.TMDBID)
		}
		media.TMDBID, media.IMDBID = title.TMDBID, ""
		media.TMDBDetails, media.Rating, media.Overview, media.Genres, media.Runtime = nil, 0, "", nil, 0
	}
	if title.Kind == pobreflix.KindMovie {
		media.MediaType = models.MediaTypeMovie
	} else if media.MediaType == models.MediaTypeMovie || media.MediaType == "" {
		media.MediaType = models.MediaTypeTV
	}
	if title.Name != "" && media.Name == "" {
		media.Name = title.Name
	}
	if title.Kind == pobreflix.KindMovie {
		_, _ = pobreflixFiles(media.URL, title.Servers)
	}
	return title, nil
}

// pobreflixPanel opens a Pobreflix title page and returns the panel its TMDB
// id names. Used only when the page embeds no playable servers (the old
// theme) or the embedded servers cannot be resolved.
func pobreflixPanel(ctx context.Context, media *models.Anime) (startflix.Panel, error) {
	title, err := loadPobreflixTitle(ctx, media)
	if err != nil {
		return startflix.Panel{}, err
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
	var noStream *startflix.NoStreamError
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrPobreflixNoIMDb):
		return &friendlyError{cause: err, msg: "GoAnime couldn't link this Pobreflix movie to the servers it plays from. Try searching it on StartFlix."}
	case errors.Is(err, pobreflix.ErrNoEpisode):
		return &friendlyError{cause: err, msg: "Pobreflix has this episode listed but its playable files are missing right now. Try another episode, or come back later."}
	case errors.As(err, &noStream) && len(noStream.Failures) == 0 && len(noStream.Unsupported) > 0:
		return &friendlyError{cause: err, msg: pobreflixDeadServersMessage(media, noStream.Unsupported)}
	case errors.Is(err, startflix.ErrNotOnPanel), errors.Is(err, startflix.ErrNoPanel):
		return &friendlyError{cause: err, msg: "Pobreflix has this title only on servers that don't carry it. Try searching it on StartFlix or another source."}
	default:
		return describePanelErr(pobreflix.SourceName, err)
	}
}

// pobreflixDeadServersMessage explains a title whose every panel player sits
// on a host that no longer answers — old titles whose panel record predates
// the panel's current player lineup ("Todo Poderoso (2003)": actelecup.com,
// letsupload.co, fembed and friends, all gone as of 2026-10-17). The panel's
// record is upstream-maintained and gets re-pointed as its players die, so
// the honest advice is to give up for today, not to follow the generic
// "try another episode" — which for a movie reads as nonsense. Nothing is
// hidden here: the title stays in search results (markPanelDeadErr).
func pobreflixDeadServersMessage(media *models.Anime, unsupported []string) string {
	what := "title"
	if media.MediaType == models.MediaTypeMovie {
		what = "movie"
	}
	return "Pobreflix has this " + what + " only on servers that no longer answer (" +
		strings.Join(unsupported, ", ") + "). Try again later, or search it on another source."
}

// fillPobreflixEpisodeTitles names the episodes the panel lists without a
// name, from Pobreflix's own episode list. Best effort: the title page is
// already cached in the client, and the page lists the episodes with
// Portuguese names.
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
// episodes of the season and audio the user picks. Current pages are read
// directly; an old-theme page falls back to the StartFlix panel.
func GetPobreflixEpisodes(media *models.Anime) ([]models.Episode, error) {
	if media == nil || media.URL == "" {
		return nil, fmt.Errorf("no Pobreflix page for this title")
	}
	var (
		title pobreflix.Title
		err   error
	)
	runWithSpinner("Loading title...", func() {
		ctx, cancel := context.WithTimeout(context.Background(), sfxListBudget)
		defer cancel()
		title, err = loadPobreflixTitle(ctx, media)
	})
	if err == nil && !title.HasDirect() && title.Kind == pobreflix.KindMovie && pobreflixGateResolvable(title.Servers) {
		// The movie has no embedded file, but its player page can be cleared.
		// Listing it here keeps the direct playback path (and its real audio
		// check) instead of sending it through the panel first.
		return []models.Episode{{
			Number: "1",
			Num:    1,
			URL:    media.URL,
			Title:  models.TitleDetails{English: media.Name, Romaji: media.Name},
		}}, nil
	}
	if err == nil && title.HasDirect() {
		if title.Kind == pobreflix.KindMovie {
			return []models.Episode{{
				Number: "1",
				Num:    1,
				URL:    media.URL,
				Title:  models.TitleDetails{English: media.Name, Romaji: media.Name},
			}}, nil
		}
		episodes, directErr := pobreflixDirectEpisodes(media, title)
		if directErr == nil {
			return episodes, nil
		}
		if errors.Is(directErr, tui.ErrPickBack) || errors.Is(directErr, tui.ErrPickCancelled) {
			return nil, ErrBackToSearch
		}
		util.Debug("Pobreflix direct listing failed, falling back to the panel", "err", directErr)
	}
	return GetStartFlixEpisodes(media)
}

// pobreflixDirectEpisodes lists a current-site series in the season and audio
// the user picks.
func pobreflixDirectEpisodes(media *models.Anime, title pobreflix.Title) ([]models.Episode, error) {
	seasons := title.Seasons
	if len(seasons) == 0 {
		return nil, fmt.Errorf("pobreflix: %q lists no episodes", media.Name)
	}
	keys := make([]string, len(seasons))
	panelSeasons := make([]startflix.Season, len(seasons))
	for i, s := range seasons {
		keys[i] = strconv.Itoa(s.Number)
		panelSeasons[i] = pobreflixSeasonAsPanel(s)
	}
	key, err := selectSeasonWith(sfxSeasonPickFn, media, keys, func() []tui.PickItem {
		return startFlixSeasonItems(panelSeasons)
	})
	if err != nil {
		return nil, err
	}
	number, _ := strconv.Atoi(key)
	season, ok := title.Season(number)
	if !ok {
		return nil, fmt.Errorf("season %d is not on Pobreflix for %q", number, media.Name)
	}
	return pobreflixDirectSeasonEpisodes(media, title, season), nil
}

// pobreflixDirectSeasonEpisodes builds a season's episode list in the audio
// the user picks and records the season as the title's current one.
func pobreflixDirectSeasonEpisodes(media *models.Anime, title pobreflix.Title, season pobreflix.Season) []models.Episode {
	audio := selectPobreflixAudio(media.URL, season)
	media.CurrentSeason = season.Number
	rememberListedAudio(media.URL, strconv.Itoa(season.Number), startflixAudio(audio))

	base := pbClientFn()
	list := make([]pobreflix.Episode, 0, len(season.Episodes))
	for _, ep := range season.Episodes {
		if ep.Audio != pobreflix.AudioUnknown && ep.Audio != audio {
			continue
		}
		list = append(list, ep)
	}
	if len(list) == 0 {
		list = season.Episodes
	}

	episodes := make([]models.Episode, 0, len(list))
	for _, ep := range list {
		pageURL := ep.URL
		if !pobreflix.IsEpisodePage(pageURL) {
			pageURL = base.EpisodeURL(title, season.Number, ep.Number)
		}
		episodes = append(episodes, models.Episode{
			Number:   strconv.Itoa(ep.Number),
			Num:      ep.Number,
			URL:      pageURL,
			DataID:   strconv.Itoa(title.InternalID),
			SeasonID: strconv.Itoa(season.Number),
			Title:    models.TitleDetails{English: ep.Title, Romaji: ep.Title},
		})
	}
	util.Debug("Pobreflix episodes loaded directly", "title", media.Name, "season", season.Number,
		"audio", audio, "count", len(episodes))
	return episodes
}

// selectPobreflixAudio picks the audio track to list. It reuses the panel's
// remembered/pinned choice and only asks when the season really has both.
func selectPobreflixAudio(titleURL string, season pobreflix.Season) pobreflix.Audio {
	panel := pobreflixSeasonAsPanel(season)
	audio := selectPanelAudio(pobreflix.SourceName, titleURL, panel)
	if audio == startflix.AudioSubtitled {
		return pobreflix.AudioSubtitled
	}
	return pobreflix.AudioDubbed
}

// pobreflixSeasonAsPanel converts a current-site season to the panel's
// audio-grouped shape, so the shared season/audio pickers can render it.
func pobreflixSeasonAsPanel(s pobreflix.Season) startflix.Season {
	out := startflix.Season{Number: s.Number}
	for _, ep := range s.Episodes {
		pe := startflix.Episode{Number: ep.Number, Title: ep.Title, Audio: startflix.AudioDubbed}
		if ep.Audio == pobreflix.AudioSubtitled {
			pe.Audio = startflix.AudioSubtitled
			out.Subtitled = append(out.Subtitled, pe)
			continue
		}
		out.Dubbed = append(out.Dubbed, pe)
	}
	return out
}

// startflixAudio maps Pobreflix's audio label to the panel's.
func startflixAudio(a pobreflix.Audio) startflix.Audio {
	if a == pobreflix.AudioSubtitled {
		return startflix.AudioSubtitled
	}
	return startflix.AudioDubbed
}

// GetPobreflixSeasonNumbers lists a series' season numbers, for downloads
// that name the seasons rather than pick one.
func GetPobreflixSeasonNumbers(media *models.Anime) ([]int, error) {
	if media == nil || media.URL == "" {
		return nil, fmt.Errorf("no Pobreflix page for this title")
	}
	ctx, cancel := context.WithTimeout(context.Background(), sfxListBudget)
	defer cancel()
	title, err := loadPobreflixTitle(ctx, media)
	if err == nil && title.Kind == pobreflix.KindSeries && title.HasDirect() {
		nums := make([]int, 0, len(title.Seasons))
		for _, s := range title.Seasons {
			nums = append(nums, s.Number)
		}
		return nums, nil
	}
	return GetStartFlixSeasonNumbers(media)
}

// GetPobreflixSeasonEpisodes lists one season of a series without the season
// picker; see GetStartFlixSeasonEpisodes.
func GetPobreflixSeasonEpisodes(media *models.Anime, seasonNum int) ([]models.Episode, error) {
	if media == nil || media.URL == "" {
		return nil, fmt.Errorf("no Pobreflix page for this title")
	}
	ctx, cancel := context.WithTimeout(context.Background(), sfxListBudget)
	defer cancel()
	title, err := loadPobreflixTitle(ctx, media)
	if err == nil && title.Kind == pobreflix.KindSeries && title.HasDirect() {
		if season, ok := title.Season(seasonNum); ok {
			return pobreflixDirectSeasonEpisodes(media, title, season), nil
		}
		available := make([]string, 0, len(title.Seasons))
		for _, s := range title.Seasons {
			available = append(available, strconv.Itoa(s.Number))
		}
		return nil, fmt.Errorf("season %d is not on Pobreflix for %q (seasons: %s)",
			seasonNum, media.Name, strings.Join(available, ", "))
	}
	return GetStartFlixSeasonEpisodes(media, seasonNum)
}

// GetPobreflixStreamURL resolves an episode or movie to a playable URL and
// hands mpv/the downloader the audio preference. Current pages play their
// embedded MP4s directly; old-theme pages fall back to the panel. It clears
// what a previous stream left behind, then publishes this one's context.
func GetPobreflixStreamURL(media *models.Anime, episode *models.Episode, quality string) (string, error) {
	stream, err := ResolvePobreflixStream(media, episode, quality)
	if err != nil {
		return "", err
	}
	util.ClearGlobalReferer()
	util.ClearGlobalSubtitles()
	stream.Publish()
	return stream.URL, nil
}

// ResolvePobreflixStream is GetPobreflixStreamURL without the side effects: the
// referer and the audio hint come back in the Stream.
func ResolvePobreflixStream(media *models.Anime, episode *models.Episode, quality string) (source.Stream, error) {
	if media == nil || episode == nil {
		return source.Stream{}, fmt.Errorf("no Pobreflix title or episode to play")
	}
	ctx, cancel := context.WithTimeout(context.Background(), sfxStreamBudget)
	defer cancel()
	title, err := loadPobreflixTitle(ctx, media)
	if err == nil && pobreflixCanPlayDirectly(title) {
		var stream source.Stream
		if title.Kind == pobreflix.KindMovie {
			stream, err = pobreflixDirectMovieURL(ctx, media, title, quality)
		} else {
			stream, err = pobreflixDirectEpisodeURL(ctx, media, title, episode, quality)
		}
		if err == nil && stream.URL != "" {
			return stream, nil
		}
		if err != nil {
			util.Debug("Pobreflix direct playback failed, falling back to the panel", "err", err)
		}
	}
	return ResolveStartFlixStream(media, episode, quality)
}

// pobreflixDirectMovieURL returns a movie's file, from the page's own servers
// or, when only player pages are listed, from behind the Turnstile gate.
func pobreflixDirectMovieURL(ctx context.Context, media *models.Anime, title pobreflix.Title, quality string) (source.Stream, error) {
	want := pobreflixAudio(wantedStreamAudio(media, nil))
	files, filesErr := pobreflixFiles(media.URL, title.Servers)
	if filesErr == nil && pobreflixFilesHaveAudio(files, want) {
		return pobreflixPlayMovieFile(ctx, media, files, want, quality)
	}

	var gateErr error
	if pobreflixGateResolvable(title.Servers) {
		target := pobreflix.PlayerTarget{TMDBID: title.TMDBID, Kind: pobreflix.KindMovie}
		if stream, err := pobreflixResolveGate(ctx, media, title.Servers, target, want, quality); err == nil && stream.URL != "" {
			return stream, nil
		} else if err != nil {
			gateErr = err
			util.Debug("Pobreflix player gate did not resolve the movie", "title", media.Name, "err", err)
		}
	}
	if filesErr != nil {
		if gateErr != nil {
			return source.Stream{}, gateErr
		}
		return source.Stream{}, filesErr
	}
	return pobreflixPlayMovieFile(ctx, media, files, want, quality)
}

// pobreflixPlayMovieFile picks the movie's embedded server and returns its
// public MP4. It is also the fallback when the gate could not be cleared.
func pobreflixPlayMovieFile(ctx context.Context, media *models.Anime, files []pobreflix.Server, want pobreflix.Audio, quality string) (source.Stream, error) {
	server, have, err := choosePobreflixServer(files, want, quality, media)
	if err != nil {
		return source.Stream{}, err
	}
	rawURL := server.URL
	if rawURL == "" {
		rawURL = server.Download
	}
	if rawURL == "" {
		return source.Stream{}, fmt.Errorf("pobreflix: movie %q has no playable file", media.Name)
	}
	rawURL = pobreflixResolveDownload(ctx, rawURL)
	return pobreflixStream(rawURL, have, ""), nil
}

// pobreflixResolveGate answers the Turnstile on the title's player page and
// returns the media URL behind it. It asks the usual audio question if the
// player serves a different track than the one wanted.
func pobreflixResolveGate(ctx context.Context, media *models.Anime, servers []pobreflix.Server, target pobreflix.PlayerTarget, want pobreflix.Audio, quality string) (source.Stream, error) {
	// Every listed player is a candidate, best first. One that cannot work here
	// (it needs a solver and none is installed) is skipped and one that fails
	// is not the end of the title: the next is tried.
	var stream pobreflix.Stream
	var failure error
	resolved := false
	for _, server := range rankPobreflixGatedServers(servers, want, quality) {
		if err := ctx.Err(); err != nil {
			return source.Stream{}, err
		}
		s, err := resolvePobreflixPlayer(ctx, server, target, want)
		if errors.Is(err, errPobreflixGated) {
			continue // needs a solver that is not installed
		}
		if err != nil {
			failure = err
			util.Debug("Pobreflix player did not resolve; trying the next one", "player", server.Name, "err", err)
			continue
		}
		stream, resolved = s, true
		break
	}
	if !resolved {
		if failure != nil {
			return source.Stream{}, failure
		}
		return source.Stream{}, errPobreflixGated
	}
	have := stream.Audio
	if have == pobreflix.AudioUnknown {
		have = want
	}
	if have != want {
		swap := audioSwap{
			Source: pobreflix.SourceName,
			Title:  media.Name,
			Want:   startflixAudio(want),
			Have:   startflixAudio(have),
			Why: []string{fmt.Sprintf(
				"Pobreflix lists this title %s, but its player served the %s.",
				want, have)},
		}
		accepted, err := sfxAudioSwapFn(swap)
		if err != nil || !accepted {
			return source.Stream{}, &friendlyError{
				cause: fmt.Errorf("%w: %s plays, %s does not", ErrAudioUnavailable, have, want),
				msg: fmt.Sprintf("No Pobreflix server has this title %s right now, so the %s was not played in its place.",
					want, audioDescription(startflixAudio(have))),
			}
		}
	}
	util.Debug("Pobreflix stream obtained through its player", "tmdb", target.TMDBID, "audio", have)
	return pobreflixStream(stream.URL, have, stream.Referer), nil
}

// resolvePobreflixPlayer opens one player page: the RedeFlix player needs no
// solver; a page in the Turnstile format needs the registered one, and without
// it the answer is errPobreflixGated so the caller moves on.
func resolvePobreflixPlayer(ctx context.Context, server pobreflix.Server, target pobreflix.PlayerTarget, want pobreflix.Audio) (pobreflix.Stream, error) {
	pb := pbClientFn()
	if pb.IsSolverFreePlayer(server.URL) {
		stream, err := pb.ResolveRedeflix(ctx, server.URL, want)
		if err == nil || !errors.Is(err, pobreflix.ErrNotRedeflixPlayer) {
			return stream, err
		}
	}
	if !netx.TurnstileSolverAvailable() {
		return pobreflix.Stream{}, errPobreflixGated
	}
	return pb.ResolvePlayer(ctx, server.URL, target, want)
}

// rankPobreflixGatedServers lists the player pages most likely to carry want
// first, then the best quality, one per URL: the site lists the same player for
// both audios, and the track is checked after the page is opened.
func rankPobreflixGatedServers(servers []pobreflix.Server, want pobreflix.Audio, quality string) []pobreflix.Server {
	type scored struct {
		server pobreflix.Server
		score  int
	}
	var all []scored
	seen := map[string]bool{}
	for _, s := range servers {
		if !s.Gated() || seen[s.URL] {
			continue
		}
		seen[s.URL] = true
		score := pobreflix.QualityRank(s.Quality)
		if s.Audio == want {
			score += 100000
		}
		if quality != "" && !strings.EqualFold(quality, "best") && strings.EqualFold(s.Quality, quality) {
			score += 1000000
		}
		if s.URL != "" {
			score += 100
		}
		all = append(all, scored{s, score})
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].score > all[j].score })
	out := make([]pobreflix.Server, len(all))
	for i, sc := range all {
		out[i] = sc.server
	}
	return out
}

// pobreflixResolveDownload follows only the site's own download endpoints;
// public R2 MP4s are already final and a HEAD to them is wasted latency.
func pobreflixResolveDownload(ctx context.Context, rawURL string) string {
	if !strings.Contains(rawURL, "/api?") {
		return rawURL
	}
	return pbClientFn().ResolveDownload(ctx, rawURL)
}

// pobreflixDirectEpisodeURL opens the episode page and resolves its embedded
// download endpoint to the final MP4.
func pobreflixDirectEpisodeURL(ctx context.Context, media *models.Anime, title pobreflix.Title, episode *models.Episode, quality string) (source.Stream, error) {
	season := media.CurrentSeason
	if season <= 0 {
		season, _ = strconv.Atoi(episode.SeasonID)
	}
	number := sfxEpisodeNumber(episode)
	if season <= 0 || number <= 0 {
		return source.Stream{}, fmt.Errorf("pobreflix: episode %q has no season/number", episode.Number)
	}
	pageURL := episode.URL
	if !pobreflix.IsEpisodePage(pageURL) {
		pageURL = pbClientFn().EpisodeURL(title, season, number)
	}
	if pageURL == "" {
		return source.Stream{}, fmt.Errorf("pobreflix: no episode page for season %d episode %d", season, number)
	}
	data, err := pbClientFn().Episode(ctx, pageURL)
	if err != nil {
		return source.Stream{}, err
	}
	want := pobreflixAudio(wantedStreamAudio(media, episode))
	files, filesErr := pobreflixFiles(media.URL, data.Servers)
	if filesErr == nil && pobreflixFilesHaveAudio(files, want) {
		return pobreflixPlayEpisodeFile(ctx, files, want, quality, media, season, number)
	}

	var gateErr error
	if pobreflixGateResolvable(data.Servers) {
		tmdb := data.Series.TMDBID
		if tmdb <= 0 {
			tmdb = title.TMDBID
		}
		target := pobreflix.PlayerTarget{
			TMDBID:  tmdb,
			Kind:    pobreflix.KindSeries,
			Season:  season,
			Episode: number,
		}
		if stream, err := pobreflixResolveGate(ctx, media, data.Servers, target, want, quality); err == nil && stream.URL != "" {
			return stream, nil
		} else if err != nil {
			gateErr = err
			util.Debug("Pobreflix player gate did not resolve the episode", "title", media.Name,
				"season", season, "episode", number, "err", err)
		}
	}
	if filesErr != nil {
		if gateErr != nil {
			return source.Stream{}, gateErr
		}
		return source.Stream{}, filesErr
	}
	return pobreflixPlayEpisodeFile(ctx, files, want, quality, media, season, number)
}

// pobreflixPlayEpisodeFile picks an embedded episode server and resolves its
// download endpoint. It is also the fallback when the gate could not be
// cleared.
func pobreflixPlayEpisodeFile(ctx context.Context, files []pobreflix.Server, want pobreflix.Audio, quality string, media *models.Anime, season, number int) (source.Stream, error) {
	server, have, err := choosePobreflixServer(files, want, quality, media)
	if err != nil {
		return source.Stream{}, err
	}
	rawURL := server.Download
	if rawURL == "" {
		rawURL = server.URL
	}
	if rawURL == "" {
		return source.Stream{}, fmt.Errorf("pobreflix: episode %d has no playable file", number)
	}
	rawURL = pobreflixResolveDownload(ctx, rawURL)
	util.Debug("Pobreflix episode stream", "season", season, "episode", number, "audio", have, "quality", server.Quality)
	return pobreflixStream(rawURL, have, ""), nil
}

// choosePobreflixServer picks the best server for the wanted audio and
// quality. When the wanted audio is absent, it follows the panel's rule: a
// single-audio listing is taken silently; a real choice is asked once.
func choosePobreflixServer(servers []pobreflix.Server, want pobreflix.Audio, quality string, media *models.Anime) (pobreflix.Server, pobreflix.Audio, error) {
	if len(servers) == 0 {
		return pobreflix.Server{}, "", fmt.Errorf("pobreflix: %q lists no playable servers", media.Name)
	}
	available := make([]pobreflix.Audio, 0, 2)
	has := func(a pobreflix.Audio) bool {
		for _, s := range servers {
			if s.Audio == a {
				return true
			}
		}
		return false
	}
	if has(pobreflix.AudioDubbed) {
		available = append(available, pobreflix.AudioDubbed)
	}
	if has(pobreflix.AudioSubtitled) {
		available = append(available, pobreflix.AudioSubtitled)
	}
	if len(available) == 0 {
		available = append(available, pobreflix.AudioUnknown)
	}

	have := want
	if !has(want) {
		if len(available) == 1 {
			have = available[0]
		} else {
			swap := audioSwap{
				Source: pobreflix.SourceName,
				Title:  media.Name,
				Want:   startflixAudio(want),
				Have:   startflixAudio(available[0]),
				Why: []string{fmt.Sprintf(
					"Pobreflix lists this title %s, but its %s files are missing right now.",
					want, want)},
			}
			ok, err := sfxAudioSwapFn(swap)
			if err != nil || !ok {
				return pobreflix.Server{}, "", &friendlyError{
					cause: fmt.Errorf("%w: %s plays, %s does not", ErrAudioUnavailable, available[0], want),
					msg: fmt.Sprintf("No Pobreflix server has this title %s right now, so the %s was not played in its place.",
						want, audioDescription(startflixAudio(available[0]))),
				}
			}
			have = available[0]
		}
	}

	best := pobreflix.Server{}
	bestScore := -1
	for _, s := range servers {
		if s.Audio != pobreflix.AudioUnknown && s.Audio != have {
			continue
		}
		score := pobreflix.QualityRank(s.Quality)
		if s.Audio == have {
			score += 100000
		}
		if quality != "" && !strings.EqualFold(quality, "best") && strings.EqualFold(s.Quality, quality) {
			score += 1000000
		}
		if s.URL != "" {
			score += 100
		}
		if score > bestScore {
			best, bestScore = s, score
		}
	}
	if bestScore < 0 {
		return pobreflix.Server{}, "", fmt.Errorf("pobreflix: no server for %s", have)
	}
	return best, have, nil
}

// pobreflixAudio maps the panel's audio label to Pobreflix's.
func pobreflixAudio(a startflix.Audio) pobreflix.Audio {
	if a == startflix.AudioSubtitled {
		return pobreflix.AudioSubtitled
	}
	return pobreflix.AudioDubbed
}

// pobreflixStream is a resolved Pobreflix URL with its playback context: the
// Referer a proof-bearing player stream requires, and the track the file
// carries so mpv can be told before it starts.
func pobreflixStream(url string, audio pobreflix.Audio, referer string) source.Stream {
	s := source.Stream{URL: url, Referer: referer}
	if startFlixPinnedAudioLang() == "" {
		s.AudioLang = audioLangForStartFlix(startflixAudio(audio))
	}
	return s
}
