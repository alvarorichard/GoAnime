package startflix

// family_isolation_test.go is the meta layer: the properties the other
// tests rely on without saying it. No panel URL may belong to two families
// at once (double routing would make two resolvers fight over one player),
// every failure the upn resolver leaves must carry the "upn:" prefix the
// aggregate failure reads, and the dispatch ranking must order the families
// exactly as the resolver choice assumes.

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// familySamples are one representative panel player URL per family, the
// shapes the panel has actually listed (2026-10-21 evidence in comments
// across the resolvers).
var familySamples = []struct {
	name string
	url  string
	typ  string // Player.Type, for the direct-file family
}{
	{"byse", "https://playerbyse.example/v/abc123", ""},
	{"abyss", "https://playembedapi.site/?v=naqbaSRZ5", ""},
	{"vidsrc.sh", "https://vidsrc.sh/embed/movie/tt0315327?ds_lang=pt", ""},
	{"upns", "https://embedplayapiupn.upns.xyz/#8txjhj", ""},
	{"direct file", "https://cdn.example/files/film.mp4", "jwplayer"},
}

// knows reports the family predicates for one player; byse and abyss share
// a URL-level predicate, direct files ride Player.Type.
func knows(p Player) map[string]bool {
	u, err := url.Parse(p.URL)
	if err != nil || u.Host == "" {
		return nil
	}
	return map[string]bool{
		"byse":        isByseHost(u.Host),
		"abyss":       isAbyssHost(u.Host),
		"vidsrc.sh":   isVIDSRCPanelPlayer(u),
		"upns":        isUPNsPanelPlayer(u),
		"direct file": p.IsDirectFile(),
	}
}

// TestFamilyPredicatesDisjointPerPlayer — every sample belongs to exactly
// one family. A second "true" here would mean ResolveStreamFor has two
// claims on one player and its switch hides which one wins.
func TestFamilyPredicatesDisjointPerPlayer(t *testing.T) {
	t.Parallel()
	for _, s := range familySamples {
		know := knows(Player{URL: s.url, Type: s.typ})
		require.NotNil(t, know, "%s must parse as a player URL", s.name)
		var claimed []string
		for fam, yes := range know {
			if yes {
				claimed = append(claimed, fam)
			}
		}
		require.Len(t, claimed, 1, "%s claimed by: %v", s.name, claimed)
		assert.Equal(t, s.name, claimed[0], s.url)
	}
}

// TestFamilySwapsStayForeign — each family predicate must refuse every
// other family's sample, so a rename or a new host suffix cannot quietly
// stretch a mapping.
func TestFamilySwapsStayForeign(t *testing.T) {
	t.Parallel()
	for _, s := range familySamples {
		for _, other := range familySamples {
			if other.name == s.name {
				continue
			}
			know := knows(Player{URL: other.url, Type: other.typ})
			assert.False(t, know[s.name],
				"%s must not claim %s's URL %q", s.name, other.name, other.url)
		}
	}
}

// TestPlayerRankOrder — the dispatch ranking the resolver choice assumes:
// Byse (one JSON call) first, Abyss next (cheap page), direct files after
// (the panel's own <video>, usually dead), embeds the rest; anything
// unparseable sinks last so it never blocks the walk.
func TestPlayerRankOrder(t *testing.T) {
	t.Parallel()
	byse := Player{URL: "https://playerbyse.example/v/a"}
	abyss := Player{URL: "https://playembedapi.site/?v=v"}
	direct := Player{URL: "https://cdn.example/f.mp4", Type: "plyr"}
	other := Player{URL: "https://vidsrc.sh/embed/movie/tt0315327"}
	broken := Player{URL: "ht tps://cdn.example/f.mp4"}

	assert.Less(t, playerRank(byse), playerRank(abyss))
	assert.Less(t, playerRank(abyss), playerRank(direct))
	assert.Less(t, playerRank(direct), playerRank(other))
	assert.Equal(t, 4, playerRank(broken), "an unparseable URL sinks last")
}

// TestUPNsErrorPrefixContract — every failure the upn resolver surfaces
// names its family. The dispatcher's aggregate failure reads these strings
// and the app's dialog quotes them (the 2026-10-21 dialog showed the upn
// verbatim message), so a bare "HTTP 404" would be a regression.
func TestUPNsErrorPrefixContract(t *testing.T) {
	cases := []struct {
		name string
		fake func(t *testing.T) *fakeUPNsPlayer
		key  string // the player URL fragment; "" means none
	}{
		{"dead id", func(t *testing.T) *fakeUPNsPlayer {
			return &fakeUPNsPlayer{Status: 404, Message: deadVideo}
		}, "8txjhj"},
		{"http 500", func(t *testing.T) *fakeUPNsPlayer {
			return &fakeUPNsPlayer{Status: 500}
		}, "8txjhj"},
		{"non-hex body", func(t *testing.T) *fakeUPNsPlayer {
			return &fakeUPNsPlayer{Raw: "<html>decoy</html>"}
		}, "8txjhj"},
		{"non-JSON plaintext", func(t *testing.T) *fakeUPNsPlayer {
			return &fakeUPNsPlayer{Raw: string(upnsEncryptHex(t, []byte("x")))}
		}, "8txjhj"},
		{"no source", func(t *testing.T) *fakeUPNsPlayer {
			return &fakeUPNsPlayer{Confs: []string{`{"player":true}`}}
		}, "8txjhj"},
		{"not ready", func(t *testing.T) *fakeUPNsPlayer {
			return &fakeUPNsPlayer{Confs: []string{`{"player":false}`}}
		}, "8txjhj"},
		{"no key on the URL", func(t *testing.T) *fakeUPNsPlayer {
			return &fakeUPNsPlayer{}
		}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := tc.fake(t)
			ts := f.serve(t)
			fakeUPNsSeams(t, ts)
			player := ts.URL + "/fake/" + tc.key
			u, err := url.Parse(player)
			require.NoError(t, err)
			_, err = resolveUPNsPanelPlayer(context.Background(), u)
			require.Error(t, err)
			assert.True(t, stringsHasUpn(err), "error %q is not family-branded", err)
		})
	}
}

// TestVIDSRCPanelErrorContract — the vidsrc chain's own refusals name their
// layer, so the aggregate can tell an "unreadable stream envelope" from an
// API outage without a fixture per wording.
func TestVIDSRCPanelErrorContract(t *testing.T) {
	playerPage := fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tok1")
	f := &fakeVIDSRC{
		kind:         KindMovie,
		id:           "tt0315327",
		playerPages:  []string{playerPage},
		plainEnvJSON: `{"status_code":"503","data":{"stream_urls":[]}}`,
	}
	vidsrcHTTPClientForTest(t, f.start(t))
	u, err := url.Parse("https://vidsrc.sh/embed/movie/tt0315327")
	require.NoError(t, err)
	_, err = resolveVIDSRCPanelPlayer(context.Background(), u)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "vidsrc fallback: stream API answered 503",
		"the verdict names the stream API's own status")

	f2 := &fakeVIDSRC{
		kind:          KindMovie,
		id:            "tt0315327",
		playerPages:   []string{playerPage},
		plainEnvJSON:  `{"status_code":"200","data":{"stream_urls":["SRV/pl/x/master.m3u8"]}}`,
		mastersBroken: true,
	}
	vidsrcHTTPClientForTest(t, f2.start(t))
	_, err = resolveVIDSRCPanelPlayer(context.Background(), u)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "1 URL(s) listed, none playable")
}

// TestKeylessFamilyHostLandsUnsupported — the upn mapping is "host AND
// key": a family host carrying no key at all is not the family's to key,
// and the dispatcher hands it to the generic sweep instead (the Unsupported
// list), not to a resolver that would bail. The resolver's own keyless
// verdict is pinned at unit level (TestUPNsWithoutKey).
func TestKeylessFamilyHostLandsUnsupported(t *testing.T) {
	f := &fakeUPNsPlayer{}
	ts := f.serve(t)
	fakeUPNsSeams(t, ts)
	c := NewClientForTest(nil, "https://www.startflix.test")
	_, err := c.ResolveStreamFor(context.Background(),
		[]Player{{URL: "https://embedplayapiupn.upns.xyz/", Label: "#2 Dublado"}}, AudioDubbed)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNoSupportedServer)
	var noStream *NoStreamError
	require.True(t, errors.As(err, &noStream))
	assert.Contains(t, noStream.Unsupported, "embedplayapiupn.upns.xyz",
		"a keyless family host is the generic sweep's business, not the resolver's")
	assert.Empty(t, noStream.Failures, "no dedicated resolver ran, so nothing failed in it")
	assert.Empty(t, f.Video, "the resolver is never called for a keyless URL")
}

// stringsHasUpn reports whether an error is family-branded.
func stringsHasUpn(err error) bool {
	return len(err.Error()) >= 5 && err.Error()[:5] == "upn: "
}
