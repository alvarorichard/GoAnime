package superflix

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// The FirePlayer host replaced its /player/index.php?do=getVideo endpoint with
// a POST to /layer/<key>/<hash>/. Same job, same JSON shape (videoSource /
// securedLink / videoImage), but the path carries a key that rotates per
// session, so it has to be read from the player's own script bundle first.
//
// This is what makes a repeat play browser-free again. getVideo now answers 404
// (403 before the header contract), so the cached (host, hash) pair could not be
// replayed and every play fell through to a ~7s Cloudflare solve. With the layer
// call the cached pair signs a fresh URL over plain HTTP in two requests —
// and unlike a cached signed URL, it never expires.
//
// Confirmed live 2026-09-01 against xn--tckasiu6cvova0eb5fua2449g98vg.best:
// the flow answers 200 with the client's own User-Agent, needs no Cloudflare
// clearance cookie, and does not require loading the player page first. The
// signed URL it returns plays under the same CDN header contract as one
// captured from the browser.
const (
	// sfPlayerScriptPath serves the player bundle that carries the current
	// layer key.
	sfPlayerScriptPath = "/player/assets/scripts.php?v=7"
	// sfLayerBudget caps the two-request signing flow. It runs on the play path
	// in place of a browser solve, so it must fail fast enough to leave time
	// for that fallback.
	sfLayerBudget = 20 * time.Second
)

// sfLayerKeyRe extracts the rotating key out of the player bundle. The observed
// shape is 32 hex + 8 hex + 22 base64url characters, but only the length and
// alphabet are pinned here — the internal structure is not ours to rely on.
var sfLayerKeyRe = regexp.MustCompile(`/layer/([A-Za-z0-9_-]{40,128})/`)

// fetchLayerKey reads the current layer key from the player's script bundle.
func (c *SuperFlixClient) fetchLayerKey(ctx context.Context, playerBaseURL, referer string) (string, error) {
	scriptURL := strings.TrimSuffix(playerBaseURL, "/") + sfPlayerScriptPath

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, scriptURL, http.NoBody)
	if err != nil {
		return "", err
	}
	applyCDNPlaybackHeaders(req, referer, c.effectiveUserAgent())

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("player script fetch failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("player script returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return "", fmt.Errorf("player script read failed: %w", err)
	}

	m := sfLayerKeyRe.FindSubmatch(body)
	if m == nil {
		return "", fmt.Errorf("no layer key in the player script (%s) — the contract may have moved again", scriptURL)
	}
	return string(m[1]), nil
}

// getVideoViaLayer signs a fresh stream URL through the /layer/ endpoint.
func (c *SuperFlixClient) getVideoViaLayer(ctx context.Context, playerBaseURL, videoHash, referer string) (streamURL, thumbURL string, err error) {
	ctx, cancel := context.WithTimeout(ctx, sfLayerBudget)
	defer cancel()

	key, err := c.fetchLayerKey(ctx, playerBaseURL, referer)
	if err != nil {
		return "", "", err
	}

	base := strings.TrimSuffix(playerBaseURL, "/")
	layerURL := fmt.Sprintf("%s/layer/%s/%s/", base, key, videoHash)

	form := url.Values{
		"hash": {videoHash},
		"r":    {c.base() + "/"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, layerURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", err
	}
	applyCDNPlaybackHeaders(req, referer, c.effectiveUserAgent())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", base)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("layer request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return "", "", fmt.Errorf("layer read failed: %w", err)
	}
	if err := ensureJSONResponse("layer API", resp, body); err != nil {
		return "", "", err
	}

	var result getVideoResponse
	if err := jsonx.Unmarshal(body, &result); err != nil {
		return "", "", fmt.Errorf("failed to decode layer response: %w", err)
	}
	streamURL = preferredGetVideoURL(result)
	if streamURL == "" {
		return "", "", fmt.Errorf("no stream URL in layer response")
	}
	return streamURL, result.VideoImage, nil
}

// signStreamURL obtains a freshly signed stream URL for a cached (host, hash)
// pair over plain HTTP — no browser.
//
// The current /layer/ contract is tried first because it is the one the live
// player uses; the legacy getVideo endpoint is kept as a fallback for any
// player host still serving it, and because it is a single request when it
// works. Both return the same JSON shape.
func (c *SuperFlixClient) signStreamURL(ctx context.Context, playerBaseURL, videoHash, referer string) (streamURL, thumbURL string, err error) {
	streamURL, thumbURL, err = c.getVideoViaLayer(ctx, playerBaseURL, videoHash, referer)
	if err == nil {
		return streamURL, thumbURL, nil
	}
	util.Debug("SuperFlix: layer signing failed, trying the legacy getVideo endpoint", "err", err)

	streamURL, thumbURL, legacyErr := c.GetVideoAPI(ctx, playerBaseURL, videoHash, referer)
	if legacyErr != nil {
		// Report the layer failure: it is the contract the live player uses, so
		// it is the one worth acting on when both are gone.
		return "", "", fmt.Errorf("layer signing failed (%w); legacy getVideo also failed: %v", err, legacyErr)
	}
	return streamURL, thumbURL, nil
}
