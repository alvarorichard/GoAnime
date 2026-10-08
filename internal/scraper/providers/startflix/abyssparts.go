package startflix

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

// Chunked Abyss renditions are fetched one player part at a time (2 MiB by
// default). Seeks revisit parts, so the proxy keeps a per-file LRU of
// complete, successfully fetched parts and coalesces concurrent fetches of
// the same part: one origin request serves every caller, and a failed fetch
// is never cached as data.
const (
	abyssPartCacheBytes = 64 << 20 // per-file budget for cached parts
	abyssPartCacheMax   = 256      // per-file entry cap, bounds the map and list
	abyssPartRetries    = 2        // beyond the first attempt, transient failures only
	abyssPartTimeout    = 30 * time.Second
)

// abyssPart is one cache entry: a complete part of the file, decrypted.
type abyssPart struct {
	idx  int64
	data []byte
}

// abyssWaiter coalesces concurrent fetches of one part. The fetch runs
// detached from any single caller's cancellation, so a caller that goes away
// cannot take the result down with it while others still wait for it.
type abyssWaiter struct {
	done chan struct{}
	data []byte
	err  error
}

// abyssPartCache is a per-file LRU of parts plus the in-flight fetches. The
// list is small (bounded by abyssPartCacheMax), so a linear scan beats a map
// plus intrusive list bookkeeping.
type abyssPartCache struct {
	mu       sync.Mutex
	lru      []*abyssPart // index 0 = most recently used
	bytes    int64
	inflight map[int64]*abyssWaiter
}

func newAbyssPartCache() *abyssPartCache {
	return &abyssPartCache{inflight: map[int64]*abyssWaiter{}}
}

func (c *abyssPartCache) get(idx int64) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, p := range c.lru {
		if p.idx == idx {
			copy(c.lru[1:i+1], c.lru[:i])
			c.lru[0] = p
			return p.data, true
		}
	}
	return nil, false
}

func (c *abyssPartCache) put(idx int64, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, p := range c.lru {
		if p.idx == idx {
			return
		}
	}
	c.lru = append([]*abyssPart{{idx: idx, data: data}}, c.lru...)
	c.bytes += int64(len(data)) // #nosec G115 -- bounded by the file size
	for len(c.lru) > abyssPartCacheMax || c.bytes > abyssPartCacheBytes {
		evicted := c.lru[len(c.lru)-1]
		c.lru = c.lru[:len(c.lru)-1]
		c.bytes -= int64(len(evicted.data)) // #nosec G115 -- see above
	}
}

// abyssGetPart returns the decrypted bytes of file part idx through the cache.
// Concurrent callers for the same part share one origin request.
func (f *abyssFile) abyssGetPart(ctx context.Context, client *http.Client, idx int64) ([]byte, error) {
	c := f.partsCache()
	if data, ok := c.get(idx); ok {
		return data, nil
	}

	c.mu.Lock()
	if w, ok := c.inflight[idx]; ok {
		c.mu.Unlock()
		select {
		case <-w.done:
			return w.data, w.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	w := &abyssWaiter{done: make(chan struct{})}
	c.inflight[idx] = w
	c.mu.Unlock()

	// Detached from the caller's cancellation: the result is cached for the
	// next caller even if this one hangs up mid-seek.
	fctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), abyssPartTimeout)
	data, err := f.abyssFetchPart(fctx, client, idx)
	cancel()

	c.mu.Lock()
	delete(c.inflight, idx)
	w.data, w.err = data, err
	close(w.done)
	c.mu.Unlock()
	if err == nil {
		c.put(idx, data)
	}
	return data, err
}

// abyssFetchPart fetches one complete part from the CDN and decrypts what
// needs it. Transient failures (network errors, 5xx, 429) are retried a
// couple of times with a short backoff; 404s and 403s fail immediately, as
// retrying a rejected path would only stall the response.
func (f *abyssFile) abyssFetchPart(ctx context.Context, client *http.Client, idx int64) ([]byte, error) {
	c := f.chunk
	lo, hi := f.abyssPartBounds(idx)
	if lo < 0 || hi < lo {
		return nil, fmt.Errorf("part %d out of range", idx)
	}
	want := hi - lo + 1

	var req string
	var from, through int64 // absolute .fd range; -1 sends no Range header
	var seed string
	if c.firstSize > 0 && lo < c.firstSize {
		req, from, through, seed = c.firstURL, lo, min(hi, c.firstSize-1), c.firstSeed
	} else {
		token := abyssChunkToken(c.partTokenPath(f.abyssPartSize(), idx), f.size)
		if token == "" {
			return nil, errors.New("could not build the Abyss chunk token")
		}
		req = "https://" + c.chunkHost + "/sora/" + strconv.FormatInt(f.size, 10) + "/" + token
		from, through = -1, -1
	}

	started := time.Now()
	var body []byte
	var err error
	for attempt := 0; ; attempt++ {
		body, err = abyssPartRequest(ctx, client, req, from, through, want, f.origin)
		if err == nil {
			break
		}
		if ctx.Err() != nil || attempt >= abyssPartRetries || !abyssRetryableFailure(err) {
			util.Debug("StartFlix chunk part failed", "part", idx, "range", fmt.Sprintf("%d-%d", lo, hi),
				"attempt", attempt+1, "dur", time.Since(started).Round(time.Millisecond).String(), "err", err)
			return nil, err
		}
		util.Debug("StartFlix chunk part retry", "part", idx, "attempt", attempt+1, "err", err)
		if err := abyssSleep(ctx, time.Duration(attempt+1)*250*time.Millisecond); err != nil {
			return nil, err
		}
	}
	if seed != "" && from < abyssEncryptedHead {
		n := min(int64(len(body)), abyssEncryptedHead-from) // #nosec G115 -- from is a file offset
		if err := abyssXOR(seed, from, body[:n]); err != nil {
			return nil, err
		}
	}
	return body, nil
}

// abyssPartRequest performs one upstream request for a part. The response
// must be a complete part: short or overlong bodies are rejected rather than
// padded or truncated into the cache.
func abyssPartRequest(ctx context.Context, client *http.Client, rawURL string, from, through, want int64, origin string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", origin)
	req.Header.Set("Referer", origin+"/")
	if from >= 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", from, through))
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, abyssRedactTransportErr(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return nil, netx.NewHTTPStatusError(SourceName, "abyss chunk", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, want+1)) // #nosec G115 -- want is a part length
	if err != nil {
		return nil, abyssRedactTransportErr(err)
	}
	if int64(len(body)) != want { // #nosec G115 -- see above
		return nil, fmt.Errorf("abyss chunk returned %d bytes, expected %d", len(body), want)
	}
	return body, nil
}

// abyssRetryableFailure reports whether a part fetch is worth retrying:
// network errors and origin-side statuses, never client errors like 404/403.
func abyssRetryableFailure(err error) bool {
	var diag *netx.SourceDiagnostic
	if errors.As(err, &diag) {
		switch diag.StatusCode {
		case http.StatusRequestTimeout, http.StatusTooManyRequests,
			http.StatusInternalServerError, http.StatusBadGateway,
			http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return true
		}
		return false
	}
	// Transport error: retry unless it is our own deadline or cancellation.
	return !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
}

// abyssRedactTransportErr strips the upstream URL from transport errors: the
// /sora/ URL embeds the chunk token, which never belongs in a log.
func abyssRedactTransportErr(err error) error {
	var ue *url.Error
	if errors.As(err, &ue) {
		return &url.Error{Op: ue.Op, URL: "abyss-chunk", Err: ue.Err}
	}
	return err
}

func abyssSleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// abyssPartSize is the player's part granularity, min(size, 2 MiB). It is part
// of the token path, so it must match the player exactly.
func (f *abyssFile) abyssPartSize() int64 {
	const part = int64(2 << 20)
	if f.size < part {
		return f.size
	}
	return part
}

// abyssPartBounds returns the [lo, hi] file byte range of part idx.
func (f *abyssFile) abyssPartBounds(idx int64) (lo, hi int64) {
	part := f.abyssPartSize()
	lo = idx * part
	hi = min(lo+part-1, f.size-1)
	return
}

// partTokenPath builds the virtual path the player's token encrypts.
func (c *abyssChunkSource) partTokenPath(chunkSize, idx int64) string {
	return c.chunkPath + "/" + strconv.FormatInt(chunkSize, 10) + "/" + strconv.FormatInt(idx, 10)
}
