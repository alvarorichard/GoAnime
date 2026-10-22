package startflix

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// upns.go resolves the panel's upn player family — the second Dublado embed
// StartFlix lists for movies (live 2026-10-21: "Velozes e Furiosos" (2001,
// tt0232500) listed embedplayapiupn.upns.xyz/#8txjhj as Dublado #2). Until
// then the family fell through ResolveStreamFor's yt-dlp sweep only to get
// burned there, because the player is a Vite SPA (assets/index-*.js) whose
// whole API is encrypted.
//
// The wire protocol was reverse-engineered from that bundle on 2026-10-21
// (string table resolved: obfuscator.io accessor be(n) = tbl[(n-254-259)%550]
// after its rotate-until checksum loop, confirmed against the SPA's own
// runtime by hooking TextEncoder/TextDecoder, which the app builds through
// `new window[N]`):
//
//	GET {base}/api/v1/info?id={key}                  → hex(AES-CBC(meta JSON))
//	GET {base}/api/v1/video?id={key}&w={px}&h={px}&r={embedderHostNoWWW}
//	                                                 → hex(AES-CBC(config JSON))
//
// with a fixed AES-128-CBC key and IV the SPA builds programmatically and
// the runtime captures named verbatim. The video config carries the sources
// as strings or {url, params, domain} objects keyed "cf" (Cloudflare),
// "cfNative", "hlsVideoGoogle", "hlsVideoTiktok" and "source" (In-House),
// ordered by streamingConfig.order under the same names; delivery.inHouse
// "saturated" makes the player re-fetch with its capacityToken pair, so the
// resolver does one such refresh too. All sources are HLS (the player marks
// every one application/x-mpegurl), so the master prober validates them and
// reads their height, the way the vidsrc chain does.
//
// Known state: the only key the panel ever showed for this id answers
// "Video not found or deleted" on /api/v1/video — verified twice live through
// the SPA's own authentic flow (2026-10-21), so the Dublado entry is dead at
// the source, not merely unsupported. The resolver still ships because the
// panel keys are stable (re-fetched fresh after an hour, same #8txjhj) and
// this family will serve other titles' audio tracks when their videos exist;
// the fake-server tests pin the protocol so a live healthy id drops in raw.
//
// Note the fragment key is NOT ephemeral and /api/v1/info still decrypts for
// this dead id — only the video stage names the message, which is why it is
// surfaced verbatim: "not found or deleted" means the panel's entry is stale.

const (
	// upnsAESKey / upnsAESIV are the player family's fixed AES-128-CBC
	// secret pair, captured at runtime on 2026-10-21 (the SPA assembles
	// both from an obfuscated char-code arithmetic, but they are constants
	// per bundle).
	upnsAESKey = "kiemtienmua911ca"
	upnsAESIV  = "1234567890oiuytr"

	// upnsViewport is the w/h the resolver reports — the SPA sends the
	// browser's innerWidth/screen.height; over the wire any plausible
	// viewport works (tested live: the endpoint does not tie the id to it).
	upnsViewportW = 1280
	upnsViewportH = 720

	upnsBodyLimit = 64 << 10
)

var upnsHTTPOnce sync.Once
var upnsHTTP *http.Client

// upnsHTTPDefault builds the family's client: the same fast
// Chrome-impersonating client the rest of the chain uses — the upn API
// decloaks non-browser fingerprints (a plain curl's HTTP fetch returns an
// opaque decoy body of the same length; the SPA's fetch returned the real
// hex ciphertext live on 2026-10-21), so TLS-grade impersonation is the
// entry fee.
var upnsHTTPDefault = func() *http.Client {
	upnsHTTPOnce.Do(func() { upnsHTTP = util.NewFastClient() })
	return upnsHTTP
}

// upnsHTTPFn returns the client the resolver runs on; a seam so tests route
// it through one httptest server, as the vidsrc tests do.
var upnsHTTPFn = upnsHTTPDefault

// isUPNsPanelPlayer reports whether a panel player URL is one this file can
// key: any host on upns.xyz carrying a key in its fragment (the SPA reads
// the same fragment) or an ?id= query. Everything else stays unsupported.
func isUPNsPanelPlayer(u *url.URL) bool {
	if u == nil || !strings.HasSuffix(u.Hostname(), "upns.xyz") {
		return false
	}
	_, ok := upnsPanelKey(u)
	return ok
}

// upnsPanelKey reads the video key the player URL names — the fragment for
// the embed form seen live (…/#8txjhj), the ?id= query as a courtesy form.
func upnsPanelKey(u *url.URL) (string, bool) {
	if u == nil {
		return "", false
	}
	if k := strings.Trim(u.Fragment, "/ "); k != "" {
		return k, true
	}
	if k := u.Query().Get("id"); k != "" {
		return k, true
	}
	return "", false
}

// upnsVideoConfig is the decrypted /api/v1/video shape: the sources under
// their four keys (strings or {url, params, domain} objects), the order the
// player walks them in, and the delivery state that names saturation.
type upnsVideoConfig struct {
	Player bool   `json:"player"`
	Title  string `json:"title"`

	CF       any `json:"cf"`
	CFNative any `json:"cfNative"`
	Google   any `json:"hlsVideoGoogle"`
	TikTok   any `json:"hlsVideoTiktok"`
	Source   any `json:"source"`

	StreamingConfig *struct {
		Order  []string                  `json:"order"`
		Adjust map[string]map[string]any `json:"adjust"`
	} `json:"streamingConfig"`

	Delivery *struct {
		InHouse             string `json:"inHouse"`
		CapacityToken       string `json:"capacityToken"`
		CapacityTokenExpire string `json:"capacityTokenExpire"`
	} `json:"delivery"`
}

// upnsSourceOrder is the player's fallback crawl order when the config
// carries no streamingConfig.order: the four keys in iu()'s literal.
var upnsSourceOrder = []string{"Cloudflare", "Tiktok", "Google", "In-House"}

// upnsCandidates lists the HLS master URLs the config offers, in the order
// the player crawls them. Cloudflare has a native-HLS twin the player uses
// whenever the SPA cannot lift hls.js — our player is native too, so the
// native form is the preferred statement of that key.
func upnsCandidates(cfg *upnsVideoConfig) []string {
	names := upnsSourceOrder
	var adjust map[string]map[string]any
	if cfg.StreamingConfig != nil {
		if len(cfg.StreamingConfig.Order) > 0 {
			names = cfg.StreamingConfig.Order
		}
		adjust = cfg.StreamingConfig.Adjust
	}
	sourceFor := func(name string) any {
		switch name {
		case "Cloudflare":
			// cfNative is the player's own native-HLS twin of cf (the SPA
			// uses it when hls.js is absent — ours is a native player too).
			return firstAnyNonEmpty(cfg.CFNative, cfg.CF)
		case "Tiktok":
			return cfg.TikTok
		case "Google":
			return cfg.Google
		case "In-House":
			return cfg.Source
		}
		return nil
	}
	var out []string
	seen := map[string]bool{}
	for _, name := range names {
		v := sourceFor(name)
		if v == nil {
			continue
		}
		if s, ok := upnsSourceURL(v, adjust[name]); ok && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// upnsSourceURL turns one config source into a plain URL. The player's tu()
// accepts a string as-is, or an object carrying url plus optional params
// (merged into the query) and domain (rewrites an /hls/ path to
// /hlsmod/{domain}/); the per-key adjust entry from streamingConfig supplies
// more of the same for the object form.
func upnsSourceURL(v, adjust any) (string, bool) {
	obj, ok := v.(map[string]any)
	if ok {
		raw := anyToString(obj["url"])
		if raw == "" {
			return "", false
		}
		u, err := url.Parse(raw)
		if err != nil {
			return "", false
		}
		params := map[string]any{}
		domain := ""
		for _, layer := range []map[string]any{obj, toStringMap(adjust)} {
			if layer == nil {
				continue
			}
			if dm := anyToString(layer["domain"]); dm != "" && domain == "" {
				domain = dm
			}
			for k, pv := range toStringMap(layer["params"]) {
				if _, dup := params[k]; !dup {
					params[k] = pv
				}
			}
		}
		q := u.Query()
		for k, pv := range params {
			if _, dup := q[k]; !dup {
				q.Set(k, anyToString(pv))
			}
		}
		u.RawQuery = q.Encode()
		if domain != "" && strings.Contains(u.Path, "/hls/") {
			u.Path = strings.Replace(u.Path, "/hls/", "/hlsmod/"+domain+"/", 1)
		}
		return u.String(), true
	}
	s := anyToString(v)
	return s, s != ""
}

func anyToString(v any) string {
	s, _ := v.(string)
	return s
}

func firstAnyNonEmpty(vals ...any) any {
	for _, v := range vals {
		if anyToString(v) != "" || len(toStringMap(v)) > 0 {
			return v
		}
	}
	return nil
}

func toStringMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

// upnsDecryptHex decodes the family's response body: a hex string of
// AES-128-CBC ciphertext with the fixed key and IV, PKCS#7-padded JSON.
func upnsDecryptHex(hexBody string) ([]byte, error) {
	data, err := hex.DecodeString(strings.TrimSpace(hexBody))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext length %d is not AES-blocked", len(data))
	}
	block, err := aes.NewCipher([]byte(upnsAESKey))
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(data))
	cipher.NewCBCDecrypter(block, []byte(upnsAESIV)).CryptBlocks(out, data)
	pad := int(out[len(out)-1])
	if pad == 0 || pad > aes.BlockSize || pad > len(out) {
		return nil, fmt.Errorf("bad PKCS#7 padding")
	}
	for _, b := range out[len(out)-pad:] {
		if int(b) != pad {
			return nil, fmt.Errorf("bad PKCS#7 padding")
		}
	}
	return out[:len(out)-pad], nil
}

// upnsEmbedderHost names the site the panel embeds the player from, the
// value the SPA derives from document.referrer and sends as &r= (its ne()
// strips the www.). It is advisory, not authenticating — live the endpoint
// answered readers with and without it the same way.
func upnsEmbedderHost() string {
	u, err := url.Parse(panelBase())
	if err != nil || u.Hostname() == "" {
		return ""
	}
	return strings.TrimPrefix(u.Hostname(), "www.")
}

func upnsVideoURL(base, key, extra string) string {
	// The SPA always appends &r= even when its referrer read is empty, so
	// the empty form is part of the family's contract.
	return fmt.Sprintf("%s/api/v1/video?id=%s&w=%d&h=%d&r=%s%s",
		base, url.QueryEscape(key), upnsViewportW, upnsViewportH,
		url.QueryEscape(upnsEmbedderHost()), extra)
}

func upnsGetBody(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Referer", originOf(rawURL)+"/")
	resp, err := upnsHTTPFn().Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, upnsBodyLimit))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		var env struct {
			Message string `json:"message"`
		}
		if jsonx.Unmarshal(body, &env) == nil && env.Message != "" {
			return nil, fmt.Errorf("upn: %s", env.Message)
		}
		return nil, fmt.Errorf("upn: HTTP %d", resp.StatusCode)
	}
	return body, nil
}

func upnsFetchConfig(ctx context.Context, base, key, extra string) (*upnsVideoConfig, error) {
	body, err := upnsGetBody(ctx, upnsVideoURL(base, key, extra))
	if err != nil {
		return nil, err
	}
	plain, err := upnsDecryptHex(string(body))
	if err != nil {
		// A non-hex 200 (proxy injection, contract change) is not readable.
		return nil, fmt.Errorf("upn: unreadable response: %w", err)
	}
	var cfg upnsVideoConfig
	if err := jsonx.Unmarshal(plain, &cfg); err != nil {
		return nil, fmt.Errorf("upn: decrypted payload is not the player config: %w", err)
	}
	return &cfg, nil
}

// resolveUPNsPanelPlayer resolves one upn player URL end to end: decrypt the
// video config, walk the sources in the config's own order, and hand back
// the tallest master that actually loads. The stream's audio is left empty —
// the panel labels its upn players like every other family label, and
// ResolveStreamFor applies that label (this file has nothing truthful to add
// about the tracks).
func resolveUPNsPanelPlayer(ctx context.Context, u *url.URL) (*Stream, error) {
	key, ok := upnsPanelKey(u)
	if !ok {
		return nil, fmt.Errorf("upn: without a key: %s", u.RequestURI())
	}
	base := u.Scheme + "://" + u.Host
	cfg, err := upnsFetchConfig(ctx, base, key, "")
	if err != nil {
		return nil, err
	}
	// delivery.inHouse "saturated" is the family's own "retry me" contract:
	// refresh once with the capacity token pair the config carried.
	if d := cfg.Delivery; d != nil && d.InHouse == "saturated" &&
		d.CapacityToken != "" && d.CapacityTokenExpire != "" {
		extra := "&capacityToken=" + url.QueryEscape(d.CapacityToken) +
			"&capacityTokenExpire=" + url.QueryEscape(d.CapacityTokenExpire)
		if retry, err := upnsFetchConfig(ctx, base, key, extra); err == nil {
			cfg = retry
		}
	}
	if !cfg.Player {
		return nil, fmt.Errorf("upn: %q is not ready yet", key)
	}
	candidates := upnsCandidates(cfg)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("upn: no sources listed for %q", key)
	}
	referer := masterReferer(base + "/")
	best := -1
	bestHeight := 0
	for i, raw := range candidates {
		_, height := probeVIDSRCMaster(ctx, raw, referer)
		if height > bestHeight {
			best, bestHeight = i, height
		}
	}
	if best < 0 {
		return nil, fmt.Errorf("upn: %d source(s) listed, none playable", len(candidates))
	}
	return &Stream{
		URL:     candidates[best],
		Referer: referer,
		Host:    u.Host,
		Height:  bestHeight,
		Audio:   "",
	}, nil
}
