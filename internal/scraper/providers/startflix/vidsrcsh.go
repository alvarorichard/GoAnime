package startflix

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
)

// vidsrcsh.go resolves the StartFlix panel's own vidsrc.sh player. The panel
// lists it as a data-source URL in two shapes — path-keyed
// (https://vidsrc.sh/embed/movie/tt…, https://vidsrc.sh/embed/tv/194583/2/3)
// and query-keyed (https://vidsrc.sh/embed/tv?tmdb=194583&season=1&episode=1,
// the series shape seen live that day) — and until 2026-10-21 it reached
// only the yt-dlp sweep, because ResolveStreamFor had no case for the host.
// But vidsrc.sh is the vidsrc family: it speaks the same vs_src protocol and
// keys the same ids (live on 2026-10-21 one panel URL resolved end to end
// through the existing chain in about two seconds), so the player URL
// already names the exact base and id the fallback chain walks on.
// resolveVIDSRCPanelPlayer enters that chain, unchanged, through the player
// URL.
//
// superflixbr.com was the first candidate for this layer and is a dead end
// (probed 2026-10-21): its players pass through an ad-gate whose real video
// only loads via a pop-under monetization funnel, with no id-keyed endpoint.
// The panel's vidsrc.sh embed is the fallback this file serves instead.

// vidsrcEmbedPathRe matches the embed path head: the kind, an optional id
// segment, and the optional season/episode tail. The id may also ride the
// query (?tmdb= / ?imdb=) with no path segment at all — the panel's series
// shape — so an empty match here still counts when the query keys it. The
// trailing ?ds_lang query is the panel's language hint, not a key.
var vidsrcEmbedPathRe = regexp.MustCompile(`^/embed/(movie|tv|series)(?:/(tt\d+|\d+))?(?:/(\d+))?(?:/(\d+))?/?$`)

// vidsrcEmbedIDRe names a keyable id: an IMDb id (tt…) or a TMDB id (digits).
var vidsrcEmbedIDRe = regexp.MustCompile(`^(?:tt)?\d+$`)

// vidsrcPanelKey reads the id the panel embed keys its title on, whichever
// shape the URL took: the path's id segment when there is one, otherwise the
// query (?imdb= for movies, ?tmdb= otherwise). Season and episode come from
// the path tail first, from the query when the path carries neither.
func vidsrcPanelKey(u *url.URL) (kind, id string, season, episode int, ok bool) {
	m := vidsrcEmbedPathRe.FindStringSubmatch(u.Path)
	if m == nil {
		return "", "", 0, 0, false
	}
	kind, id = m[1], m[2]
	if id == "" {
		q := u.Query()
		switch kind {
		case "movie":
			id = firstString(q.Get("imdb"), q.Get("tmdb"))
		default:
			id = firstString(q.Get("tmdb"), q.Get("imdb"))
		}
	}
	if !vidsrcEmbedIDRe.MatchString(id) {
		return "", "", 0, 0, false
	}
	if m[3] != "" || m[4] != "" {
		season, _ = strconv.Atoi(m[3])
		episode, _ = strconv.Atoi(m[4])
		return kind, id, season, episode, true
	}
	q := u.Query()
	season, _ = strconv.Atoi(q.Get("season"))
	episode, _ = strconv.Atoi(q.Get("episode"))
	return kind, id, season, episode, true
}

// firstString returns the first value that is not empty.
func firstString(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// isVIDSRCPanelPlayer reports whether a panel player URL is one the vidsrc
// family can key directly: the vidsrc.sh host with an embed path that names
// an id. Other hosts, non-embed paths and no-key embeds stay on the generic
// yt-dlp sweep.
func isVIDSRCPanelPlayer(u *url.URL) bool {
	if u == nil || u.Hostname() != "vidsrc.sh" {
		return false
	}
	_, _, _, _, ok := vidsrcPanelKey(u)
	return ok
}

// resolveVIDSRCPanelPlayer runs the vidsrc chain against the player's own
// base for the id the embed names — pinned there, not walked through
// vidsrcBasesFn, because the panel named this player on this host and the
// fallback's base-walk serves the no-panel flow. Without a named season and
// episode the chain probes the series' own defaults, exactly as the
// standalone fallback does. A movie key is the IMDb id (tt…) the chain
// prefers, a tv key the TMDB id.
func resolveVIDSRCPanelPlayer(ctx context.Context, u *url.URL) (*Stream, error) {
	kind, id, season, episode, ok := vidsrcPanelKey(u)
	if !ok {
		return nil, fmt.Errorf("not a keyable vidsrc embed: %s", u.RequestURI())
	}
	mediaKind := KindSeries
	if kind == "movie" {
		mediaKind = KindMovie
	}
	return resolveVIDSRCOn(ctx, u.Scheme+"://"+u.Host, mediaKind, id, season, episode)
}
