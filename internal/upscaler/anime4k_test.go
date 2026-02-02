package upscaler

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

// createTestImage creates a simple test image for testing
func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with gradient
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				uint8(x * 255 / width),
				uint8(y * 255 / height),
				128,
				255,
			})
		}
	}

	// Draw some lines (simulating anime-style edges)
	for i := 0; i < width && i < height; i++ {
		img.Set(i, height/2, color.RGBA{0, 0, 0, 255})
		img.Set(width/2, i, color.RGBA{0, 0, 0, 255})
	}

	return img
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Passes != 2 {
		t.Errorf("Expected Passes=2, got %d", opts.Passes)
	}
	if opts.FastMode != false {
		t.Error("Expected FastMode=false")
	}
	if opts.ScaleFactor != 2 {
		t.Errorf("Expected ScaleFactor=2, got %d", opts.ScaleFactor)
	}
}

func TestFastOptions(t *testing.T) {
	opts := FastOptions()

	if opts.Passes != 1 {
		t.Errorf("Expected Passes=1 for fast mode, got %d", opts.Passes)
	}
	if opts.FastMode != true {
		t.Error("Expected FastMode=true")
	}
}

func TestHighQualityOptions(t *testing.T) {
	opts := HighQualityOptions()

	if opts.Passes != 4 {
		t.Errorf("Expected Passes=4 for HQ mode, got %d", opts.Passes)
	}
	if opts.FastMode != false {
		t.Error("Expected FastMode=false for HQ")
	}
}

func TestNewAnime4KUpscaler(t *testing.T) {
	opts := DefaultOptions()
	upscaler := NewAnime4KUpscaler(opts)

	if upscaler == nil {
		t.Fatal("Expected non-nil upscaler")
	}

	// Test Close doesn't panic
	upscaler.Close()
}

func TestUpscaleImage(t *testing.T) {
	opts := FastOptions() // Use fast mode for quicker test
	upscaler := NewAnime4KUpscaler(opts)
	defer upscaler.Close()

	// Create a small test image
	srcImg := createTestImage(50, 50)

	// Upscale the image
	result, err := upscaler.UpscaleImage(srcImg)
	if err != nil {
		t.Fatalf("UpscaleImage failed: %v", err)
	}

	// Verify dimensions (should be 2x)
	expectedWidth := 100
	expectedHeight := 100

	bounds := result.Bounds()
	if bounds.Dx() != expectedWidth || bounds.Dy() != expectedHeight {
		t.Errorf("Expected dimensions %dx%d, got %dx%d",
			expectedWidth, expectedHeight, bounds.Dx(), bounds.Dy())
	}
}

func TestUpscaleImageParallel(t *testing.T) {
	opts := FastOptions()
	upscaler := NewAnime4KUpscaler(opts)
	defer upscaler.Close()

	srcImg := createTestImage(30, 30)

	result, err := upscaler.UpscaleImageParallel(srcImg)
	if err != nil {
		t.Fatalf("UpscaleImageParallel failed: %v", err)
	}

	bounds := result.Bounds()
	if bounds.Dx() != 60 || bounds.Dy() != 60 {
		t.Errorf("Expected dimensions 60x60, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestGetUpscaledDimensions(t *testing.T) {
	tests := []struct {
		inputW, inputH   int
		expectW, expectH int
	}{
		{100, 100, 200, 200},
		{50, 75, 100, 150},
		{1920, 1080, 3840, 2160},
	}

	for _, test := range tests {
		w, h := GetUpscaledDimensions(test.inputW, test.inputH)
		if w != test.expectW || h != test.expectH {
			t.Errorf("GetUpscaledDimensions(%d, %d) = (%d, %d), expected (%d, %d)",
				test.inputW, test.inputH, w, h, test.expectW, test.expectH)
		}
	}
}

func TestAnime4KImage_Process(t *testing.T) {
	srcImg := createTestImage(40, 40)
	a4kImg := NewImageFromImage(srcImg)

	if a4kImg.W != 40 || a4kImg.H != 40 {
		t.Errorf("Expected initial dimensions 40x40, got %dx%d", a4kImg.W, a4kImg.H)
	}

	// Process with fast mode
	a4kImg.Process(1, 1.0/3.0, 1.0, true)

	// After processing, dimensions should be 2x
	if a4kImg.W != 80 || a4kImg.H != 80 {
		t.Errorf("Expected processed dimensions 80x80, got %dx%d", a4kImg.W, a4kImg.H)
	}

	// GetImage should return non-nil
	resultImg := a4kImg.GetImage()
	if resultImg == nil {
		t.Error("GetImage returned nil")
	}
}

func TestLoadImage_FileNotFound(t *testing.T) {
	_, err := LoadImage("/nonexistent/path/to/image.png")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestUpscaleImageFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.png")
	outputPath := filepath.Join(tmpDir, "output.png")

	// Create test image and save it
	srcImg := createTestImage(30, 30)
	a4kImg := NewImageFromImage(srcImg)
	if err := a4kImg.SaveImage(inputPath); err != nil {
		t.Fatalf("Failed to save test image: %v", err)
	}

	// Upscale the image file
	opts := FastOptions()
	if err := UpscaleImageFile(inputPath, outputPath, opts); err != nil {
		t.Fatalf("UpscaleImageFile failed: %v", err)
	}

	// Verify output exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	// Load and verify dimensions
	resultImg, err := LoadImage(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output image: %v", err)
	}

	if resultImg.W != 60 || resultImg.H != 60 {
		t.Errorf("Expected output dimensions 60x60, got %dx%d", resultImg.W, resultImg.H)
	}
}

func TestAnime4KUpscaler_UpscaleAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.png")
	outputPath := filepath.Join(tmpDir, "output.png")

	// Create and save test image
	srcImg := createTestImage(25, 25)
	a4kImg := NewImageFromImage(srcImg)
	if err := a4kImg.SaveImage(inputPath); err != nil {
		t.Fatalf("Failed to save test image: %v", err)
	}

	// Create upscaler and process
	opts := FastOptions()
	upscaler := NewAnime4KUpscaler(opts)
	defer upscaler.Close()

	if err := upscaler.UpscaleAndSave(inputPath, outputPath); err != nil {
		t.Fatalf("UpscaleAndSave failed: %v", err)
	}

	// Verify output
	resultImg, err := LoadImage(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}

	if resultImg.W != 50 || resultImg.H != 50 {
		t.Errorf("Expected 50x50, got %dx%d", resultImg.W, resultImg.H)
	}
}
