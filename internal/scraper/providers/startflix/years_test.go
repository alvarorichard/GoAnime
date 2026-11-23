package startflix

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// StartFlix lists iCarly (2007) and iCarly (2021) under one title and its
// search page gives neither a year (checked 2026-11-23: only the IMDb score),
// so the results screen folded them into one row. The year comes from the
// title page, for the results that need it. Note the slugs run the other way
// round from TopCine's: /series/icarly-2/ is the 2021 one here.

func titleDoc(t *testing.T, html string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestParseTitleYear(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, html, want string }{
		{"a series page", `<div class="extra"><span class="date">Jun. 17, 2021</span></div>`, "2021"},
		{"an old title", `<div class="extra"><span class="date">Sep. 08, 2007</span></div>`, "2007"},
		{"a full date", `<div class="extra"><span class="date" itemprop="dateCreated">2019-03-01</span></div>`, "2019"},
		{"no date line", `<div class="extra"><span>AMC+</span></div>`, ""},
		{"a date without a year", `<div class="extra"><span class="date">soon</span></div>`, ""},
		{"a date elsewhere on the page is not the title's", `<div class="sidebar"><span class="date">Jan. 01, 1999</span></div>`, ""},
		{"digits that are not a year", `<div class="extra"><span class="date">12345</span></div>`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := parseTitleYear(titleDoc(t, tc.html)); got != tc.want {
				t.Errorf("year = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAmbiguousYearless(t *testing.T) {
	t.Parallel()
	got := ambiguousYearless([]Media{
		{Title: "iCarly"},                      // 0: shares its title, no year
		{Title: " ICARLY ", Year: "2007"},      // 1: has a year already
		{Title: "Dexter"},                      // 2: unique
		{Title: "icarly"},                      // 3: shares its title, no year
		{Title: "Dexter: New Blood", Year: ""}, // 4: unique
	})
	if fmt.Sprint(got) != "[0 3]" {
		t.Errorf("ambiguousYearless = %v, want [0 3]: only a yearless result with a twin needs its page", got)
	}
}

// icarlyServer serves the iCarly search and the two title pages, counting the
// title pages asked for.
func icarlyServer(t *testing.T, titleHits *atomic.Int32) *Client {
	t.Helper()
	return newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			_, _ = w.Write(fixture(t, "search_icarly_2026_11_23.html"))
		case "/series/icarly-2/":
			titleHits.Add(1)
			_, _ = w.Write(fixture(t, "title_icarly_2021_2026_11_23.html"))
		case "/series/icarly/":
			titleHits.Add(1)
			_, _ = w.Write(fixture(t, "title_icarly_2007_2026_11_23.html"))
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestSearchFillsTheYearsOfSameTitleResults(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	c := icarlyServer(t, &hits)
	results, err := c.Search(context.Background(), "icarly")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	years := map[string]string{}
	for _, m := range results {
		if m.Title != "iCarly" || m.Kind != KindSeries {
			t.Errorf("result %+v lost its title or kind while its year was read", m)
		}
		years[m.URL[strings.LastIndex(strings.TrimSuffix(m.URL, "/"), "/")+1:]] = m.Year
	}
	if years["icarly-2/"] != "2021" || years["icarly/"] != "2007" {
		t.Errorf("years = %v, want icarly-2 -> 2021 and icarly -> 2007", years)
	}
}

func TestSearchYearsAreCachedPerPage(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	c := icarlyServer(t, &hits)
	for range 3 {
		if _, err := c.Search(context.Background(), "icarly"); err != nil {
			t.Fatal(err)
		}
	}
	if got := hits.Load(); got != 2 {
		t.Errorf("title pages fetched %d times over three searches, want 2 (once each)", got)
	}
}

func TestSearchLeavesUniqueTitlesAlone(t *testing.T) {
	t.Parallel()
	var other atomic.Int32
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			other.Add(1)
		}
		_, _ = w.Write(fixture(t, "search_2026_10_07.html"))
	}))
	if _, err := c.Search(context.Background(), "naruto"); err != nil {
		t.Fatal(err)
	}
	if other.Load() != 0 {
		t.Errorf("%d title pages fetched for results that share no title", other.Load())
	}
}

func TestSearchSurvivesATitlePageThatFails(t *testing.T) {
	t.Parallel()
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(fixture(t, "search_icarly_2026_11_23.html"))
	}))
	results, err := c.Search(context.Background(), "icarly")
	if err != nil {
		t.Fatalf("a year that cannot be read must not fail the search: %v", err)
	}
	if len(results) != 2 || results[0].Year != "" || results[1].Year != "" {
		t.Fatalf("results = %+v, want both kept with no year", results)
	}
	for _, m := range results {
		if m.Title != "iCarly" || m.URL == "" {
			t.Errorf("result %+v lost its title or link when its page failed", m)
		}
	}
}

func TestSearchCapsTheYearLookups(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	var page strings.Builder
	for i := range 10 {
		fmt.Fprintf(&page, `<div class="result-item"><article><div class="details"><div class="title"><a href="https://www.startflix.biz/series/one-%d/">One</a></div></div></article></div>`, i)
	}
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_, _ = w.Write([]byte("<html><body>" + page.String() + "</body></html>"))
			return
		}
		hits.Add(1)
		_, _ = w.Write(fixture(t, "title_icarly_2021_2026_11_23.html"))
	}))
	if _, err := c.Search(context.Background(), "one"); err != nil {
		t.Fatal(err)
	}
	if got := hits.Load(); got != maxYearLookups {
		t.Errorf("fetched %d title pages for ten twins, want the cap %d", got, maxYearLookups)
	}
}
