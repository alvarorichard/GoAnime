package animesfhd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/alvarorichard/Goanime/v2/internal/util/jsonx"
)

const (
	// SourceName is the canonical source name used in saved titles.
	SourceName = "AnimesFHD"
	// DefaultBase is the public site and API origin, checked on 2026-11-11.
	DefaultBase     = "https://animesfhd.net"
	maxBodySize     = 4 << 20
	maxEpisodePages = 100
)

// Client reads AnimesFHD without authentication or browser state.
type Client struct {
	http       *http.Client
	baseURL    string
	userAgent  string
	maxRetries int
	retryDelay time.Duration
}

// Version identifies one season/part/audio release, never a merged episode list.
type Version struct {
	URL    string
	Label  string
	Audio  string
	Season int
}

// Subtitle is a caption track for the selected source.
type Subtitle struct {
	URL      string `json:"url"`
	Language string `json:"language"`
	Label    string `json:"label"`
}

// Stream carries the native media and its playback context.
type Stream struct {
	URL       string
	Referer   string
	AudioLang string
	Subtitles []Subtitle
}

type seriesDTO struct {
	ID       int64    `json:"id"`
	Title    string   `json:"title"`
	Cover    string   `json:"cover_image"`
	Category string   `json:"category"`
	Year     int      `json:"year"`
	Audios   []string `json:"audios"`
}

type episodeDTO struct {
	ID          int64   `json:"id"`
	Number      float64 `json:"episode_number"`
	Title       string  `json:"title"`
	Aired       string  `json:"aired_at"`
	Category    string  `json:"series_category"`
	SubtitleURL string  `json:"subtitle_url"`
}

type versionDTO struct {
	ID     int64  `json:"series_id"`
	Title  string `json:"title"`
	Season int    `json:"season"`
	Part   int    `json:"part"`
	Label  string `json:"season_label"`
	Audio  string `json:"audio"`
}

type seriesDetail struct {
	Series   seriesDTO    `json:"series"`
	Episodes []episodeDTO `json:"episodes"`
	Versions []versionDTO `json:"versions"`
	HasMore  bool         `json:"has_more"`
}

type sourceDTO struct {
	URL         string `json:"url"`
	Kind        string `json:"kind"`
	Native      bool   `json:"native_video"`
	Audio       string `json:"audio"`
	SubtitleURL string `json:"subtitle_url"`
}

// NewAnimesFHDClient constructs a client without making network requests.
func NewAnimesFHDClient() *Client {
	c := NewClientForTest(util.NewFastClient(), DefaultBase)
	c.maxRetries = 2
	c.retryDelay = 350 * time.Millisecond
	return c
}

// NewClientForTest uses an injected HTTP client and origin with retries disabled.
func NewClientForTest(hc *http.Client, baseURL string) *Client {
	// Copy the client so installing a redirect policy never mutates a shared client.
	client := *hc
	c := &Client{http: &client, baseURL: strings.TrimRight(baseURL, "/"), userAgent: netx.UserAgent}
	previous := client.CheckRedirect
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("AnimesFHD: too many redirects")
		}
		if previous != nil {
			if err := previous(req, via); err != nil {
				return err
			}
		}
		c.decorateRequest(req)
		return nil
	}
	return c
}

// SearchAnime searches the first 50 matching titles, or popular titles if empty.
func (c *Client) SearchAnime(query string) ([]*models.Anime, error) {
	return c.SearchAnimeContext(context.Background(), query)
}

// SearchAnimeContext is the cancellable search used by the source fan-out.
func (c *Client) SearchAnimeContext(ctx context.Context, query string) ([]*models.Anime, error) {
	q := url.Values{"pagination": {"1"}, "page": {"1"}, "per_page": {"50"}}
	if query = strings.TrimSpace(query); query != "" {
		q.Set("q", query)
	} else {
		q.Set("sort", "popular")
	}
	var page struct {
		Series []seriesDTO `json:"series"`
	}
	if err := c.getJSON(ctx, c.baseURL+"/api/series?"+q.Encode(), &page); err != nil {
		return nil, err
	}
	if page.Series == nil {
		return nil, netx.NewParserError(SourceName, "search", "missing series array", nil)
	}
	results := make([]*models.Anime, 0, len(page.Series))
	seen := make(map[int64]bool)
	for _, s := range page.Series {
		if s.ID <= 0 || strings.TrimSpace(s.Title) == "" || seen[s.ID] {
			continue
		}
		seen[s.ID] = true
		mediaType := models.MediaTypeAnime
		if s.Category == "filme" {
			mediaType = models.MediaTypeMovie
		}
		dub, sub := audioName(s.Category) == "Dublado", audioName(s.Category) == "Legendado"
		for _, audio := range s.Audios {
			dub = dub || audioName(audio) == "Dublado"
			sub = sub || audioName(audio) == "Legendado"
		}
		year := ""
		if s.Year > 0 {
			year = strconv.Itoa(s.Year)
		}
		results = append(results, &models.Anime{Name: s.Title, URL: c.seriesURL(s.ID), ImageURL: s.Cover,
			Source: SourceName, MediaType: mediaType, Year: year, Audio: models.AudioLabel(dub, sub)})
	}
	return results, nil
}

// GetVersionsContext lists the exact season/audio IDs advertised for a title.
func (c *Client) GetVersionsContext(ctx context.Context, animeURL string) ([]Version, error) {
	id, err := c.pageID(animeURL, "serie")
	if err != nil {
		return nil, err
	}
	detail, err := c.detail(ctx, id, 1)
	if err != nil {
		return nil, err
	}
	versions := make([]Version, 0, len(detail.Versions))
	seen := make(map[int64]bool)
	for _, v := range detail.Versions {
		if v.ID <= 0 || seen[v.ID] {
			continue
		}
		seen[v.ID] = true
		label := v.Label
		if label == "" {
			label = v.Title
			if v.Season > 0 {
				label = fmt.Sprintf("Temporada %d", v.Season)
			}
			if v.Part > 0 {
				label += fmt.Sprintf(" - Parte %d", v.Part)
			}
		}
		audio := audioName(v.Audio)
		if audio != "" {
			label += " — " + audio
		}
		versions = append(versions, Version{URL: c.seriesURL(v.ID), Label: label, Audio: audio, Season: v.Season})
	}
	return versions, nil
}

// GetAnimeEpisodes lists only the selected version, in numeric episode order.
func (c *Client) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	return c.GetAnimeEpisodesContext(context.Background(), animeURL)
}

// GetAnimeEpisodesContext follows every episode page, rejecting stalled pagination.
func (c *Client) GetAnimeEpisodesContext(ctx context.Context, animeURL string) ([]models.Episode, error) {
	id, err := c.pageID(animeURL, "serie")
	if err != nil {
		return nil, err
	}
	var episodes []models.Episode
	seen := make(map[int64]bool)
	for page := 1; page <= maxEpisodePages; page++ {
		detail, err := c.detail(ctx, id, page)
		if err != nil {
			return nil, err
		}
		if detail.Episodes == nil {
			return nil, netx.NewParserError(SourceName, "episodes", "missing episodes array", nil)
		}
		before := len(episodes)
		for _, e := range detail.Episodes {
			if e.ID <= 0 || seen[e.ID] {
				continue
			}
			seen[e.ID] = true
			episodes = append(episodes, models.Episode{
				URL:    c.baseURL + "/episod/" + strconv.FormatInt(e.ID, 10),
				Number: strconv.FormatFloat(e.Number, 'f', -1, 64), Num: int(e.Number),
				Title: models.TitleDetails{Romaji: e.Title}, Aired: e.Aired,
				Audio: detail.audio(),
			})
		}
		if !detail.HasMore {
			sort.SliceStable(episodes, func(i, j int) bool {
				a, _ := strconv.ParseFloat(episodes[i].Number, 64)
				b, _ := strconv.ParseFloat(episodes[j].Number, 64)
				return a < b
			})
			return episodes, nil
		}
		if len(episodes) == before {
			return nil, netx.NewParserError(SourceName, "episodes", "pagination made no progress", nil)
		}
	}
	return nil, netx.NewParserError(SourceName, "episodes", "episode page limit reached", nil)
}

func (c *Client) detail(ctx context.Context, id int64, page int) (seriesDetail, error) {
	var detail seriesDetail
	target := fmt.Sprintf("%s/api/series/%d?page=%d&per_page=200", c.baseURL, id, page)
	if err := c.getJSON(ctx, target, &detail); err != nil {
		return detail, err
	}
	if detail.Series.ID != id {
		return detail, netx.NewParserError(SourceName, "series", "response has a different series ID", nil)
	}
	return detail, nil
}

func (d *seriesDetail) audio() string {
	if audio := audioName(d.Series.Category); audio != "" {
		return audio
	}
	// Movies use category=filme; only the matching version proves their audio.
	for _, v := range d.Versions {
		if v.ID == d.Series.ID {
			return audioName(v.Audio)
		}
	}
	return ""
}

// GetEpisodeStreamURL returns the native URL without its playback metadata.
func (c *Client) GetEpisodeStreamURL(episodeURL string) (string, error) {
	stream, err := c.GetEpisodeStreamContext(context.Background(), episodeURL)
	return stream.URL, err
}

// GetEpisodeStreamContext selects the first working native source in API order.
func (c *Client) GetEpisodeStreamContext(ctx context.Context, episodeURL string) (Stream, error) {
	id, err := c.pageID(episodeURL, "episod")
	if err != nil {
		return Stream{}, err
	}
	var detail struct {
		Episode episodeDTO  `json:"episode"`
		Sources []sourceDTO `json:"sources"`
	}
	if err := c.getJSON(ctx, fmt.Sprintf("%s/api/episodes/%d", c.baseURL, id), &detail); err != nil {
		return Stream{}, err
	}
	if detail.Episode.ID != id {
		return Stream{}, netx.NewParserError(SourceName, "episode", "response has a different episode ID", nil)
	}
	var lastErr error
	for _, candidate := range detail.Sources {
		if !candidate.Native || strings.EqualFold(candidate.Kind, "embed") || strings.TrimSpace(candidate.URL) == "" {
			continue
		}
		if want := audioName(detail.Episode.Category); want != "" && audioName(candidate.Audio) != "" && want != audioName(candidate.Audio) {
			continue
		}
		streamURL, err := c.resolveURL(candidate.URL)
		if err == nil {
			streamURL, err = c.resolveMedia(ctx, streamURL)
		}
		if err != nil {
			if ctx.Err() != nil {
				return Stream{}, ctx.Err()
			}
			lastErr = err
			continue
		}
		stream := Stream{URL: streamURL, Referer: DefaultBase + "/"}
		if audioName(candidate.Audio) == "Dublado" || audioName(detail.Episode.Category) == "Dublado" {
			stream.AudioLang = "por"
		}
		subtitle := candidate.SubtitleURL
		if subtitle == "" {
			subtitle = detail.Episode.SubtitleURL
		}
		if subtitle != "" {
			if target, err := c.resolveURL(subtitle); err == nil {
				stream.Subtitles = []Subtitle{{URL: target, Language: "por", Label: "Português"}}
			}
		}
		return stream, nil
	}
	if lastErr != nil {
		return Stream{}, fmt.Errorf("AnimesFHD native sources failed: %w", lastErr)
	}
	return Stream{}, netx.NewParserError(SourceName, "stream", "no native video source", nil)
}

func (c *Client) resolveMedia(ctx context.Context, target string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, http.NoBody)
	if err != nil {
		return "", err
	}
	c.decorateRequest(req)
	req.Header.Set("Range", "bytes=0-0")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("AnimesFHD media request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if err := netx.CheckHTTPStatus(resp, SourceName); err != nil {
		return "", err
	}
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/json") {
		return "", netx.NewParserError(SourceName, "stream", "resolver returned a page instead of media", nil)
	}
	finalURL := resp.Request.URL.String()
	if !isHLS(finalURL) && !strings.HasSuffix(strings.ToLower(resp.Request.URL.Path), ".mp4") && !strings.HasPrefix(contentType, "video/") && !strings.Contains(contentType, "mpegurl") {
		return "", netx.NewParserError(SourceName, "stream", "unrecognized media response", nil)
	}
	return netx.ValidateStreamURL(finalURL, SourceName)
}

func isHLS(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && (strings.HasSuffix(strings.ToLower(u.Path), ".m3u8") || u.Query().Get("playlist") == ".m3u8")
}

func (c *Client) getJSON(ctx context.Context, target string, dst any) error {
	for attempt := 0; ; attempt++ {
		retry, err := c.getJSONOnce(ctx, target, dst)
		if err == nil || !retry || !c.shouldRetry(attempt) {
			return err
		}
		if err := c.sleep(ctx); err != nil {
			return err
		}
	}
}

func (c *Client) getJSONOnce(ctx context.Context, target string, dst any) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, http.NoBody)
	if err != nil {
		return false, err
	}
	c.decorateRequest(req)
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return ctx.Err() == nil, fmt.Errorf("AnimesFHD API request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if err := netx.CheckHTTPStatus(resp, SourceName); err != nil {
		return resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500, err
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize+1))
	if err != nil {
		return false, fmt.Errorf("AnimesFHD response: %w", err)
	}
	if len(body) > maxBodySize {
		return false, netx.NewParserError(SourceName, "json", "response exceeds size limit", nil)
	}
	if err := netx.CheckHTMLResponse(resp, body, SourceName); err != nil {
		return false, err
	}
	if err := jsonx.Unmarshal(body, dst); err != nil {
		return false, netx.NewParserError(SourceName, "json", "invalid API JSON", err)
	}
	return false, nil
}

func (c *Client) decorateRequest(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Referer", DefaultBase+"/")
	req.Header.Set("Accept-Language", netx.AcceptLanguage)
}

func (c *Client) shouldRetry(attempt int) bool { return attempt < c.maxRetries }

func (c *Client) sleep(ctx context.Context) error {
	timer := time.NewTimer(c.retryDelay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (c *Client) resolveURL(ref string) (string, error) {
	base, err := url.Parse(c.baseURL + "/")
	if err != nil {
		return "", err
	}
	u, err := url.Parse(strings.TrimSpace(ref))
	if err != nil || u.User != nil {
		return "", netx.NewParserError(SourceName, "url", "invalid media URL", err)
	}
	return netx.ValidateStreamURL(base.ResolveReference(u).String(), SourceName)
}

func (c *Client) seriesURL(id int64) string {
	return c.baseURL + "/serie/" + strconv.FormatInt(id, 10)
}

func (c *Client) pageID(raw, kind string) (int64, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return 0, netx.NewParserError(SourceName, "url", "invalid page URL", err)
	}
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return 0, err
	}
	prefix := "/" + kind + "/"
	if u.Scheme != base.Scheme || u.Host != base.Host || u.User != nil || !strings.HasPrefix(u.Path, prefix) {
		return 0, netx.NewParserError(SourceName, "url", "unexpected page origin or path", nil)
	}
	rawID := strings.TrimSuffix(strings.TrimPrefix(u.Path, prefix), "/")
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 || strconv.FormatInt(id, 10) != rawID {
		return 0, netx.NewParserError(SourceName, "url", "invalid page ID", err)
	}
	return id, nil
}

func audioName(audio string) string {
	switch strings.ToLower(audio) {
	case "dublado":
		return "Dublado"
	case "legendado":
		return "Legendado"
	default:
		return ""
	}
}
