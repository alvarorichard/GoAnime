package plutotv

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
)

// Every way the rebase can be refused or fail, each on its own. A branch no
// test takes is a branch a mutant can break unseen; these exist so that every
// one is taken, and the meta tests in internal/api hold them to it.

// ---- the transport stream's corners ----------------------------------------

func TestAPacketWithNoPayloadHasNoTimestampAndIsLeftAlone(t *testing.T) {
	// Adaptation field only (AFC 2), carrying a PCR: it moves, nothing else.
	p := tsPES(0x100, 0xE0, 900000, -1, 450000)
	p[3] = 0x20
	payload, pcr := tsLayout(p)
	assert.Equal(t, -1, payload)
	assert.Equal(t, 6, pcr)
	_, ok := firstTimestamp(p)
	assert.False(t, ok, "a PES header in a packet that says it has no payload is not read")

	seg := append(append([]byte{}, p...), tsSegment(90000, 1)...)
	require.NoError(t, shiftTimestamps(seg, 90000))
	base := int64(seg[6])<<25 | int64(seg[7])<<17 | int64(seg[8])<<9 | int64(seg[9])<<1 | int64(seg[10]>>7)
	assert.EqualValues(t, 540000, base, "its PCR moved")
	assert.Equal(t, p[12:], seg[12:tsPacket], "and its bytes after the adaptation field did not")

	// An adaptation field that fills the packet leaves no payload either.
	full := tsFiller(0x100, 0)
	full[3], full[4] = 0x30, 183
	payload, _ = tsLayout(full)
	assert.Equal(t, -1, payload)
}

func TestAPCRFlagWithTooShortAnAdaptationFieldIsNoPCR(t *testing.T) {
	p := tsFiller(0x100, 0)
	p[3], p[4], p[5] = 0x30, 3, 0x10 // says PCR, has room for none
	_, pcr := tsLayout(p)
	assert.Equal(t, -1, pcr)
}

func TestStreamsWithNoOptionalHeaderCarryNoTimestamp(t *testing.T) {
	for _, id := range []byte{0xBC, 0xBE, 0xBF, 0xF0, 0xF1, 0xF2, 0xF8, 0xFF} {
		p := tsPES(0x100, id, 90000, -1, -1)
		payload, _ := tsLayout(p)
		pts, dts := pesTimes(p, payload)
		assert.Equal(t, []int{-1, -1}, []int{pts, dts}, "stream id %#x", id)
		before := append([]byte{}, p...)
		require.NoError(t, shiftTimestamps(p, 777))
		assert.Equal(t, before, p, "stream id %#x is not rewritten", id)
	}
	for _, id := range []byte{0xE0, 0xC0, 0xBD} {
		p := tsPES(0x100, id, 90000, -1, -1)
		payload, _ := tsLayout(p)
		pts, _ := pesTimes(p, payload)
		assert.NotEqual(t, -1, pts, "stream id %#x has a PTS", id)
	}
}

func TestAPESThatSaysPTSAndDTSWithNoRoomForTheDTSIsNotRead(t *testing.T) {
	// The header starts 16 bytes before the end of the packet: room for its
	// PTS (bytes 9-13), none for a DTS (14-18).
	full := tsPES(0x100, 0xE0, 90000, 45000, -1)
	header := full[4 : 4+19]
	p := tsFiller(0x100, 0)
	p[1] |= 0x40
	p[3] = 0x30
	const payload = tsPacket - 16
	p[4], p[5] = byte(payload-5), 0
	copy(p[payload:], header[:16])
	got, _ := tsLayout(p)
	require.Equal(t, payload, got)
	pts, dts := pesTimes(p, got)
	assert.Equal(t, []int{-1, -1}, []int{pts, dts}, "a header cut short is not read at all")
	before := append([]byte{}, p...)
	require.NoError(t, shiftTimestamps(p, 90000))
	assert.Equal(t, before, p)
}

func TestBytesThatLookLikeAPESHeaderAreNotOne(t *testing.T) {
	// A packet in the middle of a PES whose payload happens to start with a
	// start code and a PTS flag: without the start indicator it is data.
	mid := tsPES(0x100, 0xE0, 90000, -1, -1)
	mid[1] &^= 0x40
	before := append([]byte{}, mid...)
	require.NoError(t, shiftTimestamps(mid, 12345))
	assert.Equal(t, before, mid, "data in the middle of a PES is not rewritten")
	_, ok := firstTimestamp(mid)
	assert.False(t, ok)

	// A packet that starts a unit but whose payload is no start code.
	odd := tsPES(0x100, 0xE0, 90000, -1, -1)
	odd[4+2] = 0x02 // 00 00 02: not a start code
	before = append([]byte{}, odd...)
	require.NoError(t, shiftTimestamps(odd, 12345))
	assert.Equal(t, before, odd, "bytes that are not a PES header are not one")
	_, ok = firstTimestamp(odd)
	assert.False(t, ok)
}

func TestShiftTimestampsMovesAPTSThatHasNoDTS(t *testing.T) {
	p := tsPES(0x101, 0xC0, 90000, -1, -1) // audio: a PTS alone
	require.NoError(t, shiftTimestamps(p, 450000))
	assert.EqualValues(t, 540000, readTS(p[4+9:]))
	at, ok := firstTimestamp(p)
	require.True(t, ok)
	assert.EqualValues(t, 540000, at)
}

func TestWriteTSKeepsThePrefixAndTheMarkerBits(t *testing.T) {
	for _, prefix := range []byte{0x20, 0x30, 0x10} {
		for _, v := range []int64{0, 1, 90000, ptsMask, 1 << 32, 0x15555_5555} {
			b := []byte{prefix | 1, 0, 1, 0, 1}
			writeTS(b, v)
			assert.Equal(t, prefix, b[0]&0xF0, "the 4-bit prefix says PTS or DTS: it is kept")
			assert.Equal(t, byte(1), b[0]&1, "marker bit")
			assert.Equal(t, byte(1), b[2]&1, "marker bit")
			assert.Equal(t, byte(1), b[4]&1, "marker bit")
			assert.Equal(t, v&ptsMask, readTS(b))
		}
	}
	b := make([]byte, 5)
	writeTS(b, -1)
	assert.EqualValues(t, ptsMask, readTS(b), "a negative timestamp wraps as the 33-bit clock does")
	writeTS(b, ptsMask+91)
	assert.EqualValues(t, 90, readTS(b), "and so does one past the top")
}

func TestTheFirstClockOfVideoAloneIsItsDTS(t *testing.T) {
	// B-frames: the video's first PTS is ahead of its DTS, and with no audio
	// to set the clock the DTS is where the segment starts.
	p := tsPES(0x100, 0xE0, 90000+6006, 90000, -1)
	at, ok := firstTimestamp(p)
	require.True(t, ok)
	assert.EqualValues(t, 90000, at)
}

func TestTheCacheOfATitlePlayedThroughStaysSmall(t *testing.T) {
	f := newFakePluto(t)
	f.serveSegMedia(10, true)
	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	local := PlayURL(stream)
	id := strings.TrimSuffix(local[strings.LastIndex(local, "/")+1:], ".m3u8")
	for n := range 30 {
		status, _, _ := get(t, fmt.Sprintf("%s/%d.ts", local[:len(local)-len(".m3u8")], n))
		require.Equal(t, http.StatusOK, status)
	}
	entry, ok := playlists.lookup(id)
	require.True(t, ok)
	entry.cache.mu.Lock()
	defer entry.cache.mu.Unlock()
	assert.LessOrEqual(t, len(entry.cache.items), segmentKeepBehind+1+segmentPrefetch, "a whole film is not kept in memory")
}

func TestCheckTSRefusesALaterPacketOutOfSync(t *testing.T) {
	seg := tsSegment(0, 2)
	seg[tsPacket*2] = 0x48
	assert.ErrorIs(t, checkTS(seg), errNotTS)
	assert.NoError(t, checkTS(tsSegment(0, 2)))
}

func TestFirstTimestampStopsAtALaterPacketOutOfSync(t *testing.T) {
	seg := append(tsFiller(0x100, 1), tsSegment(90000, 1)...)
	seg[tsPacket] = 0
	_, ok := firstTimestamp(seg)
	assert.False(t, ok, "a stream that loses sync is not read on as if it had not")
}

func TestDecryptRefusesAKeyOfTheWrongLength(t *testing.T) {
	_, err := decryptPrefix(make([]byte, 64), []byte("short"), [16]byte{})
	assert.Error(t, err)
	_, err = decryptAES128(make([]byte, 64), []byte("short"), [16]byte{})
	assert.Error(t, err)
}

// ---- the plan's corners ------------------------------------------------------

func TestPlanRebaseOfWhatIsNotATitlePlaylistIsAnError(t *testing.T) {
	f := newFakePluto(t)
	plan, err := f.client().planRebase(t.Context(), "#EXTM3U\n#EXT-X-ENDLIST\n")
	assert.Nil(t, plan)
	assert.Error(t, err)
}

func TestPlanRebaseThatTheCallerGivesUpOnGivesUp(t *testing.T) {
	f := newFakePluto(t)
	f.serveSegMedia(10, true)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	f.onHit = func(path string) {
		if strings.HasSuffix(path, "seg_00000.ts") {
			cancel() // the user left while the ranges were being read
		}
	}
	_, err := f.client().Stream(ctx, episodeURL, "best")
	assert.ErrorIs(t, err, context.Canceled, "not a title handed over as if nothing happened")
}

func TestProbeOfASegmentWhoseStartCannotBeDecrypted(t *testing.T) {
	plan := &rebasePlan{fetch: func(_ context.Context, url string, _ int64, _ bool) ([]byte, error) {
		if strings.HasSuffix(url, ".key") {
			return []byte("0123456789abcdef"), nil
		}
		return []byte{1, 2, 3}, nil // less than one block
	}}
	_, err := plan.probe(t.Context(), segmentRef{url: "https://x.pluto.tv/a.ts", keyURL: "https://x.pluto.tv/k.key"})
	assert.Error(t, err)
}

func TestASegmentPastTheEndOrBeforeTheStartIsNone(t *testing.T) {
	plan := &rebasePlan{segs: make([]segmentRef, 3)}
	for _, i := range []int{-1, 3, 99} {
		_, err := plan.segment(t.Context(), i, &keyCache{})
		assert.Error(t, err, "segment %d", i)
	}
}

func TestASegmentWhoseKeyCannotBeHadIsRefused(t *testing.T) {
	plan := &rebasePlan{segs: []segmentRef{{url: "u", keyURL: "k"}}, fetch: func(_ context.Context, url string, _ int64, _ bool) ([]byte, error) {
		if url == "k" {
			return nil, errors.New("gone")
		}
		return tsSegment(0, 1), nil
	}}
	_, err := plan.segment(t.Context(), 0, &keyCache{})
	assert.ErrorContains(t, err, "gone")
}

func TestASegmentThatIsNotTSIsRefused(t *testing.T) {
	plan := &rebasePlan{segs: []segmentRef{{url: "u"}}, fetch: func(context.Context, string, int64, bool) ([]byte, error) {
		return []byte("<html>a CDN's error page</html>"), nil
	}}
	_, err := plan.segment(t.Context(), 0, &keyCache{})
	assert.ErrorIs(t, err, errNotTS, "a page served as a segment is not handed to the player")
}

func TestAnUnencryptedSegmentIsMovedWithoutAKey(t *testing.T) {
	plan := &rebasePlan{segs: []segmentRef{{url: "u", shift: 90000}}, fetch: func(context.Context, string, int64, bool) ([]byte, error) {
		return tsSegment(90000, 1), nil
	}}
	data, err := plan.segment(t.Context(), 0, &keyCache{})
	require.NoError(t, err)
	at, _ := firstTimestamp(data)
	assert.EqualValues(t, 180000, at)
}

func TestTheKeyCacheAsksForAKeyOnceAndRefusesOneThatIsNot16Bytes(t *testing.T) {
	asked := 0
	fetch := func(_ context.Context, url string, _ int64, _ bool) ([]byte, error) {
		asked++
		if url == "bad" {
			return []byte("fifteen bytes!!"[:15]), nil
		}
		return []byte("0123456789abcdef"), nil
	}
	var keys keyCache
	for range 5 {
		key, err := keys.get(t.Context(), "good", fetch)
		require.NoError(t, err)
		assert.Len(t, key, 16)
	}
	assert.Equal(t, 1, asked, "one fetch for the range, whatever number of segments ask")

	_, err := keys.get(t.Context(), "bad", fetch)
	assert.Error(t, err)
	_, err = keys.get(t.Context(), "bad", fetch)
	assert.Error(t, err, "and a refused key is not remembered as a good one")
	assert.Equal(t, 3, asked)
}

// ---- reading from the CDN ----------------------------------------------------

func TestGetBytesRefusesAHostThatIsNotPlutos(t *testing.T) {
	spy := newSpyHost(t, "x")
	_, err := NewClient().getBytes(t.Context(), spy.URL+"/a.ts", 10, false)
	assert.Error(t, err)
	assert.Zero(t, spy.hits.Load())
}

func TestGetBytesReportsWhatCannotBeAsked(t *testing.T) {
	f := newFakePluto(t)
	var nilCtx context.Context
	_, err := f.client().getBytes(nilCtx, f.url()+"/media/a.ts", 10, false)
	assert.Error(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	addr := srv.URL
	srv.Close() // nothing listens there any more
	c := NewClientForTest(http.DefaultClient, addr)
	_, err = c.getBytes(t.Context(), addr+"/a.ts", 10, false)
	assert.Error(t, err)
}

func TestGetBytesReportsAStatusOtherThanOK(t *testing.T) {
	f := newFakePluto(t)
	f.setRoute("/media/a.ts", http.StatusForbidden, "no")
	_, err := f.client().getBytes(t.Context(), f.url()+"/media/a.ts", 10, false)
	assert.ErrorContains(t, err, "403")
}

func TestGetBytesHoldsTheLimit(t *testing.T) {
	f := newFakePluto(t)
	f.set("/media/a.ts", strings.Repeat("x", 100))
	c := f.client()

	_, err := c.getBytes(t.Context(), f.url()+"/media/a.ts", 99, false)
	assert.ErrorContains(t, err, "size limit", "a whole read that runs over is refused")
	body, err := c.getBytes(t.Context(), f.url()+"/media/a.ts", 100, false)
	require.NoError(t, err)
	assert.Len(t, body, 100, "one that fits is whole")

	body, err = c.getBytes(t.Context(), f.url()+"/media/a.ts", 10, true)
	require.NoError(t, err)
	assert.Len(t, body, 10, "a prefix stops where it was asked to, whatever the server sent")
	assert.Equal(t, "bytes=0-9", f.lastRequest("/media/a.ts").Header.Get("Range"))
	assert.Equal(t, netx.UserAgent, f.lastRequest("/media/a.ts").Header.Get("User-Agent"), "the CDN refuses a request that names no User-Agent")
}

func TestGetBytesReportsABodyCutShort(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		conn, buf, err := w.(http.Hijacker).Hijack()
		if err != nil {
			return
		}
		_, _ = fmt.Fprint(buf, "HTTP/1.1 200 OK\r\nContent-Length: 1000\r\n\r\nonly a little")
		_ = buf.Flush()
		_ = conn.Close()
	}))
	t.Cleanup(srv.Close)
	c := NewClientForTest(srv.Client(), srv.URL)
	_, err := c.getBytes(t.Context(), srv.URL+"/a.ts", 10_000, false)
	assert.Error(t, err, "a segment cut short is not a segment")
}

// ---- serving -----------------------------------------------------------------

func TestPublishPlanRefusesAPlanWithNothingToServe(t *testing.T) {
	s := newPlaylistServer()
	_, err := s.publishPlan(nil)
	assert.Error(t, err)
	_, err = s.publishPlan(&rebasePlan{})
	assert.Error(t, err)
}

func TestPlayURLOfARebasedTitleThatCannotBeServedIsTheCDNsPlaylist(t *testing.T) {
	prev := publishPlanFn
	t.Cleanup(func() { publishPlanFn = prev })
	publishPlanFn = func(*rebasePlan) (string, error) { return "", errors.New("no port") }

	s := Stream{URL: "https://stitcher.pluto.tv/x.m3u8", Playlist: samplePlaylist, AdFree: true, rebase: &rebasePlan{segs: make([]segmentRef, 1)}}
	got := PlayURL(s)
	assert.True(t, strings.HasPrefix(got, "http://127.0.0.1:"), "the title still plays, from its absolute playlist: %s", got)
	status, _, body := fetchLocal(t, http.MethodGet, got)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, samplePlaylist, body)
}

// A range whose clock starts later than its place (an encode that kept the
// whole film's clock in every range, but the ranges reordered or the first one
// trimmed): the shift is negative and it is put right too.
func TestPlanRebaseOfRangesThatRunBackwardsInTime(t *testing.T) {
	for name, hours := range map[string]func(r int) int64{
		"later ranges start an hour later each": func(r int) int64 { return int64(r) },
		"later ranges start an hour earlier":    func(r int) int64 { return int64(2 - r) },
	} {
		t.Run(name, func(t *testing.T) {
			f := newFakePluto(t)
			f.serveSegMedia(10, true)
			for r, rng := range testRanges {
				for s := range 10 {
					start := int64(clockStart) + int64(r*10+s)*segTicks + hours(r)*3600*90000
					enc := encrypt(tsSegment(start, 12+s), keyOf(r), ivOf(s+1))
					f.set(fmt.Sprintf("%s%s/hls/video_pt_video_300/seg_%05d.ts", fakeClipRoot, rng, s), string(enc))
				}
			}
			stream, err := f.client().Stream(t.Context(), episodeURL, "best")
			require.NoError(t, err)
			require.NotNil(t, stream.rebase, "a range an hour off is put right whichever way it is off")
			for i, s := range stream.rebase.segs {
				r := i / 10
				assert.EqualValues(t, -(hours(r)-hours(0))*3600*90000, s.shift, "segment %d", i)
			}
		})
	}
}

// A failure forgets the failed fetch, and only that one: a retry that started
// meanwhile under the same number is not dropped with it.
func TestASegmentCacheForgetsOnlyTheFailureItself(t *testing.T) {
	var c segmentCache
	release := make(chan struct{})
	failed := c.start(1, func() ([]byte, error) { <-release; return nil, errors.New("down") })
	// Someone else (a trim, a later start) replaced the entry meanwhile.
	c.mu.Lock()
	newer := &segmentResult{done: make(chan struct{})}
	c.items[1] = newer
	c.mu.Unlock()
	close(release)
	<-failed.done

	require.Eventually(t, func() bool {
		c.mu.Lock()
		defer c.mu.Unlock()
		return c.items[1] == newer
	}, time.Second, time.Millisecond, "the newer entry is still there")

	// And a failure that is the entry is forgotten.
	gone := c.start(2, func() ([]byte, error) { return nil, errors.New("down") })
	<-gone.done
	require.Eventually(t, func() bool {
		c.mu.Lock()
		defer c.mu.Unlock()
		_, ok := c.items[2]
		return !ok
	}, time.Second, time.Millisecond)
}
