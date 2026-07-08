// Package scraper provides web scraping functionality for anime sources
package allanime

import (
	"net/http"
	"regexp"
	"sync"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

const (
	// AllAnimeReferer is the Referer/Origin AllAnime binds its API to. Rotated
	// 2026-07-08 to youtu-chan.com (ani-cli PR #1772); requests carrying the old
	// allmanga.to referer now receive a stripped response.
	AllAnimeReferer = "https://youtu-chan.com"
	AllAnimeBase    = "allanime.day"
	AllAnimeAPI     = "https://api.allanime.day/api"

	// allAnimeKeyHex is the AES-256 key (32 bytes, hex) used for BOTH the
	// aaReq request-signing token and the `tobeparsed` response decryption.
	// Rotated 2026-07-08 (ani-cli PR #1772); the old SHA-256("Xot36i3lK3:v1")
	// derivation no longer matches, so the key is now carried literally.
	allAnimeKeyHex = "22196fa6afca95309fdabe9a3534b87cd2454e50efeabfcbdbdfd3de678b3982"

	// allAnimePersistedQueryHash is the Apollo persistedQuery sha256 for the
	// `episode { sourceUrls / tobeparsed }` query, and doubles as the `qh`
	// field bound into the aaReq token.
	allAnimePersistedQueryHash = "d405d0edd690624b66baba3068e0edc3ac90f1597d898a1ec8db4e5c43c00fec"

	// allAnimePersistedQueryOrigin is the Origin the GET path must send.
	// AllAnime returns a stripped (no `tobeparsed`) response for any other Origin.
	allAnimePersistedQueryOrigin = "https://youtu-chan.com"

	// allAnimeAAReqEpoch and allAnimeAAReqBuildID are fixed protocol constants
	// bound into the aaReq token's payload and IV derivation (ani-cli PR #1772).
	allAnimeAAReqEpoch   = "4128"
	allAnimeAAReqBuildID = "9"
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
// Only for tests.
func NewClientForTest(serverURL string) *AllAnimeClient {
	return &AllAnimeClient{
		client:    util.GetFastClient(),
		referer:   AllAnimeReferer,
		apiBase:   serverURL,
		userAgent: netx.UserAgent,
	}
}
