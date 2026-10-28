package startflix

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// truncatingUpstream serves several chunked renditions on one host. Each
// rendition whose size is in truncate answers its final part one 64 KiB block
// short — the live shape of a bad upload: the origin declares the part whole
// and then stops early. Every other rendition serves whole.
type truncatingUpstream struct {
	plain    map[int64][]byte
	chunk    int64
	origin   string
	truncate map[int64]bool
}

func (u *truncatingUpstream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != u.origin {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, "/sora/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) != 2 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	size, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	plain, ok := u.plain[size]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	token, err := decodeChunkToken(parts[1])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	token = decryptSoraToken(token, size)
	path := string(token)
	slash := strings.LastIndex(path, "/")
	if slash < 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	part, err := strconv.ParseInt(path[slash+1:], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	start := part * u.chunk
	end := min(start+u.chunk, size)
	if start < 0 || start >= size {
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return
	}
	seg := plain[start:end]
	if u.truncate[size] && end == size {
		cut := int64(65536)
		if int64(len(seg)) > cut {
			seg = seg[:int64(len(seg))-cut]
		} else {
			seg = seg[:0]
		}
	}
	serveSoraBytes(w, r, seg)
}

// TestSelectAbyssFileSkipsTruncatedTail pins the live fix: the best-ranked
// rendition answers its final part short, so the probe rejects it and the next
// rendition of the same height is chosen. Without the tail probe the download
// stalls at 99% retrying the missing part forever.
func TestSelectAbyssFileSkipsTruncatedTail(t *testing.T) {
	t.Parallel()
	const chunk = int64(2 << 20)
	// The broken rendition, like the live one, has a final part that is a
	// whole number of 64 KiB blocks; the healthy one ends mid-block.
	const brokenSize = 2*chunk + 10*65536
	const healthySize = 2*chunk + 7*65536 + 1234

	plain := map[int64][]byte{}
	for _, size := range []int64{brokenSize, healthySize} {
		b := make([]byte, size)
		for i := range b {
			b[i] = byte(i % 251)
		}
		copy(b, "\x00\x00\x00\x20ftypisom\x00\x00\x02\x00")
		plain[size] = b
	}
	up := &truncatingUpstream{
		plain: plain, chunk: chunk, origin: "https://playembedapi.site",
		truncate: map[int64]bool{brokenSize: true},
	}
	srv := httptest.NewTLSServer(up)
	t.Cleanup(srv.Close)
	host := strings.TrimPrefix(srv.URL, "https://")

	mk := func(size int64) *abyssFile {
		return &abyssFile{
			size:   size,
			origin: "https://playembedapi.site",
			chunk: &abyssChunkSource{
				md5ID: 99, resID: 5, sub: host, codec: "h264", label: "1080p",
				chunkPath: "/mp4/99/5/" + strconv.FormatInt(size, 10),
				chunkHost: host,
			},
		}
	}
	c := &Client{http: srv.Client(), baseURL: "https://www.startflix.test", userAgent: userAgent}
	files := []*abyssFile{mk(brokenSize), mk(healthySize)}

	got, err := c.selectAbyssFile(context.Background(), files)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if got.size != healthySize {
		t.Fatalf("picked the %d-byte rendition, want the %d-byte one", got.size, healthySize)
	}
	// The chosen rendition must really fetch whole.
	body, err := got.abyssFetchPart(context.Background(), srv.Client(), (got.size-1)/chunk)
	if err != nil {
		t.Fatalf("healthy tail: %v", err)
	}
	lo, hi := got.abyssPartBounds((got.size - 1) / chunk)
	if len(body) != int(hi-lo+1) {
		t.Fatalf("healthy tail: got %d bytes, want %d", len(body), hi-lo+1)
	}
}

// TestSelectAbyssFileFailsWhenAllTailsTruncated pins that a batch whose every
// rendition is cut short is refused rather than streamed into the 99% stall.
func TestSelectAbyssFileFailsWhenAllTailsTruncated(t *testing.T) {
	t.Parallel()
	const chunk = int64(2 << 20)
	const size = 2*chunk + 10*65536
	b := make([]byte, size)
	for i := range b {
		b[i] = byte(i % 241)
	}
	copy(b, "\x00\x00\x00\x20ftypisom\x00\x00\x02\x00")
	up := &truncatingUpstream{
		plain: map[int64][]byte{size: b}, chunk: chunk,
		origin: "https://playembedapi.site", truncate: map[int64]bool{size: true},
	}
	srv := httptest.NewTLSServer(up)
	t.Cleanup(srv.Close)
	host := strings.TrimPrefix(srv.URL, "https://")
	f := &abyssFile{
		size: size, origin: "https://playembedapi.site",
		chunk: &abyssChunkSource{
			md5ID: 99, resID: 5, sub: host, codec: "h264", label: "1080p",
			chunkPath: "/mp4/99/5/" + strconv.FormatInt(size, 10),
			chunkHost: host,
		},
	}
	c := &Client{http: srv.Client(), baseURL: "https://www.startflix.test", userAgent: userAgent}
	if _, err := c.selectAbyssFile(context.Background(), []*abyssFile{f}); err == nil {
		t.Fatal("a batch with only truncated tails was accepted")
	}
}

// TestMakeAbyssFilesOrdersFallbacks pins the candidate order: the quality rule
// that always picked one file now orders the whole list, so a first choice the
// CDN serves badly yields to the renditions behind it.
func TestMakeAbyssFilesOrdersFallbacks(t *testing.T) {
	t.Parallel()
	embed := mustParseURL(t, "https://playembedapi.site/?v=x")
	datas := &abyssDatas{Slug: "x", MD5ID: 42, UserID: 7}
	media := &abyssMedia{}
	media.MP4.Domains = []string{"alpha.cdn.test", "bravo.cdn.test", "charlie.cdn.test", "delta.cdn.test"}
	const size = 1 << 22
	media.MP4.Sources = []abyssSource{
		{Label: "1080p", ResID: 5, Size: size, Codec: "h264", Sub: "alpha"},
		{Label: "720p", ResID: 4, Size: size, Codec: "h264", Sub: "delta"},
		{Label: "1080p", ResID: 5, Size: size, Codec: "h264", Sub: "bravo"},
		{Label: "1080p", ResID: 5, Size: size, Codec: "av1", Sub: "charlie"},
	}
	files := makeAbyssFiles(embed, datas, media)
	if len(files) != 4 {
		t.Fatalf("candidates = %d, want 4", len(files))
	}
	want := []string{"alpha.cdn.test", "bravo.cdn.test", "charlie.cdn.test", "delta.cdn.test"}
	for i, f := range files {
		if f.chunk == nil || f.chunk.chunkHost != want[i] {
			t.Errorf("candidate %d host = %v, want %s", i, f.chunk, want[i])
		}
	}
	if files[0].height != 1080 {
		t.Errorf("first candidate height = %d, want 1080", files[0].height)
	}
}
