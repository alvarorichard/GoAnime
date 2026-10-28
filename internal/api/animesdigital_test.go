package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/animesdigital"
)

func TestGetAnimesDigitalEpisodesAudioChoice(t *testing.T) {
	oldClient, oldPick := animesDigitalClientFn, animesDigitalPickFn
	t.Cleanup(func() {
		animesDigitalClientFn, animesDigitalPickFn = oldClient, oldPick
	})

	var requested string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = r.URL.Path
		_, _ = w.Write([]byte(`<html><a href="/video/a/1/">Anime Episódio 1</a></html>`))
	}))
	defer srv.Close()
	client := animesdigital.NewClientForTest(srv.Client(), srv.URL)
	animesDigitalClientFn = func() *animesdigital.Client { return client }

	paired := &models.Anime{
		Source: "AnimesDigital",
		URL:    pairedAnimesDigitalURLForTest(srv.URL+"/anime/a/show", srv.URL+"/anime/a/show-dublado"),
	}

	t.Run("legendado is the default", func(t *testing.T) {
		animesDigitalPickFn = func([]string) (int, error) { return 0, errors.New("no TTY") }
		if _, err := GetAnimesDigitalEpisodes(paired); err != nil {
			t.Fatal(err)
		}
		if requested != "/anime/a/show" {
			t.Fatalf("requested %q, want subtitled page", requested)
		}
	})

	t.Run("dubbed can be selected", func(t *testing.T) {
		animesDigitalPickFn = func(labels []string) (int, error) {
			if len(labels) != 2 || labels[0] != "Legendado" {
				t.Fatalf("labels = %v", labels)
			}
			return 1, nil
		}
		if _, err := GetAnimesDigitalEpisodes(paired); err != nil {
			t.Fatal(err)
		}
		if requested != "/anime/a/show-dublado" {
			t.Fatalf("requested %q, want dubbed page", requested)
		}
	})
}

func TestGetAnimesDigitalEpisodesSkipsPickerForSingleAudio(t *testing.T) {
	oldClient, oldPick := animesDigitalClientFn, animesDigitalPickFn
	t.Cleanup(func() {
		animesDigitalClientFn, animesDigitalPickFn = oldClient, oldPick
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><a href="/video/a/1/">Anime Episódio 1</a></html>`))
	}))
	defer srv.Close()
	client := animesdigital.NewClientForTest(srv.Client(), srv.URL)
	animesDigitalClientFn = func() *animesdigital.Client { return client }
	animesDigitalPickFn = func([]string) (int, error) {
		t.Fatal("picker must not open for a single-audio title")
		return 0, nil
	}

	_, err := GetAnimesDigitalEpisodes(&models.Anime{Source: "AnimesDigital", URL: srv.URL + "/anime/a/show"})
	if err != nil {
		t.Fatal(err)
	}
}

// Build the same opaque fragment SearchAnime uses without exporting an
// implementation-only encoder.
func pairedAnimesDigitalURLForTest(subtitled, dubbed string) string {
	doc := `<html><a href="` + subtitled + `"><span class="title_anime">Show</span></a>` +
		`<a href="` + dubbed + `"><span class="title_anime">Show Dublado</span></a></html>`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(doc))
	}))
	defer srv.Close()
	client := animesdigital.NewClientForTest(srv.Client(), srv.URL)
	results, err := client.SearchAnime("show")
	if err != nil || len(results) != 1 {
		panic("could not build paired AnimesDigital test URL")
	}
	return results[0].URL
}
