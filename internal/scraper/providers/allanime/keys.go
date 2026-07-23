// Package allanime — dynamic per-epoch key derivation for the AllAnime API.
//
// Since ani-cli PR #1779 AllAnime stopped shipping a static AES key. The key
// now rotates with a server "epoch" and must be reconstructed on the client:
//
//  1. GET the referer page (mkissa.to); it embeds `"epoch":<n>` and
//     `"partB":"<base64(32 bytes)>"`, plus a reference to the entry bundle
//     `<cdn>/entry/app.<hash>.js`.
//  2. GET that entry bundle; it imports a handful of code-split chunks
//     (`../chunks/<name>.js`). One of the first few embeds a 64-hex "mask".
//  3. key = mask XOR partB (byte-for-byte, both 32 bytes).
//
// The same key both signs the aaReq token and decrypts the `tobeparsed` blob,
// so a stale key breaks every episode-source query.
package allanime

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
)

// aaKeys is the per-epoch secret material AllAnime's API requires: a 32-byte
// AES-256 key plus the numeric epoch. Both are bound into the aaReq token, and
// the key also decrypts the `tobeparsed` response (ani-cli PR #1779).
type aaKeys struct {
	key   []byte
	epoch string
}

// aaKeysTTL bounds how long a derived key is reused before a refetch. The epoch
// rotates server-side; a few minutes keeps requests valid while avoiding a
// bundle re-scrape on every episode.
const aaKeysTTL = 3 * time.Minute

// Pre-compiled scrapers for the key-derivation flow.
var (
	aaEpochRe = regexp.MustCompile(`"epoch":(\d+)`)
	aaPartBRe = regexp.MustCompile(`"partB":"([^"]*)"`)
	// aaAppRe matches the entry bundle URL host-agnostically. Deriving the CDN
	// root from this match (rather than pinning allAnimeCDN) keeps derivation
	// working across CDN host rotations (ani-cli PR #1779 moved it once already).
	aaAppRe   = regexp.MustCompile(`https?://[^"'\s]+/entry/app\.[A-Za-z0-9_.-]+\.js`)
	aaChunkRe = regexp.MustCompile(`"\.\./chunks/[A-Za-z0-9_.-]+\.js"`)
	aaMaskRe  = regexp.MustCompile(`[0-9a-f]{64}`)
)

// getAAKeys returns the cached per-epoch key, deriving (and caching) a fresh one
// when absent or expired. Safe for concurrent callers.
func (c *AllAnimeClient) getAAKeys() (*aaKeys, error) {
	c.keyMu.Lock()
	defer c.keyMu.Unlock()

	if c.keys != nil && time.Now().Before(c.keysExp) {
		return c.keys, nil
	}

	keys, err := c.fetchAAKeys()
	if err != nil {
		return nil, err
	}
	c.keys = keys
	c.keysExp = time.Now().Add(aaKeysTTL)
	return keys, nil
}

// fetchAAKeys performs the full scrape-and-XOR key derivation described in the
// package doc. It issues live HTTP requests and must not run under tests
// (test clients inject a fixture key via NewClientForTest/newTestClient).
func (c *AllAnimeClient) fetchAAKeys() (*aaKeys, error) {
	page, err := c.getText(c.referer)
	if err != nil {
		return nil, fmt.Errorf("aaKeys: fetch referer page: %w", err)
	}

	epochM := aaEpochRe.FindStringSubmatch(page)
	partBM := aaPartBRe.FindStringSubmatch(page)
	appURL := aaAppRe.FindString(page)
	if len(epochM) < 2 || len(partBM) < 2 || appURL == "" {
		return nil, fmt.Errorf("aaKeys: epoch/partB/app bundle not found on %s", c.referer)
	}

	app, err := c.getText(appURL)
	if err != nil {
		return nil, fmt.Errorf("aaKeys: fetch app bundle: %w", err)
	}

	// Chunks are "../chunks/<name>.js" relative to the "/entry/" bundle, so the
	// CDN root is the appURL prefix before "/entry/". Derive it from the page
	// instead of the allAnimeCDN const so a CDN host rotation can't strand us.
	cdnRoot, _, ok := strings.Cut(appURL, "/entry/")
	if !ok {
		return nil, fmt.Errorf("aaKeys: app bundle URL missing /entry/ segment: %s", appURL)
	}

	// The mask lives in one of the first few code-split chunks the entry imports.
	maskHex := ""
	for i, m := range aaChunkRe.FindAllString(app, -1) {
		if i >= 5 {
			break
		}
		chunk := strings.TrimPrefix(strings.Trim(m, `"`), "../")
		js, err := c.getText(cdnRoot + "/" + chunk)
		if err != nil {
			continue
		}
		if hit := aaMaskRe.FindString(js); hit != "" {
			maskHex = hit
			break
		}
	}
	if maskHex == "" {
		return nil, fmt.Errorf("aaKeys: mask not found in %s chunks", cdnRoot)
	}

	mask, err := hex.DecodeString(maskHex)
	if err != nil || len(mask) != 32 {
		return nil, fmt.Errorf("aaKeys: bad mask hex (len=%d): %w", len(mask), err)
	}
	partB, err := base64.StdEncoding.DecodeString(partBM[1])
	if err != nil || len(partB) != 32 {
		return nil, fmt.Errorf("aaKeys: bad partB (len=%d): %w", len(partB), err)
	}

	key := make([]byte, 32)
	for i := range key {
		key[i] = mask[i] ^ partB[i]
	}
	util.Debugf("AllAnime derived aaKey (epoch=%s): %x", epochM[1], key)
	return &aaKeys{key: key, epoch: epochM[1]}, nil
}

// getText fetches a URL and returns its body as a string, capped at 8 MiB. Used
// for the referer page and CDN JS bundles during key derivation.
func (c *AllAnimeClient) getText(rawURL string) (string, error) {
	req, err := http.NewRequest("GET", rawURL, http.NoBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Referer", c.referer)

	resp, err := c.client.Do(req) // #nosec G704
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: status %d", rawURL, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024))
	if err != nil {
		return "", err
	}
	return string(body), nil
}
