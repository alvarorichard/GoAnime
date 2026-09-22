package hianime

import (
	"context"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// zokoServerName is the only server this scraper reads. doc.go explains why the
// others are left alone.
const zokoServerName = "ZokoAnime"

// server is one entry of the episode's server list.
type server struct {
	Name  string // "ZokoAnime", "HD-1", …
	Audio string // "sub", "dub", "hsub"
	Embed string // decoded from data-hash
}

// GetEpisodeStreamURL resolves an episode URL to a playable HLS URL plus the
// metadata the player needs.
//
// quality may be "best", "" or a label such as "720p"/"1080". When a matching
// variant exists its playlist is returned; otherwise the master playlist is,
// which lets the player pick.
//
// The metadata map carries, besides the usual "source":
//
//	referer    the origin the stream CDN checks (see originOf); some of its
//	           URLs answer 403 without it
//	audio_lang the track that was actually picked, which may not be the one
//	           preferred if the episode does not carry it
//	subtitles  a JSON array of {"url","language","label"}, the shape
//	           source_providers.go feeds to util.SetGlobalSubtitles
func (c *HiAnimeClient) GetEpisodeStreamURL(ctx context.Context, episodeURL, quality string) (streamURL string, metadata map[string]string, err error) {
	epID, err := EpisodeID(episodeURL)
	if err != nil {
		return "", nil, err
	}

	serversURL := c.restURL("episode/servers?episodeId=" + epID)
	fragment, err := c.getFragment(ctx, serversURL, "servers", episodeURL)
	if err != nil {
		return "", nil, err
	}
	servers, err := parseServers(fragment)
	if err != nil {
		return "", nil, err
	}

	chosen, ok := selectServer(servers, preferredAudio())
	if !ok {
		return "", nil, netx.NewParserError(sourceLabel, "servers",
			"episode "+epID+" lists "+strconv.Itoa(len(servers))+" server(s), none of them "+
				zokoServerName+" (the only one this source can read)", nil)
	}
	util.Debug("HiAnime server selected", "episode", epID, "server", chosen.Name, "audio", chosen.Audio)

	embedOrigin := originOf(chosen.Embed)
	page, err := c.getBody(ctx, chosen.Embed, "embed", func(req *http.Request) {
		req.Header.Set("Referer", c.baseURL+"/")
	})
	if err != nil {
		return "", nil, err
	}
	payload, err := decodePayload(page)
	if err != nil {
		return "", nil, err
	}

	// Variant selection has to send the Referer too: the variant playlist is
	// precisely the level where the CDN was measured to enforce it.
	referer := embedOrigin + "/"
	streamURL = payload.Src
	if v, ok := c.selectVariant(ctx, payload.Src, quality, referer); ok {
		streamURL = v
	}

	metadata = map[string]string{
		"source":     "hianime",
		"referer":    referer,
		"audio_lang": chosen.Audio,
	}
	if subs := encodeSubtitles(payload.Subtitles); subs != "" {
		metadata["subtitles"] = subs
	}
	return streamURL, metadata, nil
}

// parseServers reads the server list fragment.
func parseServers(fragment string) ([]server, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(fragment))
	if err != nil {
		return nil, netx.NewParserError(sourceLabel, "servers", "failed to parse server list", err)
	}

	var out []server
	doc.Find(".server-item").Each(func(_ int, s *goquery.Selection) {
		hash := strings.TrimSpace(s.AttrOr("data-hash", ""))
		if hash == "" {
			return
		}
		// data-hash is plain base64 of the embed URL — no cipher, despite the
		// name the site gives the attribute.
		embed, decErr := base64.StdEncoding.DecodeString(hash)
		if decErr != nil {
			return
		}
		out = append(out, server{
			Name:  strings.TrimSpace(s.AttrOr("data-server-name", "")),
			Audio: strings.ToLower(strings.TrimSpace(s.AttrOr("data-type", ""))),
			Embed: strings.TrimSpace(string(embed)),
		})
	})

	if len(out) == 0 {
		return nil, netx.NewParserError(sourceLabel, "servers",
			"no servers in the list fragment (layout changed?)", nil)
	}
	return out, nil
}

// selectServer picks the readable server for the preferred audio track, falling
// back to any other track rather than failing an episode that exists in only
// one language.
func selectServer(servers []server, prefer string) (server, bool) {
	for _, s := range servers {
		if strings.EqualFold(s.Name, zokoServerName) && s.Audio == prefer && s.Embed != "" {
			return s, true
		}
	}
	for _, s := range servers {
		if strings.EqualFold(s.Name, zokoServerName) && s.Embed != "" {
			return s, true
		}
	}
	return server{}, false
}

// encodeSubtitles renders the player's subtitle list into the metadata map.
//
// The payload's own "lang" field is unreliable — every track on the episodes
// sampled claimed "en", including the Portuguese and Spanish ones — so the
// label is what gets carried as the language too. The label is what the picker
// shows, and a wrong two-letter code would make PT-BR unselectable.
func encodeSubtitles(subs []embedSubtitle) string {
	type track struct {
		URL      string `json:"url"`
		Language string `json:"language"`
		Label    string `json:"label"`
	}
	out := make([]track, 0, len(subs))
	for _, s := range subs {
		if strings.TrimSpace(s.Src) == "" {
			continue
		}
		label := strings.TrimSpace(s.Label)
		if label == "" {
			label = strings.TrimSpace(s.Lang)
		}
		out = append(out, track{URL: s.Src, Language: label, Label: label})
	}
	if len(out) == 0 {
		return ""
	}
	encoded, err := jsonx.Marshal(out)
	if err != nil {
		util.Debug("HiAnime could not encode subtitles", "error", err)
		return ""
	}
	return string(encoded)
}

// selectVariant reads the master playlist and returns the variant whose
// resolution matches the requested quality. Reports false for "best"/"" and
// whenever the requested height is absent, so the caller keeps the master.
func (c *HiAnimeClient) selectVariant(ctx context.Context, masterURL, quality, referer string) (string, bool) {
	want := normalizeQuality(quality)
	if want == 0 {
		return "", false
	}
	body, err := c.getBody(ctx, masterURL, "playlist", func(req *http.Request) {
		req.Header.Set("Referer", referer)
	})
	if err != nil {
		util.Debug("HiAnime could not read master playlist; falling back to it", "error", err)
		return "", false
	}

	lines := strings.Split(string(body), "\n")
	for i, line := range lines {
		m := variantRe.FindStringSubmatch(line)
		if m == nil || i+1 >= len(lines) {
			continue
		}
		height, convErr := strconv.Atoi(m[1])
		if convErr != nil || height != want {
			continue
		}
		next := strings.TrimSpace(lines[i+1])
		if next == "" || strings.HasPrefix(next, "#") {
			continue
		}
		return resolveRef(masterURL, next), true
	}
	util.Debug("HiAnime quality not available; using master playlist", "requested", quality)
	return "", false
}
