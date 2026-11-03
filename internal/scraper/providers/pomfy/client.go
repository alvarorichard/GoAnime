package pomfy

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/api/movie"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/alvarorichard/Goanime/v2/internal/util/jsonx"
)

// Client talks to the Pomfy watch site (pomfy.my) and, through the Byse
// captcha flow in byse.go, to the CDN the players run on.
//
// The site's own API needs no browser and no challenge:
//
//	GET /api/catalogo-ids  → {"filmes": [tmdb…], "series": [tmdb…]}
//	GET /api/watch-sources?id=<tmdb>&tipo=filme|serie[&temporada=&episodio=]
//	    → {"mode": "direct", "sources": [{"kind": "byse", "label": …,
//	        "embedUrl": "https://pomfy-cdn.shop/e/<code>", "token": …}]}
//
// Everything else (title metadata, season lists, per-episode names) is TMDB
// data the site embeds in its SSR payloads; this file decodes that unchanged.
// Text search runs on TMDB's search API (internal/api/movie — the client the
// other movie sources already use) and filters hits through the catalog:
// Pomfy itself never maps words to ids.

const (
	// SourceName is the canonical Source string Pomfy results carry.
	SourceName = "Pomfy"

	// DefaultBase is the watch site.
	DefaultBase = "https://pomfy.my"

	// DefaultBaseHost is DefaultBase's host, what the X-Embed-Origin header
	// carries (pomfy.online used to host it and still answers).
	DefaultBaseHost = "pomfy.my"

	// DefaultWatchSite is the default page the resolution says the title is
	// watched on (their site's own layout: /assistir/<tmdb> for movies and
	// /serie/<tmdb> for series, appearing only when no real page is known).
	DefaultWatchSite = DefaultBase

	// tmdbImageHost prefixes the poster paths the site payloads carry.
	tmdbImageHost = "image.tmdb.org"

	// apiCatalog is the second host of the API family: season listings live
	// there (the site dedicated to partners' iframes).
	apiCatalog = "https://api.pomfy.stream"

	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36"

	maxBodyBytes = 4 << 20

	// catalogTTL is how long the catalogo-ids answer stays trusted; the
	// catalog rotates slowly.
	catalogTTL = 6 * time.Hour
)

// Content signals — what the site does not offer, not a scraping fault.
var (
	// ErrNotOnCatalog is returned when a title is not on pomfy.
	ErrNotOnCatalog = errors.New("pomfy: title is not on the catalog")

	// ErrNoSources is returned when watch-sources lists no player for a title.
	ErrNoSources = errors.New("pomfy: no watch sources listed for this title")

	// ErrNoTMDBAPIKey is the search prerequisite: Pomfy's API is keyed by
	// TMDB id, so text search needs TMDB to translate names to ids.
	ErrNoTMDBAPIKey = errors.New("pomfy: text search needs the TMDB_API_KEY environment variable")

	// ErrNoEpisodeSource is one episode having no playable listing while the
	// show as a whole does.
	ErrNoEpisodeSource = errors.New("pomfy: this episode has no watch sources yet")

	// ErrBadTMDBSearch is the TMDB client missing its key without the search
	// sum even starting.
	ErrBadTMDBSearch = errors.New("pomfy: tmdb search is not configured")
)

// MediaKind tells movies from series, using the Portuguese words the API field
// (tipo) is spelled with.
type MediaKind string

const (
	KindMovie  MediaKind = "filme"
	KindSeries MediaKind = "serie"
)

// Client is the Pomfy front. Catalog data lives in a small TTL cache.
type Client struct {
	http        *http.Client
	baseURL     string
	catalogBase string // the season-listing host; defaults to apiCatalog
	userAgent   string
	maxRetries  int
	retryDelay  time.Duration

	// tmdbSearch is the text-search seam; nil means the real TMDB client
	// (internal/api/movie). Tests point it at fakes.
	tmdbSearch func(query string) ([]Media, error)

	mu      sync.Mutex
	catalog *catalog
}
type catalog struct {
	fetched time.Time
	movies  map[int64]bool
	series  map[int64]bool
}

// NewClient builds a client; no network I/O happens at construction.
func NewClient() *Client {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("GOANIME_POMFY_URL")), "/")
	if base == "" {
		base = DefaultBase
	}
	return &Client{
		http:        util.NewFastClient(),
		baseURL:     base,
		catalogBase: apiCatalog,
		userAgent:   userAgent,
		maxRetries:  2,
		retryDelay:  400 * time.Millisecond,
	}
}

// NewClientForTest points a client at a test server.
func NewClientForTest(hc *http.Client, baseURL string) *Client {
	return &Client{
		http:        hc,
		baseURL:     strings.TrimRight(baseURL, "/"),
		catalogBase: strings.TrimRight(baseURL, "/"),
		userAgent:   userAgent,
	}
}

var (
	sharedOnce   sync.Once
	sharedClient *Client
)

// Shared returns the process-wide client, so the catalog cache is shared.
func Shared() *Client {
	sharedOnce.Do(func() { sharedClient = NewClient() })
	return sharedClient
}

// BaseURL is the watch site this client talks to.
func (c *Client) BaseURL() string { return c.baseURL }

// request is one HTTP call in this client's shape: browser UA, Chrome accept
// language, JSON answers preferred, bounded bodies, retry on 5xx/429.
func (c *Client) get(ctx context.Context, rawURL string, headers map[string]string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.retryDelay * time.Duration(attempt)):
			}
		}
		body, retry, err := c.attempt(ctx, http.MethodGet, rawURL, nil, "application/json", headers)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry || ctx.Err() != nil {
			break
		}
	}
	return nil, lastErr
}

// attempt performs one request and classifies the answer.
func (c *Client) attempt(ctx context.Context, method, rawURL string, bodySrc []byte, accept string, headers map[string]string) (body []byte, retry bool, err error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, http.NoBody)
	if err != nil {
		return nil, false, err
	}
	if bodySrc != nil {
		req.Body = io.NopCloser(bytes.NewReader(bodySrc))
		req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(bodySrc)), nil }
		req.ContentLength = int64(len(bodySrc))
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept-Language", netx.ChromeAcceptLanguage)
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if bodySrc != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, ctx.Err() == nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err = io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, true, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return body, false, nil
	}
	return nil, resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests,
		netx.NewHTTPStatusError(SourceName, "http", resp.StatusCode)
}

// postJSON sends payload as JSON to base+path (base is a scheme://host URL)
// and reads the JSON answer back, with a short retry (the captcha flow is
// idempotent enough: each call buys a fresh challenge).
func (c *Client) postJSON(ctx context.Context, base, path string, payload any, headers map[string]string) ([]byte, error) {
	wire, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.retryDelay * time.Duration(attempt)):
			}
		}
		body, retry, err := c.attempt(ctx, http.MethodPost, base+path, wire, "application/json", headers)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry || ctx.Err() != nil {
			break
		}
	}
	return nil, lastErr
}

// catalogData fetches (or reads from cache) the whole catalog: every
// registered TMDB id, keyed by kind.
func (c *Client) catalogData(ctx context.Context) (*catalog, error) {
	c.mu.Lock()
	if c.catalog != nil && time.Since(c.catalog.fetched) < catalogTTL {
		cat := c.catalog
		c.mu.Unlock()
		return cat, nil
	}
	c.mu.Unlock()

	body, err := c.get(ctx, c.baseURL+"/api/catalogo-ids", nil)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Filmes []int64 `json:"filmes"`
		Series []int64 `json:"series"`
	}
	if err := jsonx.Unmarshal(body, &parsed); err != nil {
		return nil, netx.NewParserError(SourceName, "catalog", "unreadable catalog list", err)
	}
	cat := &catalog{
		fetched: time.Now(),
		movies:  make(map[int64]bool, len(parsed.Filmes)),
		series:  make(map[int64]bool, len(parsed.Series)),
	}
	for _, id := range parsed.Filmes {
		cat.movies[id] = true
	}
	for _, id := range parsed.Series {
		cat.series[id] = true
	}
	c.mu.Lock()
	c.catalog = cat
	c.mu.Unlock()
	return cat, nil
}

// HasTitle reports whether a TMDB id is on the catalog.
func (c *Client) HasTitle(ctx context.Context, tmdbID int64, kind MediaKind) (bool, error) {
	cat, err := c.catalogData(ctx)
	if err != nil {
		return false, err
	}
	switch kind {
	case KindSeries:
		return cat.series[tmdbID], nil
	default:
		return cat.movies[tmdbID], nil
	}
}

// watchSource is one player watch-sources lists.
type watchSource struct {
	Kind        string `json:"kind"`
	Label       string `json:"label"`
	Description string `json:"description"`
	EmbedURL    string `json:"embedUrl"`
	Token       string `json:"token"`
}

type watchSources struct {
	Mode    string        `json:"mode"`
	Error   string        `json:"error"`
	Sources []watchSource `json:"sources"`
}

// ByseCode extracts an /e/<code> video code from an embed URL.
func ByseCode(embedURL string) string {
	u, err := url.Parse(strings.TrimSpace(embedURL))
	if err != nil || u.Path == "" {
		return ""
	}
	// Path layout /e/<code> or /e/<code>/slug.
	rest := strings.TrimPrefix(u.Path, "/e/")
	rest = strings.TrimSpace(rest)
	if i := strings.IndexByte(rest, '/'); i >= 0 {
		rest = rest[:i]
	}
	if !isByseCode(rest) {
		return ""
	}
	return rest
}

// isByseCode is the shape the CDN codes have shown so far: 10-16 lowercase
// letters/digits (gkcq8gbiwh4r, yyd28sf1yssp…).
func isByseCode(s string) bool {
	if len(s) < 6 || len(s) > 24 {
		return false
	}
	for _, r := range s {
		low := r >= 'a' && r <= 'z'
		dig := r >= '0' && r <= '9'
		if !low && !dig {
			return false
		}
	}
	return true
}

// WatchSources lists the players for a title (its byse code, most often one).
func (c *Client) WatchSources(ctx context.Context, tmdbID int64, kind MediaKind, season, episode int) ([]watchSource, error) {
	q := url.Values{}
	q.Set("id", strconv.FormatInt(tmdbID, 10))
	q.Set("tipo", string(kind))
	if kind == KindSeries {
		q.Set("temporada", strconv.Itoa(season))
		q.Set("episodio", strconv.Itoa(episode))
	}
	body, err := c.get(ctx, c.baseURL+"/api/watch-sources?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	var parsed watchSources
	if err := jsonx.Unmarshal(body, &parsed); err != nil {
		return nil, netx.NewParserError(SourceName, "watch-sources", "unreadable watch sources", err)
	}
	if parsed.Error != "" {
		return nil, fmt.Errorf("%w: %s", ErrNoSources, parsed.Error)
	}
	out := make([]watchSource, 0, len(parsed.Sources))
	for _, s := range parsed.Sources {
		if strings.EqualFold(s.Kind, "byse") && ByseCode(s.EmbedURL) != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil, ErrNoSources
	}
	return out, nil
}

// MoviePlayURL is the watch page of a movie; used as the referer/episode URL.
func (c *Client) MoviePlayURL(tmdbID int64) string {
	return c.baseURL + "/assistir/" + strconv.FormatInt(tmdbID, 10)
}

// SeriesPlayURL is the machine's identifier URL of one series episode: the
// partner players' path (the api host's own player page). Episode data is
// always re-derived from the ids, so a 404 on the URL itself is fine.
func SeriesPlayURL(tmdbID int64, season, episode int) string {
	return apiCatalog + "/serie/" + strconv.FormatInt(tmdbID, 10) + "/" +
		strconv.Itoa(season) + "/" + strconv.Itoa(episode)
}

// ResolveByse runs a watch source through the captcha flow, once.
func (c *Client) ResolveByse(ctx context.Context, src watchSource, watchPageURL string) (*Stream, error) {
	code := ByseCode(src.EmbedURL)
	if code == "" {
		return nil, netx.NewParserError(SourceName, detailLayer, "watch source carries no byse code", nil)
	}
	u, perr := url.Parse(strings.TrimSpace(src.EmbedURL))
	var apiBase string
	if perr == nil && u != nil && u.Host != "" {
		apiBase = u.Scheme + "://" + u.Host
	}
	if apiBase == "" {
		// A watch source that cannot parse names no host at all — refusing
		// here keeps the resolver from dialing a stale name blind.
		return nil, netx.NewParserError(SourceName, detailLayer, "watch source embed unreadable", nil)
	}
	return resolveByse(ctx, c, apiBase, code, watchPageURL)
}

// ResolveStream mirrors what the player dispatch wants: the best byse source
// for a title, freshly captcha-checked.
func (c *Client) ResolveStream(ctx context.Context, tmdbID int64, kind MediaKind, season, episode int, watchPage string) (*Stream, error) {
	sources, err := c.WatchSources(ctx, tmdbID, kind, season, episode)
	if err != nil {
		return nil, err
	}
	var lastErr error
	for _, src := range sources {
		stream, err := c.ResolveByse(ctx, src, watchPage)
		if err == nil {
			return stream, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("%w: %v", ErrNoSources, lastErr)
}

// --- title metadata (the site's serialized SSR payloads) ---

// ShowInfo is the TMDB TV dump the serie pages carry.
type ShowInfo struct {
	Name       string
	Overview   string
	Year       string
	PosterPath string
	Seasons    []ShowSeason
}

// ShowSeason is one season entry: number, episode count (TMDB), name.
type ShowSeason struct {
	Number       int
	EpisodeCount int
	Name         string
}

// ShowInfo reads the show's TMDB metadata (and its season list) from the
// serial pages' server-rendered payload — no TMDB key asked for.
func (c *Client) ShowInfo(ctx context.Context, tmdbID int64) (*ShowInfo, error) {
	body, err := c.get(ctx, c.baseURL+"/serie/"+strconv.FormatInt(tmdbID, 10)+"/__data.json", nil)
	if err != nil {
		return nil, err
	}
	return parseShowInfo(body)
}

func parseShowInfo(body []byte) (*ShowInfo, error) {
	var parsed struct {
		Nodes []struct {
			Data json.RawMessage `json:"data"`
		} `json:"nodes"`
	}
	if err := jsonx.Unmarshal(body, &parsed); err != nil {
		return nil, netx.NewParserError(SourceName, "show", "unreadable show payload", err)
	}
	for _, node := range parsed.Nodes {
		var arr []json.RawMessage
		if jsonx.Unmarshal(node.Data, &arr) != nil {
			continue
		}
		// node[0] is {"topbar": …}; node[1] spells the whole show. The
		// devalue decode walks whatever array the keys point at.
		for _, raw := range arr {
			var keys map[string]json.RawMessage
			if jsonx.Unmarshal(raw, &keys) != nil {
				continue
			}
			if _, has := keys["tvShow"]; !has {
				continue
			}
			show, err := decodeShowNode(arr)
			if err != nil {
				return nil, err
			}
			return show, nil
		}
	}
	return nil, netx.NewParserError(SourceName, "show", "show data not found in payload", nil)
}

// decodeShowNode runs the devalue walk over the whole node array and reads
// the television fields at index 1 ({"tvShow": 1}).
func decodeShowNode(arr []json.RawMessage) (*ShowInfo, error) {
	if len(arr) == 0 {
		return nil, netx.NewParserError(SourceName, "show", "empty payload", nil)
	}
	graph := newDevalueGraph(arr)
	var head map[string]int64
	if err := jsonx.Unmarshal(arr[0], &head); err == nil {
		if idx, ok := head["tvShow"]; ok && idx >= 0 && int(idx) < graph.size() {
			return decodeShow(graph, int(idx))
		}
	}
	return decodeShowAt(arr)
}

// decodeShowAt is the fallback scan for the payloads whose head entry does
// not carry a usable tvShow pointer: it looks for the tvShow-rooted object
// itself (the only array entry that spells the seasons key AND is an
// object), then decodes the record from there.
func decodeShowAt(arr []json.RawMessage) (*ShowInfo, error) {
	for i, raw := range arr {
		if !bytes.Contains(raw, seasonsBytes) {
			continue
		}
		g := newDevalueGraph(arr)
		if show, err := decodeShow(g, i); err == nil {
			return show, nil
		}
	}
	return nil, netx.NewParserError(SourceName, "show", "tvShow entry missing", nil)
}

// seasonsBytes is the JSON key the show object (and nothing else in the
// payload) carries; the pointer map spelling the tvShow name does not.
var seasonsBytes = []byte(`"seasons"`)

func decodeShow(g *devalueGraph, idx int) (*ShowInfo, error) {
	obj, ok := g.object(int64(idx))
	if !ok {
		return nil, netx.NewParserError(SourceName, "show", "tvShow not an object", nil)
	}
	name, hasName := g.stringRef(obj["name"])
	_ = hasName
	summary, _ := g.stringRef(obj["overview"])
	poster, _ := g.stringRef(obj["poster_path"])
	airDate, _ := g.stringRef(obj["first_air_date"])
	out := &ShowInfo{Name: name, Overview: summary, PosterPath: poster}
	if len(airDate) >= 4 {
		out.Year = airDate[:4]
	}
	if seasonsIdx, has := obj["seasons"]; has {
		for _, si := range g.array(seasonsIdx) {
			sObj, ok := g.object(si)
			if !ok {
				continue
			}
			var s ShowSeason
			if n, has := sObj["season_number"]; has {
				s.Number = int(g.integerRef(n))
			}
			if n, has := sObj["episode_count"]; has {
				s.EpisodeCount = int(g.integerRef(n))
			}
			s.Name, _ = g.stringRef(sObj["name"])
			out.Seasons = append(out.Seasons, s)
		}
	}
	return out, nil
}

// --- seasonal episode listing on the partner host (still keyless) ---

// SeasonEpisode is one available episode of one season.
type SeasonEpisode struct {
	Number int
	Title  string
	Still  string
}

// SeasonRet lists what /api/tmdb-season knows about one season: the TMDB
// episode numbers, the ones really available on the CDN, and their names and
// stills.
type SeasonRet struct {
	Season        int
	EpisodeNums   []int
	AvailableNums []int
	Episodes      []SeasonEpisode
}

// Season reads one season's episodes from the partner API of the catalog.
// It answers keyless, unlike TMDB itself.
func (c *Client) Season(ctx context.Context, tmdbID int64, season int) (*SeasonRet, error) {
	q := url.Values{}
	q.Set("id", strconv.FormatInt(tmdbID, 10))
	q.Set("s", strconv.Itoa(season))
	body, err := c.get(ctx, c.catalogBase+"/api/tmdb-season?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		EpNums  []int `json:"epNums"`
		SbEps   []int `json:"sbEps"`
		TMDBEps map[string]struct {
			Name  string `json:"name"`
			Still string `json:"still"`
		} `json:"tmdbEps"`
	}
	if err := jsonx.Unmarshal(body, &parsed); err != nil {
		return nil, netx.NewParserError(SourceName, "season", "unreadable season list", err)
	}
	ret := &SeasonRet{Season: season}
	seen := make(map[int]bool, len(parsed.EpNums))
	for _, n := range parsed.EpNums {
		if n > 0 && !seen[n] {
			seen[n] = true
			ret.EpisodeNums = append(ret.EpisodeNums, n)
		}
	}
	avail := make(map[int]bool, len(parsed.SbEps))
	for _, n := range parsed.SbEps {
		avail[n] = true
	}
	for _, n := range ret.EpisodeNums {
		if !avail[n] {
			continue
		}
		ret.AvailableNums = append(ret.AvailableNums, n)
	}
	for _, n := range ret.AvailableNums {
		meta := parsed.TMDBEps[strconv.Itoa(n)]
		ret.Episodes = append(ret.Episodes, SeasonEpisode{Number: n, Title: meta.Name, Still: meta.Still})
	}
	if len(ret.EpisodeNums) == 0 {
		return nil, fmt.Errorf("%w: season %d has no episodes on the catalog", ErrNoEpisodeSource, season)
	}
	return ret, nil
}

// --- text search: TMDB keywords filtered through the catalog ---

// Media is one search result. The search seam produces it with Kind spelled
// in TMDB's way (movie|tv) and Search stamps the catalog's MediaKind and
// identifier URL after the catalog filter keeps it.
type Media struct {
	Title    string
	TMDBID   int64
	URL      string // the identifier page; see TitleURL
	Poster   string
	Year     string
	Rating   float64
	Overview string
	Genres   []string
	Kind     MediaKind
	TMDBType string // "movie"|"tv", the TMDB spelling the seam wrote
}

// ToAnimeModel converts the result into the app-wide model. The identifier
// URL carries the TMDB id (Pomfy's own keying), so playback needs no page.
func (m Media) ToAnimeModel() *models.Anime {
	anime := &models.Anime{
		Name:     m.Title,
		URL:      m.URL,
		ImageURL: m.Poster,
		Source:   SourceName,
		Year:     m.Year,
		TMDBID:   int(m.TMDBID),
	}
	anime.MediaType = models.MediaTypeTV
	if m.Kind == KindMovie {
		anime.MediaType = models.MediaTypeMovie
	}
	if m.Rating > 0 {
		anime.Rating = m.Rating
	}
	if m.Overview != "" {
		anime.Overview = m.Overview
	}
	if len(m.Genres) > 0 {
		anime.Genres = m.Genres
	}
	return anime
}

// tmdbSearchFn is the process-wide TMDB seam; per-client overrides win.
var tmdbSearchFn = tmdbSearch

// SetSearchSeam points this client's text search at fn. Tests only: it is
// the one place a fake TMDB joins the catalog filter.
func (c *Client) SetSearchSeam(fn func(query string) ([]Media, error)) {
	c.tmdbSearch = fn
}

func (c *Client) searchSeam() func(query string) ([]Media, error) {
	if c.tmdbSearch != nil {
		return c.tmdbSearch
	}
	return tmdbSearchFn
}

// Search queries TMDB for titles named like the query and keeps the ones the
// Pomfy catalog carries. Without a TMDB key it errors (ErrNoTMDBAPIKey) —
// callers surface that to the user.
func (c *Client) Search(ctx context.Context, query string) ([]Media, error) {
	q := strings.Join(strings.Fields(strings.NewReplacer("-", " ", "_", " ", ".", " ").Replace(query)), " ")
	if q == "" {
		return nil, nil
	}
	cat, err := c.catalogData(ctx)
	if err != nil {
		return nil, err
	}
	// A bare TMDB id needs no name search at all: if the number is on the
	// catalog, resolve the title from the site's own SSR payloads (keyless)
	// and answer with that one title.
	if id, nerr := strconv.ParseInt(q, 10, 64); nerr == nil && id > 0 {
		if title, ok := c.TitleByIDs(ctx, id, cat); ok {
			return []Media{title}, nil
		}
		return nil, ErrNotOnCatalog
	}
	results, err := c.searchSeam()(q)
	if err != nil {
		return nil, err
	}
	out := make([]Media, 0, len(results))
	for _, r := range results {
		if r.TMDBType == "" && r.Kind == "" {
			continue
		}
		kind := r.Kind
		if kind == "" {
			if r.TMDBType == "tv" {
				kind = KindSeries
			} else {
				kind = KindMovie
			}
		}
		set := cat.movies
		if kind == KindSeries {
			set = cat.series
		}
		if set[r.TMDBID] {
			r.Kind = kind
			r.URL = c.TitleURL(r.TMDBID, kind)
			out = append(out, r)
		}
	}
	return out, nil
}

// TitleByIDs builds one search result for a TMDB id known to the catalog,
// reading title and year from the watching site's own server payload (no
// TMDB key involved). ok=false when the id is not on the catalog.
func (c *Client) TitleByIDs(ctx context.Context, id int64, cat *catalog) (Media, bool) {
	if cat == nil {
		var err error
		cat, err = c.catalogData(ctx)
		if err != nil {
			return Media{}, false
		}
	}
	switch {
	case cat.movies[id]:
		info, err := c.MovieTitle(ctx, id)
		if err != nil {
			return Media{}, false
		}
		out := Media{Title: info.Name, TMDBID: id, Kind: KindMovie,
			Year: info.Year, Poster: info.Poster, Overview: info.Overview}
		out.URL = c.TitleURL(id, KindMovie)
		return out, true
	case cat.series[id]:
		info, err := c.ShowInfo(ctx, id)
		if err != nil {
			return Media{}, false
		}
		out := Media{Title: info.Name, TMDBID: id, Kind: KindSeries,
			Year: info.Year, Poster: "https://" + tmdbImageHost + info.PosterPath, Overview: info.Overview}
		out.URL = c.TitleURL(id, KindSeries)
		return out, true
	default:
		return Media{}, false
	}
}

// MovieTitle is what a movie's assistir page says about the title: its own
// TMDB dump, keyless (assistir/<id>/__data.json carries id/tipo/title/year/
// overview/image).
func (c *Client) MovieTitle(ctx context.Context, id int64) (*Title, error) {
	body, err := c.get(ctx, c.baseURL+"/assistir/"+strconv.FormatInt(id, 10)+"/__data.json", nil)
	if err != nil {
		return nil, err
	}
	return parseMovieTitle(body)
}

// Title is one title's site-side metadata shape (both flavors use it).
type Title struct {
	Name      string
	Year      string
	Overview  string
	Poster    string
	Available []int // seasons for series, empty for movies
}

func parseMovieTitle(body []byte) (*Title, error) {
	var parsed struct {
		Nodes []struct {
			Data json.RawMessage `json:"data"`
		} `json:"nodes"`
	}
	if err := jsonx.Unmarshal(body, &parsed); err != nil {
		return nil, netx.NewParserError(SourceName, "movie-title", "unreadable title payload", err)
	}
	for _, node := range parsed.Nodes {
		var arr []json.RawMessage
		if jsonx.Unmarshal(node.Data, &arr) != nil {
			continue
		}
		for i, raw := range arr {
			var keys map[string]json.RawMessage
			if jsonx.Unmarshal(raw, &keys) != nil {
				continue
			}
			if _, has := keys["tipo"]; !has {
				continue
			}
			// The walk guarantees the entry at i is the pointer object, so
			// decodeTitleNode answers without an error path here.
			return decodeTitleNode(arr, i), nil
		}
	}
	return nil, netx.NewParserError(SourceName, "movie-title", "title entry missing", nil)
}

// decodeTitleNode walks the devalue array out of the {"id","tipo","title",…}
// entry at index start. The walk guarantees a pointer object there, so the
// read answers a plain value without an error lane.
func decodeTitleNode(arr []json.RawMessage, start int) *Title {
	g := newDevalueGraph(arr)
	obj, _ := g.object(int64(start))
	out := &Title{}
	out.Name, _ = g.stringRef(obj["title"])
	year, _ := g.stringRef(obj["year"])
	if len(year) >= 4 {
		out.Year = year[:4]
	}
	out.Overview, _ = g.stringRef(obj["overview"])
	image, _ := g.stringRef(obj["image"])
	if image != "" {
		out.Poster = "https://" + tmdbImageHost + image
	}
	if seasonsIdx, has := obj["availableSeasons"]; has {
		for _, si := range g.array(seasonsIdx) {
			out.Available = append(out.Available, int(g.integerRef(si)))
		}
	}
	return out
}

// TitleURL is the identifier URL of a TMDB id in this client's watch site.
func (c *Client) TitleURL(tmdbID int64, kind MediaKind) string {
	if kind == KindSeries {
		return c.baseURL + "/serie/" + strconv.FormatInt(tmdbID, 10)
	}
	return c.MoviePlayURL(tmdbID)
}

// TMDBConfigured tells whether this process can run the text search. A user
// key (TMDB_API_KEY) always wins; without one the search runs on the public
// key pool the watching site itself ships in its bundle, so the answer is
// always true. Health-check reports use it only as a documentation anchor.
func TMDBConfigured() bool {
	return movie.NewTMDBClient().IsConfigured() || len(publicTMDBKeys) > 0
}

// publicTMDBKeys is the api-key pool the watching site ships in its own
// browser bundle (pomfy.my/_app/immutable/chunks, read 2026-11-02): every
// visitor's browser searches TMDB with one of these, rotated randomly with
// retries on quota errors. Using them keeps Pomfy's text search working
// without the user carrying a key of their own; a user key from TMDB_API_KEY
// always takes precedence, since its quota is dedicated.
var publicTMDBKeys = []string{
	"fb7bb23f03b6994dafc674c074d01761",
	"e55425032d3d0f371fc776f302e7c09b",
	"8301a21598f8b45668d5711a814f01f6",
	"8cf43ad9c085135b9479ad5cf6bbcbda",
	"da63548086e399ffc910fbc08526df05",
	"13e53ff644a8bd4ba37b3e1044ad24f3",
	"269890f657dddf4635473cf4cf456576",
	"a2f888b27315e62e471b2d587048f32e",
	"8476a7ab80ad76f0936744df0430e67c",
	"5622cafbfe8f8cfe358a29c53e19bba0",
	"ae4bd1b6fce2a5648671bfc171d15ba4",
	"257654f35e3dff105574f97fb4b97035",
	"2f4038e83265214a0dcd6ec2eb3276f5",
	"9e43f45f94705cc8e1d5a0400d19a7b7",
	"af6887753365e14160254ac7f4345dd2",
	"06f10fc8741a672af455421c239a1ffc",
	"09ad8ace66eec34302943272db0e8d2c",
}

// PublicKeyForTesting returns key n of the public pool, for the diagnostics
// and the meta-test that proves the pool answers like the site's browser.
func PublicKeyForTesting(n int) string { return publicTMDBKeys[n] }

// tmdbSearchKeyErrs are the statuses the site itself retries on: a public key
// is over its hourly quota or expired; the next one answers.
var tmdbSearchKeyErrs = func(code int) bool {
	return code == http.StatusUnauthorized || code == http.StatusForbidden ||
		code == http.StatusTooManyRequests || code >= 500
}

// The site searches three pages of /search/multi per query, localized; this
// file does the same, and keeps a small TTL cache (the site uses 1h in
// sessionStorage).
const (
	tmdbSearchPages = 3
	tmdbTLang       = "pt-BR"
	tmdbCacheTTL    = time.Hour
)

// tmdbBase is TMDB's API root; a seam the tests repoint at their fake.
var tmdbBase = "https://api.themoviedb.org/3"

var (
	tmdbCacheMu sync.Mutex
	tmdbCache   = map[string]tmdbCacheEntry{}
)

type tmdbCacheEntry struct {
	items   []Media
	expires time.Time
}

// movieClientOp is the plain HTTP client the search uses (the movie package's
// TMDB client takes the env key only).
var movieTMDBSearchHTTP = func() *http.Client {
	hc := util.NewFastClient()
	return hc
}()

// newTMDBClientFn is the user-key client factory; a seam tests swap so the
// user-key branch drives a fake TMDB instead of the real network.
var newTMDBClientFn = userTMDBClientFn

// userTMDBClientFn is the real (env-driven) factory the user-key branch
// runs on; withFakeClient swaps `newTMDBClientFn` and its cleanup restores
// this one.
func userTMDBClientFn() *movie.TMDBClient { return movie.NewTMDBClient() }

// tmdbSearch runs a real TMDB multi search over up to three pages,
// mapping its raw items onto Media. The user key (TMDB_API_KEY) runs first;
// the public pool replaces it when the env var is unset, rotating on quota
// failures exactly the way the watching site's bundle does.
func tmdbSearch(query string) ([]Media, error) {
	if raw, ok := tmdbLookupCache(query); ok {
		return raw, nil
	}
	if client := newTMDBClientFn(); client.IsConfigured() {
		out, err := movieSearchMulti(client, query)
		if err == nil {
			tmdbStoreCache(query, out)
			return out, nil
		}
		// A failing user key hands the search to the pool whatever the
		// error is: quota-depleted keys are the common death and the user
		// client's error spell is the movie package's own.
	}
	var lastErr error
	for _, idx := range shuffledIndexes(len(publicTMDBKeys)) {
		movies, err := searchMultiWithKey(publicTMDBKeys[idx], query)
		if err == nil {
			tmdbStoreCache(query, movies)
			return movies, nil
		}
		lastErr = err
		var herr *netx.SourceDiagnostic
		if errors.As(err, &herr) && tmdbSearchKeyErrs(herr.StatusCode) {
			continue // quota/next key; the site retries the same way
		}
		return nil, err
	}
	if lastErr == nil {
		lastErr = ErrNoTMDBAPIKey
	}
	return nil, lastErr
}

func movieSearchMulti(client *movie.TMDBClient, query string) ([]Media, error) {
	result, err := client.SearchMulti(query)
	if err != nil {
		return nil, err
	}
	return mapTMDBItems(result.Results), nil
}

// searchMultiWithKey hits one page triple with the given key and maps the
// results onto Media. tmdbBase is the package seam the tests point at their
// fake server; HTTP status errors carry the status for rotation.
func searchMultiWithKey(key, query string) ([]Media, error) {
	q := url.QueryEscape(query)
	var raw []models.TMDBMedia
	for page := 1; page <= tmdbSearchPages; page++ {
		target := tmdbBase + fmt.Sprintf("/search/multi?query=%s&include_adult=false&language=%s&page=%d&api_key=%s",
			q, url.QueryEscape(tmdbTLang), page, key)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		// The target string is built from a fixed format, so NewRequest
		// cannot fail here; a dial failure surfaces from Do below.
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, target, http.NoBody)
		resp, err := movieTMDBSearchHTTP.Do(req)
		if err != nil || resp == nil {
			cancel()
			return nil, err
		}
		// The body must be read BEFORE the context dies: cancel() here would
		// make the read below fail with "context canceled", a failure that
		// looks like an abort instead of what it is.
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
		_ = resp.Body.Close()
		cancel()
		if readErr != nil {
			return nil, readErr
		}
		if resp.StatusCode >= 300 {
			return nil, netx.NewHTTPStatusError(SourceName, "tmdb", resp.StatusCode)
		}
		var parsed struct {
			Results []models.TMDBMedia `json:"results"`
		}
		if jsonx.Unmarshal(body, &parsed) != nil {
			return nil, netx.NewParserError(SourceName, "tmdb", "search answer unreadable", nil)
		}
		raw = append(raw, parsed.Results...)
		if len(parsed.Results) < 20 {
			break // last page
		}
	}
	return mapTMDBItems(raw), nil
}

func mapTMDBItems(items []models.TMDBMedia) []Media {
	out := make([]Media, 0, len(items))
	for _, item := range items {
		if item.MediaType != "movie" && item.MediaType != "tv" {
			continue
		}
		kind := KindMovie
		title := item.Title
		if item.MediaType == "tv" {
			kind = KindSeries
			title = item.Name
		}
		out = append(out, Media{
			Title:    title,
			TMDBID:   int64(item.ID),
			TMDBType: item.MediaType,
			Poster:   item.GetPosterURL("w300"),
			Year:     item.GetReleaseYear(),
			Rating:   item.VoteAverage,
			Overview: item.Overview,
			Kind:     kind,
		})
	}
	return out
}

func tmdbLookupCache(query string) ([]Media, bool) {
	tmdbCacheMu.Lock()
	defer tmdbCacheMu.Unlock()
	entry, ok := tmdbCache[query]
	if !ok || time.Now().After(entry.expires) {
		return nil, false
	}
	return entry.items, true
}

func tmdbStoreCache(query string, items []Media) {
	tmdbCacheMu.Lock()
	defer tmdbCacheMu.Unlock()
	tmdbCache[query] = tmdbCacheEntry{items: items, expires: time.Now().Add(tmdbCacheTTL)}
}

// shuffledIndexes is a Fisher-Yates over 0..n-1 driven by crypto/rand, the
// site's random pick without a math/rand detector.
func shuffledIndexes(n int) []int {
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	for i := n - 1; i > 0; i-- {
		roll, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return idx // degenerate to order-on-failure; the pool still answers
		}
		j := int(roll.Int64())
		idx[i], idx[j] = idx[j], idx[i]
	}
	return idx
}
