package pobreflix

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// The catalog's search corpus. Production asks the site's /api catalog first;
// this index is the browserless fallback when that API is unavailable or the
// mirror still runs the old theme. The old theme had no guest search guests
// could call (/buscar is client-side, the IPS quick-search endpoint is
// members-only, and its full search page answers every query with the same
// slider noise — checked 2026-10-16), but it publishes the whole grid as
// content sitemaps: sitemap.php lists sitemap_content_videobox_Video_<n> shard
// files, 1000 URLs each, with every URL carrying its audio. Current mirrors
// publish sitemap-movies.xml and sitemap-series.xml instead. A search builds
// an index of those files once and ranks the slugs against the query.

const (
	// corpusTTL is how long an in-date cache saves the shard fetches. The
	// catalog adds titles every day, but a two-day-old index still answers
	// most queries — a miss is visible ("nothing found"), never a wrong hit.
	corpusTTL = 48 * time.Hour

	// corpusConcurrency bounds the shard fetches; each is ~180 KB and the
	// per-source search deadline is 12s, so the whole index must land inside
	// a dozen parallel requests' time.
	corpusConcurrency = 12

	// maxSearchResults is how many corpus matches one search returns. The
	// panel-ordered list is capped at the winners' end — a 74k-record catalog
	// matched on one or two tokens can rank hundreds; the user reads 25.
	maxSearchResults = 25

	// refinementBudget is how many pre-filtered candidates go through
	// models.RelevanceScore, the same engine that orders every other source's
	// results. Pre-filtering keeps the per-search cost bounded on a
	// 74k-record corpus; the cheap pass drops nothing a word match would
	// have kept.
	refinementBudget = 500
)

// videoRecordRe pulls an old record URL apart:
//
//	https://www.pobreflixtv.sale/filmes/online/verity-dublado-84506/
//	https://www.pobreflixtv.sale/series/online/ben-hur-a-serie-dublado-83633/
var videoRecordRe = regexp.MustCompile(
	`^/(filmes|series|episodios)/online/([^/]+)`)

// Current-site paths: /filme/<slug>.html, /serie/<slug>.html and
// /assistir/<slug>-t<season>-ep<episode>.html. The sitemaps list the same
// pages without ".html" (2026-10-26), so the suffix is optional here. The slug is the record's
// identifier; it often ends in a random suffix ("todo-poderoso-53qvb") or the
// TMDB id ("ben-hur-40360").
var (
	newMoviePathRe   = regexp.MustCompile(`^/filme/([^/]+?)(?:\.html)?$`)
	newSeriesPathRe  = regexp.MustCompile(`^/serie/([^/]+?)(?:\.html)?$`)
	newEpisodePathRe = regexp.MustCompile(`^/assistir/(.+)-t\d+-ep\d+(?:\.html)?$`)
	currentRecordRe  = regexp.MustCompile(`^/(filme|serie)/([^/]+?)(?:\.html)?$`)
)

// trailingIDRe splits the record id off the end of a slug.
var trailingIDRe = regexp.MustCompile(`^(.*)-(\d+)$`)

// audioOfSlug reads the audio a slug names, without changing the slug.
func audioOfSlug(slug string) Audio {
	am := audioSuffixRe.FindStringSubmatch(slug)
	if am == nil {
		return AudioUnknown
	}
	if am[2] == "legendado" {
		return AudioSubtitled
	}
	return AudioDubbed
}

// isRandomSlugSuffix reports whether the last slug token looks like the
// current site's random disambiguator ("4hsp3", "0vl7bwm", "g6JBa"). The
// tokens are 5-8 alphanumerics with at least one digit and one letter. A
// numeric token of any length is the TMDB id the site sometimes appends.
func isRandomSlugSuffix(tok string) bool {
	if tok == "" {
		return false
	}
	if n, err := strconv.Atoi(tok); err == nil {
		// Four-digit years are part of many titles ("Blade Runner 2049");
		// TMDB ids appended as disambiguators are usually longer or clearly
		// outside the year range.
		if n >= 1900 && n <= 2099 {
			return false
		}
		return n > 0
	}
	if len(tok) < 5 || len(tok) > 8 {
		return false
	}
	digit, letter := false, false
	for _, r := range tok {
		switch {
		case r >= '0' && r <= '9':
			digit = true
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
			letter = true
		default:
			return false
		}
	}
	return digit && letter
}

// displaySlug removes the record id / random suffix the current site appends
// to a slug, keeping at least one token. It is for display and matching only;
// record URLs keep the original slug.
func displaySlug(slug string) string {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return ""
	}
	parts := strings.Split(slug, "-")
	if len(parts) < 2 {
		return slug
	}
	if isRandomSlugSuffix(parts[len(parts)-1]) {
		trimmed := strings.Join(parts[:len(parts)-1], "-")
		if strings.Trim(trimmed, "-") != "" {
			return trimmed
		}
	}
	return slug
}

// audioSuffixRe reads the audio a record slug names: the last dash words when
// they name a known track.
var audioSuffixRe = regexp.MustCompile(`^(.*?)-(dublado|legendado|dual|dual-dublado)$`)

// locRe extracts every <loc> URL of a sitemap document. The content type is
// text/html here, so no XML parser is asked to be strict about it.
var locRe = regexp.MustCompile(`(?i)<loc>\s*(.*?)\s*</loc>`)

// entry is one indexed record URL.
type entry struct {
	URL string
	// Kind classifies the record; /episodios/ URLs never enter the corpus.
	Kind MediaKind
	// BaseSlug is the slug with the audio suffix and the record id removed —
	// "verity" out of "verity-dublado-84506". It is the dedup key.
	BaseSlug string
	// Tokens is the folded, spaced form of BaseSlug ("o apanhador de
	// sonhos"), computed once when the entry enters the index.
	Tokens string
	Audio  Audio
	// VideoID is the record id the URL ends in; used for recency ordering.
	VideoID int
}

// parseRecordPath reads every title-page layout the app routes. An episode
// URL reports ok=false for the corpus (that is per-episode noise), while title
// pages keep their raw slug for the episode marker check.
func parseRecordPath(path string) (kind MediaKind, baseSlug string, audio Audio, videoID int, ok bool) {
	if m := currentRecordRe.FindStringSubmatch(path); m != nil {
		kind = KindMovie
		if m[1] == "serie" {
			kind = KindSeries
		}
		slug := strings.TrimSpace(m[2])
		audio = audioOfSlug(slug)
		if am := audioSuffixRe.FindStringSubmatch(slug); am != nil {
			slug = am[1]
		}
		// The trailing id/random suffix is not part of the title. Keep the
		// original URL; only the display slug is trimmed.
		if tm := trailingIDRe.FindStringSubmatch(slug); tm != nil {
			if n, err := strconv.Atoi(tm[2]); err == nil && n > 0 && (n < 1900 || n > 2099) {
				slug = tm[1]
			}
		}
		slug = displaySlug(slug)
		if slug == "" {
			return "", "", AudioUnknown, 0, false
		}
		return kind, slug, audio, 0, true
	}
	if newEpisodePathRe.MatchString(path) {
		return "", "", AudioUnknown, 0, false
	}
	m := videoRecordRe.FindStringSubmatch(path)
	if m == nil {
		return "", "", AudioUnknown, 0, false
	}
	if m[1] == "episodios" {
		return "", "", AudioUnknown, 0, false
	}
	kind = KindMovie
	if m[1] == "series" {
		kind = KindSeries
	}
	slug := strings.TrimSuffix(m[2], "/")

	// The record id ends the slug.
	idm := trailingIDRe.FindStringSubmatch(slug)
	if idm == nil {
		return "", "", AudioUnknown, 0, false
	}
	vid, err := strconv.Atoi(idm[2])
	if err != nil || vid <= 0 {
		return "", "", AudioUnknown, 0, false
	}
	slug = idm[1]

	// Then the audio suffix.
	audio = AudioUnknown
	if am := audioSuffixRe.FindStringSubmatch(slug); am != nil {
		slug = am[1]
		switch am[2] {
		case "legendado":
			audio = AudioSubtitled
		default:
			audio = AudioDubbed // dublado and dual tracks play dubbed first
		}
	}
	if slug == "" {
		return "", "", AudioUnknown, 0, false
	}
	return kind, slug, audio, vid, true
}

// slugTokens folds a slug into compare tokens: lowercase, spaces instead of
// separators, single-spaced. Slugs are already accent-free — the site writes
// them that way — so no diacritic work is needed here.
var slugNoiseRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugTokens(slug string) string {
	return strings.TrimSpace(slugNoiseRe.ReplaceAllString(strings.ToLower(slug), " "))
}

// queryFold lowers and accent-folds a user query so it can meet the slugs. A
// package-level Replacer, because the per-call build inside models' fold is
// fine for dozens of titles and wrong for a 74k-record corpus.
var queryFoldRe = strings.NewReplacer(
	"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ç", "c", "ñ", "n",
	"Á", "a", "À", "a", "Â", "a", "Ã", "a", "Ä", "a",
	"É", "e", "È", "e", "Ê", "e", "Ë", "e",
	"Í", "i", "Ì", "i", "Î", "i", "Ï", "i",
	"Ó", "o", "Ò", "o", "Ô", "o", "Õ", "o", "Ö", "o",
	"Ú", "u", "Ù", "u", "Û", "u", "Ü", "u",
	"Ç", "c", "Ñ", "n",
)

// corpus is the in-memory search index.
type corpus struct {
	mu      sync.Mutex
	entries []entry
	// fetchedAt is when the index was last read from the network. A stale
	// date makes the next search resume missing shards.
	fetchedAt time.Time
	// shardsDone lists the shard files folded into the index.
	shardsDone map[string]bool
	// cachePath is the disk file; NewClient fills the user-cache default,
	// tests point it at a throwaway file.
	cachePath string
}

// corpusCacheFile is the disk format. Entries are raw URLs so a parser change
// re-derives everything without a migration.
type corpusCacheFile struct {
	FetchedAt int64           `json:"fetched_at"`
	Shards    map[string]bool `json:"shards_done"`
	Entries   []string        `json:"entries"`
}

// corpusCachePath is where the index is cached by default: next to the
// app's other caches.
func corpusCachePath() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "GoAnime")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	return filepath.Join(dir, "pobreflix-index.json"), nil
}

// loadCache reads the disk cache, if it parses and matches the host this
// client talks to. A missing or broken file is a cold cache, not an error.
func (cc *corpus) loadCache(baseURL string) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return
	}
	path := cc.cachePath
	if path == "" {
		path, err = corpusCachePath()
		if err != nil {
			return
		}
	}
	// Scope access to the cache directory so a substituted symlink cannot
	// turn the cache into a read of a file outside that directory.
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return
	}
	defer func() { _ = root.Close() }()
	raw, err := root.ReadFile(filepath.Base(path))
	if err != nil {
		return
	}
	var file corpusCacheFile
	if err := json.Unmarshal(raw, &file); err != nil || len(file.Entries) == 0 {
		util.Debug("Pobreflix corpus cache unreadable", "path", path)
		return
	}
	entries := make([]entry, 0, len(file.Entries))
	for _, rawURL := range file.Entries {
		e, ok := entryFromURL(rawURL)
		if !ok || !sameHost(uHost(e.URL), base.Host) {
			// A cache from a mirror (or a hand-edited file) whose records do
			// not answer at this host is dropped wholesale.
			continue
		}
		entries = append(entries, e)
	}
	if len(entries) == 0 {
		return
	}
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.entries = entries
	cc.fetchedAt = time.Unix(file.FetchedAt, 0)
	if file.Shards != nil {
		cc.shardsDone = file.Shards
	}
}

// sameHost compares hosts ignoring the www. prefix.
func sameHost(a, b string) bool {
	return strings.TrimPrefix(strings.ToLower(a), "www.") == strings.TrimPrefix(strings.ToLower(b), "www.")
}

// uHost is the host of a URL, "" when it does not parse — entries in the
// cache all came from here, but a hand-edited cache file should not panic.
func uHost(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Host
}

// saveCache persists the index. Best-effort: a read-only cache dir degrades a
// later search to a slower build, never a failure.
func (cc *corpus) saveCache() {
	path := cc.cachePath
	if path == "" {
		var err error
		if path, err = corpusCachePath(); err != nil {
			return
		}
	}
	cc.mu.Lock()
	if len(cc.entries) == 0 {
		cc.mu.Unlock()
		return
	}
	urls := make([]string, 0, len(cc.entries))
	for _, e := range cc.entries {
		urls = append(urls, e.URL)
	}
	shards := cc.shardsDone
	if shards == nil {
		shards = map[string]bool{}
	}
	fetched := cc.fetchedAt.Unix()
	cc.mu.Unlock()

	raw, err := json.Marshal(corpusCacheFile{FetchedAt: fetched, Shards: shards, Entries: urls})
	if err != nil {
		return
	}
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return
	}
	defer func() { _ = root.Close() }()
	cacheFile, err := root.OpenFile(filepath.Base(path), os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		util.Debug("Pobreflix corpus cache save failed", "path", path, "err", err)
		return
	}
	// Tighten legacy caches too, using the opened file so a path replacement
	// cannot redirect chmod. Apply permissions before writing the new index.
	if err = cacheFile.Chmod(0o600); err == nil {
		err = cacheFile.Truncate(0)
	}
	if err == nil {
		_, err = cacheFile.Write(raw)
	}
	if closeErr := cacheFile.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		util.Debug("Pobreflix corpus cache save failed", "path", path, "err", err)
	}
}

// fresh reports a full index fetched within the TTL.
//
// "Full" is load-bearing: entries with no completed shard recorded mean the
// build timed out partway — a partial round that never resumes while it
// counts as fresh. A shard holds thousands of records, so a missing one is
// not a niche gap; "todo poderoso" (record 450) was unreachable for two days
// because the first build died before its shard. An incomplete index keeps
// answering through refresh's rules and finishes itself on the next search.
func (cc *corpus) fresh() bool {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	return len(cc.entries) > 0 && len(cc.shardsDone) > 0 && time.Since(cc.fetchedAt) < corpusTTL
}

// entryFromURL parses a record URL into an index entry.
func entryFromURL(raw string) (entry, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return entry{}, false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return entry{}, false
	}
	kind, slug, audio, vid, ok := parseRecordPath(u.Path)
	if !ok {
		return entry{}, false
	}
	// Every URL the app opens is the ".html" form the catalog API returns.
	if currentRecordRe.MatchString(u.Path) && !strings.HasSuffix(u.Path, ".html") {
		u.Path += ".html"
		u.RawQuery, u.Fragment = "", ""
		raw = u.String()
	}
	return entry{
		URL:      raw,
		Kind:     kind,
		BaseSlug: slug,
		Tokens:   slugTokens(slug),
		Audio:    audio,
		VideoID:  vid,
	}, true
}

// refresh builds the index: reads the sitemap index, then fetches every shard
// the cache does not already cover. It fails only when nothing was learned —
// a partial build is kept and searched, so a timed-out first build still
// returns results and finishes itself on the next search.
func (cc *corpus) refresh(ctx context.Context, client *Client) error {
	shards, err := client.sitemapShards(ctx)
	if err != nil {
		// Without a sitemap index a stale cache is still worth reusing
		// (it covers the catalog minus new uploads).
		cc.mu.Lock()
		hasEntries := len(cc.entries) > 0
		cc.mu.Unlock()
		if hasEntries {
			util.Debug("Pobreflix sitemap index failed, reusing the cache", "err", err)
			return nil
		}
		return err
	}
	if len(shards) == 0 {
		return errors.New("pobreflix: sitemap lists no video shards")
	}

	// Shards to fetch: everything not recorded done. A first build starts
	// empty; a partial-build resume fetches only the misses.
	var todo []string
	cc.mu.Lock()
	if cc.shardsDone == nil {
		cc.shardsDone = map[string]bool{}
	}
	for _, sh := range shards {
		if !cc.shardsDone[sh.name] {
			todo = append(todo, sh.name)
		}
	}
	fullBuild := len(todo) == len(shards)
	cc.mu.Unlock()
	if len(todo) == 0 {
		// Everything is done but the date is stale: a full rescan replaces
		// the entries with this round's catalog.
		fullBuild = true
		for _, sh := range shards {
			todo = append(todo, sh.name)
		}
	}

	if err := cc.fetchShards(ctx, client, shards, todo, fullBuild); err != nil {
		return err
	}
	cc.mu.Lock()
	learned := len(cc.entries) > 0 && len(cc.shardsDone) > 0
	cc.mu.Unlock()
	if !learned {
		return errors.New("pobreflix: sitemap shards produced no catalog entries")
	}
	return nil
}

// shard is one content sitemap file. Old layouts name
// sitemap_content_videobox_Video_<n>; current ones name sitemap-movies.xml and
// sitemap-series.xml (with an explicit order).
type shard struct {
	name  string // sitemap_content_videobox_Video_12 or movies
	url   string
	order int // explicit order for current sitemaps, 0 for old names
}

// shardNumber orders shards deterministically by their numeric suffix.
func (s shard) number() int {
	if s.order != 0 {
		return s.order
	}
	i := strings.LastIndex(s.name, "_")
	if i < 0 {
		return 0
	}
	n, err := strconv.Atoi(s.name[i+1:])
	if err != nil {
		return 0
	}
	return n
}

// fetchShards pulls the todo shards in bounded parallel batches and folds the
// result in. A shard that fails or is cancelled is skipped — its records are
// simply absent from this round, and the next search resumes them.
func (cc *corpus) fetchShards(ctx context.Context, client *Client, shards []shard, todo []string, fullBuild bool) error {
	byName := make(map[string]shard, len(shards))
	for _, sh := range shards {
		byName[sh.name] = sh
	}
	sort.Slice(todo, func(i, j int) bool {
		return byName[todo[i]].number() < byName[todo[j]].number()
	})
	sem := make(chan struct{}, corpusConcurrency)
	var (
		wg    sync.WaitGroup
		added syncMap // shard name -> fresh entries
	)
	for _, name := range todo {
		select {
		case <-ctx.Done():
			wg.Wait()
			cc.fold(&added, fullBuild, false, len(shards))
			return ctx.Err()
		case sem <- struct{}{}:
		}
		wg.Add(1)
		go func(name string) {
			defer func() { <-sem; wg.Done() }()
			body, err := client.get(ctx, byName[name].url, getOptions{
				accept: "text/xml,application/xml,text/html,*/*;q=0.8",
			})
			if err != nil {
				// A shard loss is not fatal: the round keeps whatever other
				// shards answered, and the next search resumes the misses.
				util.Debug("Pobreflix shard fetch failed", "shard", name, "err", err)
				return
			}
			added.store(name, parseShardLocs(body))
		}(name)
	}
	wg.Wait()
	cc.fold(&added, fullBuild, true, len(shards))
	util.Debug("Pobreflix corpus updated", "shards", len(todo), "of", len(shards))
	return nil
}

// syncMap collects per-shard results from the fetch goroutines.
type syncMap struct {
	mu sync.Mutex
	m  map[string][]string
}

func (s *syncMap) store(k string, v []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.m == nil {
		s.m = map[string][]string{}
	}
	s.m[k] = v
}

// fold merges fetched shards into the index. fullBuild replaces the entries
// with this round's scan (a complete rescan is the catalog as of now);
// partialBuild appends the fresh shards to what is held. done marks a round
// that finished its goroutines; the fetchedAt stamp moves only when every
// listed shard is covered — a partial round stays stale so the next search
// resumes the missing shards.
func (cc *corpus) fold(added *syncMap, fullBuild, done bool, totalShards int) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	if cc.shardsDone == nil {
		cc.shardsDone = map[string]bool{}
	}
	if fullBuild {
		cc.entries = nil
	}
	covered := 0
	for name, urls := range added.m {
		for _, raw := range urls {
			if e, ok := entryFromURL(raw); ok {
				cc.entries = append(cc.entries, e)
			}
		}
		cc.shardsDone[name] = true
		covered++
	}
	// Only a round that left the index complete counts as fresh; anything
	// else keeps the date stale and resumes on the next search.
	if done && totalShards > 0 && len(cc.shardsDone) >= totalShards && len(cc.entries) > 0 {
		cc.fetchedAt = time.Now()
	}
}

// parseShardLocs extracts the record URLs one shard carries.
func parseShardLocs(body []byte) []string {
	raws := locRe.FindAllStringSubmatch(string(body), -1)
	out := make([]string, 0, len(raws))
	for _, m := range raws {
		out = append(out, m[1])
	}
	return out
}

// hasEntries reports whether any index entries exist.
func (cc *corpus) hasEntries() bool {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	return len(cc.entries) > 0
}

// rankStopWords are the glue words models' scorer ignores. A query made only
// of them has no catalog match to find; the corpus stays silent instead of
// ranking containment noise. Mirrors models.queryTokens' list.
var rankStopWords = map[string]bool{
	"a": true, "o": true, "os": true, "as": true,
	"de": true, "do": true, "da": true, "dos": true, "das": true,
	"e": true, "em": true, "no": true, "na": true, "nos": true, "nas": true,
	"um": true, "uma": true, "uns": true, "umas": true,
	"para": true, "com": true, "que": true, "the": true, "of": true,
}

// rank scores the corpus against the query: a cheap token pre-filter over
// the 74k records, then models.RelevanceScore — the same engine that orders
// every other source's results — over the survivors. Ties break dub-first
// (the record an audio picker would choose anyway), then newest-first.
func (cc *corpus) rank(q string) []Media {
	nq := slugTokens(queryFoldRe.Replace(q))
	if nq == "" {
		return nil
	}
	words := strings.Fields(nq)
	significant := 0
	for _, w := range words {
		if len([]rune(w)) > 1 && !rankStopWords[w] {
			significant++
		}
	}
	if significant == 0 {
		return nil
	}

	cc.mu.Lock()
	defer cc.mu.Unlock()
	if len(cc.entries) == 0 {
		return nil
	}

	type survivor struct {
		i  int
		sc int
	}
	survivors := make([]survivor, 0, 64)
	// The pre-filter keeps records where every query word appears in the
	// tokens (substring is enough: "todo poderoso" must hit "a volta do
	// todo poderoso" — token-level prefix work is RelevanceScore's job).
	for i := range cc.entries {
		t := cc.entries[i].Tokens
		if t == "" {
			continue
		}
		every := true
		for _, w := range words {
			if !strings.Contains(t, w) {
				every = false
				break
			}
		}
		if every {
			survivors = append(survivors, survivor{i: i})
			if len(survivors) >= refinementBudget {
				break
			}
		}
	}
	if len(survivors) == 0 {
		return nil
	}
	for j := range survivors {
		e := &cc.entries[survivors[j].i]
		survivors[j].sc = models.RelevanceScore(nq, e.Tokens, "")
	}
	sort.SliceStable(survivors, func(a, b int) bool {
		ei, ej := &cc.entries[survivors[a].i], &cc.entries[survivors[b].i]
		if survivors[a].sc != survivors[b].sc {
			return survivors[a].sc > survivors[b].sc
		}
		if ei.Audio != ej.Audio {
			return ei.Audio == AudioDubbed
		}
		return ei.VideoID > ej.VideoID
	})

	out := make([]Media, 0, maxSearchResults)
	seen := map[string]bool{}
	for _, s := range survivors {
		e := cc.entries[s.i]
		if s.sc <= 0 {
			continue
		}
		key := e.Tokens
		if seen[key] {
			continue
		}
		if len(out) >= maxSearchResults {
			break
		}
		seen[key] = true
		out = append(out, Media{
			Title:   slugToTitle(e.BaseSlug),
			URL:     e.URL,
			Kind:    e.Kind,
			Audio:   e.Audio,
			VideoID: e.VideoID,
		})
	}
	return out
}

// slugToTitle turns the sitemap's slug into a display title: dashes become
// spaces, each word capitalized. The result is accent-free — slugs are — but
// reads like the site's own.
func slugToTitle(slug string) string {
	words := strings.Fields(slugTokens(slug))
	out := make([]byte, 0, len(slug)+2)
	for i, w := range words {
		if i > 0 {
			out = append(out, ' ')
		}
		if w == "" {
			continue
		}
		b := []byte(w)
		if b[0] >= 'a' && b[0] <= 'z' {
			b[0] -= 'a' - 'A'
		}
		out = append(out, b...)
	}
	return string(out)
}
