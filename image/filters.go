package image

import (
	"image"
	"image/color"
)

type RGBA struct {
	*image.RGBA
}

// Greyscale converts an image to greyscale.
func (m RGBA) Greyscale() {
	b := m.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, b, _ := m.At(x, y).RGBA()
			new := 0.3*float64(r) + 0.59*float64(g) + 0.11*float64(b)
			m.Set(x, y, color.Gray16{uint16(new)})
		}
	}
}
