package pngmask

import (
	"centerline-go/internal/domain"
	"image"
	"image/png"
	"os"
)

func DecodeFile(path string, threshold uint8) (domain.Mask, error) {
	f, err := os.Open(path)
	if err != nil {
		return domain.Mask{}, err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return domain.Mask{}, err
	}
	return DecodeImage(img, threshold), nil
}

func DecodeImage(img image.Image, threshold uint8) domain.Mask {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	data := make([]bool, w*h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			a8 := uint8(a >> 8)
			if a8 <= 32 {
				continue
			}
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)
			luma := uint8((299*uint32(r8) + 587*uint32(g8) + 114*uint32(b8)) / 1000)
			if luma < threshold {
				data[y*w+x] = true
			}
		}
	}
	return domain.Mask{W: w, H: h, Data: data}
}
