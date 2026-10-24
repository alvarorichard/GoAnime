package pobreflix

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// The site moved from a server-rendered videobox page to a JSON-first one in
// October 2026. Every watch page now embeds the record the page is about in a
// <script id="initialMovieData">, <script id="initialSeriesData"> or
// <script id="initialEpisodeData"> block; the #view element that remains on
// movie pages is a fixed player shell and its data-tmdb-id is NOT the title's
// id. The embedded JSON is the authoritative source for the TMDB id, the
// display name, the poster, the servers and the episode list.

type embeddedItem struct {
	ID          int    `json:"id"`
	InternalID  int    `json:"internal_id"`
	TMDBID      int    `json:"tmdb_id"`
	Slug        string `json:"slug"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Poster      string `json:"poster"`
	PosterLarge string `json:"poster_large"`
	Audio       string `json:"audio"`
}

type embeddedServer struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Quality  string `json:"quality"`
	URL      string `json:"url"`
	Download string `json:"download"`
	Type     string `json:"type"`
	Audio    string `json:"audio"`
}

type embeddedMovie struct {
	Type    string                      `json:"type"`
	Item    embeddedItem                `json:"item"`
	Stream  string                      `json:"stream"`
	Servers map[string][]embeddedServer `json:"servers"`
}

type embeddedEpisode struct {
	ID        int    `json:"id"`
	TMDBID    int    `json:"tmdb_id"`
	Season    int    `json:"season"`
	Episode   int    `json:"episode"`
	Title     string `json:"title"`
	Available bool   `json:"available"`
	Audio     string `json:"audio"`
}

type embeddedSeason struct {
	Season int    `json:"season"`
	Name   string `json:"name"`
}

type embeddedSeries struct {
	Type     string            `json:"type"`
	Item     embeddedItem      `json:"item"`
	Episodes []embeddedEpisode `json:"episodes"`
	Seasons  []embeddedSeason  `json:"seasons"`
}

type embeddedEpisodePage struct {
	Type    string                      `json:"type"`
	Series  embeddedItem                `json:"series"`
	Episode embeddedEpisode             `json:"episode"`
	Servers map[string][]embeddedServer `json:"servers"`
}

// embeddedJSON reads one embedded JSON block by script id. An absent or
// unreadable block is a miss, never an error: the page may be an older theme.
func embeddedJSON(doc *goquery.Document, id string) []byte {
	if doc == nil {
		return nil
	}
	raw := strings.TrimSpace(doc.Find("script#" + id).First().Text())
	if raw == "" {
		return nil
	}
	return []byte(raw)
}

// embeddedTMDBID is the record's TMDB id. Current pages repeat it in both
// "tmdb_id" and "id"; older API rows sometimes carry only one.
func embeddedTMDBID(item embeddedItem) int {
	if item.TMDBID > 0 {
		return item.TMDBID
	}
	return item.ID
}

// embeddedMovieData returns the movie block, if the page carries one.
func embeddedMovieData(doc *goquery.Document) (embeddedMovie, bool) {
	raw := embeddedJSON(doc, "initialMovieData")
	if raw == nil {
		return embeddedMovie{}, false
	}
	var m embeddedMovie
	if json.Unmarshal(raw, &m) != nil || embeddedTMDBID(m.Item) <= 0 {
		return embeddedMovie{}, false
	}
	return m, true
}

// embeddedSeriesData returns the series block, if the page carries one.
func embeddedSeriesData(doc *goquery.Document) (embeddedSeries, bool) {
	raw := embeddedJSON(doc, "initialSeriesData")
	if raw == nil {
		return embeddedSeries{}, false
	}
	var s embeddedSeries
	if json.Unmarshal(raw, &s) != nil || embeddedTMDBID(s.Item) <= 0 {
		return embeddedSeries{}, false
	}
	return s, true
}

// embeddedEpisodeData returns the episode block, if the page carries one.
func embeddedEpisodeData(doc *goquery.Document) (embeddedEpisodePage, bool) {
	raw := embeddedJSON(doc, "initialEpisodeData")
	if raw == nil {
		return embeddedEpisodePage{}, false
	}
	var e embeddedEpisodePage
	if json.Unmarshal(raw, &e) != nil || embeddedTMDBID(e.Series) <= 0 {
		return embeddedEpisodePage{}, false
	}
	return e, true
}

// parseEmbeddedTitle reads the authoritative Title of a current-site page.
// The movie and series blocks describe the title itself; the episode block
// describes the series an episode belongs to (Episode reads the slot and
// servers).
func parseEmbeddedTitle(doc *goquery.Document, pageURL string) (Title, bool) {
	if m, ok := embeddedMovieData(doc); ok {
		t := titleFromEmbeddedItem(m.Item, KindMovie, pageURL)
		t.Servers = serversFromEmbedded(m.Servers)
		return t, true
	}
	if s, ok := embeddedSeriesData(doc); ok {
		t := titleFromEmbeddedItem(s.Item, KindSeries, pageURL)
		t.Seasons = seasonsFromEmbedded(doc, s.Item, pageURL)
		return t, true
	}
	if e, ok := embeddedEpisodeData(doc); ok {
		t := titleFromEmbeddedItem(e.Series, KindSeries, pageURL)
		t.Audio = audioFromEmbedded(e.Episode.Audio)
		if t.Audio == AudioUnknown {
			t.Audio = audioFromEmbedded(e.Series.Audio)
		}
		return t, true
	}
	return Title{}, false
}

// titleFromEmbeddedItem builds the common Title fields from an embedded item.
func titleFromEmbeddedItem(item embeddedItem, kind MediaKind, pageURL string) Title {
	tmdb := embeddedTMDBID(item)
	internal := item.InternalID
	if internal <= 0 {
		internal = item.ID
	}
	slug := strings.TrimSpace(item.Slug)
	if slug == "" {
		_, slug, _, _ = titleRecordFromPath(pathOfURL(pageURL))
	}
	poster := strings.TrimSpace(item.PosterLarge)
	if poster == "" {
		poster = strings.TrimSpace(item.Poster)
	}
	name := strings.Join(strings.Fields(item.Name), " ")
	return Title{
		Kind:       kind,
		VideoID:    internal,
		TMDBID:     tmdb,
		InternalID: internal,
		Audio:      audioFromEmbedded(item.Audio),
		Name:       name,
		Poster:     poster,
		Slug:       slug,
	}
}

// serversFromEmbedded flattens the site's per-audio server map. The map key is
// the authoritative audio; a server that repeats it is left alone.
func serversFromEmbedded(grouped map[string][]embeddedServer) []Server {
	if len(grouped) == 0 {
		return nil
	}
	audioForKey := func(key string) Audio {
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "dub", "dublado", "dual", "dual-dublado":
			return AudioDubbed
		case "leg", "legendado":
			return AudioSubtitled
		default:
			return AudioUnknown
		}
	}
	out := make([]Server, 0, len(grouped))
	seen := make(map[string]bool)
	for key, list := range grouped {
		audio := audioForKey(key)
		for _, s := range list {
			url := strings.TrimSpace(s.URL)
			download := strings.TrimSpace(s.Download)
			if url == "" && download == "" {
				continue
			}
			dedup := url + "\x00" + download
			if seen[dedup] {
				continue
			}
			seen[dedup] = true
			serverAudio := audioFromEmbedded(s.Audio)
			if serverAudio == AudioUnknown {
				serverAudio = audio
			}
			out = append(out, Server{
				URL:      url,
				Download: download,
				Quality:  strings.TrimSpace(s.Quality),
				Audio:    serverAudio,
				Name:     strings.TrimSpace(s.Name),
			})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Audio != out[j].Audio {
			return out[i].Audio == AudioDubbed
		}
		return QualityRank(out[i].Quality) > QualityRank(out[j].Quality)
	})
	return out
}

// audioFromEmbedded maps the site's short audio codes onto the package's
// labels.
func audioFromEmbedded(raw string) Audio {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "dub", "dublado", "dual", "dual-dublado":
		return AudioDubbed
	case "leg", "legendado", "sub", "subtitled":
		return AudioSubtitled
	default:
		return AudioUnknown
	}
}

// QualityRank turns a server's quality label into a sortable height. Unknown
// labels rank below every explicit resolution.
func QualityRank(raw string) int {
	raw = strings.ToLower(strings.TrimSpace(raw))
	switch raw {
	case "4k", "2160p", "uhd":
		return 2160
	case "1440p", "2k":
		return 1440
	case "1080p", "fullhd", "full hd", "fhd":
		return 1080
	case "720p", "hd":
		return 720
	case "480p", "sd":
		return 480
	case "360p":
		return 360
	}
	return 0
}

// seasonsFromEmbedded builds the season/episode list from the embedded series
// record. Episode URLs use the site's current
// /assistir/<slug>-t<season>-ep<episode>.html layout and are resolved against
// the page URL, so the api layer can open them directly.
func seasonsFromEmbedded(doc *goquery.Document, item embeddedItem, pageURL string) []Season {
	data, ok := embeddedSeriesData(doc)
	if !ok || len(data.Episodes) == 0 {
		return nil
	}
	byNumber := make(map[int]*Season)
	for _, ep := range data.Episodes {
		if !ep.Available || ep.Season <= 0 || ep.Episode <= 0 {
			continue
		}
		s := byNumber[ep.Season]
		if s == nil {
			s = &Season{Number: ep.Season}
			byNumber[ep.Season] = s
		}
		audio := audioFromEmbedded(ep.Audio)
		if audio == AudioUnknown {
			audio = audioFromEmbedded(item.Audio)
		}
		if s.Audio == AudioUnknown {
			s.Audio = audio
		}
		slug := strings.TrimSpace(item.Slug)
		path := ""
		if slug != "" {
			path = "/assistir/" + slug + "-t" + itoa(ep.Season) + "-ep" + itoa(ep.Episode) + ".html"
		}
		s.Episodes = append(s.Episodes, Episode{
			Number: ep.Episode,
			Title:  strings.Join(strings.Fields(ep.Title), " "),
			Audio:  audio,
			URL:    urlResolve(pageURL, path),
		})
	}
	if len(byNumber) == 0 {
		return nil
	}
	out := make([]Season, 0, len(byNumber))
	for _, s := range byNumber {
		sortEpisodes(s.Episodes)
		out = append(out, *s)
	}
	sortSeasons(out)
	return out
}

// sortEpisodes orders a season's episodes by number, keeping the dub record
// first when the same slot exists in both audios.
func sortEpisodes(episodes []Episode) {
	sort.SliceStable(episodes, func(i, j int) bool {
		if episodes[i].Number != episodes[j].Number {
			return episodes[i].Number < episodes[j].Number
		}
		return episodes[i].Audio == AudioDubbed && episodes[j].Audio != AudioDubbed
	})
}

// sortSeasons orders seasons by number.
func sortSeasons(seasons []Season) {
	sort.SliceStable(seasons, func(i, j int) bool { return seasons[i].Number < seasons[j].Number })
}

// itoa avoids importing strconv in two files for a single small use.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
