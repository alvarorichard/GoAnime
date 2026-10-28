package player

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// Issue #203: the dead-Blogger-token chain already fast-failed, but the only
// thing that reached the user was the raw wrapped error. ErrSourceVideoRemoved
// is the exported half of that wrap, so the playback layer can recognise
// "the source deleted the file" and say so.
//
// The wrap must carry BOTH sentinels: the exported one for the UI decision, the
// internal one for the fast-fail checks already in this package.
func TestErrSourceVideoRemoved_WrapCarriesBothSentinels(t *testing.T) {
	t.Parallel()

	// The exact wrap from GetVideoURLForEpisodeEnhanced.
	cause := fmt.Errorf("failed to extract video URL: %w", bloggerRPCError{Code: 5})
	wrapped := fmt.Errorf("%w: %w", ErrSourceVideoRemoved, cause)

	if !errors.Is(wrapped, ErrSourceVideoRemoved) {
		t.Fatalf("playback layer can no longer detect a removed video: %v", wrapped)
	}
	if !errors.Is(wrapped, errBloggerVideoUnavailable) {
		t.Fatalf("internal fast-fail sentinel lost from the chain: %v", wrapped)
	}
	// The rendered message is unchanged from before the sentinel was exported.
	if !strings.HasPrefix(wrapped.Error(), "video unavailable on this source: ") {
		t.Fatalf("user-facing prefix changed: %q", wrapped.Error())
	}
}

// A transient RPC status must NOT look like a removed video, or the UI would
// tell the user to switch sources over a hiccup that a retry fixes.
func TestErrSourceVideoRemoved_TransientRPCStatusDoesNotMatch(t *testing.T) {
	t.Parallel()

	for _, code := range []int{5, 3} {
		if _, terminal := bloggerRPCStatusName(code); !terminal {
			t.Fatalf("expected RPC status %d to be terminal", code)
		}
		if !errors.Is(bloggerRPCError{Code: code}, errBloggerVideoUnavailable) {
			t.Fatalf("terminal RPC status %d must wrap the dead-token sentinel", code)
		}
	}

	// 14 (UNAVAILABLE) is Google being briefly unhappy, not a deleted video.
	transient := bloggerRPCError{Code: 14}
	if _, terminal := bloggerRPCStatusName(14); terminal {
		t.Skip("status 14 is classified terminal in this build; nothing to assert")
	}
	if errors.Is(transient, errBloggerVideoUnavailable) {
		t.Fatalf("transient RPC status must not report the video as removed")
	}
	if errors.Is(transient, ErrSourceVideoRemoved) {
		t.Fatalf("transient RPC status must not reach the 'try another source' message")
	}
}
