package startflix

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

// streamProxy is a loopback HTTP server that hands mpv and the downloader a
// plain MP4 for an Abyss file: it forwards each Range to the CDN with the
// headers the CDN wants, and decrypts whatever part of the response falls in
// the encrypted head. Both consumers already speak Range over HTTP, so neither
// needs to know Abyss exists.
type streamProxy struct {
	mu    sync.Mutex
	base  string // http://127.0.0.1:port
	files map[string]*proxiedFile
	order []string
}

type proxiedFile struct {
	*abyssFile
	client *http.Client
}

// maxProxiedFiles bounds the registry: a binge registers one file per episode,
// and only the latest few can still be playing or downloading.
const maxProxiedFiles = 64

var (
	proxyOnce   sync.Once
	proxyShared *streamProxy
)

func sharedStreamProxy() *streamProxy {
	proxyOnce.Do(func() { proxyShared = &streamProxy{files: map[string]*proxiedFile{}} })
	return proxyShared
}

// proxyClient is the client the proxy uses upstream: SSRF-guarded like every
// scraper request, but with no whole-request timeout, because a playback
// response is read for as long as the episode plays.
func (c *Client) proxyClient() *http.Client {
	if c.stream != nil {
		return c.stream
	}
	return c.http
}

func newStreamClient() *http.Client {
	t := netx.SafeScraperTransport(30 * time.Second)
	t.ResponseHeaderTimeout = 30 * time.Second
	return &http.Client{Transport: t}
}

// serve registers f and returns its local URL, starting the listener on first
// use.
func (p *streamProxy) serve(f *abyssFile, client *http.Client) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.base == "" {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return "", fmt.Errorf("local stream proxy: %w", err)
		}
		srv := &http.Server{Handler: p, ReadHeaderTimeout: 10 * time.Second}
		go func() {
			if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
				util.Debug("StartFlix stream proxy stopped", "err", err)
			}
		}()
		p.base = "http://" + ln.Addr().String()
		util.RegisterLocalProxyHost(ln.Addr().String())
		util.Debug("StartFlix stream proxy listening", "addr", p.base)
	}

	var raw [12]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	id := hex.EncodeToString(raw[:])
	p.files[id] = &proxiedFile{abyssFile: f, client: client}
	p.order = append(p.order, id)
	for len(p.order) > maxProxiedFiles {
		delete(p.files, p.order[0])
		p.order = p.order[1:]
	}
	return p.base + "/abyss/" + id + ".mp4", nil
}

func (p *streamProxy) lookup(path string) (*proxiedFile, bool) {
	id := strings.TrimSuffix(strings.TrimPrefix(path, "/abyss/"), ".mp4")
	p.mu.Lock()
	defer p.mu.Unlock()
	f, ok := p.files[id]
	return f, ok
}

// parseRange reads a single "bytes=a-b" / "bytes=a-" / "bytes=-n" range against
// a file of size bytes. ok is false for anything else, which is answered 416.
func parseRange(header string, size int64) (start, end int64, ok bool) {
	spec, found := strings.CutPrefix(strings.TrimSpace(header), "bytes=")
	if !found || strings.Contains(spec, ",") {
		return 0, 0, false
	}
	first, last, found := strings.Cut(spec, "-")
	if !found {
		return 0, 0, false
	}
	var err error
	switch first {
	case "":
		n, err := strconv.ParseInt(last, 10, 64)
		if err != nil || n <= 0 {
			return 0, 0, false
		}
		start, end = max(size-n, 0), size-1
	default:
		if start, err = strconv.ParseInt(first, 10, 64); err != nil {
			return 0, 0, false
		}
		end = size - 1
		if last != "" {
			if end, err = strconv.ParseInt(last, 10, 64); err != nil {
				return 0, 0, false
			}
			end = min(end, size-1)
		}
	}
	if start < 0 || start > end || start >= size {
		return 0, 0, false
	}
	return start, end, true
}

func (p *streamProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, ok := p.lookup(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", "video/mp4")

	start, end := int64(0), f.size-1
	partial := false
	if h := r.Header.Get("Range"); h != "" {
		s, e, ok := parseRange(h, f.size)
		if !ok {
			w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", f.size))
			w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}
		start, end, partial = s, e, true
	}

	switch r.Method {
	case http.MethodHead:
		writeRangeHeaders(w, start, end, f.size, partial)
		return
	case http.MethodGet:
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if f.chunk != nil {
		// parseRange has already bounded [start, end] to the file.
		rd := f.openPartReader(f.client, start, end)
		defer rd.close()
		// Open the first part this response needs before promising mpv any
		// bytes: a dead origin must fail as 502 here, not as a 206 that
		// truncates mid-stream once the headers have gone out.
		first := start / f.abyssPartSize()
		if _, err := rd.part(r.Context(), first); err != nil {
			util.Debug("StartFlix chunk proxy failed before headers", "err", err,
				"range", fmt.Sprintf("%d-%d", start, end), "part", first)
			http.Error(w, "upstream chunk unavailable", http.StatusBadGateway)
			return
		}
		writeRangeHeaders(w, start, end, f.size, partial)
		if err := rd.copyRange(r.Context(), start, end, w); err != nil && r.Context().Err() == nil {
			// Headers (and possibly bytes) are already out; all we can do is
			// stop writing. The short body tells mpv the response ended badly.
			util.Debug("StartFlix chunk proxy failed mid-stream", "err", err,
				"range", fmt.Sprintf("%d-%d", start, end))
		}
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, f.upstream, http.NoBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	f.decorate(req, userAgent)
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	resp, err := f.upstreamWithRetry(r.Context(), req)
	if err != nil {
		util.Debug("StartFlix proxy upstream failed", "err", err, "range", req.Header.Get("Range"))
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusPartialContent {
		util.Debug("StartFlix proxy upstream refused", "status", resp.StatusCode, "range", req.Header.Get("Range"))
		http.Error(w, "upstream "+resp.Status, http.StatusBadGateway)
		return
	}

	writeRangeHeaders(w, start, end, f.size, partial)
	body := io.Reader(resp.Body)
	if start < abyssEncryptedHead {
		body = &headDecrypter{r: resp.Body, seed: f.seed, pos: start}
	}
	if _, err := io.Copy(w, body); err != nil && r.Context().Err() == nil {
		util.Debug("StartFlix proxy copy ended", "err", err)
	}
}

// upstreamWithRetry sends a single-file upstream request, retrying transient
// failures (network errors, 5xx, 429) a couple of times with a short backoff
// before the response headers reach the client.
func (p *proxiedFile) upstreamWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	started := time.Now()
	var resp *http.Response
	var err error
	for attempt := 0; ; attempt++ {
		resp, err = p.client.Do(req)
		if err == nil && resp.StatusCode != http.StatusRequestTimeout &&
			resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < 500 {
			return resp, nil
		}
		if ctx.Err() != nil || attempt >= abyssPartRetries {
			break
		}
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil, err
			}
		} else {
			_ = resp.Body.Close()
		}
		util.Debug("StartFlix proxy upstream retry", "attempt", attempt+1,
			"status", statusOrZero(resp, err), "err", errStringOrNone(err))
		if rerr := abyssSleep(ctx, time.Duration(attempt+1)*250*time.Millisecond); rerr != nil {
			return nil, rerr
		}
	}
	if err == nil {
		util.Debug("StartFlix proxy upstream unavailable", "status", resp.StatusCode,
			"range", req.Header.Get("Range"), "dur", time.Since(started).Round(time.Millisecond).String())
		return resp, nil
	}
	return nil, err
}

func statusOrZero(resp *http.Response, err error) int {
	if err != nil || resp == nil {
		return 0
	}
	return resp.StatusCode
}

func errStringOrNone(err error) string {
	if err == nil {
		return "none"
	}
	return err.Error()
}

// readAbyssRange fetches a small range through the same chunk path used by the
// player's Service Worker. It is used during resolution to reject stale URLs
// before handing the stream to mpv.
func readAbyssRange(ctx context.Context, client *http.Client, f *abyssFile, start, end int64) ([]byte, error) {
	var out bytes.Buffer
	if err := copyAbyssChunkRange(ctx, client, f, start, end, &out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// copyAbyssChunkRange streams [start, end] out of the player's parts through
// a reader of its own: each part comes from the shared cache (concurrent
// callers share the fetch) and the requested slice is written from it.
func copyAbyssChunkRange(ctx context.Context, client *http.Client, f *abyssFile, start, end int64, dst io.Writer) error {
	if err := f.checkChunkRange(start, end); err != nil {
		return err
	}
	rd := f.openPartReader(client, start, end)
	defer rd.close()
	return rd.copyRange(ctx, start, end, dst)
}

func writeRangeHeaders(w http.ResponseWriter, start, end, size int64, partial bool) {
	w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
	if partial {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
		w.WriteHeader(http.StatusPartialContent)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// headDecrypter decrypts the bytes of r that fall in the encrypted head, given
// that r starts at file offset pos. Everything past the head passes through.
type headDecrypter struct {
	r    io.Reader
	seed string
	pos  int64
}

func (d *headDecrypter) Read(b []byte) (int, error) {
	n, err := d.r.Read(b)
	if n > 0 && d.pos < abyssEncryptedHead {
		k := int(min(int64(n), abyssEncryptedHead-d.pos))
		if xerr := abyssXOR(d.seed, d.pos, b[:k]); xerr != nil {
			return 0, xerr
		}
	}
	d.pos += int64(n)
	return n, err
}
