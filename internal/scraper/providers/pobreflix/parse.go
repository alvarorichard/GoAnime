package pobreflix

import (
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Title page anatomy (server-rendered, checked 2026-10-16):
//
//	<div id="view" data-video-id="84752" data-tmdb-id="793002" data-temp="" data-ep="">
//
// A movie page mounts one #view; an episode page mounts a #view.eps with
// data-temp and data-ep filled; a series page adds the vbEpisodes block â€” a
// season select (options data-season / data-for-audio) and episode cards
// (li.vbEpisodeCard, data-title / data-number, linking the episode records).
// The player itself is assembled client-side by the site's script bundle, so
// playback rides the StartFlix panel through the TMDB id â€” the api layer's
// job, never this parser's.

// parseTitle reads a title page. Two mount shapes exist (checked against the
// captured pages, 2026-10-16):
//
//   - movies and episodes mount <div id="view" data-video-id= data-tmdb-id=
//     [data-temp= data-ep=]>; the episode page adds class="eps";
//   - a series page carries the ids on the episodes block itself
//     (<div class="vbEpisodes" data-video-id= data-tmdb-id=>).
//
// ok=false covers a page with no mount or one without a TMDB id, which names
// nothing to play the title by.
func parseTitle(doc *goquery.Document, pageURL string) (Title, bool) {
	mount := doc.Find(`div#view`).First()
	if mount.Length() == 0 {
		mount = doc.Find(`.vbEpisodes`).First()
	}
	if mount.Length() == 0 {
		return Title{}, false
	}

	t := Title{}
	t.VideoID, _ = strconv.Atoi(strings.TrimSpace(attrOr(mount, "data-video-id")))
	t.TMDBID, _ = strconv.Atoi(strings.TrimSpace(attrOr(mount, "data-tmdb-id")))
	if t.TMDBID <= 0 {
		return Title{}, false
	}

	_, slug, audio, slugID := titleRecordFromPath(pathOfURL(pageURL))
	if slug == "" {
		return Title{}, false
	}
	if t.VideoID == 0 {
		t.VideoID = slugID
	}
	t.Audio = audio
	t.Kind = KindMovie
	if strings.Contains(pathOfURL(pageURL), "/series/") || strings.Contains(pathOfURL(pageURL), "/episodios/") {
		t.Kind = KindSeries
	}

	// The record's own name: the h2.one element is the clean Portuguese
	// title; the h1 reads "Assistir â€¦ Online".
	name := strings.Join(strings.Fields(doc.Find("h2.one").First().Text()), " ")
	if name == "" {
		name = displayTitle(doc.Find("h1.type").First().Text())
	}
	if name == "" {
		name = slugToTitle(slug)
	}
	t.Name = name

	// The poster is the record's TMDB image; og:image is the canonical form.
	t.Poster = strings.TrimSpace(doc.Find(`meta[property="og:image"]`).First().AttrOr("content", ""))

	if t.Kind == KindSeries {
		parseSeries(doc, &t)
	}
	return t, true
}

// parseSeries fills a series Title's seasons and episode names from the
// vbEpisodes block. Only the rendered audio/season combination is listed â€”
// other audios are other records â€” so the fill is best-effort by design:
// playback rides the panel and never depends on it.
func parseSeries(doc *goquery.Document, t *Title) {
	list := doc.Find(".vbEpisodes")
	if list.Length() == 0 {
		return
	}
	// Cards may not carry data-season; the block's data-current-season says
	// which season the rendered list belongs to.
	current, _ := strconv.Atoi(strings.TrimSpace(attrOr(list, "data-current-season")))
	seasons := map[int]*Season{}
	seasonFor := func(number int) *Season {
		if number <= 0 {
			number = current
		}
		if number <= 0 {
			number = 1
		}
		s := seasons[number]
		if s == nil {
			s = &Season{Number: number, Audio: t.Audio}
			seasons[number] = s
		}
		return s
	}
	list.Find("button.vbSeasonSelect__option").Each(func(_ int, opt *goquery.Selection) {
		season, _ := strconv.Atoi(strings.TrimSpace(attrOr(opt, "data-season")))
		if season <= 0 {
			return
		}
		s := seasonFor(season)
		if a := Audio(strings.TrimSpace(attrOr(opt, "data-for-audio"))); a != "" && s.Audio == "" {
			s.Audio = a
		}
	})
	list.Find("li.vbEpisodeCard").Each(func(_ int, card *goquery.Selection) {
		season, _ := strconv.Atoi(strings.TrimSpace(attrOr(card, "data-season")))
		number, _ := strconv.Atoi(strings.TrimSpace(attrOr(card, "data-number")))
		if number <= 0 {
			return
		}
		title := strings.Join(strings.Fields(attrOr(card, "data-title")), " ")
		href := attrOr(card.Find("a[href]").First(), "href")
		s := seasonFor(season)
		s.Episodes = append(s.Episodes, Episode{
			Number: number,
			Title:  title,
			URL:    urlResolve(pageURL0(doc), href),
		})
	})
	if len(seasons) == 0 {
		return
	}
	t.Seasons = t.Seasons[:0]
	for _, s := range seasons {
		sort.SliceStable(s.Episodes, func(i, j int) bool { return s.Episodes[i].Number < s.Episodes[j].Number })
		t.Seasons = append(t.Seasons, *s)
	}
	sort.Slice(t.Seasons, func(i, j int) bool { return t.Seasons[i].Number < t.Seasons[j].Number })
}

// titleRecordFromPath reads a title page's URL into its audio and record id,
// accepting the three record layouts — including /episodios/online/, whose
// slug carries an episode marker ("ben-hur-a-serie-1x1") that stays on the
// base slug; the corpus filters those pages out, a title page does not.
func titleRecordFromPath(path string) (kind MediaKind, baseSlug string, audio Audio, videoID int) {
	segment := ""
	for _, prefix := range []string{"/filmes/online/", "/series/online/", "/episodios/online/"} {
		if i := strings.Index(path, prefix); i >= 0 {
			segment = path[i+len(prefix):]
			if prefix == "/series/online/" || prefix == "/episodios/online/" {
				kind = KindSeries
			} else {
				kind = KindMovie
			}
			break
		}
	}
	if segment == "" {
		return "", "", AudioUnknown, 0
	}
	segment = strings.TrimSuffix(segment, "/")
	segment = strings.SplitN(segment, "?", 2)[0]
	idm := trailingIDRe.FindStringSubmatch(segment)
	if idm == nil {
		return "", "", AudioUnknown, 0
	}
	vid, err := strconv.Atoi(idm[2])
	if err != nil || vid <= 0 {
		return "", "", AudioUnknown, 0
	}
	segment = idm[1]
	audio = AudioUnknown
	if am := audioSuffixRe.FindStringSubmatch(segment); am != nil {
		segment = am[1]
		switch am[2] {
		case "legendado":
			audio = AudioSubtitled
		default:
			audio = AudioDubbed
		}
	}
	return kind, segment, audio, vid
}

// attrOr is goquery's Selection.AttrOr spelled shorter.
func attrOr(sel *goquery.Selection, attr string) string {
	v, ok := sel.Attr(attr)
	if !ok {
		return ""
	}
	return v
}

// pathOfURL returns the URL's path or the raw string when it does not parse â€”
// record URLs are absolute, but fixtures pass paths, and the classifier must
// not confuse them.
func pathOfURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if i := strings.Index(raw, "://"); i >= 0 {
		rest := raw[i+3:]
		if j := strings.Index(rest, "/"); j >= 0 {
			return rest[j:]
		}
		return "/"
	}
	if strings.HasPrefix(raw, "/") {
		return raw
	}
	if i := strings.Index(raw, "/"); i >= 0 {
		return raw[i:]
	}
	return "/" + raw
}

// hostOfURL returns the scheme+host of an absolute URL, "" otherwise.
func hostOfURL(raw string) string {
	raw = strings.TrimSpace(raw)
	i := strings.Index(raw, "://")
	if i < 0 {
		return ""
	}
	rest := raw[i+3:]
	if j := strings.Index(rest, "/"); j >= 0 {
		return raw[:i+3] + rest[:j]
	}
	return raw
}

// pageURL0 recovers an absolute base for href resolution from the document's
// canonical/base link; falls back to empty (test fixtures then carry
// absolute episode URLs by design).
func pageURL0(doc *goquery.Document) string {
	for _, sel := range []string{"link[rel=canonical]", "base"} {
		if v := strings.TrimSpace(doc.Find(sel).First().AttrOr("href", "")); v != "" && strings.Contains(v, "://") {
			return v
		}
	}
	return ""
}

// urlResolve resolves href against a page URL. Only the host matters here â€”
// record links are site-absolute ("/episodios/online/â€¦").
func urlResolve(pageURL, href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}
	if strings.Contains(href, "://") {
		return href
	}
	base := hostOfURL(pageURL)
	if base == "" {
		return href
	}
	if !strings.HasPrefix(href, "/") {
		href = "/" + href
	}
	return base + href
}
