package metadata

import (
	"net/http"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setAniListHeaders must present a NON-browser User-Agent: AniList answers
// browser UAs with a 403 ("The AniList API has been temporarily disabled due to
// severe stability issues") and serves plain API clients normally (issue #184).
func TestSetAniListHeaders_UsesNonBrowserUserAgent(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest("POST", "https://graphql.anilist.co", nil)
	require.NoError(t, err)

	setAniListHeaders(req)

	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Equal(t, "application/json", req.Header.Get("Accept"))
	assert.Equal(t, netx.APIUserAgent, req.Header.Get("User-Agent"))

	ua := req.Header.Get("User-Agent")
	for _, marker := range []string{"Mozilla", "AppleWebKit", "Chrome", "Firefox", "Safari", "Gecko"} {
		assert.NotContains(t, ua, marker,
			"a browser-shaped UA gets AniList to 403 us — %q must not appear", marker)
	}
}

// The AniList calls must not ride the shared surf client, which would overwrite
// the User-Agent with Chrome's and get us blocked regardless of the header above.
func TestNewEnricher_UsesPlainClientForAniList(t *testing.T) {
	t.Parallel()

	e := NewEnricher()
	require.NotNil(t, e.aniListClient)

	plain, ok := e.aniListClient.(*http.Client)
	require.True(t, ok, "AniList must use a plain *http.Client, not the impersonating shared client")
	assert.NotNil(t, plain)
	assert.NotSame(t, e.client, e.aniListClient,
		"AniList must not share the Chrome-impersonating client")
}

// Tests that inject a client must still control AniList traffic.
func TestNewEnricherWithClient_InjectedClientAlsoServesAniList(t *testing.T) {
	t.Parallel()

	stub := &stubClient{}
	e := NewEnricherWithClient(stub)
	assert.Same(t, stub, e.client)
	assert.Same(t, stub, e.aniListClient,
		"an injected client must serve AniList too, or tests would hit the real API")
}

type stubClient struct{ calls int }

func (s *stubClient) Do(_ *http.Request) (*http.Response, error) {
	s.calls++
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody, Header: make(http.Header)}, nil
}
