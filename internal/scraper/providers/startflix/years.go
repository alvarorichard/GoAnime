package startflix

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// The search page names a release year only for some titles. StartFlix lists
// a remake or revival under the same title as the original (iCarly 2007 and
// iCarly 2021), and a result with no year cannot be told from the other one.
// The title page always carries the date, so the year is read from there for
// the results that need it.

const (
	// yearLookupBudget bounds all the title-page fetches of one search.
	yearLookupBudget = 6 * time.Second
	// maxYearLookups caps the pages one search may fetch: a broad query
	// ("one") repeats titles too, and the year is only a label.
	maxYearLookups = 6
)

var releaseYearRe = regexp.MustCompile(`\b(?:19|20)\d{2}\b`)

// parseTitleYear reads the release year from a title page's date line
// ("Jun. 17, 2021"), or "" when the page has none.
func parseTitleYear(doc *goquery.Document) string {
	return releaseYearRe.FindString(doc.Find(".extra .date").First().Text())
}

// ambiguousYearless returns the indexes of results that have no year and share
// their title with another result, in result order.
func ambiguousYearless(results []Media) []int {
	count := map[string]int{}
	for _, m := range results {
		count[strings.ToLower(strings.TrimSpace(m.Title))]++
	}
	var idx []int
	for i, m := range results {
		if m.Year == "" && count[strings.ToLower(strings.TrimSpace(m.Title))] > 1 {
			idx = append(idx, i)
		}
	}
	return idx
}

// fillAmbiguousYears sets the year of each yearless result that shares its
// title with another, from its title page. Best effort: a page that cannot be
// read leaves the year empty.
func (c *Client) fillAmbiguousYears(ctx context.Context, results []Media) {
	idx := ambiguousYearless(results)
	if len(idx) == 0 {
		return
	}
	if len(idx) > maxYearLookups {
		idx = idx[:maxYearLookups]
	}
	ctx, cancel := context.WithTimeout(ctx, yearLookupBudget)
	defer cancel()
	var wg sync.WaitGroup
	for _, i := range idx {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i].Year = c.titleYear(ctx, results[i].URL)
		}()
	}
	wg.Wait()
}

// titleYear is a title page's release year. Cached per page.
func (c *Client) titleYear(ctx context.Context, pageURL string) string {
	if y, ok := c.years.Load(pageURL); ok {
		return y.(string)
	}
	doc, _, err := c.getDocument(ctx, pageURL, getOptions{referer: c.baseURL + "/", layer: "title"})
	if err != nil {
		util.Debug("StartFlix title year unavailable", "page", pageURL, "err", err)
		return ""
	}
	year := parseTitleYear(doc)
	c.years.Store(pageURL, year)
	return year
}
