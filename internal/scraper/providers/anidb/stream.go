package anidb

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

// apiLanguage mirrors one entry of /api/frontend/episode/<id>/languages.
type apiLanguage struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	EmbedURL string `json:"embed_url"`
}

type apiLanguagesResponse struct {
	Languages []apiLanguage `json:"languages"`
}

// GetEpisodeStreamURL resolves an episode URL to a playable HLS URL plus the
// metadata the player needs (referer, audio language).
//
// quality may be "best", "" or a label such as "720p"/"1080". When a matching
// variant exists its playlist is returned; otherwise the master playlist is,
// which lets the player pick.
func (c *AniDBClient) GetEpisodeStreamURL(ctx context.Context, episodeURL, quality string) (streamURL string, metadata map[string]string, err error) {
	epID, err := EpisodeID(episodeURL)
	if err != nil {
		return "", nil, err
	}

	apiURL := fmt.Sprintf("%s/api/frontend/episode/%s/languages", c.baseURL, epID)
	var payload apiLanguagesResponse
	if err := c.getJSON(ctx, apiURL, "languages", &payload); err != nil {
		return "", nil, err
	}
	lang, ok := selectLanguage(payload.Languages, preferredLanguage())
	if !ok {
		return "", nil, netx.NewParserError(sourceLabel, "languages",
			fmt.Sprintf("no playable language for episode %s", epID), nil)
	}
	util.Debug("AniDB language selected", "episode", epID, "code", lang.Code, "name", lang.Name)

	master, err := c.masterPlaylistURL(ctx, lang.EmbedURL)
	if err != nil {
		return "", nil, err
	}

	streamURL = master
	if v, ok := c.selectVariant(ctx, master, quality); ok {
		streamURL = v
	}

	metadata = map[string]string{
		"source":     "anidb",
		"referer":    c.baseURL + "/",
		"audio_lang": lang.Code,
	}
	return streamURL, metadata, nil
}

// masterPlaylistURL fetches an embed page and pulls the HLS master URL out of
// the player config.
func (c *AniDBClient) masterPlaylistURL(ctx context.Context, embedURL string) (string, error) {
	if embedURL == "" {
		return "", netx.NewParserError(sourceLabel, "embed", "empty embed URL", nil)
	}
	body, err := c.getBody(ctx, embedURL, "embed")
	if err != nil {
		return "", err
	}
	m := masterPlaylistRe.FindSubmatch(body)
	if m == nil {
		return "", netx.NewParserError(sourceLabel, "embed",
			"no m3u8 in embed page (player layout changed?)", nil)
	}
	return string(m[1]), nil
}

// selectVariant reads the master playlist and returns the variant whose
// resolution matches the requested quality. Reports false for "best"/"" and
// whenever the requested height is absent, so the caller keeps the master.
func (c *AniDBClient) selectVariant(ctx context.Context, masterURL, quality string) (string, bool) {
	want := normalizeQuality(quality)
	if want == 0 {
		return "", false
	}
	body, err := c.getBody(ctx, masterURL, "playlist")
	if err != nil {
		util.Debug("AniDB could not read master playlist; falling back to it", "error", err)
		return "", false
	}

	lines := strings.Split(string(body), "\n")
	for i, line := range lines {
		m := variantRe.FindStringSubmatch(line)
		if m == nil || i+1 >= len(lines) {
			continue
		}
		height, err := strconv.Atoi(m[1])
		if err != nil || height != want {
			continue
		}
		next := strings.TrimSpace(lines[i+1])
		if next == "" || strings.HasPrefix(next, "#") {
			continue
		}
		return resolveRef(masterURL, next), true
	}
	util.Debug("AniDB quality not available; using master playlist", "requested", quality)
	return "", false
}
