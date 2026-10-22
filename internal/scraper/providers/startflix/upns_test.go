package startflix

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// upnsEncryptHex is the server side of the family's contract the tests stand
// up: PKCS#7-pad the plaintext, AES-128-CBC it with the player's fixed key
// and IV, and hex-encode it — exactly the body shape /api/v1/video answers
// with (the key/IV pair was captured from the SPA's own runtime and the
// pipeline round-tripped against it live on 2026-10-21).
func upnsEncryptHex(t *testing.T, plain []byte) []byte {
	t.Helper()
	block, err := aes.NewCipher([]byte(upnsAESKey))
	require.NoError(t, err)
	pad := aes.BlockSize - len(plain)%aes.BlockSize
	padded := append(bytes.Clone(plain), bytes.Repeat([]byte{byte(pad)}, pad)...)
	out := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, []byte(upnsAESIV)).CryptBlocks(out, padded)
	return []byte(hex.EncodeToString(out))
}

// TestUPNsPanelPlayerMapping pins which panel player URLs belong to the upn
// family: any *.upns.xyz host keyed by its fragment — the live shape the
// panel listed (embedplayapiupn.upns.xyz/#8txjhj, Dublado #2 for Velozes e
// Furiosos) — with an ?id= query as a courtesy form. Other hosts, and keyless
// URLs, stay unsupported.
func TestUPNsPanelPlayerMapping(t *testing.T) {
	for _, raw := range []string{
		"https://embedplayapiupn.upns.xyz/#8txjhj",
		"https://embedplayupn.upns.xyz/#abc12",
		"https://embedplayapiupn.upns.xyz/?id=8txjhj",
	} {
		u, err := url.Parse(raw)
		require.NoError(t, err)
		assert.True(t, isUPNsPanelPlayer(u), raw)
	}
	for _, raw := range []string{
		"https://embedplayapiupn.upns.xyz/",   // front page, no key anywhere
		"https://embedplayapiupn.upns.xyz/#/", // stripped to nothing
		"https://upn.example.org/#8txjhj",     // another host entirely
		"https://upnsxyz.example/#8txjhj",     // suffix must end the hostname
	} {
		u, err := url.Parse(raw)
		require.NoError(t, err)
		assert.False(t, isUPNsPanelPlayer(u), raw)
	}
}

// TestUPNsDecryptHex pins the pipeline at its sharpest edge: the same
// key/IV pair both sides use, block alignment, and the refusal to guess
// about a body that is not hex — which is how a proxy's injected HTML would
// arrive.
func TestUPNsDecryptHex(t *testing.T) {
	plain := []byte(`{"title":"Velozes e Furiosos"}`)
	out, err := upnsDecryptHex(string(upnsEncryptHex(t, plain)))
	require.NoError(t, err)
	assert.Equal(t, plain, out)

	_, err = upnsDecryptHex("<html>decoy</html>")
	assert.Error(t, err, "non-hex body must not decrypt")

	_, err = upnsDecryptHex(hex.EncodeToString(make([]byte, 15)))
	assert.Error(t, err, "unblocked ciphertext must be refused")
}

// fakeUPNsPlayer is one httptest server speaking both halves of the family's
// wire: the encrypted /api/v1/video config stage and the HLS master the
// config points at. Configs are staged in order ("|STAGE|" separated) and
// every video-stage hit is recorded, so tests can pin the request shape
// (id/w/h/r) the SPA's own flow established — including the capacity-token
// refresh replay.
type fakeUPNsPlayer struct {
	Confs   []string // plaintext configs, one per response ("|STAGE|" to vary)
	Video   []string // query strings of each /api/v1/video hit
	Referer string   // last video-stage Referer header
	Status  int      // when set, the video stage answers this code with a JSON message
	Message string   // the plain-JSON error body's message
	Raw     string   // when set, the video stage answers 200 with this raw body

	// Media overrides the master stage by exact path ("/media/hi.m3u8"
	// → playlist body); a path with no entry answers 404. nil keeps the
	// default: every /media/ path serves a 1080p master. Master, when set,
	// is that default body instead of the 1080p one.
	Media     map[string]string
	Master    string
	MediaHits []string // media paths probed, in order
}

func (f *fakeUPNsPlayer) serve(t *testing.T) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/video":
			f.Video = append(f.Video, r.URL.RawQuery)
			f.Referer = r.Header.Get("Referer")
			if f.Status != 0 {
				w.WriteHeader(f.Status)
				_, _ = fmt.Fprintf(w, `{"message":%q}`, f.Message)
				return
			}
			if f.Raw != "" {
				_, _ = w.Write([]byte(f.Raw))
				return
			}
			idx := len(f.Video) - 1
			defaultConf := fmt.Sprintf(`{"player":true,"title":"Velozes e Furiosos","source":%q}`,
				baseOf(r)+"/media/master.m3u8")
			stages := strings.Split(strings.Join(f.Confs, "|STAGE|"), "|STAGE|")
			if len(stages) == 1 && stages[0] == "" {
				stages = []string{defaultConf}
			}
			if idx >= len(stages) {
				idx = len(stages) - 1
			}
			plain := []byte(strings.ReplaceAll(stages[idx], "{{SRC}}", baseOf(r)))
			w.Header().Set("Content-Type", "application/octet-stream")
			_, _ = w.Write(upnsEncryptHex(t, plain))
		case strings.HasPrefix(r.URL.Path, "/media/"):
			f.MediaHits = append(f.MediaHits, r.URL.Path)
			if f.Media != nil {
				if body, ok := f.Media[r.URL.Path]; ok {
					w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
					_, _ = w.Write([]byte(body))
				} else {
					http.NotFound(w, r)
				}
				return
			}
			if f.Master != "" {
				w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
				_, _ = w.Write([]byte(f.Master))
				return
			}
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			_, _ = w.Write([]byte("#EXTM3U\n" +
				"#EXT-X-STREAM-INF:BANDWIDTH=4500000,RESOLUTION=1920x1080\n" +
				"1080.m3u8\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(ts.Close)
	return ts
}

func baseOf(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

func fakeUPNsSeams(t *testing.T, ts *httptest.Server) {
	t.Helper()
	oldUpns, oldVIDSRC := upnsHTTPFn, vidsrcHTTPFn
	upnsHTTPFn, vidsrcHTTPFn = ts.Client, ts.Client
	t.Cleanup(func() { upnsHTTPFn, vidsrcHTTPFn = oldUpns, oldVIDSRC })
}

// TestUPNsPanelPlayerResolve walks one player URL end to end through the
// fake: fragment key, the SPA's request shape (id/w/h/r as its runtime
// showed live), the AES envelope, the master prober's height read, and a
// stream that carries the player's own base as referrer.
func TestUPNsPanelPlayerResolve(t *testing.T) {
	f := &fakeUPNsPlayer{}
	ts := f.serve(t)
	fakeUPNsSeams(t, ts)

	player := ts.URL + "/fake/#8txjhj"
	u, err := url.Parse(player)
	require.NoError(t, err)

	stream, err := resolveUPNsPanelPlayer(context.Background(), u)
	require.NoError(t, err)
	require.NotNil(t, stream)
	assert.Equal(t, ts.URL+"/media/master.m3u8", stream.URL)
	assert.Equal(t, strings.TrimPrefix(ts.URL, "http://"), stream.Host)
	assert.Equal(t, ts.URL+"/", stream.Referer)
	assert.Equal(t, 1080, stream.Height)
	assert.Empty(t, stream.Audio, "the panel label decides, not this file")
	require.Len(t, f.Video, 1)
	assert.Equal(t, fmt.Sprintf("id=8txjhj&w=%d&h=%d&r=painel-aso.sbs",
		upnsViewportW, upnsViewportH), f.Video[0])
	assert.Equal(t, ts.URL+"/", f.Referer)
}

// TestUPNsPanelPlayerSaturated runs the family's own retry contract: a
// "saturated" delivery re-fetches once with the capacity token pair, then
// resolves.
func TestUPNsPanelPlayerSaturated(t *testing.T) {
	f := &fakeUPNsPlayer{
		Confs: []string{
			`{"player":true,"delivery":{"inHouse":"saturated","capacityToken":"ct1","capacityTokenExpire":"900000"}}`,
			`{"player":true,"source":"{{SRC}}/media/master.m3u8"}`,
		},
	}
	ts := f.serve(t)
	fakeUPNsSeams(t, ts)

	u, err := url.Parse(ts.URL + "/fake/#8txjhj")
	require.NoError(t, err)

	stream, err := resolveUPNsPanelPlayer(context.Background(), u)
	require.NoError(t, err)
	assert.Equal(t, ts.URL+"/media/master.m3u8", stream.URL)
	require.Len(t, f.Video, 2)
	assert.NotContains(t, f.Video[0], "capacityToken")
	assert.Contains(t, f.Video[1], "capacityToken=ct1")
	assert.Contains(t, f.Video[1], "capacityTokenExpire=900000")
}

// TestUPNsPanelPlayerDeadVol records the state the reverse engineering met
// live: the panel's key exists (info still decrypts) but the video stage
// answers its "removed" message, which the resolver surfaces verbatim so
// the aggregate failure names the dead Dublado source instead of a mystery
// host.
func TestUPNsPanelPlayerDeadVol(t *testing.T) {
	f := &fakeUPNsPlayer{Status: http.StatusNotFound,
		Message: "Video not found or deleted"}
	ts := f.serve(t)
	fakeUPNsSeams(t, ts)

	u, err := url.Parse(ts.URL + "/fake/#8txjhj")
	require.NoError(t, err)

	_, err = resolveUPNsPanelPlayer(context.Background(), u)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Video not found or deleted")
	require.Len(t, f.Video, 1, "the dead id answers its message from the video stage")
}

// TestUPNsCandidatesOrder pins the source walk: the config's own
// streamingConfig.order decides, the Cloudflare key prefers its native-HLS
// twin, the object form carries params and the /hlsmod/ rewrite, and a
// missing order falls back to the player's fixed crawl order.
func TestUPNsCandidatesOrder(t *testing.T) {
	cfg := &upnsVideoConfig{
		CF:     "https://cf.example/master.m3u8",
		Source: "https://inhouse.example/hls/k/master.m3u8",
		Google: map[string]any{
			"url":    "https://gw.example/master.m3u8",
			"params": map[string]any{"t": "a1"},
		},
		StreamingConfig: &struct {
			Order  []string                  `json:"order"`
			Adjust map[string]map[string]any `json:"adjust"`
		}{
			Order: []string{"In-House", "Cloudflare", "Google"},
			Adjust: map[string]map[string]any{
				"Google": {"domain": "gw.example"},
			},
		},
	}
	assert.Equal(t, []string{
		"https://inhouse.example/hls/k/master.m3u8",
		"https://cf.example/master.m3u8",
		"https://gw.example/master.m3u8?t=a1",
	}, upnsCandidates(cfg))

	cfg2 := &upnsVideoConfig{
		CF:       "https://cf.example/master.m3u8",
		CFNative: "https://cfnative.example/master.m3u8",
		TikTok:   "https://tok.example/master.m3u8",
	}
	assert.Equal(t, []string{
		"https://cfnative.example/master.m3u8",
		"https://tok.example/master.m3u8",
	}, upnsCandidates(cfg2))

	// The object form's own domain rewrites /hls/ paths the player's way.
	cfg3 := &upnsVideoConfig{
		Source: map[string]any{
			"url":    "https://inhouse.example/hls/k/master.m3u8",
			"domain": "inhouse.example",
		},
	}
	assert.Equal(t, []string{
		"https://inhouse.example/hlsmod/inhouse.example/k/master.m3u8",
	}, upnsCandidates(cfg3))
}

// TestUPNsPanelPlayerNotReady — the SPA bails when the config says the
// video exists but the player is not staged; so does the resolver.
func TestUPNsPanelPlayerNotReady(t *testing.T) {
	f := &fakeUPNsPlayer{
		Confs: []string{`{"player":false,"title":"t"}`},
	}
	ts := f.serve(t)
	fakeUPNsSeams(t, ts)

	u, err := url.Parse(ts.URL + "/fake/#8txjhj")
	require.NoError(t, err)

	_, err = resolveUPNsPanelPlayer(context.Background(), u)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not ready yet")
}
