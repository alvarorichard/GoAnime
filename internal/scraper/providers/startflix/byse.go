package startflix

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// Byse is the player behind StartFlix's embedplaybyse.top servers, and the one
// filemoon.sx now serves too (both answer "Byse Frontend" with the same bundle,
// checked 2026-10-07). It is the only host on the panel that resolves over
// plain HTTP: GET /api/videos/<code> returns the stream list sealed with
// AES-256-GCM, and the key ships in the same response.
//
// What the bundle does with it (videoPagesBundle, version "6" on 2026-10-07):
// key_parts holds 30 base64url strings, most of them decoys. For a version v in
// 1..20 the key is parts v and 31-v (1-based) concatenated; for any other
// version every part is used. The plaintext is JSON with a "sources" array of
// HLS masters. The CDN serves those to any client — no Referer, no cookie, no
// User-Agent binding.

var byseEmbedPathRe = regexp.MustCompile(`^/(?:e|d|v|download)/([A-Za-z0-9]{6,})`)

// isByseHost reports whether a player host runs the Byse frontend.
func isByseHost(host string) bool {
	host = strings.ToLower(host)
	return strings.Contains(host, "byse") || strings.Contains(host, "filemoon")
}

// byseCode extracts the video code from a Byse embed URL.
func byseCode(u *url.URL) (string, bool) {
	m := byseEmbedPathRe.FindStringSubmatch(u.Path)
	if m == nil {
		return "", false
	}
	return m[1], true
}

type byseVideo struct {
	Error    string         `json:"error"`
	Title    string         `json:"title"`
	Playback *bysePlayback  `json:"playback"`
	Tracks   []byseTrackRaw `json:"tracks"`
}

type bysePlayback struct {
	Algorithm string   `json:"algorithm"`
	IV        string   `json:"iv"`
	Payload   string   `json:"payload"`
	KeyParts  []string `json:"key_parts"`
	Version   string   `json:"version"`
}

type byseSource struct {
	URL      string `json:"url"`
	Label    string `json:"label"`
	MimeType string `json:"mime_type"`
	Height   int    `json:"height"`
}

// byseTrackRaw is a subtitle track. No sample on 2026-10-07 carried one, so
// every plausible field name is accepted rather than guessing one.
type byseTrackRaw struct {
	URL      string `json:"url"`
	File     string `json:"file"`
	Src      string `json:"src"`
	Label    string `json:"label"`
	Language string `json:"language"`
	Lang     string `json:"lang"`
	Kind     string `json:"kind"`
}

type byseSources struct {
	Sources []byseSource   `json:"sources"`
	Tracks  []byseTrackRaw `json:"tracks"`
}

// byseKey selects and joins the key parts the way the player bundle does.
func byseKey(pb *bysePlayback) ([]byte, error) {
	parts := pb.KeyParts
	if v, err := strconv.Atoi(strings.TrimSpace(pb.Version)); err == nil && v >= 1 && v <= 20 && 31-v <= len(parts) {
		parts = []string{parts[v-1], parts[30-v]}
	}
	var key []byte
	for _, p := range parts {
		if p == "" {
			continue
		}
		b, err := decodeBase64URL(p)
		if err != nil {
			return nil, fmt.Errorf("key part: %w", err)
		}
		key = append(key, b...)
	}
	switch len(key) {
	case 16, 24, 32:
		return key, nil
	default:
		return nil, fmt.Errorf("unexpected key length %d", len(key))
	}
}

// decryptBysePlayback opens the sealed stream list.
func decryptBysePlayback(pb *bysePlayback) (*byseSources, error) {
	if pb == nil {
		return nil, errors.New("no playback payload")
	}
	if a := strings.ToUpper(pb.Algorithm); a != "" && !strings.Contains(a, "GCM") {
		return nil, fmt.Errorf("unsupported algorithm %q", pb.Algorithm)
	}
	key, err := byseKey(pb)
	if err != nil {
		return nil, err
	}
	iv, err := decodeBase64URL(pb.IV)
	if err != nil {
		return nil, fmt.Errorf("iv: %w", err)
	}
	payload, err := decodeBase64URL(pb.Payload)
	if err != nil {
		return nil, fmt.Errorf("payload: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(iv) < 8 {
		return nil, fmt.Errorf("iv too short (%d bytes)", len(iv))
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, len(iv))
	if err != nil {
		return nil, err
	}
	plain, err := gcm.Open(nil, iv, payload, nil)
	if err != nil {
		return nil, err
	}
	var out byseSources
	if err := jsonx.Unmarshal(plain, &out); err != nil {
		return nil, fmt.Errorf("decrypted payload: %w", err)
	}
	return &out, nil
}

func decodeBase64URL(s string) ([]byte, error) {
	s = strings.TrimRight(strings.TrimSpace(s), "=")
	s = strings.NewReplacer("+", "-", "/", "_").Replace(s)
	return base64.RawURLEncoding.DecodeString(s)
}

// pickByseSource prefers the tallest rendition; a master playlist carries its
// own variants, so this mostly matters when a video lists several files.
func pickByseSource(sources []byseSource) (byseSource, bool) {
	playable := make([]byseSource, 0, len(sources))
	for _, s := range sources {
		if strings.HasPrefix(s.URL, "https://") || strings.HasPrefix(s.URL, "http://") {
			playable = append(playable, s)
		}
	}
	if len(playable) == 0 {
		return byseSource{}, false
	}
	sort.SliceStable(playable, func(i, j int) bool { return playable[i].Height > playable[j].Height })
	return playable[0], true
}

func byseSubtitles(tracks ...[]byseTrackRaw) []Subtitle {
	var out []Subtitle
	seen := map[string]bool{}
	for _, list := range tracks {
		for _, t := range list {
			if k := strings.ToLower(t.Kind); k != "" && k != "captions" && k != "subtitles" {
				continue
			}
			u := firstNonEmpty(t.URL, t.File, t.Src)
			if u == "" || seen[u] {
				continue
			}
			seen[u] = true
			lang := strings.ToLower(firstNonEmpty(t.Language, t.Lang, t.Label))
			out = append(out, Subtitle{URL: u, Language: lang, Label: firstNonEmpty(t.Label, t.Language, t.Lang)})
		}
	}
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v = strings.TrimSpace(v); v != "" {
			return v
		}
	}
	return ""
}

// resolveByse turns a Byse embed URL into its HLS master.
func (c *Client) resolveByse(ctx context.Context, embed *url.URL) (*Stream, error) {
	code, ok := byseCode(embed)
	if !ok {
		return nil, netx.NewParserError(SourceName, "byse", "no video code in embed URL", nil)
	}
	origin := embed.Scheme + "://" + embed.Host
	body, err := c.get(ctx, origin+"/api/videos/"+url.PathEscape(code), getOptions{
		referer: origin + "/e/" + code,
		accept:  "application/json",
		layer:   "byse",
		allow:   []int{404},
	})
	if err != nil {
		return nil, err
	}

	var video byseVideo
	if err := json.Unmarshal(body, &video); err != nil {
		return nil, netx.NewParserError(SourceName, "byse", "unreadable video record", err)
	}
	if video.Error != "" {
		return nil, fmt.Errorf("byse %s: %s", code, video.Error)
	}
	sources, err := decryptBysePlayback(video.Playback)
	if err != nil {
		return nil, netx.NewDecryptError(SourceName, "byse", "could not open the playback payload", err)
	}
	src, ok := pickByseSource(sources.Sources)
	if !ok {
		return nil, fmt.Errorf("byse %s: no playable source", code)
	}
	return &Stream{
		URL:       src.URL,
		Referer:   origin + "/",
		Subtitles: byseSubtitles(sources.Tracks, video.Tracks),
		Host:      embed.Host,
	}, nil
}
