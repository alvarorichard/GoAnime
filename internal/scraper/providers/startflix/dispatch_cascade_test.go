package startflix

// dispatch_cascade_test.go chains whole resolvers as ResolveStreamFor sees
// them: two independent fake fakes (one per family) behind a host-routing
// transport, panel players naming real family hosts, and the walk deciding
// among dead players, wrong audio, and taller-but-wrong picture. The unit
// layer pins each resolver; this file pins the decisions they feed.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hostRouter routes by requested hostname: hosts a stub named (the panel's
// real family hosts) are rewritten onto their fake server's URL, anything
// else goes out as-is (the fakes' own http host, the panel fake, etc.).
type hostRouter struct{ byHost map[string]*httptest.Server }

func (h hostRouter) RoundTrip(req *http.Request) (*http.Response, error) {
	if srv, ok := h.byHost[req.URL.Hostname()]; ok {
		target, err := url.Parse(srv.URL)
		if err != nil {
			return nil, err
		}
		req.URL.Scheme, req.URL.Host, req.Host = target.Scheme, target.Host, target.Host
	}
	return http.DefaultTransport.RoundTrip(req)
}

// routeBothFamilies points the two family seams at one client for the test's
// lifetime: the upn resolver and the vidsrc chain (including its master
// prober) both ride the same host router, the way the panel's players sit
// on real hosts side by side.
func routeBothFamilies(t *testing.T, router hostRouter) {
	t.Helper()
	oldUpns, oldVIDSRC := upnsHTTPFn, vidsrcHTTPFn
	upnsHTTPFn = func() *http.Client { return &http.Client{Transport: router} }
	vidsrcHTTPFn = func() *http.Client { return &http.Client{Transport: router} }
	t.Cleanup(func() { upnsHTTPFn, vidsrcHTTPFn = oldUpns, oldVIDSRC })
}

// panelServer serves one players page listing the given buttons, the markup
// Players() parses from the panel's real pages.
func panelServer(t *testing.T, buttons string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(buttons))
	}))
	t.Cleanup(ts.Close)
	return ts
}

const (
	// The panel's real player hosts; the router rewrites them onto fakes.
	upnHost = "https://embedplayapiupn.upns.xyz/#8txjhj"
	vidHost = "https://vidsrc.sh/embed/movie/tt0315327?ds_lang=pt"
	// A simple dead video-stage message, the shape the reverse engineering
	// met for the live Dublado id.
	deadVideo = "Video not found or deleted"
)

// deadUpnFake is the on-the-wire state of a deleted upn video.
func deadUpnFake(t *testing.T) *httptest.Server {
	f := &fakeUPNsPlayer{Status: 404, Message: deadVideo}
	return f.serve(t)
}

// aliveUpnFake answers with one source at a given height (360 or 1080).
func aliveUpnFake(t *testing.T, height int) *httptest.Server {
	res := "1920x1080"
	if height == 360 {
		res = "640x360"
	}
	f := &fakeUPNsPlayer{
		Media: map[string]string{"/media/master.m3u8": masterAt(res)},
		Confs: []string{`{"player":true,"source":"{{SRC}}/media/master.m3u8"}`},
	}
	return f.serve(t)
}

// TestCascade_PanelPlayersToStream chains the panel parse to the walk: a
// players page listing a dead Dublado upn player and a working Legendado
// vidsrc player, queried as the user asks — the Dublado list. The vidsrc
// stream wins as the only survivor, carrying its listing's Legendado audio,
// which is exactly what sends the swap dialog to the user.
func TestCascade_PanelPlayersToStream(t *testing.T) {
	srvUpn := deadUpnFake(t)
	playerPage := fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tok1")
	f := &fakeVIDSRC{
		kind:         KindMovie,
		id:           "tt0315327",
		playerPages:  []string{playerPage},
		plainEnvJSON: `{"status_code":"200","data":{"stream_urls":["SRV/pl/cascade/master.m3u8"]}}`,
	}
	srvVIDSRC := f.start(t)

	router := hostRouter{byHost: map[string]*httptest.Server{
		"embedplayapiupn.upns.xyz": srvUpn,
		"vidsrc.sh":                srvVIDSRC,
	}}
	routeBothFamilies(t, router)

	psrv := panelServer(t,
		`<button data-show-player="true" data-source="https://embedplayapiupn.upns.xyz/#8txjhj" data-type="iframe">#1 Dublado</button>`+
			`<button data-show-player="true" data-source="https://vidsrc.sh/embed/movie/tt0315327?ds_lang=pt" data-type="iframe">Legendado</button>`)

	c := NewClientForTest(psrv.Client(), "https://www.startflix.test")
	players, err := c.Players(context.Background(), psrv.URL+"/episodio/x")
	require.NoError(t, err)
	require.Len(t, players, 2)
	assert.Equal(t, AudioDubbed, players[0].Audio(), "the parse reads the #1 Dublado label")
	assert.Equal(t, AudioSubtitled, players[1].Audio(), "the parse reads the Legendado label")

	stream, err := c.ResolveStreamFor(context.Background(), players, AudioDubbed)
	require.NoError(t, err)
	base, err := url.Parse(srvVIDSRC.URL)
	require.NoError(t, err)
	assert.Equal(t, base.String()+"/pl/cascade/master.m3u8?token=testtoken", stream.URL)
	assert.Equal(t, AudioSubtitled, stream.Audio,
		"the surviving stream keeps its listing's audio; the caller asks the user")
	assert.Equal(t, 1080, stream.Height)
	require.Len(t, f.dataFetched, 1, "one stream-data call; the movie keys off the panel URL's id")
	assert.Contains(t, f.dataFetched[0], "imdb=tt0315327")
}

// TestCascade_AudioBeatsHeight — the same two-alive panel watched in either
// audio. A Dublado listing tuned lower beats a taller Legendado when the
// user asked for the dub, and the tallest Legendado wins when they asked for
// the original; with no preference, height decides. The wire is identical in
// all three cases — only the walk's decision moves.
func TestCascade_AudioBeatsHeight(t *testing.T) {
	srvUpn := aliveUpnFake(t, 360)
	playerPage := fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tok1")
	f := &fakeVIDSRC{
		kind:         KindMovie,
		id:           "tt0315327",
		playerPages:  []string{playerPage},
		plainEnvJSON: `{"status_code":"200","data":{"stream_urls":["SRV/pl/tall/master.m3u8"]}}`,
	}
	srvVIDSRC := f.start(t)
	router := hostRouter{byHost: map[string]*httptest.Server{
		"embedplayapiupn.upns.xyz": srvUpn,
		"vidsrc.sh":                srvVIDSRC,
	}}
	routeBothFamilies(t, router)
	c := NewClientForTest(nil, "https://www.startflix.test")
	players := []Player{
		{URL: upnHost, Label: "#1 Dublado"},
		{URL: vidHost, Label: "Legendado"},
	}

	upnBase, err := url.Parse(srvUpn.URL)
	require.NoError(t, err)

	t.Run("want dub: short-but-dubbed wins over tall-but-original", func(t *testing.T) {
		stream, err := c.ResolveStreamFor(context.Background(), players, AudioDubbed)
		require.NoError(t, err)
		assert.Equal(t, upnBase.String()+"/media/master.m3u8", stream.URL)
		assert.Equal(t, AudioDubbed, stream.Audio)
		assert.Equal(t, 360, stream.Height, "audio preference is worth any height")
	})
	t.Run("want original: the tall one wins", func(t *testing.T) {
		stream, err := c.ResolveStreamFor(context.Background(), players, AudioSubtitled)
		require.NoError(t, err)
		assert.Contains(t, stream.URL, "/pl/tall/master.m3u8")
		assert.Equal(t, AudioSubtitled, stream.Audio)
		assert.Equal(t, 1080, stream.Height)
	})
	t.Run("no preference: the tallest wins", func(t *testing.T) {
		stream, err := c.ResolveStreamFor(context.Background(), players, "")
		require.NoError(t, err)
		assert.Contains(t, stream.URL, "/pl/tall/master.m3u8")
		assert.Equal(t, 1080, stream.Height)
	})
}

// TestCascade_AllPlayersDeadAggregates — every family player dead: one
// NoStreamError carrying both hosts' readings, the Dublado partner's verbatim
// message among them, so the app's dialog can say which walls it hit. No
// unsupported host is claimed (both players had dedicated resolvers), so the
// yt-dlp sweep stays out.
func TestCascade_AllPlayersDeadAggregates(t *testing.T) {
	srvUpn := deadUpnFake(t)
	playerPage := fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tok1")
	f := &fakeVIDSRC{
		kind:          KindMovie,
		id:            "tt0315327",
		playerPages:   []string{playerPage},
		plainEnvJSON:  `{"status_code":"200","data":{"stream_urls":["SRV/pl/dead/master.m3u8"]}}`,
		mastersBroken: true,
	}
	srvVIDSRC := f.start(t)
	router := hostRouter{byHost: map[string]*httptest.Server{
		"embedplayapiupn.upns.xyz": srvUpn,
		"vidsrc.sh":                srvVIDSRC,
	}}
	routeBothFamilies(t, router)
	c := NewClientForTest(nil, "https://www.startflix.test")
	players := []Player{
		{URL: upnHost, Label: "#1 Dublado"},
		{URL: vidHost, Label: "Legendado"},
	}

	require.Len(t, players, 2)
	_, err := c.ResolveStreamFor(context.Background(), players, AudioDubbed)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNoSupportedServer)
	var noStream *NoStreamError
	require.True(t, errors.As(err, &noStream))
	require.Len(t, noStream.Failures, 2, "both dead legs are named")
	require.Len(t, noStream.Unsupported, 0, "no host lacked a resolver; yt-dlp stays out")
	joined := ""
	for _, e := range noStream.Failures {
		joined += e.Error() + " | "
	}
	assert.Contains(t, joined, "embedplayapiupn.upns.xyz: upn: Video not found or deleted")
	assert.Contains(t, joined, "vidsrc.sh: vidsrc fallback: 1 URL(s) listed, none playable")
}

// TestCascade_UnusableHosts — malformed URLs are skipped silently, an
// unsupported host keeps its name for the aggregate, a players list with
// nothing usable fails with its own letter, and an empty list is ErrNoPlayers.
// Under the test binary the yt-dlp sweep is off, so the unsupported host
// fails without touching the network.
func TestCascade_UnusableHosts(t *testing.T) {
	c := NewClientForTest(nil, "https://www.startflix.test")
	ctx := context.Background()

	_, err := c.ResolveStreamFor(ctx, nil, AudioDubbed)
	require.ErrorIs(t, err, ErrNoPlayers)

	_, err = c.ResolveStreamFor(ctx, []Player{
		{URL: ":"},
		{URL: "https://vidsrcme.su/x", Label: "#1 Dublado"},
	}, AudioDubbed)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNoSupportedServer)
	var noStream *NoStreamError
	require.True(t, errors.As(err, &noStream))
	assert.Contains(t, noStream.Unsupported, "vidsrcme.su", "the unsupported host keeps its name")

	_, err = c.ResolveStreamFor(ctx, []Player{{URL: ":"}, {URL: "://"}}, AudioDubbed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no usable URL among 2 players")
}

// TestCascade_SupportedHostsNeverReachYtDlp — a supported family that dies
// must not hand its failure to the yt-dlp sweep: Unsupported stays empty,
// and the error remains the aggregate of the families' own readings.
func TestCascade_SupportedHostsNeverReachYtDlp(t *testing.T) {
	srvUpn := deadUpnFake(t)
	router := hostRouter{byHost: map[string]*httptest.Server{
		"embedplayapiupn.upns.xyz": srvUpn,
	}}
	routeBothFamilies(t, router)
	c := NewClientForTest(nil, "https://www.startflix.test")
	players := []Player{{URL: upnHost, Label: "#1 Dublado"}}

	_, err := c.ResolveStreamFor(context.Background(), players, AudioDubbed)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNoSupportedServer)
	var noStream *NoStreamError
	require.True(t, errors.As(err, &noStream))
	assert.Empty(t, noStream.Unsupported, "a supported family never becomes yt-dlp's problem")
	require.Len(t, noStream.Failures, 1)
}
