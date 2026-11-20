package player

import (
	"errors"
	"fmt"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pickMock records what the quality menu was shown and answers with a fixed
// choice.
type pickMock struct {
	idx    int
	err    error
	calls  int
	labels []string
	res    []int
}

func (p *pickMock) fn(labels []string, res []int) (int, error) {
	p.calls++
	p.labels, p.res = labels, res
	return p.idx, p.err
}

const threeQualities = `{"data":[` +
	`{"src":"https://cdn.test/360.mp4","label":"360p"},` +
	`{"src":"https://cdn.test/1080.mp4","label":"FHD 1080p"},` +
	`{"src":"https://cdn.test/720.mp4","label":"720p"}]}`

type bodyCase struct {
	name      string
	body      string
	quality   string
	pick      pickMock
	want      animeFireResolution
	wantErr   error
	wantErrIs string
	wantAsked bool
}

var bodyCases = []bodyCase{
	{name: "one quality needs no choice", body: `{"data":[{"src":"https://cdn.test/a.mp4","label":"480p"}]}`, quality: "best",
		want: animeFireResolution{src: "https://cdn.test/a.mp4"}},
	{name: "a session quality picks without asking", body: threeQualities, quality: "720p",
		want: animeFireResolution{src: "https://cdn.test/720.mp4"}},
	{name: "worst picks the lowest", body: threeQualities, quality: "worst",
		want: animeFireResolution{src: "https://cdn.test/360.mp4"}},
	{name: "best asks the user and remembers the pick", body: threeQualities, quality: "best", pick: pickMock{idx: 1}, wantAsked: true,
		want: animeFireResolution{src: "https://cdn.test/720.mp4", remember: "720p"}},
	{name: "no quality yet asks and remembers", body: threeQualities, quality: "", pick: pickMock{idx: 0}, wantAsked: true,
		want: animeFireResolution{src: "https://cdn.test/1080.mp4", remember: "1080p"}},
	{name: "the menu is sorted highest first", body: threeQualities, quality: "best", pick: pickMock{idx: 2}, wantAsked: true,
		want: animeFireResolution{src: "https://cdn.test/360.mp4", remember: "360p"}},
	{name: "back leaves the quality menu", body: threeQualities, quality: "best", pick: pickMock{err: tui.ErrPickBack}, wantAsked: true,
		wantErr: ErrBackRequested},
	{name: "cancel leaves the quality menu", body: threeQualities, quality: "best", pick: pickMock{err: tui.ErrPickCancelled}, wantAsked: true,
		wantErr: ErrBackRequested},
	{name: "a menu failure is reported", body: threeQualities, quality: "best", pick: pickMock{err: errors.New("tty gone")}, wantAsked: true,
		wantErrIs: "failed to select quality"},
	{name: "index past the end", body: threeQualities, quality: "best", pick: pickMock{idx: 3}, wantAsked: true,
		wantErrIs: "out of range"},
	{name: "negative index", body: threeQualities, quality: "best", pick: pickMock{idx: -1}, wantAsked: true,
		wantErrIs: "out of range"},
	{name: "a session quality with no exact match takes the nearest", body: threeQualities, quality: "4320p",
		want: animeFireResolution{src: "https://cdn.test/1080.mp4"}},
	{name: "a session quality that matches nothing takes the first source", body: threeQualities, quality: "dublado",
		want: animeFireResolution{src: "https://cdn.test/360.mp4"}},
	{name: "empty list with a blogger token", body: `{"data":[],"token":"https://www.blogger.com/video.g?token=abc"}`, quality: "best",
		want: animeFireResolution{blogger: "https://www.blogger.com/video.g?token=abc"}},
	{name: "empty list with another token", body: `{"data":[],"token":"https://elsewhere.test/x"}`, quality: "best",
		want: animeFireResolution{}},
	{name: "direct video link in a non-JSON body", body: `<html><script>play("https://cdn.test/v/e.mp4")</script></html>`, quality: "best",
		want: animeFireResolution{src: "https://cdn.test/v/e.mp4"}},
	{name: "blogger link in a non-JSON body", body: `<html>https://www.blogger.com/video.g?token=Zz_9-a</html>`, quality: "best",
		want: animeFireResolution{blogger: "https://www.blogger.com/video.g?token=Zz_9-a"}},
	{name: "direct link wins over a blogger link", body: `https://cdn.test/v/e.m3u8 https://www.blogger.com/video.g?token=abc`, quality: "best",
		want: animeFireResolution{src: "https://cdn.test/v/e.m3u8"}},
	{name: "nothing usable", body: `<html>blocked</html>`, quality: "best", want: animeFireResolution{}},
	{name: "empty body", body: ``, quality: "best", want: animeFireResolution{}},
}

func bodyFailures(impl func([]byte, string, qualityPicker) (animeFireResolution, error)) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	for _, tc := range bodyCases {
		p := tc.pick
		got, err := impl([]byte(tc.body), tc.quality, p.fn)
		switch {
		case tc.wantErr != nil:
			if !errors.Is(err, tc.wantErr) {
				f = append(f, fmt.Sprintf("%s: error %v", tc.name, err))
			}
		case tc.wantErrIs != "":
			if err == nil || !containsStr(err.Error(), tc.wantErrIs) {
				f = append(f, fmt.Sprintf("%s: error %v", tc.name, err))
			}
		default:
			if err != nil || got != tc.want {
				f = append(f, fmt.Sprintf("%s: got (%+v,%v) want %+v", tc.name, got, err, tc.want))
			}
		}
		if (p.calls > 0) != tc.wantAsked {
			f = append(f, fmt.Sprintf("%s: asked=%v want %v", tc.name, p.calls > 0, tc.wantAsked))
		}
	}
	return f
}

func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestResolveAnimeFireBody(t *testing.T) {
	assert.Empty(t, bodyFailures(resolveAnimeFireBody))
}

func TestResolveAnimeFireBodyShowsTheMenuSortedWithMirrorLabels(t *testing.T) {
	p := &pickMock{idx: 0}
	_, err := resolveAnimeFireBody([]byte(threeQualities), "best", p.fn)
	require.NoError(t, err)
	assert.Equal(t, []string{"1080p", "720p", "360p"}, p.labels)
	assert.Equal(t, []int{1080, 720, 360}, p.res)

	p = &pickMock{idx: 1}
	dup := `{"data":[{"src":"a","label":"720p"},{"src":"b","label":"720p"}]}`
	got, err := resolveAnimeFireBody([]byte(dup), "best", p.fn)
	require.NoError(t, err)
	assert.Equal(t, []string{"720p", "720p (mirror 2)"}, p.labels)
	assert.Equal(t, "b", got.src, "the index lines up with the labels")
}

func TestResolveAnimeFireBodyRemembersUnparsableLabels(t *testing.T) {
	body := `{"data":[{"src":"a","label":"  Dublado HD "},{"src":"b","label":"Legendado"}]}`
	got, err := resolveAnimeFireBody([]byte(body), "best", (&pickMock{idx: 0}).fn)
	require.NoError(t, err)
	assert.Equal(t, "a", got.src)
	assert.Equal(t, "dublado hd", got.remember, "no resolution to canonicalise: the trimmed lowercase label")
}

func TestResolveAnimeFireBodyChoiceIsNotRememberedOverASessionQuality(t *testing.T) {
	got, err := resolveAnimeFireBody([]byte(threeQualities), "dublado", (&pickMock{idx: 0}).fn)
	require.NoError(t, err)
	assert.Empty(t, got.remember, "a quality the user already set is not overwritten")
}

func TestResolveAnimeFireBodyMeta(t *testing.T) {
	require.Empty(t, bodyFailures(resolveAnimeFireBody))
	type fn = func([]byte, string, qualityPicker) (animeFireResolution, error)
	mutants := map[string]fn{
		"always asks": func(b []byte, q string, p qualityPicker) (animeFireResolution, error) {
			return resolveAnimeFireBody(b, "best", p)
		},
		"never asks": func(b []byte, q string, p qualityPicker) (animeFireResolution, error) {
			return resolveAnimeFireBody(b, q, func([]string, []int) (int, error) { return 0, nil })
		},
		"does not remember": func(b []byte, q string, p qualityPicker) (animeFireResolution, error) {
			r, err := resolveAnimeFireBody(b, q, p)
			r.remember = ""
			return r, err
		},
		"remembers over a session quality": func(b []byte, q string, p qualityPicker) (animeFireResolution, error) {
			r, err := resolveAnimeFireBody(b, q, p)
			if r.src != "" && r.remember == "" && q != "" && q != "best" && p != nil {
				r.remember = "x"
			}
			return r, err
		},
		"back is an error": func(b []byte, q string, p qualityPicker) (animeFireResolution, error) {
			r, err := resolveAnimeFireBody(b, q, p)
			if errors.Is(err, ErrBackRequested) {
				return r, errors.New("failed to select quality")
			}
			return r, err
		},
		"unchecked index": func(b []byte, q string, p qualityPicker) (animeFireResolution, error) {
			r, err := resolveAnimeFireBody(b, q, p)
			if err != nil && containsStr(err.Error(), "out of range") {
				return animeFireResolution{src: "x"}, nil
			}
			return r, err
		},
		"blogger token ignored": func(b []byte, q string, p qualityPicker) (animeFireResolution, error) {
			r, err := resolveAnimeFireBody(b, q, p)
			r.blogger = ""
			return r, err
		},
		"blogger and src swapped": func(b []byte, q string, p qualityPicker) (animeFireResolution, error) {
			r, err := resolveAnimeFireBody(b, q, p)
			r.blogger, r.src = r.src, r.blogger
			return r, err
		},
		"single quality still asks": func(b []byte, q string, p qualityPicker) (animeFireResolution, error) {
			r, err := resolveAnimeFireBody(b, q, p)
			if r.src != "" && p != nil {
				_, _ = p(nil, nil)
			}
			return r, err
		},
		"fallbacks dropped": func(b []byte, q string, p qualityPicker) (animeFireResolution, error) {
			r, err := resolveAnimeFireBody(b, q, p)
			if len(b) > 0 && b[0] != '{' {
				return animeFireResolution{}, err
			}
			return r, err
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, bodyFailures(m), "mutant %q survived", name)
	}
}

// extractActualVideoURL end to end over the fake web: the referer AnimeFire's
// CDN demands is set, and a single quality is returned untouched.
func TestExtractActualVideoURLAnimeFire(t *testing.T) {
	restore := snapshotGlobalReferer()
	defer restore()
	util.ClearGlobalReferer()
	withQuality(t, "best")

	useFakeWeb(t, &fakeWeb{pages: map[string]fakePage{
		fireAPI:                          page(`{"data":[{"src":"https://cdn.test/only.mp4","label":"720p"}]}`),
		"https://animefire.io/video/bad": page(`<html>blocked</html>`),
		"https://animefire.io/video/dn":  {err: errors.New("refused")},
	}})

	got, err := extractActualVideoURL(fireAPI)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.test/only.mp4", got)
	assert.Equal(t, "https://animefire.io", util.GetGlobalReferer(), "the CDN rejects requests without it")

	_, err = extractActualVideoURL("https://animefire.io/video/bad")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid video URL found")

	_, err = extractActualVideoURL("https://animefire.io/video/dn")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch video page")

	_, err = extractActualVideoURL("https://other.test/page")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid video URL found", "a source this function does not know ends in the same error")
}

func TestExtractActualVideoURLKeepsAnExistingReferer(t *testing.T) {
	restore := snapshotGlobalReferer()
	defer restore()
	util.SetGlobalReferer("https://already.set")
	withQuality(t, "best")
	useFakeWeb(t, &fakeWeb{pages: map[string]fakePage{fireAPI: page(`{"data":[{"src":"https://cdn.test/o.mp4","label":"720p"}]}`)}})

	_, err := extractActualVideoURL(fireAPI)
	require.NoError(t, err)
	assert.Equal(t, "https://already.set", util.GetGlobalReferer())
}
