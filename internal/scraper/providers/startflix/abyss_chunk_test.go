package startflix

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"maps"
	"math"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func mustUnmarshalMedia(t *testing.T, name string) *abyssMedia {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	var media abyssMedia
	if err := jsonx.Unmarshal(raw, &media); err != nil {
		t.Fatal(err)
	}
	return &media
}

// Quirky vectors: the player's md5 hashes the decimal digits' VALUES for a
// numeric seed (see abyssChunkKey), verified against the live player.
func TestAbyssChunkKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		size int64
		want string
	}{
		{392440465, "880d78d848175fee6b096c173909eca7"},
		{555324642, "93786ae9d29f48bff5ae832d094365a5"},
		{1098017024, "f085981736f919b7fe766d6c5e513176"},
		{3112249341, "800b914bc931aad099a1060c52c099b1"},
		{844573812, "a37606ea7a3cffbb529736d3073b26cf"},
		{10, "25daad3d9e60b45043a70c4ab7d3b1c6"},
	}
	for _, tt := range tests {
		key, iv := abyssChunkKey(tt.size)
		if got := string(key); got != tt.want {
			t.Errorf("abyssChunkKey(%d) = %s, want %s", tt.size, got, tt.want)
		}
		if string(iv) != tt.want[:aes.BlockSize] {
			t.Errorf("abyssChunkKey(%d) iv = %x, want key prefix", tt.size, iv)
		}
	}
}

// TestAbyssChunkTokenPlayerParity pins the token builder against a token the
// live player minted on 2026-10-08 (captured from its /sora/ request): same
// path, same size, byte-for-byte identical output. It also pins that the key
// is NOT the naive standard-md5 derivation, which the CDN rejects with 404.
func TestAbyssChunkTokenPlayerParity(t *testing.T) {
	t.Parallel()
	const (
		size = 392440465
		path = "/mp4/26611060/2/392440465/2097152/3"
		want = "dTdWOW5Mb2xYa3BtOXNYdXNZTDFQS2JzNWRzQXgyK0VLV0ZScjgxRzZzU3JpOW8"
	)
	if got := abyssChunkToken(path, size); got != want {
		t.Fatalf("token = %s, want player token %s", got, want)
	}
	naive := md5.Sum([]byte(strconv.FormatInt(size, 10)))
	if key, _ := abyssChunkKey(size); string(key) == hex.EncodeToString(naive[:]) {
		t.Fatal("chunk key equals the naive standard-md5 derivation")
	}
}

func TestAbyssChunkDomain(t *testing.T) {
	t.Parallel()
	domains := []string{"aa.test", "bb.test", "cc.test"}
	if got := abyssChunkDomain(domains, 392440465, "bb"); got != "bb.test" {
		t.Errorf("sub match = %q", got)
	}
	// No sub match: the player falls back to domains[size%len(domains)].
	if got, want := abyssChunkDomain(domains, 4, "zz"), "bb.test"; got != want {
		t.Errorf("fallback = %q, want %q", got, want)
	}
	if got := abyssChunkDomain(nil, 4, "zz"); got != "" {
		t.Errorf("empty domains = %q", got)
	}
}

// TestMakeAbyssFileChunked uses a real decrypted media list (H5QfVbK-c,
// 2026-10-08): no url/path rendition exists, so the file must resolve to the
// chunked layout with the sub-matched domain and head entry.
func TestMakeAbyssFileChunked(t *testing.T) {
	t.Parallel()
	embed := mustParseURL(t, "https://playembedapi.site/?v=H5QfVbK-c")
	datas := &abyssDatas{Slug: "H5QfVbK-c", MD5ID: 26611060, UserID: 9133}
	media := mustUnmarshalMedia(t, "abyss_media_chunked_2026_10_08.json")
	f, ok := makeAbyssFile(embed, datas, media)
	if !ok {
		t.Fatal("chunked-only media resolved to nothing")
	}
	if f.chunk == nil {
		t.Fatal("expected the chunked layout, got single-file")
	}
	if f.size != 1217612566 {
		t.Errorf("size = %d, want tallest h264 1217612566", f.size)
	}
	if f.origin != "https://playembedapi.site" {
		t.Errorf("origin = %q", f.origin)
	}
	c := f.chunk
	if c.chunkHost != "o25chikcb28.sssrr.org" {
		t.Errorf("host = %q", c.chunkHost)
	}
	if want := "https://o25chikcb28.sssrr.org/mp4/26611060/5/1217612566"; c.chunkURL != want {
		t.Errorf("chunkURL = %q, want %q", c.chunkURL, want)
	}
	if want := "/mp4/26611060/5/1217612566"; c.chunkPath != want {
		t.Errorf("chunkPath = %q, want %q", c.chunkPath, want)
	}
	if c.firstSize != 18874368 || !strings.HasSuffix(c.firstURL, ".fd") {
		t.Errorf("head = %q (%d bytes)", c.firstURL, c.firstSize)
	}
	if want := "e495aeb7e575f93364ba6daa6f4d9.1217612566.5.fd"; c.firstSeed != want {
		t.Errorf("head seed = %q, want %q", c.firstSeed, want)
	}
}

// TestMakeAbyssFileChunkedNoHead covers renditions with sub/domains but no
// fristDatas entry — and a partSize that is not a whole number of parts,
// which the player refuses to split — so everything goes through the chunk
// tokens.
func TestMakeAbyssFileChunkedNoHead(t *testing.T) {
	t.Parallel()
	embed := mustParseURL(t, "https://playembedapi.site/?v=x")
	datas := &abyssDatas{Slug: "x", MD5ID: 42, UserID: 7}
	media := &abyssMedia{}
	media.MP4.Sources = []abyssSource{
		{Label: "720p", ResID: 4, Size: 1000000, Codec: "h264", Sub: "qq"},
		{Label: "480p", ResID: 3, Size: 900000, Codec: "h264", Sub: "rr"},
	}
	media.MP4.Domains = []string{"qq.cdn.test", "rr.cdn.test"}
	media.MP4.FristDatas = []abyssFirstData{
		// partSize 3 MiB is not a multiple of the 2 MiB part: refused.
		{ResID: 4, Size: 1000000, Codec: "h264", URL: "https://qq.cdn.test/h.1000000.4.fd", PartSize: 3 << 20},
	}
	f, ok := makeAbyssFile(embed, datas, media)
	if !ok || f.chunk == nil {
		t.Fatalf("file = %+v, ok = %v; want chunked without head", f, ok)
	}
	if f.chunk.firstSize != 0 || f.chunk.firstURL != "" {
		t.Errorf("head should be empty: %+v", f.chunk)
	}
}

// fakeChunkUpstream serves a synthetic plaintext file through both chunk
// transports: an .fd head URL (head-encrypted like the player) and a /sora/
// endpoint that decrypts the token with the quirky key and serves the part.
// Both require the embed Origin, like the live CDN. Failures can be queued
// per target ("fd" or "sora:<part>"); hits are counted per target.
type fakeChunkUpstream struct {
	plain  []byte
	fdSeed string
	fdSize int64
	size   int64
	md5ID  int64
	resID  int
	chunk  int64
	origin string

	mu       sync.Mutex
	hits     map[string]int
	failures map[string][]int
	delay    time.Duration   // applied to a target's first hit
	hold     map[string]bool // targets whose hits hang until the request is cancelled
}

// record counts a hit on target, consumes a queued failure status (0 means
// none) and returns it; a target's first successful hit sleeps once, to hold
// the fetch open for the coalescing test. A held target never answers: ok is
// false once the client gives up on it.
func (f *fakeChunkUpstream) record(ctx context.Context, target string) (status int, ok bool) {
	f.mu.Lock()
	if f.hits == nil {
		f.hits = map[string]int{}
	}
	f.hits[target]++
	if q := f.failures[target]; len(q) > 0 {
		status, f.failures[target] = q[0], q[1:]
	}
	first, d, held := f.hits[target] == 1, f.delay, f.hold[target]
	f.mu.Unlock()
	if held {
		<-ctx.Done()
		return 0, false
	}
	if status == 0 && first && d > 0 {
		select {
		case <-time.After(d):
		case <-ctx.Done():
		}
	}
	return status, true
}

func (f *fakeChunkUpstream) hitsFor(target string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.hits[target]
}

func (f *fakeChunkUpstream) queue(target string, statuses ...int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failures == nil {
		f.failures = map[string][]int{}
	}
	f.failures[target] = append(f.failures[target], statuses...)
}

func (f *fakeChunkUpstream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != f.origin {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if strings.HasSuffix(r.URL.Path, ".fd") {
		status, ok := f.record(r.Context(), "fd")
		if !ok {
			return // the client gave up on a held target
		}
		if status != 0 {
			w.WriteHeader(status)
			return
		}
		head := append([]byte(nil), f.plain[:abyssEncryptedHead]...)
		if err := abyssXOR(f.fdSeed, 0, head); err != nil {
			panic(err)
		}
		f.serveBytes(w, r, append(head, f.plain[abyssEncryptedHead:f.fdSize]...))
		return
	}
	// /sora/<size>/<token>
	rest := strings.TrimPrefix(r.URL.Path, "/sora/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) != 2 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	token, err := decodeChunkToken(parts[1])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	key, iv := abyssChunkKey(f.size)
	block, _ := aes.NewCipher(key)
	ctr := make([]byte, aes.BlockSize)
	copy(ctr, iv)
	stream := cipher.NewCTR(block, ctr)
	stream.XORKeyStream(token, token)
	path, want := string(token), fmt.Sprintf("/mp4/%d/%d/%d/%d/", f.md5ID, f.resID, f.size, f.chunk)
	if !strings.HasPrefix(path, want) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	part, err := strconv.Atoi(strings.TrimPrefix(path, want))
	if err != nil || part < 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	status, ok := f.record(r.Context(), "sora:"+strconv.Itoa(part))
	if !ok {
		return // the client gave up on a held target
	}
	if status != 0 {
		w.WriteHeader(status)
		return
	}
	start := int64(part) * f.chunk
	f.serveBytes(w, r, f.plain[start:min(start+f.chunk, f.size)])
}

func (f *fakeChunkUpstream) serveBytes(w http.ResponseWriter, r *http.Request, b []byte) {
	w.Header().Set("Accept-Ranges", "bytes")
	start, end := int64(0), int64(len(b))-1
	if h := r.Header.Get("Range"); h != "" {
		var s, e int64
		if _, err := fmt.Sscanf(h, "bytes=%d-%d", &s, &e); err != nil || s < 0 || e >= int64(len(b)) || s > e {
			w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}
		start, end = s, e
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", s, e, len(b)))
		w.WriteHeader(http.StatusPartialContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_, _ = w.Write(b[start : end+1])
}

func decodeChunkToken(tok string) ([]byte, error) {
	pad := func(s string) string {
		if m := len(s) % 4; m != 0 {
			s += strings.Repeat("=", 4-m)
		}
		return s
	}
	outer, err := base64.StdEncoding.DecodeString(pad(tok))
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(pad(string(outer)))
}

// newChunkedFile builds a chunked abyssFile over a fake upstream.
func newChunkedFile(srv *httptest.Server, size, fdSize int64) *abyssFile {
	host := strings.TrimPrefix(srv.URL, "https://")
	return &abyssFile{
		size:   size,
		origin: "https://playembedapi.site",
		chunk: &abyssChunkSource{
			md5ID: 99, resID: 2, sub: "x", codec: "h264", label: "360p",
			chunkURL:  srv.URL + "/mp4/99/2/" + strconv.FormatInt(size, 10),
			chunkPath: "/mp4/99/2/" + strconv.FormatInt(size, 10),
			chunkHost: host,
			firstURL:  srv.URL + "/d/head.fd",
			firstSize: fdSize,
			firstSeed: "head.fd",
		},
	}
}

// TestChunkedProxyRoundTrip serves a synthetic file through the local proxy:
// head via encrypted .fd, body via token-checked /sora/ parts, then reads it
// back the way mpv would, including ranges straddling the head/part split.
func TestChunkedProxyRoundTrip(t *testing.T) {
	t.Parallel()
	const size = 5_000_100
	plain := make([]byte, size)
	copy(plain, "\x00\x00\x00\x20ftypisom\x00\x00\x02\x00")
	for i := 16; i < size; i++ {
		plain[i] = byte(i % 251)
	}
	const chunk = int64(2 << 20)
	const fdSize = chunk // first part through .fd
	up := &fakeChunkUpstream{
		plain: plain, fdSeed: "head.fd", fdSize: fdSize,
		size: size, md5ID: 99, resID: 2, chunk: chunk, origin: "https://playembedapi.site",
	}
	srv := httptest.NewTLSServer(up)
	t.Cleanup(srv.Close)
	f := newChunkedFile(srv, size, fdSize)

	ctx := context.Background()
	// Probe path: head decrypts to ftyp, and the first token part opens.
	head, err := readAbyssRange(ctx, srv.Client(), f, 0, 15)
	if err != nil {
		t.Fatalf("head: %v", err)
	}
	if string(head[4:8]) != "ftyp" {
		t.Fatalf("head is not an MP4: % x", head[:16])
	}
	split, err := readAbyssRange(ctx, srv.Client(), f, fdSize, fdSize+15)
	if err != nil {
		t.Fatalf("first token part: %v", err)
	}
	if !bytes.Equal(split, plain[fdSize:fdSize+16]) {
		t.Fatal("token part bytes differ from the source file")
	}
	if up.hitsFor("fd") == 0 || up.hitsFor("sora:1") == 0 {
		t.Error("upstream never saw both transports")
	}

	local, err := sharedStreamProxy().serve(f, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	get := func(rng string) (int, http.Header, []byte) {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, local, http.NoBody)
		req.Header.Set("Range", rng)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("GET %s: %v", rng, err)
		}
		defer func() { _ = resp.Body.Close() }()
		body, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, resp.Header, body
	}
	if code, _, body := get("bytes=0-15"); code != http.StatusPartialContent || string(body[4:8]) != "ftyp" {
		t.Errorf("head range: HTTP %d % x", code, body)
	}
	// Straddles the .fd/sora split and the 64 KiB decryption edge.
	for _, span := range [][2]int64{{65530, 65545}, {fdSize - 100, fdSize + 100}} {
		lo, hi := span[0], span[1]
		code, _, body := get(fmt.Sprintf("bytes=%d-%d", lo, hi))
		if code != http.StatusPartialContent || !bytes.Equal(body, plain[lo:hi+1]) {
			t.Errorf("range %d-%d: HTTP %d, %d bytes", lo, hi, code, len(body))
		}
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, local, http.NoBody)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	full, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if !bytes.Equal(full, plain) {
		t.Errorf("full file through proxy differs: got %d bytes, want %d", len(full), len(plain))
	}
}

// TestChunkedProxyFailsFastBeforeHeaders pins the failure contract: when the
// first needed part is unavailable, mpv gets an explicit 502 with no partial
// headers, and the failed fetch is never cached — the next request retries
// the origin once it is healthy again.
func TestChunkedProxyFailsFastBeforeHeaders(t *testing.T) {
	t.Parallel()
	const size = 5_000_100
	plain := make([]byte, size)
	const chunk = int64(2 << 20)
	up := &fakeChunkUpstream{
		plain: plain, fdSeed: "head.fd", fdSize: chunk,
		size: size, md5ID: 99, resID: 2, chunk: chunk, origin: "https://playembedapi.site",
	}
	up.queue("sora:1", http.StatusNotFound)
	srv := httptest.NewTLSServer(up)
	t.Cleanup(srv.Close)
	f := newChunkedFile(srv, size, chunk)
	local, err := sharedStreamProxy().serve(f, srv.Client())
	if err != nil {
		t.Fatal(err)
	}

	get := func(rng string) *http.Response {
		req, _ := http.NewRequest(http.MethodGet, local, http.NoBody)
		req.Header.Set("Range", rng)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = resp.Body.Close() })
		return resp
	}
	// Deep range whose first part 404s: explicit 502, no partial headers.
	resp := get(fmt.Sprintf("bytes=%d-%d", chunk, chunk+15))
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("want 502 before headers, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Content-Range") != "" {
		t.Error("502 must not carry partial-response headers")
	}
	if got := up.hitsFor("sora:1"); got != 1 {
		t.Errorf("404 should not be retried: %d origin hits", got)
	}
	// The failure was not cached: once the origin recovers, the same range
	// serves normally.
	resp2 := get(fmt.Sprintf("bytes=%d-%d", chunk, chunk+15))
	if resp2.StatusCode != http.StatusPartialContent {
		t.Fatalf("range after recovery: %d", resp2.StatusCode)
	}
	body, _ := io.ReadAll(resp2.Body)
	if !bytes.Equal(body, plain[chunk:chunk+16]) {
		t.Error("bytes after recovery differ from the source file")
	}
}

// TestChunkedProxyRetriesTransient pins the retry policy: a 503 part is
// retried once and then succeeds; a 503 that never clears fails after the
// bounded attempts, not forever.
func TestChunkedProxyRetriesTransient(t *testing.T) {
	t.Parallel()
	const size = 5_000_100
	plain := make([]byte, size)
	const chunk = int64(2 << 20)
	up := &fakeChunkUpstream{
		plain: plain, fdSeed: "head.fd", fdSize: chunk,
		size: size, md5ID: 99, resID: 2, chunk: chunk, origin: "https://playembedapi.site",
	}
	up.queue("sora:1", http.StatusServiceUnavailable)
	srv := httptest.NewTLSServer(up)
	t.Cleanup(srv.Close)
	f := newChunkedFile(srv, size, chunk)
	local, err := sharedStreamProxy().serve(f, srv.Client())
	if err != nil {
		t.Fatal(err)
	}

	get := func() (*http.Response, []byte) {
		req, _ := http.NewRequest(http.MethodGet, local, http.NoBody)
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", chunk, chunk+15))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = resp.Body.Close() })
		body, _ := io.ReadAll(resp.Body)
		return resp, body
	}
	// One 503 then success: the retry serves the range.
	resp, body := get()
	if resp.StatusCode != http.StatusPartialContent || !bytes.Equal(body, plain[chunk:chunk+16]) {
		t.Fatalf("after one 503: HTTP %d, % x", resp.StatusCode, body)
	}
	if got, want := up.hitsFor("sora:1"), 2; got != want {
		t.Errorf("origin hits = %d, want %d (one retry)", got, want)
	}

	// A part that stays down: bounded attempts, then an explicit 502.
	up.queue("sora:2", http.StatusServiceUnavailable, http.StatusServiceUnavailable, http.StatusServiceUnavailable)
	resp2, _ := func() (*http.Response, []byte) {
		req, _ := http.NewRequest(http.MethodGet, local, http.NoBody)
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", 2*chunk, 2*chunk+15))
		r, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = r.Body.Close() })
		b, _ := io.ReadAll(r.Body)
		return r, b
	}()
	if resp2.StatusCode != http.StatusBadGateway {
		t.Fatalf("permanently failing part: %d, want 502", resp2.StatusCode)
	}
	if got, want := up.hitsFor("sora:2"), 1+abyssPartRetries; got != want {
		t.Errorf("origin hits = %d, want %d (bounded retries)", got, want)
	}
}

// TestChunkedProxyCachesParts pins reuse: a second read of the same range
// comes from the cache, so the origin sees one hit, and the bytes match.
func TestChunkedProxyCachesParts(t *testing.T) {
	t.Parallel()
	const size = 5_000_100
	plain := make([]byte, size)
	for i := range plain {
		plain[i] = byte(i % 199)
	}
	const chunk = int64(2 << 20)
	up := &fakeChunkUpstream{
		plain: plain, fdSeed: "head.fd", fdSize: chunk,
		size: size, md5ID: 99, resID: 2, chunk: chunk, origin: "https://playembedapi.site",
	}
	srv := httptest.NewTLSServer(up)
	t.Cleanup(srv.Close)
	f := newChunkedFile(srv, size, chunk)
	local, err := sharedStreamProxy().serve(f, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	rng := fmt.Sprintf("bytes=%d-%d", chunk, chunk+999)
	bodies := make([][]byte, 2)
	for i := range bodies {
		req, _ := http.NewRequest(http.MethodGet, local, http.NoBody)
		req.Header.Set("Range", rng)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusPartialContent {
			t.Fatalf("read %d: HTTP %d", i, resp.StatusCode)
		}
		bodies[i] = b
	}
	if !bytes.Equal(bodies[0], plain[chunk:chunk+1000]) || !bytes.Equal(bodies[1], bodies[0]) {
		t.Error("cached read differs from the source file")
	}
	if got := up.hitsFor("sora:1"); got != 1 {
		t.Errorf("origin hits = %d, want 1 (second read from cache)", got)
	}
}

// TestChunkedProxyCoalescesFetches pins the concurrency contract: two
// simultaneous reads of the same part share one origin request, and going
// away does not take the shared fetch down.
func TestChunkedProxyCoalescesFetches(t *testing.T) {
	t.Parallel()
	const size = 5_000_100
	plain := make([]byte, size)
	const chunk = int64(2 << 20)
	up := &fakeChunkUpstream{
		plain: plain, fdSeed: "head.fd", fdSize: chunk,
		size: size, md5ID: 99, resID: 2, chunk: chunk, origin: "https://playembedapi.site",
		delay: 200 * time.Millisecond, // hold the first fetch open
	}
	srv := httptest.NewTLSServer(up)
	t.Cleanup(srv.Close)
	f := newChunkedFile(srv, size, chunk)
	local, err := sharedStreamProxy().serve(f, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	rng := fmt.Sprintf("bytes=%d-%d", chunk, chunk+63)
	var wg sync.WaitGroup
	bodies := make([][]byte, 2)
	errs := make([]error, 2)
	for i := range bodies {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req, _ := http.NewRequest(http.MethodGet, local, http.NoBody)
			req.Header.Set("Range", rng)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				errs[i] = err
				return
			}
			body, rerr := io.ReadAll(resp.Body)
			cerr := resp.Body.Close()
			bodies[i] = body
			if rerr != nil {
				errs[i] = rerr
			} else {
				errs[i] = cerr
			}
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("read %d: %v", i, err)
		}
		if !bytes.Equal(bodies[i], plain[chunk:chunk+64]) {
			t.Fatalf("read %d differs from the source file", i)
		}
	}
	if got := up.hitsFor("sora:1"); got != 1 {
		t.Errorf("origin hits = %d, want 1 (concurrent reads coalesced)", got)
	}
}

// newPartsFixture serves a size-byte file with known bytes through a fake
// upstream whose head is the first part; configure runs before it serves.
func newPartsFixture(t *testing.T, size int64, configure func(*fakeChunkUpstream)) (*fakeChunkUpstream, *httptest.Server, *abyssFile, []byte) {
	t.Helper()
	const chunk = int64(2 << 20)
	plain := make([]byte, size)
	for i := range plain {
		plain[i] = byte(i % 241)
	}
	up := &fakeChunkUpstream{
		plain: plain, fdSeed: "head.fd", fdSize: chunk,
		size: size, md5ID: 99, resID: 2, chunk: chunk, origin: "https://playembedapi.site",
	}
	if configure != nil {
		configure(up)
	}
	srv := httptest.NewTLSServer(up)
	t.Cleanup(srv.Close)
	return up, srv, newChunkedFile(srv, size, chunk), plain
}

func partCached(f *abyssFile, idx int64) bool {
	_, ok := abyssParts.get(abyssPartKey{f.partsCache(), idx})
	return ok
}

func inflightParts(f *abyssFile) map[int64]*abyssWaiter {
	c := f.partsCache()
	c.mu.Lock()
	defer c.mu.Unlock()
	return maps.Clone(c.inflight)
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestChunkedReadAheadPrefetches pins the read-ahead: a short probe read
// prefetches nothing, while a reader holding its part gets the next
// abyssReadAhead parts of its range unasked — and nothing past them — so
// reading them later costs no origin hit.
func TestChunkedReadAheadPrefetches(t *testing.T) {
	t.Parallel()
	const chunk = int64(2 << 20)
	const size = 10*chunk + 123
	up, srv, f, plain := newPartsFixture(t, size, nil)
	ctx := context.Background()

	if _, err := readAbyssRange(ctx, srv.Client(), f, 0, 15); err != nil {
		t.Fatal(err)
	}
	if n := len(inflightParts(f)); n != 0 {
		t.Fatalf("a 16-byte probe started %d prefetches", n)
	}

	rd := f.openPartReader(srv.Client(), chunk, size-1)
	defer rd.close()
	if _, err := rd.part(ctx, 1); err != nil {
		t.Fatal(err)
	}
	for idx := int64(2); idx <= 1+abyssReadAhead; idx++ {
		waitFor(t, fmt.Sprintf("part %d prefetched", idx), func() bool { return partCached(f, idx) })
	}
	if got := up.hitsFor("sora:" + strconv.Itoa(2+abyssReadAhead)); got != 0 {
		t.Errorf("part past the read-ahead window fetched %d times", got)
	}
	data, err := rd.part(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, plain[2*chunk:3*chunk]) {
		t.Error("prefetched part differs from the source file")
	}
	if got := up.hitsFor("sora:2"); got != 1 {
		t.Errorf("origin hits for a prefetched part = %d, want 1", got)
	}
}

// TestChunkedSeekDropsStalePrefetch pins seeking: a reader that hangs up
// leaves its prefetches running (mpv may reopen at the same spot), but a
// reader landing elsewhere cancels the ones no window covers, and nothing
// from them is cached.
func TestChunkedSeekDropsStalePrefetch(t *testing.T) {
	t.Parallel()
	const chunk = int64(2 << 20)
	const size = 18 * chunk
	_, srv, f, plain := newPartsFixture(t, size, func(up *fakeChunkUpstream) {
		up.hold = map[string]bool{}
		for idx := 2; idx <= 1+abyssReadAhead; idx++ {
			up.hold["sora:"+strconv.Itoa(idx)] = true
		}
	})
	ctx := context.Background()

	rd := f.openPartReader(srv.Client(), chunk, size-1)
	if _, err := rd.part(ctx, 1); err != nil {
		t.Fatal(err)
	}
	held := inflightParts(f)
	if len(held) != abyssPrefetchers {
		t.Fatalf("prefetches in flight = %d, want %d", len(held), abyssPrefetchers)
	}
	rd.close()
	if n := len(inflightParts(f)); n != len(held) {
		t.Fatalf("hanging up dropped prefetches: %d of %d left", n, len(held))
	}

	seek := f.openPartReader(srv.Client(), 15*chunk, size-1)
	defer seek.close()
	for idx := range inflightParts(f) {
		if idx < 15 {
			t.Errorf("part %d still in flight after the seek", idx)
		}
	}
	for idx, w := range held {
		select {
		case <-w.done:
			if w.err == nil {
				t.Errorf("dropped prefetch of part %d succeeded", idx)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("prefetch of part %d still running after the seek", idx)
		}
		if partCached(f, idx) {
			t.Errorf("dropped part %d was cached", idx)
		}
	}
	data, err := seek.part(ctx, 15)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, plain[15*chunk:16*chunk]) {
		t.Error("part at the seek target differs from the source file")
	}
}

// TestChunkedPrefetchFailureLeftToReader pins that a failed prefetch is not
// retried in a loop: the parts past it still prefetch, and the failed one is
// fetched again only when the reader gets there.
func TestChunkedPrefetchFailureLeftToReader(t *testing.T) {
	t.Parallel()
	const chunk = int64(2 << 20)
	const size = 6 * chunk
	up, srv, f, plain := newPartsFixture(t, size, func(up *fakeChunkUpstream) {
		up.failures = map[string][]int{"sora:2": {http.StatusNotFound}}
	})
	ctx := context.Background()

	rd := f.openPartReader(srv.Client(), chunk, size-1)
	defer rd.close()
	if _, err := rd.part(ctx, 1); err != nil {
		t.Fatal(err)
	}
	for idx := int64(3); idx <= 5; idx++ {
		waitFor(t, fmt.Sprintf("part %d prefetched", idx), func() bool { return partCached(f, idx) })
	}
	waitFor(t, "prefetching to settle", func() bool { return len(inflightParts(f)) == 0 })
	if got := up.hitsFor("sora:2"); got != 1 {
		t.Errorf("failed prefetch hit the origin %d times, want 1", got)
	}
	data, err := rd.part(ctx, 2)
	if err != nil {
		t.Fatalf("reader fetch after a failed prefetch: %v", err)
	}
	if !bytes.Equal(data, plain[2*chunk:3*chunk]) {
		t.Error("part differs from the source file")
	}
	if got := up.hitsFor("sora:2"); got != 2 {
		t.Errorf("origin hits = %d, want 2 (prefetch, then the reader)", got)
	}
}

// TestAbyssPartStoreLRU pins the shared store: least recently used goes
// first, a read refreshes a part, both caps hold, and files never collide.
func TestAbyssPartStoreLRU(t *testing.T) {
	t.Parallel()
	a, b := newAbyssPartCache(), newAbyssPartCache()
	s := newAbyssPartStore(10, 100)
	s.put(abyssPartKey{a, 1}, []byte("aaaa"))
	s.put(abyssPartKey{b, 1}, []byte("bbbb"))
	if got, ok := s.get(abyssPartKey{a, 1}); !ok || string(got) != "aaaa" {
		t.Fatalf("a/1 = %q, %v", got, ok)
	}
	s.put(abyssPartKey{a, 2}, []byte("cccc")) // 12 bytes: b/1 is least recent
	if _, ok := s.get(abyssPartKey{b, 1}); ok {
		t.Error("b/1 survived past the byte budget")
	}
	for _, k := range []abyssPartKey{{a, 1}, {a, 2}} {
		if _, ok := s.get(k); !ok {
			t.Errorf("part %d evicted out of LRU order", k.idx)
		}
	}
	if s.bytes != 8 {
		t.Errorf("bytes = %d, want 8", s.bytes)
	}

	s = newAbyssPartStore(1<<20, 2)
	for idx := int64(1); idx <= 3; idx++ {
		s.put(abyssPartKey{a, idx}, []byte{byte(idx)})
	}
	if _, ok := s.get(abyssPartKey{a, 1}); ok || s.order.Len() != 2 {
		t.Errorf("entry cap not held: %d entries", s.order.Len())
	}
}

// addToCounterBytewise is the byte-at-a-time carry loop addToCounter
// replaced, kept as the reference the 64-bit version must agree with.
func addToCounterBytewise(ctr []byte, n uint64) {
	for i := len(ctr) - 1; i >= 0 && n > 0; i-- {
		sum := uint64(ctr[i]) + (n & 0xFF)
		ctr[i] = byte(sum)
		n = (n >> 8) + (sum >> 8)
	}
}

// TestAddToCounterMatchesBytewiseCarry pins addToCounter to the loop it
// replaced on the carries that matter — none, low half into high half, a
// chain through every byte, the wrap at 2^128 — and on random counters.
func TestAddToCounterMatchesBytewiseCarry(t *testing.T) {
	t.Parallel()
	ff := bytes.Repeat([]byte{0xFF}, aes.BlockSize)
	type tc struct {
		ctr []byte
		n   uint64
	}
	cases := []tc{
		{make([]byte, aes.BlockSize), 0},
		{make([]byte, aes.BlockSize), 1},
		{append(make([]byte, 8), ff[:8]...), 1},
		{ff, 1},
		{ff, math.MaxUint64},
		{append(bytes.Repeat([]byte{0x12}, 8), ff[:8]...), math.MaxUint64},
	}
	rng := rand.New(rand.NewPCG(1, 2))
	for range 20000 {
		ctr := make([]byte, aes.BlockSize)
		for i := range ctr {
			ctr[i] = byte(rng.Uint32())
		}
		n := rng.Uint64()
		if rng.IntN(2) == 0 {
			n >>= rng.UintN(64) // small offsets too, like real file positions
		}
		cases = append(cases, tc{ctr, n})
	}
	for _, c := range cases {
		want := bytes.Clone(c.ctr)
		addToCounterBytewise(want, c.n)
		var got [aes.BlockSize]byte
		copy(got[:], c.ctr)
		addToCounter(&got, c.n)
		if !bytes.Equal(got[:], want) {
			t.Fatalf("ctr % x + %d = % x, want % x", c.ctr, c.n, got, want)
		}
	}
}

// TestAbyssCTRSeekMatchesStream pins seeking against the standard library's
// own CTR: decrypting from any offset equals the same slice of one continuous
// keystream, including counters whose low 64 bits, or all 128, roll over
// inside the span. Negative offsets are refused, not turned into a counter.
func TestAbyssCTRSeekMatchesStream(t *testing.T) {
	t.Parallel()
	key := bytes.Repeat([]byte{7}, 32)
	ivs := [][]byte{
		make([]byte, aes.BlockSize),
		{0, 0, 0, 0, 0, 0, 0, 1, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFD},
		append(bytes.Repeat([]byte{0xFF}, aes.BlockSize-1), 0xFE),
	}
	const span = 8 * aes.BlockSize
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	for _, iv := range ivs {
		stream := make([]byte, span)
		cipher.NewCTR(block, iv).XORKeyStream(stream, stream)
		for off := range span {
			for _, n := range []int{1, aes.BlockSize, span - off} {
				if off+n > span {
					continue
				}
				got := make([]byte, n)
				if err := abyssCTR(key, iv, int64(off), got); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got, stream[off:off+n]) {
					t.Fatalf("iv % x: keystream at %d+%d differs from the continuous stream", iv, off, n)
				}
			}
		}
	}
	if err := abyssCTR(key, ivs[0], -1, make([]byte, 1)); err == nil {
		t.Error("negative offset accepted")
	}
}

// TestUTF8ToLatin1Bounds pins the narrowing: every byte value round-trips,
// and nothing outside 0x00-0xFF — including what invalid UTF-8 decodes to —
// is ever truncated into a byte.
func TestUTF8ToLatin1Bounds(t *testing.T) {
	t.Parallel()
	all := make([]byte, 256)
	for i := range all {
		all[i] = byte(i)
	}
	got, err := utf8ToLatin1(string(latin1ToUTF8(all)))
	if err != nil || !bytes.Equal(got, all) {
		t.Fatalf("round trip: %v, % x", err, got)
	}
	for _, bad := range []string{"Ā", "Ł", "�", "\xff", "ok\xc3", "\U0001F600"} {
		if out, err := utf8ToLatin1(bad); err == nil {
			t.Errorf("%q narrowed to % x instead of failing", bad, out)
		}
	}
}

// recordingTransport records every upstream URL requested through it.
type recordingTransport struct {
	next http.RoundTripper
	mu   sync.Mutex
	urls []string
}

func (r *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r.mu.Lock()
	r.urls = append(r.urls, req.URL.String())
	r.mu.Unlock()
	return r.next.RoundTrip(req)
}

func (r *recordingTransport) seen() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.urls)
}

// TestChunkedProxyRangeOnlySelectsParts is the G704 proof for the proxy. The
// Range header is the only part of a client request that reaches upstream
// fetches; whatever it says, every upstream request — reads and prefetches —
// is an entry of the file's own URL table, and ranges that are malformed or
// out of bounds never reach upstream at all.
func TestChunkedProxyRangeOnlySelectsParts(t *testing.T) {
	t.Parallel()
	const chunk = int64(2 << 20)
	const size = 4*chunk + 77
	_, srv, f, plain := newPartsFixture(t, size, nil)
	rec := &recordingTransport{next: srv.Client().Transport}
	client := &http.Client{Transport: rec}
	local, err := sharedStreamProxy().serve(f, client)
	if err != nil {
		t.Fatal(err)
	}
	table, err := f.abyssPartURLs()
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{}
	for _, u := range table {
		allowed[u] = true
	}

	get := func(rng string) int {
		req, _ := http.NewRequest(http.MethodGet, local, http.NoBody)
		req.Header.Set("Range", rng)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Range %q: %v", rng, err)
		}
		defer func() { _ = resp.Body.Close() }()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusPartialContent {
			var s, e, total int64
			if _, err := fmt.Sscanf(resp.Header.Get("Content-Range"), "bytes %d-%d/%d", &s, &e, &total); err != nil {
				t.Fatalf("Range %q: Content-Range %q", rng, resp.Header.Get("Content-Range"))
			}
			if !bytes.Equal(body, plain[s:e+1]) {
				t.Errorf("Range %q: body differs from the file", rng)
			}
		}
		return resp.StatusCode
	}

	refused := []string{
		fmt.Sprintf("bytes=%d-%d", size, size+10),
		"bytes=9223372036854775807-", "bytes=-9223372036854775808", "bytes=-0",
		"bytes=0-1,5-6", "bytes=5-1", "bytes=abc-", "items=0-1",
		"bytes=https://evil.test/-", "bytes=-1-2", "bytes= 0-1;evil.test",
	}
	for _, rng := range refused {
		if code := get(rng); code != http.StatusRequestedRangeNotSatisfiable {
			t.Errorf("Range %q: HTTP %d, want 416", rng, code)
		}
	}
	if n := len(rec.seen()); n != 0 {
		t.Fatalf("refused ranges reached upstream %d times", n)
	}

	served := []string{
		"bytes=-1", "bytes=-99999999999", fmt.Sprintf("bytes=%d-", size-1),
		fmt.Sprintf("bytes=%d-%d", chunk-1, chunk), "bytes=0-9223372036854775807",
	}
	for _, rng := range served {
		if code := get(rng); code != http.StatusPartialContent {
			t.Errorf("Range %q: HTTP %d, want 206", rng, code)
		}
	}
	waitFor(t, "prefetching to settle", func() bool { return len(inflightParts(f)) == 0 })
	urls := rec.seen()
	if len(urls) == 0 {
		t.Fatal("no upstream request was recorded")
	}
	// Indices past either end of the table fail before any request.
	for _, idx := range []int64{-1, int64(len(table)), math.MaxInt64} {
		if _, err := f.abyssFetchPart(context.Background(), client, idx); err == nil {
			t.Errorf("part %d fetched", idx)
		}
	}
	if n := len(rec.seen()); n != len(urls) {
		t.Errorf("out-of-range parts reached upstream %d times", n-len(urls))
	}
	for _, u := range urls {
		if !allowed[u] {
			t.Errorf("upstream request outside the file's URL table: %s", u)
		}
	}
}

// TestIsAbyssHostname pins what the media list may name as a host: a bare
// DNS name, never a port, user info, path, query, fragment or IP literal.
func TestIsAbyssHostname(t *testing.T) {
	t.Parallel()
	for _, ok := range []string{"o25chikcb28.sssrr.org", "qq.cdn.test", "a-b.c", "localhost"} {
		if !isAbyssHostname(ok) {
			t.Errorf("%q refused", ok)
		}
	}
	for _, bad := range []string{
		"", "127.0.0.1", "169.254.169.254", "::1", "[::1]", "host:443", "user@host",
		"host/path", "host?q", "host#f", "-a.b", "a-.b", "a..b", "a b", "h\x00st",
		strings.Repeat("a", 64) + ".com", strings.Repeat("a.", 127) + "com",
	} {
		if isAbyssHostname(bad) {
			t.Errorf("%q accepted", bad)
		}
	}
}

// TestMakeAbyssFileRejectsSteeredHosts pins that the remote media list cannot
// steer the proxy off a plain https host name: such renditions are skipped, a
// bad head entry is dropped (its parts then come through the tokens), and a
// bad single-file URL falls through to the chunked renditions.
func TestMakeAbyssFileRejectsSteeredHosts(t *testing.T) {
	t.Parallel()
	embed := mustParseURL(t, "https://playembedapi.site/?v=x")
	datas := &abyssDatas{Slug: "x", MD5ID: 42, UserID: 7}
	chunked := func(domains []string, sub, head string) *abyssMedia {
		m := &abyssMedia{}
		m.MP4.Sources = []abyssSource{{Label: "720p", ResID: 4, Size: 1 << 22, Codec: "h264", Sub: sub}}
		m.MP4.Domains = domains
		if head != "" {
			m.MP4.FristDatas = []abyssFirstData{{ResID: 4, Size: 1 << 22, Codec: "h264", URL: head, PartSize: 2 << 20}}
		}
		return m
	}
	// Each domain matches the rendition's sub, so only the host check can
	// refuse it.
	for _, d := range []struct{ domain, sub string }{
		{"qq.cdn.test:8443", "qq"}, {"qq.cdn.test/x?", "qq"}, {"user@qq.cdn.test", "qq"},
		{"127.0.0.1", "127"}, {"169.254.169.254", "169"},
	} {
		if f, ok := makeAbyssFile(embed, datas, chunked([]string{d.domain}, d.sub, "")); ok {
			t.Errorf("domain %q resolved to host %q", d.domain, f.chunk.chunkHost)
		}
	}
	for _, head := range []string{"https://127.0.0.1/h.fd", "http://qq.cdn.test/h.fd", "https://u@qq.cdn.test/h.fd", "https://qq.cdn.test:1/h.fd"} {
		f, ok := makeAbyssFile(embed, datas, chunked([]string{"qq.cdn.test"}, "qq", head))
		if !ok || f.chunk == nil || f.chunk.firstURL != "" {
			t.Errorf("head %q: ok=%v file=%+v, want chunked without the head", head, ok, f)
		}
	}
	m := chunked([]string{"qq.cdn.test"}, "qq", "")
	m.MP4.Sources = append(m.MP4.Sources, abyssSource{Label: "1080p", ResID: 5, Size: 1 << 22, Codec: "h264", URL: "http://10.0.0.1", Path: "v.mp4"})
	if f, ok := makeAbyssFile(embed, datas, m); !ok || f.chunk == nil || f.upstream != "" {
		t.Errorf("bad single-file URL: ok=%v file=%+v, want the chunked rendition", ok, f)
	}
}

// TestMakeAbyssFilePicksTallestRendition pins the quality rule for Abyss: the
// tallest picture wins whatever its codec or layout; H.264 only breaks a tie
// in height, and a single file only a tie with a chunked rendition.
func TestMakeAbyssFilePicksTallestRendition(t *testing.T) {
	t.Parallel()
	embed := mustParseURL(t, "https://playembedapi.site/?v=x")
	datas := &abyssDatas{Slug: "x", MD5ID: 42, UserID: 7}
	const size = 1 << 22
	chunked := func(label string, res int, codec string) abyssSource {
		return abyssSource{Label: label, ResID: res, Size: size, Codec: codec, Sub: "qq"}
	}
	single := func(label string, res int, codec string) abyssSource {
		return abyssSource{Label: label, ResID: res, Size: size, Codec: codec, URL: "https://files.cdn.test/" + label, Path: "v.mp4"}
	}
	tests := []struct {
		name       string
		sources    []abyssSource
		wantHeight int
		wantCodec  string // chunked picks
		wantSingle bool
	}{
		{name: "AV1-only 1080p beats H.264 720p",
			sources:    []abyssSource{chunked("720p", 4, "h264"), chunked("1080p", 5, "av1")},
			wantHeight: 1080, wantCodec: "av1"},
		{name: "same height prefers H.264",
			sources:    []abyssSource{chunked("1080p", 5, "av1"), chunked("1080p", 5, "h264"), chunked("720p", 4, "h264")},
			wantHeight: 1080, wantCodec: "h264"},
		{name: "chunked 1080p beats a 720p single file",
			sources:    []abyssSource{single("720p", 4, "h264"), chunked("1080p", 5, "h264")},
			wantHeight: 1080, wantCodec: "h264"},
		{name: "single file wins a tie",
			sources:    []abyssSource{chunked("1080p", 5, "h264"), single("1080p", 5, "h264")},
			wantHeight: 1080, wantSingle: true},
		{name: "taller single file beats chunked",
			sources:    []abyssSource{chunked("720p", 4, "h264"), single("1080p", 5, "av1")},
			wantHeight: 1080, wantSingle: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			media := &abyssMedia{}
			media.MP4.Sources = tt.sources
			media.MP4.Domains = []string{"qq.cdn.test"}
			f, ok := makeAbyssFile(embed, datas, media)
			if !ok {
				t.Fatal("no rendition picked")
			}
			if f.height != tt.wantHeight {
				t.Errorf("height = %d, want %d", f.height, tt.wantHeight)
			}
			if gotSingle := f.chunk == nil; gotSingle != tt.wantSingle {
				t.Fatalf("single file = %v, want %v", gotSingle, tt.wantSingle)
			}
			if !tt.wantSingle && f.chunk.codec != tt.wantCodec {
				t.Errorf("codec = %s, want %s", f.chunk.codec, tt.wantCodec)
			}
		})
	}
}

// TestChunkedAdvanceKeepsAnotherRangesFetch pins that sequential reading
// never cancels anything. A downloader reads one file as several ranges:
// range A timed out with part 2 in flight and was about to retry at the same
// spot, while range B, further in, moved on — and B's move cancelled A's
// fetch, so the retry started over (seen live: "part=85 ... context
// canceled"). Only a reader opening elsewhere — a seek — drops fetches now.
func TestChunkedAdvanceKeepsAnotherRangesFetch(t *testing.T) {
	t.Parallel()
	const chunk = int64(2 << 20)
	const size = 18 * chunk
	_, srv, f, _ := newPartsFixture(t, size, func(up *fakeChunkUpstream) {
		up.hold = map[string]bool{"sora:2": true}
	})
	ctx := context.Background()
	b := f.openPartReader(srv.Client(), 10*chunk, size-1)
	defer b.close()
	if _, err := b.part(ctx, 10); err != nil {
		t.Fatal(err)
	}

	a := f.openPartReader(srv.Client(), 2*chunk, 9*chunk)
	actx, cancel := context.WithCancel(ctx)
	go func() { _, _ = a.part(actx, 2) }()
	waitFor(t, "part 2 in flight", func() bool { _, ok := inflightParts(f)[2]; return ok })
	held := inflightParts(f)[2]
	t.Cleanup(held.cancel) // release the held request when the test ends
	cancel()
	a.close()

	if _, err := b.part(ctx, 11); err != nil {
		t.Fatal(err)
	}
	select {
	case <-held.done:
		t.Fatalf("range B moving on cancelled range A's in-flight part: %v", held.err)
	default:
	}
	retry := f.openPartReader(srv.Client(), 2*chunk, 9*chunk) // A's retry, same spot
	defer retry.close()
	if got := inflightParts(f)[2]; got != held {
		t.Error("the retry did not find the original fetch still going")
	}
}

// TestProbeAbyssFetchesHeadAndFirstPartAtOnce pins the probe's parallelism:
// it opens the encrypted head and the first token part concurrently. Each
// request here waits for the other to arrive, so a probe that fetched them
// one after the other would fail.
func TestProbeAbyssFetchesHeadAndFirstPartAtOnce(t *testing.T) {
	t.Parallel()
	const chunk = int64(2 << 20)
	const size = 2*chunk + 1000
	plain := make([]byte, size)
	copy(plain, "\x00\x00\x00\x20ftypisom\x00\x00\x02\x00")
	up := &fakeChunkUpstream{
		plain: plain, fdSeed: "head.fd", fdSize: chunk,
		size: size, md5ID: 99, resID: 2, chunk: chunk, origin: "https://playembedapi.site",
	}
	var arrived atomic.Int32
	var bothOnce sync.Once
	both := make(chan struct{})
	barrier := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retries can bring more than two requests; the second arrival opens
		// the barrier for every one of them.
		if arrived.Add(1) >= 2 {
			bothOnce.Do(func() { close(both) })
		}
		select {
		case <-both:
			up.ServeHTTP(w, r)
		case <-time.After(3 * time.Second):
			w.WriteHeader(http.StatusGatewayTimeout) // the other request never came
		}
	})
	srv := httptest.NewTLSServer(barrier)
	t.Cleanup(srv.Close)
	f := newChunkedFile(srv, size, chunk)
	c := &Client{http: srv.Client(), baseURL: "https://www.startflix.test", userAgent: userAgent}
	if err := c.probeAbyss(context.Background(), f); err != nil {
		t.Fatalf("probe: %v (head and first part were not fetched at once)", err)
	}
}
