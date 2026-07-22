package superflix

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// benchSearchHTML builds a search results page with n media cards in the
// current SuperFlix markup (group/card, data-msg buttons, mt-3 span metadata).
func benchSearchHTML(n int) string {
	var b strings.Builder
	b.WriteString(`<html><body><div id="results">`)
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, `
<div class="group/card">
  <img alt="Título de Teste %d" src="https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/poster%d.jpg">
  <h3>Título de Teste %d</h3>
  <button data-msg="Copiar TMDB" data-copy="%d"></button>
  <button data-msg="Copiar IMDB" data-copy="tt%07d"></button>
  <button data-msg="Copiar Link" data-copy="https://superflixapi.pro/serie/%d"></button>
  <div class="mt-3"><span>2021</span><span>Série</span></div>
</div>`, i, i, i, 100000+i, i, 100000+i)
	}
	b.WriteString(`</div></body></html>`)
	return b.String()
}

func BenchmarkParseCards(b *testing.B) {
	html := benchSearchHTML(48) // a full search results page
	c := &SuperFlixClient{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			b.Fatal(err)
		}
		if got := c.parseCards(doc); len(got) != 48 {
			b.Fatalf("expected 48 cards, got %d", len(got))
		}
	}
}

// benchEpisodesHTML builds a serie page carrying a window.allEpisodes blob with
// the given seasons × episodes.
func benchEpisodesHTML(seasons, episodes int) string {
	var b strings.Builder
	b.WriteString(`<html><head><script>window.allEpisodes = {`)
	for s := 1; s <= seasons; s++ {
		if s > 1 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, `"%d":[`, s)
		for e := 1; e <= episodes; e++ {
			if e > 1 {
				b.WriteString(",")
			}
			fmt.Fprintf(&b, `{"epi_num":%d,"title":"Episódio %d","air_date":"2021-04-%02d"}`, e, e, (e%28)+1)
		}
		b.WriteString("]")
	}
	b.WriteString(`};</script></head><body></body></html>`)
	return b.String()
}

func BenchmarkExtractEpisodes(b *testing.B) {
	html := benchEpisodesHTML(10, 24) // long-running dorama/anime scale
	c := &SuperFlixClient{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err := c.ExtractEpisodes(html)
		if err != nil {
			b.Fatal(err)
		}
		if len(out) != 10 {
			b.Fatalf("expected 10 seasons, got %d", len(out))
		}
	}
}
