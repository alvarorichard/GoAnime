package player

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeWeb stands in for the SSRF-guarded GET: it serves canned pages and
// records every URL asked for.
type fakeWeb struct {
	pages map[string]fakePage
	asked []string
}

type fakePage struct {
	status int
	body   string
	err    error
}

func page(body string) fakePage { return fakePage{status: http.StatusOK, body: body} }

func (w *fakeWeb) get(url string) (*http.Response, error) {
	w.asked = append(w.asked, url)
	p, ok := w.pages[url]
	if !ok {
		return nil, fmt.Errorf("no such host in the fake web: %s", url)
	}
	if p.err != nil {
		return nil, p.err
	}
	status := p.status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(p.body)),
	}, nil
}

// useFakeWeb installs w as the package's GET. Tests using it are not parallel.
func useFakeWeb(t *testing.T, w *fakeWeb) {
	t.Helper()
	prev := safeGetFn
	safeGetFn = w.get
	t.Cleanup(func() { safeGetFn = prev })
}

func withQuality(t *testing.T, q string) {
	t.Helper()
	prev := util.GlobalQuality
	util.GlobalQuality = q
	t.Cleanup(func() { util.GlobalQuality = prev })
}

// ---- qualityOfLabel / qualityOfURL ---------------------------------------------------

var labelCases = map[string]int{"720p": 720, "1080p": 1080, "FHD": 0, "Auto": 0, "": 0, "480": 480, "HD 720": 720, "4K 2160p": 4, "v2 360p": 2}

func labelFailures(impl func(string) int) (f []string) {
	for in, want := range labelCases {
		if got := impl(in); got != want {
			f = append(f, fmt.Sprintf("%q: got %d want %d", in, got, want))
		}
	}
	return f
}

func TestQualityOfLabel(t *testing.T) {
	assert.Empty(t, labelFailures(qualityOfLabel))
}

func TestQualityOfLabelMeta(t *testing.T) {
	require.Empty(t, labelFailures(qualityOfLabel))
	mutants := map[string]func(string) int{
		"always zero": func(string) int { return 0 },
		"last number": func(s string) int {
			all := digitsRe.FindAllString(s, -1)
			if len(all) == 0 {
				return 0
			}
			var n int
			_, _ = fmt.Sscanf(all[len(all)-1], "%d", &n)
			return n
		},
		"first digit only": func(s string) int {
			if d := digitsRe.FindString(s); d != "" {
				return int(d[0] - '0')
			}
			return 0
		},
		"label with letters is zero": func(s string) int {
			if strings.ContainsAny(s, "pP") && false {
				return 0
			}
			if strings.ContainsAny(s, " ") {
				return 0
			}
			return qualityOfLabel(s)
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, labelFailures(m), "mutant %q survived", name)
	}
}

var urlQualityCases = map[string]int{
	"https://cdn.test/v/movie_720p.mp4": 720,
	"https://cdn.test/v/1080.mp4":       1080,
	"https://cdn.test/v/480p.mp4":       480,
	"https://cdn.test/v/movie.mp4":      0,
	"https://cdn.test/v/master.m3u8":    0,
	"":                                  0,
	"https://cdn.test/v/720p.m3u8":      0,
}

func urlQualityFailures(impl func(string) int) (f []string) {
	for in, want := range urlQualityCases {
		if got := impl(in); got != want {
			f = append(f, fmt.Sprintf("%q: got %d want %d", in, got, want))
		}
	}
	return f
}

func TestQualityOfURL(t *testing.T) {
	assert.Empty(t, urlQualityFailures(qualityOfURL))
}

func TestQualityOfURLMeta(t *testing.T) {
	require.Empty(t, urlQualityFailures(qualityOfURL))
	mutants := map[string]func(string) int{
		"always zero": func(string) int { return 0 },
		"any extension": func(s string) int {
			if strings.HasSuffix(s, ".m3u8") {
				return 720
			}
			return qualityOfURL(s)
		},
		"ignores the number": func(s string) int {
			if qualityOfURL(s) > 0 {
				return 1
			}
			return 0
		},
		"requires the p": func(s string) int {
			if !strings.Contains(s, "p.mp4") {
				return 0
			}
			return qualityOfURL(s)
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, urlQualityFailures(m), "mutant %q survived", name)
	}
}

func TestSourcesFromEntriesKeepsOrderAndLabels(t *testing.T) {
	got := sourcesFromEntries([]videoEntry{{Src: "a", Label: "360p"}, {Src: "b", Label: "FHD"}, {Src: "c", Label: "1080p"}})
	require.Len(t, got, 3)
	assert.Equal(t, 360, got[0].Quality)
	assert.Equal(t, "a", got[0].URL)
	assert.Equal(t, 0, got[1].Quality)
	assert.Equal(t, "c", got[2].URL)
	assert.Equal(t, 1080, got[2].Quality)
	assert.Empty(t, sourcesFromEntries(nil))
}

// ---- ExtractVideoSources ------------------------------------------------------------

const (
	epPage      = "https://animefire.io/animes/x/1"
	fireAPI     = "https://animefire.io/video/x"
	apiTwoItems = `{"data":[{"src":"https://cdn.test/360.mp4","label":"360p"},{"src":"https://cdn.test/720.mp4","label":"720p"}]}`
)

func pageWithVideoSrc(src string) string {
	if strings.Contains(src, `"`) { // JSON: the attribute must be single-quoted
		return `<html><body><div data-video-src='` + src + `'></div></body></html>`
	}
	return `<html><body><div data-video-src="` + src + `"></div></body></html>`
}

type srcRow struct {
	q   int
	url string
}

func rows(in []struct {
	Quality int
	URL     string
}) []srcRow {
	var out []srcRow
	for _, s := range in {
		out = append(out, srcRow{s.Quality, s.URL})
	}
	return out
}

func TestExtractVideoSources(t *testing.T) {
	tests := []struct {
		name    string
		web     map[string]fakePage
		want    []srcRow
		wantErr bool
	}{
		{
			"animefire API with two qualities",
			map[string]fakePage{epPage: page(pageWithVideoSrc(fireAPI)), fireAPI: page(apiTwoItems)},
			[]srcRow{{360, "https://cdn.test/360.mp4"}, {720, "https://cdn.test/720.mp4"}}, false,
		},
		{
			"video src that is itself the JSON",
			map[string]fakePage{epPage: page(pageWithVideoSrc(`{"data":[{"src":"https://x/a.mp4","label":"480p"}]}`))},
			[]srcRow{{480, "https://x/a.mp4"}}, false,
		},
		{
			"animefire API with a single quality",
			map[string]fakePage{epPage: page(pageWithVideoSrc(fireAPI)), fireAPI: page(`{"data":[{"src":"https://cdn.test/only.mp4","label":"480p"}]}`)},
			[]srcRow{{480, "https://cdn.test/only.mp4"}}, false,
		},
		{
			"video src that is JSON with no entries falls back to the text",
			map[string]fakePage{epPage: page(pageWithVideoSrc(`{"data":[]}`))},
			[]srcRow{{0, `{"data":[]}`}}, false,
		},
		{
			"direct mp4 spelling its resolution",
			map[string]fakePage{epPage: page(pageWithVideoSrc("https://cdn.test/e_1080p.mp4"))},
			[]srcRow{{1080, "https://cdn.test/e_1080p.mp4"}}, false,
		},
		{
			"direct link with no resolution",
			map[string]fakePage{epPage: page(pageWithVideoSrc("https://cdn.test/master.m3u8"))},
			[]srcRow{{0, "https://cdn.test/master.m3u8"}}, false,
		},
		{
			"animefire API with no data falls back to the link itself",
			map[string]fakePage{epPage: page(pageWithVideoSrc(fireAPI)), fireAPI: page(`{"data":[]}`)},
			[]srcRow{{0, fireAPI}}, false,
		},
		{
			"animefire API that is not JSON falls back to the link itself",
			map[string]fakePage{epPage: page(pageWithVideoSrc(fireAPI)), fireAPI: page(`<html>blocked</html>`)},
			[]srcRow{{0, fireAPI}}, false,
		},
		{
			"animefire API unreachable",
			map[string]fakePage{epPage: page(pageWithVideoSrc(fireAPI)), fireAPI: {err: errors.New("dial tcp: refused")}},
			nil, true,
		},
		{
			"episode page unreachable",
			map[string]fakePage{epPage: {err: errors.New("dial tcp: refused")}},
			nil, true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			web := &fakeWeb{pages: tc.web}
			useFakeWeb(t, web)
			got, err := ExtractVideoSources(epPage)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, rows(got))
		})
	}
}

func TestExtractVideoSourcesFetchesTheAPIOnlyForAnimefireLinks(t *testing.T) {
	web := &fakeWeb{pages: map[string]fakePage{epPage: page(pageWithVideoSrc("https://cdn.test/e.mp4"))}}
	useFakeWeb(t, web)
	_, err := ExtractVideoSources(epPage)
	require.NoError(t, err)
	assert.Equal(t, []string{epPage}, web.asked, "a plain link costs no second request")
}

// ---- extractVideoURL ----------------------------------------------------------------

func TestExtractVideoURL(t *testing.T) {
	const u = "https://site.test/ep"
	tests := []struct {
		name    string
		html    string
		want    string
		wantErr string
	}{
		{"data-video-src on a div", `<div data-video-src="https://a/1.mp4"></div>`, "https://a/1.mp4", ""},
		{"video element src", `<video src="https://a/2.mp4"></video>`, "https://a/2.mp4", ""},
		{"video wins over a later iframe", `<iframe src="https://a/player"></iframe><video src="https://a/3.mp4"></video>`, "https://a/3.mp4", ""},
		{"data-video-src wins over src on the same element", `<video data-video-src="https://a/d.mp4" src="https://a/s.mp4"></video>`, "https://a/d.mp4", ""},
		{"empty attribute is skipped", `<video data-video-src="" src="https://a/s.mp4"></video>`, "https://a/s.mp4", ""},
		{"blogger iframe", `<iframe src="https://www.blogger.com/video.g?token=abc"></iframe>`, "https://www.blogger.com/video.g?token=abc", ""},
		{"blogger link in the page text", `<p>watch https://www.blogger.com/video.g?token=AbC_9-x here</p>`, "https://www.blogger.com/video.g?token=AbC_9-x", ""},
		{"direct mp4 in the page text", `<script>var u="https://cdn.test/v/e.mp4";</script>`, "https://cdn.test/v/e.mp4", ""},
		{"nothing at all", `<p>no video here</p>`, "", "no video source found"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			useFakeWeb(t, &fakeWeb{pages: map[string]fakePage{u: page(tc.html)}})
			got, err := extractVideoURL(u)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}

	t.Run("fetch failure", func(t *testing.T) {
		useFakeWeb(t, &fakeWeb{pages: map[string]fakePage{u: {err: errors.New("refused")}}})
		_, err := extractVideoURL(u)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch URL")
	})
}

func TestFetchContent(t *testing.T) {
	useFakeWeb(t, &fakeWeb{pages: map[string]fakePage{
		"https://ok.test":   page("hello"),
		"https://gone.test": {status: http.StatusNotFound, body: "nope"},
		"https://down.test": {err: errors.New("refused")},
		"https://big.test":  page(strings.Repeat("x", 11*1024*1024)),
	}})

	body, err := fetchContent("https://ok.test")
	require.NoError(t, err)
	assert.Equal(t, "hello", body)

	_, err = fetchContent("https://gone.test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 404")

	_, err = fetchContent("https://down.test")
	assert.Error(t, err)

	body, err = fetchContent("https://big.test")
	require.NoError(t, err)
	assert.Len(t, body, 10*1024*1024, "reads at most 10 MiB")
}

// ---- resolveDownloadURL / resolveAnimeFireFallbackDownloadURL --------------------------

func TestResolveDownloadURL(t *testing.T) {
	withQuality(t, "best")
	api720 := `{"data":[{"src":"https://cdn.test/360.mp4","label":"360p"},{"src":"https://cdn.test/720.mp4","label":"720p"}]}`

	t.Run("a non-AnimeFire URL is returned as is, with no request", func(t *testing.T) {
		web := &fakeWeb{}
		useFakeWeb(t, web)
		got, err := resolveDownloadURL("https://cdn.test/file.mp4")
		require.NoError(t, err)
		assert.Equal(t, "https://cdn.test/file.mp4", got)
		assert.Empty(t, web.asked)
	})
	t.Run("best picks the highest quality", func(t *testing.T) {
		useFakeWeb(t, &fakeWeb{pages: map[string]fakePage{fireAPI: page(api720)}})
		got, err := resolveDownloadURL(fireAPI)
		require.NoError(t, err)
		assert.Equal(t, "https://cdn.test/720.mp4", got)
	})
	t.Run("worst picks the lowest", func(t *testing.T) {
		withQuality(t, "worst")
		useFakeWeb(t, &fakeWeb{pages: map[string]fakePage{fireAPI: page(api720)}})
		got, err := resolveDownloadURL(fireAPI)
		require.NoError(t, err)
		assert.Equal(t, "https://cdn.test/360.mp4", got)
	})
	t.Run("a named quality picks that one", func(t *testing.T) {
		withQuality(t, "360p")
		useFakeWeb(t, &fakeWeb{pages: map[string]fakePage{fireAPI: page(api720)}})
		got, err := resolveDownloadURL(fireAPI)
		require.NoError(t, err)
		assert.Equal(t, "https://cdn.test/360.mp4", got)
	})
	t.Run("request failure", func(t *testing.T) {
		useFakeWeb(t, &fakeWeb{pages: map[string]fakePage{fireAPI: {err: errors.New("refused")}}})
		_, err := resolveDownloadURL(fireAPI)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch AnimeFire video API")
	})
	t.Run("invalid JSON", func(t *testing.T) {
		useFakeWeb(t, &fakeWeb{pages: map[string]fakePage{fireAPI: page("<html>")}})
		_, err := resolveDownloadURL(fireAPI)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse AnimeFire video API")
	})
}

func TestResolveAnimeFireFallbackDownloadURL(t *testing.T) {
	withQuality(t, "best")
	api3 := `{"data":[{"src":"https://cdn.test/360.mp4","label":"360p"},{"src":"https://cdn.test/720.mp4","label":"720p"},{"src":"https://cdn.test/1080.mp4","label":"1080p"}]}`

	t.Run("skips the source that failed", func(t *testing.T) {
		useFakeWeb(t, &fakeWeb{pages: map[string]fakePage{fireAPI: page(api3)}})
		got, err := resolveAnimeFireFallbackDownloadURL(fireAPI, "https://cdn.test/1080.mp4")
		require.NoError(t, err)
		assert.Equal(t, "https://cdn.test/720.mp4", got, "the next best after the one that failed")
	})
	t.Run("not an AnimeFire API URL", func(t *testing.T) {
		web := &fakeWeb{}
		useFakeWeb(t, web)
		_, err := resolveAnimeFireFallbackDownloadURL("https://other.test/video/x", "x")
		require.Error(t, err)
		assert.Empty(t, web.asked)
	})
	t.Run("no other source", func(t *testing.T) {
		useFakeWeb(t, &fakeWeb{pages: map[string]fakePage{fireAPI: page(`{"data":[{"src":"https://cdn.test/only.mp4","label":"720p"}]}`)}})
		_, err := resolveAnimeFireFallbackDownloadURL(fireAPI, "https://cdn.test/only.mp4")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no fallback source")
	})
	t.Run("request failure", func(t *testing.T) {
		useFakeWeb(t, &fakeWeb{pages: map[string]fakePage{fireAPI: {err: errors.New("refused")}}})
		_, err := resolveAnimeFireFallbackDownloadURL(fireAPI, "x")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch AnimeFire fallback sources")
	})
}

// ---- metatest for the HTTP-backed resolvers ------------------------------------------------

// resolveFingerprint runs a resolver over a fixed fake web and summarises the
// answers, so a broken resolver shows up as a different fingerprint.
func resolveFingerprint(resolve func(string) (string, error)) string {
	prevGet, prevQ := safeGetFn, util.GlobalQuality
	defer func() { safeGetFn, util.GlobalQuality = prevGet, prevQ }()
	web := &fakeWeb{pages: map[string]fakePage{
		fireAPI:                           page(`{"data":[{"src":"https://cdn.test/360.mp4","label":"360p"},{"src":"https://cdn.test/720.mp4","label":"720p"}]}`),
		"https://animefire.io/video/down": {err: errors.New("refused")},
	}}
	safeGetFn = web.get
	var out []string
	for _, q := range []string{"best", "worst", "360p"} {
		util.GlobalQuality = q
		got, err := resolve(fireAPI)
		out = append(out, fmt.Sprintf("%s=%s/%v", q, got, err != nil))
	}
	got, err := resolve("https://animefire.io/video/down")
	out = append(out, fmt.Sprintf("down=%s/%v", got, err != nil))
	got, err = resolve("https://cdn.test/plain.mp4")
	out = append(out, fmt.Sprintf("plain=%s/%v/asked=%d", got, err != nil, len(web.asked)))
	return strings.Join(out, ";")
}

func TestResolveDownloadURLMeta(t *testing.T) {
	want := resolveFingerprint(resolveDownloadURL)
	require.Contains(t, want, "best=https://cdn.test/720.mp4/false")
	mutants := map[string]func(string) (string, error){
		"always the first source": func(u string) (string, error) {
			if !isAnimeFireVideoAPIURL(u) {
				return u, nil
			}
			return "https://cdn.test/360.mp4", nil
		},
		"ignores the quality": func(u string) (string, error) {
			prev := util.GlobalQuality
			util.GlobalQuality = "best"
			defer func() { util.GlobalQuality = prev }()
			return resolveDownloadURL(u)
		},
		"always requests": func(u string) (string, error) {
			if !isAnimeFireVideoAPIURL(u) {
				_, _ = safeGetFn("https://cdn.test/unused")
			}
			return resolveDownloadURL(u)
		},
		"errors swallowed": func(u string) (string, error) {
			got, err := resolveDownloadURL(u)
			if err != nil {
				return u, nil //nolint:nilerr // deliberate mutant: hides the failure
			}
			return got, nil
		},
		"plain URLs rewritten": func(u string) (string, error) {
			got, err := resolveDownloadURL(u)
			if !isAnimeFireVideoAPIURL(u) {
				return got + "?x", err
			}
			return got, err
		},
	}
	for name, m := range mutants {
		assert.NotEqual(t, want, resolveFingerprint(m), "mutant %q survived", name)
	}
}
