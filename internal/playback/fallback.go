package playback

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/alvarorichard/Goanime/v2/internal/api/providers"
	"github.com/alvarorichard/Goanime/v2/internal/api/providers/naming"
	apisource "github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/player"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// When an episode will not load on the source the user picked (AnimeFire has
// stretches of instability), the same episode of the same title is looked for
// on the other anime sources, and played from the first that has it. Only the
// video comes from there: the title the user chose stays the one they watch,
// so their history and the CONTINUE list keep pointing at it.
//
// A wrong episode is worse than none, so a candidate is taken only when its
// title is the same title (see sameTitle) and it lists the episode's number.

// fallbackOrder is the order other sources are tried in: the Portuguese ones
// first, as the search lists them.
var fallbackOrder = []apisource.SourceKind{
	apisource.AnimeFire,
	apisource.AnimesDigital,
	apisource.Goyabu,
	apisource.Anroll,
	apisource.AnimeDrive,
	apisource.HiAnime,
}

// fallbackSearchTimeout bounds the search and the episode listing on each
// source, so a dead source costs seconds rather than the user's patience.
const fallbackSearchTimeout = 20 * time.Second

// Seams for the lookups on another source, so a test can drive the cascade
// without the network.
var (
	searchOnSource = func(ctx context.Context, kind apisource.SourceKind, query string) ([]*models.Anime, error) {
		s, ok := apisource.Registered(kind)
		if !ok {
			return nil, fmt.Errorf("%s is not registered", kind)
		}
		searchable, ok := s.(apisource.Searchable)
		if !ok {
			return nil, fmt.Errorf("%s cannot be searched", kind)
		}
		return searchable.Search(ctx, query)
	}
	episodesOnSource = providers.FetchEpisodes
)

// fallbackKinds are the sources worth trying for a title that failed on
// current: registered, reached by the search (not turned off by
// GOANIME_DISABLED_SOURCES, not shipped off and left off), and not current.
func fallbackKinds(current string) []apisource.SourceKind {
	var kinds []apisource.SourceKind
	for _, k := range fallbackOrder {
		s, ok := apisource.Registered(k)
		if !ok || !apisource.IsSearchEnabled(s.Describe()) {
			continue
		}
		// Source labels are not always the bare kind ("Animefire.io").
		if strings.HasPrefix(strings.ToLower(current), strings.ToLower(string(k))) {
			continue
		}
		kinds = append(kinds, k)
	}
	return kinds
}

// worthTryingElsewhere reports whether err is a source failing, as opposed to
// the user leaving or the program stopping.
func worthTryingElsewhere(ctx context.Context, err error) bool {
	if err == nil || ctx.Err() != nil {
		return false
	}
	for _, leaving := range []error{
		player.ErrBackToEpisodeSelection,
		player.ErrBackRequested,
		player.ErrBackToAnimeSelection,
		player.ErrUserQuit,
		context.Canceled,
	} {
		if errors.Is(err, leaving) {
			return false
		}
	}
	return true
}

// videoFromAnotherSource looks for episode episodeNum of anime on the other
// sources. It returns the video URL and the source it came from, or an empty
// URL and the sources it tried.
func videoFromAnotherSource(ctx context.Context, anime *models.Anime, episodeNumberStr string, episodeNum int) (videoURL, from string, tried []string) {
	// Movies and series have their own cascade, and a different catalogue.
	if anime == nil || anime.MediaType == models.MediaTypeMovie || anime.MediaType == models.MediaTypeTV {
		return "", "", nil
	}
	wanted := titleKey(anime.Name)
	if wanted == "" {
		return "", "", nil
	}
	query := naming.CleanTitle(anime.Name)

	for _, kind := range fallbackKinds(anime.Source) {
		if ctx.Err() != nil {
			break
		}
		tried = append(tried, string(kind))
		util.Infof("Trying %s for %s...", kind, strings.TrimSpace(query))

		url, ok := episodeVideoOn(ctx, kind, query, wanted, anime.Audio, episodeNumberStr, episodeNum)
		if ok {
			return url, string(kind), tried
		}
	}
	return "", "", tried
}

// episodeVideoOn finds the title and the episode on one source and resolves
// the episode's video.
func episodeVideoOn(ctx context.Context, kind apisource.SourceKind, query, wanted, audio, episodeNumberStr string, episodeNum int) (string, bool) {
	lookup, cancel := context.WithTimeout(ctx, fallbackSearchTimeout)
	defer cancel()

	found, err := searchOnSource(lookup, kind, query)
	if err != nil {
		util.Debugf("fallback: search on %s failed: %v", kind, err)
		return "", false
	}
	candidate := pickSameTitle(found, wanted, audio)
	if candidate == nil {
		util.Debugf("fallback: %s has no title called %q", kind, wanted)
		return "", false
	}
	// Some sources leave the field empty; the dispatcher resolves by it.
	if candidate.Source == "" {
		candidate.Source = string(kind)
	}

	episodes, err := episodesOnSource(lookup, candidate)
	if err != nil {
		util.Debugf("fallback: listing episodes on %s failed: %v", kind, err)
		return "", false
	}
	episode := findEpisodeNumbered(episodes, episodeNumberStr, episodeNum)
	if episode == nil {
		util.Debugf("fallback: %s does not list episode %d of %q", kind, episodeNum, candidate.Name)
		return "", false
	}

	candidate.Episodes = []models.Episode{*episode}
	// Not the lookup context: resolving the video may wait on the user (a
	// quality picker), and has its own limits.
	videoURL, err := videoURLFor(ctx, episode, candidate)
	if err != nil || videoURL == "" {
		util.Debugf("fallback: episode %d of %q did not load on %s: %v", episodeNum, candidate.Name, kind, err)
		return "", false
	}
	return videoURL, true
}

var (
	bracketed  = regexp.MustCompile(`\[[^\]]*\]|\([^)]*\)`)
	audioWords = regexp.MustCompile(`\b(dublado|legendado|dub|sub|todos os episodios)\b`)
	digitRun   = regexp.MustCompile(`\d+`)
)

// titleKey folds a title into what two sources would agree on: language and
// source tags, bracketed notes and the dub/sub wording off, lowercase, letters
// and digits only. A season marker stays ("... 2" is another entry than "...").
func titleKey(name string) string {
	title := strings.ToLower(naming.CleanTitle(name))
	title = bracketed.ReplaceAllString(title, " ")
	title = audioWords.ReplaceAllString(title, " ")
	var sb strings.Builder
	space := false
	for _, r := range title {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			sb.WriteRune(r)
			space = false
		case unicode.IsSpace(r), r == '-', r == ':':
			if !space {
				sb.WriteRune(' ')
				space = true
			}
		}
	}
	return strings.TrimSpace(sb.String())
}

// sameTitle reports whether name is the title wanted (a titleKey).
func sameTitle(name, wanted string) bool {
	return wanted != "" && titleKey(name) == wanted
}

// pickSameTitle chooses among a source's results the one that is the title,
// preferring the same audio (dubbed/subtitled) as the one that failed. It
// returns nil when none is the same title.
func pickSameTitle(results []*models.Anime, wanted, audio string) *models.Anime {
	var first *models.Anime
	for _, r := range results {
		if r == nil || !sameTitle(r.Name, wanted) {
			continue
		}
		if audio != "" && strings.EqualFold(r.Audio, audio) {
			return r
		}
		if first == nil {
			first = r
		}
	}
	return first
}

// findEpisodeNumbered returns the episode numbered like the one that failed.
// Numbers are compared as numbers ("Episódio 06" is 6); an episode with no
// number in its label never matches, since guessing one would play the wrong
// episode.
func findEpisodeNumbered(episodes []models.Episode, numberStr string, num int) *models.Episode {
	want := num
	if want <= 0 {
		n, ok := numberIn(numberStr)
		if !ok {
			return nil
		}
		want = n
	}
	for i := range episodes {
		// The label first: Num is an index in some sources.
		n, ok := numberIn(episodes[i].Number)
		if !ok {
			n = episodes[i].Num
		}
		if n == want {
			return &episodes[i]
		}
	}
	return nil
}

// numberIn is the first number in an episode label.
func numberIn(label string) (int, bool) {
	n, err := strconv.Atoi(digitRun.FindString(label))
	return n, err == nil
}
