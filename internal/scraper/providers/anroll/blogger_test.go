package anroll

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
)

type statusTransport struct {
	status int
	calls  int
}

func (s *statusTransport) RoundTrip(*http.Request) (*http.Response, error) {
	s.calls++
	return &http.Response{StatusCode: s.status, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}}, nil
}

func TestCheckBloggerLink(t *testing.T) {
	const locked = "https://redirector.googlevideo.com/videoplayback?expire=1&ip=172.70.100.77&source=blogger&itag=22"
	tests := []struct {
		name      string
		url       string
		status    int
		wantErr   bool
		wantProbe bool
	}{
		{"refused blogger link", locked, http.StatusForbidden, true, true},
		{"playable blogger link", locked, http.StatusPartialContent, false, true},
		{"not blogger", "https://redirector.googlevideo.com/videoplayback?ip=1.2.3.4&source=youtube", http.StatusForbidden, false, false},
		{"not googlevideo", "https://static.anidrive.click/master.m3u8", http.StatusForbidden, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := &statusTransport{status: tt.status}
			c := NewClientForTest(&http.Client{Transport: rt}, "https://anroll.test")
			c.probeBlogger = true
			err := c.checkBloggerLink(context.Background(), tt.url)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if (rt.calls > 0) != tt.wantProbe {
				t.Fatalf("probe calls = %d, wantProbe %v", rt.calls, tt.wantProbe)
			}
			if err != nil {
				var d *netx.SourceDiagnostic
				if !errors.As(err, &d) || d.Kind != netx.DiagnosticDownloadExpired || !strings.Contains(err.Error(), "Blogger") {
					t.Fatalf("unexpected diagnostic: %v", err)
				}
			}
		})
	}
}
