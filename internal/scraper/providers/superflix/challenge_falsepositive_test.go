package superflix

import (
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A page that CAN show a Turnstile is not a page that IS one.
//
// SuperFlix added a login modal backed by Turnstile, so from 2026-09-24 the
// widget's script tag shipped on every page it serves. cfChallengeMarkers
// listed that URL, so every SuperFlix response — including a search results
// page with three matches on it — was classified as a captcha block. The source
// then failed its search, tripped its circuit breaker and vanished from the
// fan-out, while plain curl against the same URL got 200 and the results.
//
// The user saw it as "no SuperFlix results". Nothing was rate limiting them and
// nothing was gated; we were reading a good page and calling it a wall.
//
// This is the second marker on that list to be too broad in exactly this way
// (see the /cdn-cgi/challenge-platform note in cf.go), which is why the fixture
// is the real captured page rather than a hand-written snippet: a snippet only
// proves what its author already believed.

func liveSearchPage(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/search_results_2026_09_24.html")
	require.NoError(t, err)
	return b
}

func TestRealSearchPageIsNotAChallenge(t *testing.T) {
	t.Parallel()
	body := liveSearchPage(t)

	require.Contains(t, string(body), "challenges.cloudflare.com/turnstile",
		"the fixture must still carry the script tag, or it is not testing the false positive")

	assert.False(t, bodyHasChallengeMarker(body),
		"a search results page was classified as a Cloudflare challenge; "+
			"SuperFlix drops out of every search when this is wrong")
}

// And the page it was hiding really does parse. Without this the fix could be
// "stop calling it a challenge" while the results were unreadable anyway.
func TestRealSearchPageStillParses(t *testing.T) {
	t.Parallel()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(liveSearchPage(t))))
	require.NoError(t, err)

	results := (&SuperFlixClient{}).parseCards(doc)
	require.NotEmpty(t, results, "the page the false positive was hiding parses to nothing")

	first := results[0]
	assert.Equal(t, "Cowboy Bebop: O Filme", first.Title)
	assert.Equal(t, "11299", first.TMDBID, "without the TMDB id the result cannot be played")
	assert.Equal(t, "filme", first.SFType)
	assert.Equal(t, "2001", first.Year)
}

// The gates must still be caught. A fix that widened the hole would be worse
// than the false positive: an unrecognised gate is handed to the parser, which
// reads the interstitial as content.
func TestRealChallengesAreStillDetected(t *testing.T) {
	t.Parallel()
	gates := map[string]string{
		"cloudflare managed challenge": `<html><head><title>Just a moment...</title></head>` +
			`<body><div id="cf-chl-widget"></div><script>window._cf_chl_opt={cvId:"3"}</script></body></html>`,
		"cloudflare js challenge": `<html><body><form id="challenge-form" action="/x?__cf_chl_f_tk=abc"></form></body></html>`,
		"superflix's own gate": `<html><head><title>Verificação</title></head>` +
			`<body><form id="cf-turnstile-form"><div class="cf-turnstile"></div></form>` +
			`<script src="https://challenges.cloudflare.com/turnstile/v0/api.js"></script></body></html>`,
		"challenge platform orchestrate": `<html><body>` +
			`<script src="/cdn-cgi/challenge-platform/h/b/orchestrate/chl_page/v1"></script></body></html>`,
		"checking your browser": `<html><body>Checking your browser before accessing the site.</body></html>`,
	}
	for name, body := range gates {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Truef(t, bodyHasChallengeMarker([]byte(body)),
				"a real gate went undetected; the parser would read the interstitial as content")
		})
	}
}

// The passive telemetry snippet Cloudflare injects into ORDINARY responses must
// stay a non-signal. It is the first lesson this list learned and the regression
// that would undo it is the same shape as the one above.
func TestPassiveCloudflareSnippetsAreNotChallenges(t *testing.T) {
	t.Parallel()
	benign := map[string]string{
		"bot telemetry on a normal page": `<html><body>content` +
			`<script src="/cdn-cgi/challenge-platform/scripts/jsd/main.js"></script></body></html>`,
		"a login modal's turnstile mount": `<html><body><div data-api-auth-turnstile="login"></div>` +
			`<script src="https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit"></script>` +
			`<div class="group/card"><img alt="Some Movie"></div></body></html>`,
	}
	for name, body := range benign {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Falsef(t, bodyHasChallengeMarker([]byte(body)),
				"an ordinary page was called a challenge; the source disappears from search when this is wrong")
		})
	}
}
