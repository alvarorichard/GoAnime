package providers

import (
	"context"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
)

func resetMovieTVPlaybackState() {
	util.ResetPlaybackState()
	util.SetSubtitlesDisabled(false)
}

func TestFlixHQProviderFetchEpisodesMovie(t *testing.T) {
	t.Parallel()

	provider := &flixHQProvider{}
	anime := &models.Anime{
		Name:      "Dexter",
		Source:    "FlixHQ",
		MediaType: models.MediaTypeMovie,
		URL:       "https://flixhq.to/movie/watch-dexter-39448",
	}

	episodes, err := provider.FetchEpisodes(context.Background(), anime)
	if err != nil {
		t.Fatalf("FetchEpisodes returned error: %v", err)
	}
	if len(episodes) != 1 {
		t.Fatalf("FetchEpisodes returned %d episodes, want 1", len(episodes))
	}
	if episodes[0].URL != "39448" {
		t.Fatalf("movie episode URL = %q, want %q", episodes[0].URL, "39448")
	}
}

func TestSFlixProviderFetchEpisodesMovie(t *testing.T) {
	t.Parallel()

	provider := &sflixProvider{}
	anime := &models.Anime{
		Name:      "Inception",
		Source:    "SFlix",
		MediaType: models.MediaTypeMovie,
		URL:       "https://sflix.to/movie/free-inception-hd-27205",
	}

	episodes, err := provider.FetchEpisodes(context.Background(), anime)
	if err != nil {
		t.Fatalf("FetchEpisodes returned error: %v", err)
	}
	if len(episodes) != 1 {
		t.Fatalf("FetchEpisodes returned %d episodes, want 1", len(episodes))
	}
	if episodes[0].URL != "27205" {
		t.Fatalf("movie episode URL = %q, want %q", episodes[0].URL, "27205")
	}
}

func TestSuperFlixProviderFetchEpisodesMovie(t *testing.T) {
	t.Parallel()

	provider := &superFlixProvider{}
	anime := &models.Anime{
		Name:      "Dexter",
		Source:    "SuperFlix",
		MediaType: models.MediaTypeMovie,
		URL:       "1405",
	}

	episodes, err := provider.FetchEpisodes(context.Background(), anime)
	if err != nil {
		t.Fatalf("FetchEpisodes returned error: %v", err)
	}
	if len(episodes) != 1 {
		t.Fatalf("FetchEpisodes returned %d episodes, want 1", len(episodes))
	}
	if episodes[0].URL != "1405" {
		t.Fatalf("movie episode URL = %q, want %q", episodes[0].URL, "1405")
	}
}

func TestApplyFlixHQPlaybackStateStoresMetadata(t *testing.T) {
	t.Cleanup(resetMovieTVPlaybackState)

	applyFlixHQPlaybackState(&scraper.FlixHQStreamInfo{
		Referer: "https://flixhq.to/",
		Subtitles: []scraper.FlixHQSubtitle{
			{URL: "https://cdn.example.com/en.vtt", Language: "english", Label: "English"},
		},
	})

	if got := util.GetGlobalReferer(); got != "https://flixhq.to/" {
		t.Fatalf("GlobalReferer = %q, want %q", got, "https://flixhq.to/")
	}
	subtitles := util.GetGlobalSubtitles()
	if len(subtitles) != 1 {
		t.Fatalf("stored %d subtitle tracks, want 1", len(subtitles))
	}
	if subtitles[0].Label != "English" {
		t.Fatalf("subtitle label = %q, want %q", subtitles[0].Label, "English")
	}
}

func TestApplySFlixPlaybackStateStoresMetadata(t *testing.T) {
	t.Cleanup(resetMovieTVPlaybackState)

	applySFlixPlaybackState(&scraper.SFlixStreamInfo{
		Referer: "https://sflix.to/",
		Subtitles: []scraper.SFlixSubtitle{
			{URL: "https://cdn.example.com/en.vtt", Language: "english", Label: "English"},
		},
	})

	if got := util.GetGlobalReferer(); got != "https://sflix.to/" {
		t.Fatalf("GlobalReferer = %q, want %q", got, "https://sflix.to/")
	}
	subtitles := util.GetGlobalSubtitles()
	if len(subtitles) != 1 {
		t.Fatalf("stored %d subtitle tracks, want 1", len(subtitles))
	}
	if subtitles[0].Language != "english" {
		t.Fatalf("subtitle language = %q, want %q", subtitles[0].Language, "english")
	}
}

func TestApplySuperFlixPlaybackResultStoresMetadata(t *testing.T) {
	t.Cleanup(resetMovieTVPlaybackState)

	anime := &models.Anime{Name: "Dexter"}
	applySuperFlixPlaybackResult(anime, &scraper.SuperFlixStreamResult{
		Referer: "https://superflixapi.rest/",
		Thumb:   "https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/poster.jpg",
		Subtitles: []scraper.SuperFlixSubtitle{
			{Lang: "Portuguese", URL: "https://cdn.example.com/pt.vtt"},
		},
	})

	if got := util.GetGlobalReferer(); got != "https://superflixapi.rest/" {
		t.Fatalf("GlobalReferer = %q, want %q", got, "https://superflixapi.rest/")
	}
	if anime.ImageURL != "https://image.tmdb.org/t/p/w500/poster.jpg" {
		t.Fatalf("anime.ImageURL = %q, want normalized TMDB URL", anime.ImageURL)
	}
	subtitles := util.GetGlobalSubtitles()
	if len(subtitles) != 1 {
		t.Fatalf("stored %d subtitle tracks, want 1", len(subtitles))
	}
	if subtitles[0].Language != "portuguese" {
		t.Fatalf("subtitle language = %q, want %q", subtitles[0].Language, "portuguese")
	}
}

func TestApplyFlixHQPlaybackStateRespectsNoSubs(t *testing.T) {
	t.Cleanup(resetMovieTVPlaybackState)

	util.SetSubtitlesDisabled(true)

	applyFlixHQPlaybackState(&scraper.FlixHQStreamInfo{
		Referer: "https://flixhq.to/",
		Subtitles: []scraper.FlixHQSubtitle{
			{URL: "https://cdn.example.com/en.vtt", Language: "english", Label: "English"},
		},
	})

	if got := util.GetGlobalReferer(); got != "https://flixhq.to/" {
		t.Fatalf("GlobalReferer = %q, want %q", got, "https://flixhq.to/")
	}
	if subtitles := util.GetGlobalSubtitles(); len(subtitles) != 0 {
		t.Fatalf("stored %d subtitle tracks, want 0 with no-subs", len(subtitles))
	}
}
