package startflix

import (
	"container/list"
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
// default). Three things keep playback and seeking smooth on top of that:
//
//   - Every fetched part lands in one LRU shared by all files, under a single
//     memory budget. Seeking back replays recent parts without the network,
//     and the parts of episodes nobody watches any more age out instead of
//     piling up per registered file.
//   - Each open read (one mpv connection, one downloader range) is a reader
//     with a position. Once its current part is in hand, the next parts of
//     its range are prefetched in parallel, nearest first, so sequential
//     playback rarely waits on the origin.
//   - Concurrent fetches of a part share one origin request, and when a new
//     reader opens somewhere else (a seek: players seek by opening a new
//     connection), in-flight fetches that no reader wants any more are
//     cancelled so the new position gets the bandwidth.
//
// A failed fetch is never cached as data.
const (
	abyssPartCacheBytes = 128 << 20 // shared budget for cached parts, across every file
	abyssPartCacheMax   = 512       // shared entry cap, bounds the map and list
	abyssReadAhead      = 6         // parts prefetched past a reader's position (12 MiB)
	abyssPrefetchers    = 3         // concurrent prefetches per file
	abyssPartRetries    = 2         // beyond the first attempt, transient failures only
	abyssPartTimeout    = 30 * time.Second
)

// abyssPartKey names one part of one file.
type abyssPartKey struct {
	c   *abyssPartCache
	idx int64
}

// abyssPart is one cache entry: a complete part of a file, decrypted.
type abyssPart struct {
	key  abyssPartKey
	data []byte
}

// abyssPartStore is the LRU of complete parts shared by every file. Its lock
// is a leaf: callers may hold a file's abyssPartCache.mu, never the reverse.
type abyssPartStore struct {
	mu         sync.Mutex
	budget     int64
	maxEntries int
	order      *list.List // front = most recently used; values are *abyssPart
	index      map[abyssPartKey]*list.Element
	bytes      int64
}

var abyssParts = newAbyssPartStore(abyssPartCacheBytes, abyssPartCacheMax)

func newAbyssPartStore(budget int64, maxEntries int) *abyssPartStore {
	return &abyssPartStore{
		budget:     budget,
		maxEntries: maxEntries,
		order:      list.New(),
		index:      map[abyssPartKey]*list.Element{},
	}
}

// get returns a cached part and marks it most recently used.
func (s *abyssPartStore) get(k abyssPartKey) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.index[k]
	if !ok {
		return nil, false
	}
	s.order.MoveToFront(e)
	return e.Value.(*abyssPart).data, true
}

// put stores one complete, successfully fetched part as most recent and
// evicts from the back until both caps hold.
func (s *abyssPartStore) put(k abyssPartKey, data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.index[k]; ok {
		s.order.MoveToFront(e)
		return
	}
	s.index[k] = s.order.PushFront(&abyssPart{key: k, data: data})
	s.bytes += int64(len(data))
	for s.order.Len() > s.maxEntries || s.bytes > s.budget {
		e := s.order.Back()
		p := e.Value.(*abyssPart)
		s.order.Remove(e)
		delete(s.index, p.key)
		s.bytes -= int64(len(p.data))
	}
}

// abyssWaiter is one in-flight part fetch; every caller that wants the part
// waits on it. The fetch runs detached from all of them, so a caller that
// goes away cannot take it down for the others. It ends on its own deadline,
// or early once a seek leaves it outside every reader's window.
type abyssWaiter struct {
	done     chan struct{}
	data     []byte
	err      error
	cancel   context.CancelFunc
	prefetch bool
}

// abyssPartCache is one file's side of the shared store: its in-flight
// fetches and where each reader streaming it is, under one lock so a lookup,
// the in-flight check and the store of a finished fetch are atomic together.
type abyssPartCache struct {
	mu          sync.Mutex
	inflight    map[int64]*abyssWaiter
	readers     map[*abyssPartReader]*abyssReadWindow
	prefetching int                // in-flight fetches started as prefetches
	failed      map[int64]struct{} // prefetches that failed: left to the reader that gets there
}

func newAbyssPartCache() *abyssPartCache {
	return &abyssPartCache{
		inflight: map[int64]*abyssWaiter{},
		readers:  map[*abyssPartReader]*abyssReadWindow{},
		failed:   map[int64]struct{}{},
	}
}

// abyssGetPart returns the decrypted bytes of file part idx through the cache.
// Concurrent callers for the same part share one origin request.
func (f *abyssFile) abyssGetPart(ctx context.Context, client *http.Client, idx int64) ([]byte, error) {
	c := f.partsCache()
	c.mu.Lock()
	if data, ok := abyssParts.get(abyssPartKey{c, idx}); ok {
		c.mu.Unlock()
		return data, nil
	}
	w, ok := c.inflight[idx]
	if !ok {
		w = c.fetchLocked(f, client, idx, false)
	}
	c.mu.Unlock()
	select {
	case <-w.done:
		return w.data, w.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// fetchLocked starts a detached fetch of part idx and registers it in flight.
// When it ends, the result is stored (on success) and the file's prefetching
// is topped up again. The caller holds c.mu.
func (c *abyssPartCache) fetchLocked(f *abyssFile, client *http.Client, idx int64, prefetch bool) *abyssWaiter {
	ctx, cancel := context.WithTimeout(context.Background(), abyssPartTimeout)
	w := &abyssWaiter{done: make(chan struct{}), cancel: cancel, prefetch: prefetch}
	c.inflight[idx] = w
	if prefetch {
		c.prefetching++
	}
	go func() {
		data, err := f.abyssFetchPart(ctx, client, idx)
		cancel()

		c.mu.Lock()
		defer c.mu.Unlock()
		// A fetch dropped as stale is no longer in inflight, and a newer
		// fetch of the same part may have taken its slot since.
		stale := c.inflight[idx] != w
		if !stale {
			delete(c.inflight, idx)
		}
		switch {
		case err == nil:
			abyssParts.put(abyssPartKey{c, idx}, data)
			delete(c.failed, idx)
		case prefetch && !stale:
			c.failed[idx] = struct{}{}
		}
		w.data, w.err = data, err
		close(w.done)
		if prefetch {
			c.prefetching--
		}
		c.prefetchLocked(f, client)
	}()
	return w
}

// wantedLocked reports whether some reader's window — its current part and
// the read-ahead past it, within its range — covers part idx. The caller
// holds c.mu.
func (c *abyssPartCache) wantedLocked(idx int64) bool {
	for _, w := range c.readers {
		if idx >= w.pos && idx <= min(w.pos+abyssReadAhead, w.last) {
			return true
		}
	}
	return false
}

// dropStaleLocked cancels the in-flight fetches no reader's window covers any
// more — what a seek leaves behind — so they stop competing with the new
// position for bandwidth. It runs only when a reader opens: a seek always
// arrives as a new connection, while a reader moving on is just sequential
// reading. Hanging up does not trigger it either, so a client retrying at the
// same spot — mpv reopening, a downloader range that timed out — finds its
// fetches still going, even while other ranges of the file keep advancing.
// The caller holds c.mu.
func (c *abyssPartCache) dropStaleLocked() {
	for idx, w := range c.inflight {
		if !c.wantedLocked(idx) {
			delete(c.inflight, idx)
			w.cancel()
		}
	}
}

// prefetchLocked starts fetches for the nearest wanted parts past the readers
// that hold their current part, up to abyssPrefetchers at a time for the
// file. The caller holds c.mu.
func (c *abyssPartCache) prefetchLocked(f *abyssFile, client *http.Client) {
	for c.prefetching < abyssPrefetchers {
		idx, ok := c.nextPrefetchLocked()
		if !ok {
			return
		}
		c.fetchLocked(f, client, idx, true)
	}
}

// nextPrefetchLocked picks the nearest part past a ready reader that is not
// cached, in flight, or a failed prefetch. Checking the cache also refreshes
// the window's parts in the LRU, so what a reader is about to play is the
// last thing evicted. The caller holds c.mu.
func (c *abyssPartCache) nextPrefetchLocked() (int64, bool) {
	for d := int64(1); d <= abyssReadAhead; d++ {
		for _, w := range c.readers {
			idx := w.pos + d
			if !w.ready || idx > w.last {
				continue
			}
			if _, ok := c.inflight[idx]; ok {
				continue
			}
			if _, ok := c.failed[idx]; ok {
				continue
			}
			if _, ok := abyssParts.get(abyssPartKey{c, idx}); ok {
				continue
			}
			return idx, true
		}
	}
	return 0, false
}

// abyssPartReader is one sequential read of a byte range: an mpv connection,
// a downloader range, a probe. The handle itself never changes; where the
// read is lives in its abyssReadWindow, inside the cache it is guarded by.
type abyssPartReader struct {
	f      *abyssFile
	client *http.Client
	c      *abyssPartCache
}

// abyssReadWindow is where one reader is, in parts. Guarded by the cache's mu.
type abyssReadWindow struct {
	pos   int64 // part being read
	last  int64 // part holding the range's end: nothing past it is prefetched
	ready bool  // pos is in hand, so prefetching past it may start
}

// openPartReader registers a reader over [start, end], a range the caller
// has already checked (checkChunkRange). Opening one is where a seek shows
// up, so it drops the fetches nobody wants any more.
func (f *abyssFile) openPartReader(client *http.Client, start, end int64) *abyssPartReader {
	part := f.abyssPartSize()
	c := f.partsCache()
	rd := &abyssPartReader{f: f, client: client, c: c}
	c.mu.Lock()
	c.readers[rd] = &abyssReadWindow{pos: start / part, last: end / part}
	c.dropStaleLocked()
	c.mu.Unlock()
	return rd
}

// checkChunkRange rejects a byte range outside a chunked file.
func (f *abyssFile) checkChunkRange(start, end int64) error {
	if f.chunk == nil {
		return errors.New("not a chunked Abyss file")
	}
	if f.abyssPartSize() <= 0 || start < 0 || end < start || end >= f.size {
		return fmt.Errorf("invalid Abyss chunk range %d-%d", start, end)
	}
	return nil
}

// close unregisters the reader. Its in-flight fetches keep going; see
// dropStaleLocked.
func (rd *abyssPartReader) close() {
	rd.c.mu.Lock()
	delete(rd.c.readers, rd)
	rd.c.mu.Unlock()
}

// part returns part idx, moving the reader there first. Prefetching past it
// waits until the part is in hand, so after a seek the part mpv is blocked on
// gets the bandwidth before the read-ahead does. Moving on is sequential
// reading, not a seek, so it cancels nothing (see dropStaleLocked).
func (rd *abyssPartReader) part(ctx context.Context, idx int64) ([]byte, error) {
	c := rd.c
	c.mu.Lock()
	w := c.readers[rd]
	w.pos, w.ready = idx, false
	c.mu.Unlock()

	data, err := rd.f.abyssGetPart(ctx, rd.client, idx)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	w.ready = true
	c.prefetchLocked(rd.f, rd.client)
	c.mu.Unlock()
	return data, nil
}

// copyRange writes [start, end] — the range the reader was opened over — to
// dst, part by part.
func (rd *abyssPartReader) copyRange(ctx context.Context, start, end int64, dst io.Writer) error {
	if err := rd.f.checkChunkRange(start, end); err != nil {
		return err
	}
	for pos := start; pos <= end; {
		idx := pos / rd.f.abyssPartSize()
		lo, _ := rd.f.abyssPartBounds(idx)
		data, err := rd.part(ctx, idx)
		if err != nil {
			return fmt.Errorf("abyss part %d: %w", idx, err)
		}
		from := pos - lo
		to := min(end, lo+int64(len(data))-1) - lo
		if from < 0 || to < from {
			return fmt.Errorf("invalid Abyss part slice %d-%d", from, to)
		}
		if _, err := dst.Write(data[from : to+1]); err != nil {
			return err
		}
		pos = lo + to + 1
	}
	return nil
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

	urls, err := f.abyssPartURLs()
	if err != nil {
		return nil, err
	}
	if idx >= int64(len(urls)) {
		return nil, fmt.Errorf("part %d out of range", idx)
	}
	req := urls[idx]
	var from, through int64 = -1, -1 // absolute .fd range; -1 sends no Range header
	var seed string
	if c.firstSize > 0 && lo < c.firstSize {
		from, through, seed = lo, min(hi, c.firstSize-1), c.firstSeed
	}

	started := time.Now()
	var body []byte
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
	if diag, ok := errors.AsType[*netx.SourceDiagnostic](err); ok {
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
	if ue, ok := errors.AsType[*url.Error](err); ok {
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

// abyssPartURLs returns the upstream URL of every part of the file, built once
// from what resolution fixed: the .fd URL for the parts inside the encrypted
// head, a /sora/ token URL for each part after it. A proxied read only selects
// an entry by index, so nothing a client sends — a Range header, a path —
// goes into building an upstream URL: it can pick a part of this file from
// this file's hosts, and nothing else.
func (f *abyssFile) abyssPartURLs() ([]string, error) {
	f.urlsOnce.Do(func() {
		c, part := f.chunk, f.abyssPartSize()
		if c == nil || part <= 0 {
			f.urlsErr = errors.New("not a chunked Abyss file")
			return
		}
		n := (f.size + part - 1) / part
		urls := make([]string, 0, n)
		sora := "https://" + c.chunkHost + "/sora/" + strconv.FormatInt(f.size, 10) + "/"
		for idx := range n {
			if c.firstSize > 0 && idx*part < c.firstSize {
				urls = append(urls, c.firstURL)
				continue
			}
			token := abyssChunkToken(c.partTokenPath(part, idx), f.size)
			if token == "" {
				f.urlsErr = errors.New("could not build the Abyss chunk token")
				return
			}
			urls = append(urls, sora+token)
		}
		f.urls = urls
	})
	return f.urls, f.urlsErr
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
