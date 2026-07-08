package superflix

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

// ErrSuperFlixNoServers is returned when /player/bootstrap responds with an
// empty options list. This is a content-availability signal from SuperFlix
// (the upstream JS shows a "not yet released" screen in the same case), not
// a system or scraping error — callers should surface it to the user as

// GetPlayerPage loads the player page for a given content
func (c *SuperFlixClient) GetPlayerPage(ctx context.Context, mediaType, mediaID, season, episode string) (string, error) {
	path := fmt.Sprintf("/%s/%s", mediaType, mediaID)
	if season != "" {
		path += "/" + season
	}
	if episode != "" {
		path += "/" + episode
	}

	pageURL := c.baseURL + path
	util.Debug("SuperFlix player page", "url", pageURL)

	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)
	req.Header.Set("Referer", c.baseURL+"/")
	req.Header.Set("Sec-Fetch-Dest", "iframe")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

// ExtractTokens extracts CSRF_TOKEN, PAGE_TOKEN, etc. from player HTML
func (c *SuperFlixClient) ExtractTokens(html string) *SuperFlixTokens {
	tokens := &SuperFlixTokens{}
	if m := sfCSRFTokenRe.FindStringSubmatch(html); len(m) > 1 {
		tokens.CSRF = m[1]
	}
	if m := sfPageTokenRe.FindStringSubmatch(html); len(m) > 1 {
		tokens.PageToken = m[1]
	}
	if m := sfContentIDRe.FindStringSubmatch(html); len(m) > 1 {
		tokens.ContentID = m[1]
	}
	if m := sfContentTypeRe.FindStringSubmatch(html); len(m) > 1 {
		tokens.ContentType = m[1]
	}
	if m := sfTitleRe.FindStringSubmatch(html); len(m) > 1 {
		tokens.Title = m[1]
	}
	return tokens
}

// Bootstrap calls /player/bootstrap to get server list
func (c *SuperFlixClient) Bootstrap(ctx context.Context, tokens *SuperFlixTokens) ([]SuperFlixServer, error) {
	bootstrapURL := c.baseURL + "/player/bootstrap"

	form := url.Values{
		"contentid":  {tokens.ContentID},
		"type":       {tokens.ContentType},
		"_token":     {tokens.CSRF},
		"page_token": {tokens.PageToken},
		"pageToken":  {tokens.PageToken},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", bootstrapURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", c.baseURL+"/")
	req.Header.Set("X-Page-Token", tokens.PageToken)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", c.baseURL)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if err := ensureJSONResponse("bootstrap", resp, body); err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Options []SuperFlixServer `json:"options"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode bootstrap response: %w", err)
	}

	return result.Data.Options, nil
}

// GetSourceURL calls /player/source to get the redirect URL for a video
func (c *SuperFlixClient) GetSourceURL(ctx context.Context, videoID string, tokens *SuperFlixTokens) (string, error) {
	sourceURL := c.baseURL + "/player/source"

	form := url.Values{
		"video_id":   {videoID},
		"page_token": {tokens.PageToken},
		"host":       {""},
		"site":       {""},
		"_token":     {tokens.CSRF},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", sourceURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", c.baseURL+"/")
	req.Header.Set("X-Page-Token", tokens.PageToken)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", c.baseURL)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if err := ensureJSONResponse("source", resp, body); err != nil {
		return "", err
	}

	var result struct {
		Data struct {
			VideoURL string `json:"video_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to decode source response: %w", err)
	}

	if result.Data.VideoURL == "" {
		return "", fmt.Errorf("no video URL in source response")
	}

	return result.Data.VideoURL, nil
}

// ResolveRedirect follows the SuperFlix redirect to get the external player URL
func (c *SuperFlixClient) ResolveRedirect(ctx context.Context, redirectURL string) (baseURL, videoHash, playerHTML string, err error) {
	// Use the client's transport if available, otherwise fall back to safe transport
	transport := c.client.Transport
	if transport == nil {
		transport = netx.SafeScraperTransport(30 * time.Second)
	}

	// Use a client that does NOT follow redirects automatically
	noRedirectClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", redirectURL, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)
	req.Header.Set("Referer", c.baseURL+"/")

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	location := redirectURL
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location = resp.Header.Get("Location")
		if location == "" {
			return "", "", "", fmt.Errorf("redirect with no Location header")
		}
	}

	// Follow to the final page
	req2, err := http.NewRequestWithContext(ctx, "GET", location, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create follow request: %w", err)
	}
	c.decorateRequest(req2)
	req2.Header.Set("Referer", c.baseURL+"/")

	followClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
	resp2, err := followClient.Do(req2)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to follow redirect: %w", err)
	}
	defer func() { _ = resp2.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp2.Body, 5*1024*1024))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read player page: %w", err)
	}

	finalURL := resp2.Request.URL.String()

	if strings.Contains(finalURL, "/video/") {
		parts := strings.SplitN(finalURL, "/video/", 2)
		baseURL = parts[0]
		videoHash = strings.SplitN(parts[1], "?", 2)[0]
		videoHash = strings.SplitN(videoHash, "#", 2)[0]
	} else {
		idx := strings.LastIndex(finalURL, "/")
		if idx > 0 {
			baseURL = finalURL[:idx]
			videoHash = strings.SplitN(finalURL[idx+1:], "?", 2)[0]
		}
	}

	return baseURL, videoHash, string(body), nil
}

// ExtractPlayerExtras extracts defaultAudio and subtitles from the external player HTML
func (c *SuperFlixClient) ExtractPlayerExtras(html string) (defaultAudio []string, subtitles []SuperFlixSubtitle) {
	if m := sfDefaultAudioRe.FindStringSubmatch(html); len(m) > 1 {
		_ = json.Unmarshal([]byte(m[1]), &defaultAudio)
	}

	if m := sfSubtitleRe.FindStringSubmatch(html); len(m) > 1 {
		for part := range strings.SplitSeq(m[1], ",") {
			sm := sfSubPartRe.FindStringSubmatch(part)
			if len(sm) > 2 {
				subtitles = append(subtitles, SuperFlixSubtitle{
					Lang: sm[1],
					URL:  sm[2],
				})
			}
		}
	}
	return
}

// GetVideoAPI calls the external player's API to get the actual stream URL
func (c *SuperFlixClient) GetVideoAPI(ctx context.Context, playerBaseURL, videoHash, referer string) (streamURL, thumbURL string, err error) {
	apiURL := fmt.Sprintf("%s/player/index.php?data=%s&do=getVideo", playerBaseURL, videoHash)

	form := url.Values{
		"hash": {videoHash},
		"r":    {c.baseURL + "/"},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", referer)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return "", "", fmt.Errorf("failed to read response: %w", err)
	}

	if err := ensureJSONResponse("video API", resp, body); err != nil {
		return "", "", err
	}

	var result struct {
		SecuredLink string `json:"securedLink"`
		VideoSource string `json:"videoSource"`
		VideoImage  string `json:"videoImage"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", fmt.Errorf("failed to decode video API response: %w", err)
	}

	switch {
	case result.SecuredLink != "":
		streamURL = result.SecuredLink
	case result.VideoSource != "":
		streamURL = result.VideoSource
	default:
		return "", "", fmt.Errorf("no stream URL in video API response")
	}

	return streamURL, result.VideoImage, nil
}

// embedStreamSolver is the subset of the browser solver used to extract a live
// stream by driving the warezcdn embed through its Turnstile gate and capturing
// the player's getVideo response. The transport's cfSolver interface stays
// minimal (Solve only); only the real *cfBrowserSolver implements this, so the
// type assertion in GetStreamURL is false for test fakes / a nil solver.
type embedStreamSolver interface {
	SniffEmbedStream(ctx context.Context, embedURL string, timeout time.Duration) (*CFStreamResult, error)
}

// sniffEmbedStreamAttempts is how many times the browser path drives the embed
// through the Turnstile gate before giving up. The managed challenge is
// probabilistic — it can fail to auto-pass on a cold tick — so a single
// re-solve rescues the common first-play flake instead of bouncing the user
// back to the search screen. Capped at 2 so the worst case (2×90s solve) still
// fits inside the caller's 210s context budget (see GetSuperFlixStreamURL).
const sniffEmbedStreamAttempts = 2

// sniffEmbedStreamWithRetry drives SniffEmbedStream up to
// sniffEmbedStreamAttempts times, retrying only transient solve failures (a gate
// that did not clear in time). It stops immediately once the context is done, so
// a user abort or an exhausted deadline is never retried, and it emits a
// user-facing notice before a retry so the wait does not look like a hang.
func sniffEmbedStreamWithRetry(ctx context.Context, solver embedStreamSolver, embedURL string) (*CFStreamResult, error) {
	var lastErr error
	for attempt := 1; attempt <= sniffEmbedStreamAttempts; attempt++ {
		res, err := solver.SniffEmbedStream(ctx, embedURL, 0)
		if err == nil {
			return res, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			break // context cancelled/expired — do not burn another solve
		}
		if attempt < sniffEmbedStreamAttempts {
			util.Info("Verification didn't complete on the first try — retrying once...")
			util.Debug("SuperFlix embed sniff retry", "attempt", attempt, "err", err)
		}
	}
	return nil, lastErr
}

// GetStreamURL resolves the playable stream for SuperFlix content.
//
// The legacy player-page→tokens→bootstrap→source pipeline is dead: the current
// site serves a Turnstile-gated embed with no inline tokens. So in production
// (browser solver present) we drive the embed through the gate and sniff the
// player's getVideo response for the signed HLS master. The legacy pipeline
// below is retained only for the httptest-backed unit tests (which null out the
// browser solver via SetTestConfig).
func (c *SuperFlixClient) GetStreamURL(ctx context.Context, mediaType, mediaID, season, episode string) (*SuperFlixStreamResult, error) {
	if solver, ok := c.browserSolver.(embedStreamSolver); ok {
		return c.getStreamViaBrowser(ctx, solver, mediaType, mediaID, season, episode)
	}

	html, err := c.GetPlayerPage(ctx, mediaType, mediaID, season, episode)
	if err != nil {
		return nil, fmt.Errorf("failed to load player page: %w", err)
	}

	tokens := c.ExtractTokens(html)
	if tokens.CSRF == "" || tokens.PageToken == "" {
		return nil, fmt.Errorf("failed to extract tokens from player page")
	}

	servers, err := c.Bootstrap(ctx, tokens)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap: %w", err)
	}
	if len(servers) == 0 {
		// Empty bootstrap on a fully-loaded player page means SuperFlix has
		// no provider for this specific content (typically placeholder
		// episodes whose `air_date` is null in ALL_EPISODES). The upstream
		// site renders a "not yet released" screen in the same case.
		// Annotate the error with the player URL and contentid so triage
		// doesn't confuse this with a network or scraping failure.
		playerPath := fmt.Sprintf("/%s/%s", mediaType, mediaID)
		if season != "" {
			playerPath += "/" + season
		}
		if episode != "" {
			playerPath += "/" + episode
		}
		return nil, fmt.Errorf("%w (url=%s%s, contentid=%s) — try another episode or source",
			ErrSuperFlixNoServers, c.baseURL, playerPath, tokens.ContentID)
	}

	// Pick first non-fallback server
	var videoIDStr string
	for _, s := range servers {
		var raw string
		if err := json.Unmarshal(s.ID, &raw); err == nil {
			if !strings.HasPrefix(raw, "fallback") {
				videoIDStr = raw
				break
			}
		}
		// Try as number
		var num json.Number
		if err := json.Unmarshal(s.ID, &num); err == nil {
			videoIDStr = num.String()
			break
		}
	}
	if videoIDStr == "" {
		// Fallback: use first server
		var raw string
		if err := json.Unmarshal(servers[0].ID, &raw); err == nil {
			videoIDStr = raw
		} else {
			var num json.Number
			if err := json.Unmarshal(servers[0].ID, &num); err == nil {
				videoIDStr = num.String()
			} else {
				return nil, fmt.Errorf("failed to parse server ID")
			}
		}
	}

	redirectURL, err := c.GetSourceURL(ctx, videoIDStr, tokens)
	if err != nil {
		return nil, fmt.Errorf("failed to get source URL: %w", err)
	}

	playerBaseURL, videoHash, playerHTML, err := c.ResolveRedirect(ctx, redirectURL)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve redirect: %w", err)
	}

	referer := fmt.Sprintf("%s/video/%s", playerBaseURL, videoHash)
	streamURL, thumbURL, err := c.GetVideoAPI(ctx, playerBaseURL, videoHash, referer)
	if err != nil {
		return nil, fmt.Errorf("failed to get video from API: %w", err)
	}

	result := &SuperFlixStreamResult{
		StreamURL: streamURL,
		Title:     tokens.Title,
		Referer:   playerBaseURL + "/",
		Thumb:     NormalizeSuperFlixImageURL(thumbURL),
	}

	defaultAudio, subtitles := c.ExtractPlayerExtras(playerHTML)
	result.DefaultAudio = defaultAudio
	result.Subtitles = subtitles

	return result, nil
}

// getStreamViaBrowser resolves the stream, preferring a browser-free path.
//
// The only browser-gated step is mapping tmdb→(playerHost, videoHash) through
// the embed host's Cloudflare Turnstile gate. The player host's getVideo
// endpoint that turns that
// pair into a fresh signed HLS link is NOT gated, so once the pair is cached we
// replay over plain HTTP with no browser. The headed browser therefore runs only
// on the FIRST play of a title — or when the cached host rotates out and the
// HTTP getVideo fails, which transparently falls back to a re-solve.
func (c *SuperFlixClient) getStreamViaBrowser(ctx context.Context, solver embedStreamSolver, mediaType, mediaID, season, episode string) (*SuperFlixStreamResult, error) {
	key := streamCacheKey(mediaType, mediaID, season, episode)

	// 1. Cached (host, hash) → pure-HTTP getVideo, no browser.
	if ent, ok := defaultStreamCache.get(key); ok {
		referer := ent.Host + "/video/" + ent.Hash
		streamURL, thumb, err := c.GetVideoAPI(ctx, ent.Host, ent.Hash, referer)
		if err == nil && streamURL != "" {
			util.Debug("SuperFlix stream from cache (no browser)", "key", key, "host", ent.Host)
			return &SuperFlixStreamResult{
				StreamURL: streamURL,
				Referer:   ent.Host + "/",
				Thumb:     NormalizeSuperFlixImageURL(thumb),
			}, nil
		}
		util.Debug("SuperFlix cached stream stale, re-solving", "key", key, "err", err)
	}

	// 2. Cache miss / stale → drive the headed browser through the gate once,
	//    capture the stream + (host, hash), and cache the pair for next time.
	var embedURL string
	if mediaType == "serie" {
		s, e := season, episode
		if s == "" {
			s = "1"
		}
		if e == "" {
			e = "1"
		}
		embedURL = fmt.Sprintf("https://%s/serie/%s/%s/%s", SuperFlixEmbedHost, mediaID, s, e)
	} else {
		embedURL = fmt.Sprintf("https://%s/filme/%s", SuperFlixEmbedHost, mediaID)
	}

	res, err := sniffEmbedStreamWithRetry(ctx, solver, embedURL)
	if err != nil {
		return nil, fmt.Errorf("superflix embed stream sniff failed (%s): %w", embedURL, err)
	}

	// The raw-media fallback capture has no player host/hash (those come only
	// from the getVideo URL); caching an empty pair would just force a doomed
	// GetVideoAPI round-trip before every future re-solve.
	if res.PlayerHost != "" && res.VideoHash != "" {
		defaultStreamCache.put(key, streamCacheEntry{Host: res.PlayerHost, Hash: res.VideoHash})
	}

	referer := res.Referer
	if referer == "" {
		referer = "https://" + SuperFlixEmbedHost + "/"
	}
	return &SuperFlixStreamResult{
		StreamURL: res.StreamURL,
		Referer:   referer,
	}, nil
}
