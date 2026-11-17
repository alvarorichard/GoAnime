package plutotv

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Pluto TV's media host is several CDNs behind one name. One of them answers
// 403 for every playlist under a clip while serving the segments and keys the
// playlist lists, so whether the title's own playlist can be fetched depends on
// the DNS answer of the moment, for this client and for the player alike. What
// Stream hands over is therefore the playlist's text, every URI absolute; and
// when the playlist is refused, the same title read out of the stitched one.

// playlistLines splits a playlist into its tags and its media lines.
func playlistLines(playlist string) (tags, media []string) {
	for _, line := range strings.Split(playlist, "\n") {
		switch {
		case line == "":
		case strings.HasPrefix(line, "#"):
			tags = append(tags, line)
		default:
			media = append(media, line)
		}
	}
	return tags, media
}

func countPrefix(lines []string, prefix string) int {
	n := 0
	for _, l := range lines {
		if strings.HasPrefix(l, prefix) {
			n++
		}
	}
	return n
}

func TestStreamHandsOverTheTitlesPlaylistWithEveryURIAbsolute(t *testing.T) {
	f := newFakePluto(t)
	f.serveEpisode(fakeEpisodeID)

	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	require.True(t, stream.AdFree)
	assert.Equal(t, f.url()+f.contentPath("hls_2400"), stream.URL)

	tags, media := playlistLines(stream.Playlist)
	require.Len(t, media, 9, "the nine segments of the title")
	base := f.url() + fakeClip + "0-end/"
	for i, line := range media {
		assert.Equal(t, base+"hls_2400-0000"+string(rune('0'+i))+".ts", line, "segment %d, absolute and in order", i)
	}
	assert.Equal(t, 9, countPrefix(tags, `#EXT-X-KEY:METHOD=AES-128,URI="`+base+`hls_2400_keyfile_0.key",IV=0x`), "each key file is absolute, its IV untouched")
	assert.Equal(t, "#EXTM3U", tags[0])
	assert.Equal(t, "#EXT-X-ENDLIST", tags[len(tags)-1])
	assert.NotContains(t, stream.Playlist, "DISCONTINUITY", "the title's own playlist is one piece")
}

func TestStreamReadsTheTitleOutOfTheStitchedPlaylistWhenItsOwnIsRefused(t *testing.T) {
	refusals := map[string]func(f *fakePluto){
		"forbidden, as one of the CDNs answers": func(f *fakePluto) { f.setRoute(f.contentPath("hls_2400"), http.StatusForbidden, "") },
		"a page":                                func(f *fakePluto) { f.set(f.contentPath("hls_2400"), "<html>blocked</html>") },
		"a playlist that lists no media":        func(f *fakePluto) { f.set(f.contentPath("hls_2400"), "#EXTM3U\n#EXT-X-ENDLIST\n") },
		"gone":                                  func(f *fakePluto) { f.setRoute(f.contentPath("hls_2400"), http.StatusNotFound, "") },
	}
	for name, refuse := range refusals {
		t.Run(name, func(t *testing.T) {
			f := newFakePluto(t)
			f.serveEpisode(fakeEpisodeID)
			refuse(f)

			stream, err := f.client().Stream(t.Context(), episodeURL, "best")
			require.NoError(t, err)
			assert.True(t, stream.AdFree, "the title alone is still what plays")
			assert.Contains(t, stream.URL, "/master.m3u8", "the URL is one that answers: the stitched master")

			tags, media := playlistLines(stream.Playlist)
			var want []string
			for _, r := range []string{"0-100", "101-200", "201-end"} {
				for s := range 3 {
					want = append(want, f.url()+fakeClip+r+"/hls_2400-0000"+string(rune('0'+s))+".ts")
				}
			}
			assert.Equal(t, want, media, "the title's ranges in order, and nothing spliced between them")
			assert.Equal(t, 2, countPrefix(tags, "#EXT-X-DISCONTINUITY"), "one between each pair of ranges")
			assert.Equal(t, 9, countPrefix(tags, "#EXT-X-KEY:METHOD=AES-128,URI=\""+f.url()+fakeClip), "every segment keeps its key")
			assert.Equal(t, 9, countPrefix(tags, "#EXTINF:5,"))
			assert.Contains(t, tags, "#EXT-X-TARGETDURATION:5")
			assert.Contains(t, tags, "#EXT-X-PLAYLIST-TYPE:VOD")
			assert.Equal(t, "#EXT-X-ENDLIST", tags[len(tags)-1])
			assert.NotContains(t, stream.Playlist, "/v1/mp4/", "no ad creative")
			assert.NotContains(t, stream.Playlist, "bumper", "nor the service's closing bumper")
			assert.NotContains(t, stream.Playlist, "METHOD=NONE", "the unkeyed state of the ads does not leak onto the title")

			assert.Zero(t, f.hitCount(f.variantPath(fakeEpisodeID, 2294080)), "the quality asked for is kept: no lower rung is tried")
		})
	}
}

func TestStreamKeepsTheQualityAskedForWhenReadingTheStitchedPlaylist(t *testing.T) {
	f := newFakePluto(t)
	f.serveEpisode(fakeEpisodeID)
	for _, rung := range fakeLadder {
		f.setRoute(f.contentPath(rung.rendition), http.StatusForbidden, "")
	}
	stream, err := f.client().Stream(t.Context(), episodeURL, "worst")
	require.NoError(t, err)
	require.True(t, stream.AdFree)
	_, media := playlistLines(stream.Playlist)
	require.NotEmpty(t, media)
	for _, line := range media {
		assert.Contains(t, line, "/hls_300-", "the lowest rung's segments")
	}
}

func TestStreamOfATitleInSeveralClipsHasNoPlaylistOfItsOwn(t *testing.T) {
	f := newFakePluto(t)
	f.serveEpisode(fakeEpisodeID)
	a := f.stitched("hls_2400", 4, "0-100", "101-end")
	b := strings.ReplaceAll(f.stitched("hls_2400", 4, "0-100", "101-end"), "abc_Title", "other_Title")
	f.set(f.variantPath(fakeEpisodeID, 3321280), a+b)

	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	assert.False(t, stream.AdFree)
	assert.Empty(t, stream.Playlist, "nothing is assembled for a layout that was not understood")
}

// ---- titlePlaylist ---------------------------------------------------------

const (
	tpBase    = "https://stitcher.pluto.tv/v2/stitch/hls/episode/x/3321280/playlist.m3u8?sid=1"
	tpDir     = "https://siloh.plutotv.net/1_pluto/clip/abc_Title/720p/2022/hls/"
	tpContent = tpDir + "0-end/hls_2400.m3u8"
)

func tpKey(dir string, iv int) string {
	return `#EXT-X-KEY:METHOD=AES-128,URI="` + tpDir + dir + `/hls_2400_keyfile_0.key",IV=0x0000000000000000000000000000000` + string(rune('0'+iv))
}

func tpSeg(dir string, n int) string {
	return tpDir + dir + "/hls_2400-0000" + string(rune('0'+n)) + ".ts"
}

func TestTitlePlaylistKeepsOnlyTheTitlesSegments(t *testing.T) {
	stitched := strings.Join([]string{
		"#EXTM3U", "#EXT-X-VERSION:5", "#EXT-X-TARGETDURATION:9", "#EXT-X-PROGRAM-DATE-TIME:2026-11-16T14:02:30.697Z",
		tpKey("0-100", 1), "#EXTINF:5,", tpSeg("0-100", 0),
		tpKey("0-100", 2), "#EXTINF:5.5,", tpSeg("0-100", 1),
		"#EXT-X-DISCONTINUITY", "#EXT-X-KEY:METHOD=NONE",
		"#EXTINF:9,", "https://siloh.plutotv.net/v1/mp4/c(ts)/p(0_ad%2Fcreative%2Fx_ad%2F720p%2Fvideo_2400.mp4)/frag(1-2)",
		"#EXTINF:5,", "https://siloh.plutotv.net/0_ad/creative/697e_ad/720p/2026/hls/0-end/hls_2400-00000.ts",
		"#EXTINF:5,", tpDir + "0-100/hls_600-00000.ts",
		"#EXTINF:5,", "https://siloh.plutotv.net/1_pluto/clip/other_Title/720p/2022/hls/0-100/hls_2400-00000.ts",
		"#EXT-X-DISCONTINUITY",
		tpKey("101-end", 1), "#EXTINF:4.25,", tpSeg("101-end", 0),
		"#EXT-X-ENDLIST", "",
	}, "\n")

	got, ok := titlePlaylist(stitched, tpBase, tpContent)
	require.True(t, ok)
	want := strings.Join([]string{
		"#EXTM3U", "#EXT-X-VERSION:3", "#EXT-X-PLAYLIST-TYPE:VOD", "#EXT-X-TARGETDURATION:6", "#EXT-X-MEDIA-SEQUENCE:0",
		tpKey("0-100", 1), "#EXTINF:5,", tpSeg("0-100", 0),
		tpKey("0-100", 2), "#EXTINF:5.5,", tpSeg("0-100", 1),
		"#EXT-X-DISCONTINUITY",
		tpKey("101-end", 1), "#EXTINF:4.25,", tpSeg("101-end", 0),
		"#EXT-X-ENDLIST", "",
	}, "\n")
	assert.Equal(t, want, got, "the ad, its creative under /hls/, another rendition and another clip are all left out; the target duration is the title's longest segment, rounded up")
}

func TestTitlePlaylistResolvesRelativeURIsAndReadsCRLF(t *testing.T) {
	stitched := "#EXTM3U\r\n#EXT-X-KEY:METHOD=AES-128,URI=\"/1_pluto/clip/abc_Title/720p/2022/hls/0-end/k.key\",IV=0x01\r\n#EXTINF:5,\r\n/1_pluto/clip/abc_Title/720p/2022/hls/0-end/hls_2400-00000.ts\r\n#EXT-X-ENDLIST\r\n"
	got, ok := titlePlaylist(stitched, "https://siloh.plutotv.net/x/playlist.m3u8", tpContent)
	require.True(t, ok)
	assert.Contains(t, got, `URI="https://siloh.plutotv.net/1_pluto/clip/abc_Title/720p/2022/hls/0-end/k.key",IV=0x01`)
	assert.Contains(t, got, "\n"+tpDir+"0-end/hls_2400-00000.ts\n")
	assert.NotContains(t, got, "\r")
}

func TestTitlePlaylistCarriesAnUnkeyedTitleAsUnkeyed(t *testing.T) {
	stitched := "#EXTM3U\n#EXT-X-KEY:METHOD=NONE\n#EXTINF:5,\n" + tpSeg("0-end", 0) + "\n#EXT-X-ENDLIST\n"
	got, ok := titlePlaylist(stitched, tpBase, tpContent)
	require.True(t, ok)
	assert.Contains(t, got, "#EXT-X-KEY:METHOD=NONE\n#EXTINF:5,\n"+tpSeg("0-end", 0))

	plain, ok := titlePlaylist("#EXTM3U\n#EXTINF:5,\n"+tpSeg("0-end", 0)+"\n", tpBase, tpContent)
	require.True(t, ok)
	assert.NotContains(t, plain, "#EXT-X-KEY", "no key line is invented for a playlist that has none")
}

func TestTitlePlaylistRefusesWhatItCannotCarryOverExactly(t *testing.T) {
	cases := map[string]struct{ stitched, content string }{
		"a key that leaves the IV to the sequence number": {
			"#EXTM3U\n#EXT-X-KEY:METHOD=AES-128,URI=\"" + tpDir + "0-end/k.key\"\n#EXTINF:5,\n" + tpSeg("0-end", 0) + "\n", tpContent},
		"a segment with no duration": {
			"#EXTM3U\n" + tpKey("0-end", 1) + "\n" + tpSeg("0-end", 0) + "\n", tpContent},
		"a duration that belongs to the ad before it": {
			"#EXTM3U\n#EXTINF:9,\nhttps://siloh.plutotv.net/v1/mp4/ad\n" + tpSeg("0-end", 0) + "\n", tpContent},
		"a duration that is not a number": {
			"#EXTM3U\n#EXTINF:abc,\n" + tpSeg("0-end", 0) + "\n", tpContent},
		"a duration no number can hold": {
			"#EXTM3U\n#EXTINF:1e999,\n" + tpSeg("0-end", 0) + "\n", tpContent},
		"a duration of zero": {
			"#EXTM3U\n#EXTINF:0,\n" + tpSeg("0-end", 0) + "\n", tpContent},
		"no segment of the title at all": {
			"#EXTM3U\n#EXTINF:5,\nhttps://siloh.plutotv.net/v1/mp4/ad\n#EXT-X-ENDLIST\n", tpContent},
		"a content URL that is not the 0-end playlist": {
			"#EXTM3U\n#EXTINF:5,\n" + tpSeg("0-end", 0) + "\n", tpDir + "hls_2400.m3u8"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, ok := titlePlaylist(tc.stitched, tpBase, tc.content)
			assert.False(t, ok)
			assert.Empty(t, got)
		})
	}
}

// ---- absolutePlaylist ------------------------------------------------------

func TestAbsolutePlaylistResolvesSegmentsAndKeyFiles(t *testing.T) {
	in := "#EXTM3U\r\n#EXT-X-KEY:METHOD=AES-128,URI=\"hls_300_keyfile_0.key\",IV=0x01\r\n#EXTINF:5.000000,\r\nhls_300-00000.ts\r\n\r\n#EXT-X-KEY:METHOD=AES-128,URI=\"https://other.example/k.key\",IV=0x02\n#EXTINF:5,\nhttps://other.example/a.ts\n  ../up/b.ts  \n#EXT-X-ENDLIST\n"
	got := absolutePlaylist(in, "https://cdn.example/clip/hls/0-end/hls_300.m3u8?x=1")
	want := "#EXTM3U\n#EXT-X-KEY:METHOD=AES-128,URI=\"https://cdn.example/clip/hls/0-end/hls_300_keyfile_0.key\",IV=0x01\n#EXTINF:5.000000,\nhttps://cdn.example/clip/hls/0-end/hls_300-00000.ts\n\n#EXT-X-KEY:METHOD=AES-128,URI=\"https://other.example/k.key\",IV=0x02\n#EXTINF:5,\nhttps://other.example/a.ts\nhttps://cdn.example/clip/hls/up/b.ts\n#EXT-X-ENDLIST\n"
	assert.Equal(t, want, got)
}

func TestAbsolutePlaylistLeavesWhatItCannotResolve(t *testing.T) {
	in := "#EXTM3U\n#EXTINF:5,\na.ts\n"
	assert.Equal(t, in, absolutePlaylist(in, "http://[::1"), "a base that is no URL changes nothing")
	odd := "#EXTM3U\n#EXT-X-KEY:METHOD=AES-128,URI=\"%zz\"\n#EXTINF:5,\n%zz\n"
	assert.Equal(t, odd, absolutePlaylist(odd, "https://cdn.example/a/b.m3u8"), "a reference that is no URL is kept as written")
}

// ---- real answers ----------------------------------------------------------

// The stitched playlists captured from the service, read the way Stream reads
// them when the title's own playlist is refused.
func TestRegressionTheTitleIsReadOutOfTheRealStitchedPlaylists(t *testing.T) {
	for _, name := range []string{"stitched_series_2026_11_16.m3u8", "stitched_movie_2026_11_16.m3u8", "stitched_movie_pt_2026_11_16.m3u8"} {
		t.Run(name, func(t *testing.T) {
			stitched := fixture(t, name)
			contentURL, ok := contentPlaylistURL(stitched)
			require.True(t, ok)

			own, ok := titlePlaylist(stitched, "https://stitcher.pluto.tv/v2/stitch/hls/episode/x/1/playlist.m3u8", contentURL)
			require.True(t, ok)
			tags, media := playlistLines(own)
			require.NotEmpty(t, media)

			dir := contentURL[:strings.LastIndex(contentURL, "0-end/")]
			titleSegments := 0
			for _, line := range strings.Split(strings.ReplaceAll(stitched, "\r\n", "\n"), "\n") {
				if strings.HasPrefix(line, dir) {
					titleSegments++
				}
			}
			assert.Len(t, media, titleSegments, "every segment of the title, and only those")
			for _, line := range media {
				assert.True(t, strings.HasPrefix(line, dir), "not the title's: %s", line)
				assert.NotContains(t, line, "_ad")
				assert.NotContains(t, line, "bumper")
			}
			assert.Equal(t, len(media), countPrefix(tags, "#EXTINF:"))
			assert.Equal(t, len(media), countPrefix(tags, "#EXT-X-KEY:METHOD=AES-128"), "the real title is keyed segment by segment")
			assert.Zero(t, countPrefix(tags, "#EXT-X-KEY:METHOD=NONE"))
			assert.False(t, isProtected(own))
		})
	}
}

// A stitched playlist the title cannot be read out of exactly is not guessed
// at: the next rung, whose own playlist answers, is used instead.
func TestStreamTriesTheNextRungWhenTheTitleCannotBeReadOutOfTheStitchedPlaylist(t *testing.T) {
	f := newFakePluto(t)
	f.serveEpisode(fakeEpisodeID)
	f.setRoute(f.contentPath("hls_2400"), http.StatusForbidden, "")
	withoutIV := strings.ReplaceAll(f.stitched("hls_2400", 3, "0-100", "101-end"), ",IV=0x00000000000000000000000000000001", "")
	f.set(f.variantPath(fakeEpisodeID, 3321280), withoutIV)

	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	assert.True(t, stream.AdFree)
	assert.Equal(t, f.url()+f.contentPath("hls_1600"), stream.URL)
	_, media := playlistLines(stream.Playlist)
	require.Len(t, media, 9)
	assert.Contains(t, media[0], "/0-end/hls_1600-00000.ts")
}

// When no rung yields the title alone, the stitched master is what is left.
func TestStreamPlaysTheStitchedMasterWhenNoRungYieldsTheTitle(t *testing.T) {
	f := newFakePluto(t)
	f.serveEpisode(fakeEpisodeID)
	for _, rung := range fakeLadder {
		f.setRoute(f.contentPath(rung.rendition), http.StatusForbidden, "")
		withoutIV := strings.ReplaceAll(f.stitched(rung.rendition, 3, "0-100", "101-end"), ",IV=0x00000000000000000000000000000001", "")
		f.set(f.variantPath(fakeEpisodeID, rung.bandwidth), withoutIV)
	}
	stream, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	assert.False(t, stream.AdFree)
	assert.Empty(t, stream.Playlist)
	assert.Contains(t, stream.URL, "/master.m3u8")
	assert.Zero(t, f.hitCount(f.contentPath("hls_600")), "three rungs at most, as before")
}
