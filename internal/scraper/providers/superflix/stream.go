package superflix

import (
	"context"
	"encoding/json"
	"errors"
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

// isNativePlayerHost reports whether a resolved player "host" is SuperFlix's own
// native player (…/player/native/media/<id>) rather than an external warezcdn
// host. Observed live 2026-07-22: /player/source for some titles now redirects to
// this native player instead of an external host. It does NOT implement the
// /player/index.php?do=getVideo contract (answers 405 HTML), so its (host, hash)
// pair must never be cached or replayed — doing so poisoned the stream cache and
// cost a wasted browser solve plus a doomed getVideo round-trip on every play.
func isNativePlayerHost(host string) bool {
	return strings.Contains(host, "/player/native")
}

// errSuperFlixNativePlayer signals that the chosen server resolved to the native
// player, which the HTTP pipeline cannot extract from — the caller falls back to
// the embed sniff, which handles it.
var errSuperFlixNativePlayer = errors.New("superflix: server resolved to the native player (no getVideo API); use the embed sniff")

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
// maxPlayerPageAttempts bounds the retry in playerPageWithTokens, and
// playerPageRetryDelay spaces the attempts out.
//
// SuperFlix answers the same player URL with several page variants, and only some
// carry PAGE_TOKEN. With the browser solve allowed the tokened variant often
// arrives on the first attempt (measured ~6s warm); the retries cover the site's
// habit of serving a token-less variant even after the gate is cleared. Once the
// gate is warm each retry is a cheap ~300ms fetch (no re-solve), so a handful of
// them costs little and materially lifts the hit rate.
//
// The delay is not decoration: hammering the origin earns a plain-text "Too many
// requests" that locks the server list out for minutes — a self-inflicted wound,
// since the fallback then has to do all the work anyway.
const (
	maxPlayerPageAttempts = 6
	// 300ms is enough spacing to avoid the "Too many requests" wall within a
	// single play's handful of attempts (that wall came from HUNDREDS of requests
	// across probing, not from one play), while keeping a full miss — 6 attempts
	// that all get the token-less variant — down to ~2s before we fall back,
	// instead of the ~5s the old 800ms spacing cost.
	playerPageRetryDelay = 300 * time.Millisecond
)

// ErrSuperFlixRateLimited is returned when SuperFlix answers "Too many requests".
//
// It arrives as a 200 with a 17-byte plain-text body, so nothing upstream treats
// it as an error — callers must not retry into it, only back off.
var ErrSuperFlixRateLimited = errors.New("superflix: rate limited (too many requests)")

// isRateLimited reports whether a response body is SuperFlix's rate-limit notice
// rather than a page.
func isRateLimited(html string) bool {
	if len(html) > 200 {
		return false
	}
	return strings.Contains(strings.ToLower(html), "too many requests")
}

// playerPageWithTokens loads the REAL player page — the one carrying PAGE_TOKEN
// and the content id, without which /player/bootstrap cannot be called — retrying
// past the token-less page variants SuperFlix keeps serving.
//
// The retry deliberately REUSES the client, cf_clearance cookie and all. An
// earlier version rebuilt the transport between attempts, on the theory that the
// shell was sticky per connection. That threw away the Cloudflare clearance, so
// every attempt re-armed the gate and paid a fresh headed-browser solve: measured
// against Mushoku Tensei, attempt 1 took 3m02s, attempt 2 took 1m56s, and the rest
// blew the deadline. Whatever a fresh transport might buy, it cannot be worth
// re-solving Cloudflare eight times.
func (c *SuperFlixClient) playerPageWithTokens(ctx context.Context, mediaType, mediaID, season, episode string) (*SuperFlixTokens, error) {
	var lastErr error

	for attempt := range maxPlayerPageAttempts {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if attempt > 0 {
			select {
			case <-time.After(playerPageRetryDelay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		html, err := c.GetPlayerPage(ctx, mediaType, mediaID, season, episode)
		if err != nil {
			lastErr = err
			continue
		}

		// SuperFlix serves its rate-limit notice as a 200 with a plain-text body, so
		// it reaches us looking like a page. Retrying into it only digs the hole
		// deeper — give up on the server list and let the caller fall back.
		if isRateLimited(html) {
			return nil, ErrSuperFlixRateLimited
		}
		// Once the browser has had its short, genuine-iframe grace period, a
		// restricted shell cannot become a tokened player by issuing the same
		// top-level request again.  Retrying it six times only keeps the browser
		// window on screen and delays the stream fallback.
		if isRestrictedEmbedPage([]byte(html)) {
			return nil, ErrSuperFlixRestricted
		}

		// CSRF_TOKEN is deliberately empty on the current pages (`var CSRF_TOKEN = ""`),
		// so requiring it — as this code used to — rejected every real page and made
		// the server list unreachable. PAGE_TOKEN is the one bootstrap validates.
		tokens := c.ExtractTokens(html)
		if tokens.PageToken != "" && tokens.ContentID != "" {
			util.Debug("SuperFlix: got the real player page", "attempt", attempt+1, "contentID", tokens.ContentID)
			return tokens, nil
		}
		lastErr = fmt.Errorf("player page carried no tokens (%d bytes — the shell, not the player)", len(html))
	}

	return nil, fmt.Errorf("could not load a SuperFlix player page with tokens after %d attempts: %w", maxPlayerPageAttempts, lastErr)
}

// GetServers lists the servers SuperFlix offers for a title/episode, each tagged
// Dublado or Legendado (see SuperFlixServer.Type).
//
// This is the only path that exposes them. The embed page the browser sniff lands
// on is a bare player shell with no server list, which is why stream resolution
// used to silently take whatever the embed happened to play — the user could
// choose neither the server nor the audio.
//
// Placeholder entries (the site's own "fallback" options) are dropped, exactly as
// the upstream player drops them before rendering its list.
func (c *SuperFlixClient) GetServers(ctx context.Context, mediaType, mediaID, season, episode string) ([]SuperFlixServer, *SuperFlixTokens, error) {
	// The tokened player page — the only page that yields the server list — sits
	// behind Cloudflare, so getting it REQUIRES the browser solve. Measured live:
	// with the persistent profile warm (which it is after any prior SuperFlix play,
	// and the profile persists on disk across sessions) the solve is ~6s. An earlier
	// version forbade the solve here to dodge a hang; that just made the whole
	// feature dead — the page never carried tokens without it.
	//
	// The hang that prompted the ban was a DIFFERENT bug: rebuilding the transport
	// between attempts discarded cf_clearance and re-solved a cold gate every time
	// (Mushoku Tensei: 3m19s + 1m56s + …). That is fixed by reusing the client (see
	// playerPageWithTokens). The remaining cold-profile cost — one solve on the very
	// first play — is unavoidable: the stream itself needs a solve too, so paying it
	// here, where it also buys the server choice, is strictly better. The caller
	// bounds this whole call with a budget and falls back to the sniff if it is
	// exceeded (see sfServerListBudget).
	tokens, err := c.playerPageWithTokens(ctx, mediaType, mediaID, season, episode)
	if err != nil {
		return nil, nil, err
	}

	servers, err := c.Bootstrap(ctx, tokens)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list servers: %w", err)
	}

	playable := make([]SuperFlixServer, 0, len(servers))
	for _, s := range servers {
		if s.IsFallback() || s.IDString() == "" {
			continue
		}
		playable = append(playable, s)
	}
	if len(playable) == 0 {
		return nil, nil, fmt.Errorf("%w (contentid=%s)", ErrSuperFlixNoServers, tokens.ContentID)
	}

	util.Debug("SuperFlix servers", "count", len(playable), "contentID", tokens.ContentID)
	return playable, tokens, nil
}

// StreamFromServer resolves the stream for one specific server, so the caller can
// honor the user's pick instead of guessing.
func (c *SuperFlixClient) StreamFromServer(ctx context.Context, tokens *SuperFlixTokens, serverID, mediaType, mediaID, season, episode string) (*SuperFlixStreamResult, error) {
	redirectURL, err := c.GetSourceURL(ctx, serverID, tokens)
	if err != nil {
		return nil, fmt.Errorf("failed to get source URL: %w", err)
	}

	playerBaseURL, videoHash, playerHTML, err := c.ResolveRedirect(ctx, redirectURL)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve redirect: %w", err)
	}

	// The native player answers getVideo with 405 HTML — bail out NOW so the
	// caller falls back to the embed sniff without the doomed round-trip, and
	// without caching a (host, hash) pair that can never replay.
	if isNativePlayerHost(playerBaseURL) {
		return nil, fmt.Errorf("%w (%s)", errSuperFlixNativePlayer, playerBaseURL)
	}

	// Cache the (host, hash) so the NEXT play of this exact episode replays over
	// plain HTTP with no browser and no server-list fetch (see TryCachedStream).
	// This is what turns a re-watch or a resume from an ~8s open into a ~1s one.
	cacheKey := streamCacheKey(mediaType, mediaID, season, episode)
	defaultStreamCache.put(cacheKey, streamCacheEntry{Host: playerBaseURL, Hash: videoHash})

	referer := fmt.Sprintf("%s/video/%s", playerBaseURL, videoHash)
	streamURL, thumbURL, err := c.GetVideoAPI(ctx, playerBaseURL, videoHash, referer)
	if err != nil {
		return nil, fmt.Errorf("failed to get video from API: %w", err)
	}
	defaultAudio, subtitles := c.ExtractPlayerExtras(playerHTML)
	// Server resolution already loaded the player HTML, so retain the audio and
	// subtitle metadata with the stream key.  A replay can then start with the
	// complete track list without a second player-page request.
	if len(defaultAudio) > 0 || len(subtitles) > 0 {
		defaultStreamCache.put(cacheKey, streamCacheEntry{
			Host:         playerBaseURL,
			Hash:         videoHash,
			DefaultAudio: defaultAudio,
			Subtitles:    subtitles,
			ExtrasCached: true,
		})
	}
	return &SuperFlixStreamResult{
		StreamURL:    streamURL,
		Title:        tokens.Title,
		Referer:      playerBaseURL + "/",
		Thumb:        NormalizeSuperFlixImageURL(thumbURL),
		DefaultAudio: defaultAudio,
		Subtitles:    subtitles,
	}, nil
}

// TryCachedStream replays a previously-resolved stream over plain HTTP — no
// browser, no server-list fetch. It is the fast path for a repeat or resumed play:
// the (host, hash) captured last time yields a fresh signed HLS link from the
// ungated getVideo endpoint, so a re-watch opens in ~1s instead of paying the full
// Cloudflare + resolution cost again.
//
// Returns false on a cache miss, or when the cached host has rotated out (getVideo
// fails), so the caller re-resolves from scratch.
func (c *SuperFlixClient) TryCachedStream(ctx context.Context, mediaType, mediaID, season, episode string) (*SuperFlixStreamResult, bool) {
	return c.streamFromCache(ctx, streamCacheKey(mediaType, mediaID, season, episode))
}

// streamFromCache is the shared cache-replay used by both TryCachedStream and the
// browser path's first step.
func (c *SuperFlixClient) streamFromCache(ctx context.Context, key string) (*SuperFlixStreamResult, bool) {
	ent, ok := defaultStreamCache.get(key)
	if !ok {
		return nil, false
	}
	// Self-heal entries written before the native player was recognized: they
	// point at SuperFlix's own /player/native/… URL, whose getVideo always 405s.
	// Replaying one wastes round-trips (and a browser solve) on every play.
	if isNativePlayerHost(ent.Host) {
		util.Debug("SuperFlix cache entry points at the native player (never replayable); dropping it", "key", key, "host", ent.Host)
		defaultStreamCache.del(key)
		return nil, false
	}
	referer := ent.Host + "/video/" + ent.Hash

	// The fresh signed HLS URL and the player metadata are independent HTTP
	// requests. Run them in the shared bounded parallel executor and wait for
	// both, so cache replay keeps every subtitle/audio track without paying their
	// latencies serially.
	var (
		streamURL, thumb string
		streamErr        error
		dead             bool
		audio            = ent.DefaultAudio
		subs             = ent.Subtitles
	)
	util.ParallelExecute(2,
		func() {
			streamURL, thumb, streamErr = c.GetVideoAPI(ctx, ent.Host, ent.Hash, referer)
			// getVideo still signs URLs on the cached host even after that host
			// rotates out of the CDN — the signed link then answers 403/404 and mpv
			// dies a few seconds after launch. Probe the freshly signed URL right
			// here, inside the parallel branch, so the CDN check overlaps the
			// extras fetch instead of adding a third serial round-trip.
			if streamErr == nil && streamURL != "" {
				dead = c.streamURLDead(WithoutBrowserSolve(ctx), streamURL, ent.Host+"/")
			}
		},
		func() {
			if !ent.ExtrasCached {
				audio, subs = c.fetchPlayerExtras(ctx, ent.Host, ent.Hash)
			}
		},
	)
	if streamErr != nil || streamURL == "" {
		util.Debug("SuperFlix cached stream stale, will re-resolve", "key", key, "err", streamErr)
		return nil, false
	}
	// The CDN definitively rejected the freshly signed URL: drop the entry so
	// the caller re-resolves through the browser instead of handing a dead link
	// to the player.
	if dead {
		util.Debug("SuperFlix cached host rotated (CDN rejected signed URL), re-resolving", "key", key, "host", ent.Host)
		defaultStreamCache.del(key)
		return nil, false
	}
	if !ent.ExtrasCached {
		defaultStreamCache.put(key, streamCacheEntry{
			Host:         ent.Host,
			Hash:         ent.Hash,
			DefaultAudio: audio,
			Subtitles:    subs,
			ExtrasCached: len(audio) > 0 || len(subs) > 0,
		})
	}
	util.Debug("SuperFlix stream from cache (no browser, no server list)", "key", key, "host", ent.Host)
	return &SuperFlixStreamResult{
		StreamURL:    streamURL,
		Referer:      ent.Host + "/",
		Thumb:        NormalizeSuperFlixImageURL(thumb),
		DefaultAudio: audio,
		Subtitles:    subs,
	}, true
}

// streamURLDead reports whether the CDN definitively rejects a freshly signed
// HLS URL. It issues a tiny ranged GET with the same Referer mpv will use, so
// a rotated-out cache host is caught here (403/404/410) instead of by a dying
// mpv. Ambiguous outcomes (network blip, timeout, other statuses) return false
// so a transient failure never nukes an otherwise good cache entry.
func (c *SuperFlixClient) streamURLDead(ctx context.Context, streamURL, referer string) bool {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, streamURL, nil)
	if err != nil {
		return false
	}
	c.decorateRequest(req)
	req.Header.Set("Referer", referer)
	req.Header.Set("Range", "bytes=0-1")
	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	switch resp.StatusCode {
	case http.StatusForbidden, http.StatusNotFound, http.StatusGone:
		return true
	default:
		return false
	}
}

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

	// A rotated-out player host answers /video/<hash> with a 404 page. Without
	// this guard we would still parse the hash out of the final URL and hand a
	// dead (host, hash) to the cache and getVideo, surfacing the upstream 404 to
	// the player instead of failing over to the next source.
	if resp2.StatusCode >= 400 {
		return "", "", "", fmt.Errorf("player page dead (%d): %s", resp2.StatusCode, finalURL)
	}

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

// fetchPlayerExtras loads the external player page for (host, hash) and pulls out
// the HLS audio-track languages and the external subtitle tracks.
//
// This exists because the browser path — the one production actually takes — used
// to return a stream with NO subtitles and NO audio info at all: it sniffs the
// media URL out of the network traffic and never reads the player page that
// carries them. The player host is not behind the Cloudflare gate (that is why
// GetVideoAPI can replay over plain HTTP), so recovering them costs one ordinary
// GET and no browser.
//
// Failure is non-fatal: extras enrich playback (subtitle tracks, dub-vs-original
// audio) but the stream plays without them, so callers ignore the error.
func (c *SuperFlixClient) fetchPlayerExtras(ctx context.Context, host, hash string) (defaultAudio []string, subtitles []SuperFlixSubtitle) {
	if host == "" || hash == "" || isNativePlayerHost(host) {
		return nil, nil
	}
	// Extras are optional and some rotating player hosts retire /video/<hash>
	// immediately after issuing the signed HLS URL.  Their tombstone page can
	// look like a Cloudflare/restricted shell to cfFallbackTransport; allowing
	// the normal escalation here opens that dead URL in the headed browser even
	// though the stream was already resolved successfully.  Keep this best-effort
	// enrichment HTTP-only and simply continue without tracks when it is gone.
	ctx = WithoutBrowserSolve(ctx)
	pageURL := host + "/video/" + hash

	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		util.Debug("SuperFlix: player extras request failed", "url", pageURL, "err", err)
		return nil, nil
	}
	req.Header.Set("Referer", host+"/")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		util.Debug("SuperFlix: player extras fetch failed", "url", pageURL, "err", err)
		return nil, nil
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		util.Debug("SuperFlix: player extras read failed", "url", pageURL, "err", err)
		return nil, nil
	}

	defaultAudio, subtitles = c.ExtractPlayerExtras(string(body))
	util.Debug("SuperFlix player extras", "audio", defaultAudio, "subtitles", len(subtitles))
	return defaultAudio, subtitles
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

	var result getVideoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", fmt.Errorf("failed to decode video API response: %w", err)
	}

	streamURL = preferredGetVideoURL(result)
	if streamURL == "" {
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
		if errors.Is(err, ErrSuperFlixRestricted) {
			break // terminal: the content is access-restricted, retrying won't help
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
	if res, ok := c.streamFromCache(ctx, key); ok {
		return res, nil
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
	referer := res.Referer
	if referer == "" {
		referer = "https://" + SuperFlixEmbedHost + "/"
	}

	// The browser can capture a signed URL for a player host that has already
	// rotated out of the CDN — getVideo still signs on a dead host, so mpv would
	// launch and then 404 a few seconds in. The cache path already probes for
	// exactly this (see streamFromCache); the fresh path used to skip it. Probe
	// here so a dead capture fails over to the next source instead of surfacing
	// the 404 to the player, and so we never cache a doomed (host, hash).
	if res.StreamURL != "" && c.streamURLDead(WithoutBrowserSolve(ctx), res.StreamURL, referer) {
		return nil, fmt.Errorf("superflix sniffed a dead stream host (%s)", res.PlayerHost)
	}

	// The sniff only yields the media URL; the subtitle tracks and the HLS audio
	// languages live on the player page, which is not gated. Fetch them so the
	// browser path is as rich as the plain-HTTP one (it used to return neither).
	audio, subs := c.fetchPlayerExtras(ctx, res.PlayerHost, res.VideoHash)
	// ALWAYS cache the (host, hash) — it is the browser-gated fact that makes the
	// next play skip the solve entirely. Gating the write on extras (which are a
	// bonus and can transiently fail to fetch) meant a hiccup there forced a full
	// re-solve on every subsequent play, and left a stale entry in place. Attach the
	// extras only when we actually got them.
	if res.PlayerHost != "" && res.VideoHash != "" {
		ent := streamCacheEntry{Host: res.PlayerHost, Hash: res.VideoHash}
		if len(audio) > 0 || len(subs) > 0 {
			ent.DefaultAudio = audio
			ent.Subtitles = subs
			ent.ExtrasCached = true
		}
		defaultStreamCache.put(key, ent)
	}

	return &SuperFlixStreamResult{
		StreamURL:    res.StreamURL,
		Referer:      referer,
		DefaultAudio: audio,
		Subtitles:    subs,
	}, nil
}
