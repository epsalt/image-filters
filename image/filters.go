package image

import (
	"image"
	"image/color"
	"math"
)

type RGBA struct {
	*image.RGBA
}

type qerror struct {
	r, g, b float64
}

func (e1 qerror) add(e2 qerror) qerror {
	return qerror{
		e1.r + e2.r,
		e1.g + e2.g,
		e1.b + e2.b,
	}
}

func (e qerror) scale(n float64) qerror {
	return qerror{
		e.r * n,
		e.g * n,
		e.b * n,
	}
}

func (e qerror) apply(c color.Color) color.Color {
	cn := color.NRGBAModel.Convert(c).(color.NRGBA)

	r := clamp(float64(cn.R) + e.r)
	g := clamp(float64(cn.G) + e.g)
	b := clamp(float64(cn.B) + e.b)

	return color.NRGBA{r, g, b, cn.A}
}

func newQerror(old, new color.Color) qerror {
	o := color.NRGBAModel.Convert(old).(color.NRGBA)
	n := color.NRGBAModel.Convert(new).(color.NRGBA)

	return qerror{
		float64(int(o.R) - int(n.R)),
		float64(int(o.G) - int(n.G)),
		float64(int(o.B) - int(n.B)),
	}
}

func clamp(band float64) uint8 {
	if band > 255 {
		band = 255
	} else if band < 0 {
		band = 0
	}
	return uint8(band)
}

// Greyscale converts an image to greyscale.
func (m *RGBA) Greyscale() {
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
func (m *RGBA) Dither(p color.Palette) {
	b := m.Bounds()
	coefficients := [4]float64{7.0 / 16.0, 3.0 / 16.0, 5.0 / 16.0, 1.0 / 16.0}

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

func gauss(sigma float64, dx int) float64 {
	x := float64(dx)

	t1 := 1.0 / math.Sqrt(2.0*math.Pi*math.Pow(sigma, 2))
	t2 := math.Pow(math.E, -1*math.Pow(x, 2)/2*math.Pow(sigma, 2))

	return t1 * t2
}

// Blur blurs an image using a gaussian filter.
func (m *RGBA) Blur(sigma float64, k int) {
	bounds := m.Bounds()
	filter := make([]float64, k*2+1)

	for i := 0; i < k+1; i++ {
		g := gauss(sigma, k-i)
		filter[i] = g
		filter[len(filter)-i-1] = g
	}

	sum := 0.0
	for i := 0; i < len(filter); i++ {
		sum += filter[i]
	}

	// First horizontal pass
	next := make([]color.NRGBA, bounds.Dx())
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var r, g, b float64
			curr := color.NRGBAModel.Convert(m.At(x, y)).(color.NRGBA)

			for kx := x - k; kx < x+k+1; kx++ {
				var bkx = kx

				switch {
				case kx < 0:
					bkx = 0
				case kx >= bounds.Max.X:
					bkx = bounds.Max.X - 1
				}

				kp := color.NRGBAModel.Convert(m.At(bkx, y)).(color.NRGBA)
				coeff := filter[k+x-kx]
				r += float64(kp.R) * coeff
				g += float64(kp.G) * coeff
				b += float64(kp.B) * coeff
			}
			new := color.NRGBA{
				clamp(r / sum),
				clamp(g / sum),
				clamp(b / sum),
				curr.A,;
			}
			next[x-bounds.Min.X] = new
		}
		for i := 0; i < len(next); i++ {
			m.Set(i+bounds.Min.X, y, next[i])
		}

	}

	// Second vertical pass
	next = make([]color.NRGBA, bounds.Dy())
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			var r, g, b float64
			curr := color.NRGBAModel.Convert(m.At(x, y)).(color.NRGBA)

			for ky := y - k; ky < y+k+1; ky++ {
				bky := ky

				switch {
				case bky < 0:
					bky = 0
				case ky >= bounds.Max.Y:
					bky = bounds.Max.Y - 1
				}

				coeff := filter[k+y-ky]
				kp := color.NRGBAModel.Convert(m.At(x, bky)).(color.NRGBA)
				r += float64(kp.R) * coeff
				g += float64(kp.G) * coeff
				b += float64(kp.B) * coeff
			}
			new := color.NRGBA{
				clamp(r / sum),
				clamp(g / sum),
				clamp(b / sum),
				curr.A,
			}
			next[y-bounds.Min.Y] = new
		}
		for j := 0; j < len(next); j++ {
			m.Set(x, j+bounds.Min.Y, next[j])
		}
	}
}
