package superflix

import (
	"net/http"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
)

func TestSameSiteFromPlaywright(t *testing.T) {
	t.Parallel()
	strict := playwright.SameSiteAttribute("Strict")
	lax := playwright.SameSiteAttribute("Lax")
	none := playwright.SameSiteAttribute("None")
	bogus := playwright.SameSiteAttribute("Whatever")

	assert.Equal(t, http.SameSiteDefaultMode, sameSiteFromPlaywright(nil), "nil → default")
	assert.Equal(t, http.SameSiteStrictMode, sameSiteFromPlaywright(&strict))
	assert.Equal(t, http.SameSiteLaxMode, sameSiteFromPlaywright(&lax))
	assert.Equal(t, http.SameSiteNoneMode, sameSiteFromPlaywright(&none))
	assert.Equal(t, http.SameSiteDefaultMode, sameSiteFromPlaywright(&bogus), "unknown → default")
}

func TestEmbedHostParentURL(t *testing.T) {
	t.Parallel()
	// The fallback is liveBase(), NOT the compiled constant. Those are the same
	// value only while runtime discovery has not resolved a rotation, so
	// spelling the constant here made this test fail the day the domain moved
	// (.beer → .baby) even though discovery was doing its job and the app kept
	// working. Asserting against liveBase() pins the actual contract — "fall
	// back to whatever host we would target" — and survives the next rotation.
	fallback := liveBase() + "/"
	tests := map[string]string{
		"https://superflixapi.pro/filme/1": "https://superflixapi.pro/",
		"https://sub.host.com/serie/1/2/3": "https://sub.host.com/",
		"http://other.tld/x":               "http://other.tld/",
		"":                                 fallback, // no host → fallback
		"not a url with spaces":            fallback, // unparseable host → fallback
	}
	for in, want := range tests {
		assert.Equalf(t, want, embedHostParentURL(in), "embedHostParentURL(%q)", in)
	}
}

func TestIsRealPlayerHTML(t *testing.T) {
	t.Parallel()
	assert.True(t, isRealPlayerHTML(`<script>var ALL_EPISODES = {};</script>`), "ALL_EPISODES is real")
	assert.True(t, isRealPlayerHTML(`<a data-episode-id="9">ep</a>`), "frontend anchors are real")
	assert.False(t, isRealPlayerHTML(`<h1>Acesso Restrito</h1>`), "restricted shell is not real")
	assert.False(t, isRealPlayerHTML(""), "empty is not real")
}
