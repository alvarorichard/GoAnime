// Package upscaler provides video upscaling using Anime4K algorithm
package upscaler

import (
	"bufio"
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

// VideoUpscaleConfig holds configuration for video upscaling
type VideoUpscaleConfig struct {
	InputPath      string
	OutputPath     string
	Anime4KOptions Anime4KOptions
	FFmpegPath     string
	FFprobePath    string
	Workers        int  // Number of parallel frame processors
	KeepTempFiles  bool // Keep temporary frame files for debugging
	UseGPUEncoding bool // Use GPU hardware encoding if available
	PreserveAudio  bool // Preserve original audio track
	OutputFormat   string
	VideoBitrate   string
	AudioBitrate   string
	FrameRate      float64 // 0 means preserve original
}

// DefaultVideoConfig returns default video upscaling configuration
func DefaultVideoConfig() VideoUpscaleConfig {
	ffmpegPath := "ffmpeg"
	ffprobePath := "ffprobe"

	// Try to find FFmpeg in common locations
	if runtime.GOOS == "darwin" {
		// Check Homebrew locations
		if _, err := os.Stat("/opt/homebrew/bin/ffmpeg"); err == nil {
			ffmpegPath = "/opt/homebrew/bin/ffmpeg"
			ffprobePath = "/opt/homebrew/bin/ffprobe"
		} else if _, err := os.Stat("/usr/local/bin/ffmpeg"); err == nil {
			ffmpegPath = "/usr/local/bin/ffmpeg"
			ffprobePath = "/usr/local/bin/ffprobe"
		}
	}

	return VideoUpscaleConfig{
		Anime4KOptions: DefaultOptions(),
		FFmpegPath:     ffmpegPath,
		FFprobePath:    ffprobePath,
		Workers:        runtime.NumCPU(),
		KeepTempFiles:  false,
		UseGPUEncoding: false,
		PreserveAudio:  true,
		OutputFormat:   "mp4",
		VideoBitrate:   "8M",
		AudioBitrate:   "192k",
		FrameRate:      0,
	}
}

// VideoUpscaler handles video upscaling operations
type VideoUpscaler struct {
	config    VideoUpscaleConfig
	upscaler  *Anime4KUpscaler
	tempDir   string
	frameInfo VideoFrameInfo
}

// VideoFrameInfo contains information about the video
type VideoFrameInfo struct {
	Width       int
	Height      int
	FrameRate   float64
	Duration    float64
	TotalFrames int
	Codec       string
	AudioCodec  string
}

// NewVideoUpscaler creates a new video upscaler
func NewVideoUpscaler(config VideoUpscaleConfig) (*VideoUpscaler, error) {
	// Verify FFmpeg is available
	if err := verifyFFmpeg(config.FFmpegPath); err != nil {
		return nil, fmt.Errorf("FFmpeg not available: %w", err)
	}

	return &VideoUpscaler{
		config:   config,
		upscaler: NewAnime4KUpscaler(config.Anime4KOptions),
	}, nil
}

// Close cleans up resources used by the video upscaler
func (v *VideoUpscaler) Close() {
	if v.upscaler != nil {
		v.upscaler.Close()
	}
}

// verifyFFmpeg checks if FFmpeg is available and working
func verifyFFmpeg(ffmpegPath string) error {
	cmd := exec.Command(ffmpegPath, "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("FFmpeg not found at %s: %w", ffmpegPath, err)
	}
	return nil
}

// UpscaleVideo upscales a video file using the Anime4K algorithm
func (v *VideoUpscaler) UpscaleVideo(ctx context.Context) error {
	util.Info("Starting video upscaling with Anime4K algorithm...")

	// Create temporary directory for frames
	tempDir, err := os.MkdirTemp("", "goanime_upscale_")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	v.tempDir = tempDir

	if !v.config.KeepTempFiles {
		defer func() {
			util.Debugf("Cleaning up temp directory: %s", tempDir)
			if err := os.RemoveAll(tempDir); err != nil {
				util.Warnf("Failed to clean up temp directory: %v", err)
			}
		}()
	} else {
		util.Infof("Temp files will be kept at: %s", tempDir)
	}

	// Step 1: Probe video for information
	util.Info("Analyzing video...")
	if err := v.probeVideo(); err != nil {
		return fmt.Errorf("failed to probe video: %w", err)
	}
	util.Infof("Video: %dx%d @ %.2f fps, Duration: %.2f seconds, ~%d frames",
		v.frameInfo.Width, v.frameInfo.Height, v.frameInfo.FrameRate,
		v.frameInfo.Duration, v.frameInfo.TotalFrames)

	// Step 2: Extract frames
	util.Info("Extracting frames...")
	framesDir := filepath.Join(tempDir, "frames")
	if err := os.MkdirAll(framesDir, 0750); err != nil {
		return fmt.Errorf("failed to create frames directory: %w", err)
	}
	if err := v.extractFrames(ctx, framesDir); err != nil {
		return fmt.Errorf("failed to extract frames: %w", err)
	}

	// Step 3: Upscale frames with progress
	util.Info("Upscaling frames with Anime4K...")
	upscaledDir := filepath.Join(tempDir, "upscaled")
	if err := os.MkdirAll(upscaledDir, 0750); err != nil {
		return fmt.Errorf("failed to create upscaled directory: %w", err)
	}
	if err := v.upscaleFrames(ctx, framesDir, upscaledDir); err != nil {
		return fmt.Errorf("failed to upscale frames: %w", err)
	}

	// Step 4: Encode video
	util.Info("Encoding upscaled video...")
	if err := v.encodeVideo(ctx, upscaledDir); err != nil {
		return fmt.Errorf("failed to encode video: %w", err)
	}

	util.Infof("Video upscaling complete! Output: %s", v.config.OutputPath)
	return nil
}

// probeVideo gets video information using ffprobe
func (v *VideoUpscaler) probeVideo() error {
	// Get video stream info
	// #nosec G204 -- FFprobe path is application-controlled, not user input
	cmd := exec.Command(v.config.FFprobePath,
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,r_frame_rate,codec_name,duration",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		filepath.Clean(v.config.InputPath),
	)

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("ffprobe failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		return fmt.Errorf("no video stream found")
	}

	// Parse first line (video stream info)
	parts := strings.Split(lines[0], ",")
	if len(parts) >= 4 {
		v.frameInfo.Width, _ = strconv.Atoi(parts[0])
		v.frameInfo.Height, _ = strconv.Atoi(parts[1])

		// Parse frame rate (e.g., "24000/1001" or "30/1")
		if strings.Contains(parts[2], "/") {
			fpsParts := strings.Split(parts[2], "/")
			if len(fpsParts) == 2 {
				num, _ := strconv.ParseFloat(fpsParts[0], 64)
				den, _ := strconv.ParseFloat(fpsParts[1], 64)
				if den > 0 {
					v.frameInfo.FrameRate = num / den
				}
			}
		} else {
			v.frameInfo.FrameRate, _ = strconv.ParseFloat(parts[2], 64)
		}

		v.frameInfo.Codec = parts[3]

		// Try to get duration from stream
		if len(parts) >= 5 && parts[4] != "" && parts[4] != "N/A" {
			v.frameInfo.Duration, _ = strconv.ParseFloat(parts[4], 64)
		}
	}

	// If duration not found in stream, try format duration
	if v.frameInfo.Duration == 0 && len(lines) > 1 {
		v.frameInfo.Duration, _ = strconv.ParseFloat(strings.TrimSpace(lines[1]), 64)
	}

	// If still no duration, use ffprobe with different options
	if v.frameInfo.Duration == 0 {
		// #nosec G204 -- FFprobe path is application-controlled
		durationCmd := exec.Command(v.config.FFprobePath,
			"-v", "error",
			"-show_entries", "format=duration",
			"-of", "default=noprint_wrappers=1:nokey=1",
			filepath.Clean(v.config.InputPath),
		)
		durationOutput, err := durationCmd.Output()
		if err == nil {
			v.frameInfo.Duration, _ = strconv.ParseFloat(strings.TrimSpace(string(durationOutput)), 64)
		}
	}

	// Calculate total frames
	if v.frameInfo.FrameRate > 0 && v.frameInfo.Duration > 0 {
		v.frameInfo.TotalFrames = int(v.frameInfo.FrameRate * v.frameInfo.Duration)
	}

	// Get audio codec
	// #nosec G204 -- FFprobe path is application-controlled
	audioCmd := exec.Command(v.config.FFprobePath,
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=codec_name",
		"-of", "csv=p=0",
		filepath.Clean(v.config.InputPath),
	)
	audioOutput, err := audioCmd.Output()
	if err == nil {
		v.frameInfo.AudioCodec = strings.TrimSpace(string(audioOutput))
	}

	return nil
}

// extractFrames extracts all frames from the video
func (v *VideoUpscaler) extractFrames(ctx context.Context, outputDir string) error {
	args := []string{
		"-i", filepath.Clean(v.config.InputPath),
		"-vsync", "0",
		"-frame_pts", "1",
		filepath.Join(outputDir, "frame_%08d.png"),
	}

	// #nosec G204 -- FFmpeg path is application-controlled
	cmd := exec.CommandContext(ctx, v.config.FFmpegPath, args...)
	cmd.Stderr = os.Stderr

	if util.IsDebug {
		util.Debugf("FFmpeg command: %s %v", v.config.FFmpegPath, args)
	}

	return cmd.Run()
}

// upscaleFrames processes all frames with Anime4K
func (v *VideoUpscaler) upscaleFrames(ctx context.Context, inputDir, outputDir string) error {
	// Get list of frames
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("failed to read frames directory: %w", err)
	}

	var frames []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".png") {
			frames = append(frames, entry.Name())
		}
	}

	if len(frames) == 0 {
		return fmt.Errorf("no frames found to upscale")
	}

	totalFrames := len(frames)
	util.Infof("Processing %d frames with %d workers...", totalFrames, v.config.Workers)

	// Create progress model
	p := progress.New(progress.WithDefaultGradient())
	model := &upscaleProgressModel{
		progress:    p,
		totalFrames: totalFrames,
	}

	program := tea.NewProgram(model)

	// Channel for frame jobs
	jobs := make(chan string, totalFrames)
	results := make(chan error, totalFrames)

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < v.config.Workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for frameName := range jobs {
				select {
				case <-ctx.Done():
					results <- ctx.Err()
					return
				default:
				}

				inputPath := filepath.Join(inputDir, frameName)
				outputPath := filepath.Join(outputDir, frameName)

				err := v.upscaleSingleFrame(inputPath, outputPath)
				if err != nil {
					results <- fmt.Errorf("frame %s: %w", frameName, err)
				} else {
					results <- nil
					model.mu.Lock()
					model.completed++
					program.Send(frameProgressMsg{completed: model.completed})
					model.mu.Unlock()
				}
			}
		}()
	}

	// Send jobs
	go func() {
		for _, frame := range frames {
			jobs <- frame
		}
		close(jobs)
	}()

	// Wait for completion in background
	go func() {
		wg.Wait()
		close(results)
		program.Send(upscaleCompleteMsg{})
	}()

	// Run progress UI
	if _, err := program.Run(); err != nil {
		util.Warnf("Progress display error: %v", err)
	}

	// Check for errors
	var errors []error
	for err := range results {
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%d frames failed to upscale, first error: %v", len(errors), errors[0])
	}

	return nil
}

// upscaleSingleFrame processes a single frame using Anime4KGo
func (v *VideoUpscaler) upscaleSingleFrame(inputPath, outputPath string) error {
	// Open input image
	// #nosec G304 -- inputPath is from temp directory created by this application
	file, err := os.Open(filepath.Clean(inputPath))
	if err != nil {
		return fmt.Errorf("failed to open frame: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode frame: %w", err)
	}

	// Upscale using Anime4KGo via our wrapper
	upscaled, err := v.upscaler.UpscaleImage(img)
	if err != nil {
		return fmt.Errorf("failed to upscale frame: %w", err)
	}

	// Save output
	// #nosec G304 -- outputPath is from temp directory created by this application
	outFile, err := os.Create(filepath.Clean(outputPath))
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}

	encoder := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := encoder.Encode(outFile, upscaled); err != nil {
		_ = outFile.Close()
		return fmt.Errorf("failed to encode output: %w", err)
	}

	if err := outFile.Close(); err != nil {
		return fmt.Errorf("failed to close output file: %w", err)
	}

	return nil
}

// encodeVideo encodes the upscaled frames back to video
func (v *VideoUpscaler) encodeVideo(ctx context.Context, framesDir string) error {
	// Calculate new dimensions (Anime4KGo always upscales 2x)
	newWidth := v.frameInfo.Width * 2
	newHeight := v.frameInfo.Height * 2

	// Build FFmpeg arguments
	args := []string{
		"-framerate", fmt.Sprintf("%.6f", v.frameInfo.FrameRate),
		"-i", filepath.Join(framesDir, "frame_%08d.png"),
	}

	// Add audio from original if preserving
	if v.config.PreserveAudio && v.frameInfo.AudioCodec != "" {
		args = append(args, "-i", v.config.InputPath, "-map", "0:v", "-map", "1:a")
	}

	// Video codec selection
	videoCodec := "libx264"
	if v.config.UseGPUEncoding {
		// Try to use hardware encoding
		switch runtime.GOOS {
		case "darwin":
			videoCodec = "h264_videotoolbox"
		case "linux":
			// Check for NVIDIA GPU
			if _, err := exec.LookPath("nvidia-smi"); err == nil {
				videoCodec = "h264_nvenc"
			}
		case "windows":
			videoCodec = "h264_nvenc"
		}
	}

	args = append(args,
		"-c:v", videoCodec,
		"-b:v", v.config.VideoBitrate,
		"-pix_fmt", "yuv420p",
		"-s", fmt.Sprintf("%dx%d", newWidth, newHeight),
	)

	// Audio codec
	if v.config.PreserveAudio && v.frameInfo.AudioCodec != "" {
		args = append(args, "-c:a", "aac", "-b:a", v.config.AudioBitrate)
	}

	// Output file
	args = append(args, "-y", filepath.Clean(v.config.OutputPath))

	// #nosec G204 -- FFmpeg path is application-controlled
	cmd := exec.CommandContext(ctx, v.config.FFmpegPath, args...)

	if util.IsDebug {
		util.Debugf("FFmpeg encode command: %s %v", v.config.FFmpegPath, args)
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

// Progress model for Bubble Tea
type upscaleProgressModel struct {
	progress    progress.Model
	totalFrames int
	completed   int
	done        bool
	mu          sync.Mutex
}

type frameProgressMsg struct {
	completed int
}

type upscaleCompleteMsg struct{}

func (m *upscaleProgressModel) Init() tea.Cmd {
	return nil
}

func (m *upscaleProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
	case frameProgressMsg:
		m.mu.Lock()
		m.completed = msg.completed
		m.mu.Unlock()
		return m, nil
	case upscaleCompleteMsg:
		m.done = true
		return m, tea.Quit
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}
	return m, nil
}

func (m *upscaleProgressModel) View() string {
	m.mu.Lock()
	completed := m.completed
	m.mu.Unlock()

	percent := float64(completed) / float64(m.totalFrames)
	if m.done {
		return fmt.Sprintf("\n✓ Upscaled %d/%d frames (100%%)\n\n", m.totalFrames, m.totalFrames)
	}
	return fmt.Sprintf("\n%s\nUpscaling: %d/%d frames (%.1f%%)\n",
		m.progress.ViewAs(percent),
		completed, m.totalFrames, percent*100)
}

// UpscaleVideoFile is a convenience function to upscale a video file
func UpscaleVideoFile(inputPath, outputPath string, opts Anime4KOptions) error {
	config := DefaultVideoConfig()
	config.InputPath = inputPath
	config.OutputPath = outputPath
	config.Anime4KOptions = opts

	upscaler, err := NewVideoUpscaler(config)
	if err != nil {
		return err
	}

	return upscaler.UpscaleVideo(context.Background())
}

// GetVideoInfo returns information about a video file
func GetVideoInfo(videoPath string) (*VideoFrameInfo, error) {
	config := DefaultVideoConfig()
	v := &VideoUpscaler{config: config}
	v.config.InputPath = videoPath

	if err := v.probeVideo(); err != nil {
		return nil, err
	}

	return &v.frameInfo, nil
}

// EstimateUpscaleTime estimates the time required to upscale a video
func EstimateUpscaleTime(videoPath string, opts Anime4KOptions) (time.Duration, error) {
	info, err := GetVideoInfo(videoPath)
	if err != nil {
		return 0, err
	}

	// Rough estimate based on frame count and processing speed
	// Assumes ~0.5 seconds per frame on average hardware with parallel processing
	baseTimePerFrame := 500 * time.Millisecond

	// Adjust for quality settings
	timePerFrame := baseTimePerFrame * time.Duration(opts.Passes)
	if opts.FastMode {
		timePerFrame = timePerFrame / 2
	}

	totalTime := time.Duration(info.TotalFrames) * timePerFrame / time.Duration(runtime.NumCPU())

	return totalTime, nil
}

// ValidateFFmpeg checks if FFmpeg is installed and returns version info
func ValidateFFmpeg() (string, error) {
	config := DefaultVideoConfig()

	// #nosec G204 -- FFmpegPath is from DefaultVideoConfig, not user input
	cmd := exec.Command(config.FFmpegPath, "-version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("FFmpeg not found: %w\nPlease install FFmpeg: https://ffmpeg.org/download.html", err)
	}

	// Extract version from first line
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	if scanner.Scan() {
		return scanner.Text(), nil
	}

	return string(output), nil
}
