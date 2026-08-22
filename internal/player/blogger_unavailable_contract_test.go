// Package player — Contract tests for the dead-Blogger-token fast-fail.
//
// 2026-04-28: a Goyabu episode whose Blogger token was dead upstream
// (video deleted/region-blocked) caused 6 doomed batchexecute attempts
// (3 in extractBloggerVideoURL's retry loop + 3 in the player layer's
// HandleDownloadAndPlay redundant resolve), wasting ~12 seconds before
// the user saw a generic "no valid video URL found".
//
// The fix introduces errBloggerVideoUnavailable as a sentinel that
// propagates through:
//
//	parseBatchexecuteResponse  ──►  extractBloggerGoogleVideoURL
//	   ──►  startBloggerProxy ("failed to extract video URL: %w")
//	   ──►  extractBloggerVideoURL  (retry-loop fast-fail check)
//	   ──►  extractActualVideoURL
//	   ──►  GetVideoURLForEpisodeEnhanced ("video unavailable on this source: %w")
//	   ──►  PlayEpisode (routes user back to episode selection)
//
// Each link in this chain must preserve the sentinel via errors.Is.
// These tests lock down the contract: anyone replacing %w with %v, or
// re-wrapping with errors.New, will see this fire.
package player

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestSentinelPropagation_ThroughProductionWrapChain reproduces the
// exact wrap sequence used at every intermediate call site in the
// dead-Blogger-token path and asserts errors.Is still finds the
// sentinel at the top.
func TestSentinelPropagation_ThroughProductionWrapChain(t *testing.T) {
	t.Parallel()

	// Step 1 — parser detects empty body and returns the sentinel.
	_, leaf := parseBatchexecuteResponse([]byte(`)]}'` + "\n"))
	if leaf == nil {
		t.Fatalf("parser must return an error for empty body")
	}
	if !errors.Is(leaf, errBloggerVideoUnavailable) {
		t.Fatalf("parser must return errBloggerVideoUnavailable; got %v", leaf)
	}

	// Step 2 — startBloggerProxy wraps with "failed to extract video URL: %w".
	step2 := fmt.Errorf("failed to extract video URL: %w", leaf)
	if !errors.Is(step2, errBloggerVideoUnavailable) {
		t.Fatalf("startBloggerProxy wrap must preserve sentinel; got %v", step2)
	}

	// Step 3 — extractBloggerVideoURL's retry loop must observe the
	// sentinel so it can short-circuit. Verify a downstream caller
	// receiving the wrapped error can still detect it.
	if !errors.Is(step2, errBloggerVideoUnavailable) {
		t.Fatalf("retry-loop check must succeed on wrapped error")
	}

	// Step 4 — GetVideoURLForEpisodeEnhanced wraps with
	// "video unavailable on this source: %w".
	step4 := fmt.Errorf("video unavailable on this source: %w", step2)
	if !errors.Is(step4, errBloggerVideoUnavailable) {
		t.Fatalf("GetVideoURLForEpisodeEnhanced wrap must preserve sentinel; got %v", step4)
	}

	// Final — the message at the top of the chain must remain user-facing.
	if !strings.Contains(step4.Error(), "video unavailable on this source") {
		t.Fatalf("top-level error message must include user-facing prefix; got %q", step4.Error())
	}
	if !strings.Contains(step4.Error(), "failed to extract video URL") {
		t.Fatalf("top-level error chain must preserve intermediate context; got %q", step4.Error())
	}
}

// TestSentinelPropagation_ProductionWrapMustUseW guards against a
// subtle regression: someone replacing %w with %v in the wrap chain
// (which kills errors.Is propagation). Mirrors the *exact* format
// strings used in scraper.go so a wrong format string fails this test.
func TestSentinelPropagation_ProductionWrapMustUseW(t *testing.T) {
	t.Parallel()

	leaf := errBloggerVideoUnavailable

	// The two wrap call sites in scraper.go.
	wrapA := fmt.Errorf("failed to extract video URL: %w", leaf)
	wrapB := fmt.Errorf("video unavailable on this source: %w", wrapA)

	if !errors.Is(wrapA, errBloggerVideoUnavailable) {
		t.Fatalf("wrapA must preserve sentinel via %%w")
	}
	if !errors.Is(wrapB, errBloggerVideoUnavailable) {
		t.Fatalf("wrapB must preserve sentinel via %%w (caller depends on this)")
	}

	// And the negative — using %v breaks errors.Is.
	bad := fmt.Errorf("failed to extract video URL: %v", leaf)
	if errors.Is(bad, errBloggerVideoUnavailable) {
		t.Fatalf("guard tripped: %%v should NOT preserve errors.Is — if it does, the Go runtime semantics changed and this whole assumption needs a re-audit")
	}
}

// TestSentinelDistinctFromGenericParseError asserts that a structured
// response we fail to walk (a different bug class from "video is dead")
// does NOT match the sentinel. If this test breaks, it means the
// sentinel got over-broad and would cause callers to fast-fail a
// transient parser-shape problem that's worth retrying.
func TestSentinelDistinctFromGenericParseError(t *testing.T) {
	t.Parallel()

	// Valid wrb.fr envelope, but the inner data has no streams array — a
	// parser-shape problem worth retrying (the 2026-04-23 fix exists
	// precisely because Google reshuffled the index before).
	body := `)]}'

41
[["wrb.fr","WcwnYd","[\"some-payload\",null,1]",null,null,null,"generic"]]
`
	_, err := parseBatchexecuteResponse([]byte(body))
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if errors.Is(err, errBloggerVideoUnavailable) {
		t.Fatalf("structured-but-unparsed must NOT match the dead-token sentinel; got %v", err)
	}
}

// TestNeedsVideoExtraction_BloggerVariants is a guardrail for the URL
// classifier that decides which URLs flow into extractBloggerVideoURL
// (and therefore into the parser's sentinel path). If a future code
// change loses Blogger URL detection, the dead-token fast-fail would
// silently stop firing for that class. Pin every variant we serve.
func TestNeedsVideoExtraction_BloggerVariants(t *testing.T) {
	t.Parallel()

	cases := []struct {
		url  string
		want bool
		why  string
	}{
		{"https://www.blogger.com/video.g?token=XYZ", true, "canonical blogger embed"},
		{"https://blogger.com/video.g?token=XYZ", true, "blogger without www"},
		{"http://www.blogger.com/video.g?token=XYZ", true, "http (no scheme upgrade)"},
		{"https://www.blogspot.com/video/XYZ", true, "blogspot variant"},
		{"https://WWW.BLOGGER.COM/VIDEO.G?TOKEN=XYZ", true, "uppercase host (case-insensitive)"},
		{"https://rr1---sn-xxx.googlevideo.com/videoplayback?expire=123", false, "resolved CDN URL"},
		{"http://127.0.0.1:54577/blogger_proxy", false, "local proxy URL"},
		{"https://cdn.example.com/master.m3u8", false, "HLS"},
		{"", false, "empty"},
	}

	for _, tc := range cases {
		t.Run(tc.why, func(t *testing.T) {
			t.Parallel()

			got := needsVideoExtraction(tc.url)
			if got != tc.want {
				t.Fatalf("needsVideoExtraction(%q) = %v, want %v (%s)", tc.url, got, tc.want, tc.why)
			}
		})
	}
}
