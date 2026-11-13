package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/providers/naming"
	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/animesfhd"
)

var _ source.StreamFetcher = (*animesfhdProvider)(nil)
var _ source.Searchable = (*animesfhdProvider)(nil)
var _ scraper.ContextualScraper = (*scraper.AnimesFHDAdapter)(nil)

func TestAnimesFHDDescriptor20261111(t *testing.T) {
	t.Parallel()
	p := &animesfhdProvider{}
	d := p.Describe()
	if d.Kind != source.AnimesFHD || d.Priority != 70 || d.ProbeURL != "https://animesfhd.net" || d.DefaultDisabled || !p.HasSeasons() {
		t.Fatalf("descriptor=%+v", d)
	}
	adapter, err := p.scraper()
	if err != nil || adapter.GetType() != scraper.AnimesFHDType {
		t.Fatalf("adapter=%v err=%v", adapter, err)
	}
	if kind, ok := source.KindFor(scraper.AnimesFHDType); !ok || kind != source.AnimesFHD {
		t.Fatalf("kind=%s ok=%v", kind, ok)
	}
	if got := naming.CleanTitle("[AnimesFHD] Frieren"); got != "Frieren" {
		t.Fatalf("source tag leaked into title: %q", got)
	}
}

func TestAnimesFHDProviderHTTP(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/series":
			_, _ = w.Write([]byte(`{"series":[{"id":1,"title":"Frieren","category":"dublado"}]}`))
		case "/api/episodes/2":
			_, _ = w.Write([]byte(`{"episode":{"id":2},"sources":[{"native_video":true,"kind":"file","url":"/api/v/s6?source=3&playlist=.m3u8","audio":"dublado","subtitle_url":"/caption.vtt"}]}`))
		case "/api/v/s6":
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			_, _ = w.Write([]byte("#EXTM3U\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	p := &animesfhdProvider{}
	p.adapter.UnifiedScraper = scraper.NewAnimesFHDAdapterWithClient(animesfhd.NewClientForTest(srv.Client(), srv.URL))
	p.once.Do(func() {})
	results, err := p.Search(context.Background(), "frieren")
	if err != nil || len(results) != 1 || results[0].Name != "[PT-BR] Frieren" || results[0].Source != "AnimesFHD" {
		t.Fatalf("search=%v err=%v", results, err)
	}
	stream, err := p.FetchStream(context.Background(), &models.Episode{URL: srv.URL + "/episod/2"}, results[0], "")
	if err != nil || stream.Referer != "https://animesfhd.net/" || stream.AudioLang != "por" || len(stream.Subtitles) != 1 || stream.Subtitles[0].URL != srv.URL+"/caption.vtt" {
		t.Fatalf("stream=%+v err=%v", stream, err)
	}
}
