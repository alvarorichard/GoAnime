package miruro

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/alvarorichard/Goanime/v2/internal/util/jsonx"
)

// sourcesData is the watch page's list of sources for one episode: per audio
// track, the providers the site aggregates, each with its servers.
type sourcesData struct {
	Tracks        []sourceTrack `json:"tracks"`
	ProviderOrder []string      `json:"providerOrder"`
}

type sourceTrack struct {
	Track     string           `json:"track"` // "ssub", "sub", "dub"
	Providers []sourceProvider `json:"providers"`
}

type sourceProvider struct {
	Provider  string           `json:"provider"`
	Servers   []sourceServer   `json:"servers"`
	Subtitles []sourceSubtitle `json:"subtitles"`
}

// sourceServer is one mirror. Headers are what its CDN checks: some answer
// 403 without the Referer, one family also wants the Origin (measured
// 2026-11-26). A server with only an embed page and no streams is skipped.
type sourceServer struct {
	Server  string            `json:"server"`
	Headers map[string]string `json:"headers"`
	Streams []sourceStream    `json:"streams"`
}

type sourceStream struct {
	URL    string `json:"url"`
	Format string `json:"format"` // "hls" on every server seen; "mp4" is accepted too
}

type sourceSubtitle struct {
	Language string `json:"language"`
	Label    string `json:"label"`
	File     string `json:"file"`
	Default  bool   `json:"default"`
}

// candidate is one stream that could play the episode, in the order tried.
type candidate struct {
	track     string
	provider  string
	server    string
	url       string
	format    string
	referer   string
	origin    string
	subtitles []sourceSubtitle // offered to the player; empty unless the track is ssub
}

// maxCandidates bounds how many streams one episode tries before giving up,
// so a page full of dead mirrors fails in seconds rather than minutes.
const maxCandidates = 8

// GetEpisodeStreamURL resolves a watch URL to a playable stream plus the
// metadata the player needs.
//
// The page offers many mirrors; candidates orders them and the first whose
// playlist answers is used, so a dead mirror does not fail the episode.
// quality may be "best", "" or a label such as "720p"/"1080"; when the master
// playlist has that height its variant is returned, otherwise the master.
//
// The metadata map carries, besides "source":
//
//	referer, origin  the headers the stream's CDN checks
//	audio_lang       "sub" or "dub": what was actually picked, which may not be
//	                 the preference when the episode lacks it
//	subtitles        a JSON array of {"url","language","label"}
func (c *Client) GetEpisodeStreamURL(ctx context.Context, episodeURL, quality string) (streamURL string, metadata map[string]string, err error) {
	ref, err := parseEpisodeURL(episodeURL)
	if err != nil {
		return "", nil, err
	}
	page, _, err := c.fetchWatchPage(ctx, ref.ID, ref.Slug, ref.Number, "sources")
	if err != nil {
		return "", nil, err
	}
	if page.SourcesData == nil || len(page.SourcesData.Tracks) == 0 {
		// The site lists the episode and has nothing to play for it.
		return "", nil, fmt.Errorf("%s episode %s: %w", sourceLabel, ref.Number, netx.ErrMediaOffline)
	}

	cands := candidates(page.SourcesData, preferredAudio())
	if len(cands) == 0 {
		return "", nil, netx.NewParserError(sourceLabel, "sources",
			"episode "+ref.Number+" lists only embed players, no stream this source can read", nil)
	}

	for i, cand := range cands {
		if i == maxCandidates {
			break
		}
		if err := ctx.Err(); err != nil {
			return "", nil, err
		}
		body, ok := c.probe(ctx, cand)
		if !ok {
			continue
		}
		util.Debug("Miruro stream selected", "episode", ref.Number, "track", cand.track,
			"provider", cand.provider, "server", cand.server)

		streamURL = cand.url
		if v, ok := selectVariant(body, cand.url, quality); ok {
			streamURL = v
		}
		metadata = map[string]string{
			"source":     "miruro",
			"audio_lang": audioOf(cand.track),
		}
		if cand.referer != "" {
			metadata["referer"] = cand.referer
		}
		if cand.origin != "" {
			metadata["origin"] = cand.origin
		}
		if subs := encodeSubtitles(cand.subtitles); subs != "" {
			metadata["subtitles"] = subs
		}
		return streamURL, metadata, nil
	}
	return "", nil, netx.NewParserError(sourceLabel, "sources",
		"none of the "+strconv.Itoa(min(len(cands), maxCandidates))+" streams tried for episode "+ref.Number+" answered", nil)
}

// audioOf names a track the way the metadata contract does: the two subtitled
// tracks are both "sub".
func audioOf(track string) string {
	if track == trackDub {
		return trackDub
	}
	return trackHardSub
}

// planStep is one pass over a track. needSubs restricts the pass to providers
// that ship an English or Portuguese subtitle file: an ssub stream without one
// plays with no subtitles at all, which is worse than the burned-in kind.
type planStep struct {
	track    string
	needSubs bool
}

// trackPlan is the order the tracks are tried in for an audio preference.
func trackPlan(prefer string) []planStep {
	subbed := []planStep{
		{track: trackSoftSub, needSubs: true},
		{track: trackHardSub},
		{track: trackSoftSub},
	}
	if prefer == trackDub {
		return append([]planStep{{track: trackDub}}, subbed...)
	}
	return append(subbed, planStep{track: trackDub})
}

// candidates lists the streams of an episode in the order worth trying them:
// the tracks by trackPlan, the providers by the site's own providerOrder,
// their servers as listed.
func candidates(data *sourcesData, prefer string) []candidate {
	rank := make(map[string]int, len(data.ProviderOrder))
	for i, p := range data.ProviderOrder {
		rank[p] = i
	}
	seen := make(map[string]struct{})
	var out []candidate
	for _, step := range trackPlan(prefer) {
		for _, tr := range data.Tracks {
			if tr.Track != step.track {
				continue
			}
			for _, p := range orderedProviders(tr.Providers, rank) {
				var subs []sourceSubtitle
				if tr.Track == trackSoftSub {
					subs = wantedSubtitles(p.Subtitles)
				}
				if step.needSubs && len(subs) == 0 {
					continue
				}
				for _, s := range p.Servers {
					for _, st := range s.Streams {
						u := strings.TrimSpace(st.URL)
						format := strings.ToLower(strings.TrimSpace(st.Format))
						if !strings.HasPrefix(u, "https://") && !strings.HasPrefix(u, "http://") {
							continue
						}
						if format != "hls" && format != "mp4" {
							continue
						}
						if _, dup := seen[u]; dup {
							continue
						}
						seen[u] = struct{}{}
						out = append(out, candidate{
							track:     tr.Track,
							provider:  p.Provider,
							server:    s.Server,
							url:       u,
							format:    format,
							referer:   header(s.Headers, "Referer"),
							origin:    header(s.Headers, "Origin"),
							subtitles: subs,
						})
					}
				}
			}
		}
	}
	return out
}

// orderedProviders sorts a track's providers by the site's providerOrder;
// one the order does not name keeps its place after the named ones.
func orderedProviders(providers []sourceProvider, rank map[string]int) []sourceProvider {
	out := append([]sourceProvider(nil), providers...)
	pos := func(p sourceProvider) int {
		if r, ok := rank[p.Provider]; ok {
			return r
		}
		return len(rank)
	}
	// Stable insertion sort: the lists are a handful long.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && pos(out[j]) < pos(out[j-1]); j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

func header(h map[string]string, name string) string {
	for k, v := range h {
		if strings.EqualFold(k, name) {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// probe fetches a candidate the way the player will, with its headers, and
// returns the playlist for an HLS stream. A failure is logged and reported as
// false: the next candidate gets its turn.
func (c *Client) probe(ctx context.Context, cand candidate) ([]byte, bool) {
	pctx, cancel := context.WithTimeout(ctx, c.probeTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(pctx, http.MethodGet, cand.url, http.NoBody)
	if err != nil {
		return nil, false
	}
	req.Header.Set("User-Agent", c.userAgent)
	if cand.referer != "" {
		req.Header.Set("Referer", cand.referer)
	}
	if cand.origin != "" {
		req.Header.Set("Origin", cand.origin)
	}
	if cand.format == "mp4" {
		req.Header.Set("Range", "bytes=0-0")
	}
	resp, err := c.client.Do(req) // #nosec G704 -- URL comes from the site's own source list
	if err != nil {
		util.Debug("Miruro stream unreachable", "provider", cand.provider, "server", cand.server, "error", err)
		return nil, false
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		util.Debug("Miruro stream refused", "provider", cand.provider, "server", cand.server, "status", resp.StatusCode)
		return nil, false
	}
	if cand.format == "mp4" {
		return nil, true
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxPlaylistBytes))
	if err != nil || !bytes.HasPrefix(bytes.TrimSpace(body), []byte("#EXTM3U")) {
		util.Debug("Miruro stream is not a playlist", "provider", cand.provider, "server", cand.server)
		return nil, false
	}
	return body, true
}

// selectVariant returns the variant of a master playlist whose height matches
// the requested quality. Reports false for "best"/"", for a media playlist and
// whenever the height is absent, so the caller keeps the URL it has.
func selectVariant(playlist []byte, playlistURL, quality string) (string, bool) {
	want := normalizeQuality(quality)
	if want == 0 {
		return "", false
	}
	lines := strings.Split(string(playlist), "\n")
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
		return resolveRef(playlistURL, next), true
	}
	util.Debug("Miruro quality not available; using the playlist as is", "requested", quality)
	return "", false
}

// subtitleLanguage names the languages offered to the player: English, the
// site's own, and Portuguese, GoAnime's. Anything else returns "".
//
// The site's "default" flag is not trusted for this: on the episodes sampled
// one provider flagged Arabic as its default.
func subtitleLanguage(s sourceSubtitle) string {
	lang := strings.ToLower(strings.TrimSpace(s.Language))
	label := strings.ToLower(s.Label)
	switch {
	case lang == "en" || lang == "eng" || strings.HasPrefix(lang, "en-") || strings.Contains(label, "english"):
		return "English"
	case lang == "pt" || lang == "por" || strings.HasPrefix(lang, "pt-") || strings.Contains(label, "portugu"):
		return "Portuguese"
	default:
		return ""
	}
}

// wantedSubtitles keeps the English and Portuguese files of a provider,
// English first and, within a language, the site's default first. Every
// language a provider lists would put a picker of twenty entries in front of
// every episode.
func wantedSubtitles(subs []sourceSubtitle) []sourceSubtitle {
	var out []sourceSubtitle
	for _, lang := range []string{"English", "Portuguese"} {
		for _, wantDefault := range []bool{true, false} {
			for _, s := range subs {
				if s.Default != wantDefault || subtitleLanguage(s) != lang {
					continue
				}
				f := strings.TrimSpace(s.File)
				if !strings.HasPrefix(f, "https://") && !strings.HasPrefix(f, "http://") {
					continue
				}
				out = append(out, s)
			}
		}
	}
	return out
}

// encodeSubtitles renders the subtitle files into the metadata map, in the
// shape source_providers.go hands to util.SetGlobalSubtitles. The language is
// the plain name, so a choice the player remembers carries over between
// episodes whose labels differ; the label is the site's.
func encodeSubtitles(subs []sourceSubtitle) string {
	type track struct {
		URL      string `json:"url"`
		Language string `json:"language"`
		Label    string `json:"label"`
	}
	out := make([]track, 0, len(subs))
	for _, s := range subs {
		lang := subtitleLanguage(s)
		label := strings.TrimSpace(s.Label)
		if label == "" {
			label = lang
		}
		out = append(out, track{URL: strings.TrimSpace(s.File), Language: lang, Label: label})
	}
	if len(out) == 0 {
		return ""
	}
	encoded, err := jsonx.Marshal(out)
	if err != nil {
		util.Debug("Miruro could not encode subtitles", "error", err)
		return ""
	}
	return string(encoded)
}
