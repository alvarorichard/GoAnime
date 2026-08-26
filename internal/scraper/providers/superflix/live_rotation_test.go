package superflix

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
)

// TestLiveSuperFlixAfterHostRotation walks the browser-free half of the
// SuperFlix chain against the live host, so a host rotation is caught as a
// failing stage rather than "it stopped working". Opt in with GOANIME_LIVE=1.
func TestLiveSuperFlixAfterHostRotation(t *testing.T) {
	if os.Getenv("GOANIME_LIVE") == "" {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
	util.InitLogger()

	stage := func(name string, fn func() error) bool {
		start := time.Now()
		if err := fn(); err != nil {
			fmt.Printf("FAIL  %-34s %-8s  %v\n", name, time.Since(start).Round(time.Millisecond), err)
			t.Fail()
			return false
		}
		fmt.Printf("ok    %-34s %s\n", name, time.Since(start).Round(time.Millisecond))
		return true
	}

	// 1. the canonical host must answer directly, without a redirect: a 301
	//    downgrades the player POSTs to GETs and breaks bootstrap.
	stage("canonical host answers 200 (no 301)", func() error {
		c := NewSuperFlixClient()
		resp, err := c.client.Get(SuperFlixBase + "/")
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.Request.URL.Host != SuperFlixEmbedHost {
			return fmt.Errorf("redirected to %s — SuperFlixBase is stale, rotate it", resp.Request.URL.Host)
		}
		return nil
	})

	// 2. search must return parseable cards.
	var tmdbID string
	stage("search returns cards", func() error {
		results, err := NewSuperFlixClient().SearchMediaWithContext(context.Background(), "jojo")
		if err != nil {
			return err
		}
		if len(results) == 0 {
			return fmt.Errorf("no results")
		}
		for _, r := range results {
			if r.TMDBID != "" {
				tmdbID = r.TMDBID
				break
			}
		}
		fmt.Printf("      %d results, first=%q tmdb=%s\n", len(results), results[0].Title, tmdbID)
		if tmdbID == "" {
			return fmt.Errorf("no result carried a TMDB id")
		}
		return nil
	})

	// 3. the embed URL must be built on the live host.
	stage("embed URL targets the live host", func() error {
		if tmdbID == "" {
			return fmt.Errorf("no tmdb id from search")
		}
		embed := fmt.Sprintf("%s/filme/%s", SuperFlixBase, tmdbID)
		if !strings.Contains(embed, SuperFlixEmbedHost) {
			return fmt.Errorf("embed %s does not use %s", embed, SuperFlixEmbedHost)
		}
		fmt.Printf("      %s\n", embed)
		return nil
	})
}
