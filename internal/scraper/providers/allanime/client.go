// Package scraper provides web scraping functionality for anime sources
package allanime

import (
	"bytes"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

const (
	// AllAnimeReferer is the Referer/Origin AllAnime binds its API to. Rotated
	// 2026-07-22 to mkissa.to (ani-cli PR #1779); requests carrying the old
	// youtu-chan.com referer now receive a stripped response.
	AllAnimeReferer = "https://mkissa.to"
	// AllAnimeBase is the host that internal ("--"-encoded) source URLs resolve
	// to (e.g. the /clock.json embeds). Unchanged by the mkissa migration.
	AllAnimeBase = "allanime.day"
	// AllAnimeAPI is the GraphQL endpoint. Moved off api.allanime.day to the
	// mkissa mirror 2026-07-22 (ani-cli PR #1779).
	AllAnimeAPI = "https://api.mkissa.net/api"

	// allAnimePersistedQueryHash is the Apollo persistedQuery sha256 for the
	// `episode { sourceUrls / tobeparsed }` query, and doubles as the `qh`
	// field bound into the aaReq token. Rotated 2026-07-22 (ani-cli PR #1779).
	allAnimePersistedQueryHash = "f4662f4b7510b26795dd53ef824a0bf1740fbbc5d1273fab18222ac831bca8d0"

	// allAnimePersistedQueryOrigin is the Origin the GET path must send.
	// AllAnime returns a stripped (no `tobeparsed`) response for any other Origin.
	allAnimePersistedQueryOrigin = "https://mkissa.to"
)

// Pre-compiled regexes for AllAnime scraper (avoid per-call compilation)
var (
	allAnimeSourceURLFallbackRe = regexp.MustCompile(`"sourceUrl":"--([^"]*)".*?"sourceName":"([^"]*)"`)
	allAnimeVideoLinkRe         = regexp.MustCompile(`"link":"([^"]*)".*?"resolutionStr":"([^"]*)"`)
	allAnimeM3U8Re              = regexp.MustCompile(`"hls":true.*?"link":"([^"]*)"`)
)

// AllAnimeClient handles interactions with AllAnime API
type AllAnimeClient struct {
	client    *http.Client
	referer   string
	apiBase   string
	userAgent string

	// keyMu guards the cached per-epoch AES material (keys/keysExp). The key is
	// scraped from the mkissa CDN bundle (fetchAAKeys) and reused until keysExp.
	keyMu   sync.Mutex
	keys    *aaKeys
	keysExp time.Time
}

// allAnimeClientInstance is a singleton for connection reuse
var (
	allAnimeClientInstance *AllAnimeClient
	allAnimeClientOnce     sync.Once
)

// NewAllAnimeClient creates a new AllAnime client (returns cached instance for connection reuse)
func NewAllAnimeClient() *AllAnimeClient {
	allAnimeClientOnce.Do(func() {
		allAnimeClientInstance = &AllAnimeClient{
			client:    util.NewFastClient(), // Own client to avoid http2 transport race
			referer:   AllAnimeReferer,
			apiBase:   AllAnimeAPI,
			userAgent: netx.UserAgent,
		}
	})
	return allAnimeClientInstance
}

// NewClientForTest returns a client whose API base points at a test server.
// Only for tests. A fixed fixture key is injected so the client never scrapes
// the live mkissa key bundle over the network during tests.
func NewClientForTest(serverURL string) *AllAnimeClient {
	return &AllAnimeClient{
		client:    util.GetFastClient(),
		referer:   AllAnimeReferer,
		apiBase:   serverURL,
		userAgent: netx.UserAgent,
		keys:      &aaKeys{key: bytes.Repeat([]byte{0x2a}, 32), epoch: "0"},
		keysExp:   time.Now().Add(time.Hour),
	}
}
