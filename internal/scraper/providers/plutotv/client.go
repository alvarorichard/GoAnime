package plutotv

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/alvarorichard/Goanime/v2/internal/util/jsonx"
)

const (
	// SourceName is the canonical source name used in saved titles.
	SourceName = "PlutoTV"
	// DefaultBootURL is the session service, checked on 2026-11-16.
	DefaultBootURL = "https://boot.pluto.tv"
	// siteBase is the origin the identifier URLs are spelled under.
	siteBase = "https://pluto.tv"

	// The app identity the session is opened with: the site's own web client.
	appName       = "web"
	appVersion    = "9.0.0"
	deviceVersion = "126.0.0"

	// searchLimit is how many hits one search asks for. Channels and timelines
	// share the list with movies and series, so it is larger than what is shown.
	searchLimit = 60

	// A long-running series lists every episode of every season in one answer
	// (about 1.4KB each), and a two-hour movie's playlist is close to 1MB.
	maxJSONBytes     = 16 << 20
	maxPlaylistBytes = 8 << 20

	// sessionLifetime bounds how long a session is reused when the boot answer
	// does not say; a 401 renews it sooner.
	sessionLifetime = 30 * time.Minute

	// maxInterstitialSegments is the longest run of segments a second clip may
	// have and still be read as a bumper rather than as part of the title.
	maxInterstitialSegments = 6
)

// ErrProtected reports a title whose playlist asks for a key system this
// package does not handle. It is content state, not a parser failure.
var ErrProtected = errors.New("pluto tv: this title is DRM-protected and cannot be played here")

var (
	// idRe is a catalog id: a 24-digit hex ObjectID.
	idRe = regexp.MustCompile(`^[0-9a-f]{24}$`)
	// pagePathRe reads the identifier URLs documented in doc.go.
	pagePathRe = regexp.MustCompile(`^/on-demand/(series|movies)/([0-9a-f]{24})(?:/episode/([0-9a-f]{24}))?/?$`)
	// slugYearRe reads the release year a movie slug carries
	// ("god-of-war-2017-1-1-ptv2").
	slugYearRe = regexp.MustCompile(`-((?:19|20)\d{2})-\d+-\d+(?:-[a-z0-9]+)?$`)
	// streamInfRe / resolutionRe read one #EXT-X-STREAM-INF line.
	bandwidthRe  = regexp.MustCompile(`[:,]BANDWIDTH=(\d+)`)
	resolutionRe = regexp.MustCompile(`RESOLUTION=\d+x(\d+)`)
	// tagURIRe matches the URI attribute of a playlist tag (a key file).
	tagURIRe = regexp.MustCompile(`URI="[^"]*"`)
	// keyMethodRe reads the method of an #EXT-X-KEY or #EXT-X-SESSION-KEY.
	keyMethodRe = regexp.MustCompile(`(?m)^#EXT-X-(?:SESSION-)?KEY:.*?METHOD=([A-Za-z0-9-]+)`)
	// contentSegmentRe matches a segment of a title's own clip and captures
	// the clip directory, the range and the rendition name:
	// …/hls/498333-941317/hls_600-00012.ts
	contentSegmentRe = regexp.MustCompile(`^(https?://[^\s?#]+/hls/)(\d+)-(\d+|end)/(hls_\d+)-\d+\.ts$`)
	// qualityDigitsRe pulls "1080" out of "1080p".
	qualityDigitsRe = regexp.MustCompile(`(\d{3,4})`)
)

// Kind is what a catalog entry is.
type Kind string

const (
	// KindMovie is a single film.
	KindMovie Kind = "movie"
	// KindSeries is a show with seasons.
	KindSeries Kind = "series"
)

// Title is one search hit.
type Title struct {
	ID       string
	Name     string
	Slug     string
	Language string // the catalog's two-letter code, "PT", "EN", …
	Kind     Kind
}

// Episode is one entry of a season.
type Episode struct {
	ID          string
	Name        string
	Description string
	Number      int
	Season      int
	Duration    time.Duration
}

// Season is a numbered season and its episodes, in order.
type Season struct {
	Number   int
	Episodes []Episode
}

// Show is a series and its seasons, in order.
type Show struct {
	ID      string
	Name    string
	Seasons []Season
}

// Stream is a resolved playlist.
type Stream struct {
	URL string
	// Playlist is the title alone as a media playlist whose every URI is
	// absolute, set whenever AdFree is. It is what should be played: the CDN
	// is several, and one of them refuses the playlist at URL while serving
	// the segments it lists, so a player handed URL fails or not by the DNS
	// answer it happens to get. Serve this text to the player instead.
	Playlist string
	// AdFree is false when the title's own playlist could not be found and
	// URL is the stitched one, with the service's breaks in it.
	AdFree bool
}

// Client talks to Pluto TV with an anonymous session.
type Client struct {
	http      *http.Client
	bootURL   string
	userAgent string
	clientID  string
	// allow names hosts (with port) trusted beyond Pluto TV's own, so a test
	// can point the services at a local server and still be refused elsewhere.
	allow map[string]bool

	mu   sync.Mutex
	sess *session
}

// session is one boot answer: the token and where to use it.
type session struct {
	token          string
	stitcherParams string
	search         string
	vod            string
	stitcher       string
	expires        time.Time
}

var (
	sharedOnce   sync.Once
	sharedClient *Client
)

// Shared returns the process-wide client, so one session serves the search,
// the listing and the stream.
func Shared() *Client {
	sharedOnce.Do(func() { sharedClient = NewClient() })
	return sharedClient
}

// NewClient builds a client without making network requests.
func NewClient() *Client {
	return &Client{
		http:      util.NewFastClient(),
		bootURL:   DefaultBootURL,
		userAgent: netx.UserAgent,
		clientID:  newClientID(),
	}
}

// NewClientForTest uses an injected HTTP client and boot origin, and trusts
// that origin's host in addition to Pluto TV's own. Only for tests.
func NewClientForTest(hc *http.Client, bootURL string) *Client {
	c := NewClient()
	c.http = hc
	c.bootURL = strings.TrimRight(bootURL, "/")
	if u, err := url.Parse(c.bootURL); err == nil && u.Host != "" {
		c.allow = map[string]bool{u.Host: true}
	}
	return c
}

// newClientID is a random UUIDv4: the device id an anonymous session is keyed
// by. It is made per process and never stored.
func newClientID() string {
	var b [16]byte
	_, _ = rand.Read(b[:]) // never fails since Go 1.24
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// ---------------------------------------------------------------------------
// Identifier URLs
// ---------------------------------------------------------------------------

// SeriesURL is the identifier URL of a series.
func SeriesURL(id string) string { return siteBase + "/on-demand/series/" + id }

// MovieURL is the identifier URL of a movie. It is also what plays it.
func MovieURL(id string) string { return siteBase + "/on-demand/movies/" + id }

// EpisodeURL is the identifier URL of one episode of a series.
func EpisodeURL(seriesID, episodeID string) string {
	return SeriesURL(seriesID) + "/episode/" + episodeID
}

// URL is the title's identifier URL.
func (t Title) URL() string {
	if t.Kind == KindMovie {
		return MovieURL(t.ID)
	}
	return SeriesURL(t.ID)
}

// Year is the release year a movie's slug carries, or "".
func (t Title) Year() string {
	if t.Kind != KindMovie {
		return ""
	}
	if m := slugYearRe.FindStringSubmatch(t.Slug); m != nil {
		return m[1]
	}
	return ""
}

// languageTag words the catalog's language code the way result names carry it.
func languageTag(code string) string {
	switch strings.ToUpper(strings.TrimSpace(code)) {
	case "PT":
		return "[PT-BR]"
	case "EN":
		return "[English]"
	default:
		return "[Multilanguage]"
	}
}

// ToAnimeModel converts a hit to the shared model. The name carries the media
// type and language tags the other movie/TV sources use.
func (t Title) ToAnimeModel() *models.Anime {
	mediaType, typeTag := models.MediaTypeTV, "[TV]"
	if t.Kind == KindMovie {
		mediaType, typeTag = models.MediaTypeMovie, "[Movie]"
	}
	return &models.Anime{
		Name:      fmt.Sprintf("%s %s %s", typeTag, languageTag(t.Language), t.Name),
		URL:       t.URL(),
		Source:    SourceName,
		MediaType: mediaType,
		Year:      t.Year(),
	}
}

// pageRef is a parsed identifier URL.
type pageRef struct {
	kind      Kind
	id        string // the series or movie id
	episodeID string // set on an episode URL
}

// playID is the id the stitcher plays: the episode's, or the movie's own.
func (r pageRef) playID() (string, bool) {
	if r.episodeID != "" {
		return r.episodeID, true
	}
	return r.id, r.kind == KindMovie
}

func parsePageURL(raw string) (pageRef, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.User != nil || u.Scheme != "https" || strings.TrimPrefix(strings.ToLower(u.Host), "www.") != "pluto.tv" {
		return pageRef{}, netx.NewParserError(SourceName, "url", "not a Pluto TV on-demand URL", err)
	}
	m := pagePathRe.FindStringSubmatch(u.Path)
	if m == nil || (m[1] == "movies" && m[3] != "") {
		// A movie has no episodes: its page and its player are the same id.
		return pageRef{}, netx.NewParserError(SourceName, "url", "not a Pluto TV on-demand URL", nil)
	}
	ref := pageRef{kind: KindSeries, id: m[2], episodeID: m[3]}
	if m[1] == "movies" {
		ref.kind = KindMovie
	}
	return ref, nil
}

// IsMovieURL reports whether raw is the identifier URL of a movie.
func IsMovieURL(raw string) bool {
	ref, err := parsePageURL(raw)
	return err == nil && ref.kind == KindMovie
}

// ---------------------------------------------------------------------------
// Session
// ---------------------------------------------------------------------------

// trusted reports whether raw is an https URL on one of Pluto TV's own hosts.
// The service hosts come from the boot answer and the media host from a
// playlist, so neither is followed blindly.
func (c *Client) trusted(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil {
		return false
	}
	if c.allow[u.Host] {
		return u.Scheme == "http" || u.Scheme == "https"
	}
	if u.Scheme != "https" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	for _, domain := range []string{"pluto.tv", "plutotv.net"} {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}
	return false
}

// currentSession returns the cached session, opening one when there is none
// or it has run out.
func (c *Client) currentSession(ctx context.Context) (*session, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sess != nil && time.Now().Before(c.sess.expires) {
		return c.sess, nil
	}
	s, err := c.boot(ctx)
	if err != nil {
		return nil, err
	}
	c.sess = s
	return s, nil
}

// dropSession forgets s, so the next call opens a new one.
func (c *Client) dropSession(s *session) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sess == s {
		c.sess = nil
	}
}

func (c *Client) boot(ctx context.Context) (*session, error) {
	q := url.Values{
		"appName":             {appName},
		"appVersion":          {appVersion},
		"deviceVersion":       {deviceVersion},
		"deviceModel":         {"web"},
		"deviceMake":          {"chrome"},
		"deviceType":          {"web"},
		"clientID":            {c.clientID},
		"clientModelNumber":   {"1.0.0"},
		"serverSideAds":       {"false"},
		"drmCapabilities":     {""},
		"blockingMode":        {""},
		"notificationVersion": {"1"},
		"appLaunchCount":      {"0"},
		"lastAppLaunchDate":   {""},
	}
	var answer struct {
		SessionToken   string `json:"sessionToken"`
		StitcherParams string `json:"stitcherParams"`
		RefreshInSec   int    `json:"refreshInSec"`
		Servers        struct {
			Search   string `json:"search"`
			VOD      string `json:"vod"`
			Stitcher string `json:"stitcher"`
		} `json:"servers"`
	}
	if _, err := c.getJSON(ctx, c.bootURL+"/v4/start?"+q.Encode(), "", "session", &answer); err != nil {
		return nil, err
	}
	s := &session{
		token:          strings.TrimSpace(answer.SessionToken),
		stitcherParams: strings.Trim(strings.TrimSpace(answer.StitcherParams), "&?"),
		search:         strings.TrimRight(answer.Servers.Search, "/"),
		vod:            strings.TrimRight(answer.Servers.VOD, "/"),
		stitcher:       strings.TrimRight(answer.Servers.Stitcher, "/"),
	}
	if s.token == "" {
		return nil, netx.NewParserError(SourceName, "session", "the session answer carries no token", nil)
	}
	for _, host := range []string{s.search, s.vod, s.stitcher} {
		if !c.trusted(host) {
			return nil, netx.NewParserError(SourceName, "session", "the session answer names an unexpected service host", nil)
		}
	}
	lifetime := sessionLifetime
	if refresh := time.Duration(answer.RefreshInSec) * time.Second; refresh > 0 && refresh < lifetime {
		lifetime = refresh
	}
	s.expires = time.Now().Add(lifetime)
	return s, nil
}

// withSession runs call with the current session and, when the service
// answers 401, once more with a new one.
func (c *Client) withSession(ctx context.Context, call func(*session) (int, error)) error {
	for attempt := 0; ; attempt++ {
		s, err := c.currentSession(ctx)
		if err != nil {
			return err
		}
		status, err := call(s)
		if status == http.StatusUnauthorized && attempt == 0 {
			c.dropSession(s)
			continue
		}
		return err
	}
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

// Search returns the movies and series that match query. Live channels and
// their schedules, which the same endpoint lists, are left out.
func (c *Client) Search(ctx context.Context, query string) ([]Title, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	var page struct {
		Data []struct {
			ID       string `json:"id"`
			Slug     string `json:"slug"`
			Name     string `json:"name"`
			Type     string `json:"type"`
			Language string `json:"language"`
		} `json:"data"`
	}
	err := c.withSession(ctx, func(s *session) (int, error) {
		q := url.Values{"q": {query}, "limit": {fmt.Sprint(searchLimit)}}
		return c.getJSON(ctx, s.search+"/v1/search?"+q.Encode(), s.token, "search", &page)
	})
	if err != nil {
		return nil, err
	}
	titles := make([]Title, 0, len(page.Data))
	seen := make(map[string]bool, len(page.Data))
	for _, hit := range page.Data {
		kind := Kind(hit.Type)
		name := strings.TrimSpace(hit.Name)
		if (kind != KindMovie && kind != KindSeries) || !idRe.MatchString(hit.ID) || name == "" || seen[hit.ID] {
			continue
		}
		seen[hit.ID] = true
		titles = append(titles, Title{ID: hit.ID, Name: name, Slug: hit.Slug, Language: hit.Language, Kind: kind})
	}
	return titles, nil
}

// ---------------------------------------------------------------------------
// Seasons and episodes
// ---------------------------------------------------------------------------

// Show lists the seasons and episodes of the series at seriesURL.
func (c *Client) Show(ctx context.Context, seriesURL string) (Show, error) {
	ref, err := parsePageURL(seriesURL)
	if err != nil {
		return Show{}, err
	}
	if ref.kind != KindSeries {
		return Show{}, netx.NewParserError(SourceName, "episodes", "a movie has no seasons", nil)
	}
	var answer struct {
		ID      string `json:"_id"`
		Name    string `json:"name"`
		Seasons []struct {
			Number   int `json:"number"`
			Episodes []struct {
				ID          string `json:"_id"`
				Name        string `json:"name"`
				Description string `json:"description"`
				Number      int    `json:"number"`
				Season      int    `json:"season"`
				Duration    int64  `json:"originalContentDuration"`
			} `json:"episodes"`
		} `json:"seasons"`
	}
	err = c.withSession(ctx, func(s *session) (int, error) {
		target := fmt.Sprintf("%s/v4/vod/series/%s/seasons?offset=1000&page=1", s.vod, ref.id)
		return c.getJSON(ctx, target, s.token, "episodes", &answer)
	})
	if err != nil {
		return Show{}, err
	}
	if answer.ID != ref.id {
		return Show{}, netx.NewParserError(SourceName, "episodes", "the answer is for a different series", nil)
	}
	show := Show{ID: answer.ID, Name: strings.TrimSpace(answer.Name)}
	seen := make(map[string]bool)
	for _, s := range answer.Seasons {
		season := Season{Number: s.Number}
		for _, e := range s.Episodes {
			if !idRe.MatchString(e.ID) || seen[e.ID] {
				continue
			}
			seen[e.ID] = true
			season.Episodes = append(season.Episodes, Episode{
				ID:          e.ID,
				Name:        strings.TrimSpace(e.Name),
				Description: strings.TrimSpace(e.Description),
				Number:      e.Number,
				Season:      s.Number,
				Duration:    time.Duration(e.Duration) * time.Millisecond,
			})
		}
		if len(season.Episodes) == 0 {
			continue
		}
		sort.SliceStable(season.Episodes, func(i, j int) bool {
			return season.Episodes[i].Number < season.Episodes[j].Number
		})
		// A season whose numbers are missing or repeated is numbered by its
		// order, so every episode can still be asked for by number.
		for i := range season.Episodes {
			if season.Episodes[i].Number <= 0 || (i > 0 && season.Episodes[i].Number <= season.Episodes[i-1].Number) {
				for j := range season.Episodes {
					season.Episodes[j].Number = j + 1
				}
				break
			}
		}
		show.Seasons = append(show.Seasons, season)
	}
	sort.SliceStable(show.Seasons, func(i, j int) bool { return show.Seasons[i].Number < show.Seasons[j].Number })
	if len(show.Seasons) == 0 {
		return Show{}, netx.NewParserError(SourceName, "episodes", "the series lists no episodes", nil)
	}
	return show, nil
}

// ---------------------------------------------------------------------------
// Stream
// ---------------------------------------------------------------------------

// variant is one rendition of the stitched master playlist.
type variant struct {
	URL       string
	Bandwidth int
	Height    int // 0 when the master does not say
}

// Stream resolves the movie or episode at pageURL to a playlist. quality is
// "best", "worst" or a height such as "720p".
func (c *Client) Stream(ctx context.Context, pageURL, quality string) (Stream, error) {
	ref, err := parsePageURL(pageURL)
	if err != nil {
		return Stream{}, err
	}
	id, ok := ref.playID()
	if !ok {
		return Stream{}, netx.NewParserError(SourceName, "stream", "a series URL names no episode to play", nil)
	}

	var masterURL, master string
	err = c.withSession(ctx, func(s *session) (int, error) {
		masterURL = fmt.Sprintf("%s/v2/stitch/hls/episode/%s/master.m3u8?%s&jwt=%s&masterJWTPassthrough=true",
			s.stitcher, id, s.stitcherParams, url.QueryEscape(s.token))
		var status int
		var fetchErr error
		master, status, fetchErr = c.getText(ctx, masterURL, "stream")
		return status, fetchErr
	})
	if err != nil {
		return Stream{}, err
	}
	if isProtected(master) {
		return Stream{}, ErrProtected
	}
	variants := parseMaster(master, masterURL)
	if len(variants) == 0 {
		return Stream{}, netx.NewParserError(SourceName, "stream", "the master playlist lists no renditions", nil)
	}

	// The closest rendition first; the others are tried only when its own
	// playlist cannot be fetched.
	for i, v := range orderVariants(variants, quality) {
		if i == 3 {
			break
		}
		if !c.trusted(v.URL) {
			continue
		}
		playlist, _, err := c.getText(ctx, v.URL, "stream")
		if err != nil {
			if ctx.Err() != nil {
				return Stream{}, ctx.Err()
			}
			util.Debug("Pluto TV rendition playlist unavailable", "bandwidth", v.Bandwidth, "error", err)
			continue
		}
		if isProtected(playlist) {
			return Stream{}, ErrProtected
		}
		contentURL, ok := contentPlaylistURL(playlist)
		if !ok || !c.trusted(contentURL) {
			// The layout is the title's, not the rendition's: another
			// rendition would read the same.
			util.Debug("Pluto TV title does not follow the single-clip layout; playing the stitched playlist")
			break
		}
		body, err := c.checkMediaPlaylist(ctx, contentURL)
		if err == nil {
			return Stream{URL: contentURL, Playlist: absolutePlaylist(body, contentURL), AdFree: true}, nil
		}
		if ctx.Err() != nil {
			return Stream{}, ctx.Err()
		}
		if errors.Is(err, ErrProtected) {
			// The title itself is keyed; another rung of it is too.
			return Stream{}, err
		}
		// One of the CDNs answers 403 for every playlist under the clip and
		// still serves its segments. The stitched playlist lists the same
		// title in ranges, so the title alone is read out of it instead.
		util.Debug("Pluto TV content playlist unavailable", "bandwidth", v.Bandwidth, "error", err)
		if own, ok := titlePlaylist(playlist, v.URL, contentURL); ok {
			return Stream{URL: masterURL, Playlist: own, AdFree: true}, nil
		}
	}
	return Stream{URL: masterURL}, nil
}

// absolutePlaylist rewrites every URI of a media playlist (segments and key
// files) against base, so the text plays from wherever it is served.
func absolutePlaylist(playlist, base string) string {
	origin, err := url.Parse(base)
	if err != nil {
		return playlist
	}
	resolve := func(ref string) string {
		u, err := url.Parse(strings.TrimSpace(ref))
		if err != nil {
			return ref
		}
		return origin.ResolveReference(u).String()
	}
	lines := strings.Split(strings.ReplaceAll(playlist, "\r\n", "\n"), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
		case strings.HasPrefix(trimmed, "#"):
			lines[i] = tagURIRe.ReplaceAllStringFunc(trimmed, func(attr string) string {
				return `URI="` + resolve(attr[len(`URI="`):len(attr)-1]) + `"`
			})
		default:
			lines[i] = resolve(trimmed)
		}
	}
	return strings.Join(lines, "\n")
}

// titlePlaylist builds the title's own playlist out of the stitched one: the
// segments of the clip contentURL names, in order, each with the key and the
// duration the stitcher gave it, and nothing of what was spliced between them.
// The ranges of a clip are cut from one encode, so their timestamps run on.
//
// It answers false when a segment cannot be carried over exactly: one without
// a duration, or one whose key leaves the IV to the sequence number, which is
// not the number it has here.
func titlePlaylist(stitched, base, contentURL string) (string, bool) {
	cut := strings.LastIndex(contentURL, "0-end/")
	if cut < 0 {
		return "", false
	}
	dir, rendition := contentURL[:cut], strings.TrimSuffix(contentURL[cut+len("0-end/"):], ".m3u8")

	var body strings.Builder
	var key, duration, lastRange string
	segments, longest := 0, 0.0
	for _, line := range strings.Split(absolutePlaylist(stitched, base), "\n") {
		switch {
		case strings.HasPrefix(line, "#EXT-X-KEY:"):
			key = line
		case strings.HasPrefix(line, "#EXTINF:"):
			duration = line
		case line == "" || strings.HasPrefix(line, "#"):
		default:
			m := contentSegmentRe.FindStringSubmatch(line)
			seconds, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimPrefix(duration, "#EXTINF:"), ","), 64)
			segmentDuration := duration
			duration = "" // a duration belongs to the one segment after it
			if m == nil || m[1] != dir || m[4] != rendition {
				continue
			}
			if err != nil || seconds <= 0 {
				return "", false
			}
			if key != "" && !strings.Contains(key, "METHOD=NONE") && !strings.Contains(key, "IV=") {
				return "", false
			}
			if r := m[2] + "-" + m[3]; r != lastRange {
				if lastRange != "" {
					body.WriteString("#EXT-X-DISCONTINUITY\n")
				}
				lastRange = r
			}
			if key != "" {
				body.WriteString(key + "\n")
			}
			body.WriteString(segmentDuration + "\n" + line + "\n")
			segments++
			longest = max(longest, seconds)
		}
	}
	if segments == 0 {
		return "", false
	}
	header := fmt.Sprintf("#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-PLAYLIST-TYPE:VOD\n#EXT-X-TARGETDURATION:%d\n#EXT-X-MEDIA-SEQUENCE:0\n", int(math.Ceil(longest)))
	return header + body.String() + "#EXT-X-ENDLIST\n", true
}

// isProtected reports a playlist keyed by anything other than plain AES-128.
func isProtected(playlist string) bool {
	for _, m := range keyMethodRe.FindAllStringSubmatch(playlist, -1) {
		if method := strings.ToUpper(m[1]); method != "NONE" && method != "AES-128" {
			return true
		}
	}
	return false
}

// parseMaster reads the renditions of a master playlist, resolving their URIs
// against the playlist's own URL.
func parseMaster(master, masterURL string) []variant {
	base, err := url.Parse(masterURL)
	if err != nil {
		return nil
	}
	var variants []variant
	var pending *variant
	for _, line := range strings.Split(master, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "#EXT-X-STREAM-INF:"):
			pending = &variant{}
			if m := bandwidthRe.FindStringSubmatch(line); m != nil {
				_, _ = fmt.Sscan(m[1], &pending.Bandwidth)
			}
			if m := resolutionRe.FindStringSubmatch(line); m != nil {
				_, _ = fmt.Sscan(m[1], &pending.Height)
			}
		case line == "" || strings.HasPrefix(line, "#"):
		case pending != nil:
			if ref, err := url.Parse(line); err == nil {
				pending.URL = base.ResolveReference(ref).String()
				variants = append(variants, *pending)
			}
			pending = nil
		}
	}
	return variants
}

// bandwidthBudget is the highest bitrate read as fitting a height, for masters
// that name no resolution. Pluto TV's ladder (checked 2026-11-16) is 240p at
// 0.6 Mbps, 360p at 1.0, 480p at 1.6, 576p at 2.3 and 720p at 3.3.
func bandwidthBudget(height int) int {
	switch {
	case height < 360:
		return 800_000
	case height < 480:
		return 1_300_000
	case height < 576:
		return 2_000_000
	case height < 720:
		return 2_800_000
	default:
		return int(^uint(0) >> 1)
	}
}

// orderVariants sorts the renditions by how well they answer quality: the best
// for "best" (and anything unreadable), the smallest for "worst", and for a
// height the largest one that fits, then the nearest above it.
func orderVariants(variants []variant, quality string) []variant {
	ordered := append([]variant(nil), variants...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Bandwidth > ordered[j].Bandwidth })

	quality = strings.ToLower(strings.TrimSpace(quality))
	if quality == "worst" {
		sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Bandwidth < ordered[j].Bandwidth })
		return ordered
	}
	m := qualityDigitsRe.FindStringSubmatch(quality)
	if m == nil {
		return ordered
	}
	var height int
	_, _ = fmt.Sscan(m[1], &height)
	fits := func(v variant) bool {
		if v.Height > 0 {
			return v.Height <= height
		}
		return v.Bandwidth <= bandwidthBudget(height)
	}
	var within, above []variant
	for _, v := range ordered {
		if fits(v) {
			within = append(within, v)
		} else {
			above = append([]variant{v}, above...)
		}
	}
	return append(within, above...)
}

// contentPlaylistURL reads a stitched rendition playlist and returns the URL
// of the title's own uninterrupted playlist (see doc.go).
//
// It answers false unless the playlist is one clip from its first range to its
// last: a title made of several clips has no single playlist to point at.
func contentPlaylistURL(playlist string) (string, bool) {
	type clip struct {
		dir, rendition      string
		segments            int
		fromStart, toTheEnd bool
	}
	var clips []*clip
	byKey := make(map[string]*clip)
	for _, line := range strings.Split(playlist, "\n") {
		m := contentSegmentRe.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		key := m[1] + "\x00" + m[4]
		c := byKey[key]
		if c == nil {
			c = &clip{dir: m[1], rendition: m[4]}
			byKey[key] = c
			clips = append(clips, c)
		}
		c.segments++
		c.fromStart = c.fromStart || m[2] == "0"
		c.toTheEnd = c.toTheEnd || m[3] == "end"
	}
	var title *clip
	for _, c := range clips {
		if title == nil || c.segments > title.segments {
			title = c
		}
	}
	if title == nil || !title.fromStart || !title.toTheEnd {
		return "", false
	}
	for _, c := range clips {
		if c != title && c.segments > maxInterstitialSegments {
			return "", false
		}
	}
	return title.dir + "0-end/" + title.rendition + ".m3u8", true
}

// checkMediaPlaylist confirms target answers with a media playlist.
func (c *Client) checkMediaPlaylist(ctx context.Context, target string) (string, error) {
	body, _, err := c.getText(ctx, target, "stream")
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(strings.TrimSpace(body), "#EXTM3U") || !strings.Contains(body, "#EXTINF") {
		return "", netx.NewParserError(SourceName, "stream", "the content URL is not a media playlist", nil)
	}
	if isProtected(body) {
		return "", ErrProtected
	}
	return body, nil
}

// ---------------------------------------------------------------------------
// HTTP plumbing
// ---------------------------------------------------------------------------

func (c *Client) get(ctx context.Context, target, token, accept, layer string, limit int64) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, http.NoBody)
	if err != nil {
		return nil, 0, netx.NewParserError(SourceName, layer, "bad request URL", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", accept)
	req.Header.Set("Accept-Language", netx.AcceptLanguage)
	req.Header.Set("Origin", siteBase)
	req.Header.Set("Referer", siteBase+"/")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.http.Do(req) // #nosec G704 -- hosts are checked by trusted()
	if err != nil {
		return nil, 0, netx.NewParserError(SourceName, layer, "request failed", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, netx.NewHTTPStatusError(SourceName, layer, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, resp.StatusCode, netx.NewParserError(SourceName, layer, "failed to read response", err)
	}
	if int64(len(body)) > limit {
		return nil, resp.StatusCode, netx.NewParserError(SourceName, layer, "response exceeds size limit", nil)
	}
	return body, resp.StatusCode, nil
}

func (c *Client) getJSON(ctx context.Context, target, token, layer string, dst any) (int, error) {
	body, status, err := c.get(ctx, target, token, "application/json", layer, maxJSONBytes)
	if err != nil {
		return status, err
	}
	if err := jsonx.Unmarshal(body, dst); err != nil {
		return status, netx.NewParserError(SourceName, layer, "invalid API JSON", err)
	}
	return status, nil
}

func (c *Client) getText(ctx context.Context, target, layer string) (string, int, error) {
	body, status, err := c.get(ctx, target, "", "application/x-mpegURL, */*", layer, maxPlaylistBytes)
	return string(body), status, err
}
