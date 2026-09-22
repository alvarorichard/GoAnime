package hianime

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
)

// AnimeID extracts the numeric anime id from a permalink, or accepts a bare id.
func AnimeID(animeURL string) (string, error) {
	s := strings.TrimSpace(animeURL)
	if s == "" {
		return "", netx.NewParserError(sourceLabel, "identity", "empty anime URL", nil)
	}
	if isAllDigits(s) {
		return s, nil
	}
	if m := animeHrefRe.FindStringSubmatch(s); m != nil {
		return m[2], nil
	}
	return "", netx.NewParserError(sourceLabel, "identity",
		fmt.Sprintf("not a hianime.at anime URL: %s", animeURL), nil)
}

// EpisodeID extracts the episode id from a watch URL's ?ep=, or accepts a bare
// id.
//
// Note that an episode's id is NOT the anime's: a watch URL carries both, and
// picking the wrong one lists somebody else's servers. Hence the deliberate
// refusal below rather than a fallback to the trailing number.
func EpisodeID(episodeURL string) (string, error) {
	s := strings.TrimSpace(episodeURL)
	if s == "" {
		return "", netx.NewParserError(sourceLabel, "identity", "empty episode URL", nil)
	}
	if isAllDigits(s) {
		return s, nil
	}
	if m := episodeIDRe.FindStringSubmatch(s); m != nil {
		return m[1], nil
	}
	return "", netx.NewParserError(sourceLabel, "identity",
		fmt.Sprintf("not a hianime.at episode URL (no ?ep= in %s)", episodeURL), nil)
}

func (c *HiAnimeClient) animeURL(slug, id string) string {
	return fmt.Sprintf("%s/%s-%s", c.baseURL, slug, id)
}

// episodeURL is the page a human would open, and the only channel this scraper
// has back to itself: GetEpisodeStreamURL reads the episode id out of it later.
func (c *HiAnimeClient) episodeURL(animeSlug, animeID string, episodeID int) string {
	return fmt.Sprintf("%s/watch/%s-%s?ep=%d", c.baseURL, animeSlug, animeID, episodeID)
}

// Audio tracks, as the server list labels them. "hsub" appears on some titles
// alongside the usual pair.
const (
	audioSub = "sub"
	audioDub = "dub"
)

// preferredAudio returns the audio track to try first. Subtitled by default;
// GOANIME_HIANIME_AUDIO accepts "sub"/"jpn" or "dub"/"eng".
func preferredAudio() string {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("GOANIME_HIANIME_AUDIO"))) {
	case "dub", "dubbed", "eng":
		return audioDub
	default:
		return audioSub
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

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// resolveRef resolves a playlist-relative reference against the playlist URL.
func resolveRef(base, ref string) string {
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ref
	}
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

// originOf reduces a URL to scheme://host, which is what the stream CDN wants
// as a Referer.
//
// The CDN enforces it on part of its URLs, not all of them. Sampled 2026-09-22
// across three titles with freshly resolved links: the master playlist answered
// 200 either way, while the variant playlist below it answered 403 without a
// Referer on two of the three. So a player that omits it appears to work and
// then stalls on the title where it matters; sending it always costs nothing.
func originOf(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}
