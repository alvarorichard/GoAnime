package api

import (
	"errors"
	"strings"
	"sync/atomic"

	"github.com/alvarorichard/Goanime/internal/util"
)

// AniList can take its GraphQL API down without taking the site down. When it
// does, graphql.anilist.co answers EVERY request — any User-Agent, any query —
// with HTTP 403 and a body reading
//
//	"The AniList API has been temporarily disabled due to severe stability issues."
//
// Confirmed live 2026-09-09: the API UA, no UA, curl's default and a browser UA
// all get the same 403, while anilist.co itself serves 200. So this is not the
// User-Agent filtering documented on netx.APIUserAgent (where a browser UA was
// rejected and an API UA accepted) — that distinction no longer applies while
// the API is off, and no header change can work around it.
//
// Two things follow, and both are handled here rather than at each call site:
// the failure has to be *named* so a user reading the log can tell an upstream
// outage from a broken install, and it has to be *remembered* so a session does
// not re-ask an API that has already said it is switched off.

// ErrAniListAPIDisabled reports that AniList has turned its API off upstream.
// Nothing on this side can fix it; callers should fall back to another metadata
// source rather than treat it as a lookup miss.
var ErrAniListAPIDisabled = errors.New(
	"AniList API is temporarily disabled upstream (not a GoAnime problem); using MyAnimeList metadata instead")

// aniListDisabledMarkers are substrings of AniList's own outage message. Two
// independent fragments are matched so a reworded notice still trips at least
// one, and so an unrelated 403 (a proxy, a captive portal) does not.
var aniListDisabledMarkers = []string{
	"api has been temporarily disabled",
	"severe stability issues",
}

// aniListDisabled latches once the outage has been observed, so the rest of the
// session skips AniList entirely.
//
// Latched rather than time-boxed on purpose: an outage announced by the API
// itself is not a blip that clears in seconds, and a binge would otherwise pay
// a doomed round trip per title (and per search variation within a title). A
// restart is the natural reset — the alternative, a timed retry, adds a knob to
// tune for a case where the answer is already known.
var aniListDisabled atomic.Bool

// bodySaysAniListDisabled reports whether an AniList error body is the upstream
// "API disabled" notice rather than an ordinary rejection.
func bodySaysAniListDisabled(body []byte) bool {
	lower := strings.ToLower(string(body))
	for _, marker := range aniListDisabledMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

// noteAniListDisabled records the outage, logging once so a session says it
// plainly without repeating itself for every title.
func noteAniListDisabled() {
	if aniListDisabled.CompareAndSwap(false, true) {
		util.Debug("AniList has disabled its API upstream; skipping it for the rest of this session")
	}
}

// aniListIsDisabled reports whether the outage has already been seen.
func aniListIsDisabled() bool { return aniListDisabled.Load() }

// resetAniListDisabledForTest clears the latch. Test-only: the flag is
// process-wide, so a test that trips it would otherwise leak into its
// neighbours.
func resetAniListDisabledForTest() { aniListDisabled.Store(false) }
