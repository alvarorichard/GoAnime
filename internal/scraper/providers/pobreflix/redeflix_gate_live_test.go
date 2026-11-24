//go:build live

package pobreflix_test

import (
	"context"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/cfsolver"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/pobreflix"
)

// TestLivePobreflixRedeflixGateWithSolver resolves Interestelar through the
// RedeFlixApi Turnstile gate with the installed cfsolve.
func TestLivePobreflixRedeflixGateWithSolver(t *testing.T) {
	if !cfsolver.Register() {
		t.Skip("cfsolve is not installed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
	defer cancel()
	c := pobreflix.NewClient()
	s, err := c.ResolveRedeflix(ctx, "https://redeflixapi.click/filme/157336", pobreflix.AudioDubbed)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	t.Logf("resolved: host %s audio %q", hostOf(s.URL), s.Audio)
}

func hostOf(raw string) string {
	for i := 0; i+2 < len(raw); i++ {
		if raw[i:i+3] == "://" {
			rest := raw[i+3:]
			for j := range rest {
				if rest[j] == '/' {
					return rest[:j]
				}
			}
			return rest
		}
	}
	return raw
}
