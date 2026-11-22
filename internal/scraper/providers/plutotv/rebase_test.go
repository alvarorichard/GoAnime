package plutotv

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The titles that broke seeking (Jumper, Super, O Monge à Prova de Balas,
// 2026-11-21): each range a file of its own whose clock restarts at one second.
// These tests build that title, and its well-behaved twin whose clock runs on,
// on the fake service, and hold what the player is handed to a timeline a seek
// can use.

const (
	segTicks   = 6 * 90000 // the duration stitchedSeg gives every segment
	clockStart = 90000     // where a range's first segment starts its clock
)

var testRanges = []string{"0-100", "100-200", "200-end"}

func ivOf(n int) (iv [16]byte) {
	iv[15], iv[14] = byte(n), byte(n>>8)
	return
}

func keyOf(rng int) []byte { return []byte(fmt.Sprintf("range-key-%05d!", rng))[:16] }

// serveSegMedia serves the seg-layout title of stitchedSeg: its stitched
// playlist, its keys and its encrypted segments. With restart each range
// counts from clockStart again; without, the clock runs on across them.
func (f *fakePluto) serveSegMedia(perRange int, restart bool) string {
	f.t.Helper()
	f.serveEpisode(fakeEpisodeID)
	stitched := f.stitchedSeg(perRange, testRanges...)
	f.set(f.variantPath(fakeEpisodeID, 3321280), stitched)
	f.setRoute(f.contentPath("hls_2400"), http.StatusForbidden, "Forbidden")
	for r, rng := range testRanges {
		f.set(fakeClipRoot+rng+"/hls/video_pt_video_300/seg_keyfile_0.key", string(keyOf(r)))
		for s := range perRange {
			start := int64(clockStart) + int64(s)*segTicks
			if !restart {
				start += int64(r*perRange) * segTicks
			}
			enc := encrypt(tsSegment(start, 12+s), keyOf(r), ivOf(s+1))
			f.set(fmt.Sprintf("%s%s/hls/video_pt_video_300/seg_%05d.ts", fakeClipRoot, rng, s), string(enc))
		}
	}
	return stitched
}

// ---- reading the playlist ---------------------------------------------------

func TestParseTitlePlaylist(t *testing.T) {
	const own = "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:6\n" +
		"#EXT-X-KEY:METHOD=AES-128,URI=\"https://c.example/k0.key\",IV=0x00000000000000000000000000000001\n#EXTINF:6,\nhttps://c.example/a.ts\n" +
		"#EXT-X-KEY:METHOD=AES-128,URI=\"https://c.example/k0.key\",IV=0xbd\n#EXTINF:5.5,\nhttps://c.example/b.ts\n" +
		"#EXT-X-DISCONTINUITY\n#EXT-X-KEY:METHOD=NONE\n#EXTINF:6,\nhttps://c.example/c.ts\n#EXT-X-ENDLIST\n"
	header, segs, err := parseTitlePlaylist(own)
	require.NoError(t, err)
	assert.Equal(t, "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:6\n", header)
	require.Len(t, segs, 3)
	assert.Equal(t, "https://c.example/k0.key", segs[0].keyURL)
	assert.Equal(t, ivOf(1), segs[0].iv)
	assert.Equal(t, ivOf(0xbd), segs[1].iv, "a short IV is the low bytes")
	assert.Equal(t, []int{0, 0, 1}, []int{segs[0].group, segs[1].group, segs[2].group})
	assert.Empty(t, segs[2].keyURL, "METHOD=NONE is no key")
	assert.InDelta(t, 5.5, extinfSeconds(segs[1].extinf), 0)

	// A tag after the first segment is the segment's, not the header's.
	header, _, err = parseTitlePlaylist("#EXTM3U\n#EXTINF:6,\nhttps://c.example/a.ts\n#EXT-X-PROGRAM-DATE-TIME:2026-11-21T00:00:00Z\n#EXTINF:6,\nhttps://c.example/b.ts\n")
	require.NoError(t, err)
	assert.Equal(t, "#EXTM3U\n", header)

	for name, bad := range map[string]string{
		"a second segment with no duration of its own": "#EXTM3U\n#EXTINF:6,\nhttps://c.example/a.ts\nhttps://c.example/b.ts\n",
		"no segments":       "#EXTM3U\n#EXT-X-ENDLIST\n",
		"no duration":       "#EXTM3U\nhttps://c.example/a.ts\n",
		"a key with no IV":  "#EXTM3U\n#EXT-X-KEY:METHOD=AES-128,URI=\"https://c.example/k\"\n#EXTINF:6,\nhttps://c.example/a.ts\n",
		"a key with no URI": "#EXTM3U\n#EXT-X-KEY:METHOD=AES-128,IV=0x01\n#EXTINF:6,\nhttps://c.example/a.ts\n",
	} {
		_, _, err := parseTitlePlaylist(bad)
		assert.Error(t, err, name)
	}
}

func TestRenderAddressesTheSegmentsUnderTheID(t *testing.T) {
	plan := &rebasePlan{header: "#EXTM3U\n", segs: []segmentRef{
		{extinf: "#EXTINF:6,", group: 0}, {extinf: "#EXTINF:6,", group: 0}, {extinf: "#EXTINF:5,", group: 1},
	}}
	assert.Equal(t, "#EXTM3U\n#EXTINF:6,\nabc/0.ts\n#EXTINF:6,\nabc/1.ts\n#EXT-X-DISCONTINUITY\n#EXTINF:5,\nabc/2.ts\n#EXT-X-ENDLIST\n", plan.render("abc"))
}

// ---- deciding ---------------------------------------------------------------

func TestPlanRebaseOfRangesThatRestartTheirClock(t *testing.T) {
	f := newFakePluto(t)
	f.serveSegMedia(10, true)
	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	require.True(t, stream.AdFree)
	plan := stream.rebase
	require.NotNil(t, plan, "ranges that each restart at one second are put on one timeline")
	require.Len(t, plan.segs, 30)

	// Range r starts where the durations before it say: r*10 segments in.
	for i, s := range plan.segs {
		r := i / 10
		assert.EqualValues(t, int64(r*10)*segTicks, s.shift, "segment %d", i)
	}
	assert.NotContains(t, plan.render("x"), "EXT-X-KEY", "what is served is not encrypted")
	assert.Equal(t, 2, strings.Count(plan.render("x"), "#EXT-X-DISCONTINUITY"))
	assert.Equal(t, 30, strings.Count(stream.Playlist, "#EXTINF"), "the absolute playlist is still there for whoever wants it")
}

func TestPlanRebaseLeavesATitleWhoseClockRunsOnAlone(t *testing.T) {
	f := newFakePluto(t)
	f.serveSegMedia(10, false)
	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	assert.True(t, stream.AdFree)
	assert.Nil(t, stream.rebase, "nothing to put right: the CDN's segments are played as they are")
	assert.Equal(t, 30, strings.Count(stream.Playlist, "#EXTINF"))
}

func TestPlanRebaseToleratesTheDriftOfRoundedDurations(t *testing.T) {
	// The playlist says 6 s a segment, the encode makes a little more: by the
	// last range the clock is seconds off the playlist's, which is not a
	// restart and must not be "put right".
	f := newFakePluto(t)
	f.serveSegMedia(10, false)
	for r, rng := range testRanges {
		for s := range 10 {
			start := int64(clockStart) + int64(r*10+s)*segTicks + int64(r*10+s)*9000 // 100 ms a segment: 2 s off by the last range
			enc := encrypt(tsSegment(start, 12+s), keyOf(r), ivOf(s+1))
			f.set(fmt.Sprintf("%s%s/hls/video_pt_video_300/seg_%05d.ts", fakeClipRoot, rng, s), string(enc))
		}
	}
	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	assert.Nil(t, stream.rebase)
}

func TestPlanRebaseNeverBreaksPlayback(t *testing.T) {
	cases := map[string]func(f *fakePluto){
		"a probe is refused": func(f *fakePluto) {
			f.setRoute(fakeClipRoot+"100-200/hls/video_pt_video_300/seg_00000.ts", http.StatusForbidden, "no")
		},
		"a key is gone": func(f *fakePluto) {
			f.setRoute(fakeClipRoot+"200-end/hls/video_pt_video_300/seg_keyfile_0.key", http.StatusNotFound, "no")
		},
		"a range is not a transport stream": func(f *fakePluto) {
			f.set(fakeClipRoot+"100-200/hls/video_pt_video_300/seg_00000.ts", string(encrypt([]byte("not ts at all, not ts at all....."), keyOf(1), ivOf(1))))
		},
	}
	for name, break_ := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFakePluto(t)
			f.serveSegMedia(10, true)
			break_(f)
			stream, err := f.client().Stream(t.Context(), episodeURL, "best")
			require.NoError(t, err, "the title still resolves")
			assert.True(t, stream.AdFree)
			assert.Nil(t, stream.rebase, "and plays as the CDN has it")
			assert.Equal(t, 30, strings.Count(stream.Playlist, "#EXTINF"))
			assert.True(t, strings.HasPrefix(PlayURL(stream), "http://127.0.0.1:"), "from its own playlist")
		})
	}
}

func TestPlanRebaseRefusesAHostThatIsNotPlutos(t *testing.T) {
	spy := newSpyHost(t, "x")
	f := newFakePluto(t)
	plan, err := f.client().planRebase(t.Context(), "#EXTM3U\n#EXTINF:6,\n"+spy.URL+"/a.ts\n#EXT-X-DISCONTINUITY\n#EXTINF:6,\n"+spy.URL+"/b.ts\n")
	assert.Nil(t, plan)
	assert.ErrorContains(t, err, "segment on a host that is not Pluto TV's", "refused before any probe is started, not by the fetch that would follow")
	assert.Zero(t, spy.hits.Load(), "nothing was asked of it")

	// A key on another host is refused the same way.
	_, err = f.client().planRebase(t.Context(), "#EXTM3U\n#EXT-X-KEY:METHOD=AES-128,URI=\""+spy.URL+"/k.key\",IV=0x01\n#EXTINF:6,\n"+f.url()+"/media/a.ts\n#EXT-X-DISCONTINUITY\n#EXTINF:6,\n"+f.url()+"/media/b.ts\n")
	assert.ErrorContains(t, err, "segment on a host that is not Pluto TV's")
}

func TestPlanRebaseOfOneRangeHasNoOtherToDisagreeWith(t *testing.T) {
	f := newFakePluto(t)
	plan, err := f.client().planRebase(t.Context(), "#EXTM3U\n#EXTINF:6,\n"+f.url()+"/media/a.ts\n")
	assert.NoError(t, err)
	assert.Nil(t, plan)
	assert.Zero(t, f.hitCount("/media/a.ts"), "and nothing was fetched to find out")
}

func TestProbeReadsOnlyThePrefixOfASegment(t *testing.T) {
	f := newFakePluto(t)
	f.serveSegMedia(10, true)
	var sent atomic.Int64
	f.onHit = func(path string) {
		if strings.HasSuffix(path, "seg_00000.ts") {
			sent.Add(1)
		}
	}
	_, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	r := f.lastRequest(fakeClipRoot + "100-200/hls/video_pt_video_300/seg_00000.ts")
	require.NotNil(t, r)
	assert.Equal(t, fmt.Sprintf("bytes=0-%d", probeBytes-1), r.Header.Get("Range"))
	assert.EqualValues(t, 3, sent.Load(), "one probe a range, and no more")
}

// ---- serving ----------------------------------------------------------------

func get(t *testing.T, target string, hdr ...string) (int, http.Header, []byte) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, target, http.NoBody)
	require.NoError(t, err)
	for i := 0; i+1 < len(hdr); i += 2 {
		req.Header.Set(hdr[i], hdr[i+1])
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, resp.Header, body
}

func TestARebasedTitleIsServedOnOneTimeline(t *testing.T) {
	f := newFakePluto(t)
	f.serveSegMedia(10, true)
	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	local := PlayURL(stream)
	require.True(t, strings.HasPrefix(local, "http://127.0.0.1:"), local)

	status, hdr, body := get(t, local)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, "application/vnd.apple.mpegurl", hdr.Get("Content-Type"))
	playlist := string(body)
	require.Len(t, segmentURLs(playlist), 30)
	assert.NotContains(t, playlist, "EXT-X-KEY")
	assert.NotContains(t, playlist, f.url(), "segments come from here, not from the CDN")

	// Every segment, resolved the way a player does, is a transport stream
	// whose first timestamp is where the playlist's durations put it.
	base := local[:strings.LastIndex(local, "/")+1]
	for i, rel := range segmentURLs(playlist) {
		status, hdr, seg := get(t, base+rel)
		require.Equal(t, http.StatusOK, status, rel)
		assert.Equal(t, "video/mp2t", hdr.Get("Content-Type"))
		require.NoError(t, checkTS(seg), rel)
		at, ok := firstTimestamp(seg)
		require.True(t, ok)
		assert.EqualValues(t, clockStart+int64(i)*segTicks, at, "segment %d starts where the title's timeline says", i)
	}
}

func TestARebasedSegmentAnswersRangeAndHeadRequests(t *testing.T) {
	f := newFakePluto(t)
	f.serveSegMedia(10, true)
	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	local := PlayURL(stream)
	seg := local[:len(local)-len(".m3u8")] + "/12.ts"

	_, _, whole := get(t, seg)
	status, hdr, part := get(t, seg, "Range", "bytes=188-375")
	assert.Equal(t, http.StatusPartialContent, status)
	assert.Equal(t, whole[188:376], part)
	assert.Equal(t, fmt.Sprintf("bytes 188-375/%d", len(whole)), hdr.Get("Content-Range"))

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodHead, seg, http.NoBody)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.EqualValues(t, len(whole), resp.ContentLength)
}

func TestARebasedTitleServesNothingElse(t *testing.T) {
	f := newFakePluto(t)
	f.serveSegMedia(10, true)
	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	local := PlayURL(stream)
	id := strings.TrimSuffix(local[strings.LastIndex(local, "/")+1:], ".m3u8")
	base := local[:strings.LastIndex(local, "/")+1]
	for name, target := range map[string]string{
		"past the last segment":  base + id + "/30.ts",
		"a negative segment":     base + id + "/-1.ts",
		"not a number":           base + id + "/x.ts",
		"not a segment":          base + id + "/1.mp4",
		"a number with no .ts":   base + id + "/1",
		"a title never made":     base + strings.Repeat("0", 32) + "/1.ts",
		"a path inside":          base + id + "/1/2.ts",
		"the id with a trailing": base + id + "/",
	} {
		status, _, _ := get(t, target)
		assert.Equal(t, http.StatusNotFound, status, name)
	}
	// A title with no segments of its own has none to give.
	plain, err := playlists.publish(samplePlaylist)
	require.NoError(t, err)
	status, _, _ := get(t, plain[:len(plain)-len(".m3u8")]+"/0.ts")
	assert.Equal(t, http.StatusNotFound, status)
}

func TestARebasedSegmentThatFailsIsAskedAgainNotRemembered(t *testing.T) {
	f := newFakePluto(t)
	f.serveSegMedia(10, true)
	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	local := PlayURL(stream)
	seg := local[:len(local)-len(".m3u8")] + "/25.ts"
	path := fakeClipRoot + "200-end/hls/video_pt_video_300/seg_00005.ts"

	good := f.routes[path]
	f.setRoute(path, http.StatusBadGateway, "down")
	status, _, _ := get(t, seg)
	assert.Equal(t, http.StatusBadGateway, status, "the player is told, not handed garbage")

	f.mu.Lock()
	f.routes[path] = good
	f.mu.Unlock()
	status, _, body := get(t, seg)
	assert.Equal(t, http.StatusOK, status, "the CDN came back, and so did the segment")
	assert.NoError(t, checkTS(body))
}

func TestARebasedSegmentWithTheWrongKeyIsRefused(t *testing.T) {
	f := newFakePluto(t)
	f.serveSegMedia(10, true)
	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	local := PlayURL(stream)
	f.set(fakeClipRoot+"100-200/hls/video_pt_video_300/seg_keyfile_0.key", "0123456789abcdef")
	status, _, _ := get(t, local[:len(local)-len(".m3u8")]+"/13.ts")
	assert.Equal(t, http.StatusBadGateway, status, "a segment that did not decrypt is not served as one")
}

func TestSegmentsAheadAreFetchedWhileThePlayerIsBusy(t *testing.T) {
	f := newFakePluto(t)
	f.serveSegMedia(10, true)
	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	local := PlayURL(stream)
	seg := local[:len(local)-len(".m3u8")]

	get(t, seg+"/3.ts")
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) && (f.hitCount(fakeClipRoot+"0-100/hls/video_pt_video_300/seg_00005.ts") == 0) {
		time.Sleep(10 * time.Millisecond)
	}
	assert.Positive(t, f.hitCount(fakeClipRoot+"0-100/hls/video_pt_video_300/seg_00004.ts"), "the next segment")
	assert.Positive(t, f.hitCount(fakeClipRoot+"0-100/hls/video_pt_video_300/seg_00005.ts"), "and the one after")
	assert.Zero(t, f.hitCount(fakeClipRoot+"0-100/hls/video_pt_video_300/seg_00007.ts"), "and no more")
}

func TestTwoRequestsForOneSegmentFetchItOnce(t *testing.T) {
	f := newFakePluto(t)
	f.serveSegMedia(10, true)
	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	local := PlayURL(stream)
	seg := local[:len(local)-len(".m3u8")] + "/17.ts"
	path := fakeClipRoot + "100-200/hls/video_pt_video_300/seg_00007.ts"
	before := f.hitCount(path)

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, _, _ := get(t, seg)
			assert.Equal(t, http.StatusOK, status)
		}()
	}
	wg.Wait()
	assert.Equal(t, before+1, f.hitCount(path), "the player's retries and the prefetch share one fetch")
}

func TestASegmentCacheKeepsOnlyWhatIsNear(t *testing.T) {
	var c segmentCache
	for i := range 20 {
		res := c.start(i, func() ([]byte, error) { return []byte{byte(i)}, nil })
		<-res.done
	}
	c.trim(10)
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.items {
		assert.GreaterOrEqual(t, k, 10-segmentKeepBehind)
		assert.LessOrEqual(t, k, 10+segmentPrefetch)
	}
	assert.Len(t, c.items, segmentKeepBehind+1+segmentPrefetch)
}

func TestARequestThatGivesUpDoesNotPoisonTheFetch(t *testing.T) {
	f := newFakePluto(t)
	f.serveSegMedia(10, true)
	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	local := PlayURL(stream)
	entry, ok := playlists.lookup(strings.TrimSuffix(local[strings.LastIndex(local, "/")+1:], ".m3u8"))
	require.True(t, ok)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = entry.segment(ctx, 6)
	assert.Error(t, err, "the caller that left is told so")
	data, err := entry.segment(t.Context(), 6)
	require.NoError(t, err, "and the next one gets the segment")
	assert.NoError(t, checkTS(data))
}
