package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/appflow"
	"github.com/alvarorichard/Goanime/internal/discord"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/playback"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
)

// HandleContinue reopens a CONTINUE row: enter resumes the tracked episode
// directly, r starts it over, e stops at the episode list. Anything the
// direct path cannot open falls back to a normal search of the title.
func HandleContinue(decision SearchHomeDecision) {
	if decision.Entry == nil {
		return
	}
	if err := openContinue(decision); err != nil {
		util.Warnf("Could not reopen directly (%v); searching instead.", err)
		clearContinueOverrides()
		playbackModeFn()(fallbackTitle(decision))
	}
}

// fallbackTitle is the title a failed reopen searches for: the decision's,
// else the row's show title.
func fallbackTitle(decision SearchHomeDecision) string {
	if decision.Title != "" {
		return decision.Title
	}
	if decision.Entry != nil {
		return decision.Entry.AnimeTitle
	}
	return ""
}

// clearContinueOverrides drops a direct open's one-shot state so a fallback
// search behaves like a fresh one.
func clearContinueOverrides() {
	playback.PreselectEpisodeNum = 0
	player.ClearResumeOverride()
}

// Flow seams: production wires the real search, fetch and playback. Tests
// swap them for scripted fakes to drive openContinue headless end to end.
var (
	searchAnimeAutoFn   = api.SearchAnimeAuto
	fetchAnimeDetailsFn = appflow.FetchAnimeDetails
	getAnimeEpisodesFn  = appflow.GetAnimeEpisodes
	handleSeriesFn      = playback.HandleSeries
	handleMovieFn       = playback.HandleMovie
	// handlePlaybackModeFn routes fallbacks back to search. Nil means
	// HandlePlaybackMode itself; tests inject a recorder. It must stay
	// uninitialized at declaration: pointing it at HandlePlaybackMode
	// here is an initialization cycle (it calls back into HandleContinue).
	handlePlaybackModeFn func(string)
)

// playbackModeFn resolves the fallback route.
func playbackModeFn() func(string) {
	if handlePlaybackModeFn != nil {
		return handlePlaybackModeFn
	}
	return HandlePlaybackMode
}

// openContinue searches the tracked title, auto-picks its source and plays
// the tracked episode.
func openContinue(decision SearchHomeDecision) error {
	entry := decision.Entry
	title := cleanContinueTitle(entry)
	if title == "" {
		return fmt.Errorf("no title to search for")
	}

	src := continueSourceToken(entry.Source)
	if src == "" && isMovieEntry(entry) {
		// A sourceless movie restricts to the movie catalogs: searching
		// everywhere lets an anime of the same name win the pick.
		src = "movietv"
	}
	anime, err := searchAnimeAutoFn(title, src, pickContinueVersion(entry.Source, entry.MediaType))
	if err != nil {
		return err
	}
	if isMovieEntry(entry) && !anime.IsMovie() && anime.MediaType != "" {
		// Never play the wrong content: a tracked movie that resolves to
		// an anime falls back to the results screen instead.
		util.Debugf("Continue mismatch: tracked movie %q resolved to %q (%s)", title, anime.Name, anime.MediaType)
		return fmt.Errorf("tracked movie %q resolved to an anime title", title)
	}

	fetchAnimeDetailsFn(anime)
	episodes, err := getAnimeEpisodesFn(anime)
	if err != nil || len(episodes) == 0 {
		if err == nil {
			err = fmt.Errorf("no episodes")
		}
		reportEpisodesUnavailable(anime, err)
		return err
	}

	if !decision.EpisodesOnly {
		series := !anime.IsMovie() && len(episodes) > 1
		epNum, pos, resume, start := continueTarget(*entry, decision.FromStart)
		if series {
			playback.PreselectEpisodeNum = epNum
		} else {
			// Movies and lists never preselect: a stale number would
			// leak into the next series flow (HandleMovie reads none).
			playback.PreselectEpisodeNum = 0
		}
		if start {
			player.ForceStartNext()
		} else if resume {
			player.ForceResumeNext(pos)
		}
		util.Infof("Reopening %q (%s, episode %d)", title, displayContinueSource(anime), epNum)
	} else {
		util.Infof("Opening episodes of %q (%s)", title, displayContinueSource(anime))
	}

	ctx := context.Background()
	discordManager := discord.NewManager()
	_ = discordManager.Initialize()
	defer discordManager.Shutdown()

	totalEpisodes := len(episodes)
	var playbackErr error
	if !anime.IsMovie() && totalEpisodes > 1 {
		playbackErr = handleSeriesFn(ctx, anime, episodes, totalEpisodes, discordManager.IsEnabled())
	} else {
		playbackErr = handleMovieFn(ctx, anime, episodes, discordManager.IsEnabled())
	}
	if errors.Is(playbackErr, player.ErrBackToAnimeSelection) {
		clearContinueOverrides()
		playbackModeFn()(title)
		return nil
	}
	return playbackErr
}

// continueSourceToken maps a tracked source display name to the search's
// source selector ("Animefire.io" to "animefire"). Unknown sources search
// everywhere.
func continueSourceToken(source string) string {
	token := strings.ToLower(strings.TrimSpace(source))
	token = strings.TrimSuffix(token, ".io")
	switch token {
	case "animefire", "goyabu", "hianime", "animesdigital", "startflix", "topcine", "pobreflix", "animedrive":
		return token
	default:
		return ""
	}
}

// isMovieEntry reports whether a CONTINUE row tracks a movie.
func isMovieEntry(entry *tui.HomeContinueEntry) bool {
	return strings.EqualFold(strings.TrimSpace(entry.MediaType), "movie")
}

// cleanContinueTitle strips scraper tags ("[PT-BR]", "[Movie]") and audio
// words ("(Legendado)", "(Dublado)") so a tracked title searches clean.
// Tracked titles come from scraped sites and may carry controls or ANSI:
// SingleLine sanitizes first so nothing hostile reaches the search.
func cleanContinueTitle(entry *tui.HomeContinueEntry) string {
	title := tui.SingleLine(strings.TrimSpace(entry.AnimeTitle))
	if title == "" {
		title = tui.SingleLine(strings.TrimSpace(entry.EpisodeTitle))
	}
	for {
		trimmed := strings.TrimSpace(title)
		if !strings.HasPrefix(trimmed, "[") {
			title = trimmed
			break
		}
		end := strings.Index(trimmed, "]")
		if end < 0 {
			break
		}
		title = strings.TrimSpace(trimmed[end+1:])
	}
	lower := strings.ToLower(title)
	for _, audio := range []string{"(legendado)", "(dublado)"} {
		if strings.HasSuffix(lower, audio) {
			title = strings.TrimSpace(title[:len(title)-len(audio)])
			lower = strings.ToLower(title)
		}
	}
	return title
}

// displayContinueSource is how the picked source reads in the reopen line.
func displayContinueSource(anime *models.Anime) string {
	if source := strings.TrimSpace(anime.Source); source != "" {
		return source
	}
	return "unknown source"
}

// pickContinueVersion prefers the tracked title's copy on its source, then
// the same media type (a tracked movie must not reopen as the anime of the
// same name); anything else keeps the search's closest match first.
func pickContinueVersion(source, mediaType string) func([]*models.Anime) *models.Anime {
	return func(animes []*models.Anime) *models.Anime {
		if len(animes) == 0 {
			return nil
		}
		want := strings.ToLower(strings.TrimSpace(source))
		want = strings.TrimSuffix(want, ".io")
		for _, anime := range animes {
			if anime == nil {
				continue
			}
			have := strings.ToLower(strings.TrimSpace(anime.Source))
			have = strings.TrimSuffix(have, ".io")
			if want != "" && have == want {
				return anime
			}
		}
		if wantType := strings.ToLower(strings.TrimSpace(mediaType)); wantType != "" {
			for _, anime := range animes {
				if anime == nil {
					continue
				}
				if have := strings.ToLower(strings.TrimSpace(string(anime.MediaType))); have != "" && have == wantType {
					return anime
				}
			}
		}
		return animes[0]
	}
}
