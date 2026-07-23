package superflix

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/util"
)

// ErrSuperFlixNoServers is returned when /player/bootstrap responds with an
// empty options list. This is a content-availability signal from SuperFlix
// (the upstream JS shows a "not yet released" screen in the same case), not
// a system or scraping error — callers should surface it to the user as

// NormalizeSuperFlixImageURL converts SuperFlix CloudFront proxy URLs to direct TMDB image URLs.
// Discord's image proxy cannot handle the double-URL format used by SuperFlix:
//
//	https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/poster.jpg
//
// This extracts the embedded TMDB URL and upgrades to w500 quality:
//
//	https://image.tmdb.org/t/p/w500/poster.jpg
func NormalizeSuperFlixImageURL(imageURL string) string {
	if imageURL == "" {
		return ""
	}
	const tmdbPrefix = "https://image.tmdb.org/t/p/"
	if idx := strings.Index(imageURL, tmdbPrefix); idx > 0 {
		direct := imageURL[idx:]
		// Upgrade thumbnail size for Discord display
		direct = strings.Replace(direct, "/w342/", "/w500/", 1)
		direct = strings.Replace(direct, "/w185/", "/w500/", 1)
		direct = strings.Replace(direct, "/w154/", "/w500/", 1)
		return direct
	}
	return imageURL
}

// SearchMedia searches SuperFlix for movies/series/animes
func (c *SuperFlixClient) SearchMedia(query string) ([]*SuperFlixMedia, error) {
	return c.SearchMediaWithContext(context.Background(), query)
}

// SearchMediaWithContext searches with context support.
//
// Search never escalates to the headed browser (WithoutBrowserSolve): the
// user is querying ALL sources at once here, and only the play path — where
// SuperFlix content was explicitly chosen — may open a browser window. If the
// gate is closed, search just returns no SuperFlix results (R6).
func (c *SuperFlixClient) SearchMediaWithContext(ctx context.Context, query string) ([]*SuperFlixMedia, error) {
	ctx = WithoutBrowserSolve(ctx)
	// CLI args arrive hyphenated like "the-boys" (TreatingAnimeName joins
	// words with dashes), but SuperFlix's search engine treats the dash as
	// a literal character and returns "Nenhum resultado encontrado".
	// Restore spaces so titles like "The Boys" actually match.
	normalized := strings.TrimSpace(query)
	normalized = strings.ReplaceAll(normalized, "-", " ")
	normalized = strings.ReplaceAll(normalized, "_", " ")
	for strings.Contains(normalized, "  ") {
		normalized = strings.ReplaceAll(normalized, "  ", " ")
	}

	cacheKey := strings.ToLower(normalized)
	if cached, ok := c.searchCache.Load(cacheKey); ok {
		return cached.([]*SuperFlixMedia), nil
	}

	searchURL := fmt.Sprintf("%s/pesquisar?s=%s", c.baseURL, url.QueryEscape(normalized))
	util.Debug("SuperFlix search", "query", query, "normalized", normalized, "url", searchURL)

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	results := c.parseCards(doc)
	c.searchCache.Store(cacheKey, results)
	return results, nil
}

// parseCards extracts media cards from SuperFlix HTML
func (c *SuperFlixClient) parseCards(doc *goquery.Document) []*SuperFlixMedia {
	var results []*SuperFlixMedia
	seen := make(map[string]bool)

	doc.Find("div.group\\/card").Each(func(i int, card *goquery.Selection) {
		var title, imageURL string
		if img := card.Find("img"); img.Length() > 0 {
			title, _ = img.Attr("alt")
			// Extract cover image URL from src, data-src, or srcset
			if src, ok := img.Attr("src"); ok && src != "" && !strings.HasPrefix(src, "data:") {
				imageURL = src
			}
			if imageURL == "" {
				if dataSrc, ok := img.Attr("data-src"); ok && dataSrc != "" {
					imageURL = dataSrc
				}
			}
			if imageURL == "" {
				if srcset, ok := img.Attr("srcset"); ok && srcset != "" {
					// Take the first URL from srcset (format: "url size, url size, ...")
					parts := strings.Fields(strings.Split(srcset, ",")[0])
					if len(parts) > 0 {
						imageURL = parts[0]
					}
				}
			}
		}
		if title == "" {
			if h3 := card.Find("h3"); h3.Length() > 0 {
				title = strings.TrimSpace(h3.Text())
			}
		}
		if title == "" {
			return
		}

		var tmdbID, imdbID, linkURL string

		card.Find("button").Each(func(_ int, btn *goquery.Selection) {
			msg, _ := btn.Attr("data-msg")
			copyVal, _ := btn.Attr("data-copy")
			switch {
			case strings.Contains(msg, "TMDB"):
				tmdbID = copyVal
			case strings.Contains(msg, "IMDB"):
				imdbID = copyVal
			case strings.Contains(msg, "Link"):
				linkURL = copyVal
			}
		})

		// Extract type and year from metadata.
		// The div.mt-3 contains child spans: one for the year (e.g. "2017")
		// and one for the type (e.g. "Anime", "Filme", "Série").
		// Older HTML used "|" separators; current HTML uses separate <span>s.
		var tipo, year string
		card.Find("div.mt-3 span").Each(func(_ int, span *goquery.Selection) {
			text := strings.TrimSpace(span.Text())
			if text == "" {
				return
			}
			// A 4-digit number starting with 1 or 2 is a year.
			if len(text) == 4 && (text[0] == '1' || text[0] == '2') {
				if _, err := strconv.Atoi(text); err == nil {
					year = text
					return
				}
			}
			tipo = text
		})
		// Fallback: try legacy "|"-separated format inside div.mt-3.
		if tipo == "" && year == "" {
			metaText := strings.TrimSpace(card.Find("div.mt-3").Text())
			metaParts := splitAndTrim(metaText, "|")
			if len(metaParts) > 0 {
				tipo = metaParts[len(metaParts)-1]
			}
			if len(metaParts) > 1 {
				year = metaParts[1]
			}
		}

		sfType := "serie"
		if strings.Contains(linkURL, "/filme/") {
			sfType = "filme"
		}

		key := tmdbID
		if key == "" {
			key = title
		}
		if seen[key] {
			return
		}
		seen[key] = true

		if tipo == "" {
			if sfType == "filme" {
				tipo = "Filme"
			} else {
				tipo = "Série"
			}
		}

		results = append(results, &SuperFlixMedia{
			Title:    title,
			Year:     year,
			Type:     tipo,
			SFType:   sfType,
			TMDBID:   tmdbID,
			IMDBID:   imdbID,
			ImageURL: NormalizeSuperFlixImageURL(imageURL),
		})
	})

	return results
}
