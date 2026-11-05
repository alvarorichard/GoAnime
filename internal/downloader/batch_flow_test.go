package downloader

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/iotest"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/alvarorichard/Goanime/v2/internal/models"
)

// ---- shared fakes ----

// answers scripts the prompts' input: each call fills the pointer it is
// given with the next answer; once the script runs out the input fails.
type answers struct {
	mu    sync.Mutex
	queue []any
	asked int
}

func (a *answers) scan(dst ...any) (int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.asked++
	if len(a.queue) == 0 {
		return 0, errors.New("EOF")
	}
	next := a.queue[0]
	a.queue = a.queue[1:]
	if err, ok := next.(error); ok {
		return 0, err
	}
	switch p := dst[0].(type) {
	case *string:
		*p = next.(string)
	case *int:
		*p = next.(int)
	}
	return 1, nil
}

// playRecorder records what was played.
type playRecorder struct {
	mu     sync.Mutex
	played []string
	err    error
}

func (p *playRecorder) start(path string, _ []string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.played = append(p.played, path)
	return "/tmp/sock", p.err
}

// flow is a downloader wired entirely to fakes: URLs come from a map, bytes
// from a loopback server, progress to a mock sender, input from a script.
type flow struct {
	d      *EpisodeDownloader
	sender *mockProgressSender
	in     *answers
	play   *playRecorder
	srv    *httptest.Server
	body   []byte

	mu       sync.Mutex
	requests []string // "METHOD path"
	inflight atomic.Int32
	maxSeen  atomic.Int32
	failPath map[string]int // path -> status to answer with
	hold     time.Duration  // delay per GET, to overlap downloads
	models   []*progressModel
}

func newFlow(t *testing.T, nEps int, answersList ...any) *flow {
	t.Helper()
	f := &flow{body: bytesOfSize(4096), failPath: map[string]int{}}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.requests = append(f.requests, r.Method+" "+r.URL.Path)
		status := f.failPath[r.URL.Path]
		f.mu.Unlock()
		if r.Method == http.MethodGet {
			n := f.inflight.Add(1)
			defer f.inflight.Add(-1)
			for {
				m := f.maxSeen.Load()
				if n <= m || f.maxSeen.CompareAndSwap(m, n) {
					break
				}
			}
			time.Sleep(f.hold)
		}
		if status != 0 {
			w.WriteHeader(status)
			return
		}
		w.Header().Set("Content-Length", fmt.Sprint(len(f.body)))
		if r.Method == http.MethodGet {
			_, _ = w.Write(f.body)
		}
	}))
	t.Cleanup(f.srv.Close)

	f.sender = newMockSender()
	f.in = &answers{queue: answersList}
	f.play = &playRecorder{}
	f.d = makeEpisodeDownloader(t, "Foo", makeEpisodes(nEps), false)
	f.d.opts = downloaderOptions{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		sleep:      func(time.Duration) {},
		newSender: func(m tea.Model) progressSender {
			if pm, ok := m.(*progressModel); ok {
				f.mu.Lock()
				f.models = append(f.models, pm)
				f.mu.Unlock()
			}
			return f.sender
		},
		videoURL: func(episodeURL string) (string, error) {
			// ep3 -> <srv>/v3.mp4
			return f.srv.URL + "/v" + strings.TrimPrefix(episodeURL, "https://example.test/ep") + ".mp4", nil
		},
		scan: f.in.scan,
		play: f.play.start,
	}
	return f
}

func (f *flow) path(ep int) string {
	return filepath.Join(f.d.episodeDir(ep), f.d.episodeFilename(ep))
}

func (f *flow) have(t *testing.T, eps ...int) {
	t.Helper()
	for _, ep := range eps {
		if err := os.MkdirAll(filepath.Dir(f.path(ep)), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(f.path(ep), f.body, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func (f *flow) count(prefix string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, r := range f.requests {
		if strings.HasPrefix(r, prefix) {
			n++
		}
	}
	return n
}

func mustErr(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), substr) {
		t.Fatalf("error = %v, want one containing %q", err, substr)
	}
}

func mustNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func fileSize(p string) int64 {
	fi, err := os.Stat(p)
	if err != nil {
		return -1
	}
	return fi.Size()
}

// ---- partitionRange / pendingEpisodes ----

type partitionFn func(d *EpisodeDownloader, start, end int) (todo, existing []int)

func partitionFailures(t *testing.T, fn partitionFn) []string {
	t.Helper()
	f := newFlow(t, 5)
	f.have(t, 2, 4)
	var bad []string
	chk := func(name string, start, end int, wantTodo, wantExisting []int) {
		todo, existing := fn(f.d, start, end)
		if !slices.Equal(todo, wantTodo) || !slices.Equal(existing, wantExisting) {
			bad = append(bad, fmt.Sprintf("%s: got %v/%v, want %v/%v", name, todo, existing, wantTodo, wantExisting))
		}
	}
	chk("whole range", 1, 5, []int{1, 3, 5}, []int{2, 4})
	chk("a single missing episode", 3, 3, []int{3}, nil)
	chk("a single existing episode", 4, 4, nil, []int{4})
	chk("unknown episodes are skipped", 4, 9, []int{5}, []int{4})
	chk("nothing in the source", 7, 9, nil, nil)
	chk("inverted range is empty", 5, 1, nil, nil)
	return bad
}

func TestPartitionRange(t *testing.T) {
	if bad := partitionFailures(t, (*EpisodeDownloader).partitionRange); len(bad) > 0 {
		t.Fatal(strings.Join(bad, "\n"))
	}
}

func TestMeta_PartitionRange_CatchesBrokenImplementations(t *testing.T) {
	mutants := map[string]partitionFn{
		"swapped lists": func(d *EpisodeDownloader, s, e int) ([]int, []int) {
			a, b := d.partitionRange(s, e)
			return b, a
		},
		"downloads everything": func(d *EpisodeDownloader, s, e int) ([]int, []int) {
			a, b := d.partitionRange(s, e)
			return append(slices.Clone(a), b...), nil
		},
		"does not skip unknown episodes": func(d *EpisodeDownloader, s, e int) ([]int, []int) {
			a, b := d.partitionRange(s, e)
			for n := s; n <= e; n++ {
				if _, ok := d.findEpisodeByNumber(n); !ok {
					a = append(a, n)
				}
			}
			return a, b
		},
		"end exclusive":   func(d *EpisodeDownloader, s, e int) ([]int, []int) { return d.partitionRange(s, e-1) },
		"start exclusive": func(d *EpisodeDownloader, s, e int) ([]int, []int) { return d.partitionRange(s+1, e) },
	}
	for name, fn := range mutants {
		t.Run(name, func(t *testing.T) {
			if len(partitionFailures(t, fn)) == 0 {
				t.Fatalf("mutant %q survived", name)
			}
		})
	}
}

type pendingFn func(d *EpisodeDownloader) ([]int, int)

func pendingFailures(t *testing.T, fn pendingFn) []string {
	t.Helper()
	f := newFlow(t, 0)
	f.d.episodes = []models.Episode{
		{Num: 1, Number: "9"}, // Num wins even at the boundary value 1
		{Num: 0, Number: "1"}, // text-only episode 1
		{Num: 2, Number: "2"},
		{Num: 0, Number: "3"},   // number only in the text
		{Num: 0, Number: "x"},   // no usable number
		{Num: -1, Number: "0"},  // zero is not a number
		{Num: 0, Number: "-4"},  // negative is not a number
		{Num: 6, Number: "six"}, // Num wins over the text
	}
	f.have(t, 2, 3)
	todo, existing := fn(f.d)
	var bad []string
	if !slices.Equal(todo, []int{1, 1, 6}) || existing != 2 {
		bad = append(bad, fmt.Sprintf("got %v/%d, want [1 1 6]/2", todo, existing))
	}
	return bad
}

func TestPendingEpisodes(t *testing.T) {
	if bad := pendingFailures(t, (*EpisodeDownloader).pendingEpisodes); len(bad) > 0 {
		t.Fatal(strings.Join(bad, "\n"))
	}
}

func TestMeta_PendingEpisodes_CatchesBrokenImplementations(t *testing.T) {
	mutants := map[string]pendingFn{
		"ignores the text number": func(d *EpisodeDownloader) ([]int, int) {
			keep := d.episodes
			defer func() { d.episodes = keep }()
			d.episodes = slices.DeleteFunc(slices.Clone(keep), func(e models.Episode) bool { return e.Num <= 0 })
			return d.pendingEpisodes()
		},
		"counts nothing": func(d *EpisodeDownloader) ([]int, int) { a, _ := d.pendingEpisodes(); return a, 0 },
		"includes existing": func(d *EpisodeDownloader) ([]int, int) {
			a, n := d.pendingEpisodes()
			return append(a, 2), n
		},
		"text number one rejected": func(d *EpisodeDownloader) ([]int, int) {
			a, n := d.pendingEpisodes()
			return a[1:], n
		},
		"keeps unusable numbers": func(d *EpisodeDownloader) ([]int, int) {
			a, n := d.pendingEpisodes()
			return append(a, 0), n
		},
	}
	for name, fn := range mutants {
		t.Run(name, func(t *testing.T) {
			if len(pendingFailures(t, fn)) == 0 {
				t.Fatalf("mutant %q survived", name)
			}
		})
	}
}

// ---- tiny pure helpers ----

func TestIsYes(t *testing.T) {
	for _, c := range []struct {
		in   string
		want bool
	}{{"y", true}, {"Y", true}, {"yes", true}, {"YES", true}, {"Yes", true}, {"n", false}, {"no", false}, {"", false}, {"yess", false}, {" y", false}, {"ye", false}} {
		if got := isYes(c.in); got != c.want {
			t.Errorf("isYes(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestEpisodeChoice(t *testing.T) {
	nums := []int{3, 5, 8}
	for _, c := range []struct {
		in   int
		want int
		ok   bool
	}{{3, 3, true}, {5, 5, true}, {8, 8, true}, {0, 0, false}, {4, 0, false}, {-3, 0, false}, {9, 0, false}} {
		if got, ok := episodeChoice(c.in, nums); got != c.want || ok != c.ok {
			t.Errorf("episodeChoice(%d) = (%d,%v), want (%d,%v)", c.in, got, ok, c.want, c.ok)
		}
	}
	if _, ok := episodeChoice(1, nil); ok {
		t.Error("an empty list must resolve nothing")
	}
}

func TestIsOpaqueStreamURL(t *testing.T) {
	for url, want := range map[string]bool{
		"https://x.sharepoint.com/v.aspx":    true,
		"https://repackager.wixmp.com/v.mp4": true,
		"https://cdn/a/master.m3u8":          true,
		"https://cdn/a.m3u8?t=1":             true,
		"https://www.blogger.com/video.g":    true,
		"https://cdn.example.com/v.mp4":      false,
		"":                                   false,
	} {
		if got := isOpaqueStreamURL(url); got != want {
			t.Errorf("isOpaqueStreamURL(%q) = %v, want %v", url, got, want)
		}
	}
}

// ---- prompts ----

type promptCase struct {
	name       string
	run        func(f *flow) error
	answers    []any
	wantErr    string
	wantPlayed []int // episodes whose path must have been played
	wantAsked  int
}

func promptCases() []promptCase {
	played := func(eps ...int) []int { return eps }
	single := func(name string, run func(f *flow) error) []promptCase {
		return []promptCase{
			{name: name + ": yes plays", run: run, answers: []any{"y"}, wantPlayed: played(7), wantAsked: 1},
			{name: name + ": YES plays", run: run, answers: []any{"YES"}, wantPlayed: played(7), wantAsked: 1},
			{name: name + ": no does not play", run: run, answers: []any{"n"}, wantAsked: 1},
			{name: name + ": an input error is not fatal and does not play", run: run, answers: []any{errors.New("closed")}, wantAsked: 1},
		}
	}
	rng := func(name string, run func(f *flow) error) []promptCase {
		return []promptCase{
			{name: name + ": a listed episode plays", run: run, answers: []any{5}, wantPlayed: played(5), wantAsked: 1},
			{name: name + ": the first listed episode plays", run: run, answers: []any{1}, wantPlayed: played(1), wantAsked: 1},
			{name: name + ": 0 exits", run: run, answers: []any{0}, wantAsked: 1},
			{name: name + ": an unlisted episode plays nothing", run: run, answers: []any{4}, wantAsked: 1},
			{name: name + ": 2 is unlisted and plays nothing", run: run, answers: []any{2}, wantAsked: 1},
			{name: name + ": an input error is returned", run: run, answers: []any{errors.New("closed")}, wantErr: "failed to read input", wantAsked: 1},
		}
	}
	var all []promptCase
	all = append(all, single("existing", func(f *flow) error {
		return f.d.promptPlayExisting(7, f.path(7))
	})...)
	all = append(all, single("downloaded", func(f *flow) error {
		return f.d.promptPlayDownloaded(7, f.path(7))
	})...)
	all = append(all, rng("downloaded range", func(f *flow) error {
		return f.d.promptPlayDownloadedRangeHuh([]int{1, 5, 8})
	})...)
	all = append(all, rng("existing range", func(f *flow) error {
		return f.d.promptPlayExistingRangeHuh([]int{1, 5, 8})
	})...)
	all = append(all,
		promptCase{name: "downloaded range: empty list does not even ask", run: func(f *flow) error { return f.d.promptPlayDownloadedRangeHuh(nil) }, answers: []any{3}},
		promptCase{name: "existing range: empty list does not even ask", run: func(f *flow) error { return f.d.promptPlayExistingRangeHuh(nil) }, answers: []any{3}},
	)
	return all
}

func promptFailures(t *testing.T, tweak func(f *flow)) []string {
	t.Helper()
	var bad []string
	for _, c := range promptCases() {
		f := newFlow(t, 9, c.answers...)
		if tweak != nil {
			tweak(f)
		}
		err := c.run(f)
		if (c.wantErr == "") != (err == nil) || (err != nil && !strings.Contains(err.Error(), c.wantErr)) {
			bad = append(bad, fmt.Sprintf("%s: error %v, want %q", c.name, err, c.wantErr))
		}
		var want []string
		for _, ep := range c.wantPlayed {
			want = append(want, f.path(ep))
		}
		if !slices.Equal(f.play.played, want) {
			bad = append(bad, fmt.Sprintf("%s: played %v, want %v", c.name, f.play.played, want))
		}
		if f.in.asked != c.wantAsked {
			bad = append(bad, fmt.Sprintf("%s: asked %d times, want %d", c.name, f.in.asked, c.wantAsked))
		}
	}
	return bad
}

func TestPrompts(t *testing.T) {
	if bad := promptFailures(t, nil); len(bad) > 0 {
		t.Fatal(strings.Join(bad, "\n"))
	}
}

func TestPlayEpisode(t *testing.T) {
	f := newFlow(t, 1)
	mustNoErr(t, f.d.playEpisode("/x/ep1.mp4", 1))
	if !slices.Equal(f.play.played, []string{"/x/ep1.mp4"}) {
		t.Fatalf("played %v", f.play.played)
	}
	f.play.err = errors.New("no mpv")
	mustErr(t, f.d.playEpisode("/x/ep1.mp4", 1), "failed to start video")
}

func TestMeta_Prompts_CatchBrokenInput(t *testing.T) {
	mutants := map[string]func(f *flow){
		"player always errors": func(f *flow) { f.play.err = errors.New("boom") },
		"input answers wrong type": func(f *flow) {
			orig := f.d.opts.scan
			f.d.opts.scan = func(a ...any) (int, error) {
				n, err := orig(a...)
				if p, ok := a[0].(*int); ok && err == nil {
					*p++
				}
				return n, err
			}
		},
		"input never fails": func(f *flow) {
			orig := f.d.opts.scan
			f.d.opts.scan = func(a ...any) (int, error) { n, _ := orig(a...); return max(n, 1), nil }
		},
	}
	for name, tweak := range mutants {
		t.Run(name, func(t *testing.T) {
			if len(promptFailures(t, tweak)) == 0 {
				t.Fatalf("mutant %q survived", name)
			}
		})
	}
}

// ---- getContentLength ----

func TestGetContentLength(t *testing.T) {
	f := newFlow(t, 1)

	n, err := f.d.getContentLength(f.srv.URL + "/v1.mp4")
	mustNoErr(t, err)
	if n != int64(len(f.body)) {
		t.Fatalf("size = %d, want %d", n, len(f.body))
	}
	if f.count("HEAD") != 1 {
		t.Fatalf("expected one HEAD, got %v", f.requests)
	}

	// An HLS playlist is never probed.
	before := len(f.requests)
	n, err = f.d.getContentLength(f.srv.URL + "/a/master.m3u8")
	mustNoErr(t, err)
	if n != 400*1024*1024 || len(f.requests) != before {
		t.Fatalf("m3u8: size %d, requests %v", n, f.requests[before:])
	}

	// A server that answers without a length: plain hosts fail, opaque ones estimate.
	empty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Transfer-Encoding", "chunked")
		w.WriteHeader(200)
	}))
	t.Cleanup(empty.Close)
	n, err = f.d.getContentLength(empty.URL + "/plain.mp4")
	mustErr(t, err, "content-length header missing")
	if n != 0 {
		t.Fatalf("a failed probe must report size 0, got %d", n)
	}
	n, err = f.d.getContentLength(empty.URL + "/x.mp4?host=wixmp.com")
	mustNoErr(t, err)
	if n != 300*1024*1024 {
		t.Fatalf("opaque estimate = %d", n)
	}

	// A server that is down: plain hosts fail, opaque ones estimate.
	dead := httptest.NewServer(http.NotFoundHandler())
	deadURL := dead.URL
	dead.Close()
	n, err = f.d.getContentLength(deadURL + "/v.mp4")
	if err == nil || n != 0 {
		t.Fatalf("a dead plain host must fail with size 0, got %d, %v", n, err)
	}
	n, err = f.d.getContentLength(deadURL + "/v.mp4?blogger.com")
	mustNoErr(t, err)
	if n != 300*1024*1024 {
		t.Fatalf("dead opaque estimate = %d", n)
	}

	// A malformed URL: plain fails, opaque estimates.
	n, err = f.d.getContentLength("http://[::1")
	if err == nil || n != 0 {
		t.Fatalf("a malformed plain URL must fail with size 0, got %d, %v", n, err)
	}
	n, err = f.d.getContentLength("http://[::1/wixmp.com")
	mustNoErr(t, err)
	if n != 300*1024*1024 {
		t.Fatalf("malformed opaque estimate = %d", n)
	}
}

func TestGetContentLength_SendsAnOriginReferrerOnlyToOpaqueHosts(t *testing.T) {
	var gotReferer atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReferer.Store(r.Header.Get("Referer"))
		w.Header().Set("Content-Length", "10")
	}))
	t.Cleanup(srv.Close)
	f := newFlow(t, 1)
	_, _ = f.d.getContentLength(srv.URL + "/plain.mp4")
	if r := gotReferer.Load(); r != "" {
		t.Fatalf("plain host got a referer %q", r)
	}
	_, _ = f.d.getContentLength(srv.URL + "/x.mp4?sharepoint.com")
	if r := gotReferer.Load(); r != "https://allmanga.to" {
		t.Fatalf("opaque host referer = %v", r)
	}
}

// ---- the single / range / all entry points ----

func TestDownloadSingleEpisode(t *testing.T) {
	t.Run("unknown episode", func(t *testing.T) {
		f := newFlow(t, 2)
		mustErr(t, f.d.DownloadSingleEpisode(9), "episode 9 not found")
		if len(f.requests) != 0 {
			t.Fatalf("unexpected requests %v", f.requests)
		}
	})
	t.Run("already on disk asks to play and downloads nothing", func(t *testing.T) {
		f := newFlow(t, 2, "y")
		f.have(t, 1)
		mustNoErr(t, f.d.DownloadSingleEpisode(1))
		if len(f.requests) != 0 || !slices.Equal(f.play.played, []string{f.path(1)}) {
			t.Fatalf("requests %v played %v", f.requests, f.play.played)
		}
	})
	t.Run("url resolution failure", func(t *testing.T) {
		f := newFlow(t, 1)
		f.d.opts.videoURL = func(string) (string, error) { return "", errors.New("provider down") }
		mustErr(t, f.d.DownloadSingleEpisode(1), "failed to get video URL")
		if fileSize(f.path(1)) != -1 {
			t.Fatal("nothing must be written")
		}
	})
	t.Run("downloads then offers to play", func(t *testing.T) {
		f := newFlow(t, 1, "y")
		mustNoErr(t, f.d.DownloadSingleEpisode(1))
		if fileSize(f.path(1)) != int64(len(f.body)) {
			t.Fatalf("size %d", fileSize(f.path(1)))
		}
		if !slices.Equal(f.play.played, []string{f.path(1)}) {
			t.Fatalf("played %v", f.play.played)
		}
		if !f.sender.hasStatusContaining("Download complete") {
			t.Fatal("no completion status")
		}
	})
	t.Run("declining to play still succeeds", func(t *testing.T) {
		f := newFlow(t, 1, "n")
		mustNoErr(t, f.d.DownloadSingleEpisode(1))
		if len(f.play.played) != 0 {
			t.Fatalf("played %v", f.play.played)
		}
	})
	t.Run("a server error surfaces", func(t *testing.T) {
		f := newFlow(t, 1)
		f.failPath["/v1.mp4"] = 500
		if err := f.d.DownloadSingleEpisode(1); err == nil {
			t.Fatal("expected an error")
		}
	})
}

func TestDownloadEpisodeRange(t *testing.T) {
	t.Run("inverted range", func(t *testing.T) {
		f := newFlow(t, 5)
		mustErr(t, f.d.DownloadEpisodeRange(4, 2), "cannot be greater")
	})
	t.Run("a one-episode range is valid", func(t *testing.T) {
		f := newFlow(t, 4, 0)
		mustNoErr(t, f.d.DownloadEpisodeRange(3, 3))
		if fileSize(f.path(3)) != int64(len(f.body)) {
			t.Fatal("episode 3 missing")
		}
	})
	t.Run("everything exists: offers to play and downloads nothing", func(t *testing.T) {
		f := newFlow(t, 4, 2)
		f.have(t, 1, 2, 3)
		mustNoErr(t, f.d.DownloadEpisodeRange(1, 3))
		if len(f.requests) != 0 || !slices.Equal(f.play.played, []string{f.path(2)}) {
			t.Fatalf("requests %v played %v", f.requests, f.play.played)
		}
	})
	t.Run("downloads only what is missing", func(t *testing.T) {
		f := newFlow(t, 4, 0) // then chooses 0 = exit
		f.have(t, 2)
		mustNoErr(t, f.d.DownloadEpisodeRange(1, 4))
		for _, ep := range []int{1, 3, 4} {
			if fileSize(f.path(ep)) != int64(len(f.body)) {
				t.Errorf("episode %d size %d", ep, fileSize(f.path(ep)))
			}
		}
		if f.count("GET /v2") != 0 {
			t.Error("the existing episode was downloaded again")
		}
		if !f.sender.hasStatusContaining("All downloads complete") {
			t.Error("no completion status")
		}
	})
	t.Run("one failing episode fails the batch but keeps the others", func(t *testing.T) {
		f := newFlow(t, 3)
		f.failPath["/v2.mp4"] = 500
		err := f.d.DownloadEpisodeRange(1, 3)
		mustErr(t, err, "1 download(s) failed")
		if fileSize(f.path(1)) != int64(len(f.body)) || fileSize(f.path(3)) != int64(len(f.body)) {
			t.Error("the healthy episodes must still be saved")
		}
		if !f.sender.hasStatusContaining("Some downloads failed") {
			t.Error("no failure status")
		}
		if f.in.asked != 0 {
			t.Error("no play prompt after a failed batch")
		}
	})
}

func TestDownloadAllEpisodes(t *testing.T) {
	t.Run("no episodes", func(t *testing.T) {
		f := newFlow(t, 0)
		mustErr(t, f.d.DownloadAllEpisodes(), "no episodes available")
	})
	t.Run("everything already downloaded", func(t *testing.T) {
		f := newFlow(t, 3)
		f.have(t, 1, 2, 3)
		mustNoErr(t, f.d.DownloadAllEpisodes())
		if len(f.requests) != 0 || f.sender.runCount.Load() != 0 {
			t.Fatalf("requests %v runs %d", f.requests, f.sender.runCount.Load())
		}
	})
	t.Run("a single missing episode is still downloaded", func(t *testing.T) {
		f := newFlow(t, 3, 0)
		f.have(t, 1, 3)
		mustNoErr(t, f.d.DownloadAllEpisodes())
		if fileSize(f.path(2)) != int64(len(f.body)) {
			t.Fatal("episode 2 missing")
		}
	})
	t.Run("downloads the rest", func(t *testing.T) {
		f := newFlow(t, 4, 0)
		f.have(t, 1, 4)
		mustNoErr(t, f.d.DownloadAllEpisodes())
		for _, ep := range []int{2, 3} {
			if fileSize(f.path(ep)) != int64(len(f.body)) {
				t.Errorf("episode %d missing", ep)
			}
		}
		if f.count("GET /v1") != 0 || f.count("GET /v4") != 0 {
			t.Error("existing episodes were downloaded again")
		}
	})
}

func TestDownloadConcurrentWithProgress(t *testing.T) {
	t.Run("nothing to do", func(t *testing.T) {
		f := newFlow(t, 1)
		mustNoErr(t, f.d.downloadConcurrentWithProgress(nil))
		if f.sender.runCount.Load() != 0 {
			t.Fatal("no progress screen for an empty batch")
		}
	})
	t.Run("no episode resolves", func(t *testing.T) {
		f := newFlow(t, 2)
		f.d.opts.videoURL = func(string) (string, error) { return "", errors.New("down") }
		mustErr(t, f.d.downloadConcurrentWithProgress([]int{1, 2}), "no episodes could be resolved")
	})
	t.Run("an unknown number is skipped, the rest download", func(t *testing.T) {
		f := newFlow(t, 2, 0)
		mustNoErr(t, f.d.downloadConcurrentWithProgress([]int{1, 99}))
		if fileSize(f.path(1)) != int64(len(f.body)) {
			t.Fatal("episode 1 missing")
		}
	})
	t.Run("a progress screen error is reported", func(t *testing.T) {
		f := newFlow(t, 1)
		f.sender.runErr = errors.New("tty lost")
		mustErr(t, f.d.downloadConcurrentWithProgress([]int{1}), "progress display error")
	})
	t.Run("the total is the sum of the sizes, with a default for unknown ones", func(t *testing.T) {
		f := newFlow(t, 2, 0)
		// Episode 2's size probe fails: it counts as the 100 MB default.
		f.failPath["/v2.mp4"] = 0
		orig := f.d.opts.httpClient
		f.d.opts.httpClient = &http.Client{Timeout: 5 * time.Second, Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method == http.MethodHead && strings.HasSuffix(r.URL.Path, "/v2.mp4") {
				return nil, errors.New("probe failed")
			}
			return orig.Do(r)
		})}
		_ = f.d.downloadConcurrentWithProgress([]int{1, 2})
		f.mu.Lock()
		defer f.mu.Unlock()
		if len(f.models) != 1 {
			t.Fatalf("models = %d", len(f.models))
		}
		want := int64(len(f.body)) + 100*1024*1024
		if got := f.models[0].totalBytes; got != want {
			t.Fatalf("totalBytes = %d, want %d", got, want)
		}
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// ---- concurrency limit ----

func concurrencyFailures(t *testing.T, tweak func(d *EpisodeDownloader)) []string {
	t.Helper()
	f := newFlow(t, 6, 0)
	f.hold = 60 * time.Millisecond
	f.d.config.Concurrent = 2
	if tweak != nil {
		tweak(f.d)
	}
	var bad []string
	if err := f.d.DownloadEpisodeRange(1, 6); err != nil {
		bad = append(bad, "batch failed: "+err.Error())
	}
	if got := f.maxSeen.Load(); got > 2 {
		bad = append(bad, fmt.Sprintf("%d downloads ran at once, limit is 2", got))
	}
	if got := f.maxSeen.Load(); got < 2 {
		bad = append(bad, fmt.Sprintf("only %d download ran at once: downloads are not concurrent", got))
	}
	for ep := 1; ep <= 6; ep++ {
		if fileSize(f.path(ep)) != int64(len(f.body)) {
			bad = append(bad, fmt.Sprintf("episode %d not saved", ep))
		}
	}
	return bad
}

func TestDownloadMultiple_RespectsTheConcurrencyLimit(t *testing.T) {
	if bad := concurrencyFailures(t, nil); len(bad) > 0 {
		t.Fatal(strings.Join(bad, "\n"))
	}
}

func TestMeta_Concurrency_CatchesBrokenLimits(t *testing.T) {
	mutants := map[string]func(d *EpisodeDownloader){
		"limit ignored (unbounded)": func(d *EpisodeDownloader) { d.config.Concurrent = 6 },
		"serial":                    func(d *EpisodeDownloader) { d.config.Concurrent = 1 },
	}
	for name, tweak := range mutants {
		t.Run(name, func(t *testing.T) {
			if len(concurrencyFailures(t, tweak)) == 0 {
				t.Fatalf("mutant %q survived", name)
			}
		})
	}
}

// ---- downloadEpisodeWithSharedProgress ----

func TestDownloadEpisodeWithSharedProgress(t *testing.T) {
	t.Run("saves the file and tracks shared progress", func(t *testing.T) {
		f := newFlow(t, 2)
		pm := newProgressModelForTest(int64(2 * len(f.body)))
		var ep1, ep2, total int64
		var mu sync.Mutex
		mustNoErr(t, f.d.downloadEpisodeWithSharedProgress(f.srv.URL+"/v1.mp4", f.path(1), &ep1, &total, &mu, pm, f.sender))
		mustNoErr(t, f.d.downloadEpisodeWithSharedProgress(f.srv.URL+"/v2.mp4", f.path(2), &ep2, &total, &mu, pm, f.sender))
		if ep1 != int64(len(f.body)) || ep2 != int64(len(f.body)) || total != 2*int64(len(f.body)) {
			t.Fatalf("ep1=%d ep2=%d total=%d", ep1, ep2, total)
		}
		if pm.received != total {
			t.Fatalf("model received %d, want %d", pm.received, total)
		}
		msgs := f.sender.sentMessages()
		last, ok := msgs[len(msgs)-1].(progressMsg)
		if !ok || last.received != total || last.totalBytes != pm.totalBytes {
			t.Fatalf("last progress message %+v", msgs[len(msgs)-1])
		}
		got, _ := os.ReadFile(f.path(2))
		if !bytes.Equal(got, f.body) {
			t.Fatal("file content differs from the served body")
		}
	})
	t.Run("a bad status writes nothing", func(t *testing.T) {
		f := newFlow(t, 1)
		f.failPath["/v1.mp4"] = 404
		var a, b int64
		var mu sync.Mutex
		err := f.d.downloadEpisodeWithSharedProgress(f.srv.URL+"/v1.mp4", f.path(1), &a, &b, &mu, newProgressModelForTest(1), f.sender)
		mustErr(t, err, "bad status")
		if fileSize(f.path(1)) != -1 {
			t.Fatal("a failed download must not leave a file")
		}
	})
	t.Run("an unreachable server fails to start", func(t *testing.T) {
		f := newFlow(t, 1)
		var a, b int64
		var mu sync.Mutex
		err := f.d.downloadEpisodeWithSharedProgress("http://127.0.0.1:1/v.mp4", f.path(1), &a, &b, &mu, newProgressModelForTest(1), f.sender)
		mustErr(t, err, "failed to start download")
	})
	t.Run("a path outside the output directory is refused", func(t *testing.T) {
		f := newFlow(t, 1)
		var a, b int64
		var mu sync.Mutex
		escape := filepath.Join(f.d.config.OutputDir, "..", "evil.mp4")
		err := f.d.downloadEpisodeWithSharedProgress(f.srv.URL+"/v1.mp4", escape, &a, &b, &mu, newProgressModelForTest(1), f.sender)
		mustErr(t, err, "invalid destination path")
	})
	t.Run("an uncreatable file is reported", func(t *testing.T) {
		f := newFlow(t, 1)
		var a, b int64
		var mu sync.Mutex
		missingDir := filepath.Join(f.d.config.OutputDir, "no", "such", "dir", "v.mp4")
		err := f.d.downloadEpisodeWithSharedProgress(f.srv.URL+"/v1.mp4", missingDir, &a, &b, &mu, newProgressModelForTest(1), f.sender)
		mustErr(t, err, "failed to create file")
	})
}

// A body that arrives one byte at a time must be saved whole: the copy loop
// may not drop short reads.
func TestDownloadEpisodeWithSharedProgress_ShortReads(t *testing.T) {
	f := newFlow(t, 1)
	body := bytesOfSize(300)
	f.d.opts.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200, Status: "200 OK",
			Body: io.NopCloser(iotest.OneByteReader(bytes.NewReader(body))),
		}, nil
	})}
	var a, total int64
	var mu sync.Mutex
	pm := newProgressModelForTest(int64(len(body)))
	mustNoErr(t, f.d.downloadEpisodeWithSharedProgress("http://x/v.mp4", f.path(1), &a, &total, &mu, pm, f.sender))
	got, _ := os.ReadFile(f.path(1))
	if !bytes.Equal(got, body) || total != int64(len(body)) {
		t.Fatalf("saved %d bytes (total %d), want %d", len(got), total, len(body))
	}
}
