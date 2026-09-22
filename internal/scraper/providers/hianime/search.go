package hianime

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

// SearchAnime searches hianime.at and returns the anime cards on the first page.
func (c *HiAnimeClient) SearchAnime(ctx context.Context, query string) ([]*models.Anime, error) {
	if strings.TrimSpace(query) == "" {
		return nil, netx.NewParserError(sourceLabel, "search", "empty query", nil)
	}
	searchURL := fmt.Sprintf("%s/search?keyword=%s", c.baseURL, url.QueryEscape(query))
	util.Debug("HiAnime search", "query", query, "url", searchURL)

	// The search page is HTML by design, so netx.CheckHTMLResponse (which flags
	// HTML as a challenge on JSON endpoints) does not apply here;
	// CheckChallengeDocument below is the right guard for an HTML source.
	body, err := c.getBody(ctx, searchURL, "search", nil)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, netx.NewParserError(sourceLabel, "search", "failed to parse search page", err)
	}
	if err := netx.CheckChallengeDocument(doc, sourceLabel+" search"); err != nil {
		return nil, err
	}

	results := c.extractSearchResults(doc)
	if len(results) == 0 {
		util.Debug("HiAnime search returned no cards", "query", query)
	}
	return results, nil
}

// extractSearchResults reads the anime cards out of a search page.
//
// Scoped to .flw-item rather than scanning every anchor: the page's chrome —
// the A-Z index, the footer, the sign-in links — carries hrefs that a bare
// "<slug>-<digits>" pattern happily matches (/az-list/0-9 among them). The
// card is the structure that actually means "a result".
func (c *HiAnimeClient) extractSearchResults(doc *goquery.Document) []*models.Anime {
	seen := make(map[string]struct{})
	var out []*models.Anime

	doc.Find(".flw-item").Each(func(_ int, card *goquery.Selection) {
		link := card.Find(".film-name a").First()
		href := strings.TrimSpace(link.AttrOr("href", ""))
		m := animeHrefRe.FindStringSubmatch(href)
		if m == nil {
			return
		}
		slug, id := m[1], m[2]
		if _, dup := seen[id]; dup {
			return
		}

		// The anchor's text is whitespace-padded by the template, so the title
		// attribute is the cleaner source; the poster's alt text is the fallback.
		title := strings.TrimSpace(link.AttrOr("title", ""))
		if title == "" {
			title = strings.TrimSpace(link.Text())
		}
		img := card.Find("img.film-poster-img").First()
		if title == "" {
			title = strings.TrimSpace(img.AttrOr("alt", ""))
		}
		if title == "" {
			return // a card with no title is chrome, not a result
		}
		seen[id] = struct{}{}

		poster := strings.TrimSpace(img.AttrOr("src", ""))
		if poster == "" {
			// Cards below the fold are lazy-loaded.
			poster = strings.TrimSpace(img.AttrOr("data-src", ""))
		}

		out = append(out, &models.Anime{
			Name:      title,
			URL:       c.animeURL(slug, id),
			ImageURL:  poster,
			Source:    sourceLabel,
			MediaType: models.MediaTypeAnime,
		})
	})
	return out
}
