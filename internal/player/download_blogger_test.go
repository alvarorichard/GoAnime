package player

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	charmlog "charm.land/log/v2"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rangeHost mocks the googlevideo CDN: it serves one body, honours Range, and
// can drop connections, ignore Range or fail, per request.
type rangeHost struct {
	body []byte

	mu          sync.Mutex
	ranges      []string // Range header of each GET, "" when absent
	heads       int
	gets        int
	dropFirstAt int    // bytes written before the first GET's connection is cut (0: never)
	dropEvery   int    // every GET writes at most this many bytes, then the connection is cut (0: never)
	ignoreRange bool   // answer 200 with the whole body
	failStatus  int    // answer every GET with this status
	failFirst   int    // the first n GETs answer 403
	headLength  *int   // Content-Length to claim on HEAD (default len(body))
	zeroBytes   bool   // every GET declares its payload and sends none
	failRange   [2]int // GETs numbered from..to (1-based, inclusive) answer 500
}

func (h *rangeHost) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.mu.Lock()
		n := len(h.body)
		if h.headLength != nil {
			n = *h.headLength
		}
		if r.Method == http.MethodHead {
			h.heads++
			h.mu.Unlock()
			w.Header().Set("Content-Length", strconv.Itoa(n))
			w.WriteHeader(http.StatusOK)
			return
		}
		h.gets++
		call := h.gets
		h.ranges = append(h.ranges, r.Header.Get("Range"))
		failStatus, failFirst, ignore, dropFirst, dropEvery := h.failStatus, h.failFirst, h.ignoreRange, h.dropFirstAt, h.dropEvery
		zero, fr := h.zeroBytes, h.failRange
		h.mu.Unlock()

		if fr[0] > 0 && call >= fr[0] && call <= fr[1] {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if failStatus != 0 {
			w.WriteHeader(failStatus)
			return
		}
		if call <= failFirst {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		from, to := 0, len(h.body)-1
		status := http.StatusOK
		if rng := r.Header.Get("Range"); rng != "" && !ignore {
			if _, err := fmt.Sscanf(rng, "bytes=%d-%d", &from, &to); err != nil {
				w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
				return
			}
			status = http.StatusPartialContent
		}
		to = min(to, len(h.body)-1)
		payload := h.body[from : to+1]

		w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
		w.WriteHeader(status)
		cut := len(payload)
		if call == 1 && dropFirst > 0 {
			cut = min(cut, dropFirst)
		}
		if dropEvery > 0 {
			cut = min(cut, dropEvery)
		}
		if zero {
			cut = 0
		}
		_, _ = w.Write(payload[:cut])
		// Returning with fewer bytes than declared makes net/http cut the
		// connection: the client sees an unexpected EOF mid-body.
	})
}

func content(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte((i*31 + 7) % 251)
	}
	return b
}

// bloggerHarness installs the seams for one test and records the sleeps.
type bloggerHarness struct {
	host *rangeHost
	srv  *httptest.Server

	mu     sync.Mutex // parts download in parallel and all report here
	sleeps []time.Duration
	guard  []string
}

func newHarness(t *testing.T, host *rangeHost) *bloggerHarness {
	t.Helper()
	h := &bloggerHarness{host: host}
	h.srv = httptest.NewServer(host.handler())
	t.Cleanup(h.srv.Close)

	oldGuard, oldClient, oldSleep, oldLogger := validateURLFn, newBloggerClientFn, chunkRetrySleep, util.Logger
	t.Cleanup(func() {
		validateURLFn, newBloggerClientFn, chunkRetrySleep, util.Logger = oldGuard, oldClient, oldSleep, oldLogger
	})
	if util.Logger == nil { // the program sets it up at start; a test must not crash on an error log
		util.Logger = charmlog.New(io.Discard)
	}
	validateURLFn = func(u string) error {
		h.mu.Lock()
		defer h.mu.Unlock()
		h.guard = append(h.guard, u)
		return nil
	}
	newBloggerClientFn = func() *http.Client { return &http.Client{Timeout: 5 * time.Second} }
	chunkRetrySleep = func(d time.Duration) {
		h.mu.Lock()
		defer h.mu.Unlock()
		h.sleeps = append(h.sleeps, d)
	}
	return h
}

func (h *bloggerHarness) url() string { return h.srv.URL + "/video.mp4" }

func readFile(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	require.NoError(t, err)
	return b
}

// ---- downloadBloggerChunk -----------------------------------------------------------

func TestDownloadBloggerChunkWritesExactlyItsRange(t *testing.T) {
	body := content(1000)
	h := newHarness(t, &rangeHost{body: body})
	dest := filepath.Join(t.TempDir(), "ep.mp4")
	m := &model{}

	require.NoError(t, downloadBloggerChunk(h.url(), 100, 299, 2, dest, m))

	part, _ := safePartPath(dest, 2)
	assert.Equal(t, body[100:300], readFile(t, part))
	assert.Equal(t, []string{"bytes=100-299"}, h.host.ranges)
	assert.Equal(t, int64(200), m.received, "progress counts the bytes of this part")
	assert.Equal(t, []string{h.url()}, h.guard, "the guard saw the URL before the request")
	assert.Empty(t, h.sleeps)
}

func TestDownloadBloggerChunkResumesAfterADroppedConnection(t *testing.T) {
	body := content(1000)
	h := newHarness(t, &rangeHost{body: body, dropFirstAt: 50})
	dest := filepath.Join(t.TempDir(), "ep.mp4")
	m := &model{}

	require.NoError(t, downloadBloggerChunk(h.url(), 0, 199, 0, dest, m))

	part, _ := safePartPath(dest, 0)
	assert.Equal(t, body[:200], readFile(t, part), "no byte lost or repeated across the drop")
	assert.Equal(t, []string{"bytes=0-199", "bytes=50-199"}, h.host.ranges, "the second request resumes where the first stopped")
	assert.Equal(t, []time.Duration{500 * time.Millisecond}, h.sleeps)
	assert.Equal(t, int64(200), m.received)
}

func TestDownloadBloggerChunkProgressResetsTheStaleCounter(t *testing.T) {
	// 300 bytes arrive 10 at a time: 30 requests, more than the 20 stale
	// retries allowed, because every one of them makes progress.
	body := content(300)
	h := newHarness(t, &rangeHost{body: body, dropEvery: 10})
	dest := filepath.Join(t.TempDir(), "ep.mp4")

	require.NoError(t, downloadBloggerChunk(h.url(), 0, 299, 0, dest, &model{}))

	part, _ := safePartPath(dest, 0)
	assert.Equal(t, body, readFile(t, part))
	assert.Equal(t, 30, h.host.gets)
}

func TestDownloadBloggerChunkGivesUpAfterTwentyStaleRetries(t *testing.T) {
	h := newHarness(t, &rangeHost{body: content(100), failStatus: http.StatusInternalServerError})
	dest := filepath.Join(t.TempDir(), "ep.mp4")

	err := downloadBloggerChunk(h.url(), 0, 99, 0, dest, &model{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "max retries (20)")
	assert.Contains(t, err.Error(), "chunk 0")
	assert.Equal(t, 20, h.host.gets, "exactly 20 attempts, no more")
	assert.Len(t, h.sleeps, 19, "a pause before every retry, none before the first try")
}

func TestDownloadBloggerChunkRetriesAfterAnErrorStatus(t *testing.T) {
	body := content(200)
	h := newHarness(t, &rangeHost{body: body, failFirst: 2})
	dest := filepath.Join(t.TempDir(), "ep.mp4")

	require.NoError(t, downloadBloggerChunk(h.url(), 0, 199, 0, dest, &model{}))
	part, _ := safePartPath(dest, 0)
	assert.Equal(t, body, readFile(t, part))
	assert.Equal(t, 3, h.host.gets)
}

func TestDownloadBloggerChunkWhenTheServerIgnoresRange(t *testing.T) {
	body := content(1000)
	for _, tc := range []struct{ from, to int64 }{{0, 99}, {100, 299}, {900, 999}} {
		t.Run(fmt.Sprintf("%d-%d", tc.from, tc.to), func(t *testing.T) {
			h := newHarness(t, &rangeHost{body: body, ignoreRange: true})
			dest := filepath.Join(t.TempDir(), "ep.mp4")
			m := &model{}

			require.NoError(t, downloadBloggerChunk(h.url(), tc.from, tc.to, 1, dest, m))

			part, _ := safePartPath(dest, 1)
			assert.Equal(t, body[tc.from:tc.to+1], readFile(t, part), "a 200 answer is trimmed to this part's span")
			assert.Equal(t, tc.to-tc.from+1, m.received)
		})
	}
}

func TestDownloadBloggerChunkRefusesBlockedURLsBeforeAnyRequest(t *testing.T) {
	h := newHarness(t, &rangeHost{body: content(10)})
	validateURLFn = func(string) error { return errors.New("URL resolves to disallowed IP 10.0.0.1") }

	err := downloadBloggerChunk(h.url(), 0, 9, 0, filepath.Join(t.TempDir(), "ep.mp4"), &model{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SSRF blocked")
	assert.Zero(t, h.host.gets)
}

func TestDownloadBloggerChunkRejectsAPartPathOutsideTheDirectory(t *testing.T) {
	h := newHarness(t, &rangeHost{body: content(10)})
	dir := t.TempDir()
	// A destination that is a directory cannot hold a part file next to it.
	err := downloadBloggerChunk(h.url(), 0, 9, 0, filepath.Join(dir, "missing-dir", "ep.mp4"), &model{})
	require.Error(t, err)
	assert.Zero(t, h.host.gets, "nothing is fetched when the part file cannot be created")
}

// chunkScenarioFailures runs the chunk downloader over three scenarios and
// returns what is wrong; the metatest feeds it corrupting wrappers.
func chunkScenarioFailures(t *testing.T, impl func(url string, from, to int64, part int, dest string, m *model) error) []string {
	t.Helper()
	var f []string
	body := content(600)
	scenarios := []struct {
		name string
		host *rangeHost
		from int64
		to   int64
	}{
		{"plain", &rangeHost{body: body}, 100, 399},
		{"drop", &rangeHost{body: body, dropFirstAt: 40}, 0, 299},
		{"ignores range", &rangeHost{body: body, ignoreRange: true}, 200, 449},
	}
	for _, sc := range scenarios {
		h := newHarness(t, sc.host)
		dest := filepath.Join(t.TempDir(), "ep.mp4")
		m := &model{}
		if err := impl(h.url(), sc.from, sc.to, 0, dest, m); err != nil {
			f = append(f, fmt.Sprintf("%s: error %v", sc.name, err))
			continue
		}
		part, _ := safePartPath(dest, 0)
		got, err := os.ReadFile(part)
		if err != nil || !bytes.Equal(got, body[sc.from:sc.to+1]) {
			f = append(f, fmt.Sprintf("%s: part file differs", sc.name))
		}
		if m.received != sc.to-sc.from+1 {
			f = append(f, fmt.Sprintf("%s: progress %d", sc.name, m.received))
		}
	}
	return f
}

func TestDownloadBloggerChunkMeta(t *testing.T) {
	require.Empty(t, chunkScenarioFailures(t, downloadBloggerChunk))
	type fn = func(string, int64, int64, int, string, *model) error
	mutants := map[string]fn{
		"truncates the part": func(u string, a, b int64, p int, d string, m *model) error {
			err := downloadBloggerChunk(u, a, b, p, d, m)
			part, _ := safePartPath(d, p)
			if st, e := os.Stat(part); e == nil && st.Size() > 0 {
				_ = os.Truncate(part, st.Size()-1)
			}
			return err
		},
		"appends a stray byte": func(u string, a, b int64, p int, d string, m *model) error {
			err := downloadBloggerChunk(u, a, b, p, d, m)
			part, _ := safePartPath(d, p)
			if f, e := os.OpenFile(part, os.O_APPEND|os.O_WRONLY, 0o600); e == nil {
				_, _ = f.Write([]byte{0})
				_ = f.Close()
			}
			return err
		},
		"swallows a failure": func(u string, a, b int64, p int, d string, m *model) error {
			_ = downloadBloggerChunk(u, a+1, b, p, d, m)
			return nil
		},
		"progress doubled": func(u string, a, b int64, p int, d string, m *model) error {
			err := downloadBloggerChunk(u, a, b, p, d, m)
			m.addProgressReceived(b - a + 1)
			return err
		},
		"wrong span": func(u string, a, b int64, p int, d string, m *model) error {
			return downloadBloggerChunk(u, a, b-1, p, d, m)
		},
		"wrong part file": func(u string, a, b int64, p int, d string, m *model) error {
			return downloadBloggerChunk(u, a, b, p+1, d, m)
		},
		"fails on the first drop": func(u string, a, b int64, p int, d string, m *model) error {
			err := downloadBloggerChunk(u, a, b, p, d, m)
			if err == nil && strings.Contains(u, "x") && false {
				return errors.New("unreachable")
			}
			return err
		},
	}
	delete(mutants, "fails on the first drop") // equivalent to the real function: kept out
	for name, m := range mutants {
		assert.NotEmpty(t, chunkScenarioFailures(t, m), "mutant %q survived", name)
	}
}

// ---- downloadBloggerDirect ----------------------------------------------------------

func TestDownloadBloggerDirectAssemblesTheParts(t *testing.T) {
	for _, threads := range []int{1, 2, 3, 4, 7} {
		t.Run(fmt.Sprintf("%d threads", threads), func(t *testing.T) {
			body := content(1003) // not divisible by the thread counts: the last part takes the rest
			h := newHarness(t, &rangeHost{body: body})
			dest := filepath.Join(t.TempDir(), "out", "ep.mp4")
			m := &model{}

			require.NoError(t, downloadBloggerDirect(h.url(), dest, threads, m))

			assert.Equal(t, body, readFile(t, dest))
			assert.Equal(t, int64(1003), m.totalBytes)
			assert.Equal(t, int64(1003), m.received)
			assert.Equal(t, threads, h.host.gets, "one request per part")
			left, _ := filepath.Glob(dest + ".part*")
			assert.Empty(t, left, "part files are removed once combined")
		})
	}
}

func TestDownloadBloggerDirectWithMoreThreadsThanBytes(t *testing.T) {
	body := content(5)
	h := newHarness(t, &rangeHost{body: body})
	dest := filepath.Join(t.TempDir(), "ep.mp4")

	require.NoError(t, downloadBloggerDirect(h.url(), dest, 16, &model{}))
	assert.Equal(t, body, readFile(t, dest))
	assert.Equal(t, 5, h.host.gets, "never more parts than bytes")
}

func TestDownloadBloggerDirectErrors(t *testing.T) {
	t.Run("blocked URL", func(t *testing.T) {
		h := newHarness(t, &rangeHost{body: content(10)})
		validateURLFn = func(string) error { return errors.New("disallowed") }
		err := downloadBloggerDirect(h.url(), filepath.Join(t.TempDir(), "e.mp4"), 2, &model{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SSRF blocked")
		assert.Zero(t, h.host.heads+h.host.gets)
	})
	t.Run("zero content length", func(t *testing.T) {
		zero := 0
		h := newHarness(t, &rangeHost{body: content(10), headLength: &zero})
		err := downloadBloggerDirect(h.url(), filepath.Join(t.TempDir(), "e.mp4"), 2, &model{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "content length is zero")
		assert.Zero(t, h.host.gets)
	})
	t.Run("a failing part fails the download", func(t *testing.T) {
		h := newHarness(t, &rangeHost{body: content(100), failStatus: http.StatusInternalServerError})
		dest := filepath.Join(t.TempDir(), "e.mp4")
		err := downloadBloggerDirect(h.url(), dest, 2, &model{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "chunk download failed")
		_, statErr := os.Stat(dest)
		assert.True(t, os.IsNotExist(statErr), "no half-assembled file is left as the result")
	})
	t.Run("the output directory cannot be created", func(t *testing.T) {
		h := newHarness(t, &rangeHost{body: content(10)})
		blocker := filepath.Join(t.TempDir(), "file")
		require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o600))
		err := downloadBloggerDirect(h.url(), filepath.Join(blocker, "sub", "e.mp4"), 2, &model{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create output directory")
		assert.Zero(t, h.host.heads+h.host.gets)
	})
	t.Run("the length request fails", func(t *testing.T) {
		h := newHarness(t, &rangeHost{body: content(10)})
		h.srv.Close()
		err := downloadBloggerDirect(h.url(), filepath.Join(t.TempDir(), "e.mp4"), 2, &model{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get content length")
	})
}

// ---- the retry accounting ------------------------------------------------------------

type failingTransport struct {
	mu    sync.Mutex
	calls int
}

func (f *failingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	f.mu.Lock()
	f.calls++
	f.mu.Unlock()
	return nil, errors.New("dial tcp: connection refused")
}

func TestDownloadBloggerChunkGivesUpWhenTheRequestNeverSucceeds(t *testing.T) {
	h := newHarness(t, &rangeHost{body: content(10)})
	ft := &failingTransport{}
	newBloggerClientFn = func() *http.Client { return &http.Client{Transport: ft} }

	err := downloadBloggerChunk(h.url(), 0, 9, 0, filepath.Join(t.TempDir(), "ep.mp4"), &model{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "max retries (20)")
	assert.Equal(t, 20, ft.calls, "a refused connection counts like any other failed attempt")
}

func TestDownloadBloggerChunkGivesUpWhenConnectionsDeliverNothing(t *testing.T) {
	h := newHarness(t, &rangeHost{body: content(100), zeroBytes: true})
	err := downloadBloggerChunk(h.url(), 0, 99, 0, filepath.Join(t.TempDir(), "ep.mp4"), &model{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max retries (20)")
	assert.Equal(t, 20, h.host.gets)
}

func TestDownloadBloggerChunkProgressGrantsAFreshBudgetOfRetries(t *testing.T) {
	// call 1 delivers 10 bytes, calls 2..20 fail (19 stale attempts), call 21 works.
	body := content(100)
	h := newHarness(t, &rangeHost{body: body, dropFirstAt: 10, failRange: [2]int{2, 20}})
	dest := filepath.Join(t.TempDir(), "ep.mp4")

	require.NoError(t, downloadBloggerChunk(h.url(), 0, 99, 0, dest, &model{}))
	part, _ := safePartPath(dest, 0)
	assert.Equal(t, body, readFile(t, part))
	assert.Equal(t, 21, h.host.gets, "progress zeroes the stale count, so 19 more failures are survivable")
}

func TestDownloadBloggerChunkLastBytesAreNotLost(t *testing.T) {
	body := content(200)
	h := newHarness(t, &rangeHost{body: body, dropFirstAt: 199}) // leaves exactly one byte
	dest := filepath.Join(t.TempDir(), "ep.mp4")

	require.NoError(t, downloadBloggerChunk(h.url(), 0, 199, 0, dest, &model{}))
	part, _ := safePartPath(dest, 0)
	assert.Equal(t, body, readFile(t, part), "the final byte is fetched, not skipped")
	assert.Equal(t, []string{"bytes=0-199", "bytes=199-199"}, h.host.ranges)
}

func TestDownloadBloggerChunkOneByteBeforeAnIgnoredRange(t *testing.T) {
	body := content(300)
	h := newHarness(t, &rangeHost{body: body, ignoreRange: true, dropFirstAt: 1})
	dest := filepath.Join(t.TempDir(), "ep.mp4")

	require.NoError(t, downloadBloggerChunk(h.url(), 0, 299, 0, dest, &model{}))
	part, _ := safePartPath(dest, 0)
	assert.Equal(t, body, readFile(t, part), "a resumed 200 skips the one byte already written")
}

func TestDownloadBloggerChunkIgnoredRangeShorterThanTheOffset(t *testing.T) {
	h := newHarness(t, &rangeHost{body: content(100), ignoreRange: true})
	err := downloadBloggerChunk(h.url(), 150, 199, 0, filepath.Join(t.TempDir(), "ep.mp4"), &model{})
	require.Error(t, err, "a body that ends before the part's offset can never deliver it")
	assert.Contains(t, err.Error(), "max retries (20)")
}

func TestDownloadBloggerChunkReadsLargePartsInOneRequest(t *testing.T) {
	body := content(300 * 1024) // several 64 KiB reads
	h := newHarness(t, &rangeHost{body: body})
	dest := filepath.Join(t.TempDir(), "ep.mp4")

	require.NoError(t, downloadBloggerChunk(h.url(), 0, int64(len(body)-1), 0, dest, &model{}))
	part, _ := safePartPath(dest, 0)
	assert.True(t, bytes.Equal(body, readFile(t, part)))
	assert.Equal(t, 1, h.host.gets, "a healthy connection is read to the end, not re-opened per read")
}
