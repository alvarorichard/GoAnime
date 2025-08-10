package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/lrstanley/go-ytdlp"
)

// DownloadAllAnimeSmartRange downloads a range of episodes exclusively for AllAnime.
// It prioritizes high-quality mirrors and writes AniSkip sidecar files for intro/outro skipping.
func DownloadAllAnimeSmartRange(anime *models.Anime, startEp, endEp int, quality string) error {
	// Validate source
	if !isAllAnimeSourceAPI(anime) {
		return fmt.Errorf("AllAnime Smart Range is only available for AllAnime sources")
	}

	if quality == "" {
		quality = "best"
	}

	util.Debug("AllAnime Smart Range start",
		"anime", anime.Name,
		"range", fmt.Sprintf("%d-%d", startEp, endEp),
		"quality", quality)

	// Fetch episodes using enhanced path (this enables AniSkip enrichment when MAL ID is available)
	episodes, err := GetAnimeEpisodesEnhanced(anime)
	if err != nil {
		return fmt.Errorf("failed to get episodes: %w", err)
	}
	if startEp < 1 || endEp > len(episodes) || startEp > endEp {
		return fmt.Errorf("invalid range %d-%d (available: 1-%d)", startEp, endEp, len(episodes))
	}

	// Prepare output directory
	outDir, err := smartOutputDir(anime)
	if err != nil {
		return err
	}
	if mkErr := os.MkdirAll(outDir, 0755); mkErr != nil {
		return fmt.Errorf("failed to create output directory: %w", mkErr)
	}

	// Iterate episodes and download
	for i := startEp; i <= endEp; i++ {
		ep := episodes[i-1]
		filePath := filepath.Join(outDir, fmt.Sprintf("%d.mp4", i))

		if stat, err := os.Stat(filePath); err == nil && stat.Size() > 1024 {
			util.Infof("Episode %d already exists, skipping", i)
			// Even if file exists, (re)write sidecar if needed
			_ = WriteAniSkipSidecar(filePath, &ep)
			continue
		}

		util.Infof("Resolving stream URL for episode %d...", i)
		streamURL, err := GetEpisodeStreamURLEnhanced(&ep, anime, quality)
		if err != nil || streamURL == "" {
			// Fallback to non-enhanced AllAnime-aware function
			streamURL, err = GetEpisodeStreamURL(&ep, anime, quality)
			if err != nil || streamURL == "" {
				util.Errorf("Failed to get stream URL for episode %d: %v", i, err)
				continue
			}
		}

		util.Debug("Stream URL resolved", "episode", i, "len", len(streamURL))

		// Prefer yt-dlp for HLS and known hosters used by AllAnime
		if err := smartDownload(streamURL, filePath); err != nil {
			util.Errorf("Download failed for episode %d: %v", i, err)
			continue
		}

		// Write AniSkip sidecar markers if available
		if err := WriteAniSkipSidecar(filePath, &ep); err != nil {
			util.Debugf("Failed to write AniSkip sidecar for episode %d: %v", i, err)
		}

		util.Infof("Episode %d downloaded successfully", i)
	}

	return nil
}

// smartDownload chooses the best method to download AllAnime links (HLS/hosters)
func smartDownload(url, dest string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	// Use yt-dlp for most AllAnime links (m3u8, wixmp, blogger, sharepoint, etc.)
	if strings.Contains(url, ".m3u8") ||
		strings.Contains(url, "master.m3u8") ||
		strings.Contains(url, "wixmp.com") || strings.Contains(url, "repackager.wixmp.com") ||
		strings.Contains(url, "blogger.com") ||
		strings.Contains(url, "sharepoint.com") ||
		strings.Contains(url, "allanime") || strings.Contains(url, "allmanga") {
		ctx := context.Background()
		ytdlp.MustInstall(ctx, nil)
		dl := ytdlp.New().Output(dest)
		_, err := dl.Run(ctx, url)
		if err != nil {
			return fmt.Errorf("yt-dlp failed: %w", err)
		}
		// Verify
		if st, err := os.Stat(dest); err != nil || st.Size() < 1024 {
			return fmt.Errorf("download verification failed for %s", dest)
		}
		return nil
	}

	// Otherwise, simple HTTP download
	client := &http.Client{Timeout: 0}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			util.Logger.Warn("Error closing response body", "error", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil {
			util.Logger.Warn("Error closing output file", "error", err)
		}
	}()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	return nil
}

// writeAniSkipSidecar writes a JSON file with OP/ED skip times next to the video
func writeAniSkipSidecar(videoPath string, ep *models.Episode) error {
	if ep == nil {
		return nil
	}
	// Only write if we have at least one skip window
	if ep.SkipTimes.Op.Start == 0 && ep.SkipTimes.Op.End == 0 && ep.SkipTimes.Ed.Start == 0 && ep.SkipTimes.Ed.End == 0 {
		return nil
	}

	type skipFile struct {
		Format  string `json:"format"`
		OPStart int    `json:"op_start"`
		OPEnd   int    `json:"op_end"`
		EDStart int    `json:"ed_start"`
		EDEnd   int    `json:"ed_end"`
		Updated string `json:"updated"`
		Episode string `json:"episode"`
		Source  string `json:"source"`
	}

	payload := skipFile{
		Format:  "aniskip",
		OPStart: ep.SkipTimes.Op.Start,
		OPEnd:   ep.SkipTimes.Op.End,
		EDStart: ep.SkipTimes.Ed.Start,
		EDEnd:   ep.SkipTimes.Ed.End,
		Updated: time.Now().Format(time.RFC3339),
		Episode: ep.Number,
		Source:  "AllAnime",
	}

	b, _ := json.MarshalIndent(payload, "", "  ")
	sidecar := strings.TrimSuffix(videoPath, filepath.Ext(videoPath)) + ".skips.json"
	return os.WriteFile(sidecar, b, 0644)
}

// WriteAniSkipSidecar is an exported wrapper to write AniSkip sidecar files.
// It delegates to the internal writeAniSkipSidecar and exists to allow other
// packages (e.g., player) to generate the sidecar after downloads.
func WriteAniSkipSidecar(videoPath string, ep *models.Episode) error {
	return writeAniSkipSidecar(videoPath, ep)
}

func smartOutputDir(anime *models.Anime) (string, error) {
	userHome, _ := os.UserHomeDir()
	safeName := sanitizeSmart(anime.Name)
	return filepath.Join(userHome, ".local", "goanime", "downloads", "anime", safeName), nil
}

func sanitizeSmart(name string) string {
	name = strings.ReplaceAll(name, "[AllAnime]", "")
	name = strings.ReplaceAll(name, "[AnimeFire]", "")
	name = strings.TrimSpace(name)
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, ch := range invalid {
		name = strings.ReplaceAll(name, ch, "_")
	}
	return name
}
