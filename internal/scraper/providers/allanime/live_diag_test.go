package allanime

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
)

// TestLiveAllAnimeDiagnostic walks the whole AllAnime chain against the live
// service and reports which stage breaks. Opt in with GOANIME_LIVE=1.
func TestLiveAllAnimeDiagnostic(t *testing.T) {
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
	util.InitLogger()
	c := NewAllAnimeClient()

	stage := func(name string, fn func() error) bool {
		start := time.Now()
		err := fn()
		if err != nil {
			fmt.Printf("FAIL  %-26s %-8s  %v\n", name, time.Since(start).Round(time.Millisecond), err)
			return false
		}
		fmt.Printf("ok    %-26s %-8s\n", name, time.Since(start).Round(time.Millisecond))
		return true
	}

	// 1. key bundle
	stage("fetch AES keys", func() error {
		k, err := c.getAAKeys()
		if err != nil {
			return err
		}
		fmt.Printf("      keys=%+v\n", k != nil)
		return nil
	})

	// 2. search
	var animeID string
	stage("search", func() error {
		res, err := c.SearchAnime("jojo")
		if err != nil {
			return err
		}
		if len(res) == 0 {
			return fmt.Errorf("no results")
		}
		fmt.Printf("      %d results, first=%q url=%s\n", len(res), res[0].Name, res[0].URL)
		animeID = res[0].URL
		return nil
	})
	if animeID == "" {
		t.Fatal("search stage produced no anime to continue with")
	}

	// 3. episodes
	var eps []string
	stage("episode list (sub)", func() error {
		var err error
		eps, err = c.GetEpisodesList(lastPathSegment(animeID), "sub")
		if err != nil {
			return err
		}
		if len(eps) == 0 {
			return fmt.Errorf("empty episode list")
		}
		fmt.Printf("      %d episodes, first=%s last=%s\n", len(eps), eps[0], eps[len(eps)-1])
		return nil
	})
	if len(eps) == 0 {
		t.Fatal("no episodes to continue with")
	}

	// 4. stream URL
	stage("stream URL for ep "+eps[0], func() error {
		url, meta, err := c.GetEpisodeURL(lastPathSegment(animeID), eps[0], "sub", "best")
		if err != nil {
			return err
		}
		if url == "" {
			return fmt.Errorf("empty stream URL")
		}
		fmt.Printf("      url=%.90s...\n      meta=%v\n", url, meta)
		return nil
	})
}

func lastPathSegment(s string) string {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '/' {
			return s[i+1:]
		}
	}
	return s
}
