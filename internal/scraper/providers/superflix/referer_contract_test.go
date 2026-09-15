package superflix

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// The Referer contract.
//
// SuperFlix's CDN answers 403 to a signed playlist requested with anything but
// the player's own /video/<hash> page as Referer. That value is produced in
// FOUR places, and the 2026-08-26 outage was two of them quietly disagreeing:
// each built the correct Referer for its own getVideo call and then handed mpv
// the bare origin on the very next line.
//
// Pinning the four call sites one by one is not enough — the defect is that a
// NEW producer can be added and drift again. So this file guards three levels:
//
//  1. every producer, exercised end to end, returns the /video/<hash> form;
//  2. no result literal may build a Referer by string concatenation at all;
//  3. (live, opt-in) the upstream rules our arguments encode still hold.
// =============================================================================

// sfFullServer stands up the whole non-browser pipeline: player page → bootstrap
// → source → redirect → getVideo, so GetStreamURL can be driven offline.
func sfFullServer(t *testing.T) string {
	t.Helper()
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	mux.HandleFunc("/player/bootstrap", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, sfBootstrapJSON)
	})
	mux.HandleFunc("/player/source", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":{"video_url":"%s/video/hash123"}}`, srv.URL)
	})
	mux.HandleFunc("/video/hash123", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, realPlayerPage)
	})
	mux.HandleFunc("/player/index.php", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"securedLink":"https://cdn.example/master.m3u8"}`)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, sfTokenedPlayerPage)
	})
	return srv.URL
}

// sfTokenedPlayerPage is the page variant that carries BOTH tokens. The shared
// sfRealPlayerPage fixture has an empty CSRF_TOKEN, which GetStreamURL rejects
// before it ever reaches the code under test here.
const sfTokenedPlayerPage = `<html><head><title>Player | Matrix</title></head><body><script>
var CSRF_TOKEN = "csrf-abc123";
var PAGE_TOKEN = "eyJleHAiOjE3ODM4MDg3MzJ9";
var INITIAL_CONTENT_ID = 127972;
var CONTENT_TYPE = "serie";
var CURRENT_EPISODE = 1;
</script></body></html>`

// TestEveryStreamProducerUsesVideoPageReferer drives each way a caller can get a
// SuperFlixStreamResult and asserts they all agree on the Referer. Two of these
// shipped the bare origin; the cached replay and the browser sniff were fixed
// first, and GetStreamURL's own server path was still wrong after that.
func TestEveryStreamProducerUsesVideoPageReferer(t *testing.T) {
	// Not parallel: these swap the package-global stream cache.
	produce := map[string]func(t *testing.T, base string) *SuperFlixStreamResult{
		"StreamFromServer": func(t *testing.T, base string) *SuperFlixStreamResult {
			c := NewClientForTest(base)
			tokens := &SuperFlixTokens{ContentID: "1", PageToken: "tok"}
			res, err := c.StreamFromServer(context.Background(), tokens, "159462", "serie", "42821", "1", "3")
			require.NoError(t, err)
			return res
		},
		"TryCachedStream": func(t *testing.T, base string) *SuperFlixStreamResult {
			c := NewClientForTest(base)
			defaultStreamCache.put(streamCacheKey("serie", "42821", "1", "3"),
				streamCacheEntry{Host: base, Hash: "hash123"})
			res, ok := c.TryCachedStream(context.Background(), "serie", "42821", "1", "3")
			require.True(t, ok, "a seeded episode must replay from cache")
			return res
		},
		"GetStreamURL": func(t *testing.T, base string) *SuperFlixStreamResult {
			c := NewClientForTest(base)
			res, err := c.GetStreamURL(context.Background(), "serie", "42821", "1", "3")
			require.NoError(t, err)
			return res
		},
	}

	for name, fn := range produce {
		t.Run(name, func(t *testing.T) {
			withFreshStreamCache(t)
			base := sfFullServer(t)
			res := fn(t, base)
			require.NotNil(t, res)

			assert.Equal(t, base+"/video/hash123", res.Referer,
				"%s must hand mpv the player's /video/<hash> page", name)
			// The exact value the CDN rejects. Stated separately so a failure
			// says *which* mistake was made, not just that two strings differ.
			assert.NotEqual(t, base+"/", res.Referer,
				"%s regressed to the bare origin — the CDN 403s every signed playlist for it", name)
		})
	}
}

// TestNoStreamResultBuildsRefererByConcatenation is the structural guard: it is
// what would have caught the two drifted sites without anyone knowing to look.
//
// Every Referer in a result literal must come from a variable or from
// playerRefererFor. `playerBaseURL + "/"` is exactly the shape of the bug, and
// string concatenation is how it got written both times.
func TestNoStreamResultBuildsRefererByConcatenation(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	require.NoError(t, err)

	resultTypes := map[string]bool{"SuperFlixStreamResult": true, "CFStreamResult": true}
	var checked int

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, pErr := parser.ParseFile(fset, name, nil, 0)
		require.NoError(t, pErr)

		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}
			name, ok := litTypeName(lit)
			if !ok || !resultTypes[name] {
				return true
			}
			for _, elt := range lit.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				key, ok := kv.Key.(*ast.Ident)
				if !ok || key.Name != "Referer" {
					continue
				}
				checked++
				pos := fset.Position(kv.Pos())
				switch v := kv.Value.(type) {
				case *ast.Ident:
					// A variable — its own construction is covered by the
					// behavioural test above.
				case *ast.CallExpr:
					fn, _ := v.Fun.(*ast.Ident)
					assert.NotNil(t, fn, "%s: unexpected Referer call form", pos)
					if fn != nil {
						assert.Equal(t, "playerRefererFor", fn.Name,
							"%s: build the Referer with playerRefererFor", pos)
					}
				default:
					t.Errorf("%s: Referer is built inline (%T) instead of via "+
						"playerRefererFor — this is the exact shape of the 2026-08-26 "+
						"outage, where `playerBaseURL + \"/\"` shipped a Referer the CDN 403s",
						pos, kv.Value)
				}
			}
			return true
		})
	}
	assert.NotZero(t, checked, "found no Referer fields to check — did the result type get renamed?")
}

// litTypeName returns the bare type name of a composite literal (T{} or &T{}).
func litTypeName(lit *ast.CompositeLit) (string, bool) {
	switch t := lit.Type.(type) {
	case *ast.Ident:
		return t.Name, true
	case *ast.SelectorExpr:
		return t.Sel.Name, true
	}
	return "", false
}
