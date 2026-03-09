// Package downloader provides download functionality for 9animetv.to anime content.
// This file implements a dedicated 9anime downloader that resolves streams via the
// NineAnimeClient scraper and downloads using native HLS or yt-dlp, with a Bubble Tea
// progress UI consistent with the rest of GoAnime.
package downloader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/downloader/hls"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/lrstanley/go-ytdlp"
	"github.com/manifoldco/promptui"
)

// NineAnimeDownloadConfig holds configuration for 9anime downloads
type NineAnimeDownloadConfig struct {
	OutputDir    string
	AudioType    string // "sub" or "dub"
	Quality      string // "best", "1080p", "720p", etc.
	AnimeName    string
	Season       int
	Concurrent   int    // max concurrent episode downloads (default 1 for 9anime rate limits)
	SubsLanguage string // subtitle language chosen by user; empty = prompt interactively
}

// NineAnimeDownloader handles downloading anime episodes from 9animetv.to
type NineAnimeDownloader struct {
	config          NineAnimeDownloadConfig
	client          *scraper.NineAnimeClient
	chosenSubLang   string // cached subtitle language choice for the session
	subLangResolved bool   // true once the user has been prompted (or "none" was chosen)
}

// NewNineAnimeDownloader creates a new 9anime downloader
func NewNineAnimeDownloader(config NineAnimeDownloadConfig) *NineAnimeDownloader {
	if config.OutputDir == "" {
		config.OutputDir = util.DefaultDownloadDir()
	}
	if config.AudioType == "" {
		config.AudioType = "sub"
	}
	if config.Quality == "" {
		config.Quality = "best"
	}
	if config.Season < 1 {
		config.Season = 1
	}
	if config.Concurrent < 1 {
		config.Concurrent = 1
	}

	return &NineAnimeDownloader{
		config: config,
		client: scraper.NewNineAnimeClient(),
	}
}

// DownloadAllEpisodes downloads every available episode from 9anime
func (d *NineAnimeDownloader) DownloadAllEpisodes(anime *models.Anime) error {
	if anime == nil {
		return fmt.Errorf("anime is nil")
	}

	animeID := anime.URL
	util.Infof("Downloading ALL episodes of %s from 9Anime (ID: %s)", anime.Name, animeID)

	episodes, err := d.client.GetEpisodes(animeID)
	if err != nil {
		return fmt.Errorf("failed to get episodes from 9Anime: %w", err)
	}
	if len(episodes) == 0 {
		return fmt.Errorf("no episodes found for %s", anime.Name)
	}

	fmt.Printf("Found %d episode(s) for %s\n", len(episodes), anime.Name)

	outputDir := d.buildOutputDir(anime)
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return d.downloadBatchWithProgress(anime, episodes, outputDir)
}

// DownloadSingleEpisode downloads a single episode by number from 9anime
func (d *NineAnimeDownloader) DownloadSingleEpisode(anime *models.Anime, episodeNum int) error {
	if anime == nil {
		return fmt.Errorf("anime is nil")
	}

	animeID := anime.URL // For 9anime, URL stores the anime ID
	util.Infof("Downloading episode %d of %s from 9Anime (ID: %s)", episodeNum, anime.Name, animeID)

	// Fetch episodes
	episodes, err := d.client.GetEpisodes(animeID)
	if err != nil {
		return fmt.Errorf("failed to get episodes from 9Anime: %w", err)
	}

	// Find the requested episode
	var targetEp *scraper.NineAnimeEpisode
	for i := range episodes {
		if episodes[i].Number == episodeNum {
			targetEp = &episodes[i]
			break
		}
	}
	if targetEp == nil {
		return fmt.Errorf("episode %d not found (available: %d episodes)", episodeNum, len(episodes))
	}

	// Resolve stream and download
	outputDir := d.buildOutputDir(anime)
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return d.downloadEpisodeWithProgress(anime, targetEp, outputDir)
}

// DownloadEpisodeRange downloads a range of episodes from 9anime
func (d *NineAnimeDownloader) DownloadEpisodeRange(anime *models.Anime, startEp, endEp int) error {
	if anime == nil {
		return fmt.Errorf("anime is nil")
	}
	if startEp > endEp {
		return fmt.Errorf("start episode (%d) cannot be greater than end episode (%d)", startEp, endEp)
	}

	animeID := anime.URL
	util.Infof("Downloading episodes %d-%d of %s from 9Anime (ID: %s)", startEp, endEp, anime.Name, animeID)

	// Fetch episodes
	episodes, err := d.client.GetEpisodes(animeID)
	if err != nil {
		return fmt.Errorf("failed to get episodes from 9Anime: %w", err)
	}

	// Filter episodes in range
	var toDownload []scraper.NineAnimeEpisode
	for _, ep := range episodes {
		if ep.Number >= startEp && ep.Number <= endEp {
			toDownload = append(toDownload, ep)
		}
	}

	if len(toDownload) == 0 {
		return fmt.Errorf("no episodes found in range %d-%d (total available: %d)", startEp, endEp, len(episodes))
	}

	outputDir := d.buildOutputDir(anime)
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return d.downloadBatchWithProgress(anime, toDownload, outputDir)
}

// buildOutputDir returns the Plex-compatible output directory for this anime
func (d *NineAnimeDownloader) buildOutputDir(anime *models.Anime) string {
	animeName := d.config.AnimeName
	if animeName == "" {
		animeName = anime.Name
	}
	// SanitizeForFilename now handles all bracket tags, parenthesized 9anime
	// metadata (HD, SUB, DUB, Multilanguage, Ep N/N), and trailing ratings.
	safeName := util.SanitizeForFilename(animeName)
	if safeName == "" {
		safeName = fmt.Sprintf("9Anime_%s", anime.URL)
	}
	return filepath.Join(d.config.OutputDir, safeName, fmt.Sprintf("Season %02d", d.config.Season))
}

// episodeFilename returns a Plex-compatible filename for an episode
func (d *NineAnimeDownloader) episodeFilename(anime *models.Anime, epNum int) string {
	animeName := d.config.AnimeName
	if animeName == "" {
		animeName = anime.Name
	}
	// SanitizeForFilename handles all cleaning (bracket tags, parens metadata, ratings).
	safeName := util.SanitizeForFilename(animeName)
	if safeName == "" {
		safeName = fmt.Sprintf("9Anime_%s", anime.URL)
	}
	return util.PlexEpisodeFilename(safeName, d.config.Season, epNum)
}

// resolveStream resolves the m3u8 stream URL and metadata for a 9anime episode
func (d *NineAnimeDownloader) resolveStream(episodeID string) (streamURL, referer string, subtitles []scraper.NineAnimeSubtitleTrack, err error) {
	// Get servers
	servers, err := d.client.GetServers(episodeID)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to get servers: %w", err)
	}
	if len(servers) == 0 {
		return "", "", nil, fmt.Errorf("no servers available")
	}

	// Filter by preferred audio type
	var filtered []scraper.NineAnimeServer
	for _, s := range servers {
		if s.AudioType == d.config.AudioType {
			filtered = append(filtered, s)
		}
	}
	if len(filtered) == 0 {
		filtered = servers // Fallback to any available
	}

	// Try each server until one works
	var lastErr error
	for _, server := range filtered {
		source, sErr := d.client.GetSource(server.DataID)
		if sErr != nil {
			lastErr = sErr
			util.Debugf("9anime server %s failed: %v", server.Name, sErr)
			continue
		}

		streamInfo, siErr := d.client.GetStreamInfo(source.EmbedURL)
		if siErr != nil {
			lastErr = siErr
			util.Debugf("9anime stream resolution failed for %s: %v", server.Name, siErr)
			continue
		}

		if streamInfo.M3U8URL == "" {
			lastErr = fmt.Errorf("empty m3u8 URL from %s", server.Name)
			continue
		}

		// Extract referer from embed domain
		ref := ""
		if parsed, pErr := url.Parse(source.EmbedURL); pErr == nil {
			ref = fmt.Sprintf("%s://%s/", parsed.Scheme, parsed.Host)
		}

		util.Debugf("9anime stream resolved via %s (%s): %s",
			server.Name, server.AudioType, streamInfo.M3U8URL[:min(len(streamInfo.M3U8URL), 80)])

		return streamInfo.M3U8URL, ref, streamInfo.Tracks, nil
	}

	if lastErr != nil {
		return "", "", nil, fmt.Errorf("all servers failed, last error: %w", lastErr)
	}
	return "", "", nil, fmt.Errorf("no working stream found")
}

// promptSubtitleLanguage shows the available subtitle languages and lets the user choose.
// The choice is cached so subsequent episodes in a batch use the same language.
// Returns the chosen label (or empty string to skip subtitles).
func (d *NineAnimeDownloader) promptSubtitleLanguage(tracks []scraper.NineAnimeSubtitleTrack) string {
	// If already resolved for this session, reuse
	if d.subLangResolved {
		return d.chosenSubLang
	}

	// If caller pre-configured a language, try to match it
	if d.config.SubsLanguage != "" {
		d.subLangResolved = true
		if strings.EqualFold(d.config.SubsLanguage, "none") || strings.EqualFold(d.config.SubsLanguage, "off") {
			d.chosenSubLang = ""
			return ""
		}
		if strings.EqualFold(d.config.SubsLanguage, "all") {
			d.chosenSubLang = "__all__"
			return "__all__"
		}
		// Try exact or partial match
		for _, t := range tracks {
			if strings.EqualFold(t.Label, d.config.SubsLanguage) ||
				strings.Contains(strings.ToLower(t.Label), strings.ToLower(d.config.SubsLanguage)) {
				d.chosenSubLang = t.Label
				return t.Label
			}
		}
		// No match found — fall through to interactive prompt
		util.Warnf("Subtitle language %q not found, showing available options...", d.config.SubsLanguage)
	}

	if len(tracks) == 0 {
		d.subLangResolved = true
		d.chosenSubLang = ""
		return ""
	}

	// Build options: each language + "All" + "None"
	var items []string
	for _, t := range tracks {
		label := t.Label
		if t.Default {
			label += " (default)"
		}
		items = append(items, label)
	}
	items = append(items, "All (download all subtitles)")
	items = append(items, "None (skip subtitles)")

	fmt.Printf("\n%d subtitle language(s) available:\n", len(tracks))

	prompt := promptui.Select{
		Label: "Select subtitle language",
		Items: items,
		Size:  len(items),
	}

	idx, _, err := prompt.Run()
	d.subLangResolved = true

	if err != nil {
		// On error (Ctrl+C, etc.) default to the track marked as default, or skip
		for _, t := range tracks {
			if t.Default {
				d.chosenSubLang = t.Label
				fmt.Printf("Defaulting to subtitle: %s\n", t.Label)
				return t.Label
			}
		}
		d.chosenSubLang = ""
		return ""
	}

	if idx == len(items)-1 {
		// "None" selected
		d.chosenSubLang = ""
		fmt.Println("Subtitles: disabled")
		return ""
	}
	if idx == len(items)-2 {
		// "All" selected
		d.chosenSubLang = "__all__"
		fmt.Println("Subtitles: downloading all languages")
		return "__all__"
	}

	d.chosenSubLang = tracks[idx].Label
	fmt.Printf("Subtitles: %s\n", d.chosenSubLang)
	return d.chosenSubLang
}

// downloadSubtitles downloads subtitle tracks alongside the video file,
// filtered by the user's language choice (already prompted before download started).
func (d *NineAnimeDownloader) downloadSubtitles(tracks []scraper.NineAnimeSubtitleTrack, videoPath string) {
	if len(tracks) == 0 {
		return
	}

	// Use the cached subtitle language choice
	choice := d.chosenSubLang
	if !d.subLangResolved || choice == "" {
		return // no subtitle language chosen or user chose "None"
	}

	// Check ffmpeg availability and validate the resolved binary path.
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		util.Warnf("ffmpeg not found — cannot embed subtitles into the video file")
		return
	}
	// Resolve symlinks and ensure the path is absolute to prevent PATH-based injection.
	ffmpegPath, err = filepath.EvalSymlinks(ffmpegPath)
	if err != nil {
		util.Warnf("failed to resolve ffmpeg path: %v", err)
		return
	}
	if !filepath.IsAbs(ffmpegPath) {
		util.Warnf("ffmpeg resolved to a non-absolute path — refusing to execute")
		return
	}
	if fi, statErr := os.Stat(ffmpegPath); statErr != nil || fi.IsDir() {
		util.Warnf("ffmpeg path is not a valid file: %s", ffmpegPath)
		return
	}

	dir := filepath.Dir(videoPath)

	type subEntry struct {
		tmpPath  string
		label    string
		langCode string
	}
	var entries []subEntry

	for _, track := range tracks {
		if track.File == "" {
			continue
		}

		// Filter by chosen language unless "all" was selected
		if choice != "__all__" && !strings.EqualFold(track.Label, choice) {
			continue
		}

		ext := "vtt"
		lower := strings.ToLower(track.File)
		if strings.Contains(lower, ".srt") {
			ext = "srt"
		} else if strings.Contains(lower, ".ass") {
			ext = "ass"
		}

		lang := util.SanitizeForFilename(track.Label)
		if lang == "" {
			lang = "unknown"
		}

		tmpPath := filepath.Join(dir, fmt.Sprintf(".tmp_sub_%s.%s", lang, ext))
		if dlErr := d.downloadFile(track.File, tmpPath); dlErr != nil {
			util.Warnf("Failed to download subtitle (%s): %v", track.Label, dlErr)
			continue
		}
		entries = append(entries, subEntry{tmpPath: tmpPath, label: track.Label, langCode: lang})
	}

	if len(entries) == 0 {
		return
	}

	// Mux subtitles into the video container with ffmpeg
	// --- Mux subtitles into the video container ---
	fmt.Printf("Embedding %d subtitle(s) into video...\n", len(entries))

	buildMuxArgs := func(subCodec, outPath string) []string {
		// Clean all file paths to prevent directory traversal in arguments.
		a := []string{"-y", "-fflags", "+genpts", "-i", filepath.Clean(videoPath)}
		for _, e := range entries {
			a = append(a, "-i", filepath.Clean(e.tmpPath))
		}
		// Map only video and audio from input — skip data streams like timed_id3
		// which are present in MPEG-TS from HLS downloads and crash MP4/MKV muxing.
		a = append(a, "-map", "0:v", "-map", "0:a")
		for i := range entries {
			a = append(a, "-map", fmt.Sprintf("%d", i+1))
		}
		a = append(a, "-c:v", "copy", "-c:a", "copy", "-c:s", subCodec)
		for i, e := range entries {
			a = append(a, fmt.Sprintf("-metadata:s:s:%d", i), fmt.Sprintf("language=%s", e.langCode))
			a = append(a, fmt.Sprintf("-metadata:s:s:%d", i), fmt.Sprintf("title=%s", e.label))
		}
		a = append(a, filepath.Clean(outPath))
		return a
	}

	runMux := func(subCodec, outPath string) error {
		args := buildMuxArgs(subCodec, outPath)
		util.Debugf("ffmpeg mux cmd: %s %v", "ffmpeg", args)
		cmd := exec.Command(ffmpegPath, args...) // #nosec G204 -- ffmpegPath is validated: resolved via EvalSymlinks, confirmed absolute and a regular file
		var stderrBuf bytes.Buffer
		cmd.Stdout = nil
		cmd.Stderr = &stderrBuf
		if err := cmd.Run(); err != nil {
			util.Debugf("ffmpeg mux failed: %v\nstderr: %s", err, stderrBuf.String())
			_ = os.Remove(outPath)
			return err
		}
		return nil
	}

	embedded := false

	// Attempt 1: MP4 container with mov_text subtitle codec
	tmpMP4 := videoPath + ".muxing.mp4"
	if err := runMux("mov_text", tmpMP4); err == nil {
		if renErr := os.Rename(tmpMP4, videoPath); renErr != nil {
			util.Warnf("Failed to replace video: %v", renErr)
			_ = os.Remove(tmpMP4)
		} else {
			embedded = true
		}
	} else {
		util.Debugf("MP4 mux failed: %v — trying MKV fallback", err)
	}

	// Attempt 2: MKV container (more tolerant of various subtitle formats / TS inputs)
	if !embedded {
		mkvPath := strings.TrimSuffix(videoPath, filepath.Ext(videoPath)) + ".mkv"
		tmpMKV := mkvPath + ".tmp.mkv"
		if err := runMux("srt", tmpMKV); err == nil {
			if renErr := os.Rename(tmpMKV, mkvPath); renErr != nil {
				util.Warnf("Failed to save MKV: %v", renErr)
				_ = os.Remove(tmpMKV)
			} else {
				embedded = true
				fmt.Printf("Note: saved as .mkv for better subtitle compatibility\n")
			}
		}
	}

	if embedded {
		fmt.Printf("Subtitles embedded successfully!\n")
	} else {
		util.Warnf("Could not embed subtitles — both MP4 and MKV muxing failed")
	}

	// Clean up temp subtitle files
	for _, e := range entries {
		_ = os.Remove(e.tmpPath)
	}
}

// downloadFile downloads a URL to a local file
func (d *NineAnimeDownloader) downloadFile(fileURL, destPath string) error {
	client := &http.Client{
		Transport: api.SafeTransport(30 * time.Second),
		Timeout:   60 * time.Second,
	}

	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", scraper.NineAnimeUserAgent)

	resp, err := client.Do(req) // #nosec G704
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0700); err != nil {
		return err
	}

	out, err := os.Create(filepath.Clean(destPath))
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	_, err = io.Copy(out, resp.Body)
	return err
}

// downloadEpisodeWithProgress downloads a single episode with a Bubble Tea progress bar
func (d *NineAnimeDownloader) downloadEpisodeWithProgress(anime *models.Anime, episode *scraper.NineAnimeEpisode, outputDir string) error {
	epPath := filepath.Join(outputDir, d.episodeFilename(anime, episode.Number))

	// Skip if already exists and is reasonably sized
	if info, err := os.Stat(epPath); err == nil && info.Size() > 1024*1024 {
		sizeMB := float64(info.Size()) / (1024 * 1024)
		fmt.Printf("Episode %d already exists (%.1f MB): %s\n", episode.Number, sizeMB, epPath)
		return nil
	}

	fmt.Printf("Resolving stream for episode %d (%s)...\n", episode.Number, episode.Title)

	streamURL, referer, subtitles, err := d.resolveStream(episode.EpisodeID)
	if err != nil {
		return fmt.Errorf("failed to resolve stream for ep %d: %w", episode.Number, err)
	}

	// Prompt subtitle language BEFORE starting the Bubble Tea progress UI.
	// This is the only safe moment: stdin is free, stream is resolved, tracks are known.
	// The choice is cached — subsequent episodes reuse the same selection.
	if len(subtitles) > 0 && !d.subLangResolved {
		fmt.Printf("\nFound %d subtitle track(s) for this anime:\n", len(subtitles))
		d.promptSubtitleLanguage(subtitles)
		fmt.Println() // blank line before progress bar
	}

	// Estimate size for progress bar
	estimatedSize := int64(400 * 1024 * 1024) // 400MB default for HLS

	m := &progressModel{
		progress:   progress.New(progress.WithDefaultBlend()),
		totalBytes: estimatedSize,
	}

	p := tea.NewProgram(m)

	downloadComplete := make(chan error, 1)
	go func() {
		dlErr := d.downloadStream(streamURL, epPath, referer, m)

		if dlErr == nil {
			// Download subtitles (uses cached language choice)
			d.downloadSubtitles(subtitles, epPath)

			p.Send(statusMsg("Download completed!"))
			time.Sleep(800 * time.Millisecond)
		} else {
			p.Send(statusMsg(fmt.Sprintf("Download failed: %v", dlErr)))
			time.Sleep(500 * time.Millisecond)
		}

		m.mu.Lock()
		m.done = true
		m.mu.Unlock()
		p.Quit()

		downloadComplete <- dlErr
	}()

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("progress display error: %w", err)
	}

	if err := <-downloadComplete; err != nil {
		return err
	}

	// Verify
	if info, err := os.Stat(epPath); err != nil || info.Size() < 1024 {
		return fmt.Errorf("download verification failed for episode %d", episode.Number)
	}

	fmt.Printf("Episode %d downloaded successfully!\n", episode.Number)
	return nil
}

// downloadBatchWithProgress downloads multiple episodes with a shared progress bar
func (d *NineAnimeDownloader) downloadBatchWithProgress(anime *models.Anime, episodes []scraper.NineAnimeEpisode, outputDir string) error {
	total := len(episodes)

	// Check which episodes need downloading
	var toDownload []scraper.NineAnimeEpisode
	var skipped int
	for _, ep := range episodes {
		epPath := filepath.Join(outputDir, d.episodeFilename(anime, ep.Number))
		if info, err := os.Stat(epPath); err == nil && info.Size() > 1024*1024 {
			skipped++
			fmt.Printf("  Skipping episode %d (already exists)\n", ep.Number)
			continue
		}
		toDownload = append(toDownload, ep)
	}

	if len(toDownload) == 0 {
		fmt.Printf("All %d episode(s) already downloaded!\n", total)
		return nil
	}

	if skipped > 0 {
		fmt.Printf("Skipped %d existing episode(s), downloading %d...\n", skipped, len(toDownload))
	}

	// Download sequentially for 9anime (rate limiting)
	var successCount, failCount int
	var totalSizeMB float64

	for i, ep := range toDownload {
		fmt.Printf("\n[%d/%d] Episode %d: %s\n", i+1, len(toDownload), ep.Number, ep.Title)

		err := d.downloadEpisodeWithProgress(anime, &ep, outputDir)
		if err != nil {
			util.Errorf("Failed to download episode %d: %v", ep.Number, err)
			failCount++
			continue
		}
		successCount++

		// Track size
		epPath := filepath.Join(outputDir, d.episodeFilename(anime, ep.Number))
		if info, statErr := os.Stat(epPath); statErr == nil {
			totalSizeMB += float64(info.Size()) / (1024 * 1024)
		}

		// Small delay between episodes to be friendly to the server
		if i < len(toDownload)-1 {
			time.Sleep(1 * time.Second)
		}
	}

	// Summary
	fmt.Printf("\n%s\n", strings.Repeat("═", 60))
	fmt.Printf("  Download Summary\n")
	fmt.Printf("  ✓ Success: %d/%d\n", successCount, len(toDownload))
	if failCount > 0 {
		fmt.Printf("  ✗ Failed:  %d\n", failCount)
	}
	fmt.Printf("  📦 Total:  %.1f MB\n", totalSizeMB)
	fmt.Printf("  📁 Path:   %s\n", outputDir)
	fmt.Printf("%s\n", strings.Repeat("═", 60))

	if failCount > 0 {
		return fmt.Errorf("%d episode(s) failed to download", failCount)
	}
	return nil
}

// downloadStream downloads an HLS stream to a file, trying native HLS first then yt-dlp
func (d *NineAnimeDownloader) downloadStream(streamURL, destPath, referer string, m *progressModel) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// For m3u8 streams, try native HLS first (handles obfuscated segments)
	if strings.Contains(streamURL, ".m3u8") {
		err := d.downloadNativeHLS(streamURL, destPath, referer, m)
		if err == nil {
			return nil
		}
		util.Warnf("Native HLS failed, falling back to yt-dlp: %v", err)
	}

	// Fallback to yt-dlp
	return d.downloadWithYtDlp(streamURL, destPath, referer, m)
}

// downloadNativeHLS downloads using the native HLS downloader with progress tracking
func (d *NineAnimeDownloader) downloadNativeHLS(streamURL, destPath, referer string, m *progressModel) error {
	headers := map[string]string{
		"User-Agent": scraper.NineAnimeUserAgent,
		"Accept":     "*/*",
	}

	if referer != "" {
		headers["Referer"] = referer
		headers["Origin"] = strings.TrimSuffix(referer, "/")
	}

	ctx := context.Background()

	err := hls.DownloadToFile(ctx, streamURL, destPath, headers, func(bytesWritten int64, segmentsWritten, totalSegments int) {
		if m == nil || totalSegments <= 0 {
			return
		}

		m.mu.Lock()
		defer m.mu.Unlock()

		m.received = bytesWritten

		// Dynamically estimate total from average segment size
		if segmentsWritten >= 3 {
			avgBytesPerSeg := bytesWritten / int64(segmentsWritten)
			estimatedTotal := avgBytesPerSeg * int64(totalSegments)
			if estimatedTotal > m.totalBytes {
				m.totalBytes = estimatedTotal
			}
		}

		// Cap at 98% to prevent showing 100% prematurely
		if m.totalBytes > 0 && m.received >= m.totalBytes {
			m.received = int64(float64(m.totalBytes) * 0.98)
		}
	})

	if err != nil {
		return fmt.Errorf("native HLS download failed: %w", err)
	}

	// Set real 100% from actual file size
	if m != nil {
		if fi, statErr := os.Stat(destPath); statErr == nil && fi.Size() > 0 {
			m.mu.Lock()
			m.totalBytes = fi.Size()
			m.received = fi.Size()
			m.mu.Unlock()
		}
	}

	return nil
}

// downloadWithYtDlp downloads using yt-dlp with Chrome impersonation
func (d *NineAnimeDownloader) downloadWithYtDlp(streamURL, destPath, referer string, m *progressModel) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	// Install yt-dlp if needed
	if _, installErr := ytdlp.Install(ctx, nil); installErr != nil {
		return fmt.Errorf("failed to install yt-dlp: %w", installErr)
	}

	dl := ytdlp.New().
		Output(destPath).
		Format("bestvideo+bestaudio/best").
		Downloader("ffmpeg").
		DownloaderArgs("ffmpeg_i:-allowed_extensions ALL").
		ConcurrentFragments(4).
		FragmentRetries("5").
		Retries("5").
		SocketTimeout(30).
		Impersonate("chrome")

	if referer != "" {
		dl.AddHeaders("Referer:" + referer)
		if parsed, err := url.Parse(referer); err == nil {
			origin := fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
			dl.AddHeaders("Origin:" + origin)
		}
	}

	// Real-time progress via yt-dlp callback
	var lastReportedBytes int64
	var lastProgressFile string
	if m != nil {
		dl.ProgressFunc(200*time.Millisecond, func(update ytdlp.ProgressUpdate) {
			if update.Status == ytdlp.ProgressStatusPostProcessing ||
				update.Status == ytdlp.ProgressStatusFinished {
				return
			}

			m.mu.Lock()
			defer m.mu.Unlock()

			if update.Filename != "" && update.Filename != lastProgressFile {
				lastProgressFile = update.Filename
				lastReportedBytes = 0
			}

			downloaded := int64(update.DownloadedBytes)
			if delta := downloaded - lastReportedBytes; delta > 0 {
				m.received += delta
				lastReportedBytes = downloaded
			}

			if update.TotalBytes > 0 && m.totalBytes < int64(update.TotalBytes) {
				m.totalBytes = int64(update.TotalBytes)
			}
		})
	}

	// Run with retry logic
	var runErr error
	const maxRetries = 2
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			util.Infof("Retrying download (attempt %d/%d)...", attempt+1, maxRetries+1)
			time.Sleep(time.Duration(attempt*2) * time.Second)
			lastReportedBytes = 0
			lastProgressFile = ""
		}

		_, runErr = dl.Run(ctx, streamURL, "--hls-use-mpegts")
		if runErr == nil {
			break
		}

		if attempt < maxRetries && isRetryableDownloadError(runErr) {
			continue
		}
		break
	}

	if runErr != nil {
		return fmt.Errorf("yt-dlp download failed: %w", runErr)
	}

	return nil
}

// isRetryableDownloadError checks if a download error is retryable
func isRetryableDownloadError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "timeout") ||
		strings.Contains(s, "connection") ||
		strings.Contains(s, "network") ||
		strings.Contains(s, "reset") ||
		strings.Contains(s, "refused") ||
		strings.Contains(s, "temporary")
}
