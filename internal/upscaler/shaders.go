// Package upscaler provides anime video upscaling using the Anime4K algorithm
// This file handles real-time upscaling via mpv GLSL shaders
package upscaler

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alvarorichard/Goanime/internal/util"
)

const (
	// Anime4K shader release URL
	anime4kShaderURL = "https://github.com/bloc97/Anime4K/releases/download/v4.0.1/Anime4K_v4.0.zip"
	// Anime4K GAN UUL shaders URL (experimental for 360p->4K)
	anime4kGANShaderBaseURL = "https://raw.githubusercontent.com/bloc97/Anime4K/master/glsl/"
	// Shader directory name
	shaderDirName = "anime4k-shaders"
)

// ShaderMode represents the upscaling quality mode
type ShaderMode int

const (
	// ShaderModeOff disables real-time upscaling
	ShaderModeOff ShaderMode = iota
	// ShaderModeFast uses Mode A (Fast) - for text-heavy anime like subtitled content
	ShaderModeFast
	// ShaderModeBalanced uses Mode B (Balanced) - general purpose anime upscaling
	ShaderModeBalanced
	// ShaderModeQuality uses Mode C (Quality) - for high-quality anime films
	ShaderModeQuality
	// ShaderModePerformance uses lighter shaders for weaker GPUs
	ShaderModePerformance
	// ShaderModeUltra uses maximum enhancement - very visible on low quality sources
	ShaderModeUltra
	// ShaderModeGAN_UUL uses experimental SRGAN shaders for 360p->4K upscaling
	// Best for very low resolution anime (360p, 480p). Requires powerful GPU.
	ShaderModeGAN_UUL
	// ShaderModeAdvancedAA uses Mode A+A (highest perceptual quality, double restore)
	ShaderModeAdvancedAA
	// ShaderModeAdvancedBB uses Mode B+B (high perceptual quality for 720p content)
	ShaderModeAdvancedBB
	// ShaderModeAdvancedCA uses Mode C+A (quality mode with added restore for upscaled content)
	ShaderModeAdvancedCA
)

// CurrentShaderMode holds the current real-time upscaling mode
var CurrentShaderMode = ShaderModeOff

// GetShaderDir returns the path to the shader directory
func GetShaderDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home directory
		home, _ := os.UserHomeDir()
		configDir = home
	}
	return filepath.Join(configDir, "goanime", shaderDirName)
}

// ShadersInstalled checks if Anime4K shaders are installed
func ShadersInstalled() bool {
	shaderDir := GetShaderDir()
	// Check for key shader files
	requiredFiles := []string{
		"Anime4K_Clamp_Highlights.glsl",
		"Anime4K_Restore_CNN_M.glsl",
		"Anime4K_Upscale_CNN_x2_M.glsl",
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(filepath.Join(shaderDir, file)); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// GANShadersInstalled checks if experimental GAN UUL shaders are installed
func GANShadersInstalled() bool {
	shaderDir := GetShaderDir()
	// Check for GAN UUL shader files
	requiredFiles := []string{
		"Anime4K_Restore_GAN_UUL.glsl",
		"Anime4K_Upscale_GAN_x4_UUL.glsl",
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(filepath.Join(shaderDir, file)); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// InstallShaders downloads and installs Anime4K shaders
func InstallShaders() error {
	shaderDir := GetShaderDir()

	// Create shader directory
	if err := os.MkdirAll(shaderDir, 0750); err != nil {
		return fmt.Errorf("failed to create shader directory: %w", err)
	}

	util.Info("Downloading Anime4K shaders...")

	// Download the shader zip
	resp, err := http.Get(anime4kShaderURL)
	if err != nil {
		return fmt.Errorf("failed to download shaders: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			util.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download shaders: HTTP %d", resp.StatusCode)
	}

	// Create temp file for zip
	tmpFile, err := os.CreateTemp("", "anime4k-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		if removeErr := os.Remove(tmpPath); removeErr != nil {
			util.Debugf("Failed to remove temp file: %v", removeErr)
		}
	}()

	// Copy download to temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to save download: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Extract zip
	util.Info("Extracting shaders...")
	if err := extractZip(tmpPath, shaderDir); err != nil {
		return fmt.Errorf("failed to extract shaders: %w", err)
	}

	util.Infof("Anime4K shaders installed to: %s", shaderDir)
	return nil
}

// InstallGANShaders downloads and installs experimental GAN UUL shaders for 360p->4K upscaling
// These shaders are very heavy and require a powerful GPU
func InstallGANShaders() error {
	shaderDir := GetShaderDir()

	// Create shader directory
	if err := os.MkdirAll(shaderDir, 0750); err != nil {
		return fmt.Errorf("failed to create shader directory: %w", err)
	}

	// GAN UUL shader files to download
	ganShaders := []string{
		"Restore/Anime4K_Restore_GAN_UUL.glsl",
		"Upscale/Anime4K_Upscale_GAN_x4_UUL.glsl",
	}

	util.Info("Downloading experimental GAN UUL shaders for 360p->4K upscaling...")
	util.Warn("Note: These shaders are very heavy and require a powerful GPU!")

	for _, shaderPath := range ganShaders {
		url := anime4kGANShaderBaseURL + shaderPath
		fileName := filepath.Base(shaderPath)
		destPath := filepath.Join(shaderDir, fileName)

		util.Infof("Downloading %s...", fileName)

		// #nosec G107 -- URL is constructed from trusted constant anime4kGANShaderBaseURL
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to download %s: %w", fileName, err)
		}

		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			return fmt.Errorf("failed to download %s: HTTP %d", fileName, resp.StatusCode)
		}

		// Create destination file
		// #nosec G304 -- destPath is constructed from trusted config directory
		outFile, err := os.Create(destPath)
		if err != nil {
			_ = resp.Body.Close()
			return fmt.Errorf("failed to create file %s: %w", fileName, err)
		}

		// Copy with size limit (GAN shaders can be large, up to 10MB)
		const maxShaderSize = 10 * 1024 * 1024 // 10MB
		limitedReader := io.LimitReader(resp.Body, maxShaderSize)
		_, copyErr := io.Copy(outFile, limitedReader)
		closeOutErr := outFile.Close()
		closeRespErr := resp.Body.Close()

		if copyErr != nil {
			return fmt.Errorf("failed to save %s: %w", fileName, copyErr)
		}
		if closeOutErr != nil {
			return fmt.Errorf("failed to close %s: %w", fileName, closeOutErr)
		}
		if closeRespErr != nil {
			util.Warnf("Failed to close response body: %v", closeRespErr)
		}
	}

	util.Infof("GAN UUL shaders installed to: %s", shaderDir)
	util.Info("Recommended usage: For 360p to 4K upscaling of low resolution anime")
	return nil
}

// extractZip extracts a zip file to the destination directory
func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			util.Warnf("Failed to close zip reader: %v", closeErr)
		}
	}()

	for _, f := range r.File {
		// Only extract .glsl files
		if !strings.HasSuffix(f.Name, ".glsl") {
			continue
		}

		// Get just the filename (strip any directory path)
		fileName := filepath.Base(f.Name)
		destPath := filepath.Join(destDir, fileName)

		// Open source file in zip
		rc, err := f.Open()
		if err != nil {
			return err
		}

		// Create destination file
		// #nosec G304 G703 -- destPath is constructed from trusted config directory
		outFile, err := os.Create(destPath)
		if err != nil {
			_ = rc.Close()
			return err
		}

		// Copy contents with size limit to prevent decompression bomb
		// Anime4K shaders are typically under 100KB each, so 1MB is generous
		const maxShaderSize = 1 * 1024 * 1024 // 1MB
		limitedReader := io.LimitReader(rc, maxShaderSize)
		// #nosec G110 -- size limited by LimitReader to prevent decompression bomb
		_, copyErr := io.Copy(outFile, limitedReader)
		closeOutErr := outFile.Close()
		closeRcErr := rc.Close()

		if copyErr != nil {
			return copyErr
		}
		if closeOutErr != nil {
			return fmt.Errorf("failed to close output file: %w", closeOutErr)
		}
		if closeRcErr != nil {
			return fmt.Errorf("failed to close zip entry: %w", closeRcErr)
		}
	}

	return nil
}

// GetMPVShaderArgs returns mpv arguments for real-time upscaling
func GetMPVShaderArgs(mode ShaderMode) []string {
	if mode == ShaderModeOff {
		return nil
	}

	shaderDir := GetShaderDir()
	if !ShadersInstalled() {
		util.Warn("Anime4K shaders not installed. Run upscale setup first.")
		return nil
	}

	// Build shader path helper
	shader := func(name string) string {
		return filepath.Join(shaderDir, name)
	}

	var shaders []string

	switch mode {
	case ShaderModeFast:
		// Mode A (Fast) - Good for text-heavy anime
		// Optimized for subtitled content where text clarity is important
		shaders = []string{
			shader("Anime4K_Clamp_Highlights.glsl"),
			shader("Anime4K_Restore_CNN_M.glsl"),
			shader("Anime4K_Upscale_CNN_x2_M.glsl"),
			shader("Anime4K_AutoDownscalePre_x2.glsl"),
			shader("Anime4K_AutoDownscalePre_x4.glsl"),
			shader("Anime4K_Upscale_CNN_x2_S.glsl"),
		}

	case ShaderModeBalanced:
		// Mode B (Balanced) - General purpose anime
		// Good balance between quality and performance
		shaders = []string{
			shader("Anime4K_Clamp_Highlights.glsl"),
			shader("Anime4K_Restore_CNN_Soft_M.glsl"),
			shader("Anime4K_Upscale_CNN_x2_M.glsl"),
			shader("Anime4K_AutoDownscalePre_x2.glsl"),
			shader("Anime4K_AutoDownscalePre_x4.glsl"),
			shader("Anime4K_Upscale_CNN_x2_S.glsl"),
		}

	case ShaderModeQuality:
		// Mode C (Quality) - High quality for anime films
		// Best quality but requires good GPU
		shaders = []string{
			shader("Anime4K_Clamp_Highlights.glsl"),
			shader("Anime4K_Upscale_Denoise_CNN_x2_M.glsl"),
			shader("Anime4K_AutoDownscalePre_x2.glsl"),
			shader("Anime4K_AutoDownscalePre_x4.glsl"),
			shader("Anime4K_Upscale_CNN_x2_L.glsl"),
		}

	case ShaderModePerformance:
		// Mode for weaker GPUs - minimal shader chain
		shaders = []string{
			shader("Anime4K_Clamp_Highlights.glsl"),
			shader("Anime4K_Upscale_CNN_x2_S.glsl"),
		}

	case ShaderModeUltra:
		// Ultra mode - maximum enhancement chain, very visible on SD sources
		// Uses multiple denoise + upscale passes for dramatic improvement
		shaders = []string{
			shader("Anime4K_Clamp_Highlights.glsl"),
			shader("Anime4K_Denoise_Bilateral_Mode.glsl"),
			shader("Anime4K_Deblur_DoG.glsl"),
			shader("Anime4K_Darken_HQ.glsl"),
			shader("Anime4K_Thin_HQ.glsl"),
			shader("Anime4K_Upscale_Denoise_CNN_x2_VL.glsl"),
			shader("Anime4K_AutoDownscalePre_x2.glsl"),
			shader("Anime4K_AutoDownscalePre_x4.glsl"),
			shader("Anime4K_Upscale_CNN_x2_L.glsl"),
		}

	case ShaderModeGAN_UUL:
		// GAN UUL mode - experimental SRGAN shaders for 360p->4K upscaling
		// Recommended for very low resolution anime (360p, 480p)
		// WARNING: Very heavy shaders, requires powerful GPU!
		if !GANShadersInstalled() {
			util.Warn("GAN UUL shaders not installed. Run 'goanime --upscale-install-gan' first.")
			return nil
		}
		shaders = []string{
			shader("Anime4K_Restore_GAN_UUL.glsl"),
			shader("Anime4K_Upscale_GAN_x4_UUL.glsl"),
			shader("Anime4K_Restore_CNN_Soft_M.glsl"),
			shader("Anime4K_Upscale_CNN_x2_M.glsl"),
		}

	case ShaderModeAdvancedAA:
		// Mode A+A - highest perceptual quality, double restore
		// Best for 1080p anime with high compression artifacts
		// WARNING: Can cause severe ringing on some content
		shaders = []string{
			shader("Anime4K_Clamp_Highlights.glsl"),
			shader("Anime4K_Restore_CNN_M.glsl"),
			shader("Anime4K_Upscale_CNN_x2_M.glsl"),
			shader("Anime4K_Restore_CNN_S.glsl"),
			shader("Anime4K_AutoDownscalePre_x2.glsl"),
			shader("Anime4K_AutoDownscalePre_x4.glsl"),
			shader("Anime4K_Upscale_CNN_x2_S.glsl"),
		}

	case ShaderModeAdvancedBB:
		// Mode B+B - high perceptual quality for 720p content
		// Good for anime with ringing/aliasing artifacts
		shaders = []string{
			shader("Anime4K_Clamp_Highlights.glsl"),
			shader("Anime4K_Restore_CNN_Soft_M.glsl"),
			shader("Anime4K_Upscale_CNN_x2_M.glsl"),
			shader("Anime4K_Restore_CNN_Soft_S.glsl"),
			shader("Anime4K_AutoDownscalePre_x2.glsl"),
			shader("Anime4K_AutoDownscalePre_x4.glsl"),
			shader("Anime4K_Upscale_CNN_x2_S.glsl"),
		}

	case ShaderModeAdvancedCA:
		// Mode C+A - quality mode with added restore for upscaled content
		// Good for downscaled 1080p->480p anime
		shaders = []string{
			shader("Anime4K_Clamp_Highlights.glsl"),
			shader("Anime4K_Upscale_Denoise_CNN_x2_M.glsl"),
			shader("Anime4K_Restore_CNN_M.glsl"),
			shader("Anime4K_AutoDownscalePre_x2.glsl"),
			shader("Anime4K_AutoDownscalePre_x4.glsl"),
			shader("Anime4K_Upscale_CNN_x2_M.glsl"),
		}
	}

	// Check that all shaders exist
	var validShaders []string
	for _, s := range shaders {
		if _, err := os.Stat(s); err == nil {
			validShaders = append(validShaders, s)
		} else {
			util.Warnf("Shader not found: %s", s)
		}
	}

	if len(validShaders) == 0 {
		return nil
	}

	// Build mpv arguments for shaders
	var args []string

	// Create shader cache directory
	cacheDir := filepath.Join(shaderDir, "cache")
	if err := os.MkdirAll(cacheDir, 0750); err == nil {
		args = append(args, "--gpu-shader-cache-dir="+cacheDir)
	}

	// Use --glsl-shaders with colon-separated paths for gpu-next/libplacebo
	// This format works better with modern mpv
	shaderPaths := strings.Join(validShaders, ":")
	args = append(args, "--glsl-shaders="+shaderPaths)

	return args
}

// GetShaderModeName returns a human-readable name for the shader mode
func GetShaderModeName(mode ShaderMode) string {
	switch mode {
	case ShaderModeOff:
		return "Off"
	case ShaderModeFast:
		return "Fast (Mode A)"
	case ShaderModeBalanced:
		return "Balanced (Mode B)"
	case ShaderModeQuality:
		return "Quality (Mode C)"
	case ShaderModePerformance:
		return "Performance (Low GPU)"
	case ShaderModeUltra:
		return "Ultra (Max Enhancement)"
	case ShaderModeGAN_UUL:
		return "GAN UUL (360p→4K)"
	case ShaderModeAdvancedAA:
		return "Advanced A+A (Max Perceptual)"
	case ShaderModeAdvancedBB:
		return "Advanced B+B (720p Optimized)"
	case ShaderModeAdvancedCA:
		return "Advanced C+A (Upscaled Content)"
	default:
		return "Unknown"
	}
}

// GetShaderModeDescription returns a detailed description for the shader mode
func GetShaderModeDescription(mode ShaderMode) string {
	switch mode {
	case ShaderModeOff:
		return "Real-time upscaling disabled"
	case ShaderModeFast:
		return "Mode A: Optimized for text-heavy anime with subtitles"
	case ShaderModeBalanced:
		return "Mode B: General purpose anime upscaling, good balance"
	case ShaderModeQuality:
		return "Mode C: High quality for anime films, requires good GPU"
	case ShaderModePerformance:
		return "Minimal shaders for weaker GPUs"
	case ShaderModeUltra:
		return "Maximum enhancement with denoise + deblur + upscale"
	case ShaderModeGAN_UUL:
		return "Experimental SRGAN for 360p→4K, VERY heavy, requires powerful GPU"
	case ShaderModeAdvancedAA:
		return "Double restore pass, highest perceptual quality, may cause ringing"
	case ShaderModeAdvancedBB:
		return "Double soft restore, good for 720p with aliasing artifacts"
	case ShaderModeAdvancedCA:
		return "Quality + restore, good for downscaled 480p anime"
	default:
		return "Unknown mode"
	}
}

// GetAllShaderModes returns all available shader modes
func GetAllShaderModes() []ShaderMode {
	return []ShaderMode{
		ShaderModeOff,
		ShaderModeFast,
		ShaderModeBalanced,
		ShaderModeQuality,
		ShaderModePerformance,
		ShaderModeUltra,
		ShaderModeGAN_UUL,
		ShaderModeAdvancedAA,
		ShaderModeAdvancedBB,
		ShaderModeAdvancedCA,
	}
}

// GetAdvancedShaderModes returns only the advanced shader modes (for menu)
func GetAdvancedShaderModes() []ShaderMode {
	return []ShaderMode{
		ShaderModeGAN_UUL,
		ShaderModeAdvancedAA,
		ShaderModeAdvancedBB,
		ShaderModeAdvancedCA,
	}
}

// CycleShaderMode cycles through shader modes
func CycleShaderMode() ShaderMode {
	CurrentShaderMode = (CurrentShaderMode + 1) % 10
	return CurrentShaderMode
}

// SetShaderMode sets the shader mode
func SetShaderMode(mode ShaderMode) {
	CurrentShaderMode = mode
}
