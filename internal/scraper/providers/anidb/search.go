package anidb

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

// SearchAnime searches anidb.app and returns the anime cards on the first page.
func (c *AniDBClient) SearchAnime(ctx context.Context, query string) ([]*models.Anime, error) {
	if strings.TrimSpace(query) == "" {
		return nil, netx.NewParserError(sourceLabel, "search", "empty query", nil)
	}
	searchURL := fmt.Sprintf("%s/browse?q=%s", c.baseURL, url.QueryEscape(query))
	util.Debug("AniDB search", "query", query, "url", searchURL)

	// The browse page is HTML by design, so netx.CheckHTMLResponse (which flags
	// HTML as a challenge on JSON endpoints) does not apply here;
	// CheckChallengeDocument below is the right guard for an HTML source.
	body, err := c.getBody(ctx, searchURL, "search")
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
		util.Debug("AniDB search returned no cards", "query", query)
	}
	return results, nil
}

// extractSearchResults reads the anime cards out of a browse page. Cards are
// anchors whose href is an /anime/<slug>-<id> permalink; the human title is the
// anchor's title attribute, with the poster's alt text as a fallback.
func (c *AniDBClient) extractSearchResults(doc *goquery.Document) []*models.Anime {
	seen := make(map[string]struct{})
	var out []*models.Anime

	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		m := animeHrefRe.FindStringSubmatch(strings.TrimSpace(href))
		if m == nil {
			return
		}
		id := m[2]
		if _, dup := seen[id]; dup {
			return
		}

		title := strings.TrimSpace(s.AttrOr("title", ""))
		if title == "" {
			title = strings.TrimSpace(s.Find("img").AttrOr("alt", ""))
		}
		if title == "" {
			return // a card with no title is navigation chrome, not a result
		}
		seen[id] = struct{}{}

		out = append(out, &models.Anime{
			Name:      title,
			URL:       c.animeURL(m[1], id),
			ImageURL:  strings.TrimSpace(s.Find("img").AttrOr("src", "")),
			Source:    sourceLabel,
			MediaType: models.MediaTypeAnime,
		})
	})
	return out
}
