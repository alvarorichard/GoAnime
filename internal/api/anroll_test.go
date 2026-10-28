package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/anroll"
)

func TestGetAnrollEpisodesAudioChoice(t *testing.T) {
	oldClient, oldPick := anrollClientFn, anrollPickFn
	t.Cleanup(func() { anrollClientFn, anrollPickFn = oldClient, oldPick })

	var requested string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = r.URL.Path
		_, _ = w.Write([]byte(`<html><a class="ep-text-item" data-ep="0001" href="/1/">Episode</a></html>`))
	}))
	defer srv.Close()
	client := anroll.NewClientForTest(srv.Client(), srv.URL)
	anrollClientFn = func() *anroll.Client { return client }
	paired := &models.Anime{Source: "AnimesROLL", URL: pairedAnrollURLForTest(t, srv.URL)}

	t.Run("legendado is the default", func(t *testing.T) {
		anrollPickFn = func([]string) (int, error) { return 0, errors.New("no TTY") }
		if _, err := GetAnrollEpisodes(paired); err != nil {
			t.Fatal(err)
		}
		if requested != "/anime/show" {
			t.Fatalf("requested %q, want subtitled page", requested)
		}
	})

	t.Run("dubbed can be selected", func(t *testing.T) {
		anrollPickFn = func(labels []string) (int, error) {
			if len(labels) != 2 || labels[0] != "Legendado" {
				t.Fatalf("labels = %v", labels)
			}
			return 1, nil
		}
		if _, err := GetAnrollEpisodes(paired); err != nil {
			t.Fatal(err)
		}
		if requested != "/anime/show-dublado" {
			t.Fatalf("requested %q, want dubbed page", requested)
		}
	})
}

func pairedAnrollURLForTest(t *testing.T, base string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<article class="anime-card"><a href="` + base + `/anime/show"><img alt="Show"><span class="card-sub">Legendado</span></a></article>` +
			`<article class="anime-card"><a href="` + base + `/anime/show-dublado"><img alt="Show Dublado"><span class="card-sub">Dublado</span></a></article>`))
	}))
	defer srv.Close()
	searchClient := anroll.NewClientForTest(srv.Client(), srv.URL)
	results, err := searchClient.SearchAnime("show")
	if err != nil || len(results) != 1 {
		t.Fatalf("paired search results=%d err=%v", len(results), err)
	}
	return results[0].URL
}
