package upscaler

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnFloat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   float64
		want uint8
	}{
		{0, 0},
		{-1, 0},
		{255, 255},
		{300, 255},
		{100, 100},
		{99.5, 100},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, unFloat(tt.in))
	}
}

func TestMaxUint8(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a, b, c, want uint8
	}{
		{1, 2, 3, 3},
		{5, 2, 3, 5},
		{1, 5, 3, 5},
		{2, 2, 2, 2},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, maxUint8(tt.a, tt.b, tt.c))
	}
}

func TestMinUint8(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a, b, c, want uint8
	}{
		{1, 2, 3, 1},
		{5, 2, 3, 2},
		{5, 5, 3, 3},
		{2, 2, 2, 2},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, minUint8(tt.a, tt.b, tt.c))
	}
}

func makeTestPNG(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "in.png")
	f, err := os.Create(p)
	require.NoError(t, err)
	defer f.Close()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := range 4 {
		for x := range 4 {
			img.Set(x, y, color.RGBA{uint8(x * 60), uint8(y * 60), 200, 255})
		}
	}
	require.NoError(t, png.Encode(f, img))
	return p
}

func TestLoadImage_PNGFromTempDir(t *testing.T) {
	t.Parallel()
	p := makeTestPNG(t)
	a, err := LoadImage(p)
	require.NoError(t, err)
	assert.Equal(t, 4, a.W)
	assert.Equal(t, 4, a.H)
}

func TestNewImageFromImage(t *testing.T) {
	t.Parallel()
	img := image.NewRGBA(image.Rect(0, 0, 6, 8))
	a := NewImageFromImage(img)
	assert.Equal(t, 6, a.W)
	assert.Equal(t, 8, a.H)
	assert.Equal(t, "png", a.FmtType)
}

func TestAnime4KImage_GetImage(t *testing.T) {
	t.Parallel()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	a := NewImageFromImage(img)
	assert.NotNil(t, a.GetImage())
}

func TestAnime4KImage_SaveImage(t *testing.T) {
	t.Parallel()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	a := NewImageFromImage(img)
	out := filepath.Join(t.TempDir(), "out.png")
	require.NoError(t, a.SaveImage(out))
	info, err := os.Stat(out)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}
