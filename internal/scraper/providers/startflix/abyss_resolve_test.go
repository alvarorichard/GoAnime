package startflix

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
)

// These tests drive resolveAbyss end to end against fake hosts: the embed
// page with its sealed datas, the media list inside it, the CDN probe, and the
// local proxy mpv and the downloader read from. The sealed fixture is built
// with the package's own functions, so it round-trips exactly what the
// player's page carries.

const (
	fakeEmbedHost   = "playembedapi.test"
	fakeEmbedOrigin = "https://" + fakeEmbedHost
)

// sealAbyssDatas builds the embed page's datas value for media: the media JSON
// AES-CTR sealed under md5hex("<user>:<slug>:<md5>"), carried as a Latin-1
// string inside base64 JSON.
func sealAbyssDatas(t *testing.T, slug string, md5ID, userID int64, media any) string {
	t.Helper()
	plain, err := json.Marshal(media)
	if err != nil {
		t.Fatal(err)
	}
	ct := bytes.Clone(plain)
	seed := strconv.FormatInt(userID, 10) + ":" + slug + ":" + strconv.FormatInt(md5ID, 10)
	if err := abyssXOR(seed, 0, ct); err != nil {
		t.Fatal(err)
	}
	inner, err := json.Marshal(map[string]any{
		"slug": slug, "md5_id": md5ID, "user_id": userID,
		"media": string(latin1ToUTF8(ct)),
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := utf8ToLatin1(string(inner))
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(raw)
}

func embedPage(datas string) string {
	return `<html><head><script>const datas = "` + datas + `";</script></head><body></body></html>`
}

// fakeAbyssFile is a single-file Abyss CDN: the file's first 64 KiB are
// sealed under its last path segment, it serves Range only to the embed's
// Origin, and it can be told to answer 503 a few times or 403 throughout.
type fakeAbyssFile struct {
	plain     []byte
	sealed    []byte
	path      string
	unavail   atomic.Int32 // next N requests get 503
	forbidden atomic.Bool
	hits      atomic.Int32
}

func newFakeAbyssFile(t *testing.T, size int, path string) *fakeAbyssFile {
	t.Helper()
	plain := make([]byte, size)
	copy(plain, "\x00\x00\x00\x20ftypisom\x00\x00\x02\x00")
	for i := 16; i < size; i++ {
		plain[i] = byte(i % 233)
	}
	sealed := bytes.Clone(plain)
	if err := abyssXOR(path[strings.LastIndex(path, "/")+1:], 0, sealed[:abyssEncryptedHead]); err != nil {
		t.Fatal(err)
	}
	return &fakeAbyssFile{plain: plain, sealed: sealed, path: path}
}

func (f *fakeAbyssFile) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.hits.Add(1)
	switch {
	case r.URL.Path != f.path:
		http.NotFound(w, r)
	case r.Header.Get("Origin") != fakeEmbedOrigin || r.Header.Get("Referer") != fakeEmbedOrigin+"/" || f.forbidden.Load():
		w.WriteHeader(http.StatusForbidden)
	case f.unavail.Add(-1) >= 0:
		w.WriteHeader(http.StatusServiceUnavailable)
	default:
		http.ServeContent(w, r, "v.mp4", time.Time{}, bytes.NewReader(f.sealed))
	}
}

// abyssSite routes the embed page and a CDN by the host the client asked for.
func abyssSite(datas string, cdn http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Original-Host") == fakeEmbedHost {
			_, _ = w.Write([]byte(embedPage(datas)))
			return
		}
		cdn.ServeHTTP(w, r)
	})
}

func singleFileMedia(size int, fileURL, path string) map[string]any {
	return map[string]any{"mp4": map[string]any{"sources": []map[string]any{
		{"label": "360p", "res_id": 2, "size": size, "codec": "h264", "url": fileURL, "path": path, "status": true},
		{"label": "720p", "res_id": 4, "size": size, "codec": "h264", "url": fileURL, "path": path, "status": true},
	}}}
}

func abyssPlayer() (Player, *url.URL) {
	u, _ := url.Parse(fakeEmbedOrigin + "/?v=slug1")
	return Player{URL: u.String(), Type: "iframe", referer: "https://painel.test/"}, u
}

func readLocal(t *testing.T, local, rng string) (int, []byte) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, local, http.NoBody)
	if rng != "" {
		req.Header.Set("Range", rng)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", rng, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, body
}

// TestResolveAbyssSingleFile runs the single-file layout end to end: the
// embed's datas open to a media list, the tallest rendition is probed with
// the embed's Origin and Referer, and the local proxy serves plain MP4 bytes
// — decrypted across the 64 KiB edge, for ranges and for a whole read.
func TestResolveAbyssSingleFile(t *testing.T) {
	t.Parallel()
	const size = 200_000
	cdn := newFakeAbyssFile(t, size, "/x/abc.mp4")
	datas := sealAbyssDatas(t, "slug1", 11, 22, singleFileMedia(size, "https://files.abyss.test", "x/abc.mp4"))
	c := newTestClient(t, abyssSite(datas, cdn))
	p, u := abyssPlayer()

	s, err := c.resolveAbyss(context.Background(), u, p)
	if err != nil {
		t.Fatal(err)
	}
	if s.Height != 720 || s.Host != fakeEmbedHost {
		t.Errorf("stream = %+v, want the 720p rendition from %s", s, fakeEmbedHost)
	}
	if !strings.HasPrefix(s.URL, "http://127.0.0.1:") {
		t.Fatalf("stream URL %q is not the local proxy", s.URL)
	}

	for _, span := range [][2]int{{0, 15}, {abyssEncryptedHead - 10, abyssEncryptedHead + 10}, {size - 100, size - 1}} {
		code, body := readLocal(t, s.URL, fmt.Sprintf("bytes=%d-%d", span[0], span[1]))
		if code != http.StatusPartialContent || !bytes.Equal(body, cdn.plain[span[0]:span[1]+1]) {
			t.Errorf("range %v: HTTP %d, %d bytes, plain=%v", span, code, len(body), bytes.Equal(body, cdn.plain[span[0]:span[1]+1]))
		}
	}
	code, body := readLocal(t, s.URL, "")
	if code != http.StatusOK || !bytes.Equal(body, cdn.plain) {
		t.Errorf("whole file: HTTP %d, %d bytes, want the plain file", code, len(body))
	}
}

// TestAbyssProxyUpstreamRetries pins the single-file proxy's failure contract:
// a transient 503 is retried before mpv sees anything, and a refusal that
// stays is an explicit 502, not a truncated 206.
func TestAbyssProxyUpstreamRetries(t *testing.T) {
	t.Parallel()
	const size = 100_000
	cdn := newFakeAbyssFile(t, size, "/x/retry.mp4")
	datas := sealAbyssDatas(t, "slug1", 11, 22, singleFileMedia(size, "https://files.abyss.test", "x/retry.mp4"))
	c := newTestClient(t, abyssSite(datas, cdn))
	p, u := abyssPlayer()
	s, err := c.resolveAbyss(context.Background(), u, p)
	if err != nil {
		t.Fatal(err)
	}

	cdn.unavail.Store(1)
	before := cdn.hits.Load()
	code, body := readLocal(t, s.URL, "bytes=10-20")
	if code != http.StatusPartialContent || !bytes.Equal(body, cdn.plain[10:21]) {
		t.Errorf("after one 503: HTTP %d, % x", code, body)
	}
	if got := cdn.hits.Load() - before; got != 2 {
		t.Errorf("upstream hits = %d, want 2 (one retry)", got)
	}

	cdn.forbidden.Store(true)
	before = cdn.hits.Load()
	if code, _ := readLocal(t, s.URL, "bytes=10-20"); code != http.StatusBadGateway {
		t.Errorf("forbidden upstream: HTTP %d, want 502", code)
	}
	if got := cdn.hits.Load() - before; got != 1 {
		t.Errorf("a 403 was retried: %d upstream hits", got)
	}
}

// TestResolveAbyssChunked runs the chunked layout through resolveAbyss: the
// media list names a sub-matched domain and an encrypted .fd head; the probe
// opens both the head and the first token part before the proxy URL is handed
// out, and reads through the proxy straddle the head/part split.
func TestResolveAbyssChunked(t *testing.T) {
	t.Parallel()
	const chunk = int64(2 << 20)
	const size = chunk + 300_000
	plain := make([]byte, size)
	copy(plain, "\x00\x00\x00\x20ftypisom\x00\x00\x02\x00")
	for i := 16; i < int(size); i++ {
		plain[i] = byte(i % 211)
	}
	up := &fakeChunkUpstream{
		plain: plain, fdSeed: "head.fd", fdSize: chunk,
		size: size, md5ID: 77, resID: 4, chunk: chunk, origin: fakeEmbedOrigin,
	}
	media := map[string]any{"mp4": map[string]any{
		"sources": []map[string]any{
			{"label": "720p", "res_id": 4, "size": size, "codec": "h264", "sub": "qq", "status": true},
		},
		"domains": []string{"zz.abyss.test", "qq.abyss.test"},
		"fristDatas": []map[string]any{
			{"res_id": 4, "size": size, "codec": "h264", "url": "https://qq.abyss.test/d/head.fd", "partSize": chunk},
		},
	}}
	datas := sealAbyssDatas(t, "slug1", 77, 22, media)
	c := newTestClient(t, abyssSite(datas, up))
	p, u := abyssPlayer()

	s, err := c.resolveAbyss(context.Background(), u, p)
	if err != nil {
		t.Fatal(err)
	}
	if s.Height != 720 {
		t.Errorf("height = %d, want 720", s.Height)
	}
	if up.hitsFor("fd") == 0 || up.hitsFor("sora:1") == 0 {
		t.Error("the probe did not open both the head and the first token part")
	}
	for _, span := range [][2]int64{{0, 15}, {chunk - 50, chunk + 50}, {size - 16, size - 1}} {
		code, body := readLocal(t, s.URL, fmt.Sprintf("bytes=%d-%d", span[0], span[1]))
		if code != http.StatusPartialContent || !bytes.Equal(body, plain[span[0]:span[1]+1]) {
			t.Errorf("range %v: HTTP %d, %d bytes", span, code, len(body))
		}
	}
}

// TestResolveAbyssFailures pins each way resolution gives up before a proxy
// URL exists, and that each says why.
func TestResolveAbyssFailures(t *testing.T) {
	t.Parallel()
	const size = 100_000
	good := singleFileMedia(size, "https://files.abyss.test", "x/abc.mp4")
	noRendition := map[string]any{"mp4": map[string]any{"sources": []map[string]any{
		{"label": "720p", "res_id": 4, "size": size, "codec": "h264"}, // neither a file nor chunked
	}}}
	tests := []struct {
		name   string
		page   string
		cdn    func(*fakeAbyssFile)
		want   string
		wantAs netx.DiagnosticKind
	}{
		{name: "page without datas", page: "<html>nothing here</html>", want: "unreadable embed page", wantAs: netx.DiagnosticParserBroken},
		{name: "media that does not open", page: embedPage(base64.StdEncoding.EncodeToString([]byte(`{"slug":"s","md5_id":1,"user_id":2,"media":"garbage"}`))),
			want: "could not open the media list", wantAs: netx.DiagnosticDecryptBroken},
		{name: "no supported rendition", page: embedPage(sealAbyssDatas(t, "s", 1, 2, noRendition)), want: "no supported MP4 rendition"},
		{name: "CDN refuses the probe", page: embedPage(sealAbyssDatas(t, "s", 1, 2, good)),
			cdn: func(f *fakeAbyssFile) { f.forbidden.Store(true) }, want: "403"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cdn := newFakeAbyssFile(t, size, "/x/abc.mp4")
			if tt.cdn != nil {
				tt.cdn(cdn)
			}
			page := tt.page
			c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-Original-Host") == fakeEmbedHost {
					_, _ = w.Write([]byte(page))
					return
				}
				cdn.ServeHTTP(w, r)
			}))
			p, u := abyssPlayer()
			s, err := c.resolveAbyss(context.Background(), u, p)
			if err == nil {
				t.Fatalf("resolved to %+v, want an error", s)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("err = %v, want it to mention %q", err, tt.want)
			}
			if tt.wantAs != "" {
				var diag *netx.SourceDiagnostic
				if !errors.As(err, &diag) || diag.Kind != tt.wantAs {
					t.Errorf("err = %v, want a %s diagnostic", err, tt.wantAs)
				}
			}
		})
	}
}

// TestProbeAbyssChunkedRejectsNonMP4 pins the chunked probe's sanity check:
// a head that decrypts to something other than an MP4 means the scheme
// changed, and resolution fails rather than handing mpv noise.
func TestProbeAbyssChunkedRejectsNonMP4(t *testing.T) {
	t.Parallel()
	const chunk = int64(2 << 20)
	const size = chunk + 1000
	plain := make([]byte, size) // all zeros: no "ftyp" at offset 4
	up := &fakeChunkUpstream{
		plain: plain, fdSeed: "head.fd", fdSize: chunk,
		size: size, md5ID: 99, resID: 2, chunk: chunk, origin: "https://playembedapi.site",
	}
	c := newTestClient(t, up)
	f := &abyssFile{
		size: size, origin: "https://playembedapi.site",
		chunk: &abyssChunkSource{
			md5ID: 99, resID: 2, chunkPath: "/mp4/99/2/" + strconv.FormatInt(size, 10),
			chunkHost: "qq.abyss.test", firstURL: "https://qq.abyss.test/d/head.fd", firstSize: chunk, firstSeed: "head.fd",
		},
	}
	if err := c.probeAbyss(context.Background(), f); err == nil || !strings.Contains(err.Error(), "not an MP4") {
		t.Fatalf("err = %v, want the non-MP4 head refused", err)
	}
}

// TestHeadDecrypterStreams pins the streaming decrypter the single-file proxy
// wraps responses in: reads of any size, starting before or after the 64 KiB
// edge, come out plain, and bytes past the head pass through untouched.
func TestHeadDecrypterStreams(t *testing.T) {
	t.Parallel()
	const seed = "abc.mp4"
	plain := make([]byte, abyssEncryptedHead+5000)
	for i := range plain {
		plain[i] = byte(i % 251)
	}
	sealed := bytes.Clone(plain)
	if err := abyssXOR(seed, 0, sealed[:abyssEncryptedHead]); err != nil {
		t.Fatal(err)
	}
	for _, start := range []int{0, 1, abyssEncryptedHead - 7, abyssEncryptedHead, abyssEncryptedHead + 100} {
		// iotest-style small reads exercise every boundary inside Read.
		r := &headDecrypter{r: &smallReader{b: sealed[start:], n: 333}, seed: seed, pos: int64(start)}
		got, err := io.ReadAll(r)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, plain[start:]) {
			t.Errorf("start %d: decrypted stream differs from the plain file", start)
		}
	}
}

// smallReader returns at most n bytes per Read.
type smallReader struct {
	b []byte
	n int
}

func (s *smallReader) Read(p []byte) (int, error) {
	if len(s.b) == 0 {
		return 0, io.EOF
	}
	k := min(len(p), s.n, len(s.b))
	copy(p, s.b[:k])
	s.b = s.b[k:]
	return k, nil
}

// TestClientStream runs a players page through Client.Stream: its buttons
// are parsed, every supported host is resolved, and the best stream comes
// back with the panel's subtitle attached; a page with no buttons, or the
// panel's "not found" answer, says so.
func TestClientStream(t *testing.T) {
	t.Parallel()
	pb := sealByse(t, map[string]any{
		"sources": []map[string]any{{"url": "https://cdn.test/fhd/master.m3u8", "height": 1080}},
	}, 6)
	var mu sync.Mutex
	var referers []string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch host := r.Header.Get("X-Original-Host"); {
		case host == "painel.test" && r.URL.Path == "/episodio/9":
			mu.Lock()
			referers = append(referers, r.Header.Get("Referer"))
			mu.Unlock()
			_, _ = w.Write([]byte(`<div id="players">
				<button data-show-player="true" data-source="https://files.test/ok.mp4" data-type="jwplayer" data-subtitles="" data-id="2">Player #2</button>
				<button data-show-player="true" data-source="https://embedplaybyse.test/e/abc12345/x" data-type="iframe" data-subtitles="https://painel.test/pt.vtt" data-id="1">Player #1</button>
				<button data-show-player="true" data-source="https://vidsrcme.su/embed/1" data-type="iframe" data-id="3">Player #3</button>
			</div>`))
		case host == "painel.test" && r.URL.Path == "/episodio/empty":
			_, _ = w.Write([]byte(`<div id="players"></div>`))
		case host == "painel.test" && r.URL.Path == "/filme/tt0000404":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`Conteúdo não encontrado`))
		default:
			fakeHosts(t, &pb, "").ServeHTTP(w, r)
		}
	}))
	ctx := context.Background()

	s, err := c.Stream(ctx, "https://painel.test/episodio/9")
	if err != nil {
		t.Fatal(err)
	}
	if s.URL != "https://cdn.test/fhd/master.m3u8" || s.Height != 1080 {
		t.Errorf("stream = %s (%dp), want the 1080p Byse stream over the unlabelled file", s.URL, s.Height)
	}
	if len(s.Subtitles) != 1 || s.Subtitles[0].URL != "https://painel.test/pt.vtt" {
		t.Errorf("subtitles = %+v, want the panel's track", s.Subtitles)
	}
	mu.Lock()
	if len(referers) != 1 || referers[0] != "https://painel.test/" {
		t.Errorf("episode endpoint referer = %v, want the panel origin", referers)
	}
	mu.Unlock()

	if _, err := c.Stream(ctx, "https://painel.test/episodio/empty"); !errors.Is(err, ErrNoPlayers) {
		t.Errorf("empty players page: err = %v, want ErrNoPlayers", err)
	}
	if _, err := c.Stream(ctx, "https://painel.test/filme/tt0000404"); !errors.Is(err, ErrNotOnPanel) {
		t.Errorf("panel not found: err = %v, want ErrNotOnPanel", err)
	}
	if _, err := c.Stream(ctx, "::not a url"); err == nil {
		t.Error("an invalid players URL was accepted")
	}
}

func TestSeasonAccessors(t *testing.T) {
	t.Parallel()
	s := Season{
		Number:    3,
		Dubbed:    []Episode{{Number: 1, Audio: AudioDubbed}},
		Subtitled: []Episode{{Number: 1, Audio: AudioSubtitled}, {Number: 2, Audio: AudioSubtitled}},
	}
	if s.Key() != "3" {
		t.Errorf("Key = %q", s.Key())
	}
	if got := s.Episodes(AudioDubbed); len(got) != 1 || got[0].Audio != AudioDubbed {
		t.Errorf("dubbed = %+v", got)
	}
	if got := s.Episodes(AudioSubtitled); len(got) != 2 {
		t.Errorf("subtitled = %+v", got)
	}
	if got := s.Episodes(""); len(got) != 1 {
		t.Errorf("no audio named falls back to the dub: %+v", got)
	}
}

// TestNoStreamErrorUnwrap: each host's failure stays reachable through the
// aggregate, so callers can still ask errors.Is about a specific cause.
func TestNoStreamErrorUnwrap(t *testing.T) {
	t.Parallel()
	cause := errors.New("byse gone")
	err := error(&NoStreamError{Failures: []error{fmt.Errorf("embedplaybyse.test: %w", cause)}})
	if !errors.Is(err, cause) {
		t.Error("the failure cause is not reachable through NoStreamError")
	}
	if !errors.Is(err, ErrNoSupportedServer) {
		t.Error("NoStreamError no longer matches ErrNoSupportedServer")
	}
}

// TestAbyssRedactTransportErr pins that a transport error never carries the
// /sora/ URL — it embeds the chunk token — into a log, even wrapped, while
// keeping what failed; other errors pass through untouched.
func TestAbyssRedactTransportErr(t *testing.T) {
	t.Parallel()
	const tokenURL = "https://qq.abyss.test/sora/123/SECRETTOKEN"
	cause := errors.New("connection reset by peer")
	for _, in := range []error{
		&url.Error{Op: "Get", URL: tokenURL, Err: cause},
		fmt.Errorf("abyss part 3: %w", &url.Error{Op: "Get", URL: tokenURL, Err: cause}),
	} {
		got := abyssRedactTransportErr(in)
		if strings.Contains(got.Error(), "SECRETTOKEN") {
			t.Errorf("%q leaks the chunk token", got)
		}
		if !strings.Contains(got.Error(), "abyss-chunk") || !errors.Is(got, cause) {
			t.Errorf("%q lost what failed", got)
		}
	}
	plain := errors.New("abyss chunk returned 10 bytes, expected 20")
	if got := abyssRedactTransportErr(plain); got != plain {
		t.Errorf("a non-transport error was rewritten: %v", got)
	}
}
