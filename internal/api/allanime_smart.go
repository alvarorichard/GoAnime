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
	// Validate
	if err := validateSmartRangeInputs(anime, startEp, endEp, &quality); err != nil {
		return err
	}

	util.Debug("AllAnime Smart Range start",
		"anime", anime.Name,
		"range", fmt.Sprintf("%d-%d", startEp, endEp),
		"quality", quality)

	// Fetch episodes using enhanced path (enables AniSkip enrichment)
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
	if mkErr := os.MkdirAll(outDir, 0700); mkErr != nil {
		return fmt.Errorf("failed to create output directory: %w", mkErr)
	}

	// Iterate episodes and download
	for i := startEp; i <= endEp; i++ {
		ep := episodes[i-1]
		filePath := filepath.Join(outDir, fmt.Sprintf("%d.mp4", i))

		if alreadyDownloaded(filePath) {
			util.Infof("Episode %d already exists, skipping", i)
			_ = WriteAniSkipSidecar(filePath, &ep) // refresh sidecar if needed
			continue
		}

		streamURL, err := resolveStreamURLForEpisode(&ep, anime, quality)
		if err != nil {
			util.Errorf("Failed to get stream URL for episode %d: %v", i, err)
			continue
		}

		if err := smartDownload(streamURL, filePath); err != nil {
			util.Errorf("Download failed for episode %d: %v", i, err)
			continue
		}

		if err := WriteAniSkipSidecar(filePath, &ep); err != nil {
			util.Debugf("Failed to write AniSkip sidecar for episode %d: %v", i, err)
		}
		util.Infof("Episode %d downloaded successfully", i)
	}
	return nil
}

// smartDownload chooses the best method to download AllAnime links (HLS/hosters)
func smartDownload(url, dest string) error {
	// Sanitize and validate destination path under the downloads root
	safeDest, err := sanitizeSmartDest(dest)
	if err != nil {
		return err
	}
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(safeDest), 0700); err != nil {
		return err
	}

	// Use yt-dlp for HLS/known hosters
	if shouldUseYtDlp(url) {
		ctx := context.Background()
		ytdlp.MustInstall(ctx, nil)
		dl := ytdlp.New().Output(safeDest)
		_, err := dl.Run(ctx, url)
		if err != nil {
			return fmt.Errorf("yt-dlp failed: %w", err)
		}
		// Verify
		if st, err := os.Stat(safeDest); err != nil || st.Size() < 1024 {
			return fmt.Errorf("download verification failed for %s", safeDest)
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
	// #nosec G304: path validated by sanitizeSmartDest to remain within the GoAnime downloads root
	out, err := os.Create(safeDest)
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
	// Restrictive permissions: owner read/write only
	return os.WriteFile(sidecar, b, 0600)
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

// sanitizeSmartDest ensures destination path is within the GoAnime downloads root under the user's home
func sanitizeSmartDest(p string) (string, error) {
	if strings.TrimSpace(p) == "" {
		return "", fmt.Errorf("empty destination path")
	}
	if strings.HasPrefix(p, "-") || strings.ContainsAny(p, "\x00\n\r") {
		return "", fmt.Errorf("invalid destination path")
	}
	cleaned := filepath.Clean(p)
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	root := filepath.Join(home, ".local", "goanime", "downloads", "anime")
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	absFile, err := filepath.Abs(cleaned)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(absRoot, absFile)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("destination escapes downloads root: %s", cleaned)
	}
	return absFile, nil
}

// Helpers to reduce complexity

// validateSmartRangeInputs ensures correct source and quality defaulting
func validateSmartRangeInputs(anime *models.Anime, startEp, endEp int, quality *string) error {
	if !isAllAnimeSourceAPI(anime) {
		return fmt.Errorf("AllAnime Smart Range is only available for AllAnime sources")
	}
	if quality != nil && *quality == "" {
		*quality = "best"
	}
	if startEp < 1 || endEp < startEp {
		return fmt.Errorf("invalid range %d-%d", startEp, endEp)
	}
	return nil
}

// shouldUseYtDlp decides if yt-dlp is preferred for a given URL
func shouldUseYtDlp(u string) bool {
	l := strings.ToLower(u)
	return strings.Contains(l, ".m3u8") ||
		strings.Contains(l, "master.m3u8") ||
		strings.Contains(l, "wixmp.com") || strings.Contains(l, "repackager.wixmp.com") ||
		strings.Contains(l, "blogger.com") ||
		strings.Contains(l, "sharepoint.com") ||
		strings.Contains(l, "allanime") || strings.Contains(l, "allmanga")
}

// alreadyDownloaded checks if the file exists and seems valid (>1KB)
func alreadyDownloaded(path string) bool {
	if st, err := os.Stat(path); err == nil {
		return st.Size() > 1024
	}
	return false
}

// resolveStreamURLForEpisode resolves the streaming URL with enhanced fallback
func resolveStreamURLForEpisode(ep *models.Episode, anime *models.Anime, quality string) (string, error) {
	if ep == nil || anime == nil {
		return "", fmt.Errorf("nil episode or anime")
	}
	url, err := GetEpisodeStreamURLEnhanced(ep, anime, quality)
	if err == nil && url != "" {
		util.Debug("Stream URL resolved (enhanced)", "len", len(url))
		return url, nil
	}
	url, err = GetEpisodeStreamURL(ep, anime, quality)
	if err != nil || url == "" {
		return "", fmt.Errorf("fallback stream URL resolution failed: %w", err)
	}
	util.Debug("Stream URL resolved (fallback)", "len", len(url))
	return url, nil
}
