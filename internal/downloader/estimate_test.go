package downloader

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	estimateHLS     = int64(500 * 1024 * 1024)
	estimateDefault = int64(300 * 1024 * 1024)
)

// rangeServer mocks a media host: it records the Range header it was sent and
// answers with the given Content-Range (omitted when empty).
func rangeServer(t *testing.T, contentRange string, seen *string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if seen != nil {
			*seen = r.Header.Get("Range")
		}
		if contentRange != "" {
			w.Header().Set("Content-Range", contentRange)
		}
		w.WriteHeader(http.StatusPartialContent)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// estimateCases is the behaviour table of estimateStreamContentLength against a
// mock host; the metatest reuses it to judge broken estimators.
var estimateCases = []struct {
	name         string
	contentRange string
	want         int64
}{
	{"exact size from Content-Range", "bytes 0-4095/12345678", 12345678},
	{"unknown total", "bytes 0-4095/*", estimateDefault},
	{"no header", "", estimateDefault},
	{"total is not a number", "bytes 0-4095/lots", estimateDefault},
	{"no slash", "bytes 0-4095", estimateDefault},
	{"two slashes", "bytes 0-4095/12/34", estimateDefault},
	{"huge file", "bytes 0-4095/9000000000", 9000000000},
	{"zero total is taken at face value", "bytes 0-0/0", 0},
}

func estimateFailures(impl func(url string, c *http.Client) (int64, error)) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprintf("panic: %v", r))
		}
	}()
	for _, tc := range estimateCases {
		var seen string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seen = r.Header.Get("Range")
			if tc.contentRange != "" {
				w.Header().Set("Content-Range", tc.contentRange)
			}
			w.WriteHeader(http.StatusPartialContent)
		}))
		got, err := impl(srv.URL+"/v.mp4", &http.Client{Timeout: 2 * time.Second})
		srv.Close()
		if err != nil || got != tc.want {
			f = append(f, fmt.Sprintf("%s: got %d,%v want %d", tc.name, got, err, tc.want))
		}
		if seen != "bytes=0-4095" {
			f = append(f, fmt.Sprintf("%s: Range header %q", tc.name, seen))
		}
	}
	for _, u := range []string{"https://x/y.m3u8", "https://x/y.m3u8?token=1"} {
		if got, err := impl(u, &http.Client{Timeout: time.Second}); err != nil || got != estimateHLS {
			f = append(f, fmt.Sprintf("hls %s: got %d,%v", u, got, err))
		}
	}
	if got, err := impl("http://127.0.0.1:1/v.mp4", &http.Client{Timeout: time.Second}); err != nil || got != estimateDefault {
		f = append(f, fmt.Sprintf("unreachable host: got %d,%v", got, err))
	}
	if _, err := impl("http://bad host/\x7f", &http.Client{Timeout: time.Second}); err == nil {
		f = append(f, "malformed URL must error")
	}
	return f
}

func TestEstimateStreamContentLengthAgainstMockHost(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	for _, tc := range estimateCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var seen string
			srv := rangeServer(t, tc.contentRange, &seen)
			got, err := d.estimateStreamContentLength(srv.URL+"/v.mp4", &http.Client{Timeout: 2 * time.Second})
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, "bytes=0-4095", seen, "asks for the first 4 KiB only")
		})
	}
}

func TestEstimateStreamContentLengthHLSNeverTouchesTheNetwork(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hits++ }))
	t.Cleanup(srv.Close)
	got, err := d.estimateStreamContentLength(srv.URL+"/master.m3u8", srv.Client())
	require.NoError(t, err)
	assert.Equal(t, estimateHLS, got)
	assert.Zero(t, hits)
}

func TestEstimateStreamContentLengthMalformedURL(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	n, err := d.estimateStreamContentLength("http://bad host/\x7f", &http.Client{})
	assert.Zero(t, n, "no size is claimed for a request that never happened")
	assert.Error(t, err, "a URL that cannot even form a request is an error, not a guess")
}

func TestEstimateStreamContentLengthMeta(t *testing.T) {
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	reference := d.estimateStreamContentLength
	require.Empty(t, estimateFailures(reference), "the real estimator satisfies the contract")

	mutants := map[string]func(string, *http.Client) (int64, error){
		"always default": func(string, *http.Client) (int64, error) { return estimateDefault, nil },
		"ignores Content-Range": func(u string, c *http.Client) (int64, error) {
			n, err := reference(u, c)
			if n != estimateHLS && n != 0 {
				return estimateDefault, err
			}
			return n, err
		},
		"hls gets the default": func(u string, c *http.Client) (int64, error) {
			if len(u) > 5 && u[len(u)-5:] == ".m3u8" {
				return estimateDefault, nil
			}
			return reference(u, c)
		},
		"no Range header sent": func(u string, c *http.Client) (int64, error) {
			if strings.Contains(u, ".m3u8") {
				return estimateHLS, nil
			}
			req, err := http.NewRequest("GET", u, http.NoBody)
			if err != nil {
				return 0, err
			}
			resp, err := c.Do(req)
			if err != nil {
				return estimateDefault, nil
			}
			defer func() { _ = resp.Body.Close() }()
			parts := strings.Split(resp.Header.Get("Content-Range"), "/")
			if len(parts) == 2 && parts[1] != "*" {
				if n, perr := strconv.ParseInt(parts[1], 10, 64); perr == nil {
					return n, nil
				}
			}
			return estimateDefault, nil
		},
		"unreachable host errors": func(u string, c *http.Client) (int64, error) {
			n, err := reference(u, c)
			if n == estimateDefault && u == "http://127.0.0.1:1/v.mp4" {
				return 0, fmt.Errorf("dial failed")
			}
			return n, err
		},
		"malformed URL swallowed": func(u string, c *http.Client) (int64, error) {
			n, err := reference(u, c)
			if err != nil {
				return estimateDefault, nil //nolint:nilerr // deliberate mutant: swallows the error
			}
			return n, err
		},
		"unknown total parsed as zero": func(u string, c *http.Client) (int64, error) {
			n, err := reference(u, c)
			if n == estimateDefault {
				return 0, err
			}
			return n, err
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, estimateFailures(m), "mutant %q survived", name)
	}
}
