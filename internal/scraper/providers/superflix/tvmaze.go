package superflix

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// tvmazeBaseURL is the keyless TVmaze API. Overridable in tests.
var tvmazeBaseURL = "https://api.tvmaze.com"

// GetEpisodesFromTVmaze lists a series' seasons + episodes from the public,
// KEYLESS TVmaze API, keyed by the IMDB id SuperFlix already gives us in search
// results. This is the browser-free episode source: SuperFlix's own per-series
// episode data lives only behind its Turnstile gate, and TVmaze covers the same
// standard season/episode numbering the SuperFlix player embed
// (/serie/{tmdb}/{s}/{e}) expects.
//
// Returns map[season]→episodes (future/undated episodes filtered out, matching
// the gated path's behaviour). The caller falls back to the browser path only if
// this fails (no IMDB id, show not on TVmaze, or network error).
func GetEpisodesFromTVmaze(ctx context.Context, httpClient *http.Client, imdbID string) (map[string][]SuperFlixEpisode, error) {
	if imdbID == "" {
		return nil, fmt.Errorf("tvmaze: empty imdb id")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}

	// 1. imdb → tvmaze show id.
	lookupURL := fmt.Sprintf("%s/lookup/shows?imdb=%s", tvmazeBaseURL, url.QueryEscape(imdbID))
	var show struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := tvmazeGetJSON(ctx, httpClient, lookupURL, &show); err != nil {
		return nil, fmt.Errorf("tvmaze lookup %s: %w", imdbID, err)
	}
	if show.ID == 0 {
		return nil, fmt.Errorf("tvmaze: no show for imdb %s", imdbID)
	}

	// 2. show id → episodes.
	epsURL := fmt.Sprintf("%s/shows/%d/episodes", tvmazeBaseURL, show.ID)
	var raw []struct {
		Season  int    `json:"season"`
		Number  int    `json:"number"`
		Name    string `json:"name"`
		AirDate string `json:"airdate"`
	}
	if err := tvmazeGetJSON(ctx, httpClient, epsURL, &raw); err != nil {
		return nil, fmt.Errorf("tvmaze episodes for %d: %w", show.ID, err)
	}

	out := make(map[string][]SuperFlixEpisode)
	for _, ep := range raw {
		if ep.Season < 1 || ep.Number < 1 { // skip specials / unnumbered
			continue
		}
		season := strconv.Itoa(ep.Season)
		out[season] = append(out[season], SuperFlixEpisode{
			EpiNum:  json.Number(strconv.Itoa(ep.Number)),
			Title:   ep.Name,
			AirDate: ep.AirDate,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("tvmaze: no episodes for imdb %s", imdbID)
	}

	out = filterEpisodesByAirDate(out, time.Now())
	// Drop seasons emptied by the air-date filter so the picker shows only
	// seasons that actually have watchable episodes.
	for season, eps := range out {
		if len(eps) == 0 {
			delete(out, season)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("tvmaze: no aired episodes for imdb %s", imdbID)
	}
	return out, nil
}

func tvmazeGetJSON(ctx context.Context, httpClient *http.Client, u string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", SuperFlixUserAgent)
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("not found (404)")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
	if err != nil {
		return err
	}
	return json.Unmarshal(body, dst)
}
