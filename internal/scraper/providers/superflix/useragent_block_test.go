package superflix

import (
	"net/http"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This host throttles by User-Agent string, and we were sending one it blocks.
//
// Measured 2026-09-24 against /pesquisar, one IP, requests 25s apart, identical
// Accept-Language throughout:
//
//	Firefox/125.0 (X11; Linux x86_64)   429, 429, 429, 429
//	Firefox/125.0 (Windows NT 10.0)     200
//	Firefox/121.0 (Windows NT 10.0)     200, 200
//	Chrome/124 (X11; Linux x86_64)      200
//
// The symptom was SuperFlix returning nothing from every search while curl on
// the same URL got 200 and three results — reported simply as "no SuperFlix
// results". Nothing about it looked like a User-Agent problem from the inside:
// the client saw a 429 and reported a rate limit, which was true and useless.

// blockedUserAgent is the exact string measured at 429. It stays here so the
// value that cost a day cannot quietly come back.
const blockedUserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:125.0) Gecko/20100101 Firefox/125.0"

func TestUserAgentIsNotTheBlockedString(t *testing.T) {
	t.Parallel()
	assert.NotEqual(t, blockedUserAgent, SuperFlixUserAgent,
		"this User-Agent was measured at 429 four times out of four; SuperFlix returns nothing with it")
}

// The declared UA has to describe the browser the solver actually drives.
//
// Cloudflare binds cf_clearance to the UA that solved the challenge, so a
// mismatch is re-challenged in a loop. The old value said Firefox and the
// solver has driven Chromium for a long time — the comment explaining the
// Firefox choice was describing a browser that had been replaced.
func TestUserAgentDescribesTheBrowserWeDrive(t *testing.T) {
	t.Parallel()
	assert.Contains(t, SuperFlixUserAgent, "Chrome/",
		"the solver drives Chromium (bundled or the system chrome channel); the UA must say so")
	assert.NotContains(t, SuperFlixUserAgent, "Firefox",
		"claiming Firefox while driving Chromium is what left the blocked string in place unnoticed")
}

// Accept-Language must describe the same browser as the UA — the rule
// netx.AcceptLanguage exists to enforce. When the UA moved to Chrome the ladder
// had to move with it.
func TestAcceptLanguageStillPairsWithTheUserAgent(t *testing.T) {
	t.Parallel()
	require.Contains(t, SuperFlixUserAgent, "Chrome/")
	assert.Equal(t, netx.ChromeAcceptLanguage, superFlixAcceptLanguage,
		"a Chrome UA with Firefox's q-ladder is a pair no real browser produces")
}

// The override is the escape hatch. This host has blocked a User-Agent once; the
// person it happens to next must not have to wait for a release.
func TestUserAgentOverride(t *testing.T) {
	t.Run("unset keeps the compiled value", func(t *testing.T) {
		assert.Equal(t, SuperFlixUserAgent, resolveUserAgent(SuperFlixUserAgent))
	})
	t.Run("set wins", func(t *testing.T) {
		t.Setenv(userAgentOverrideEnv, "Mozilla/5.0 (Custom) Whatever/1.0")
		assert.Equal(t, "Mozilla/5.0 (Custom) Whatever/1.0", resolveUserAgent(SuperFlixUserAgent))
	})
	t.Run("blank is ignored", func(t *testing.T) {
		t.Setenv(userAgentOverrideEnv, "   ")
		assert.Equal(t, SuperFlixUserAgent, resolveUserAgent(SuperFlixUserAgent))
	})
}

// And the override has to reach the wire, not just the resolver.
func TestDecorateRequestSendsTheResolvedUserAgent(t *testing.T) {
	t.Setenv(userAgentOverrideEnv, "Mozilla/5.0 (Override) Test/1.0")

	c := NewSuperFlixClient()
	req, err := http.NewRequest(http.MethodGet, "https://example.invalid/pesquisar", http.NoBody)
	require.NoError(t, err)
	c.decorateRequest(req)

	assert.Equal(t, "Mozilla/5.0 (Override) Test/1.0", req.Header.Get("User-Agent"))
	assert.Equal(t, superFlixAcceptLanguage, req.Header.Get("Accept-Language"))
	assert.True(t, strings.HasPrefix(req.Header.Get("Accept"), "text/html"))
}
