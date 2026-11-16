package download

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alvarorichard/Goanime/v2/internal/api"
	"github.com/alvarorichard/Goanime/v2/internal/api/providers"
	"github.com/alvarorichard/Goanime/v2/internal/downloader/hls"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/player"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/plutotv"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// These tests follow a Pluto TV anime all the way to a file on disk, the way
// -d and -dm do it: the season listing (a real captured answer for Death Note),
// the stream through the provider registry, then the repo's HLS downloader
// against segments that are really AES-128 encrypted. What they pin is that
// the file holds the title alone: nothing of the ads the stitcher splices in.

const (
	plutoKey   = "0123456789abcdef"
	plutoClip  = "/media/clip/abc_DeathNote/720p/20220214/hls/"
	plutoSeg   = 4 // segments of the title
	plutoFixed = "../scraper/providers/plutotv/testdata/seasons_2026_11_16.json"
)

var (
	plutoEpisodeRe = regexp.MustCompile(`^/v2/stitch/hls/episode/[0-9a-f]{24}/(master\.m3u8|(\d+)/playlist\.m3u8)$`)
	plutoLadder    = map[string]string{"3321280": "hls_2400", "1042180": "hls_600"}
)

// plutoPayload is segment n as the player should see it: whole MPEG-TS packets
// (sync byte 0x47) that differ per segment, so order and decryption show.
func plutoPayload(segment int) []byte {
	p := bytes.Repeat([]byte{byte(segment + 1)}, 188*4)
	for i := 0; i < len(p); i += 188 {
		p[i] = 0x47
	}
	return p
}

// plutoEncrypt is AES-128-CBC with PKCS#7 padding (RFC 8216 §5.2); the IV of
// segment n is n+1, which is what the title's playlist declares.
func plutoEncrypt(t *testing.T, segment int) []byte {
	t.Helper()
	block, err := aes.NewCipher([]byte(plutoKey))
	require.NoError(t, err)
	iv := make([]byte, aes.BlockSize)
	iv[aes.BlockSize-1] = byte(segment + 1)
	plain := plutoPayload(segment)
	pad := aes.BlockSize - len(plain)%aes.BlockSize
	plain = append(plain, bytes.Repeat([]byte{byte(pad)}, pad)...)
	out := make([]byte, len(plain))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, plain)
	return out
}

// plutoService is a Pluto TV that serves any episode: a master, the stitched
// playlist with an ad between two ranges of the title, the title's own
// playlist, its key and its segments.
type plutoService struct {
	srv  *httptest.Server
	mu   sync.Mutex
	hits []string
	// refusePlaylists makes the CDN answer 403 for playlists, not segments.
	refusePlaylists bool
}

func newPlutoService(t *testing.T) *plutoService {
	t.Helper()
	seasons, err := os.ReadFile(plutoFixed)
	require.NoError(t, err)
	s := &plutoService{}
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.hits = append(s.hits, r.URL.Path)
		s.mu.Unlock()
		p := r.URL.Path
		switch {
		case p == "/v4/start":
			_, _ = fmt.Fprintf(w, `{"sessionToken":"tok","stitcherParams":"appName=web&sid=sid-1&deviceId=D&serverSideAds=false","refreshInSec":3000,"servers":{"search":%q,"vod":%q,"stitcher":%q}}`, s.srv.URL, s.srv.URL, s.srv.URL)
		case strings.HasPrefix(p, "/v4/vod/series/") && strings.HasSuffix(p, "/seasons"):
			_, _ = w.Write(seasons)
		case plutoEpisodeRe.MatchString(p):
			s.stitcher(w, p)
		case strings.HasPrefix(p, "/media/"):
			if r.Header.Get("User-Agent") == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			s.media(t, w, r, p)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(s.srv.Close)
	return s
}

func (s *plutoService) stitcher(w http.ResponseWriter, p string) {
	m := plutoEpisodeRe.FindStringSubmatch(p)
	if m[1] == "master.m3u8" {
		fmt.Fprint(w, "#EXTM3U\n#EXT-X-VERSION:5\n")
		for _, bw := range []string{"1042180", "3321280"} {
			fmt.Fprintf(w, "#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=%s\n%s/playlist.m3u8?sid=sid-1\n", bw, bw)
		}
		return
	}
	rendition := plutoLadder[m[2]]
	var b strings.Builder
	b.WriteString("#EXTM3U\n#EXT-X-VERSION:5\n#EXT-X-PLAYLIST-TYPE:VOD\n#EXT-X-TARGETDURATION:5\n#EXT-X-MEDIA-SEQUENCE:0\n")
	for i, rng := range []string{"0-100", "101-end"} {
		if i > 0 {
			fmt.Fprintf(&b, "#EXT-X-DISCONTINUITY\n#EXT-X-KEY:METHOD=NONE\n#EXTINF:5,\n%s/v1/mp4/c(ts)/p(1_ad%%2Fcreative_ad%%2F720p%%2Fvideo_600.mp4)\n#EXT-X-DISCONTINUITY\n", s.srv.URL)
		}
		for n := 0; n < plutoSeg/2; n++ {
			fmt.Fprintf(&b, "#EXT-X-KEY:METHOD=AES-128,URI=\"%s%s%s/%s_keyfile_0.key\",IV=0x%032x\n#EXTINF:5,\n%s%s%s/%s-%05d.ts\n", s.srv.URL, plutoClip, rng, rendition, n+1, s.srv.URL, plutoClip, rng, rendition, n)
		}
	}
	fmt.Fprintf(&b, "#EXT-X-DISCONTINUITY\n#EXT-X-KEY:METHOD=NONE\n#EXTINF:5,\n%s/media/clip/bumper/1080pDRM/hls/0-end/hls_bumper-00000.ts\n#EXT-X-ENDLIST\n", s.srv.URL)
	fmt.Fprint(w, b.String())
}

func (s *plutoService) media(t *testing.T, w http.ResponseWriter, r *http.Request, p string) {
	t.Helper()
	switch {
	case strings.HasSuffix(p, ".m3u8") && s.refusePlaylists:
		http.Error(w, "", http.StatusForbidden)
	case strings.HasSuffix(p, ".m3u8"):
		rendition := strings.TrimSuffix(p[strings.LastIndex(p, "/")+1:], ".m3u8")
		var b strings.Builder
		b.WriteString("#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-PLAYLIST-TYPE:VOD\n#EXT-X-TARGETDURATION:5\n#EXT-X-MEDIA-SEQUENCE:0\n")
		for n := range plutoSeg {
			fmt.Fprintf(&b, "#EXT-X-KEY:METHOD=AES-128,URI=\"%s_keyfile_0.key\",IV=0x%032x\n#EXTINF:5.0,\n%s-%05d.ts\n", rendition, n+1, rendition, n)
		}
		b.WriteString("#EXT-X-ENDLIST\n")
		fmt.Fprint(w, b.String())
	case strings.HasSuffix(p, ".key"):
		_, _ = w.Write([]byte(plutoKey))
	case strings.HasSuffix(p, ".ts"):
		var n int
		_, err := fmt.Sscanf(p[strings.LastIndex(p, "-")+1:], "%d.ts", &n)
		if err != nil || n >= plutoSeg {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(plutoEncrypt(t, n))
	default:
		http.NotFound(w, r)
	}
}

func (s *plutoService) paths() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.hits...)
}

func useServiceFor(t *testing.T, s *plutoService) {
	t.Helper()
	prev := api.PlutoTVClientFn
	api.PlutoTVClientFn = func() *plutotv.Client { return plutotv.NewClientForTest(s.srv.Client(), s.srv.URL) }
	t.Cleanup(func() { api.PlutoTVClientFn = prev })
}

func deathNote() *models.Anime {
	return &models.Anime{Name: "Death Note", URL: plutotv.SeriesURL("6206bba315c8610013683974"), Source: "PlutoTV", MediaType: models.MediaTypeTV}
}

// downloadPlutoEpisode runs the -dm path for the first episode of the first
// season and returns the bytes written.
func downloadPlutoEpisode(t *testing.T, quality string) (file []byte, svc *plutoService, rendition string) {
	t.Helper()
	svc = newPlutoService(t)
	useServiceFor(t, svc)

	anime := deathNote()
	numbers, err := seriesSeasonsFor(anime)(anime)
	require.NoError(t, err)
	require.NotEmpty(t, numbers)
	eps, err := seriesEpisodesFor(anime)(anime, numbers[0])
	require.NoError(t, err)
	require.NotEmpty(t, eps)

	stream, err := providers.FetchStream(t.Context(), &eps[0], anime, quality)
	require.NoError(t, err)
	// What the player and the downloader are handed is GoAnime's own loopback
	// address for the playlist: the CDN's URL for it is refused by one of the
	// CDNs behind the name. The download paths know the address as their own.
	require.True(t, util.IsLocalProxyURL(stream.URL), "%s", stream.URL)
	require.True(t, player.LooksLikeHLS(stream.URL), "the downloader takes the HLS path for it")
	rendition = "hls_2400"
	if strings.Contains(playlistAt(t, stream.URL), "/hls_600-") {
		rendition = "hls_600"
	}

	out := filepath.Join(t.TempDir(), "episode.ts")
	headers := map[string]string{"User-Agent": "Mozilla/5.0 (test)", "Accept": "*/*", "Referer": stream.Referer}
	// The client the real download uses, so its guards are part of the test.
	require.NoError(t, hls.DownloadToFileWithClient(t.Context(), util.GetDownloadClient(), stream.URL, out, headers, nil))
	file, err = os.ReadFile(out)
	require.NoError(t, err)
	return file, svc, rendition
}

// playlistAt reads what a stream URL serves.
func playlistAt(t *testing.T, target string) string {
	t.Helper()
	resp, err := http.Get(target) //nolint:noctx // GoAnime's own loopback playlist
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(body)
}

func TestPlutoTVEpisodeDownloadsAsTheTitleAloneAndDecrypted(t *testing.T) {
	file, svc, rendition := downloadPlutoEpisode(t, "best")
	assert.Equal(t, "hls_2400", rendition)

	var want []byte
	for n := range plutoSeg {
		want = append(want, plutoPayload(n)...)
	}
	assert.True(t, bytes.Equal(want, file), "the file is the %d segments of the title, decrypted and in order (%d bytes, want %d)", plutoSeg, len(file), len(want))

	for _, p := range svc.paths() {
		assert.False(t, strings.HasPrefix(p, "/v1/mp4/"), "an ad creative was fetched: %s", p)
		assert.NotContains(t, p, "bumper", "the service's closing bumper was fetched")
	}
}

func TestPlutoTVDownloadFollowsTheQualityAsked(t *testing.T) {
	file, _, rendition := downloadPlutoEpisode(t, "worst")
	assert.Equal(t, "hls_600", rendition)
	assert.Len(t, file, plutoSeg*len(plutoPayload(0)))
}

func TestPlutoTVMovieIsOneDownloadableItem(t *testing.T) {
	movie := &models.Anime{Name: "Some Movie", URL: plutotv.MovieURL("62b9ea9ec4eb8f00140cb633"), Source: "PlutoTV", MediaType: models.MediaTypeMovie}
	eps, err := api.GetPlutoTVEpisodes(movie)
	require.NoError(t, err)
	require.Len(t, eps, 1, "-dm of a movie downloads exactly one item")
	assert.Equal(t, movie.URL, eps[0].URL)
}

// One of the CDNs behind Pluto TV's media host answers 403 for every playlist
// under a clip and still serves the segments. The download must not depend on
// which CDN answered: the title is then read out of the stitched playlist, and
// the file is the same title, whole and without the ads.
func TestPlutoTVDownloadsTheTitleWhenTheCDNRefusesItsPlaylist(t *testing.T) {
	svc := newPlutoService(t)
	svc.refusePlaylists = true
	useServiceFor(t, svc)

	anime := deathNote()
	eps, err := seriesEpisodesFor(anime)(anime, 1)
	require.NoError(t, err)
	stream, err := providers.FetchStream(t.Context(), &eps[0], anime, "best")
	require.NoError(t, err)
	require.True(t, util.IsLocalProxyURL(stream.URL))

	playlist := playlistAt(t, stream.URL)
	assert.Contains(t, playlist, plutoClip+"0-100/hls_2400-00000.ts", "the title's ranges, as the stitcher lists them")
	assert.Contains(t, playlist, plutoClip+"101-end/hls_2400-00001.ts")
	assert.NotContains(t, playlist, "/v1/mp4/", "without the ad between them")
	assert.NotContains(t, playlist, "bumper")

	out := filepath.Join(t.TempDir(), "episode.ts")
	headers := map[string]string{"User-Agent": "Mozilla/5.0 (test)", "Accept": "*/*"}
	require.NoError(t, hls.DownloadToFileWithClient(t.Context(), util.GetDownloadClient(), stream.URL, out, headers, nil))
	file, err := os.ReadFile(out)
	require.NoError(t, err)
	// Each range numbers its segments from zero: two ranges of two segments.
	var want []byte
	for range 2 {
		want = append(want, plutoPayload(0)...)
		want = append(want, plutoPayload(1)...)
	}
	assert.True(t, bytes.Equal(want, file), "the four segments of the title, decrypted, in order (%d bytes, want %d)", len(file), len(want))
	for _, p := range svc.paths() {
		assert.False(t, strings.HasPrefix(p, "/v1/mp4/"), "an ad creative was fetched: %s", p)
		assert.NotContains(t, p, "bumper")
	}
}
