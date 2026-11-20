package pobreflix

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
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

// The cached corpus holds absolute URLs on the host they were read from. When
// the client moves to the mirror, its entries move with it: left behind they
// would redirect on every open, and fail when the old domain goes.
func TestMovingToAMirrorRewritesTheCorpusHost(t *testing.T) {
	moved := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) }))
	t.Cleanup(moved.Close)
	old := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, moved.URL+r.URL.Path, http.StatusMovedPermanently)
	}))
	t.Cleanup(old.Close)
	cache := t.TempDir() + "/corpus.json"
	c := NewClientForTest(http.DefaultClient, old.URL, cache)
	c.search.entries = []entry{
		{URL: old.URL + "/filme/todo-poderoso.html", BaseSlug: "todo-poderoso", Kind: KindMovie},
		{URL: "https://unrelated.example/filme/x.html", BaseSlug: "x", Kind: KindMovie},
	}

	if _, err := c.get(context.Background(), old.URL+"/filme/x", getOptions{}); err != nil {
		t.Fatal(err)
	}

	got := c.search.entries
	if want := moved.URL + "/filme/todo-poderoso.html"; got[0].URL != want {
		t.Fatalf("entry = %q; want it on the mirror: %q", got[0].URL, want)
	}
	if got[1].URL != "https://unrelated.example/filme/x.html" {
		t.Fatalf("an entry on another host was rewritten: %q", got[1].URL)
	}
	if raw, err := os.ReadFile(cache); err != nil || !strings.Contains(string(raw), moved.URL) {
		t.Fatalf("the cache on disk still names the old host: err=%v\n%s", err, raw)
	}
}

// Many requests at once, the first of which moves the base: no race, and the
// base ends on the mirror exactly once.
func TestFollowingAMirrorIsRaceFree(t *testing.T) {
	moved := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) }))
	t.Cleanup(moved.Close)
	old := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, moved.URL+r.URL.Path, http.StatusMovedPermanently)
	}))
	t.Cleanup(old.Close)
	c := NewClientForTest(http.DefaultClient, old.URL, t.TempDir()+"/corpus.json")

	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = c.get(context.Background(), old.URL+"/a", getOptions{})
			_ = c.BaseURL()
		}()
	}
	wg.Wait()

	if c.BaseURL() != moved.URL {
		t.Fatalf("BaseURL() = %q; want %q", c.BaseURL(), moved.URL)
	}
}
