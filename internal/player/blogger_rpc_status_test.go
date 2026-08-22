package player

import (
	"errors"
	"strings"
	"testing"
)

// realNotFoundBody is a verbatim capture from Google (issue #195, episode 5 of
// "JoJo no Kimyou na Bouken Part 3" on AnimeFire): HTTP 200, no payload, and
// status code 5 in the envelope.
const realNotFoundBody = ")]}'\n\n106\n" +
	`[["wrb.fr","WcwnYd",null,null,null,[5],"generic"],["di",160],["af.httprm",160,"-6632312378740051841",2]]` +
	"\n25\n" + `[["e",4,null,null,142]]` + "\n"

func TestBatchexecuteRPCStatus_ReadsRealNotFoundEnvelope(t *testing.T) {
	code, ok := batchexecuteRPCStatus([]byte(realNotFoundBody))
	if !ok {
		t.Fatal("failed to read the status code out of a real NOT_FOUND envelope")
	}
	if code != 5 {
		t.Fatalf("code = %d, want 5", code)
	}
}

func TestBatchexecuteRPCStatus_IgnoresEnvelopeWithPayload(t *testing.T) {
	body := ")]}'\n\n1461\n" +
		`[["wrb.fr","WcwnYd","[1,null,[[\"https://rr7.googlevideo.com/videoplayback?mime=video/mp4\",[18]]]]",null,null,null,"generic"]]`
	if _, ok := batchexecuteRPCStatus([]byte(body)); ok {
		t.Fatal("an envelope carrying a payload must not be read as an error status")
	}
}

func TestBatchexecuteRPCStatus_IgnoresUnrelatedBodies(t *testing.T) {
	for _, body := range []string{"", ")]}'", ")]}'\n\n[]", "not json at all", `[["wrb.fr","OtherRpc",null,null,null,[5]]]`} {
		if _, ok := batchexecuteRPCStatus([]byte(body)); ok {
			t.Errorf("body %q must not yield a status code", body)
		}
	}
}

// TestParseBatchexecute_NotFoundIsReportedAsUnavailable is the issue #195 fix:
// the opaque "no video URL found in batchexecute response" is replaced by a
// message that names the real cause, and it fast-fails via the existing
// sentinel instead of burning the retry budget.
func TestParseBatchexecute_NotFoundIsReportedAsUnavailable(t *testing.T) {
	_, err := parseBatchexecuteResponse([]byte(realNotFoundBody))
	if err == nil {
		t.Fatal("expected an error")
	}
	if !errors.Is(err, errBloggerVideoUnavailable) {
		t.Fatalf("NOT_FOUND must map to errBloggerVideoUnavailable so callers fast-fail; got %v", err)
	}
	msg := err.Error()
	for _, want := range []string{"NOT_FOUND", "removed"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error message should mention %q, got %q", want, msg)
		}
	}
	if strings.Contains(msg, "no video URL found") {
		t.Errorf("the opaque parser-bug wording must be gone, got %q", msg)
	}
}

// TestParseBatchexecute_TransientStatusStaysRetryable keeps rate limiting from
// being misfiled as a dead video: those codes must NOT fast-fail.
func TestParseBatchexecute_TransientStatusStaysRetryable(t *testing.T) {
	for _, code := range []string{"8", "14", "4", "13"} {
		body := ")]}'\n\n106\n" + `[["wrb.fr","WcwnYd",null,null,null,[` + code + `],"generic"]]`
		_, err := parseBatchexecuteResponse([]byte(body))
		if err == nil {
			t.Fatalf("code %s: expected an error", code)
		}
		if errors.Is(err, errBloggerVideoUnavailable) {
			t.Errorf("code %s is transient and must stay retryable, but it fast-fails: %v", code, err)
		}
	}
}

func TestParseBatchexecute_TerminalStatusesFastFail(t *testing.T) {
	for _, code := range []string{"3", "5", "7", "9", "16"} {
		body := ")]}'\n\n106\n" + `[["wrb.fr","WcwnYd",null,null,null,[` + code + `],"generic"]]`
		_, err := parseBatchexecuteResponse([]byte(body))
		if !errors.Is(err, errBloggerVideoUnavailable) {
			t.Errorf("code %s is terminal and must fast-fail, got %v", code, err)
		}
	}
}

// TestParseBatchexecute_UnknownStatusStillReportsTheCode makes sure a status
// Google adds later is still surfaced with its number rather than swallowed.
func TestParseBatchexecute_UnknownStatusStillReportsTheCode(t *testing.T) {
	body := ")]}'\n\n106\n" + `[["wrb.fr","WcwnYd",null,null,null,[42],"generic"]]`
	_, err := parseBatchexecuteResponse([]byte(body))
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "42") {
		t.Errorf("error should carry the unrecognised code, got %q", err)
	}
}

// TestParseBatchexecute_StillParsesAWorkingResponse guards the happy path.
func TestParseBatchexecute_StillParsesAWorkingResponse(t *testing.T) {
	body := ")]}'\n\n1461\n" +
		`[["wrb.fr","WcwnYd","[1,null,[[\"https://rr7.googlevideo.com/videoplayback?itag=18\\u0026mime=video/mp4\",[18]]]]",null,null,null,"generic"]]`
	got, err := parseBatchexecuteResponse([]byte(body))
	if err != nil {
		t.Fatalf("a valid payload must still parse: %v", err)
	}
	if !strings.Contains(got, "googlevideo.com") {
		t.Fatalf("unexpected URL %q", got)
	}
}
