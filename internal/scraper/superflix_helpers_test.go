package scraper

import (
	"net/http"
	"testing"

	"github.com/playwright-community/playwright-go"
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
	tests := map[string]string{
		"https://warezcdn.lat/filme/1":     "https://warezcdn.lat/",
		"https://sub.host.com/serie/1/2/3": "https://sub.host.com/",
		"http://other.tld/x":               "http://other.tld/",
		"":                                 "https://" + SuperFlixEmbedHost + "/", // no host → fallback
		"not a url with spaces":            "https://" + SuperFlixEmbedHost + "/", // unparseable host → fallback
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
