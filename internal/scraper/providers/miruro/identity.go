package miruro

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
)

// Package-level regexes: compiled once, per the provider contract.
var (
	// idRe is a Miruro anime id: 32 URL-safe base64 characters on every title
	// seen ("wHWR9m4VQ8GLW4GP9_ZTblQXn9f0phB8"). The lower bound is looser so
	// a longer or shorter id after a site change still parses.
	idRe = regexp.MustCompile(`^[A-Za-z0-9_-]{16,64}$`)
	// pageRe reads the id, and the slug when there is one, out of an /info or
	// /watch URL. Host-agnostic so a domain rotation only needs the const.
	pageRe = regexp.MustCompile(`/(?:info|watch)/([A-Za-z0-9_-]{16,64})(?:/([^/?#]+))?/?(?:[?#]|$)`)
	// epRe reads the episode number out of a watch URL.
	epRe = regexp.MustCompile(`[?&]ep=(\d+(?:\.\d+)?)(?:&|#|$)`)
	// variantRe reads one #EXT-X-STREAM-INF line's resolution height.
	variantRe = regexp.MustCompile(`RESOLUTION=\d+x(\d+)`)
	// qualityDigitsRe pulls "1080" out of "1080p".
	qualityDigitsRe = regexp.MustCompile(`(\d{3,4})`)
)

// AnimeID extracts the anime id from an /info or /watch URL, or accepts a bare
// id.
func AnimeID(animeURL string) (string, error) {
	s := strings.TrimSpace(animeURL)
	if s == "" {
		return "", netx.NewParserError(sourceLabel, "identity", "empty anime URL", nil)
	}
	if idRe.MatchString(s) {
		return s, nil
	}
	if m := pageRe.FindStringSubmatch(s); m != nil {
		return m[1], nil
	}
	return "", netx.NewParserError(sourceLabel, "identity",
		fmt.Sprintf("not a miruro.cx anime URL: %s", animeURL), nil)
}

// episodeRef is what a watch URL names: the anime, its slug and the episode.
type episodeRef struct {
	ID     string
	Slug   string
	Number string // as the site writes it: "1", "12.5"
}

// parseEpisodeURL reads a watch URL. The slug may be missing (the site
// redirects to the canonical one), the episode number may not.
func parseEpisodeURL(episodeURL string) (episodeRef, error) {
	s := strings.TrimSpace(episodeURL)
	if s == "" {
		return episodeRef{}, netx.NewParserError(sourceLabel, "identity", "empty episode URL", nil)
	}
	m := pageRe.FindStringSubmatch(s)
	if m == nil {
		return episodeRef{}, netx.NewParserError(sourceLabel, "identity",
			fmt.Sprintf("not a miruro.cx watch URL: %s", episodeURL), nil)
	}
	ep := epRe.FindStringSubmatch(s)
	if ep == nil {
		return episodeRef{}, netx.NewParserError(sourceLabel, "identity",
			fmt.Sprintf("not a miruro.cx episode URL (no ?ep= in %s)", episodeURL), nil)
	}
	return episodeRef{ID: m[1], Slug: m[2], Number: ep[1]}, nil
}

// animeURL is the title's page; GetAnimeEpisodes reads the id back out of it.
func (c *MiruroClient) animeURL(id string) string {
	return c.baseURL + "/info/" + id
}

// episodeURL is the page a human would open, and the only channel this scraper
// has back to itself: GetEpisodeStreamURL reads the id and number out of it.
func (c *MiruroClient) episodeURL(id, slug, number string) string {
	return fmt.Sprintf("%s%s?ep=%s", c.baseURL, watchPath(id, slug), number)
}

// watchPath is the watch page of a title. The site wants a slug segment but
// accepts any (it redirects to the right one), which is why a missing slug is
// written "x", as ani-cli does.
func watchPath(id, slug string) string {
	if slug == "" {
		slug = "x"
	}
	return "/watch/" + id + "/" + slug
}

// slugFromPath reads the slug out of a /watch/<id>/<slug> path.
func slugFromPath(path string) string {
	if m := pageRe.FindStringSubmatch(path); m != nil {
		return m[2]
	}
	return ""
}

// Audio as the site names its tracks. "ssub" carries the subtitles as
// separate files, "sub" burns them into the picture.
const (
	trackSoftSub = "ssub"
	trackHardSub = "sub"
	trackDub     = "dub"
)

// preferredAudio returns the audio to try first: subtitled by default;
// GOANIME_MIRURO_AUDIO accepts "sub"/"jpn" or "dub"/"eng".
func preferredAudio() string {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("GOANIME_MIRURO_AUDIO"))) {
	case "dub", "dubbed", "eng":
		return trackDub
	default:
		return trackHardSub
	}
}

// normalizeQuality turns "1080p", "1080", "hd" into a pixel height, and returns
// 0 for "best"/"" (meaning: leave the choice to the player).
func normalizeQuality(quality string) int {
	q := strings.ToLower(strings.TrimSpace(quality))
	if q == "" || q == "best" || q == "auto" {
		return 0
	}
	if m := qualityDigitsRe.FindStringSubmatch(q); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil {
			return n
		}
	}
	return 0
}

// resolveRef resolves a playlist-relative reference against the playlist URL;
// an absolute one comes back as it is.
func resolveRef(base, ref string) string {
	b, err := url.Parse(base)
	if err != nil {
		return ref
	}
	r, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return b.ResolveReference(r).String()
}
