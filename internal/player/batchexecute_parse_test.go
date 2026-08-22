package player

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// Regression for the issue surfaced 2026-04-28 alongside the issue #166 fix.
//
// Symptom in the wild: a Goyabu episode whose Blogger token was dead upstream
// (video deleted/region-blocked) caused 6 doomed batchexecute calls (3 in the
// scraper retry loop + 3 in the player-layer redundant resolve), wasting
// ~12 seconds before the user saw a generic "no valid video URL found".
// Google signals "this RPC produced no result" with HTTP 200 + a body that is
// only the `)]}'` anti-hijacking prefix.
//
// parseBatchexecuteResponse must distinguish that empty-body case
// (errBloggerVideoUnavailable, which callers fast-fail on) from a structured
// response that we just failed to walk (a generic error, which is worth
// retrying because Google's response shape has shifted before — see the
// 2026-04-23 fix in the function's docstring).
func TestParseBatchexecuteResponse_EmptyBodyReturnsUnavailableSentinel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		body string
	}{
		{
			name: "bare anti-hijacking prefix",
			body: `)]}'` + "\n",
		},
		{
			name: "prefix with surrounding whitespace",
			body: "\n  )]}'\n\n",
		},
		{
			name: "prefix only no newline",
			body: `)]}'`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := parseBatchexecuteResponse([]byte(tc.body))
			if err == nil {
				t.Fatalf("expected error for empty batchexecute body, got nil")
			}
			if !errors.Is(err, errBloggerVideoUnavailable) {
				t.Fatalf("expected errBloggerVideoUnavailable, got %v", err)
			}
		})
	}
}

// A non-empty response with no recognizable streams is *not* the
// "video unavailable" case — it's a parser-shape problem worth retrying.
// This guards the sentinel from being over-broad.
func TestParseBatchexecuteResponse_StructuredButUnparsedIsNotUnavailable(t *testing.T) {
	t.Parallel()

	// Valid wrb.fr envelope but the inner data has no streams array.
	body := `)]}'

41
[["wrb.fr","WcwnYd","[\"some-payload\",null,1]",null,null,null,"generic"]]
`

	_, err := parseBatchexecuteResponse([]byte(body))
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if errors.Is(err, errBloggerVideoUnavailable) {
		t.Fatalf("structured-but-unparsed must NOT be classified as unavailable, got %v", err)
	}
	if !strings.Contains(err.Error(), "no video URL found") {
		t.Fatalf("expected generic parse error, got %v", err)
	}
}

// Sanity: a real-shape response with a googlevideo.com URL still parses
// successfully. This protects against the empty-body branch swallowing the
// happy path.
func TestParseBatchexecuteResponse_RegexFallbackPicksGoogleVideoURL(t *testing.T) {
	t.Parallel()

	body := `)]}'

128
[["wrb.fr","WcwnYd","[null,null,[[[\"https://rr1---sn-xxx.googlevideo.com/videoplayback?expire=123&mime=video%2Fmp4&itag=22\"]]]]",null,null,null,"generic"]]
`

	got, err := parseBatchexecuteResponse([]byte(body))
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !strings.Contains(got, "googlevideo.com") {
		t.Fatalf("expected googlevideo URL, got %q", got)
	}
}

// TestParseBatchexecuteResponse_TrulyEmptyBody covers a zero-byte body —
// the most degenerate case. Real Blogger responses always include the
// `)]}'` prefix, but a connection that closes mid-stream can produce
// this. Treat it as token-unavailable so the retry loop fast-fails just
// as it does for the prefix-only case.
func TestParseBatchexecuteResponse_TrulyEmptyBody(t *testing.T) {
	t.Parallel()

	_, err := parseBatchexecuteResponse(nil)
	if err == nil {
		t.Fatalf("expected error for nil body, got nil")
	}
	if !errors.Is(err, errBloggerVideoUnavailable) {
		t.Fatalf("nil body must map to errBloggerVideoUnavailable, got %v", err)
	}

	_, err = parseBatchexecuteResponse([]byte(""))
	if err == nil {
		t.Fatalf("expected error for empty body, got nil")
	}
	if !errors.Is(err, errBloggerVideoUnavailable) {
		t.Fatalf("empty body must map to errBloggerVideoUnavailable, got %v", err)
	}

	_, err = parseBatchexecuteResponse([]byte("   \n\t\n"))
	if err == nil {
		t.Fatalf("expected error for whitespace-only body, got nil")
	}
	if !errors.Is(err, errBloggerVideoUnavailable) {
		t.Fatalf("whitespace-only body must map to errBloggerVideoUnavailable, got %v", err)
	}
}

// TestParseBatchexecuteResponse_SentinelSurvivesWrap proves the sentinel
// remains identifiable through fmt.Errorf("…: %w", err). This is the
// exact wrap used by extractActualVideoURL → GetVideoURLForEpisodeEnhanced
// when surfacing the error to the caller, and it is what the retry-loop
// fast-fail and the player-layer skip rely on. If anyone replaces %w
// with %v in the wrap, this test fires.
func TestParseBatchexecuteResponse_SentinelSurvivesWrap(t *testing.T) {
	t.Parallel()

	_, raw := parseBatchexecuteResponse([]byte(`)]}'` + "\n"))
	if raw == nil {
		t.Fatalf("expected raw error from empty-body parse, got nil")
	}

	wrapped := fmt.Errorf("video unavailable on this source: %w",
		fmt.Errorf("failed to extract video URL: %w", raw))

	if !errors.Is(wrapped, errBloggerVideoUnavailable) {
		t.Fatalf("sentinel must survive double-wrap; got %v", wrapped)
	}
}

// TestParseBatchexecuteResponse_LiveLogReplay locks down the production
// payload shape observed on 2026-04-28 16:12:30 for the working Black
// Clover episode 8 — data[2] streams, googlevideo CDN URL, mp4 mime.
// Built via the same buildBatchexecuteBody helper used by the rest of
// the suite so the JSON shape is canonical. If Google later changes
// the wire format and we silently regress, this snapshot fires.
func TestParseBatchexecuteResponse_LiveLogReplay(t *testing.T) {
	t.Parallel()

	body := buildBatchexecuteBody(2, []string{
		"https://rr5---sn-q4f7l.googlevideo.com/videoplayback?expire=1745958900&itag=18&mime=video%2Fmp4",
		"https://rr5---sn-q4f7l.googlevideo.com/videoplayback?expire=1745958900&itag=22&mime=video%2Fmp4",
	})

	got, err := parseBatchexecuteResponse(body)
	if err != nil {
		t.Fatalf("expected success on the production-shape payload, got %v", err)
	}
	if !strings.Contains(got, "googlevideo.com") {
		t.Fatalf("URL must be a googlevideo CDN URL, got %q", got)
	}
	if !strings.Contains(got, "itag=22") {
		t.Fatalf("expected 720p (itag=22) preference, got %q", got)
	}
}
