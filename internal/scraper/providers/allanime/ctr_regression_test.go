// Package scraper — regression suite for the AllAnime cipher-mode + transport
// drift that broke episode source resolution starting 2026-04-22.
//
// Discovered:  2026-04-29 — comparison
//
//	https://github.com/pystardust/ani-cli/compare/allanime-fix...justchokingaround:ani-cli:allanime-fix
//	shows ani-cli switched decode_tobeparsed from AES-256-GCM to
//	AES-256-CTR (counter = nonce||0x00000002), and replaced the
//	direct POST request with a persisted-query GET that requires
//	`Origin: https://youtu-chan.com`. GoAnime still issued the
//	legacy POST and treated the trailing 16 bytes as a GCM auth
//	tag, so every episode lookup failed silently.
//
// Fixed:       2026-04-29 — this commit. Production code switched to CTR;
//
//	GetEpisodeURL now tries the persisted-query GET first and
//	falls back to POST only when the response lacks `tobeparsed`.
//	Filemoon (sourceName "Fm-mp4") sources are now decrypted via
//	their own AES-CTR key-parts protocol.
//
// Root cause:  Two unrelated upstream drifts that landed within a week:
//  1. AllAnime's `tobeparsed` blobs no longer carry a valid
//     GCM auth tag in the trailing 16 bytes (ani-cli commit
//     e5523a9b). `cipher.GCM.Open` rejected every blob with
//     "AES-GCM decryption failed" at allanime.go:114-117.
//  2. AllAnime's POST endpoint stopped returning `tobeparsed`
//     unless the request matches the persisted-query GET shape
//     with `Origin: https://youtu-chan.com` (ani-cli commit
//     1ccbf71f).
//
// Blast radius:total — every AllAnime episode resolution failed since
//
//	2026-04-22. Users saw "no source URLs found for episode" at
//	allanime.go:651 with no path forward (since switching scraper
//	didn't help — AllAnime is the primary EN/sub source).
//
// The tests below pin the fix in place. Reverting any of:
//   - decodeToBeParsed back to AES-GCM
//   - GetEpisodeURL back to POST-only
//   - the filemoon-specific decrypt path
//
// will fail one of these tests loudly.
package allanime

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers — build blobs the way the real (post-2026-04-22) AllAnime API does:
// AES-256-CTR with counter = nonce||0x00000002, plus 16 trailing bytes that
// are NOT a valid GCM auth tag.
// ---------------------------------------------------------------------------

// encryptToBeParsedCTR encrypts plaintext the way ani-cli's decode_tobeparsed
// expects: [0x01 version][12-byte nonce][CTR ciphertext][16 trailing bytes].
// The trailing 16 bytes are random — explicitly NOT a GCM auth tag — so any
// GCM-based decoder would reject the blob.
func encryptToBeParsedCTR(t *testing.T, plaintext string) string {
	t.Helper()
	nonce := make([]byte, 12)
	_, err := io.ReadFull(rand.Reader, nonce)
	require.NoError(t, err)
	return encryptToBeParsedCTRWithNonce(t, plaintext, nonce)
}

func encryptToBeParsedCTRWithNonce(t *testing.T, plaintext string, nonce []byte) string {
	t.Helper()
	require.Len(t, nonce, 12)

	key := sha256.Sum256([]byte(allAnimeKeyPhrase))
	block, err := aes.NewCipher(key[:])
	require.NoError(t, err)

	// CTR initial counter: nonce (12 bytes) || 0x00000002 (4-byte big-endian).
	// This matches GCM's encryption block 0 (J0+1 where J0 = nonce||0x00000001),
	// which is exactly what ani-cli does:
	//     ctr="${iv}00000002"; openssl enc -d -aes-256-ctr -K key -iv "$ctr"
	iv := make([]byte, 16)
	copy(iv[:12], nonce)
	iv[12], iv[13], iv[14], iv[15] = 0x00, 0x00, 0x00, 0x02

	stream := cipher.NewCTR(block, iv)
	ciphertext := make([]byte, len(plaintext))
	stream.XORKeyStream(ciphertext, []byte(plaintext))

	// 16 trailing bytes — these would be a GCM tag if AllAnime were still
	// using GCM. They aren't anymore. Fill with non-zero garbage so any
	// GCM-based decoder cannot get lucky.
	trailing := make([]byte, 16)
	_, err = io.ReadFull(rand.Reader, trailing)
	require.NoError(t, err)

	payload := append([]byte{0x01}, nonce...)
	payload = append(payload, ciphertext...)
	payload = append(payload, trailing...)
	return base64.StdEncoding.EncodeToString(payload)
}

// ---------------------------------------------------------------------------
// 1. Cipher-mode regression: AES-CTR (no auth) vs GCM (rejects bad tag)
// ---------------------------------------------------------------------------

// TestDecodeToBeParsedCTR_RealWorldBlobShape proves the bug exists and
// pins the fix. AllAnime now sends blobs whose trailing 16 bytes are NOT
// a GCM auth tag; GCM rejected them with "AES-GCM decryption failed".
// CTR ignores those bytes by construction, so a correct decoder must
// recover the plaintext regardless of what the trailing 16 bytes are.
//
// If you revert decodeToBeParsed to AES-GCM, this test will fail with
// "AES-GCM decryption failed" because the random trailing 16 bytes
// won't authenticate.
func TestDecodeToBeParsedCTR_RealWorldBlobShape(t *testing.T) {
	t.Parallel()
	plaintext := `{"data":{"episode":{"sourceUrls":[{"sourceUrl":"--504c4c484b021717","sourceName":"Fm-mp4"}]}}}`
	blob := encryptToBeParsedCTR(t, plaintext)

	sources, err := decodeToBeParsed(blob)
	require.NoError(t, err, "post-2026-04-22 blobs use CTR; trailing 16 bytes are NOT a GCM tag")
	require.Len(t, sources, 1)
	assert.Equal(t, "Fm-mp4", sources[0].sourceName)
	assert.Equal(t, "504c4c484b021717", sources[0].sourceURL)
}

// TestDecodeToBeParsedCTR_DeterministicWithFixedNonce verifies counter
// math is correct: encrypt with a known nonce, decrypt, recover plaintext.
// If counter init drifts (e.g. someone uses 0x00000001 — the GCM J0 value
// reserved for the auth tag — instead of 0x00000002), this fails.
func TestDecodeToBeParsedCTR_DeterministicWithFixedNonce(t *testing.T) {
	t.Parallel()
	nonce, _ := hex.DecodeString("aabbccddeeff00112233aabb")
	plaintext := `{"data":{"episode":{"sourceUrls":[{"sourceUrl":"--08","sourceName":"P1"}]}}}`

	blob := encryptToBeParsedCTRWithNonce(t, plaintext, nonce)
	sources, err := decodeToBeParsed(blob)
	require.NoError(t, err)
	require.Len(t, sources, 1)
	assert.Equal(t, "P1", sources[0].sourceName)
	assert.Equal(t, "08", sources[0].sourceURL)
}

// TestDecodeToBeParsedCTR_TrailingBytesIgnored is the explicit pin: if
// someone re-introduces tag verification, this test fails because the
// trailing 16 bytes are deterministically wrong (all 0xFF).
func TestDecodeToBeParsedCTR_TrailingBytesIgnored(t *testing.T) {
	t.Parallel()
	plaintext := `{"data":{"episode":{"sourceUrls":[{"sourceUrl":"--0809","sourceName":"NoTag"}]}}}`

	nonce, _ := hex.DecodeString("000102030405060708090a0b")
	key := sha256.Sum256([]byte(allAnimeKeyPhrase))
	block, err := aes.NewCipher(key[:])
	require.NoError(t, err)

	iv := make([]byte, 0, len(nonce)+4)
	iv = append(iv, nonce...)
	iv = append(iv, 0x00, 0x00, 0x00, 0x02)
	stream := cipher.NewCTR(block, iv)
	ct := make([]byte, len(plaintext))
	stream.XORKeyStream(ct, []byte(plaintext))

	// Deterministic non-tag trailing bytes — all 0xFF.
	trailing := bytes.Repeat([]byte{0xFF}, 16)

	payload := append([]byte{0x01}, nonce...)
	payload = append(payload, ct...)
	payload = append(payload, trailing...)
	blob := base64.StdEncoding.EncodeToString(payload)

	sources, err := decodeToBeParsed(blob)
	require.NoError(t, err, "trailing 16 bytes are not a tag — must be ignored")
	require.Len(t, sources, 1)
	assert.Equal(t, "NoTag", sources[0].sourceName)
}

// TestDecodeToBeParsedCTR_LongPlaintextSpansMultipleBlocks ensures the CTR
// counter increments correctly across AES blocks (16-byte boundaries). A
// stuck counter would corrupt every block past the first.
func TestDecodeToBeParsedCTR_LongPlaintextSpansMultipleBlocks(t *testing.T) {
	t.Parallel()
	// Build 200 source entries → ~10 KB of plaintext, well past block boundary.
	var entries []string
	for i := range 200 {
		entries = append(entries, fmt.Sprintf(`{"sourceUrl":"--0809","sourceName":"P%d"}`, i))
	}
	plaintext := `{"data":{"episode":{"sourceUrls":[` + strings.Join(entries, ",") + `]}}}`

	blob := encryptToBeParsedCTR(t, plaintext)
	sources, err := decodeToBeParsed(blob)
	require.NoError(t, err)
	require.Len(t, sources, 200)
	assert.Equal(t, "P0", sources[0].sourceName)
	assert.Equal(t, "P199", sources[199].sourceName)
}

// ---------------------------------------------------------------------------
// 2. Transport regression: persisted-query GET path with Origin header
// ---------------------------------------------------------------------------

// TestGetEpisodeURL_UsesPersistedQueryGETFirst pins ani-cli patch 1ccbf71f.
// The modern AllAnime endpoint returns `tobeparsed` only on the persisted-
// query GET path with Origin: https://youtu-chan.com. POST returns a
// stripped response that has no source URLs.
//
// This test stands up a server that ONLY honors the GET path + Origin
// header. If GetEpisodeURL reverts to POST-first, the test fails because
// the server returns empty sourceUrls.
func TestGetEpisodeURL_UsesPersistedQueryGETFirst(t *testing.T) {
	t.Parallel()

	const expectedHash = "d405d0edd690624b66baba3068e0edc3ac90f1597d898a1ec8db4e5c43c00fec"

	linkServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, buildLinksJSON(
			struct{ quality, url string }{"720p", "https://cdn.example.com/720.mp4"},
		))
	}))
	defer linkServer.Close()

	var sawGET, sawPOST atomic.Bool
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			sawGET.Store(true)
			// Verify persisted query shape and Origin header.
			q := r.URL.Query()
			vars := q.Get("variables")
			ext := q.Get("extensions")
			assert.Contains(t, vars, `"showId":"abc"`, "GET must carry showId in variables")
			assert.Contains(t, vars, `"episodeString":"5"`, "GET must carry episodeString")
			assert.Contains(t, ext, expectedHash, "GET must carry persistedQuery sha256Hash")
			assert.Equal(t, "https://youtu-chan.com", r.Header.Get("Origin"),
				"GET must send Origin: https://youtu-chan.com (else AllAnime strips response)")
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, buildSourceURLsJSON(
				struct{ url, name string }{linkServer.URL, "Default"},
			))
		case http.MethodPost:
			sawPOST.Store(true)
			// Server-side guard: POST is the legacy path, must not be used
			// when GET succeeds. If the impl regresses, return empty so the
			// test fails on "no source URLs found".
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"data":{"episode":{"episodeString":"5","sourceUrls":[]}}}`)
		}
	}))
	defer apiServer.Close()

	url, _, err := newTestClient(apiServer.URL).GetEpisodeURL("abc", "5", "sub", "best")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/720.mp4", url)
	assert.True(t, sawGET.Load(), "GET path must be tried first")
	assert.False(t, sawPOST.Load(), "POST must not be issued when GET returns sourceUrls")
}

// TestGetEpisodeURL_FallsBackToPOSTWhenGETLacksTobeparsed pins the
// ani-cli fallback: when GET returns empty / no `tobeparsed`, retry via
// POST. This guarantees we don't regress on older AllAnime mirrors that
// still serve the legacy shape.
func TestGetEpisodeURL_FallsBackToPOSTWhenGETLacksTobeparsed(t *testing.T) {
	t.Parallel()

	linkServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, buildLinksJSON(
			struct{ quality, url string }{"480p", "https://cdn.example.com/480.mp4"},
		))
	}))
	defer linkServer.Close()

	var sawGET, sawPOST atomic.Bool
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			sawGET.Store(true)
			// GET returns empty body — triggers POST fallback.
			w.Header().Set("Content-Type", "application/json")
		case http.MethodPost:
			sawPOST.Store(true)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, buildSourceURLsJSON(
				struct{ url, name string }{linkServer.URL, "Default"},
			))
		}
	}))
	defer apiServer.Close()

	url, _, err := newTestClient(apiServer.URL).GetEpisodeURL("abc", "1", "sub", "best")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/480.mp4", url)
	assert.True(t, sawGET.Load(), "GET must be tried first")
	assert.True(t, sawPOST.Load(), "POST must be the fallback when GET response is empty")
}

// ---------------------------------------------------------------------------
// 3. Filemoon regression: separate AES-CTR with key_parts protocol
// ---------------------------------------------------------------------------

// b64urlNoPad encodes raw bytes to base64url with NO padding — matches
// the on-the-wire format AllAnime's filemoon endpoint uses for `iv`,
// `payload`, and `key_parts` entries.
func b64urlNoPad(raw []byte) string {
	return base64.RawURLEncoding.EncodeToString(raw)
}

// buildFilemoonResponse constructs a payload identical in shape to what
// AllAnime's filemoon source endpoint returns (post-2026-04-25):
//
//	{"iv":"<b64url>","payload":"<b64url>","key_parts":["<b64url>","<b64url>"]}
//
// payload = AES-256-CTR-encrypt(plaintext, key=kp1||kp2, iv=iv||00000002)
//
//	|| 16 trailing bytes (ignored, just like the main tobeparsed blob).
func buildFilemoonResponse(t *testing.T, plaintext string) string {
	t.Helper()

	iv := make([]byte, 12)
	_, err := io.ReadFull(rand.Reader, iv)
	require.NoError(t, err)

	kp1 := make([]byte, 16)
	kp2 := make([]byte, 16)
	_, _ = io.ReadFull(rand.Reader, kp1)
	_, _ = io.ReadFull(rand.Reader, kp2)
	key := make([]byte, 0, len(kp1)+len(kp2))
	key = append(key, kp1...)
	key = append(key, kp2...) // 32-byte AES-256 key

	block, err := aes.NewCipher(key)
	require.NoError(t, err)

	ctr := make([]byte, 0, len(iv)+4)
	ctr = append(ctr, iv...)
	ctr = append(ctr, 0x00, 0x00, 0x00, 0x02)
	stream := cipher.NewCTR(block, ctr)
	ct := make([]byte, len(plaintext))
	stream.XORKeyStream(ct, []byte(plaintext))

	// 16 trailing bytes — discarded by the decoder, just like decodeToBeParsed.
	payload := make([]byte, 0, len(ct)+16)
	payload = append(payload, ct...)
	payload = append(payload, bytes.Repeat([]byte{0x00}, 16)...)

	wrapper := map[string]any{
		"iv":        b64urlNoPad(iv),
		"payload":   b64urlNoPad(payload),
		"key_parts": []string{b64urlNoPad(kp1), b64urlNoPad(kp2)},
	}
	out, _ := json.Marshal(wrapper)
	return string(out)
}

// TestGetFilemoonLinks_DecryptsAndExtractsURLs pins ani-cli patch 156bf9b7.
// Filemoon sources (sourceName "Fm-mp4") are NOT generic JSON link blobs —
// their response is itself AES-encrypted with a key split across two
// `key_parts` fields. Without this path, every filemoon source silently
// fails and AllAnime's video resolution hits the fallback chain.
func TestGetFilemoonLinks_DecryptsAndExtractsURLs(t *testing.T) {
	t.Parallel()

	plaintext := `{"sources":[{"url":"https://cdn.filemoon.example/1080.m3u8","height":1080},{"url":"https://cdn.filemoon.example/720.m3u8","height":720}]}`
	body := buildFilemoonResponse(t, plaintext)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, body)
	}))
	defer server.Close()

	links, err := newTestClient(server.URL).getFilemoonLinks(server.URL)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.filemoon.example/1080.m3u8", links["1080p"])
	assert.Equal(t, "https://cdn.filemoon.example/720.m3u8", links["720p"])
}

// TestProcessSourceURLsConcurrent_RoutesFmMp4ToFilemoonDecoder pins the
// dispatch logic: any source named "Fm-mp4" must be fetched via
// getFilemoonLinks, never the generic getLinks. If someone removes the
// dispatch, generic JSON parsing returns no links and the test fails.
func TestProcessSourceURLsConcurrent_RoutesFmMp4ToFilemoonDecoder(t *testing.T) {
	t.Parallel()

	plaintext := `{"sources":[{"url":"https://cdn.filemoon.example/720.m3u8","height":720}]}`
	body := buildFilemoonResponse(t, plaintext)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, body)
	}))
	defer server.Close()

	client := newTestClient("")
	url, _, err := client.processSourceEntriesConcurrent(
		[]sourceEntry{{URL: server.URL, Name: "Fm-mp4"}},
		"best", "anime-id", "1",
	)
	require.NoError(t, err)
	assert.Contains(t, url, "filemoon.example")
}

// ---------------------------------------------------------------------------
// 4. Anti-revert pin: confirm the cipher mode is CTR, not GCM.
// ---------------------------------------------------------------------------

// TestDecodeToBeParsed_AnyTrailing16BytesAccepted is the clearest signal
// that we are NOT using GCM. We craft 5 distinct blobs that share the
// same nonce + ciphertext but have different (random) trailing 16 bytes.
// All five must decrypt to the same plaintext.
//
// Under GCM, at most ONE trailing-byte sequence (the genuine tag) would
// authenticate; the other four would error out. Under CTR, all five
// succeed — that's the property we want to enforce.
func TestDecodeToBeParsed_AnyTrailing16BytesAccepted(t *testing.T) {
	t.Parallel()

	plaintext := `{"data":{"episode":{"sourceUrls":[{"sourceUrl":"--08","sourceName":"X"}]}}}`
	nonce, _ := hex.DecodeString("0123456789abcdef01234567")

	key := sha256.Sum256([]byte(allAnimeKeyPhrase))
	block, _ := aes.NewCipher(key[:])
	iv := make([]byte, 0, len(nonce)+4)
	iv = append(iv, nonce...)
	iv = append(iv, 0x00, 0x00, 0x00, 0x02)
	ct := make([]byte, len(plaintext))
	cipher.NewCTR(block, iv).XORKeyStream(ct, []byte(plaintext))

	for i := range 5 {
		trailing := make([]byte, 16)
		_, _ = io.ReadFull(rand.Reader, trailing)

		payload := append([]byte{0x01}, nonce...)
		payload = append(payload, ct...)
		payload = append(payload, trailing...)
		blob := base64.StdEncoding.EncodeToString(payload)

		sources, err := decodeToBeParsed(blob)
		require.NoErrorf(t, err, "iteration %d: CTR decoder must accept any trailing bytes", i)
		require.Lenf(t, sources, 1, "iteration %d", i)
		assert.Equal(t, "X", sources[0].sourceName)
	}
}
