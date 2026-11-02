package pobreflix

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestMirrorOfAdoptsOnlyTheSameSite(t *testing.T) {
	cases := []struct {
		name, base, final string
		want              string
		ok                bool
	}{
		{"mirror domain", "https://pobreflixtv.gay", "https://pobreflixtv.christmas/filme/x", "https://pobreflixtv.christmas", true},
		{"same host", "https://pobreflixtv.gay", "https://pobreflixtv.gay/filme/x", "", false},
		{"another site", "https://pobreflixtv.gay", "https://evil.example/filme/x", "", false},
		{"look-alike prefix", "https://pobreflixtv.gay", "https://pobreflixtv-evil.example/", "", false},
		{"suffix trick", "https://pobreflixtv.gay", "https://evil.example/pobreflixtv.christmas", "", false},
		{"downgrade to http", "https://pobreflixtv.gay", "http://pobreflixtv.christmas/", "", false},
		{"www is another label", "https://pobreflixtv.gay", "https://www.pobreflixtv.christmas/", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(tc.final)
			if err != nil {
				t.Fatal(err)
			}
			got, ok := mirrorOf(tc.base, u)
			if ok != tc.ok || got != tc.want {
				t.Fatalf("mirrorOf(%q, %q) = %q, %v; want %q, %v", tc.base, tc.final, got, ok, tc.want, tc.ok)
			}
		})
	}
}

// A catalog request that the site redirects to its new mirror moves the client
// there, so later requests go straight to the live host.
func TestClientFollowsAMirrorRedirect(t *testing.T) {
	var movedHits int
	moved := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		movedHits++
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(moved.Close)
	old := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, moved.URL+r.URL.Path, http.StatusMovedPermanently)
	}))
	t.Cleanup(old.Close)

	c := NewClientForTest(http.DefaultClient, old.URL, t.TempDir()+"/corpus.json")
	if _, err := c.get(context.Background(), old.URL+"/filme/x", getOptions{}); err != nil {
		t.Fatal(err)
	}
	if got := c.BaseURL(); got != moved.URL {
		t.Fatalf("BaseURL() = %q after the redirect; want the mirror %q", got, moved.URL)
	}
}

// A redirect to a different site is followed for that one request, but the
// client never adopts it as the catalog.
func TestClientIgnoresARedirectToAnotherSite(t *testing.T) {
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) }))
	t.Cleanup(other.Close)
	// 127.0.0.1 and localhost differ in the first label, so rewrite the target
	// to localhost: same server, different site as far as mirrorOf can tell.
	oURL, _ := url.Parse(other.URL)
	redirTo := "http://localhost:" + oURL.Port()
	old2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, redirTo, http.StatusFound)
	}))
	t.Cleanup(old2.Close)

	c := NewClientForTest(http.DefaultClient, old2.URL, t.TempDir()+"/corpus.json")
	if _, err := c.get(context.Background(), old2.URL+"/x", getOptions{}); err != nil {
		t.Fatal(err)
	}
	if got := c.BaseURL(); got != old2.URL {
		t.Fatalf("BaseURL() = %q; a redirect to another site must not become the catalog", got)
	}
}
