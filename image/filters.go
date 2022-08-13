package image

import (
	"image"
	"image/color"
)

type RGBA struct {
	*image.RGBA
}

type qerror struct {
	r, g, b float32
}

func (e1 qerror) add(e2 qerror) qerror {
	return qerror{
		e1.r + e2.r,
		e1.g + e2.g,
		e1.b + e2.b,
	}
}

func (e qerror) scale(n float32) qerror {
	return qerror{
		e.r * n,
		e.g * n,
		e.b * n,
	}
}

func (e qerror) apply(c color.Color) color.Color {
	cn := color.NRGBAModel.Convert(c).(color.NRGBA)

	r := clamp(float32(cn.R) + e.r)
	g := clamp(float32(cn.G) + e.g)
	b := clamp(float32(cn.B) + e.b)

	return color.NRGBA{r, g, b, cn.A}
}

func newQerror(old, new color.Color) qerror {
	o := color.NRGBAModel.Convert(old).(color.NRGBA)
	n := color.NRGBAModel.Convert(new).(color.NRGBA)

	return qerror{
		float32(int(o.R) - int(n.R)),
		float32(int(o.G) - int(n.G)),
		float32(int(o.B) - int(n.B)),
	}
}

func clamp(band float32) uint8 {
	if band > 255 {
		band = 255
	} else if band < 0 {
		band = 0
	}
	return uint8(band)
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

// Dither coverts an image to use a color palette with the
// Floyd-Steinberg dithering algorithm.
func (m RGBA) Dither(p color.Palette) {
	b := m.Bounds()
	coefficients := [4]float32{7.0 / 16.0, 3.0 / 16.0, 5.0 / 16.0, 1.0 / 16.0}

	errors := make([][]qerror, b.Min.Y+b.Dy())
	for y := b.Min.Y; y < b.Max.Y; y++ {
		errors[y] = make([]qerror, b.Max.X+b.Dx())
	}

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			currentError := errors[y][x]

			oldPixel := currentError.apply(m.At(x, y))
			newPixel := p.Convert(oldPixel)
			m.Set(x, y, newPixel)

			propagatedError := newQerror(oldPixel, newPixel)
			for ci, neighbor := range [][]int{{y, x + 1}, {y + 1, x - 1}, {y + 1, x}, {y + 1, x + 1}} {
				j, i := neighbor[0], neighbor[1]

				if j >= b.Min.Y && j < b.Max.Y && i >= b.Min.X && i < b.Max.X {
					errors[j][i] = propagatedError.scale(coefficients[ci]).add(errors[j][i])
				}
			}
		}
	}
}
