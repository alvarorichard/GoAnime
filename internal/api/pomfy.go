package api

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/pomfy"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// Pomfy flow. The catalog is keyed by TMDB id and every watch page hands a
// Byse video code over; the streaming itself — the proof-of-work captcha, the
// AES-256-GCM seal — lives in the pomfy package. What the api layer adds is
// the shape the rest of GoAnime expects: a movie is one episode; a series
// lists its seasons and asks which to open (the shared season picker), with
// the available episode numbers and names read from Pomfy's keyless season
// API rather than from TMDB.

// pomfyClientFn is the Pomfy client; a seam for tests.
var PomfyClientFn = pomfy.Shared

// listBudget bounds the catalog and season requests of a listing.
const pomfyListBudget = 20 * time.Second

// streamBudget mirrors the panel's: one watch-sources call plus up to four
// Byse calls with a PoW solve in between.
const pomfyStreamBudget = sfxStreamBudget + 2*time.Second

// pomfyKindOf maps a media type onto the catalog's own spelling.
func pomfyKindOf(media *models.Anime) pomfy.MediaKind {
	if media.MediaType == models.MediaTypeTV {
		return pomfy.KindSeries
	}
	return pomfy.KindMovie
}

// pomfyTMDBID returns the title's TMDB id, parsable from the identifier URL
// (assistir/serie paths or a bare id) when the field is empty.
func pomfyTMDBID(media *models.Anime, episode *models.Episode) (string, bool) {
	if media.TMDBID > 0 {
		return strconv.Itoa(media.TMDBID), true
	}
	if id := pomfyPathID(media.URL); id != "" {
		return id, true
	}
	if episode != nil {
		if id := pomfyPathID(episode.URL); id != "" {
			return id, true
		}
	}
	return "", false
}

// pomfyPathID reads the resource's own TMDB id from a URL: the first number
// after the site's /assistir/ or /serie/ marker (the segments after it on a
// series page are the season and episode), or a bare numeric string.
func pomfyPathID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	for _, marker := range []string{"assistir/", "serie/"} {
		i := strings.Index(raw, marker)
		if i < 0 {
			continue
		}
		rest := raw[i+len(marker):]
		seg := rest
		if j := strings.IndexByte(seg, '/'); j >= 0 {
			seg = seg[:j]
		}
		if n, err := strconv.Atoi(strings.TrimSpace(seg)); err == nil && n > 0 {
			return strconv.Itoa(n)
		}
	}
	if n, err := strconv.Atoi(raw); err == nil && n > 0 {
		return strconv.Itoa(n)
	}
	return ""
}

func firstNonEmptyStr(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// GetPomfyEpisodes lists a Pomfy title: the movie itself, or the episodes of
// the season the user picks. Every episode carries the TMDB id and the
// partner player's episode URL, which is all GetPomfyStreamURL re-derives
// playback from.
func GetPomfyEpisodes(media *models.Anime) ([]models.Episode, error) {
	if media == nil {
		return nil, fmt.Errorf("no Pomfy title to list")
	}
	tmdbID, ok := pomfyTMDBID(media, nil)
	if !ok {
		return nil, fmt.Errorf("%q has no TMDB id to open on Pomfy", media.Name)
	}
	kind := pomfyKindOf(media)
	ctx, cancel := context.WithTimeout(context.Background(), pomfyListBudget)
	defer cancel()

	if kind == pomfy.KindMovie {
		available, err := PomfyClientFn().HasTitle(ctx, atoint64(tmdbID), kind)
		if err != nil {
			return nil, describePomfyErr(err)
		}
		if !available {
			return nil, fmt.Errorf("%w: %s", pomfy.ErrNotOnCatalog, media.Name)
		}
		return []models.Episode{{
			Number:   "1",
			Num:      1,
			URL:      PomfyClientFn().MoviePlayURL(atoint64(tmdbID)),
			DataID:   tmdbID,
			SeasonID: "0",
			Title:    models.TitleDetails{English: media.Name, Romaji: media.Name},
		}}, nil
	}

	// Series: seasons and per-episode availability.
	info, err := PomfyClientFn().ShowInfo(ctx, atoint64(tmdbID))
	if err != nil || len(info.Seasons) == 0 {
		return nil, describePomfyErrLess(media, err)
	}
	nums := make([]string, 0, len(info.Seasons))
	items := make([]tui.PickItem, 0, len(info.Seasons))
	for _, s := range info.Seasons {
		if s.EpisodeCount <= 0 && s.Number > 1 {
			// TMDB lists empty season shells; specials (0) still go through,
			// episodes come from the season API and drop what is absent.
			continue
		}
		nums = append(nums, strconv.Itoa(s.Number))
		label := "Season " + strconv.Itoa(s.Number)
		if s.Name != "" {
			label = s.Name
		}
		detail := ""
		if s.EpisodeCount > 0 {
			detail = fmt.Sprintf("%d episodes", s.EpisodeCount)
		}
		items = append(items, tui.PickItem{
			Label:   label,
			Details: detail,
			Tags:    []string{"Pomfy"},
		})
	}
	if len(nums) == 0 {
		return nil, fmt.Errorf("pomfy: %q has no seasons", media.Name)
	}
	numStr, err := selectSeasonWith(sfxSeasonPickFn, media, nums, func() []tui.PickItem { return items })
	if err != nil {
		if errors.Is(err, tui.ErrPickBack) || errors.Is(err, tui.ErrPickCancelled) {
			return nil, ErrBackToSearch
		}
		return nil, fmt.Errorf("season selection cancelled: %w", err)
	}
	seasonNum, _ := strconv.Atoi(numStr)
	return pomfySeasonEpisodes(media, tmdbID, seasonNum)
}

// pomfySeasonEpisodes builds one season's episode list and records the season
// as the title's current one, the way the panel flow does.
func pomfySeasonEpisodes(media *models.Anime, tmdbID string, seasonNum int) ([]models.Episode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), pomfyListBudget)
	defer cancel()
	season, err := PomfyClientFn().Season(ctx, atoint64(tmdbID), seasonNum)
	if err != nil {
		return nil, describePomfyErr(err)
	}
	media.CurrentSeason = seasonNum
	episodes := make([]models.Episode, 0, len(season.Episodes))
	for _, ep := range season.Episodes {
		title := ep.Title
		if title == "" {
			title = fmt.Sprintf("Episode %d", ep.Number)
		}
		episodes = append(episodes, models.Episode{
			Number:   strconv.Itoa(ep.Number),
			Num:      ep.Number,
			URL:      pomfy.SeriesPlayURL(atoint64(tmdbID), seasonNum, ep.Number),
			DataID:   tmdbID,
			SeasonID: strconv.Itoa(seasonNum),
			Title:    models.TitleDetails{English: title, Romaji: title},
		})
	}
	if len(episodes) == 0 {
		return nil, fmt.Errorf("%w: season %d lists no episodes yet", pomfy.ErrNoEpisodeSource, seasonNum)
	}
	util.Debug("Pomfy episodes loaded", "tmdb", tmdbID, "season", seasonNum, "count", len(episodes))
	return episodes, nil
}

// GetPomfySeasonNumbers lists a series' season numbers, for downloads that
// name the seasons rather than pick one.
func GetPomfySeasonNumbers(media *models.Anime) ([]int, error) {
	if media == nil {
		return nil, fmt.Errorf("no Pomfy title")
	}
	tmdbID, ok := pomfyTMDBID(media, nil)
	if !ok {
		return nil, fmt.Errorf("%q has no TMDB id to open on Pomfy", media.Name)
	}
	ctx, cancel := context.WithTimeout(context.Background(), pomfyListBudget)
	defer cancel()
	info, err := PomfyClientFn().ShowInfo(ctx, atoint64(tmdbID))
	if err != nil {
		return nil, describePomfyErr(err)
	}
	nums := make([]int, 0, len(info.Seasons))
	for _, s := range info.Seasons {
		if s.EpisodeCount > 0 || s.Number <= 1 {
			nums = append(nums, s.Number)
		}
	}
	return nums, nil
}

// GetPomfySeasonEpisodes lists one season of a series without the picker.
func GetPomfySeasonEpisodes(media *models.Anime, seasonNum int) ([]models.Episode, error) {
	if media == nil {
		return nil, fmt.Errorf("no Pomfy title")
	}
	tmdbID, ok := pomfyTMDBID(media, nil)
	if !ok {
		return nil, fmt.Errorf("%q has no TMDB id to open on Pomfy", media.Name)
	}
	return pomfySeasonEpisodes(media, tmdbID, seasonNum)
}

// GetPomfyStreamURL resolves an episode or movie to its HLS master. The input
// it needs is exactly what episode listing stores: the TMDB id and, for
// series, the season/episode numbers — names ride along in the identifiers.
func GetPomfyStreamURL(media *models.Anime, episode *models.Episode, _ string) (string, error) {
	if media == nil || episode == nil {
		return "", errors.New("no Pomfy title or episode to play")
	}
	tmdbID, ok := pomfyTMDBID(media, episode)
	if !ok {
		return "", fmt.Errorf("%q has no TMDB id to play on Pomfy", media.Name)
	}
	id := atoint64(tmdbID)
	kind := pomfyKindOf(media)
	season, num := pomfySeasonAndEpisode(media, episode, kind)

	var stream *pomfy.Stream
	var err error
	runWithSpinner("Resolving stream (Pomfy)...", func() {
		sctx, cancel := context.WithTimeout(context.Background(), pomfyStreamBudget)
		defer cancel()
		watch := PomfyClientFn().BaseURL() + "/assistir/" + strconv.FormatInt(id, 10)
		stream, err = PomfyClientFn().ResolveStream(sctx, id, kind, season, num, watch)
	})
	if err != nil {
		return "", describePomfyErr(err)
	}
	if stream.Referer != "" {
		util.SetGlobalReferer(stream.Referer)
	}
	// The HLS master's URL spells the audio language its file carries
	// (…/lang/por/… on series); hint mpv before the first track plays.
	if lang := pomfyStreamLang(stream.URL); lang != "" {
		util.SetGlobalAudioLanguage(lang)
	}
	if len(stream.Subtitles) > 0 && !util.GlobalNoSubs {
		subs := make([]util.SubtitleInfo, 0, len(stream.Subtitles))
		for _, s := range stream.Subtitles {
			subs = append(subs, util.SubtitleInfo{URL: s.URL, Language: s.Language, Label: s.Label})
		}
		util.SetGlobalSubtitles(subs)
	}
	util.Debug("Pomfy stream URL obtained", "host", stream.Host, "url", stream.URL[:min(len(stream.URL), 80)])
	return stream.URL, nil
}

// pomfySeasonAndEpisode derives the season and episode numbers. Movies are
// always 1/1 (watch-sources ignores both); series read the episode's own
// season and number first — batch downloads hand over history-built
// episodes — and only then the media's current season.
func pomfySeasonAndEpisode(media *models.Anime, episode *models.Episode, kind pomfy.MediaKind) (int, int) {
	if kind != pomfy.KindSeries {
		return 0, 0
	}
	season, num := 0, 0
	if seasonStr := firstNonEmptyStr(episode.SeasonID, pomfyPathSeason(episode.URL)); seasonStr != "" {
		season, _ = strconv.Atoi(seasonStr)
	}
	if numStr := firstNonEmptyStr(episode.Number, strconv.Itoa(episode.Num)); numStr != "" && numStr != "0" {
		num, _ = strconv.Atoi(numStr)
	}
	if season <= 0 && media != nil {
		season = media.CurrentSeason
	}
	if season <= 0 {
		season = 1
	}
	if num <= 0 {
		num = 1
	}
	return season, num
}

// pomfyPathSeason reads the /serie/<id>/<season>/<episode> season segment.
func pomfyPathSeason(raw string) string {
	parts := strings.Split(strings.TrimRight(raw, "/"), "/")
	if len(parts) >= 3 && strings.Contains(raw, "/serie/") {
		return parts[len(parts)-2]
	}
	return ""
}

// pomfyStreamLang reads the JSON-ish HLS path's audio language, spelled one
// compo in: the masters label series audio like ",c_x,lang/por/c_por,
// .urlset" — no slash in front of "lang" (the leading one belongs to the
// numbering segment), so the scan anchors on "lang/" itself.
func pomfyStreamLang(raw string) string {
	const marker = "lang/"
	i := strings.Index(raw, marker)
	if i < 0 {
		return ""
	}
	rest := raw[i+len(marker):]
	end := strings.IndexByte(rest, '/')
	if end < 0 {
		end = len(rest)
	}
	lang := strings.ToLower(rest[:end])
	switch {
	case strings.HasPrefix(lang, "pt"):
		return "por"
	case strings.HasPrefix(lang, "ja"):
		return "jpn"
	case strings.HasPrefix(lang, "en"):
		return "eng"
	case len(lang) == 3:
		for _, r := range lang {
			if r < 'a' || r > 'z' {
				return ""
			}
		}
		return lang
	default:
		return ""
	}
}

// describePomfyErr turns a Pomfy failure into a plain-language message,
// keeping the cause reachable through Unwrap.
func describePomfyErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, pomfy.ErrNoTMDBAPIKey):
		return &friendlyError{cause: err, msg: "Pomfy's title search could not reach TMDB — even on the site's public key pool. Set TMDB_API_KEY (free at themoviedb.org) for a reliable quota, or search other sources."}
	case errors.Is(err, pomfy.ErrNotOnCatalog):
		return &friendlyError{cause: err, msg: "This title is not on Pomfy yet. Try searching it on another source."}
	case errors.Is(err, pomfy.ErrNoSources):
		return &friendlyError{cause: err, msg: "Pomfy has no player listed for this title yet. Try again later, or search it on another source."}
	case errors.Is(err, pomfy.ErrNoEpisodeSource):
		return &friendlyError{cause: err, msg: "This episode is not on Pomfy yet. Try again later, or pick another one."}
	default:
		return &friendlyError{cause: err, msg: "Pomfy playback failed: " + err.Error()}
	}
}

// describePomfyErrLess treats a missing show payload specially: a series
// whose seasons cannot be listed is a Pomfy-side outage, not a user error.
func describePomfyErrLess(media *models.Anime, err error) error {
	if err == nil {
		return &friendlyError{msg: "Pomfy lists no seasons for " + media.Name + "."}
	}
	return describePomfyErr(err)
}

// atoint64 parses s as an id, 0 when unparsable.
func atoint64(s string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return n
}
