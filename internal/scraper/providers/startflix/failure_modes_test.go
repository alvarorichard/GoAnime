package startflix

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// byseKey takes parts v and 31-v for versions 1..20. Bounding only 31-v let
// v in 16..20 index past a list shorter than v: a runtime panic.
func TestByseKeyShortPartListDoesNotPanic(t *testing.T) {
	t.Parallel()
	enc := base64.RawURLEncoding.EncodeToString
	var panicked []string
	for v := 16; v <= 20; v++ {
		for n := 31 - v; n < v; n++ {
			parts := make([]string, n)
			for i := range parts {
				parts[i] = enc(make([]byte, 16))
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = append(panicked, fmt.Sprintf("v=%d parts=%d: %v", v, n, r))
					}
				}()
				if _, err := byseKey(&bysePlayback{Version: strconv.Itoa(v), KeyParts: parts}); err == nil {
					t.Errorf("v=%d parts=%d: a %d-byte key was accepted", v, n, 16*n)
				}
			}()
		}
	}
	if len(panicked) > 0 {
		t.Fatalf("byseKey panicked for %d inputs:\n%s", len(panicked), strings.Join(panicked, "\n"))
	}
}

// The same record served by a Byse host. ResolveStream resolves players in
// goroutines, where a panic cannot be recovered by any caller: it ends the
// process, so one malformed API answer must come back as an error instead.
func TestResolveStreamSurvivesShortByseKeyParts(t *testing.T) {
	t.Parallel()
	pb := sealByse(t, map[string]any{
		"sources": []map[string]any{{"url": "https://cdn.test/master.m3u8", "height": 1080}},
	}, 20)
	pb.KeyParts = pb.KeyParts[:11] // passes a 31-v bound, too short for part v
	c := newTestClient(t, fakeHosts(t, &pb, ""))

	_, err := c.ResolveStream(context.Background(), []Player{{URL: "https://embedplaybyse.test/e/abc12345/x", Type: "iframe"}})
	if !errors.Is(err, ErrNoSupportedServer) {
		t.Fatalf("err = %v, want a NoStreamError recording the Byse failure", err)
	}
}

// A player whose URL does not parse or names no host was skipped without a
// trace, so a listing of only such players came back as an empty
// NoStreamError: no unsupported host, no failure. Nothing usable was listed,
// which is what ErrNoPlayers says.
func TestResolveStreamOnlyUnusablePlayerURLs(t *testing.T) {
	t.Parallel()
	c := NewClientForTest(&http.Client{Transport: failTransport{t}}, "https://www.startflix.test")
	_, err := c.ResolveStream(context.Background(), []Player{
		{URL: "/player/relative", Type: "iframe"},
		{URL: "https://%zz/e/x", Type: "iframe"},
	})
	if !errors.Is(err, ErrNoPlayers) {
		t.Fatalf("err = %#v (%v), want ErrNoPlayers", err, err)
	}
}

// An unusable URL beside an unsupported host must not hide that host.
func TestResolveStreamUnusableURLBesideUnsupportedHost(t *testing.T) {
	t.Parallel()
	c := NewClientForTest(&http.Client{Transport: failTransport{t}}, "https://www.startflix.test")
	_, err := c.ResolveStream(context.Background(), []Player{
		{URL: "/player/relative", Type: "iframe"},
		{URL: "https://vidsrcme.su/embed/tv?tmdb=1", Type: "iframe"},
	})
	var ns *NoStreamError
	if !errors.As(err, &ns) || len(ns.Unsupported) != 1 || ns.Unsupported[0] != "vidsrcme.su" {
		t.Fatalf("err = %v, want a NoStreamError naming vidsrcme.su", err)
	}
}

// failTransport fails the test on any request: the players above must be
// rejected before anything goes out.
type failTransport struct{ t *testing.T }

func (f failTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	f.t.Errorf("unexpected request to %s", r.URL)
	return nil, errors.New("no network in this test")
}

// probeAbyss reads 16 bytes at the first token part. A file that ends fewer
// than 16 bytes past the encrypted head has a shorter last part, and reading
// past it was refused as an invalid range — rejecting a playable file.
func TestProbeAbyssShortFirstTokenPart(t *testing.T) {
	t.Parallel()
	const chunk = int64(2 << 20)
	const size = chunk + 10
	plain := make([]byte, size)
	copy(plain, "\x00\x00\x00\x20ftypisom\x00\x00\x02\x00")
	up := &fakeChunkUpstream{
		plain: plain, fdSeed: "head.fd", fdSize: chunk,
		size: size, md5ID: 99, resID: 2, chunk: chunk, origin: "https://playembedapi.site",
	}
	srv := httptest.NewTLSServer(up)
	t.Cleanup(srv.Close)
	f := newChunkedFile(srv, size, chunk)
	c := &Client{http: srv.Client(), baseURL: "https://www.startflix.test", userAgent: userAgent}

	if err := c.probeAbyss(context.Background(), f); err != nil {
		t.Fatalf("probe rejected a playable file: %v", err)
	}
	if got := up.hitsFor("sora:1"); got != 1 {
		t.Errorf("token part 1 fetched %d times, want 1: the probe must still check it", got)
	}
}
