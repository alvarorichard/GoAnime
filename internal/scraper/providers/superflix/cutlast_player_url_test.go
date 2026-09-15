package superflix

import (
	"strings"
	"testing"
)

// legacySplitPlayerURL is the pre-Go-1.27 implementation that lived inline in
// ResolveRedirect. It is the oracle for the differential test.
func legacySplitPlayerURL(finalURL string) (baseURL, videoHash string) {
	if strings.Contains(finalURL, "/video/") {
		parts := strings.SplitN(finalURL, "/video/", 2)
		baseURL = parts[0]
		videoHash = strings.SplitN(parts[1], "?", 2)[0]
		videoHash = strings.SplitN(videoHash, "#", 2)[0]
	} else {
		idx := strings.LastIndex(finalURL, "/")
		if idx > 0 {
			baseURL = finalURL[:idx]
			videoHash = strings.SplitN(finalURL[idx+1:], "?", 2)[0]
		}
	}
	return baseURL, videoHash
}

func TestSplitPlayerURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantBase string
		wantHash string
	}{
		{"canonical video path", "https://player.example.com/video/abc123", "https://player.example.com", "abc123"},
		{"query is dropped", "https://player.example.com/video/abc123?t=10", "https://player.example.com", "abc123"},
		{"fragment is dropped", "https://player.example.com/video/abc123#start", "https://player.example.com", "abc123"},
		{"query and fragment", "https://player.example.com/video/abc123?t=1#x", "https://player.example.com", "abc123"},
		{"nested path before video", "https://player.example.com/e/video/abc123", "https://player.example.com/e", "abc123"},
		{"fallback to last segment", "https://player.example.com/embed/abc123", "https://player.example.com/embed", "abc123"},
		{"fallback drops query", "https://player.example.com/embed/abc123?x=1", "https://player.example.com/embed", "abc123"},
		{"no slash yields nothing", "abc123", "", ""},
		{"empty input", "", "", ""},
		{"leading slash only", "/abc123", "", ""},
		{"trailing slash yields empty hash", "https://player.example.com/embed/", "https://player.example.com/embed", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, hash := splitPlayerURL(tt.input)
			if base != tt.wantBase || hash != tt.wantHash {
				t.Errorf("splitPlayerURL(%q) = (%q, %q), want (%q, %q)",
					tt.input, base, hash, tt.wantBase, tt.wantHash)
			}
		})
	}
}

func TestSplitPlayerURLMatchesLegacyImplementation(t *testing.T) {
	corpus := []string{
		"", "/", "//", "abc", "/abc", "abc/", "a/b",
		"https://player.example.com/video/abc123",
		"https://player.example.com/video/abc123?t=10#x",
		"https://player.example.com/video/",
		"https://player.example.com/video/a/b",
		"https://player.example.com/embed/abc123",
		"https://player.example.com/embed/abc123?x=1",
		"https://player.example.com/embed/",
		"https://player.example.com",
		"https://player.example.com/",
		"/video/abc",
		"video/abc",
	}
	for _, in := range corpus {
		gotBase, gotHash := splitPlayerURL(in)
		wantBase, wantHash := legacySplitPlayerURL(in)
		if gotBase != wantBase || gotHash != wantHash {
			t.Errorf("input %q: CutLast version = (%q,%q), legacy = (%q,%q)",
				in, gotBase, gotHash, wantBase, wantHash)
		}
	}
}
