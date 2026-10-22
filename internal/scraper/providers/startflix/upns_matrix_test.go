package startflix

// upns_matrix_test.go is the unit layer over the upn family: every helper
// the resolver is built from gets a table, and every failure path the wire
// can produce gets named. The e2e layer (upns_test.go) pins the happy path
// and the retry; this file pins the edges one component at a time, so a
// regression names its piece instead of a mystery end to end.

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// masterAt renders a master playlist of one variant at the given
// RESOLUTION, the way the family's masters look to the prober.
func masterAt(res string) string {
	return "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=2000000,RESOLUTION=" + res + "\nv.m3u8\n"
}

// TestUPNsPanelKeyTable — every fragment/query shape the panel has produced
// or the SPA's own reader accepts, plus the shapes that must stay keyless.
func TestUPNsPanelKeyTable(t *testing.T) {
	t.Parallel()
	cases := []struct {
		raw  string
		want string // "" means no key
	}{
		{"https://embedplayapiupn.upns.xyz/#8txjhj", "8txjhj"},
		{"https://embedplayapiupn.upns.xyz/#8txjhj/", "8txjhj"},
		{"https://embedplayapiupn.upns.xyz/?id=8txjhj", "8txjhj"},
		{"https://embedplayapiupn.upns.xyz/#8txjhj?id=other", "8txjhj?id=other"}, // the fragment wins, spoils and all
		{"https://embedplayapiupn.upns.xyz/#/", ""},                              // trims to nothing
		{"https://embedplayapiupn.upns.xyz/#", ""},
	}
	for _, tc := range cases {
		u, err := url.Parse(tc.raw)
		require.NoError(t, err)
		key, ok := upnsPanelKey(u)
		if tc.want == "" {
			assert.False(t, ok, tc.raw)
			continue
		}
		require.True(t, ok, tc.raw)
		assert.Equal(t, tc.want, key, tc.raw)
	}

	for _, raw := range []string{"", ":", "https://"} {
		u, err := url.Parse(raw)
		if err != nil {
			continue
		}
		_, ok := upnsPanelKey(u)
		assert.False(t, ok, "a keyless parse stays keyless: %q", raw)
	}

	_, ok := upnsPanelKey(nil)
	assert.False(t, ok, "a nil URL carries no key")
}

// TestUPNsSourceURLTable — the two source forms and every knob the player's
// tu() applies: params from the object and the adjust layer, the /hlsmod/
// domain rewrite, and the refusals (no url, unparseable).
func TestUPNsSourceURLTable(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   any
		adj  any
		want string
		ok   bool
	}{
		{"plain string", "https://h.example/m.m3u8", nil, "https://h.example/m.m3u8", true},
		{"empty string", "", nil, "", false},
		{"object with url only", map[string]any{"url": "https://h.example/m.m3u8", "proxy": true},
			nil, "https://h.example/m.m3u8", true},
		{"object without url", map[string]any{"params": map[string]any{"t": "1"}}, nil, "", false},
		{"object params", map[string]any{"url": "https://h.example/m.m3u8",
			"params": map[string]any{"t": "a1"}}, nil, "https://h.example/m.m3u8?t=a1", true},
		{"adjust params fill in", map[string]any{"url": "https://h.example/m.m3u8"},
			map[string]any{"params": map[string]any{"t": "a1"}}, "https://h.example/m.m3u8?t=a1", true},
		{"object params win over adjust", map[string]any{"url": "https://h.example/m.m3u8",
			"params": map[string]any{"t": "obj"}}, map[string]any{"params": map[string]any{"t": "adj"}},
			"https://h.example/m.m3u8?t=obj", true},
		{"params merge, object wins per key", map[string]any{"url": "https://h.example/m.m3u8",
			"params": map[string]any{"t": "obj"}}, map[string]any{"params": map[string]any{"u": "adj", "t": "adj2"}},
			"https://h.example/m.m3u8?t=obj&u=adj", true},
		{"params never clobber the query they came with", map[string]any{
			"url": "https://h.example/m.m3u8?x=orig", "params": map[string]any{"x": "new"}},
			nil, "https://h.example/m.m3u8?x=orig", true},
		{"domain rewrites an /hls/ path", map[string]any{"url": "https://h.example/hls/k/m.m3u8",
			"domain": "cdn.example"}, nil, "https://h.example/hlsmod/cdn.example/k/m.m3u8", true},
		{"domain without /hls/ changes nothing", map[string]any{"url": "https://h.example/v/k/m.m3u8",
			"domain": "cdn.example"}, nil, "https://h.example/v/k/m.m3u8", true},
		{"adjust domain rewrites too", map[string]any{"url": "https://h.example/hls/k/m.m3u8"},
			map[string]any{"domain": "adj.example"}, "https://h.example/hlsmod/adj.example/k/m.m3u8", true},
		{"unparseable url refused", map[string]any{"url": "ht tps://h.example/m.m3u8"}, nil, "", false},
	}
	for _, tc := range cases {
		got, ok := upnsSourceURL(tc.in, tc.adj)
		require.Equal(t, tc.ok, ok, tc.name)
		assert.Equal(t, tc.want, got, tc.name)
	}
}

// TestUPNsCandidatesTable — the walk: order honored, unknown names ignored,
// duplicates dropped once, Cloudflare preferring its native-HLS twin, and
// the fixed crawl order when the config names none.
func TestUPNsCandidatesTable(t *testing.T) {
	t.Parallel()
	same := "https://dup.example/a.m3u8"
	build := func(order []string, cfg *upnsVideoConfig) []string {
		if len(order) > 0 {
			cfg.StreamingConfig = &struct {
				Order  []string                  `json:"order"`
				Adjust map[string]map[string]any `json:"adjust"`
			}{Order: order}
		}
		return upnsCandidates(cfg)
	}
	t.Run("config order decides and unknown names are skipped", func(t *testing.T) {
		cfg := &upnsVideoConfig{
			CF:     "https://cf.example/a.m3u8",
			Source: "https://in.example/a.m3u8",
		}
		assert.Equal(t, []string{"https://in.example/a.m3u8", "https://cf.example/a.m3u8"},
			build([]string{"In-House", "Nonsense", "Cloudflare"}, cfg))
	})
	t.Run("duplicates across keys collapse once", func(t *testing.T) {
		cfg := &upnsVideoConfig{
			CF:       same,
			CFNative: same, // the same URL under both halves of Cloudflare
		}
		assert.Equal(t, []string{same}, build([]string{"Cloudflare"}, cfg))
	})
	t.Run("no order falls back to the crawl order", func(t *testing.T) {
		cfg := &upnsVideoConfig{
			CFNative: "https://cf.example/a.m3u8",
			TikTok:   "https://tok.example/a.m3u8",
			Google:   "https://gw.example/a.m3u8",
			Source:   "https://in.example/a.m3u8",
		}
		assert.Equal(t, []string{
			"https://cf.example/a.m3u8", "https://tok.example/a.m3u8",
			"https://gw.example/a.m3u8", "https://in.example/a.m3u8",
		}, build(nil, cfg))
	})
	t.Run("an unfilled source key yields nothing", func(t *testing.T) {
		cfg := &upnsVideoConfig{TikTok: "https://tok.example/a.m3u8"}
		assert.Equal(t, []string{"https://tok.example/a.m3u8"},
			build([]string{"Google", "Tiktok"}, cfg))
	})
}

// TestUPNsVideoURLShape — the wire query the SPA's runtime established
// (id/w/h/r), including the empty-r contract and the panel-env override the
// embedder host takes its name from.
func TestUPNsVideoURLShape(t *testing.T) {
	t.Setenv("GOANIME_STARTFLIX_PANEL_URL", "https://www.painel-zzz.sbs")
	assert.Equal(t,
		"https://embedplayapiupn.upns.xyz/api/v1/video?id=8txjhj&w=1280&h=720&r=painel-zzz.sbs",
		upnsVideoURL("https://embedplayapiupn.upns.xyz", "8txjhj", ""))

	t.Setenv("GOANIME_STARTFLIX_PANEL_URL", "https://painel-slash.sbs/") // trailing slash trimmed
	assert.Equal(t,
		"https://embedplayapiupn.upns.xyz/api/v1/video?id=k&w=1280&h=720&r=painel-slash.sbs&capacityToken=t",
		upnsVideoURL("https://embedplayapiupn.upns.xyz", "k", "&capacityToken=t"))

	// The SPA appends &r= even when its referrer read is empty.
	t.Setenv("GOANIME_STARTFLIX_PANEL_URL", "   ")
	assert.Equal(t,
		"https://embedplayapiupn.upns.xyz/api/v1/video?id=k&w=1280&h=720&r=painel-aso.sbs",
		upnsVideoURL("https://embedplayapiupn.upns.xyz", "k", ""), "blank env falls back to the default panel")
}

// TestUPNsDecryptFailsMatrix — every body shape that must refuse to
// decrypt: not hex, unblocked, empty, odd half-byte, or padding that does
// not verify.
func TestUPNsDecryptFailsMatrix(t *testing.T) {
	t.Parallel()
	oddities := []string{
		"",                                  // empty
		"zzzz",                              // not hex
		"abc",                               // odd half-byte
		hex.EncodeToString(make([]byte, 4)), // unblocked length
	}
	for _, body := range oddities {
		_, err := upnsDecryptHex(body)
		assert.Error(t, err, "body %q must not decrypt", body)
	}

	// A decryptable block whose last byte claims pad=3 while the two bytes
	// before it disagree: decryption succeeds, padding is a lie.
	block, err := aes.NewCipher([]byte(upnsAESKey))
	require.NoError(t, err)
	plain := make([]byte, aes.BlockSize)
	plain[aes.BlockSize-1] = 3
	plain[aes.BlockSize-2] = 9
	plain[aes.BlockSize-3] = 1
	out := make([]byte, aes.BlockSize)
	cipher.NewCBCEncrypter(block, []byte(upnsAESIV)).CryptBlocks(out, plain)
	_, err = upnsDecryptHex(hex.EncodeToString(out))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad PKCS#7 padding")
}

// TestUPNsGetBodyErrorMatrix walks every wire failure the resolver can meet
// and pins the message each leaves behind (the aggregate failure reads
// these verbatim; the "upn:" prefix belongs to all of them).
func TestUPNsGetBodyErrorMatrix(t *testing.T) {
	t.Run("plain 500 with no JSON body", func(t *testing.T) {
		f := &fakeUPNsPlayer{Status: 500, Message: ""}
		ts := f.serve(t)
		fakeUPNsSeams(t, ts)
		u, _ := url.Parse(ts.URL + "/fake/#8txjhj")
		_, err := resolveUPNsPanelPlayer(context.Background(), u)
		require.Error(t, err)
		assert.Equal(t, "upn: HTTP 500", err.Error())
	})

	t.Run("403 JSON message is surfaced verbatim", func(t *testing.T) {
		f := &fakeUPNsPlayer{Status: 403, Message: "Video not found or deleted"}
		ts := f.serve(t)
		fakeUPNsSeams(t, ts)
		u, _ := url.Parse(ts.URL + "/fake/#8txjhj")
		_, err := resolveUPNsPanelPlayer(context.Background(), u)
		require.Error(t, err)
		assert.Equal(t, "upn: Video not found or deleted", err.Error())
	})

	t.Run("200 with a non-hex body", func(t *testing.T) {
		f := &fakeUPNsPlayer{Raw: "<html>decoy</html>"}
		ts := f.serve(t)
		fakeUPNsSeams(t, ts)
		u, _ := url.Parse(ts.URL + "/fake/#8txjhj")
		_, err := resolveUPNsPanelPlayer(context.Background(), u)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unreadable response")
	})

	t.Run("200 hex carrying non-JSON plaintext", func(t *testing.T) {
		f := &fakeUPNsPlayer{Raw: string(upnsEncryptHex(t, []byte("this is not a config")))}
		ts := f.serve(t)
		fakeUPNsSeams(t, ts)
		u, _ := url.Parse(ts.URL + "/fake/#8txjhj")
		_, err := resolveUPNsPanelPlayer(context.Background(), u)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not the player config")
	})
}

// TestUPNsResolverChoosesTallest — the prober walks every candidate in the
// config's own order and the tallest playable wins; a dead candidate does
// not sink the walk.
func TestUPNsResolverChoosesTallest(t *testing.T) {
	f := &fakeUPNsPlayer{
		Media: map[string]string{
			"/media/tok.m3u8": masterAt("1920x1080"),
			"/media/g.m3u8":   masterAt("640x360"),
			// /media/cf.m3u8 absent on purpose: the walk must survive it.
		},
		Confs: []string{`{"player":true,"title":"t",
			"hlsVideoTiktok":{"url":"{{SRC}}/media/tok.m3u8","params":{"tok":"xyz"}},
			"hlsVideoGoogle":"{{SRC}}/media/g.m3u8",
			"cf":"{{SRC}}/media/cf.m3u8",
			"streamingConfig":{"order":["Tiktok","Google","Cloudflare"]}}`},
	}
	ts := f.serve(t)
	fakeUPNsSeams(t, ts)
	u, err := url.Parse(ts.URL + "/fake/#8txjhj")
	require.NoError(t, err)

	stream, err := resolveUPNsPanelPlayer(context.Background(), u)
	require.NoError(t, err)
	assert.Equal(t, ts.URL+"/media/tok.m3u8?tok=xyz", stream.URL)
	assert.Equal(t, 1080, stream.Height)
	assert.Equal(t, []string{"/media/tok.m3u8", "/media/g.m3u8", "/media/cf.m3u8"},
		f.MediaHits, "every candidate is probed, in the config's own order")
}

// TestUPNsGoldenConfigFixture — a full-shape payload (five source keys,
// streamingConfig.order with adjust) round-trips as the protocol's golden:
// order honored, adjust merged into the object's params, probe order
// pinned, and the stream built from the winner.
func TestUPNsGoldenConfigFixture(t *testing.T) {
	f := &fakeUPNsPlayer{
		Media: map[string]string{
			"/media/tok.m3u8": masterAt("1280x720"),
			"/media/gw.m3u8":  masterAt("1920x1080"),
			"/media/cf.m3u8":  masterAt("426x240"),
		},
		Confs: []string{`{"player":true,"title":"Velozes e Furiosos","source":"{{SRC}}/media/in.m3u8",
			"cf":"{{SRC}}/media/cf.m3u8",
			"cfNative":"{{SRC}}/media/cf.m3u8",
			"hlsVideoGoogle":{"url":"{{SRC}}/media/gw.m3u8","params":{"t":"gg"}},
			"hlsVideoTiktok":{"url":"{{SRC}}/media/tok.m3u8"},
			"streamingConfig":{"order":["Tiktok","Google","Cloudflare","In-House"],
				"adjust":{"Google":{"params":{"u":"extra"}}}}}`},
	}
	// /media/in.m3u8 has no Media entry on purpose: a dead In-House tail
	// must not sink the walk that already found taller water.
	ts := f.serve(t)
	fakeUPNsSeams(t, ts)
	u, err := url.Parse(ts.URL + "/fake/#8txjhj")
	require.NoError(t, err)

	stream, err := resolveUPNsPanelPlayer(context.Background(), u)
	require.NoError(t, err)
	assert.Equal(t, ts.URL+"/media/gw.m3u8?t=gg&u=extra", stream.URL)
	assert.Equal(t, 1080, stream.Height)
	assert.Equal(t, []string{"/media/tok.m3u8", "/media/gw.m3u8", "/media/cf.m3u8", "/media/in.m3u8"},
		f.MediaHits, "the probe walks the config's own order across all four keys")
	assert.Equal(t, ts.URL+"/", stream.Referer)
	assert.Empty(t, stream.Audio, "the panel label decides, not this file")
}

// TestUPNsHappyDefaultShape pins the default master body: no Media, no
// Master override, the standard config — exactly the live default.
func TestUPNsHappyDefaultShape(t *testing.T) {
	f := &fakeUPNsPlayer{}
	ts := f.serve(t)
	fakeUPNsSeams(t, ts)
	u, _ := url.Parse(ts.URL + "/fake/#8txjhj")
	stream, err := resolveUPNsPanelPlayer(context.Background(), u)
	require.NoError(t, err)
	assert.Equal(t, ts.URL+"/media/master.m3u8", stream.URL)
	assert.Equal(t, 1080, stream.Height)
	assert.Equal(t, []string{"/media/master.m3u8"}, f.MediaHits,
		"the default body still records its probe")
}

// TestUPNsNoSourcesAndAllMastersDead — the two graceful refusals: a config
// without any source, and a source that names a master that never loads.
func TestUPNsNoSourcesAndAllMastersDead(t *testing.T) {
	t.Run("config carries no source", func(t *testing.T) {
		f := &fakeUPNsPlayer{Confs: []string{`{"player":true,"title":"t"}`}}
		ts := f.serve(t)
		fakeUPNsSeams(t, ts)
		u, _ := url.Parse(ts.URL + "/fake/#8txjhj")
		_, err := resolveUPNsPanelPlayer(context.Background(), u)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no sources listed")
	})
	t.Run("the master refuses to load", func(t *testing.T) {
		f := &fakeUPNsPlayer{
			Confs: []string{`{"player":true,"source":"{{SRC}}/media/one.m3u8"}`},
			Media: map[string]string{
				// Any path but the configured one would 404 under a set
				// Media map; give one known path a body so only the
				// configured candidate dies.
				"/media/unused.m3u8": masterAt("640x360"),
			},
		}
		ts := f.serve(t)
		fakeUPNsSeams(t, ts)
		u, _ := url.Parse(ts.URL + "/fake/#8txjhj")
		_, err := resolveUPNsPanelPlayer(context.Background(), u)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "1 source(s) listed, none playable")
	})
}

// TestUPNsSaturatedRidesTheToken — the capacity-token retry is taken even
// when it then resolves nowhere: the retried request must carry the pair,
// and the "not ready" verdict of the retried config must surface.
func TestUPNsSaturatedRidesTheToken(t *testing.T) {
	f := &fakeUPNsPlayer{Confs: []string{
		`{"player":true,"delivery":{"inHouse":"saturated","capacityToken":"CT","capacityTokenExpire":"8"},"cf":"{{SRC}}/media/tok.m3u8"}`,
		`{"player":false}`}}
	ts := f.serve(t)
	fakeUPNsSeams(t, ts)
	u, _ := url.Parse(ts.URL + "/fake/#8txjhj")
	_, err := resolveUPNsPanelPlayer(context.Background(), u)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not ready yet")
	require.Len(t, f.Video, 2)
	assert.Contains(t, f.Video[1], "capacityToken=CT")
	assert.Contains(t, f.Video[1], "capacityTokenExpire=8")
}

// TestUPNsContextCancelled — a caller walking away mid-resolution stops the
// resolver at the wire, not after it.
func TestUPNsContextCancelled(t *testing.T) {
	f := &fakeUPNsPlayer{}
	ts := f.serve(t)
	fakeUPNsSeams(t, ts)
	u, _ := url.Parse(ts.URL + "/fake/#8txjhj")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := resolveUPNsPanelPlayer(ctx, u)
	require.Error(t, err)
	assert.Empty(t, f.Video, "a cancelled context makes no wire calls")
}

// TestUPNsWithoutKey — a keyless URL on the family's host names its own
// contract, with the aggregated "upn:" prefix the dispatcher's failure list
// reads.
func TestUPNsWithoutKey(t *testing.T) {
	f := &fakeUPNsPlayer{}
	ts := f.serve(t)
	fakeUPNsSeams(t, ts)
	u, _ := url.Parse(ts.URL + "/fake/")
	_, err := resolveUPNsPanelPlayer(context.Background(), u)
	require.Error(t, err)
	assert.Equal(t, "upn: without a key: /fake/", err.Error())
}

// TestUPNsGetBodyStatusNoMessage — a non-200 whose body is not the message
// envelope leaves the bare HTTP code, still prefixed for the aggregate.
func TestUPNsGetBodyStatusNoMessage(t *testing.T) {
	f := &fakeUPNsPlayer{Status: 503, Message: ""}
	ts := f.serve(t)
	fakeUPNsSeams(t, ts)
	u, _ := url.Parse(ts.URL + "/fake/#8txjhj")
	_, err := resolveUPNsPanelPlayer(context.Background(), u)
	require.Error(t, err)
	assert.Equal(t, "upn: HTTP 503", err.Error())
}
