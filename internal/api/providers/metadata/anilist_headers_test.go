package metadata

import (
	"context"
	"io"
	"net/http"
	"strings"
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

	req, err := http.NewRequest("POST", "https://graphql.anilist.co", http.NoBody)
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

// capturingClient records the requests that actually leave the Enricher.
type capturingClient struct {
	reqs []*http.Request
	body string
}

func (c *capturingClient) Do(req *http.Request) (*http.Response, error) {
	c.reqs = append(c.reqs, req)
	body := c.body
	if body == "" {
		body = `{"data":{"Media":{"id":1,"title":{"romaji":"X"}}}}`
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil
}

// The header helper being right is not enough — what matters is what the REAL
// AniList code paths put on the wire. If someone routes them back through the
// Chrome-impersonating shared client, the UA becomes a browser one and AniList
// starts 403ing again (issue #184). These drive the actual methods.
func TestEnricher_AniListRequestsCarryNonBrowserUA(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func(*Enricher) error
	}{
		{
			name: "EnrichFromAniList",
			call: func(e *Enricher) error {
				_, err := e.EnrichFromAniList(context.Background(), "naruto")
				return err
			},
		},
		{
			name: "EnrichFromAniListByID",
			call: func(e *Enricher) error {
				_, err := e.EnrichFromAniListByID(context.Background(), 20)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cc := &capturingClient{}
			e := NewEnricherWithClient(cc)

			_ = tt.call(e) // the response shape is irrelevant; the request is the subject

			require.NotEmpty(t, cc.reqs, "the method must have issued an AniList request")
			var sawAniList bool
			for _, req := range cc.reqs {
				if !strings.Contains(req.URL.Host, "anilist") {
					continue
				}
				sawAniList = true
				ua := req.Header.Get("User-Agent")
				assert.Equal(t, netx.APIUserAgent, ua)
				for _, marker := range []string{"Mozilla", "AppleWebKit", "Chrome", "Firefox", "Safari", "Gecko"} {
					assert.NotContains(t, ua, marker,
						"AniList 403s browser-shaped User-Agents — %q must not reach the wire", marker)
				}
			}
			assert.True(t, sawAniList, "expected a request to graphql.anilist.co")
		})
	}
}
