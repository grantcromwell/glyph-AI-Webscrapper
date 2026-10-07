// Package retina is the shared front of the eye for the analysis spells: it
// turns any image into a square luminance grid of a chosen power-of-two size
// (box-averaged, so detail is summarised not dropped). The decomposition spells
// (wavelet, spectrum) read this grid; the math then lives in internal/math.
package vision

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

// Load reads an image and returns an n×n grid of luminance (0..255), box-averaged
// from the source. n should be a power of two for the transforms.
func Load(path string, n int) ([][]float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	grid := make([][]float64, n)
	for gy := 0; gy < n; gy++ {
		grid[gy] = make([]float64, n)
		y0, y1 := gy*h/n, (gy+1)*h/n
		if y1 <= y0 {
			y1 = y0 + 1
		}
		for gx := 0; gx < n; gx++ {
			x0, x1 := gx*w/n, (gx+1)*w/n
			if x1 <= x0 {
				x1 = x0 + 1
			}
			var sum float64
			cnt := 0
			for y := y0; y < y1 && y < h; y++ {
				for x := x0; x < x1 && x < w; x++ {
					r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
					sum += 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(bl>>8)
					cnt++
				}
			}
			if cnt > 0 {
				grid[gy][gx] = sum / float64(cnt)
			}
		}
	}
	return grid, nil
}