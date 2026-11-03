package pomfy

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// The Byse player API on the CDN the embeds come from (pomfy-cdn.shop). This
// is the captcha-gated shape Pomfy serves; StartFlix's unguarded shape lives
// in the startflix package. Differences worth remembering:
//
//   - details/captcha/playback are /api/videos/<code>/embed/* endpoints and
//     read X-Embed-Origin/Referer/Parent headers naming the embedding site;
//   - POST /embed/captcha hands out a proof-of-work challenge (nonce,
//     difficulty, opaque token) whose solutions redeem at /embed/captcha/verify;
//   - POST /embed/playback then returns the AES-256-GCM sealed record under
//     the X-Captcha-Token header, and its declared key_part version — 19 in
//     samples from 2026-11 — picks the two parts of thirty that are the key.

const (
	// detailLayer names this flow in netx diagnostics.
	detailLayer = "byse"

	// powBudgetMost caps one solve: difficulty 17-18 answers in ~0.2-2s on a
	// desktop core, 180s keeps margin on thrice-slower hardware.
	powBudgetMost = 180 * time.Second

	// fingerprintConfidence is the constant the player bundle reports.
	fingerprintConfidence = 0.99

	// minHeight is the smallest rendition label counted as a height.
	minHeight = 144
)

// base64URLRaw decodes the seal's payload fields: unpadded base64url, with
// whatever "=" tails the bundles stringed in trimmed before the read.
var base64URLRaw = base64.RawURLEncoding

// EmbedTube is one resolution of an embed slot for a video code. Watch their
// fields: they fill the playback metadata contract (providers.
// applyPlaybackMetadata) almost as-is.
type Stream struct {
	URL       string
	Referer   string
	Host      string
	Height    int
	Subtitles []Subtitle
}

// Subtitle is one caption track the unsealed record lists.
type Subtitle struct {
	URL      string
	Language string
	Label    string
}

// byseVideo is what /embed/details returns.
type byseDetails struct {
	Error         string `json:"error"`
	ID            int64  `json:"id"`
	Code          string `json:"code"`
	Title         string `json:"title"`
	PosterURL     string `json:"poster_url"`
	EmbedFrameURL string `json:"embed_frame_url"`
}

// captchaChallenge is what POST /embed/captcha answers with.
type captchaChallenge struct {
	Error      string `json:"error"`
	PowNonce   string `json:"pow_nonce"`
	Difficulty int    `json:"pow_difficulty"`
	PowToken   string `json:"pow_token"`
	ExpiresIn  int    `json:"expires_in"`
}

// captchaVerify is what POST /embed/captcha/verify answers with.
type captchaVerify struct {
	Error  string `json:"error"`
	Status string `json:"status"`
	Token  string `json:"token"`
}

// playbackSeal is the sealed record blob the captcha token buys.
type playbackSeal struct {
	Algorithm string   `json:"algorithm"`
	IV        string   `json:"iv"`
	Payload   string   `json:"payload"`
	KeyParts  []string `json:"key_parts"`
	Version   string   `json:"version"`
	ExpiresAt string   `json:"expires_at"`
}

// playbackResponse wraps the seal.
type playbackResponse struct {
	Error    string       `json:"error"`
	Playback playbackSeal `json:"playback"`
}

// sealedSource is one unsealed rendition.
type sealedSource struct {
	// URL is the HLS master on the byse edge.
	URL      string `json:"url"`
	Label    string `json:"label"`
	Quality  string `json:"quality"`
	MimeType string `json:"mime_type"`
	Height   int    `json:"height"`
	Bitrate  int64  `json:"bitrate_kbps"`
}

// sealedRecord is the unsealed playback payload.
type sealedRecord struct {
	PosterURL   string         `json:"poster_url"`
	GeneratedAt string         `json:"generated_at"`
	Sources     []sealedSource `json:"sources"`
	Tracks      []byseTrack    `json:"tracks"`
}

// byseTrack accepts the subtitle track fields the players have used so far.
type byseTrack struct {
	URL      string `json:"url"`
	File     string `json:"file"`
	Src      string `json:"src"`
	Label    string `json:"label"`
	Language string `json:"language"`
	Lang     string `json:"lang"`
	Kind     string `json:"kind"`
}

// fingerprint is the identity payload the captcha and playback requests both
// carry. The bundle's value is browser-derived; every shape of that reached
// the server only as these four fields (checked 2026-11-02), so GoAnime fills
// them with random hex at the same lengths the bundle produces.
type fingerprint struct {
	Token      string  `json:"token"`
	ViewerID   string  `json:"viewer_id"`
	DeviceID   string  `json:"device_id"`
	Confidence float64 `json:"confidence"`
}

func newFingerprint() fingerprint {
	return fingerprint{
		Token:      randomHex(16),
		ViewerID:   randomHex(8),
		DeviceID:   randomHex(8),
		Confidence: fingerprintConfidence,
	}
}

// resolveByse runs the whole four-call flow for one video code and opens the
// seal. apiBase is "scheme://host" of the embed host — https://pomfy-cdn.shop
// in production, a test server URL in the flow test.
func resolveByse(ctx context.Context, c *Client, apiBase, code, watchPageURL string) (*Stream, error) {
	if code == "" {
		return nil, netx.NewParserError(SourceName, detailLayer, "no video code to resolve", nil)
	}
	parent := apiBase + "/e/" + code
	headers := map[string]string{
		"X-Embed-Origin":  watchPageOrigin(watchPageURL),
		"X-Embed-Referer": watchPageURL,
		"X-Embed-Parent":  parent,
	}

	// 1. details is metadata only; its absence decides nothing (the bundle
	//    treats a failed parse as an empty detail set).
	if body, err := c.postJSON(ctx, apiBase, "/api/videos/"+url.PathEscape(code)+"/embed/details", struct{}{}, headers); err != nil {
		util.Debug("byse details unavailable", "code", code, "err", err)
	} else {
		var det byseDetails
		if err := json.Unmarshal(body, &det); err != nil {
			util.Debug("byse details unreadable", "code", code, "err", err)
		} else if det.Error != "" {
			return nil, fmt.Errorf("byse %s: %s", code, det.Error)
		}
	}

	// 2. ask for the challenge.
	body, err := c.postJSON(ctx, apiBase, "/api/videos/"+url.PathEscape(code)+"/embed/captcha", struct{}{}, headers)
	if err != nil {
		return nil, netx.NewParserError(SourceName, detailLayer, "captcha challenge request failed", err)
	}
	var challenge captchaChallenge
	if err := json.Unmarshal(body, &challenge); err != nil {
		return nil, netx.NewParserError(SourceName, detailLayer, "captcha challenge unreadable", err)
	}
	if challenge.Error != "" {
		return nil, fmt.Errorf("byse %s captcha: %s", code, challenge.Error)
	}
	if challenge.PowNonce == "" || challenge.PowToken == "" || challenge.Difficulty <= 0 {
		return nil, fmt.Errorf("byse %s: captcha challenge incomplete", code)
	}

	// 3. solve the proof-of-work.
	solution, err := solveNonce(ctx, challenge.PowNonce+":", challenge.Difficulty, powBudgetMost)
	if err != nil {
		return nil, fmt.Errorf("byse %s: proof-of-work: %w", code, err)
	}

	// 4. redeem it.
	fp := newFingerprint()
	body, err = c.postJSON(ctx, apiBase, "/api/videos/"+url.PathEscape(code)+"/embed/captcha/verify", struct {
		PowToken    string      `json:"pow_token"`
		Solution    string      `json:"solution"`
		Fingerprint fingerprint `json:"fingerprint"`
	}{challenge.PowToken, solution, fp}, headers)
	if err != nil {
		return nil, netx.NewParserError(SourceName, detailLayer, "captcha verify failed", err)
	}
	var verify captchaVerify
	if err := json.Unmarshal(body, &verify); err != nil {
		return nil, netx.NewParserError(SourceName, detailLayer, "captcha verify answer unreadable", err)
	}
	if verify.Error != "" {
		return nil, fmt.Errorf("byse %s captcha verify: %s", code, verify.Error)
	}
	if verify.Status != "ok" || verify.Token == "" {
		return nil, fmt.Errorf("byse %s: captcha rejected the solution", code)
	}

	// 5. buy the sealed record with the captcha token.
	seedHeaders := map[string]string{}
	for k, v := range headers {
		seedHeaders[k] = v
	}
	seedHeaders["X-Captcha-Token"] = verify.Token
	body, err = c.postJSON(ctx, apiBase, "/api/videos/"+url.PathEscape(code)+"/embed/playback", struct {
		Fingerprint fingerprint `json:"fingerprint"`
	}{fp}, seedHeaders)
	if err != nil {
		return nil, netx.NewParserError(SourceName, detailLayer, "playback request failed", err)
	}
	var pb playbackResponse
	if err := json.Unmarshal(body, &pb); err != nil {
		return nil, netx.NewParserError(SourceName, detailLayer, "playback record unreadable", err)
	}

	// 6. open it.
	sources, err := decryptBysePlayback(&pb.Playback)
	if err != nil {
		return nil, netx.NewDecryptError(SourceName, detailLayer, "could not open the playback payload", err)
	}
	src, ok := pickSource(sources.Sources)
	if !ok {
		return nil, fmt.Errorf("byse %s: no playable source", code)
	}
	return &Stream{
		URL:       src.URL,
		Referer:   strings.TrimSuffix(watchPageURL, "/"),
		Host:      apiBase,
		Height:    src.height(),
		Subtitles: subtitlesFilter(sources.Tracks),
	}, nil
}

// watchPageOrigin returns the bare host the site sends in X-Embed-Origin
// ("pomfy.my"), or the default watching site when the URL does not parse.
func watchPageOrigin(raw string) string {
	if u, err := url.Parse(raw); err == nil && u.Host != "" {
		return u.Host
	}
	return DefaultBaseHost
}

// byseKey selects and concatenates the key parts the way the player bundle's
// reorderKeyParts does: version v in 1..20 with 30 parts means parts v and
// 31-v (1-based); anything else means every part. Decoded with padding and
// spelling tolerated, the result must be exactly 32 bytes for AES-256.
func byseKey(pb *playbackSeal) ([]byte, error) {
	parts := pb.KeyParts
	if v, err := strconv.Atoi(strings.TrimSpace(pb.Version)); err == nil && v >= 1 && v <= 20 &&
		len(parts) >= 30 && max(v, 31-v) <= len(parts) {
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
	if len(key) != 32 {
		return nil, fmt.Errorf("unexpected key length %d", len(key))
	}
	return key, nil
}

// decryptBysePlayback unseals the stream list: mean payload = AES-256-GCM
// over iv, whose 16-byte authentication tag rides the ciphertext's tail.
func decryptBysePlayback(pb *playbackSeal) (*sealedRecord, error) {
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

	// byseKey already validated the key at exactly 32 bytes, and the iv at
	// the shape the seal carries, so the cipher builds below cannot fail.
	block, _ := aes.NewCipher(key)
	const tagSize = 16
	if len(payload) <= tagSize {
		return nil, errors.New("payload too short to carry a GCM tag")
	}
	gcm, _ := cipher.NewGCMWithNonceSize(block, len(iv))
	plain, err := gcm.Open(nil, iv, payload, nil)
	if err != nil {
		return nil, err
	}
	var out sealedRecord
	if err := json.Unmarshal(plain, &out); err != nil {
		return nil, fmt.Errorf("decrypted payload: %w", err)
	}
	return &out, nil
}

// sealForTestFn builds the seal behind SealForTest: AES-256-GCM over a
// pointer-obvious key/iv pair, with the real key at parts 9 and 22. The
// crypto calls here cannot fail with these fixed shapes (the cipher builds
// a 32-byte key, the GCM a 12-byte nonce), so there is no error return.
func sealForTestFn(plaintext []byte) playbackSeal {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(0xA1 + i)
	}
	iv := make([]byte, 12)
	for i := range iv {
		iv[i] = byte(0x41 + i)
	}
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCMWithNonceSize(block, len(iv))
	sealed := gcm.Seal(nil, iv, plaintext, nil)
	parts := make([]string, 30)
	for i := range parts {
		parts[i] = base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%032x", i+1)))
	}
	parts[8] = base64.RawURLEncoding.EncodeToString(key[:16])
	parts[21] = base64.RawURLEncoding.EncodeToString(key[16:])
	return playbackSeal{
		Algorithm: "AES-256-GCM",
		IV:        base64.RawURLEncoding.EncodeToString(iv),
		Payload:   base64.RawURLEncoding.EncodeToString(sealed),
		KeyParts:  parts,
		Version:   "9",
	}
}

// decodeBase64URL decodes the standard_[url] alphabet with or without padding,
// forgiving the "=" tails bundles sometimes carry.
func decodeBase64URL(s string) ([]byte, error) {
	s = strings.TrimRight(strings.TrimSpace(s), "=")
	if s == "" {
		return nil, errors.New("empty value")
	}
	return base64URLRaw.DecodeString(s)
}

// pickSource prefers the tallest rendition — a master playlist carries its own
// variants, so this mostly matters when a video lists several language files.
// pickSource prefers the tallest rendition — a master playlist carries its
// own variants, so this mostly matters when a video lists several language
// files. Any http(s) master plays, labeled height or not (an unlabeled
// single rendition is the rounded "0"-live one).
func pickSource(sources []sealedSource) (sealedSource, bool) {
	best := -1
	var chosen sealedSource
	for _, s := range sources {
		u := s.URL
		if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
			continue
		}
		if h := s.height(); best == -1 || h > best {
			best, chosen = h, s
		}
	}
	return chosen, best >= 0
}

// height is the source's picture height: its height field, or the number in a
// "1080p"-style label when the field is missing.
func (s sealedSource) height() int {
	if s.Height > 0 {
		return s.Height
	}
	return labelHeight(s.Label)
}

// labelHeight reads the number in a "1080p"-style label; 0 when absent.
func labelHeight(label string) int {
	n, ok := heightDigits(label)
	if !ok || n < minHeight {
		return 0
	}
	return n
}

func heightDigits(label string) (int, bool) {
	i := strings.IndexFunc(label, func(r rune) bool { return r >= '0' && r <= '9' })
	if i < 0 {
		return 0, false
	}
	j := i
	for j < len(label) && label[j] >= '0' && label[j] <= '9' {
		j++
	}
	n, err := strconv.Atoi(label[i:j])
	if err != nil {
		return 0, false
	}
	return n, true
}

// subtitlesFilter converts raw track entries to the playback metadata shape.
func subtitlesFilter(tracks []byseTrack) []Subtitle {
	out := make([]Subtitle, 0, len(tracks))
	seen := make(map[string]bool, len(tracks))
	for _, t := range tracks {
		if k := strings.ToLower(t.Kind); k != "" && k != "captions" && k != "subtitles" {
			continue
		}
		u := firstNonEmpty(t.URL, t.File, t.Src)
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		out = append(out, Subtitle{
			URL:      u,
			Language: strings.ToLower(firstNonEmpty(t.Language, t.Lang, t.Label)),
			Label:    firstNonEmpty(t.Label, t.Language, t.Lang),
		})
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
