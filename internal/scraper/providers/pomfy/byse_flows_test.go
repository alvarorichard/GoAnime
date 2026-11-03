package pomfy

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

// The resolveByse paths: every recording shape the fake CDN can answer —
// including the failure shapes — driven through resolveByse itself. The
// resolved cases keep the chain reachable through Unwrap so callers still
// tell a down CDN from a corrupt record.

type resolveCase struct {
	name string
	// knobs are the fake's stage substitutions; zero means recorded.
	knobs byseKnobs
	// seal replaces the whole record with a real one a test built.
	seal *playbackSeal
	// noCode runs the resolver without a code at all and asserts its early
	// return; everything else runs over cta12345.
	noCode bool
	// want is what the failure chain must mention; the shape the resolver
	// must SURVIVE (details garbage) carries the "@none@" marker.
	want string
}

func TestResolveByseFailurePaths(t *testing.T) {
	cases := []resolveCase{
		{name: "no code at all", noCode: true, want: "no video code to resolve"},
		{name: "details names the video gone",
			knobs: byseKnobs{detailsBody: `{"error":"gone"}`},
			want:  "byse cta12345: gone"},
		{name: "details garbage is survivable",
			knobs: byseKnobs{detailsBody: `<html>not the details json</html>`},
			want:  "@none@"},
		{name: "captcha answers 500",
			knobs: byseKnobs{captchaStatus: http.StatusInternalServerError, captchaBody: `{"error":"dead"}`},
			want:  "captcha challenge request failed"},
		{name: "captcha body is garbage",
			knobs: byseKnobs{captchaBody: `<html>not the challenge json</html>`},
			want:  "captcha challenge unreadable"},
		{name: "captcha carries an error",
			knobs: byseKnobs{captchaBody: `{"error":"overloaded"}`},
			want:  "byse cta12345 captcha: overloaded"},
		{name: "challenge incomplete",
			knobs: byseKnobs{captchaBody: `{"pow_nonce":"","pow_token":"","pow_difficulty":0}`},
			want:  "captcha challenge incomplete"},
		{name: "verify answers 500",
			knobs: byseKnobs{verifyStatus: http.StatusInternalServerError, verifyBody: `{"error":"dead"}`},
			want:  "captcha verify failed"},
		{name: "verify body is garbage",
			knobs: byseKnobs{verifyBody: `[1,2,3]`},
			want:  "captcha verify answer unreadable"},
		{name: "verify carries an error",
			knobs: byseKnobs{verifyBody: `{"error":"expired"}`},
			want:  "byse cta12345 captcha verify: expired"},
		{name: "verify says pending with no token",
			knobs: byseKnobs{verifyBody: `{"status":"pending"}`},
			want:  "captcha rejected the solution"},
		{name: "playback answers 403",
			knobs: byseKnobs{playbackStatus: http.StatusForbidden, playbackBody: `{"error":"gate"}`},
			want:  "playback request failed"},
		{name: "playback body is garbage",
			knobs: byseKnobs{playbackBody: `{"playback":`},
			want:  "playback record unreadable"},
		{name: "seal names a foreign cipher",
			knobs: byseKnobs{playbackBody: `{"playback":{"algorithm":"AES-256-CBC","iv":"AAAA","payload":"AAAA","key_parts":["AAAA"],"version":"1"}}`},
			want:  "unsupported algorithm"},
		{name: "seal lists nothing playable",
			seal: sealFor(t, []byte(`{"sources":[],"tracks":[]}`)),
			want: "no playable source"},
		{name: "seal does not decrypt with its own key",
			seal: brokenSeal(t),
			want: "cipher: message authentication failed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.noCode {
				// reads before any side is touched
				_, err := resolveByse(context.Background(), NewClientForTest(&http.Client{}, "https://x.test"),
					"https://x.test", "", DefaultBase+"/assistir/550")
				if err == nil || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("err = %v, want %q", err, tc.want)
				}
				return
			}
			seal := tc.seal
			if seal == nil {
				fixed := sealFixture(t, []byte(`{"sources":[{"url":"https://edge.test/x/master.m3u8"}],"tracks":[]}`))
				seal = &fixed
			}
			cdn := newFakeByseCDNWith(t, "ref", 6, *seal, tc.knobs)
			c := NewClientForTest(cdn.srv.Client(), cdn.srv.URL)

			_, err := resolveByse(context.Background(), c, cdn.srv.URL, "cta12345", DefaultBase+"/assistir/550")
			if tc.want == "@none@" {
				if err != nil {
					t.Fatalf("the survivable shape must resolve, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected an error")
			}
			if !errorChainMentions(err, tc.want) {
				t.Errorf("err = %v, want a chain mention of %q", err, tc.want)
			}
		})
	}
}

// errorChainMentions walks err's chain so a wrapped decryption failure still
// names its cause ("unsupported algorithm …").
func errorChainMentions(err error, want string) bool {
	for chain := err; chain != nil; chain = errors.Unwrap(chain) {
		if strings.Contains(chain.Error(), want) {
			return true
		}
	}
	return false
}

// sealFor returns the fixture seal as a pointer for the knob table.
func sealFor(t *testing.T, plaintext []byte) *playbackSeal {
	t.Helper()
	seal := sealFixture(t, plaintext)
	return &seal
}

// brokenSeal realises a seal whose iv no longer matches its key, so the
// authenticated open fails — the "corrupt record" shape.
func brokenSeal(t *testing.T) *playbackSeal {
	seal := sealFixture(t, []byte(`{"sources":[{"url":"https://edge.test/x/master.m3u8"}]}`))
	seal.IV = "AAAAAAAAAp"
	return &seal
}

// ResolveByse / ResolveStream: the PUBLIC helpers over the world whose
// catalog names the fake CDN — the whole tree as the client has it.
func TestResolveHelpersOverWorld(t *testing.T) {
	world := newFakeWorld(t)
	c := world.client()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := c.ResolveByse(ctx, watchSource{EmbedURL: world.cdn.srv.URL + "/e/cta12345", Kind: "byse"}, DefaultBase+"/assistir/550")
	if err != nil {
		t.Fatalf("ResolveByse: %v", err)
	}
	if !strings.Contains(stream.URL, "master.m3u8") || stream.Height != 1080 {
		t.Errorf("stream = %+v", stream)
	}
	if stream.Host == "" || !strings.HasPrefix(stream.Host, "http") {
		t.Errorf("Host = %q, want the embed base", stream.Host)
	}
	if stream.Referer != DefaultBase+"/assistir/550" {
		t.Errorf("Referer = %q", stream.Referer)
	}

	stream, err = c.ResolveStream(ctx, 550, KindMovie, 0, 0, DefaultBase+"/assistir/550")
	if err != nil {
		t.Fatalf("ResolveStream: %v", err)
	}
	if !strings.Contains(stream.URL, "master.m3u8") {
		t.Errorf("stream = %q", stream.URL)
	}

	// The dead every-source case surfaces ErrNoSources.
	world.cdn.refuseSo = true
	if _, err := c.ResolveStream(ctx, 550, KindMovie, 0, 0, DefaultBase+"/assistir/550"); !errors.Is(err, ErrNoSources) {
		t.Fatalf("all-sources-dead err = %v, want ErrNoSources", err)
	}
}

// The URL→host passthrough: cdn.test has no DNS, and ResolveByse names it —
// a one-second client keeps the suite tight while failing loudly.
func TestResolveByseKeepsEmbedHost(t *testing.T) {
	c := NewClientForTest(&http.Client{Timeout: time.Second}, DefaultBase)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := c.ResolveByse(ctx, watchSource{EmbedURL: "https://cdn.test/e/gkcq8gbiwh4r"}, DefaultBase+"/assistir/550")
	if err == nil {
		t.Fatal("expected the unroutable host to fail the resolve")
	}
	if !strings.Contains(err.Error(), "cdn.test") && !dialish(err) {
		t.Errorf("err = %v, want a DNS-level failure for cdn.test", err)
	}
}

// dialish walks the chain looking for the URI dial/lookup errors.
func dialish(err error) bool {
	for chain := err; chain != nil; chain = errors.Unwrap(chain) {
		msg := chain.Error()
		if strings.Contains(msg, "lookup") || strings.Contains(msg, "dial") ||
			strings.Contains(msg, "no such host") {
			return true
		}
	}
	return false
}

// --- the small helpers' edges ---

func TestWatchPageOriginFallback(t *testing.T) {
	if got := watchPageOrigin(""); got != DefaultBaseHost {
		t.Errorf("empty → %q, want %q", got, DefaultBaseHost)
	}
	if got := watchPageOrigin("http://"); got != DefaultBaseHost {
		t.Errorf("no host → %q, want %q", got, DefaultBaseHost)
	}
	if got := watchPageOrigin("https://other.example/assistir/1"); got != "other.example" {
		t.Errorf("other host → %q", got)
	}
}

func TestDecodeBase64URLEdges(t *testing.T) {
	if got, err := decodeBase64URL(""); err == nil || got != nil {
		t.Errorf("empty = %q err=%v, want an error", got, err)
	}
	raw, err := decodeBase64URL("aGVsbG8=") // padded spelling tolerated
	if err != nil || string(raw) != "hello" {
		t.Errorf("padded = %q err=%v, want hello", raw, err)
	}
}

func TestPickSourceEdges(t *testing.T) {
	if _, ok := pickSource(nil); ok {
		t.Error("nil list must not play")
	}
	if _, ok := pickSource([]sealedSource{{URL: "rtsp://old.test"}}); ok {
		t.Error("non-http sources must not play")
	}
	best, ok := pickSource([]sealedSource{
		{URL: "https://a/x.m3u8", Label: "480p"},
		{URL: "https://b/x.m3u8", Label: "720p"},
	})
	if !ok || best.URL != "https://b/x.m3u8" {
		t.Errorf("picked %+v, want the labeled-tallest", best)
	}
	// An unlabeled single master with the height field empty still plays.
	if single, ok := pickSource([]sealedSource{{URL: "https://a/x.m3u8"}}); !ok || single.URL != "https://a/x.m3u8" {
		t.Errorf("unlabeled master = %+v ok=%v, want it playable", single, ok)
	}
}

// byseKey: a bad part decode must end the lookup with a named error.
func TestByseKeyBadPartDecode(t *testing.T) {
	pb := &playbackSeal{KeyParts: []string{"!!no base64!!"}, Version: ""} // "" joins every part
	if _, err := byseKey(pb); err == nil || !strings.Contains(err.Error(), "key part") {
		t.Fatalf("err = %v, want a key part decode error", err)
	}
}

func TestSubtitlesFilterEdges(t *testing.T) {
	filtered := subtitlesFilter([]byseTrack{
		{URL: "https://a/pt.vtt", Language: "pt-br", Label: "?"},
		{File: "https://b/en.vtt", Kind: "captions"},
		{Src: "https://c/dup.vtt", Kind: "subtitles"},
		{URL: "https://c/dup.vtt", Kind: "subtitles"}, // dup drops
		{URL: "https://d/audio.m3u8", Kind: "audio"},  // non-caption kind drops
		{Label: ""}, // urlless drops
	})
	if len(filtered) != 3 {
		t.Fatalf("filtered = %+v, want three tracks", filtered)
	}
	if filtered[1].URL != "https://b/en.vtt" {
		t.Errorf("file mapping wrong: %+v", filtered[1])
	}
	if filtered[2].Language != "" {
		t.Errorf("empty fields must land empty: %+v", filtered[2])
	}
}

func TestFirstNonEmptyRuns(t *testing.T) {
	if got := firstNonEmpty("", " ", "x"); got != "x" {
		t.Errorf("skips blanks: %q", got)
	}
	if got := firstNonEmpty(); got != "" {
		t.Errorf("empty call = %q", got)
	}
}

func TestHeightBridgesLabels(t *testing.T) {
	if got := (sealedSource{Label: "1080p"}).height(); got != 1080 {
		t.Errorf("label height = %d, want 1080", got)
	}
	if got := (sealedSource{Label: "unlabeled"}).height(); got != 0 {
		t.Errorf("unlabelable = %d, want 0", got)
	}
	if got := (sealedSource{Label: "2160"}).height(); got != 2160 {
		t.Errorf("bare-number label = %d", got)
	}
}

func TestLabelHeightTable(t *testing.T) {
	cases := []struct {
		label string
		want  int
	}{
		{"4320p", 4320},
		{"720", 720},
		{"p", 0},        // no digits
		{"24p", 0},      // below the 144 floor reads as unlabeled
		{"Stream X", 0}, // nothing to read
	}
	for _, tc := range cases {
		if got := labelHeight(tc.label); got != tc.want {
			t.Errorf("labelHeight(%q) = %d, want %d", tc.label, got, tc.want)
		}
	}
}

func TestHeightDigitsGuard(t *testing.T) {
	if _, ok := heightDigits("no digit"); ok {
		t.Error("a label with no digit must not parse")
	}
	if _, ok := heightDigits(""); ok {
		t.Error("empty must not parse")
	}
}

func TestURLParseEmbedEdges(t *testing.T) {
	if ByseCode("%zz nonsense") != "" {
		t.Error("a malformed embed must not yield a code")
	}
	if got := ByseCode("https://cdn.test/e/ABCDEF"); got != "" {
		t.Errorf("uppercase must not pass: %q", got)
	}
}
