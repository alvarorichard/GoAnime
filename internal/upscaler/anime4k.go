// Package upscaler provides anime video upscaling using the Anime4K algorithm
// Based on https://github.com/TianZerL/Anime4KGo (MIT License)
// Algorithm based on bloc97's Anime4K version 0.9
package upscaler

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"github.com/disintegration/gift"
)

// Anime4KOptions contains configuration for the Anime4K upscaling algorithm
type Anime4KOptions struct {
	// Passes is the number of processing passes (default: 2)
	Passes int
	// StrengthColor controls line thinning, range 0-1 (default: 0.3333)
	StrengthColor float64
	// StrengthGradient controls sharpening, range 0-1 (default: 1.0)
	StrengthGradient float64
	// FastMode uses faster but lower quality processing
	FastMode bool
	// ScaleFactor is the upscale multiplier (Anime4K always uses 2x)
	ScaleFactor int
}

// DefaultOptions returns the default Anime4K options
func DefaultOptions() Anime4KOptions {
	return Anime4KOptions{
		Passes:           2,
		StrengthColor:    1.0 / 3.0,
		StrengthGradient: 1.0,
		FastMode:         false,
		ScaleFactor:      2,
	}
}

// FastOptions returns fast mode options for real-time processing
func FastOptions() Anime4KOptions {
	return Anime4KOptions{
		Passes:           1,
		StrengthColor:    1.0 / 3.0,
		StrengthGradient: 1.0,
		FastMode:         true,
		ScaleFactor:      2,
	}
}

// HighQualityOptions returns high quality options for offline processing
func HighQualityOptions() Anime4KOptions {
	return Anime4KOptions{
		Passes:           4,
		StrengthColor:    0.4,
		StrengthGradient: 1.0,
		FastMode:         false,
		ScaleFactor:      2,
	}
}

// Anime4KImage represents an image for Anime4K processing
type Anime4KImage struct {
	W       int
	H       int
	FmtType string
	data    image.Image
}

// LoadImage loads an image file and returns an Anime4KImage
func LoadImage(src string) (*Anime4KImage, error) {
	// #nosec G304 -- src is user-provided file path for upscaling
	f, err := os.ReadFile(filepath.Clean(src))
	if err != nil {
		return nil, fmt.Errorf("failed to read image file: %w", err)
	}
	r := bytes.NewReader(f)
	img, fmtType, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	b := img.Bounds()
	return &Anime4KImage{W: b.Dx(), H: b.Dy(), FmtType: fmtType, data: img}, nil
}

// NewImageFromImage creates an Anime4KImage from an existing image.Image
func NewImageFromImage(img image.Image) *Anime4KImage {
	b := img.Bounds()
	return &Anime4KImage{W: b.Dx(), H: b.Dy(), FmtType: "png", data: img}
}

// Process applies the Anime4K algorithm to the image
func (img *Anime4KImage) Process(passes int, sc, sg float64, fastMode bool) {
	// Resize using cubic resampling (2x upscale)
	g := gift.New(gift.Resize(img.W*2, img.H*2, gift.CubicResampling))
	dstImg := image.NewRGBA(g.Bounds(img.data.Bounds()))
	g.Draw(dstImg, img.data)

	for i := 0; i < passes; i++ {
		getGray(dstImg)
		pushColor(dstImg, sc)
		getGradient(dstImg, fastMode)
		pushGradient(dstImg, sg)
	}

	img.data = dstImg
	img.W = dstImg.Bounds().Dx()
	img.H = dstImg.Bounds().Dy()
}

// GetImage returns the processed image
func (img *Anime4KImage) GetImage() image.Image {
	return img.data
}

// SaveImage saves the processed image to a file
func (img *Anime4KImage) SaveImage(filename string) error {
	// #nosec G304 -- filename is user-provided output path for upscaled image
	f, err := os.Create(filepath.Clean(filename))
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close file: %w", cerr)
		}
	}()

	if err := png.Encode(f, img.data); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}
	return nil
}

// getGray computes the grayscale of the image and stores it in the Alpha channel
func getGray(img *image.RGBA) {
	changeEachPixel(img, func(x, y int, p *color.RGBA) color.RGBA {
		g := 0.299*float64(p.R) + 0.587*float64(p.G) + 0.114*float64(p.B)
		return color.RGBA{p.R, p.G, p.B, uint8(g)}
	})
}

// getGradient computes the gradient of the image and stores it in the Alpha channel
func getGradient(img *image.RGBA, fastMode bool) {
	maxX := img.Bounds().Dx() - 1
	maxY := img.Bounds().Dy() - 1

	if fastMode {
		changeEachPixel(img, func(x, y int, p *color.RGBA) color.RGBA {
			if x == 0 || x == maxX || y == 0 || y == maxY {
				return *p
			}

			G := (math.Abs(float64(img.RGBAAt(x-1, y+1).A)+2*float64(img.RGBAAt(x, y+1).A)+float64(img.RGBAAt(x+1, y+1).A)-float64(img.RGBAAt(x-1, y-1).A)-2*float64(img.RGBAAt(x, y-1).A)-float64(img.RGBAAt(x+1, y-1).A)) +
				math.Abs(float64(img.RGBAAt(x-1, y-1).A)+2*float64(img.RGBAAt(x-1, y).A)+float64(img.RGBAAt(x-1, y+1).A)-float64(img.RGBAAt(x+1, y-1).A)-2*float64(img.RGBAAt(x+1, y).A)-float64(img.RGBAAt(x+1, y+1).A)))

			rst := unFloat(G / 2)
			return color.RGBA{p.R, p.G, p.B, 255 - rst}
		})
	} else {
		changeEachPixel(img, func(x, y int, p *color.RGBA) color.RGBA {
			if x == 0 || x == maxX || y == 0 || y == maxY {
				return *p
			}

			gy := float64(img.RGBAAt(x-1, y+1).A) + 2*float64(img.RGBAAt(x, y+1).A) + float64(img.RGBAAt(x+1, y+1).A) -
				float64(img.RGBAAt(x-1, y-1).A) - 2*float64(img.RGBAAt(x, y-1).A) - float64(img.RGBAAt(x+1, y-1).A)
			gx := float64(img.RGBAAt(x-1, y-1).A) + 2*float64(img.RGBAAt(x-1, y).A) + float64(img.RGBAAt(x-1, y+1).A) -
				float64(img.RGBAAt(x+1, y-1).A) - 2*float64(img.RGBAAt(x+1, y).A) - float64(img.RGBAAt(x+1, y+1).A)

			G := math.Sqrt(gy*gy + gx*gx)

			rst := unFloat(G)
			return color.RGBA{p.R, p.G, p.B, 255 - rst}
		})
	}
}

// pushColor makes the linework of the image thinner guided by the grayscale in Alpha channel
func pushColor(dst *image.RGBA, strength float64) {
	getLightest := func(mc, a, b, c *color.RGBA) {
		mc.R = unFloat(float64(mc.R)*(1.0-strength) + ((float64(a.R)+float64(b.R)+float64(c.R))/3.0)*strength)
		mc.G = unFloat(float64(mc.G)*(1.0-strength) + ((float64(a.G)+float64(b.G)+float64(c.G))/3.0)*strength)
		mc.B = unFloat(float64(mc.B)*(1.0-strength) + ((float64(a.B)+float64(b.B)+float64(c.B))/3.0)*strength)
		mc.A = unFloat(float64(mc.A)*(1.0-strength) + ((float64(a.A)+float64(b.A)+float64(c.A))/3.0)*strength)
	}

	changeEachPixel(dst, func(x, y int, p *color.RGBA) color.RGBA {
		xn, xp, yn, yp := -1, 1, -1, 1
		if x == 0 {
			xn = 0
		} else if x == dst.Bounds().Dx()-1 {
			xp = 0
		}
		if y == 0 {
			yn = 0
		} else if y == dst.Bounds().Dy()-1 {
			yp = 0
		}

		tl, tc, tr := dst.RGBAAt(x+xn, y+yn), dst.RGBAAt(x, y+yn), dst.RGBAAt(x+xp, y+yn)
		ml, mc, mr := dst.RGBAAt(x+xn, y), *p, dst.RGBAAt(x+xp, y)
		bl, bc, br := dst.RGBAAt(x+xn, y+yp), dst.RGBAAt(x, y+yp), dst.RGBAAt(x+xn, y+yp)

		// top and bottom
		maxD := maxUint8(bl.A, bc.A, br.A)
		minL := minUint8(tl.A, tc.A, tr.A)
		if minL > mc.A && mc.A > maxD {
			getLightest(&mc, &tl, &tc, &tr)
		} else {
			maxD = maxUint8(tl.A, tc.A, tr.A)
			minL = minUint8(bl.A, bc.A, br.A)
			if minL > mc.A && mc.A > maxD {
				getLightest(&mc, &bl, &bc, &br)
			}
		}

		// subdiagonal
		maxD = maxUint8(ml.A, mc.A, bc.A)
		minL = minUint8(mr.A, tc.A, tr.A)
		if minL > maxD {
			getLightest(&mc, &mr, &tc, &tr)
		} else {
			maxD = maxUint8(mc.A, mr.A, tc.A)
			minL = minUint8(bl.A, ml.A, bc.A)
			if minL > maxD {
				getLightest(&mc, &bl, &ml, &bc)
			}
		}

		// left and right
		maxD = maxUint8(tl.A, ml.A, bl.A)
		minL = minUint8(tr.A, mr.A, br.A)
		if minL > mc.A && mc.A > maxD {
			getLightest(&mc, &tr, &mr, &br)
		} else {
			maxD = maxUint8(tr.A, mr.A, br.A)
			minL = minUint8(tl.A, ml.A, bl.A)
			if minL > mc.A && mc.A > maxD {
				getLightest(&mc, &tl, &ml, &bl)
			}
		}

		// diagonal
		maxD = maxUint8(ml.A, mc.A, tc.A)
		minL = minUint8(mr.A, br.A, bc.A)
		if minL > maxD {
			getLightest(&mc, &mr, &br, &tc)
		} else {
			maxD = maxUint8(mc.A, mr.A, bc.A)
			minL = minUint8(tc.A, ml.A, tl.A)
			if minL > maxD {
				getLightest(&mc, &tc, &ml, &tl)
			}
		}

		return mc
	})
}

// pushGradient makes the linework of the image sharper guided by the gradient in Alpha channel
func pushGradient(dst *image.RGBA, strength float64) {
	getLightest := func(mc, a, b, c *color.RGBA) color.RGBA {
		mc.R = unFloat(float64(mc.R)*(1.0-strength) + ((float64(a.R)+float64(b.R)+float64(c.R))/3.0)*strength)
		mc.G = unFloat(float64(mc.G)*(1.0-strength) + ((float64(a.G)+float64(b.G)+float64(c.G))/3.0)*strength)
		mc.B = unFloat(float64(mc.B)*(1.0-strength) + ((float64(a.B)+float64(b.B)+float64(c.B))/3.0)*strength)
		mc.A = 255
		return *mc
	}

	changeEachPixel(dst, func(x, y int, p *color.RGBA) color.RGBA {
		xn, xp, yn, yp := -1, 1, -1, 1
		if x == 0 {
			xn = 0
		} else if x == dst.Bounds().Dx()-1 {
			xp = 0
		}
		if y == 0 {
			yn = 0
		} else if y == dst.Bounds().Dy()-1 {
			yp = 0
		}

		tl, tc, tr := dst.RGBAAt(x+xn, y+yn), dst.RGBAAt(x, y+yn), dst.RGBAAt(x+xp, y+yn)
		ml, mc, mr := dst.RGBAAt(x+xn, y), *p, dst.RGBAAt(x+xp, y)
		bl, bc, br := dst.RGBAAt(x+xn, y+yp), dst.RGBAAt(x, y+yp), dst.RGBAAt(x+xn, y+yp)

		// top and bottom
		maxD := maxUint8(bl.A, bc.A, br.A)
		minL := minUint8(tl.A, tc.A, tr.A)
		if minL > mc.A && mc.A > maxD {
			return getLightest(&mc, &tl, &tc, &tr)
		}
		maxD = maxUint8(tl.A, tc.A, tr.A)
		minL = minUint8(bl.A, bc.A, br.A)
		if minL > mc.A && mc.A > maxD {
			return getLightest(&mc, &bl, &bc, &br)
		}

		// subdiagonal
		maxD = maxUint8(ml.A, mc.A, bc.A)
		minL = minUint8(mr.A, tc.A, tr.A)
		if minL > maxD {
			return getLightest(&mc, &mr, &tc, &tr)
		}
		maxD = maxUint8(mc.A, mr.A, tc.A)
		minL = minUint8(bl.A, ml.A, bc.A)
		if minL > maxD {
			return getLightest(&mc, &bl, &ml, &bc)
		}

		// left and right
		maxD = maxUint8(tl.A, ml.A, bl.A)
		minL = minUint8(tr.A, mr.A, br.A)
		if minL > mc.A && mc.A > maxD {
			return getLightest(&mc, &tr, &mr, &br)
		}
		maxD = maxUint8(tr.A, mr.A, br.A)
		minL = minUint8(tl.A, ml.A, bl.A)
		if minL > mc.A && mc.A > maxD {
			return getLightest(&mc, &tl, &ml, &bl)
		}

		// diagonal
		maxD = maxUint8(ml.A, mc.A, tc.A)
		minL = minUint8(mr.A, br.A, bc.A)
		if minL > maxD {
			return getLightest(&mc, &mr, &br, &tc)
		}
		maxD = maxUint8(mc.A, mr.A, bc.A)
		minL = minUint8(tc.A, ml.A, tl.A)
		if minL > maxD {
			return getLightest(&mc, &tc, &ml, &tl)
		}

		mc.A = 255
		return mc
	})
}

// changeEachPixel traverses all pixels and applies the function to each
func changeEachPixel(img *image.RGBA, fun func(x, y int, p *color.RGBA) color.RGBA) {
	imgInfo := img.Bounds()
	temp := image.NewRGBA(imgInfo)
	dx, dy := imgInfo.Dx(), imgInfo.Dy()
	for i := 0; i < dx; i++ {
		for j := 0; j < dy; j++ {
			p := img.RGBAAt(i, j)
			temp.SetRGBA(i, j, fun(i, j, &p))
		}
	}
	*img = *temp
}

// unFloat converts float64 to uint8, range 0-255
func unFloat(n float64) uint8 {
	n += 0.5
	if n >= 255 {
		return 255
	} else if n <= 0 {
		return 0
	}
	return uint8(n)
}

func maxUint8(a, b, c uint8) uint8 {
	if a > b && a > c {
		return a
	} else if b > c {
		return b
	}
	return c
}

func minUint8(a, b, c uint8) uint8 {
	if a < b && a < c {
		return a
	} else if b < c {
		return b
	}
	return c
}

// Anime4KUpscaler performs Anime4K upscaling on images
type Anime4KUpscaler struct {
	opts Anime4KOptions
}

// NewAnime4KUpscaler creates a new Anime4K upscaler with the given options
func NewAnime4KUpscaler(opts Anime4KOptions) *Anime4KUpscaler {
	return &Anime4KUpscaler{opts: opts}
}

// Close cleans up resources (currently a no-op but maintains interface compatibility)
func (u *Anime4KUpscaler) Close() {}

// UpscaleImage upscales a single image using the Anime4K algorithm
func (u *Anime4KUpscaler) UpscaleImage(src image.Image) (image.Image, error) {
	// Create Anime4K image wrapper from the source image
	img := NewImageFromImage(src)

	// Process using the Anime4K algorithm
	img.Process(u.opts.Passes, u.opts.StrengthColor, u.opts.StrengthGradient, u.opts.FastMode)

	// Return the processed image
	return img.GetImage(), nil
}

// UpscaleImageParallel upscales an image - wrapper for UpscaleImage
func (u *Anime4KUpscaler) UpscaleImageParallel(src image.Image) (image.Image, error) {
	return u.UpscaleImage(src)
}

// UpscaleAndSave loads, upscales, and saves an image file
func (u *Anime4KUpscaler) UpscaleAndSave(inputPath, outputPath string) error {
	img, err := LoadImage(inputPath)
	if err != nil {
		return err
	}

	img.Process(u.opts.Passes, u.opts.StrengthColor, u.opts.StrengthGradient, u.opts.FastMode)

	return img.SaveImage(outputPath)
}

// UpscaleImageFile is a convenience function to upscale a single image file
func UpscaleImageFile(inputPath, outputPath string, opts Anime4KOptions) error {
	upscaler := NewAnime4KUpscaler(opts)
	return upscaler.UpscaleAndSave(inputPath, outputPath)
}

// GetUpscaledDimensions returns the dimensions after upscaling
// Anime4K always upscales by 2x
func GetUpscaledDimensions(width, height int) (int, int) {
	return width * 2, height * 2
}
