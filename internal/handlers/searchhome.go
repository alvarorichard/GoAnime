package handlers

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/appflow"
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
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return tui.SearchHomeResult{}, fmt.Errorf("no interactive terminal available for search")
	}

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

	return tui.SearchHome(tui.SearchHomeOptions{
		Query:       query,
		ScopeIndex:  scopeIndexForSource(util.GlobalSource),
		SourcesLine: sourcesLine(),
		Continue:    entries,
		Recent:      recent,
		DeleteEntry: deleteContinueEntry,
	})
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
	active := source.ActiveSources()
	disabled := source.DisabledSources()
	sep := " · "
	if tui.Plain() {
		sep = " - "
	}
	line := fmt.Sprintf("%d sources on", len(active))
	if len(disabled) > 0 {
		names := make([]string, 0, len(disabled))
		for _, kind := range disabled {
			names = append(names, string(kind))
		}
		line += sep + strings.Join(names, ", ") + " off"
	}
	return line
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

// deleteContinueEntry removes one CONTINUE row from watch history.
func deleteContinueEntry(key string) error {
	tracker := tracking.GetGlobalTracker()
	if tracker == nil {
		tracker = tracking.NewLocalTracker(tracking.DBPath())
	}
	if tracker == nil {
		return fmt.Errorf("watch history unavailable")
	}
	return tracker.DeleteAnime(0, key)
}
