//go:build ignore

// Command server stands in for the AnimeFire API while the README demo is
// recorded (see demo.tape). It is not part of GoAnime.
//
// The catalog lists only works in the public domain, and every title plays
// the same locally generated test pattern, which has no audio track. The
// recording therefore neither shows nor streams copyrighted material.
//
// It speaks the three calls the AnimeFire client makes (see
// internal/scraper/providers/animefire/api.go):
//
//	GET /animes/pesquisar?q=<query>
//	GET /anime/<id>
//	GET /episode/<id>
//
// and serves the stream itself under /media/. Like the real CDN's, the
// stream URL names an HLS playlist: the client asks for it and passes over a
// URL that does not answer with one.
//
// Usage:
//
//	go run docs/demo/server.go -addr 127.0.0.1:8765 -media <dir with index.m3u8>
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
)

type work struct {
	id     string
	title  string
	credit string // director
	year   string
}

// Japanese animated shorts published before 1931, which are in the public
// domain in Japan and in the United States. They are listed on AniList, so
// GoAnime's metadata lookup finds them as it would a real result.
var catalog = []work{
	{id: "namakura-gatana", title: "Namakura Gatana", credit: "Jun'ichi Kouchi", year: "1917"},
	{id: "saru-to-kani-no-gassen", title: "Saru to Kani no Gassen", credit: "Seitaro Kitayama", year: "1917"},
	{id: "momotarou", title: "Momotarou", credit: "Seitaro Kitayama", year: "1918"},
	{id: "mikanbune", title: "Mikanbune", credit: "Noburo Ofuji", year: "1927"},
	{id: "kuro-nyago", title: "Kuro Nyago", credit: "Noburo Ofuji", year: "1929"},
}

type anime struct {
	ID          string            `json:"id"`
	Titles      map[string]string `json:"titles"`
	Audio       string            `json:"audio"`
	Status      string            `json:"status"`
	PublishedAt string            `json:"published_at"`
}

type episode struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Audio  string `json:"audio"`
	Season int    `json:"season"`
	Number int    `json:"number"`
}

type stream struct {
	Audio     string   `json:"audio"`
	URL       string   `json:"url"`
	Qualities []string `json:"qualities"`
}

// anime describes a work the way the API does. The films are silent, with
// intertitles, so they are listed as subtitled ("Legendado"). Only the year of
// release is known for certain, and the year is all GoAnime reads from
// published_at.
func (w work) anime() anime {
	return anime{
		ID:          w.id,
		Titles:      map[string]string{"BR": w.title},
		Audio:       "Legendado",
		Status:      "completed",
		PublishedAt: w.year,
	}
}

func find(id string) (work, bool) {
	for _, w := range catalog {
		if w.id == id {
			return w, true
		}
	}
	return work{}, false
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{"data": v}); err != nil {
		log.Print(err)
	}
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8765", "address to listen on")
	media := flag.String("media", "hls", "directory holding the HLS playlist (index.m3u8) every episode plays")
	flag.Parse()

	mux := http.NewServeMux()

	// Titles that match the query come first and the rest of the catalog
	// follows, the way the real search also returns loosely related titles.
	mux.HandleFunc("GET /animes/pesquisar", func(w http.ResponseWriter, r *http.Request) {
		q := strings.ToLower(r.URL.Query().Get("q"))
		results := make([]work, len(catalog))
		copy(results, catalog)
		sort.SliceStable(results, func(i, j int) bool {
			return strings.Contains(strings.ToLower(results[i].title), q) &&
				!strings.Contains(strings.ToLower(results[j].title), q)
		})
		out := make([]anime, 0, len(results))
		for _, a := range results {
			out = append(out, a.anime())
		}
		writeJSON(w, out)
	})

	mux.HandleFunc("GET /anime/{id}", func(w http.ResponseWriter, r *http.Request) {
		a, ok := find(r.PathValue("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		// Momotarou lists three parts so the demo walks through the episodes
		// picker and the OPTIONS menu; every other work is a single short film
		// that plays straight away.
		count := 1
		if a.id == "momotarou" {
			count = 3
		}
		eps := make([]episode, 0, count)
		for n := 1; n <= count; n++ {
			eps = append(eps, episode{ID: fmt.Sprintf("%s.%d", a.id, n), Title: a.title, Audio: "Legendado", Season: 1, Number: n})
		}
		writeJSON(w, map[string]any{"hero": a.anime(), "episodes": eps})
	})

	mux.HandleFunc("GET /episode/{id}", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"id": r.PathValue("id"),
			"streams": []stream{{
				Audio:     "Legendado",
				URL:       "http://" + r.Host + "/media/index.m3u8",
				Qualities: []string{"1080p", "720p"},
			}},
		})
	})

	mux.Handle("GET /media/", http.StripPrefix("/media/", http.FileServer(http.Dir(*media))))

	for _, a := range catalog {
		log.Printf("%s (%s, %s), public domain", a.title, a.credit, a.year)
	}
	log.Printf("listening on http://%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
