package startflix

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// tailFlakyUpstream serves a synthetic file through /sora/ only (no .fd head).
// The tail part fails whole-part reads (no Range, or a Range larger than a
// slice) with a truncated 200 body, but serves Range slices of at most
// abyssFetchSliceBytes correctly — the live CDN shape that stalled a 876 MB
// download at 99% with "unexpected EOF" on its last 640 KiB.
type tailFlakyUpstream struct {
	plain  []byte
	size   int64
	md5ID  int64
	resID  int
	chunk  int64
	origin string
	tail   int64
}

func decryptSoraToken(tok []byte, size int64) []byte {
	key, iv := abyssChunkKey(size)
	block, _ := aes.NewCipher(key)
	ctr := make([]byte, aes.BlockSize)
	copy(ctr, iv)
	stream := cipher.NewCTR(block, ctr)
	stream.XORKeyStream(tok, tok)
	return tok
}

func (f *tailFlakyUpstream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != f.origin {
		w.WriteHeader(http.StatusForbidden)
		return
	}
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
	token = decryptSoraToken(token, f.size)
	part, err := parseSoraPart(token, f.md5ID, f.resID, f.size, f.chunk)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	start := part * f.chunk
	end := min(start+f.chunk, f.size)
	want := end - start
	rng := r.Header.Get("Range")
	// Tail whole-part reads (or oversized ranges) get a truncated body.
	if part == f.tail && (rng == "" || rangeLen(rng, want) > abyssFetchSliceBytes) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(f.plain[start : start+100])
		return
	}
	serveSoraBytes(w, r, f.plain[start:end])
}

func parseSoraPart(token []byte, md5ID int64, resID int, size, chunk int64) (int64, error) {
	// Token plaintext is "/mp4/<md5>/<res>/<size>/<chunk>/<part>".
	path, want := string(token), fmt.Sprintf("/mp4/%d/%d/%d/%d/", md5ID, resID, size, chunk)
	if !strings.HasPrefix(path, want) {
		return 0, fmt.Errorf("bad path %q", path)
	}
	return strconv.ParseInt(strings.TrimPrefix(path, want), 10, 64)
}

func rangeLen(h string, want int64) int64 {
	var s, e int64
	if _, err := fmt.Sscanf(h, "bytes=%d-%d", &s, &e); err != nil {
		return want + 1
	}
	return e - s + 1
}

func serveSoraBytes(w http.ResponseWriter, r *http.Request, b []byte) {
	w.Header().Set("Accept-Ranges", "bytes")
	if h := r.Header.Get("Range"); h != "" {
		var s, e int64
		if _, err := fmt.Sscanf(h, "bytes=%d-%d", &s, &e); err != nil || s < 0 || e >= int64(len(b)) || s > e {
			w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", s, e, len(b)))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(b[s : e+1])
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

func newTailFlakyFile(srv *httptest.Server, size int64) *abyssFile {
	host := strings.TrimPrefix(srv.URL, "https://")
	host = strings.TrimPrefix(host, "http://")
	return &abyssFile{
		size:   size,
		origin: "https://playembedapi.site",
		chunk: &abyssChunkSource{
			md5ID: 99, resID: 2, sub: "x", codec: "h264", label: "1080p",
			chunkURL:  srv.URL + "/mp4/99/2/" + strconv.FormatInt(size, 10),
			chunkPath: "/mp4/99/2/" + strconv.FormatInt(size, 10),
			chunkHost: host,
		},
	}
}

// TestAbyssTailRecoversViaSlices pins the 99%-stall fix: a tail part whose
// whole-part read always truncates is re-fetched as Range slices and
// assembles byte-for-byte.
func TestAbyssTailRecoversViaSlices(t *testing.T) {
	t.Parallel()
	const part = int64(2 << 20)
	const tailSize = 600 * 1024 // two slices: 512 KiB + ~88 KiB
	const size = 2*part + tailSize
	plain := make([]byte, size)
	for i := range plain {
		plain[i] = byte(i % 251)
	}
	up := &tailFlakyUpstream{
		plain: plain, size: size, md5ID: 99, resID: 2, chunk: part,
		origin: "https://playembedapi.site", tail: 2,
	}
	srv := httptest.NewTLSServer(up)
	t.Cleanup(srv.Close)
	f := newTailFlakyFile(srv, size)

	got, err := f.abyssFetchPart(context.Background(), srv.Client(), 2)
	if err != nil {
		t.Fatalf("tail part: %v", err)
	}
	if len(got) != int(tailSize) || string(got) != string(plain[2*part:]) {
		t.Fatalf("tail bytes differ: got %d bytes, want %d", len(got), tailSize)
	}
}

// siblingPrimaryFailsTail serves 500 for the tail part on every request;
// the sibling serves it correctly.
type siblingPrimaryFailsTail struct {
	plain  []byte
	size   int64
	origin string
	tail   int64
	fail   bool // this server is the failing primary
}

func (f *siblingPrimaryFailsTail) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != f.origin {
		w.WriteHeader(http.StatusForbidden)
		return
	}
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
	token = decryptSoraToken(token, f.size)
	// Token plaintext ends in "/<part>".
	path := string(token)
	slash := strings.LastIndex(path, "/")
	part, _ := strconv.ParseInt(path[slash+1:], 10, 64)
	if f.fail && part == f.tail {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	start := part * int64(2<<20)
	serveSoraBytes(w, r, f.plain[start:min(start+int64(2<<20), f.size)])
}

// TestAbyssTailFailsOverToSibling pins the sibling-domain fallback: when the
// primary host answers 500 for the tail on every attempt, the same token
// succeeds on the sibling host.
func TestAbyssTailFailsOverToSibling(t *testing.T) {
	t.Parallel()
	const part = int64(2 << 20)
	const tailSize = 1000
	const size = 2*part + tailSize
	plain := make([]byte, size)
	for i := range plain {
		plain[i] = byte(i % 241)
	}
	primary := &siblingPrimaryFailsTail{plain: plain, size: size, origin: "https://playembedapi.site", tail: 2, fail: true}
	sibling := &siblingPrimaryFailsTail{plain: plain, size: size, origin: "https://playembedapi.site", tail: 2}
	primarySrv := httptest.NewTLSServer(primary)
	t.Cleanup(primarySrv.Close)
	siblingSrv := httptest.NewTLSServer(sibling)
	t.Cleanup(siblingSrv.Close)

	hostOf := func(srv *httptest.Server) string {
		h := strings.TrimPrefix(srv.URL, "https://")
		return strings.TrimPrefix(h, "http://")
	}
	f := &abyssFile{
		size:   size,
		origin: "https://playembedapi.site",
		chunk: &abyssChunkSource{
			md5ID: 99, resID: 2, sub: "x", codec: "h264", label: "1080p",
			chunkPath: "/mp4/99/2/" + strconv.FormatInt(size, 10),
			chunkHost: hostOf(primarySrv),
			domains:   []string{hostOf(primarySrv), hostOf(siblingSrv)},
		},
	}

	// The primary client works for both test servers (same test CA).
	got, err := f.abyssFetchPart(context.Background(), primarySrv.Client(), 2)
	if err != nil {
		t.Fatalf("tail via sibling: %v", err)
	}
	if string(got) != string(plain[2*part:]) {
		t.Fatalf("sibling tail bytes differ: got %d bytes, want %d", len(got), tailSize)
	}
}
