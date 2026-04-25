package scraper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// newTestClient builds an AllAnimeClient pointed at the given httptest server.
func newTestClient(serverURL string) *AllAnimeClient {
	return &AllAnimeClient{
		client:    util.GetFastClient(),
		referer:   AllAnimeReferer,
		apiBase:   serverURL,
		userAgent: UserAgent,
	}
}

// encryptToBeParsed encrypts plaintext JSON the same way the AllAnime API does:
// base64( 0x01 || nonce(12) || AES-256-GCM(plaintext) )
// Updated 2026-04-24: cipher changed from CTR to GCM; key rotated to "Xot36i3lK3:v1".
func encryptToBeParsed(t *testing.T, plaintext string) string {
	t.Helper()
	nonce := make([]byte, 12)
	_, err := io.ReadFull(rand.Reader, nonce)
	require.NoError(t, err)
	return encryptToBeParsedWithNonce(t, plaintext, nonce)
}

// encryptToBeParsedWithNonce uses an explicit nonce for deterministic tests.
func encryptToBeParsedWithNonce(t *testing.T, plaintext string, nonce []byte) string {
	t.Helper()
	key := sha256.Sum256([]byte(allAnimeKeyPhrase))
	block, err := aes.NewCipher(key[:])
	require.NoError(t, err)
	gcm, err := cipher.NewGCM(block)
	require.NoError(t, err)

	sealed := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	payload := append([]byte{0x01}, nonce...)
	payload = append(payload, sealed...)
	return base64.StdEncoding.EncodeToString(payload)
}

// buildSourceURLsJSON builds a valid AllAnime sourceUrls JSON response.
func buildSourceURLsJSON(sources ...struct{ url, name string }) string {
	type su struct {
		SourceURL  string `json:"sourceUrl"`
		SourceName string `json:"sourceName"`
	}
	entries := make([]su, len(sources))
	for i, s := range sources {
		entries[i] = su{SourceURL: s.url, SourceName: s.name}
	}
	wrapper := map[string]any{
		"data": map[string]any{
			"episode": map[string]any{
				"episodeString": "1",
				"sourceUrls":    entries,
			},
		},
	}
	b, _ := json.Marshal(wrapper)
	return string(b)
}

// buildToBeParsedResponse wraps encrypted blob in an API response.
func buildToBeParsedResponse(blob string) string {
	return fmt.Sprintf(`{"data":{"episode":{"tobeparsed":"%s"}}}`, blob)
}

// hexEncodeSourceURL encodes a URL using the hex substitution table (inverse mapping).
// This is the inverse of decodeSourceURL's table lookup — it produces the hex-encoded
// string that decodeSourceURL would decode back to the original URL.
func hexEncodeSourceURL(url string) string {
	// Build inverse table: char -> hex pair
	inv := make(map[byte]string, len(hexSubstitutionTable))
	for hex, char := range hexSubstitutionTable {
		if len(char) == 1 {
			inv[char[0]] = hex
		}
	}
	var b strings.Builder
	for i := 0; i < len(url); i++ {
		if hex, ok := inv[url[i]]; ok {
			b.WriteString(hex)
		} else {
			// Characters not in the table shouldn't appear in real encoded URLs,
			// but for safety pass them through as-is (will break round-trip).
			b.WriteByte(url[i])
		}
	}
	return b.String()
}

// buildLinksJSON builds a JSON links response like what source endpoints return.
func buildLinksJSON(links ...struct{ quality, url string }) string {
	type linkEntry struct {
		Link          string `json:"link"`
		ResolutionStr string `json:"resolutionStr"`
	}
	entries := make([]linkEntry, len(links))
	for i, l := range links {
		entries[i] = linkEntry{Link: l.url, ResolutionStr: l.quality}
	}
	b, _ := json.Marshal(map[string]any{"links": entries})
	return string(b)
}

// ---------------------------------------------------------------------------
// 1. Existing tests (preserved) — HTTP error classification
// ---------------------------------------------------------------------------

func TestAllAnimeSearchAnimeClassifiesHTMLPayloadAsSourceUnavailable(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `<!DOCTYPE html><html><head><title>Just a moment...</title></head><body><div id="cf-wrapper">Cloudflare block</div></body></html>`)
	}))
	defer server.Close()

	_, err := newTestClient(server.URL).SearchAnime("One Piece")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable))
}

func TestAllAnimeSearchAnimeValidJSONParsesCorrectly(t *testing.T) {
	t.Parallel()
	const validJSON = `{"data":{"shows":{"edges":[{"_id":"abc","name":"One Piece","englishName":"One Piece","availableEpisodes":{"sub":1100}}]}}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, validJSON)
	}))
	defer server.Close()

	results, err := newTestClient(server.URL).SearchAnime("One Piece")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Contains(t, results[0].Name, "One Piece")
}

func TestAllAnimeSearchAnimeClassifies403AsSourceUnavailable(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	_, err := newTestClient(server.URL).SearchAnime("One Piece")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable))
}

func TestAllAnimeGetEpisodesListClassifiesHTMLPayloadAsSourceUnavailable(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, `<html><body>Access Denied</body></html>`)
	}))
	defer server.Close()

	_, err := newTestClient(server.URL).GetEpisodesList("some-id", "sub")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable))
}

func TestAllAnimeGetEpisodeURLClassifiesHTMLAsSourceUnavailable(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, `<html><body>Rate limited</body></html>`)
	}))
	defer server.Close()

	_, _, err := newTestClient(server.URL).GetEpisodeURL("anime-id", "1", "sub", "best")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable))
}

func TestAllAnimeGetEpisodeURL503ClassifiesAsSourceUnavailable(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	_, _, err := newTestClient(server.URL).GetEpisodeURL("anime-id", "1", "sub", "best")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable))
}

func TestCheckHTMLResponseByteFallback(t *testing.T) {
	t.Parallel()
	resp := &http.Response{Header: make(http.Header)}
	body := []byte("\r\n<!DOCTYPE html><html><body>blocked</body></html>")
	err := checkHTMLResponse(resp, body, "test-source")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable))
}

func TestCheckHTTPStatusNonBlockingCodeReturnsPlainError(t *testing.T) {
	t.Parallel()
	resp := &http.Response{StatusCode: http.StatusNotFound}
	err := checkHTTPStatus(resp, "test-source")
	require.Error(t, err)
	assert.False(t, errors.Is(err, ErrSourceUnavailable))
	assert.Contains(t, err.Error(), "404")
}

func TestAllAnimeGetLinksClassifiesHTMLContentTypeAsSourceUnavailable(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, `<!DOCTYPE html><html><head><title>Just a moment...</title></head><body><div id="cf-wrapper"></div></body></html>`)
	}))
	defer server.Close()

	_, err := newTestClient(server.URL).getLinks(server.URL + "/links")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable))
}

func TestAllAnimeGetLinksClassifiesHTMLBodyAsSourceUnavailable(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body>Access Denied</body></html>`)
	}))
	defer server.Close()

	_, err := newTestClient(server.URL).getLinks(server.URL + "/links")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable))
}

// ---------------------------------------------------------------------------
// 2. AES key derivation
// ---------------------------------------------------------------------------

func TestAllAnimeKeyMatchesOpenSSL(t *testing.T) {
	t.Parallel()
	// sha256("Xot36i3lK3:v1") — updated 2026-04-24 when AllAnime rotated the key.
	// Verify: printf '%s' 'Xot36i3lK3:v1' | openssl dgst -sha256
	expected := "a254aa27c410f297bd04ba33a0c0df7ff4e706bf3ae27271c6703f84e750f552"
	assert.Equal(t, expected, hex.EncodeToString(allAnimeKey))
}

func TestAllAnimeKeyLength(t *testing.T) {
	t.Parallel()
	assert.Len(t, allAnimeKey, 32, "AES-256 key must be 32 bytes")
}

// ---------------------------------------------------------------------------
// 3. Hex substitution cipher — decodeSourceURL
// ---------------------------------------------------------------------------

func TestDecodeSourceURLCompleteMapping(t *testing.T) {
	t.Parallel()
	client := NewAllAnimeClient()

	tests := []struct {
		name     string
		encoded  string
		expected string
	}{
		{
			name:     "path /clock gets base prepended and .json suffix",
			encoded:  "175b54575b53",
			expected: "https://allanime.day/clock.json",
		},
		{name: "lowercase letters",
			encoded:  "595a5b5c5d5e5f505152535455565748494a4b4c4d4e4f404142",
			expected: "abcdefghijklmnopqrstuvwxyz"},
		{name: "uppercase letters",
			encoded:  "797a7b7c7d7e7f707172737475767768696a6b6c6d6e6f606162",
			expected: "ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{name: "digits",
			encoded:  "08090a0b0c0d0e0f0001",
			expected: "0123456789"},
		{name: "special characters",
			encoded:  "151667460217071b636578191c1e101112131403051d",
			expected: "-._~:/?#[]@!$&()*+,;=%"},
		{name: "full https URL",
			encoded:  "504c4c484b021717595454595651555d165c5941175b54575b53",
			expected: "https://allanime.day/clock.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, client.decodeSourceURL(tt.encoded))
		})
	}
}

func TestDecodeSourceURLEmptyInput(t *testing.T) {
	t.Parallel()
	client := NewAllAnimeClient()
	assert.Equal(t, "", client.decodeSourceURL(""))
}

func TestDecodeSourceURLOddLengthDropsTrailingChar(t *testing.T) {
	t.Parallel()
	client := NewAllAnimeClient()
	// "0809X" -> pairs "08","09" decoded to "01", trailing "X" ignored
	result := client.decodeSourceURL("0809X")
	assert.Equal(t, "01", result)
}

func TestDecodeSourceURLUnknownPairsPassThrough(t *testing.T) {
	t.Parallel()
	client := NewAllAnimeClient()
	// "zz" is not in the table — should pass through unchanged
	result := client.decodeSourceURL("zz0809")
	assert.Equal(t, "zz01", result)
}

func TestDecodeSourceURLClockReplacementMultiple(t *testing.T) {
	t.Parallel()
	client := NewAllAnimeClient()
	// Encode "/clock/clock" -> both should be replaced
	// /=17 c=5b l=54 o=57 c=5b k=53
	result := client.decodeSourceURL("175b54575b53175b54575b53")
	assert.Equal(t, "https://allanime.day/clock.json/clock.json", result)
}

func TestDecodeSourceURLDoesNotPrependBaseForAbsoluteURL(t *testing.T) {
	t.Parallel()
	client := NewAllAnimeClient()
	// Encode "https://example.com/video.mp4"
	// h=50 t=4c t=4c p=48 s=4b :=02 /=17 /=17 e=5d x=40 a=59 m=55 p=48 l=54 e=5d .=16 c=5b o=57 m=55
	encoded := "504c4c484b02171759165b5755"
	result := client.decodeSourceURL(encoded)
	// Starts with "h" not "/" so no base prefix
	assert.False(t, strings.HasPrefix(result, "https://allanime.day/"))
}

func TestDecodeSourceURLBijectiveProperty(t *testing.T) {
	// Verify every entry in the substitution table is unique (no collisions)
	t.Parallel()
	seen := make(map[string]string)
	for hex, char := range hexSubstitutionTable {
		if prevHex, dup := seen[char]; dup {
			t.Errorf("duplicate mapping: both %q and %q map to %q", prevHex, hex, char)
		}
		seen[char] = hex
	}
}

func TestDecodeSourceURLTableCoversAllExpectedChars(t *testing.T) {
	t.Parallel()
	// Verify coverage: a-z, A-Z, 0-9, and all special chars from the bash script
	expected := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~:/?#[]@!$&()*+,;=%"
	covered := make(map[rune]bool)
	for _, ch := range hexSubstitutionTable {
		for _, r := range ch {
			covered[r] = true
		}
	}
	for _, r := range expected {
		assert.True(t, covered[r], "character %q not covered by substitution table", string(r))
	}
}

// ---------------------------------------------------------------------------
// 4. AES-256-GCM — decodeToBeParsed (updated 2026-04-24: CTR → GCM, new key)
// ---------------------------------------------------------------------------

func TestDecodeToBeParsedRoundTrip(t *testing.T) {
	t.Parallel()
	plaintext := `{"data":{"episode":{"sourceUrls":[{"sourceUrl":"--504c4c484b021717","sourceName":"TestProvider"}]}}}`
	blob := encryptToBeParsed(t, plaintext)

	sources, err := decodeToBeParsed(blob)
	require.NoError(t, err)
	require.Len(t, sources, 1)
	assert.Equal(t, "TestProvider", sources[0].sourceName)
	assert.Equal(t, "504c4c484b021717", sources[0].sourceURL)
}

func TestDecodeToBeParsedMultipleSources(t *testing.T) {
	t.Parallel()
	plaintext := `{"data":{"episode":{"sourceUrls":[
		{"sourceUrl":"--0809","sourceName":"Provider1"},
		{"sourceUrl":"--0a0b","sourceName":"Provider2"},
		{"sourceUrl":"--0c0d","sourceName":"Provider3"}
	]}}}`
	blob := encryptToBeParsed(t, plaintext)

	sources, err := decodeToBeParsed(blob)
	require.NoError(t, err)
	require.Len(t, sources, 3)
	assert.Equal(t, "Provider1", sources[0].sourceName)
	assert.Equal(t, "0809", sources[0].sourceURL)
	assert.Equal(t, "Provider2", sources[1].sourceName)
	assert.Equal(t, "Provider3", sources[2].sourceName)
}

func TestDecodeToBeParsedTooShort(t *testing.T) {
	t.Parallel()
	blob := base64.StdEncoding.EncodeToString([]byte("short"))
	_, err := decodeToBeParsed(blob)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too short")
}

func TestDecodeToBeParsedBadBase64(t *testing.T) {
	t.Parallel()
	_, err := decodeToBeParsed("!!!not-base64!!!")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "base64")
}

func TestDecodeToBeParsedExactly12BytesNoCiphertext(t *testing.T) {
	t.Parallel()
	// 12 bytes < 30 (minimum for GCM: 1+12+16+1) → too short
	blob := base64.StdEncoding.EncodeToString(make([]byte, 12))
	_, err := decodeToBeParsed(blob)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too short")
}

func TestDecodeToBeParsedExactly13BytesMinimal(t *testing.T) {
	t.Parallel()
	// 13 bytes < 30 → too short
	blob := base64.StdEncoding.EncodeToString(make([]byte, 13))
	_, err := decodeToBeParsed(blob)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too short")
}

func TestDecodeToBeParsedExactly29BytesStillTooShort(t *testing.T) {
	t.Parallel()
	// Updated 2026-04-24: GCM requires ≥ 30 bytes (1 version + 12 nonce + 16 tag + 1 plaintext).
	// 29 bytes is still too short.
	blob := base64.StdEncoding.EncodeToString(make([]byte, 29))
	_, err := decodeToBeParsed(blob)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too short")
}

func TestDecodeToBeParsedCorruptedCiphertext(t *testing.T) {
	t.Parallel()
	plaintext := `{"data":{"episode":{"sourceUrls":[{"sourceUrl":"--0809","sourceName":"P1"}]}}}`
	blob := encryptToBeParsed(t, plaintext)

	// Decode, flip bits in the ciphertext+tag region (bytes 13+), re-encode.
	// AES-GCM authentication will reject the tampered payload.
	raw, err := base64.StdEncoding.DecodeString(blob)
	require.NoError(t, err)
	for i := 13; i < len(raw); i++ {
		raw[i] ^= 0xFF
	}
	corruptBlob := base64.StdEncoding.EncodeToString(raw)

	_, err = decodeToBeParsed(corruptBlob)
	assert.Error(t, err, "GCM authentication must reject tampered ciphertext")
	assert.Contains(t, err.Error(), "AES-GCM decryption failed")
}

func TestDecodeToBeParsedTruncatedCiphertext(t *testing.T) {
	t.Parallel()
	plaintext := `{"data":{"episode":{"sourceUrls":[{"sourceUrl":"--0809","sourceName":"P1"}]}}}`
	blob := encryptToBeParsed(t, plaintext)

	raw, err := base64.StdEncoding.DecodeString(blob)
	require.NoError(t, err)
	// 16 bytes < 30 minimum → "too short" before GCM is even attempted
	truncated := base64.StdEncoding.EncodeToString(raw[:16])

	_, err = decodeToBeParsed(truncated)
	assert.Error(t, err, "truncated blob should fail")
}

func TestDecodeToBeParsedRegexFallbackSourceUrlBeforeSourceName(t *testing.T) {
	t.Parallel()
	// JSON that is NOT valid for the struct parser but matches the regex
	plaintext := `[{"sourceUrl":"--5959","sourceName":"Fallback1"}]`
	blob := encryptToBeParsed(t, plaintext)

	sources, err := decodeToBeParsed(blob)
	require.NoError(t, err)
	require.Len(t, sources, 1)
	assert.Equal(t, "Fallback1", sources[0].sourceName)
	assert.Equal(t, "5959", sources[0].sourceURL)
}

func TestDecodeToBeParsedRegexFallbackReversedFieldOrder(t *testing.T) {
	t.Parallel()
	// sourceName comes before sourceUrl — tests the reverse-order regex
	plaintext := `[{"sourceName":"Reversed","sourceUrl":"--0a0b"}]`
	blob := encryptToBeParsed(t, plaintext)

	sources, err := decodeToBeParsed(blob)
	require.NoError(t, err)
	require.Len(t, sources, 1)
	assert.Equal(t, "Reversed", sources[0].sourceName)
	assert.Equal(t, "0a0b", sources[0].sourceURL)
}

func TestDecodeToBeParsedDeterministicWithFixedNonce(t *testing.T) {
	t.Parallel()
	plaintext := `{"data":{"episode":{"sourceUrls":[{"sourceUrl":"--08","sourceName":"Det"}]}}}`
	nonce, _ := hex.DecodeString("000000000000000000000000")

	blob1 := encryptToBeParsedWithNonce(t, plaintext, nonce)
	blob2 := encryptToBeParsedWithNonce(t, plaintext, nonce)
	assert.Equal(t, blob1, blob2, "same nonce + plaintext must produce same blob")

	sources, err := decodeToBeParsed(blob1)
	require.NoError(t, err)
	require.Len(t, sources, 1)
	assert.Equal(t, "Det", sources[0].sourceName)
}

func TestDecodeToBeParsedLargePayload(t *testing.T) {
	t.Parallel()
	// Build a payload with 100 source URLs
	var entries []string
	for i := 0; i < 100; i++ {
		entries = append(entries, fmt.Sprintf(`{"sourceUrl":"--0809","sourceName":"P%d"}`, i))
	}
	plaintext := `{"data":{"episode":{"sourceUrls":[` + strings.Join(entries, ",") + `]}}}`
	blob := encryptToBeParsed(t, plaintext)

	sources, err := decodeToBeParsed(blob)
	require.NoError(t, err)
	assert.Len(t, sources, 100)
}

// ---------------------------------------------------------------------------
// 5. extractSourceURLs — integration tests
// ---------------------------------------------------------------------------

func TestExtractSourceURLsHandlesToBeParsed(t *testing.T) {
	t.Parallel()
	plaintext := `{"data":{"episode":{"sourceUrls":[{"sourceUrl":"--175b54575b53","sourceName":"TestProv"}]}}}`
	blob := encryptToBeParsed(t, plaintext)
	response := buildToBeParsedResponse(blob)

	urls := NewAllAnimeClient().extractSourceURLs(response)
	require.Len(t, urls, 1)
	assert.Equal(t, "https://allanime.day/clock.json", urls[0])
}

func TestExtractSourceURLsStandardPath(t *testing.T) {
	t.Parallel()
	response := buildSourceURLsJSON(struct{ url, name string }{"--08090a", "Default"})

	urls := NewAllAnimeClient().extractSourceURLs(response)
	require.Len(t, urls, 1)
	assert.Equal(t, "012", urls[0])
}

func TestExtractSourceURLsDirectURLWithoutPrefix(t *testing.T) {
	t.Parallel()
	response := buildSourceURLsJSON(struct{ url, name string }{"https://example.com/video.mp4", "Direct"})

	urls := NewAllAnimeClient().extractSourceURLs(response)
	require.Len(t, urls, 1)
	assert.Equal(t, "https://example.com/video.mp4", urls[0])
}

func TestExtractSourceURLsMixedEncodedAndDirect(t *testing.T) {
	t.Parallel()
	response := buildSourceURLsJSON(
		struct{ url, name string }{"--0809", "Encoded"},
		struct{ url, name string }{"https://cdn.example.com/stream.m3u8", "Direct"},
	)

	urls := NewAllAnimeClient().extractSourceURLs(response)
	require.Len(t, urls, 2)
	assert.Equal(t, "01", urls[0])
	assert.Equal(t, "https://cdn.example.com/stream.m3u8", urls[1])
}

func TestExtractSourceURLsToBeParsedFallsBackToStandard(t *testing.T) {
	t.Parallel()
	// Response has "tobeparsed" but with garbage blob; should fall back to sourceUrls
	response := `{"data":{"episode":{"tobeparsed":"not-valid-base64","sourceUrls":[{"sourceUrl":"--0809","sourceName":"Fallback"}]}}}`

	urls := NewAllAnimeClient().extractSourceURLs(response)
	require.Len(t, urls, 1)
	assert.Equal(t, "01", urls[0])
}

func TestExtractSourceURLsRegexFallbackOnMalformedJSON(t *testing.T) {
	t.Parallel()
	// Invalid JSON but contains regex-matchable pattern
	response := `not json {"sourceUrl":"--0809","blah":"x","sourceName":"Regex"} garbage`

	urls := NewAllAnimeClient().extractSourceURLs(response)
	require.Len(t, urls, 1)
	assert.Equal(t, "01", urls[0])
}

func TestExtractSourceURLsEmptyResponse(t *testing.T) {
	t.Parallel()
	urls := NewAllAnimeClient().extractSourceURLs("")
	assert.Empty(t, urls)
}

func TestExtractSourceURLsNoSourceUrls(t *testing.T) {
	t.Parallel()
	response := `{"data":{"episode":{"episodeString":"1","sourceUrls":[]}}}`
	urls := NewAllAnimeClient().extractSourceURLs(response)
	assert.Empty(t, urls)
}

func TestExtractToBeParsedBlobExtractsCorrectly(t *testing.T) {
	t.Parallel()
	response := `{"data":{"episode":{"tobeparsed":"AQID"}}}`
	assert.Equal(t, "AQID", extractToBeParsedBlob(response))
}

func TestExtractToBeParsedBlobMissing(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", extractToBeParsedBlob(`{"data":{}}`))
}

// ---------------------------------------------------------------------------
// 6. extractVideoLinks — unit tests
// ---------------------------------------------------------------------------

func TestExtractVideoLinksJSON(t *testing.T) {
	t.Parallel()
	response := buildLinksJSON(
		struct{ quality, url string }{"1080p", "https://cdn.example.com/1080.mp4"},
		struct{ quality, url string }{"720p", "https://cdn.example.com/720.mp4"},
	)

	links := NewAllAnimeClient().extractVideoLinks(response)
	assert.Equal(t, "https://cdn.example.com/1080.mp4", links["1080p"])
	assert.Equal(t, "https://cdn.example.com/720.mp4", links["720p"])
}

func TestExtractVideoLinksHLS(t *testing.T) {
	t.Parallel()
	response := `{"links":[{"link":"https://cdn.example.com/master.m3u8","hls":true}]}`
	links := NewAllAnimeClient().extractVideoLinks(response)
	assert.Contains(t, links, "hls")
	assert.Equal(t, "https://cdn.example.com/master.m3u8", links["hls"])
}

func TestExtractVideoLinksUnknownQuality(t *testing.T) {
	t.Parallel()
	response := `{"links":[{"link":"https://cdn.example.com/stream.mp4"}]}`
	links := NewAllAnimeClient().extractVideoLinks(response)
	assert.Contains(t, links, "unknown")
}

func TestExtractVideoLinksEscapedBackslashes(t *testing.T) {
	t.Parallel()
	response := `{"links":[{"link":"https:\\/\\/cdn.example.com\\/video.mp4","resolutionStr":"480p"}]}`
	links := NewAllAnimeClient().extractVideoLinks(response)
	assert.Equal(t, "https://cdn.example.com/video.mp4", links["480p"])
}

func TestExtractVideoLinksEmptyResponse(t *testing.T) {
	t.Parallel()
	assert.Empty(t, NewAllAnimeClient().extractVideoLinks(""))
}

func TestExtractVideoLinksGarbageInput(t *testing.T) {
	t.Parallel()
	assert.Empty(t, NewAllAnimeClient().extractVideoLinks("this is not json at all"))
}

func TestExtractVideoLinksRegexFallback(t *testing.T) {
	t.Parallel()
	// Non-JSON but regex-matchable
	response := `something "link":"https://cdn.example.com/vid.mp4" and "resolutionStr":"360p" end`
	links := NewAllAnimeClient().extractVideoLinks(response)
	assert.Equal(t, "https://cdn.example.com/vid.mp4", links["360p"])
}

// ---------------------------------------------------------------------------
// 7. selectQuality — unit tests
// ---------------------------------------------------------------------------

func TestSelectQualityBestPicks1080p(t *testing.T) {
	t.Parallel()
	links := map[string]string{
		"360p":  "https://cdn.example.com/360.mp4",
		"720p":  "https://cdn.example.com/720.mp4",
		"1080p": "https://cdn.example.com/1080.mp4",
	}
	url, meta := NewAllAnimeClient().selectQuality(links, "best")
	assert.Equal(t, "https://cdn.example.com/1080.mp4", url)
	assert.Equal(t, "1080p", meta["quality"])
}

func TestSelectQualityWorstPicks360p(t *testing.T) {
	t.Parallel()
	links := map[string]string{
		"360p":  "https://cdn.example.com/360.mp4",
		"720p":  "https://cdn.example.com/720.mp4",
		"1080p": "https://cdn.example.com/1080.mp4",
	}
	url, meta := NewAllAnimeClient().selectQuality(links, "worst")
	assert.Equal(t, "https://cdn.example.com/360.mp4", url)
	assert.Equal(t, "360p", meta["quality"])
}

func TestSelectQualityExactMatch(t *testing.T) {
	t.Parallel()
	links := map[string]string{
		"360p":  "https://cdn.example.com/360.mp4",
		"720p":  "https://cdn.example.com/720.mp4",
		"1080p": "https://cdn.example.com/1080.mp4",
	}
	url, meta := NewAllAnimeClient().selectQuality(links, "720p")
	assert.Equal(t, "https://cdn.example.com/720.mp4", url)
	assert.Equal(t, "720p", meta["quality"])
}

func TestSelectQualityPriorityPreferred(t *testing.T) {
	t.Parallel()
	links := map[string]string{
		"1080p":          "https://other.example.com/1080.mp4",
		"1080p_priority": "https://sharepoint.com/1080.mp4",
	}
	url, meta := NewAllAnimeClient().selectQuality(links, "best")
	assert.Equal(t, "https://sharepoint.com/1080.mp4", url)
	assert.Equal(t, "high", meta["priority"])
}

func TestSelectQualityFallbackToHLS(t *testing.T) {
	t.Parallel()
	links := map[string]string{
		"hls": "https://cdn.example.com/master.m3u8",
	}
	url, meta := NewAllAnimeClient().selectQuality(links, "best")
	assert.Equal(t, "https://cdn.example.com/master.m3u8", url)
	assert.Equal(t, "hls", meta["quality"])
	assert.Equal(t, "m3u8", meta["type"])
}

func TestSelectQualityEmptyLinks(t *testing.T) {
	t.Parallel()
	url, _ := NewAllAnimeClient().selectQuality(map[string]string{}, "best")
	assert.Equal(t, "", url)
}

func TestSelectQualityNoExactMatchFallsThrough(t *testing.T) {
	t.Parallel()
	links := map[string]string{
		"480p": "https://cdn.example.com/480.mp4",
	}
	url, meta := NewAllAnimeClient().selectQuality(links, "1080p")
	// Should fall back to HLS or any available
	assert.NotEmpty(t, url)
	assert.Equal(t, "480p", meta["quality"])
}

// ---------------------------------------------------------------------------
// 8. getPriorityScore — unit tests
// ---------------------------------------------------------------------------

func TestGetPriorityScoreKnownDomains(t *testing.T) {
	t.Parallel()
	client := NewAllAnimeClient()

	assert.Greater(t, client.getPriorityScore("https://sharepoint.com/video.mp4"), 0)
	assert.Greater(t, client.getPriorityScore("https://wixmp.com/video.mp4"), 0)
	assert.Greater(t, client.getPriorityScore("https://dropbox.com/video.mp4"), 0)
}

func TestGetPriorityScoreSharepointHighest(t *testing.T) {
	t.Parallel()
	client := NewAllAnimeClient()
	sp := client.getPriorityScore("https://sharepoint.com/x")
	wx := client.getPriorityScore("https://wixmp.com/x")
	assert.Greater(t, sp, wx, "sharepoint should have higher priority than wixmp")
}

func TestGetPriorityScoreUnknownDomain(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 0, NewAllAnimeClient().getPriorityScore("https://unknown.example.com/x"))
}

// ---------------------------------------------------------------------------
// 9. prioritizeLinks — unit tests
// ---------------------------------------------------------------------------

func TestPrioritizeLinksAddsPriorityEntries(t *testing.T) {
	t.Parallel()
	links := map[string]string{
		"1080p": "https://wixmp.com/1080.mp4",
		"720p":  "https://other.com/720.mp4",
	}
	result := NewAllAnimeClient().prioritizeLinks(links)
	assert.Contains(t, result, "1080p_priority")
	assert.NotContains(t, result, "720p_priority")
	// Original links preserved
	assert.Contains(t, result, "1080p")
	assert.Contains(t, result, "720p")
}

func TestPrioritizeLinksNoHighPriority(t *testing.T) {
	t.Parallel()
	links := map[string]string{
		"720p": "https://generic.com/720.mp4",
	}
	result := NewAllAnimeClient().prioritizeLinks(links)
	assert.NotContains(t, result, "720p_priority")
	assert.Contains(t, result, "720p")
}

// ---------------------------------------------------------------------------
// 10. parseEpisodeNum — unit tests
// ---------------------------------------------------------------------------

func TestParseEpisodeNum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected int
	}{
		{"1", 1},
		{"25", 25},
		{"100", 100},
		{"0", 1},   // zero defaults to 1
		{"abc", 1}, // non-numeric defaults to 1
		{"", 1},    // empty defaults to 1
		{"12.5", 12},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, parseEpisodeNum(tt.input))
		})
	}
}

// ---------------------------------------------------------------------------
// 11. extractEpisodes — unit tests
// ---------------------------------------------------------------------------

func TestExtractEpisodesSubMode(t *testing.T) {
	t.Parallel()
	detail := map[string]any{
		"sub": []any{"3", "1", "2"},
	}
	eps := extractEpisodes(detail, "sub")
	assert.Equal(t, []string{"1", "2", "3"}, eps)
}

func TestExtractEpisodesDubMode(t *testing.T) {
	t.Parallel()
	detail := map[string]any{
		"dub": []any{"5", "10"},
		"sub": []any{"1", "2"},
	}
	eps := extractEpisodes(detail, "dub")
	assert.Equal(t, []string{"5", "10"}, eps)
}

func TestExtractEpisodesMissingMode(t *testing.T) {
	t.Parallel()
	detail := map[string]any{
		"sub": []any{"1"},
	}
	eps := extractEpisodes(detail, "dub")
	assert.Empty(t, eps)
}

func TestExtractEpisodesNilMap(t *testing.T) {
	t.Parallel()
	assert.Empty(t, extractEpisodes(nil, "sub"))
}

func TestExtractEpisodesFloatEpisodes(t *testing.T) {
	t.Parallel()
	detail := map[string]any{
		"sub": []any{1.0, 2.0, 3.5},
	}
	eps := extractEpisodes(detail, "sub")
	require.Len(t, eps, 3)
	assert.Equal(t, "1", eps[0])
	assert.Equal(t, "2", eps[1])
	assert.Equal(t, "3.5", eps[2])
}

// ---------------------------------------------------------------------------
// 12. SendSkipTimesToMPV — mock tests
// ---------------------------------------------------------------------------

func TestSendSkipTimesToMPVSuccess(t *testing.T) {
	t.Parallel()
	client := NewAllAnimeClient()
	ep := &models.Episode{
		SkipTimes: models.SkipTimes{
			Op: models.Skip{Start: 10, End: 100},
			Ed: models.Skip{Start: 1200, End: 1290},
		},
	}

	var capturedArgs []any
	mockCmd := func(_ string, args []any) (any, error) {
		capturedArgs = args
		return nil, nil
	}

	err := client.SendSkipTimesToMPV(ep, "/tmp/mpv.sock", mockCmd)
	require.NoError(t, err)
	require.Len(t, capturedArgs, 3)
	assert.Equal(t, "set_property", capturedArgs[0])
	assert.Equal(t, "chapter-list", capturedArgs[1])

	chapters := capturedArgs[2].([]map[string]any)
	// Should have: Pre-Opening, Opening, Main, Ending, Post-Credits
	assert.GreaterOrEqual(t, len(chapters), 4)
	// Verify chapter titles exist
	titles := make([]string, len(chapters))
	for i, ch := range chapters {
		titles[i] = ch["title"].(string)
	}
	assert.Contains(t, titles, "Opening")
	assert.Contains(t, titles, "Ending")
	assert.Contains(t, titles, "Main")
}

func TestSendSkipTimesToMPVNoSkipTimes(t *testing.T) {
	t.Parallel()
	ep := &models.Episode{} // zero SkipTimes
	mockCmd := func(_ string, _ []any) (any, error) { return nil, nil }

	err := NewAllAnimeClient().SendSkipTimesToMPV(ep, "/tmp/mpv.sock", mockCmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no skip times")
}

func TestSendSkipTimesToMPVCommandError(t *testing.T) {
	t.Parallel()
	ep := &models.Episode{
		SkipTimes: models.SkipTimes{
			Op: models.Skip{Start: 10, End: 100},
		},
	}
	mockCmd := func(_ string, _ []any) (any, error) {
		return nil, fmt.Errorf("IPC connection refused")
	}

	err := NewAllAnimeClient().SendSkipTimesToMPV(ep, "/tmp/mpv.sock", mockCmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "IPC connection refused")
}

func TestSendSkipTimesToMPVOnlyOpening(t *testing.T) {
	t.Parallel()
	ep := &models.Episode{
		SkipTimes: models.SkipTimes{
			Op: models.Skip{Start: 5, End: 90},
		},
	}

	var capturedChapters []map[string]any
	mockCmd := func(_ string, args []any) (any, error) {
		capturedChapters = args[2].([]map[string]any)
		return nil, nil
	}

	err := NewAllAnimeClient().SendSkipTimesToMPV(ep, "/tmp/mpv.sock", mockCmd)
	require.NoError(t, err)

	titles := make(map[string]bool)
	for _, ch := range capturedChapters {
		titles[ch["title"].(string)] = true
	}
	assert.True(t, titles["Opening"])
	assert.True(t, titles["Main"])
	assert.False(t, titles["Ending"], "no ED skip times -> no Ending chapter")
}

func TestSendSkipTimesToMPVOnlyEnding(t *testing.T) {
	t.Parallel()
	ep := &models.Episode{
		SkipTimes: models.SkipTimes{
			Ed: models.Skip{Start: 1200, End: 1290},
		},
	}

	var capturedChapters []map[string]any
	mockCmd := func(_ string, args []any) (any, error) {
		capturedChapters = args[2].([]map[string]any)
		return nil, nil
	}

	err := NewAllAnimeClient().SendSkipTimesToMPV(ep, "/tmp/mpv.sock", mockCmd)
	require.NoError(t, err)

	titles := make(map[string]bool)
	for _, ch := range capturedChapters {
		titles[ch["title"].(string)] = true
	}
	assert.True(t, titles["Ending"])
	assert.True(t, titles["Post-Credits"])
	assert.False(t, titles["Pre-Opening"])
}

// ---------------------------------------------------------------------------
// 13. SearchAnime — mock server tests
// ---------------------------------------------------------------------------

func TestSearchAnimeMultipleResults(t *testing.T) {
	t.Parallel()
	response := `{"data":{"shows":{"edges":[
		{"_id":"id1","name":"Naruto","englishName":"Naruto","availableEpisodes":{"sub":220}},
		{"_id":"id2","name":"Naruto Shippuden","englishName":"Naruto Shippuden","availableEpisodes":{"sub":500}},
		{"_id":"id3","name":"Boruto","englishName":"Boruto","availableEpisodes":{"sub":293}}
	]}}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, response)
	}))
	defer server.Close()

	results, err := newTestClient(server.URL).SearchAnime("Naruto")
	require.NoError(t, err)
	require.Len(t, results, 3)
	// Sorted by episode count descending
	assert.Contains(t, results[0].Name, "Naruto Shippuden")
	assert.Equal(t, "id2", results[0].URL)
}

func TestSearchAnimePrefersEnglishName(t *testing.T) {
	t.Parallel()
	response := `{"data":{"shows":{"edges":[
		{"_id":"id1","name":"Shingeki no Kyojin","englishName":"Attack on Titan","availableEpisodes":{"sub":75}}
	]}}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, response)
	}))
	defer server.Close()

	results, err := newTestClient(server.URL).SearchAnime("attack")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Contains(t, results[0].Name, "Attack on Titan")
	assert.NotContains(t, results[0].Name, "Shingeki")
}

func TestSearchAnimeFallsBackToJapaneseName(t *testing.T) {
	t.Parallel()
	response := `{"data":{"shows":{"edges":[
		{"_id":"id1","name":"Shingeki no Kyojin","englishName":"","availableEpisodes":{"sub":75}}
	]}}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, response)
	}))
	defer server.Close()

	results, err := newTestClient(server.URL).SearchAnime("shingeki")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Contains(t, results[0].Name, "Shingeki no Kyojin")
}

func TestSearchAnimeEmptyResults(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"data":{"shows":{"edges":[]}}}`)
	}))
	defer server.Close()

	results, err := newTestClient(server.URL).SearchAnime("nonexistent12345")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearchAnimeInvalidJSON(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{invalid json}`)
	}))
	defer server.Close()

	_, err := newTestClient(server.URL).SearchAnime("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}

func TestSearchAnimeConnectionRefused(t *testing.T) {
	t.Parallel()
	// Use a URL that will definitely fail to connect
	_, err := newTestClient("http://127.0.0.1:1").SearchAnime("test")
	assert.Error(t, err)
}

func TestSearchAnimeVerifiesRequestHeaders(t *testing.T) {
	t.Parallel()
	var capturedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"data":{"shows":{"edges":[]}}}`)
	}))
	defer server.Close()

	_, _ = newTestClient(server.URL).SearchAnime("test")
	assert.Equal(t, "application/json", capturedHeaders.Get("Content-Type"))
	assert.Equal(t, AllAnimeReferer, capturedHeaders.Get("Referer"))
	assert.NotEmpty(t, capturedHeaders.Get("User-Agent"))
}

func TestSearchAnimeVerifiesGraphQLBody(t *testing.T) {
	t.Parallel()
	var capturedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		capturedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"data":{"shows":{"edges":[]}}}`)
	}))
	defer server.Close()

	_, _ = newTestClient(server.URL).SearchAnime("test query")
	assert.Contains(t, capturedBody, "test query")
	assert.Contains(t, capturedBody, "query")
	assert.Contains(t, capturedBody, "variables")
}

// ---------------------------------------------------------------------------
// 14. GetEpisodesList — mock server tests
// ---------------------------------------------------------------------------

func TestGetEpisodesListSuccess(t *testing.T) {
	t.Parallel()
	response := `{"data":{"show":{"_id":"abc","availableEpisodesDetail":{"sub":["3","1","2"]}}}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, response)
	}))
	defer server.Close()

	eps, err := newTestClient(server.URL).GetEpisodesList("abc", "sub")
	require.NoError(t, err)
	assert.Equal(t, []string{"1", "2", "3"}, eps)
}

func TestGetEpisodesListDefaultsToSub(t *testing.T) {
	t.Parallel()
	response := `{"data":{"show":{"_id":"abc","availableEpisodesDetail":{"sub":["1"]}}}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, response)
	}))
	defer server.Close()

	eps, err := newTestClient(server.URL).GetEpisodesList("abc", "")
	require.NoError(t, err)
	assert.Len(t, eps, 1)
}

func TestGetEpisodesListNoEpisodes(t *testing.T) {
	t.Parallel()
	response := `{"data":{"show":{"_id":"abc","availableEpisodesDetail":{"sub":[]}}}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, response)
	}))
	defer server.Close()

	eps, err := newTestClient(server.URL).GetEpisodesList("abc", "sub")
	require.NoError(t, err)
	assert.Empty(t, eps)
}

func TestGetEpisodesListRateLimited(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	_, err := newTestClient(server.URL).GetEpisodesList("abc", "sub")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable))
}

// ---------------------------------------------------------------------------
// 15. GetAnimeEpisodes — mock server tests
// ---------------------------------------------------------------------------

func TestGetAnimeEpisodesConvertsToModelFormat(t *testing.T) {
	t.Parallel()
	response := `{"data":{"show":{"_id":"abc","availableEpisodesDetail":{"sub":["1","2","3"]}}}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, response)
	}))
	defer server.Close()

	episodes, err := newTestClient(server.URL).GetAnimeEpisodes("abc")
	require.NoError(t, err)
	require.Len(t, episodes, 3)
	assert.Equal(t, "1", episodes[0].Number)
	assert.Equal(t, 1, episodes[0].Num)
	assert.Equal(t, "1", episodes[0].URL)
}

// ---------------------------------------------------------------------------
// 16. GetAnimeEpisodesWithAniSkip — mock tests
// ---------------------------------------------------------------------------

func TestGetAnimeEpisodesWithAniSkipEnrichesData(t *testing.T) {
	t.Parallel()
	response := `{"data":{"show":{"_id":"abc","availableEpisodesDetail":{"sub":["1","2"]}}}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, response)
	}))
	defer server.Close()

	callCount := 0
	aniSkipFunc := func(malID int, epNum int, ep *models.Episode) error {
		callCount++
		ep.SkipTimes = models.SkipTimes{
			Op: models.Skip{Start: 10, End: 90},
		}
		return nil
	}

	episodes, err := newTestClient(server.URL).GetAnimeEpisodesWithAniSkip("abc", 12345, aniSkipFunc)
	require.NoError(t, err)
	assert.Equal(t, 2, callCount)
	assert.Equal(t, 10, episodes[0].SkipTimes.Op.Start)
}

func TestGetAnimeEpisodesWithAniSkipHandlesError(t *testing.T) {
	t.Parallel()
	response := `{"data":{"show":{"_id":"abc","availableEpisodesDetail":{"sub":["1"]}}}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, response)
	}))
	defer server.Close()

	aniSkipFunc := func(_ int, _ int, _ *models.Episode) error {
		return fmt.Errorf("AniSkip API unavailable")
	}

	// Should not propagate error — just log it
	episodes, err := newTestClient(server.URL).GetAnimeEpisodesWithAniSkip("abc", 12345, aniSkipFunc)
	require.NoError(t, err)
	require.Len(t, episodes, 1)
}

func TestGetAnimeEpisodesWithAniSkipZeroMalID(t *testing.T) {
	t.Parallel()
	response := `{"data":{"show":{"_id":"abc","availableEpisodesDetail":{"sub":["1"]}}}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, response)
	}))
	defer server.Close()

	called := false
	aniSkipFunc := func(_ int, _ int, _ *models.Episode) error {
		called = true
		return nil
	}

	// malID=0 should skip calling aniSkipFunc
	_, err := newTestClient(server.URL).GetAnimeEpisodesWithAniSkip("abc", 0, aniSkipFunc)
	require.NoError(t, err)
	assert.False(t, called)
}

func TestGetAnimeEpisodesWithAniSkipNilFunc(t *testing.T) {
	t.Parallel()
	response := `{"data":{"show":{"_id":"abc","availableEpisodesDetail":{"sub":["1"]}}}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, response)
	}))
	defer server.Close()

	// nil aniSkipFunc should not panic
	episodes, err := newTestClient(server.URL).GetAnimeEpisodesWithAniSkip("abc", 12345, nil)
	require.NoError(t, err)
	require.Len(t, episodes, 1)
}

// ---------------------------------------------------------------------------
// 17. getLinks — mock server tests
// ---------------------------------------------------------------------------

func TestGetLinksSuccess(t *testing.T) {
	t.Parallel()
	linksJSON := buildLinksJSON(
		struct{ quality, url string }{"1080p", "https://cdn.example.com/1080.mp4"},
		struct{ quality, url string }{"720p", "https://cdn.example.com/720.mp4"},
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, linksJSON)
	}))
	defer server.Close()

	links, err := newTestClient(server.URL).getLinks(server.URL)
	require.NoError(t, err)
	assert.Contains(t, links, "1080p")
	assert.Contains(t, links, "720p")
}

func TestGetLinksEmptyJSON(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"links":[]}`)
	}))
	defer server.Close()

	links, err := newTestClient(server.URL).getLinks(server.URL)
	require.NoError(t, err)
	assert.Empty(t, links)
}

func TestGetLinksVerifiesRefererHeader(t *testing.T) {
	t.Parallel()
	var capturedReferer string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReferer = r.Header.Get("Referer")
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"links":[]}`)
	}))
	defer server.Close()

	_, _ = newTestClient(server.URL).getLinks(server.URL)
	assert.Equal(t, AllAnimeReferer, capturedReferer)
}

// ---------------------------------------------------------------------------
// 18. End-to-end: GetEpisodeURL with mock API + mock link server
// ---------------------------------------------------------------------------

func TestGetEpisodeURLEndToEndStandard(t *testing.T) {
	t.Parallel()

	// Link server that returns video links
	linkServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, buildLinksJSON(
			struct{ quality, url string }{"1080p", "https://cdn.example.com/ep1_1080.mp4"},
			struct{ quality, url string }{"720p", "https://cdn.example.com/ep1_720.mp4"},
		))
	}))
	defer linkServer.Close()

	// API server that returns source URLs pointing to the link server
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, buildSourceURLsJSON(
			struct{ url, name string }{linkServer.URL, "TestSource"},
		))
	}))
	defer apiServer.Close()

	url, meta, err := newTestClient(apiServer.URL).GetEpisodeURL("anime-id", "1", "sub", "best")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/ep1_1080.mp4", url)
	assert.Equal(t, "1080p", meta["quality"])
}

func TestGetEpisodeURLEndToEndToBeParsed(t *testing.T) {
	t.Parallel()

	// Link server
	linkServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, buildLinksJSON(
			struct{ quality, url string }{"720p", "https://cdn.example.com/ep1_720.mp4"},
		))
	}))
	defer linkServer.Close()

	// Real tobeparsed data has sourceUrl values that are "--" + hex-encoded.
	// Hex-encode the link server URL and prefix with "--" to match real API format.
	hexURL := hexEncodeSourceURL(linkServer.URL)
	plaintext := fmt.Sprintf(`{"data":{"episode":{"sourceUrls":[{"sourceUrl":"--%s","sourceName":"Encrypted"}]}}}`, hexURL)
	blob := encryptToBeParsed(t, plaintext)

	// API server returns tobeparsed response
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, buildToBeParsedResponse(blob))
	}))
	defer apiServer.Close()

	url, meta, err := newTestClient(apiServer.URL).GetEpisodeURL("anime-id", "1", "sub", "best")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/ep1_720.mp4", url)
	assert.Equal(t, "720p", meta["quality"])
}

func TestGetEpisodeURLEndToEndNoSourceURLs(t *testing.T) {
	t.Parallel()
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"data":{"episode":{"episodeString":"1","sourceUrls":[]}}}`)
	}))
	defer apiServer.Close()

	_, _, err := newTestClient(apiServer.URL).GetEpisodeURL("anime-id", "1", "sub", "best")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no source URLs")
}

func TestGetEpisodeURLDefaultsMode(t *testing.T) {
	t.Parallel()

	var capturedBody string
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		capturedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"data":{"episode":{"episodeString":"1","sourceUrls":[]}}}`)
	}))
	defer apiServer.Close()

	_, _, _ = newTestClient(apiServer.URL).GetEpisodeURL("anime-id", "1", "", "")
	assert.Contains(t, capturedBody, "sub")
}

// ---------------------------------------------------------------------------
// 19. Concurrent processing — race condition and timeout tests
// ---------------------------------------------------------------------------

func TestProcessSourceURLsConcurrentAllFail(t *testing.T) {
	t.Parallel()

	// All source URLs return errors
	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failServer.Close()

	client := newTestClient(failServer.URL)
	startedAt := time.Now()
	_, _, err := client.processSourceURLsConcurrent(
		[]string{failServer.URL + "/1", failServer.URL + "/2"},
		"best", "anime-id", "1",
	)
	require.Error(t, err)
	assert.Less(t, time.Since(startedAt), 2*time.Second, "failed sources should not wait for the global timeout")
	assert.NotContains(t, err.Error(), "timeout waiting for results")
}

func TestProcessSourceURLsConcurrentPartialFailure(t *testing.T) {
	t.Parallel()

	var reqCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := reqCount.Add(1)
		if n == 1 {
			// First request fails
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Second request succeeds
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, buildLinksJSON(
			struct{ quality, url string }{"720p", "https://cdn.example.com/720.mp4"},
		))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	startedAt := time.Now()
	url, _, err := client.processSourceURLsConcurrent(
		[]string{server.URL + "/fail", server.URL + "/ok"},
		"best", "anime-id", "1",
	)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/720.mp4", url)
	assert.Less(t, time.Since(startedAt), 2*time.Second, "successful fallback should not wait for the global timeout")
}

func TestProcessSourceURLsConcurrentFallsBackToFast4SpeedDirectSource(t *testing.T) {
	t.Parallel()

	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failServer.Close()

	client := newTestClient(failServer.URL)
	directURL := "https://tools.fast4speed.rsvp//media9/videos/gHQe2eBBh57QdC9hZ/sub/1"

	url, meta, err := client.processSourceURLsConcurrent(
		[]string{failServer.URL + "/clock.json", directURL},
		"worst", "gHQe2eBBh57QdC9hZ", "1",
	)

	require.NoError(t, err)
	assert.Equal(t, directURL, url)
	assert.Equal(t, "direct", meta["quality"])
	assert.Equal(t, "direct", meta["type"])
	assert.Equal(t, AllAnimeReferer, meta["referer"])
	assert.Equal(t, "gHQe2eBBh57QdC9hZ", meta["anime_id"])
	assert.Equal(t, "1", meta["episode"])
}

func TestProcessSourceURLsConcurrentHighPriorityWins(t *testing.T) {
	t.Parallel()

	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond) // Slow response
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, buildLinksJSON(
			struct{ quality, url string }{"1080p", "https://generic.com/1080.mp4"},
		))
	}))
	defer slowServer.Close()

	fastServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// High priority domain (sharepoint.com)
		_, _ = fmt.Fprint(w, buildLinksJSON(
			struct{ quality, url string }{"720p", "https://sharepoint.com/720.mp4"},
		))
	}))
	defer fastServer.Close()

	client := newTestClient("")
	url, meta, err := client.processSourceURLsConcurrent(
		[]string{slowServer.URL, fastServer.URL},
		"best", "anime-id", "1",
	)
	require.NoError(t, err)
	assert.Contains(t, url, "sharepoint.com")
	assert.Equal(t, "high", meta["priority"])
}

func TestProcessSourceURLsConcurrentRaceSafety(t *testing.T) {
	t.Parallel()

	// Test that concurrent processing doesn't race on shared state
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, buildLinksJSON(
			struct{ quality, url string }{"720p", "https://cdn.example.com/720.mp4"},
		))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	// Run multiple concurrent calls
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, _ = client.processSourceURLsConcurrent(
				[]string{server.URL + "/a", server.URL + "/b", server.URL + "/c"},
				"best", "anime-id", "1",
			)
		}()
	}
	wg.Wait() // If there's a race, -race detector will catch it
}

func TestProcessSourceURLsConcurrentSingleSource(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, buildLinksJSON(
			struct{ quality, url string }{"480p", "https://cdn.example.com/480.mp4"},
		))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	url, _, err := client.processSourceURLsConcurrent(
		[]string{server.URL},
		"best", "anime-id", "1",
	)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/480.mp4", url)
}

// ---------------------------------------------------------------------------
// 20. getLinks with priority domains — integration
// ---------------------------------------------------------------------------

func TestGetLinksAddsPriorityForKnownDomains(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, buildLinksJSON(
			struct{ quality, url string }{"1080p", "https://wixmp.com/video/1080.mp4"},
			struct{ quality, url string }{"720p", "https://generic.com/720.mp4"},
		))
	}))
	defer server.Close()

	links, err := newTestClient(server.URL).getLinks(server.URL)
	require.NoError(t, err)
	assert.Contains(t, links, "1080p_priority", "wixmp.com link should get _priority suffix")
	assert.NotContains(t, links, "720p_priority", "generic.com should not get _priority")
}

// ---------------------------------------------------------------------------
// 21. HTTP error codes — exhaustive coverage
// ---------------------------------------------------------------------------

func TestHTTPStatusCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code                int
		shouldBeUnavailable bool
	}{
		{200, false},
		{301, true}, // redirect is non-2xx
		{400, true}, // bad request
		{403, true}, // forbidden -> ErrSourceUnavailable
		{429, true}, // rate limited -> ErrSourceUnavailable
		{500, true}, // internal server error
		{503, true}, // service unavailable -> ErrSourceUnavailable
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.code), func(t *testing.T) {
			t.Parallel()
			resp := &http.Response{StatusCode: tt.code}
			err := checkHTTPStatus(resp, "test")
			if tt.code >= 200 && tt.code < 300 {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
			if tt.code == 403 || tt.code == 429 || tt.code == 503 {
				assert.True(t, errors.Is(err, ErrSourceUnavailable))
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 22. Slow / timeout server simulation
// ---------------------------------------------------------------------------

func TestGetEpisodeURLSlowAPIResponse(t *testing.T) {
	t.Parallel()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"data":{"episode":{"episodeString":"1","sourceUrls":[]}}}`)
	}))
	defer apiServer.Close()

	_, _, err := newTestClient(apiServer.URL).GetEpisodeURL("id", "1", "sub", "best")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no source URLs")
}

// ---------------------------------------------------------------------------
// 23. Intermittent failures — flaky server simulation
// ---------------------------------------------------------------------------

func TestGetLinksIntermittentFailures(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := callCount.Add(1)
		if n%2 == 1 {
			// Odd calls fail
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, buildLinksJSON(
			struct{ quality, url string }{"720p", "https://cdn.example.com/720.mp4"},
		))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	// First call fails
	_, err := client.getLinks(server.URL)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable))

	// Second call succeeds
	links, err := client.getLinks(server.URL)
	require.NoError(t, err)
	assert.Contains(t, links, "720p")
}

// ---------------------------------------------------------------------------
// 24. Edge case: server returns empty body
// ---------------------------------------------------------------------------

func TestGetLinksEmptyBody(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// empty body
	}))
	defer server.Close()

	links, err := newTestClient(server.URL).getLinks(server.URL)
	require.NoError(t, err)
	assert.Empty(t, links)
}

// ---------------------------------------------------------------------------
// 25. Hex substitution table completeness: round-trip via inverse map
// ---------------------------------------------------------------------------

func TestHexSubstitutionTableInverseRoundTrip(t *testing.T) {
	t.Parallel()

	// Build inverse map
	inverse := make(map[string]string, len(hexSubstitutionTable))
	for hex, char := range hexSubstitutionTable {
		_, exists := inverse[char]
		require.False(t, exists, "duplicate output %q in hexSubstitutionTable", char)
		inverse[char] = hex
	}

	// For every character in the table, encode then decode and verify round-trip.
	// decodeSourceURL applies post-processing: "/" gets the base URL prepended,
	// so we skip "/" — that behaviour is tested in TestDecodeSourceURLCompleteMapping.
	client := NewAllAnimeClient()
	for char, hex := range inverse {
		if char == "/" {
			continue // post-processing prepends base URL; tested elsewhere
		}
		decoded := client.decodeSourceURL(hex)
		assert.Equal(t, char, decoded, "round-trip failed for hex %q -> expected %q", hex, char)
	}
}

// ---------------------------------------------------------------------------
// 26. GetType and GetStreamURL interface methods
// ---------------------------------------------------------------------------

func TestGetTypeReturnsAllAnime(t *testing.T) {
	t.Parallel()
	assert.Equal(t, AllAnimeType, NewAllAnimeClient().GetType())
}

func TestGetStreamURLReturnsError(t *testing.T) {
	t.Parallel()
	_, _, err := NewAllAnimeClient().GetStreamURL("any-url")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not fully implemented")
}

func TestGetAnimeDetailsReturnsError(t *testing.T) {
	t.Parallel()
	_, err := NewAllAnimeClient().GetAnimeDetails("any-url")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// 27. Multiple concurrent GetEpisodeURL calls (race detector stress)
// ---------------------------------------------------------------------------

func TestGetEpisodeURLConcurrentCalls(t *testing.T) {
	t.Parallel()

	linkServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, buildLinksJSON(
			struct{ quality, url string }{"720p", "https://cdn.example.com/720.mp4"},
		))
	}))
	defer linkServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, buildSourceURLsJSON(
			struct{ url, name string }{linkServer.URL, "Src"},
		))
	}))
	defer apiServer.Close()

	client := newTestClient(apiServer.URL)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(ep int) {
			defer wg.Done()
			url, _, err := client.GetEpisodeURL("anime-id", fmt.Sprintf("%d", ep), "sub", "best")
			if err == nil {
				assert.NotEmpty(t, url)
			}
		}(i)
	}
	wg.Wait()
}

// ---------------------------------------------------------------------------
// 28. decodeToBeParsed with actual OpenSSL-generated blob (cross-validation)
// ---------------------------------------------------------------------------

func TestDecodeToBeParsedCrossValidateWithOpenSSL(t *testing.T) {
	// Deterministic test: verify our Go GCM encryption/decryption round-trips correctly.
	// Updated 2026-04-24: CTR → GCM, key rotated to "Xot36i3lK3:v1".
	// ani-cli reference: https://github.com/pystardust/ani-cli/commit/e5523a9b480f67ee878a0cc075043313cc58e07d
	t.Parallel()

	nonce, _ := hex.DecodeString("aabbccddeeff00112233aabb")
	plaintext := `{"data":{"episode":{"sourceUrls":[{"sourceUrl":"--504c4c484b021717","sourceName":"TestProvider"}]}}}`

	// Encrypt using new key + GCM
	key := sha256.Sum256([]byte("Xot36i3lK3:v1"))
	assert.Equal(t, allAnimeKey, key[:], "key derivation must be consistent")

	block, err := aes.NewCipher(key[:])
	require.NoError(t, err)
	gcm, err := cipher.NewGCM(block)
	require.NoError(t, err)

	sealed := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	payload := append([]byte{0x01}, nonce...)
	payload = append(payload, sealed...)
	blob := base64.StdEncoding.EncodeToString(payload)

	// Decrypt using production code
	sources, err := decodeToBeParsed(blob)
	require.NoError(t, err)
	require.Len(t, sources, 1)
	assert.Equal(t, "TestProvider", sources[0].sourceName)
	assert.Equal(t, "504c4c484b021717", sources[0].sourceURL)
}

// ---------------------------------------------------------------------------
// 29. SearchAnime with unknown episode format
// ---------------------------------------------------------------------------

func TestSearchAnimeUnknownEpisodeCount(t *testing.T) {
	t.Parallel()
	response := `{"data":{"shows":{"edges":[{"_id":"id1","name":"Mystery Show","englishName":"","availableEpisodes":{"sub":"unknown"}}]}}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, response)
	}))
	defer server.Close()

	results, err := newTestClient(server.URL).SearchAnime("mystery")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Contains(t, results[0].Name, "Unknown episodes")
}

func TestSearchAnimeNullEpisodes(t *testing.T) {
	t.Parallel()
	response := `{"data":{"shows":{"edges":[{"_id":"id1","name":"No Eps","englishName":"","availableEpisodes":null}]}}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, response)
	}))
	defer server.Close()

	results, err := newTestClient(server.URL).SearchAnime("test")
	require.NoError(t, err)
	require.Len(t, results, 1)
}

// ---------------------------------------------------------------------------
// 30. Server drops connection mid-response
// ---------------------------------------------------------------------------

func TestSearchAnimeServerDropsConnection(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Write partial response then close
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			t.Skip("server doesn't support hijacking")
		}
		conn, _, err := hijacker.Hijack()
		if err != nil {
			t.Skip("hijack failed")
		}
		_, _ = conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n{\"data\":"))
		_ = conn.Close()
	}))
	defer server.Close()

	_, err := newTestClient(server.URL).SearchAnime("test")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// 31. selectQuality edge: only priority entries
// ---------------------------------------------------------------------------

func TestSelectQualityOnlyPriorityEntries(t *testing.T) {
	t.Parallel()
	links := map[string]string{
		"720p_priority": "https://sharepoint.com/720.mp4",
	}
	url, meta := NewAllAnimeClient().selectQuality(links, "best")
	assert.NotEmpty(t, url)
	assert.Equal(t, "high", meta["priority"])
}

// ---------------------------------------------------------------------------
// 32. AES decryption does not panic on various malformed inputs
// ---------------------------------------------------------------------------

func TestDecodeToBeParsedNoPanicOnMalformed(t *testing.T) {
	t.Parallel()

	inputs := []string{
		"",                     // empty
		"AA==",                 // 1 byte
		"AAAAAAAAAAAAAAAA",     // 12 bytes exactly (nonce only, too short)
		"AAAAAAAAAAAAAAAAAAAA", // 15 bytes
		base64.StdEncoding.EncodeToString(make([]byte, 100)), // 100 zero bytes
		"YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo=",               // "abcdefghijklmnopqrstuvwxyz"
	}

	for i, input := range inputs {
		t.Run(fmt.Sprintf("input_%d", i), func(t *testing.T) {
			t.Parallel()
			// Must not panic
			_, _ = decodeToBeParsed(input)
		})
	}
}
