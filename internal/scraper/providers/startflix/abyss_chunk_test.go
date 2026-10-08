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
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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
	delay    time.Duration // applied to a target's first hit
}

// record counts a hit on target, consumes a queued failure status (0 means
// none) and returns it; a target's first successful hit sleeps once, to hold
// the fetch open for the coalescing test.
func (f *fakeChunkUpstream) record(target string) int {
	f.mu.Lock()
	if f.hits == nil {
		f.hits = map[string]int{}
	}
	f.hits[target]++
	status := 0
	if q := f.failures[target]; len(q) > 0 {
		status, f.failures[target] = q[0], q[1:]
	}
	first, d := f.hits[target] == 1, f.delay
	f.mu.Unlock()
	if status == 0 && first && d > 0 {
		time.Sleep(d)
	}
	return status
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
		if status := f.record("fd"); status != 0 {
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
	if status := f.record("sora:" + strconv.Itoa(part)); status != 0 {
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
func newChunkedFile(srv *httptest.Server, plain []byte, size, chunk, fdSize int64) *abyssFile {
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
	f := newChunkedFile(srv, plain, size, chunk, fdSize)

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
	f := newChunkedFile(srv, plain, size, chunk, chunk)
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
	f := newChunkedFile(srv, plain, size, chunk, chunk)
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
	f := newChunkedFile(srv, plain, size, chunk, chunk)
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
	f := newChunkedFile(srv, plain, size, chunk, chunk)
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
