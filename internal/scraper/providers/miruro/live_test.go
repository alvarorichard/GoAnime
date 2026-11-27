package miruro

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// TestLiveMiruroChain walks search → episodes → stream against the real site,
// for each audio track, and then fetches the playlist and the first subtitle
// the way the player will: with the referer and origin the scraper handed out.
// Those headers are the one failure a parser test cannot see, since one of the
// site's CDN families answers 403 without the Origin.
//
// Opt in with GOANIME_LIVE=1; never runs in CI.
func TestLiveMiruroChain(t *testing.T) {
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
	util.InitLogger()
	c := NewMiruroClient()

	results, err := c.SearchAnime(context.Background(), "frieren")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("search returned no results")
	}
	fmt.Printf("search   : %d results\n", len(results))
	for i, r := range results[:min(3, len(results))] {
		fmt.Printf("           [%d] %-52s %s anilist=%d mal=%d year=%s audio=%q\n",
			i, truncate(r.Name, 52), r.URL, r.AnilistID, r.MalID, r.Year, r.Audio)
	}
	target := results[0]
	if target.AnilistID == 0 {
		t.Errorf("the first result carries no AniList id: %+v", target)
	}

	eps, err := c.GetAnimeEpisodes(context.Background(), target.URL)
	if err != nil {
		t.Fatalf("episodes for %s: %v", target.URL, err)
	}
	fmt.Printf("episodes : %d for %q (first=%s last=%s)\n           %s\n",
		len(eps), truncate(target.Name, 40), eps[0].Number, eps[len(eps)-1].Number, eps[0].URL)

	for _, audio := range []string{"sub", "dub"} {
		t.Setenv("GOANIME_MIRURO_AUDIO", audio)
		for _, q := range []string{"best", "720p"} {
			streamURL, meta, err := c.GetEpisodeStreamURL(context.Background(), eps[0].URL, q)
			if err != nil {
				t.Errorf("stream (%s, %s): %v", audio, q, err)
				continue
			}
			fmt.Printf("stream %s %-5s: %s\n           meta=%s\n", audio, q, truncate(streamURL, 88), truncate(fmt.Sprint(meta), 300))
			if meta["audio_lang"] != audio {
				t.Errorf("stream (%s, %s): picked %q", audio, q, meta["audio_lang"])
			}
			if status := fetchStatus(t, streamURL, meta["referer"], meta["origin"]); status != http.StatusOK {
				t.Errorf("stream (%s, %s): playlist answered %d with referer %q origin %q; playback would fail",
					audio, q, status, meta["referer"], meta["origin"])
			}
			if subs := meta["subtitles"]; subs != "" {
				first := strings.SplitN(strings.SplitN(subs, `"url":"`, 2)[1], `"`, 2)[0]
				if status := fetchStatus(t, first, meta["referer"], meta["origin"]); status != http.StatusOK {
					t.Errorf("subtitle (%s): answered %d", audio, status)
				}
			}
		}
	}
}

// fetchStatus GETs a URL with the headers the scraper handed out.
func fetchStatus(t *testing.T, rawURL, referer, origin string) int {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	resp, err := util.NewFastClient().Do(req) // #nosec G704 -- URL came from the live scraper above
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
