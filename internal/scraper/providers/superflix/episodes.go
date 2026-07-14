package superflix

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/util"
)

// ErrSuperFlixNoServers is returned when /player/bootstrap responds with an
// empty options list. This is a content-availability signal from SuperFlix
// (the upstream JS shows a "not yet released" screen in the same case), not
// a system or scraping error — callers should surface it to the user as

// ExtractEpisodes extracts episodes from a SuperFlix player/serie page.
//
// Two formats are supported:
//   - Legacy player page: a `var ALL_EPISODES = {...}` JS object (air-date
//     filtered, since it can contain unreleased placeholders).
//   - Current rotating frontend (superflix.bond / primeflix.mom /
//     lospobreflix.site / …): episodes rendered as
//     `<a data-season data-episode data-episode-id href="/episodio/...">`.
//     These are already release-filtered by the site, so no air-date pass.
func (c *SuperFlixClient) ExtractEpisodes(html string) (map[string][]SuperFlixEpisode, error) {
	if m := sfAllEpisodesRe.FindStringSubmatch(html); len(m) >= 2 {
		var result map[string][]SuperFlixEpisode
		if err := json.Unmarshal([]byte(m[1]), &result); err != nil {
			return nil, fmt.Errorf("failed to parse ALL_EPISODES: %w", err)
		}
		return filterEpisodesByAirDate(result, time.Now()), nil
	}

	if blob := parseWindowAllEpisodes(html); len(blob) > 0 {
		return blob, nil
	}

	if fe := parseFrontendEpisodes(html); len(fe) > 0 {
		return fe, nil
	}
	return nil, nil
}

// parseWindowAllEpisodes reads the `window.allEpisodes = {...};` blob the current
// rotating frontend injects. Unlike the rendered anchors (which expose only
// season/episode numbers), the blob carries every season at once with full
// metadata — title, epi_num and air_date. Episodes are air-date filtered the
// same way as the legacy ALL_EPISODES blob, since the dataset can include
// unreleased placeholders.
func parseWindowAllEpisodes(html string) map[string][]SuperFlixEpisode {
	m := sfWindowAllEpisodesRe.FindStringSubmatch(html)
	if len(m) < 2 {
		return nil
	}
	var result map[string][]SuperFlixEpisode
	if err := json.Unmarshal([]byte(m[1]), &result); err != nil {
		util.Debug("SuperFlix: failed to parse window.allEpisodes", "err", err)
		return nil
	}
	return filterEpisodesByAirDate(result, time.Now())
}

// parseFrontendEpisodes reads episodes from the rotating SuperFlix frontend
// serie page. Each `<a data-episode-id>` anchor carries the season and episode
// numbers we need to build the player URL later. Only the currently-loaded
// season's episodes are present on a given page (other seasons live at
// /serie/<slug>/<n>); GetEpisodes fetches those separately and merges.
func parseFrontendEpisodes(html string) map[string][]SuperFlixEpisode {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil
	}
	out := make(map[string][]SuperFlixEpisode)
	doc.Find("a[data-episode-id]").Each(func(_ int, a *goquery.Selection) {
		season, _ := a.Attr("data-season")
		epnum, _ := a.Attr("data-episode")
		if season == "" || epnum == "" {
			return
		}
		out[season] = append(out[season], SuperFlixEpisode{
			EpiNum: json.Number(epnum),
			Title:  "Episódio " + epnum,
		})
	})
	return out
}

// parseFrontendSeasons returns the distinct season numbers linked on a frontend
// serie page (the season dropdown / "/serie/<slug>/<n>" links).
func parseFrontendSeasons(html string) []string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil
	}
	seen := make(map[string]bool)
	var seasons []string
	re := regexp.MustCompile(`/serie/[a-z0-9-]+/(\d+)$`)
	doc.Find(`a[href]`).Each(func(_ int, a *goquery.Selection) {
		href, _ := a.Attr("href")
		href = strings.SplitN(href, "?", 2)[0]
		href = strings.SplitN(href, "#", 2)[0]
		if mm := re.FindStringSubmatch(href); len(mm) > 1 {
			if !seen[mm[1]] {
				seen[mm[1]] = true
				seasons = append(seasons, mm[1])
			}
		}
	})
	sort.Slice(seasons, func(i, j int) bool {
		ai, _ := strconv.Atoi(seasons[i])
		aj, _ := strconv.Atoi(seasons[j])
		return ai < aj
	})
	return seasons
}

// filterEpisodesByAirDate drops episodes with empty/"null" air_date and
// episodes whose air_date is strictly after the current UTC day.
//
// Comparison is done at day granularity in UTC so the result does not drift
// across midnight: a previous version used `t.After(now.Add(24*time.Hour))`,
// which kept tomorrow's episodes any time `now`'s UTC time-of-day was past
// 00:00 — flaky for any caller running with `now` in a timezone west of UTC.
func filterEpisodesByAirDate(result map[string][]SuperFlixEpisode, now time.Time) map[string][]SuperFlixEpisode {
	utcNow := now.UTC()
	today := time.Date(utcNow.Year(), utcNow.Month(), utcNow.Day(), 0, 0, 0, 0, time.UTC)
	for season, episodes := range result {
		var validEpisodes []SuperFlixEpisode
		for _, ep := range episodes {
			if ep.AirDate == "" || ep.AirDate == "null" {
				continue
			}
			if t, err := time.Parse("2006-01-02", ep.AirDate); err == nil {
				if t.After(today) {
					continue
				}
			}
			validEpisodes = append(validEpisodes, ep)
		}
		result[season] = validEpisodes
	}
	return result
}

// GetEpisodes returns the seasons and episodes for a series.
//
// Legacy player pages embed every season in one ALL_EPISODES blob. The current
// rotating frontend renders only the loaded season's episodes per page and
// links the others at /serie/<slug>/<n>, so when we detect that format we fetch
// each remaining season (via the gateway serie/<tmdb>/<n>, which redirects to
// the right frontend season) and merge. Per-season fetches reuse the already
// cleared CF profile, so they don't re-trigger the challenge.
func (c *SuperFlixClient) GetEpisodes(ctx context.Context, tmdbID string) (map[string][]SuperFlixEpisode, error) {
	// Production path: drive the browser solver directly. It returns the final
	// (rotating) frontend URL, which we need to resolve the per-season links
	// onto the right domain. The transport/HTTP path can't expose that.
	if c.browserSolver != nil {
		return c.getEpisodesViaBrowser(ctx, tmdbID)
	}

	// Test path (SetTestConfig): plain HTTP against an httptest server.
	html, err := c.GetPlayerPage(ctx, "serie", tmdbID, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to load player page: %w", err)
	}
	return c.ExtractEpisodes(html)
}

// getEpisodesViaBrowser solves the serie page, parses the loaded season, then
// solves each remaining season's frontend URL and merges. Per-season solves
// reuse the warm CF profile, so they don't re-trigger the challenge.
func (c *SuperFlixClient) getEpisodesViaBrowser(ctx context.Context, tmdbID string) (map[string][]SuperFlixEpisode, error) {
	base := strings.TrimSuffix(c.baseURL, "/")
	res, err := c.browserSolver.Solve(ctx, base+"/serie/"+tmdbID, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to load serie page: %w", err)
	}

	episodes := make(map[string][]SuperFlixEpisode)

	// Legacy ALL_EPISODES (rare now) carries every season in one shot.
	if legacy, lErr := c.ExtractEpisodes(res.HTML); lErr == nil && sfAllEpisodesRe.MatchString(res.HTML) {
		return legacy, nil
	}

	// Current frontend injects every season with air_date in window.allEpisodes,
	// so a single solve covers all seasons — no per-season fetch needed.
	if blob := parseWindowAllEpisodes(res.HTML); len(blob) > 0 {
		return blob, nil
	}

	for s, eps := range parseFrontendEpisodes(res.HTML) {
		episodes[s] = eps
	}

	// Resolve the other seasons' URLs against the solved frontend domain and
	// fetch each that we don't already have.
	for season, seasonURL := range resolveFrontendSeasonURLs(res.HTML, res.FinalURL) {
		if _, ok := episodes[season]; ok {
			continue
		}
		sres, sErr := c.browserSolver.Solve(ctx, seasonURL, 0)
		if sErr != nil {
			util.Debug("SuperFlix: failed to load season page", "season", season, "url", seasonURL, "err", sErr)
			continue
		}
		for s, eps := range parseFrontendEpisodes(sres.HTML) {
			if _, ok := episodes[s]; !ok {
				episodes[s] = eps
			}
		}
	}

	if len(episodes) == 0 {
		// Returning (nil, nil) here used to make an unparseable page look like a
		// title that simply has no episodes, which surfaced to the user as a bare
		// "no seasons found" with nothing to act on. Report it as the scrape
		// failure it is, and say which page defeated us.
		return nil, fmt.Errorf("%w: solved %s but it exposed no episode list (title=%q)",
			ErrSuperFlixNoEpisodeList, res.FinalURL, pageTitle(res.HTML))
	}
	return episodes, nil
}

// pageTitle pulls <title> out of a page for diagnostics. SuperFlix increasingly
// answers with an "Embed | <name>" shell that carries only a player iframe and
// no episode list, so the title is the quickest way to see that in a log.
func pageTitle(html string) string {
	m := sfTitleRe.FindStringSubmatch(html)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}

// resolveFrontendSeasonURLs maps season number -> absolute URL for every
// /serie/<slug>/<n> link on a frontend serie page, resolved against the page's
// final (post-redirect) URL so they hit the correct rotating domain.
func resolveFrontendSeasonURLs(html, finalURL string) map[string]string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil
	}
	var basePtr *url.URL
	if finalURL != "" {
		basePtr, _ = url.Parse(finalURL)
	}
	re := regexp.MustCompile(`/serie/[a-z0-9-]+/(\d+)$`)
	out := make(map[string]string)
	doc.Find(`a[href]`).Each(func(_ int, a *goquery.Selection) {
		href, _ := a.Attr("href")
		clean := strings.SplitN(href, "?", 2)[0]
		clean = strings.SplitN(clean, "#", 2)[0]
		m := re.FindStringSubmatch(clean)
		if m == nil {
			return
		}
		season := m[1]
		if _, ok := out[season]; ok {
			return
		}
		abs := clean
		if basePtr != nil {
			if ref, perr := url.Parse(clean); perr == nil {
				abs = basePtr.ResolveReference(ref).String()
			}
		}
		out[season] = abs
	})
	return out
}
