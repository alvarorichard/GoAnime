// Package scraper provides web scraping functionality for anime sources
package allanime

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
)

// buildAAReq builds the "aaReq" proof token AllAnime requires on the
// episode-sources query — its absence yields an AA_CRYPTO_MISSING error
// (ani-cli PR #1772/#1779). The token is:
//
//	base64( 0x01 || iv(12) || AES-256-GCM(key, iv, payload) )
//
// where payload is a small JSON blob pinned to the current 5-minute time
// window and iv is the first 12 bytes of SHA-256 over the epoch + qh + window.
// The server recomputes and validates it, so byte-exactness of both the payload
// and the seed matters. Key and epoch come from the per-epoch derived material
// (see fetchAAKeys).
func buildAAReq(qh string, keys *aaKeys) (string, error) {
	return buildAAReqAt(qh, keys.key, keys.epoch, time.Now().UnixMilli())
}

// aaReqWindowMillis is the clock-bucket width the token's timestamp is rounded
// to; it must match the server's window or the token is rejected.
const aaReqWindowMillis = 300000 // 5 minutes

// buildAAReqAt is buildAAReq with the key/epoch and an injectable wall-clock
// (milliseconds since epoch) passed explicitly, so the token is deterministic
// in tests. Production derives key/epoch from fetchAAKeys and passes time.Now.
func buildAAReqAt(qh string, key []byte, epoch string, nowMillis int64) (string, error) {
	ts := (nowMillis / aaReqWindowMillis) * aaReqWindowMillis

	payload := fmt.Sprintf(`{"v":1,"ts":%d,"epoch":%s,"qh":%q}`, ts, epoch, qh)

	ivSeed := fmt.Sprintf("%s:%s:%d", epoch, qh, ts)
	ivHash := sha256.Sum256([]byte(ivSeed))
	iv := ivHash[:12]

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aaReq cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("aaReq gcm: %w", err)
	}
	// Seal returns ciphertext with the 16-byte tag appended (ct || tag),
	// matching ani-cli's Buffer.concat([ct, tag]).
	sealed := gcm.Seal(nil, iv, []byte(payload), nil)

	out := make([]byte, 0, 1+len(iv)+len(sealed))
	out = append(out, 0x01)
	out = append(out, iv...)
	out = append(out, sealed...)
	return base64.StdEncoding.EncodeToString(out), nil
}

// sourceInfo holds a decoded source URL and its provider name.
type sourceInfo struct {
	sourceName string
	sourceURL  string
}

// sourceEntry pairs a decoded source URL with its provider name so callers
// can dispatch by provider (e.g. filemoon needs its own decryption flow,
// while everything else goes through the generic getLinks path).
type sourceEntry struct {
	URL  string
	Name string
}

// decodeToBeParsed decrypts the "tobeparsed" blob from the AllAnime API.
//
// Blob format (restored to AES-256-GCM 2026-07-08, ani-cli PR #1772):
//
//	base64( 0x01 || nonce(12) || ciphertext || tag(16) )
//
// The trailing 16 bytes are a valid GCM auth tag again (they were a discarded
// dead slot during the 2026-04 CTR interlude), so Open both decrypts and
// authenticates in one step. Minimum valid size: 1 + 12 + 0 + 16 = 29 bytes.
func decodeToBeParsed(blob string, key []byte) ([]sourceInfo, error) {
	util.Debugf("AllAnime tobeparsed raw blob (first 60 chars): %q", blob[:min(60, len(blob))])

	// Try standard base64 first, then URL-safe (AllAnime may use either)
	data, err := base64.StdEncoding.DecodeString(blob)
	if err != nil {
		data, err = base64.URLEncoding.DecodeString(blob)
		if err != nil {
			data, err = base64.RawURLEncoding.DecodeString(blob)
			if err != nil {
				return nil, fmt.Errorf("base64 decode failed: %w", err)
			}
		}
	}

	util.Debugf("AllAnime tobeparsed decoded length: %d bytes, first 16 bytes: %x", len(data), data[:min(16, len(data))])

	// 1 (version) + 12 (nonce) + 16 (tag) = 29 bytes minimum (empty plaintext).
	if len(data) < 29 {
		return nil, fmt.Errorf("tobeparsed blob too short (%d bytes)", len(data))
	}

	// Slice: [version][nonce(12)][ciphertext || tag]. Go's GCM Open expects the
	// tag appended to the ciphertext, which is exactly the wire layout.
	nonce := data[1:13]
	sealed := data[13:]
	util.Debugf("AllAnime tobeparsed nonce: %x", nonce)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	plaintext, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return nil, fmt.Errorf("tobeparsed GCM decrypt failed: %w", err)
	}

	// Parse the decrypted JSON to extract sourceUrl/sourceName pairs.
	// The bash script does:
	//   sed -nE 's|.*"sourceUrl":"--([^"]*)".*"sourceName":"([^"]*)".*|\2 :\1|p'
	// We parse each sourceUrls entry from the JSON structure.
	var result struct {
		Data struct {
			Episode struct {
				SourceUrls []struct {
					SourceURL  string `json:"sourceUrl"`
					SourceName string `json:"sourceName"`
				} `json:"sourceUrls"`
			} `json:"episode"`
		} `json:"data"`
	}

	util.Debugf("AllAnime tobeparsed decrypted (first 200 bytes): %q", string(plaintext[:min(200, len(plaintext))]))

	// The plaintext may contain the full GraphQL response or just the sourceUrls array.
	// Try parsing as the full response first.
	if err := json.Unmarshal(plaintext, &result); err == nil && len(result.Data.Episode.SourceUrls) > 0 {
		var sources []sourceInfo
		for _, su := range result.Data.Episode.SourceUrls {
			url := su.SourceURL
			url = strings.TrimPrefix(url, "--")
			sources = append(sources, sourceInfo{
				sourceName: su.SourceName,
				sourceURL:  url,
			})
		}
		return sources, nil
	}

	// Fallback: try to extract using regex (like the bash sed pattern).
	// The plaintext might not be perfectly structured JSON.
	matches := sourceURLNameRe.FindAllSubmatch(plaintext, -1)
	if len(matches) == 0 {
		// Also try reverse order (sourceName before sourceUrl)
		matches = sourceNameURLRe.FindAllSubmatch(plaintext, -1)
		if len(matches) == 0 {
			return nil, fmt.Errorf("no source URLs found in decrypted tobeparsed data")
		}
		var sources []sourceInfo
		for _, m := range matches {
			sources = append(sources, sourceInfo{
				sourceName: string(m[1]),
				sourceURL:  string(m[2]),
			})
		}
		return sources, nil
	}

	var sources []sourceInfo
	for _, m := range matches {
		sources = append(sources, sourceInfo{
			sourceName: string(m[2]),
			sourceURL:  string(m[1]),
		})
	}
	return sources, nil
}

// Pre-compiled patterns for decrypted source extraction.
var (
	sourceURLNameRe = regexp.MustCompile(`"sourceUrl"\s*:\s*"--([^"]*)"[^}]*"sourceName"\s*:\s*"([^"]*)"`)
	sourceNameURLRe = regexp.MustCompile(`"sourceName"\s*:\s*"([^"]*)"[^}]*"sourceUrl"\s*:\s*"--([^"]*)"`)
	toBeParsedRe    = regexp.MustCompile(`"tobeparsed"\s*:\s*"([^"]*)"`)
)

// extractToBeParsedBlob extracts the base64 "tobeparsed" value from the API response JSON.
func extractToBeParsedBlob(response string) string {
	match := toBeParsedRe.FindStringSubmatch(response)
	if len(match) >= 2 {
		return match[1]
	}
	return ""
}

// hexSubstitutionTable is the complete hex-pair substitution cipher from ani-cli's provider_init.
// Each 2-char hex pair maps to its decoded ASCII character.
var hexSubstitutionTable = map[string]string{
	// Uppercase letters
	"79": "A", "7a": "B", "7b": "C", "7c": "D", "7d": "E", "7e": "F", "7f": "G",
	"70": "H", "71": "I", "72": "J", "73": "K", "74": "L", "75": "M", "76": "N", "77": "O",
	"68": "P", "69": "Q", "6a": "R", "6b": "S", "6c": "T", "6d": "U", "6e": "V", "6f": "W",
	"60": "X", "61": "Y", "62": "Z",
	// Lowercase letters
	"59": "a", "5a": "b", "5b": "c", "5c": "d", "5d": "e", "5e": "f", "5f": "g",
	"50": "h", "51": "i", "52": "j", "53": "k", "54": "l", "55": "m", "56": "n", "57": "o",
	"48": "p", "49": "q", "4a": "r", "4b": "s", "4c": "t", "4d": "u", "4e": "v", "4f": "w",
	"40": "x", "41": "y", "42": "z",
	// Digits
	"08": "0", "09": "1", "0a": "2", "0b": "3", "0c": "4", "0d": "5", "0e": "6", "0f": "7",
	"00": "8", "01": "9",
	// Special characters
	"15": "-", "16": ".", "67": "_", "46": "~",
	"02": ":", "17": "/", "07": "?", "1b": "#",
	"63": "[", "65": "]", "78": "@",
	"19": "!", "1c": "$", "1e": "&",
	"10": "(", "11": ")", "12": "*", "13": "+", "14": ",",
	"03": ";", "05": "=", "1d": "%",
}

// decodeSourceURL decodes the encoded source URL using the hex substitution cipher from ani-cli
func (c *AllAnimeClient) decodeSourceURL(encoded string) string {
	// Split into 2-char hex pairs and substitute
	var result strings.Builder
	result.Grow(len(encoded))
	for i := 0; i+1 < len(encoded); i += 2 {
		pair := encoded[i : i+2]
		if val, exists := hexSubstitutionTable[pair]; exists {
			result.WriteString(val)
		} else {
			result.WriteString(pair)
		}
	}

	decoded := result.String()

	// Replace "/clock" with "/clock.json" like in ani-cli
	decoded = strings.ReplaceAll(decoded, "/clock", "/clock.json")

	// If it starts with /, it's a path that needs the AllAnime base
	if strings.HasPrefix(decoded, "/") {
		decoded = fmt.Sprintf("https://%s%s", AllAnimeBase, decoded)
	}

	return decoded
}

// decryptFilemoonPayload performs the AES-256-CTR decrypt with key=kp1||kp2
// and counter=iv||0x00000002, then strips the trailing 16 padding bytes.
func decryptFilemoonPayload(w filemoonResponse) ([]byte, error) {
	iv, err := decodeFilemoonField(w.IV)
	if err != nil {
		return nil, fmt.Errorf("filemoon: decode iv: %w", err)
	}
	if len(iv) != 12 {
		return nil, fmt.Errorf("filemoon: iv must be 12 bytes, got %d", len(iv))
	}

	kp1, err := decodeFilemoonField(w.KeyParts[0])
	if err != nil {
		return nil, fmt.Errorf("filemoon: decode key_parts[0]: %w", err)
	}
	kp2, err := decodeFilemoonField(w.KeyParts[1])
	if err != nil {
		return nil, fmt.Errorf("filemoon: decode key_parts[1]: %w", err)
	}
	key := append([]byte{}, kp1...)
	key = append(key, kp2...)
	if len(key) != 32 {
		return nil, fmt.Errorf("filemoon: key must be 32 bytes (AES-256), got %d", len(key))
	}

	payload, err := decodeFilemoonField(w.Payload)
	if err != nil {
		return nil, fmt.Errorf("filemoon: decode payload: %w", err)
	}
	if len(payload) < 16 {
		return nil, fmt.Errorf("filemoon: payload too short (%d < 16)", len(payload))
	}
	ciphertext := payload[:len(payload)-16]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("filemoon: cipher init: %w", err)
	}
	// Counter = iv (12 bytes per-message, decoded from wire) || 0x00000002.
	// gosec flags the 0x02 as a hardcoded IV — false positive: this is the
	// decryption path; the iv comes from the encrypted response and the
	// trailing counter is fixed by the protocol (GCM J0+1 convention).
	counter := make([]byte, 16)
	copy(counter[:12], iv)
	counter[15] = 0x02

	plaintext := make([]byte, len(ciphertext))
	cipher.NewCTR(block, counter).XORKeyStream(plaintext, ciphertext) // #nosec G407 -- decrypt path; iv is per-message from response, 0x02 is protocol-fixed counter
	return plaintext, nil
}

// decodeFilemoonField accepts either base64url-no-pad (the canonical
// AllAnime form) or padded base64url, since edge nodes occasionally pad.
func decodeFilemoonField(s string) ([]byte, error) {
	if b, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return base64.URLEncoding.DecodeString(s)
}
