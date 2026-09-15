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

	// 1. runtime discovery must land on a host that answers without
	//    redirecting. This is the stage that actually keeps playback working
	//    across a rotation, so it is checked before the constant.
	stage("discovery finds the live host", func() error {
		host, err := probeLiveHost(context.Background())
		if err != nil {
			return err
		}
		fmt.Printf("      discovered %s (seed %s)\n", host, SuperFlixEmbedHost)
		if host == "" {
			return fmt.Errorf("discovery returned an empty host")
		}
		return nil
	})

	// 2. the compiled default vs. the live host. Advisory, NOT a failure:
	//    discovery walks every retired alias, so a stale default no longer
	//    breaks playback — it only means the next release should rotate it.
	//    This used to fail the whole test, and it went red on a day when
	//    discovery had already found .monster and search was working. It also
	//    fetched through the Cloudflare-solving client, spending 3 minutes on a
	//    browser solve just to learn a host name the probe had already reported.
	if host, err := probeLiveHost(context.Background()); err == nil && host != SuperFlixEmbedHost {
		fmt.Printf("note  compiled default %s is stale; live host is %s — rotate SuperFlixBase next release\n",
			SuperFlixEmbedHost, host)
	}

	// 3. search must return parseable cards.
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

	// 4. the embed URL must be built on the live host.
	//
	// This used to format SuperFlixBase and then check it contained
	// SuperFlixEmbedHost — two constants compared with each other, so it passed
	// while printing a URL on the dead .baby host. It now builds the URL the way
	// production does and checks it against what discovery found.
	stage("embed URL targets the live host", func() error {
		if tmdbID == "" {
			return fmt.Errorf("no tmdb id from search")
		}
		live, err := probeLiveHost(context.Background())
		if err != nil {
			return err
		}
		embed := fmt.Sprintf("%s/filme/%s", LiveBase(context.Background()), tmdbID)
		if !strings.Contains(embed, "https://"+live+"/") {
			return fmt.Errorf("embed %s does not use the live host %s", embed, live)
		}
		fmt.Printf("      %s\n", embed)
		return nil
	})
}
