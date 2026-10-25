package animefire

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// AnimeFire is read through its JSON API, not by scraping pages.
//
// The site was rewritten as an Angular single-page app and moved from
// animefire.io to animefire.one. What the old scraper fetched is now a ~13KB
// shell with an empty <title> and no content at all — every search returned
// nothing, silently, because zero parsed results reads the same as zero
// matches. The source had been dead for a while and nothing said so.
//
// The app it serves talks to api.animefire.one, and that is what this file
// speaks. It is a better contract than the markup ever was: stable field names,
// no layout to drift, and the episode list arrives in one request instead of a
// paginated crawl.
//
// The three calls, mapped 2026-09-21 by reading the app's bundle and watching
// its network traffic:
//
//	GET /animes/pesquisar?q=<query>   search      → data[]  {id, titles, audio, poster}
//	GET /anime/<animeID>              episodes    → data.episodes[] {id, title, season, number}
//	GET /episode/<episodeID>          playback    → data.streams[]  {audio, url, qualities}
//
// A stream url looks like a picture (…/h.jpg) and is an HLS master playlist —
// the server sends application/vnd.apple.mpegurl. mpv and ffmpeg play it
// directly; the extension is camouflage, not a format.

// defaultAPIBase is the JSON API behind the site's app.
const defaultAPIBase = "https://api.animefire.one"

// apiBaseEnvOverride points the client at a different API origin.
//
// It is an escape hatch for a host that moves: this site has already
// moved once (animefire.io → animefire.one, and a scraped page → this API), and
// when it moves again an env var repairs an installed binary without waiting
// for a release. It also lets a test drive the whole client against an httptest
// server instead of the live API.
const apiBaseEnvOverride = "GOANIME_ANIMEFIRE_API"

// resolveAPIBase returns the API origin to talk to.
func resolveAPIBase() string {
	if v := strings.TrimSpace(os.Getenv(apiBaseEnvOverride)); v != "" {
		return strings.TrimSuffix(v, "/")
	}
	return defaultAPIBase
}

// apiVersion is the ?v= the app sends on every call. Kept because an API that
// bothers to version its requests may one day answer differently without it.
const apiVersion = "3"

// siteBase is the human-facing site, used to build the URLs carried in
// models.Anime.URL / models.Episode.URL.
//
// Those fields are the only channel this scraper has to the rest of GoAnime,
// which passes them back to it later (GetAnimeEpisodes, GetEpisodeStreamURL).
// They could hold a bare id, but a real page URL costs nothing and stays
// meaningful for anyone reading a log or a bookmark.
const siteBase = "https://animefire.one"

// apiSearchResponse is GET /animes/pesquisar.
type apiSearchResponse struct {
	Data []apiAnime `json:"data"`
}

type apiAnime struct {
	ID          string            `json:"id"`
	Titles      map[string]string `json:"titles"`
	Audio       string            `json:"audio"`
	PosterSrc   string            `json:"poster_src"`
	Status      string            `json:"status"`
	PublishedAt string            `json:"published_at"`
}

// title picks the Brazilian title, falling back through the other languages the
// API carries rather than returning an anime with no name.
func (a apiAnime) title() string {
	for _, k := range []string{"BR", "US", "JP"} {
		if t := strings.TrimSpace(a.Titles[k]); t != "" {
			return t
		}
	}
	for _, t := range a.Titles {
		if t = strings.TrimSpace(t); t != "" {
			return t
		}
	}
	return ""
}

// year is the year of the first release, from published_at ("2002-10-03").
func (a apiAnime) year() string {
	if y, _, _ := strings.Cut(strings.TrimSpace(a.PublishedAt), "-"); len(y) == 4 {
		if _, err := strconv.Atoi(y); err == nil {
			return y
		}
	}
	return ""
}

// audioLabel reads the API's audio field ("Dublado", "Legendado" or
// "Dublado & Legendado") into the wording every source uses.
func audioLabel(audio string) string {
	var dubbed, subtitled bool
	for part := range strings.SplitSeq(strings.ToLower(audio), "&") {
		switch strings.TrimSpace(part) {
		case "dublado":
			dubbed = true
		case "legendado":
			subtitled = true
		}
	}
	return models.AudioLabel(dubbed, subtitled)
}

// apiAnimeResponse is GET /anime/<id>.
type apiAnimeResponse struct {
	Data struct {
		Hero     apiAnime     `json:"hero"`
		Seasons  []apiSeason  `json:"seasons"`
		Episodes []apiEpisode `json:"episodes"`
	} `json:"data"`
}

// apiSeason is one entry of data.seasons. Episode numbers restart in every
// season; first_episode_number places the season in the whole run.
type apiSeason struct {
	Number             int `json:"number"`
	FirstEpisodeNumber int `json:"first_episode_number"`
}

type apiEpisode struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Audio    string `json:"audio"`
	Season   int    `json:"season"`
	Number   int    `json:"number"`
	Synopsis string `json:"synopsis"`
}

// apiEpisodeResponse is GET /episode/<id>.
type apiEpisodeResponse struct {
	Data struct {
		ID      string      `json:"id"`
		Title   string      `json:"title"`
		Number  int         `json:"number"`
		Streams []apiStream `json:"streams"`
	} `json:"data"`
}

type apiStream struct {
	Audio     string   `json:"audio"`
	IsOffline bool     `json:"is_offline"`
	URL       string   `json:"url"`
	Qualities []string `json:"qualities"`
}

// animeURLFor and episodeURLFor build the URLs carried in the models.
func animeURLFor(id string) string   { return siteBase + "/anime/" + id }
func episodeURLFor(id string) string { return siteBase + "/episode/" + id }

// idFromURL pulls the trailing API id out of one of those URLs.
//
// It also accepts a bare id, so a caller that already has one — or a stored
// entry from an older run — still works.
func idFromURL(raw string) string {
	raw = strings.TrimSpace(strings.TrimSuffix(raw, "/"))
	if raw == "" {
		return ""
	}
	if i := strings.LastIndex(raw, "/"); i >= 0 {
		return raw[i+1:]
	}
	return raw
}

// getJSON performs one API call and decodes it.
func (c *AnimefireClient) getJSON(path string, out any) error {
	endpoint := c.apiBase + path
	sep := "?"
	if strings.Contains(endpoint, "?") {
		sep = "&"
	}
	endpoint += sep + "v=" + apiVersion

	req, err := http.NewRequest(http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)
	req.Header.Set("Accept", "application/json")
	// The API is a different origin from the site the app runs on, so it
	// expects the browser's CORS headers.
	req.Header.Set("Origin", siteBase)
	req.Header.Set("Referer", siteBase+"/")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := netx.CheckHTTPStatus(resp, "AnimeFire API"); err != nil {
		return err
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	if err := json.Unmarshal(body, out); err != nil {
		// An HTML body here means we reached the site instead of the API — a
		// rotation, or a proxy in the way. Say that rather than surfacing a
		// bare JSON syntax error.
		return netx.NewParserError("AnimeFire", "api",
			"the API did not answer with JSON (the endpoint may have moved)", err)
	}
	return nil
}

// searchAPI runs a search through the JSON API.
func (c *AnimefireClient) searchAPI(query string) ([]*models.Anime, error) {
	var out apiSearchResponse
	if err := c.getJSON("/animes/pesquisar?q="+url.QueryEscape(query), &out); err != nil {
		return nil, err
	}

	animes := make([]*models.Anime, 0, len(out.Data))
	for _, a := range out.Data {
		name := a.title()
		if a.ID == "" || name == "" {
			continue
		}
		animes = append(animes, &models.Anime{
			Name:      name,
			URL:       animeURLFor(a.ID),
			ImageURL:  a.PosterSrc,
			MediaType: models.MediaTypeAnime,
			Year:      a.year(),
			Audio:     audioLabel(a.Audio),
		})
	}
	return animes, nil
}

// episodesAPI lists an anime's episodes.
func (c *AnimefireClient) episodesAPI(animeURL string) ([]models.Episode, error) {
	id := idFromURL(animeURL)
	if id == "" {
		return nil, fmt.Errorf("AnimeFire: no anime id in %q", animeURL)
	}

	var out apiAnimeResponse
	if err := c.getJSON("/anime/"+url.PathEscape(id), &out); err != nil {
		return nil, err
	}

	episodes := make([]models.Episode, 0, len(out.Data.Episodes))
	for _, e := range out.Data.Episodes {
		if e.ID == "" {
			continue
		}
		episodes = append(episodes, models.Episode{
			Num:      e.Number,
			URL:      episodeURLFor(e.ID),
			Title:    models.TitleDetails{Romaji: e.Title, English: e.Title},
			Synopsis: e.Synopsis,
			SeasonID: strconv.Itoa(e.Season),
		})
	}
	numberAcrossSeasons(episodes, out.Data.Seasons)
	c.warnIfTitleLooksOffline(episodes)
	return episodes, nil
}

// numberAcrossSeasons gives every episode its number in the whole run.
//
// The API numbers episodes within their season (One Piece's 1180 episodes
// repeat 983 numbers), and everything after the picker finds an episode by
// its number and takes the first match: next episode, resume, ranged
// downloads. Season 2 episode 1 played season 1 episode 1.
//
// Each season's first_episode_number gives the run number. Checked on
// 2026-10-12 against 24 titles: it numbered every list 1..N, in list order,
// with no repeat. If the numbers still repeat (a season the list does not
// describe), the list order, which the API gives by season and number, is
// used instead.
func numberAcrossSeasons(episodes []models.Episode, seasons []apiSeason) {
	first := make(map[string]int, len(seasons))
	for _, s := range seasons {
		if s.FirstEpisodeNumber > 0 {
			first[strconv.Itoa(s.Number)] = s.FirstEpisodeNumber
		}
	}
	seen := make(map[int]bool, len(episodes))
	unique := true
	for i := range episodes {
		if f, ok := first[episodes[i].SeasonID]; ok {
			episodes[i].Num = f + episodes[i].Num - 1
		}
		if seen[episodes[i].Num] {
			unique = false
		}
		seen[episodes[i].Num] = true
	}
	for i := range episodes {
		if !unique {
			episodes[i].Num = i + 1
		}
		episodes[i].Number = strconv.Itoa(episodes[i].Num)
	}
}

// warnIfTitleLooksOffline says so BEFORE the user starts picking episodes.
//
// AnimeFire lists titles it hosts nothing for. Naruto Shippuden is listed with
// 500 episodes and not one of them has a file — sampled across the range on
// 2026-09-21, every single probe came back is_offline with a null url. Nothing
// in the anime payload says so: status is "completed", the audio reads
// "Dublado & Legendado", the episode list looks complete.
//
// Without this the only way to find out is to pick an episode, be told it is
// offline, and repeat — which is exactly what happened: four episodes tried,
// four identical refusals, no way to tell it was the whole title.
//
// Since 2026-10-12 a probe also counts a URL whose playlist the CDN answers
// 404 for, which is most of the catalogue on that date (see checkPlaylist).
//
// Two probes, not one: a single gap is normal and must not condemn a title
// that mostly works. Both offline is a strong signal. A title whose first
// episode plays costs one probe; one with both offline cost about 0.75s on
// 2026-10-12 (an API call and a playlist check per track, per probe).
func (c *AnimefireClient) warnIfTitleLooksOffline(episodes []models.Episode) {
	if len(episodes) < 2 {
		return
	}
	for _, idx := range []int{0, len(episodes) / 2} {
		if _, err := c.streamAPI(episodes[idx].URL); !errors.Is(err, netx.ErrMediaOffline) {
			return // something is playable; nothing to warn about
		}
	}
	util.Warnf("AnimeFire may have no video files for this title — "+
		"the first and middle episodes are both offline (%d episodes listed).", len(episodes))
	util.Infof("Try the same title on another source (StartFlix or HiAnime).")
}

// streamAPI resolves an episode's playable URL.
func (c *AnimefireClient) streamAPI(episodeURL string) (string, error) {
	id := idFromURL(episodeURL)
	if id == "" {
		return "", fmt.Errorf("AnimeFire: no episode id in %q", episodeURL)
	}

	var out apiEpisodeResponse
	if err := c.getJSON("/episode/"+url.PathEscape(id), &out); err != nil {
		return "", err
	}

	// A URL whose check could not tell (network error, 5xx) is kept as a
	// last resort: it says nothing about the file, and hiding a file that is
	// there would be worse than letting the player try.
	var unchecked string
	for _, s := range orderStreams(out.Data.Streams) {
		err := c.checkPlaylist(s.URL)
		if err == nil {
			return s.URL, nil
		}
		util.Debug("AnimeFire track not playable", "episode", id, "audio", s.Audio, "err", err)
		if !errors.Is(err, errNoPlaylist) && unchecked == "" {
			unchecked = s.URL
		}
	}
	if unchecked != "" {
		return unchecked, nil
	}
	// The API said there is nothing to play (every track is_offline or with a
	// null url), or the CDN has no playlist behind the URLs it gave. That is
	// content state, not a scraping failure, and the player has to say so
	// rather than blaming itself.
	return "", fmt.Errorf("AnimeFire episode %s: %w", id, netx.ErrMediaOffline)
}

// orderStreams lists the tracks worth trying, in the order to try them.
//
// The API returns one entry per audio track, and GoAnime's audience is
// Brazilian, so a dubbed track comes first when both exist — the same
// preference the language tags elsewhere in the app encode. Offline entries
// are dropped: the API still lists them, but they have nothing behind them.
func orderStreams(streams []apiStream) []apiStream {
	var dubbed, others []apiStream
	for _, s := range streams {
		if s.IsOffline || strings.TrimSpace(s.URL) == "" {
			continue
		}
		if strings.EqualFold(s.Audio, "dublado") {
			dubbed = append(dubbed, s)
		} else {
			others = append(others, s)
		}
	}
	return append(dubbed, others...)
}

// errNoPlaylist means the CDN answered for the URL and there is no playlist
// there: 404/410, or a 200 that is not an HLS playlist.
var errNoPlaylist = errors.New("the video server has no playlist at this URL")

// checkPlaylist asks the CDN for the HLS master playlist a stream URL names.
//
// On 2026-10-12 akumast.net answered 404 for nearly every URL the API gave,
// fresh API answers included, while the API kept is_offline false. The
// master is a few hundred bytes, so asking for it is cheap. The CDN answers
// 403 to a request without a User-Agent, so the client's is sent.
func (c *AnimefireClient) checkPlaylist(rawURL string) error {
	req, err := http.NewRequest(http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("%w: %w", errNoPlaylist, err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", siteBase)
	req.Header.Set("Referer", siteBase+"/")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
		return fmt.Errorf("%w (HTTP %d)", errNoPlaylist, resp.StatusCode)
	case resp.StatusCode < 200 || resp.StatusCode > 299:
		return fmt.Errorf("playlist check: HTTP %d", resp.StatusCode)
	}
	head, err := io.ReadAll(io.LimitReader(resp.Body, 64))
	if err != nil {
		return err
	}
	if !strings.HasPrefix(strings.TrimSpace(string(head)), "#EXTM3U") {
		return fmt.Errorf("%w (the answer is not an HLS playlist)", errNoPlaylist)
	}
	return nil
}
