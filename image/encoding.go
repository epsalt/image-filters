package image

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/draw"
	"image/png"
	"strings"
)

// Encode encodes an image to a base64 string.
func Encode(m RGBA) string {
	var buf bytes.Buffer
	png.Encode(&buf, m)
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// Decode converts a base64 string into an editable image.
func Decode(data string) (RGBA, error) {
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	m, _, err := image.Decode(reader)
	if err != nil {
		return RGBA{}, err
	}

	rgba, ok := m.(*image.RGBA)
	if !ok {
		b := m.Bounds()
		rgba = image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		draw.Draw(rgba, rgba.Bounds(), m, b.Min, draw.Src)
	}
	return RGBA{rgba}, nil
}
