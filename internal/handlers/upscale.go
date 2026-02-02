package handlers

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/upscaler"
	"github.com/alvarorichard/Goanime/internal/util"
)

// HandleUpscaleRequest processes upscale requests
func HandleUpscaleRequest() error {
	// Initialize logger for upscale process
	util.InitLogger()

	if util.GlobalUpscaleRequest == nil {
		return fmt.Errorf("upscale request is nil")
	}

	req := util.GlobalUpscaleRequest

	// Validate FFmpeg is available
	ffmpegVersion, err := upscaler.ValidateFFmpeg()
	if err != nil {
		return fmt.Errorf("FFmpeg validation failed: %w", err)
	}
	util.Infof("Using %s", ffmpegVersion)

	// Determine output path if not specified
	outputPath := req.OutputPath
	if outputPath == "" {
		ext := filepath.Ext(req.InputPath)
		base := strings.TrimSuffix(req.InputPath, ext)
		outputPath = fmt.Sprintf("%s_upscaled%s", base, ext)
	}

	// Determine if it's an image or video based on extension
	ext := strings.ToLower(filepath.Ext(req.InputPath))
	isImage := isImageExtension(ext)

	// Create Anime4K options
	anime4kOpts := upscaler.DefaultOptions()
	anime4kOpts.ScaleFactor = req.ScaleFactor
	anime4kOpts.Passes = req.Passes
	anime4kOpts.StrengthColor = req.StrengthColor
	anime4kOpts.StrengthGradient = req.StrengthGradient
	anime4kOpts.FastMode = req.FastMode

	if req.HighQuality {
		anime4kOpts = upscaler.HighQualityOptions()
		anime4kOpts.ScaleFactor = req.ScaleFactor
	} else if req.FastMode {
		anime4kOpts = upscaler.FastOptions()
		anime4kOpts.ScaleFactor = req.ScaleFactor
	}

	if isImage {
		return handleImageUpscale(req.InputPath, outputPath, anime4kOpts)
	}

	return handleVideoUpscale(req, outputPath, anime4kOpts)
}

// handleImageUpscale processes a single image
func handleImageUpscale(inputPath, outputPath string, opts upscaler.Anime4KOptions) error {
	util.Infof("Upscaling image: %s", inputPath)
	util.Infof("Output: %s", outputPath)
	util.Infof("Scale factor: %dx, Passes: %d, Fast mode: %v",
		opts.ScaleFactor, opts.Passes, opts.FastMode)

	startTime := time.Now()

	if err := upscaler.UpscaleImageFile(inputPath, outputPath, opts); err != nil {
		return fmt.Errorf("image upscale failed: %w", err)
	}

	elapsed := time.Since(startTime)
	util.Infof("Image upscaling complete in %v", elapsed.Round(time.Millisecond))
	util.Infof("Output saved to: %s", outputPath)

	return nil
}

// handleVideoUpscale processes a video file
func handleVideoUpscale(req *util.UpscaleRequest, outputPath string, opts upscaler.Anime4KOptions) error {
	util.Infof("Upscaling video: %s", req.InputPath)
	util.Infof("Output: %s", outputPath)
	util.Infof("Scale factor: %dx, Passes: %d, Fast mode: %v",
		opts.ScaleFactor, opts.Passes, opts.FastMode)

	// Get video info for estimation
	info, err := upscaler.GetVideoInfo(req.InputPath)
	if err != nil {
		util.Warnf("Could not get video info: %v", err)
	} else {
		util.Infof("Video: %dx%d @ %.2f fps, Duration: %.2f seconds",
			info.Width, info.Height, info.FrameRate, info.Duration)

		// Estimate time
		estimated, err := upscaler.EstimateUpscaleTime(req.InputPath, opts)
		if err == nil {
			util.Infof("Estimated processing time: %v", estimated.Round(time.Minute))
		}
	}

	// Create video upscaler config
	config := upscaler.DefaultVideoConfig()
	config.InputPath = req.InputPath
	config.OutputPath = outputPath
	config.Anime4KOptions = opts
	config.PreserveAudio = req.PreserveAudio
	config.UseGPUEncoding = req.UseGPU
	config.VideoBitrate = req.VideoBitrate

	if req.Workers > 0 {
		config.Workers = req.Workers
	}

	// Create upscaler
	videoUpscaler, err := upscaler.NewVideoUpscaler(config)
	if err != nil {
		return fmt.Errorf("failed to create video upscaler: %w", err)
	}
	defer videoUpscaler.Close()

	startTime := time.Now()

	// Run upscaling
	ctx := context.Background()
	if err := videoUpscaler.UpscaleVideo(ctx); err != nil {
		return fmt.Errorf("video upscale failed: %w", err)
	}

	elapsed := time.Since(startTime)
	util.Infof("Video upscaling complete in %v", elapsed.Round(time.Second))
	util.Infof("Output saved to: %s", outputPath)

	return nil
}

// isImageExtension checks if the file extension is a supported image format
func isImageExtension(ext string) bool {
	imageExts := map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
		".gif":  true,
		".bmp":  true,
		".tiff": true,
		".webp": true,
	}
	return imageExts[ext]
}
