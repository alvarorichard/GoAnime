package startflix

import (
	"net/http"
	"testing"
)

// TestStreamProxyKeepsEveryFileOfABatch pins that a URL the proxy hands out
// stays valid while a batch download works through its episodes. The batch
// downloader resolves every episode's stream before it starts downloading
// any, and each Abyss episode registers one file; the registry used to keep
// only the last 64, first in first out, so in a longer batch the first
// episodes' URLs were dead (404) by the time their download began.
func TestStreamProxyKeepsEveryFileOfABatch(t *testing.T) {
	t.Parallel()
	p := newStreamProxy()
	const batch = 150
	urls := make([]string, 0, batch)
	for range batch {
		u, err := p.serve(&abyssFile{upstream: "https://files.abyss.test/v.mp4", size: 1 << 20}, http.DefaultClient)
		if err != nil {
			t.Fatal(err)
		}
		urls = append(urls, u)
	}
	for _, i := range []int{0, 1, batch / 2, batch - 1} {
		req, _ := http.NewRequest(http.MethodHead, urls[i], http.NoBody)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("episode %d of %d: HTTP %d, want its file still served", i+1, batch, resp.StatusCode)
		}
	}
}

func headStatus(t *testing.T, u string) int {
	t.Helper()
	req, _ := http.NewRequest(http.MethodHead, u, http.NoBody)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	return resp.StatusCode
}

// TestStreamProxyEvictsLeastRecentlyUsed pins the eviction order once the
// registry is full: the file read least recently goes, so one being played
// or downloaded survives however many episodes are registered after it.
func TestStreamProxyEvictsLeastRecentlyUsed(t *testing.T) {
	t.Parallel()
	p := newStreamProxy()
	register := func() string {
		u, err := p.serve(&abyssFile{upstream: "https://files.abyss.test/v.mp4", size: 1 << 20}, http.DefaultClient)
		if err != nil {
			t.Fatal(err)
		}
		return u
	}
	playing, idle := register(), register()
	for range maxProxiedFiles - 2 {
		register()
	}
	if headStatus(t, playing) != http.StatusOK { // reading it makes it recent
		t.Fatal("a file within the cap is not served")
	}
	register() // one past the cap: the least recently used goes
	if got := headStatus(t, playing); got != http.StatusOK {
		t.Errorf("the file being read was evicted: HTTP %d", got)
	}
	if got := headStatus(t, idle); got != http.StatusNotFound {
		t.Errorf("the least recently used file is still served: HTTP %d", got)
	}
	if n := p.lru.Len(); n != maxProxiedFiles {
		t.Errorf("registry holds %d files, want the cap %d", n, maxProxiedFiles)
	}
}
