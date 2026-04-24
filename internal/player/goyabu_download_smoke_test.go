package player

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
)

func TestSmokeGoyabuQuickDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live Goyabu download smoke test in short mode")
	}
	if os.Getenv("GOANIME_LIVE_SMOKE") == "" {
		t.Skip("set GOANIME_LIVE_SMOKE=1 to run the live Goyabu download smoke test")
	}

	StopBloggerProxy()
	t.Cleanup(StopBloggerProxy)

	anime := &models.Anime{
		Name:   "[PT-BR] Naruto Shippuden",
		URL:    "https://goyabu.io/anime/naruto-shippuden-online-hd-3",
		Source: "Goyabu",
	}

	episodes, err := api.GetAnimeEpisodesEnhanced(anime)
	if err != nil {
		if errors.Is(err, scraper.ErrSourceUnavailable) {
			t.Skipf("skipping live Goyabu download smoke test while upstream is unavailable: %v", err)
		}
		t.Fatalf("GetAnimeEpisodesEnhanced failed: %v", err)
	}
	if len(episodes) == 0 {
		t.Fatal("GetAnimeEpisodesEnhanced returned no episodes")
	}

	videoURL, err := GetVideoURLForEpisodeEnhanced(&episodes[0], anime)
	if err != nil {
		if errors.Is(err, scraper.ErrSourceUnavailable) {
			t.Skipf("skipping live Goyabu download smoke stream setup while upstream is unavailable: %v", err)
		}
		t.Fatalf("GetVideoURLForEpisodeEnhanced failed: %v", err)
	}
	if videoURL == "" {
		t.Fatal("GetVideoURLForEpisodeEnhanced returned an empty URL")
	}

	req, err := http.NewRequest(http.MethodGet, videoURL, nil)
	if err != nil {
		t.Fatalf("creating download request failed: %v", err)
	}
	req.Header.Set("Range", "bytes=0-4095")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("quick download request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		t.Fatalf("quick download returned status %d with body %q", resp.StatusCode, string(body))
	}

	outPath := filepath.Join(t.TempDir(), "goyabu_quick_download.bin")
	outFile, err := os.Create(outPath)
	if err != nil {
		t.Fatalf("creating temp output file failed: %v", err)
	}
	defer func() { _ = outFile.Close() }()

	written, err := io.Copy(outFile, io.LimitReader(resp.Body, 4096))
	if err != nil {
		t.Fatalf("writing quick download sample failed: %v", err)
	}
	if written == 0 {
		t.Fatal("quick download wrote 0 bytes")
	}

	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("stat temp output file failed: %v", err)
	}
	if info.Size() != written {
		t.Fatalf("temp output size = %d, want %d", info.Size(), written)
	}
}

func TestSmokeGoyabuDownloadResolver(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live Goyabu download resolver smoke test in short mode")
	}
	if os.Getenv("GOANIME_LIVE_SMOKE") == "" {
		t.Skip("set GOANIME_LIVE_SMOKE=1 to run the live Goyabu download resolver smoke test")
	}

	anime := &models.Anime{
		Name:   "[PT-BR] Naruto Shippuden",
		URL:    "https://goyabu.io/anime/naruto-shippuden-online-hd-3",
		Source: "Goyabu",
	}

	episodes, err := api.GetAnimeEpisodesEnhanced(anime)
	if err != nil {
		if errors.Is(err, scraper.ErrSourceUnavailable) {
			t.Skipf("skipping live Goyabu download resolver smoke test while upstream is unavailable: %v", err)
		}
		t.Fatalf("GetAnimeEpisodesEnhanced failed: %v", err)
	}
	if len(episodes) == 0 {
		t.Fatal("GetAnimeEpisodesEnhanced returned no episodes")
	}

	videoURL, err := ResolveDownloadStreamURL(episodes[0], anime)
	if err != nil {
		if errors.Is(err, scraper.ErrSourceUnavailable) {
			t.Skipf("skipping live Goyabu download resolver smoke stream setup while upstream is unavailable: %v", err)
		}
		t.Fatalf("ResolveDownloadStreamURL failed: %v", err)
	}
	if videoURL == "" {
		t.Fatal("ResolveDownloadStreamURL returned an empty URL")
	}
}

func TestSmokeGoyabuBatchDownloadResolverSample(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live Goyabu batch resolver smoke test in short mode")
	}
	if os.Getenv("GOANIME_LIVE_SMOKE") == "" {
		t.Skip("set GOANIME_LIVE_SMOKE=1 to run the live Goyabu batch resolver smoke test")
	}

	anime := &models.Anime{
		Name:   "[PT-BR] Naruto Shippuden",
		URL:    "https://goyabu.io/anime/naruto-shippuden-online-hd-3",
		Source: "Goyabu",
	}

	episodes, err := api.GetAnimeEpisodesEnhanced(anime)
	if err != nil {
		if errors.Is(err, scraper.ErrSourceUnavailable) {
			t.Skipf("skipping live Goyabu batch resolver smoke test while upstream is unavailable: %v", err)
		}
		t.Fatalf("GetAnimeEpisodesEnhanced failed: %v", err)
	}
	if len(episodes) < 3 {
		t.Fatalf("GetAnimeEpisodesEnhanced returned %d episodes, want at least 3", len(episodes))
	}

	for i := 0; i < 3; i++ {
		videoURL, err := ResolveDownloadStreamURL(episodes[i], anime)
		if err != nil {
			if errors.Is(err, scraper.ErrSourceUnavailable) {
				t.Skipf("skipping live Goyabu batch resolver smoke after upstream challenge on sample %d: %v", i+1, err)
			}
			t.Fatalf("ResolveDownloadStreamURL failed for sample %d: %v", i+1, err)
		}
		if videoURL == "" {
			t.Fatalf("ResolveDownloadStreamURL returned empty URL for sample %d", i+1)
		}
	}
}
