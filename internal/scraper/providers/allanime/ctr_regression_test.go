// Package allanime — regression suite for the AllAnime episode-source
// resolution protocol.
//
// History of upstream drift this file pins:
//   - 2026-04-22: AllAnime moved `tobeparsed` to AES-256-CTR and required a
//     persisted-query GET with Origin: youtu-chan.com.
//   - 2026-07-08 (ani-cli PR #1772): AllAnime rotated the AES key, restored
//     AES-256-GCM for `tobeparsed`, and began rejecting the episode-sources
//     query outright (AA_CRYPTO_MISSING) unless it carries an `aaReq` proof
//     token. Production now sends aaReq and decrypts `tobeparsed` with GCM.
//
// The `tobeparsed` GCM round-trip itself is pinned in client_test.go
// (TestDecodeToBeParsedCrossValidateWithOpenSSL). This file pins the two
// pieces that are specific to the transport and the still-CTR Filemoon path:
//   - GetEpisodeURL uses the persisted-query GET first (with aaReq + Origin),
//     falling back to POST only when GET yields no source URLs;
//   - Filemoon (sourceName "Fm-mp4") sources are decrypted via their own
//     AES-256-CTR key-parts protocol, which upstream did NOT change.
package allanime

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// 1. Transport regression: persisted-query GET path with aaReq + Origin header
// ---------------------------------------------------------------------------

// TestGetEpisodeURL_UsesPersistedQueryGETFirst pins the modern transport: the
// AllAnime endpoint returns `sourceUrls` only on the persisted-query GET path
// carrying the aaReq token and Origin: https://youtu-chan.com. POST is the
// legacy fallback and must not be issued when GET succeeds.
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
			q := r.URL.Query()
			vars := q.Get("variables")
			ext := q.Get("extensions")
			assert.Contains(t, vars, `"showId":"abc"`, "GET must carry showId in variables")
			assert.Contains(t, vars, `"episodeString":"5"`, "GET must carry episodeString")
			assert.Contains(t, ext, expectedHash, "GET must carry persistedQuery sha256Hash")
			assert.Contains(t, ext, `"aaReq"`, "GET must carry the aaReq crypto token")
			assert.Equal(t, "https://youtu-chan.com", r.Header.Get("Origin"),
				"GET must send Origin: https://youtu-chan.com (else AllAnime strips response)")
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, buildSourceURLsJSON(
				struct{ url, name string }{linkServer.URL, "Default"},
			))
		case http.MethodPost:
			sawPOST.Store(true)
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

// TestGetEpisodeURL_FallsBackToPOSTWhenGETLacksTobeparsed pins the fallback:
// when GET returns empty, retry via POST (which also carries aaReq now).
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
			w.Header().Set("Content-Type", "application/json")
		case http.MethodPost:
			sawPOST.Store(true)
			body, _ := io.ReadAll(r.Body)
			assert.Contains(t, string(body), `"aaReq"`, "POST fallback must also carry aaReq")
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
// 2. Filemoon regression: separate AES-CTR with key_parts protocol (unchanged)
// ---------------------------------------------------------------------------

// b64urlNoPad encodes raw bytes to base64url with NO padding — matches the
// on-the-wire format AllAnime's filemoon endpoint uses for iv/payload/key_parts.
func b64urlNoPad(raw []byte) string {
	return base64.RawURLEncoding.EncodeToString(raw)
}

// buildFilemoonResponse constructs a payload identical in shape to what
// AllAnime's filemoon source endpoint returns:
//
//	{"iv":"<b64url>","payload":"<b64url>","key_parts":["<b64url>","<b64url>"]}
//
// payload = AES-256-CTR-encrypt(plaintext, key=kp1||kp2, iv=iv||00000002)
//
//	|| 16 trailing bytes (ignored by the CTR decoder).
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

// TestGetFilemoonLinks_DecryptsAndExtractsURLs pins the filemoon path:
// Filemoon sources are AES-encrypted with a key split across two key_parts
// fields. Without this path, every filemoon source silently fails.
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
// dispatch: any source named "Fm-mp4" must be fetched via getFilemoonLinks,
// never the generic getLinks.
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
