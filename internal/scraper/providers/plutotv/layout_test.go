package plutotv

import (
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The failure these tests exist for (2026-11-21, "Jumper"): the stitcher
// served a title in a segment layout this code did not know, the closing ad
// bumper was the one clip in a layout it did know, and the bumper was played
// as the title: five seconds of an animation, then the end of the file.
//
// The lesson is not "learn this layout". The service will change its layout
// again. The lesson is that what Stream hands the player must be the title or
// else the whole stitched playlist, and never a small clip taken for the title.
// So three things are held here: the known layouts are read right (fixtures of
// real answers), an unknown layout is never read as a bumper (mutations of
// those answers and a generated property), and the whole path through Stream
// ends in a playlist that is most of the title or is the stitched master.

var (
	extinfRe = regexp.MustCompile(`(?m)^#EXTINF:([0-9.]+)`)
	// adMarkerRe is anything the service names an ad or a bumper by.
	adMarkerRe = regexp.MustCompile(`(?i)(ad%2F|_ad/|_ad_|bumper|/v1/mp4/|/mp4/v1/)`)
)

// segmentURLs lists the segment lines of a playlist, in order.
func segmentURLs(playlist string) []string {
	var out []string
	for _, line := range strings.Split(strings.ReplaceAll(playlist, "\r\n", "\n"), "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			out = append(out, line)
		}
	}
	return out
}

// seconds is the playlist's total duration.
func seconds(playlist string) float64 {
	total := 0.0
	for _, m := range extinfRe.FindAllStringSubmatch(playlist, -1) {
		d, _ := strconv.ParseFloat(m[1], 64)
		total += d
	}
	return total
}

// titleSegments counts the segments of a stitched playlist that are the
// title's: the ones that are not an ad creative and not the bumper.
func titleSegments(stitched string) int {
	n := 0
	for _, u := range segmentURLs(stitched) {
		if !adMarkerRe.MatchString(u) {
			n++
		}
	}
	return n
}

// assertTitleOnly is the contract of an own playlist read out of a stitched
// one: it has the title's segments, all of them, and nothing else.
func assertTitleOnly(t *testing.T, stitched, own string) {
	t.Helper()
	require.True(t, strings.HasPrefix(own, "#EXTM3U"))
	assert.Contains(t, own, "#EXT-X-ENDLIST")
	assert.NotRegexp(t, adMarkerRe, own, "no ad and no bumper in the title's playlist")
	assert.Equal(t, titleSegments(stitched), len(segmentURLs(own)), "every segment of the title, and only those")
	assert.Equal(t, len(segmentURLs(own)), strings.Count(own, "#EXTINF"), "a duration for each")
	for _, u := range segmentURLs(own) {
		assert.Regexp(t, `^https?://`, u, "absolute")
	}
}

// ---- the real answers ------------------------------------------------------

func TestRealSegLayoutPlaylistIsReadAsTheTitle(t *testing.T) {
	stitched := fixture(t, "stitched_movie_seg_2026_11_21.m3u8")
	require.Contains(t, stitched, "ad_bumper", "the fixture carries the bumper that was mistaken for the title")
	require.Contains(t, stitched, "ad%2Fcreative", "and the ads")
	require.Contains(t, stitched, "/hls/video_pt_video_300/seg_", "in the layout that broke")

	contentURL, ok := contentPlaylistURL(stitched)
	require.True(t, ok, "the title is recognised, not left to the stitched playlist")
	assert.NotContains(t, contentURL, "bumper")
	assert.Contains(t, contentURL, "6ad10c31a586c4b3c9e505a5", "the title's clip, not the bumper's")

	own, ok := titlePlaylist(stitched, "https://stitcher.example/x.m3u8", contentURL)
	require.True(t, ok)
	assertTitleOnly(t, stitched, own)
	assert.Greater(t, len(segmentURLs(own)), 100, "a film, not five seconds")
	assert.Greater(t, seconds(own), 500.0)
	ranges := map[string]bool{}
	for _, u := range segmentURLs(own) {
		if m := contentSegmentV2Re.FindStringSubmatch(u); m != nil {
			ranges[m[2]+"-"+m[3]] = true
		}
	}
	assert.Greater(t, len(ranges), 1, "the title is in several ranges")
	assert.Equal(t, len(ranges)-1, strings.Count(own, "#EXT-X-DISCONTINUITY"), "and the ranges are joined")
	assert.Equal(t, len(segmentURLs(own)), strings.Count(own, "#EXT-X-KEY:"), "every segment keeps its key")
}

// Every real answer kept as a fixture, in every layout: whatever is read out of
// it is the title and all of the title. A layout added later gets its fixture
// in this list and is held to the same thing.
func TestEveryRealStitchedPlaylistYieldsTheWholeTitleOrNothing(t *testing.T) {
	for _, name := range []string{
		"stitched_series_2026_11_16.m3u8",
		"stitched_movie_2026_11_16.m3u8",
		"stitched_movie_pt_2026_11_16.m3u8",
		"stitched_movie_seg_2026_11_21.m3u8",
	} {
		t.Run(name, func(t *testing.T) {
			stitched := fixture(t, name)
			contentURL, ok := contentPlaylistURL(stitched)
			require.True(t, ok)
			own, ok := titlePlaylist(stitched, "https://stitcher.example/x.m3u8", contentURL)
			if !ok {
				return // allowed: the stitched playlist plays whole instead
			}
			assertTitleOnly(t, stitched, own)
		})
	}
}

// ---- a layout nobody has seen ----------------------------------------------

// renameLayout rewrites every segment of the title's clip into a layout this
// code has never been told about, leaving the ads and the bumper as they are:
// exactly what the service did to Jumper, with a different name.
func renameLayout(stitched string, layout func(i int, u string) string) string {
	lines := strings.Split(stitched, "\n")
	i := 0
	for n, l := range lines {
		if l != "" && !strings.HasPrefix(l, "#") && !adMarkerRe.MatchString(l) {
			lines[n] = layout(i, l)
			i++
		}
	}
	return strings.Join(lines, "\n")
}

func TestAnUnknownLayoutIsNeverReadAsTheBumper(t *testing.T) {
	stitched := fixture(t, "stitched_movie_seg_2026_11_21.m3u8")
	layouts := map[string]func(int, string) string{
		"flat numbered parts": func(i int, u string) string {
			return fmt.Sprintf("https://siloh-ns1.plutotv.net/529_pluto/clip/T/720p/part-%05d.ts", i)
		},
		"a query string": func(i int, u string) string {
			return fmt.Sprintf("https://siloh-ns1.plutotv.net/529_pluto/clip/T/hls/0-end/hls_600-%05d.ts?x=1", i)
		},
		"a relative path": func(i int, u string) string { return fmt.Sprintf("hls/0-end/hls_600-%05d.ts", i) },
		"fragmented mp4": func(i int, u string) string {
			return fmt.Sprintf("https://siloh-ns1.plutotv.net/529_pluto/clip/T/seg_%05d.m4s", i)
		},
		"the title's range name": func(i int, u string) string { return strings.Replace(u, "/hls/video_", "/stream/video_", 1) },
	}
	for name, layout := range layouts {
		t.Run(name, func(t *testing.T) {
			mutated := renameLayout(stitched, layout)
			require.NotEqual(t, stitched, mutated)
			require.Contains(t, mutated, "ad_bumper", "the bumper is still there, in the layout that is known")

			got, ok := contentPlaylistURL(mutated)
			assert.False(t, ok, "the bumper must not be taken for the title (got %q)", got)
			assert.Empty(t, got)
		})
	}
}

// Through Stream: an unknown layout plays the stitched master, which has the
// whole title and the ads, and carries no playlist text of its own.
func TestStreamOfAnUnknownLayoutPlaysTheWholeStitchedPlaylist(t *testing.T) {
	f := newFakePluto(t)
	f.serveEpisode(fakeEpisodeID)
	unknown := renameLayout(f.stitched("hls_2400", 12, "0-100", "101-200", "201-end"), func(i int, _ string) string {
		return fmt.Sprintf("%s/media/clip/abc_Title/720p/part-%05d.ts", f.url(), i)
	})
	unknown = strings.ReplaceAll(unknown, "hls_bumper-", "hls_300-") // a bumper as the real service names one
	require.Contains(t, unknown, "hls_300-00000.ts", "the bumper is in the layout that is known")
	f.set(f.variantPath(fakeEpisodeID, 3321280), unknown)

	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	assert.False(t, stream.AdFree, "the result says the breaks are still in it")
	assert.Empty(t, stream.Playlist, "no playlist of ours: a one-clip playlist would be the bumper")
	assert.Contains(t, stream.URL, "/v2/stitch/hls/episode/"+fakeEpisodeID+"/master.m3u8")
}

// ---- the layout that broke, through Stream ---------------------------------

// stitchedSeg is a stitched playlist in the seg layout, `perRange` segments in
// each range with an ad between, and the closing bumper in the older layout.
func (f *fakePluto) stitchedSeg(perRange int, ranges ...string) string {
	var b strings.Builder
	b.WriteString("#EXTM3U\n#EXT-X-VERSION:5\n#EXT-X-PLAYLIST-TYPE:VOD\n#EXT-X-TARGETDURATION:6\n#EXT-X-MEDIA-SEQUENCE:0\n")
	for i, r := range ranges {
		if i > 0 {
			b.WriteString("#EXT-X-DISCONTINUITY\n#EXT-X-KEY:METHOD=NONE\n")
			fmt.Fprintf(&b, "#EXTINF:6,\n%s/v1/mp4/c(ts)/p(%d_ad%%2Fcreative%%2Fx%%2Fvideo_300.mp4)\n#EXT-X-DISCONTINUITY\n", f.url(), i)
		}
		for s := 0; s < perRange; s++ {
			fmt.Fprintf(&b, "#EXT-X-KEY:METHOD=AES-128,URI=\"%s%s%s/hls/video_pt_video_300/seg_keyfile_0.key\",IV=0x%032x\n", f.url(), fakeClipRoot, r, s+1)
			fmt.Fprintf(&b, "#EXTINF:6,\n%s%s%s/hls/video_pt_video_300/seg_%05d.ts\n", f.url(), fakeClipRoot, r, s)
		}
	}
	fmt.Fprintf(&b, "#EXT-X-DISCONTINUITY\n#EXT-X-KEY:METHOD=NONE\n#EXTINF:5,\n%s/media/clip/x_ad_bumper_y/1080pDRM/hls/0-end/hls_300-00000.ts\n#EXT-X-ENDLIST\n", f.url())
	return b.String()
}

// fakeClipRoot is a clip of the seg layout on the fake CDN.
const fakeClipRoot = "/media/clip/abc_Title/720pDRM/20261015/"

func TestStreamOfTheSegLayoutPlaysTheTitleAndNotTheBumper(t *testing.T) {
	f := newFakePluto(t)
	f.serveEpisode(fakeEpisodeID)
	stitched := f.stitchedSeg(40, "0-100", "100-200", "200-end")
	f.set(f.variantPath(fakeEpisodeID, 3321280), stitched)
	// What the CDN did to Jumper: the title's own playlist is refused, under
	// every name it might have.
	f.setRoute(f.contentPath("hls_2400"), http.StatusForbidden, "Forbidden")

	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	require.True(t, stream.AdFree, "the title alone was found")
	assertTitleOnly(t, stitched, stream.Playlist)
	assert.Equal(t, 120, len(segmentURLs(stream.Playlist)))
	assert.InDelta(t, 120*6, seconds(stream.Playlist), 0.01)

	// And what the player is handed is that text, from loopback.
	local := PlayURL(stream)
	assert.True(t, strings.HasPrefix(local, "http://127.0.0.1:"), local)
	status, _, body := fetchLocal(t, http.MethodGet, local)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, strings.TrimLeft(stream.Playlist, "\xef\xbb\xbf \t\r\n"), body)
}

// ---- the property ----------------------------------------------------------

// Whatever the layout, the playlist of our own that Stream may hand over is
// the title's: never a clip that is a sliver of what the stitcher listed, never
// one named for an ad. Generated: layouts the code knows and does not, titles
// short and long, ads and bumpers in every placement.
func TestPropertyWhatIsReadAsTheTitleIsTheTitle(t *testing.T) {
	titleLayouts := []func(dir, rng string, i int) string{
		func(dir, rng string, i int) string { return fmt.Sprintf("%s/hls/%s/hls_600-%05d.ts", dir, rng, i) },
		func(dir, rng string, i int) string {
			return fmt.Sprintf("%s/%s/hls/video_pt_video_300/seg_%05d.ts", dir, rng, i)
		},
		func(dir, rng string, i int) string {
			return fmt.Sprintf("%s/%s/hls/video_en_video_600/seg_%05d.ts", dir, rng, i)
		},
		func(dir, rng string, i int) string { return fmt.Sprintf("%s/part-%s-%05d.ts", dir, rng, i) },
		func(dir, rng string, i int) string { return fmt.Sprintf("%s/%s/%05d.m4s", dir, rng, i) },
		func(dir, rng string, i int) string {
			return fmt.Sprintf("%s/hls/%s/hls_600-%05d.ts?token=%d", dir, rng, i, i)
		},
	}
	bumperLayouts := []func(i int) string{
		func(i int) string {
			return fmt.Sprintf("https://cdn.example/259_O/clip/x_ptv_ad_bumper_y/1080pDRM/hls/0-end/hls_300-%05d.ts", i)
		},
		func(i int) string {
			return fmt.Sprintf("https://cdn.example/259_O/clip/bumper/hls/0-end/hls_300-%05d.ts", i)
		},
		func(i int) string {
			return fmt.Sprintf("https://cdn.example/259_O/clip/intro/hls/0-end/hls_300-%05d.ts", i)
		},
	}
	rnd := rand.New(rand.NewSource(20261121))
	for n := range 600 {
		layout := titleLayouts[rnd.Intn(len(titleLayouts))]
		perRange, ranges := 1+rnd.Intn(80), 1+rnd.Intn(8)
		// A title is longer than twice the longest bumper: with fewer segments
		// than that, a title in a layout nobody knows cannot be told from one.
		for perRange*ranges < 4*maxInterstitialSegments {
			perRange++
		}
		var b strings.Builder
		b.WriteString("#EXTM3U\n#EXT-X-VERSION:5\n")
		titleCount, edge := 0, 0
		for r := range ranges {
			from, to := strconv.Itoa(edge), strconv.Itoa(edge+1000)
			if r == 0 {
				from = "0"
			}
			if r == ranges-1 {
				to = "end"
			}
			edge += 1000
			if r > 0 {
				for a := range 1 + rnd.Intn(6) {
					fmt.Fprintf(&b, "#EXT-X-DISCONTINUITY\n#EXTINF:6,\nhttps://cdn.example/v1/mp4/c(ts)/p(%d_ad%%2Fcreative%%2F%d%%2Fvideo.mp4)/frag(%d)/s\n", r, a, a)
				}
			}
			for i := range perRange {
				fmt.Fprintf(&b, "#EXTINF:6,\n%s\n", layout("https://cdn.example/529_p/clip/T/720p/x_1", from+"-"+to, i))
				titleCount++
			}
		}
		bumper := bumperLayouts[rnd.Intn(len(bumperLayouts))]
		for i := range rnd.Intn(maxInterstitialSegments + 1) {
			fmt.Fprintf(&b, "#EXT-X-DISCONTINUITY\n#EXTINF:5,\n%s\n", bumper(i))
		}
		b.WriteString("#EXT-X-ENDLIST\n")
		stitched := b.String()

		contentURL, ok := contentPlaylistURL(stitched)
		if !ok {
			continue // the stitched playlist plays whole: always allowed
		}
		own, ok := titlePlaylist(stitched, "https://cdn.example/s.m3u8", contentURL)
		if !ok {
			continue
		}
		got := len(segmentURLs(own))
		if got != titleCount || adMarkerRe.MatchString(own) {
			t.Fatalf("case %d: the title has %d segments, its own playlist has %d (or carries an ad)\ncontent URL %s", n, titleCount, got, contentURL)
		}
	}
}

// The worst way to be wrong, said once on its own: a clip that is a sliver of
// the playlist never wins the title, whatever layout it is in.
func TestAClipThatIsASliverOfThePlaylistIsNotTheTitle(t *testing.T) {
	unknown := strings.Repeat("#EXTINF:6,\nhttps://cdn.example/clip/T/part-1.ts\n", 500)
	bumperOnly := "#EXTINF:5,\nhttps://cdn.example/259/clip/Q/hls/0-end/hls_300-00000.ts\n"
	_, ok := contentPlaylistURL(unknown + bumperOnly)
	assert.False(t, ok)

	// And the same clip, alone, is a title of one segment: short, but whole.
	_, ok = contentPlaylistURL(bumperOnly)
	assert.True(t, ok, "a playlist of one clip is that clip")
	// Unless the service names it an ad.
	_, ok = contentPlaylistURL("#EXTINF:5,\nhttps://cdn.example/259/clip/x_ptv_ad_bumper_y/hls/0-end/hls_300-00000.ts\n")
	assert.False(t, ok, "a bumper is never the title, even when it is all there is")
}
