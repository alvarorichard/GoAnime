package handlers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/appflow"
	"github.com/alvarorichard/Goanime/v2/internal/favorites"
	"github.com/alvarorichard/Goanime/v2/internal/suggest"
	"github.com/alvarorichard/Goanime/v2/internal/tracking"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"golang.org/x/term"
)

// SearchHomeDecision is the SEARCH home screen's decision.
type SearchHomeDecision struct {
	Query        string // new search text
	Scope        string // source selector: "", "ptbr", "hianime" or "movietv"
	Title        string // title to search when reopening fails
	Entry        *tui.HomeContinueEntry
	FromStart    bool // reopen from zero
	EpisodesOnly bool // stop at the episode list
	Cancelled    bool
}

// maxHomeContinue caps the CONTINUE list; maxHomeRecent caps RECENT.
const (
	maxHomeContinue = 30
	maxHomeRecent   = 5
)

// HandleSearchHome opens the SEARCH home screen (tela 1): the query box,
// scope tabs, CONTINUE history and RECENT searches.
func HandleSearchHome() (SearchHomeDecision, error) {
	res, err := OpenSearchHome("")
	if err != nil {
		if errors.Is(err, tui.ErrSearchHomeCancelled) {
			return SearchHomeDecision{Cancelled: true}, nil
		}
		return SearchHomeDecision{}, err
	}
	switch res.Action {
	case tui.HomeResume:
		entry := res.Entry
		return SearchHomeDecision{Entry: &entry, Title: entry.AnimeTitle, Scope: res.Scope}, nil
	case tui.HomeFromStart:
		entry := res.Entry
		return SearchHomeDecision{Entry: &entry, Title: entry.AnimeTitle, Scope: res.Scope, FromStart: true}, nil
	case tui.HomeEpisodes:
		entry := res.Entry
		return SearchHomeDecision{Entry: &entry, Title: entry.AnimeTitle, Scope: res.Scope, EpisodesOnly: true}, nil
	default:
		return SearchHomeDecision{Query: res.Query, Scope: res.Scope}, nil
	}
}

// OpenSearchHome opens the SEARCH home screen with query kept in the box.
// The retry loop uses it to go back to search (tela 3's "edit search").
// Empty query starts blank. Cancel surfaces tui.ErrSearchHomeCancelled.
func OpenSearchHome(query string) (tui.SearchHomeResult, error) {
	if !stdinIsTerminal() {
		return tui.SearchHomeResult{}, fmt.Errorf("no interactive terminal available for search")
	}
	return openSearchHome(query, runSearchHome)
}

// Seams for tests: whether stdin is a terminal, and the screen runner.
var (
	stdinIsTerminal = func() bool { return term.IsTerminal(int(os.Stdin.Fd())) }
	runSearchHome   = tui.SearchHome
)

// openSearchHome gathers the home screen's lists and options, with the
// autocomplete attached, and hands them to run (tui.SearchHome in
// production, a capture in tests).
func openSearchHome(query string, run func(tui.SearchHomeOptions) (tui.SearchHomeResult, error)) (tui.SearchHomeResult, error) {

	continueAll, err := tracking.ContinueEntries()
	if err != nil {
		util.Debugf("Continue history unavailable: %v", err)
	}
	entries := groupContinueEntries(continueAll)
	if len(entries) > maxHomeContinue {
		entries = entries[:maxHomeContinue]
	}

	past, err := tracking.RecentSearches(maxHomeRecent)
	if err != nil {
		util.Debugf("Recent searches unavailable: %v", err)
	}
	recent := make([]tui.HomeRecentSearch, 0, len(past))
	for _, r := range past {
		if query := tui.SingleLine(r.Query); query != "" {
			recent = append(recent, tui.HomeRecentSearch{Query: query, Count: r.ResultCount})
		}
	}

	// Read first: reading a damaged file is what sets it aside and earns the
	// notice.
	favs := homeFavorites()
	favNotice := favoritesNotice()

	res, err := run(tui.SearchHomeOptions{
		Query:         query,
		ScopeIndex:    scopeIndexForSource(util.GlobalSource),
		SourcesLine:   sourcesLine(),
		Continue:      entries,
		Recent:        recent,
		DeleteEntry:   deleteContinueEntry,
		DeleteRecent:  forgetRecentSearch,
		ClearContinue: clearContinueHistory,
		ClearRecent:   clearRecentHistory,
		Suggest:       homeSuggest,
		Sources:       homeSources,
		SetSource:     setHomeSource,

		Favorites:      favs,
		Notice:         favNotice,
		DeleteFavorite: forgetFavorite,
		ClearFavorites: clearFavorites,

		Restore: homeRestoreFor(query),
	})
	if err == nil {
		rememberHomeSearch(res)
	}
	return res, err
}

// favoritesStore is the FAVORITES file; a variable so tests use a temporary one.
var favoritesStore = favorites.Default

// homeFavorites reads the FAVORITES rows, newest first.
func homeFavorites() []tui.HomeFavorite {
	list, err := favoritesStore().List()
	if err != nil {
		util.Debugf("Favorites unavailable: %v", err)
		return nil
	}
	rows := make([]tui.HomeFavorite, 0, len(list))
	for _, f := range list {
		if title := tui.SingleLine(f.Title); title != "" {
			rows = append(rows, tui.HomeFavorite{Key: f.Key, Title: title, Year: f.Year, Media: f.Media, Source: f.Source})
		}
	}
	return rows
}

// favoritesNotice warns when the favorites file was damaged and set aside, so
// the empty list is not a mystery.
func favoritesNotice() string {
	aside := favoritesStore().TakeDamaged()
	if aside == "" {
		return ""
	}
	return "Your favorites file was damaged, so the list starts empty. A copy was kept as " + filepath.Base(aside) + "."
}

// forgetFavorite takes one work off FAVORITES.
func forgetFavorite(key string) error { return favoritesStore().Remove(key) }

// clearFavorites empties FAVORITES.
func clearFavorites() error { return favoritesStore().Clear() }

// homeSuggest adapts the keyless suggestion engine (Kitsu for anime, IMDb for
// movies and series) to the search home's autocomplete.
func homeSuggest(ctx context.Context, query, scope string) ([]tui.HomeSuggestion, error) {
	return homeSuggestWith(ctx, query, scope, suggest.Default().Suggest)
}

// homeSuggestWith narrows the lookup to the scope's kinds and adapts the rows.
func homeSuggestWith(ctx context.Context, query, scope string, ask func(context.Context, string, []suggest.Kind) ([]suggest.Suggestion, error)) ([]tui.HomeSuggestion, error) {
	return adaptSuggestions(ask(ctx, query, suggest.KindsForScope(scope)))
}

// adaptSuggestions turns engine suggestions into autocomplete rows.
func adaptSuggestions(items []suggest.Suggestion, err error) ([]tui.HomeSuggestion, error) {
	if err != nil {
		return nil, err
	}
	rows := make([]tui.HomeSuggestion, 0, len(items))
	for _, s := range items {
		rows = append(rows, tui.HomeSuggestion{Title: s.Title, Detail: s.Detail()})
	}
	return rows, nil
}

func init() {
	appflow.OpenSearchHomeFn = OpenSearchHome
}

// scopeIndexForSource maps a --source flag to the home screen's scope tab.
func scopeIndexForSource(src string) int {
	switch strings.ToLower(strings.TrimSpace(src)) {
	case "hianime":
		return 2
	case "ptbr", "pt-br", "animefire", "goyabu", "animesdigital", "anroll":
		return 1
	case "movietv", "startflix", "topcine", "pobreflix":
		return 3
	default:
		return 0
	}
}

// sourcesLine reads "6 sources on · Goyabu off" at the right of the header.
func sourcesLine() string {
	return tui.SourcesSummary(homeSources())
}

// Seams for tests: where the choices are saved and how.
var (
	sourcePrefsPath   = util.SourcePrefsPath
	persistSourcePref = util.PersistSourcePref
)

// homeSources lists the registered sources for the ctrl+s screen.
func homeSources() []tui.SourceRow {
	return homeSourceRows(source.SourceStates())
}

// homeSourceRows adapts source states to the screen's rows.
func homeSourceRows(states []source.State) []tui.SourceRow {
	rows := make([]tui.SourceRow, 0, len(states))
	for _, st := range states {
		rows = append(rows, tui.SourceRow{
			ID:     string(st.Kind),
			Name:   string(st.Kind),
			Note:   st.Note,
			On:     st.On,
			Locked: st.Locked,
		})
	}
	return rows
}

// setHomeSource saves the user's choice for one source.
func setHomeSource(id string, on bool) error {
	path, err := sourcePrefsPath()
	if err != nil {
		return err
	}
	return persistSourcePref(path, id, on)
}

// groupContinueEntries collapses watch history to one row per show: episodes
// of the same title share a row, and the latest watched episode wins. Rows
// never played (no position, unfinished) are noise, not history. Old rows
// without a show title group by episode URL instead.
func groupContinueEntries(rows []tracking.Anime) []tui.HomeContinueEntry {
	byShow := make(map[string]int)
	var entries []tui.HomeContinueEntry
	for _, e := range rows {
		finished := e.Duration > 0 && e.PlaybackTime >= e.Duration*9/10
		if !finished && e.PlaybackTime <= 0 {
			continue
		}
		title := tui.SingleLine(e.AnimeTitle)
		if title == "" {
			title = tui.SingleLine(e.Title)
		}
		key := continueGroupKey(e)
		srcName := tui.SingleLine(e.Source)
		if srcName == "" {
			// Rows written before the show context existed carry no
			// source: recover it from the tracked URL's domain so old
			// movies still route to their catalog.
			srcName = sourceFromURL(e.AllanimeID)
		}
		entry := tui.HomeContinueEntry{
			Key:          e.AllanimeID,
			AnimeTitle:   title,
			EpisodeTitle: tui.SingleLine(e.Title),
			EpisodeNum:   e.EpisodeNumber,
			Position:     e.PlaybackTime,
			Duration:     e.Duration,
			Source:       srcName,
			Audio:        tui.SingleLine(e.Audio),
			MediaType:    tui.SingleLine(e.MediaType),
			UpdatedAt:    e.LastUpdated,
		}
		if pos, ok := byShow[key]; ok {
			if e.LastUpdated.After(entries[pos].UpdatedAt) {
				entries[pos] = entry
			}
			continue
		}
		byShow[key] = len(entries)
		entries = append(entries, entry)
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].UpdatedAt.After(entries[j].UpdatedAt)
	})
	return entries
}

// sourceFromURL recovers a source display name from a tracked URL's domain
// ("https://www.startflix.test/series/x:ep1" to "StartFlix") for rows
// written before the show context existed.
func sourceFromURL(rawURL string) string {
	host := strings.ToLower(strings.TrimSpace(rawURL))
	if i := strings.Index(host, "://"); i >= 0 {
		host = host[i+3:]
	}
	if i := strings.Index(host, "/"); i >= 0 {
		host = host[:i]
	}
	host = strings.TrimPrefix(host, "www.")
	switch {
	case strings.Contains(host, "animefire"):
		return "Animefire.io"
	case strings.Contains(host, "goyabu"):
		return "Goyabu"
	case strings.Contains(host, "hianime"):
		return "HiAnime"
	case strings.Contains(host, "animesdigital"):
		return "AnimesDigital"
	case strings.Contains(host, "anroll"):
		return "AnimesROLL"
	case strings.Contains(host, "startflix"):
		return "StartFlix"
	case strings.Contains(host, "topcine"):
		return "TopCine"
	case strings.Contains(host, "pobreflixtv"):
		return "Pobreflix"
	case strings.Contains(host, "animedrive") || strings.Contains(host, "animesdrive"):
		return "AnimeDrive"
	default:
		return ""
	}
}

// title is known, the episode URL with its episode suffix stripped
// ("...:ep5") otherwise.
func continueGroupKey(e tracking.Anime) string {
	if title := strings.ToLower(strings.TrimSpace(e.AnimeTitle)); title != "" {
		return "t:" + title + "\x00" + strings.ToLower(strings.TrimSpace(e.Source))
	}
	base := e.AllanimeID
	if i := strings.LastIndex(base, ":ep"); i >= 0 {
		if _, err := strconv.Atoi(base[i+3:]); err == nil {
			base = base[:i]
		}
	}
	return "u:" + strings.ToLower(strings.TrimSpace(base))
}

// continueTarget decides what reopening a row plays: a finished episode
// offers the next one from zero (resuming at the credits is never right); a
// finished movie replays itself from the start; anything else resumes where
// it stopped. FromStart always starts over.
func continueTarget(entry tui.HomeContinueEntry, fromStart bool) (episodeNum, position int, resume, start bool) {
	if fromStart {
		return entry.EpisodeNum, 0, false, true
	}
	finished := entry.Duration > 0 && entry.Position >= entry.Duration*9/10
	if !finished {
		return entry.EpisodeNum, entry.Position, true, false
	}
	if strings.EqualFold(strings.TrimSpace(entry.MediaType), "movie") {
		return entry.EpisodeNum, 0, false, true
	}
	return entry.EpisodeNum + 1, 0, false, false
}

// Seams for forgetting history, so a test can say what the history answers and
// read back what was asked of it.
var (
	continueRows    = tracking.ContinueEntries
	deleteTracked   = deleteTrackedRow
	forgetSearchFn  = tracking.ForgetSearch
	clearContinueFn = tracking.ClearContinue
	clearRecentFn   = tracking.ClearRecent
)

// deleteTrackedRow removes one episode's row from watch history.
func deleteTrackedRow(key string) error {
	tracker := tracking.GetGlobalTracker()
	if tracker == nil {
		tracker = tracking.NewLocalTracker(tracking.DBPath())
	}
	if tracker == nil {
		return fmt.Errorf("watch history unavailable")
	}
	return tracker.DeleteAnime(0, key)
}

// keysOfShow is the history key of every episode of the show the row with key
// belongs to. The CONTINUE list shows one row per show, standing for the
// latest episode watched, so removing the row has to remove them all or the
// show would come back with an older episode. A key not in rows is just itself.
func keysOfShow(rows []tracking.Anime, key string) []string {
	var group string
	for _, r := range rows {
		if r.AllanimeID == key {
			group = continueGroupKey(r)
			break
		}
	}
	if group == "" {
		return []string{key}
	}
	var keys []string
	for _, r := range rows {
		if continueGroupKey(r) == group {
			keys = append(keys, r.AllanimeID)
		}
	}
	return keys
}

// deleteContinueEntry removes one CONTINUE row, and so the show it stands
// for, from watch history.
func deleteContinueEntry(key string) error {
	rows, err := continueRows()
	if err != nil {
		return err
	}
	for _, k := range keysOfShow(rows, key) {
		if err := deleteTracked(k); err != nil {
			return err
		}
	}
	return nil
}

// forgetRecentSearch removes one search from the RECENT list.
func forgetRecentSearch(query string) error {
	return forgetSearchFn(query)
}

// clearContinueHistory empties the CONTINUE list.
func clearContinueHistory() error {
	_, err := clearContinueFn()
	return err
}

// clearRecentHistory empties the RECENT list.
func clearRecentHistory() error {
	_, err := clearRecentFn()
	return err
}
