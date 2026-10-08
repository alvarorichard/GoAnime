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
	mu     sync.Mutex
	base   string // http://127.0.0.1:port
	files  map[string]*proxiedFile
	order  []string
	client *http.Client
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
	switch {
	case first == "":
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
		writeRangeHeaders(w, start, end, f.size, partial)
		if err := copyAbyssChunkRange(r.Context(), f.client, f.abyssFile, start, end, w); err != nil && r.Context().Err() == nil {
			util.Debug("StartFlix chunk proxy failed", "err", err, "range", fmt.Sprintf("%d-%d", start, end))
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
	resp, err := f.client.Do(req)
	if err != nil {
		util.Debug("StartFlix proxy upstream failed", "err", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
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

func copyAbyssChunkRange(ctx context.Context, client *http.Client, f *abyssFile, start, end int64, dst io.Writer) error {
	c := f.chunk
	chunkSize := int64(2 << 20) // player core's `min(size, 0x200000)`
	if f.size < chunkSize {
		chunkSize = f.size
	}
	for pos := start; pos <= end; {
		part := pos / chunkSize
		partStart := part * chunkSize
		partEnd := min(partStart+chunkSize-1, f.size-1)
		to := min(end, partEnd)
		upstream := ""
		from, through := pos-partStart, to-partStart
		seed := ""
		if pos < c.firstSize {
			upstream = c.firstURL
			from, through = pos, min(to, c.firstSize-1)
			seed = c.firstSeed
			to = through
			if from > through {
				return fmt.Errorf("invalid Abyss first-data range %d-%d", from, through)
			}
		} else {
			logicalPath := strings.TrimRight(c.chunkURL, "/") + "/" + strconv.FormatInt(chunkSize, 10) + "/" + strconv.FormatInt(part, 10)
			token := abyssChunkToken(logicalPath, f.size)
			if token == "" {
				return errors.New("could not build Abyss chunk token")
			}
			upstream = "https://" + c.chunkHost + "/sora/" + strconv.FormatInt(f.size, 10) + "/" + token
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, upstream, http.NoBody)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Origin", "https://player.abyssplayer.com")
		req.Header.Set("Referer", "https://player.abyssplayer.com/")
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", from, through))
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return readErr
		}
		if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
			return netx.NewHTTPStatusError(SourceName, "abyss chunk", resp.StatusCode)
		}
		if resp.StatusCode == http.StatusOK && len(body) > int(through-from+1) {
			body = body[from : through+1]
		}
		if int64(len(body)) != through-from+1 {
			return fmt.Errorf("Abyss chunk returned %d bytes, expected %d", len(body), through-from+1)
		}
		if seed != "" {
			if from < abyssEncryptedHead {
				n := min(int64(len(body)), abyssEncryptedHead-from)
				if err := abyssXOR(seed, from, body[:n]); err != nil {
					return err
				}
			}
		}
		if _, err := dst.Write(body); err != nil {
			return err
		}
		pos = to + 1
	}
	return nil
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
