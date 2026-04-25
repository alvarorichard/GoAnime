// ===========================================================================
// goyabu_blogger_fix_test.go — Regression tests for the Goyabu/Blogger batchexecute bug
//
// Issue detected: 2026-04-23
//   Diiver reported via Discord that Goyabu (PT-BR source) episodes were not
//   playing. The debug log showed the error
//   "no video URL found in batchexecute response" on all 3 attempts
//   of the Blogger URL extractor.
//
// Root cause (player/scraper.go — before the 2026-04-23 fix):
//   The batchexecute parser assumed the streams array was always at index
//   data[2] of the inner JSON payload. When Google changed the index
//   (or returned data with fewer than 3 elements), the code silently executed
//   `continue` without extracting any URL. There was no regex fallback.
//
//   Buggy code (removed on 2026-04-23):
//     if len(data) < 3 { continue }        // fails when data has 0-2 elements
//     streams, ok := data[2].([]any)       // hardcoded index — breaks when Google changes it
//
// Fix applied: 2026-04-23
//   1. The parser iterates all indices of data[], identifying streams as
//      the first element that is an array of arrays.
//   2. Regex fallback extracts *.googlevideo.com URLs directly from the raw body
//      if the structured parsing produces no result.
//
// Functions tested:
//   - parseBatchexecuteResponse (real — player/scraper.go, fix 2026-04-23)
//   - parseBatchexecuteResponseLegacy (pre-fix logic, inlined here)
// ===========================================================================

package player

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// URL constants used in test fixtures
// ---------------------------------------------------------------------------

const (
	googleVideoURL720p = "https://rr3---sn-q4f7l.googlevideo.com/videoplayback?expire=9999999999&itag=22&mime=video%2Fmp4&source=blogger"
	googleVideoURL360p = "https://rr3---sn-q4f7l.googlevideo.com/videoplayback?expire=9999999999&itag=18&mime=video%2Fmp4&source=blogger"
	googleVideoNoMIME  = "https://rr3---sn-q4f7l.googlevideo.com/videoplayback?expire=9999999999&itag=22&source=blogger"
)

// ---------------------------------------------------------------------------
// parseBatchexecuteResponseLegacy — logic BEFORE the 2026-04-23 fix
//
// Kept here to demonstrate the bug: only checks data[2] and has no regex
// fallback. Any response with streams outside index 2 returns an error.
// ---------------------------------------------------------------------------
func parseBatchexecuteResponseLegacy(body []byte) (string, error) {
	var videoURL string
	for _, line := range strings.Split(string(body), "\n") {
		if !strings.Contains(line, "wrb.fr") {
			continue
		}
		var outer []any
		if err := json.Unmarshal([]byte(line), &outer); err != nil {
			continue
		}
		for _, entry := range outer {
			arr, ok := entry.([]any)
			if !ok || len(arr) < 3 {
				continue
			}
			if fmt.Sprint(arr[0]) != "wrb.fr" || fmt.Sprint(arr[1]) != "WcwnYd" {
				continue
			}
			var data []any
			if err := json.Unmarshal(fmt.Append(nil, arr[2]), &data); err != nil {
				continue
			}
			// BUG (before 2026-04-23): hardcoded index; fails when data has
			// fewer than 3 elements or if Google moves streams to a different index.
			if len(data) < 3 {
				continue
			}
			streams, ok := data[2].([]any)
			if !ok {
				continue
			}
			for _, s := range streams {
				stream, ok := s.([]any)
				if !ok || len(stream) < 1 {
					continue
				}
				u, ok := stream[0].(string)
				if !ok {
					continue
				}
				if strings.Contains(u, "mime=video%2Fmp4") || strings.Contains(u, "mime=video/mp4") {
					videoURL = u
					break
				}
			}
			break
		}
		if videoURL != "" {
			break
		}
	}
	if videoURL == "" {
		return "", errors.New("no video URL found in batchexecute response")
	}
	return videoURL, nil
}

// ---------------------------------------------------------------------------
// buildBatchexecuteBody builds a realistic batchexecute response body.
//
// The actual Google format is:
//
//	)]}'\n
//	[["wrb.fr","WcwnYd","<inner_json_string>",null,...,"generic"]]\n
//
// where <inner_json_string> is the inner JSON payload serialized as a string.
// streams is placed at data[dataIdx] in the inner payload.
// ---------------------------------------------------------------------------
func buildBatchexecuteBody(dataIdx int, streamURLs []string) []byte {
	streams := make([]any, len(streamURLs))
	for i, u := range streamURLs {
		streams[i] = []any{u, float64(360)}
	}

	data := make([]any, dataIdx+1)
	data[dataIdx] = streams

	innerJSON, _ := json.Marshal(data)

	outerLine, _ := json.Marshal([][]any{
		{"wrb.fr", "WcwnYd", string(innerJSON), nil, nil, nil, "generic"},
	})

	return []byte(")]}'\n" + string(outerLine) + "\n")
}

// ---------------------------------------------------------------------------
// Tests: classic format (streams at data[2]) — both parsers work
// ---------------------------------------------------------------------------

func TestParseBatchexecuteResponse_StreamsAtIndex2_FormatoClassico(t *testing.T) {
	t.Parallel()

	body := buildBatchexecuteBody(2, []string{googleVideoURL720p, googleVideoURL360p})

	// Legacy parser (pre-fix) works because streams are at data[2].
	legacyURL, err := parseBatchexecuteResponseLegacy(body)
	require.NoError(t, err, "legacy logic should work with streams at data[2]")
	assert.Equal(t, googleVideoURL720p, legacyURL)

	// Fixed parser also works.
	fixedURL, err := parseBatchexecuteResponse(body)
	require.NoError(t, err)
	assert.Equal(t, googleVideoURL720p, fixedURL)
}

// ---------------------------------------------------------------------------
// Tests: simulation of bug 2026-04-23 — streams at index != 2
// ---------------------------------------------------------------------------

// TestParseBatchexecuteResponse_Bug_StreamsEmIndex0 demonstrates the bug:
// when Google returns streams at data[0], the legacy parser fails while
// the fixed parser (fix 2026-04-23) finds the URL.
func TestParseBatchexecuteResponse_Bug_StreamsEmIndex0(t *testing.T) {
	t.Parallel()

	body := buildBatchexecuteBody(0, []string{googleVideoURL720p})

	// Simulated BUG: legacy parser (pre 2026-04-23) fails with streams at data[0].
	_, err := parseBatchexecuteResponseLegacy(body)
	assert.Error(t, err, "BUG 2026-04-23: legacy parser does NOT find streams at data[0]")

	// FIX: updated parser finds streams at any index.
	url, err := parseBatchexecuteResponse(body)
	require.NoError(t, err, "fix 2026-04-23: parser should find streams at data[0]")
	assert.Equal(t, googleVideoURL720p, url)
}

// TestParseBatchexecuteResponse_Bug_StreamsEmIndex1 covers the case where
// Google returns only two elements in data (indices 0 and 1).
func TestParseBatchexecuteResponse_Bug_StreamsEmIndex1(t *testing.T) {
	t.Parallel()

	body := buildBatchexecuteBody(1, []string{googleVideoURL360p})

	// Simulated BUG: legacy parser fails (data has 2 elements, len(data) < 3).
	_, err := parseBatchexecuteResponseLegacy(body)
	assert.Error(t, err, "BUG 2026-04-23: legacy parser fails when len(data) < 3")

	// FIX: updated parser finds streams at data[1].
	url, err := parseBatchexecuteResponse(body)
	require.NoError(t, err, "fix 2026-04-23: should find streams at data[1]")
	assert.Equal(t, googleVideoURL360p, url)
}

// TestParseBatchexecuteResponse_Bug_DataComUmElemento covers the edge case
// where data has only one element (data[0] = streams).
func TestParseBatchexecuteResponse_Bug_DataComUmElemento(t *testing.T) {
	t.Parallel()

	body := buildBatchexecuteBody(0, []string{googleVideoURL360p})

	_, err := parseBatchexecuteResponseLegacy(body)
	assert.Error(t, err, "BUG 2026-04-23: len(data)==1 < 3, legacy parser discards it")

	url, err := parseBatchexecuteResponse(body)
	require.NoError(t, err)
	assert.Equal(t, googleVideoURL360p, url)
}

// ---------------------------------------------------------------------------
// Tests: quality selection
// ---------------------------------------------------------------------------

func TestParseBatchexecuteResponse_Prefere720p(t *testing.T) {
	t.Parallel()

	// Resposta com 360p listado antes do 720p.
	body := buildBatchexecuteBody(2, []string{googleVideoURL360p, googleVideoURL720p})

	url, err := parseBatchexecuteResponse(body)
	require.NoError(t, err)
	assert.Equal(t, googleVideoURL720p, url, "should prefer 720p (itag=22) over 360p (itag=18)")
}

func TestParseBatchexecuteResponse_Only360p(t *testing.T) {
	t.Parallel()

	body := buildBatchexecuteBody(2, []string{googleVideoURL360p})

	url, err := parseBatchexecuteResponse(body)
	require.NoError(t, err)
	assert.Equal(t, googleVideoURL360p, url)
}

// ---------------------------------------------------------------------------
// Tests: regex fallback (fix 2026-04-23)
// ---------------------------------------------------------------------------

// TestParseBatchexecuteResponse_FallbackRegex verifies that when structured
// parsing finds no streams, the regex still captures googlevideo.com URLs
// present in the raw response body.
func TestParseBatchexecuteResponse_FallbackRegex(t *testing.T) {
	t.Parallel()

	// Raw body with a googlevideo.com URL but no valid wrb.fr/WcwnYd structure.
	body := []byte(`)]}'\n` +
		`[["wrb.fr","WcwnYd","{}",null,null,null,"generic"]]` + "\n" +
		`debug: ` + googleVideoURL720p + ` extra`)

	// Legacy parser (no regex fallback) fails.
	_, err := parseBatchexecuteResponseLegacy(body)
	assert.Error(t, err, "legacy parser has no regex fallback")

	// Fixed parser finds URL via regex.
	url, err := parseBatchexecuteResponse(body)
	require.NoError(t, err, "fix 2026-04-23: regex fallback should find googlevideo.com URL")
	assert.Contains(t, url, ".googlevideo.com")
}

// TestParseBatchexecuteResponse_FallbackRegex_StreamsSemMIME covers streams
// that exist in the structure but lack mime=video%2Fmp4, while the raw body
// contains a valid googlevideo.com URL.
func TestParseBatchexecuteResponse_FallbackRegex_StreamsSemMIME(t *testing.T) {
	t.Parallel()

	body := buildBatchexecuteBody(2, []string{googleVideoNoMIME})
	// Appends a valid URL as free text at the end of the body.
	body = append(body, []byte("\ninfo: "+googleVideoURL360p)...)

	url, err := parseBatchexecuteResponse(body)
	require.NoError(t, err)
	assert.Contains(t, url, ".googlevideo.com")
}

// ---------------------------------------------------------------------------
// Tests: error cases / invalid response
// ---------------------------------------------------------------------------

func TestParseBatchexecuteResponse_CorpoVazio(t *testing.T) {
	t.Parallel()

	_, err := parseBatchexecuteResponse([]byte{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no video URL found")
}

func TestParseBatchexecuteResponse_JSONInvalido(t *testing.T) {
	t.Parallel()

	body := []byte(")]}'\n[this is not valid json]\n")

	_, err := parseBatchexecuteResponse(body)
	assert.Error(t, err)
}

func TestParseBatchexecuteResponse_SemLinhaWrbFr(t *testing.T) {
	t.Parallel()

	body := []byte(")]}'\n[[\"other.method\",\"SomeRPC\",\"[]\",null]]\n")

	_, err := parseBatchexecuteResponse(body)
	assert.Error(t, err)
}

func TestParseBatchexecuteResponse_DataSemArrayDeArrays(t *testing.T) {
	t.Parallel()

	// data contains only strings and numbers, no array of arrays.
	innerJSON := `["string_value", 42, null]`
	outerLine, _ := json.Marshal([][]any{
		{"wrb.fr", "WcwnYd", innerJSON, nil, nil, nil, "generic"},
	})
	body := []byte(")]}'\n" + string(outerLine) + "\n")

	_, err := parseBatchexecuteResponse(body)
	assert.Error(t, err)
}

func TestParseBatchexecuteResponse_StreamsVazios(t *testing.T) {
	t.Parallel()

	// data[2] is an empty array — does not satisfy len(s) > 0.
	innerJSON := `[null, null, []]`
	outerLine, _ := json.Marshal([][]any{
		{"wrb.fr", "WcwnYd", innerJSON, nil, nil, nil, "generic"},
	})
	body := []byte(")]}'\n" + string(outerLine) + "\n")

	_, err := parseBatchexecuteResponse(body)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: multiple indices — parser should use the first valid one found
// ---------------------------------------------------------------------------

func TestParseBatchexecuteResponse_UsaPrimeiroIndiceValido(t *testing.T) {
	t.Parallel()

	// data[1] = streams with 360p; data[3] = streams with 720p.
	// Parser should return the first valid one found (data[1], 360p).
	streams360 := []any{[]any{googleVideoURL360p, float64(360)}}
	streams720 := []any{[]any{googleVideoURL720p, float64(720)}}

	data := []any{nil, streams360, nil, streams720}
	innerJSON, _ := json.Marshal(data)
	outerLine, _ := json.Marshal([][]any{
		{"wrb.fr", "WcwnYd", string(innerJSON), nil, nil, nil, "generic"},
	})
	body := []byte(")]}'\n" + string(outerLine) + "\n")

	url, err := parseBatchexecuteResponse(body)
	require.NoError(t, err)
	// Streams at data[1] (360p) should be found first.
	assert.Equal(t, googleVideoURL360p, url)
}
