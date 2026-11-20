package pomfy

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"
)

// End-to-end of the byse chain over the shared mock: captcha handed out,
// PoW solved with the challenge's own difficulty, verify judged the way the
// server judges (hash of OUR nonce), playback gated on the captcha token,
// final record decrypted through resolveByse itself.

func TestResolveByseEndToEnd(t *testing.T) {
	cdn := newFakeByseCDN(t, "ref", 6, sealFixture(t, recordJSON(t)))
	c := NewClientForTest(cdn.srv.Client(), cdn.srv.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := resolveByse(ctx, c, cdn.srv.URL, "cta12345", DefaultBase+"/assistir/550")
	if err != nil {
		t.Fatalf("resolveByse: %v", err)
	}
	if stream.URL != "https://edge.test/hls2/high/master.m3u8" {
		t.Errorf("stream.URL = %q, want the tallest source", stream.URL)
	}
	if stream.Height != 1080 {
		t.Errorf("stream.Height = %d, want 1080", stream.Height)
	}
	if len(stream.Subtitles) != 1 || stream.Subtitles[0].Language != "pt-br" {
		t.Errorf("subtitles = %+v", stream.Subtitles)
	}
	if cdn.detailsPOSTs != 1 || cdn.captchaPOSTs != 1 || cdn.verifyPOSTs != 1 || cdn.playbackPOSTs != 1 {
		t.Errorf("flow counts: details=%d captcha=%d verify=%d playback=%d, want 1/1/1/1",
			cdn.detailsPOSTs, cdn.captchaPOSTs, cdn.verifyPOSTs, cdn.playbackPOSTs)
	}
}

// TestResolveByseEmbedIdentity pin that the embedding headers name the
// watching site and the player frame — what the server validates.
func TestResolveByseEmbedIdentity(t *testing.T) {
	cdn := newFakeByseCDN(t, "ref", 6, sealFixture(t, recordJSON(t)))
	var gotOrigin, gotParent, gotReferer string
	inner := cdn.srv.Client().Transport
	rt := &recordingTransport{inner: inner}
	hc := cdn.srv.Client()
	hc.Transport = rt
	c := NewClientForTest(hc, cdn.srv.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	watch := DefaultBase + "/assistir/550"
	if _, err := resolveByse(ctx, c, cdn.srv.URL, "cta12345", watch); err != nil {
		t.Fatalf("resolveByse: %v", err)
	}
	for _, h := range rt.seen {
		if got := h.Get("X-Embed-Origin"); got != "" {
			gotOrigin = got
		}
		if got := h.Get("X-Embed-Parent"); got != "" && strings.Contains(got, "/e/") {
			gotParent = got
		}
		if got := h.Get("X-Embed-Referer"); got != "" {
			gotReferer = got
		}
		if tok := h.Get("X-Captcha-Token"); tok != "" && !strings.HasPrefix(tok, "captcha-ok-") {
			t.Errorf("X-Captcha-Token = %q, want the fake ok-token prefix", tok)
		}
	}
	if gotOrigin != DefaultBaseHost {
		t.Errorf("X-Embed-Origin = %q, want %q", gotOrigin, DefaultBaseHost)
	}
	if !strings.HasPrefix(gotParent, "//"+""+cdn.srv.URL[7:]) {
		if !strings.HasPrefix(gotParent, cdn.srv.URL+"/e/") {
			t.Errorf("X-Embed-Parent = %q, want %q prefix", gotParent, cdn.srv.URL+"/e/")
		}
	}
	if gotReferer != watch {
		t.Errorf("X-Embed-Referer = %q, want %q", gotReferer, watch)
	}
}

// TestByseKeyVersionPicksTwoParts checks the reorderKeyParts rule: version v
// in 1..20 with thirty parts means parts v and 31-v (1-based); anything else
// means the whole list joined, which cannot be a 32-byte key.
func TestByseKeyVersionPicksTwoParts(t *testing.T) {
	mkParts := func(n int) []string {
		parts := make([]string, n)
		for i := range parts {
			parts[i] = base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%016x", i)))
		}
		return parts
	}

	fixture := sealFixture(t, []byte(`{"sources":[],"tracks":[]}`))
	key, err := byseKey(&fixture)
	if err != nil {
		t.Fatalf("byseKey version 9: %v", err)
	}
	if len(key) != 32 {
		t.Fatalf("byseKey version 9 length = %d, want 32", len(key))
	}

	parts23 := playbackSeal{KeyParts: mkParts(30), Version: "23"}
	if _, err := byseKey(&parts23); err == nil {
		t.Error("version 23 with 30 decoys must fail, not fake a key")
	}

	// The 1-based rule on a hand-built seal: version "5" selects exactly
	// parts 5 and 26, whatever those decode to.
	parts := mkParts(30)
	want := []byte{}
	for _, idx := range []int{4, 25} {
		raw, err := base64.RawURLEncoding.DecodeString(parts[idx])
		if err != nil {
			t.Fatal(err)
		}
		want = append(want, raw...)
	}
	five := playbackSeal{KeyParts: parts, Version: "5"}
	got, err := byseKey(&five)
	if err != nil {
		t.Fatalf("byseKey version 5: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Error("version 5 must select parts 5 and 26 (1-based)")
	}
}

// TestDecryptBysePlaybackFailures pins the guard rails.
func TestDecryptBysePlaybackFailures(t *testing.T) {
	good := sealFixture(t, []byte(`{"sources":[],"tracks":[]}`))

	wrong := good
	wrong.Algorithm = "AES-256-CBC"
	if _, err := decryptBysePlayback(&wrong); err == nil {
		t.Error("CBC must be refused")
	}

	if _, err := decryptBysePlayback(nil); err == nil {
		t.Error("nil seal must be refused")
	}

	short := good
	short.Payload = "AAAA"
	if _, err := decryptBysePlayback(&short); err == nil {
		t.Error("a payload too short to carry a tag must be refused")
	}

	trash := good
	trash.IV = "!!not-base64!!"
	if _, err := decryptBysePlayback(&trash); err == nil {
		t.Error("an unreadable iv must be refused")
	}
}
