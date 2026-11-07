package rangedl

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentLengthReadsTheHeadResponse(t *testing.T) {
	var methods []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		w.Header().Set("Content-Length", "1234")
	}))
	defer srv.Close()

	got, err := ContentLength(srv.URL, srv.Client(), ProbeOptions{})

	require.NoError(t, err)
	assert.Equal(t, int64(1234), got)
	assert.Equal(t, []string{"HEAD"}, methods, "one HEAD is enough")
}

func TestContentLengthFallsBackToARangedGetWhenHeadIsRefused(t *testing.T) {
	for _, status := range []int{http.StatusMethodNotAllowed, http.StatusNotImplemented} {
		var methods, ranges []string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			methods = append(methods, r.Method)
			ranges = append(ranges, r.Header.Get("Range"))
			if r.Method == http.MethodHead {
				w.WriteHeader(status)
				return
			}
			w.Header().Set("Content-Length", "77")
			w.WriteHeader(http.StatusPartialContent)
		}))
		got, err := ContentLength(srv.URL, srv.Client(), ProbeOptions{})
		srv.Close()

		require.NoError(t, err, "status %d", status)
		assert.Equal(t, int64(77), got)
		assert.Equal(t, []string{"HEAD", "GET"}, methods)
		assert.Equal(t, []string{"", "bytes=0-0"}, ranges, "the fallback asks for a single byte")
	}
}

func TestContentLengthFallsBackWhenTheHeadRequestFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			panic(http.ErrAbortHandler)
		}
		w.Header().Set("Content-Length", "9")
	}))
	defer srv.Close()

	got, err := ContentLength(srv.URL, srv.Client(), ProbeOptions{})

	require.NoError(t, err)
	assert.Equal(t, int64(9), got)
}

func TestContentLengthPreparesEveryRequest(t *testing.T) {
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Header.Get("Referer"))
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Length", "1")
	}))
	defer srv.Close()

	_, err := ContentLength(srv.URL, srv.Client(), ProbeOptions{Prepare: func(r *http.Request) { r.Header.Set("Referer", "https://r.example/") }})

	require.NoError(t, err)
	assert.Equal(t, []string{"https://r.example/", "https://r.example/"}, seen, "the HEAD and the fallback GET both carry the headers")
}

func TestContentLengthExpiredLinkIsTyped(t *testing.T) {
	for _, status := range []int{http.StatusForbidden, http.StatusNotFound} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(status) }))
		_, err := ContentLength(srv.URL, srv.Client(), ProbeOptions{})
		srv.Close()
		require.Error(t, err, "status %d", status)
		var diag *netx.SourceDiagnostic
		require.True(t, errors.As(err, &diag), "status %d: got %T", status, err)
		assert.Equal(t, netx.DiagnosticDownloadExpired, diag.Kind)
		assert.Equal(t, status, diag.StatusCode)
		assert.Equal(t, "content-length", diag.Layer)
	}
}

func TestContentLengthOtherStatusesAreErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusInternalServerError) }))
	defer srv.Close()
	_, err := ContentLength(srv.URL, srv.Client(), ProbeOptions{})
	require.Error(t, err)
	assert.ErrorContains(t, err, "status code 500")
}

func TestContentLengthAcceptsPartialContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "5")
		w.WriteHeader(http.StatusPartialContent)
	}))
	defer srv.Close()
	got, err := ContentLength(srv.URL, srv.Client(), ProbeOptions{})
	require.NoError(t, err)
	assert.Equal(t, int64(5), got)
}

func TestContentLengthMissingHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header()["Content-Length"] = nil // no Content-Length at all
	}))
	defer srv.Close()

	_, err := ContentLength(srv.URL, srv.Client(), ProbeOptions{})
	require.Error(t, err, "without OnMissing a missing length is an error")

	called := 0
	got, err := ContentLength(srv.URL, srv.Client(), ProbeOptions{OnMissing: func() (int64, error) { called++; return 42, nil }})
	require.NoError(t, err)
	assert.Equal(t, int64(42), got)
	assert.Equal(t, 1, called)

	boom := errors.New("boom")
	_, err = ContentLength(srv.URL, srv.Client(), ProbeOptions{OnMissing: func() (int64, error) { return 0, boom }})
	assert.ErrorIs(t, err, boom)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestContentLengthHeaderThatIsNotANumber(t *testing.T) {
	// net/http refuses such a header itself, so hand ContentLength the response
	// a lenient transport would produce.
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Length": []string{"12x"}},
			Body:       http.NoBody,
			Request:    r,
		}, nil
	})}

	got, err := ContentLength("http://example.invalid/v.mp4", client, ProbeOptions{})

	require.Error(t, err)
	assert.Equal(t, int64(0), got)
}

func TestContentLengthBadURL(t *testing.T) {
	_, err := ContentLength("http://[::1\n", http.DefaultClient, ProbeOptions{})
	require.Error(t, err)
}
