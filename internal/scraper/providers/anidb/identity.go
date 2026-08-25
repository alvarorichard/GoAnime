package anidb

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
	// Tolerate a trailing "<slug>-<id>" without the /anime/ prefix.
	if i := strings.LastIndex(s, "-"); i >= 0 && isAllDigits(s[i+1:]) {
		return s[i+1:], nil
	}
	return "", netx.NewParserError(sourceLabel, "identity",
		fmt.Sprintf("not an anidb.app anime URL: %s", animeURL), nil)
}

// EpisodeID extracts the numeric episode id from a canonical episode URL, or
// accepts a bare id.
func EpisodeID(episodeURL string) (string, error) {
	s := strings.TrimSpace(episodeURL)
	if s == "" {
		return "", netx.NewParserError(sourceLabel, "identity", "empty episode URL", nil)
	}
	if isAllDigits(s) {
		return s, nil
	}
	if m := episodeHrefRe.FindStringSubmatch(s); m != nil {
		return m[1], nil
	}
	return "", netx.NewParserError(sourceLabel, "identity",
		fmt.Sprintf("not an anidb.app episode URL: %s", episodeURL), nil)
}

func (c *AniDBClient) animeURL(slug, id string) string {
	return fmt.Sprintf("%s/anime/%s-%s", c.baseURL, slug, id)
}

func (c *AniDBClient) episodeURL(id int) string {
	return fmt.Sprintf("%s/episode/%d", c.baseURL, id)
}

// preferredLanguage returns the language code to try first. Subbed by default;
// GOANIME_ANIDB_LANG accepts "jpn"/"sub" or "eng"/"dub".
func preferredLanguage() string {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("GOANIME_ANIDB_LANG"))) {
	case "eng", "dub", "dubbed":
		return "eng"
	default:
		return "jpn"
	}
}

// selectLanguage picks the preferred language, falling back to any other entry
// that carries an embed URL rather than failing the episode outright.
func selectLanguage(langs []apiLanguage, prefer string) (apiLanguage, bool) {
	for _, l := range langs {
		if strings.EqualFold(l.Code, prefer) && l.EmbedURL != "" {
			return l, true
		}
	}
	for _, l := range langs {
		if l.EmbedURL != "" {
			return l, true
		}
	}
	return apiLanguage{}, false
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
